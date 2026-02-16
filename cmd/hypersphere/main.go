// Path: cmd/hypersphere/main.go
// Description: Provide CLI entrypoints for HyperSphere workflows and default full-screen explorer launch.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/takelley1/hypersphere/internal/app"
	"github.com/takelley1/hypersphere/internal/config"
	"github.com/takelley1/hypersphere/internal/deletion"
	"github.com/takelley1/hypersphere/internal/migration"
	"github.com/takelley1/hypersphere/internal/tui"
)

type cliFlags struct {
	command        string
	startupCommand string
	headless       bool
	crumbsless     bool
	workflow       string
	mode           string
	execute        bool
	readOnly       bool
	threshold      int
	refreshSeconds float64
	logLevel       logLevel
	logFile        string
}

type logLevel string

const (
	logLevelDebug logLevel = "debug"
	logLevelInfo  logLevel = "info"
	logLevelWarn  logLevel = "warn"
	logLevelError logLevel = "error"
)

var (
	buildVersion          = "0.0.0"
	buildCommit           = "unknown"
	buildDate             = "unknown"
	defaultRefreshSeconds = 2.0
	minimumRefreshSeconds = 1.0
)

type startupFlagValues struct {
	startupCommand *string
	headless       *bool
	crumbsless     *bool
	workflow       *string
	mode           *string
	execute        *bool
	readOnly       *bool
	write          *bool
	threshold      *int
	refresh        *float64
	level          *string
	logFile        *string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, output io.Writer, errOutput io.Writer) int {
	flags, err := parseFlags(args)
	if err != nil {
		_, _ = fmt.Fprintf(errOutput, "flag parsing failed: %v\n", err)
		return 1
	}
	if err := writeStartupLog(flags.logFile, flags.logLevel); err != nil {
		_, _ = fmt.Fprintf(errOutput, "log setup failed: %v\n", err)
		return 1
	}
	if flags.command == "version" {
		writeVersion(output)
		return 0
	}
	if flags.command == "info" {
		if err := writeInfo(output); err != nil {
			_, _ = fmt.Fprintf(errOutput, "info command failed: %v\n", err)
			return 1
		}
		return 0
	}
	cfg := config.Config{Mode: flags.mode, Execute: flags.execute, ThresholdPercent: flags.threshold}
	application := app.New(output)
	switch flags.workflow {
	case "deletion":
		runDeletionWorkflow(application, cfg)
	case "explorer":
		runExplorerWorkflow(
			os.Stdout,
			flags.readOnly,
			flags.startupCommand,
			flags.headless,
			flags.crumbsless,
		)
	default:
		runMigrationWorkflow(application, cfg)
	}
	return 0
}

func parseFlags(args []string) (cliFlags, error) {
	flagSet, values := newStartupFlagSet()
	if err := flagSet.Parse(args); err != nil {
		return cliFlags{}, err
	}
	command, err := parseSubcommand(flagSet.Args())
	if err != nil {
		return cliFlags{}, err
	}
	resolvedLevel, err := parseLogLevel(*values.level)
	if err != nil {
		return cliFlags{}, err
	}
	workflow, err := validateWorkflow(*values.workflow)
	if err != nil {
		return cliFlags{}, err
	}
	readOnly, err := resolveStartupReadOnly(*values.readOnly, *values.write)
	if err != nil {
		return cliFlags{}, err
	}
	return cliFlags{
		command:        command,
		startupCommand: normalizeStartupCommand(*values.startupCommand),
		headless:       *values.headless,
		crumbsless:     *values.crumbsless,
		workflow:       workflow,
		mode:           strings.TrimSpace(*values.mode),
		execute:        *values.execute,
		readOnly:       readOnly,
		threshold:      *values.threshold,
		refreshSeconds: clampRefreshSeconds(*values.refresh),
		logLevel:       resolvedLevel,
		logFile:        strings.TrimSpace(*values.logFile),
	}, nil
}

func newStartupFlagSet() (*flag.FlagSet, startupFlagValues) {
	flagSet := flag.NewFlagSet("hypersphere", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	values := startupFlagValues{
		startupCommand: flagSet.String("command", "", "startup resource view command"),
		headless:       flagSet.Bool("headless", false, "hide table header line"),
		crumbsless:     flagSet.Bool("crumbsless", false, "hide breadcrumb line"),
		workflow:       flagSet.String("workflow", "explorer", "workflow: explorer, migration, or deletion"),
		mode:           flagSet.String("mode", "all", "mode: mark, purge, or all"),
		execute:        flagSet.Bool("execute", false, "execute mutating actions"),
		readOnly:       flagSet.Bool("readonly", false, "start in read-only mode"),
		write:          flagSet.Bool("write", false, "override config read-only default"),
		threshold:      flagSet.Int("threshold", 85, "target utilization threshold percent"),
		refresh:        flagSet.Float64("refresh", defaultRefreshSeconds, "inventory refresh interval in seconds"),
		level:          flagSet.String("log-level", string(logLevelInfo), "log level: debug, info, warn, or error"),
		logFile:        flagSet.String("log-file", "", "path to runtime log output file"),
	}
	return flagSet, values
}

func normalizeStartupCommand(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, ":")))
}

func validateWorkflow(value string) (string, error) {
	workflow := strings.ToLower(strings.TrimSpace(value))
	if workflow == "migration" || workflow == "deletion" || workflow == "explorer" {
		return workflow, nil
	}
	return "", fmt.Errorf("unsupported workflow %q", value)
}

func resolveStartupReadOnly(readOnly bool, write bool) (bool, error) {
	if write {
		return false, nil
	}
	if readOnly {
		return true, nil
	}
	return readOnlyConfigDefault()
}

func readOnlyConfigDefault() (bool, error) {
	paths, err := infoPaths()
	if err != nil {
		return false, err
	}
	content, err := os.ReadFile(paths["config"])
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return parseReadOnlyConfigValue(string(content)), nil
}

func parseReadOnlyConfigValue(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, ":") {
			continue
		}
		fields := strings.SplitN(trimmed, ":", 2)
		if strings.ToLower(strings.TrimSpace(fields[0])) != "readonly" {
			continue
		}
		parsed, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(fields[1])))
		if err == nil {
			return parsed
		}
	}
	return false
}

func clampRefreshSeconds(refreshSeconds float64) float64 {
	if refreshSeconds < minimumRefreshSeconds {
		return minimumRefreshSeconds
	}
	return refreshSeconds
}

func parseLogLevel(value string) (logLevel, error) {
	level := logLevel(strings.ToLower(strings.TrimSpace(value)))
	if level == logLevelDebug || level == logLevelInfo || level == logLevelWarn || level == logLevelError {
		return level, nil
	}
	return "", fmt.Errorf("invalid log level %q", value)
}

func writeStartupLog(logPath string, level logLevel) error {
	trimmedPath := strings.TrimSpace(logPath)
	if trimmedPath == "" {
		return nil
	}
	logFile, err := os.OpenFile(trimmedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = logFile.Close()
	}()
	_, err = fmt.Fprintf(
		logFile,
		"time=%s level=%s message=\"startup\"\n",
		time.Now().UTC().Format(time.RFC3339),
		level,
	)
	return err
}

func parseSubcommand(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	command := strings.ToLower(strings.TrimSpace(args[0]))
	if command == "version" || command == "info" {
		return command, nil
	}
	return "", fmt.Errorf("unsupported command %q", args[0])
}

func writeVersion(output io.Writer) {
	_, _ = fmt.Fprintf(
		output,
		"version=%s commit=%s buildDate=%s\n",
		buildVersion,
		buildCommit,
		buildDate,
	)
}

func writeInfo(output io.Writer) error {
	paths, err := infoPaths()
	if err != nil {
		return err
	}
	keys := []string{"config", "logs", "dumps", "skins", "plugins", "hotkeys"}
	for _, key := range keys {
		_, _ = fmt.Fprintf(output, "%s=%s\n", key, paths[key])
	}
	return nil
}

func infoPaths() (map[string]string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configRoot, err := filepath.Abs(filepath.Join(homePath, ".hypersphere"))
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"config":  filepath.Join(configRoot, "config.yaml"),
		"logs":    filepath.Join(configRoot, "logs"),
		"dumps":   filepath.Join(configRoot, "dumps"),
		"skins":   filepath.Join(configRoot, "skins.yaml"),
		"plugins": filepath.Join(configRoot, "plugins.yaml"),
		"hotkeys": filepath.Join(configRoot, "hotkeys.yaml"),
	}, nil
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
		VMs:           defaultVMRows(),
		LUNs:          defaultLUNRows(),
		Clusters:      defaultClusterRows(),
		Datacenters:   defaultDatacenterRows(),
		ResourcePools: defaultResourcePoolRows(),
		Networks:      defaultNetworkRows(),
		Templates:     defaultTemplateRows(),
		Snapshots:     defaultSnapshotRows(),
		Tasks:         defaultTaskRows(),
		Events:        defaultEventRows(),
		Alarms:        defaultAlarmRows(),
		Folders:       defaultFolderRows(),
		Tags:          defaultTagRows(),
		Hosts:         defaultHostRows(),
		Datastores:    defaultDatastoreRows(),
	}
}

func defaultVMRows() []tui.VMRow {
	return []tui.VMRow{
		{Name: "vm-a", Tags: "prod,linux", Cluster: "cluster-east", Host: "esxi-01", Network: "dvpg-prod-100", PowerState: "on", Datastore: "ds-1", AttachedStorage: "vsan-east", IPAddress: "10.10.1.21", DNSName: "vm-a.prod.local", CPUCount: 4, MemoryMB: 8192, UsedCPUPercent: 63, UsedMemoryMB: 5632, UsedStorageGB: 76, LargestDiskGB: 80, SnapshotTotalGB: 9, Owner: "a@example.com", SnapshotCount: 2},
		{Name: "vm-b", Tags: "dev,windows", Cluster: "cluster-west", Host: "esxi-02", Network: "dvpg-dev-200", PowerState: "off", Datastore: "ds-2", AttachedStorage: "nfs-west", IPAddress: "10.20.2.34", DNSName: "vm-b.dev.local", CPUCount: 2, MemoryMB: 4096, UsedCPUPercent: 0, UsedMemoryMB: 0, UsedStorageGB: 48, LargestDiskGB: 60, SnapshotTotalGB: 4, Owner: "b@example.com", SnapshotCount: 1},
		{Name: "vm-c", Tags: "prod,db", Cluster: "cluster-east", Host: "esxi-05", Network: "dvpg-prod-100", PowerState: "on", Datastore: "ds-3", AttachedStorage: "vvol-central", IPAddress: "10.10.1.45", DNSName: "vm-c.db.local", CPUCount: 8, MemoryMB: 16384, UsedCPUPercent: 71, UsedMemoryMB: 13240, UsedStorageGB: 220, LargestDiskGB: 200, SnapshotTotalGB: 26, Owner: "c@example.com", SnapshotCount: 3},
		{Name: "vm-d", Tags: "qa,linux", Cluster: "cluster-central", Host: "esxi-03", Network: "dvpg-storage-120", PowerState: "suspended", Datastore: "ds-4", AttachedStorage: "iscsi-edge", IPAddress: "10.30.3.18", DNSName: "vm-d.qa.local", CPUCount: 2, MemoryMB: 6144, UsedCPUPercent: 9, UsedMemoryMB: 840, UsedStorageGB: 32, LargestDiskGB: 40, SnapshotTotalGB: 2, Owner: "d@example.com", SnapshotCount: 1},
		{Name: "vm-e", Tags: "edge,linux", Cluster: "cluster-edge", Host: "esxi-04", Network: "dvpg-edge-trunk", PowerState: "on", Datastore: "ds-5", AttachedStorage: "ds-5", IPAddress: "172.16.40.11", DNSName: "vm-e.edge.local", CPUCount: 4, MemoryMB: 8192, UsedCPUPercent: 44, UsedMemoryMB: 2980, UsedStorageGB: 54, LargestDiskGB: 64, SnapshotTotalGB: 0, Owner: "e@example.com", SnapshotCount: 0},
		{Name: "vm-f", Tags: "dev,api", Cluster: "cluster-west", Host: "esxi-06", Network: "dvpg-dev-200", PowerState: "off", Datastore: "ds-6", AttachedStorage: "ds-6", IPAddress: "10.20.2.58", DNSName: "vm-f.api.local", CPUCount: 6, MemoryMB: 12288, UsedCPUPercent: 0, UsedMemoryMB: 0, UsedStorageGB: 89, LargestDiskGB: 100, SnapshotTotalGB: 7, Owner: "f@example.com", SnapshotCount: 2},
		{Name: "vm-g", Tags: "ops,jump", Cluster: "cluster-east", Host: "esxi-07", Network: "vmk-mgmt", PowerState: "on", Datastore: "ds-7", AttachedStorage: "ds-7", IPAddress: "10.50.5.7", DNSName: "vm-g.ops.local", CPUCount: 2, MemoryMB: 4096, UsedCPUPercent: 37, UsedMemoryMB: 2112, UsedStorageGB: 28, LargestDiskGB: 32, SnapshotTotalGB: 0, Owner: "g@example.com", SnapshotCount: 0},
		{Name: "vm-h", Tags: "prod,cache", Cluster: "cluster-central", Host: "esxi-08", Network: "dvpg-storage-120", PowerState: "on", Datastore: "ds-8", AttachedStorage: "ds-8", IPAddress: "10.30.3.88", DNSName: "vm-h.cache.local", CPUCount: 4, MemoryMB: 8192, UsedCPUPercent: 58, UsedMemoryMB: 4760, UsedStorageGB: 66, LargestDiskGB: 80, SnapshotTotalGB: 3, Owner: "h@example.com", SnapshotCount: 1},
	}
}

func defaultLUNRows() []tui.LUNRow {
	return []tui.LUNRow{
		{Name: "lun-001", Tags: "gold", Cluster: "cluster-east", Datastore: "san-a", CapacityGB: 1000, UsedGB: 450},
		{Name: "lun-002", Tags: "silver", Cluster: "cluster-west", Datastore: "san-b", CapacityGB: 2000, UsedGB: 900},
		{Name: "lun-003", Tags: "bronze", Cluster: "cluster-central", Datastore: "san-c", CapacityGB: 1500, UsedGB: 700},
		{Name: "lun-004", Tags: "archive", Cluster: "cluster-edge", Datastore: "san-d", CapacityGB: 3000, UsedGB: 1200},
		{Name: "lun-005", Tags: "flash", Cluster: "cluster-east", Datastore: "san-e", CapacityGB: 1200, UsedGB: 840},
		{Name: "lun-006", Tags: "backup", Cluster: "cluster-west", Datastore: "san-f", CapacityGB: 2500, UsedGB: 1250},
		{Name: "lun-007", Tags: "gold", Cluster: "cluster-central", Datastore: "san-g", CapacityGB: 1800, UsedGB: 1080},
		{Name: "lun-008", Tags: "silver", Cluster: "cluster-edge", Datastore: "san-h", CapacityGB: 1600, UsedGB: 640},
	}
}

func defaultClusterRows() []tui.ClusterRow {
	return []tui.ClusterRow{
		{Name: "cluster-east", Tags: "prod", Datacenter: "dc-1", Hosts: 8, VMCount: 120, CPUUsagePercent: 63, MemUsagePercent: 58, ResourcePoolCount: 4, NetworkCount: 9},
		{Name: "cluster-west", Tags: "dev", Datacenter: "dc-2", Hosts: 6, VMCount: 90, CPUUsagePercent: 52, MemUsagePercent: 49, ResourcePoolCount: 3, NetworkCount: 7},
		{Name: "cluster-central", Tags: "qa", Datacenter: "dc-1", Hosts: 5, VMCount: 64, CPUUsagePercent: 57, MemUsagePercent: 55, ResourcePoolCount: 2, NetworkCount: 6},
		{Name: "cluster-edge", Tags: "edge", Datacenter: "dc-3", Hosts: 4, VMCount: 33, CPUUsagePercent: 47, MemUsagePercent: 44, ResourcePoolCount: 2, NetworkCount: 5},
	}
}

func defaultDatacenterRows() []tui.DatacenterRow {
	return []tui.DatacenterRow{
		{Name: "dc-1", ClusterCount: 2, HostCount: 13, VMCount: 184, DatastoreCount: 6, CPUUsagePercent: 60, MemUsagePercent: 57},
		{Name: "dc-2", ClusterCount: 1, HostCount: 6, VMCount: 90, DatastoreCount: 4, CPUUsagePercent: 52, MemUsagePercent: 49},
		{Name: "dc-3", ClusterCount: 1, HostCount: 4, VMCount: 33, DatastoreCount: 3, CPUUsagePercent: 46, MemUsagePercent: 43},
	}
}

func defaultResourcePoolRows() []tui.ResourcePoolRow {
	return []tui.ResourcePoolRow{
		{Name: "rp-prod", Cluster: "cluster-east", CPUReservationMHz: 6400, MemReservationMB: 8192, VMCount: 24, CPULimitMHz: 12000, MemLimitMB: 16384},
		{Name: "rp-dev", Cluster: "cluster-west", CPUReservationMHz: 3200, MemReservationMB: 4096, VMCount: 18, CPULimitMHz: 9000, MemLimitMB: 12288},
		{Name: "rp-qa", Cluster: "cluster-central", CPUReservationMHz: 2800, MemReservationMB: 3072, VMCount: 12, CPULimitMHz: 7000, MemLimitMB: 10240},
		{Name: "rp-edge", Cluster: "cluster-edge", CPUReservationMHz: 2000, MemReservationMB: 2048, VMCount: 9, CPULimitMHz: 5000, MemLimitMB: 8192},
	}
}

func defaultNetworkRows() []tui.NetworkRow {
	return []tui.NetworkRow{
		{Name: "dvpg-prod-100", Type: "distributed-portgroup", VLAN: "100", Switch: "dvs-core-a", AttachedVMs: 41, MTU: 9000, Uplinks: 4},
		{Name: "dvpg-dev-200", Type: "distributed-portgroup", VLAN: "200", Switch: "dvs-core-b", AttachedVMs: 27, MTU: 9000, Uplinks: 4},
		{Name: "vmk-mgmt", Type: "vmkernel", VLAN: "10", Switch: "vss-mgmt-01", AttachedVMs: 8, MTU: 1500, Uplinks: 2},
		{Name: "dvpg-storage-120", Type: "distributed-portgroup", VLAN: "120", Switch: "dvs-storage", AttachedVMs: 19, MTU: 9000, Uplinks: 2},
		{Name: "dvpg-backup-130", Type: "distributed-portgroup", VLAN: "130", Switch: "dvs-backup", AttachedVMs: 11, MTU: 9000, Uplinks: 2},
		{Name: "dvpg-edge-trunk", Type: "distributed-portgroup", VLAN: "trunk", Switch: "dvs-edge", AttachedVMs: 7, MTU: 1600, Uplinks: 2},
	}
}

func defaultTemplateRows() []tui.TemplateRow {
	return []tui.TemplateRow{
		{Name: "tpl-rhel9-base", OS: "rhel9", Datastore: "vsan-east", Folder: "/Templates/Linux", Age: "45d", CPUCount: 4, MemoryMB: 8192},
		{Name: "tpl-ubuntu2204-base", OS: "ubuntu22.04", Datastore: "vvol-central", Folder: "/Templates/Linux", Age: "32d", CPUCount: 2, MemoryMB: 4096},
		{Name: "tpl-windows2022-base", OS: "windows2022", Datastore: "nfs-west", Folder: "/Templates/Windows", Age: "54d", CPUCount: 4, MemoryMB: 8192},
		{Name: "tpl-sles15-base", OS: "sles15", Datastore: "vsan-east", Folder: "/Templates/Linux", Age: "27d", CPUCount: 2, MemoryMB: 4096},
		{Name: "tpl-centos7-legacy", OS: "centos7", Datastore: "ds-6", Folder: "/Templates/Legacy", Age: "143d", CPUCount: 2, MemoryMB: 2048},
		{Name: "tpl-debian12-app", OS: "debian12", Datastore: "ds-7", Folder: "/Templates/App", Age: "18d", CPUCount: 4, MemoryMB: 6144},
	}
}

func defaultSnapshotRows() []tui.SnapshotRow {
	return []tui.SnapshotRow{
		{VM: "vm-a", Snapshot: "pre-patch", Size: "12G", Created: "2026-02-10T12:00:00Z", Age: "6d", Quiesced: "yes", Owner: "a@example.com"},
		{VM: "vm-b", Snapshot: "before-upgrade", Size: "8G", Created: "2026-02-08T09:15:00Z", Age: "8d", Quiesced: "no", Owner: "b@example.com"},
		{VM: "vm-c", Snapshot: "monthly-backup", Size: "24G", Created: "2026-01-31T22:40:00Z", Age: "16d", Quiesced: "yes", Owner: "c@example.com"},
		{VM: "vm-d", Snapshot: "pre-maintenance", Size: "6G", Created: "2026-02-14T03:05:00Z", Age: "2d", Quiesced: "no", Owner: "d@example.com"},
		{VM: "vm-e", Snapshot: "schema-change", Size: "10G", Created: "2026-02-12T18:20:00Z", Age: "4d", Quiesced: "yes", Owner: "e@example.com"},
		{VM: "vm-f", Snapshot: "pre-hotfix", Size: "5G", Created: "2026-02-15T07:55:00Z", Age: "1d", Quiesced: "no", Owner: "f@example.com"},
	}
}

func defaultTaskRows() []tui.TaskRow {
	return []tui.TaskRow{
		{Entity: "vm-a", Action: "power-off", State: "success", Started: "2026-02-16T08:10:00Z", Duration: "24s", Owner: "ops@example.com"},
		{Entity: "vm-b", Action: "clone", State: "running", Started: "2026-02-16T08:16:00Z", Duration: "2m14s", Owner: "dev@example.com"},
		{Entity: "esxi-01", Action: "enter-maintenance", State: "queued", Started: "2026-02-16T08:20:00Z", Duration: "0s", Owner: "infra@example.com"},
		{Entity: "ds-7", Action: "rescan", State: "success", Started: "2026-02-16T08:01:00Z", Duration: "11s", Owner: "storage@example.com"},
		{Entity: "vm-c", Action: "snapshot-create", State: "failed", Started: "2026-02-16T07:55:00Z", Duration: "38s", Owner: "dba@example.com"},
		{Entity: "cluster-west", Action: "rebalance", State: "running", Started: "2026-02-16T08:05:00Z", Duration: "6m02s", Owner: "sre@example.com"},
	}
}

func defaultEventRows() []tui.EventRow {
	return []tui.EventRow{
		{Time: "2026-02-16T08:04:00Z", Severity: "info", Entity: "vm-a", Message: "power state changed to on", User: "ops@example.com"},
		{Time: "2026-02-16T08:09:00Z", Severity: "warning", Entity: "esxi-06", Message: "host entered disconnected state", User: "infra@example.com"},
		{Time: "2026-02-16T08:12:00Z", Severity: "error", Entity: "ds-7", Message: "datastore latency threshold exceeded", User: "storage@example.com"},
		{Time: "2026-02-16T08:18:00Z", Severity: "info", Entity: "vm-b", Message: "snapshot created", User: "dev@example.com"},
		{Time: "2026-02-16T08:20:00Z", Severity: "warning", Entity: "cluster-west", Message: "demand imbalance detected", User: "sre@example.com"},
		{Time: "2026-02-16T08:23:00Z", Severity: "info", Entity: "vm-c", Message: "guest tools upgraded", User: "dba@example.com"},
	}
}

func defaultAlarmRows() []tui.AlarmRow {
	return []tui.AlarmRow{
		{Entity: "vm-a", Alarm: "CPU usage high", Status: "yellow", Triggered: "2026-02-16T08:05:00Z", AckedBy: "-"},
		{Entity: "vm-c", Alarm: "Datastore latency critical", Status: "red", Triggered: "2026-02-16T08:12:00Z", AckedBy: "storage@example.com"},
		{Entity: "esxi-06", Alarm: "Host disconnected", Status: "red", Triggered: "2026-02-16T08:09:00Z", AckedBy: "infra@example.com"},
		{Entity: "cluster-west", Alarm: "Imbalance detected", Status: "yellow", Triggered: "2026-02-16T08:20:00Z", AckedBy: "-"},
		{Entity: "ds-7", Alarm: "Space utilization warning", Status: "yellow", Triggered: "2026-02-16T08:18:00Z", AckedBy: "ops@example.com"},
		{Entity: "vm-b", Alarm: "Snapshot chain length high", Status: "yellow", Triggered: "2026-02-16T08:22:00Z", AckedBy: "-"},
	}
}

func defaultFolderRows() []tui.FolderRow {
	return []tui.FolderRow{
		{Path: "/Datacenters/dc-1/vm/Prod", Type: "vm-folder", Children: 6, VMCount: 74},
		{Path: "/Datacenters/dc-1/vm/QA", Type: "vm-folder", Children: 3, VMCount: 28},
		{Path: "/Datacenters/dc-2/vm/Dev", Type: "vm-folder", Children: 5, VMCount: 52},
		{Path: "/Datacenters/dc-3/vm/Edge", Type: "vm-folder", Children: 2, VMCount: 17},
		{Path: "/Datacenters/dc-1/host/Compute", Type: "host-folder", Children: 4, VMCount: 0},
		{Path: "/Datacenters/dc-2/network/Distributed", Type: "network-folder", Children: 7, VMCount: 0},
	}
}

func defaultTagRows() []tui.TagRow {
	return []tui.TagRow{
		{Tag: "env:prod", Category: "environment", Cardinality: "single", AttachedObjects: 74},
		{Tag: "env:dev", Category: "environment", Cardinality: "single", AttachedObjects: 53},
		{Tag: "tier:gold", Category: "service-tier", Cardinality: "single", AttachedObjects: 38},
		{Tag: "tier:silver", Category: "service-tier", Cardinality: "single", AttachedObjects: 47},
		{Tag: "backup:daily", Category: "backup-policy", Cardinality: "multiple", AttachedObjects: 26},
		{Tag: "compliance:pci", Category: "compliance", Cardinality: "multiple", AttachedObjects: 12},
	}
}

func defaultHostRows() []tui.HostRow {
	return []tui.HostRow{
		{Name: "esxi-01", Tags: "gpu", Cluster: "cluster-east", CPUUsagePercent: 72, MemUsagePercent: 67, ConnectionState: "connected", CoreCount: 24, ThreadCount: 48, VMCount: 29},
		{Name: "esxi-02", Tags: "general", Cluster: "cluster-west", CPUUsagePercent: 44, MemUsagePercent: 52, ConnectionState: "connected", CoreCount: 20, ThreadCount: 40, VMCount: 21},
		{Name: "esxi-03", Tags: "storage", Cluster: "cluster-central", CPUUsagePercent: 51, MemUsagePercent: 60, ConnectionState: "connected", CoreCount: 20, ThreadCount: 40, VMCount: 17},
		{Name: "esxi-04", Tags: "compute", Cluster: "cluster-edge", CPUUsagePercent: 38, MemUsagePercent: 41, ConnectionState: "maintenance", CoreCount: 16, ThreadCount: 32, VMCount: 9},
		{Name: "esxi-05", Tags: "gpu", Cluster: "cluster-east", CPUUsagePercent: 68, MemUsagePercent: 73, ConnectionState: "connected", CoreCount: 24, ThreadCount: 48, VMCount: 26},
		{Name: "esxi-06", Tags: "general", Cluster: "cluster-west", CPUUsagePercent: 40, MemUsagePercent: 46, ConnectionState: "disconnected", CoreCount: 20, ThreadCount: 40, VMCount: 14},
		{Name: "esxi-07", Tags: "network", Cluster: "cluster-central", CPUUsagePercent: 49, MemUsagePercent: 58, ConnectionState: "connected", CoreCount: 20, ThreadCount: 40, VMCount: 15},
		{Name: "esxi-08", Tags: "edge", Cluster: "cluster-edge", CPUUsagePercent: 36, MemUsagePercent: 39, ConnectionState: "connected", CoreCount: 16, ThreadCount: 32, VMCount: 11},
	}
}

func defaultDatastoreRows() []tui.DatastoreRow {
	return []tui.DatastoreRow{
		{Name: "vsan-east", Tags: "flash", Cluster: "cluster-east", CapacityGB: 8000, UsedGB: 4200, FreeGB: 3800, Type: "vsan", LatencyMS: 2},
		{Name: "nfs-west", Tags: "archive", Cluster: "cluster-west", CapacityGB: 12000, UsedGB: 7200, FreeGB: 4800, Type: "nfs", LatencyMS: 6},
		{Name: "vvol-central", Tags: "tier-1", Cluster: "cluster-central", CapacityGB: 9000, UsedGB: 5100, FreeGB: 3900, Type: "vvol", LatencyMS: 4},
		{Name: "iscsi-edge", Tags: "edge", Cluster: "cluster-edge", CapacityGB: 4000, UsedGB: 1900, FreeGB: 2100, Type: "iscsi", LatencyMS: 5},
		{Name: "ds-5", Tags: "backup", Cluster: "cluster-east", CapacityGB: 6000, UsedGB: 2500, FreeGB: 3500, Type: "nfs", LatencyMS: 7},
		{Name: "ds-6", Tags: "dev", Cluster: "cluster-west", CapacityGB: 5500, UsedGB: 2100, FreeGB: 3400, Type: "vsan", LatencyMS: 3},
		{Name: "ds-7", Tags: "prod", Cluster: "cluster-central", CapacityGB: 10000, UsedGB: 6900, FreeGB: 3100, Type: "vvol", LatencyMS: 4},
		{Name: "ds-8", Tags: "qa", Cluster: "cluster-edge", CapacityGB: 5000, UsedGB: 2200, FreeGB: 2800, Type: "iscsi", LatencyMS: 5},
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
