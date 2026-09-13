package main

import (
	"github.com/gdamore/tcell/v2"
)

type LineEditor struct {
	buf    []rune // the text, as runes so multi-byte input stays intact
	cursor int    // insertion point as a rune index, 0 <= cursor <= len(buf)
}

// NewLineEditor returns an editor holding text with the cursor at pos
// OOR pos clamps into [0, len(text)]
func NewLineEditor(text string, pos int) LineEditor {
	b := []rune(text)
	if pos < 0 {
		pos = 0
	}
	if pos > len(b) {
		pos = len(b)
	}
	return LineEditor{buf: b, cursor: pos}
}

// Text returns current buffer contents
func (e *LineEditor) Text() string {
	return string(e.buf)
}

// Cursor returns insertion point as a rune index
func (e *LineEditor) Cursor() int {
	return e.cursor
}

// SetText replaces buffer and put cursor at the end
func (e *LineEditor) SetText(s string) {
	e.buf = []rune(s)
	e.cursor = len(e.buf)
}

// SetCursor moves insertion point, clamped into range
func (e *LineEditor) SetCursor(pos int) {
	e.cursor = max(0, min(pos, len(e.buf)))
}

// Insert puts r at cursor and step past it
func (e *LineEditor) Insert(r rune) {
	e.buf = append(e.buf[:e.cursor:e.cursor], append([]rune{r}, e.buf[e.cursor:]...)...)
	e.cursor++
}

// Backspace deletes rune before cursor
func (e *LineEditor) Backspace() {
	if e.cursor > 0 {
		e.buf = append(e.buf[:e.cursor-1], e.buf[e.cursor:]...)
		e.cursor--
	}
}

// Delete removes rune under cursor
func (e *LineEditor) Delete() {
	if e.cursor < len(e.buf) {
		e.buf = append(e.buf[:e.cursor], e.buf[e.cursor+1:]...)
	}
}

// MoveLeft steps cursor back n runes
func (e *LineEditor) MoveLeft(n int) {
	e.cursor = max(0, e.cursor-n)
}

// MoveRight steps cursor forward n runes
func (e *LineEditor) MoveRight(n int) {
	e.cursor = min(len(e.buf), e.cursor+n)
}

// Home jumps to the start of the line
func (e *LineEditor) Home() {
	e.cursor = 0
}

// End jumps to end of the line
func (e *LineEditor) End() {
	e.cursor = len(e.buf)
}

// KillToEnd deletes from the cursor to the end of the line (C-k)
func (e *LineEditor) KillToEnd() {
	e.buf = e.buf[:e.cursor]
}

// KillAll clears the whole line (C-u)
func (e *LineEditor) KillAll() {
	e.buf = nil
	e.cursor = 0
}

func isEditorSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

// KillWord deletes word before cursor, and any spaces between it and the cursor (like readline unix-word-rubout)
func (e *LineEditor) KillWord() {
	i := e.cursor
	for i > 0 && isEditorSpace(e.buf[i-1]) {
		i--
	}
	for i > 0 && !isEditorSpace(e.buf[i-1]) {
		i--
	}
	e.buf = append(e.buf[:i], e.buf[e.cursor:]...)
	e.cursor = i
}

// TextBeforeCursor returns buffer contents before the insertion point
func (e *LineEditor) TextBeforeCursor() string {
	return string(e.buf[:e.cursor])
}

// HandleKey feeds one key event to the editor. report whether the caller should submit (Enter) or cancel (Escape/C-c); anything else is an edit the caller just needs to redraw for

func (e *LineEditor) HandleKey(ev *tcell.EventKey) (submit, cancel bool) {
	switch ev.Key() {
	case tcell.KeyEnter:
		return true, false
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return false, true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.Backspace()
	case tcell.KeyDelete:
		e.Delete()
	case tcell.KeyLeft:
		e.MoveLeft(1)
	case tcell.KeyRight:
		e.MoveRight(1)
	case tcell.KeyHome:
		e.Home()
	case tcell.KeyEnd:
		e.End()
	case tcell.KeyCtrlA:
		e.Home()
	case tcell.KeyCtrlE:
		e.End()
	case tcell.KeyCtrlK:
		e.KillToEnd()
	case tcell.KeyCtrlU:
		e.KillAll()
	case tcell.KeyCtrlW:
		e.KillWord()
	case tcell.KeyRune:
		e.Insert(ev.Rune())
	}
	return false, false
}
