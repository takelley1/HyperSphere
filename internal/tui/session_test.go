// Path: internal/tui/session_test.go
// Description: Validate table interactions for selection, sorting hotkeys, and bulk actions.
package tui

import (
	"reflect"
	"testing"
)

type fakeExecutor struct {
	resource Resource
	action   string
	ids      []string
}

func (f *fakeExecutor) Execute(resource Resource, action string, ids []string) error {
	f.resource = resource
	f.action = action
	f.ids = append([]string{}, ids...)
	return nil
}

func TestVMViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(Catalog{VMs: []VMRow{{Name: "vm-a", Tags: "prod", Cluster: "c1", PowerState: "on", Datastore: "ds-1", Owner: "a@example.com"}}})
	view, err := navigator.Execute(":vm")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "TAGS", "CLUSTER", "POWER", "DATASTORE", "OWNER"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected vm columns: got %v want %v", view.Columns, want)
	}
}

func TestLUNViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(Catalog{LUNs: []LUNRow{{Name: "lun-1", Tags: "tier1", Cluster: "c1", Datastore: "san-a", CapacityGB: 100, UsedGB: 60}}})
	view, err := navigator.Execute(":lun")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "TAGS", "CLUSTER", "DATASTORE", "CAPACITY_GB", "USED_GB"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected lun columns: got %v want %v", view.Columns, want)
	}
}

func TestClusterViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(Catalog{Clusters: []ClusterRow{{Name: "cluster-a", Tags: "gold", Datacenter: "dc1", Hosts: 5, VMCount: 50, CPUUsagePercent: 60, MemUsagePercent: 58}}})
	view, err := navigator.Execute(":cluster")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "TAGS", "DATACENTER", "HOSTS", "VMS", "CPU_PERCENT", "MEM_PERCENT"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected cluster columns: got %v want %v", view.Columns, want)
	}
}

func TestSessionSpaceSelectAndBulkAction(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a", Cluster: "c1"}, {Name: "vm-b", Cluster: "c1"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	if err := session.HandleKey("SPACE"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	if err := session.HandleKey("DOWN"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	if err := session.HandleKey("SPACE"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-off", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	if executor.resource != ResourceVM || executor.action != "power-off" {
		t.Fatalf("unexpected execution context: %+v", executor)
	}
	wantIDs := []string{"vm-a", "vm-b"}
	if !reflect.DeepEqual(executor.ids, wantIDs) {
		t.Fatalf("unexpected selected ids: got %v want %v", executor.ids, wantIDs)
	}
}

func TestSessionActionFallsBackToCurrentRow(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	_ = session.ExecuteCommand(":vm")
	_ = session.HandleKey("DOWN")
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	wantIDs := []string{"vm-b"}
	if !reflect.DeepEqual(executor.ids, wantIDs) {
		t.Fatalf("unexpected selected ids: got %v want %v", executor.ids, wantIDs)
	}
}

func TestSessionSortHotkeysMirrorK9sPattern(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-z", Cluster: "c2", Owner: "z@example.com"}, {Name: "vm-a", Cluster: "c1", Owner: "a@example.com"}}})
	_ = session.ExecuteCommand(":vm")
	if err := session.HandleKey("N"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	view := session.CurrentView()
	if view.Rows[0][0] != "vm-a" {
		t.Fatalf("expected name sort asc first row vm-a, got %s", view.Rows[0][0])
	}
	if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	if err := session.HandleKey("SHIFT+O"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	view = session.CurrentView()
	if view.Rows[0][2] != "c1" {
		t.Fatalf("expected selected-column sort by cluster ascending")
	}
}

func TestSessionRejectsInvalidActionAndHotkey(t *testing.T) {
	session := NewSession(Catalog{LUNs: []LUNRow{{Name: "lun-1"}}})
	_ = session.ExecuteCommand(":lun")
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on", executor); err == nil {
		t.Fatalf("expected invalid action error")
	}
	if err := session.HandleKey("INVALID"); err == nil {
		t.Fatalf("expected invalid hotkey error")
	}
}
