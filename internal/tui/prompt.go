// Path: internal/tui/prompt.go
// Description: Manage prompt command history and suggestions for command-mode UX.
package tui

import (
	"sort"
	"strings"
)

const defaultPromptHistoryMax = 200

// PromptState tracks command history and suggestions.
type PromptState struct {
	history []string
	cursor  int
	maxSize int
}

// NewPromptState creates prompt state with a bounded history size.
func NewPromptState(maxSize int) PromptState {
	if maxSize < 1 {
		maxSize = defaultPromptHistoryMax
	}
	return PromptState{history: []string{}, cursor: 0, maxSize: maxSize}
}

// Record stores one command line in history.
func (p *PromptState) Record(line string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}
	p.history = append(p.history, trimmed)
	if len(p.history) > p.maxSize {
		p.history = p.history[len(p.history)-p.maxSize:]
	}
	p.cursor = len(p.history)
}

// Previous returns the previous history entry.
func (p *PromptState) Previous() (string, bool) {
	if len(p.history) == 0 {
		return "", false
	}
	if p.cursor > 0 {
		p.cursor--
	}
	return p.history[p.cursor], true
}

// Next returns the next history entry.
func (p *PromptState) Next() (string, bool) {
	if len(p.history) == 0 {
		return "", false
	}
	if p.cursor >= len(p.history) {
		p.cursor = len(p.history) - 1
		return p.history[p.cursor], true
	}
	if p.cursor >= len(p.history)-1 {
		return p.history[p.cursor], true
	}
	p.cursor++
	return p.history[p.cursor], true
}

// Suggest returns sorted command suggestions for the given prefix.
func (p *PromptState) Suggest(prefix string, view ResourceView) []string {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		return nil
	}
	candidates := suggestionCandidates(view)
	return filterCandidates(candidates, trimmed)
}

func suggestionCandidates(view ResourceView) []string {
	candidates := resourceCommandAliases()
	for _, action := range view.Actions {
		candidates = append(candidates, "!"+action)
	}
	for _, key := range sortedMapKeys(view.SortHotKeys) {
		candidates = append(candidates, key)
	}
	return candidates
}

func resourceCommandAliases() []string {
	candidates := append([]string{}, ResourceCommandAliases()...)
	return append(candidates, ":help", ":q", ":readonly", ":ro", ":history up", ":history down", ":ctx")
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func filterCandidates(candidates []string, prefix string) []string {
	prefixLower := strings.ToLower(prefix)
	seen := map[string]struct{}{}
	filtered := []string{}
	for _, candidate := range candidates {
		if !strings.HasPrefix(strings.ToLower(candidate), prefixLower) {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		filtered = append(filtered, candidate)
	}
	sort.Strings(filtered)
	return filtered
}
