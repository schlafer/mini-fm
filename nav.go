package main

import (
	"fmt"
	"strings"
)

func togglePrevDir() {
	if PrevDir == "" {
		return
	}
	SwitchDir(PrevDir)
}

func upDir() {
	splitPwd := strings.Split(strings.TrimSuffix(Pwd, "/"), "/")
	if len(splitPwd) > 1 {
		newPwd := strings.Join(splitPwd[:len(splitPwd)-1], "/")
		SwitchDir(fmt.Sprint("/", newPwd))
	}
}

func backspace(fullWord bool) {
	if len(Input) < 1 {
		upDir()
		return
	}

	modified := Input[:len(Input)-1]
	if fullWord {
		fields := strings.Fields(Input)
		if len(fields) > 0 {
			fields = fields[:len(fields)-1]
		}
		modified = strings.Join(fields, " ")
	}

	results := search(modified)
	Input = modified
	if len(results) == 0 {
		Results = nil
	} else {
		Results = results
	}
	invalidateList()
	Selected = 0
	TopIndex = 0
}

func doInput(r rune) {
	if len(Input) >= maxInputLength {
		return
	}

	modified := Input + string(r)
	results := search(modified)

	if len(results) == 0 {
		return
	}

	Input = modified
	Results = results
	invalidateList()
	Selected = 0
	TopIndex = 0
}
