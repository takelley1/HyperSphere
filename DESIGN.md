# DESIGN

## Short-Term Goals

### CLI Startup Parity
- [ ] Sub-task: add startup command validation status when `--command` is unknown.
- [ ] Sub-task: add startup alias coverage for `--command` values (`vm`, `vms`, `ds`).
- [ ] Sub-task: add integration coverage for `--headless` with in-session view switching.
- [ ] Sub-task: add integration coverage for `--crumbsless` with startup view routing.

### Table Widget Parity
- [ ] Sub-task: add sticky-header viewport offset indicator (`rows x-y of z`) in the footer.
- [ ] Sub-task: add sticky-header viewport size configuration for dense vs standard layouts.
- [ ] Sub-task: add ASCII-compatible sort-direction glyph fallback (`^`/`v`) for NO_COLOR mode.
- [ ] Add per-resource status badges in the header line.
- [ ] Sub-task: extend describe-panel formatting coverage for host/lun/datastore focused fields.
- [ ] Sub-task: add describe-panel empty-selection status messaging integration coverage.
- [ ] Sub-task: add integration coverage for repeated terminal resize width changes.
- [ ] Sub-task: add integration coverage for overflow markers across column-offset transitions.
- [ ] Sub-task: surface overflow markers in screenshot-baseline frame/title styling.
- [ ] Sub-task: add selection-sync integration coverage for headless mode table clicks.
- [ ] Sub-task: add mouse click row-selection behavior for mixed compact/non-compact views.

### Terminal Readability and Styles
- [ ] Add configurable high-contrast, standard, and dim color palettes.
- [ ] Add monochrome fallback indicators for status and power-state cells.
- [ ] Add optional unicode/ASCII symbol mode for compatibility across terminals.
- [ ] Add screenshot-baseline palette preset (black canvas, cyan frames, yellow/cyan accents).
- [ ] Add selected-row inversion style for dense table readability.
- [ ] Add status color mapping for healthy/degraded/faulted rows.

### Header Layout Parity
- [ ] Add right-aligned ASCII logo block with safe clipping bounds.
- [ ] Add responsive header degradation: collapse hotkeys first, then hide logo.
- [ ] Add runtime width tracking tests for top-header zone clipping and padding.
- [ ] Add top-header render integration coverage for resize-driven redraw behavior.
- [ ] Sub-task: add integration coverage for center legend one-entry-per-line rendering.
- [ ] Sub-task: replace placeholder metadata values with live cluster/user/vCenter stats.
- [ ] Sub-task: add metadata trend suffix rendering for CPU and MEM values.

### Rendering Performance
- [ ] Add batched table diff rendering to avoid full-cell rebuilds on each keystroke.
- [ ] Add render instrumentation counters for frame timing in debug mode.
- [ ] Add optional debounce for high-frequency key-repeat scenarios.
- [ ] Add adaptive redraw policy for large inventories (>10k rows).

### Prompt UX Parity
- [ ] Add command palette help for `:history`, `:suggest`, and `:ro`.
- [ ] Add prompt-history position badge while traversing command history.
- [ ] Add contextual completion list rendering for view/action/filter commands.
- [ ] Add ghost-text suggestion preview while typing in prompt mode.
- [ ] Add alias-registry parse error surfacing in prompt status and startup status.
- [ ] Add alias-registry hot-reload command for iterative alias editing.
- [ ] Add integration coverage for help-modal lifecycle across view switches.
- [ ] Add `:ctx` completion hints for configured endpoint names.
- [ ] Add context switch status badge to the header metadata area.
- [ ] Add prompt validation for unknown `:ctx` endpoint names before submit.
- [ ] Add integration coverage for alias-palette lifecycle (`ctrl-a`, `Esc`, alias execution).
- [ ] Add prompt validation style parity between label and input field in color and NO_COLOR modes.
- [ ] Add integration coverage for `:-` last-view toggle behavior across repeated prompt invocations.
- [ ] Sub-task: add focused tests for prompt validation reset when exiting prompt mode.
- [ ] Sub-task: add focused tests for pending-input states with trailing spaces across valid commands.

## Medium-Term Goals

### vSphere Data Layer Integration
- [ ] Add adapter interfaces for VM, host, datastore, cluster, and LUN-like storage listing.
- [ ] Add watch/refresh adapters for periodic updates and row identity stability.
- [ ] Normalize model shaping so all table views share one canonical row pipeline.
- [ ] Add data-source health state and stale-data indicators.
- [ ] Sub-task: define canonical adapter errors for disconnected and permission-denied states.
- [ ] Sub-task: add fake adapter fixtures for deterministic watch/update tests.

### Action Execution Pipeline
- [ ] Add async task execution model for queued/running/success/failure states.
- [ ] Add a task/status view for bulk operations.
- [ ] Add cancellable task hooks where VMware APIs support cancellation.
- [ ] Add per-action timeout and retry policy wiring.

### Navigation and UX Depth
- [ ] Add breadcrumb navigation for datacenter -> cluster -> host -> VM context.
- [ ] Add split-pane detail drawers for selected object metadata.
- [ ] Add search result jump list with next/previous navigation.
- [ ] Add keyboard cheatsheet modal with per-view hotkeys.
- [ ] Sub-task: centralize per-view legend definitions for table and log contexts.
- [ ] Add context-scoped alias overlays so aliases can vary by active vCenter target.
- [ ] Add pulses-style health dashboard for live utilization and alarms.
- [ ] Add xray-style dependency explorer for VM/host/datastore/network relationships.
- [ ] Add endpoint health probing during context reconnect to classify degraded/disconnected states.
- [ ] Add startup-view badge in header metadata to show active routed command.

### Write-Path UX
- [ ] Add confirmation dialogs for destructive actions.
- [ ] Add action preview summaries (targets, impact, rollback notes).
- [ ] Add consistent error presentation and retry affordances.
- [ ] Add audit trail summaries for completed bulk actions.

## Long-Term Goals

### Extensibility Model
- [ ] Add plugin-style action hooks for custom vSphere operations.
- [ ] Add per-view keymap customization in config.
- [ ] Add customizable command aliases and action macros.
- [ ] Add skin/theme customization parity for terminal visuals.

### End-to-End Quality
- [ ] Add fake govmomi-backed integration tests for view routing and table refresh.
- [ ] Add integration tests for mark semantics and bulk action execution.
- [ ] Add integration tests for failure handling, retries, and read-only enforcement.
- [ ] Add integration tests for alias registry loading and optional-argument expansion.
- [ ] Add benchmark tests for large inventory rendering and sorting.
- [ ] Add integration tests for `:ctx` switch flow and active-view refresh behavior.
- [ ] Add integration tests for startup `--command` view routing and first-frame rendering.

### API and Plugin Surface
- [ ] Add versioned explorer API contracts for external adapters.
- [ ] Add compatibility tests for plugin action registration.
- [ ] Add plugin sandboxing and permission prompts for write actions.
- [ ] Add plugin telemetry hooks for action lifecycle events.

### Product Parity Hardening
- [ ] Add cross-view navigation polish and consistent status bars.
- [ ] Add large-environment performance benchmarks for table rendering and refresh.
- [ ] Add release criteria and acceptance checklist for k9s-parity milestones.
- [ ] Add parity matrix mapping k9s features to vSphere analogs with completion state.
- [ ] Add a canonical requirements ledger linking parity items to failing-test IDs.
- [ ] Sub-task: validate center legend parity items against screenshot baselines.
- [ ] Sub-task: emit machine-readable parity matrix artifact from `REQUIREMENTS.md`.
- [ ] Sub-task: gate release checklist on parity matrix `done|validated` thresholds.
