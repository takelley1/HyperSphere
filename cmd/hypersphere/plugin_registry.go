// Path: cmd/hypersphere/plugin_registry.go
// Description: Load and validate plugin registry entries before runtime activation.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const pluginRegistryEnvPath = "HYPERSPHERE_PLUGINS_FILE"

type pluginRegistry struct {
	entries []pluginEntry
}

type pluginEntry struct {
	Name     string   `json:"name"`
	Command  string   `json:"command"`
	Scopes   []string `json:"scopes"`
	Shortcut string   `json:"shortcut"`
}

func defaultPluginRegistryPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv(pluginRegistryEnvPath)); override != "" {
		return override, nil
	}
	paths, err := infoPaths()
	if err != nil {
		return "", err
	}
	return paths["plugins"], nil
}

func loadPluginRegistry(path string) (pluginRegistry, error) {
	registry := pluginRegistry{entries: []pluginEntry{}}
	if strings.TrimSpace(path) == "" {
		return registry, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return registry, nil
		}
		return pluginRegistry{}, err
	}
	entries, err := parsePluginRegistry(string(content))
	if err != nil {
		return pluginRegistry{}, err
	}
	registry.entries = entries
	return registry, nil
}

func loadDefaultPluginRegistry() (pluginRegistry, error) {
	path, err := defaultPluginRegistryPath()
	if err != nil {
		return pluginRegistry{}, err
	}
	return loadPluginRegistry(path)
}

func parsePluginRegistry(content string) ([]pluginEntry, error) {
	entries := []pluginEntry{}
	if strings.TrimSpace(content) == "" {
		return entries, nil
	}
	if err := json.Unmarshal([]byte(content), &entries); err != nil {
		return nil, err
	}
	for index, entry := range entries {
		if err := validatePluginEntry(entry, index); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func validatePluginEntry(entry pluginEntry, index int) error {
	if strings.TrimSpace(entry.Name) == "" {
		return pluginFieldError(index, "name")
	}
	if strings.TrimSpace(entry.Command) == "" {
		return pluginFieldError(index, "command")
	}
	if len(entry.Scopes) == 0 {
		return pluginFieldError(index, "scopes")
	}
	for _, scope := range entry.Scopes {
		if strings.TrimSpace(scope) == "" {
			return pluginFieldError(index, "scopes")
		}
	}
	return nil
}

func pluginFieldError(index int, field string) error {
	return fmt.Errorf("invalid plugin field plugins[%d].%s", index, field)
}
