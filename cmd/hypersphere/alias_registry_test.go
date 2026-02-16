// Path: cmd/hypersphere/alias_registry_test.go
// Description: Validate alias registry loading and command resolution behavior.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCommandAliasRegistryMissingFileReturnsEmpty(t *testing.T) {
	registry, err := loadCommandAliasRegistry(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("expected missing alias file to be ignored, got %v", err)
	}
	if resolved := registry.Resolve(":vmx"); resolved != ":vmx" {
		t.Fatalf("expected unknown alias to pass through unchanged, got %q", resolved)
	}
}

func TestLoadCommandAliasRegistryParsesAndResolvesAliases(t *testing.T) {
	aliasPath := filepath.Join(t.TempDir(), "aliases.yaml")
	content := "# aliases\nvmprod: :vm /prod\nhost: :host\n"
	if err := os.WriteFile(aliasPath, []byte(content), 0o600); err != nil {
		t.Fatalf("expected alias file write to succeed: %v", err)
	}
	registry, err := loadCommandAliasRegistry(aliasPath)
	if err != nil {
		t.Fatalf("expected alias file load to succeed: %v", err)
	}
	if resolved := registry.Resolve(":vmprod owner=team-a"); resolved != ":vm /prod owner=team-a" {
		t.Fatalf("expected alias command with arguments, got %q", resolved)
	}
	if resolved := registry.Resolve("!power-off"); resolved != "!power-off" {
		t.Fatalf("expected non-colon command to pass through unchanged, got %q", resolved)
	}
}

func TestLoadCommandAliasRegistryRejectsInvalidEntries(t *testing.T) {
	aliasPath := filepath.Join(t.TempDir(), "aliases.yaml")
	content := "broken-entry\n"
	if err := os.WriteFile(aliasPath, []byte(content), 0o600); err != nil {
		t.Fatalf("expected alias file write to succeed: %v", err)
	}
	if _, err := loadCommandAliasRegistry(aliasPath); err == nil {
		t.Fatalf("expected invalid alias entry to fail")
	}
}

func TestDefaultAliasRegistryPathUsesEnvironmentOverride(t *testing.T) {
	expected := filepath.Join(t.TempDir(), "aliases.yaml")
	t.Setenv(aliasRegistryEnvPath, expected)
	path, err := defaultAliasRegistryPath()
	if err != nil {
		t.Fatalf("expected env override path to resolve, got %v", err)
	}
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}
