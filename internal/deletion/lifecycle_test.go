// Path: internal/deletion/lifecycle_test.go
// Description: Validate pending-deletion lifecycle planning and idempotent state changes.
package deletion

import (
	"testing"
	"time"
)

func fixedNow() time.Time {
	return time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)
}

func TestPlanAllActions(t *testing.T) {
	policy := Policy{MarkAfterDays: 30, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"}
	engine := NewEngine(policy)
	vms := []VM{
		{Name: "mark-me", Folder: "WORKLOADS", PoweredOffDays: 45, OwnerEmail: "a@example.com"},
		{Name: "purge-me", Folder: "PENDING_DELETION", PoweredOffDays: 60, Metadata: map[string]string{
			FieldPendingSince: "2026-01-01",
			FieldDeleteOn:     "2026-02-10",
		}},
		{Name: "remind-me", Folder: "PENDING_DELETION", PoweredOffDays: 60, Metadata: map[string]string{
			FieldPendingSince: "2026-02-01",
			FieldDeleteOn:     "2026-02-25",
		}},
		{Name: "reset-me", Folder: "WORKLOADS", PoweredOffDays: 60, Metadata: map[string]string{
			FieldPendingSince: "2026-01-01",
			FieldDeleteOn:     "2026-02-20",
			FieldOriginalName: "vm-original",
		}},
	}
	plan := engine.Plan(vms, ModeAll, fixedNow())
	if len(plan) != 4 {
		t.Fatalf("expected 4 actions, got %d", len(plan))
	}
	if plan[0].Type != ActionMark {
		t.Fatalf("expected mark action, got %s", plan[0].Type)
	}
	if plan[1].Type != ActionPurge {
		t.Fatalf("expected purge action, got %s", plan[1].Type)
	}
	if plan[2].Type != ActionRemind {
		t.Fatalf("expected remind action, got %s", plan[2].Type)
	}
	if plan[3].Type != ActionReset {
		t.Fatalf("expected reset action, got %s", plan[3].Type)
	}
}

func TestApplyMarkIsIdempotent(t *testing.T) {
	policy := Policy{MarkAfterDays: 30, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"}
	engine := NewEngine(policy)
	vm := VM{Name: "vm", Folder: "WORKLOADS", PoweredOffDays: 45, OwnerEmail: "a@example.com", Metadata: map[string]string{}}
	action := Action{Type: ActionMark, VMName: vm.Name}
	first := engine.Apply(vm, action, fixedNow())
	second := engine.Apply(first, action, fixedNow())
	if first.Metadata[FieldPendingSince] != second.Metadata[FieldPendingSince] {
		t.Fatalf("expected idempotent pending since")
	}
	if second.Metadata[FieldInitialNoticeSent] != "true" {
		t.Fatalf("expected initial notice sent true")
	}
}

func TestPlanRespectsMode(t *testing.T) {
	policy := Policy{MarkAfterDays: 30, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"}
	engine := NewEngine(policy)
	vms := []VM{{Name: "mark-me", Folder: "WORKLOADS", PoweredOffDays: 45}, {Name: "purge-me", Folder: "PENDING_DELETION", Metadata: map[string]string{FieldDeleteOn: "2026-01-01"}}}
	plan := engine.Plan(vms, ModeMark, fixedNow())
	if len(plan) != 1 || plan[0].Type != ActionMark {
		t.Fatalf("expected only mark action, got %+v", plan)
	}
}
