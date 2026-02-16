// Path: cmd/hypersphere/alias_registry.go
// Description: Load and resolve prompt command aliases from a registry file.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const aliasRegistryEnvPath = "HYPERSPHERE_ALIASES_FILE"

type commandAliasRegistry struct {
	aliases map[string]string
}

func loadDefaultCommandAliasRegistry() (commandAliasRegistry, error) {
	path, err := defaultAliasRegistryPath()
	if err != nil {
		return commandAliasRegistry{}, err
	}
	return loadCommandAliasRegistry(path)
}

func defaultAliasRegistryPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv(aliasRegistryEnvPath)); override != "" {
		return override, nil
	}
	paths, err := infoPaths()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(paths["config"]), "aliases.yaml"), nil
}

func loadCommandAliasRegistry(path string) (commandAliasRegistry, error) {
	registry := commandAliasRegistry{aliases: map[string]string{}}
	if strings.TrimSpace(path) == "" {
		return registry, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return registry, nil
		}
		return commandAliasRegistry{}, err
	}
	aliases, err := parseAliasRegistry(string(content))
	if err != nil {
		return commandAliasRegistry{}, err
	}
	registry.aliases = aliases
	return registry, nil
}

func parseAliasRegistry(content string) (map[string]string, error) {
	aliases := map[string]string{}
	for index, line := range strings.Split(content, "\n") {
		name, command, ok, err := parseAliasLine(line, index+1)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		aliases[name] = command
	}
	return aliases, nil
}

func parseAliasLine(line string, lineNumber int) (string, string, bool, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false, nil
	}
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", "", false, fmt.Errorf("invalid alias entry on line %d", lineNumber)
	}
	name := strings.ToLower(strings.TrimSpace(parts[0]))
	command := strings.TrimSpace(parts[1])
	if name == "" || command == "" || !strings.HasPrefix(command, ":") {
		return "", "", false, fmt.Errorf("invalid alias entry on line %d", lineNumber)
	}
	return name, command, true, nil
}

func (r commandAliasRegistry) Resolve(line string) string {
	alias, arguments, ok := parseAliasInvocation(line)
	if !ok {
		return line
	}
	command, exists := r.aliases[alias]
	if !exists {
		return line
	}
	if arguments == "" {
		return command
	}
	return command + " " + arguments
}

func parseAliasInvocation(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ":") {
		return "", "", false
	}
	payload := strings.TrimSpace(strings.TrimPrefix(trimmed, ":"))
	if payload == "" {
		return "", "", false
	}
	parts := strings.Fields(payload)
	name := strings.ToLower(parts[0])
	arguments := strings.TrimSpace(strings.TrimPrefix(payload, parts[0]))
	return name, arguments, true
}
