package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// copySelectedPath puts the selected entry's full path on the system clipboard
func copySelectedPath() {
	if p := SelectedPath(); p != "" {
		screen.SetClipboard([]byte(p))
	}
}

// openSelected launches the os's default opener, or runs the file directly if it's executable
func openSelected() {
	name, ok := currentName()
	if !ok {
		return
	}
	fullPath := path.Join(Pwd, name)
	os.Chdir(Pwd)

	stat, err := os.Stat(fullPath)
	isExec := err == nil && stat.Mode()&0o111 != 0

	go func(p string) {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			if isExec {
				cmd = exec.Command(p)
			} else {
				cmd = exec.Command("xdg-open", p)
			}
		case "darwin":
			if isExec {
				cmd = exec.Command(p)
			} else {
				cmd = exec.Command("open", p)
			}
		default:
			screen.Fini()
			fmt.Println("dont actually know how to open a file on your OS, pls submit an issue")
			os.Exit(0)
		}
		_ = cmd.Run()
	}(fullPath)
}

// ask for y/n confirmation, then remove selected entry
func promptDelete() {
	name, ok := currentName()
	if !ok {
		return
	}
	fullPath := path.Join(Pwd, name)

	OpenPrompt("delete "+name+"? (y/n): ", "", 0, func(input string) {
		if strings.ToLower(input) == "y" {
			os.RemoveAll(fullPath)
			SwitchDir(Pwd)
		}
	})
}

// begin inline rename of selected entry
func startRename() {
	name, ok := currentName()
	if !ok {
		return
	}
	Edit = editRename
	EditOrig = name
	EditBuf.SetText(name)
}

// begins inline copy of selected entry to new destination
func startCopy() {
	name, ok := currentName()
	if !ok {
		return
	}
	Edit = editCopy
	EditOrig = name
	EditBuf.SetText(name)

	// make room for the edit line below the source
	vh := height - reservedRows
	if Selected-TopIndex >= vh-1 {
		TopIndex++
	}
}

// enter selection mode on first press;
// on the second press it runs a bash command against everything that's been marked
func handleMultiSelect() {
	if !Selecting {
		name, ok := currentName()
		if !ok {
			return
		}
		Selecting = true
		Sel = map[string]bool{name: true}
		LastMarked = name
		MoveCursor(1)
		return
	}

	names := selectedNames()
	if len(names) == 0 {
		Selecting = false
		Sel = nil
		return
	}

	// second C-x will jump back to last marked entry before asking
	// what 2 run, so you see what youre working with
	jumpTo(LastMarked)

	token := braceList(names)

	// prefill " %" and park cursor behind the space,
	// so typing replaces selection placeholder
	OpenPrompt("bash (%=sel): ", " % ", 0, func(cmd string) {
		if strings.TrimSpace(cmd) == "" {
			return
		}
		final := cmd
		if strings.Contains(final, "%") {
			final = strings.ReplaceAll(final, "%", token)
		} else {
			final = final + " " + token
		}
		c := exec.Command("bash", "-c", final)
		c.Dir = Pwd
		_ = c.Run()
		Selecting = false
		Sel = nil
	})
}

// ask for new file/dir name (trailing "/" makes a dir),
// creating missing parent directories along the way
func promptCreate() {
	OpenPrompt("create: ", "", 0, func(name string) {
		if name == "" {
			return
		}
		fullPath := path.Join(Pwd, name)
		lastDir := fullPath
		if strings.HasSuffix(name, "/") {
			os.MkdirAll(fullPath, 0o755)
		} else {
			dir := filepath.Dir(fullPath)
			os.MkdirAll(dir, 0o755)
			lastDir = dir
			if f, err := os.Create(fullPath); err == nil {
				f.Close()
			}
		}
		SwitchDir(lastDir)
	})
}

// exit multi-select mode, or quit mini-fm entirely
func cancelOrQuit() {
	if Selecting {
		Selecting = false
		Sel = nil
		return
	}
	kittyClear()
	screen.Fini()
	os.Exit(0)
}

// mark current entry in selection mode, or open/enter it otherwise
func selectOrToggle() {
	if Selecting {
		toggleSelect()
		return
	}
	if Select() != "" {
		quitOnSelect()
	}
}

// jump to $HOME, or to / if we're already there
func toggleHome() {
	homeDir, err := os.UserHomeDir()
	targetDir := homeDir
	if err != nil || path.Clean(Pwd) == path.Clean(homeDir) {
		targetDir = path.Clean("/")
	}
	SwitchDir(path.Clean(targetDir))
}

// quit mini-fm and ask shell to open $EDITOR on selected file
func quitOnSelect() {
	kittyClear()
	screen.Fini()
	selectedPath := Select()
	if selectedPath == "" {
		os.Exit(0)
	}
	fmt.Printf("$EDITOR %s\n", escapePath(selectedPath))
	os.Exit(0)
}

// quit mini-fm and ask shell to cd into current directory or selection
func quitOnPwd() {
	kittyClear()
	screen.Fini()
	var p string
	if Input == "" {
		p = Pwd
	} else {
		items := CurrentList()
		if len(items) > 0 && Selected < len(items) {
			p = filepath.Join(Pwd, items[Selected])
		} else {
			p = Pwd
		}
	}
	fmt.Printf("cd %s\n", escapePath(p))
	os.Exit(0)
}
