// Path: internal/config/config_test.go
// Description: Validate configuration precedence, prompting, and validation rules.
package config

import (
	"errors"
	"testing"
)

type fakePrompter struct {
	responses map[string]string
	err       error
}

func (f fakePrompter) Ask(key string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.responses[key], nil
}

func TestResolvePrefersCLIThenEnvThenPrompt(t *testing.T) {
	cli := CLIInput{Mode: "all", Execute: true, ThresholdPercent: 70}
	env := map[string]string{
		"HYPERSPHERE_MODE":      "mark",
		"HYPERSPHERE_EXECUTE":   "false",
		"HYPERSPHERE_THRESHOLD": "80",
	}
	cfg, err := Resolve(cli, env, fakePrompter{responses: map[string]string{}})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if cfg.Mode != "all" {
		t.Fatalf("expected CLI mode, got %s", cfg.Mode)
	}
	if !cfg.Execute {
		t.Fatalf("expected CLI execute true")
	}
	if cfg.ThresholdPercent != 70 {
		t.Fatalf("expected CLI threshold 70, got %d", cfg.ThresholdPercent)
	}
}

func TestResolveReadsEnvWhenCLIMissing(t *testing.T) {
	cfg, err := Resolve(CLIInput{}, map[string]string{
		"HYPERSPHERE_MODE":      "purge",
		"HYPERSPHERE_EXECUTE":   "true",
		"HYPERSPHERE_THRESHOLD": "85",
		"HOME":                  "/home/tester",
	}, fakePrompter{responses: map[string]string{}})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if cfg.Mode != "purge" || !cfg.Execute || cfg.ThresholdPercent != 85 {
		t.Fatalf("unexpected resolved config: %+v", cfg)
	}
	if cfg.ConfigDir != "/home/tester/.config/hypersphere" {
		t.Fatalf("expected default config dir from HOME, got %q", cfg.ConfigDir)
	}
}

func TestResolveUsesConfigDirEnvOverride(t *testing.T) {
	cfg, err := Resolve(
		CLIInput{Mode: "mark", Execute: true, ThresholdPercent: 80},
		map[string]string{
			"HOME":                   "/home/tester",
			"HYPERSPHERE_CONFIG_DIR": "/tmp/hs-config",
		},
		fakePrompter{},
	)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if cfg.ConfigDir != "/tmp/hs-config" {
		t.Fatalf("expected config dir override from env, got %q", cfg.ConfigDir)
	}
}

func TestResolvePromptsWhenInteractiveAndMissing(t *testing.T) {
	prompt := fakePrompter{responses: map[string]string{
		"mode":      "mark",
		"execute":   "false",
		"threshold": "85",
	}}
	cfg, err := Resolve(CLIInput{}, map[string]string{}, prompt)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if cfg.Mode != "mark" || cfg.Execute || cfg.ThresholdPercent != 85 {
		t.Fatalf("unexpected prompted config: %+v", cfg)
	}
}

func TestResolveFailsWhenNonInteractiveAndMissing(t *testing.T) {
	_, err := Resolve(CLIInput{NonInteractive: true}, map[string]string{}, fakePrompter{})
	if err == nil {
		t.Fatalf("expected error for missing config in non-interactive mode")
	}
}

func TestResolveReturnsPromptError(t *testing.T) {
	want := errors.New("prompt failed")
	_, err := Resolve(CLIInput{}, map[string]string{}, fakePrompter{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("expected prompt error %v, got %v", want, err)
	}
}
