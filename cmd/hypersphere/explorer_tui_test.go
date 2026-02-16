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

func TestComposeTableTitleUsesCenteredViewNameScopeCountFormat(t *testing.T) {
	view := tui.ResourceView{
		Resource: tui.ResourceVM,
		Rows:     [][]string{{"vm-a"}, {"vm-b"}},
	}
	title := composeTableTitle(view, false, false)
	if !strings.Contains(title, "VM(all)[2]") {
		t.Fatalf("expected ViewName(scope)[count] title format, got %q", title)
	}
}

func TestComposeTableTitleAddsDividerSegmentsOnBothSides(t *testing.T) {
	view := tui.ResourceView{
		Resource: tui.ResourceHost,
		Rows:     [][]string{{"host-a"}},
	}
	title := composeTableTitle(view, false, false)
	if !strings.Contains(title, "─ HOST(all)[1] ─") {
		t.Fatalf("expected divider segments around centered title, got %q", title)
	}
}

func TestComposeTableTitleUsesTaskResourceLabel(t *testing.T) {
	view := tui.ResourceView{
		Resource: tui.ResourceTask,
		Rows:     [][]string{{"vm-a"}},
	}
	title := composeTableTitle(view, false, false)
	if !strings.Contains(title, "TASK(all)[1]") {
		t.Fatalf("expected task resource label in title, got %q", title)
	}
}

func TestComposeLogTitleIncludesObjectPathAndTarget(t *testing.T) {
	title := composeLogTitle("vm/vm-a", "vmware.log")
	if !strings.Contains(title, "Logs vm/vm-a (target=vmware.log)") {
		t.Fatalf("expected log title with object path and target, got %q", title)
	}
}

func TestHandlePromptDoneLogCommandSetsLogFrameTitleWithTarget(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.startPrompt(":log vm/vm-a target=vmware.log")
	runtime.handlePromptDone(tcell.KeyEnter)
	runtime.renderTableWithWidth(compactModeWidthThreshold + 10)
	title := runtime.body.GetTitle()
	if !strings.Contains(title, "Logs vm/vm-a (target=vmware.log)") {
		t.Fatalf("expected log frame title with target, got %q", title)
	}
}

func TestRenderLogLinesUsesTimestampLevelAndWrappedContinuationIndentation(t *testing.T) {
	entries := []runtimeLogEntry{
		{
			Timestamp: "2026-02-16T11:20:00Z",
			Level:     "INFO",
			Message:   "task completed with very long detail string for wrapped output validation",
		},
	}
	lines := renderLogLines(entries, 24)
	if len(lines) < 2 {
		t.Fatalf("expected wrapped log output to span multiple lines, got %d (%q)", len(lines), lines)
	}
	if !strings.HasPrefix(lines[0], "2026-02-16T11:20:00Z INFO ") {
		t.Fatalf("expected timestamp+level prefix in first log line, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], strings.Repeat(" ", logContinuationIndentWidth())) {
		t.Fatalf("expected wrapped continuation indentation, got %q", lines[1])
	}
}

func TestRenderTableWithWidthInLogModeShowsTimestampedMonospaceRows(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.startPrompt(":log")
	runtime.handlePromptDone(tcell.KeyEnter)
	runtime.renderTableWithWidth(320)
	logCell := runtime.body.GetCell(1, 1).Text
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T`).MatchString(logCell) {
		t.Fatalf("expected timestamped log cell, got %q", logCell)
	}
	if !strings.Contains(logCell, " INFO ") && !strings.Contains(logCell, " WARN ") && !strings.Contains(logCell, " ERROR ") {
		t.Fatalf("expected level marker in log row, got %q", logCell)
	}
}

func TestRenderTableWithWidthUsesResourceCompactColumnsOnNarrowWidths(t *testing.T) {
	testCases := []struct {
		command string
		want    []string
	}{
		{command: "vm", want: []string{"NAME", "POWER", "ATTACHED_STORAGE"}},
		{command: "lun", want: []string{"NAME", "DATASTORE", "USED_GB"}},
		{command: "cluster", want: []string{"NAME", "HOSTS", "VMS"}},
		{command: "task", want: []string{"ENTITY", "ACTION", "STATE"}},
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

func TestTableRowColorMapsCanonicalStatusValues(t *testing.T) {
	theme := explorerTheme{
		UseColor:    true,
		EvenRowText: tcell.ColorWhite,
		OddRowText:  tcell.ColorLightGray,
		RowHealthy:  tcell.ColorGreen,
		RowDegraded: tcell.ColorYellow,
		RowFaulted:  tcell.ColorRed,
	}
	view := tui.ResourceView{
		Columns: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"obj-a", "healthy"},
			{"obj-b", "degraded"},
			{"obj-c", "faulted"},
		},
	}
	if color := tableRowColor(theme, view, 0); color != tcell.ColorGreen {
		t.Fatalf("expected healthy row to render green, got %v", color)
	}
	if color := tableRowColor(theme, view, 1); color != tcell.ColorYellow {
		t.Fatalf("expected degraded row to render yellow, got %v", color)
	}
	if color := tableRowColor(theme, view, 2); color != tcell.ColorRed {
		t.Fatalf("expected faulted row to render red, got %v", color)
	}
}

func TestTableRowColorFallsBackToAlternatingPaletteForUnknownStatus(t *testing.T) {
	theme := explorerTheme{
		UseColor:    true,
		EvenRowText: tcell.ColorWhite,
		OddRowText:  tcell.ColorLightGray,
		RowHealthy:  tcell.ColorGreen,
		RowDegraded: tcell.ColorYellow,
		RowFaulted:  tcell.ColorRed,
	}
	view := tui.ResourceView{
		Columns: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"obj-a", "unknown"},
			{"obj-b", "n/a"},
		},
	}
	if color := tableRowColor(theme, view, 0); color != tcell.ColorWhite {
		t.Fatalf("expected unknown even row to use even fallback color, got %v", color)
	}
	if color := tableRowColor(theme, view, 1); color != tcell.ColorLightGray {
		t.Fatalf("expected unknown odd row to use odd fallback color, got %v", color)
	}
}

func TestRenderTopHeaderCenterIncludesMovedHelpHintsAndPromptState(t *testing.T) {
	lines := strings.Split(renderTopHeaderCenter(false, true), "\n")
	want := []string{
		"<:> Command    </> Filter",
		"<?> Help       <!> Action",
		"<Tab> Complete <h/j/k/l> Move",
		"<Shift+O> Sort Prompt: ON | <q> Quit",
	}
	if len(lines) != len(want) {
		t.Fatalf("expected %d center lines, got %d (%q)", len(want), len(lines), lines)
	}
	for index, expected := range want {
		if lines[index] != expected {
			t.Fatalf("expected center help line %d to be %q, got %q", index, expected, lines[index])
		}
	}
}

func TestRenderTableWithWidthHighlightsSelectedRowAcrossFullTableWidth(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	if err := runtime.session.HandleKey("DOWN"); err != nil {
		t.Fatalf("expected row move to succeed: %v", err)
	}
	rows := tableRows(runtime.session.CurrentView(), runtime.session.IsMarked, true)
	availableWidth := tableRenderWidth(naturalColumnWidths(rows)) + 8
	runtime.renderTableWithWidth(availableWidth)

	selectedRow, selectedColumn := selectionForTable(runtime.session, true)
	cell := runtime.body.GetCell(selectedRow, selectedColumn)
	if cell.Text != "vm-b" {
		t.Fatalf("expected selected cell text to remain unchanged, got %q", cell.Text)
	}
	if runtime.body.GetColumnCount() <= len(runtime.session.CurrentView().Columns)+1 {
		t.Fatalf("expected filler column for full-width highlight")
	}
	expected := selectedRowBackgroundColor(runtime.theme)
	for columnIndex := 0; columnIndex < runtime.body.GetColumnCount(); columnIndex++ {
		rowCell := runtime.body.GetCell(selectedRow, columnIndex)
		if rowCell == nil {
			t.Fatalf("expected selected row cell at column %d", columnIndex)
		}
		_, background, _ := rowCell.Style.Decompose()
		if background != expected {
			t.Fatalf(
				"expected selected row background %v at column %d, got %v",
				expected,
				columnIndex,
				background,
			)
		}
	}
}

func TestRenderTableWithWidthUsesMarkedRowColorAcrossEntireRow(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	selectedID := runtime.session.CurrentView().IDs[runtime.session.SelectedRow()]
	if err := runtime.session.HandleKey("SPACE"); err != nil {
		t.Fatalf("expected mark key to succeed: %v", err)
	}
	if err := runtime.session.HandleKey("DOWN"); err != nil {
		t.Fatalf("expected row move to succeed: %v", err)
	}
	runtime.renderTableWithWidth(compactModeWidthThreshold + 20)
	markedDataRow := rowIndexForID(runtime.session.CurrentView(), selectedID)
	if markedDataRow < 0 {
		t.Fatalf("expected marked row id to exist")
	}
	markedRow := markedDataRow + 1
	expected := markedRowBackgroundColor(runtime.theme)
	for columnIndex := 0; columnIndex < runtime.body.GetColumnCount(); columnIndex++ {
		rowCell := runtime.body.GetCell(markedRow, columnIndex)
		if rowCell == nil {
			t.Fatalf("expected marked row cell at column %d", columnIndex)
		}
		_, background, _ := rowCell.Style.Decompose()
		if background != expected {
			t.Fatalf(
				"expected marked row background %v at column %d, got %v",
				expected,
				columnIndex,
				background,
			)
		}
	}
}

func TestNewExplorerRuntimeUsesCyanFrameForActiveContentView(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	runtime := newExplorerRuntime()
	if runtime.body.GetBorderColor() != tcell.ColorAqua {
		t.Fatalf(
			"expected active content frame border color %v, got %v",
			tcell.ColorAqua,
			runtime.body.GetBorderColor(),
		)
	}
}

func TestContentFrameColorUsesWhiteWhenColorDisabled(t *testing.T) {
	if color := contentFrameColor(explorerTheme{UseColor: false}); color != tcell.ColorWhite {
		t.Fatalf("expected white border when color is disabled, got %v", color)
	}
}

func TestContentFrameColorUsesCyanWhenColorEnabled(t *testing.T) {
	if color := contentFrameColor(explorerTheme{UseColor: true}); color != tcell.ColorAqua {
		t.Fatalf("expected cyan border when color is enabled, got %v", color)
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

func TestRenderTopHeaderCenterUsesMultipleColumnsForTableHints(t *testing.T) {
	lines := strings.Split(renderTopHeaderCenter(false, false), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected multi-column legend lines, got %d (%q)", len(lines), lines)
	}
	multiColumnCount := 0
	for _, line := range lines {
		if strings.Count(line, "<") >= 2 {
			multiColumnCount++
		}
	}
	if multiColumnCount < 2 {
		t.Fatalf("expected at least two multi-column hint rows, got %d (%q)", multiColumnCount, lines)
	}
}

func TestRenderTopHeaderCenterUsesLogNavigationLegendInLogView(t *testing.T) {
	lines := strings.Split(renderTopHeaderCenter(true, false), "\n")
	want := []string{
		"<g> Top         <G> Bottom",
		"<PgUp/PgDn> Scroll",
		"Prompt: OFF | <q> Quit",
	}
	if len(lines) != len(want) {
		t.Fatalf("expected %d log legend lines, got %d (%q)", len(want), len(lines), lines)
	}
	for index, expected := range want {
		if lines[index] != expected {
			t.Fatalf("expected log legend line %d to be %q, got %q", index, expected, lines[index])
		}
	}
}

func TestHandlePromptDoneSwitchesHeaderLegendForLogViewAndRestoresTableLegend(t *testing.T) {
	runtime := newExplorerRuntime()

	runtime.startPrompt(":log")
	runtime.handlePromptDone(tcell.KeyEnter)
	if !runtime.logMode {
		t.Fatalf("expected runtime to enter log view mode")
	}
	runtime.renderTopHeaderWithWidth(120)
	if !strings.Contains(runtime.topHeader.GetText(false), "<PgUp/PgDn> Scroll") {
		t.Fatalf("expected log-view legend after :log command")
	}
	if !strings.Contains(runtime.topHeader.GetText(false), "Prompt: OFF | <q> Quit") {
		t.Fatalf("expected moved prompt/quit help in top header during log view")
	}

	runtime.startPrompt(":table")
	runtime.handlePromptDone(tcell.KeyEnter)
	if runtime.logMode {
		t.Fatalf("expected runtime to leave log view mode")
	}
	runtime.renderTopHeaderWithWidth(120)
	if !strings.Contains(runtime.topHeader.GetText(false), "<:> Command") {
		t.Fatalf("expected table legend after :table command")
	}
	if !strings.Contains(runtime.topHeader.GetText(false), "<Tab> Complete") {
		t.Fatalf("expected moved help hints in top header after returning to table view")
	}
}

func TestHandlePromptDoneColumnSelectionPersistsPerViewAndResets(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")

	runtime.startPrompt(":cols set NAME,POWER")
	runtime.handlePromptDone(tcell.KeyEnter)
	if runtime.body.GetCell(0, 1).Text != "NAME" || runtime.body.GetCell(0, 2).Text != "POWER" {
		t.Fatalf("expected vm columns NAME/POWER after :cols set")
	}
	for columnIndex := 0; columnIndex < runtime.body.GetColumnCount(); columnIndex++ {
		if runtime.body.GetCell(0, columnIndex).Text == "USED_CPU_PERCENT" {
			t.Fatalf("did not expect hidden vm column after :cols set")
		}
	}

	runtime.startPrompt(":host")
	runtime.handlePromptDone(tcell.KeyEnter)
	runtime.startPrompt(":vm")
	runtime.handlePromptDone(tcell.KeyEnter)
	if runtime.body.GetCell(0, 1).Text != "NAME" || runtime.body.GetCell(0, 2).Text != "POWER" {
		t.Fatalf("expected per-view vm column selection to persist across view switches")
	}

	runtime.startPrompt(":cols reset")
	runtime.handlePromptDone(tcell.KeyEnter)
	runtime.renderTableWithWidth(500)
	foundUsedCPU := false
	for columnIndex := 0; columnIndex < runtime.body.GetColumnCount(); columnIndex++ {
		if runtime.body.GetCell(0, columnIndex).Text == "USED_CPU_PERCENT" {
			foundUsedCPU = true
		}
	}
	if !foundUsedCPU {
		t.Fatalf("expected vm full columns restored after :cols reset")
	}
}

func TestRenderTopHeaderRightUsesMultilineASCIILogoBlock(t *testing.T) {
	lines := strings.Split(renderTopHeaderRight(), "\n")
	want := []string{
		"          .------------.        ",
		"       .-'   +------+   '-.     ",
		"     .'    /|      /|      '.   ",
		"    /     +------+ |         \\  ",
		"    \\     | +----|-+        /   ",
		"     '.   |/     |/      .-'    ",
		"       '-. +------+   .-'       ",
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
		"CPU: 63%(+)",
		"MEM: 58%(-)",
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

func TestFormatMetricWithTrendSupportsPositiveNegativeAndFlatValues(t *testing.T) {
	if value := formatMetricWithTrend(63, 1); value != "63%(+)" {
		t.Fatalf("expected positive trend value, got %q", value)
	}
	if value := formatMetricWithTrend(58, -1); value != "58%(-)" {
		t.Fatalf("expected negative trend value, got %q", value)
	}
	if value := formatMetricWithTrend(40, 0); value != "40%" {
		t.Fatalf("expected flat trend value, got %q", value)
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

func TestComputeLogViewportOffsetSupportsTopBottomAndPaging(t *testing.T) {
	totalRows := 18
	pageSize := 5
	if offset := computeLogViewportOffset(7, totalRows, pageSize, "top"); offset != 0 {
		t.Fatalf("expected top to move offset to 0, got %d", offset)
	}
	if offset := computeLogViewportOffset(0, totalRows, pageSize, "bottom"); offset != 13 {
		t.Fatalf("expected bottom to move offset to 13, got %d", offset)
	}
	if offset := computeLogViewportOffset(0, totalRows, pageSize, "pagedown"); offset != 5 {
		t.Fatalf("expected pagedown from 0 to move offset to 5, got %d", offset)
	}
	if offset := computeLogViewportOffset(12, totalRows, pageSize, "pagedown"); offset != 13 {
		t.Fatalf("expected pagedown to clamp at bottom offset 13, got %d", offset)
	}
	if offset := computeLogViewportOffset(13, totalRows, pageSize, "pageup"); offset != 8 {
		t.Fatalf("expected pageup from bottom to move offset to 8, got %d", offset)
	}
	if offset := computeLogViewportOffset(3, totalRows, pageSize, "pageup"); offset != 0 {
		t.Fatalf("expected pageup to clamp at top offset 0, got %d", offset)
	}
}

func TestHandleGlobalKeyLogViewportControlsMoveToExpectedOffsets(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.startPrompt(":log")
	runtime.handlePromptDone(tcell.KeyEnter)
	runtime.logEntries = make([]runtimeLogEntry, 24)
	for index := range runtime.logEntries {
		runtime.logEntries[index] = runtimeLogEntry{
			Timestamp: "2026-02-16T12:00:00Z",
			Level:     "INFO",
			Message:   "log row",
		}
	}
	runtime.body.SetRect(0, 0, 120, 10)
	runtime.renderTableWithWidth(120)

	maxOffset := runtime.logViewportMaxOffset(120)
	pageSize := runtime.logViewportPageSize()
	if maxOffset <= 0 || pageSize <= 0 {
		t.Fatalf("expected positive viewport size and max offset, got max=%d page=%d", maxOffset, pageSize)
	}

	runtime.body.SetOffset(0, 0)
	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone))
	downOffset, _ := runtime.body.GetOffset()
	if downOffset != minInt(pageSize, maxOffset) {
		t.Fatalf("expected pagedown offset %d, got %d", minInt(pageSize, maxOffset), downOffset)
	}

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone))
	bottomOffset, _ := runtime.body.GetOffset()
	if bottomOffset != maxOffset {
		t.Fatalf("expected end key to jump to bottom offset %d, got %d", maxOffset, bottomOffset)
	}

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone))
	topOffset, _ := runtime.body.GetOffset()
	if topOffset != 0 {
		t.Fatalf("expected home key to jump to top offset 0, got %d", topOffset)
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

func TestReadThemeUsesScreenshotPalettePreset(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	theme := readTheme()
	if theme.HeaderBackground != tcell.ColorAqua {
		t.Fatalf("expected cyan table header background, got %v", theme.HeaderBackground)
	}
	if theme.CanvasBackground != tcell.ColorBlack {
		t.Fatalf("expected black canvas background, got %v", theme.CanvasBackground)
	}
	if theme.HeaderAccentLeft != "yellow" {
		t.Fatalf("expected yellow left header accent, got %q", theme.HeaderAccentLeft)
	}
	if theme.HeaderAccentCenter != "aqua" {
		t.Fatalf("expected cyan center header accent, got %q", theme.HeaderAccentCenter)
	}
	if theme.HeaderAccentRight != "fuchsia" {
		t.Fatalf("expected magenta right header accent, got %q", theme.HeaderAccentRight)
	}
}

func TestReadThemeUsesYellowSelectionHighlights(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	theme := readTheme()
	if theme.RowSelected != tcell.ColorYellow {
		t.Fatalf("expected selected row highlight yellow, got %v", theme.RowSelected)
	}
	if theme.RowMarked == theme.RowSelected {
		t.Fatalf(
			"expected marked row highlight to differ from selected row highlight, got %v",
			theme.RowMarked,
		)
	}
}

func TestRenderTopHeaderWithWidthUsesScreenshotAccentColors(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	runtime := newExplorerRuntime()
	runtime.renderTopHeaderWithWidth(120)
	text := runtime.topHeader.GetText(false)
	if !strings.Contains(text, "[yellow]") {
		t.Fatalf("expected yellow accent in top header, got %q", text)
	}
	if !strings.Contains(text, "[aqua]") {
		t.Fatalf("expected cyan accent in top header, got %q", text)
	}
	if !strings.Contains(text, "[fuchsia]") {
		t.Fatalf("expected magenta accent in top header, got %q", text)
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

func TestRenderTopHeaderCenterOmitsClockForEventDrivenRedraw(t *testing.T) {
	footer := renderTopHeaderCenter(false, true)
	clock := regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}\b`)
	if clock.MatchString(footer) {
		t.Fatalf("expected top-header help without realtime clock: %q", footer)
	}
}

func TestRenderTopHeaderWithWidthCollapsesCenterLegendBeforeHidingLogo(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.renderTopHeaderWithWidth(90)
	text := runtime.topHeader.GetText(false)
	if !strings.Contains(text, "Context: vc-primary") {
		t.Fatalf("expected left metadata to remain visible at compact width, got %q", text)
	}
	if strings.Contains(text, "<Tab> Complete") {
		t.Fatalf("expected center legend to collapse before logo hide, got %q", text)
	}
	if !strings.Contains(text, "status: ready") {
		t.Fatalf("expected compact status hint in center header, got %q", text)
	}
	if !strings.Contains(text, "+------+ |") {
		t.Fatalf("expected logo body to remain visible before hide threshold, got %q", text)
	}
}

func TestRenderTopHeaderWithWidthHidesLogoAfterCenterCollapse(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.renderTopHeaderWithWidth(70)
	text := runtime.topHeader.GetText(false)
	if !strings.Contains(text, "Context: vc-primary") {
		t.Fatalf("expected left metadata to remain visible at narrow width, got %q", text)
	}
	if strings.Contains(text, ".------------.") || strings.Contains(text, "+------+") {
		t.Fatalf("expected logo to hide at narrow width, got %q", text)
	}
	if !strings.Contains(runtime.body.GetTitle(), "VM(all)") {
		t.Fatalf("expected active view title to remain visible, got %q", runtime.body.GetTitle())
	}
}

func TestRuntimeActionExecutorRoutesVMPowerLifecycleActions(t *testing.T) {
	executor := &runtimeActionExecutor{}
	cases := []struct {
		action string
		method string
	}{
		{action: "power-on", method: "power_on"},
		{action: "power-off", method: "power_off"},
		{action: "reset", method: "reset"},
		{action: "suspend", method: "suspend"},
	}
	for _, tc := range cases {
		if err := executor.Execute(tui.ResourceVM, tc.action, []string{"vm-a"}); err != nil {
			t.Fatalf("Execute returned error for %s: %v", tc.action, err)
		}
		if !strings.Contains(executor.last, "method="+tc.method) {
			t.Fatalf("expected method routing %q in %q", tc.method, executor.last)
		}
	}
}

func TestExecutePromptCommandDatastoreEvacuateReportsMigratedVMCount(t *testing.T) {
	session := tui.NewSession(defaultCatalog())
	if err := session.ExecuteCommand(":datastore"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
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
		"!evacuate",
	)
	if !keepRunning {
		t.Fatalf("expected first evacuate command to keep runtime alive")
	}
	if !strings.Contains(message, "ERR_CONFIRMATION_REQUIRED") {
		t.Fatalf("expected confirmation-required error for first evacuate, got %q", message)
	}

	message, keepRunning = executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		nil,
		"!evacuate",
	)
	if !keepRunning {
		t.Fatalf("expected second evacuate command to keep runtime alive")
	}
	if !strings.Contains(message, "migrated_vm_count=") {
		t.Fatalf("expected migrate count reporting in status, got %q", message)
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

func TestRuntimeLoadsAliasRegistryAtStartupForMultiTokenArgs(t *testing.T) {
	aliasPath := filepath.Join(t.TempDir(), "aliases.yaml")
	content := "go-host: :host\n"
	if err := os.WriteFile(aliasPath, []byte(content), 0o600); err != nil {
		t.Fatalf("expected alias file write to succeed: %v", err)
	}
	t.Setenv(aliasRegistryEnvPath, aliasPath)
	runtime := newExplorerRuntime()

	message, keepRunning := executePromptCommand(
		&runtime.session,
		&runtime.promptState,
		runtime.actionExec,
		&runtime.contexts,
		&runtime.aliasRegistry,
		":go-host owner=team-a cluster=prod",
	)
	if !keepRunning {
		t.Fatalf("expected alias command to keep runtime alive")
	}
	if message != "view: host" {
		t.Fatalf("expected host view status from startup-loaded alias, got %q", message)
	}
	if runtime.session.CurrentView().Resource != tui.ResourceHost {
		t.Fatalf("expected startup-loaded alias registry to resolve host view command")
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

func TestExecutePromptCommandRegexFilterRejectsInvalidPatternAndKeepsPriorFilter(t *testing.T) {
	session := tui.NewSession(
		tui.Catalog{
			VMs: []tui.VMRow{{Name: "vm-a"}, {Name: "vm-b"}, {Name: "db-a"}},
		},
	)
	promptState := tui.NewPromptState(20)
	executor := &runtimeActionExecutor{}
	contexts := newRuntimeContextManager()

	okMessage, keepRunning := executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		nil,
		"/^vm-",
	)
	if !keepRunning {
		t.Fatalf("expected regex filter command to keep runtime alive")
	}
	if okMessage != "filter: ^vm-" {
		t.Fatalf("expected regex filter status, got %q", okMessage)
	}
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected regex filter to scope rows, got %d", len(session.CurrentView().Rows))
	}

	errMessage, keepRunning := executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		nil,
		"/[",
	)
	if !keepRunning {
		t.Fatalf("expected invalid regex command to keep runtime alive")
	}
	if !strings.Contains(errMessage, "command error:") {
		t.Fatalf("expected command error status for invalid regex, got %q", errMessage)
	}
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected invalid regex to preserve prior filtered rows")
	}
}

func TestExecutePromptCommandInverseRegexFilterExcludesMatches(t *testing.T) {
	session := tui.NewSession(
		tui.Catalog{
			VMs: []tui.VMRow{{Name: "vm-a"}, {Name: "vm-b"}, {Name: "db-a"}},
		},
	)
	promptState := tui.NewPromptState(20)
	executor := &runtimeActionExecutor{}
	contexts := newRuntimeContextManager()

	message, keepRunning := executePromptCommand(
		&session,
		&promptState,
		executor,
		&contexts,
		nil,
		"/!^vm-",
	)
	if !keepRunning {
		t.Fatalf("expected inverse regex filter command to keep runtime alive")
	}
	if message != "filter: !^vm-" {
		t.Fatalf("expected inverse regex filter status, got %q", message)
	}
	if len(session.CurrentView().Rows) != 1 || session.CurrentView().Rows[0][0] != "db-a" {
		t.Fatalf("expected inverse regex filter to exclude matching vm rows")
	}
}

func TestExecutePromptCommandTagFilterRequiresAllPairs(t *testing.T) {
	session := tui.NewSession(
		tui.Catalog{
			Hosts: []tui.HostRow{
				{Name: "host-a", Tags: "env=prod,tier=gold"},
				{Name: "host-b", Tags: "env=prod,tier=silver"},
				{Name: "host-c", Tags: "env=dev,tier=gold"},
			},
		},
	)
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
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
		"/-t env=prod,tier=gold",
	)
	if !keepRunning {
		t.Fatalf("expected tag filter command to keep runtime alive")
	}
	if message != "filter: -t env=prod,tier=gold" {
		t.Fatalf("expected tag filter status, got %q", message)
	}
	if len(session.CurrentView().Rows) != 1 || session.CurrentView().Rows[0][0] != "host-a" {
		t.Fatalf("expected tag filter to include only rows matching all requested pairs")
	}
}

func TestExecutePromptCommandFuzzyFilterRanksMatches(t *testing.T) {
	session := tui.NewSession(
		tui.Catalog{
			Hosts: []tui.HostRow{
				{Name: "prod-edge", Tags: "env=prod"},
				{Name: "host-prod", Tags: "env=prod"},
				{Name: "dev-a", Tags: "env=dev"},
			},
		},
	)
	if err := session.ExecuteCommand(":host"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
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
		"/-f prod",
	)
	if !keepRunning {
		t.Fatalf("expected fuzzy filter command to keep runtime alive")
	}
	if message != "filter: -f prod" {
		t.Fatalf("expected fuzzy filter status, got %q", message)
	}
	if len(session.CurrentView().Rows) != 2 || session.CurrentView().Rows[0][0] != "prod-edge" {
		t.Fatalf("expected fuzzy filter to rank highest score first")
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
	if !strings.Contains(message, "action_error code=ERR_READ_ONLY") {
		t.Fatalf("expected standardized action error code, got %q", message)
	}
	if !strings.Contains(message, "message=\"read-only mode\"") {
		t.Fatalf("expected standardized action error message, got %q", message)
	}
	if !strings.Contains(message, "entity=\"vm-a\"") {
		t.Fatalf("expected standardized action error entity, got %q", message)
	}
	if !strings.Contains(message, "retryable=false") {
		t.Fatalf("expected standardized action error retryable flag, got %q", message)
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
	if !strings.Contains(runtime.breadcrumb.GetText(true), "home > dc-1 > cluster-east > esxi-01") {
		t.Fatalf("expected breadcrumb text to render active view")
	}
}

func TestNewExplorerRuntimeCrumbslessOmitsBreadcrumbWidget(t *testing.T) {
	runtime := newExplorerRuntimeWithRenderOptions(false, "host", false, true)
	if runtime.breadcrumb.GetText(true) != "" {
		t.Fatalf("expected crumbsless runtime to keep breadcrumb empty")
	}
	if runtime.layout.GetItemCount() != 3 {
		t.Fatalf("expected compact layout with top header, table, and prompt only")
	}
}

func TestNewExplorerRuntimeRemovesBottomHelpBarFromLayout(t *testing.T) {
	runtime := newExplorerRuntime()
	if runtime.layout.GetItemCount() != 3 {
		t.Fatalf("expected runtime layout without standalone breadcrumb/status/footer widgets")
	}
}

func TestNewExplorerRuntimeRendersPathAndStatusBelowCenterShortcuts(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.renderTopHeaderWithWidth(200)
	header := strings.ToLower(runtime.topHeader.GetText(false))
	if !strings.Contains(header, "<shift+o> sort") {
		t.Fatalf("expected center shortcuts in top header, got %q", header)
	}
	sortIndex := strings.Index(header, "<shift+o> sort")
	pathIndex := strings.Index(header, "path: home > dc-1 > cluster-east > esxi-01 > vm-a")
	statusIndex := strings.Index(header, "status: ready")
	if pathIndex == -1 || statusIndex == -1 {
		t.Fatalf("expected compact path/status in top header, got %q", header)
	}
	if pathIndex <= sortIndex || statusIndex <= sortIndex {
		t.Fatalf("expected path/status below center shortcuts, got %q", header)
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

func TestDescribeDrawerKeepsTableNavigationActive(t *testing.T) {
	runtime := newExplorerRuntimeWithStartupCommand(false, "vm")
	startRow := runtime.session.SelectedRow()

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	if !runtime.isDescribePanelOpen() {
		t.Fatalf("expected describe drawer to open on d")
	}

	runtime.handleGlobalKey(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	if runtime.session.SelectedRow() == startRow {
		t.Fatalf("expected table navigation to remain active while describe drawer is open")
	}
}
