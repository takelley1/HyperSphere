// Path: internal/deletion/lifecycle_extra_test.go
// Description: Cover lifecycle apply branches and invalid metadata scenarios.
package deletion

import (
	"testing"
	"time"
)

func TestApplySupportsRemindPurgeAndReset(t *testing.T) {
	engine := NewEngine(Policy{PurgeAfterDays: 10, PendingFolder: "PENDING_DELETION"})
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	vm := VM{Name: "vm", Metadata: map[string]string{FieldOriginalName: "old", FieldPendingSince: "2026-01-01", FieldDeleteOn: "2026-01-05", FieldOwnerEmail: "a@b", FieldInitialNoticeSent: "true", FieldReminderNoticeSent: "true"}}
	reminded := engine.Apply(vm, Action{Type: ActionRemind}, now)
	if reminded.Metadata[FieldReminderNoticeSent] != "true" {
		t.Fatalf("expected reminder flag to remain true")
	}
	purged := engine.Apply(reminded, Action{Type: ActionPurge}, now)
	if !purged.Deleted {
		t.Fatalf("expected deleted flag after purge")
	}
	reset := engine.Apply(purged, Action{Type: ActionReset}, now)
	if reset.Name != "old" || len(reset.Metadata) != 0 {
		t.Fatalf("expected metadata cleared and original name restored: %+v", reset)
	}
}

func TestPlanHandlesInvalidDatesAndModePurge(t *testing.T) {
	engine := NewEngine(Policy{MarkAfterDays: 10, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"})
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	vms := []VM{{Name: "invalid-purge", Folder: "PENDING_DELETION", Metadata: map[string]string{FieldDeleteOn: "invalid"}}, {Name: "invalid-remind", Folder: "PENDING_DELETION", Metadata: map[string]string{FieldPendingSince: "invalid", FieldDeleteOn: "2026-02-20"}}, {Name: "purge", Folder: "PENDING_DELETION", Metadata: map[string]string{FieldDeleteOn: "2026-01-01"}}}
	actions := engine.Plan(vms, ModePurge, now)
	if len(actions) != 1 || actions[0].Type != ActionPurge {
		t.Fatalf("expected only purge action, got %+v", actions)
	}
}

func TestPlanNoActionForUnknownModeAndPendingFolder(t *testing.T) {
	engine := NewEngine(Policy{MarkAfterDays: 10, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"})
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	vms := []VM{{Name: "vm", Folder: "PENDING_DELETION", PoweredOffDays: 99, Metadata: map[string]string{}}, {Name: "mark", Folder: "WORKLOADS", PoweredOffDays: 20, Metadata: map[string]string{FieldPendingSince: "x"}}}
	actions := engine.Plan(vms, Mode("other"), now)
	if len(actions) != 0 {
		t.Fatalf("expected no actions for unknown mode, got %+v", actions)
	}
}
