// Path: internal/tui/session_test.go
// Description: Validate table interactions for selection, sorting hotkeys, and bulk actions.
package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeExecutor struct {
	resource Resource
	action   string
	ids      []string
	err      error
	errors   []error
	calls    int
}

func (f *fakeExecutor) Execute(resource Resource, action string, ids []string) error {
	f.resource = resource
	f.action = action
	f.ids = append([]string{}, ids...)
	f.calls++
	if len(f.errors) > 0 {
		err := f.errors[0]
		f.errors = f.errors[1:]
		return err
	}
	return f.err
}

type retriableExecutorError struct {
	message string
}

func (e retriableExecutorError) Error() string {
	return e.message
}

func (e retriableExecutorError) Retriable() bool {
	return true
}

type fakeCanceler struct {
	resource Resource
	action   string
	ids      []string
	err      error
}

func (f *fakeCanceler) Cancel(resource Resource, action string, ids []string) error {
	f.resource = resource
	f.action = action
	f.ids = append([]string{}, ids...)
	return f.err
}

type fakePerObjectExecutor struct {
	resource Resource
	action   string
	ids      []string
	failures map[string]error
}

func (f *fakePerObjectExecutor) Execute(resource Resource, action string, ids []string) error {
	f.resource = resource
	f.action = action
	f.ids = append([]string{}, ids...)
	return nil
}

func (f *fakePerObjectExecutor) ExecuteEach(
	resource Resource,
	action string,
	ids []string,
) map[string]error {
	f.resource = resource
	f.action = action
	f.ids = append([]string{}, ids...)
	return f.failures
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

func TestVMViewActionsIncludePowerLifecycleSet(t *testing.T) {
	navigator := NewNavigator(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	view, err := navigator.Execute(":vm")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"power-on", "power-off", "reset", "suspend", "migrate", "edit-tags"}
	if !reflect.DeepEqual(view.Actions, want) {
		t.Fatalf("unexpected vm actions: got %v want %v", view.Actions, want)
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

func TestSnapshotViewActionsRequireLifecycleSet(t *testing.T) {
	navigator := NewNavigator(Catalog{Snapshots: []SnapshotRow{{VM: "vm-a", Snapshot: "snap-1"}}})
	view, err := navigator.Execute(":ss")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"create", "remove", "revert", "edit-tags"}
	if !reflect.DeepEqual(view.Actions, want) {
		t.Fatalf("unexpected snapshot actions: got %v want %v", view.Actions, want)
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

func TestDatastoreViewActionsIncludeEvacuation(t *testing.T) {
	navigator := NewNavigator(Catalog{Datastores: []DatastoreRow{{Name: "ds-1"}}})
	view, err := navigator.Execute(":datastore")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"enter-maintenance", "exit-maintenance", "evacuate", "refresh", "edit-tags"}
	if !reflect.DeepEqual(view.Actions, want) {
		t.Fatalf("unexpected datastore actions: got %v want %v", view.Actions, want)
	}
}

func TestPulseViewRendersUtilizationAndAlarmSummary(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Clusters: []ClusterRow{
				{Name: "cluster-a", CPUUsagePercent: 60, MemUsagePercent: 50},
				{Name: "cluster-b", CPUUsagePercent: 40, MemUsagePercent: 30},
			},
			Datastores: []DatastoreRow{
				{Name: "ds-1", CapacityGB: 100, UsedGB: 70},
				{Name: "ds-2", CapacityGB: 100, UsedGB: 50},
			},
			Alarms: []AlarmRow{
				{Entity: "vm-a", Status: "red"},
				{Entity: "vm-b", Status: "green"},
				{Entity: "host-a", Status: "yellow"},
			},
		},
	)
	view, err := navigator.Execute(":pulse")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	wantColumns := []string{
		"CPU_PERCENT",
		"MEM_PERCENT",
		"DATASTORE_PERCENT",
		"ACTIVE_ALARMS",
		"REFRESH_TIMER",
	}
	if !reflect.DeepEqual(view.Columns, wantColumns) {
		t.Fatalf("unexpected pulse columns: got %v want %v", view.Columns, wantColumns)
	}
	if len(view.Rows) != 1 {
		t.Fatalf("expected one pulse summary row, got %d", len(view.Rows))
	}
	wantRow := []string{"50", "40", "60", "2", "15s"}
	if !reflect.DeepEqual(view.Rows[0], wantRow) {
		t.Fatalf("unexpected pulse summary row: got %v want %v", view.Rows[0], wantRow)
	}
}

func TestPulseMetricHelpersHandleEmptyInputs(t *testing.T) {
	if value := averageClusterCPU(nil); value != 0 {
		t.Fatalf("expected zero cpu average for empty input, got %d", value)
	}
	if value := averageClusterMemory(nil); value != 0 {
		t.Fatalf("expected zero memory average for empty input, got %d", value)
	}
	if value := datastoreUsagePercent(nil); value != 0 {
		t.Fatalf("expected zero datastore usage for empty input, got %d", value)
	}
	if value := datastoreUsagePercent([]DatastoreRow{{Name: "ds-1", CapacityGB: 0, UsedGB: 10}}); value != 0 {
		t.Fatalf("expected zero datastore usage when total capacity is zero, got %d", value)
	}
}

func TestXRayViewRendersPathAndSupportsOneLevelExpansion(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs: []VMRow{
				{
					Name:      "vm-a",
					Host:      "esxi-01",
					Cluster:   "cluster-east",
					Datastore: "ds-1",
					Network:   "dvpg-10",
				},
			},
		},
	)
	if err := session.ExecuteCommand(":xray"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	view := session.CurrentView()
	if view.Resource != ResourceXRay {
		t.Fatalf("expected xray resource view, got %s", view.Resource)
	}
	if len(view.Rows) == 0 || !strings.Contains(view.Rows[0][0], "vm-a") {
		t.Fatalf("expected xray row to include selected vm path, got %v", view.Rows)
	}
	initialRowCount := len(view.Rows)
	executor := &fakeExecutor{}
	if err := session.ApplyAction("expand", executor); err != nil {
		t.Fatalf("expected xray expand action to succeed: %v", err)
	}
	expanded := session.CurrentView()
	if len(expanded.Rows) <= initialRowCount {
		t.Fatalf(
			"expected xray expand to add one-level dependencies, before=%d after=%d",
			initialRowCount,
			len(expanded.Rows),
		)
	}
	if executor.calls != 0 {
		t.Fatalf("expected xray expand to be handled locally without executor call")
	}
}

func TestXRayViewHandlesEmptyInventory(t *testing.T) {
	session := NewSession(Catalog{})
	if err := session.ExecuteCommand(":xray"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	view := session.CurrentView()
	if len(view.Rows) != 1 {
		t.Fatalf("expected placeholder xray row for empty inventory, got %d", len(view.Rows))
	}
	want := []string{"-", "none", "-"}
	if !reflect.DeepEqual(view.Rows[0], want) {
		t.Fatalf("unexpected placeholder xray row: got %v want %v", view.Rows[0], want)
	}
}

func TestSessionFaultToggleFiltersAndRestoresRows(t *testing.T) {
	session := NewSession(
		Catalog{
			Hosts: []HostRow{
				{Name: "esxi-01", ConnectionState: "connected"},
				{Name: "esxi-02", ConnectionState: "maintenance"},
			},
		},
	)
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected full host row count before fault toggle")
	}
	if err := session.HandleKey("SHIFT+F"); err != nil {
		t.Fatalf("expected fault toggle hotkey to be supported: %v", err)
	}
	filtered := session.CurrentView()
	if len(filtered.Rows) != 1 || filtered.Rows[0][0] != "esxi-02" {
		t.Fatalf("expected fault toggle to keep only faulted host rows, got %v", filtered.Rows)
	}
	if err := session.HandleKey("SHIFT+F"); err != nil {
		t.Fatalf("expected second fault toggle hotkey to restore rows: %v", err)
	}
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected second fault toggle to restore full row set")
	}
}

func TestRowHasFaultSignalMatchesKeywordsAndCleanRows(t *testing.T) {
	if !rowHasFaultSignal([]string{"disk failure detected"}) {
		t.Fatalf("expected fault substring to be detected")
	}
	if !rowHasFaultSignal([]string{"warning", "faulted"}) {
		t.Fatalf("expected fault keyword to be detected")
	}
	if rowHasFaultSignal([]string{"connected", "green", "healthy"}) {
		t.Fatalf("expected clean row to be excluded from fault mode")
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

func TestTaskViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Tasks: []TaskRow{
				{Entity: "vm-a", Action: "power-off", State: "success", Started: "2026-02-16T00:00:00Z", Duration: "31s", Owner: "ops@example.com"},
			},
		},
	)
	view, err := navigator.Execute(":task")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"ENTITY", "ACTION", "STATE", "STARTED", "DURATION", "OWNER"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected task columns: got %v want %v", view.Columns, want)
	}
}

func TestEventViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Events: []EventRow{
				{Time: "2026-02-16T09:04:00Z", Severity: "warning", Entity: "vm-a", Message: "guest tools out of date", User: "ops@example.com"},
			},
		},
	)
	view, err := navigator.Execute(":event")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"TIME", "SEVERITY", "ENTITY", "MESSAGE", "USER"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected event columns: got %v want %v", view.Columns, want)
	}
}

func TestAlarmViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Alarms: []AlarmRow{
				{Entity: "vm-a", Alarm: "CPU usage high", Status: "yellow", Triggered: "2026-02-16T10:00:00Z", AckedBy: "ops@example.com"},
			},
		},
	)
	view, err := navigator.Execute(":alarm")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"ENTITY", "ALARM", "STATUS", "TRIGGERED", "ACKED_BY"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected alarm columns: got %v want %v", view.Columns, want)
	}
}

func TestFolderViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Folders: []FolderRow{
				{Path: "/Datacenters/dc-1/vm/Prod", Type: "vm-folder", Children: 4, VMCount: 23},
			},
		},
	)
	view, err := navigator.Execute(":folder")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"PATH", "TYPE", "CHILDREN", "VM_COUNT"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected folder columns: got %v want %v", view.Columns, want)
	}
}

func TestTagViewColumnsAreRelevant(t *testing.T) {
	navigator := NewNavigator(
		Catalog{
			Tags: []TagRow{
				{Tag: "env:prod", Category: "environment", Cardinality: "single", AttachedObjects: 74},
			},
		},
	)
	view, err := navigator.Execute(":tag")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := []string{"TAG", "CATEGORY", "CARDINALITY", "ATTACHED_OBJECTS"}
	if !reflect.DeepEqual(view.Columns, want) {
		t.Fatalf("unexpected tag columns: got %v want %v", view.Columns, want)
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
		Tasks:         []TaskRow{{Entity: "vm-a", Action: "power-off", State: "running", Started: "2026-02-16T00:00:00Z", Duration: "5s", Owner: "ops@example.com"}},
		Events:        []EventRow{{Time: "2026-02-16T00:00:00Z", Severity: "info", Entity: "vm-a", Message: "vm powered on", User: "ops@example.com"}},
		Alarms:        []AlarmRow{{Entity: "vm-a", Alarm: "CPU usage high", Status: "yellow", Triggered: "2026-02-16T00:00:00Z", AckedBy: "ops@example.com"}},
		Folders:       []FolderRow{{Path: "/Datacenters/dc-1/vm/Prod", Type: "vm-folder", Children: 4, VMCount: 23}},
		Tags:          []TagRow{{Tag: "env:prod", Category: "environment", Cardinality: "single", AttachedObjects: 74}},
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
		{command: ":task", column: "DURATION"},
		{command: ":event", column: "MESSAGE"},
		{command: ":alarm", column: "ACKED_BY"},
		{command: ":folder", column: "VM_COUNT"},
		{command: ":tag", column: "ATTACHED_OBJECTS"},
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
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	if executor.resource != ResourceVM || executor.action != "power-on" {
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

func TestSessionApplyActionMigrateRequiresPlacementTarget(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("migrate", executor); err == nil {
		t.Fatalf("expected migrate action to require a target host or datastore")
	}
	if executor.calls != 0 {
		t.Fatalf("expected executor not to run when migrate target is missing")
	}
}

func TestSessionApplyActionMigrateValidatesHostTarget(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs:   []VMRow{{Name: "vm-a"}},
			Hosts: []HostRow{{Name: "esxi-01"}},
		},
	)
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("migrate host=missing", executor); err == nil {
		t.Fatalf("expected migrate action to reject unknown host target")
	}
	if err := session.ApplyAction("migrate host=esxi-01", executor); err != nil {
		t.Fatalf("expected known host target to pass validation: %v", err)
	}
	if executor.action != "migrate host=esxi-01" {
		t.Fatalf("expected migrate action to carry placement target, got %q", executor.action)
	}
}

func TestSessionApplyActionMigrateValidatesDatastoreTarget(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs:        []VMRow{{Name: "vm-a"}},
			Datastores: []DatastoreRow{{Name: "ds-01"}},
		},
	)
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("migrate datastore=missing", executor); err == nil {
		t.Fatalf("expected migrate action to reject unknown datastore target")
	}
	if err := session.ApplyAction("migrate datastore=ds-01", executor); err != nil {
		t.Fatalf("expected known datastore target to pass validation: %v", err)
	}
	if executor.action != "migrate datastore=ds-01" {
		t.Fatalf("expected migrate action to carry datastore target, got %q", executor.action)
	}
}

func TestSessionApplyActionRejectsUnsupportedOptionsForNonMigrate(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on host=esxi-01", executor); err == nil {
		t.Fatalf("expected non-migrate action options to be rejected")
	}
	if executor.calls != 0 {
		t.Fatalf("expected executor not to run for invalid option usage")
	}
}

func TestParseActionInputRejectsInvalidShapes(t *testing.T) {
	cases := []string{
		"",
		"power-on invalid",
		"power-on host=",
		"power-on =value",
	}
	for _, value := range cases {
		if _, _, err := parseActionInput(value); err == nil {
			t.Fatalf("expected parseActionInput to reject %q", value)
		}
	}
}

func TestSessionApplyActionRejectsEmptyActionString(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("   ", executor); err == nil {
		t.Fatalf("expected empty action string to return parse error")
	}
	if executor.calls != 0 {
		t.Fatalf("expected executor not to run for empty action input")
	}
}

func TestSessionApplyActionSnapshotCreateRequiresSnapshotName(t *testing.T) {
	session := NewSession(
		Catalog{
			Snapshots: []SnapshotRow{{VM: "vm-a", Snapshot: "snap-1"}},
		},
	)
	if err := session.ExecuteCommand(":ss"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("create", executor); err == nil {
		t.Fatalf("expected create to require snapshot=<name>")
	}
	if err := session.ApplyAction("create snapshot=snap-new", executor); err != nil {
		t.Fatalf("expected create snapshot action with name to succeed: %v", err)
	}
	if executor.action != "create snapshot=snap-new" {
		t.Fatalf("expected create action to carry snapshot name, got %q", executor.action)
	}
}

func TestSessionApplyActionSnapshotRemoveAndRevertValidateSnapshotID(t *testing.T) {
	session := NewSession(
		Catalog{
			Snapshots: []SnapshotRow{
				{VM: "vm-a", Snapshot: "snap-1"},
				{VM: "vm-a", Snapshot: "snap-2"},
			},
		},
	)
	if err := session.ExecuteCommand(":ss"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("remove", executor); err == nil {
		t.Fatalf("expected remove to require snapshot=<id>")
	}
	if err := session.ApplyAction("remove snapshot=missing", executor); err == nil {
		t.Fatalf("expected remove to reject unknown snapshot id")
	}
	if err := session.ApplyAction("remove snapshot=snap-1", executor); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("expected destructive remove action to require confirmation, got %v", err)
	}
	if err := session.ApplyAction("remove snapshot=snap-1", executor); err != nil {
		t.Fatalf("expected confirmed remove action to succeed: %v", err)
	}
	if executor.action != "remove snapshot=snap-1" {
		t.Fatalf("expected remove action to carry snapshot id, got %q", executor.action)
	}
	if err := session.ApplyAction("revert snapshot=missing", executor); err == nil {
		t.Fatalf("expected revert to reject unknown snapshot id")
	}
	if err := session.ApplyAction("revert snapshot=snap-2", executor); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("expected destructive revert action to require confirmation, got %v", err)
	}
	if err := session.ApplyAction("revert snapshot=snap-2", executor); err != nil {
		t.Fatalf("expected confirmed revert action to succeed: %v", err)
	}
	if executor.action != "revert snapshot=snap-2" {
		t.Fatalf("expected revert action to carry snapshot id, got %q", executor.action)
	}
}

func TestSessionApplyActionSnapshotEditTagsRejectsOptionsButAllowsBareAction(t *testing.T) {
	session := NewSession(
		Catalog{
			Snapshots: []SnapshotRow{{VM: "vm-a", Snapshot: "snap-1"}},
		},
	)
	if err := session.ExecuteCommand(":ss"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("edit-tags key=value", executor); err == nil {
		t.Fatalf("expected snapshot edit-tags with options to be rejected")
	}
	if err := session.ApplyAction("edit-tags", executor); err != nil {
		t.Fatalf("expected snapshot edit-tags without options to succeed: %v", err)
	}
	if executor.action != "edit-tags" {
		t.Fatalf("expected bare edit-tags action payload, got %q", executor.action)
	}
}

func TestSessionApplyActionHostMaintenanceUpdatesStateAndTransitions(t *testing.T) {
	session := NewSession(
		Catalog{
			Hosts: []HostRow{
				{Name: "esxi-01", ConnectionState: "connected"},
			},
		},
	)
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("enter-maintenance", executor); err != nil {
		t.Fatalf("enter-maintenance returned error: %v", err)
	}
	if value := session.CurrentView().Rows[0][5]; value != "maintenance" {
		t.Fatalf("expected host connection state to switch to maintenance, got %q", value)
	}
	transitions := session.ActionTransitions()
	if transitions[len(transitions)-1].Status != "maintenance-enabled" {
		t.Fatalf("expected maintenance-enabled transition, got %q", transitions[len(transitions)-1].Status)
	}
	if err := session.ApplyAction("exit-maintenance", executor); err != nil {
		t.Fatalf("exit-maintenance returned error: %v", err)
	}
	if value := session.CurrentView().Rows[0][5]; value != "connected" {
		t.Fatalf("expected host connection state to switch back to connected, got %q", value)
	}
	transitions = session.ActionTransitions()
	if transitions[len(transitions)-1].Status != "maintenance-disabled" {
		t.Fatalf("expected maintenance-disabled transition, got %q", transitions[len(transitions)-1].Status)
	}
}

func TestSessionApplyActionDatastoreEvacuateRequiresConfirmation(t *testing.T) {
	session := NewSession(
		Catalog{
			Datastores: []DatastoreRow{{Name: "ds-1"}},
		},
	)
	if err := session.ExecuteCommand(":datastore"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("evacuate", executor); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("expected evacuate action to require confirmation, got %v", err)
	}
	if err := session.ApplyAction("evacuate", executor); err != nil {
		t.Fatalf("expected confirmed evacuate action to succeed: %v", err)
	}
	if executor.action != "evacuate" {
		t.Fatalf("expected evacuate action payload, got %q", executor.action)
	}
}

func TestSessionApplyActionTagAssignReportsPerObjectFailures(t *testing.T) {
	session := NewSession(
		Catalog{
			Tags: []TagRow{
				{Tag: "tag-a", Category: "env"},
				{Tag: "tag-b", Category: "env"},
			},
		},
	)
	if err := session.ExecuteCommand(":tag"); err != nil {
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
	executor := &fakePerObjectExecutor{
		failures: map[string]error{"tag-b": errors.New("assign failed")},
	}
	err := session.ApplyAction("assign", executor)
	if err == nil {
		t.Fatalf("expected per-object failure to return an error")
	}
	if !strings.Contains(err.Error(), "tag-b") {
		t.Fatalf("expected failure report to mention failed id, got %v", err)
	}
	wantIDs := []string{"tag-a", "tag-b"}
	if !reflect.DeepEqual(executor.ids, wantIDs) {
		t.Fatalf("expected assign action to operate on marked ids, got %v", executor.ids)
	}
}

func TestSessionApplyActionTagAssignPerObjectSuccessAndUnassignBranch(t *testing.T) {
	session := NewSession(
		Catalog{
			Tags: []TagRow{
				{Tag: "tag-a", Category: "env"},
				{Tag: "tag-b", Category: "env"},
			},
		},
	)
	if err := session.ExecuteCommand(":tag"); err != nil {
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
	executor := &fakePerObjectExecutor{failures: map[string]error{}}
	if err := session.ApplyAction("assign", executor); err != nil {
		t.Fatalf("expected per-object assign success, got %v", err)
	}
	if err := session.ApplyAction("unassign", executor); err != nil {
		t.Fatalf("expected per-object unassign success, got %v", err)
	}
}

func TestSessionApplyActionTagAssignFallsBackWithoutPerObjectExecutor(t *testing.T) {
	session := NewSession(
		Catalog{
			Tags: []TagRow{{Tag: "tag-a", Category: "env"}},
		},
	)
	if err := session.ExecuteCommand(":tag"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("assign", executor); err != nil {
		t.Fatalf("expected assign to fall back to Execute path, got %v", err)
	}
	if executor.action != "assign" {
		t.Fatalf("expected assign action payload, got %q", executor.action)
	}
}

func TestUpdateHostConnectionCellsGuardsUnsupportedShapes(t *testing.T) {
	targets := map[string]struct{}{"esxi-01": {}}
	nonHostView := ResourceView{
		Resource: ResourceVM,
		Columns:  []string{"NAME", "POWER"},
		Rows:     [][]string{{"vm-a", "on"}},
		IDs:      []string{"vm-a"},
	}
	updateHostConnectionCells(&nonHostView, targets, "maintenance")
	if nonHostView.Rows[0][1] != "on" {
		t.Fatalf("expected non-host rows to remain unchanged")
	}

	hostWithoutConnection := ResourceView{
		Resource: ResourceHost,
		Columns:  []string{"NAME", "CPU_PERCENT"},
		Rows:     [][]string{{"esxi-01", "42"}},
		IDs:      []string{"esxi-01"},
	}
	updateHostConnectionCells(&hostWithoutConnection, targets, "maintenance")
	if hostWithoutConnection.Rows[0][1] != "42" {
		t.Fatalf("expected host rows without connection column to remain unchanged")
	}

	hostWithMismatchedRows := ResourceView{
		Resource: ResourceHost,
		Columns:  []string{"NAME", "CONNECTION"},
		Rows:     [][]string{},
		IDs:      []string{"esxi-01"},
	}
	updateHostConnectionCells(&hostWithMismatchedRows, targets, "maintenance")

	hostWithMixedTargets := ResourceView{
		Resource: ResourceHost,
		Columns:  []string{"NAME", "CONNECTION"},
		Rows: [][]string{
			{"esxi-01", "connected"},
			{"esxi-02", "connected"},
		},
		IDs: []string{"esxi-01", "esxi-02"},
	}
	updateHostConnectionCells(&hostWithMixedTargets, targets, "maintenance")
	if hostWithMixedTargets.Rows[0][1] != "maintenance" {
		t.Fatalf("expected matched target host to update connection state")
	}
	if hostWithMixedTargets.Rows[1][1] != "connected" {
		t.Fatalf("expected unmatched target host to remain unchanged")
	}
}

func TestSessionApplyActionRecordsQueuedRunningAndSuccessTransitions(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	transitions := session.ActionTransitions()
	if len(transitions) != 3 {
		t.Fatalf("expected three action transitions, got %d", len(transitions))
	}
	expected := []string{"queued", "running", "success"}
	for index, status := range expected {
		if transitions[index].Status != status {
			t.Fatalf("expected status %q at index %d, got %q", status, index, transitions[index].Status)
		}
		if transitions[index].Timestamp == "" {
			t.Fatalf("expected non-empty timestamp for transition index %d", index)
		}
	}
}

func TestSessionApplyActionRecordsFailureTransition(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{err: errors.New("boom")}
	if err := session.ApplyAction("power-on", executor); err == nil {
		t.Fatalf("expected ApplyAction to return executor error")
	}
	transitions := session.ActionTransitions()
	if len(transitions) != 3 {
		t.Fatalf("expected three action transitions, got %d", len(transitions))
	}
	if transitions[2].Status != "failure" {
		t.Fatalf("expected terminal failure transition, got %q", transitions[2].Status)
	}
}

func TestSessionCancelLastActionRecordsCancelledTransition(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	canceler := &fakeCanceler{}
	if err := session.CancelLastAction(canceler); err != nil {
		t.Fatalf("CancelLastAction returned error: %v", err)
	}
	transitions := session.ActionTransitions()
	if transitions[len(transitions)-1].Status != "cancelled" {
		t.Fatalf("expected terminal cancelled transition, got %q", transitions[len(transitions)-1].Status)
	}
	if canceler.action != "power-on" || len(canceler.ids) != 1 || canceler.ids[0] != "vm-a" {
		t.Fatalf("expected cancel request context to match last action")
	}
}

func TestSessionApplyActionRequiresConfirmationForDestructiveAction(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-off", executor); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("expected destructive action confirmation error, got %v", err)
	}
	if executor.calls != 0 {
		t.Fatalf("expected executor to not run before confirmation")
	}
	if err := session.ApplyAction("power-off", executor); err != nil {
		t.Fatalf("expected second destructive action call to confirm and execute: %v", err)
	}
	if executor.calls != 1 {
		t.Fatalf("expected executor to run once after confirmation")
	}
}

func TestSessionDenyPendingActionClearsDestructiveConfirmation(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-off", executor); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("expected destructive action confirmation error, got %v", err)
	}
	session.DenyPendingAction()
	if err := session.ApplyAction("power-off", executor); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("expected destructive action to require confirmation again after deny, got %v", err)
	}
	if executor.calls != 0 {
		t.Fatalf("expected executor to remain uncalled after deny path")
	}
}

func TestSameActionRequestMismatchBranches(t *testing.T) {
	left := actionRequest{resource: ResourceVM, action: "power-off", ids: []string{"vm-a"}}
	if sameActionRequest(left, actionRequest{resource: ResourceHost, action: "power-off", ids: []string{"vm-a"}}) {
		t.Fatalf("expected resource mismatch to be false")
	}
	if sameActionRequest(left, actionRequest{resource: ResourceVM, action: "delete", ids: []string{"vm-a"}}) {
		t.Fatalf("expected action mismatch to be false")
	}
	if sameActionRequest(left, actionRequest{resource: ResourceVM, action: "power-off", ids: []string{"vm-a", "vm-b"}}) {
		t.Fatalf("expected id length mismatch to be false")
	}
	if sameActionRequest(left, actionRequest{resource: ResourceVM, action: "power-off", ids: []string{"vm-b"}}) {
		t.Fatalf("expected id value mismatch to be false")
	}
}

func TestSessionPreviewActionIncludesTargetCountIDsAndSideEffects(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
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
	preview, err := session.PreviewAction("power-off")
	if err != nil {
		t.Fatalf("PreviewAction returned error: %v", err)
	}
	if preview.TargetCount != 2 {
		t.Fatalf("expected target count 2, got %d", preview.TargetCount)
	}
	wantIDs := []string{"vm-a", "vm-b"}
	if !reflect.DeepEqual(preview.TargetIDs, wantIDs) {
		t.Fatalf("unexpected preview target ids: got %v want %v", preview.TargetIDs, wantIDs)
	}
	if len(preview.SideEffects) == 0 {
		t.Fatalf("expected preview side effects to be populated")
	}
}

func TestSessionPreviewActionErrorBranches(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	if _, err := session.PreviewAction("not-real"); err == nil {
		t.Fatalf("expected invalid action preview error")
	}
	session.ApplyFilter("missing")
	if _, err := session.PreviewAction("power-on"); err == nil {
		t.Fatalf("expected preview error with no selected rows")
	}
}

func TestActionSideEffectsBranchCoverage(t *testing.T) {
	if len(actionSideEffects("power-on")) == 0 {
		t.Fatalf("expected power-on side effects")
	}
	if len(actionSideEffects("migrate")) == 0 {
		t.Fatalf("expected migrate side effects")
	}
	if len(actionSideEffects("unknown-action")) == 0 {
		t.Fatalf("expected default side effects")
	}
}

func TestSessionApplyActionRecordsAuditSummaryForSuccess(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
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
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	audits := session.ActionAudits()
	if len(audits) == 0 {
		t.Fatalf("expected action audit summary entry")
	}
	latest := audits[len(audits)-1]
	if latest.Actor == "" || latest.Timestamp == "" {
		t.Fatalf("expected actor and timestamp fields in action audit summary")
	}
	if latest.Action != "power-on" || latest.Outcome != "success" {
		t.Fatalf("unexpected audit action summary outcome: %+v", latest)
	}
	if len(latest.Targets) != 2 {
		t.Fatalf("expected two audit targets, got %d", len(latest.Targets))
	}
	if len(latest.FailedIDs) != 0 {
		t.Fatalf("expected no failed ids for successful action")
	}
}

func TestSessionApplyActionRecordsAuditSummaryForFailure(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{err: errors.New("boom")}
	if err := session.ApplyAction("power-on", executor); err == nil {
		t.Fatalf("expected action failure")
	}
	audits := session.ActionAudits()
	if len(audits) == 0 {
		t.Fatalf("expected action audit summary entry")
	}
	latest := audits[len(audits)-1]
	if latest.Outcome != "failure" {
		t.Fatalf("expected failure outcome in action audit summary, got %q", latest.Outcome)
	}
	if len(latest.FailedIDs) != 1 || latest.FailedIDs[0] != "vm-a" {
		t.Fatalf("expected failed ids to include failed action targets")
	}
}

func TestSessionCancelLastActionReturnsErrorWhenUnsupported(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	canceler := &fakeCanceler{}
	if err := session.CancelLastAction(canceler); err == nil {
		t.Fatalf("expected cancel error when no prior action request exists")
	}
}

func TestSessionCancelLastActionReturnsErrorWhenCancelerUnavailable(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	if err := session.CancelLastAction(nil); err == nil {
		t.Fatalf("expected cancel error when canceler is nil")
	}
}

func TestSessionCancelLastActionPropagatesCancelerError(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("ApplyAction returned error: %v", err)
	}
	canceler := &fakeCanceler{err: errors.New("cancel failed")}
	if err := session.CancelLastAction(canceler); err == nil {
		t.Fatalf("expected canceler error to be returned")
	}
	transitions := session.ActionTransitions()
	if transitions[len(transitions)-1].Status == "cancelled" {
		t.Fatalf("did not expect cancelled transition when canceler returns error")
	}
}

func TestSessionApplyActionReturnsTimeoutWhenElapsedExceedsPolicy(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	session.now = sequentialClock([]time.Time{
		time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 16, 12, 0, 2, 0, time.UTC),
		time.Date(2026, 2, 16, 12, 0, 2, 0, time.UTC),
	})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	session.SetActionTimeout("power-on", time.Second)
	executor := &fakeExecutor{}
	err := session.ApplyAction("power-on", executor)
	if err == nil {
		t.Fatalf("expected timeout error when action exceeds timeout policy")
	}
	if !errors.Is(err, ErrActionTimeout) {
		t.Fatalf("expected ErrActionTimeout, got %v", err)
	}
	transitions := session.ActionTransitions()
	if transitions[len(transitions)-1].Status != "failure" {
		t.Fatalf("expected terminal failure transition for timeout")
	}
}

func TestSessionSetActionTimeoutBranchCoverage(t *testing.T) {
	session := NewSession(Catalog{})
	session.SetActionTimeout("power-on", 2*time.Second)
	if timeout, ok := session.actionTimeouts["power-on"]; !ok || timeout != 2*time.Second {
		t.Fatalf("expected action timeout to be stored")
	}
	session.SetActionTimeout("power-on", 0)
	if _, ok := session.actionTimeouts["power-on"]; ok {
		t.Fatalf("expected non-positive timeout to remove action timeout policy")
	}
	session.SetActionTimeout("   ", time.Second)
	if len(session.actionTimeouts) != 0 {
		t.Fatalf("expected empty action name to be ignored")
	}
}

func TestSessionApplyActionRetriesRetriableErrorsUpToLimit(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	session.SetActionRetryLimit("power-on", 1)
	executor := &fakeExecutor{
		errors: []error{
			retriableExecutorError{message: "temporary"},
			nil,
		},
	}
	if err := session.ApplyAction("power-on", executor); err != nil {
		t.Fatalf("expected retry flow to succeed, got error: %v", err)
	}
	if executor.calls != 2 {
		t.Fatalf("expected two execution attempts, got %d", executor.calls)
	}
}

func TestSessionApplyActionDoesNotRetryNonRetriableError(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	session.SetActionRetryLimit("power-on", 3)
	executor := &fakeExecutor{
		errors: []error{
			errors.New("fatal"),
		},
	}
	if err := session.ApplyAction("power-on", executor); err == nil {
		t.Fatalf("expected non-retriable error")
	}
	if executor.calls != 1 {
		t.Fatalf("expected one execution attempt for non-retriable error, got %d", executor.calls)
	}
}

func TestSessionApplyActionStopsAfterRetryLimitExhausted(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	session.SetActionRetryLimit("power-on", 1)
	executor := &fakeExecutor{
		errors: []error{
			retriableExecutorError{message: "temporary-1"},
			retriableExecutorError{message: "temporary-2"},
		},
	}
	if err := session.ApplyAction("power-on", executor); err == nil {
		t.Fatalf("expected retry exhaustion error")
	}
	if executor.calls != 2 {
		t.Fatalf("expected two attempts when retry limit is one, got %d", executor.calls)
	}
}

func TestSessionSetActionRetryLimitBranchCoverage(t *testing.T) {
	session := NewSession(Catalog{})
	session.SetActionRetryLimit("power-on", 2)
	if retries, ok := session.actionRetries["power-on"]; !ok || retries != 2 {
		t.Fatalf("expected action retry policy to be stored")
	}
	session.SetActionRetryLimit("power-on", 0)
	if _, ok := session.actionRetries["power-on"]; ok {
		t.Fatalf("expected non-positive retry limit to remove retry policy")
	}
	session.SetActionRetryLimit("   ", 1)
	if len(session.actionRetries) != 0 {
		t.Fatalf("expected empty action name to be ignored")
	}
	if isRetriableError(nil) {
		t.Fatalf("expected nil error to not be retriable")
	}
}

func sequentialClock(values []time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		if len(values) == 0 {
			return time.Time{}
		}
		if index >= len(values) {
			return values[len(values)-1]
		}
		value := values[index]
		index++
		return value
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

func TestSessionVisibleColumnsPersistPerViewAndReset(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs:   []VMRow{{Name: "vm-a", PowerState: "on", Cluster: "cluster-a"}},
			Hosts: []HostRow{{Name: "esxi-01", Cluster: "cluster-a", ConnectionState: "connected"}},
		},
	)
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	if err := session.SetVisibleColumns([]string{"NAME", "POWER"}); err != nil {
		t.Fatalf("SetVisibleColumns returned error: %v", err)
	}
	if !reflect.DeepEqual(session.CurrentView().Columns, []string{"NAME", "POWER"}) {
		t.Fatalf("expected narrowed vm columns, got %v", session.CurrentView().Columns)
	}
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	if !reflect.DeepEqual(session.CurrentView().Columns, []string{"NAME", "POWER"}) {
		t.Fatalf("expected vm column selection to persist, got %v", session.CurrentView().Columns)
	}
	if err := session.ResetVisibleColumns(); err != nil {
		t.Fatalf("ResetVisibleColumns returned error: %v", err)
	}
	if findColumnIndex(session.CurrentView().Columns, "USED_CPU_PERCENT") == -1 {
		t.Fatalf("expected reset to restore full vm columns, got %v", session.CurrentView().Columns)
	}
}

func TestSessionSetVisibleColumnsRejectsUnknownColumn(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.SetVisibleColumns([]string{"NAME", "NOT_REAL"}); err == nil {
		t.Fatalf("expected invalid column selection error")
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
