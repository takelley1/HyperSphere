// Path: internal/tui/render.go
// Description: Render migration and deletion plans as terminal-friendly text tables.
package tui

import (
	"fmt"
	"strings"

	"github.com/takelley1/hypersphere/internal/deletion"
	"github.com/takelley1/hypersphere/internal/migration"
)

// RenderMigrationPlan format migration plan rows.
func RenderMigrationPlan(plan []migration.PlanStep) string {
	builder := &strings.Builder{}
	builder.WriteString("Migration Plan\n")
	builder.WriteString("# VM SOURCE TARGET TIER STATUS\n")
	for _, step := range plan {
		status := "READY"
		if step.SkipReason != "" {
			status = step.SkipReason
		}
		line := fmt.Sprintf("%d %s %s %s %s %s\n", step.Order, step.VMName, step.SourceDatastore, step.TargetDatastore, step.Tier, status)
		builder.WriteString(line)
	}
	return builder.String()
}

// RenderDeletionPlan format lifecycle action rows.
func RenderDeletionPlan(actions []deletion.Action) string {
	builder := &strings.Builder{}
	builder.WriteString("Pending Deletion Plan\n")
	builder.WriteString("ACTION VM NOTES\n")
	for _, action := range actions {
		line := fmt.Sprintf("%s %s %s\n", action.Type, action.VMName, action.Notes)
		builder.WriteString(line)
	}
	return builder.String()
}
