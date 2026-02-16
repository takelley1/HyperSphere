// Path: internal/migration/planner_extra_test.go
// Description: Cover migration planner edge branches and failure handling.
package migration

import (
	"errors"
	"testing"
)

func TestExecutePlanCountsFailuresAndSkips(t *testing.T) {
	planner := NewPlanner(85)
	plan := []PlanStep{{VMName: "vm-a", TargetDatastore: "ds-1"}, {VMName: "vm-b", SkipReason: SkipOverThreshold}}
	mover := &fakeMover{errAt: map[string]error{"vm-a": errors.New("boom")}}
	summary := planner.ExecutePlan(plan, true, 0, mover)
	if summary.FailedCount != 1 || summary.MigratedCount != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(mover.calls) != 1 {
		t.Fatalf("expected one attempt with retries=0 normalized to 1")
	}
}

func TestSortTargetsUsesTierAndNameTiebreakers(t *testing.T) {
	planner := NewPlanner(90)
	vms := []VM{{Name: "vm", SizeGB: 1, SourceDatastore: "src"}}
	stores := []Datastore{{Name: "b", CapacityGB: 100, UsedGB: 10, Tier: TierSecondary}, {Name: "a", CapacityGB: 100, UsedGB: 10, Tier: TierPrimary}}
	plan := planner.BuildPlan(vms, stores)
	if plan[0].TargetDatastore != "a" {
		t.Fatalf("expected primary tier target first, got %s", plan[0].TargetDatastore)
	}
}

func TestProjectedUtilHandlesZeroCapacity(t *testing.T) {
	planner := NewPlanner(85)
	plan := planner.BuildPlan([]VM{{Name: "vm", SizeGB: 1, SourceDatastore: "src"}}, []Datastore{{Name: "ds", CapacityGB: 0, UsedGB: 0, Tier: TierPrimary}})
	if plan[0].SkipReason != SkipOverThreshold {
		t.Fatalf("expected over-threshold when capacity is zero, got %+v", plan[0])
	}
}
