package main

import (
	"fmt"
	"io"
	"os"
	"path"
)

// copies a file or a whole dir, keeping perms
func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyPath(path.Join(src, e.Name()), path.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Chmod(info.Mode())
}

func SelectedPath() string {
	name, ok := currentName()
	if !ok {
		return ""
	}
	return path.Join(Pwd, name)
}

func Select() string {
	var list []os.DirEntry
	if len(Results) > 0 {
		list = Results
	} else {
		list = Files
	}

	if len(list) == 0 {
		return ""
	}

	selected := list[Selected]

	if isDirEntry(path.Join(Pwd, selected.Name()), selected) {
		SwitchDir(path.Join(Pwd, selected.Name()))
		return ""
	}
	return path.Join(Pwd, selected.Name())
}

func SwitchDir(where string) error {
	if where == "" {
		return fmt.Errorf("cannot switch to empty directory")
	}

	cleanPath := path.Clean(where)
	if !path.IsAbs(cleanPath) {
		return fmt.Errorf("must provide an absolute path")
	}

	// remember where the cursor was in the dir we're leaving
	if LastSel == nil {
		LastSel = make(map[string]string)
	}
	if list := CurrentList(); len(list) > 0 && Selected < len(list) {
		LastSel[Pwd] = list[Selected]
	}

	// read first so a dir we cant open doesnt leave us in a broken state
	newPwd := cleanPath + "/"
	files, err := os.ReadDir(newPwd)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", newPwd, err)
	}

	// remember the dir we came from so ~ can jump back
	if Pwd != "" && Pwd != newPwd {
		PrevDir = Pwd
	}

	Pwd = newPwd
	Files = files
	Input = ""
	Selected = 0
	TopIndex = 0
	Results = nil
	invalidateList()

	// retain the last selection when coming back to this dir
	if name, ok := LastSel[Pwd]; ok {
		for i, f := range files {
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
	return nil
}

// makes sure symlinks to directories work right
func isDirEntry(path string, entry os.DirEntry) bool {
	info, err := entry.Info()
	if err != nil {
		return false
	}

	if info.IsDir() {
		return true
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Stat(path)
		if err == nil && target.IsDir() {
			return true
		}
	}
	return false
}
