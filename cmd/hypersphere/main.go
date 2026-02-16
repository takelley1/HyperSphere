// Path: cmd/hypersphere/main.go
// Description: Provide CLI entrypoints for HyperSphere workflows and default full-screen explorer launch.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	workflow       string
	mode           string
	execute        bool
	threshold      int
	refreshSeconds float64
	logLevel       logLevel
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

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, output io.Writer, errOutput io.Writer) int {
	flags, err := parseFlags(args)
	if err != nil {
		_, _ = fmt.Fprintf(errOutput, "flag parsing failed: %v\n", err)
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
		runExplorerWorkflow(os.Stdout)
	default:
		runMigrationWorkflow(application, cfg)
	}
	return 0
}

func parseFlags(args []string) (cliFlags, error) {
	flagSet := flag.NewFlagSet("hypersphere", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	workflow := flagSet.String("workflow", "explorer", "workflow: explorer, migration, or deletion")
	mode := flagSet.String("mode", "all", "mode: mark, purge, or all")
	execute := flagSet.Bool("execute", false, "execute mutating actions")
	threshold := flagSet.Int("threshold", 85, "target utilization threshold percent")
	refresh := flagSet.Float64("refresh", defaultRefreshSeconds, "inventory refresh interval in seconds")
	level := flagSet.String("log-level", string(logLevelInfo), "log level: debug, info, warn, or error")
	if err := flagSet.Parse(args); err != nil {
		return cliFlags{}, err
	}
	command, err := parseSubcommand(flagSet.Args())
	if err != nil {
		return cliFlags{}, err
	}
	resolvedLevel, err := parseLogLevel(*level)
	if err != nil {
		return cliFlags{}, err
	}
	value := strings.ToLower(strings.TrimSpace(*workflow))
	if value != "migration" && value != "deletion" && value != "explorer" {
		return cliFlags{}, fmt.Errorf("unsupported workflow %q", *workflow)
	}
	return cliFlags{
		command:        command,
		workflow:       value,
		mode:           strings.TrimSpace(*mode),
		execute:        *execute,
		threshold:      *threshold,
		refreshSeconds: clampRefreshSeconds(*refresh),
		logLevel:       resolvedLevel,
	}, nil
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
