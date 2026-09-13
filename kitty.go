package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/SerenaFontaine/kgp"
	"golang.org/x/image/draw"
	"golang.org/x/sys/unix"
)

// holds png bytes already resized for a given placement,
// so scrolling back to an image at the same size is free
type kittyCacheEntry struct {
	mtime int64
	size  int64
	png   []byte
}

var kittyCache = map[string]kittyCacheEntry{}

// previewImagePath returns the selected file if it's an image we can show, else ""
func previewImagePath() string {
	if !kittyOK || !showPreview {
		return ""
	}
	files := Files
	if len(Results) > 0 {
		files = Results
	}
	if len(files) == 0 || Selected < 0 || Selected >= len(files) {
		return ""
	}
	entry := files[Selected]
	full := path.Join(Pwd, entry.Name())
	if isDirEntry(full, entry) {
		return ""
	}
	info, err := entry.Info()
	if err != nil || info.Size() == 0 || info.Size() > maxImagePreviewSize {
		return ""
	}
	f, err := os.Open(full)
	if err != nil {
		return ""
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if strings.HasPrefix(http.DetectContentType(buf[:n]), "image/") {
		return full
	}
	return ""
}

// reconcileKitty draws want in the preview pane, or clears it, only when it changes
func reconcileKitty(want string) {
	if !kittyOK || want == KittyShown {
		return
	}
	if KittyShown != "" {
		kittyClear()
	}
	KittyShown = ""
	if want == "" {
		return
	}

	data, err := os.ReadFile(want)
	if err != nil {
		return
	}
	// pull the pixel dims before resizing, to keep the aspect ratio
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return
	}

	cols := width - width/2 - 1
	rows := height - 1
	if cols < 1 || rows < 1 {
		return
	}
	// fit into the pane without stretching, then place at its top-left (1-based)
	c, r := fitCells(cfg.Width, cfg.Height, cols, rows)
	data, err = kittyPNG(want, data, cfg.Width, cfg.Height, c, r)
	if err != nil {
		return
	}

	kittyPlace(data, width/2+1, 1, c, r)
	KittyShown = want
}

// ret: png bytes of want resized to fit a c x r cell placement
// small pngs that already fit are passed through untouched
// anything larger or non-png is decoded, downscaled to pane pixels,
// and re-encoded, because rendering megabytes is really really slow
//
// results are cached by path + mtime + size + placement geometry
func kittyPNG(imgPath string, data []byte, imgW, imgH, cols, rows int) ([]byte, error) {
	cw, ch := cellSize()
	tw, th := cols*cw, rows*ch

	st, err := os.Stat(imgPath)
	if err != nil {
		return nil, err
	}
	key := strings.Join([]string{
		imgPath,
		fmt.Sprint(st.ModTime().UnixNano()),
		fmt.Sprint(st.Size()),
		fmt.Sprint(cols), fmt.Sprint(rows),
		fmt.Sprint(cw), fmt.Sprint(ch),
	}, "\x00")
	if e, ok := kittyCache[key]; ok && e.mtime == st.ModTime().UnixNano() && e.size == st.Size() {
		return e.png, nil
	}

	// happy path path:
	// a png that already fits the pane needs no pixel work
	if isPNG(data) && imgW <= tw && imgH <= th {
		return data, nil
	}

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	b := src.Bounds()
	imgW, imgH = b.Dx(), b.Dy()

	scale := math.Min(float64(tw)/float64(imgW), float64(th)/float64(imgH))
	if scale <= 0 {
		return nil, fmt.Errorf("bad placement size")
	}
	if scale > 1 {
		// the terminal scales up for zero moneys
		scale = 1
	}
	dw, dh := max(int(math.Round(float64(imgW)*scale)), 1), max(int(math.Round(float64(imgH)*scale)), 1)

	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	out := buf.Bytes()

	if len(kittyCache) > 32 {
		clear(kittyCache)
	}
	kittyCache[key] = kittyCacheEntry{mtime: st.ModTime().UnixNano(), size: st.Size(), png: out}
	return out, nil
}

func isPNG(data []byte) bool {
	return len(data) > 8 && string(data[1:4]) == "PNG"
}

// fitCells shrinks imgW x imgH (pixels) into at most cols x rows cells, keeping aspect
// without this function shit gets shrinked and looks so fucking funny lmao.
// the main limitation of this is that as it's going to be an approximation however
// hard you try and fit it, just by the nature of having columns and rows so
// ¯\_(ツ)_/¯
func fitCells(imgW, imgH, cols, rows int) (int, int) {
	if imgW < 1 || imgH < 1 {
		return cols, rows
	}
	cw, ch := cellSize()
	scale := math.Min(float64(cols*cw)/float64(imgW), float64(rows*ch)/float64(imgH))
	c := int(math.Round(float64(imgW) * scale / float64(cw)))
	r := int(math.Round(float64(imgH) * scale / float64(ch)))
	// clamp into [1, pane]
	c = min(max(c, 1), cols)
	r = min(max(r, 1), rows)
	return c, r
}

// cellSize asks the terminal for its cell size in pixels, falling back to a ~1:2 guess
func cellSize() (int, int) {
	cw, ch := 10, 20
	if ttyFile == nil {
		return cw, ch
	}
	ws, err := unix.IoctlGetWinsize(int(ttyFile.Fd()), unix.TIOCGWINSZ)
	if err == nil && ws.Xpixel > 0 && ws.Ypixel > 0 && ws.Col > 0 && ws.Row > 0 {
		cw = int(ws.Xpixel) / int(ws.Col)
		ch = int(ws.Ypixel) / int(ws.Row)
	}
	return cw, ch
}

// wraps a kitty graphics escape sequence for tmux's dcs passthrough
// tmux will just drop them otherwise because for no reason other than it hates us and mr. goyal
func wrapTmux(seq string) string {
	if !inTmux {
		return seq
	}
	escaped := strings.ReplaceAll(seq, "\x1b", "\x1b\x1b")
	return "\x1bPtmux;" + escaped + "\x1b\\"
}

func kittyClear() {
	if !kittyOK || ttyFile == nil {
		return
	} else {
		fmt.Fprint(ttyFile, wrapTmux(kgp.DeleteAllFree().Encode()))
	}
}

// kittyPlace transmits a png and displays it, scaled into cols x rows cells at (col,row)
func kittyPlace(png []byte, col, row, cols, rows int) {
	if ttyFile == nil {
		return
	}
	cmd := kgp.NewTransmitDisplay().
		Format(kgp.FormatPNG).
		TransmitDirect(png).
		DisplaySize(cols, rows).
		// if we let the terminal move it past the image itll render the image at 0, 0 sometimes
		CursorMovement(false).
		// otherwise it might pollute tcell
		ResponseSuppression(kgp.ResponseOKOnly).
		Build()

	// tmux chokes on very long passthrough payloads so use smaller chunks there
	// both satisfy kgp's EncodeChunked which is <=4096, divisible by 4
	chunk := 4096
	if inTmux {
		chunk = 1024
	}

	// this HAS to stay one buffered flush
	w := bufio.NewWriter(ttyFile)
	fmt.Fprintf(w, "\x1b[%d;%dH", row, col)
	for _, seq := range cmd.EncodeChunked(chunk) {
		w.WriteString(wrapTmux(seq))
	}
	w.Flush()
}
