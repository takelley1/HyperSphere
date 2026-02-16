// Path: internal/migration/planner_branch_test.go
// Description: Cover datastore sorting branch for equal utilization and tier.
package migration

import "testing"

func TestSortTargetsUsesNameWhenUtilAndTierMatch(t *testing.T) {
	planner := NewPlanner(90)
	vms := []VM{{Name: "vm", SizeGB: 1, SourceDatastore: "src"}}
	stores := []Datastore{{Name: "src", CapacityGB: 100, UsedGB: 5, Tier: TierPrimary}, {Name: "z-ds", CapacityGB: 100, UsedGB: 10, Tier: TierPrimary}, {Name: "a-ds", CapacityGB: 100, UsedGB: 10, Tier: TierPrimary}}
	plan := planner.BuildPlan(vms, stores)
	if plan[0].TargetDatastore != "a-ds" {
		t.Fatalf("expected lexical target ordering, got %s", plan[0].TargetDatastore)
	}
}
