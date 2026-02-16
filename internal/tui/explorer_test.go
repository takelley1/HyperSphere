// Path: internal/tui/explorer_test.go
// Description: Validate k9s-style command navigation and resource table rendering.
package tui

import (
	"strings"
	"testing"
)

func sampleCatalog() Catalog {
	return Catalog{
		VMs:      []VMRow{{Name: "vm-a", PowerState: "on", Datastore: "ds-1", Owner: "a@example.com"}},
		LUNs:     []LUNRow{{Name: "lun-001", Datastore: "san-a", CapacityGB: 1000, UsedGB: 450}},
		Clusters: []ClusterRow{{Name: "cluster-east", Hosts: 8, VMCount: 120, CPUUsagePercent: 63}},
	}
}

func TestExecuteSwitchesViewFromColonCommand(t *testing.T) {
	navigator := NewNavigator(sampleCatalog())
	view, err := navigator.Execute(":vm")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if view.Resource != ResourceVM {
		t.Fatalf("expected vm resource, got %s", view.Resource)
	}
	if navigator.ActiveView() != ResourceVM {
		t.Fatalf("expected active view to switch to vm")
	}
}

func TestExecuteTrimsWhitespaceAndSupportsLUN(t *testing.T) {
	navigator := NewNavigator(sampleCatalog())
	view, err := navigator.Execute("  :lun   ")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if view.Resource != ResourceLUN {
		t.Fatalf("expected lun resource, got %s", view.Resource)
	}
}

func TestExecuteRejectsUnknownResource(t *testing.T) {
	navigator := NewNavigator(sampleCatalog())
	_, err := navigator.Execute(":unknown")
	if err == nil {
		t.Fatalf("expected unknown resource error")
	}
}

func TestExecuteRejectsMissingColon(t *testing.T) {
	navigator := NewNavigator(sampleCatalog())
	_, err := navigator.Execute("vm")
	if err == nil {
		t.Fatalf("expected missing colon error")
	}
}

func TestRenderResourceViewLooksLikeTable(t *testing.T) {
	navigator := NewNavigator(sampleCatalog())
	view, err := navigator.Execute(":cluster")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	out := RenderResourceView(view)
	if !strings.Contains(out, "HyperSphere") {
		t.Fatalf("expected header in output: %s", out)
	}
	if !strings.Contains(out, "cluster-east") {
		t.Fatalf("expected row in output: %s", out)
	}
	if !strings.Contains(out, "HOSTS") {
		t.Fatalf("expected uppercase columns in output: %s", out)
	}
}

func TestTableForReturnsErrorForInvalidResource(t *testing.T) {
	navigator := NewNavigator(sampleCatalog())
	_, err := navigator.TableFor(Resource("bad"))
	if err == nil {
		t.Fatalf("expected invalid resource error")
	}
}
