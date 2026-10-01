package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func discoverHosts(configPath string) ([]string, error) {
	hosts := make([]string, 0)
	seenHosts := make(map[string]bool)
	seenFiles := make(map[string]bool)
	if err := readSSHConfig(configPath, seenFiles, seenHosts, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func readSSHConfig(path string, seenFiles, seenHosts map[string]bool, hosts *[]string) error {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve SSH config path: %w", err)
	}
	if seenFiles[absolutePath] {
		return nil
	}
	seenFiles[absolutePath] = true

	file, err := os.Open(absolutePath)
	if err != nil {
		return fmt.Errorf("open SSH config %s: %w", absolutePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		for i, field := range fields {
			if strings.HasPrefix(field, "#") {
				fields = fields[:i]
				break
			}
		}
		if len(fields) < 2 {
			continue
		}
		switch strings.ToLower(fields[0]) {
		case "host":
			for _, pattern := range fields[1:] {
				if strings.ContainsAny(pattern, "*?![]") || seenHosts[pattern] {
					continue
				}
				seenHosts[pattern] = true
				*hosts = append(*hosts, pattern)
			}
		case "include":
			for _, include := range fields[1:] {
				include = expandHome(include, mustHome())
				if !filepath.IsAbs(include) {
					include = filepath.Join(filepath.Dir(absolutePath), include)
				}
				matches, err := filepath.Glob(include)
				if err != nil {
					return fmt.Errorf("%s:%d: invalid Include pattern: %w", absolutePath, lineNumber, err)
				}
				for _, match := range matches {
					if err := readSSHConfig(match, seenFiles, seenHosts, hosts); err != nil {
						return err
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read SSH config %s: %w", absolutePath, err)
	}
	return nil
}

func mustHome() string {
	home, _ := os.UserHomeDir()
	return home
}
