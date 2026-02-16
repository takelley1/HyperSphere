// Path: internal/tui/prompt_test.go
// Description: Validate prompt history and suggestion behavior for command-mode UX parity.
package tui

import "testing"

func TestPromptStateHistoryNavigation(t *testing.T) {
	state := NewPromptState(3)
	state.Record(":vm")
	state.Record(":cluster")
	state.Record("!power-off")

	entry, ok := state.Previous()
	if !ok || entry != "!power-off" {
		t.Fatalf("unexpected history previous #1: %q", entry)
	}
	entry, ok = state.Previous()
	if !ok || entry != ":cluster" {
		t.Fatalf("unexpected history previous #2: %q", entry)
	}
	entry, ok = state.Next()
	if !ok || entry != "!power-off" {
		t.Fatalf("unexpected history next #1: %q", entry)
	}
}

func TestPromptStateHistoryLimitAndIgnoreBlank(t *testing.T) {
	state := NewPromptState(2)
	state.Record(" ")
	state.Record(":vm")
	state.Record(":lun")
	state.Record(":cluster")
	first, _ := state.Previous()
	second, _ := state.Previous()
	if first != ":cluster" || second != ":lun" {
		t.Fatalf("unexpected bounded history entries: %q %q", first, second)
	}
}

func TestPromptStateSuggestions(t *testing.T) {
	state := NewPromptState(10)
	view := ResourceView{Resource: ResourceVM, Actions: []string{"power-on", "migrate"}}
	suggestions := state.Suggest(":v", view)
	if len(suggestions) == 0 {
		t.Fatalf("expected resource suggestions")
	}
	actionSuggestions := state.Suggest("!m", view)
	if len(actionSuggestions) != 1 || actionSuggestions[0] != "!migrate" {
		t.Fatalf("unexpected action suggestions: %v", actionSuggestions)
	}
}

func TestPromptStateEdgeBranches(t *testing.T) {
	state := NewPromptState(0)
	if state.maxSize <= 0 {
		t.Fatalf("expected default max size")
	}
	if _, ok := state.Previous(); ok {
		t.Fatalf("expected no previous history for empty state")
	}
	if _, ok := state.Next(); ok {
		t.Fatalf("expected no next history for empty state")
	}
	state.Record(":vm")
	if _, ok := state.Next(); ok {
		t.Fatalf("expected no next history when cursor is at tail")
	}
	if suggestions := state.Suggest("   ", ResourceView{}); suggestions != nil {
		t.Fatalf("expected nil suggestions for blank prefix")
	}
}

func TestPromptSuggestionHelpers(t *testing.T) {
	view := ResourceView{
		Actions:     []string{"power-on"},
		SortHotKeys: map[string]string{"N": "NAME", "A": "AGE"},
	}
	candidates := suggestionCandidates(view)
	if len(candidates) == 0 {
		t.Fatalf("expected non-empty suggestion candidates")
	}
	keys := sortedMapKeys(map[string]string{})
	if len(keys) != 0 {
		t.Fatalf("expected empty sorted keys")
	}
	filtered := filterCandidates([]string{":vm", ":vm", ":lun"}, ":v")
	if len(filtered) != 1 || filtered[0] != ":vm" {
		t.Fatalf("expected duplicate-filtered suggestions, got %v", filtered)
	}
}
