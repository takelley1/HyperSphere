// Path: internal/app/explorer_test.go
// Description: Validate app integration for command-mode resource view rendering.
package app

import (
	"bytes"
	"testing"

	"github.com/takelley1/hypersphere/internal/tui"
)

type fakeNavigator struct {
	view tui.ResourceView
	err  error
}

func (f fakeNavigator) Execute(_ string) (tui.ResourceView, error) {
	return f.view, f.err
}

func TestRunExplorerCommandRendersView(t *testing.T) {
	buf := &bytes.Buffer{}
	application := New(buf)
	view := tui.ResourceView{Resource: tui.ResourceVM, Columns: []string{"NAME"}, Rows: [][]string{{"vm-a"}}}
	err := application.RunExplorerCommand(":vm", fakeNavigator{view: view})
	if err != nil {
		t.Fatalf("RunExplorerCommand returned error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected rendered output")
	}
}

func TestRunExplorerCommandReturnsNavigatorError(t *testing.T) {
	buf := &bytes.Buffer{}
	application := New(buf)
	err := application.RunExplorerCommand(":vm", fakeNavigator{err: errBoom{}})
	if err == nil {
		t.Fatalf("expected navigator error")
	}
}

type errBoom struct{}

func (errBoom) Error() string {
	return "boom"
}
