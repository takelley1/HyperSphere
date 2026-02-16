# CHANGELOG

## 2026-02-16
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
