// Path: internal/app/app_extra_test.go
// Description: Cover app noop mover by running with the concrete planner.
package app

import (
	"bytes"
	"testing"

	"github.com/takelley1/hypersphere/internal/config"
	"github.com/takelley1/hypersphere/internal/migration"
)

func TestRunMigrationWithConcretePlannerCoversNoopMover(t *testing.T) {
	buf := &bytes.Buffer{}
	application := New(buf)
	planner := migration.NewPlanner(90)
	cfg := config.Config{Execute: true}
	vms := []migration.VM{{Name: "vm", SizeGB: 1, SourceDatastore: "src"}}
	stores := []migration.Datastore{{Name: "src", CapacityGB: 100, UsedGB: 10, Tier: migration.TierPrimary}, {Name: "dst", CapacityGB: 100, UsedGB: 20, Tier: migration.TierPrimary}}
	summary := application.RunMigration(cfg, vms, stores, planner)
	if summary.MigratedCount != 1 {
		t.Fatalf("expected one migrated VM, got %+v", summary)
	}
}
