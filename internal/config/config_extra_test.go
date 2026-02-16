// Path: internal/config/config_extra_test.go
// Description: Cover invalid and edge configuration parsing branches.
package config

import "testing"

func TestResolveHandlesCLIExecuteFalseWhenSet(t *testing.T) {
	cfg, err := Resolve(CLIInput{Mode: "mark", Execute: false, ExecuteSet: true, ThresholdPercent: 85}, map[string]string{}, fakePrompter{})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if cfg.Execute {
		t.Fatalf("expected execute false from CLI")
	}
}

func TestResolveFailsForInvalidEnvBool(t *testing.T) {
	_, err := Resolve(CLIInput{Mode: "mark", ThresholdPercent: 85}, map[string]string{envExecute: "nope"}, fakePrompter{})
	if err == nil {
		t.Fatalf("expected bool parsing error")
	}
}

func TestResolveFailsForInvalidEnvInt(t *testing.T) {
	_, err := Resolve(CLIInput{Mode: "mark", Execute: true}, map[string]string{envThreshold: "x"}, fakePrompter{})
	if err == nil {
		t.Fatalf("expected int parsing error")
	}
}

func TestResolveFailsForNonInteractiveExecuteAndThreshold(t *testing.T) {
	_, err := Resolve(CLIInput{Mode: "mark", NonInteractive: true}, map[string]string{}, fakePrompter{})
	if err == nil {
		t.Fatalf("expected execute missing error")
	}
	_, err = Resolve(CLIInput{Mode: "mark", Execute: true, NonInteractive: true}, map[string]string{}, fakePrompter{})
	if err == nil {
		t.Fatalf("expected threshold missing error")
	}
}
