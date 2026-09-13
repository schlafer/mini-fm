package main

import (
	"os"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gdamore/tcell/v2"
)

// ok to tune
const (
	reservedRows        = 3                // rows of chrome (yk like chrome) (pwd line + margins) subtracted from height for the file list
	maxInputLength      = 100              // max chars in the search/filter box
	maxTextPreviewSize  = 50 * 1000        // skip syntax-highlighted preview above this big
	maxImagePreviewSize = 10 * 1000 * 1000 // skip image preview above this big
	previewReadLimit    = 10000            // bytes read for syntax highlighting
)

var (
	width                     = 20
	height                    = 20
	styleBG                   = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	styleDir                  = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorDarkCyan)
	styleDirSel               = tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite)
	styleFG                   = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	styleMID                  = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorGrey)
	showPreview               = false
	hlStyle     *chroma.Style = styles.Get("modus-vivendi")
	screen      tcell.Screen
	kittyOK     = false                   // terminal speaks the kitty graphics protocol
	inTmux      = os.Getenv("TMUX") != "" // /dev/tty is tmux's pty, not the real terminal
	ttyFile     *os.File                  // where we write kitty escapes (stdout is eval'd)

	Pwd       string
	Input     string
	Files     []os.DirEntry
	Results   []os.DirEntry
	Selected  int
	TopIndex  int
	listCache []string // cached CurrentList() cleared by invalidateList

	LastSel map[string]string // remembers the cursor per dir
	PrevDir string            // last dir we were in, for ~ toggle

	Edit     editKind   // which inline edit (rename/copy) is active, if any
	EditOrig string     // the file being renamed/copied
	EditBuf  LineEditor // the edited name/destination

	Selecting  bool            // multiselect mode is on
	Sel        map[string]bool // names marked in current dir
	LastMarked string          // most recently marked name for C-x to jump back to

	KittyShown string // path of the image currently drawn via kitty (for caching)

	ActivePrompt Prompt
)

type editKind int

// go enums look do like this
const (
	editNone editKind = iota
	editRename
	editCopy
)
