// Path: cmd/hypersphere/explorer_tui.go
// Description: Run a full-screen real-time TUI explorer using tview/tcell with k9s-inspired interactions.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/takelley1/hypersphere/internal/tui"
)

const defaultPromptHistorySize = 200
const minAutosizeColumnWidth = 3
const compactModeWidthThreshold = 15
const fixedTableColumns = 1
const defaultTopHeaderWidth = 120
const topHeaderPanelHeight = 7
const defaultCPUPercent = 63
const defaultCPUTrend = 1
const defaultMEMPercent = 58
const defaultMEMTrend = -1
const logTimestampWidth = 20
const logLevelWidth = 5
const logMessageMinWidth = 16
const compactHeaderCollapseWidth = 100
const compactHeaderHideLogoWidth = 80
const describeDrawerHeight = 9

var compactColumnsByResource = map[tui.Resource][]string{
	tui.ResourceVM:        {"NAME", "POWER", "ATTACHED_STORAGE"},
	tui.ResourceLUN:       {"NAME", "DATASTORE", "USED_GB"},
	tui.ResourceCluster:   {"NAME", "HOSTS", "VMS"},
	tui.ResourceTask:      {"ENTITY", "ACTION", "STATE"},
	tui.ResourceHost:      {"NAME", "CLUSTER", "CONNECTION"},
	tui.ResourceDatastore: {"NAME", "CLUSTER", "FREE_GB"},
}

type explorerRuntime struct {
	app            *tview.Application
	session        tui.Session
	promptState    tui.PromptState
	actionExec     *runtimeActionExecutor
	contexts       runtimeContextManager
	headless       bool
	crumbsless     bool
	theme          explorerTheme
	pages          *tview.Pages
	layout         *tview.Flex
	contentPane    *tview.Flex
	helpModal      *tview.Modal
	aliasModal     *tview.Modal
	topHeader      *tview.TextView
	body           *tview.Table
	describeDrawer *tview.TextView
	breadcrumb     *tview.TextView
	status         *tview.TextView
	prompt         *tview.InputField
	aliasEntries   []string
	promptMode     bool
	helpOpen       bool
	aliasOpen      bool
	describeOpen   bool
	helpText       string
	describeText   string
	lastWidth      int
	wideColumns    bool
	headerVisible  bool
	logMode        bool
	logObjectPath  string
	logTarget      string
	logEntries     []runtimeLogEntry
}

type runtimeActionExecutor struct {
	last string
}

type runtimeLogEntry struct {
	Timestamp string
	Level     string
	Message   string
}

type runtimeContextConnector interface {
	List() []string
	Active() string
	Switch(name string) error
}

type runtimeContextManager struct {
	connector runtimeContextConnector
}

type inMemoryContextConnector struct {
	active    string
	endpoints []string
}

type explorerTheme struct {
	UseColor           bool
	CanvasBackground   tcell.Color
	HeaderText         tcell.Color
	HeaderBackground   tcell.Color
	HeaderAccentLeft   string
	HeaderAccentCenter string
	HeaderAccentRight  string
	EvenRowText        tcell.Color
	OddRowText         tcell.Color
	RowHealthy         tcell.Color
	RowDegraded        tcell.Color
	RowFaulted         tcell.Color
	RowSelected        tcell.Color
	RowMarked          tcell.Color
	RowMarkedSelected  tcell.Color
	StatusError        string
}

func (r *runtimeActionExecutor) Execute(resource tui.Resource, action string, ids []string) error {
	r.last = fmt.Sprintf(
		"vmware-api action=%s resource=%s targets=%s",
		action,
		resource,
		strings.Join(ids, ","),
	)
	return nil
}

func newRuntimeContextManager() runtimeContextManager {
	return runtimeContextManager{
		connector: &inMemoryContextConnector{
			active:    "vc-primary",
			endpoints: []string{"vc-primary", "vc-lab"},
		},
	}
}

func (m runtimeContextManager) List() []string {
	return m.connector.List()
}

func (m runtimeContextManager) Active() string {
	return m.connector.Active()
}

func (m runtimeContextManager) Switch(name string) error {
	return m.connector.Switch(name)
}

func (c *inMemoryContextConnector) List() []string {
	values := append([]string{}, c.endpoints...)
	return values
}

func (c *inMemoryContextConnector) Active() string {
	return c.active
}

func (c *inMemoryContextConnector) Switch(name string) error {
	for _, endpoint := range c.endpoints {
		if endpoint == name {
			c.active = name
			return nil
		}
	}
	return fmt.Errorf("unknown context: %s", name)
}

func runExplorerWorkflow(
	output io.Writer,
	readOnly bool,
	startupCommand string,
	headless bool,
	crumbsless bool,
) {
	runtime := newExplorerRuntimeWithRenderOptions(readOnly, startupCommand, headless, crumbsless)
	if err := runtime.run(); err != nil {
		_, _ = fmt.Fprintf(output, "tui error: %v\n", err)
	}
}

func newExplorerRuntime() explorerRuntime {
	return newExplorerRuntimeWithOptions(false, "", false)
}

func newExplorerRuntimeWithReadOnly(readOnly bool) explorerRuntime {
	return newExplorerRuntimeWithRenderOptions(readOnly, "", false, false)
}

func newExplorerRuntimeWithStartupCommand(readOnly bool, startupCommand string) explorerRuntime {
	return newExplorerRuntimeWithRenderOptions(readOnly, startupCommand, false, false)
}

func newExplorerRuntimeWithOptions(
	readOnly bool,
	startupCommand string,
	headless bool,
) explorerRuntime {
	return newExplorerRuntimeWithRenderOptions(readOnly, startupCommand, headless, false)
}

func newExplorerRuntimeWithRenderOptions(
	readOnly bool,
	startupCommand string,
	headless bool,
	crumbsless bool,
) explorerRuntime {
	runtime := explorerRuntime{
		app:            tview.NewApplication(),
		session:        tui.NewSession(defaultCatalog()),
		promptState:    tui.NewPromptState(defaultPromptHistorySize),
		actionExec:     &runtimeActionExecutor{},
		contexts:       newRuntimeContextManager(),
		headless:       headless,
		crumbsless:     crumbsless,
		theme:          readTheme(),
		pages:          tview.NewPages(),
		helpModal:      tview.NewModal(),
		aliasModal:     tview.NewModal(),
		topHeader:      tview.NewTextView(),
		body:           tview.NewTable(),
		describeDrawer: tview.NewTextView(),
		breadcrumb:     tview.NewTextView(),
		status:         tview.NewTextView(),
		prompt:         tview.NewInputField(),
		wideColumns:    true,
		headerVisible:  !headless,
		logEntries:     defaultRuntimeLogEntries(),
	}
	runtime.session.SetReadOnly(readOnly)
	message := startupCommandStatus(&runtime.session, startupCommand)
	runtime.configureWidgets()
	runtime.configureHandlers()
	runtime.render(message)
	return runtime
}

func startupCommandStatus(session *tui.Session, startupCommand string) string {
	trimmed := strings.TrimSpace(startupCommand)
	if trimmed == "" {
		return "ready"
	}
	if !strings.HasPrefix(trimmed, ":") {
		trimmed = ":" + trimmed
	}
	return statusFromError(session.ExecuteCommand(trimmed), "view: "+strings.TrimPrefix(trimmed, ":"))
}

func (r *explorerRuntime) configureWidgets() {
	r.body.SetSelectable(true, true)
	r.body.SetFixed(fixedHeaderRows(r.tableHeaderVisible()), fixedTableColumns)
	r.body.SetBorders(false)
	r.body.SetSeparator(' ')
	r.body.SetBorder(true)
	r.body.SetBorderColor(contentFrameColor(r.theme))
	r.body.SetTitleAlign(tview.AlignCenter)
	r.body.SetTitleColor(contentFrameColor(r.theme))
	r.body.SetTitle(composeTableTitle(r.session.CurrentView(), false, false))
	r.body.SetSelectedStyle(
		tcell.StyleDefault.
			Background(selectedRowBackgroundColor(r.theme)).
			Foreground(tcell.ColorBlack),
	)
	r.topHeader.SetBackgroundColor(r.theme.CanvasBackground)
	r.body.SetBackgroundColor(r.theme.CanvasBackground)
	r.breadcrumb.SetDynamicColors(true)
	r.breadcrumb.SetBorder(true)
	r.breadcrumb.SetBackgroundColor(r.theme.CanvasBackground)
	r.breadcrumb.SetTitle(" Breadcrumbs ")
	r.status.SetDynamicColors(true)
	r.status.SetBorder(true)
	r.status.SetBackgroundColor(r.theme.CanvasBackground)
	r.status.SetTitle(" Status ")
	r.prompt.SetLabel("Command: ")
	r.prompt.SetFieldBackgroundColor(r.theme.CanvasBackground)
	applyPromptValidationState(r.prompt, "")
	r.helpModal.SetBorder(true)
	r.helpModal.SetTitle(" Keymap Help ")
	r.aliasEntries = tui.ResourceCommandAliases()
	r.aliasModal.SetBorder(true)
	r.aliasModal.SetTitle(" Alias Palette ")
	r.aliasModal.SetText("Select a resource alias")
	r.aliasModal.AddButtons(r.aliasEntries)
	r.aliasModal.SetDoneFunc(r.handleAliasSelection)
	r.describeDrawer.SetBorder(true)
	r.describeDrawer.SetTitle(" Details Drawer ")
	r.describeDrawer.SetBackgroundColor(r.theme.CanvasBackground)
	r.describeDrawer.SetDynamicColors(true)
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.body, 0, 1, true).
		AddItem(r.describeDrawer, 0, 0, false)
	r.contentPane = content
	r.topHeader.SetDynamicColors(true)
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.topHeader, topHeaderPanelHeight, 0, false).
		AddItem(r.contentPane, 0, 1, true).
		AddItem(r.prompt, 1, 0, false)
	r.layout = layout
	r.layout.SetBackgroundColor(r.theme.CanvasBackground)
	r.pages.AddPage("main", layout, true, true)
	r.pages.AddPage("help", r.helpModal, true, false)
	r.pages.AddPage("alias", r.aliasModal, true, false)
	r.pages.SetBackgroundColor(r.theme.CanvasBackground)
	r.app.SetRoot(r.pages, true)
	r.app.SetFocus(r.body)
}

func (r *explorerRuntime) configureHandlers() {
	r.app.SetInputCapture(r.handleGlobalKey)
	r.app.SetBeforeDrawFunc(r.handleScreenResize)
	r.body.SetSelectionChangedFunc(r.handleTableSelectionChanged)
	r.prompt.SetDoneFunc(r.handlePromptDone)
	r.prompt.SetInputCapture(r.handlePromptHistory)
	r.prompt.SetChangedFunc(r.handlePromptChanged)
}

func (r *explorerRuntime) handleTableSelectionChanged(row int, column int) {
	selectedRow := row
	if r.tableHeaderVisible() {
		selectedRow--
	}
	r.session.SetSelection(selectedRow, column-1)
}

func (r *explorerRuntime) handleScreenResize(screen tcell.Screen) bool {
	width, _ := screen.Size()
	if width <= 0 || width == r.lastWidth {
		return false
	}
	r.lastWidth = width
	tableWidth := tableAvailableWidth(r.body)
	if tableWidth <= 0 {
		tableWidth = width
	}
	r.renderTopHeaderWithWidth(width)
	r.renderTableWithWidth(tableWidth)
	return false
}

func (r *explorerRuntime) run() error {
	return r.app.Run()
}

func (r *explorerRuntime) handleGlobalKey(evt *tcell.EventKey) *tcell.EventKey {
	if r.helpOpen {
		if evt.Key() == tcell.KeyEscape {
			r.closeHelpModal()
		}
		return nil
	}
	if r.aliasOpen {
		if evt.Key() == tcell.KeyEscape {
			r.closeAliasPalette()
		}
		return nil
	}
	if r.describeOpen && evt.Key() == tcell.KeyEscape {
		r.closeDescribePanel()
		return nil
	}
	if r.promptMode {
		return evt
	}
	if evt.Key() == tcell.KeyCtrlA {
		r.openAliasPalette()
		return nil
	}
	if evt.Key() == tcell.KeyRune && evt.Rune() == '?' {
		r.openHelpModal()
		return nil
	}
	if isDescribeEvent(evt) {
		r.openDescribePanel()
		return nil
	}
	if isPromptActivation(evt) {
		r.startPrompt(string(evt.Rune()))
		return nil
	}
	if isQuitEvent(evt) {
		r.app.Stop()
		return nil
	}
	if r.handleLogViewportControl(evt) {
		return nil
	}
	command, ok := eventToHotKey(evt)
	if !ok {
		return evt
	}
	if r.handleRuntimeToggle(command) {
		return nil
	}
	r.emitStatus(r.session.HandleKey(command))
	r.render("")
	return nil
}

func (r *explorerRuntime) handleLogViewportControl(evt *tcell.EventKey) bool {
	command, ok := logViewportCommand(evt)
	if !ok || !r.logMode {
		return false
	}
	rowOffset, columnOffset := r.body.GetOffset()
	availableWidth := tableAvailableWidth(r.body)
	nextOffset := computeLogViewportOffset(
		rowOffset,
		r.logScrollableRowCount(availableWidth),
		r.logViewportPageSize(),
		command,
	)
	r.body.SetOffset(nextOffset, columnOffset)
	return true
}

func logViewportCommand(evt *tcell.EventKey) (string, bool) {
	if evt == nil {
		return "", false
	}
	switch evt.Key() {
	case tcell.KeyPgUp:
		return "pageup", true
	case tcell.KeyPgDn:
		return "pagedown", true
	case tcell.KeyHome:
		return "top", true
	case tcell.KeyEnd:
		return "bottom", true
	}
	if evt.Key() == tcell.KeyRune && evt.Modifiers() == tcell.ModNone && evt.Rune() == 'g' {
		return "top", true
	}
	if evt.Key() == tcell.KeyRune && evt.Rune() == 'G' {
		return "bottom", true
	}
	return "", false
}

func (r *explorerRuntime) handleRuntimeToggle(command string) bool {
	switch command {
	case "CTRL+W":
		r.toggleWideColumns()
	case "CTRL+E":
		r.toggleHeaderVisibility()
	default:
		return false
	}
	r.render("")
	return true
}

func (r *explorerRuntime) toggleWideColumns() {
	selectedID := selectedIDForRow(r.session.CurrentView(), r.session.SelectedRow())
	r.wideColumns = !r.wideColumns
	if selectedID == "" {
		return
	}
	nextRow := rowIndexForID(r.session.CurrentView(), selectedID)
	if nextRow >= 0 {
		r.session.SetSelection(nextRow, r.session.SelectedColumn())
	}
}

func (r *explorerRuntime) toggleHeaderVisibility() {
	if r.headless {
		return
	}
	selectedID := selectedIDForRow(r.session.CurrentView(), r.session.SelectedRow())
	r.headerVisible = !r.headerVisible
	if selectedID == "" {
		return
	}
	nextRow := rowIndexForID(r.session.CurrentView(), selectedID)
	if nextRow >= 0 {
		r.session.SetSelection(nextRow, r.session.SelectedColumn())
	}
}

func (r *explorerRuntime) openHelpModal() {
	r.helpText = helpModalText(r.session.CurrentView())
	r.helpModal.SetText(r.helpText)
	r.pages.ShowPage("help")
	r.helpOpen = true
}

func (r *explorerRuntime) closeHelpModal() {
	r.pages.HidePage("help")
	r.helpOpen = false
}

func (r *explorerRuntime) isHelpModalOpen() bool {
	return r.helpOpen
}

func (r *explorerRuntime) openAliasPalette() {
	r.pages.ShowPage("alias")
	r.aliasOpen = true
}

func (r *explorerRuntime) closeAliasPalette() {
	r.pages.HidePage("alias")
	r.aliasOpen = false
}

func (r *explorerRuntime) isAliasPaletteOpen() bool {
	return r.aliasOpen
}

func (r *explorerRuntime) openDescribePanel() {
	details, err := r.session.SelectedResourceDetails()
	if err != nil {
		r.emitStatus(err)
		return
	}
	r.describeText = renderResourceDetails(details)
	r.describeDrawer.SetText(r.describeText)
	r.contentPane.ResizeItem(r.describeDrawer, describeDrawerHeight, 0)
	r.describeOpen = true
	r.app.SetFocus(r.body)
}

func (r *explorerRuntime) closeDescribePanel() {
	r.contentPane.ResizeItem(r.describeDrawer, 0, 0)
	r.describeOpen = false
	r.app.SetFocus(r.body)
	row, column := selectionForTable(r.session, r.tableHeaderVisible())
	r.body.Select(row, column)
}

func (r *explorerRuntime) isDescribePanelOpen() bool {
	return r.describeOpen
}

func (r *explorerRuntime) handleAliasSelection(_ int, buttonLabel string) {
	r.closeAliasPalette()
	message := statusFromError(
		r.session.ExecuteCommand(buttonLabel),
		"view: "+strings.TrimPrefix(buttonLabel, ":"),
	)
	r.render(message)
}

func (r *explorerRuntime) handlePromptDone(key tcell.Key) {
	if key == tcell.KeyEscape {
		r.endPrompt()
		return
	}
	if key != tcell.KeyEnter {
		return
	}
	if message, handled := r.handleLocalPromptCommand(strings.TrimSpace(r.prompt.GetText())); handled {
		r.endPrompt()
		r.render(message)
		return
	}
	message, keepRunning := executePromptCommand(
		&r.session,
		&r.promptState,
		r.actionExec,
		&r.contexts,
		nil,
		r.prompt.GetText(),
	)
	r.endPrompt()
	r.render(message)
	if !keepRunning {
		r.app.Stop()
	}
}

func (r *explorerRuntime) handleLocalPromptCommand(line string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 {
		return "", false
	}
	switch strings.ToLower(fields[0]) {
	case ":log", ":logs":
		r.logMode = true
		r.logObjectPath, r.logTarget = resolveLogCommandArguments(r.session, fields[1:])
		return "view: logs", true
	case ":table":
		r.logMode = false
		r.logObjectPath = ""
		r.logTarget = ""
		return fmt.Sprintf("view: %s", r.session.CurrentView().Resource), true
	case ":cols", ":columns":
		return handleColumnsPromptCommand(&r.session, fields, line), true
	default:
		return "", false
	}
}

func handleColumnsPromptCommand(session *tui.Session, fields []string, raw string) string {
	if len(fields) == 1 || strings.EqualFold(fields[1], "list") {
		return "columns: " + strings.Join(session.VisibleColumns(), ",")
	}
	action := strings.ToLower(strings.TrimSpace(fields[1]))
	if action == "reset" {
		return statusFromError(session.ResetVisibleColumns(), "columns: reset")
	}
	if action != "set" {
		return fmt.Sprintf("command error: %s", tui.ErrInvalidColumns)
	}
	selection := parseColumnSelection(raw)
	return statusFromError(session.SetVisibleColumns(selection), "columns: "+strings.Join(selection, ","))
}

func parseColumnSelection(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	lower := strings.ToLower(trimmed)
	index := strings.Index(lower, " set ")
	if index == -1 {
		return nil
	}
	values := strings.Split(trimmed[index+len(" set "):], ",")
	selection := make([]string, 0, len(values))
	for _, value := range values {
		column := strings.TrimSpace(value)
		if column != "" {
			selection = append(selection, column)
		}
	}
	return selection
}

func resolveLogCommandArguments(
	session tui.Session,
	fields []string,
) (string, string) {
	objectPath := defaultLogObjectPath(session)
	target := ""
	for _, field := range fields {
		trimmed := strings.TrimSpace(field)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "target=") {
			target = strings.TrimSpace(trimmed[len("target="):])
			continue
		}
		if !strings.Contains(trimmed, "=") {
			objectPath = trimmed
		}
	}
	return objectPath, target
}

func defaultLogObjectPath(session tui.Session) string {
	view := session.CurrentView()
	selected := session.SelectedRow()
	if selected >= 0 && selected < len(view.IDs) {
		return fmt.Sprintf("%s/%s", view.Resource, view.IDs[selected])
	}
	return string(view.Resource)
}

func (r *explorerRuntime) handlePromptHistory(evt *tcell.EventKey) *tcell.EventKey {
	if evt.Key() == tcell.KeyTab {
		r.completePromptFromSuggestions()
		return nil
	}
	if evt.Key() == tcell.KeyUp {
		r.fillPromptFromHistory("up")
		return nil
	}
	if evt.Key() == tcell.KeyDown {
		r.fillPromptFromHistory("down")
		return nil
	}
	return evt
}

func (r *explorerRuntime) completePromptFromSuggestions() {
	value, status, changed := applyPromptCompletion(
		&r.promptState,
		r.session.CurrentView(),
		r.prompt.GetText(),
	)
	if !changed {
		return
	}
	r.prompt.SetText(value)
	r.status.SetText(status)
}

func (r *explorerRuntime) fillPromptFromHistory(direction string) {
	entry, ok := readHistoryEntry(&r.promptState, direction)
	if !ok {
		return
	}
	r.prompt.SetText(entry)
}

func (r *explorerRuntime) startPrompt(prefix string) {
	r.promptMode = true
	applyPromptValidationState(r.prompt, "")
	r.prompt.SetText(prefix)
	r.app.SetFocus(r.prompt)
}

func (r *explorerRuntime) endPrompt() {
	r.promptMode = false
	r.prompt.SetText("")
	applyPromptValidationState(r.prompt, "")
	r.app.SetFocus(r.body)
}

func (r *explorerRuntime) handlePromptChanged(text string) {
	if !r.promptMode {
		return
	}
	message := promptValidationMessage(text)
	applyPromptValidationState(r.prompt, message)
	if message == "" {
		r.status.SetText("")
		return
	}
	r.status.SetText(message)
}

func (r *explorerRuntime) emitStatus(err error) {
	if err == nil {
		return
	}
	if !r.theme.UseColor {
		r.status.SetText(err.Error())
		return
	}
	r.status.SetText(fmt.Sprintf("%s%s", r.theme.StatusError, err.Error()))
}

func (r *explorerRuntime) render(message string) {
	if message != "" {
		r.status.SetText(message)
	}
	r.renderTopHeader()
	r.renderTable()
	r.renderBreadcrumb()
}

func (r *explorerRuntime) renderTopHeader() {
	r.renderTopHeaderWithWidth(r.lastWidth)
}

func (r *explorerRuntime) renderTopHeaderWithWidth(width int) {
	if r.topHeader == nil {
		return
	}
	leftLines := strings.Split(renderTopHeaderLeft(r.contexts.Active()), "\n")
	centerLines := strings.Split(
		renderTopHeaderCenterWithContext(
			r.logMode,
			r.promptMode,
			compactTopHeaderPath(r.crumbsless, r.session.BreadcrumbPath()),
			compactTopHeaderStatus(r.status.GetText(true)),
		),
		"\n",
	)
	rightLines := strings.Split(renderTopHeaderRight(), "\n")
	centerLines, rightLines = degradeTopHeaderSections(width, centerLines, rightLines)
	headerLines := renderTopHeaderLinesWithTheme(
		normalizeTopHeaderWidth(width),
		leftLines,
		centerLines,
		rightLines,
		r.theme,
	)
	r.topHeader.SetText(strings.Join(headerLines, "\n"))
}

func degradeTopHeaderSections(
	width int,
	center []string,
	right []string,
) ([]string, []string) {
	if width < compactHeaderHideLogoWidth {
		right = nil
	}
	if width < compactHeaderCollapseWidth {
		center = compactCenterLegend(center)
	}
	return center, right
}

func compactCenterLegend(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	compact := []string{lines[0]}
	if len(lines) > 0 {
		compact = append(compact, lines[len(lines)-1])
	}
	return compact
}

func (r *explorerRuntime) renderBreadcrumb() {
	if r.crumbsless || r.breadcrumb == nil {
		return
	}
	r.breadcrumb.SetText(r.session.BreadcrumbPath())
}

func (r *explorerRuntime) renderTable() {
	r.renderTableWithWidth(tableAvailableWidth(r.body))
}

func (r *explorerRuntime) renderTableWithWidth(availableWidth int) {
	includeHeader := r.tableHeaderVisible()
	r.body.SetFixed(fixedHeaderRows(includeHeader), fixedTableColumns)
	view := viewForColumnMode(r.session.CurrentView(), r.wideColumns)
	if r.logMode {
		view = logResourceView(r.logEntries, availableWidth)
	}
	view = compactViewForWidth(view, availableWidth)
	rows := tableRows(view, r.session.IsMarked, includeHeader)
	widths := autosizedColumnWidths(view, rows, availableWidth)
	rowFillWidth := tableRowFillWidth(widths, availableWidth)
	_, columnOffset := r.body.GetOffset()
	leftOverflow, rightOverflow := tableOverflowMarkers(
		widths,
		availableWidth,
		columnOffset,
		fixedTableColumns,
	)
	r.body.SetTitle(composeRuntimeTitle(r, view, leftOverflow, rightOverflow))
	r.body.Clear()
	for rowIndex, row := range rows {
		dataRowIndex := renderedDataRowIndex(rowIndex, includeHeader)
		rowID := selectedIDForRow(view, dataRowIndex)
		isMarked := rowID != "" && r.session.IsMarked(rowID)
		isSelected := dataRowIndex == r.session.SelectedRow()
		for columnIndex, value := range row {
			cell := tview.NewTableCell(value)
			if columnIndex < len(widths) {
				cell.SetMaxWidth(widths[columnIndex])
			}
			if includeHeader && rowIndex == 0 {
				cell.SetSelectable(false)
				cell.SetTextColor(r.theme.HeaderText)
				cell.SetBackgroundColor(r.theme.HeaderBackground)
				cell.SetAttributes(tcell.AttrBold)
			} else {
				cell.SetTextColor(tableTextColor(r.theme, view, dataRowIndex, isMarked, isSelected))
				cell.SetBackgroundColor(tableRowBackgroundColor(r.theme, isMarked, isSelected))
			}
			r.body.SetCell(rowIndex, columnIndex, cell)
		}
		appendRowFillCell(r, rowIndex, len(row), rowFillWidth, includeHeader, view, dataRowIndex, isMarked, isSelected)
	}
	selectedRow, selectedColumn := selectionForRenderedView(r.session, view, includeHeader)
	r.body.Select(selectedRow, selectedColumn)
}

func appendRowFillCell(
	r *explorerRuntime,
	rowIndex int,
	columnIndex int,
	fillWidth int,
	includeHeader bool,
	view tui.ResourceView,
	dataRowIndex int,
	isMarked bool,
	isSelected bool,
) {
	if fillWidth <= 0 {
		return
	}
	fill := tview.NewTableCell(strings.Repeat(" ", fillWidth))
	fill.SetMaxWidth(fillWidth)
	fill.SetSelectable(false)
	if includeHeader && rowIndex == 0 {
		fill.SetBackgroundColor(r.theme.HeaderBackground)
	} else {
		fill.SetTextColor(tableTextColor(r.theme, view, dataRowIndex, isMarked, isSelected))
		fill.SetBackgroundColor(tableRowBackgroundColor(r.theme, isMarked, isSelected))
	}
	r.body.SetCell(rowIndex, columnIndex, fill)
}

func renderedDataRowIndex(rowIndex int, includeHeader bool) int {
	if includeHeader {
		return rowIndex - 1
	}
	return rowIndex
}

func tableRowFillWidth(widths []int, availableWidth int) int {
	remainder := availableWidth - tableRenderWidth(widths) - 1
	if remainder <= 0 {
		return 0
	}
	return remainder
}

func selectedIDForRow(view tui.ResourceView, row int) string {
	if row < 0 || row >= len(view.IDs) {
		return ""
	}
	return view.IDs[row]
}

func rowIndexForID(view tui.ResourceView, id string) int {
	for index, candidate := range view.IDs {
		if candidate == id {
			return index
		}
	}
	return -1
}

func viewForColumnMode(view tui.ResourceView, wideColumns bool) tui.ResourceView {
	if wideColumns {
		return view
	}
	columns, ok := compactColumnsByResource[view.Resource]
	if !ok {
		return view
	}
	return selectCompactColumns(view, columns)
}

func compactViewForWidth(view tui.ResourceView, availableWidth int) tui.ResourceView {
	if availableWidth <= 0 || availableWidth >= compactModeWidthThreshold {
		return view
	}
	columns, ok := compactColumnsByResource[view.Resource]
	if !ok {
		return view
	}
	return selectCompactColumns(view, columns)
}

func selectCompactColumns(view tui.ResourceView, columns []string) tui.ResourceView {
	indexes, resolved := compactColumnIndexes(view.Columns, columns)
	if len(indexes) == 0 {
		return view
	}
	rows := make([][]string, len(view.Rows))
	for rowIndex, row := range view.Rows {
		rows[rowIndex] = compactRow(row, indexes)
	}
	return tui.ResourceView{
		Resource:    view.Resource,
		Columns:     resolved,
		Rows:        rows,
		IDs:         append([]string{}, view.IDs...),
		SortHotKeys: appendSortHotKeys(view.SortHotKeys),
		Actions:     append([]string{}, view.Actions...),
	}
}

func compactColumnIndexes(all []string, wanted []string) ([]int, []string) {
	indexes := make([]int, 0, len(wanted))
	resolved := make([]string, 0, len(wanted))
	for _, name := range wanted {
		index := indexOfColumn(all, name)
		if index < 0 {
			continue
		}
		indexes = append(indexes, index)
		resolved = append(resolved, name)
	}
	return indexes, resolved
}

func compactRow(row []string, indexes []int) []string {
	compact := make([]string, len(indexes))
	for index, columnIndex := range indexes {
		if columnIndex >= 0 && columnIndex < len(row) {
			compact[index] = row[columnIndex]
		}
	}
	return compact
}

func appendSortHotKeys(values map[string]string) map[string]string {
	copyValues := make(map[string]string, len(values))
	for key, value := range values {
		copyValues[key] = value
	}
	return copyValues
}

func indexOfColumn(columns []string, name string) int {
	for index, column := range columns {
		if column == name {
			return index
		}
	}
	return -1
}

func tableAvailableWidth(table *tview.Table) int {
	_, _, width, _ := table.GetInnerRect()
	return width
}

func tableRows(
	view tui.ResourceView,
	isMarked func(string) bool,
	includeHeader bool,
) [][]string {
	size := len(view.Rows)
	if includeHeader {
		size++
	}
	rows := make([][]string, 0, size)
	if includeHeader {
		headers := append([]string{"SEL"}, view.Columns...)
		rows = append(rows, headers)
	}
	for index := 0; index < len(view.Rows); index++ {
		marker := " "
		if index < len(view.IDs) && isMarked(view.IDs[index]) {
			marker = "*"
		}
		row := append([]string{marker}, view.Rows[index]...)
		rows = append(rows, row)
	}
	return rows
}

func autosizedColumnWidths(
	view tui.ResourceView,
	rows [][]string,
	availableWidth int,
) []int {
	widths := naturalColumnWidths(rows)
	if availableWidth <= 0 || tableRenderWidth(widths) <= availableWidth {
		return widths
	}
	fixed := fixedPriorityColumnIndexes(view)
	shrinkable := shrinkableColumnIndexes(widths, fixed)
	for tableRenderWidth(widths) > availableWidth && len(shrinkable) > 0 {
		if !shrinkOneStep(widths, shrinkable) {
			break
		}
	}
	return widths
}

func naturalColumnWidths(rows [][]string) []int {
	if len(rows) == 0 {
		return nil
	}
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		updateNaturalWidths(widths, row)
	}
	return widths
}

func updateNaturalWidths(widths []int, row []string) {
	for index, value := range row {
		if index < len(widths) && len(value) > widths[index] {
			widths[index] = len(value)
		}
	}
}

func fixedPriorityColumnIndexes(
	view tui.ResourceView,
) map[int]struct{} {
	fixed := map[int]struct{}{0: {}}
	for index, column := range view.Columns {
		if strings.EqualFold(column, "NAME") {
			fixed[index+1] = struct{}{}
			break
		}
	}
	return fixed
}

func shrinkableColumnIndexes(widths []int, fixed map[int]struct{}) []int {
	indexes := make([]int, 0, len(widths))
	for index := len(widths) - 1; index >= 0; index-- {
		if _, ok := fixed[index]; ok {
			continue
		}
		indexes = append(indexes, index)
	}
	return indexes
}

func shrinkOneStep(widths []int, shrinkable []int) bool {
	for _, index := range shrinkable {
		if widths[index] > minAutosizeColumnWidth {
			widths[index]--
			return true
		}
	}
	return false
}

func tableRenderWidth(widths []int) int {
	if len(widths) == 0 {
		return 0
	}
	total := len(widths) - 1
	for _, width := range widths {
		total += width
	}
	return total
}

func composeTableTitle(
	view tui.ResourceView,
	leftOverflow bool,
	rightOverflow bool,
) string {
	indicators := ""
	if leftOverflow {
		indicators += "◀"
	}
	if rightOverflow {
		indicators += "▶"
	}
	title := fmt.Sprintf("%s(all)[%d]", titleResourceLabel(view.Resource), len(view.Rows))
	if indicators != "" {
		title += fmt.Sprintf(" [%s]", indicators)
	}
	return fmt.Sprintf(" ─ %s ─ ", title)
}

func composeRuntimeTitle(
	runtime *explorerRuntime,
	view tui.ResourceView,
	leftOverflow bool,
	rightOverflow bool,
) string {
	if runtime != nil && runtime.logMode {
		return composeLogTitle(runtime.logObjectPath, runtime.logTarget)
	}
	return composeTableTitle(view, leftOverflow, rightOverflow)
}

func composeLogTitle(objectPath string, target string) string {
	path := strings.TrimSpace(objectPath)
	if path == "" {
		path = "unknown"
	}
	resolvedTarget := strings.TrimSpace(target)
	if resolvedTarget == "" {
		return fmt.Sprintf(" ─ Logs %s ─ ", path)
	}
	return fmt.Sprintf(" ─ Logs %s (target=%s) ─ ", path, resolvedTarget)
}

func logResourceView(entries []runtimeLogEntry, availableWidth int) tui.ResourceView {
	messageWidth := computeLogMessageWidth(availableWidth)
	lines := renderLogLines(entries, messageWidth)
	rows := make([][]string, 0, len(lines))
	ids := make([]string, 0, len(lines))
	for index, line := range lines {
		rows = append(rows, []string{line})
		ids = append(ids, fmt.Sprintf("log-%d", index))
	}
	return tui.ResourceView{
		Resource: tui.ResourceVM,
		Columns:  []string{"LOG"},
		Rows:     rows,
		IDs:      ids,
	}
}

func renderLogLines(entries []runtimeLogEntry, messageWidth int) []string {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, formatLogEntry(entry, messageWidth)...)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func formatLogEntry(entry runtimeLogEntry, messageWidth int) []string {
	prefix := fmt.Sprintf(
		"%-*s %-*s ",
		logTimestampWidth,
		strings.TrimSpace(entry.Timestamp),
		logLevelWidth,
		strings.ToUpper(strings.TrimSpace(entry.Level)),
	)
	continuationPrefix := strings.Repeat(" ", logContinuationIndentWidth())
	chunks := wrapLogMessage(entry.Message, maxInt(messageWidth, logMessageMinWidth))
	lines := make([]string, 0, len(chunks))
	for index, chunk := range chunks {
		if index == 0 {
			lines = append(lines, prefix+chunk)
			continue
		}
		lines = append(lines, continuationPrefix+chunk)
	}
	return lines
}

func logContinuationIndentWidth() int {
	return logTimestampWidth + 1 + logLevelWidth + 1
}

func wrapLogMessage(message string, width int) []string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return []string{"-"}
	}
	words := strings.Fields(trimmed)
	lines := []string{}
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len(current)+1+len(word) > width {
			lines = append(lines, current)
			current = word
			continue
		}
		current += " " + word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func computeLogMessageWidth(availableWidth int) int {
	width := availableWidth - fixedTableColumns - logContinuationIndentWidth() - 2
	if width < logMessageMinWidth {
		return logMessageMinWidth
	}
	return width
}

func (r *explorerRuntime) logScrollableRowCount(availableWidth int) int {
	view := logResourceView(r.logEntries, availableWidth)
	rows := tableRows(view, r.session.IsMarked, r.tableHeaderVisible())
	scrollable := len(rows) - fixedHeaderRows(r.tableHeaderVisible())
	if scrollable < 0 {
		return 0
	}
	return scrollable
}

func (r *explorerRuntime) logViewportPageSize() int {
	_, _, _, height := r.body.GetInnerRect()
	pageSize := height - fixedHeaderRows(r.tableHeaderVisible())
	if pageSize < 1 {
		return 1
	}
	return pageSize
}

func (r *explorerRuntime) logViewportMaxOffset(availableWidth int) int {
	return maxInt(
		0,
		r.logScrollableRowCount(availableWidth)-r.logViewportPageSize(),
	)
}

func computeLogViewportOffset(
	currentOffset int,
	totalRows int,
	pageSize int,
	command string,
) int {
	if totalRows <= 0 {
		return 0
	}
	if pageSize < 1 {
		pageSize = 1
	}
	maxOffset := maxInt(0, totalRows-pageSize)
	offset := currentOffset
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	switch command {
	case "top":
		return 0
	case "bottom":
		return maxOffset
	case "pageup":
		return maxInt(0, offset-pageSize)
	case "pagedown":
		return minInt(maxOffset, offset+pageSize)
	default:
		return offset
	}
}

func titleResourceLabel(resource tui.Resource) string {
	switch resource {
	case tui.ResourceVM:
		return "VM"
	case tui.ResourceLUN:
		return "LUN"
	case tui.ResourceCluster:
		return "CLUSTER"
	case tui.ResourceTask:
		return "TASK"
	case tui.ResourceHost:
		return "HOST"
	case tui.ResourceDatastore:
		return "DATASTORE"
	default:
		return strings.ToUpper(string(resource))
	}
}

func contentFrameColor(theme explorerTheme) tcell.Color {
	if !theme.UseColor {
		return tcell.ColorWhite
	}
	return tcell.ColorAqua
}

func tableRowColor(theme explorerTheme, view tui.ResourceView, rowIndex int) tcell.Color {
	if !theme.UseColor {
		return tcell.ColorWhite
	}
	status := rowStatusAt(view, rowIndex)
	switch classifyStatusLevel(status) {
	case "healthy":
		return theme.RowHealthy
	case "degraded":
		return theme.RowDegraded
	case "faulted":
		return theme.RowFaulted
	}
	if rowIndex%2 == 0 {
		return theme.EvenRowText
	}
	return theme.OddRowText
}

func tableTextColor(
	theme explorerTheme,
	view tui.ResourceView,
	rowIndex int,
	isMarked bool,
	isSelected bool,
) tcell.Color {
	if isSelected || isMarked {
		return tcell.ColorBlack
	}
	return tableRowColor(theme, view, rowIndex)
}

func tableRowBackgroundColor(theme explorerTheme, isMarked bool, isSelected bool) tcell.Color {
	if isSelected && isMarked {
		return theme.RowMarkedSelected
	}
	if isSelected {
		return selectedRowBackgroundColor(theme)
	}
	if isMarked {
		return markedRowBackgroundColor(theme)
	}
	return theme.CanvasBackground
}

func selectedRowBackgroundColor(theme explorerTheme) tcell.Color {
	return theme.RowSelected
}

func markedRowBackgroundColor(theme explorerTheme) tcell.Color {
	return theme.RowMarked
}

func rowStatusAt(view tui.ResourceView, rowIndex int) string {
	if rowIndex < 0 || rowIndex >= len(view.Rows) {
		return ""
	}
	columnIndex := statusColumnIndex(view.Columns)
	if columnIndex < 0 || columnIndex >= len(view.Rows[rowIndex]) {
		return ""
	}
	return strings.TrimSpace(strings.ToLower(view.Rows[rowIndex][columnIndex]))
}

func statusColumnIndex(columns []string) int {
	priority := []string{"STATUS", "CONNECTION", "POWER", "STATE"}
	for _, target := range priority {
		for index, column := range columns {
			if strings.EqualFold(column, target) {
				return index
			}
		}
	}
	return -1
}

func classifyStatusLevel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "healthy", "ok", "connected", "on", "running", "ready", "success":
		return "healthy"
	case "degraded", "warn", "warning", "maintenance", "suspended":
		return "degraded"
	case "faulted", "error", "failed", "disconnected", "down", "critical":
		return "faulted"
	default:
		return ""
	}
}

func renderTopHeaderLine(width int, left string, center string, right string) string {
	leftWidth, centerWidth, rightWidth := topHeaderZoneWidths(width)
	return fitHeaderLeft(left, leftWidth) +
		fitHeaderCenter(center, centerWidth) +
		fitHeaderRight(right, rightWidth)
}

func renderTopHeaderLines(width int, left []string, center []string, right []string) []string {
	lineCount := maxHeaderLineCount(left, center, right)
	lines := make([]string, 0, lineCount)
	for index := 0; index < lineCount; index++ {
		lines = append(
			lines,
			renderTopHeaderLine(
				width,
				topHeaderLineAt(left, index),
				topHeaderLineAt(center, index),
				topHeaderLineAt(right, index),
			),
		)
	}
	return lines
}

func renderTopHeaderLinesWithTheme(
	width int,
	left []string,
	center []string,
	right []string,
	theme explorerTheme,
) []string {
	lines := renderTopHeaderLines(width, left, center, right)
	if !theme.UseColor {
		return lines
	}
	colored := make([]string, 0, len(lines))
	for index := range lines {
		leftValue := colorizeHeaderAccent(fitHeaderLeft(topHeaderLineAt(left, index), width/3), theme.HeaderAccentLeft)
		centerValue := colorizeHeaderAccent(fitHeaderCenter(topHeaderLineAt(center, index), width/3), theme.HeaderAccentCenter)
		rightWidth := width - (width / 3) - (width / 3)
		rightValue := colorizeHeaderAccent(fitHeaderRight(topHeaderLineAt(right, index), rightWidth), theme.HeaderAccentRight)
		colored = append(colored, leftValue+centerValue+rightValue)
	}
	return colored
}

func colorizeHeaderAccent(value string, accent string) string {
	if accent == "" {
		return value
	}
	return fmt.Sprintf("[%s]%s[-]", accent, value)
}

func tableOverflowMarkers(
	widths []int,
	availableWidth int,
	columnOffset int,
	fixedColumns int,
) (bool, bool) {
	if len(widths) == 0 || availableWidth <= 0 {
		return false, len(widths) > 0
	}
	fixed := clampFixedColumns(fixedColumns, len(widths))
	start := fixed + maxInt(columnOffset, 0)
	if start > len(widths) {
		start = len(widths)
	}
	left := columnOffset > 0 && start > fixed
	visible := visibleScrollableColumns(widths, availableWidth, start, fixed)
	right := start+visible < len(widths)
	return left, right
}

func clampFixedColumns(value int, max int) int {
	if value < 0 {
		return 0
	}
	if value > max {
		return max
	}
	return value
}

func visibleScrollableColumns(
	widths []int,
	availableWidth int,
	scrollStart int,
	fixedColumns int,
) int {
	visible := fixedColumns
	for index := scrollStart; index < len(widths); index++ {
		if renderWidthForVisibleCount(widths, visible+1, scrollStart) > availableWidth {
			break
		}
		visible++
	}
	return visible - fixedColumns
}

func renderWidthForVisibleCount(widths []int, visibleCount int, scrollStart int) int {
	if visibleCount <= 0 {
		return 0
	}
	lastFixed := visibleCount
	if lastFixed > scrollStart {
		lastFixed = scrollStart
	}
	total := 0
	for index := 0; index < lastFixed; index++ {
		total += widths[index]
	}
	for index := scrollStart; index < scrollStart+visibleCount-lastFixed; index++ {
		total += widths[index]
	}
	return total + visibleCount - 1
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func selectionForTable(session tui.Session, includeHeader bool) (int, int) {
	return selectionForRenderedView(session, session.CurrentView(), includeHeader)
}

func selectionForRenderedView(
	session tui.Session,
	view tui.ResourceView,
	includeHeader bool,
) (int, int) {
	row := session.SelectedRow()
	if includeHeader {
		row++
	}
	column := session.SelectedColumn() + 1
	maxColumn := len(view.Columns)
	if column > maxColumn {
		column = maxColumn
	}
	if column < 0 {
		column = 0
	}
	if row < 0 {
		row = 0
	}
	if row > len(view.Rows) {
		row = len(view.Rows)
	}
	return row, column
}

func fixedHeaderRows(includeHeader bool) int {
	if includeHeader {
		return 1
	}
	return 0
}

func (r *explorerRuntime) tableHeaderVisible() bool {
	if r.headless {
		return false
	}
	return r.headerVisible
}

func helpModalText(view tui.ResourceView) string {
	actions := "none"
	if len(view.Actions) > 0 {
		actions = strings.Join(view.Actions, ", ")
	}
	return fmt.Sprintf(
		"View: %s\nActions: %s\nKeys: Esc close | J/K move row | H/L move column | D describe",
		view.Resource,
		actions,
	)
}

func renderResourceDetails(details tui.ResourceDetails) string {
	builder := &strings.Builder{}
	builder.WriteString(details.Title + "\n")
	for _, field := range details.Fields {
		builder.WriteString(fmt.Sprintf("%s: %s\n", field.Key, field.Value))
	}
	return strings.TrimRight(builder.String(), "\n")
}

func applyPromptCompletion(
	promptState *tui.PromptState,
	view tui.ResourceView,
	text string,
) (string, string, bool) {
	suggestions := promptState.Suggest(text, view)
	if len(suggestions) == 0 {
		return text, "", false
	}
	if text == suggestions[0] {
		return text, "", false
	}
	return suggestions[0], "completion: " + suggestions[0], true
}

func promptValidationMessage(text string) string {
	trimmed := strings.TrimSpace(text)
	if isPendingPromptInput(text, trimmed) {
		return ""
	}
	if isLocalPromptCommand(trimmed) {
		return ""
	}
	if _, err := tui.ParseExplorerInput(text); err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error())
	}
	return ""
}

func isLocalPromptCommand(trimmed string) bool {
	value := strings.ToLower(strings.TrimSpace(trimmed))
	return value == ":table" ||
		value == ":log" ||
		value == ":logs" ||
		value == ":cols" ||
		value == ":columns" ||
		strings.HasPrefix(value, ":log ") ||
		strings.HasPrefix(value, ":logs ") ||
		strings.HasPrefix(value, ":cols ") ||
		strings.HasPrefix(value, ":columns ")
}

func isPendingPromptInput(raw string, trimmed string) bool {
	if trimmed == "" || trimmed == ":" || trimmed == "!" || trimmed == "/" {
		return true
	}
	if !strings.HasSuffix(raw, " ") {
		return false
	}
	return hasPendingCommandArgument(trimmed)
}

func hasPendingCommandArgument(trimmed string) bool {
	if strings.HasPrefix(trimmed, ":") {
		return len(strings.Fields(strings.TrimPrefix(trimmed, ":"))) <= 1
	}
	if strings.HasPrefix(trimmed, "!") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "!")) == ""
	}
	if strings.HasPrefix(trimmed, "/") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "/")) == ""
	}
	return false
}

func applyPromptValidationState(prompt *tview.InputField, message string) {
	if message == "" {
		prompt.SetLabelColor(tcell.ColorWhite)
		prompt.SetFieldTextColor(tcell.ColorWhite)
		return
	}
	prompt.SetLabelColor(tcell.ColorRed)
	prompt.SetFieldTextColor(tcell.ColorRed)
}

func readTheme() explorerTheme {
	theme := explorerTheme{
		UseColor:           true,
		CanvasBackground:   tcell.ColorBlack,
		HeaderText:         tcell.ColorBlack,
		HeaderBackground:   tcell.ColorAqua,
		HeaderAccentLeft:   "yellow",
		HeaderAccentCenter: "aqua",
		HeaderAccentRight:  "fuchsia",
		EvenRowText:        tcell.ColorWhite,
		OddRowText:         tcell.ColorLightGray,
		RowHealthy:         tcell.ColorGreen,
		RowDegraded:        tcell.ColorYellow,
		RowFaulted:         tcell.ColorRed,
		RowSelected:        tcell.ColorYellow,
		RowMarked:          tcell.ColorDarkCyan,
		RowMarkedSelected:  tcell.ColorYellow,
		StatusError:        "[red]",
	}
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		theme.UseColor = false
		theme.CanvasBackground = tcell.ColorBlack
		theme.HeaderText = tcell.ColorWhite
		theme.HeaderBackground = tcell.ColorBlack
		theme.HeaderAccentLeft = ""
		theme.HeaderAccentCenter = ""
		theme.HeaderAccentRight = ""
		theme.EvenRowText = tcell.ColorWhite
		theme.OddRowText = tcell.ColorWhite
		theme.RowHealthy = tcell.ColorWhite
		theme.RowDegraded = tcell.ColorWhite
		theme.RowFaulted = tcell.ColorWhite
		theme.RowSelected = tcell.ColorGray
		theme.RowMarked = tcell.ColorDarkGray
		theme.RowMarkedSelected = tcell.ColorSilver
		theme.StatusError = ""
	}
	return theme
}

func maxHeaderLineCount(left []string, center []string, right []string) int {
	lineCount := len(left)
	if len(center) > lineCount {
		lineCount = len(center)
	}
	if len(right) > lineCount {
		lineCount = len(right)
	}
	return lineCount
}

func topHeaderLineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func topHeaderZoneWidths(width int) (int, int, int) {
	leftWidth := width / 3
	centerWidth := width / 3
	rightWidth := width - leftWidth - centerWidth
	return leftWidth, centerWidth, rightWidth
}

func fitHeaderLeft(text string, width int) string {
	value := trimHeaderText(text, width)
	return value + strings.Repeat(" ", width-len(value))
}

func fitHeaderCenter(text string, width int) string {
	value := trimHeaderText(text, width)
	pad := width - len(value)
	leftPad := pad / 2
	rightPad := pad - leftPad
	return strings.Repeat(" ", leftPad) + value + strings.Repeat(" ", rightPad)
}

func fitHeaderRight(text string, width int) string {
	value := trimHeaderText(text, width)
	return strings.Repeat(" ", width-len(value)) + value
}

func trimHeaderText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(text) <= width {
		return text
	}
	return text[:width]
}

func normalizeTopHeaderWidth(width int) int {
	if width <= 0 {
		return defaultTopHeaderWidth
	}
	return width
}

func renderTopHeaderLeft(context string) string {
	return strings.Join(
		[]string{
			fmt.Sprintf("Context: %s", context),
			"Cluster: n/a",
			"User: n/a",
			fmt.Sprintf("HS Version: %s", buildVersion),
			"vCenter Version: unknown",
			"CPU: " + formatMetricWithTrend(defaultCPUPercent, defaultCPUTrend),
			"MEM: " + formatMetricWithTrend(defaultMEMPercent, defaultMEMTrend),
		},
		"\n",
	)
}

func formatMetricWithTrend(percent int, trend int) string {
	switch {
	case trend > 0:
		return fmt.Sprintf("%d%%(+)", percent)
	case trend < 0:
		return fmt.Sprintf("%d%%(-)", percent)
	default:
		return fmt.Sprintf("%d%%", percent)
	}
}

func defaultRuntimeLogEntries() []runtimeLogEntry {
	return []runtimeLogEntry{
		{Timestamp: "2026-02-16T10:31:10Z", Level: "INFO", Message: "Connected to vc-primary and initialized session"},
		{Timestamp: "2026-02-16T10:31:22Z", Level: "WARN", Message: "Datastore ds-7 latency above baseline threshold"},
		{Timestamp: "2026-02-16T10:31:40Z", Level: "ERROR", Message: "Host esxi-06 disconnected from management network"},
		{Timestamp: "2026-02-16T10:32:03Z", Level: "INFO", Message: "Refresh cycle completed with 8 VM rows and 8 datastore rows"},
	}
}

func renderTopHeaderCenter(logMode bool, promptMode bool) string {
	return renderTopHeaderCenterWithContext(logMode, promptMode, "", "")
}

func renderTopHeaderCenterWithContext(
	logMode bool,
	promptMode bool,
	path string,
	status string,
) string {
	prompt := "OFF"
	if promptMode {
		prompt = "ON"
	}
	lines := []string{}
	if logMode {
		lines = append(
			lines,
			"<g> Top         <G> Bottom",
			"<PgUp/PgDn> Scroll",
			fmt.Sprintf("Prompt: %s | <q> Quit", prompt),
		)
	} else {
		lines = append(
			lines,
			"<:> Command    </> Filter",
			"<?> Help       <!> Action",
			"<Tab> Complete <h/j/k/l> Move",
			fmt.Sprintf("<Shift+O> Sort Prompt: %s | <q> Quit", prompt),
		)
	}
	if path != "" {
		lines = append(lines, "path: "+path)
	}
	if status != "" {
		lines = append(lines, "status: "+status)
	}
	return strings.Join(lines, "\n")
}

func renderTopHeaderRight() string {
	return strings.Join(
		[]string{
			"          .------------.        ",
			"       .-'   +------+   '-.     ",
			"     .'    /|      /|      '.   ",
			"    /     +------+ |         \\  ",
			"    \\     | +----|-+        /   ",
			"     '.   |/     |/      .-'    ",
			"       '-. +------+   .-'       ",
		},
		"\n",
	)
}

func compactTopHeaderPath(crumbsless bool, path string) string {
	if crumbsless {
		return ""
	}
	return path
}

func compactTopHeaderStatus(status string) string {
	sanitized := strings.TrimSpace(stripTviewTags(status))
	if sanitized == "" {
		return "ready"
	}
	return sanitized
}

func stripTviewTags(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))
	inTag := false
	for _, value := range text {
		if value == '[' {
			inTag = true
			continue
		}
		if value == ']' {
			inTag = false
			continue
		}
		if !inTag {
			builder.WriteRune(value)
		}
	}
	return builder.String()
}

func executePromptCommand(
	session *tui.Session,
	promptState *tui.PromptState,
	executor *runtimeActionExecutor,
	contexts *runtimeContextManager,
	aliasRegistry *commandAliasRegistry,
	line string,
) (string, bool) {
	resolvedLine, err := resolveCommandAliases(aliasRegistry, line)
	if err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error()), true
	}
	parsed, err := tui.ParseExplorerInput(resolvedLine)
	if err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error()), true
	}
	if shouldRecordHistory(parsed.Kind) {
		promptState.Record(line)
	}
	return runCommand(session, promptState, executor, contexts, parsed)
}

func resolveCommandAliases(aliasRegistry *commandAliasRegistry, line string) (string, error) {
	if aliasRegistry != nil {
		return aliasRegistry.Resolve(line), nil
	}
	registry, err := loadDefaultCommandAliasRegistry()
	if err != nil {
		return "", err
	}
	return registry.Resolve(line), nil
}

func runCommand(
	session *tui.Session,
	promptState *tui.PromptState,
	executor *runtimeActionExecutor,
	contexts *runtimeContextManager,
	parsed tui.ExplorerCommand,
) (string, bool) {
	if message, keepRunning, handled := handleSimpleCommandKinds(
		session,
		promptState,
		parsed,
	); handled {
		return message, keepRunning
	}
	switch parsed.Kind {
	case tui.CommandContext:
		return handleContextCommand(session, contexts, parsed.Value), true
	case tui.CommandView:
		return statusFromError(session.ExecuteCommand(":"+parsed.Value), "view: "+parsed.Value), true
	case tui.CommandAction:
		return handleActionCommand(session, executor, parsed.Value), true
	default:
		return statusFromError(session.HandleKey(parsed.Value), "key: "+parsed.Value), true
	}
}

func handleSimpleCommandKinds(
	session *tui.Session,
	promptState *tui.PromptState,
	parsed tui.ExplorerCommand,
) (string, bool, bool) {
	switch parsed.Kind {
	case tui.CommandNoop:
		return "", true, true
	case tui.CommandQuit:
		return "bye", false, true
	case tui.CommandHelp:
		return "use :vm/:lun/:cluster/:host/:datastore, :ctx, :ro, /text, !action", true, true
	case tui.CommandReadOnly:
		applyReadOnlyMode(session, parsed.Value)
		return readOnlyStatus(*session), true, true
	case tui.CommandHistory:
		return historyStatus(promptState, parsed.Value), true, true
	case tui.CommandSuggest:
		return suggestStatus(promptState, parsed.Value, session.CurrentView()), true, true
	case tui.CommandLastView:
		return statusFromError(session.LastView(), "switched to last view"), true, true
	case tui.CommandFilter:
		filterValue := strings.TrimSpace(parsed.Value)
		if strings.HasPrefix(filterValue, "-f") {
			fuzzyQuery := strings.TrimSpace(strings.TrimPrefix(filterValue, "-f"))
			return statusFromError(
				session.ApplyFuzzyFilter(fuzzyQuery),
				fmt.Sprintf("filter: %s", parsed.Value),
			), true, true
		}
		if strings.HasPrefix(filterValue, "-t") {
			tagExpression := strings.TrimSpace(strings.TrimPrefix(filterValue, "-t"))
			return statusFromError(
				session.ApplyTagFilter(tagExpression),
				fmt.Sprintf("filter: %s", parsed.Value),
			), true, true
		}
		if strings.HasPrefix(filterValue, "!") {
			pattern := strings.TrimSpace(strings.TrimPrefix(filterValue, "!"))
			return statusFromError(
				session.ApplyInverseRegexFilter(pattern),
				fmt.Sprintf("filter: %s", parsed.Value),
			), true, true
		}
		return statusFromError(
			session.ApplyRegexFilter(parsed.Value),
			fmt.Sprintf("filter: %s", parsed.Value),
		), true, true
	default:
		return "", true, false
	}
}

func handleActionCommand(
	session *tui.Session,
	executor tui.ActionExecutor,
	action string,
) string {
	if err := session.ApplyAction(action, executor); err != nil {
		return formatActionError(err, *session)
	}
	if runtimeExecutor, ok := executor.(*runtimeActionExecutor); ok && runtimeExecutor.last != "" {
		return runtimeExecutor.last
	}
	return "action executed"
}

func formatActionError(err error, session tui.Session) string {
	return fmt.Sprintf(
		"[red]action_error code=%s message=%q entity=%q retryable=%t",
		actionErrorCode(err),
		err.Error(),
		selectedActionEntity(session),
		isRetriableCommandError(err),
	)
}

func actionErrorCode(err error) string {
	switch {
	case errors.Is(err, tui.ErrReadOnly):
		return "ERR_READ_ONLY"
	case errors.Is(err, tui.ErrActionTimeout):
		return "ERR_TIMEOUT"
	case errors.Is(err, tui.ErrConfirmationRequired):
		return "ERR_CONFIRMATION_REQUIRED"
	case errors.Is(err, tui.ErrInvalidAction):
		return "ERR_INVALID_ACTION"
	default:
		return "ERR_ACTION"
	}
}

func selectedActionEntity(session tui.Session) string {
	view := session.CurrentView()
	row := session.SelectedRow()
	if row < 0 || row >= len(view.IDs) {
		return "-"
	}
	return view.IDs[row]
}

type retriableCommandError interface {
	Retriable() bool
}

func isRetriableCommandError(err error) bool {
	var retriable retriableCommandError
	return errors.As(err, &retriable) && retriable.Retriable()
}

func handleContextCommand(
	session *tui.Session,
	contexts *runtimeContextManager,
	value string,
) string {
	if contexts == nil {
		return "[red]command error: context manager unavailable"
	}
	if strings.TrimSpace(value) == "" {
		return contextListStatus(*contexts)
	}
	if err := contexts.Switch(value); err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error())
	}
	if err := refreshActiveView(session); err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error())
	}
	return fmt.Sprintf("context: %s", contexts.Active())
}

func refreshActiveView(session *tui.Session) error {
	return session.ExecuteCommand(":" + string(session.CurrentView().Resource))
}

func contextListStatus(contexts runtimeContextManager) string {
	return fmt.Sprintf(
		"contexts: %s | active: %s",
		strings.Join(contexts.List(), ", "),
		contexts.Active(),
	)
}

func statusFromError(err error, success string) string {
	if err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error())
	}
	return success
}

func readOnlyStatus(session tui.Session) string {
	if session.ReadOnly() {
		return "mode: read-only"
	}
	return "mode: read-write"
}

func historyStatus(promptState *tui.PromptState, direction string) string {
	entry, ok := readHistoryEntry(promptState, direction)
	if !ok {
		return "history: <none>"
	}
	return "history: " + entry
}

func suggestStatus(promptState *tui.PromptState, prefix string, view tui.ResourceView) string {
	suggestions := promptState.Suggest(prefix, view)
	if len(suggestions) == 0 {
		return "suggestions: <none>"
	}
	return "suggestions: " + strings.Join(suggestions, ", ")
}

func shouldRecordHistory(kind tui.CommandKind) bool {
	return kind != tui.CommandNoop && kind != tui.CommandHistory
}

func applyReadOnlyMode(session *tui.Session, mode string) {
	switch mode {
	case "on":
		session.SetReadOnly(true)
	case "off":
		session.SetReadOnly(false)
	default:
		session.SetReadOnly(!session.ReadOnly())
	}
}

func readHistoryEntry(prompt *tui.PromptState, direction string) (string, bool) {
	if direction == "up" {
		return prompt.Previous()
	}
	return prompt.Next()
}

func isPromptActivation(evt *tcell.EventKey) bool {
	if evt.Key() != tcell.KeyRune {
		return false
	}
	r := evt.Rune()
	return r == ':' || r == '/' || r == '!'
}

func isDescribeEvent(evt *tcell.EventKey) bool {
	return evt.Key() == tcell.KeyRune && evt.Modifiers() == tcell.ModNone && evt.Rune() == 'd'
}

func isQuitEvent(evt *tcell.EventKey) bool {
	if evt.Key() == tcell.KeyCtrlC {
		return true
	}
	return evt.Key() == tcell.KeyRune && (evt.Rune() == 'q' || evt.Rune() == 'Q')
}

func eventToHotKey(evt *tcell.EventKey) (string, bool) {
	if evt.Key() == tcell.KeyCtrlW {
		return "CTRL+W", true
	}
	if evt.Key() == tcell.KeyCtrlE {
		return "CTRL+E", true
	}
	if evt.Key() == tcell.KeyCtrlSpace {
		return "CTRL+SPACE", true
	}
	if evt.Key() == tcell.KeyCtrlBackslash {
		return "CTRL+\\", true
	}
	if evt.Modifiers()&tcell.ModShift != 0 && evt.Key() == tcell.KeyLeft {
		return "SHIFT+LEFT", true
	}
	if evt.Modifiers()&tcell.ModShift != 0 && evt.Key() == tcell.KeyRight {
		return "SHIFT+RIGHT", true
	}
	if evt.Key() == tcell.KeyUp {
		return "UP", true
	}
	if evt.Key() == tcell.KeyDown {
		return "DOWN", true
	}
	if evt.Key() == tcell.KeyLeft {
		return "LEFT", true
	}
	if evt.Key() == tcell.KeyRight {
		return "RIGHT", true
	}
	if evt.Key() != tcell.KeyRune {
		return "", false
	}
	if evt.Rune() == ' ' {
		return "SPACE", true
	}
	if evt.Rune() == 'h' {
		return "LEFT", true
	}
	if evt.Rune() == 'l' {
		return "RIGHT", true
	}
	if evt.Rune() == '\\' && evt.Modifiers()&tcell.ModCtrl != 0 {
		return "CTRL+\\", true
	}
	if evt.Modifiers()&tcell.ModShift != 0 {
		return strings.ToUpper(string(evt.Rune())), true
	}
	return string(evt.Rune()), true
}
