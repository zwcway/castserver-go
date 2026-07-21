package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigFileToOptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "castserver.conf")

	content := []byte("# comment\nkey1=value1\nkey2 = value2\n\n[section]\nkey3=value3\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	options := map[string]string{"key1": "from-cli"}
	if err := parseConfigFileToOptions(path, options); err != nil {
		t.Fatalf("parse config file: %v", err)
	}

	if got := options["key1"]; got != "from-cli" {
		t.Fatalf("expected existing key to be preserved, got %q", got)
	}
	if got := options["key2"]; got != "value2" {
		t.Fatalf("expected key2 from config file, got %q", got)
	}
	if _, ok := options["key3"]; ok {
		t.Fatalf("expected section values to be ignored")
	}
}
