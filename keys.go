package main

import (
	"github.com/gdamore/tcell/v2"
)

// HandleKey dispatches a key event in normal mode (no prompt or inline edit active)
func HandleKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyCtrlS:
		copySelectedPath()
	case tcell.KeyCtrlO:
		openSelected()
	case tcell.KeyCtrlD:
		promptDelete()
	case tcell.KeyCtrlR:
		startRename()
	case tcell.KeyCtrlY:
		startCopy()
	case tcell.KeyCtrlX:
		handleMultiSelect()
	case tcell.KeyCtrlA:
		promptCreate()
	case tcell.KeyEscape, tcell.KeyCtrlC:
		cancelOrQuit()
	case tcell.KeyDown, tcell.KeyCtrlJ, tcell.KeyCtrlN:
		MoveCursor(1)
	case tcell.KeyUp, tcell.KeyCtrlK, tcell.KeyCtrlP:
		MoveCursor(-1)
	case tcell.KeyTab, tcell.KeyRight, tcell.KeyCtrlL, tcell.KeyCtrlF:
		selectOrToggle()
	case tcell.KeyEnter:
		quitOnPwd()
	// KeyCtrlH is the same code as backspace, and the actual backspace is KeyBackspace2
	case tcell.KeyCtrlB, tcell.KeyCtrlH, tcell.KeyLeft:
		upDir()
	case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
		backspace(false)
	case tcell.KeyCtrlW:
		backspace(true)
	case tcell.KeyCtrlE:
		toggleHome()
	case tcell.KeyRune:
		// ~ jumps to the last dir, but only when not mid-search
		if ev.Rune() == '~' && Input == "" {
			togglePrevDir()
		} else {
			doInput(ev.Rune())
		}
	}
}
