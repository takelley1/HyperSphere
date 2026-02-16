// Path: cmd/hypersphere/plugin_registry_test.go
// Description: Validate plugin registry loading and schema checks before activation.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPluginRegistryMissingFileReturnsEmpty(t *testing.T) {
	registry, err := loadPluginRegistry(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("expected missing plugin file to be ignored, got %v", err)
	}
	if len(registry.entries) != 0 {
		t.Fatalf("expected no plugins for missing file, got %d", len(registry.entries))
	}
}

func TestLoadPluginRegistryLoadsValidEntries(t *testing.T) {
	pluginPath := filepath.Join(t.TempDir(), "plugins.json")
	content := `[
		{"name":"Drain Host","command":"drain-host.sh","scopes":["host"],"shortcut":"D"},
		{"name":"Rebalance DS","command":"rebalance-ds.sh","scopes":["datastore"]}
	]`
	if err := os.WriteFile(pluginPath, []byte(content), 0o600); err != nil {
		t.Fatalf("expected plugin file write to succeed: %v", err)
	}
	registry, err := loadPluginRegistry(pluginPath)
	if err != nil {
		t.Fatalf("expected plugin file load to succeed: %v", err)
	}
	if len(registry.entries) != 2 {
		t.Fatalf("expected two plugin entries, got %d", len(registry.entries))
	}
}

func TestLoadPluginRegistryRejectsInvalidEntryWithFieldPath(t *testing.T) {
	pluginPath := filepath.Join(t.TempDir(), "plugins.json")
	content := `[{"name":"Broken","command":"","scopes":["vm"]}]`
	if err := os.WriteFile(pluginPath, []byte(content), 0o600); err != nil {
		t.Fatalf("expected plugin file write to succeed: %v", err)
	}
	_, err := loadPluginRegistry(pluginPath)
	if err == nil {
		t.Fatalf("expected schema validation error for invalid plugin entry")
	}
	if !strings.Contains(err.Error(), "plugins[0].command") {
		t.Fatalf("expected failing field path in error, got %v", err)
	}
}

func TestDefaultPluginRegistryPathUsesEnvironmentOverride(t *testing.T) {
	expected := filepath.Join(t.TempDir(), "plugins.json")
	t.Setenv(pluginRegistryEnvPath, expected)
	path, err := defaultPluginRegistryPath()
	if err != nil {
		t.Fatalf("expected env override path to resolve, got %v", err)
	}
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestVisiblePluginsForScopeFiltersToAllowedScopes(t *testing.T) {
	registry := pluginRegistry{
		entries: []pluginEntry{
			{Name: "VM Only", Command: "vm.sh", Scopes: []string{"vm"}},
			{Name: "Host Only", Command: "host.sh", Scopes: []string{"host"}},
			{Name: "All Views", Command: "all.sh", Scopes: []string{"all"}},
		},
	}
	visible := visiblePluginsForScope(registry, "vm")
	if len(visible) != 2 {
		t.Fatalf("expected two visible plugins for vm scope, got %d", len(visible))
	}
	if visible[0].Name != "VM Only" || visible[1].Name != "All Views" {
		t.Fatalf("unexpected visible plugin set for vm scope: %+v", visible)
	}
	hostVisible := visiblePluginsForScope(registry, "host")
	if len(hostVisible) != 2 {
		t.Fatalf("expected two visible plugins for host scope, got %d", len(hostVisible))
	}
	if hostVisible[0].Name != "Host Only" || hostVisible[1].Name != "All Views" {
		t.Fatalf("unexpected visible plugin set for host scope: %+v", hostVisible)
	}
}
