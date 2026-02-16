// Path: internal/app/app_test.go
// Description: Validate top-level workflow orchestration and output production.
package app

import (
	"bytes"
	"testing"

	"github.com/takelley1/hypersphere/internal/config"
	"github.com/takelley1/hypersphere/internal/deletion"
	"github.com/takelley1/hypersphere/internal/migration"
)

type fakePlanner struct {
	plan []migration.PlanStep
	sum  migration.ExecutionSummary
}

func (f fakePlanner) BuildPlan(_ []migration.VM, _ []migration.Datastore) []migration.PlanStep {
	return f.plan
}

func (f fakePlanner) ExecutePlan(_ []migration.PlanStep, _ bool, _ int, _ migration.Mover) migration.ExecutionSummary {
	return f.sum
}

type fakeDeletionEngine struct {
	actions []deletion.Action
}

func (f fakeDeletionEngine) Plan(_ []deletion.VM, _ deletion.Mode, _ TimeValue) []deletion.Action {
	return f.actions
}

func TestRunMigrationWorkflow(t *testing.T) {
	buf := &bytes.Buffer{}
	application := New(buf)
	cfg := config.Config{Execute: false}
	planner := fakePlanner{plan: []migration.PlanStep{{Order: 1, VMName: "vm-a", TargetDatastore: "ds-1"}}, sum: migration.ExecutionSummary{DryRunCount: 1}}
	summary := application.RunMigration(cfg, nil, nil, planner)
	if summary.DryRunCount != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected workflow output")
	}
}

func TestRunDeletionWorkflow(t *testing.T) {
	buf := &bytes.Buffer{}
	application := New(buf)
	engine := fakeDeletionEngine{actions: []deletion.Action{{Type: deletion.ActionMark, VMName: "vm-a"}}}
	application.RunDeletion([]deletion.VM{}, deletion.ModeAll, TimeValue{}, engine)
	if buf.Len() == 0 {
		t.Fatalf("expected workflow output")
	}
}
