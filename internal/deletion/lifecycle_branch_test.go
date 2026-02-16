// Path: internal/deletion/lifecycle_branch_test.go
// Description: Cover remaining lifecycle branches for reminders and unknown actions.
package deletion

import (
	"testing"
	"time"
)

func TestShouldRemindSkipsWhenAlreadySent(t *testing.T) {
	engine := NewEngine(Policy{MarkAfterDays: 10, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"})
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	vms := []VM{{Name: "vm", Folder: "PENDING_DELETION", Metadata: map[string]string{FieldPendingSince: "2026-02-01", FieldDeleteOn: "2026-02-25", FieldReminderNoticeSent: "true"}}}
	actions := engine.Plan(vms, ModePurge, now)
	if len(actions) != 0 {
		t.Fatalf("expected no reminder action, got %+v", actions)
	}
}

func TestApplyUnknownActionAndNilMetadata(t *testing.T) {
	engine := NewEngine(Policy{PurgeAfterDays: 10, PendingFolder: "PENDING_DELETION"})
	updated := engine.Apply(VM{Name: "vm"}, Action{Type: ActionType("unknown")}, time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC))
	if updated.Metadata == nil {
		t.Fatalf("expected metadata map initialization")
	}
	if len(updated.Metadata) != 0 {
		t.Fatalf("expected unchanged metadata for unknown action")
	}
}

func TestShouldRemindSkipsOutsidePendingFolder(t *testing.T) {
	engine := NewEngine(Policy{MarkAfterDays: 10, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"})
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	vms := []VM{{Name: "vm", Folder: "WORKLOADS", Metadata: map[string]string{FieldPendingSince: "2026-02-01", FieldDeleteOn: "2026-02-25"}}}
	actions := engine.Plan(vms, ModePurge, now)
	if len(actions) != 0 {
		t.Fatalf("expected no remind action outside pending folder, got %+v", actions)
	}
}

func TestShouldRemindSkipsWhenPurgeDue(t *testing.T) {
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	vm := VM{Folder: "PENDING_DELETION", Metadata: map[string]string{FieldPendingSince: "2026-02-01", FieldDeleteOn: "2026-02-10"}}
	if shouldRemind(vm, now, 14, "PENDING_DELETION") {
		t.Fatalf("expected no reminder when purge is due")
	}
}
