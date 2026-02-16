// Path: cmd/hypersphere/main.go
// Description: Provide CLI entrypoints for HyperSphere workflows and command-mode resource views.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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
		runExplorerWorkflow(os.Stdin, os.Stdout)
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

func runExplorerWorkflow(input io.Reader, output io.Writer) {
	session := tui.NewSession(defaultCatalog())
	executor := cliActionExecutor{out: output}
	prompt := tui.NewPromptState(200)
	printExplorerHelp(output)
	_, _ = fmt.Fprint(output, session.Render())
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if !handleExplorerLine(&session, &prompt, executor, output, scanner.Text()) {
			return
		}
		_, _ = fmt.Fprint(output, session.Render())
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintf(output, "input error: %v\n", err)
	}
}

func handleExplorerLine(
	session *tui.Session,
	prompt *tui.PromptState,
	executor tui.ActionExecutor,
	output io.Writer,
	line string,
) bool {
	parsed, err := tui.ParseExplorerInput(line)
	if err != nil {
		emitIfError(output, err)
		return true
	}
	if shouldRecordHistory(parsed.Kind) {
		prompt.Record(line)
	}
	switch parsed.Kind {
	case tui.CommandNoop:
		return true
	case tui.CommandQuit:
		return false
	case tui.CommandHelp:
		printExplorerHelp(output)
		return true
	case tui.CommandReadOnly:
		applyReadOnlyMode(session, parsed.Value)
		return true
	case tui.CommandHistory:
		handleHistoryCommand(prompt, parsed.Value, output)
		return true
	case tui.CommandSuggest:
		handleSuggestCommand(prompt, parsed.Value, session.CurrentView(), output)
		return true
	case tui.CommandLastView:
		emitIfError(output, session.LastView())
		return true
	case tui.CommandFilter:
		session.ApplyFilter(parsed.Value)
		return true
	case tui.CommandView:
		emitIfError(output, session.ExecuteCommand(":"+parsed.Value))
		return true
	case tui.CommandAction:
		emitIfError(output, session.ApplyAction(parsed.Value, executor))
		return true
	default:
		emitIfError(output, session.HandleKey(parsed.Value))
		return true
	}
}

func shouldRecordHistory(kind tui.CommandKind) bool {
	return kind != tui.CommandNoop && kind != tui.CommandHistory
}

func applyReadOnlyMode(session *tui.Session, mode string) {
	switch mode {
	case "on":
		session.SetReadOnly(true)
	case "off":
		session.SetReadOnly(false)
	default:
		session.SetReadOnly(!session.ReadOnly())
	}
}

func handleHistoryCommand(prompt *tui.PromptState, direction string, output io.Writer) {
	entry, ok := readHistoryEntry(prompt, direction)
	if !ok {
		_, _ = fmt.Fprintln(output, "history: <none>")
		return
	}
	_, _ = fmt.Fprintf(output, "history: %s\n", entry)
}

func readHistoryEntry(prompt *tui.PromptState, direction string) (string, bool) {
	if direction == "up" {
		return prompt.Previous()
	}
	return prompt.Next()
}

func handleSuggestCommand(
	prompt *tui.PromptState,
	prefix string,
	view tui.ResourceView,
	output io.Writer,
) {
	suggestions := prompt.Suggest(prefix, view)
	if len(suggestions) == 0 {
		_, _ = fmt.Fprintln(output, "suggestions: <none>")
		return
	}
	_, _ = fmt.Fprintf(output, "suggestions: %s\n", strings.Join(suggestions, ", "))
}

func emitIfError(output io.Writer, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(output, "command error: %v\n", err)
	}
}

func printExplorerHelp(output io.Writer) {
	_, _ = fmt.Fprintln(output, "Command mode ready.")
	_, _ = fmt.Fprintln(output, "Views: :vm :lun :cluster :host :datastore | Quit: :q")
	_, _ = fmt.Fprintln(output, "Navigation: :- toggles previous view")
	_, _ = fmt.Fprintln(output, "Aliases: :vms :luns :hosts :ds")
	_, _ = fmt.Fprintln(output, "Filter: /text (clear with /)")
	_, _ = fmt.Fprintln(output, "Modes: :ro [on|off|toggle] for read-only")
	_, _ = fmt.Fprintln(output, "Prompt: :history up/down and :suggest <prefix>")
	_, _ = fmt.Fprintln(output, "Hotkeys: SPACE, CTRL+SPACE, CTRL+\\, J/K, SHIFT+LEFT, SHIFT+RIGHT, SHIFT+O")
	_, _ = fmt.Fprintln(output, "Actions: !<action> (for example !power-off in :vm)")
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

type cliActionExecutor struct {
	out io.Writer
}

func (c cliActionExecutor) Execute(resource tui.Resource, action string, ids []string) error {
	_, _ = fmt.Fprintf(c.out, "vmware-api action=%s resource=%s targets=%s\n", action, resource, strings.Join(ids, ","))
	return nil
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
