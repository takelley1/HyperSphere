// Path: cmd/hypersphere/explorer_tui_test.go
// Description: Validate table-model helpers and prompt/runtime branch behavior for the full-screen explorer.
package main

import (
	"os"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/takelley1/hypersphere/internal/tui"
)

func TestTableRowsIncludesMarkerColumnAndMarks(t *testing.T) {
	view := tui.ResourceView{
		Columns: []string{"NAME", "CLUSTER"},
		Rows:    [][]string{{"vm-a", "cluster-a"}, {"vm-b", "cluster-b"}},
		IDs:     []string{"vm-a", "vm-b"},
	}
	rows := tableRows(view, func(id string) bool {
		return id == "vm-b"
	})
	if len(rows) != 3 {
		t.Fatalf("expected header + 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "SEL" || rows[0][1] != "NAME" {
		t.Fatalf("unexpected header row: %#v", rows[0])
	}
	if rows[1][0] != " " || rows[2][0] != "*" {
		t.Fatalf("unexpected marker values: %#v %#v", rows[1], rows[2])
	}
}

func TestSelectionForTableOffsetsHeaderAndMarkerColumns(t *testing.T) {
	session := tui.NewSession(tui.Catalog{VMs: []tui.VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.HandleKey("DOWN"); err != nil {
		t.Fatalf("HandleKey error: %v", err)
	}
	if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("HandleKey error: %v", err)
	}
	row, column := selectionForTable(session)
	if row != 2 || column != 2 {
		t.Fatalf("expected selected table position 2,2 got %d,%d", row, column)
	}
}

func TestEmitStatusWritesErrorsOnly(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.emitStatus(nil)
	if strings.Contains(runtime.status.GetText(true), "command error") {
		t.Fatalf("did not expect error status for nil")
	}
	runtime.emitStatus(tui.ErrUnsupportedHotKey)
	if !strings.Contains(runtime.status.GetText(true), "unsupported hotkey") {
		t.Fatalf("expected unsupported hotkey in status")
	}
}

func TestRenderFooterIncludesPromptMode(t *testing.T) {
	if !strings.Contains(renderFooter(true), "Prompt: ON") {
		t.Fatalf("expected prompt-on indicator in footer")
	}
	if !strings.Contains(renderFooter(false), "Prompt: OFF") {
		t.Fatalf("expected prompt-off indicator in footer")
	}
}

func TestEventToHotKeyVimColumnMovement(t *testing.T) {
	left, ok := eventToHotKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	if !ok || left != "LEFT" {
		t.Fatalf("expected h to map to LEFT, got %q ok=%v", left, ok)
	}
	right, ok := eventToHotKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	if !ok || right != "RIGHT" {
		t.Fatalf("expected l to map to RIGHT, got %q ok=%v", right, ok)
	}
	sortHotKey, ok := eventToHotKey(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModShift))
	if !ok || sortHotKey != "H" {
		t.Fatalf("expected shifted H to remain sort hotkey, got %q ok=%v", sortHotKey, ok)
	}
}

func TestReadThemeRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	theme := readTheme()
	if theme.UseColor {
		t.Fatalf("expected NO_COLOR to disable color")
	}
	os.Unsetenv("NO_COLOR")
	theme = readTheme()
	if !theme.UseColor {
		t.Fatalf("expected color mode enabled when NO_COLOR is unset")
	}
}

func TestApplyPromptCompletionUsesFirstSuggestion(t *testing.T) {
	promptState := tui.NewPromptState(20)
	view := tui.ResourceView{Actions: []string{"power-on", "power-off"}}
	value, status, changed := applyPromptCompletion(&promptState, view, "!pow")
	if !changed {
		t.Fatalf("expected completion to be applied")
	}
	if value != "!power-off" {
		t.Fatalf("expected !power-off completion, got %q", value)
	}
	if !strings.Contains(status, "!power-off") {
		t.Fatalf("expected completion status to mention !power-off, got %q", status)
	}
}

func TestApplyPromptCompletionNoMatch(t *testing.T) {
	promptState := tui.NewPromptState(20)
	view := tui.ResourceView{}
	value, status, changed := applyPromptCompletion(&promptState, view, "!does-not-exist")
	if changed {
		t.Fatalf("did not expect completion change")
	}
	if value != "!does-not-exist" {
		t.Fatalf("expected input unchanged, got %q", value)
	}
	if status != "" {
		t.Fatalf("expected empty status for no completion, got %q", status)
	}
}

func TestHandlePromptHistoryTabCompletesPrompt(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.prompt.SetText(":v")
	evt := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	consumed := runtime.handlePromptHistory(evt)
	if consumed != nil {
		t.Fatalf("expected tab event to be consumed")
	}
	if runtime.prompt.GetText() != ":vm" {
		t.Fatalf("expected prompt completion :vm, got %q", runtime.prompt.GetText())
	}
	if !strings.Contains(runtime.status.GetText(true), "completion: :vm") {
		t.Fatalf("expected status to include completion message")
	}
}
