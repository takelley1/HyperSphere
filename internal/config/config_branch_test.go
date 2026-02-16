// Path: internal/config/config_branch_test.go
// Description: Cover prompt error branches in configuration resolution.
package config

import (
	"errors"
	"testing"
)

type selectivePrompter struct {
	values map[string]string
	errKey string
}

func (s selectivePrompter) Ask(key string) (string, error) {
	if key == s.errKey {
		return "", errors.New("forced")
	}
	return s.values[key], nil
}

func TestResolveReturnsExecutePromptError(t *testing.T) {
	prompt := selectivePrompter{values: map[string]string{"threshold": "80"}, errKey: "execute"}
	_, err := Resolve(CLIInput{Mode: "mark"}, map[string]string{}, prompt)
	if err == nil {
		t.Fatalf("expected execute prompt error")
	}
}

func TestResolveReturnsThresholdPromptError(t *testing.T) {
	prompt := selectivePrompter{values: map[string]string{"execute": "true"}, errKey: "threshold"}
	_, err := Resolve(CLIInput{Mode: "mark"}, map[string]string{}, prompt)
	if err == nil {
		t.Fatalf("expected threshold prompt error")
	}
}
