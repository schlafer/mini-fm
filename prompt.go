package main

import (
	"github.com/gdamore/tcell/v2"
)

type Prompt struct {
	IsActive bool
	Label    string
	Input    LineEditor
	OnSubmit func(string)
}

func OpenPrompt(label, initial string, cursor int, onSubmit func(string)) {
	ActivePrompt = Prompt{
		IsActive: true,
		Label:    label,
		Input:    NewLineEditor(initial, cursor),
		OnSubmit: onSubmit,
	}
}

func HandlePromptInput(ev *tcell.EventKey) {
	submit, cancel := ActivePrompt.Input.HandleKey(ev)
	switch {
	case submit:
		text := ActivePrompt.Input.Text()
		ActivePrompt.OnSubmit(text)
		ActivePrompt.IsActive = false
		SwitchDir(Pwd)
		screen.HideCursor()
	case cancel:
		ActivePrompt.IsActive = false
		Selecting = false
		screen.HideCursor()
		Redraw()
	}
}
