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
		Hosts:         defaultHostRows(),
		Datastores:    defaultDatastoreRows(),
	}
}

func defaultVMRows() []tui.VMRow {
	return []tui.VMRow{
		{Name: "vm-a", Tags: "prod,linux", Cluster: "cluster-east", PowerState: "on", Datastore: "ds-1", Owner: "a@example.com"},
		{Name: "vm-b", Tags: "dev,windows", Cluster: "cluster-west", PowerState: "off", Datastore: "ds-2", Owner: "b@example.com"},
		{Name: "vm-c", Tags: "prod,db", Cluster: "cluster-east", PowerState: "on", Datastore: "ds-3", Owner: "c@example.com"},
		{Name: "vm-d", Tags: "qa,linux", Cluster: "cluster-central", PowerState: "suspended", Datastore: "ds-4", Owner: "d@example.com"},
		{Name: "vm-e", Tags: "edge,linux", Cluster: "cluster-edge", PowerState: "on", Datastore: "ds-5", Owner: "e@example.com"},
		{Name: "vm-f", Tags: "dev,api", Cluster: "cluster-west", PowerState: "off", Datastore: "ds-6", Owner: "f@example.com"},
		{Name: "vm-g", Tags: "ops,jump", Cluster: "cluster-east", PowerState: "on", Datastore: "ds-7", Owner: "g@example.com"},
		{Name: "vm-h", Tags: "prod,cache", Cluster: "cluster-central", PowerState: "on", Datastore: "ds-8", Owner: "h@example.com"},
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
		{Name: "cluster-east", Tags: "prod", Datacenter: "dc-1", Hosts: 8, VMCount: 120, CPUUsagePercent: 63, MemUsagePercent: 58},
		{Name: "cluster-west", Tags: "dev", Datacenter: "dc-2", Hosts: 6, VMCount: 90, CPUUsagePercent: 52, MemUsagePercent: 49},
		{Name: "cluster-central", Tags: "qa", Datacenter: "dc-1", Hosts: 5, VMCount: 64, CPUUsagePercent: 57, MemUsagePercent: 55},
		{Name: "cluster-edge", Tags: "edge", Datacenter: "dc-3", Hosts: 4, VMCount: 33, CPUUsagePercent: 47, MemUsagePercent: 44},
	}
}

func defaultDatacenterRows() []tui.DatacenterRow {
	return []tui.DatacenterRow{
		{Name: "dc-1", ClusterCount: 2, HostCount: 13, VMCount: 184, DatastoreCount: 6},
		{Name: "dc-2", ClusterCount: 1, HostCount: 6, VMCount: 90, DatastoreCount: 4},
		{Name: "dc-3", ClusterCount: 1, HostCount: 4, VMCount: 33, DatastoreCount: 3},
	}
}

func defaultResourcePoolRows() []tui.ResourcePoolRow {
	return []tui.ResourcePoolRow{
		{Name: "rp-prod", Cluster: "cluster-east", CPUReservationMHz: 6400, MemReservationMB: 8192, VMCount: 24},
		{Name: "rp-dev", Cluster: "cluster-west", CPUReservationMHz: 3200, MemReservationMB: 4096, VMCount: 18},
		{Name: "rp-qa", Cluster: "cluster-central", CPUReservationMHz: 2800, MemReservationMB: 3072, VMCount: 12},
		{Name: "rp-edge", Cluster: "cluster-edge", CPUReservationMHz: 2000, MemReservationMB: 2048, VMCount: 9},
	}
}

func defaultNetworkRows() []tui.NetworkRow {
	return []tui.NetworkRow{
		{Name: "dvpg-prod-100", Type: "distributed-portgroup", VLAN: "100", Switch: "dvs-core-a", AttachedVMs: 41},
		{Name: "dvpg-dev-200", Type: "distributed-portgroup", VLAN: "200", Switch: "dvs-core-b", AttachedVMs: 27},
		{Name: "vmk-mgmt", Type: "vmkernel", VLAN: "10", Switch: "vss-mgmt-01", AttachedVMs: 8},
		{Name: "dvpg-storage-120", Type: "distributed-portgroup", VLAN: "120", Switch: "dvs-storage", AttachedVMs: 19},
		{Name: "dvpg-backup-130", Type: "distributed-portgroup", VLAN: "130", Switch: "dvs-backup", AttachedVMs: 11},
		{Name: "dvpg-edge-trunk", Type: "distributed-portgroup", VLAN: "trunk", Switch: "dvs-edge", AttachedVMs: 7},
	}
}

func defaultTemplateRows() []tui.TemplateRow {
	return []tui.TemplateRow{
		{Name: "tpl-rhel9-base", OS: "rhel9", Datastore: "vsan-east", Folder: "/Templates/Linux", Age: "45d"},
		{Name: "tpl-ubuntu2204-base", OS: "ubuntu22.04", Datastore: "vvol-central", Folder: "/Templates/Linux", Age: "32d"},
		{Name: "tpl-windows2022-base", OS: "windows2022", Datastore: "nfs-west", Folder: "/Templates/Windows", Age: "54d"},
		{Name: "tpl-sles15-base", OS: "sles15", Datastore: "vsan-east", Folder: "/Templates/Linux", Age: "27d"},
		{Name: "tpl-centos7-legacy", OS: "centos7", Datastore: "ds-6", Folder: "/Templates/Legacy", Age: "143d"},
		{Name: "tpl-debian12-app", OS: "debian12", Datastore: "ds-7", Folder: "/Templates/App", Age: "18d"},
	}
}

func defaultHostRows() []tui.HostRow {
	return []tui.HostRow{
		{Name: "esxi-01", Tags: "gpu", Cluster: "cluster-east", CPUUsagePercent: 72, MemUsagePercent: 67, ConnectionState: "connected"},
		{Name: "esxi-02", Tags: "general", Cluster: "cluster-west", CPUUsagePercent: 44, MemUsagePercent: 52, ConnectionState: "connected"},
		{Name: "esxi-03", Tags: "storage", Cluster: "cluster-central", CPUUsagePercent: 51, MemUsagePercent: 60, ConnectionState: "connected"},
		{Name: "esxi-04", Tags: "compute", Cluster: "cluster-edge", CPUUsagePercent: 38, MemUsagePercent: 41, ConnectionState: "maintenance"},
		{Name: "esxi-05", Tags: "gpu", Cluster: "cluster-east", CPUUsagePercent: 68, MemUsagePercent: 73, ConnectionState: "connected"},
		{Name: "esxi-06", Tags: "general", Cluster: "cluster-west", CPUUsagePercent: 40, MemUsagePercent: 46, ConnectionState: "disconnected"},
		{Name: "esxi-07", Tags: "network", Cluster: "cluster-central", CPUUsagePercent: 49, MemUsagePercent: 58, ConnectionState: "connected"},
		{Name: "esxi-08", Tags: "edge", Cluster: "cluster-edge", CPUUsagePercent: 36, MemUsagePercent: 39, ConnectionState: "connected"},
	}
}

func defaultDatastoreRows() []tui.DatastoreRow {
	return []tui.DatastoreRow{
		{Name: "vsan-east", Tags: "flash", Cluster: "cluster-east", CapacityGB: 8000, UsedGB: 4200, FreeGB: 3800},
		{Name: "nfs-west", Tags: "archive", Cluster: "cluster-west", CapacityGB: 12000, UsedGB: 7200, FreeGB: 4800},
		{Name: "vvol-central", Tags: "tier-1", Cluster: "cluster-central", CapacityGB: 9000, UsedGB: 5100, FreeGB: 3900},
		{Name: "iscsi-edge", Tags: "edge", Cluster: "cluster-edge", CapacityGB: 4000, UsedGB: 1900, FreeGB: 2100},
		{Name: "ds-5", Tags: "backup", Cluster: "cluster-east", CapacityGB: 6000, UsedGB: 2500, FreeGB: 3500},
		{Name: "ds-6", Tags: "dev", Cluster: "cluster-west", CapacityGB: 5500, UsedGB: 2100, FreeGB: 3400},
		{Name: "ds-7", Tags: "prod", Cluster: "cluster-central", CapacityGB: 10000, UsedGB: 6900, FreeGB: 3100},
		{Name: "ds-8", Tags: "qa", Cluster: "cluster-edge", CapacityGB: 5000, UsedGB: 2200, FreeGB: 2800},
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
