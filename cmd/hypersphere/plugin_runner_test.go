// Path: cmd/hypersphere/plugin_runner_test.go
// Description: Validate plugin command runner environment contract behavior.
package main

import (
	"strings"
	"testing"
)

type fakePluginRunner struct {
	command string
	env     map[string]string
	err     error
}

func (f *fakePluginRunner) Run(command string, env map[string]string) error {
	f.command = command
	f.env = env
	return f.err
}

func TestRunPluginWithEnvPassesEndpointAndSelectedIDs(t *testing.T) {
	runner := &fakePluginRunner{}
	entry := pluginEntry{Name: "Drain Host", Command: "drain-host.sh"}
	err := runPluginWithEnv(entry, []string{"host-a", "host-b"}, "vc-primary", runner)
	if err != nil {
		t.Fatalf("expected plugin run to succeed: %v", err)
	}
	if runner.command != "drain-host.sh" {
		t.Fatalf("expected plugin command to run, got %q", runner.command)
	}
	if runner.env["HYPERSPHERE_ACTIVE_ENDPOINT"] != "vc-primary" {
		t.Fatalf("expected active endpoint env var, got %q", runner.env["HYPERSPHERE_ACTIVE_ENDPOINT"])
	}
	if runner.env["HYPERSPHERE_SELECTED_IDS"] != "host-a,host-b" {
		t.Fatalf("expected selected ids env var, got %q", runner.env["HYPERSPHERE_SELECTED_IDS"])
	}
	if !strings.Contains(runner.env["HYPERSPHERE_PLUGIN_NAME"], "Drain Host") {
		t.Fatalf("expected plugin name env var, got %q", runner.env["HYPERSPHERE_PLUGIN_NAME"])
	}
}

func TestRunPluginWithEnvRejectsMissingCommand(t *testing.T) {
	runner := &fakePluginRunner{}
	entry := pluginEntry{Name: "Broken", Command: ""}
	if err := runPluginWithEnv(entry, []string{"vm-a"}, "vc-primary", runner); err == nil {
		t.Fatalf("expected missing plugin command to fail")
	}
}
