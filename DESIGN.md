# DESIGN

## Short-Term Goals

### CLI Startup Parity
- [ ] Add startup flag parity for `--log-file`.
- [ ] Add safety mode startup parity for `--readonly` and `--write`.
- [ ] Add startup routing flags for `--command`, `--headless`, and `--crumbsless`.
- [ ] Sub-task: implement `--log-file` sink wiring and write-path tests.

### Table Widget Parity
- [ ] Sync table focus from mouse selection back into session row/column state.
- [ ] Add sticky multi-column headers with sort-direction glyphs.
- [ ] Add per-resource status badges in the header line.
- [ ] Add column-width autosizing based on terminal resize events.

### Terminal Readability and Styles
- [ ] Add configurable high-contrast, standard, and dim color palettes.
- [ ] Add monochrome fallback indicators for status and power-state cells.
- [ ] Add narrow-terminal compact mode with prioritized column sets per resource.
- [ ] Add horizontal scrolling indicators when columns overflow.
- [ ] Add optional unicode/ASCII symbol mode for compatibility across terminals.
- [ ] Add screenshot-baseline palette preset (black canvas, cyan frames, yellow/cyan accents).
- [ ] Add selected-row inversion style for dense table readability.
- [ ] Add status color mapping for healthy/degraded/faulted rows.

### Header Layout Parity
- [ ] Add three-zone header layout (left metadata, center hotkeys, right logo).
- [ ] Add fixed metadata label order for context/version/capacity fields.
- [ ] Add right-aligned ASCII logo block with safe clipping bounds.
- [ ] Add responsive header degradation: collapse hotkeys first, then hide logo.

### Rendering Performance
- [ ] Add batched table diff rendering to avoid full-cell rebuilds on each keystroke.
- [ ] Add render instrumentation counters for frame timing in debug mode.
- [ ] Add optional debounce for high-frequency key-repeat scenarios.
- [ ] Add adaptive redraw policy for large inventories (>10k rows).

### Prompt UX Parity
- [ ] Add command palette help for `:history`, `:suggest`, and `:ro`.
- [ ] Add inline prompt validation feedback before submit.
- [ ] Add contextual completion list rendering for view/action/filter commands.
- [ ] Add ghost-text suggestion preview while typing in prompt mode.
- [ ] Add `?` hotkey modal for per-view keybinding discovery.
- [ ] Add `ctrl-a` alias palette to browse and execute command shortcuts.
- [ ] Add tab-accept behavior for first prompt suggestion.
- [ ] Add `:ctx` completion hints for configured endpoint names.
- [ ] Add context switch status badge to the header metadata area.
- [ ] Add prompt validation for unknown `:ctx` endpoint names before submit.

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
- [ ] Add pulses-style health dashboard for live utilization and alarms.
- [ ] Add xray-style dependency explorer for VM/host/datastore/network relationships.
- [ ] Add endpoint health probing during context reconnect to classify degraded/disconnected states.

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
- [ ] Add benchmark tests for large inventory rendering and sorting.
- [ ] Add integration tests for `:ctx` switch flow and active-view refresh behavior.

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
- [ ] Sub-task: emit machine-readable parity matrix artifact from `REQUIREMENTS.md`.
- [ ] Sub-task: gate release checklist on parity matrix `done|validated` thresholds.
