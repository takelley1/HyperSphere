// Path: cmd/hypersphere/explorer_tui.go
// Description: Run a full-screen real-time TUI explorer using tview/tcell with k9s-inspired interactions.
package main

import (
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
const explorerTableTitle = "HyperSphere Explorer"
const fixedTableColumns = 1
const defaultTopHeaderWidth = 120
const topHeaderPanelHeight = 7

var compactColumnsByResource = map[tui.Resource][]string{
	tui.ResourceVM:        {"NAME", "POWER", "DATASTORE"},
	tui.ResourceLUN:       {"NAME", "DATASTORE", "USED_GB"},
	tui.ResourceCluster:   {"NAME", "HOSTS", "VMS"},
	tui.ResourceHost:      {"NAME", "CLUSTER", "CONNECTION"},
	tui.ResourceDatastore: {"NAME", "CLUSTER", "FREE_GB"},
}

type explorerRuntime struct {
	app           *tview.Application
	session       tui.Session
	promptState   tui.PromptState
	actionExec    *runtimeActionExecutor
	contexts      runtimeContextManager
	headless      bool
	crumbsless    bool
	theme         explorerTheme
	pages         *tview.Pages
	layout        *tview.Flex
	helpModal     *tview.Modal
	aliasModal    *tview.Modal
	describeModal *tview.Modal
	topHeader     *tview.TextView
	body          *tview.Table
	breadcrumb    *tview.TextView
	status        *tview.TextView
	prompt        *tview.InputField
	footer        *tview.TextView
	aliasEntries  []string
	promptMode    bool
	helpOpen      bool
	aliasOpen     bool
	describeOpen  bool
	helpText      string
	describeText  string
	lastWidth     int
	wideColumns   bool
	headerVisible bool
}

type runtimeActionExecutor struct {
	last string
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
	UseColor         bool
	HeaderText       tcell.Color
	HeaderBackground tcell.Color
	EvenRowText      tcell.Color
	OddRowText       tcell.Color
	StatusError      string
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
		app:           tview.NewApplication(),
		session:       tui.NewSession(defaultCatalog()),
		promptState:   tui.NewPromptState(defaultPromptHistorySize),
		actionExec:    &runtimeActionExecutor{},
		contexts:      newRuntimeContextManager(),
		headless:      headless,
		crumbsless:    crumbsless,
		theme:         readTheme(),
		pages:         tview.NewPages(),
		helpModal:     tview.NewModal(),
		aliasModal:    tview.NewModal(),
		describeModal: tview.NewModal(),
		topHeader:     tview.NewTextView(),
		body:          tview.NewTable(),
		breadcrumb:    tview.NewTextView(),
		status:        tview.NewTextView(),
		prompt:        tview.NewInputField(),
		footer:        tview.NewTextView(),
		wideColumns:   true,
		headerVisible: !headless,
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
	r.body.SetTitle(composeTableTitle(false, false))
	r.body.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorWhite))
	r.breadcrumb.SetDynamicColors(true)
	r.breadcrumb.SetBorder(true)
	r.breadcrumb.SetTitle(" Breadcrumbs ")
	r.status.SetDynamicColors(true)
	r.status.SetBorder(true)
	r.status.SetTitle(" Status ")
	r.prompt.SetLabel("Command: ")
	applyPromptValidationState(r.prompt, "")
	r.footer.SetDynamicColors(true)
	r.footer.SetBorder(true)
	r.footer.SetTitle(" Help ")
	r.helpModal.SetBorder(true)
	r.helpModal.SetTitle(" Keymap Help ")
	r.aliasEntries = tui.ResourceCommandAliases()
	r.aliasModal.SetBorder(true)
	r.aliasModal.SetTitle(" Alias Palette ")
	r.aliasModal.SetText("Select a resource alias")
	r.aliasModal.AddButtons(r.aliasEntries)
	r.aliasModal.SetDoneFunc(r.handleAliasSelection)
	r.describeModal.SetBorder(true)
	r.describeModal.SetTitle(" Describe ")
	r.topHeader.SetDynamicColors(true)
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.topHeader, topHeaderPanelHeight, 0, false).
		AddItem(r.body, 0, 1, true).
		AddItem(r.prompt, 1, 0, false)
	if !r.crumbsless {
		layout.AddItem(r.breadcrumb, 3, 0, false)
	}
	layout.AddItem(r.status, 3, 0, false).AddItem(r.footer, 3, 0, false)
	r.layout = layout
	r.pages.AddPage("main", layout, true, true)
	r.pages.AddPage("help", r.helpModal, true, false)
	r.pages.AddPage("alias", r.aliasModal, true, false)
	r.pages.AddPage("describe", r.describeModal, true, false)
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
	if r.describeOpen {
		if evt.Key() == tcell.KeyEscape {
			r.closeDescribePanel()
		}
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
	r.describeModal.SetText(r.describeText)
	r.pages.ShowPage("describe")
	r.describeOpen = true
}

func (r *explorerRuntime) closeDescribePanel() {
	r.pages.HidePage("describe")
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
	r.footer.SetText(renderFooter(r.promptMode))
}

func (r *explorerRuntime) renderTopHeader() {
	r.renderTopHeaderWithWidth(r.lastWidth)
}

func (r *explorerRuntime) renderTopHeaderWithWidth(width int) {
	if r.topHeader == nil {
		return
	}
	headerLines := renderTopHeaderLines(
		normalizeTopHeaderWidth(width),
		strings.Split(renderTopHeaderLeft(r.contexts.Active()), "\n"),
		strings.Split(renderTopHeaderCenter(), "\n"),
		strings.Split(renderTopHeaderRight(), "\n"),
	)
	r.topHeader.SetText(strings.Join(headerLines, "\n"))
}

func (r *explorerRuntime) renderBreadcrumb() {
	if r.crumbsless || r.breadcrumb == nil {
		return
	}
	r.breadcrumb.SetText("home > " + string(r.session.CurrentView().Resource))
}

func (r *explorerRuntime) renderTable() {
	r.renderTableWithWidth(tableAvailableWidth(r.body))
}

func (r *explorerRuntime) renderTableWithWidth(availableWidth int) {
	includeHeader := r.tableHeaderVisible()
	r.body.SetFixed(fixedHeaderRows(includeHeader), fixedTableColumns)
	view := viewForColumnMode(r.session.CurrentView(), r.wideColumns)
	view = compactViewForWidth(view, availableWidth)
	rows := tableRows(view, r.session.IsMarked, includeHeader)
	widths := autosizedColumnWidths(view, rows, availableWidth)
	_, columnOffset := r.body.GetOffset()
	leftOverflow, rightOverflow := tableOverflowMarkers(
		widths,
		availableWidth,
		columnOffset,
		fixedTableColumns,
	)
	r.body.SetTitle(composeTableTitle(leftOverflow, rightOverflow))
	r.body.Clear()
	for rowIndex, row := range rows {
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
				cell.SetTextColor(tableRowColor(r.theme, rowIndex))
			}
			r.body.SetCell(rowIndex, columnIndex, cell)
		}
	}
	selectedRow, selectedColumn := selectionForRenderedView(r.session, view, includeHeader)
	r.body.Select(selectedRow, selectedColumn)
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

func composeTableTitle(leftOverflow bool, rightOverflow bool) string {
	indicators := ""
	if leftOverflow {
		indicators += "◀"
	}
	if rightOverflow {
		indicators += "▶"
	}
	if indicators == "" {
		return " " + explorerTableTitle + " "
	}
	return fmt.Sprintf(" %s [%s] ", explorerTableTitle, indicators)
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

func renderFooter(promptMode bool) string {
	prompt := "OFF"
	if promptMode {
		prompt = "ON"
	}
	return fmt.Sprintf(
		": view | / filter | ! action | Tab complete | h/j/k/l + arrows move | :ro toggle | Prompt: %s | q quit",
		prompt,
	)
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
	if _, err := tui.ParseExplorerInput(text); err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error())
	}
	return ""
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
		UseColor:         true,
		HeaderText:       tcell.ColorBlack,
		HeaderBackground: tcell.ColorLightSkyBlue,
		EvenRowText:      tcell.ColorWhite,
		OddRowText:       tcell.ColorLightGray,
		StatusError:      "[red]",
	}
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		theme.UseColor = false
		theme.HeaderText = tcell.ColorWhite
		theme.HeaderBackground = tcell.ColorBlack
		theme.EvenRowText = tcell.ColorWhite
		theme.OddRowText = tcell.ColorWhite
		theme.StatusError = ""
	}
	return theme
}

func tableRowColor(theme explorerTheme, rowIndex int) tcell.Color {
	if !theme.UseColor {
		return tcell.ColorWhite
	}
	if rowIndex%2 == 0 {
		return theme.EvenRowText
	}
	return theme.OddRowText
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
			"CPU: n/a",
			"MEM: n/a",
		},
		"\n",
	)
}

func renderTopHeaderCenter() string {
	return strings.Join(
		[]string{
			"<:> Command",
			"</> Filter",
			"<?> Help",
		},
		"\n",
	)
}

func renderTopHeaderRight() string {
	return "HyperSphere"
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
		session.ApplyFilter(parsed.Value)
		return fmt.Sprintf("filter: %s", parsed.Value), true, true
	default:
		return "", true, false
	}
}

func handleActionCommand(
	session *tui.Session,
	executor *runtimeActionExecutor,
	action string,
) string {
	if err := session.ApplyAction(action, executor); err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error())
	}
	if executor.last != "" {
		return executor.last
	}
	return "action executed"
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
