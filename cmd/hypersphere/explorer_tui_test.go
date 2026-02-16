// Path: cmd/hypersphere/explorer_tui_test.go
// Description: Validate table-model helpers and prompt/runtime branch behavior for the full-screen explorer.
package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/takelley1/hypersphere/internal/tui"
)

func TestTableRowsIncludesMarkerColumnAndMarks(t *testing.T) {
	view := tui.ResourceView{
		Columns: []string{"NAME", "CLUSTER"},
		Rows:    [][]string{{"vm-a", "cluster-a"}, {"vm-b", "cluster-b"}},
		IDs:     []string{"vm-a", "vm-b"},
	}
	rows := tableRows(view, func(id string) bool {
		return id == "vm-b"
	}, true)
	if len(rows) != 3 {
		t.Fatalf("expected header + 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "SEL" || rows[0][1] != "NAME" {
		t.Fatalf("unexpected header row: %#v", rows[0])
	}
	if rows[1][0] != " " || rows[2][0] != "*" {
		t.Fatalf("unexpected marker values: %#v %#v", rows[1], rows[2])
	}
}

func TestSelectionForTableOffsetsHeaderAndMarkerColumns(t *testing.T) {
	session := tui.NewSession(tui.Catalog{VMs: []tui.VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.HandleKey("DOWN"); err != nil {
		t.Fatalf("HandleKey error: %v", err)
	}
	if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("HandleKey error: %v", err)
	}
	row, column := selectionForTable(session, true)
	if row != 2 || column != 2 {
		t.Fatalf("expected selected table position 2,2 got %d,%d", row, column)
	}
}

func TestHandleTableSelectionChangedSyncsSessionSelectionForHotkeys(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	clickedID := runtime.session.CurrentView().IDs[1]

	runtime.handleTableSelectionChanged(2, 1)

	if runtime.session.SelectedRow() != 1 {
		t.Fatalf("expected selected row 1 after click, got %d", runtime.session.SelectedRow())
	}
	if err := runtime.session.HandleKey("SPACE"); err != nil {
		t.Fatalf("HandleKey error: %v", err)
	}
	if !runtime.session.IsMarked(clickedID) {
		t.Fatalf("expected clicked row id %q to be marked", clickedID)
	}
}

func TestAutosizedColumnWidthsPreserveFixedPriorityColumns(t *testing.T) {
	view := tui.ResourceView{
		Resource: tui.ResourceVM,
		Columns:  []string{"NAME", "TAGS", "CLUSTER", "POWER"},
		Rows: [][]string{
			{"vm-super-long-name", "prod,critical", "cluster-east", "poweredOn"},
		},
		IDs: []string{"vm-super-long-name"},
	}
	rows := tableRows(view, func(string) bool { return false }, true)
	natural := naturalColumnWidths(rows)
	available := natural[0] + natural[1]
	available += (len(natural)-2)*minAutosizeColumnWidth + (len(natural) - 1)

	widths := autosizedColumnWidths(view, rows, available)

	if widths[0] != natural[0] {
		t.Fatalf("expected selection column width to remain fixed at %d, got %d", natural[0], widths[0])
	}
	if widths[1] != natural[1] {
		t.Fatalf("expected name column width to remain fixed at %d, got %d", natural[1], widths[1])
	}
	if widths[2] >= natural[2] {
		t.Fatalf("expected non-priority tags column to shrink from %d, got %d", natural[2], widths[2])
	}
	if tableRenderWidth(widths) > available {
		t.Fatalf("expected autosized widths to fit available width %d, got %d", available, tableRenderWidth(widths))
	}
}

func TestRenderTableWithWidthRecalculatesAndPreservesFixedColumns(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	rows := tableRows(runtime.session.CurrentView(), runtime.session.IsMarked, true)
	natural := naturalColumnWidths(rows)
	available := natural[0] + natural[1]
	available += (len(natural)-2)*minAutosizeColumnWidth + (len(natural) - 1)

	runtime.renderTableWithWidth(available)

	nameHeader := runtime.body.GetCell(0, 1)
	tagsHeader := runtime.body.GetCell(0, 2)
	if nameHeader.MaxWidth != natural[1] {
		t.Fatalf(
			"expected fixed name header width %d after render, got %d",
			natural[1],
			nameHeader.MaxWidth,
		)
	}
	if tagsHeader.MaxWidth >= natural[2] {
		t.Fatalf(
			"expected non-priority tags header width to shrink from %d, got %d",
			natural[2],
			tagsHeader.MaxWidth,
		)
	}
}

func TestRenderTableWithWidthShowsOverflowMarkersForHiddenColumns(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	runtime.renderTableWithWidth(compactModeWidthThreshold + 1)

	title := runtime.body.GetTitle()
	if strings.Contains(title, "◀") {
		t.Fatalf("did not expect left overflow marker at initial offset, got %q", title)
	}
	if !strings.Contains(title, "▶") {
		t.Fatalf("expected right overflow marker when columns are hidden, got %q", title)
	}

	runtime.body.SetOffset(0, 1)
	runtime.renderTableWithWidth(compactModeWidthThreshold + 1)
	title = runtime.body.GetTitle()
	if !strings.Contains(title, "◀") || !strings.Contains(title, "▶") {
		t.Fatalf("expected left and right overflow markers after horizontal offset, got %q", title)
	}
}

func TestRenderTableWithWidthHidesOverflowMarkersWhenAllColumnsVisible(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	rows := tableRows(runtime.session.CurrentView(), runtime.session.IsMarked, true)
	widths := naturalColumnWidths(rows)
	runtime.renderTableWithWidth(tableRenderWidth(widths))

	title := runtime.body.GetTitle()
	if strings.Contains(title, "◀") || strings.Contains(title, "▶") {
		t.Fatalf("did not expect overflow markers when all columns fit, got %q", title)
	}
}

func TestRenderTableWithWidthUsesResourceCompactColumnsOnNarrowWidths(t *testing.T) {
	testCases := []struct {
		command string
		want    []string
	}{
		{command: "vm", want: []string{"NAME", "POWER", "DATASTORE"}},
		{command: "lun", want: []string{"NAME", "DATASTORE", "USED_GB"}},
		{command: "cluster", want: []string{"NAME", "HOSTS", "VMS"}},
		{command: "host", want: []string{"NAME", "CLUSTER", "CONNECTION"}},
		{command: "datastore", want: []string{"NAME", "CLUSTER", "FREE_GB"}},
	}
	for _, tc := range testCases {
		runtime := newExplorerRuntimeWithStartupCommand(false, tc.command)
		runtime.renderTableWithWidth(compactModeWidthThreshold - 1)
		if runtime.body.GetColumnCount() != len(tc.want)+1 {
			t.Fatalf(
				"expected compact column count %d for %s, got %d",
				len(tc.want)+1,
				tc.command,
				runtime.body.GetColumnCount(),
			)
		}
		if runtime.body.GetCell(0, 0).Text != "SEL" {
			t.Fatalf("expected marker header to remain present for %s", tc.command)
		}
		for index, expected := range tc.want {
			if runtime.body.GetCell(0, index+1).Text != expected {
				t.Fatalf(
					"expected compact header %q at index %d for %s, got %q",
					expected,
					index+1,
					tc.command,
					runtime.body.GetCell(0, index+1).Text,
				)
			}
		}
	}
}

func TestEmitStatusWritesErrorsOnly(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.emitStatus(nil)
	if strings.Contains(runtime.status.GetText(true), "command error") {
		t.Fatalf("did not expect error status for nil")
	}
	runtime.emitStatus(tui.ErrUnsupportedHotKey)
	if !strings.Contains(runtime.status.GetText(true), "unsupported hotkey") {
		t.Fatalf("expected unsupported hotkey in status")
	}
}

func TestRenderFooterIncludesPromptMode(t *testing.T) {
	if !strings.Contains(renderFooter(true), "Prompt: ON") {
		t.Fatalf("expected prompt-on indicator in footer")
	}
	if !strings.Contains(renderFooter(false), "Prompt: OFF") {
		t.Fatalf("expected prompt-off indicator in footer")
	}
}

func TestRenderTopHeaderLineUsesThreeFixedZonesWithoutOverlapAt120Columns(t *testing.T) {
	line := renderTopHeaderLine(
		120,
		"Context: vc-primary",
		"<:> Command",
		"HyperSphere",
	)
	if len(line) != 120 {
		t.Fatalf("expected 120-column header line, got %d", len(line))
	}
	leftZone := line[:40]
	centerZone := line[40:80]
	rightZone := line[80:]
	if !strings.Contains(leftZone, "Context: vc-primary") {
		t.Fatalf("expected left metadata in left zone, got %q", leftZone)
	}
	if !strings.Contains(centerZone, "<:> Command") {
		t.Fatalf("expected center legend in center zone, got %q", centerZone)
	}
	if !strings.Contains(rightZone, "HyperSphere") {
		t.Fatalf("expected logo text in right zone, got %q", rightZone)
	}
	if strings.Contains(line[:80], "HyperSphere") {
		t.Fatalf("expected right-zone text to avoid left/center overlap: %q", line[:80])
	}
	if strings.Contains(line[40:], "Context: vc-primary") {
		t.Fatalf("expected left-zone text to avoid center/right overlap: %q", line[40:])
	}
}

func TestRenderTopHeaderCenterUsesOneAngleBracketEntryPerLine(t *testing.T) {
	lines := strings.Split(renderTopHeaderCenter(), "\n")
	want := []string{
		"<:> Command",
		"</> Filter",
		"<?> Help",
	}
	if len(lines) != len(want) {
		t.Fatalf("expected %d center legend lines, got %d (%q)", len(want), len(lines), lines)
	}
	for index, expected := range want {
		if lines[index] != expected {
			t.Fatalf(
				"expected center legend line %d to be %q, got %q",
				index,
				expected,
				lines[index],
			)
		}
	}
}

func TestRenderTopHeaderRightUsesMultilineASCIILogoBlock(t *testing.T) {
	lines := strings.Split(renderTopHeaderRight(), "\n")
	want := []string{
		" _   _                        ",
		"| | | |_   _ _ __   ___ _ __ ",
		"| |_| | | | | '_ \\ / _ \\ '__|",
		"|  _  | |_| | |_) |  __/ |   ",
		"|_| |_|\\__, | .__/ \\___|_|   ",
		"        __/ | |              ",
		"       |___/|_|              ",
	}
	if len(lines) != len(want) {
		t.Fatalf("expected %d logo lines, got %d (%q)", len(want), len(lines), lines)
	}
	for index, expected := range want {
		if lines[index] != expected {
			t.Fatalf(
				"expected logo line %d to be %q, got %q",
				index,
				expected,
				lines[index],
			)
		}
	}
}

func TestRenderTopHeaderLinesRightLogoIsRightAlignedAndClippedToZone(t *testing.T) {
	width := 60
	rightLines := strings.Split(renderTopHeaderRight(), "\n")
	lines := renderTopHeaderLines(width, []string{""}, []string{""}, rightLines)
	_, _, rightWidth := topHeaderZoneWidths(width)
	for index, line := range lines {
		if len(line) != width {
			t.Fatalf("expected header line width %d, got %d", width, len(line))
		}
		leftAndCenter := line[:width-rightWidth]
		if strings.TrimSpace(leftAndCenter) != "" {
			t.Fatalf("expected empty left and center zones for line %d, got %q", index, leftAndCenter)
		}
		rightZone := line[width-rightWidth:]
		if rightZone != fitHeaderRight(rightLines[index], rightWidth) {
			t.Fatalf("expected clipped right-zone logo at line %d, got %q", index, rightZone)
		}
	}
}

func TestRenderTopHeaderLeftUsesFixedMetadataLabelOrder(t *testing.T) {
	lines := strings.Split(renderTopHeaderLeft("vc-primary"), "\n")
	want := []string{
		"Context: vc-primary",
		"Cluster: n/a",
		"User: n/a",
		"HS Version: 0.0.0",
		"vCenter Version: unknown",
		"CPU: n/a",
		"MEM: n/a",
	}
	if len(lines) != len(want) {
		t.Fatalf("expected %d metadata lines, got %d (%q)", len(want), len(lines), lines)
	}
	for index, expected := range want {
		if lines[index] != expected {
			t.Fatalf(
				"expected metadata line %d to be %q, got %q",
				index,
				expected,
				lines[index],
			)
		}
	}
}

func TestEventToHotKeyVimColumnMovement(t *testing.T) {
	left, ok := eventToHotKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	if !ok || left != "LEFT" {
		t.Fatalf("expected h to map to LEFT, got %q ok=%v", left, ok)
	}
	right, ok := eventToHotKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	if !ok || right != "RIGHT" {
		t.Fatalf("expected l to map to RIGHT, got %q ok=%v", right, ok)
	}
	sortHotKey, ok := eventToHotKey(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModShift))
	if !ok || sortHotKey != "H" {
		t.Fatalf("expected shifted H to remain sort hotkey, got %q ok=%v", sortHotKey, ok)
	}
}

func TestHandleGlobalKeyCtrlWTogglesWideColumnsAndPreservesSelectedIdentity(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	if err := runtime.session.HandleKey("DOWN"); err != nil {
		t.Fatalf("expected row move to succeed: %v", err)
	}
	selectedID := runtime.session.CurrentView().IDs[runtime.session.SelectedRow()]
	wideWidth := compactModeWidthThreshold + 10
	runtime.renderTableWithWidth(wideWidth)
	defaultColumns := runtime.body.GetColumnCount()

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyCtrlW, 0, tcell.ModCtrl))
	runtime.renderTableWithWidth(wideWidth)

	narrowColumns := runtime.body.GetColumnCount()
	if narrowColumns >= defaultColumns {
		t.Fatalf(
			"expected ctrl-w to switch to fewer columns, got default=%d toggled=%d",
			defaultColumns,
			narrowColumns,
		)
	}
	currentID := runtime.session.CurrentView().IDs[runtime.session.SelectedRow()]
	if currentID != selectedID {
		t.Fatalf("expected selected id to stay %q, got %q", selectedID, currentID)
	}

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyCtrlW, 0, tcell.ModCtrl))
	runtime.renderTableWithWidth(wideWidth)

	if runtime.body.GetColumnCount() != defaultColumns {
		t.Fatalf(
			"expected second ctrl-w toggle to restore default columns %d, got %d",
			defaultColumns,
			runtime.body.GetColumnCount(),
		)
	}
}

func TestHandleGlobalKeyCtrlETogglesHeaderVisibilityWithoutSelectionReset(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	if err := runtime.session.HandleKey("DOWN"); err != nil {
		t.Fatalf("expected row move to succeed: %v", err)
	}
	if err := runtime.session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("expected column move to succeed: %v", err)
	}
	selectedID := runtime.session.CurrentView().IDs[runtime.session.SelectedRow()]
	selectedColumn := runtime.session.SelectedColumn()

	runtime.renderTableWithWidth(compactModeWidthThreshold + 10)
	if runtime.body.GetCell(0, 0).Text != "SEL" {
		t.Fatalf("expected header row to be visible before ctrl-e")
	}

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyCtrlE, 0, tcell.ModCtrl))
	runtime.renderTableWithWidth(compactModeWidthThreshold + 10)
	if runtime.body.GetCell(0, 0).Text == "SEL" {
		t.Fatalf("expected header row to be hidden after ctrl-e")
	}
	if runtime.session.CurrentView().IDs[runtime.session.SelectedRow()] != selectedID {
		t.Fatalf("expected selected id to remain %q after hiding header", selectedID)
	}
	if runtime.session.SelectedColumn() != selectedColumn {
		t.Fatalf("expected selected column %d after hiding header, got %d", selectedColumn, runtime.session.SelectedColumn())
	}

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyCtrlE, 0, tcell.ModCtrl))
	runtime.renderTableWithWidth(compactModeWidthThreshold + 10)
	if runtime.body.GetCell(0, 0).Text != "SEL" {
		t.Fatalf("expected header row to be visible after second ctrl-e")
	}
	if runtime.session.CurrentView().IDs[runtime.session.SelectedRow()] != selectedID {
		t.Fatalf("expected selected id to remain %q after restoring header", selectedID)
	}
	if runtime.session.SelectedColumn() != selectedColumn {
		t.Fatalf("expected selected column %d after restoring header, got %d", selectedColumn, runtime.session.SelectedColumn())
	}
}

func TestReadThemeRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	theme := readTheme()
	if theme.UseColor {
		t.Fatalf("expected NO_COLOR to disable color")
	}
	os.Unsetenv("NO_COLOR")
	theme = readTheme()
	if !theme.UseColor {
		t.Fatalf("expected color mode enabled when NO_COLOR is unset")
	}
}

func TestApplyPromptCompletionUsesFirstSuggestion(t *testing.T) {
	promptState := tui.NewPromptState(20)
	view := tui.ResourceView{Actions: []string{"power-on", "power-off"}}
	value, status, changed := applyPromptCompletion(&promptState, view, "!pow")
	if !changed {
		t.Fatalf("expected completion to be applied")
	}
	if value != "!power-off" {
		t.Fatalf("expected !power-off completion, got %q", value)
	}
	if !strings.Contains(status, "!power-off") {
		t.Fatalf("expected completion status to mention !power-off, got %q", status)
	}
}

func TestApplyPromptCompletionNoMatch(t *testing.T) {
	promptState := tui.NewPromptState(20)
	view := tui.ResourceView{}
	value, status, changed := applyPromptCompletion(&promptState, view, "!does-not-exist")
	if changed {
		t.Fatalf("did not expect completion change")
	}
	if value != "!does-not-exist" {
		t.Fatalf("expected input unchanged, got %q", value)
	}
	if status != "" {
		t.Fatalf("expected empty status for no completion, got %q", status)
	}
}

func TestHandlePromptHistoryTabCompletesPrompt(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.prompt.SetText(":v")
	evt := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	consumed := runtime.handlePromptHistory(evt)
	if consumed != nil {
		t.Fatalf("expected tab event to be consumed")
	}
	if runtime.prompt.GetText() != ":vm" {
		t.Fatalf("expected prompt completion :vm, got %q", runtime.prompt.GetText())
	}
	if !strings.Contains(runtime.status.GetText(true), "completion: :vm") {
		t.Fatalf("expected status to include completion message")
	}
}

func TestHandlePromptHistoryTabNormalizesPromptToFirstSuggestion(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.prompt.SetText(":vm ")
	evt := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	_ = runtime.handlePromptHistory(evt)
	if runtime.prompt.GetText() != ":vm" {
		t.Fatalf("expected prompt text normalized to first suggestion, got %q", runtime.prompt.GetText())
	}
}

func TestHandlePromptChangedShowsValidationBeforeExecution(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.promptMode = true

	runtime.handlePromptChanged(":readonly maybe")
	if !strings.Contains(runtime.status.GetText(true), "command error: invalid action") {
		t.Fatalf("expected prompt validation error status, got %q", runtime.status.GetText(true))
	}
	labelColor, _, _ := runtime.prompt.GetLabelStyle().Decompose()
	if labelColor != tcell.ColorRed {
		t.Fatalf("expected prompt label to highlight validation error, got %v", labelColor)
	}

	runtime.handlePromptChanged(":readonly on")
	if strings.Contains(runtime.status.GetText(true), "command error:") {
		t.Fatalf("expected prompt validation status to clear for valid input")
	}
	labelColor, _, _ = runtime.prompt.GetLabelStyle().Decompose()
	if labelColor != tcell.ColorWhite {
		t.Fatalf("expected prompt label to reset after valid input, got %v", labelColor)
	}
}

func TestHandlePromptChangedInvalidCommandWithTrailingSpaceShowsValidationError(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.promptMode = true

	runtime.handlePromptChanged(":readonly maybe ")
	if !strings.Contains(runtime.status.GetText(true), "command error: invalid action") {
		t.Fatalf(
			"expected prompt validation error with trailing space, got %q",
			runtime.status.GetText(true),
		)
	}
	labelColor, _, _ := runtime.prompt.GetLabelStyle().Decompose()
	if labelColor != tcell.ColorRed {
		t.Fatalf("expected prompt label to highlight trailing-space validation error, got %v", labelColor)
	}
}

func TestHelpModalToggleWithQuestionAndEscape(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone))
	if !runtime.isHelpModalOpen() {
		t.Fatalf("expected help modal to open on ?")
	}
	if !strings.Contains(runtime.helpText, "power-on") {
		t.Fatalf("expected help modal to include active view actions")
	}
	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if runtime.isHelpModalOpen() {
		t.Fatalf("expected help modal to close on escape")
	}
}

func TestAliasPaletteOpensOnCtrlAWithSortedAliases(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyCtrlA, 0, tcell.ModCtrl))
	if !runtime.isAliasPaletteOpen() {
		t.Fatalf("expected alias palette to open on ctrl-a")
	}
	if len(runtime.aliasEntries) == 0 {
		t.Fatalf("expected alias entries to be populated")
	}
	for index := 1; index < len(runtime.aliasEntries); index++ {
		if runtime.aliasEntries[index-1] > runtime.aliasEntries[index] {
			t.Fatalf("expected aliases sorted alphabetically: %v", runtime.aliasEntries)
		}
	}
}

func TestAliasPaletteSelectionExecutesAliasCommand(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.openAliasPalette()
	runtime.handleAliasSelection(0, ":host")
	if runtime.session.CurrentView().Resource != tui.ResourceHost {
		t.Fatalf("expected alias selection to switch to host view")
	}
	if runtime.isAliasPaletteOpen() {
		t.Fatalf("expected alias palette to close after selection")
	}
	if !strings.Contains(runtime.status.GetText(true), "view: host") {
		t.Fatalf("expected status to report selected alias execution")
	}
}

func TestRenderFooterOmitsClockForEventDrivenRedraw(t *testing.T) {
	footer := renderFooter(true)
	clock := regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}\b`)
	if clock.MatchString(footer) {
		t.Fatalf("expected footer without realtime clock: %q", footer)
	}
}

func TestExecutePromptCommandCtxListShowsConfiguredEndpoints(t *testing.T) {
	session := tui.NewSession(defaultCatalog())
	promptState := tui.NewPromptState(20)
	executor := &runtimeActionExecutor{}
	contexts := newRuntimeContextManager()

	message, keepRunning := executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		nil,
		":ctx",
	)

	if !keepRunning {
		t.Fatalf("expected ctx list command to keep runtime alive")
	}
	if !strings.Contains(message, "contexts:") {
		t.Fatalf("expected ctx list status, got %q", message)
	}
	if !strings.Contains(message, "vc-primary") {
		t.Fatalf("expected default context list to include vc-primary, got %q", message)
	}
}

func TestExecutePromptCommandResolvesAliasFileEntriesWithOptionalArgs(t *testing.T) {
	aliasPath := filepath.Join(t.TempDir(), "aliases.yaml")
	content := "go-host: :host\nrepeat: :history up\n"
	if err := os.WriteFile(aliasPath, []byte(content), 0o600); err != nil {
		t.Fatalf("expected alias file write to succeed: %v", err)
	}
	registry, err := loadCommandAliasRegistry(aliasPath)
	if err != nil {
		t.Fatalf("expected alias registry load to succeed: %v", err)
	}
	session := tui.NewSession(defaultCatalog())
	promptState := tui.NewPromptState(20)
	promptState.Record(":vm")
	executor := &runtimeActionExecutor{}
	contexts := newRuntimeContextManager()

	message, keepRunning := executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		&registry,
		":go-host",
	)

	if !keepRunning {
		t.Fatalf("expected alias view command to keep runtime alive")
	}
	if !strings.Contains(message, "view: host") {
		t.Fatalf("expected host view status from alias command, got %q", message)
	}
	if session.CurrentView().Resource != tui.ResourceHost {
		t.Fatalf("expected alias command to select host view")
	}

	message, keepRunning = executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		&registry,
		":repeat",
	)
	if !keepRunning {
		t.Fatalf("expected alias history command to keep runtime alive")
	}
	if message != "history: :go-host" {
		t.Fatalf("expected alias optional-arg command to resolve history up, got %q", message)
	}
}

func TestExecutePromptCommandHistoryTraversalIsBoundedWithoutSkipping(t *testing.T) {
	session := tui.NewSession(defaultCatalog())
	promptState := tui.NewPromptState(3)
	executor := &runtimeActionExecutor{}
	contexts := newRuntimeContextManager()
	inputs := []string{":vm", ":host", ":datastore", ":cluster"}
	for _, input := range inputs {
		if _, keepRunning := executePromptCommand(
			&session,
			&promptState,
			executor,
			&contexts,
			nil,
			input,
		); !keepRunning {
			t.Fatalf("expected command %q to keep runtime alive", input)
		}
	}
	assertHistoryMessage := func(command string, expected string) {
		t.Helper()
		message, keepRunning := executePromptCommand(
			&session,
			&promptState,
			executor,
			&contexts,
			nil,
			command,
		)
		if !keepRunning {
			t.Fatalf("expected command %q to keep runtime alive", command)
		}
		if message != expected {
			t.Fatalf("expected %q for %q, got %q", expected, command, message)
		}
	}
	assertHistoryMessage(":history up", "history: :cluster")
	assertHistoryMessage(":history up", "history: :datastore")
	assertHistoryMessage(":history up", "history: :host")
	assertHistoryMessage(":history up", "history: :host")
	assertHistoryMessage(":history down", "history: :datastore")
	assertHistoryMessage(":history down", "history: :cluster")
	assertHistoryMessage(":history down", "history: :cluster")
}

func TestExecutePromptCommandCtxSelectRefreshesActiveView(t *testing.T) {
	session := tui.NewSession(defaultCatalog())
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("unexpected initial view error: %v", err)
	}
	if err := session.HandleKey("DOWN"); err != nil {
		t.Fatalf("unexpected row move error: %v", err)
	}
	promptState := tui.NewPromptState(20)
	executor := &runtimeActionExecutor{}
	contexts := newRuntimeContextManager()

	message, keepRunning := executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		nil,
		":ctx vc-lab",
	)

	if !keepRunning {
		t.Fatalf("expected ctx select command to keep runtime alive")
	}
	if !strings.Contains(message, "context: vc-lab") {
		t.Fatalf("expected selected context status, got %q", message)
	}
	if session.CurrentView().Resource != tui.ResourceHost {
		t.Fatalf("expected active view to remain host after refresh")
	}
	if session.SelectedRow() != 0 {
		t.Fatalf("expected active view refresh to reset selection row, got %d", session.SelectedRow())
	}
}

func TestNewExplorerRuntimeWithReadOnlyBlocksMutatingAction(t *testing.T) {
	runtime := newExplorerRuntimeWithReadOnly(true)
	message, keepRunning := executePromptCommand(
		&runtime.session,
		&runtime.promptState,
		runtime.actionExec,
		&runtime.contexts,
		nil,
		"!power-off",
	)
	if !keepRunning {
		t.Fatalf("expected action command to keep runtime alive")
	}
	if message != "[red]command error: read-only mode" {
		t.Fatalf("expected deterministic read-only error, got %q", message)
	}
}

func TestNewExplorerRuntimeWithStartupCommandSelectsInitialView(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "host")
	if runtime.session.CurrentView().Resource != tui.ResourceHost {
		t.Fatalf("expected startup command host to set host view")
	}
	runtime.renderTableWithWidth(compactModeWidthThreshold + 1)
	if runtime.body.GetCell(0, 2).Text != "TAGS" {
		t.Fatalf("expected host table to render immediately for startup command")
	}
}

func TestNewExplorerRuntimeHeadlessOmitsTableHeader(t *testing.T) {
	runtime := newExplorerRuntimeWithOptions(false, "host", true)
	if runtime.body.GetCell(0, 0).Text != " " {
		t.Fatalf("expected first row to start with selection marker when headless")
	}
	if runtime.body.GetCell(0, 1).Text != "esxi-01" {
		t.Fatalf("expected first row to render host data when headless")
	}
}

func TestNewExplorerRuntimeRendersBreadcrumbByDefault(t *testing.T) {
	runtime := newExplorerRuntimeWithRenderOptions(false, "host", false, false)
	if !strings.Contains(runtime.breadcrumb.GetText(true), "home > host") {
		t.Fatalf("expected breadcrumb text to render active view")
	}
}

func TestNewExplorerRuntimeCrumbslessOmitsBreadcrumbWidget(t *testing.T) {
	runtime := newExplorerRuntimeWithRenderOptions(false, "host", false, true)
	if runtime.breadcrumb.GetText(true) != "" {
		t.Fatalf("expected crumbsless runtime to keep breadcrumb empty")
	}
	if runtime.layout.GetItemCount() != 5 {
		t.Fatalf("expected crumbsless layout to omit breadcrumb widget")
	}
}

func TestDescribePanelOpensOnDAndEscRestoresSelectionAndMarks(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone))
	selectedRow := runtime.session.SelectedRow()
	selectedColumn := runtime.session.SelectedColumn()
	selectedID := runtime.session.CurrentView().IDs[selectedRow]
	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	if !runtime.isDescribePanelOpen() {
		t.Fatalf("expected describe panel to open on d")
	}
	if !strings.Contains(runtime.describeText, "NAME: vm-b") {
		t.Fatalf("expected describe output to include selected VM details")
	}
	if !strings.Contains(runtime.describeText, "POWER_STATE: off") {
		t.Fatalf("expected describe output to include VM power state")
	}

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))

	if runtime.isDescribePanelOpen() {
		t.Fatalf("expected describe panel to close on escape")
	}
	if runtime.session.SelectedRow() != selectedRow {
		t.Fatalf("expected selected row %d after escape", selectedRow)
	}
	if runtime.session.SelectedColumn() != selectedColumn {
		t.Fatalf("expected selected column %d after escape", selectedColumn)
	}
	if !runtime.session.IsMarked(selectedID) {
		t.Fatalf("expected marked row to remain marked after describe close")
	}
}
