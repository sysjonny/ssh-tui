package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverHosts(t *testing.T) {
	directory := t.TempDir()
	include := filepath.Join(directory, "hosts")
	if err := os.WriteFile(include, []byte("Host included\n  HostName server.example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(directory, "config")
	config := "Include hosts\nHost production staging *.internal !ignored\nHost staging\nHost *.example.com\n"
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}

	hosts, err := discoverHosts(configPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"included", "production", "staging"}
	if !reflect.DeepEqual(hosts, want) {
		t.Errorf("discoverHosts() = %v, want %v", hosts, want)
	}
}

func TestSetEnvironmentReplacesExistingValues(t *testing.T) {
	got := setEnvironment([]string{"PATH=/bin", "SSH_ASKPASS=old"}, map[string]string{
		"SSH_ASKPASS":      "new",
		"SSH_TUI_PASSWORD": "secret",
	})
	values := make(map[string]string)
	for _, entry := range got {
		name, value, _ := strings.Cut(entry, "=")
		values[name] = value
	}
	if values["SSH_ASKPASS"] != "new" || values["SSH_TUI_PASSWORD"] != "secret" || values["PATH"] != "/bin" {
		t.Errorf("setEnvironment() produced unexpected values: %v", values)
	}
}

func TestConnectPassesPasswordThroughAskpass(t *testing.T) {
	directory := t.TempDir()
	sshPath := filepath.Join(directory, "ssh")
	passwordPath := filepath.Join(directory, "password")
	sourcePath := filepath.Join(directory, "source")
	sshScript := "#!/bin/sh\nprintf '%s' \"${SSH_TUI_PASSWORD-unset}\" > \"$SSH_TUI_TEST_SOURCE_FILE\"\n\"$SSH_ASKPASS\" > \"$SSH_TUI_TEST_PASSWORD_FILE\"\n"
	if err := os.WriteFile(sshPath, []byte(sshScript), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	t.Setenv("SSH_TUI_TEST_PASSWORD_FILE", passwordPath)
	t.Setenv("SSH_TUI_TEST_SOURCE_FILE", sourcePath)
	t.Setenv("SSH_TUI_PASSWORD", "test-password")

	err := connect("test-host", appConfig{
		SSHConfigPath: filepath.Join(directory, "ssh_config"),
		PasswordEnv:   "SSH_TUI_PASSWORD",
	})
	if err != nil {
		t.Fatal(err)
	}
	password, err := os.ReadFile(passwordPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(password) != "test-password\n" {
		t.Errorf("askpass returned %q, want test-password", password)
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(source) != "unset" {
		t.Errorf("source password variable remained in ssh environment: %q", source)
	}
}
