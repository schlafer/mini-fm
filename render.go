package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func Redraw() {
	wantImg := previewImagePath()
	defer reconcileKitty(wantImg) // emit after tcell has flushed, so it lands on top
	screen.Clear()

	files := Files
	if len(Results) > 0 {
		files = Results
	}

	if len(files) == 0 {
		DrawFiles()
		screen.Show()
		return
	}

	selectedEntry := files[Selected]
	fullPath := path.Join(Pwd, selectedEntry.Name())

	if showPreview && wantImg == "" {
		if isDirEntry(fullPath, selectedEntry) {
			DrawDirPreview(fullPath, width/2, 0, width-1, height-1)
		} else {
			info, err := selectedEntry.Info()
			drawWarning := func(text string) {
				drawText(width/2+2, 2, width-1, 2, styleMID, text)
			}

			// dont do for 50kB+
			if err != nil || info.Size() > maxTextPreviewSize {
				DrawFiles()
				drawWarning("*file too large (or cant be opened)*")
				return
			}
			if info.Size() == 0 {
				DrawFiles()
				drawWarning("*file empty*")
				return
			}

			file, err := os.Open(fullPath)
			if err != nil {
				DrawFiles()
				drawWarning("*file cant be opened*")
				return
			}
			defer file.Close()

			// kinda hacky but works
			buffer := make([]byte, 512)
			n, _ := file.Read(buffer)
			file.Seek(0, 0)

			contentType := http.DetectContentType(buffer[:n])

			if strings.HasPrefix(contentType, "text/") || contentType == "application/javascript" || contentType == "application/json" {
				DrawFilePreview(file, width/2, 0, width-1, height-1)
			} else {
				drawWarning(contentType)
			}
		}
	}
	DrawFiles()
	screen.Show()
}

func DrawFilePreview(handle *os.File, x1, y1, x2, y2 int) {
	content, err := io.ReadAll(io.LimitReader(handle, previewReadLimit))
	if err != nil {
		return
	}

	lexer := lexers.Match(handle.Name())
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := hlStyle
	if style == nil {
		style = styles.Fallback
	}

	iterator, err := lexer.Tokenise(nil, string(content))
	if err != nil {
		return
	}

	x, y := x1, y1
	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)

		tcellStyle := tcell.StyleDefault.
			// they map directly
			Foreground(tcell.NewRGBColor(int32(entry.Colour.Red()), int32(entry.Colour.Green()), int32(entry.Colour.Blue()))).
			Background(tcell.ColorReset)

		if entry.Bold == chroma.Yes {
			tcellStyle = tcellStyle.Bold(true)
		}

		// draw each token now
		for _, r := range token.Value {
			if r == '\n' {
				x = x1
				y++
				if y > y2 {
					return
				}
				continue
			}
			if x <= x2 {
				screen.SetContent(x, y, r, nil, tcellStyle)
				x++
			}
		}
	}
}

func DrawDirPreview(fullPath string, x1, y1, x2, y2 int) {
	dirEntries, err := os.ReadDir(fullPath)
	if err != nil {
		return
	}

	for y, entry := range dirEntries {
		isDir := isDirEntry(path.Join(fullPath, entry.Name()), entry)
		if isDir {
			drawText(x1, y, x2, y, styleDir, entry.Name()+"/")
		} else {
			drawText(x1, y, x2, y, styleBG, entry.Name())
		}
		if y >= y2 {
			break
		}
	}
}

func DrawFiles() {
	filesToShow := CurrentList()

	pwdLen := len(Pwd) + 1
	drawText(1, 1, pwdLen, 1, styleBG, Pwd)

	if len(filesToShow) > 0 && Selected < len(filesToShow) {
		drawText(pwdLen, 1, 999, 1, styleMID, filesToShow[Selected])
	}

	drawText(pwdLen, 1, 999, 1, styleBG, Input)

	// one slot, top-left above the input row: file position in normal
	// mode, selection count in selection mode
	if Selecting {
		selInfo := fmt.Sprintf("[%d/%d] selected", len(Sel), len(filesToShow))
		drawText(1, 0, 999, 0, styleDir, selInfo)
	} else {
		scrollInfo := fmt.Sprintf("[%d/%d]", Selected+1, len(filesToShow))
		drawText(1, 0, 999, 0, styleMID, scrollInfo)
	}

	if len(filesToShow) == 0 {
		drawText(1, 2, 999, 3, styleMID, "*nothing here*")
		screen.Show()
		return
	}

	Selected = min(Selected, len(filesToShow)-1)
	TopIndex = min(TopIndex, Selected)

	visibleHeight := height - reservedRows
	start := TopIndex
	end := min(start+visibleHeight, len(filesToShow))

	copyShift := 0
	// inline editors arre in the file list (left) panel; preview
	// pane starts at width/2, so clamp them short of it (TODO handle panel being toggled)
	editMaxX := max(width/2-1, 2)
	for i := start; i < end; i++ {
		y := i - start + 2 + copyShift
		name := filesToShow[i]

		// inline rename editor on the selected row
		if Edit == editRename && Selected == i {
			for x := 1; x <= editMaxX; x++ {
				screen.SetContent(x, y, ' ', nil, styleBG)
			}
			drawLineEditor(1, y, editMaxX, styleFG, &EditBuf)
			continue
		}

		style := styleBG

		isDir := false
		if len(Results) > 0 {
			if i < len(Results) {
				fullPath := path.Join(Pwd, Results[i].Name())
				isDir = isDirEntry(fullPath, Results[i])
			}
		} else if i < len(Files) {
			fullPath := path.Join(Pwd, Files[i].Name())
			isDir = isDirEntry(fullPath, Files[i])
		}

		if Selected == i {
			style = styleFG
		}
		if isDir {
			style = styleDir
			if Selected == i {
				style = styleDirSel
			}
			name += "/"
		}

		// mark selected files with a * in the left gutter
		if Selecting && Sel[filesToShow[i]] {
			screen.SetContent(0, y, '*', nil, styleFG)
		}

		drawText(1, y, 999, y, style, name)

		// on copy, edit the destination path on a new line below the source
		if Edit == editCopy && Selected == i {
			ey := y + 1
			for x := 1; x <= editMaxX; x++ {
				screen.SetContent(x, ey, ' ', nil, styleBG)
			}
			drawLineEditor(1, ey, editMaxX, styleFG, &EditBuf)
			copyShift = 1
		}
	}

	if ActivePrompt.IsActive {
		input := ActivePrompt.Input.Text()
		label := ActivePrompt.Label

		for i := 0; i < width; i++ {
			screen.SetContent(i, 1, ' ', nil, styleBG)
		}

		drawText(1, 1, len(label), 1, styleFG, label)

		lastSlash := strings.LastIndex(input, "/")

		currentX := len(label) + 1

		if lastSlash != -1 {
			dirPart := input[:lastSlash+1]
			filePart := input[lastSlash+1:]

			styleDir := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorBlue)
			drawText(currentX, 1, currentX+len(dirPart), 1, styleDir, dirPart)

			if filePart != "" {
				drawText(currentX+len(dirPart), 1, width-1, 1, styleBG, filePart)
			}
		} else {
			drawText(currentX, 1, width-1, 1, styleBG, input)
		}

		screen.ShowCursor(len(label)+runewidth.StringWidth(ActivePrompt.Input.TextBeforeCursor())+1, 1)
	}
}

// render ed on row y from column x to maxX, scrolling
// horizontally so the cursor stays visible, and shows the cursor at the right cell
func drawLineEditor(x, y, maxX int, style tcell.Style, ed *LineEditor) {
	avail := maxX - x + 1
	if avail < 1 {
		return
	}
	// first visible rune! slide right until the cursor fits on screen
	start := 0
	for start < ed.cursor && runewidth.StringWidth(string(ed.buf[start:ed.cursor])) > avail-1 {
		start++
	}
	col := x
	for i := start; i < len(ed.buf) && col <= maxX; i++ {
		w := runewidth.RuneWidth(ed.buf[i])
		if col+w-1 > maxX {
			break
		}
		screen.SetContent(col, y, ed.buf[i], nil, style)
		col += w
	}
	screen.ShowCursor(x+runewidth.StringWidth(string(ed.buf[start:ed.cursor])), y)
}

func drawText(x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		screen.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}
