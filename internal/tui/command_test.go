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
		{line: ":vm /prod", kind: CommandView, value: "vm"},
		{line: ":dc", kind: CommandView, value: "datacenter"},
		{line: ":datacenter", kind: CommandView, value: "datacenter"},
		{line: ":rp", kind: CommandView, value: "resourcepool"},
		{line: ":resourcepool", kind: CommandView, value: "resourcepool"},
		{line: ":nw", kind: CommandView, value: "network"},
		{line: ":network", kind: CommandView, value: "network"},
		{line: ":tp", kind: CommandView, value: "template"},
		{line: ":template", kind: CommandView, value: "template"},
		{line: ":ss", kind: CommandView, value: "snapshot"},
		{line: ":snap", kind: CommandView, value: "snapshot"},
		{line: ":snapshot", kind: CommandView, value: "snapshot"},
		{line: ":task", kind: CommandView, value: "task"},
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

func TestParseExplorerInputHistoryAndSuggest(t *testing.T) {
	up, err := ParseExplorerInput(":history up")
	if err != nil {
		t.Fatalf("unexpected history parse error: %v", err)
	}
	if up.Kind != CommandHistory || up.Value != "up" {
		t.Fatalf("unexpected history parse: %+v", up)
	}
	suggest, err := ParseExplorerInput(":suggest :v")
	if err != nil {
		t.Fatalf("unexpected suggest parse error: %v", err)
	}
	if suggest.Kind != CommandSuggest || suggest.Value != ":v" {
		t.Fatalf("unexpected suggest parse: %+v", suggest)
	}
	if _, err := ParseExplorerInput(":history left"); err == nil {
		t.Fatalf("expected invalid history direction error")
	}
	if _, err := parseHistoryCommand(":history"); err == nil {
		t.Fatalf("expected invalid short history command")
	}
	if _, err := parseHistoryCommand(":hist up"); err == nil {
		t.Fatalf("expected invalid history keyword")
	}
	if _, err := parseSuggestCommand(":suggest"); err == nil {
		t.Fatalf("expected empty suggest prefix error")
	}
}

func TestParseExplorerInputContextCommands(t *testing.T) {
	list, err := ParseExplorerInput(":ctx")
	if err != nil {
		t.Fatalf("unexpected ctx list parse error: %v", err)
	}
	if list.Kind != CommandContext || list.Value != "" {
		t.Fatalf("unexpected ctx list parse: %+v", list)
	}

	selectCmd, err := ParseExplorerInput(":ctx vc-west")
	if err != nil {
		t.Fatalf("unexpected ctx select parse error: %v", err)
	}
	if selectCmd.Kind != CommandContext || selectCmd.Value != "vc-west" {
		t.Fatalf("unexpected ctx select parse: %+v", selectCmd)
	}

	if _, err := ParseExplorerInput(":ctx vc-west extra"); err == nil {
		t.Fatalf("expected invalid ctx command with extra args")
	}
	if _, err := parseContextCommand(":context vc-west"); err == nil {
		t.Fatalf("expected invalid ctx keyword error")
	}
}
