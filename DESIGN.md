# DESIGN

## Short-Term Goals

### Real-Time Terminal Input Loop
- [ ] Replace line-buffered scanner input with `tcell/tview` event handling.
- [ ] Preserve existing command parser semantics for `:view`, `!action`, `/filter`, and `:-`.
- [ ] Map current hotkeys (`SPACE`, `CTRL+SPACE`, `CTRL+\\`, `SHIFT+O`, `SHIFT+I`) to real key events.
- [ ] Keep session logic as the single canonical state owner.

### Prompt UX Parity
- [ ] Add command history navigation (`UP`/`DOWN`) for prompt mode.
- [ ] Add command suggestions for resources, aliases, and actions.
- [ ] Add inline prompt feedback for invalid commands without leaving the current view.

### Safety and Modes
- [ ] Surface read-only mode status in the top-level explorer UI.
- [ ] Add a runtime toggle command for read-only mode.
- [ ] Enforce read-only checks across every mutating action path.

## Medium-Term Goals

### vSphere Data Layer Integration
- [ ] Add adapter interfaces for VM, host, datastore, cluster, and LUN-like storage listing.
- [ ] Add watch/refresh adapters for periodic updates and row identity stability.
- [ ] Normalize model shaping so all table views share one canonical row pipeline.

### Action Execution Pipeline
- [ ] Add async task execution model for queued/running/success/failure states.
- [ ] Add a task/status view for bulk operations.
- [ ] Add cancellable task hooks where VMware APIs support cancellation.

### Write-Path UX
- [ ] Add confirmation dialogs for destructive actions.
- [ ] Add action preview summaries (targets, impact, rollback notes).
- [ ] Add consistent error presentation and retry affordances.

## Long-Term Goals

### Extensibility Model
- [ ] Add plugin-style action hooks for custom vSphere operations.
- [ ] Add per-view keymap customization in config.
- [ ] Add customizable command aliases and action macros.

### End-to-End Quality
- [ ] Add fake govmomi-backed integration tests for view routing and table refresh.
- [ ] Add integration tests for mark semantics and bulk action execution.
- [ ] Add integration tests for failure handling, retries, and read-only enforcement.

### Product Parity Hardening
- [ ] Add cross-view navigation polish and consistent status bars.
- [ ] Add large-environment performance benchmarks for table rendering and refresh.
- [ ] Add release criteria and acceptance checklist for k9s-parity milestones.
