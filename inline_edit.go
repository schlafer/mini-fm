package main

import (
	"os"
	"path"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

// rename (C-r) and copy (C-y) are both "edit a name inline, then apply it
// to the filesystem", so they share one state machine distinguished by Edit

// HandleEditInput feeds a key event to the active inline edit (rename or copy)
func HandleEditInput(ev *tcell.EventKey) {
	submit, cancel := EditBuf.HandleKey(ev)
	switch {
	case submit:
		commitEdit()
	case cancel:
		Edit = editNone
		screen.HideCursor()
	}
}

// apply active rename or copy and reselects result
// leave original untouched if target dir can't be created
// or underlying fs operation fails
func commitEdit() {
	kind := Edit
	Edit = editNone
	screen.HideCursor()

	name := EditBuf.Text()
	if name == "" || name == EditOrig {
		return
	}

	srcPath := path.Join(Pwd, EditOrig)
	dstPath := path.Join(Pwd, name)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil { // lets you move/copy by typing a/b/c
		return
	}

	var err error
	switch kind {
	case editRename:
		err = os.Rename(srcPath, dstPath)
	case editCopy:
		err = copyPath(srcPath, dstPath)
	}
	if err != nil {
		return
	}

	SwitchDir(Pwd)
	selectByName(filepath.Base(name))
}

// put cursor on entry called name in current dir, if present
func selectByName(name string) {
	for i, f := range Files {
		if f.Name() == name {
			Selected = i
			break
		}
	}
	visibleHeight := height - reservedRows
	if Selected >= visibleHeight {
		TopIndex = Selected - visibleHeight + 1
	}
}
