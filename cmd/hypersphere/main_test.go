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

func TestParseFlagsRefreshClampsToMinimum(t *testing.T) {
	flags, err := parseFlags([]string{"--refresh", "0.25"})
	if err != nil {
		t.Fatalf("expected --refresh to parse, got error: %v", err)
	}
	if flags.refreshSeconds != minimumRefreshSeconds {
		t.Fatalf(
			"expected refresh seconds clamped to %.2f, got %.2f",
			minimumRefreshSeconds,
			flags.refreshSeconds,
		)
	}
}

func TestParseFlagsRefreshKeepsConfiguredValueAboveMinimum(t *testing.T) {
	flags, err := parseFlags([]string{"--refresh", "2.75"})
	if err != nil {
		t.Fatalf("expected --refresh to parse, got error: %v", err)
	}
	if flags.refreshSeconds != 2.75 {
		t.Fatalf("expected refresh seconds to stay 2.75, got %.2f", flags.refreshSeconds)
	}
}

func TestParseFlagsLogLevelMapsValidValues(t *testing.T) {
	testCases := map[string]logLevel{
		"debug": logLevelDebug,
		"info":  logLevelInfo,
		"warn":  logLevelWarn,
		"error": logLevelError,
	}
	for input, expected := range testCases {
		flags, err := parseFlags([]string{"--log-level", input})
		if err != nil {
			t.Fatalf("expected --log-level %q to parse, got error: %v", input, err)
		}
		if flags.logLevel != expected {
			t.Fatalf("expected log level %q to map to %q, got %q", input, expected, flags.logLevel)
		}
	}
}

func TestParseFlagsLogLevelRejectsInvalidValue(t *testing.T) {
	_, err := parseFlags([]string{"--log-level", "verbose"})
	if err == nil {
		t.Fatalf("expected invalid --log-level value to fail")
	}
	if !strings.Contains(err.Error(), "invalid log level") {
		t.Fatalf("expected invalid log level error, got %v", err)
	}
}
