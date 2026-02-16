// Path: cmd/hypersphere/explorer_tui.go
// Description: Run a full-screen real-time TUI explorer using tview/tcell with k9s-inspired interactions.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/takelley1/hypersphere/internal/tui"
)

const defaultPromptHistorySize = 200

type explorerRuntime struct {
	app         *tview.Application
	session     tui.Session
	promptState tui.PromptState
	actionExec  *runtimeActionExecutor
	theme       explorerTheme
	body        *tview.Table
	status      *tview.TextView
	prompt      *tview.InputField
	footer      *tview.TextView
	done        chan struct{}
	promptMode  bool
}

type runtimeActionExecutor struct {
	last string
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

func runExplorerWorkflow(output io.Writer) {
	runtime := newExplorerRuntime()
	if err := runtime.run(); err != nil {
		_, _ = fmt.Fprintf(output, "tui error: %v\n", err)
	}
}

func newExplorerRuntime() explorerRuntime {
	runtime := explorerRuntime{
		app:         tview.NewApplication(),
		session:     tui.NewSession(defaultCatalog()),
		promptState: tui.NewPromptState(defaultPromptHistorySize),
		actionExec:  &runtimeActionExecutor{},
		theme:       readTheme(),
		body:        tview.NewTable(),
		status:      tview.NewTextView(),
		prompt:      tview.NewInputField(),
		footer:      tview.NewTextView(),
		done:        make(chan struct{}),
	}
	runtime.configureWidgets()
	runtime.configureHandlers()
	runtime.render("ready")
	return runtime
}

func (r *explorerRuntime) configureWidgets() {
	r.body.SetSelectable(true, true)
	r.body.SetFixed(1, 1)
	r.body.SetBorders(false)
	r.body.SetSeparator(' ')
	r.body.SetBorder(true)
	r.body.SetTitle(" HyperSphere Explorer ")
	r.body.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorWhite))
	r.status.SetDynamicColors(true)
	r.status.SetBorder(true)
	r.status.SetTitle(" Status ")
	r.prompt.SetLabel("Command: ")
	r.footer.SetDynamicColors(true)
	r.footer.SetBorder(true)
	r.footer.SetTitle(" Help ")
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.body, 0, 1, true).
		AddItem(r.status, 3, 0, false).
		AddItem(r.prompt, 1, 0, false).
		AddItem(r.footer, 3, 0, false)
	r.app.SetRoot(layout, true)
	r.app.SetFocus(r.body)
}

func (r *explorerRuntime) configureHandlers() {
	r.app.SetInputCapture(r.handleGlobalKey)
	r.prompt.SetDoneFunc(r.handlePromptDone)
	r.prompt.SetInputCapture(r.handlePromptHistory)
}

func (r *explorerRuntime) run() error {
	go r.refreshLoop()
	err := r.app.Run()
	close(r.done)
	return err
}

func (r *explorerRuntime) refreshLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			r.app.QueueUpdateDraw(func() {
				r.render("")
			})
		}
	}
}

func (r *explorerRuntime) handleGlobalKey(evt *tcell.EventKey) *tcell.EventKey {
	if r.promptMode {
		return evt
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
	r.footer.SetText(renderFooter(r.promptMode))
}

func (r *explorerRuntime) renderTable() {
	rows := tableRows(r.session.CurrentView(), r.session.IsMarked)
	r.body.Clear()
	for rowIndex, row := range rows {
		for columnIndex, value := range row {
			cell := tview.NewTableCell(value)
			if rowIndex == 0 {
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
	selectedRow, selectedColumn := selectionForTable(r.session)
	r.body.Select(selectedRow, selectedColumn)
}

func tableRows(view tui.ResourceView, isMarked func(string) bool) [][]string {
	rows := make([][]string, 0, len(view.Rows)+1)
	headers := append([]string{"SEL"}, view.Columns...)
	rows = append(rows, headers)
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

func selectionForTable(session tui.Session) (int, int) {
	view := session.CurrentView()
	row := session.SelectedRow() + 1
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

func renderFooter(promptMode bool) string {
	prompt := "OFF"
	if promptMode {
		prompt = "ON"
	}
	return fmt.Sprintf(
		": view | / filter | ! action | Tab complete | h/j/k/l + arrows move | :ro toggle | Prompt: %s | q quit | %s",
		prompt,
		time.Now().Format("15:04:05"),
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
	if strings.TrimSpace(text) == suggestions[0] {
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
	line string,
) (string, bool) {
	parsed, err := tui.ParseExplorerInput(line)
	if err != nil {
		return fmt.Sprintf("[red]command error: %s", err.Error()), true
	}
	if shouldRecordHistory(parsed.Kind) {
		promptState.Record(line)
	}
	return runCommand(session, promptState, executor, parsed)
}

func runCommand(
	session *tui.Session,
	promptState *tui.PromptState,
	executor *runtimeActionExecutor,
	parsed tui.ExplorerCommand,
) (string, bool) {
	switch parsed.Kind {
	case tui.CommandNoop:
		return "", true
	case tui.CommandQuit:
		return "bye", false
	case tui.CommandHelp:
		return "use :vm/:lun/:cluster/:host/:datastore, :ro, /text, !action", true
	case tui.CommandReadOnly:
		applyReadOnlyMode(session, parsed.Value)
		return readOnlyStatus(*session), true
	case tui.CommandHistory:
		return historyStatus(promptState, parsed.Value), true
	case tui.CommandSuggest:
		return suggestStatus(promptState, parsed.Value, session.CurrentView()), true
	case tui.CommandLastView:
		return statusFromError(session.LastView(), "switched to last view"), true
	case tui.CommandFilter:
		session.ApplyFilter(parsed.Value)
		return fmt.Sprintf("filter: %s", parsed.Value), true
	case tui.CommandView:
		return statusFromError(session.ExecuteCommand(":"+parsed.Value), "view: "+parsed.Value), true
	case tui.CommandAction:
		if err := session.ApplyAction(parsed.Value, executor); err != nil {
			return fmt.Sprintf("[red]command error: %s", err.Error()), true
		}
		if executor.last != "" {
			return executor.last, true
		}
		return "action executed", true
	default:
		return statusFromError(session.HandleKey(parsed.Value), "key: "+parsed.Value), true
	}
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
