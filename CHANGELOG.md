# CHANGELOG

## 2026-02-16
- Implemented and fulfilled RQ-046 by adding `/-f text` fuzzy filter mode with
  score-based match ranking and stable tie ordering.
- Added failing-first coverage in `internal/tui/runtime_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` for fuzzy ranking and tie-order
  behavior.
- Implemented and fulfilled RQ-045 by adding `/-t key=value,...` label-filter
  support that requires all requested tag pairs to be present in row tags.
- Added failing-first coverage in `internal/tui/runtime_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` for all-pairs tag filtering behavior.
- Implemented and fulfilled RQ-044 by adding inverse regex filter-mode support
  for `/!pattern`, excluding rows that match the provided regex.
- Added failing-first coverage in `internal/tui/runtime_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` for inverse-filter execution paths.
- Implemented and fulfilled RQ-043 by adding regex filter-mode execution for
  `/pattern` commands with regex compile validation before applying row filters.
- Added failing-first coverage in `internal/tui/runtime_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` for regex filtering and invalid-pattern
  error behavior that preserves the prior filtered result set.
- Implemented and fulfilled RQ-042 by replacing modal describe behavior with a
  split-pane details drawer inside the main table layout.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert table navigation remains active while the details drawer is open.
- Implemented and fulfilled RQ-041 by adding `n`/`N` filtered-match jump-list
  navigation with wrap-around behavior across active filter result rows.
- Added failing-first coverage in `internal/tui/runtime_test.go` for forward and
  reverse match cycling and wrap semantics.
- Implemented and fulfilled RQ-040 by adding `Shift+W` namespace warp from
  folder/tag rows into VM view with scoped filtering from the selected key.
- Added failing-first coverage in `internal/tui/runtime_test.go` for folder/tag
  warp success flows and for unsupported/empty-scope/error branches.
- Implemented and fulfilled RQ-039 by adding `Shift+J` owner-jump behavior
  from VM rows to owning host rows, with fallback to resource pool owners.
- Added failing-first coverage in `internal/tui/runtime_test.go` for host-owner
  and resource-pool-owner jump flows.
- Implemented and fulfilled RQ-038 by adding hierarchical breadcrumb path
  rendering (`home > dc > cluster > host > vm`) derived from selected row
  context for datacenter/cluster/host/vm views.
- Added failing-first coverage in `internal/tui/runtime_test.go` for breadcrumb
  chain updates during drill-down and for `LastView` restoring parent path.
- Updated runtime header and breadcrumb widgets in
  `cmd/hypersphere/explorer_tui.go` to consume `Session.BreadcrumbPath()` and
  added coverage in `cmd/hypersphere/explorer_tui_test.go`.
- Implemented and fulfilled RQ-037 by adding the `:tag`/`:tags` view with
  columns `TAG`, `CATEGORY`, `CARDINALITY`, and `ATTACHED_OBJECTS`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for tag alias parsing and tag view column
  rendering.
- Extended `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded tag rows
  and added browse-dataset assertions in `cmd/hypersphere/main_test.go`.
- Implemented and fulfilled RQ-036 by adding the `:folder`/`:folders` view
  with columns `PATH`, `TYPE`, `CHILDREN`, and `VM_COUNT`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for folder alias parsing and folder view
  column rendering.
- Extended `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded folder
  rows and added browse-dataset assertions in `cmd/hypersphere/main_test.go`.
- Implemented and fulfilled RQ-035 by adding the `:alarm`/`:alarms` view with
  columns `ENTITY`, `ALARM`, `STATUS`, `TRIGGERED`, and `ACKED_BY`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for alarm alias parsing and alarm view column
  rendering.
- Extended `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded alarm
  rows and added browse-dataset assertions in `cmd/hypersphere/main_test.go`.
- Implemented and fulfilled RQ-034 by adding the `:event`/`:events` view with
  columns `TIME`, `SEVERITY`, `ENTITY`, `MESSAGE`, and `USER`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for event alias parsing and event view column
  rendering.
- Extended `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded event
  rows and added browse-dataset assertions in `cmd/hypersphere/main_test.go`.
- Implemented and fulfilled RQ-033 by adding the `:task`/`:tasks` view with
  columns `ENTITY`, `ACTION`, `STATE`, `STARTED`, `DURATION`, and `OWNER`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for task alias parsing and task view column
  rendering.
- Extended `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded task
  rows and added dataset assertions plus compact/task-title coverage in
  `cmd/hypersphere/main_test.go` and `cmd/hypersphere/explorer_tui_test.go`.
- Implemented and fulfilled RQ-125 by adding per-view column selection controls
  (`:cols set ...`, `:cols list`, `:cols reset`) with in-session persistence
  across view switches.
- Added failing-first coverage in `internal/tui/session_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` for per-view column selection,
  persistence, and reset behavior.
- Updated `internal/tui/explorer.go` session state to persist per-resource
  visible column sets and restore defaults on reset.
- Added and fulfilled RQ-128 by moving compact breadcrumb/status hints from the
  top-right zone into the center header below keyboard shortcut rows.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert path/status ordering below center shortcuts and keep right logo zone
  dedicated to logo rendering.
- Updated center-header rendering in `cmd/hypersphere/explorer_tui.go` to append
  context/status lines under shortcut rows while restoring logo-only right-zone
  rendering.
- Added and fulfilled RQ-127 by differentiating hovered/selected row highlight
  color from marked-row highlight color.
- Updated RQ-126 implementation so selected-row highlight remains yellow while
  marked rows use a distinct non-yellow color in color mode.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert selected and marked highlight colors are not the same.
- Added RQ-125 to `REQUIREMENTS.md` to track per-view column selection controls,
  including per-view persistence and restore behavior for hidden columns.
- Added and fulfilled RQ-124 by extending VM view with `SNAPSHOT_COUNT` and
  `SNAPSHOT_TOTAL_GB` columns.
- Added failing-first coverage in `internal/tui/session_test.go` for the
  expanded VM schema including snapshot-count/size fields.
- Updated VM row shaping in `internal/tui/explorer.go` and seeded VM sample
  values in `cmd/hypersphere/main.go` for snapshot total size data.
- Added and fulfilled RQ-123 by expanding VM table coverage to include runtime
  usage and placement columns: used CPU/memory/storage, IP, DNS, cluster, host,
  network, total CPU cores, total RAM, largest disk, and attached storage.
- Added failing-first coverage in `internal/tui/session_test.go` for the
  expanded VM column schema and updated compact-column expectations in
  `cmd/hypersphere/explorer_tui_test.go`.
- Updated `internal/tui/explorer.go` VM row model, view columns, cell mapping,
  and sort hotkeys to support the expanded VM schema.
- Expanded `defaultVMRows()` in `cmd/hypersphere/main.go` to seed representative
  values for all newly added VM columns.
- Added and fulfilled RQ-122 by converting top-center help/hotkey hints from a
  single-column list to compact multi-column rows in both table and log modes.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  multi-column center legend formatting and log-mode legend row packing.
- Added and fulfilled RQ-121 by de-emphasizing breadcrumbs/status into compact
  top-right header hints (`path:` and `status:`) near the logo.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  compact layout item count and top-header compact path/status rendering.
- Updated `cmd/hypersphere/explorer_tui.go` to remove standalone breadcrumb and
  status layout panels, keep only top-header/table/prompt rows, and render
  compact right-zone path/status lines with tag-stripped status text.
- Added and fulfilled RQ-120 by moving breadcrumb and status panels above the
  resource table in the vertical TUI layout.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert top-to-bottom layout order (`topHeader`, `breadcrumb`, `status`,
  `body`, `prompt`) when breadcrumbs are enabled.
- Added and fulfilled RQ-119 by adding explicit sort guidance (`<Shift+O> Sort`)
  to the top-center table help legend.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert sort-hint visibility in moved top-header help text.
- Added and fulfilled RQ-117 by changing resource-table selection styling from
  text-only reverse attributes to full-row background highlighting.
- Added and fulfilled RQ-118 by applying full-row background colors for marked
  selections (including a distinct marked+selected combination), so selection
  state is no longer glyph-only.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  full-width selected-row highlighting and full-row marked-selection coloring.
- Updated `cmd/hypersphere/explorer_tui.go` table rendering to append a trailing
  fill cell so selected/marked row backgrounds span the rendered table width.
- Updated theme wiring in `cmd/hypersphere/explorer_tui.go` with canonical row
  background colors for selected, marked, and marked+selected states, including
  NO_COLOR fallbacks.
- Added and fulfilled RQ-116 by expanding every resource table schema with
  additional operational columns to better use horizontal terminal space.
- Added failing-first coverage in `internal/tui/session_test.go` and
  `internal/tui/resource_extra_test.go` for expanded per-resource column
  expectations and cross-resource column presence checks.
- Updated `internal/tui/explorer.go` with expanded columns, new row fields, and
  sort-key mappings for VM, LUN, cluster, datacenter, resource pool, network,
  template, snapshot, host, and datastore views, including derived LUN
  `FREE_GB` and `UTIL_PERCENT` values.
- Expanded sample catalog row data in `cmd/hypersphere/main.go` to populate the
  new columns with representative values for immediate local browsing.
- Updated `DESIGN.md` with follow-on wide-schema truncation and compact-mode
  degradation coverage sub-tasks.
- Implemented RQ-032 snapshot-governance coverage by adding `:snapshot`,
  `:snap`, and `:ss` command aliases with columns `VM`, `SNAPSHOT`, `SIZE`,
  `CREATED`, `AGE`, and `QUIESCED`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for snapshot alias parsing and snapshot view
  column rendering.
- Expanded `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded
  snapshot rows and added browse-dataset assertions in
  `cmd/hypersphere/main_test.go`.
- Marked `RQ-032` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on snapshot dataset and startup-routing
  coverage sub-tasks.
- Implemented RQ-031 template lifecycle coverage by adding `:template` and
  `:tp` command aliases with columns `NAME`, `OS`, `DATASTORE`, `FOLDER`,
  and `AGE`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for template alias parsing and template view
  column rendering.
- Expanded `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded
  template rows and added browse-dataset assertions in
  `cmd/hypersphere/main_test.go`.
- Marked `RQ-031` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on template dataset and startup-routing
  coverage sub-tasks.
- Implemented RQ-030 distributed-network view coverage by adding `:network`
  and `:nw` command aliases with columns `NAME`, `TYPE`, `VLAN`, `SWITCH`,
  and `ATTACHED_VMS`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for network alias parsing and network view
  column rendering.
- Expanded `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded network
  rows and added browse-dataset assertions in `cmd/hypersphere/main_test.go`.
- Marked `RQ-030` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on network dataset and startup-routing
  coverage sub-tasks.
- Implemented RQ-029 resource-pool coverage by adding `:rp` and
  `:resourcepool` aliases with a dedicated view exposing `NAME`, `CLUSTER`,
  `CPU_RES`, `MEM_RES`, and `VM_COUNT`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for resource-pool command parsing and
  resource-pool column rendering.
- Expanded `defaultCatalog()` in `cmd/hypersphere/main.go` with seeded resource
  pool rows and added browse-dataset coverage in `cmd/hypersphere/main_test.go`.
- Marked `RQ-029` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on resource-pool dataset and startup-routing
  coverage sub-tasks.
- Implemented RQ-028 datacenter resource coverage by adding `:dc` and
  `:datacenter` command aliases and a dedicated datacenter table view with
  columns `NAME`, `CLUSTERS`, `HOSTS`, `VMS`, and `DATASTORES`.
- Added failing-first coverage in `internal/tui/command_test.go` and
  `internal/tui/session_test.go` for datacenter command parsing and view-column
  rendering, plus `cmd/hypersphere/main_test.go` coverage for seeded
  datacenter browsing rows.
- Expanded `defaultCatalog()` in `cmd/hypersphere/main.go` with canonical
  datacenter sample rows to keep the explorer dataset browseable across views.
- Marked `RQ-028` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on datacenter dataset and startup-routing
  coverage sub-tasks.
- Implemented RQ-106 compact-header degradation behavior with width thresholds:
  center legend collapses first, then right logo hides, while left metadata and
  active view title remain visible.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  collapse-before-hide ordering and narrow-width logo removal behavior.
- Marked `RQ-106` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for threshold boundary tuning
  and per-view compact legend variants.
- Refined the top-right ASCII logo by removing the `TESSERACT` label and
  enclosing the wireframe square/cube with a circular outline for clearer
  tetrahedron/tesseract-style geometry in the header.
- Updated logo snapshot expectations in
  `cmd/hypersphere/explorer_tui_test.go` to match the new unlabeled circular
  enclosure while preserving right-zone clipping checks.
- Updated the top-right ASCII logo to a tesseract-inspired wireframe design so
  the HyperSphere mark reads as 4D geometry instead of stylized text.
- Updated failing-first header-logo expectations in
  `cmd/hypersphere/explorer_tui_test.go` for the new tesseract shape while
  retaining right-zone clipping/alignment coverage.
- Implemented RQ-105 log viewport controls with deterministic shortcuts for
  `Top`, `Bottom`, `PageUp`, and `PageDown` in log mode.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  viewport offset math and runtime key-driven offset changes.
- Marked `RQ-105` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for legend/keybinding drift
  checks and offset persistence across resize redraws.
- Added and fulfilled RQ-115 by removing the bottom help bar and moving
  actionable help hints into the top-center header in cyan-accented lines.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  footer removal from layout, moved help hints, prompt/quit visibility, and
  clock-free top-header help rendering.
- Updated `DESIGN.md` with follow-on top-center help packing/alignment tasks for
  narrower terminal widths.
- Implemented RQ-104 monospaced log-line parity by adding timestamped log row
  rendering with fixed-width level markers and wrapped continuation indentation.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  timestamp/level formatting, continuation indentation, and log-mode table output.
- Marked `RQ-104` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for narrow-width wrap coverage
  and INFO/WARN/ERROR level-width normalization checks.
- Added and fulfilled RQ-114 by expanding the built-in sample catalog used by
  the explorer runtime so local browsing has more inventory breadth.
- Added failing-first coverage in `cmd/hypersphere/main_test.go` to assert
  expanded row counts per resource and mixed operational states.
- Expanded `defaultCatalog()` in `cmd/hypersphere/main.go` with richer VM/LUN/
  cluster/host/datastore sample rows while preserving canonical IDs used in tests.
- Updated `DESIGN.md` with new short-term demo dataset quality sub-tasks for
  dense inventories, long-field browsing, and describe-panel depth.
- Implemented RQ-103 log-view title parity by adding log frame title rendering
  in the format `Logs <object-path> (target=<value>)` when a sub-target is set.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  log title formatting and prompt-driven `:log ... target=...` runtime behavior.
- Marked `RQ-103` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for log-title transition tests
  and default object-path derivation from selected table row identity.
- Implemented RQ-102 view-specific legend switching by adding prompt-driven log
  mode (`:log`/`:logs`) and table mode (`:table`) toggles that swap center
  header legends between table commands and log-navigation keys.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  log-view legend content and legend restore behavior when returning to table mode.
- Marked `RQ-102` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for `:log`/`:table` integration
  coverage and runtime wiring of log navigation actions.
- Implemented RQ-101 top-metrics parity by rendering `CPU` and `MEM` metadata
  values in percent format with trend suffix support (`(+)` / `(-)`).
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  top-header CPU/MEM formatting and trend formatter behavior.
- Marked `RQ-101` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for live metric-source wiring
  and trend derivation from sampled values.
- Implemented RQ-100 selected-row inversion parity by applying a deterministic
  reverse attribute style to the selected table row without mutating cell text.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert selected-row inversion attributes and stable selected-cell text.
- Marked `RQ-100` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for compact/headless inversion
  coverage and inversion accessibility toggles.
- Implemented RQ-099 status-color parity by adding canonical status mapping for
  table rows: healthy (green), degraded (yellow), and faulted (red), with
  deterministic fallback to alternating row colors when status is unknown.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  canonical status mapping and fallback row-color behavior.
- Marked `RQ-099` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for `CONNECTION`/`POWER` status
  mapping coverage and severity-legend hints.
- Added and fulfilled RQ-113 by replacing the top-right ASCII text mark with a
  multiline 4D hypersphere-style logo projection in the explorer header.
- Updated failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert the new hypersphere ASCII lines while keeping right-zone clipping tests.
- Updated `DESIGN.md` with follow-on short-term sub-tasks for low-width
  hypersphere variants and symmetry/clipping checks during resize.
- Implemented RQ-098 screenshot palette preset parity by mapping header accents
  to yellow/cyan/magenta, table header background to cyan, and canvas background
  to black in the explorer theme.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` for
  palette field mapping and runtime top-header accent tag rendering.
- Marked `RQ-098` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed palette-preset item and adding
  follow-on sub-tasks for NO_COLOR snapshot parity and palette override plumbing.
- Implemented RQ-097 centered title-divider parity for the active content frame
  using `ViewName(scope)[count]` formatting and divider segments on both sides.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert the `ViewName(scope)[count]` payload and divider-segment title format.
- Marked `RQ-097` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on sub-tasks for divider-title update and
  overflow-indicator placement integration coverage.
- Implemented RQ-096 cyan content-frame parity by applying a canonical cyan
  border style to the active explorer table frame.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert the active content view border color and frame-color helper behavior.
- Marked `RQ-096` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on frame-visibility and border-consistency
  sub-tasks for screenshot and table-mode coverage.
- Implemented RQ-095 right-side ASCII logo parity by rendering a canonical
  seven-line HyperSphere ASCII block in the top-right header zone.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert exact multiline logo output and right-zone clipping/alignment bounds.
- Marked `RQ-095` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed logo-block item and adding
  follow-on header logo resize/clipping sub-tasks.
- Implemented RQ-094 center hotkey legend parity so the top-center legend now
  renders one angle-bracket hotkey entry per line (`<:> Command`,
  `</> Filter`, `<?> Help`).
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert line-by-line center legend output and fixed-zone header rendering.
- Marked `RQ-094` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on short/medium/long-term legend-parity
  subtasks for integration coverage and screenshot validation.
- Implemented RQ-093 left metadata panel parity by rendering the top-left
  header zone as a fixed seven-line metadata panel in the required label order:
  `Context`, `Cluster`, `User`, `HS Version`, `vCenter Version`, `CPU`, `MEM`.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  assert exact metadata label order and line-by-line output.
- Updated `cmd/hypersphere/explorer_tui.go` to render multi-line top-header
  zones and reserve fixed header panel height for metadata visibility.
- Marked `RQ-093` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed metadata-order item and adding
  follow-on metadata value/trend sub-tasks.
- Implemented RQ-107 `d` describe-panel parity in explorer runtime with a
  dedicated describe modal that opens on `d` and closes on `Esc`.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  verify describe-panel open/close behavior, selection/column restoration, and
  mark preservation after close.
- Added canonical selected-resource detail modeling in
  `internal/tui/explorer.go`, including VM required fields and snapshot
  identifier/timestamp rendering.
- Added focused detail-coverage tests in `internal/tui/session_test.go` for VM
  required fields, non-VM fallback detail rendering, and error/fallback paths.
- Marked `RQ-107` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed describe-panel sub-tasks and adding
  follow-on detail-format and empty-selection coverage sub-tasks.
- Implemented RQ-092 three-zone top header layout in the explorer runtime with
  fixed left/center/right zones and deterministic clipping/padding behavior.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to
  verify 120-column zone rendering and non-overlap between metadata, hotkeys,
  and logo text.
- Updated `cmd/hypersphere/explorer_tui.go` to add a dedicated top-header
  widget, resize-aware redraw wiring, and canonical zone-format helper
  functions.
- Marked `RQ-092` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed three-zone header item and
  adding follow-on top-header clipping and resize integration sub-tasks.
- Implemented RQ-027 `ctrl-e` header-visibility toggle parity in the explorer
  runtime.
- Added failing-first runtime coverage in `cmd/hypersphere/explorer_tui_test.go`
  to verify repeated header toggle behavior preserves selected row/column identity.
- Updated `cmd/hypersphere/explorer_tui.go` to map `ctrl-e`, toggle table
  header rendering on/off, and keep header-offset selection behavior consistent.
- Marked `RQ-027` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed `ctrl-e` sub-tasks and adding
  follow-on `d` describe-panel implementation/test sub-tasks.
- Implemented RQ-026 `ctrl-w` wide-column toggle parity in the explorer runtime.
- Added failing-first runtime coverage in `cmd/hypersphere/explorer_tui_test.go`
  to verify schema toggling and selected-row identity preservation across toggles.
- Updated `cmd/hypersphere/explorer_tui.go` to map `ctrl-w`, switch between
  standard and wide column schemas, and keep the selected object stable.
- Marked `RQ-026` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed `ctrl-w` sub-tasks and adding the
  next `ctrl-e` header-toggle follow-on sub-tasks.
- Implemented RQ-025 mark-count badge parity by rendering header mark totals as
  `Marks[n]` and updating the count after mark, unmark, and clear flows.
- Added failing-first runtime coverage in `internal/tui/runtime_test.go` to
  verify header badge transitions for `0 -> 1 -> 0` and clear-mark reset via
  `CTRL+\`.
- Marked `RQ-025` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with follow-on short-term sub-tasks for `ctrl-w`
  wide-column toggle state wiring and selection-preservation tests.
- Implemented RQ-024 row-focus sync parity by wiring table selection-change events to session selection state.
- Added failing-first runtime coverage in `cmd/hypersphere/explorer_tui_test.go` to verify click-selected rows become hotkey/action targets.
- Added `Session.SetSelection` clamping support in `internal/tui/explorer.go` and internal coverage for in-range, negative, high, and empty-view bounds.
- Marked `RQ-024` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed mouse selection-sync item and adding follow-on selection-sync integration subtasks.
- Implemented RQ-022 overflow indicator parity so the explorer table title renders
  left/right markers (`◀`/`▶`) only when additional columns are off-screen.
- Added failing-first runtime coverage in `cmd/hypersphere/explorer_tui_test.go`
  for right-only overflow at initial offset, both-side overflow after horizontal
  offset, and no markers when all columns fit.
- Added overflow marker helpers in `cmd/hypersphere/explorer_tui.go` that compute
  hidden-column state from autosized widths, available render width, and current
  table column offset.
- Marked `RQ-022` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed overflow-indicator sub-task and
  adding follow-on integration/style sub-tasks for overflow marker behavior.
- Implemented RQ-016 inline prompt validation state so invalid command syntax is reported before command execution.
- Added failing-first runtime coverage in `cmd/hypersphere/explorer_tui_test.go` to validate pre-submit error status and prompt highlight/reset behavior.
- Wired prompt changed-event handling in `cmd/hypersphere/explorer_tui.go` to parse command input live and surface deterministic `[red]command error: ...` status output.
- Added prompt validation styling helpers in `cmd/hypersphere/explorer_tui.go` to highlight invalid prompt input and reset styles when input becomes valid or prompt mode exits.
- Updated `DESIGN.md` by removing the completed inline prompt validation roadmap item and adding follow-on validation style/reset subtasks.
- Implemented RQ-014 command-mode history traversal parity for `:history up`
  and `:history down` with bounded, non-skipping cursor behavior.
- Added failing-first command-runtime coverage in
  `cmd/hypersphere/explorer_tui_test.go` to assert ordered traversal and
  boundary clamping at both ends of prompt history.
- Updated `internal/tui/prompt.go` history `Next()` behavior to clamp at the
  newest entry instead of moving past bounds.
- Updated `internal/tui/prompt_test.go` edge-branch assertions for bounded tail
  traversal behavior.
- Marked `RQ-014` as fulfilled in `REQUIREMENTS.md`.
- Added a follow-on prompt UX roadmap sub-task in `DESIGN.md` for displaying
  history traversal position.
- Implemented RQ-013 file-backed command alias registry support using
  `~/.hypersphere/aliases.yaml` with `HYPERSPHERE_ALIASES_FILE` override.
- Added failing-first alias execution coverage in
  `cmd/hypersphere/explorer_tui_test.go` to verify alias entries resolve to
  canonical commands, including commands with optional arguments.
- Added focused alias loader/parser tests in
  `cmd/hypersphere/alias_registry_test.go` for missing-file behavior, invalid
  entry validation, and argument-appending alias expansion.
- Wired prompt command execution through alias resolution in
  `cmd/hypersphere/explorer_tui.go` before parsing command kinds.
- Added parser coverage in `internal/tui/command_test.go` and parsing support in
  `internal/tui/explorer.go` so resource view commands accept trailing optional
  arguments (for alias-expansion compatibility).
- Marked `RQ-013` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with new short/medium/long-term follow-on tasks for alias
  registry status surfacing, hot reload, context-scoped overlays, and alias
  integration coverage.

## 2026-02-15
- Implemented RQ-021 terminal-resize column autosizing parity in the full-screen explorer table.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to verify fixed-priority column widths remain intact (`SEL` and `NAME`) while non-priority columns shrink to fit narrow widths.
- Added autosizing width helpers and runtime rendering integration in `cmd/hypersphere/explorer_tui.go`, including resize-aware redraw wiring through the `tview` before-draw hook.
- Marked `RQ-021` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed autosizing roadmap item and adding follow-on subtasks for overflow indicators and repeated-resize integration coverage.
- Implemented RQ-020 selected-column sort glyph parity so sorted headers show
  `↑` and `↓` and flip deterministically with sort direction changes.
- Added failing-first coverage in `internal/tui/runtime_test.go` to assert the
  table header line renders `[NAME↑]` then `[NAME↓]` after repeat sort input.
- Updated `internal/tui/explorer.go` header decoration logic to render
  sort-direction glyphs on the sorted column instead of a generic `*` marker.
- Updated `internal/tui/explorer_coverage_test.go` for the new sorted-header
  glyph output.
- Marked `RQ-020` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed sort-glyph roadmap item and
  adding a follow-on ASCII glyph fallback sub-task.
- Implemented RQ-019 sticky table header parity by adding a selection-driven body viewport so vertical scrolling changes visible rows while the table header remains fixed.
- Added failing-first coverage in `internal/tui/runtime_test.go` to assert sticky header persistence and body-row changes after vertical scroll.
- Added viewport helper branch coverage in `internal/tui/explorer_coverage_test.go` and removed dead defensive branching in `internal/tui/explorer.go` to preserve the enforced 100% coverage gate.
- Marked `RQ-019` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed sticky-header roadmap item and adding follow-on viewport indicator/configuration sub-tasks plus explicit sort-glyph work.
- Implemented RQ-017 prompt-mode discoverability parity by adding `:-` to command suggestions.
- Added failing-first coverage in `internal/tui/prompt_test.go` to verify `:-` suggestion matching.
- Updated `internal/tui/prompt.go` suggestion candidates to include the last-view toggle command.
- Marked `RQ-017` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` with a follow-on integration sub-task for repeated `:-` prompt-toggle coverage.
- Marked `RQ-016` as fulfilled in `REQUIREMENTS.md`.
- Added failing-first coverage in `cmd/hypersphere/explorer_tui_test.go` to verify invalid
  trailing-space prompt input still shows inline validation errors before submit.
- Updated `cmd/hypersphere/explorer_tui.go` pending-input detection so trailing spaces only
  suppress validation when required arguments are still being entered.
- Updated `DESIGN.md` Prompt UX parity sub-tasks with a follow-on focused trailing-space
  pending-input test target.
- Implemented RQ-015 Tab completion parity so pressing `Tab` in prompt mode
  always applies the first suggestion when suggestions exist, including
  whitespace-normalization cases.
- Added failing-first runtime coverage in
  `cmd/hypersphere/explorer_tui_test.go` to verify Tab completion updates
  prompt input text to canonical suggestion index `0`.
- Updated `cmd/hypersphere/explorer_tui.go` completion behavior to compare the
  full input value (not only trimmed input) before deciding no-op completion.
- Marked `RQ-015` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed prompt Tab-accept sub-task.

- Implemented RQ-012 `ctrl-a` alias palette behavior in the explorer runtime.
- Added failing-first runtime tests in `cmd/hypersphere/explorer_tui_test.go` to validate `ctrl-a` palette open behavior, sorted alias ordering, and alias-command execution.
- Added an alias palette modal to `cmd/hypersphere/explorer_tui.go` with explicit open/close state tracking and selection handlers that execute the selected `:alias` command.
- Refactored resource alias handling to a canonical alias map in `internal/tui/explorer.go` and reused it from prompt suggestion generation in `internal/tui/prompt.go`.
- Updated `DESIGN.md` by removing the completed `ctrl-a` palette sub-task and adding follow-on integration coverage for alias-palette lifecycle behavior.
- Marked `RQ-012` as fulfilled in `REQUIREMENTS.md`.
- Implemented RQ-011 `?` key behavior to open a keymap help modal in explorer runtime and close it with `Esc`.
- Added failing-first cmd-level coverage in `cmd/hypersphere/explorer_tui_test.go` for help modal open/close lifecycle and action-content rendering.
- Refactored runtime root layout to use `tview.Pages` with a dedicated help modal page and explicit modal state tracking.
- Updated `DESIGN.md` by removing the completed `?` modal prompt task and adding follow-on integration coverage for modal behavior across view switches.
- Marked `RQ-011` as fulfilled in `REQUIREMENTS.md`.
- Implemented RQ-010 `--crumbsless` startup flag to hide the breadcrumb widget in explorer mode.
- Added failing-first CLI coverage in `cmd/hypersphere/main_test.go` to validate `--crumbsless` parsing.
- Added cmd-level rendering tests in `cmd/hypersphere/explorer_tui_test.go` to assert default breadcrumb rendering and crumbsless omission behavior.
- Wired `--crumbsless` parsing in `cmd/hypersphere/main.go` and runtime breadcrumb layout/render controls in `cmd/hypersphere/explorer_tui.go`.
- Marked `RQ-010` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed crumbsless startup tasks and adding a follow-on integration sub-task for crumbsless startup routing.
- Marked `RQ-009` as fulfilled in `REQUIREMENTS.md`.
- Implemented RQ-009 `--headless` startup flag to hide the explorer table header row.
- Added failing-first tests in `cmd/hypersphere/main_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` to validate CLI parsing and headerless
  first-frame rendering behavior.
- Wired `--headless` parsing in `cmd/hypersphere/main.go` and startup runtime
  rendering behavior in `cmd/hypersphere/explorer_tui.go`, including row
  selection offset and fixed-row handling without a header line.
- Updated `DESIGN.md` by removing completed startup headless tasks and adding
  follow-on integration coverage for headless view switching.
- Implemented RQ-008 `--command` startup flag to route explorer directly into
  the requested resource view on first render.
- Added failing-first tests in `cmd/hypersphere/main_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` for CLI parsing and first-frame view
  selection/rendering from startup command input.
- Wired startup command parsing in `cmd/hypersphere/main.go` and runtime
  startup command execution in `cmd/hypersphere/explorer_tui.go`.
- Marked `RQ-008` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed `--command` startup subtasks and
  adding follow-on startup command validation/alias coverage work.
- Implemented RQ-007 `--write` startup flag override behavior for explorer mode.
- Added failing-first CLI tests in `cmd/hypersphere/main_test.go` to validate
  config-driven `readOnly: true` defaults and `--write` override precedence.
- Added startup config parsing in `cmd/hypersphere/main.go` to read
  `~/.hypersphere/config.yaml` `readOnly` values and apply canonical precedence:
  `--write` > `--readonly` > config default.
- Refactored CLI flag wiring into small helper functions for startup parsing.
- Marked `RQ-007` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed `--write` startup items and adding
  follow-on startup routing sub-tasks for `--command`, `--headless`, and
  `--crumbsless`.
- Implemented RQ-006 `--readonly` startup flag support for explorer sessions.
- Added failing-first tests in `cmd/hypersphere/main_test.go` and
  `cmd/hypersphere/explorer_tui_test.go` covering flag parsing and deterministic
  mutating-action rejection in read-only mode.
- Wired CLI startup state through `cmd/hypersphere/main.go` into
  `cmd/hypersphere/explorer_tui.go` so runtime sessions initialize read-only.
- Confirmed canonical error output for blocked mutating actions:
  `[red]command error: read-only mode`.
- Marked `RQ-006` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed `--readonly` subtasks and adding
  follow-on `--write` precedence and non-interactive read-only wiring tasks.
- Implemented RQ-005 `--log-file` startup flag support to write runtime logs
  to an operator-defined file path.
- Added failing-first CLI coverage in `cmd/hypersphere/main_test.go` to assert
  custom log-file creation and startup record content.
- Added canonical startup log emission in `cmd/hypersphere/main.go` with a
  structured `time/level/message` log line.
- Marked `RQ-005` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing completed `--log-file` parity tasks and
  adding the next startup safety sub-task for `--readonly`.
- Implemented RQ-004 `--log-level` startup flag parsing with canonical
  `debug|info|warn|error` mapping.
- Added failing-first CLI tests in `cmd/hypersphere/main_test.go` for valid
  log-level mapping and invalid value parse errors.
- Added canonical log-level parsing in `cmd/hypersphere/main.go` with explicit
  parse-time validation and error reporting.
- Marked `RQ-003` and `RQ-004` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` to remove completed `--log-level` startup parity work and
  keep remaining CLI startup tasks active.
- Implemented RQ-003 `--refresh` startup flag parsing with float-seconds support.
- Added failing-first CLI tests in `cmd/hypersphere/main_test.go` for minimum
  clamp behavior and unchanged values above the minimum.
- Added canonical `--refresh` parsing in `cmd/hypersphere/main.go` with
  `minimumRefreshSeconds` clamping logic.
- Updated `DESIGN.md` by removing the completed `--refresh` sub-task from
  short-term CLI startup parity goals.
- Implemented RQ-002 `hypersphere info` command parity with runtime path introspection output.
- Added failing-first CLI coverage in `cmd/hypersphere/main_test.go` to assert required
  `config/logs/dumps/skins/plugins/hotkeys` keys and absolute path values.
- Added `info` subcommand wiring in `cmd/hypersphere/main.go` plus canonical path
  resolution under `~/.hypersphere`.
- Marked `RQ-002` as fulfilled in `REQUIREMENTS.md`.
- Updated `DESIGN.md` by removing the completed info-command item and adding new
  short/medium/long-term parity subtasks.
- Implemented RQ-001 `hypersphere version` command parity with fielded build output.
- Added test-first CLI coverage in `cmd/hypersphere/main_test.go` to assert semantic version, commit SHA, and build date fields.
- Refactored top-level CLI startup path into a testable `run(args, stdout, stderr)` entrypoint and isolated subcommand parsing.
- Marked `RQ-001` as fulfilled in `REQUIREMENTS.md`.
- Added short-term CLI startup parity follow-on tasks in `DESIGN.md`.
- Implemented RQ-018 `:ctx` command parity in explorer prompt mode.
- Added parser support for `:ctx` and `:ctx <endpoint>` with validation for extra arguments.
- Added runtime context manager plumbing with configured endpoint listing and endpoint switch handling.
- Added active-view refresh after context switch so the current resource view reconnects and re-renders deterministically.
- Added regression tests for `:ctx` parsing, context list status output, and context switch refresh behavior.
- Updated `DESIGN.md` by removing the completed context-switch goal and adding follow-on subtasks for context completion, header status, and integration coverage.
- Marked `RQ-018` as fulfilled in `REQUIREMENTS.md`.
- Added screenshot-driven visual parity requirements (RQ-092 to RQ-106) to
  define the target k9s-style look for HyperSphere.
- Added atomic UI requirements for header zoning, cyan framing, centered title
  dividers, palette mapping, row status coloring, and log-view formatting.
- Updated delivery phases so screenshot visual parity is implemented before
  broader resource and action surface expansion.
- Added matching short-term `DESIGN.md` tasks for header layout and screenshot
  baseline style parity.
- Removed periodic 1-second explorer redraw loop and switched to input-driven rendering to reduce flicker/stutter.
- Removed footer realtime clock to avoid forced full-screen repaint cadence unrelated to user interaction.
- Added cmd-level regression test to keep footer clock-free for event-driven redraw behavior.
- Added `REQUIREMENTS.md` with 91 atomic, test-ready k9s-to-vSphere parity requirements.
- Mapped k9s feature families into concrete vSphere analogs across CLI, UX,
  resources, actions, plugins, and quality gates.
- Added explicit acceptance criteria for each requirement to support failing-test-first development.
- Expanded `DESIGN.md` with newly discovered parity tasks across short, medium,
  and long-term horizons.
- Added prompt `Tab` completion in the realtime explorer; it now accepts the first command/action suggestion from the current view context.
- Added prompt completion status messaging so accepted completions are immediately visible in the status panel.
- Updated footer help text to include prompt completion behavior.
- Added cmd-level tests for completion success, no-match behavior, and `Tab` event handling in prompt mode.
- Added a terminal theme loader with `NO_COLOR` support so the explorer can run with readable monochrome output on color-limited terminals.
- Added table styling polish for readability: fixed header styling, alternating row colors, and stronger selected-row highlighting.
- Added vim-style navigation semantics in the realtime runtime (`h`/`l` for column movement, `j`/`k` row movement preserved) plus plain left/right arrow support.
- Updated footer help text to advertise vim/arrow movement semantics directly in the TUI.
- Added tests for vim key translation, `NO_COLOR` theme behavior, and plain arrow-based column movement.
- Switched explorer runtime from line-buffered input to a full-screen real-time `tview`/`tcell` event loop as the default app entrypoint workflow.
- Replaced text body rendering with a canonical table widget (`tview.Table`) to mirror k9s-style grid interaction semantics.
- Added table rendering helpers to inject a leading `SEL` column and mark selected objects with `*` for bulk action visibility.
- Added session accessor APIs (`SelectedRow`, `SelectedColumn`, `IsMarked`) to keep table focus and mark state synchronized without duplicating state ownership.
- Added prompt-mode status visibility in footer (`Prompt: ON/OFF`) while preserving existing command hints and realtime clock updates.
- Added cmd-level tests for table row shaping, table selection offsets, status branch behavior, and footer prompt-mode indicator.
- Added internal session accessor tests and restored internal coverage gate to 100% after runtime integration changes.
- Added a greenfield Go TUI architecture inspired by k9s layering with `cmd`, `internal/app`, and workflow-focused internal packages.
- Implemented canonical datastore migration planning and execution logic with threshold checks, source exclusion, candidate re-ranking, fallback tiers, skip-reason taxonomy, dry-run gating, and retry behavior.
- Implemented canonical pending-deletion lifecycle logic with VM metadata fields (`pd_pending_since`, `pd_delete_on`, `pd_owner_email`, `pd_initial_notice_sent`, `pd_reminder_notice_sent`, `pd_original_name`) and idempotent mark/remind/purge/reset behavior.
- Implemented terminal renderers for migration and deletion action plans to support non-GUI TUI flows.
- Added an executable CLI at `cmd/hypersphere/main.go` with workflow selection and example data flows.
- Added comprehensive unit tests for config precedence, migration planning/execution, lifecycle behavior, TUI rendering, and app orchestration.
- Achieved 100% statement coverage across `internal/...` packages and added script-enforced coverage gating.
- Added `scripts/lint.sh` and `scripts/test.sh` for repeatable formatting, vetting, and test execution.
- Added a k9s-style command-mode explorer flow where users can enter `:vm`, `:lun`, or `:cluster` to switch active resource views and render tabular results.
- Implemented `internal/tui` navigator and resource-table renderer with colon-command parsing, active view state, and formatted column output.
- Added app-level command execution integration via `RunExplorerCommand` and wired CLI `--workflow explorer` interactive input handling with `:q`/`:quit` exit commands.
- Added unit tests for command parsing, unknown resource handling, table rendering, app integration, and all branch paths while keeping `internal/...` coverage at 100%.
- Added interactive table session state with k9s-style row marking semantics: `Space` toggles marks, marked rows are preserved by object identity, and bulk actions target marks or fallback to the current row.
- Added VMware action mapping per resource (`vm`, `lun`, `cluster`) and bulk action execution via an API adapter interface to support actions like VM power operations, migration, and tag edits.
- Expanded resource views with object-relevant columns (for example VM `NAME/TAGS/CLUSTER/POWER/DATASTORE/OWNER`) and added sort hotkeys per resource plus `Shift+O` selected-column sorting.
- Updated explorer CLI loop to support command-mode view switching (`:vm/:lun/:cluster`), hotkey-driven table interactions, and action execution (`!<action>`).
- Added comprehensive branch tests for session navigation, selection, sorting, rendering, action execution, and helper behaviors, preserving 100% internal coverage.
- Added command parser parity layer for explorer mode with typed commands (`view`, `action`, `hotkey`, `filter`, `last-view`, `help`, `quit`) and alias support (for example `:vms`, `:ds`, `:hosts`).
- Added additional resource domains to mirror broader k9s-style surface area in vSphere semantics: `host` and `datastore` views with resource-specific columns, sort hotkeys, and supported action sets.
- Added session runtime parity features inspired by k9s interaction flow: previous-view toggle (`:-`), filter mode (`/text`), sort inversion (`SHIFT+I`), and read-only action blocking.
- Extended mark mechanics with range marking (`CTRL+SPACE`) and mark clearing (`CTRL+\\`) alongside existing `SPACE` toggles.
- Updated explorer command loop to use canonical parsed command kinds, reducing ad-hoc branching and centralizing command behavior.
- Added comprehensive tests for parser semantics, new resources, mark controls, filter behavior, last-view toggling, sort inversion, and read-only gating while preserving 100% internal coverage.
- Added a `DESIGN.md` goal tracker for ongoing k9s-to-vSphere parity work; fulfilled goals are removed as completed.
- Updated project workflow to keep `.scratchpad.txt` for active execution notes and commit in regular validated increments.
- Reorganized `DESIGN.md` into short-term, medium-term, and long-term goals, each with explicit sub-tasks.
- Established `DESIGN.md` as the canonical active roadmap where completed goals/sub-tasks are removed when fulfilled.
- Added read-only mode command parsing (`:ro`, `:readonly on|off|toggle`) in explorer command handling.
- Added read-only state indicator in table headers (`Mode: RO` / `Mode: RW`) for top-level explorer visibility.
- Added read-only runtime APIs on session state and tests for parser branches, header rendering, and action gating behavior.
- Removed fulfilled roadmap items from `DESIGN.md` and expanded remaining short/medium/long-term subtasks.
- Updated main CLI defaults so running the app without `--workflow` now launches explorer/TUI mode.
- Added prompt state support for bounded history and command suggestions across resources, aliases, actions, and sort keys.
- Added command-mode parser support for `:history up/down` and `:suggest <prefix>` and wired these into explorer runtime handling.
- Removed fulfilled prompt-suggestion roadmap items from `DESIGN.md` and kept remaining goals/subtasks active.
