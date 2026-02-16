// Path: internal/app/explorer.go
// Description: Run command-mode resource navigation and render selected resource views.
package app

import (
	"fmt"

	"github.com/takelley1/hypersphere/internal/tui"
)

// ResourceNavigator switches resource views from colon commands.
type ResourceNavigator interface {
	Execute(command string) (tui.ResourceView, error)
}

// RunExplorerCommand executes a command-mode view switch and renders the selected table.
func (a App) RunExplorerCommand(command string, navigator ResourceNavigator) error {
	view, err := navigator.Execute(command)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprint(a.out, tui.RenderResourceView(view))
	return nil
}
