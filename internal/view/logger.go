package view

import (
	"context"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
)

// Logger represents a generic log viewer.
type Logger struct {
	*tview.TextView

	actions        ui.KeyActions
	app            *App
	title, subject string
	cmdBuff        *model.FishBuff
}

// NewLogger returns a logger viewer.
func NewLogger(app *App) *Logger {
	return &Logger{
		TextView: tview.NewTextView(),
		app:      app,
		actions:  make(ui.KeyActions),
		cmdBuff:  model.NewFishBuff('/', model.FilterBuffer),
	}
}

// Init initializes the viewer.
func (l *Logger) Init(_ context.Context) error {
	if l.title != "" {
		l.SetBorder(true)
	}
	l.SetScrollable(true).SetWrap(true).SetRegions(true)
	l.SetDynamicColors(true)
	l.SetHighlightColor(tcell.ColorOrange)
	l.SetTitleColor(tcell.ColorAqua)
	l.SetInputCapture(l.keyboard)
	l.SetBorderPadding(0, 0, 1, 1)

	l.app.Styles.AddListener(l)
	l.StylesChanged(l.app.Styles)

	l.app.Prompt().SetModel(l.cmdBuff)
	l.cmdBuff.AddListener(l)

	l.bindKeys()
	l.SetInputCapture(l.keyboard)

	return nil
}

// BufferChanged indicates the buffer was changed.
func (l *Logger) BufferChanged(s string) {}

// BufferCompleted indicates input was accepted.
func (l *Logger) BufferCompleted(s string) {
}

// BufferActive indicates the buff activity changed.
func (l *Logger) BufferActive(state bool, k model.BufferKind) {
	l.app.BufferActive(state, k)
}

func (l *Logger) bindKeys() {
	l.actions.Set(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Back", l.resetCmd, false),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", l.saveCmd, false),
		ui.KeyC:         ui.NewKeyAction("Copy", l.cpCmd, true),
		ui.KeySlash:     ui.NewSharedKeyAction("Filter Mode", l.activateCmd, false),
		tcell.KeyDelete: ui.NewSharedKeyAction("Erase", l.eraseCmd, false),
	})
}

func (l *Logger) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if a, ok := l.actions[ui.AsKey(evt)]; ok {
		return a.Action(evt)
	}

	return evt
}

// StylesChanged notifies the skin changed.
func (l *Logger) StylesChanged(s *config.Styles) {
	l.SetBackgroundColor(l.app.Styles.BgColor())
	l.SetTextColor(l.app.Styles.FgColor())
	l.SetBorderFocusColor(l.app.Styles.Frame().Border.FocusColor.Color())
}

// SetSubject updates the subject.
func (l *Logger) SetSubject(s string) {
	l.subject = s
}

// Actions returns menu actions
func (l *Logger) Actions() ui.KeyActions {
	return l.actions
}

// Name returns the component name.
func (l *Logger) Name() string { return l.title }

// Start starts the view updater.
func (l *Logger) Start() {}

// Stop terminates the updater.
func (l *Logger) Stop() {
	l.app.Styles.RemoveListener(l)
}

// Hints returns menu hints.
func (l *Logger) Hints() model.MenuHints {
	return l.actions.Hints()
}

// ExtraHints returns additional hints.
func (l *Logger) ExtraHints() map[string]string {
	return nil
}

func (l *Logger) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if l.app.InCmdMode() {
		return evt
	}
	l.app.ResetPrompt(l.cmdBuff)

	return nil
}

func (l *Logger) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !l.cmdBuff.IsActive() {
		return nil
	}
	l.cmdBuff.Delete()

	return nil
}

func (l *Logger) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !l.cmdBuff.InCmdMode() {
		l.cmdBuff.Reset()
		return l.app.PrevCmd(evt)
	}
	l.cmdBuff.SetActive(false)
	l.cmdBuff.Reset()

	return nil
}

func (l *Logger) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveYAML(l.app.Config.K9s.CurrentCluster, l.title, l.GetText(true)); err != nil {
		l.app.Flash().Err(err)
	} else {
		l.app.Flash().Infof("Log %s saved successfully!", path)
	}

	return nil
}

func (l *Logger) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	l.app.Flash().Info("Content copied to clipboard...")
	if err := clipboard.WriteAll(l.GetText(true)); err != nil {
		l.app.Flash().Err(err)
	}

	return nil
}
