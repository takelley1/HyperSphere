// Path: internal/tui/runtime_test.go
// Description: Validate extended runtime parity behaviors: filter, last-view, and read-only gating.
package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseExplorerInputFilterAndLastView(t *testing.T) {
	filterCmd, err := ParseExplorerInput("/vm-a")
	if err != nil {
		t.Fatalf("unexpected filter parse error: %v", err)
	}
	if filterCmd.Kind != CommandFilter || filterCmd.Value != "vm-a" {
		t.Fatalf("unexpected filter command: %+v", filterCmd)
	}
	lastCmd, err := ParseExplorerInput(":-")
	if err != nil {
		t.Fatalf("unexpected last-view parse error: %v", err)
	}
	if lastCmd.Kind != CommandLastView {
		t.Fatalf("unexpected last-view command: %+v", lastCmd)
	}
}

func TestSessionApplyFilterAndClear(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a", Owner: "a@example.com"}, {Name: "vm-b", Owner: "b@example.com"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.ApplyFilter("vm-a")
	if len(session.CurrentView().Rows) != 1 {
		t.Fatalf("expected one filtered row")
	}
	session.ApplyFilter("")
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected filter clear to restore rows")
	}
}

func TestSessionApplyRegexFilterAndRejectInvalidPattern(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}, {Name: "db-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.ApplyRegexFilter("^vm-"); err != nil {
		t.Fatalf("ApplyRegexFilter error: %v", err)
	}
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected regex filter to keep two vm rows")
	}
	if err := session.ApplyRegexFilter("["); err == nil {
		t.Fatalf("expected invalid regex error")
	}
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected invalid regex to keep prior filtered rows")
	}
	if err := session.ApplyRegexFilter("   "); err != nil {
		t.Fatalf("expected empty regex pattern to clear filter: %v", err)
	}
	if len(session.CurrentView().Rows) != 3 {
		t.Fatalf("expected empty regex pattern to restore all rows")
	}
}

func TestSessionApplyInverseRegexFilter(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}, {Name: "db-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.ApplyInverseRegexFilter("^vm-"); err != nil {
		t.Fatalf("ApplyInverseRegexFilter error: %v", err)
	}
	if len(session.CurrentView().Rows) != 1 || session.CurrentView().Rows[0][0] != "db-a" {
		t.Fatalf("expected inverse regex filter to exclude matching vm rows")
	}
}

func TestSessionApplyTagFilterRequiresAllTagPairs(t *testing.T) {
	session := NewSession(
		Catalog{
			Hosts: []HostRow{
				{Name: "host-a", Tags: "env=prod,tier=gold"},
				{Name: "host-b", Tags: "env=prod,tier=silver"},
				{Name: "host-c", Tags: "env=dev,tier=gold"},
			},
		},
	)
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.ApplyTagFilter("env=prod,tier=gold"); err != nil {
		t.Fatalf("ApplyTagFilter error: %v", err)
	}
	if len(session.CurrentView().Rows) != 1 || session.CurrentView().Rows[0][0] != "host-a" {
		t.Fatalf("expected tag filter to require all requested tag pairs")
	}
}

func TestSessionApplyTagFilterBranchCoverage(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.ApplyTagFilter("env=prod"); err != nil {
		t.Fatalf("expected tag filter to succeed with no tags column: %v", err)
	}
	if len(session.CurrentView().Rows) != 0 {
		t.Fatalf("expected no rows when applying tag filter without TAGS column")
	}
	if err := session.ApplyTagFilter(""); err == nil {
		t.Fatalf("expected empty tag filter error")
	}
	if err := session.ApplyTagFilter("env"); err == nil {
		t.Fatalf("expected invalid tag filter expression error")
	}
	if rowMatchesTags([]string{"env=prod"}, 3, []string{"env=prod"}) {
		t.Fatalf("expected out-of-range tag index to fail match")
	}
	if rowMatchesTags([]string{"env=prod"}, -1, []string{"env=prod"}) {
		t.Fatalf("expected negative tag index to fail match")
	}
	if rowMatchesTags([]string{"env=prod,tier=silver"}, 0, []string{"env=prod", "tier=gold"}) {
		t.Fatalf("expected missing tag criterion to fail match")
	}
	if !rowMatchesTags([]string{"env=prod,,tier=gold"}, 0, []string{"env=prod", "tier=gold"}) {
		t.Fatalf("expected empty tag tokens to be ignored for positive matches")
	}
}

func TestSessionLastViewToggle(t *testing.T) {
	session := NewSession(Catalog{})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("unexpected vm error: %v", err)
	}
	if err := session.ExecuteCommand(":cluster"); err != nil {
		t.Fatalf("unexpected cluster error: %v", err)
	}
	if err := session.LastView(); err != nil {
		t.Fatalf("unexpected LastView error: %v", err)
	}
	if session.CurrentView().Resource != ResourceVM {
		t.Fatalf("expected last view toggle to vm, got %s", session.CurrentView().Resource)
	}
}

func TestSessionBreadcrumbPathHierarchyAndLastView(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs:         []VMRow{{Name: "vm-a", Cluster: "cluster-a", Host: "esxi-01"}},
			Clusters:    []ClusterRow{{Name: "cluster-a", Datacenter: "dc-1"}},
			Datacenters: []DatacenterRow{{Name: "dc-1"}},
			Hosts:       []HostRow{{Name: "esxi-01", Cluster: "cluster-a"}},
		},
	)
	if err := session.ExecuteCommand(":dc"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if session.BreadcrumbPath() != "home > dc-1" {
		t.Fatalf("unexpected datacenter breadcrumb: %q", session.BreadcrumbPath())
	}
	if err := session.ExecuteCommand(":cluster"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if session.BreadcrumbPath() != "home > dc-1 > cluster-a" {
		t.Fatalf("unexpected cluster breadcrumb: %q", session.BreadcrumbPath())
	}
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if session.BreadcrumbPath() != "home > dc-1 > cluster-a > esxi-01" {
		t.Fatalf("unexpected host breadcrumb: %q", session.BreadcrumbPath())
	}
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if session.BreadcrumbPath() != "home > dc-1 > cluster-a > esxi-01 > vm-a" {
		t.Fatalf("unexpected vm breadcrumb: %q", session.BreadcrumbPath())
	}
	if err := session.LastView(); err != nil {
		t.Fatalf("LastView error: %v", err)
	}
	if session.BreadcrumbPath() != "home > dc-1 > cluster-a > esxi-01" {
		t.Fatalf("unexpected breadcrumb after last view: %q", session.BreadcrumbPath())
	}
}

func TestSessionBreadcrumbPathFallbackBranches(t *testing.T) {
	missing := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a", Cluster: "cluster-x", Host: "esxi-x"}}})
	if err := missing.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if missing.BreadcrumbPath() != "home > cluster-x > esxi-x > vm-a" {
		t.Fatalf("unexpected vm breadcrumb without datacenter mapping: %q", missing.BreadcrumbPath())
	}
	missing.view.IDs[0] = "missing-vm"
	if missing.BreadcrumbPath() != "home > vm" {
		t.Fatalf("unexpected vm lookup fallback breadcrumb: %q", missing.BreadcrumbPath())
	}

	empty := NewSession(Catalog{})
	if err := empty.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if empty.BreadcrumbPath() != "home > vm" {
		t.Fatalf("unexpected empty vm fallback breadcrumb: %q", empty.BreadcrumbPath())
	}
	if err := empty.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if empty.BreadcrumbPath() != "home > host" {
		t.Fatalf("unexpected empty host fallback breadcrumb: %q", empty.BreadcrumbPath())
	}
	if err := empty.ExecuteCommand(":cluster"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if empty.BreadcrumbPath() != "home > cluster" {
		t.Fatalf("unexpected empty cluster fallback breadcrumb: %q", empty.BreadcrumbPath())
	}
	if err := empty.ExecuteCommand(":dc"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if empty.BreadcrumbPath() != "home > datacenter" {
		t.Fatalf("unexpected empty datacenter fallback breadcrumb: %q", empty.BreadcrumbPath())
	}

	session := NewSession(
		Catalog{
			Hosts:       []HostRow{{Name: "esxi-01", Cluster: "cluster-a"}},
			Clusters:    []ClusterRow{{Name: "cluster-a", Datacenter: "dc-1"}},
			Datacenters: []DatacenterRow{{Name: "dc-1"}},
			LUNs:        []LUNRow{{Name: "lun-1", CapacityGB: 10, UsedGB: 2}},
		},
	)
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.view.IDs[0] = "missing-host"
	if session.BreadcrumbPath() != "home > host" {
		t.Fatalf("unexpected host fallback breadcrumb: %q", session.BreadcrumbPath())
	}
	if err := session.ExecuteCommand(":cluster"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.view.IDs[0] = "missing-cluster"
	if session.BreadcrumbPath() != "home > cluster" {
		t.Fatalf("unexpected cluster fallback breadcrumb: %q", session.BreadcrumbPath())
	}
	if err := session.ExecuteCommand(":dc"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.view.IDs[0] = "missing-dc"
	if session.BreadcrumbPath() != "home > datacenter" {
		t.Fatalf("unexpected datacenter fallback breadcrumb: %q", session.BreadcrumbPath())
	}
	if err := session.ExecuteCommand(":lun"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if session.BreadcrumbPath() != "home > lun" {
		t.Fatalf("unexpected default-resource breadcrumb: %q", session.BreadcrumbPath())
	}
}

func TestSessionShiftJOwnerJumpMovesFromVMToHost(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs:      []VMRow{{Name: "vm-a", Host: "esxi-01", Cluster: "cluster-a"}},
			Hosts:    []HostRow{{Name: "esxi-01", Cluster: "cluster-a"}, {Name: "esxi-02", Cluster: "cluster-a"}},
			Clusters: []ClusterRow{{Name: "cluster-a", Datacenter: "dc-1"}},
		},
	)
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.HandleKey("SHIFT+J"); err != nil {
		t.Fatalf("HandleKey SHIFT+J error: %v", err)
	}
	if session.CurrentView().Resource != ResourceHost {
		t.Fatalf("expected owner jump to host view, got %s", session.CurrentView().Resource)
	}
	if session.CurrentView().IDs[session.SelectedRow()] != "esxi-01" {
		t.Fatalf("expected owner jump to focus owning host row")
	}
}

func TestSessionShiftJOwnerJumpFallsBackToResourcePool(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs:           []VMRow{{Name: "vm-a", Cluster: "cluster-a"}},
			ResourcePools: []ResourcePoolRow{{Name: "rp-prod", Cluster: "cluster-a"}},
		},
	)
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.HandleKey("SHIFT+J"); err != nil {
		t.Fatalf("HandleKey SHIFT+J error: %v", err)
	}
	if session.CurrentView().Resource != ResourcePool {
		t.Fatalf("expected owner jump to resource pool view, got %s", session.CurrentView().Resource)
	}
	if session.CurrentView().IDs[session.SelectedRow()] != "rp-prod" {
		t.Fatalf("expected owner jump to focus owning resource pool row")
	}
}

func TestSessionShiftJOwnerJumpErrorBranches(t *testing.T) {
	notVM := NewSession(Catalog{Hosts: []HostRow{{Name: "esxi-01"}}})
	if err := notVM.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := notVM.jumpToOwner(); err == nil {
		t.Fatalf("expected owner jump error for non-vm view")
	}

	emptyVM := NewSession(Catalog{})
	if err := emptyVM.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := emptyVM.jumpToOwner(); err == nil {
		t.Fatalf("expected owner jump error for vm view without rows")
	}

	missingVM := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a", Host: "esxi-01"}}})
	if err := missingVM.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	missingVM.view.IDs[0] = "vm-missing"
	if err := missingVM.jumpToOwner(); err == nil {
		t.Fatalf("expected owner jump error for missing vm lookup")
	}

	unowned := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := unowned.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := unowned.jumpToOwner(); err == nil {
		t.Fatalf("expected owner jump error when no host or resource pool owner exists")
	}
}

func TestOwnerJumpHelperBranches(t *testing.T) {
	session := NewSession(Catalog{
		Hosts:         []HostRow{{Name: "esxi-01"}},
		ResourcePools: []ResourcePoolRow{{Name: "rp-a", Cluster: "cluster-a"}},
	})
	if session.jumpToOwnedRow(":unknown", "id") {
		t.Fatalf("expected jumpToOwnedRow to fail for invalid command")
	}
	if session.jumpToOwnedRow(":host", "missing") {
		t.Fatalf("expected jumpToOwnedRow to fail when target row id is absent")
	}
	if containsHostName(session.navigator.catalog.Hosts, "missing") {
		t.Fatalf("did not expect missing host name to be found")
	}
	if _, ok := firstResourcePoolForCluster(session.navigator.catalog.ResourcePools, "missing"); ok {
		t.Fatalf("did not expect missing cluster resource pool mapping")
	}
	if indexOfID([]string{"a", "b"}, "missing") != -1 {
		t.Fatalf("expected indexOfID to return -1 for absent target")
	}
}

func TestSessionShiftWWarpFromFolderToScopedVMView(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs: []VMRow{
				{Name: "vm-prod-a", Tags: "prod"},
				{Name: "vm-dev-a", Tags: "dev"},
			},
			Folders: []FolderRow{
				{Path: "/Datacenters/dc-1/vm/Prod", Type: "vm-folder", Children: 2, VMCount: 1},
			},
		},
	)
	if err := session.ExecuteCommand(":folder"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.HandleKey("SHIFT+W"); err != nil {
		t.Fatalf("HandleKey SHIFT+W error: %v", err)
	}
	if session.CurrentView().Resource != ResourceVM {
		t.Fatalf("expected warp to open vm view, got %s", session.CurrentView().Resource)
	}
	if len(session.CurrentView().Rows) != 1 || session.CurrentView().Rows[0][0] != "vm-prod-a" {
		t.Fatalf("expected folder warp to scope vm rows by selected key")
	}
}

func TestSessionShiftWWarpFromTagToScopedVMView(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs: []VMRow{
				{Name: "vm-prod-a", Tags: "prod"},
				{Name: "vm-dev-a", Tags: "dev"},
			},
			Tags: []TagRow{
				{Tag: "prod", Category: "environment", Cardinality: "single", AttachedObjects: 1},
			},
		},
	)
	if err := session.ExecuteCommand(":tag"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.HandleKey("SHIFT+W"); err != nil {
		t.Fatalf("HandleKey SHIFT+W error: %v", err)
	}
	if session.CurrentView().Resource != ResourceVM {
		t.Fatalf("expected warp to open vm view, got %s", session.CurrentView().Resource)
	}
	if len(session.CurrentView().Rows) != 1 || session.CurrentView().Rows[0][0] != "vm-prod-a" {
		t.Fatalf("expected tag warp to scope vm rows by selected key")
	}
}

func TestSessionShiftWWarpErrorBranches(t *testing.T) {
	notSupported := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := notSupported.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := notSupported.HandleKey("SHIFT+W"); err == nil {
		t.Fatalf("expected shift+w error outside folder/tag views")
	}

	emptyFolder := NewSession(Catalog{Folders: []FolderRow{{Path: "/", Type: "vm-folder"}}})
	if err := emptyFolder.ExecuteCommand(":folder"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := emptyFolder.HandleKey("SHIFT+W"); err == nil {
		t.Fatalf("expected shift+w error for folder selection with empty scope key")
	}

	noRows := NewSession(Catalog{})
	if err := noRows.ExecuteCommand(":folder"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := noRows.HandleKey("SHIFT+W"); err == nil {
		t.Fatalf("expected shift+w error for folder view without selected rows")
	}
}

func TestSessionShiftWWarpReturnsExecuteErrorWhenVMViewCannotBeBuilt(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs:     []VMRow{{Name: "vm-a", Tags: "prod"}},
			Folders: []FolderRow{{Path: "/Datacenters/dc-1/vm/Prod", Type: "vm-folder"}},
		},
	)
	if err := session.ExecuteCommand(":folder"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.columnSelection[ResourceVM] = []string{"NOT_REAL"}
	if err := session.HandleKey("SHIFT+W"); err == nil {
		t.Fatalf("expected shift+w to return vm execute error for invalid stored vm columns")
	}
}

func TestSessionSearchJumpCyclesFilteredMatchesWithWrap(t *testing.T) {
	session := NewSession(
		Catalog{
			VMs: []VMRow{
				{Name: "prod-a", Tags: "prod"},
				{Name: "prod-b", Tags: "prod"},
				{Name: "dev-a", Tags: "dev"},
			},
		},
	)
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.ApplyFilter("prod")
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected filtered vm rows to include two prod matches")
	}
	if err := session.HandleKey("n"); err != nil {
		t.Fatalf("HandleKey n error: %v", err)
	}
	if session.SelectedRow() != 1 {
		t.Fatalf("expected n to move to next filtered row, got %d", session.SelectedRow())
	}
	if err := session.HandleKey("n"); err != nil {
		t.Fatalf("HandleKey n error: %v", err)
	}
	if session.SelectedRow() != 0 {
		t.Fatalf("expected n wrap to first filtered row, got %d", session.SelectedRow())
	}
	if err := session.HandleKey("N"); err != nil {
		t.Fatalf("HandleKey N error: %v", err)
	}
	if session.SelectedRow() != 1 {
		t.Fatalf("expected N wrap to last filtered row, got %d", session.SelectedRow())
	}
}

func TestSessionLastViewFailsWithoutHistory(t *testing.T) {
	session := NewSession(Catalog{})
	if err := session.LastView(); err == nil {
		t.Fatalf("expected last view error when no history")
	}
}

func TestReadOnlyBlocksActions(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	_ = session.ExecuteCommand(":vm")
	session.SetReadOnly(true)
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-off", executor); err == nil {
		t.Fatalf("expected read-only action rejection")
	}
	if !session.ReadOnly() {
		t.Fatalf("expected read-only getter true")
	}
	if !strings.Contains(session.Render(), "Mode: RO") {
		t.Fatalf("expected read-only mode indicator in render")
	}
	session.SetReadOnly(false)
	if session.ReadOnly() {
		t.Fatalf("expected read-only getter false")
	}
	if !strings.Contains(session.Render(), "Mode: RW") {
		t.Fatalf("expected read-write mode indicator in render")
	}
}

func TestSortInvertHotkey(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-z"}}})
	_ = session.ExecuteCommand(":vm")
	_ = session.HandleKey("N")
	if session.CurrentView().Rows[0][0] != "vm-a" {
		t.Fatalf("expected ascending name sort")
	}
	if err := session.HandleKey("SHIFT+I"); err != nil {
		t.Fatalf("unexpected shift+i error: %v", err)
	}
	if session.CurrentView().Rows[0][0] != "vm-z" {
		t.Fatalf("expected inverted name sort")
	}
}

func TestSortedColumnHeaderUsesDirectionGlyph(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-z"}, {Name: "vm-a"}}})
	_ = session.ExecuteCommand(":vm")
	_ = session.HandleKey("N")
	ascendingHeader := strings.Split(session.Render(), "\n")[1]
	if !strings.Contains(ascendingHeader, "[NAME↑]") {
		t.Fatalf("expected ascending glyph in selected header: %s", ascendingHeader)
	}
	_ = session.HandleKey("N")
	descendingHeader := strings.Split(session.Render(), "\n")[1]
	if !strings.Contains(descendingHeader, "[NAME↓]") {
		t.Fatalf("expected descending glyph in selected header: %s", descendingHeader)
	}
}

func TestLastViewReturnsUnderlyingExecuteError(t *testing.T) {
	session := NewSession(Catalog{})
	session.previousView = Resource("unknown")
	if err := session.LastView(); err == nil {
		t.Fatalf("expected last view execute error")
	}
}

func TestSpanMarkEdgeBranchesAndInvertSortError(t *testing.T) {
	empty := NewSession(Catalog{})
	empty.view = ResourceView{}
	empty.spanMark()
	if len(empty.marks) != 0 {
		t.Fatalf("expected no marks for empty span")
	}

	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	_ = session.ExecuteCommand(":vm")
	session.markAnchor = 99
	session.selectedRow = 1
	session.spanMark()
	if len(session.marks) != 1 {
		t.Fatalf("expected fallback toggle mark when anchor invalid")
	}

	noSort := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	_ = noSort.ExecuteCommand(":vm")
	if err := noSort.invertSort(); err == nil {
		t.Fatalf("expected invert sort error when no active sort")
	}
}

func TestSpanMarkSwapsRangeBounds(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}, {Name: "vm-c"}}})
	_ = session.ExecuteCommand(":vm")
	session.selectedRow = 2
	session.toggleMark()
	session.selectedRow = 0
	session.spanMark()
	if len(session.marks) != 3 {
		t.Fatalf("expected full range marks after swapped bounds, got %d", len(session.marks))
	}
}

func TestSessionSelectionAndMarkAccessors(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	_ = session.ExecuteCommand(":vm")
	if err := session.HandleKey("DOWN"); err != nil {
		t.Fatalf("unexpected down key error: %v", err)
	}
	if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("unexpected shift+right error: %v", err)
	}
	if session.SelectedRow() != 1 || session.SelectedColumn() != 1 {
		t.Fatalf("unexpected selected coordinates %d,%d", session.SelectedRow(), session.SelectedColumn())
	}
	if err := session.HandleKey("SPACE"); err != nil {
		t.Fatalf("unexpected mark key error: %v", err)
	}
	if !session.IsMarked("vm-b") {
		t.Fatalf("expected vm-b to be marked")
	}
	if session.IsMarked("vm-a") {
		t.Fatalf("did not expect vm-a to be marked")
	}
}

func TestMarkCountBadgeUpdatesForMarkUnmarkAndClear(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if !strings.Contains(session.Render(), "Marks[0]") {
		t.Fatalf("expected zero mark badge in header")
	}
	if err := session.HandleKey("SPACE"); err != nil {
		t.Fatalf("unexpected mark key error: %v", err)
	}
	if !strings.Contains(session.Render(), "Marks[1]") {
		t.Fatalf("expected one mark badge after marking")
	}
	if err := session.HandleKey("SPACE"); err != nil {
		t.Fatalf("unexpected unmark key error: %v", err)
	}
	if !strings.Contains(session.Render(), "Marks[0]") {
		t.Fatalf("expected zero mark badge after unmarking")
	}
	if err := session.HandleKey("SPACE"); err != nil {
		t.Fatalf("unexpected mark key error: %v", err)
	}
	if err := session.HandleKey("CTRL+\\"); err != nil {
		t.Fatalf("unexpected clear marks key error: %v", err)
	}
	if !strings.Contains(session.Render(), "Marks[0]") {
		t.Fatalf("expected zero mark badge after clear")
	}
}

func TestArrowColumnMovementWithoutShift(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a", Tags: "prod"}}})
	_ = session.ExecuteCommand(":vm")
	if err := session.HandleKey("RIGHT"); err != nil {
		t.Fatalf("expected right arrow to move column: %v", err)
	}
	if session.SelectedColumn() != 1 {
		t.Fatalf("expected selected column 1 after right arrow, got %d", session.SelectedColumn())
	}
	if err := session.HandleKey("LEFT"); err != nil {
		t.Fatalf("expected left arrow to move column: %v", err)
	}
	if session.SelectedColumn() != 0 {
		t.Fatalf("expected selected column 0 after left arrow, got %d", session.SelectedColumn())
	}
}

func TestStickyHeaderPersistsAcrossVerticalScroll(t *testing.T) {
	session := NewSession(Catalog{VMs: manyVMRows(12)})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	topFrame := session.Render()
	header := "M  >  [NAME]  POWER  USED_CPU_PERCENT"
	if !strings.Contains(topFrame, header) {
		t.Fatalf("expected header row in top frame: %s", topFrame)
	}
	for step := 0; step < 11; step++ {
		if err := session.HandleKey("DOWN"); err != nil {
			t.Fatalf("unexpected down key error: %v", err)
		}
	}
	bottomFrame := session.Render()
	if !strings.Contains(bottomFrame, header) {
		t.Fatalf("expected sticky header row in bottom frame: %s", bottomFrame)
	}
	if strings.Contains(bottomFrame, "vm-00") {
		t.Fatalf("expected scrolled body rows to change: %s", bottomFrame)
	}
	if !strings.Contains(bottomFrame, "vm-11") {
		t.Fatalf("expected selected row near end to be visible: %s", bottomFrame)
	}
}

func TestSessionColumnSelectionBranchCoverage(t *testing.T) {
	session := NewSession(Catalog{
		VMs: []VMRow{
			{Name: "vm-a", PowerState: "on"},
			{Name: "vm-b", PowerState: "off"},
		},
		Hosts: []HostRow{{Name: "esxi-01"}},
	})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.ApplyFilter("vm-a")
	if err := session.SetVisibleColumns([]string{" NAME ", "", "POWER", "POWER"}); err != nil {
		t.Fatalf("SetVisibleColumns error: %v", err)
	}
	if len(session.CurrentView().Rows) != 1 {
		t.Fatalf("expected filter to remain applied after SetVisibleColumns")
	}
	visible := session.VisibleColumns()
	if len(visible) != 2 || visible[0] != "NAME" || visible[1] != "POWER" {
		t.Fatalf("unexpected visible columns: %v", visible)
	}
	available, err := session.AvailableColumns()
	if err != nil {
		t.Fatalf("AvailableColumns error: %v", err)
	}
	if len(available) <= len(visible) {
		t.Fatalf("expected available columns to include full view schema")
	}
	if err := session.ResetVisibleColumns(); err != nil {
		t.Fatalf("ResetVisibleColumns error: %v", err)
	}
	if len(session.CurrentView().Rows) != 1 {
		t.Fatalf("expected filter to remain applied after ResetVisibleColumns")
	}
	if err := session.SetVisibleColumns([]string{"", "   "}); err == nil {
		t.Fatalf("expected invalid empty column selection")
	}

	session.view.Resource = Resource("unknown")
	if _, err := session.AvailableColumns(); err == nil {
		t.Fatalf("expected AvailableColumns error for unknown resource")
	}
	if err := session.SetVisibleColumns([]string{"NAME"}); err == nil {
		t.Fatalf("expected SetVisibleColumns error for unknown resource")
	}
	if err := session.ResetVisibleColumns(); err == nil {
		t.Fatalf("expected ResetVisibleColumns error for unknown resource")
	}

	session = NewSession(Catalog{Hosts: []HostRow{{Name: "esxi-01"}}})
	session.columnSelection[ResourceHost] = []string{"NOT_REAL"}
	if err := session.ExecuteCommand(":host"); err == nil {
		t.Fatalf("expected ExecuteCommand error for invalid stored columns")
	}
}

func manyVMRows(count int) []VMRow {
	rows := make([]VMRow, 0, count)
	for index := 0; index < count; index++ {
		rows = append(rows, VMRow{Name: fmt.Sprintf("vm-%02d", index)})
	}
	return rows
}
