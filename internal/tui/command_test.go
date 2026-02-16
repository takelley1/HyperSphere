// Path: internal/tui/command_test.go
// Description: Validate explorer command parsing and alias resolution semantics.
package tui

import "testing"

func TestParseExplorerInputKinds(t *testing.T) {
	cases := []struct {
		line     string
		kind     CommandKind
		value    string
		hasError bool
	}{
		{line: "", kind: CommandNoop, value: ""},
		{line: ":q", kind: CommandQuit, value: ""},
		{line: ":help", kind: CommandHelp, value: ""},
		{line: ":ro", kind: CommandReadOnly, value: "toggle"},
		{line: ":readonly on", kind: CommandReadOnly, value: "on"},
		{line: ":readonly off", kind: CommandReadOnly, value: "off"},
		{line: ":vms", kind: CommandView, value: "vm"},
		{line: ":ds", kind: CommandView, value: "datastore"},
		{line: "!power-off", kind: CommandAction, value: "power-off"},
		{line: "shift+o", kind: CommandHotKey, value: "SHIFT+O"},
		{line: "!", kind: "", hasError: true},
		{line: ":readonly maybe", kind: "", hasError: true},
		{line: ":rock", kind: "", hasError: true},
		{line: ":unknown", kind: "", hasError: true},
	}
	for _, tc := range cases {
		parsed, err := ParseExplorerInput(tc.line)
		if tc.hasError {
			if err == nil {
				t.Fatalf("expected error for line %q", tc.line)
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error for line %q: %v", tc.line, err)
		}
		if parsed.Kind != tc.kind || parsed.Value != tc.value {
			t.Fatalf("unexpected parse for %q: %+v", tc.line, parsed)
		}
	}
}

func TestParseExplorerInputHandlesSpacebar(t *testing.T) {
	parsed, err := ParseExplorerInput(" ")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if parsed.Kind != CommandHotKey || parsed.Value != "SPACE" {
		t.Fatalf("unexpected space parse: %+v", parsed)
	}
}

func TestParseReadOnlyCommandEdgeBranches(t *testing.T) {
	if _, err := parseReadOnlyCommand(":"); err == nil {
		t.Fatalf("expected empty readonly command error")
	}
	if _, err := parseReadOnlyCommand(":notreadonly"); err == nil {
		t.Fatalf("expected invalid readonly command error")
	}
}
