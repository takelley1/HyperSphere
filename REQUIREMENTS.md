# REQUIREMENTS

## Scope and Source
- Goal: Reimplement the k9s feature surface for vSphere/vCenter workflows.
- Source baseline: Local `k9s/README.md`, `k9s/cmd/root.go`,
  `k9s/plugins/README.md`, and current HyperSphere `DESIGN.md`.
- Long-term anchor from `DESIGN.md`: Achieve parity milestones, maintain a parity
  matrix, and harden with integration/perf quality gates.

## Requirement Format
- Every item below is atomic and intended for a single commit.
- Every item should start with a failing test, then minimal implementation.
- "Acceptance" statements are the exact behavior to assert in tests.

## CLI and Startup Parity

### RQ-001: Add `hypersphere version` command
- vSphere mapping: Same runtime introspection contract as `k9s version`.
- Acceptance: `version` subcommand prints semantic version, commit SHA, and build date fields.
- Status: fulfilled (2026-02-16).

### RQ-002: Add `hypersphere info` command
- vSphere mapping: Runtime path introspection for config/log/plugins/skins files.
- Acceptance: `info` output lists absolute paths for config, logs, dumps, skins,
  plugins, and hotkeys.
- Status: fulfilled (2026-02-15).

### RQ-003: Add `--refresh` startup flag
- vSphere mapping: UI polling interval for vCenter inventory refresh.
- Acceptance: CLI accepts a float seconds value; values below minimum are clamped
  to configured minimum.
- Status: fulfilled (2026-02-15).

### RQ-004: Add `--log-level` startup flag
- vSphere mapping: Operational debugging for vCenter API workflows.
- Acceptance: valid values map to debug/info/warn/error logger levels;
  invalid value returns a parse error.
- Status: fulfilled (2026-02-15).

### RQ-005: Add `--log-file` startup flag
- vSphere mapping: Operator-defined destination for runtime logs.
- Acceptance: log records are written to custom path when flag is set.
- Status: fulfilled (2026-02-15).

### RQ-006: Add `--readonly` startup flag
- vSphere mapping: Block write actions for inventory safety.
- Acceptance: write action invocation in read-only mode returns a deterministic
  `read-only mode` error.
- Status: fulfilled (2026-02-15).

### RQ-007: Add `--write` startup flag override
- vSphere mapping: Explicit write enable even when config defaults read-only.
- Acceptance: when both config `readOnly=true` and `--write` is passed,
  mutating actions are permitted.
- Status: fulfilled (2026-02-15).

### RQ-008: Add `--command` startup flag
- vSphere mapping: Start directly in a resource view (`vm`, `host`, `datastore`, etc.).
- Acceptance: app starts with selected view active and rendered without extra user input.

### RQ-009: Add `--headless` startup flag
- vSphere mapping: Hide top header line for dense terminal sessions.
- Acceptance: header line is not rendered when flag is enabled.

### RQ-010: Add `--crumbsless` startup flag
- vSphere mapping: Hide breadcrumb/navigation trail.
- Acceptance: breadcrumb widget does not render when flag is enabled.

## Prompt and Command-Mode Parity

### RQ-011: Add `?` key to open keymap help modal
- vSphere mapping: discoverability of hotkeys per active vSphere view.
- Acceptance: pressing `?` opens a modal containing active view actions and closes on `Esc`.

### RQ-012: Add `ctrl-a` alias palette
- vSphere mapping: browse all resource aliases/shortcuts.
- Acceptance: palette lists aliases sorted alphabetically and executes selected alias command.

### RQ-013: Add command alias registry file
- vSphere mapping: user-defined aliases for `:vm`, `:host`, `:ds`, etc.
- Acceptance: alias file entries resolve to canonical commands with optional arguments.

### RQ-014: Add command history traversal parity
- vSphere mapping: repeat and edit prior explorer commands.
- Acceptance: `:history up/down` moves cursor through bounded history without skipping entries.

### RQ-015: Add tab-complete acceptance of first suggestion
- vSphere mapping: k9s-like prompt completion flow.
- Acceptance: pressing `Tab` on a non-empty suggestion list accepts index 0 and updates input text.

### RQ-016: Add inline prompt validation state
- vSphere mapping: immediate feedback for invalid view/action/filter syntax.
- Acceptance: invalid command highlights prompt with error message before execution.

### RQ-017: Add `:-` previous-view toggle in prompt mode
- vSphere mapping: k9s back navigation behavior.
- Acceptance: issuing `:-` swaps current and previous resource views.

### RQ-018: Add command `:ctx` analog for vCenter targets
- vSphere mapping: switch active vCenter endpoint profile.
- Acceptance: `:ctx` shows configured endpoints; selecting one reconnects and refreshes active view.
- Status: fulfilled (2026-02-15).

## Table Interaction Parity

### RQ-019: Add sticky table headers
- vSphere mapping: large inventory scan usability.
- Acceptance: vertical scroll keeps header row fixed while body rows change.

### RQ-020: Add selected-column sort glyphs
- vSphere mapping: visible ascending/descending state.
- Acceptance: sorted column header includes up/down glyph and changes when sort direction flips.

### RQ-021: Add autosized columns on terminal resize
- vSphere mapping: maintain readability across narrow/wide terminals.
- Acceptance: terminal resize recalculates column widths without truncating fixed-priority columns.

### RQ-022: Add overflow indicators for hidden columns
- vSphere mapping: horizontal navigation discoverability.
- Acceptance: left/right overflow markers appear only when additional columns exist off-screen.

### RQ-023: Add compact mode for narrow terminals
- vSphere mapping: prioritized columns for VM/host/datastore views.
- Acceptance: when width is below threshold, only compact column set renders per resource.

### RQ-024: Add row focus sync from mouse click
- vSphere mapping: mixed mouse/keyboard parity.
- Acceptance: clicking a row updates internal selection state used by hotkeys and action targets.

### RQ-025: Add mark count badge in header
- vSphere mapping: bulk action confidence.
- Acceptance: header shows exact count of marked objects and updates after mark/unmark/clear.

### RQ-026: Add `ctrl-w` wide-column toggle
- vSphere mapping: switch between standard and extended column sets.
- Acceptance: toggling key changes table schema and preserves selected object identity.

### RQ-027: Add `ctrl-e` header visibility toggle
- vSphere mapping: high-density terminal mode.
- Acceptance: repeated toggle alternates header on/off with no row selection reset.

## Screenshot Visual Baseline Parity

### RQ-092: Add three-zone top header layout
- vSphere mapping: k9s-like information density for operators.
- Acceptance: top bar renders left metadata, center hotkeys, and right ASCII logo
  in fixed zones with no overlap at 120-column width.

### RQ-093: Add left metadata panel with fixed label order
- vSphere mapping: current vCenter operational context visibility.
- Acceptance: left panel shows `Context`, `Cluster`, `User`, `HS Version`,
  `vCenter Version`, `CPU`, `MEM` in that exact order.

### RQ-094: Add center hotkey legend in angle-bracket style
- vSphere mapping: discoverable command hints matching screenshot style.
- Acceptance: hotkey legend renders entries as `<key>` plus action label,
  with one entry per line.

### RQ-095: Add right-side ASCII HyperSphere logo block
- vSphere mapping: branded parity with k9s right-header mark.
- Acceptance: logo renders as multiline ASCII art aligned to the top-right of
  header area and stays clipped within panel bounds.

### RQ-096: Add cyan framed content border style
- vSphere mapping: high-contrast visual grouping of active views.
- Acceptance: active content view is wrapped in a single cyan border frame
  with consistent corner and edge glyphs.

### RQ-097: Add cyan title divider with centered view name
- vSphere mapping: explicit active view identity.
- Acceptance: view title appears centered on the top border line as
  `ViewName(scope)[count]` with cyan divider segments on both sides.

### RQ-098: Add screenshot palette preset
- vSphere mapping: recognizable k9s-inspired look-and-feel.
- Acceptance: preset maps header accents to yellow/cyan/magenta, table header
  background to cyan, and default canvas background to black.

### RQ-099: Add status color mapping for table rows
- vSphere mapping: fast health scan of vSphere objects.
- Acceptance: healthy rows render green emphasis, degraded rows yellow, and
  faulted rows red according to canonical status field mapping.

### RQ-100: Add selected-row inversion style
- vSphere mapping: cursor clarity in dense inventories.
- Acceptance: selected row is visually distinct from non-selected rows via
  deterministic inversion or accent style that does not alter cell text.

### RQ-101: Add top metrics style parity for CPU and MEM
- vSphere mapping: quick capacity posture at glance.
- Acceptance: CPU and MEM values render in the top-left panel with percentage
  formatting and trend suffix support `(+)`/`(-)`.

### RQ-102: Add view-specific hotkey legend switching
- vSphere mapping: context-aware hints for inventory vs log view.
- Acceptance: switching to log view replaces center legend with log-navigation
  keys and restores table keys when returning.

### RQ-103: Add logs view title format parity
- vSphere mapping: identify object and sub-target for streamed logs.
- Acceptance: log frame title renders `Logs <object-path> (target=<value>)`
  when a sub-target is present.

### RQ-104: Add monospaced timestamped log line renderer
- vSphere mapping: operator-readable event/log triage.
- Acceptance: log rows render with timestamp, level marker, and message
  columns in monospaced alignment with wrapped continuation indentation.

### RQ-105: Add viewport scroll controls in log mode
- vSphere mapping: navigate high-volume logs with deterministic shortcuts.
- Acceptance: `Top`, `Bottom`, `PageUp`, and `PageDown` controls move the
  visible log window to exact expected offsets.

### RQ-106: Add compact header degradation behavior
- vSphere mapping: preserve usability on smaller terminals.
- Acceptance: below configured width, center legend collapses first,
  then logo hides, while left metadata and active view remain visible.

## Resource Coverage Parity (vSphere Analogs)

### RQ-028: Add `:dc` datacenter resource view
- vSphere mapping: top-level inventory container.
- Acceptance: view columns include `NAME`, `CLUSTERS`, `HOSTS`, `VMS`, `DATASTORES`.

### RQ-029: Add `:rp` resource pool view
- vSphere mapping: scheduler partitioning visibility.
- Acceptance: view columns include `NAME`, `CLUSTER`, `CPU_RES`, `MEM_RES`, `VM_COUNT`.

### RQ-030: Add `:network` view
- vSphere mapping: distributed switch/portgroup visibility.
- Acceptance: view columns include `NAME`, `TYPE`, `VLAN`, `SWITCH`, `ATTACHED_VMS`.

### RQ-031: Add `:template` view
- vSphere mapping: VM templates lifecycle.
- Acceptance: view columns include `NAME`, `OS`, `DATASTORE`, `FOLDER`, `AGE`.

### RQ-032: Add `:snapshot` view
- vSphere mapping: VM snapshot governance.
- Acceptance: view columns include `VM`, `SNAPSHOT`, `SIZE`, `CREATED`, `AGE`, `QUIESCED`.

### RQ-033: Add `:task` view
- vSphere mapping: vCenter task execution stream.
- Acceptance: view columns include `ENTITY`, `ACTION`, `STATE`, `STARTED`, `DURATION`, `OWNER`.

### RQ-034: Add `:event` view
- vSphere mapping: inventory event chronology.
- Acceptance: view columns include `TIME`, `SEVERITY`, `ENTITY`, `MESSAGE`, `USER`.

### RQ-035: Add `:alarm` view
- vSphere mapping: active alarm tracking.
- Acceptance: view columns include `ENTITY`, `ALARM`, `STATUS`, `TRIGGERED`, `ACKED_BY`.

### RQ-036: Add `:folder` view
- vSphere mapping: inventory hierarchy exploration.
- Acceptance: view columns include `PATH`, `TYPE`, `CHILDREN`, `VM_COUNT`.

### RQ-037: Add `:tag` view
- vSphere mapping: tag/category management parity.
- Acceptance: view columns include `TAG`, `CATEGORY`, `CARDINALITY`, `ATTACHED_OBJECTS`.

## Navigation Depth and Context

### RQ-038: Add breadcrumb chain `dc > cluster > host > vm`
- vSphere mapping: hierarchical navigation parity.
- Acceptance: selecting drill-down updates breadcrumb trail and back-navigation
  restores parent view.

### RQ-039: Add `Shift+J` owner jump analog
- vSphere mapping: jump from child object to parent owner.
- Acceptance: from VM row, owner jump moves focus to host or resource pool owning that VM.

### RQ-040: Add namespace warp analog for tags/folders
- vSphere mapping: fast scope jump by common partition key.
- Acceptance: warp key on selected folder/tag opens filtered view scoped to selected key.

### RQ-041: Add search result jump list
- vSphere mapping: next/previous navigation through matched rows.
- Acceptance: `n`/`N` hotkeys cycle through filtered match indices with wrap-around.

### RQ-042: Add split-pane details drawer
- vSphere mapping: selected object metadata inspection.
- Acceptance: toggling details pane shows key-value metadata for current row
  without leaving table view.

## Filter and Query Parity

### RQ-043: Add regex filter mode
- vSphere mapping: advanced row filtering.
- Acceptance: `/pattern` applies regex filter; invalid regex returns parse error
  and keeps prior filter.

### RQ-044: Add inverse filter mode
- vSphere mapping: exclude noisy objects.
- Acceptance: `/!pattern` excludes rows matching pattern.

### RQ-045: Add label filter analog for tags
- vSphere mapping: filter by `key=value` style vSphere tags.
- Acceptance: `/-t env=prod,tier=gold` returns rows with all requested tag pairs.

### RQ-046: Add fuzzy filter mode
- vSphere mapping: tolerant search in large inventories.
- Acceptance: `/-f text` ranks results by fuzzy score and preserves stable tie ordering.

## Action Execution Parity

### RQ-047: Add async task queue model for actions
- vSphere mapping: long-running vCenter operations.
- Acceptance: action transitions through `queued -> running -> success|failure` with timestamps.

### RQ-048: Add per-action cancellation support
- vSphere mapping: cancel migrates/snapshot tasks where API supports cancel.
- Acceptance: cancel request updates task state to `cancelled` when backend supports it.

### RQ-049: Add per-action timeout policy
- vSphere mapping: avoid hung operations.
- Acceptance: timeout breach marks task as failed with explicit timeout reason.

### RQ-050: Add retry policy wiring
- vSphere mapping: transient vCenter/API failures.
- Acceptance: retriable error retries up to policy limit; non-retriable error fails immediately.

### RQ-051: Add destructive-action confirmation dialogs
- vSphere mapping: power-off/delete/snapshot-remove safeguards.
- Acceptance: destructive key path requires explicit confirm; deny path leaves object unchanged.

### RQ-052: Add action preview summary panel
- vSphere mapping: bulk operation impact preview.
- Acceptance: preview lists target count, target IDs, and expected side effects before execution.

### RQ-053: Add consistent action error presentation
- vSphere mapping: operator triage consistency.
- Acceptance: all failed actions render standardized error structure:
  code, message, entity, retryable.

### RQ-054: Add audit summary for completed bulk actions
- vSphere mapping: post-action accountability.
- Acceptance: summary records actor, timestamp, action, targets, outcomes, and failed IDs.

## vSphere-Specific Action Mappings

### RQ-055: Add VM power lifecycle actions
- vSphere mapping: `power_on`, `power_off`, `reset`, `suspend`.
- Acceptance: each action appears in VM action list and routes to distinct backend method.

### RQ-056: Add VM migrate action with placement target
- vSphere mapping: live/cold migration.
- Acceptance: migrate action requires target host or datastore and validates target existence.

### RQ-057: Add snapshot create/remove/revert actions
- vSphere mapping: snapshot lifecycle parity.
- Acceptance: action requires snapshot name for create and existing snapshot ID for remove/revert.

### RQ-058: Add host maintenance mode enter/exit actions
- vSphere mapping: host lifecycle operations.
- Acceptance: host view includes actions and state transitions reflect maintenance mode changes.

### RQ-059: Add datastore maintenance and evacuation action
- vSphere mapping: storage maintenance workflows.
- Acceptance: datastore action requires confirmation and reports migrated VM count.

### RQ-060: Add tag assign/unassign actions
- vSphere mapping: metadata management parity.
- Acceptance: bulk assign/unassign operates on marked objects and reports per-object failures.

## Observability and Health Parity

### RQ-061: Add Pulses analog dashboard
- vSphere mapping: live cluster/host/VM utilization summary.
- Acceptance: dashboard renders CPU, memory, datastore usage, and active alarms with refresh timer.

### RQ-062: Add XRay analog dependency graph
- vSphere mapping: graph VM -> host -> datastore -> network dependencies.
- Acceptance: graph view renders selected entity path and supports expanding one level at a time.

### RQ-063: Add fault/error toggle hotkey
- vSphere mapping: show only objects with alarms or degraded state.
- Acceptance: toggle filters table to faulted rows and restores full set on second toggle.

### RQ-064: Add benchmark/perf panel for heavy operations
- vSphere mapping: action duration and throughput metrics.
- Acceptance: panel shows p50/p95 duration and success rate for selected action type.

### RQ-065: Add event watch mode
- vSphere mapping: continuous vCenter event tailing.
- Acceptance: event view auto-appends new events in timestamp order during watch mode.

## Configuration and Customization Parity

### RQ-066: Add XDG-style config discovery with env override
- vSphere mapping: predictable config locations across OSes.
- Acceptance: `HYPERSPHERE_CONFIG_DIR` overrides default config root path.

### RQ-067: Add config schema validation for main config file
- vSphere mapping: fail-fast misconfiguration handling.
- Acceptance: invalid config fields produce schema error with failing field path.

### RQ-068: Add hotkeys config file support
- vSphere mapping: user key remapping parity.
- Acceptance: custom hotkey binding overrides default action binding at runtime.

### RQ-069: Add aliases config file support
- vSphere mapping: reusable command shortcuts.
- Acceptance: aliases file loads at startup and resolves multi-token alias arguments.

### RQ-070: Add plugins config file support
- vSphere mapping: external command integration.
- Acceptance: plugin entries validate against schema before being activated.

### RQ-071: Add skins/themes config support
- vSphere mapping: terminal visual customization.
- Acceptance: selected skin file changes color palette for header/body/status regions.

### RQ-072: Add `NO_COLOR` and ASCII symbol compatibility mode
- vSphere mapping: terminal compatibility.
- Acceptance: when enabled, unicode glyphs are replaced with ASCII symbols and color is disabled.

### RQ-073: Add context-specific config overlays per vCenter
- vSphere mapping: per-endpoint preferences and overrides.
- Acceptance: active endpoint loads endpoint-specific aliases/plugins/hotkeys
  before global defaults.

## Plugin and Extensibility Parity

### RQ-074: Add plugin command runner with environment contract
- vSphere mapping: run external scripts/tools against selected objects.
- Acceptance: plugin receives selected object IDs and active endpoint via environment variables.

### RQ-075: Add plugin scope binding per view
- vSphere mapping: only show plugin where resource type is valid.
- Acceptance: plugin is visible only in listed scopes and hidden elsewhere.

### RQ-076: Add plugin shortcut binding and collision detection
- vSphere mapping: predictable key behavior.
- Acceptance: startup fails with explicit error when plugin shortcut conflicts with core hotkey.

### RQ-077: Add plugin permission prompts for write-capable plugins
- vSphere mapping: safe execution of mutating external actions.
- Acceptance: first execution prompts for confirmation; denied plugin does not execute.

### RQ-078: Add plugin lifecycle telemetry hooks
- vSphere mapping: execution visibility for plugin actions.
- Acceptance: plugin start/end/failure events are recorded in task/audit stream.

## Read-Only, Safety, and Security

### RQ-079: Enforce read-only gating across all mutating paths
- vSphere mapping: safety lock parity.
- Acceptance: every mutating action from hotkey, prompt command, and plugin
  path returns read-only error.

### RQ-080: Add credential redaction in logs and status lines
- vSphere mapping: protect vCenter secrets/tokens.
- Acceptance: logs never emit plaintext password, token, or session cookie substrings.

### RQ-081: Add connection-health indicator in header
- vSphere mapping: stale/disconnected backend visibility.
- Acceptance: header shows `CONNECTED`, `DEGRADED`, or `DISCONNECTED`
  based on adapter health checks.

### RQ-082: Add stale-data indicator for cached views
- vSphere mapping: operator confidence in watch freshness.
- Acceptance: view shows stale badge when data age exceeds configured threshold.

## Performance and Scale

### RQ-083: Add large-inventory render benchmark suite
- vSphere mapping: parity hardening for enterprise-scale inventories.
- Acceptance: benchmark covers 10k/50k/100k row scenarios and records render latency.

### RQ-084: Add stable row identity under refresh
- vSphere mapping: prevent selection jumps during polling.
- Acceptance: selected entity remains selected after refresh if entity still exists.

### RQ-085: Add incremental table diff updates
- vSphere mapping: avoid full redraw churn at scale.
- Acceptance: refresh updates only changed rows/cells and preserves scroll position.

### RQ-086: Add sort performance guardrails
- vSphere mapping: predictable interaction latency.
- Acceptance: sorting 50k rows completes under configured latency threshold in benchmark tests.

## Integration and Quality Gates

### RQ-087: Add fake govmomi-backed integration test harness
- vSphere mapping: deterministic end-to-end tests without real vCenter.
- Acceptance: harness can seed inventory and drive view/action flows in integration tests.

### RQ-088: Add integration tests for mark semantics and bulk actions
- vSphere mapping: trust in multi-target operations.
- Acceptance: tests verify mark range, mark clear, and bulk action target resolution.

### RQ-089: Add integration tests for retry/failure/read-only flows
- vSphere mapping: robust failure-path correctness.
- Acceptance: tests verify retriable vs non-retriable behavior and read-only blocking.

### RQ-090: Add parity matrix document generated from requirements
- vSphere mapping: release readiness tracking against k9s feature families.
- Acceptance: matrix lists each requirement ID with status `not-started|in-progress|done|validated`.

### RQ-091: Add release acceptance checklist for parity milestones
- vSphere mapping: objective release gates.
- Acceptance: checklist includes required pass criteria for tests, benchmarks,
  docs, and parity matrix.

## Suggested Delivery Order
- Phase 1: RQ-001 to RQ-027 (CLI + command/table interaction baseline).
- Phase 2: RQ-092 to RQ-106 (screenshot visual baseline parity).
- Phase 3: RQ-028 to RQ-046 (resource/view and query expansion).
- Phase 4: RQ-047 to RQ-065 (action pipeline + observability).
- Phase 5: RQ-066 to RQ-078 (config/customization/plugins).
- Phase 6: RQ-079 to RQ-091 (security/perf/integration/release hardening).
