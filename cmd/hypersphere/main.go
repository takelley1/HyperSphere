// Path: cmd/hypersphere/main.go
// Description: Provide CLI entrypoints for HyperSphere workflows and default full-screen explorer launch.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/takelley1/hypersphere/internal/app"
	"github.com/takelley1/hypersphere/internal/config"
	"github.com/takelley1/hypersphere/internal/deletion"
	"github.com/takelley1/hypersphere/internal/migration"
	"github.com/takelley1/hypersphere/internal/tui"
)

type cliFlags struct {
	workflow  string
	mode      string
	execute   bool
	threshold int
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "flag parsing failed: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Config{Mode: flags.mode, Execute: flags.execute, ThresholdPercent: flags.threshold}
	application := app.New(os.Stdout)
	switch flags.workflow {
	case "deletion":
		runDeletionWorkflow(application, cfg)
	case "explorer":
		runExplorerWorkflow(os.Stdout)
	default:
		runMigrationWorkflow(application, cfg)
	}
}

func parseFlags() (cliFlags, error) {
	workflow := flag.String("workflow", "explorer", "workflow: explorer, migration, or deletion")
	mode := flag.String("mode", "all", "mode: mark, purge, or all")
	execute := flag.Bool("execute", false, "execute mutating actions")
	threshold := flag.Int("threshold", 85, "target utilization threshold percent")
	flag.Parse()
	value := strings.ToLower(strings.TrimSpace(*workflow))
	if value != "migration" && value != "deletion" && value != "explorer" {
		return cliFlags{}, fmt.Errorf("unsupported workflow %q", *workflow)
	}
	return cliFlags{workflow: value, mode: strings.TrimSpace(*mode), execute: *execute, threshold: *threshold}, nil
}

func runMigrationWorkflow(application app.App, cfg config.Config) {
	planner := migration.NewPlanner(cfg.ThresholdPercent)
	vms := []migration.VM{{Name: "example-vm-01", SizeGB: 15, SourceDatastore: "source-ds"}}
	stores := []migration.Datastore{{Name: "source-ds", CapacityGB: 100, UsedGB: 30, Tier: migration.TierPrimary}, {Name: "target-ds", CapacityGB: 100, UsedGB: 20, Tier: migration.TierPrimary}}
	_ = application.RunMigration(cfg, vms, stores, planner)
}

func runDeletionWorkflow(application app.App, cfg config.Config) {
	engine := deletion.NewEngine(deletion.Policy{MarkAfterDays: 30, PurgeAfterDays: 14, PendingFolder: "PENDING_DELETION"})
	adapter := deletionAdapter{engine: engine}
	vms := []deletion.VM{{Name: "example-vm-02", Folder: "WORKLOADS", PoweredOffDays: 45, OwnerEmail: "owner@example.com", Metadata: map[string]string{}}}
	mode := deletion.Mode(cfg.Mode)
	_ = application.RunDeletion(vms, mode, app.TimeValue{Value: time.Now().UTC()}, adapter)
}

func defaultCatalog() tui.Catalog {
	return tui.Catalog{
		VMs:        []tui.VMRow{{Name: "vm-a", Tags: "prod,linux", Cluster: "cluster-east", PowerState: "on", Datastore: "ds-1", Owner: "a@example.com"}, {Name: "vm-b", Tags: "dev,windows", Cluster: "cluster-west", PowerState: "off", Datastore: "ds-2", Owner: "b@example.com"}},
		LUNs:       []tui.LUNRow{{Name: "lun-001", Tags: "gold", Cluster: "cluster-east", Datastore: "san-a", CapacityGB: 1000, UsedGB: 450}, {Name: "lun-002", Tags: "silver", Cluster: "cluster-west", Datastore: "san-b", CapacityGB: 2000, UsedGB: 900}},
		Clusters:   []tui.ClusterRow{{Name: "cluster-east", Tags: "prod", Datacenter: "dc-1", Hosts: 8, VMCount: 120, CPUUsagePercent: 63, MemUsagePercent: 58}, {Name: "cluster-west", Tags: "dev", Datacenter: "dc-2", Hosts: 6, VMCount: 90, CPUUsagePercent: 52, MemUsagePercent: 49}},
		Hosts:      []tui.HostRow{{Name: "esxi-01", Tags: "gpu", Cluster: "cluster-east", CPUUsagePercent: 72, MemUsagePercent: 67, ConnectionState: "connected"}, {Name: "esxi-02", Tags: "general", Cluster: "cluster-west", CPUUsagePercent: 44, MemUsagePercent: 52, ConnectionState: "connected"}},
		Datastores: []tui.DatastoreRow{{Name: "vsan-east", Tags: "flash", Cluster: "cluster-east", CapacityGB: 8000, UsedGB: 4200, FreeGB: 3800}, {Name: "nfs-west", Tags: "archive", Cluster: "cluster-west", CapacityGB: 12000, UsedGB: 7200, FreeGB: 4800}},
	}
}

type deletionAdapter struct {
	engine deletion.Engine
}

func (d deletionAdapter) Plan(vms []deletion.VM, mode deletion.Mode, now app.TimeValue) []deletion.Action {
	resolved := now.Value
	if resolved.IsZero() {
		resolved = time.Now().UTC()
	}
	return d.engine.Plan(vms, mode, resolved)
}
