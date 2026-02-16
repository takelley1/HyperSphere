// Path: internal/tui/resource_extra_test.go
// Description: Validate additional host/datastore resource views and mark controls.
package tui

import (
	"reflect"
	"testing"
)

func TestHostAndDatastoreViewsHaveRelevantColumns(t *testing.T) {
	navigator := NewNavigator(Catalog{
		Hosts:      []HostRow{{Name: "host-a", Tags: "gpu", Cluster: "c1", CPUUsagePercent: 70, MemUsagePercent: 62, ConnectionState: "connected"}},
		Datastores: []DatastoreRow{{Name: "ds-a", Tags: "nvme", Cluster: "c1", CapacityGB: 1000, UsedGB: 600, FreeGB: 400}},
	})
	hostView, err := navigator.Execute(":host")
	if err != nil {
		t.Fatalf("host view error: %v", err)
	}
	wantHost := []string{"NAME", "TAGS", "CLUSTER", "CPU_PERCENT", "MEM_PERCENT", "CONNECTION"}
	if !reflect.DeepEqual(hostView.Columns, wantHost) {
		t.Fatalf("unexpected host columns: %v", hostView.Columns)
	}
	dsView, err := navigator.Execute(":datastore")
	if err != nil {
		t.Fatalf("datastore view error: %v", err)
	}
	wantDS := []string{"NAME", "TAGS", "CLUSTER", "CAPACITY_GB", "USED_GB", "FREE_GB"}
	if !reflect.DeepEqual(dsView.Columns, wantDS) {
		t.Fatalf("unexpected datastore columns: %v", dsView.Columns)
	}
}

func TestSessionRangeMarkAndClear(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}, {Name: "vm-c"}}})
	_ = session.ExecuteCommand(":vm")
	_ = session.HandleKey("SPACE")
	_ = session.HandleKey("DOWN")
	_ = session.HandleKey("DOWN")
	if err := session.HandleKey("CTRL+SPACE"); err != nil {
		t.Fatalf("ctrl+space error: %v", err)
	}
	if len(session.marks) != 3 {
		t.Fatalf("expected range marks for 3 rows, got %d", len(session.marks))
	}
	if err := session.HandleKey("CTRL+\\"); err != nil {
		t.Fatalf("ctrl+\\ error: %v", err)
	}
	if len(session.marks) != 0 {
		t.Fatalf("expected cleared marks")
	}
}
