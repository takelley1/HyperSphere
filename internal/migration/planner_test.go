// Path: internal/migration/planner_test.go
// Description: Validate datastore migration planning and execution safeguards.
package migration

import (
	"errors"
	"testing"
)

type fakeMover struct {
	calls []string
	errAt map[string]error
}

func (f *fakeMover) Move(vmName string, target string) error {
	f.calls = append(f.calls, vmName+"->"+target)
	if err := f.errAt[vmName]; err != nil {
		delete(f.errAt, vmName)
		return err
	}
	return nil
}

func TestBuildPlanSelectsBestCandidateAndReranks(t *testing.T) {
	planner := NewPlanner(85)
	vms := []VM{{Name: "vm-a", SizeGB: 20, SourceDatastore: "src"}, {Name: "vm-b", SizeGB: 20, SourceDatastore: "src"}}
	candidates := []Datastore{
		{Name: "src", CapacityGB: 100, UsedGB: 20, Tier: TierPrimary},
		{Name: "ds-1", CapacityGB: 100, UsedGB: 40, Tier: TierPrimary},
		{Name: "ds-2", CapacityGB: 100, UsedGB: 30, Tier: TierSecondary},
	}
	plan := planner.BuildPlan(vms, candidates)
	if len(plan) != 2 {
		t.Fatalf("expected 2 plan steps, got %d", len(plan))
	}
	if plan[0].TargetDatastore != "ds-2" {
		t.Fatalf("expected first VM to target ds-2, got %s", plan[0].TargetDatastore)
	}
	if plan[1].TargetDatastore != "ds-1" {
		t.Fatalf("expected second VM to re-rank to ds-1, got %s", plan[1].TargetDatastore)
	}
}

func TestBuildPlanSkipsWhenNoEligibleTarget(t *testing.T) {
	planner := NewPlanner(85)
	vms := []VM{{Name: "vm-a", SizeGB: 50, SourceDatastore: "src"}}
	candidates := []Datastore{{Name: "src", CapacityGB: 100, UsedGB: 20, Tier: TierPrimary}}
	plan := planner.BuildPlan(vms, candidates)
	if plan[0].SkipReason != SkipNoEligibleTarget {
		t.Fatalf("expected %s skip, got %s", SkipNoEligibleTarget, plan[0].SkipReason)
	}
}

func TestBuildPlanSkipsWhenAllTargetsOverThreshold(t *testing.T) {
	planner := NewPlanner(85)
	vms := []VM{{Name: "vm-a", SizeGB: 20, SourceDatastore: "src"}}
	candidates := []Datastore{{Name: "ds-1", CapacityGB: 100, UsedGB: 80, Tier: TierPrimary}}
	plan := planner.BuildPlan(vms, candidates)
	if plan[0].SkipReason != SkipOverThreshold {
		t.Fatalf("expected %s skip, got %s", SkipOverThreshold, plan[0].SkipReason)
	}
}

func TestExecutePlanHonorsDryRun(t *testing.T) {
	planner := NewPlanner(85)
	plan := []PlanStep{{VMName: "vm-a", TargetDatastore: "ds-1"}}
	mover := &fakeMover{errAt: map[string]error{}}
	summary := planner.ExecutePlan(plan, false, 2, mover)
	if len(mover.calls) != 0 {
		t.Fatalf("expected no calls in dry-run")
	}
	if summary.DryRunCount != 1 {
		t.Fatalf("expected one dry-run step, got %d", summary.DryRunCount)
	}
}

func TestExecutePlanRetriesMove(t *testing.T) {
	planner := NewPlanner(85)
	plan := []PlanStep{{VMName: "vm-a", TargetDatastore: "ds-1"}}
	mover := &fakeMover{errAt: map[string]error{"vm-a": errors.New("temporary")}}
	summary := planner.ExecutePlan(plan, true, 2, mover)
	if len(mover.calls) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(mover.calls))
	}
	if summary.MigratedCount != 1 || summary.FailedCount != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
