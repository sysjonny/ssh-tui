package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type appConfig struct {
	SSHConfigPath string
	PasswordEnv   string
}

func loadConfig(path string) (appConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return appConfig{}, fmt.Errorf("find home directory: %w", err)
	}
	config := appConfig{SSHConfigPath: filepath.Join(home, ".ssh", "config")}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return appConfig{}, fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return appConfig{}, fmt.Errorf("%s:%d: expected key = value", path, lineNumber)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' || value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
		switch key {
		case "ssh_config":
			if value == "" {
				return appConfig{}, fmt.Errorf("%s:%d: ssh_config cannot be empty", path, lineNumber)
			}
			config.SSHConfigPath = expandHome(value, home)
		case "password_env":
			if value != "" && !validEnvName(value) {
				return appConfig{}, fmt.Errorf("%s:%d: password_env must be a valid environment variable name", path, lineNumber)
			}
			config.PasswordEnv = value
		default:
			return appConfig{}, fmt.Errorf("%s:%d: unknown setting %q", path, lineNumber, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return appConfig{}, fmt.Errorf("read config file: %w", err)
	}
	return config, nil
}

func expandHome(path, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func validEnvName(name string) bool {
	if name == "" || !(name[0] == '_' || name[0] >= 'A' && name[0] <= 'Z' || name[0] >= 'a' && name[0] <= 'z') {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !(c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
