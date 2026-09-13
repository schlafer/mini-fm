//
// mini-fm: https://github.com/schlafer/mini-fm
// TUI file manager
//

package main

import (
	"flag"
	_ "image/gif"
	_ "image/jpeg"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func main() {
	flag.BoolVar(&showPreview, "preview", true, "show a file preview on the right side")
	flag.BoolVar(&showPreview, "p", true, "alias for -preview")
	flag.Parse()

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	screen = s
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	width, height = screen.Size()

	// kitty images go to /dev/tty since stdout gets eval'd by the shell
	ttyFile, _ = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	term := os.Getenv("TERM")
	kittyOK = ttyFile != nil && (term == "xterm-kitty" ||
		strings.Contains(term, "ghostty") ||
		os.Getenv("KITTY_WINDOW_ID") != "" ||
		os.Getenv("GHOSTTY_RESOURCES_DIR") != "" ||
		os.Getenv("WEZTERM_PANE") != "" ||
		// tmux hides the outer terminals env vars and rewrites TERM,
		// but forwards kitty graphics if `set -g allow-passthrough on` is set
		// everyone has this enabled already
		inTmux)

	screen.SetStyle(styleBG)

	screen.Clear()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err, "getpwd")
	}

	SwitchDir(cwd)
	Redraw()

	for {
		screen.Show()

		ev := screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			width, height = screen.Size()
			// force the image to be redrawn at the new size
			if KittyShown != "" {
				kittyClear()
				KittyShown = ""
			}
			screen.Sync()
			Redraw()
		case *tcell.EventKey:
			if ActivePrompt.IsActive {
				HandlePromptInput(ev)
				Redraw()
				continue
			}
			if Edit != editNone {
				HandleEditInput(ev)
				Redraw()
				continue
			}
			HandleKey(ev)
			Redraw()
		}
	}
}
