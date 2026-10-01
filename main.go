package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to ssh-tui config file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ssh-tui:", err)
		os.Exit(1)
	}
	if err := run(config, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "ssh-tui:", err)
		os.Exit(1)
	}
}

func run(config appConfig, input io.Reader, output, errorOutput io.Writer) error {
	hosts, err := discoverHosts(config.SSHConfigPath)
	if err != nil {
		return err
	}
	if len(hosts) == 0 {
		return fmt.Errorf("no connectable Host entries found in %s", config.SSHConfigPath)
	}

	scanner := bufio.NewScanner(input)
	for {
		fmt.Fprintln(output, "\nSSH hosts:")
		for i, host := range hosts {
			fmt.Fprintf(output, "  %d) %s\n", i+1, host)
		}
		fmt.Fprintln(output, "  q) quit")
		fmt.Fprint(output, "Select a host: ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		}

		selection := strings.TrimSpace(scanner.Text())
		if strings.EqualFold(selection, "q") {
			return nil
		}
		index, err := strconv.Atoi(selection)
		if err != nil || index < 1 || index > len(hosts) {
			fmt.Fprintln(errorOutput, "Choose a listed number or q.")
			continue
		}
		if err := connect(hosts[index-1], config); err != nil {
			fmt.Fprintf(errorOutput, "Connection failed: %v\n", err)
		}
	}
}

func defaultConfigPath() string {
	if path := os.Getenv("SSH_TUI_CONFIG"); path != "" {
		return path
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "ssh-tui", "config")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ssh-tui", "config")
}
