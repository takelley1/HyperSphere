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

type explorerRuntime struct {
	app          *tview.Application
	session      tui.Session
	promptState  tui.PromptState
	actionExec   *runtimeActionExecutor
	contexts     runtimeContextManager
	headless     bool
	crumbsless   bool
	theme        explorerTheme
	pages        *tview.Pages
	layout       *tview.Flex
	helpModal    *tview.Modal
	aliasModal   *tview.Modal
	body         *tview.Table
	breadcrumb   *tview.TextView
	status       *tview.TextView
	prompt       *tview.InputField
	footer       *tview.TextView
	aliasEntries []string
	promptMode   bool
	helpOpen     bool
	aliasOpen    bool
	helpText     string
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
		app:         tview.NewApplication(),
		session:     tui.NewSession(defaultCatalog()),
		promptState: tui.NewPromptState(defaultPromptHistorySize),
		actionExec:  &runtimeActionExecutor{},
		contexts:    newRuntimeContextManager(),
		headless:    headless,
		crumbsless:  crumbsless,
		theme:       readTheme(),
		pages:       tview.NewPages(),
		helpModal:   tview.NewModal(),
		aliasModal:  tview.NewModal(),
		body:        tview.NewTable(),
		breadcrumb:  tview.NewTextView(),
		status:      tview.NewTextView(),
		prompt:      tview.NewInputField(),
		footer:      tview.NewTextView(),
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
	r.body.SetFixed(fixedHeaderRows(r.headless), 1)
	r.body.SetBorders(false)
	r.body.SetSeparator(' ')
	r.body.SetBorder(true)
	r.body.SetTitle(" HyperSphere Explorer ")
	r.body.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorWhite))
	r.breadcrumb.SetDynamicColors(true)
	r.breadcrumb.SetBorder(true)
	r.breadcrumb.SetTitle(" Breadcrumbs ")
	r.status.SetDynamicColors(true)
	r.status.SetBorder(true)
	r.status.SetTitle(" Status ")
	r.prompt.SetLabel("Command: ")
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
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
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
	r.app.SetRoot(r.pages, true)
	r.app.SetFocus(r.body)
}

func (r *explorerRuntime) configureHandlers() {
	r.app.SetInputCapture(r.handleGlobalKey)
	r.prompt.SetDoneFunc(r.handlePromptDone)
	r.prompt.SetInputCapture(r.handlePromptHistory)
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
	r.emitStatus(r.session.HandleKey(command))
	r.render("")
	return nil
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
	r.prompt.SetText(prefix)
	r.app.SetFocus(r.prompt)
}

func (r *explorerRuntime) endPrompt() {
	r.promptMode = false
	r.prompt.SetText("")
	r.app.SetFocus(r.body)
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
	r.renderTable()
	r.renderBreadcrumb()
	r.footer.SetText(renderFooter(r.promptMode))
}

func (r *explorerRuntime) renderBreadcrumb() {
	if r.crumbsless || r.breadcrumb == nil {
		return
	}
	r.breadcrumb.SetText("home > " + string(r.session.CurrentView().Resource))
}

func (r *explorerRuntime) renderTable() {
	includeHeader := !r.headless
	rows := tableRows(r.session.CurrentView(), r.session.IsMarked, includeHeader)
	r.body.Clear()
	for rowIndex, row := range rows {
		for columnIndex, value := range row {
			cell := tview.NewTableCell(value)
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
	selectedRow, selectedColumn := selectionForTable(r.session, includeHeader)
	r.body.Select(selectedRow, selectedColumn)
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

func selectionForTable(session tui.Session, includeHeader bool) (int, int) {
	view := session.CurrentView()
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

func fixedHeaderRows(headless bool) int {
	if headless {
		return 0
	}
	return 1
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
		"View: %s\nActions: %s\nKeys: Esc close | J/K move row | H/L move column",
		view.Resource,
		actions,
	)
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

func isQuitEvent(evt *tcell.EventKey) bool {
	if evt.Key() == tcell.KeyCtrlC {
		return true
	}
	return evt.Key() == tcell.KeyRune && (evt.Rune() == 'q' || evt.Rune() == 'Q')
}

func eventToHotKey(evt *tcell.EventKey) (string, bool) {
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
