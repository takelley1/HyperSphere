// Path: internal/app/app.go
// Description: Coordinate migration and deletion workflows and render outputs.
package app

import (
	"fmt"
	"io"
	"time"

	"github.com/takelley1/hypersphere/internal/config"
	"github.com/takelley1/hypersphere/internal/deletion"
	"github.com/takelley1/hypersphere/internal/migration"
	"github.com/takelley1/hypersphere/internal/tui"
)

// TimeValue wraps optional execution time values.
type TimeValue struct {
	Value time.Time
}

// MigrationPlanner defines migration workflow behavior.
type MigrationPlanner interface {
	BuildPlan(vms []migration.VM, stores []migration.Datastore) []migration.PlanStep
	ExecutePlan(plan []migration.PlanStep, execute bool, retries int, mover migration.Mover) migration.ExecutionSummary
}

// DeletionEngine defines pending deletion workflow behavior.
type DeletionEngine interface {
	Plan(vms []deletion.VM, mode deletion.Mode, now TimeValue) []deletion.Action
}

// App prints workflow outputs.
type App struct {
	out io.Writer
}

// New construct a workflow app.
func New(out io.Writer) App {
	return App{out: out}
}

// RunMigration render and execute migration plan.
func (a App) RunMigration(cfg config.Config, vms []migration.VM, stores []migration.Datastore, planner MigrationPlanner) migration.ExecutionSummary {
	plan := planner.BuildPlan(vms, stores)
	_, _ = fmt.Fprint(a.out, tui.RenderMigrationPlan(plan))
	summary := planner.ExecutePlan(plan, cfg.Execute, 2, noopMover{})
	_, _ = fmt.Fprintf(a.out, "Summary migrated=%d dry_run=%d failed=%d\n", summary.MigratedCount, summary.DryRunCount, summary.FailedCount)
	return summary
}

// RunDeletion render pending deletion actions.
func (a App) RunDeletion(vms []deletion.VM, mode deletion.Mode, now TimeValue, engine DeletionEngine) []deletion.Action {
	actions := engine.Plan(vms, mode, now)
	_, _ = fmt.Fprint(a.out, tui.RenderDeletionPlan(actions))
	return actions
}

type noopMover struct{}

func (noopMover) Move(string, string) error {
	return nil
}
