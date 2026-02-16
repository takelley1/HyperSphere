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
	want := []string{
		"NAME",
		"POWER",
		"USED_CPU_PERCENT",
		"USED_MEMORY_MB",
		"USED_STORAGE_GB",
		"IP_ADDRESS",
		"DNS_NAME",
		"CLUSTER",
		"HOST",
		"NETWORK",
		"TOTAL_CPU_CORES",
		"TOTAL_RAM_MB",
		"LARGEST_DISK_GB",
		"SNAPSHOT_COUNT",
		"SNAPSHOT_TOTAL_GB",
		"ATTACHED_STORAGE",
	}
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
	want := []string{"NAME", "TAGS", "CLUSTER", "DATASTORE", "CAPACITY_GB", "USED_GB", "FREE_GB", "UTIL_PERCENT"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected lun columns: got %v want %v", view.Columns, want)
	}
}

func TestClusterViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(Catalog{Clusters: []ClusterRow{{Name: "cluster-a", Tags: "gold", Datacenter: "dc1", Hosts: 5, VMCount: 50, CPUUsagePercent: 60, MemUsagePercent: 58, ResourcePoolCount: 4, NetworkCount: 7}}})
	view, err := navigator.Execute(":cluster")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "TAGS", "DATACENTER", "HOSTS", "VMS", "CPU_PERCENT", "MEM_PERCENT", "RESOURCE_POOLS", "NETWORKS"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected cluster columns: got %v want %v", view.Columns, want)
	}
}

func TestDatacenterViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Datacenters: []DatacenterRow{
				{Name: "dc-1", ClusterCount: 2, HostCount: 10, VMCount: 120, DatastoreCount: 8, CPUUsagePercent: 61, MemUsagePercent: 59},
			},
		},
	)
	view, err := navigator.Execute(":dc")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "CLUSTERS", "HOSTS", "VMS", "DATASTORES", "CPU_PERCENT", "MEM_PERCENT"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected datacenter columns: got %v want %v", view.Columns, want)
	}
}

func TestResourcePoolViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			ResourcePools: []ResourcePoolRow{
				{Name: "rp-prod", Cluster: "cluster-east", CPUReservationMHz: 6400, MemReservationMB: 8192, VMCount: 24, CPULimitMHz: 12000, MemLimitMB: 16384},
			},
		},
	)
	view, err := navigator.Execute(":rp")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "CLUSTER", "CPU_RES", "MEM_RES", "VM_COUNT", "CPU_LIMIT", "MEM_LIMIT"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected resource pool columns: got %v want %v", view.Columns, want)
	}
}

func TestNetworkViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Networks: []NetworkRow{
				{Name: "dvpg-prod-100", Type: "distributed-portgroup", VLAN: "100", Switch: "dvs-core", AttachedVMs: 41, MTU: 9000, Uplinks: 4},
			},
		},
	)
	view, err := navigator.Execute(":nw")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "TYPE", "VLAN", "SWITCH", "ATTACHED_VMS", "MTU", "UPLINKS"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected network columns: got %v want %v", view.Columns, want)
	}
}

func TestTemplateViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Templates: []TemplateRow{
				{Name: "tpl-rhel9-base", OS: "rhel9", Datastore: "vsan-east", Folder: "/Templates/Linux", Age: "45d", CPUCount: 4, MemoryMB: 8192},
			},
		},
	)
	view, err := navigator.Execute(":tp")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"NAME", "OS", "DATASTORE", "FOLDER", "AGE", "CPU_COUNT", "MEMORY_MB"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected template columns: got %v want %v", view.Columns, want)
	}
}

func TestSnapshotViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Snapshots: []SnapshotRow{
				{VM: "vm-a", Snapshot: "pre-upgrade", Size: "12G", Created: "2026-02-10T12:00:00Z", Age: "6d", Quiesced: "yes", Owner: "ops@example.com"},
			},
		},
	)
	view, err := navigator.Execute(":ss")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"VM", "SNAPSHOT", "SIZE", "CREATED", "AGE", "QUIESCED", "OWNER"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected snapshot columns: got %v want %v", view.Columns, want)
	}
}

func TestExpandedColumnsArePresentForEveryResourceView(t *testing.T) {
	navigator := NewNavigator(Catalog{
		VMs:           []VMRow{{Name: "vm-a"}},
		LUNs:          []LUNRow{{Name: "lun-a", CapacityGB: 100, UsedGB: 50}},
		Clusters:      []ClusterRow{{Name: "cluster-a", ResourcePoolCount: 3, NetworkCount: 5}},
		Datacenters:   []DatacenterRow{{Name: "dc-1", CPUUsagePercent: 60, MemUsagePercent: 55}},
		ResourcePools: []ResourcePoolRow{{Name: "rp-a", CPULimitMHz: 4000, MemLimitMB: 4096}},
		Networks:      []NetworkRow{{Name: "nw-a", MTU: 9000, Uplinks: 2}},
		Templates:     []TemplateRow{{Name: "tpl-a", CPUCount: 2, MemoryMB: 4096}},
		Snapshots:     []SnapshotRow{{VM: "vm-a", Snapshot: "snap-a", Owner: "ops@example.com"}},
		Hosts:         []HostRow{{Name: "host-a", CoreCount: 24, ThreadCount: 48, VMCount: 50}},
		Datastores:    []DatastoreRow{{Name: "ds-a", CapacityGB: 1000, UsedGB: 600, FreeGB: 400, Type: "vsan", LatencyMS: 3}},
	})
	cases := []struct {
		command string
		column  string
	}{
		{command: ":vm", column: "TOTAL_CPU_CORES"},
		{command: ":lun", column: "UTIL_PERCENT"},
		{command: ":cluster", column: "RESOURCE_POOLS"},
		{command: ":dc", column: "CPU_PERCENT"},
		{command: ":rp", column: "CPU_LIMIT"},
		{command: ":nw", column: "MTU"},
		{command: ":tp", column: "CPU_COUNT"},
		{command: ":ss", column: "OWNER"},
		{command: ":host", column: "THREADS"},
		{command: ":datastore", column: "LATENCY_MS"},
	}
	for _, tc := range cases {
		view, err := navigator.Execute(tc.command)
		if err != nil {
			t.Fatalf("Execute returned error for %s: %v", tc.command, err)
		}
		if findColumnIndex(view.Columns, tc.column) == -1 {
			t.Fatalf("expected column %q in view %s, got %v", tc.column, tc.command, view.Columns)
		}
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
	for step := 0; step < 7; step++ {
		if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
			t.Fatalf("HandleKey returned error: %v", err)
		}
	}
	if err := session.HandleKey("SHIFT+O"); err != nil {
		t.Fatalf("HandleKey returned error: %v", err)
	}
	view = session.CurrentView()
	if view.Rows[0][7] != "c1" {
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

func TestSelectedResourceDetailsForVMIncludesRequiredFields(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs: []VMRow{
				{
					Name:          "vm-a",
					PowerState:    "on",
					CPUCount:      8,
					MemoryMB:      16384,
					Comments:      "critical app",
					Description:   "gold workload",
					SnapshotCount: 2,
					Snapshots: []VMSnapshot{
						{Identifier: "snap-001", Timestamp: "2026-02-16T00:00:00Z"},
						{Identifier: "snap-002", Timestamp: "2026-02-16T01:00:00Z"},
					},
				},
			},
		},
	)
	details, err := session.SelectedResourceDetails()
	if err != nil {
		t.Fatalf("SelectedResourceDetails returned error: %v", err)
	}
	expectedKeys := []string{
		"NAME",
		"POWER_STATE",
		"CPU_COUNT",
		"MEMORY_MB",
		"COMMENTS",
		"DESCRIPTION",
		"SNAPSHOT_COUNT",
		"SNAPSHOT_1",
		"SNAPSHOT_2",
	}
	for _, key := range expectedKeys {
		if !hasDetailField(details.Fields, key) {
			t.Fatalf("expected details to include key %q", key)
		}
	}
	if !hasDetailFieldValue(details.Fields, "SNAPSHOT_1", "snap-001 @ 2026-02-16T00:00:00Z") {
		t.Fatalf("expected snapshot identifier and timestamp in first snapshot field")
	}
}

func TestSelectedResourceDetailsForHostUsesViewColumns(t *testing.T) {
	session := NewSession(
		Catalog{
			Hosts: []HostRow{
				{Name: "esxi-01", Cluster: "cluster-a", ConnectionState: "connected"},
			},
		},
	)
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	details, err := session.SelectedResourceDetails()
	if err != nil {
		t.Fatalf("SelectedResourceDetails returned error: %v", err)
	}
	if details.Title != "HOST DETAILS" {
		t.Fatalf("expected host details title, got %q", details.Title)
	}
	if !hasDetailFieldValue(details.Fields, "NAME", "esxi-01") {
		t.Fatalf("expected host details to include NAME field value")
	}
}

func TestSelectedResourceDetailsForVMReturnsErrorWhenIDMissingFromCatalog(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	session.view.IDs[0] = "vm-missing"
	_, err := session.SelectedResourceDetails()
	if err == nil {
		t.Fatalf("expected selected resource details error for unknown vm id")
	}
}

func TestSelectedResourceDetailsReturnsErrorWhenSelectionOutOfRange(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	session.selectedRow = 42
	_, err := session.SelectedResourceDetails()
	if err == nil {
		t.Fatalf("expected selected resource details error for out-of-range selection")
	}
}

func TestSelectedResourceDetailsForVMUsesSnapshotLengthWhenCountMissing(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs: []VMRow{
				{
					Name: "vm-a",
					Snapshots: []VMSnapshot{
						{Identifier: "snap-001", Timestamp: "2026-02-16T00:00:00Z"},
					},
				},
			},
		},
	)
	details, err := session.SelectedResourceDetails()
	if err != nil {
		t.Fatalf("SelectedResourceDetails returned error: %v", err)
	}
	if !hasDetailFieldValue(details.Fields, "SNAPSHOT_COUNT", "1") {
		t.Fatalf("expected snapshot count fallback to snapshot list length")
	}
}

func hasDetailField(fields []DetailField, key string) bool {
	for _, field := range fields {
		if field.Key == key {
			return true
		}
	}
	return false
}

func hasDetailFieldValue(fields []DetailField, key string, value string) bool {
	for _, field := range fields {
		if field.Key == key && field.Value == value {
			return true
		}
	}
	return false
}
