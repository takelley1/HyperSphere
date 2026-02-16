# DESIGN

## Short-Term Goals

### Real-Time Terminal Input Loop
- [ ] Replace line-buffered scanner input with `tcell/tview` event handling.
- [ ] Preserve existing command parser semantics for `:view`, `!action`, `/filter`, and `:-`.
- [ ] Map current hotkeys (`SPACE`, `CTRL+SPACE`, `CTRL+\\`, `SHIFT+O`, `SHIFT+I`) to real key events.
- [ ] Keep session logic as the single canonical state owner.
- [ ] Add a persistent footer pane for contextual hotkey hints.

### Prompt UX Parity
- [ ] Add command history navigation (`UP`/`DOWN`) for prompt mode.
- [ ] Add tab-complete acceptance for active suggestions.
- [ ] Add explicit prompt mode indicator in the header/status strip.

## Medium-Term Goals

### vSphere Data Layer Integration
- [ ] Add adapter interfaces for VM, host, datastore, cluster, and LUN-like storage listing.
- [ ] Add watch/refresh adapters for periodic updates and row identity stability.
- [ ] Normalize model shaping so all table views share one canonical row pipeline.
- [ ] Add data-source health state and stale-data indicators.

### Action Execution Pipeline
- [ ] Add async task execution model for queued/running/success/failure states.
- [ ] Add a task/status view for bulk operations.
- [ ] Add cancellable task hooks where VMware APIs support cancellation.
- [ ] Add per-action timeout and retry policy wiring.

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

### Product Parity Hardening
- [ ] Add cross-view navigation polish and consistent status bars.
- [ ] Add large-environment performance benchmarks for table rendering and refresh.
- [ ] Add release criteria and acceptance checklist for k9s-parity milestones.
- [ ] Add parity matrix mapping k9s features to vSphere analogs with completion state.
