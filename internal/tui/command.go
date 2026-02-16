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
	CommandReadOnly CommandKind = "readonly"
	CommandLastView CommandKind = "last_view"
	CommandHistory  CommandKind = "history"
	CommandSuggest  CommandKind = "suggest"
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
	if strings.HasPrefix(trimmed, ":ro") || strings.HasPrefix(trimmed, ":readonly") {
		return parseReadOnlyCommand(trimmed)
	}
	if trimmed == ":-" {
		return ExplorerCommand{Kind: CommandLastView}, nil
	}
	if strings.HasPrefix(trimmed, ":history") {
		return parseHistoryCommand(trimmed)
	}
	if strings.HasPrefix(trimmed, ":suggest ") {
		return parseSuggestCommand(trimmed)
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

func parseReadOnlyCommand(line string) (ExplorerCommand, error) {
	fields := strings.Fields(strings.TrimPrefix(line, ":"))
	if len(fields) == 0 {
		return ExplorerCommand{}, fmt.Errorf("%w: empty command", ErrInvalidAction)
	}
	if fields[0] != "ro" && fields[0] != "readonly" {
		return ExplorerCommand{}, fmt.Errorf("%w: %s", ErrUnsupportedHotKey, line)
	}
	if len(fields) == 1 {
		return ExplorerCommand{Kind: CommandReadOnly, Value: "toggle"}, nil
	}
	value := strings.ToLower(strings.TrimSpace(fields[1]))
	if value != "on" && value != "off" && value != "toggle" {
		return ExplorerCommand{}, fmt.Errorf("%w: readonly %s", ErrInvalidAction, value)
	}
	return ExplorerCommand{Kind: CommandReadOnly, Value: value}, nil
}

func parseHistoryCommand(line string) (ExplorerCommand, error) {
	fields := strings.Fields(strings.TrimPrefix(line, ":"))
	if len(fields) != 2 || fields[0] != "history" {
		return ExplorerCommand{}, fmt.Errorf("%w: invalid history command", ErrInvalidAction)
	}
	value := strings.ToLower(strings.TrimSpace(fields[1]))
	if value != "up" && value != "down" {
		return ExplorerCommand{}, fmt.Errorf("%w: history %s", ErrInvalidAction, value)
	}
	return ExplorerCommand{Kind: CommandHistory, Value: value}, nil
}

func parseSuggestCommand(line string) (ExplorerCommand, error) {
	value := strings.TrimSpace(strings.TrimPrefix(line, ":suggest"))
	if value == "" {
		return ExplorerCommand{}, fmt.Errorf("%w: empty suggest prefix", ErrInvalidAction)
	}
	return ExplorerCommand{Kind: CommandSuggest, Value: value}, nil
}
