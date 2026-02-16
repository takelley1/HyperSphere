# DESIGN

## Goals
- Build a real-time key-event TUI loop (non-line-buffered) using `tcell/tview` so interactions behave like k9s without per-line command entry.
- Add command history navigation and prompt suggestion behavior for aliases/resources/actions.
- Add watch-driven data refresh adapters for vSphere objects (VMs, hosts, datastores, LUN-like storage views, clusters) with consistent row identity.
- Implement async task UX for bulk actions (queued, running, success, failure) with a dedicated task/status view.
- Add dialog-driven confirmation and preview flows for mutating operations in write mode.
- Add plugin-style action hooks and per-view keymap customization similar to k9s hotkey extensibility.
- Add a read-only mode toggle and enforce it across all mutating action paths with clear status indicators.
- Add end-to-end integration tests with fake govmomi adapters for command routing, selection semantics, and action execution paths.
