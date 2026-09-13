package main

import (
	"strings"
)

func toggleSelect() {
	list := CurrentList()
	if len(list) == 0 || Selected >= len(list) {
		return
	}
	if Sel == nil {
		Sel = map[string]bool{}
	}
	name := list[Selected]
	if Sel[name] {
		delete(Sel, name)
	} else {
		Sel[name] = true
		LastMarked = name
	}
	if len(Sel) == 0 {
		Selecting = false
		Sel = nil
		return
	}
	MoveCursor(1)
}

// ret marked names in listing order, ignoring any filter
func selectedNames() []string {
	var out []string
	for _, f := range Files {
		if Sel[f.Name()] {
			out = append(out, f.Name())
		}
	}
	return out
}

// braceList makes {a,b,c} like the readme wants, or just the name for one
func braceList(names []string) string {
	if len(names) == 1 {
		return names[0]
	}
	return "{" + strings.Join(names, ",") + "}"
}
