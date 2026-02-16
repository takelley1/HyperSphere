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
- Status: fulfilled (2026-02-16).

### RQ-009: Add `--headless` startup flag
- vSphere mapping: Hide top header line for dense terminal sessions.
- Acceptance: header line is not rendered when flag is enabled.
- Status: fulfilled (2026-02-16).

### RQ-010: Add `--crumbsless` startup flag
- vSphere mapping: Hide breadcrumb/navigation trail.
- Acceptance: breadcrumb widget does not render when flag is enabled.
- Status: fulfilled (2026-02-15).

## Prompt and Command-Mode Parity

### RQ-011: Add `?` key to open keymap help modal
- vSphere mapping: discoverability of hotkeys per active vSphere view.
- Acceptance: pressing `?` opens a modal containing active view actions and closes on `Esc`.
- Status: fulfilled (2026-02-16).

### RQ-012: Add `ctrl-a` alias palette
- vSphere mapping: browse all resource aliases/shortcuts.
- Acceptance: palette lists aliases sorted alphabetically and executes selected alias command.
- Status: fulfilled (2026-02-15).

### RQ-013: Add command alias registry file
- vSphere mapping: user-defined aliases for `:vm`, `:host`, `:ds`, etc.
- Acceptance: alias file entries resolve to canonical commands with optional arguments.
- Status: fulfilled (2026-02-16).

### RQ-014: Add command history traversal parity
- vSphere mapping: repeat and edit prior explorer commands.
- Acceptance: `:history up/down` moves cursor through bounded history without skipping entries.
- Status: fulfilled (2026-02-16).

### RQ-015: Add tab-complete acceptance of first suggestion
- vSphere mapping: k9s-like prompt completion flow.
- Acceptance: pressing `Tab` on a non-empty suggestion list accepts index 0 and updates input text.
- Status: fulfilled (2026-02-15).

### RQ-016: Add inline prompt validation state
- vSphere mapping: immediate feedback for invalid view/action/filter syntax.
- Acceptance: invalid command highlights prompt with error message before execution.
- Status: fulfilled (2026-02-15).

### RQ-017: Add `:-` previous-view toggle in prompt mode
- vSphere mapping: k9s back navigation behavior.
- Acceptance: issuing `:-` swaps current and previous resource views.
- Status: fulfilled (2026-02-15).

### RQ-018: Add command `:ctx` analog for vCenter targets
- vSphere mapping: switch active vCenter endpoint profile.
- Acceptance: `:ctx` shows configured endpoints; selecting one reconnects and refreshes active view.
- Status: fulfilled (2026-02-15).

## Table Interaction Parity

### RQ-019: Add sticky table headers
- vSphere mapping: large inventory scan usability.
- Acceptance: vertical scroll keeps header row fixed while body rows change.
- Status: fulfilled (2026-02-15).

### RQ-020: Add selected-column sort glyphs
- vSphere mapping: visible ascending/descending state.
- Acceptance: sorted column header includes up/down glyph and changes when sort direction flips.
- Status: fulfilled (2026-02-15).

### RQ-021: Add autosized columns on terminal resize
- vSphere mapping: maintain readability across narrow/wide terminals.
- Acceptance: terminal resize recalculates column widths without truncating fixed-priority columns.
- Status: fulfilled (2026-02-15).

### RQ-022: Add overflow indicators for hidden columns
- vSphere mapping: horizontal navigation discoverability.
- Acceptance: left/right overflow markers appear only when additional columns exist off-screen.
- Status: fulfilled (2026-02-16).

### RQ-023: Add compact mode for narrow terminals
- vSphere mapping: prioritized columns for VM/host/datastore views.
- Acceptance: when width is below threshold, only compact column set renders per resource.
- Status: fulfilled (2026-02-16).

### RQ-024: Add row focus sync from mouse click
- vSphere mapping: mixed mouse/keyboard parity.
- Acceptance: clicking a row updates internal selection state used by hotkeys and action targets.
- Status: fulfilled (2026-02-16).

### RQ-025: Add mark count badge in header
- vSphere mapping: bulk action confidence.
- Acceptance: header shows exact count of marked objects and updates after mark/unmark/clear.
- Status: fulfilled (2026-02-16).

### RQ-026: Add `ctrl-w` wide-column toggle
- vSphere mapping: switch between standard and extended column sets.
- Acceptance: toggling key changes table schema and preserves selected object identity.
- Status: fulfilled (2026-02-16).

### RQ-027: Add `ctrl-e` header visibility toggle
- vSphere mapping: high-density terminal mode.
- Acceptance: repeated toggle alternates header on/off with no row selection reset.
- Status: fulfilled (2026-02-16).

### RQ-107: Add `d` hotkey to describe selected resource
- vSphere mapping: k9s-style describe flow for deep object inspection.
- Acceptance: pressing `d` opens a details panel for the selected row using
  resource-specific fields.
- Acceptance: for VM rows, details include at minimum `NAME`, `POWER_STATE`,
  `CPU_COUNT`, `MEMORY_MB`, `COMMENTS`, `DESCRIPTION`, `SNAPSHOT_COUNT`, and
  snapshot identifiers/timestamps when present.
- Acceptance: details panel closes on `Esc` and returns focus to the same table
  row/column without losing marks.
- Status: fulfilled (2026-02-16).

## Screenshot Visual Baseline Parity

### RQ-092: Add three-zone top header layout
- vSphere mapping: k9s-like information density for operators.
- Acceptance: top bar renders left metadata, center hotkeys, and right ASCII logo
  in fixed zones with no overlap at 120-column width.
- Status: fulfilled (2026-02-16).

### RQ-093: Add left metadata panel with fixed label order
- vSphere mapping: current vCenter operational context visibility.
- Acceptance: left panel shows `Context`, `Cluster`, `User`, `HS Version`,
  `vCenter Version`, `CPU`, `MEM` in that exact order.
- Status: fulfilled (2026-02-16).

### RQ-094: Add center hotkey legend in angle-bracket style
- vSphere mapping: discoverable command hints matching screenshot style.
- Acceptance: hotkey legend renders entries as `<key>` plus action label,
  with one entry per line.
- Status: fulfilled (2026-02-16).

### RQ-095: Add right-side ASCII HyperSphere logo block
- vSphere mapping: branded parity with k9s right-header mark.
- Acceptance: logo renders as multiline ASCII art aligned to the top-right of
  header area and stays clipped within panel bounds.
- Status: fulfilled (2026-02-16).

### RQ-113: Replace top-right ASCII mark with a 4D hypersphere logo design
- vSphere mapping: visual identity should represent HyperSphere semantics rather
  than plain project text.
- Acceptance: top-right header logo renders as a multiline ASCII projection of
  a hypersphere (4D sphere), not text lettering.
- Acceptance: logo remains right-aligned, clipped to the right header zone, and
  stable across redraw/resize flows.
- Status: fulfilled (2026-02-16).

### RQ-114: Expand built-in sample inventory data for browsing
- vSphere mapping: operators need enough local sample data to browse table
  behaviors and interactions without connecting to live endpoints.
- Acceptance: default sample catalog includes at least 8 rows each for VM, LUN,
  host, and datastore resources, and at least 4 cluster rows.
- Acceptance: sample data includes mixed operational states (for example powered
  off VMs and disconnected hosts) to exercise browsing and status cues.
- Status: fulfilled (2026-02-16).

### RQ-115: Remove bottom help bar and move help hints to top header
- vSphere mapping: keep operator guidance visible without consuming bottom-panel
  vertical space needed for status and browsing.
- Acceptance: bottom "Help" bar widget is removed from the explorer layout.
- Acceptance: relevant help hints (command/filter/help/action, completion,
  movement, prompt state, quit key) render in the top-center header region.
- Acceptance: moved help hints render with the existing cyan center accent in
  color mode and remain clock-free for event-driven redraw behavior.
- Status: fulfilled (2026-02-16).

### RQ-116: Add additional useful columns across all resource views
- vSphere mapping: improve high-density operator visibility by using horizontal
  terminal space for decision-making fields instead of sparse schemas.
- Acceptance: each resource view adds at least one new operationally useful
  column beyond its baseline schema.
- Acceptance: `:vm` includes `CPU_COUNT`, `MEMORY_MB`, `SNAPSHOTS`; `:lun`
  includes `FREE_GB`, `UTIL_PERCENT`; `:cluster` includes `RESOURCE_POOLS`,
  `NETWORKS`; `:dc` includes `CPU_PERCENT`, `MEM_PERCENT`; `:rp` includes
  `CPU_LIMIT`, `MEM_LIMIT`; `:nw` includes `MTU`, `UPLINKS`; `:tp` includes
  `CPU_COUNT`, `MEMORY_MB`; `:ss` includes `OWNER`; `:host` includes `CORES`,
  `THREADS`, `VMS`; `:datastore` includes `TYPE`, `LATENCY_MS`.
- Acceptance: built-in sample dataset provides representative values for the new
  columns so local browsing uses the wider schemas immediately.
- Status: fulfilled (2026-02-16).

### RQ-117: Highlight full selected resource row across terminal width
- vSphere mapping: improve scanability so active focus is obvious in dense
  operational tables.
- Acceptance: selected row highlight applies to all rendered cells in the row,
  not only text glyphs/characters.
- Acceptance: selected row highlight includes trailing fill so highlight visually
  spans the full table width in the terminal viewport.
- Status: fulfilled (2026-02-16).

### RQ-118: Use full-row color highlight for marked table selections
- vSphere mapping: multi-select state should be obvious for bulk actions without
  relying on marker glyphs alone.
- Acceptance: marking a row changes the background color of the entire row.
- Acceptance: marked+selected rows use a distinct combined background color.
- Acceptance: marker glyph may remain, but row color is the primary state cue.
- Status: fulfilled (2026-02-16).

### RQ-119: Add sort instructions to top help text
- vSphere mapping: improve discoverability of table ordering controls during
  terminal-first operations.
- Acceptance: top-center help legend in table mode explicitly includes sort
  guidance (`Shift+O`) so operators can discover sorting without opening docs.
- Status: fulfilled (2026-02-16).

### RQ-120: Move breadcrumbs and status panels to top of TUI layout
- vSphere mapping: keep context and feedback visible without requiring bottom
  scanning while navigating resource tables.
- Acceptance: breadcrumbs and status widgets render above the table body.
- Acceptance: prompt remains at the bottom input line; breadcrumbs/status no
  longer occupy bottom layout slots.
- Status: fulfilled (2026-02-16).

### RQ-121: Render breadcrumbs/status as compact top-right path hints
- vSphere mapping: preserve context/status visibility while reducing visual
  intrusion from dedicated bordered panels.
- Acceptance: standalone breadcrumb/status panels are removed from the vertical
  layout; only top header, table, and prompt remain.
- Acceptance: top-right header near the logo shows compact `path:` and
  `status:` lines.
- Acceptance: breadcrumbs-less mode suppresses the compact path line.
- Status: fulfilled (2026-02-16).

### RQ-122: Render top help legend in multi-column layout
- vSphere mapping: improve scan speed for hotkeys by reducing vertical scanning
  and using center header width more efficiently.
- Acceptance: table-mode top help hints render in two or three columns per row
  (not single-item rows).
- Acceptance: log-mode hotkey hints also render in multi-column rows.
- Acceptance: center legend remains cyan-accented in color mode.
- Status: fulfilled (2026-02-16).

### RQ-123: Expand VM view with runtime capacity and identity columns
- vSphere mapping: VM operations require immediate CPU/memory/storage usage and
  placement visibility without opening details panes.
- Acceptance: VM table includes columns for used CPU, used memory, used storage,
  IP address, DNS name, cluster, host, network, total CPU cores, total RAM,
  largest hard disk size, and attached storage target.
- Acceptance: built-in VM sample rows populate these columns with representative
  values for local browsing.
- Status: fulfilled (2026-02-16).

### RQ-124: Add VM snapshot count and total snapshot size columns
- vSphere mapping: snapshot governance needs immediate per-VM visibility in the
  primary VM table without opening detail panels.
- Acceptance: VM view includes `SNAPSHOT_COUNT` and `SNAPSHOT_TOTAL_GB` columns.
- Acceptance: built-in VM sample rows populate both fields.
- Status: fulfilled (2026-02-16).

### RQ-125: Add per-view column selection controls
- vSphere mapping: operators need to tailor dense tables to the fields most
  relevant for current incident/debug workflows.
- Acceptance: each resource view supports choosing visible columns from its full
  schema.
- Acceptance: selected column sets persist per view while switching between
  views during a session.
- Acceptance: hidden columns can be restored without restarting the explorer.
- Status: fulfilled (2026-02-16).

### RQ-126: Change selected-row highlight to yellow
- vSphere mapping: selected-row visibility should be high-contrast without
  muddy dark-green tones during active navigation and marking.
- Acceptance: selected-row highlight color is yellow in color mode.
- Status: fulfilled (2026-02-16).

### RQ-127: Differentiate hover row color from marked row color
- vSphere mapping: cursor focus and explicit multi-select marks should remain
  visually distinct to avoid ambiguous table state.
- Acceptance: hovered/selected row highlight color differs from marked-row
  highlight color.
- Status: fulfilled (2026-02-16).

### RQ-128: Render path/status hints in center header below shortcuts
- vSphere mapping: keep context/status discoverable in the same glance path as
  keyboard guidance while preserving right-logo visual identity.
- Acceptance: compact `path:` and `status:` hints render in the center header.
- Acceptance: path/status lines appear below shortcut rows.
- Acceptance: right header zone remains dedicated to logo rendering.
- Status: fulfilled (2026-02-16).

### RQ-096: Add cyan framed content border style
- vSphere mapping: high-contrast visual grouping of active views.
- Acceptance: active content view is wrapped in a single cyan border frame
  with consistent corner and edge glyphs.
- Status: fulfilled (2026-02-16).

### RQ-097: Add cyan title divider with centered view name
- vSphere mapping: explicit active view identity.
- Acceptance: view title appears centered on the top border line as
  `ViewName(scope)[count]` with cyan divider segments on both sides.
- Status: fulfilled (2026-02-16).

### RQ-098: Add screenshot palette preset
- vSphere mapping: recognizable k9s-inspired look-and-feel.
- Acceptance: preset maps header accents to yellow/cyan/magenta, table header
  background to cyan, and default canvas background to black.
- Status: fulfilled (2026-02-16).

### RQ-099: Add status color mapping for table rows
- vSphere mapping: fast health scan of vSphere objects.
- Acceptance: healthy rows render green emphasis, degraded rows yellow, and
  faulted rows red according to canonical status field mapping.
- Status: fulfilled (2026-02-16).

### RQ-100: Add selected-row inversion style
- vSphere mapping: cursor clarity in dense inventories.
- Acceptance: selected row is visually distinct from non-selected rows via
  deterministic inversion or accent style that does not alter cell text.
- Status: fulfilled (2026-02-16).

### RQ-101: Add top metrics style parity for CPU and MEM
- vSphere mapping: quick capacity posture at glance.
- Acceptance: CPU and MEM values render in the top-left panel with percentage
  formatting and trend suffix support `(+)`/`(-)`.
- Status: fulfilled (2026-02-16).

### RQ-102: Add view-specific hotkey legend switching
- vSphere mapping: context-aware hints for inventory vs log view.
- Acceptance: switching to log view replaces center legend with log-navigation
  keys and restores table keys when returning.
- Status: fulfilled (2026-02-16).

### RQ-103: Add logs view title format parity
- vSphere mapping: identify object and sub-target for streamed logs.
- Acceptance: log frame title renders `Logs <object-path> (target=<value>)`
  when a sub-target is present.
- Status: fulfilled (2026-02-16).

### RQ-104: Add monospaced timestamped log line renderer
- vSphere mapping: operator-readable event/log triage.
- Acceptance: log rows render with timestamp, level marker, and message
  columns in monospaced alignment with wrapped continuation indentation.
- Status: fulfilled (2026-02-16).

### RQ-105: Add viewport scroll controls in log mode
- vSphere mapping: navigate high-volume logs with deterministic shortcuts.
- Acceptance: `Top`, `Bottom`, `PageUp`, and `PageDown` controls move the
  visible log window to exact expected offsets.
- Status: fulfilled (2026-02-16).

### RQ-106: Add compact header degradation behavior
- vSphere mapping: preserve usability on smaller terminals.
- Acceptance: below configured width, center legend collapses first,
  then logo hides, while left metadata and active view remain visible.
- Status: fulfilled (2026-02-16).

## Resource Coverage Parity (vSphere Analogs)

### RQ-028: Add `:dc`/`:datacenter` datacenter resource view
- vSphere mapping: top-level inventory container.
- Acceptance: view columns include `NAME`, `CLUSTERS`, `HOSTS`, `VMS`, `DATASTORES`.
- Status: fulfilled (2026-02-16).

### RQ-029: Add `:rp`/`:resourcepool` resource pool view
- vSphere mapping: scheduler partitioning visibility.
- Acceptance: view columns include `NAME`, `CLUSTER`, `CPU_RES`, `MEM_RES`, `VM_COUNT`.
- Status: fulfilled (2026-02-16).

### RQ-030: Add `:network`/`:nw` view
- vSphere mapping: distributed switch/portgroup visibility.
- Acceptance: view columns include `NAME`, `TYPE`, `VLAN`, `SWITCH`, `ATTACHED_VMS`.
- Status: fulfilled (2026-02-16).

### RQ-031: Add `:template`/`:tp` view
- vSphere mapping: VM templates lifecycle.
- Acceptance: view columns include `NAME`, `OS`, `DATASTORE`, `FOLDER`, `AGE`.
- Status: fulfilled (2026-02-16).

### RQ-032: Add `:snapshot`/`:snap`/`:ss` view
- vSphere mapping: VM snapshot governance.
- Acceptance: view columns include `VM`, `SNAPSHOT`, `SIZE`, `CREATED`, `AGE`, `QUIESCED`.
- Status: fulfilled (2026-02-16).

### RQ-033: Add `:task` view
- vSphere mapping: vCenter task execution stream.
- Acceptance: view columns include `ENTITY`, `ACTION`, `STATE`, `STARTED`, `DURATION`, `OWNER`.
- Status: fulfilled (2026-02-16).

### RQ-034: Add `:event` view
- vSphere mapping: inventory event chronology.
- Acceptance: view columns include `TIME`, `SEVERITY`, `ENTITY`, `MESSAGE`, `USER`.
- Status: fulfilled (2026-02-16).

### RQ-035: Add `:alarm` view
- vSphere mapping: active alarm tracking.
- Acceptance: view columns include `ENTITY`, `ALARM`, `STATUS`, `TRIGGERED`, `ACKED_BY`.
- Status: fulfilled (2026-02-16).

### RQ-036: Add `:folder` view
- vSphere mapping: inventory hierarchy exploration.
- Acceptance: view columns include `PATH`, `TYPE`, `CHILDREN`, `VM_COUNT`.
- Status: fulfilled (2026-02-16).

### RQ-037: Add `:tag` view
- vSphere mapping: tag/category management parity.
- Acceptance: view columns include `TAG`, `CATEGORY`, `CARDINALITY`, `ATTACHED_OBJECTS`.
- Status: fulfilled (2026-02-16).

## Navigation Depth and Context

### RQ-038: Add breadcrumb chain `dc > cluster > host > vm`
- vSphere mapping: hierarchical navigation parity.
- Acceptance: selecting drill-down updates breadcrumb trail and back-navigation
  restores parent view.
- Status: fulfilled (2026-02-16).

### RQ-039: Add `Shift+J` owner jump analog
- vSphere mapping: jump from child object to parent owner.
- Acceptance: from VM row, owner jump moves focus to host or resource pool owning that VM.
- Status: fulfilled (2026-02-16).

### RQ-040: Add namespace warp analog for tags/folders
- vSphere mapping: fast scope jump by common partition key.
- Acceptance: warp key on selected folder/tag opens filtered view scoped to selected key.
- Status: fulfilled (2026-02-16).

### RQ-041: Add search result jump list
- vSphere mapping: next/previous navigation through matched rows.
- Acceptance: `n`/`N` hotkeys cycle through filtered match indices with wrap-around.
- Status: fulfilled (2026-02-16).

### RQ-042: Add split-pane details drawer
- vSphere mapping: selected object metadata inspection.
- Acceptance: toggling details pane shows key-value metadata for current row
  without leaving table view.
- Status: fulfilled (2026-02-16).

## Filter and Query Parity

### RQ-043: Add regex filter mode
- vSphere mapping: advanced row filtering.
- Acceptance: `/pattern` applies regex filter; invalid regex returns parse error
  and keeps prior filter.
- Status: fulfilled (2026-02-16).

### RQ-044: Add inverse filter mode
- vSphere mapping: exclude noisy objects.
- Acceptance: `/!pattern` excludes rows matching pattern.
- Status: fulfilled (2026-02-16).

### RQ-045: Add label filter analog for tags
- vSphere mapping: filter by `key=value` style vSphere tags.
- Acceptance: `/-t env=prod,tier=gold` returns rows with all requested tag pairs.
- Status: fulfilled (2026-02-16).

### RQ-046: Add fuzzy filter mode
- vSphere mapping: tolerant search in large inventories.
- Acceptance: `/-f text` ranks results by fuzzy score and preserves stable tie ordering.
- Status: fulfilled (2026-02-16).

## Action Execution Parity

### RQ-047: Add async task queue model for actions
- vSphere mapping: long-running vCenter operations.
- Acceptance: action transitions through `queued -> running -> success|failure` with timestamps.
- Status: fulfilled (2026-02-16).

### RQ-048: Add per-action cancellation support
- vSphere mapping: cancel migrates/snapshot tasks where API supports cancel.
- Acceptance: cancel request updates task state to `cancelled` when backend supports it.
- Status: fulfilled (2026-02-16).

### RQ-049: Add per-action timeout policy
- vSphere mapping: avoid hung operations.
- Acceptance: timeout breach marks task as failed with explicit timeout reason.
- Status: fulfilled (2026-02-16).

### RQ-050: Add retry policy wiring
- vSphere mapping: transient vCenter/API failures.
- Acceptance: retriable error retries up to policy limit; non-retriable error fails immediately.
- Status: fulfilled (2026-02-16).

### RQ-051: Add destructive-action confirmation dialogs
- vSphere mapping: power-off/delete/snapshot-remove safeguards.
- Acceptance: destructive key path requires explicit confirm; deny path leaves object unchanged.
- Status: fulfilled (2026-02-16).

### RQ-052: Add action preview summary panel
- vSphere mapping: bulk operation impact preview.
- Acceptance: preview lists target count, target IDs, and expected side effects before execution.
- Status: fulfilled (2026-02-16).

### RQ-053: Add consistent action error presentation
- vSphere mapping: operator triage consistency.
- Acceptance: all failed actions render standardized error structure:
  code, message, entity, retryable.
- Status: fulfilled (2026-02-16).

### RQ-054: Add audit summary for completed bulk actions
- vSphere mapping: post-action accountability.
- Acceptance: summary records actor, timestamp, action, targets, outcomes, and failed IDs.
- Status: fulfilled (2026-02-16).

## vSphere-Specific Action Mappings

### RQ-055: Add VM power lifecycle actions
- vSphere mapping: `power_on`, `power_off`, `reset`, `suspend`.
- Acceptance: each action appears in VM action list and routes to distinct backend method.
- Status: fulfilled (2026-02-16).

### RQ-056: Add VM migrate action with placement target
- vSphere mapping: live/cold migration.
- Acceptance: migrate action requires target host or datastore and validates target existence.
- Status: fulfilled (2026-02-16).

### RQ-057: Add snapshot create/remove/revert actions
- vSphere mapping: snapshot lifecycle parity.
- Acceptance: action requires snapshot name for create and existing snapshot ID for remove/revert.
- Status: fulfilled (2026-02-16).

### RQ-058: Add host maintenance mode enter/exit actions
- vSphere mapping: host lifecycle operations.
- Acceptance: host view includes actions and state transitions reflect maintenance mode changes.
- Status: fulfilled (2026-02-16).

### RQ-059: Add datastore maintenance and evacuation action
- vSphere mapping: storage maintenance workflows.
- Acceptance: datastore action requires confirmation and reports migrated VM count.
- Status: fulfilled (2026-02-16).

### RQ-060: Add tag assign/unassign actions
- vSphere mapping: metadata management parity.
- Acceptance: bulk assign/unassign operates on marked objects and reports per-object failures.
- Status: fulfilled (2026-02-16).

## Observability and Health Parity

### RQ-061: Add Pulses analog dashboard
- vSphere mapping: live cluster/host/VM utilization summary.
- Acceptance: dashboard renders CPU, memory, datastore usage, and active alarms with refresh timer.
- Status: fulfilled (2026-02-16).

### RQ-062: Add XRay analog dependency graph
- vSphere mapping: graph VM -> host -> datastore -> network dependencies.
- Acceptance: graph view renders selected entity path and supports expanding one level at a time.
- Status: fulfilled (2026-02-16).

### RQ-063: Add fault/error toggle hotkey
- vSphere mapping: show only objects with alarms or degraded state.
- Acceptance: toggle filters table to faulted rows and restores full set on second toggle.
- Status: fulfilled (2026-02-16).

### RQ-064: Add benchmark/perf panel for heavy operations
- vSphere mapping: action duration and throughput metrics.
- Acceptance: panel shows p50/p95 duration and success rate for selected action type.
- Status: fulfilled (2026-02-16).

### RQ-065: Add event watch mode
- vSphere mapping: continuous vCenter event tailing.
- Acceptance: event view auto-appends new events in timestamp order during watch mode.
- Status: fulfilled (2026-02-16).

## Configuration and Customization Parity

### RQ-066: Add XDG-style config discovery with env override
- vSphere mapping: predictable config locations across OSes.
- Acceptance: `HYPERSPHERE_CONFIG_DIR` overrides default config root path.
- Status: fulfilled (2026-02-16).

### RQ-067: Add config schema validation for main config file
- vSphere mapping: fail-fast misconfiguration handling.
- Acceptance: invalid config fields produce schema error with failing field path.
- Status: fulfilled (2026-02-16).

### RQ-068: Add hotkeys config file support
- vSphere mapping: user key remapping parity.
- Acceptance: custom hotkey binding overrides default action binding at runtime.
- Status: fulfilled (2026-02-16).

### RQ-069: Add aliases config file support
- vSphere mapping: reusable command shortcuts.
- Acceptance: aliases file loads at startup and resolves multi-token alias arguments.
- Status: fulfilled (2026-02-16).

### RQ-070: Add plugins config file support
- vSphere mapping: external command integration.
- Acceptance: plugin entries validate against schema before being activated.
- Status: fulfilled (2026-02-16).

### RQ-071: Add skins/themes config support
- vSphere mapping: terminal visual customization.
- Acceptance: selected skin file changes color palette for header/body/status regions.

### RQ-072: Add `NO_COLOR` and ASCII symbol compatibility mode
- vSphere mapping: terminal compatibility.
- Acceptance: when enabled, unicode glyphs are replaced with ASCII symbols and color is disabled.

### RQ-108: Add full Unicode glyph rendering support
- vSphere mapping: richer visual parity for headers, tables, borders, and status icons.
- Acceptance: UTF-8 box-drawing, arrows, and symbol glyphs render without
  truncation, misalignment, or replacement characters in supported terminals.
- Acceptance: column width calculations use rune/cell-width-aware logic so
  wide Unicode glyphs do not break table alignment.
- Acceptance: unsupported-glyph terminals can fall back to ASCII mode without
  panics or corrupted layout.

### RQ-109: Add emoji-capable rendering mode
- vSphere mapping: expressive status and context markers in high-signal views.
- Acceptance: emoji markers render correctly in headers/status/details panels
  when emoji mode is enabled.
- Acceptance: layout remains stable when emoji appear in table cells, including
  selection/highlight rows and sorted columns.
- Acceptance: emoji mode can be toggled off to restore plain-text glyph output
  for terminals with limited emoji support.

### RQ-110: Apply Unicode status glyphs across primary views
- vSphere mapping: faster at-a-glance interpretation of health/state in dense
  inventories.
- Acceptance: VM, host, datastore, cluster, and task rows include canonical
  Unicode state glyphs (for example healthy/warn/error/running/stopped) mapped
  from existing status fields.
- Acceptance: glyph mapping is centralized in one canonical formatter used by
  table view and details view renderers.
- Acceptance: when ASCII compatibility mode is enabled, each Unicode glyph maps
  to a deterministic ASCII equivalent with unchanged semantics.

### RQ-111: Apply extended glyphs to navigation and interaction affordances
- vSphere mapping: clearer command-mode and navigation cues with less text noise.
- Acceptance: breadcrumbs, sort indicators, overflow hints, and selection
  markers use Unicode arrows/symbols in Unicode mode.
- Acceptance: describe/detail panels use Unicode section markers and bullets for
  field grouping without breaking line wrapping.
- Acceptance: all interaction glyphs downgrade to ASCII equivalents in
  compatibility mode while preserving keyboard behavior and screen layout.

### RQ-112: Apply emoji markers to high-signal operational contexts
- vSphere mapping: prioritize urgent operator signals in status-heavy workflows.
- Acceptance: emoji markers are used only in approved contexts (alarms, task
  outcomes, maintenance state, and watch/event severity), not as decorative noise.
- Acceptance: emoji marker usage is controlled by a single feature flag and a
  centralized mapping table.
- Acceptance: disabling emoji mode removes emoji while retaining the same
  severity ordering and textual meaning.

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

### RQ-901: Add all hotkeys, controls, commands, etc. to the ? menu
- vSphere mapping: objective release gates.
- Acceptance: ? menu includes list of all controls, commands, and hotkeys

## Suggested Delivery Order
- Phase 1: RQ-001 to RQ-027 (CLI + command/table interaction baseline).
- Phase 2: RQ-092 to RQ-106 (screenshot visual baseline parity).
- Phase 3: RQ-028 to RQ-046 (resource/view and query expansion).
- Phase 4: RQ-047 to RQ-065 (action pipeline + observability).
- Phase 5: RQ-066 to RQ-078 (config/customization/plugins).
- Phase 6: RQ-079 to RQ-091 (security/perf/integration/release hardening).
