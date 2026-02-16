// Path: cmd/hypersphere/main_test.go
// Description: Validate command-line startup behavior for top-level HyperSphere subcommands.
package main

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRunVersionCommandPrintsBuildFields(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run([]string{"version"}, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	output := stdout.String()
	if !regexp.MustCompile(`version=[0-9]+\.[0-9]+\.[0-9]+`).MatchString(output) {
		t.Fatalf("expected semantic version field in output, got %q", output)
	}
	if !regexp.MustCompile(`commit=[0-9a-f]{7,40}|commit=unknown`).MatchString(output) {
		t.Fatalf("expected commit field in output, got %q", output)
	}
	if !strings.Contains(output, "buildDate=") {
		t.Fatalf("expected buildDate field in output, got %q", output)
	}
}

func TestRunInfoCommandPrintsAbsolutePaths(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run([]string{"info"}, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", exitCode, stderr.String())
	}
	output := strings.TrimSpace(stdout.String())
	lines := strings.Split(output, "\n")
	expectedKeys := []string{"config", "logs", "dumps", "skins", "plugins", "hotkeys"}
	if len(lines) != len(expectedKeys) {
		t.Fatalf("expected %d info lines, got %d (%q)", len(expectedKeys), len(lines), output)
	}
	for _, key := range expectedKeys {
		match := ""
		for _, line := range lines {
			if strings.HasPrefix(line, key+"=") {
				match = strings.TrimSpace(strings.TrimPrefix(line, key+"="))
				break
			}
		}
		if match == "" {
			t.Fatalf("missing key %q in output %q", key, output)
		}
		if !filepath.IsAbs(match) {
			t.Fatalf("expected absolute path for %q, got %q", key, match)
		}
	}
}
