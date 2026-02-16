// Path: internal/migration/planner.go
// Description: Plan and execute datastore migrations with threshold and retry controls.
package migration

import "sort"

const (
	SkipOverThreshold    = "OVER_85"
	SkipNoEligibleTarget = "NO_ELIGIBLE_TARGET"
)

// Tier labels target datastore priority.
type Tier string

const (
	TierPrimary   Tier = "primary"
	TierSecondary Tier = "secondary"
	TierTertiary  Tier = "tertiary"
)

func (t Tier) String() string {
	return string(t)
}

// VM represents a VM migration candidate.
type VM struct {
	Name            string
	SizeGB          int
	SourceDatastore string
}

// Datastore represents migration destination capacity.
type Datastore struct {
	Name       string
	CapacityGB int
	UsedGB     int
	Tier       Tier
}

// PlanStep stores one planned migration operation.
type PlanStep struct {
	Order           int
	VMName          string
	SourceDatastore string
	TargetDatastore string
	ProjectedUtil   int
	Tier            string
	SkipReason      string
}

// Mover executes one VM move.
type Mover interface {
	Move(vmName string, target string) error
}

// ExecutionSummary tracks plan execution outcomes.
type ExecutionSummary struct {
	MigratedCount int
	FailedCount   int
	DryRunCount   int
}

// Planner encapsulates migration planning and execution.
type Planner struct {
	thresholdPercent int
}

// NewPlanner build a planner with utilization guardrail.
func NewPlanner(thresholdPercent int) Planner {
	return Planner{thresholdPercent: thresholdPercent}
}

// BuildPlan create a migration plan from VM and datastore inputs.
func (p Planner) BuildPlan(vms []VM, candidates []Datastore) []PlanStep {
	state := copyDatastores(candidates)
	plan := make([]PlanStep, 0, len(vms))
	for index, vm := range vms {
		step := p.planStep(index+1, vm, state)
		if step.SkipReason == "" {
			applyProjection(state, step.TargetDatastore, vm.SizeGB)
		}
		plan = append(plan, step)
	}
	return plan
}

func (p Planner) planStep(order int, vm VM, state []Datastore) PlanStep {
	step := PlanStep{Order: order, VMName: vm.Name, SourceDatastore: vm.SourceDatastore}
	targets := targetsForSource(state, vm.SourceDatastore)
	if len(targets) == 0 {
		step.SkipReason = SkipNoEligibleTarget
		step.Tier = "-"
		return step
	}
	sortTargets(targets)
	for _, target := range targets {
		projected := projectedUtil(target, vm.SizeGB)
		if projected <= p.thresholdPercent {
			step.TargetDatastore = target.Name
			step.ProjectedUtil = projected
			step.Tier = target.Tier.String()
			return step
		}
	}
	step.SkipReason = SkipOverThreshold
	step.Tier = "-"
	return step
}

// ExecutePlan run plan steps, honoring dry-run mode and retry count.
func (p Planner) ExecutePlan(plan []PlanStep, execute bool, retries int, mover Mover) ExecutionSummary {
	summary := ExecutionSummary{}
	attemptLimit := retries
	if attemptLimit < 1 {
		attemptLimit = 1
	}
	for _, step := range plan {
		if step.SkipReason != "" {
			continue
		}
		if !execute {
			summary.DryRunCount++
			continue
		}
		if runMove(step, attemptLimit, mover) {
			summary.MigratedCount++
			continue
		}
		summary.FailedCount++
	}
	return summary
}

func runMove(step PlanStep, attempts int, mover Mover) bool {
	for i := 0; i < attempts; i++ {
		if mover.Move(step.VMName, step.TargetDatastore) == nil {
			return true
		}
	}
	return false
}

func copyDatastores(candidates []Datastore) []Datastore {
	copied := make([]Datastore, len(candidates))
	copy(copied, candidates)
	return copied
}

func applyProjection(state []Datastore, target string, size int) {
	for i := range state {
		if state[i].Name == target {
			state[i].UsedGB += size
			return
		}
	}
}

func targetsForSource(state []Datastore, source string) []Datastore {
	targets := make([]Datastore, 0, len(state))
	for _, datastore := range state {
		if datastore.Name != source {
			targets = append(targets, datastore)
		}
	}
	return targets
}

func sortTargets(targets []Datastore) {
	sort.Slice(targets, func(i int, j int) bool {
		utilI := projectedUtil(targets[i], 0)
		utilJ := projectedUtil(targets[j], 0)
		if utilI != utilJ {
			return utilI < utilJ
		}
		if targets[i].Tier != targets[j].Tier {
			return targets[i].Tier < targets[j].Tier
		}
		return targets[i].Name < targets[j].Name
	})
}

func projectedUtil(target Datastore, additionalGB int) int {
	if target.CapacityGB == 0 {
		return 100
	}
	return ((target.UsedGB + additionalGB) * 100) / target.CapacityGB
}
