package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "config")
	if err := os.WriteFile(configPath, []byte("ssh_config = ~/.ssh/work\npassword_env = SSH_WORK_PASSWORD\n"), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if config.SSHConfigPath != filepath.Join(home, ".ssh", "work") {
		t.Errorf("SSHConfigPath = %q, want home-relative SSH config", config.SSHConfigPath)
	}
	if config.PasswordEnv != "SSH_WORK_PASSWORD" {
		t.Errorf("PasswordEnv = %q, want SSH_WORK_PASSWORD", config.PasswordEnv)
	}
}

func TestLoadConfigRejectsInvalidPasswordEnv(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(configPath, []byte("password_env = SSH-PASSWORD\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(configPath); err == nil {
		t.Fatal("loadConfig accepted an invalid environment variable name")
	}
}
