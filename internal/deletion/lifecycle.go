// Path: internal/deletion/lifecycle.go
// Description: Plan and apply pending-deletion lifecycle actions using VM metadata fields.
package deletion

import "time"

const (
	FieldPendingSince       = "pd_pending_since"
	FieldDeleteOn           = "pd_delete_on"
	FieldOwnerEmail         = "pd_owner_email"
	FieldInitialNoticeSent  = "pd_initial_notice_sent"
	FieldReminderNoticeSent = "pd_reminder_notice_sent"
	FieldOriginalName       = "pd_original_name"
)

// Mode selects workflow phases.
type Mode string

const (
	ModeMark  Mode = "mark"
	ModePurge Mode = "purge"
	ModeAll   Mode = "all"
)

// ActionType identifies one lifecycle action.
type ActionType string

const (
	ActionMark   ActionType = "mark"
	ActionRemind ActionType = "remind"
	ActionPurge  ActionType = "purge"
	ActionReset  ActionType = "reset"
)

// Policy configures lifecycle timings.
type Policy struct {
	MarkAfterDays  int
	PurgeAfterDays int
	PendingFolder  string
}

// VM holds lifecycle-relevant VM state.
type VM struct {
	Name           string
	Folder         string
	PoweredOffDays int
	OwnerEmail     string
	Metadata       map[string]string
	Deleted        bool
}

// Action represents a planned lifecycle operation.
type Action struct {
	Type   ActionType
	VMName string
	Notes  string
}

// Engine plans and applies lifecycle operations.
type Engine struct {
	policy Policy
}

// NewEngine build a lifecycle engine from policy.
func NewEngine(policy Policy) Engine {
	return Engine{policy: policy}
}

// Plan generate lifecycle actions for each VM under the selected mode.
func (e Engine) Plan(vms []VM, mode Mode, now time.Time) []Action {
	actions := make([]Action, 0, len(vms))
	for _, vm := range vms {
		action, ok := e.planAction(vm, mode, now)
		if ok {
			actions = append(actions, action)
		}
	}
	return actions
}

func (e Engine) planAction(vm VM, mode Mode, now time.Time) (Action, bool) {
	if shouldReset(vm, e.policy.PendingFolder) && allows(mode, ActionReset) {
		return Action{Type: ActionReset, VMName: vm.Name}, true
	}
	if shouldPurge(vm, now, e.policy.PendingFolder) && allows(mode, ActionPurge) {
		return Action{Type: ActionPurge, VMName: vm.Name}, true
	}
	if shouldRemind(vm, now, e.policy.PurgeAfterDays, e.policy.PendingFolder) && allows(mode, ActionRemind) {
		return Action{Type: ActionRemind, VMName: vm.Name}, true
	}
	if shouldMark(vm, e.policy.MarkAfterDays, e.policy.PendingFolder) && allows(mode, ActionMark) {
		return Action{Type: ActionMark, VMName: vm.Name}, true
	}
	return Action{}, false
}

// Apply mutate VM metadata for the provided action.
func (e Engine) Apply(vm VM, action Action, now time.Time) VM {
	if vm.Metadata == nil {
		vm.Metadata = map[string]string{}
	}
	switch action.Type {
	case ActionMark:
		applyMark(&vm, e.policy.PurgeAfterDays, now)
	case ActionRemind:
		vm.Metadata[FieldReminderNoticeSent] = "true"
	case ActionPurge:
		vm.Deleted = true
	case ActionReset:
		applyReset(&vm)
	}
	return vm
}

func applyMark(vm *VM, purgeAfterDays int, now time.Time) {
	if vm.Metadata[FieldPendingSince] == "" {
		vm.Metadata[FieldPendingSince] = now.Format("2006-01-02")
		vm.Metadata[FieldDeleteOn] = now.AddDate(0, 0, purgeAfterDays).Format("2006-01-02")
		vm.Metadata[FieldOwnerEmail] = vm.OwnerEmail
		vm.Metadata[FieldOriginalName] = vm.Name
	}
	vm.Metadata[FieldInitialNoticeSent] = "true"
}

func applyReset(vm *VM) {
	if original := vm.Metadata[FieldOriginalName]; original != "" {
		vm.Name = original
	}
	delete(vm.Metadata, FieldPendingSince)
	delete(vm.Metadata, FieldDeleteOn)
	delete(vm.Metadata, FieldOwnerEmail)
	delete(vm.Metadata, FieldInitialNoticeSent)
	delete(vm.Metadata, FieldReminderNoticeSent)
	delete(vm.Metadata, FieldOriginalName)
}

func allows(mode Mode, action ActionType) bool {
	if mode == ModeAll {
		return true
	}
	if mode == ModeMark {
		return action == ActionMark
	}
	if mode == ModePurge {
		return action == ActionPurge
	}
	return false
}

func shouldMark(vm VM, markAfterDays int, pendingFolder string) bool {
	if vm.Folder == pendingFolder {
		return false
	}
	if vm.Metadata[FieldPendingSince] != "" {
		return false
	}
	return vm.PoweredOffDays >= markAfterDays
}

func shouldPurge(vm VM, now time.Time, pendingFolder string) bool {
	if vm.Folder != pendingFolder {
		return false
	}
	deleteOn := vm.Metadata[FieldDeleteOn]
	if deleteOn == "" {
		return false
	}
	dateValue, err := time.Parse("2006-01-02", deleteOn)
	if err != nil {
		return false
	}
	return !dateValue.After(now)
}

func shouldRemind(vm VM, now time.Time, purgeAfterDays int, pendingFolder string) bool {
	if vm.Folder != pendingFolder {
		return false
	}
	if vm.Metadata[FieldReminderNoticeSent] == "true" {
		return false
	}
	pendingSince, err := time.Parse("2006-01-02", vm.Metadata[FieldPendingSince])
	if err != nil {
		return false
	}
	if shouldPurge(vm, now, pendingFolder) {
		return false
	}
	halfwayDays := purgeAfterDays / 2
	halfway := pendingSince.AddDate(0, 0, halfwayDays)
	return !halfway.After(now)
}

func shouldReset(vm VM, pendingFolder string) bool {
	if vm.Folder == pendingFolder {
		return false
	}
	return vm.Metadata[FieldPendingSince] != ""
}
