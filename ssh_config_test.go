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
