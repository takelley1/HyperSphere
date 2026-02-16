// Path: cmd/hypersphere/plugin_runner.go
// Description: Execute plugin commands with endpoint and object-selection environment contracts.
package main

import (
	"fmt"
	"strings"
)

type pluginCommandRunner interface {
	Run(command string, env map[string]string) error
}

func runPluginWithEnv(
	entry pluginEntry,
	selectedIDs []string,
	activeEndpoint string,
	runner pluginCommandRunner,
) error {
	command := strings.TrimSpace(entry.Command)
	if command == "" {
		return fmt.Errorf("invalid plugin command for %q", entry.Name)
	}
	if runner == nil {
		return fmt.Errorf("plugin runner unavailable for %q", entry.Name)
	}
	environment := map[string]string{
		"HYPERSPHERE_PLUGIN_NAME":     entry.Name,
		"HYPERSPHERE_ACTIVE_ENDPOINT": strings.TrimSpace(activeEndpoint),
		"HYPERSPHERE_SELECTED_IDS":    strings.Join(selectedIDs, ","),
	}
	return runner.Run(command, environment)
}
