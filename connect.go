package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const askpassPasswordEnv = "SSH_TUI_ASKPASS_PASSWORD"

func connect(host string, config appConfig) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("find ssh executable: %w", err)
	}
	command := exec.Command(sshPath, "-F", config.SSHConfigPath, host)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if config.PasswordEnv != "" {
		password, ok := os.LookupEnv(config.PasswordEnv)
		if !ok || password == "" {
			return fmt.Errorf("environment variable %s is not set", config.PasswordEnv)
		}
		if strings.ContainsAny(password, "\r\n") {
			return fmt.Errorf("environment variable %s cannot contain a newline", config.PasswordEnv)
		}
		askpass, err := createAskpass()
		if err != nil {
			return err
		}
		defer os.Remove(askpass)

		environment := removeEnvironment(os.Environ(), config.PasswordEnv)
		command.Env = setEnvironment(environment, map[string]string{
			"SSH_ASKPASS":         askpass,
			"SSH_ASKPASS_REQUIRE": "force",
			askpassPasswordEnv:    password,
		})
		if _, ok := os.LookupEnv("DISPLAY"); !ok {
			command.Env = setEnvironment(command.Env, map[string]string{"DISPLAY": ":0"})
		}
	}
	if err := command.Run(); err != nil {
		return fmt.Errorf("ssh %s: %w", host, err)
	}
	return nil
}

func createAskpass() (string, error) {
	file, err := os.CreateTemp("", "ssh-tui-askpass-*")
	if err != nil {
		return "", fmt.Errorf("create SSH askpass helper: %w", err)
	}
	path := file.Name()
	if _, err := file.WriteString("#!/bin/sh\nprintf '%s\\n' \"$" + askpassPasswordEnv + "\"\n"); err != nil {
		file.Close()
		os.Remove(path)
		return "", fmt.Errorf("write SSH askpass helper: %w", err)
	}
	if err := file.Chmod(0700); err != nil {
		file.Close()
		os.Remove(path)
		return "", fmt.Errorf("secure SSH askpass helper: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("close SSH askpass helper: %w", err)
	}
	return filepath.Clean(path), nil
}

func removeEnvironment(environment []string, names ...string) []string {
	removed := make(map[string]bool, len(names))
	for _, name := range names {
		removed[name] = true
	}
	result := make([]string, 0, len(environment))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if !removed[name] {
			result = append(result, entry)
		}
	}
	return result
}

func setEnvironment(environment []string, values map[string]string) []string {
	result := make([]string, 0, len(environment)+len(values))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if _, replaced := values[name]; !replaced {
			result = append(result, entry)
		}
	}
	for name, value := range values {
		result = append(result, name+"="+value)
	}
	return result
}
