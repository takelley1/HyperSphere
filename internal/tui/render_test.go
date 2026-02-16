// Path: internal/tui/render_test.go
// Description: Validate plain terminal rendering for plan tables and summaries.
package tui

import (
	"strings"
	"testing"

	"github.com/takelley1/hypersphere/internal/deletion"
	"github.com/takelley1/hypersphere/internal/migration"
)

func TestRenderMigrationPlan(t *testing.T) {
	plan := []migration.PlanStep{{Order: 1, VMName: "vm-a", SourceDatastore: "src", TargetDatastore: "ds-1", SkipReason: "", Tier: migration.TierPrimary.String()}, {Order: 2, VMName: "vm-b", SourceDatastore: "src", SkipReason: migration.SkipOverThreshold, Tier: "-"}}
	out := RenderMigrationPlan(plan)
	if !strings.Contains(out, "Migration Plan") || !strings.Contains(out, "OVER_85") {
		t.Fatalf("unexpected render output: %s", out)
	}
}

func TestRenderDeletionPlan(t *testing.T) {
	actions := []deletion.Action{{Type: deletion.ActionMark, VMName: "vm-a", Notes: "delete_on=2026-03-01"}, {Type: deletion.ActionPurge, VMName: "vm-b", Notes: "expired"}}
	out := RenderDeletionPlan(actions)
	if !strings.Contains(out, "Pending Deletion Plan") || !strings.Contains(out, "purge") {
		t.Fatalf("unexpected render output: %s", out)
	}
}
