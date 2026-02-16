# CHANGELOG

## 2026-02-15
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
