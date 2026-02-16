// Path: internal/config/config.go
// Description: Resolve runtime configuration using CLI, environment, and prompting precedence.
package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	envMode      = "HYPERSPHERE_MODE"
	envExecute   = "HYPERSPHERE_EXECUTE"
	envThreshold = "HYPERSPHERE_THRESHOLD"
)

// Prompter asks users for values during interactive configuration.
type Prompter interface {
	Ask(key string) (string, error)
}

// CLIInput carries user-provided command values.
type CLIInput struct {
	Mode             string
	Execute          bool
	ExecuteSet       bool
	ThresholdPercent int
	NonInteractive   bool
}

// Config stores the resolved runtime settings.
type Config struct {
	Mode             string
	Execute          bool
	ThresholdPercent int
	NonInteractive   bool
}

// Resolve load configuration with CLI, then env, then prompt precedence.
func Resolve(cli CLIInput, env map[string]string, prompt Prompter) (Config, error) {
	cfg := Config{NonInteractive: cli.NonInteractive}
	mode, err := resolveMode(cli, env, prompt)
	if err != nil {
		return Config{}, err
	}
	cfg.Mode = mode
	cfg.Execute, err = resolveExecute(cli, env, prompt)
	if err != nil {
		return Config{}, err
	}
	cfg.ThresholdPercent, err = resolveThreshold(cli, env, prompt)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func resolveMode(cli CLIInput, env map[string]string, prompt Prompter) (string, error) {
	if cli.Mode != "" {
		return cli.Mode, nil
	}
	if value := strings.TrimSpace(env[envMode]); value != "" {
		return value, nil
	}
	if cli.NonInteractive {
		return "", fmt.Errorf("missing required configuration: mode")
	}
	return prompt.Ask("mode")
}

func resolveExecute(cli CLIInput, env map[string]string, prompt Prompter) (bool, error) {
	if cli.ExecuteSet || cli.Execute {
		return cli.Execute, nil
	}
	if value := strings.TrimSpace(env[envExecute]); value != "" {
		return parseBool(value)
	}
	if cli.NonInteractive {
		return false, fmt.Errorf("missing required configuration: execute")
	}
	value, err := prompt.Ask("execute")
	if err != nil {
		return false, err
	}
	return parseBool(value)
}

func resolveThreshold(cli CLIInput, env map[string]string, prompt Prompter) (int, error) {
	if cli.ThresholdPercent > 0 {
		return cli.ThresholdPercent, nil
	}
	if value := strings.TrimSpace(env[envThreshold]); value != "" {
		return parseInt(value)
	}
	if cli.NonInteractive {
		return 0, fmt.Errorf("missing required configuration: threshold")
	}
	value, err := prompt.Ask("threshold")
	if err != nil {
		return 0, err
	}
	return parseInt(value)
}

func parseBool(value string) (bool, error) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("invalid bool value %q: %w", value, err)
	}
	return parsed, nil
}

func parseInt(value string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid int value %q: %w", value, err)
	}
	return parsed, nil
}
