// Path: internal/tui/command.go
// Description: Parse explorer command-line inputs into typed actions for the TUI session.
package tui

import (
	"fmt"
	"strings"
)

// CommandKind identifies which explorer behavior an input line should trigger.
type CommandKind string

const (
	CommandNoop     CommandKind = "noop"
	CommandQuit     CommandKind = "quit"
	CommandHelp     CommandKind = "help"
	CommandLastView CommandKind = "last_view"
	CommandFilter   CommandKind = "filter"
	CommandView     CommandKind = "view"
	CommandAction   CommandKind = "action"
	CommandHotKey   CommandKind = "hotkey"
)

// ExplorerCommand is a parsed line from command mode.
type ExplorerCommand struct {
	Kind  CommandKind
	Value string
}

// ParseExplorerInput parses one user line into a command-mode instruction.
func ParseExplorerInput(line string) (ExplorerCommand, error) {
	if line == " " {
		return ExplorerCommand{Kind: CommandHotKey, Value: "SPACE"}, nil
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ExplorerCommand{Kind: CommandNoop}, nil
	}
	if trimmed == ":q" || trimmed == ":quit" {
		return ExplorerCommand{Kind: CommandQuit}, nil
	}
	if trimmed == ":help" || trimmed == ":h" || trimmed == "?" {
		return ExplorerCommand{Kind: CommandHelp}, nil
	}
	if trimmed == ":-" {
		return ExplorerCommand{Kind: CommandLastView}, nil
	}
	if strings.HasPrefix(trimmed, "/") {
		return ExplorerCommand{
			Kind:  CommandFilter,
			Value: strings.TrimSpace(strings.TrimPrefix(trimmed, "/")),
		}, nil
	}
	if strings.HasPrefix(trimmed, ":") {
		resource, err := parseCommand(trimmed)
		if err != nil {
			return ExplorerCommand{}, err
		}
		return ExplorerCommand{Kind: CommandView, Value: string(resource)}, nil
	}
	if strings.HasPrefix(trimmed, "!") {
		action := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "!")))
		if action == "" {
			return ExplorerCommand{}, fmt.Errorf("%w: empty action", ErrInvalidAction)
		}
		return ExplorerCommand{Kind: CommandAction, Value: action}, nil
	}
	return ExplorerCommand{Kind: CommandHotKey, Value: normalizeKey(trimmed)}, nil
}
