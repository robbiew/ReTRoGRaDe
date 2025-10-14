package tui

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// overlayString places a string onto the canvas at the given position
func (m *Model) overlayString(canvas []string, str string, startRow, startCol int) {
	lines := strings.Split(str, "\n")

	for i, line := range lines {
		row := startRow + i
		if row < 0 || row >= len(canvas) {
			continue
		}

		if startCol <= 0 {
			// Replace entire line and pad to full screen width to avoid
			// bleed from prior content on that row.
			pad := m.screenWidth - m.visualWidth(line)
			if pad > 0 {
				line = line + strings.Repeat(" ", pad)
			}
			canvas[row] = line
			continue
		}

		// Overlay only the content region, preserving existing left/right
		lineVisualWidth := m.visualWidth(line)
		left, _, right := splitByVisibleColumns(canvas[row], startCol, startCol+lineVisualWidth)
		canvas[row] = left + line + right
	}
}

// overlayArtBlock writes an art block with explicit left/right texture so
// styles do not bleed into or out of the art content.
func (m *Model) overlayArtBlock(canvas []string, lines []string, startRow, startCol, artWidth int) {
	// Ensure each line respects artWidth; then overlay, preserving existing
	// left/right portions of the canvas.
	fixed := make([]string, len(lines))
	for i, line := range lines {
		if w := m.visualWidth(line); w > artWidth {
			line = trimToVisibleWidth(line, artWidth)
		}
		fixed[i] = line
	}
	m.overlayString(canvas, strings.Join(fixed, "\n"), startRow, startCol)
}

// trimToVisibleWidth returns a prefix of s with at most target columns, keeping SGR.
func trimToVisibleWidth(s string, target int) string {
	if target <= 0 {
		return ""
	}
	var b strings.Builder
	vis := 0
	for i := 0; i < len(s) && vis < target; {
		if s[i] == '\x1b' {
			j := i + 1
			if j < len(s) && s[j] == '[' {
				j++
				for j < len(s) {
					if s[j] >= '@' && s[j] <= '~' { // final byte
						j++
						break
					}
					j++
				}
				b.WriteString(s[i:j])
				i = j
				continue
			}
		}
		_, sz := utf8.DecodeRuneInString(s[i:])
		if sz <= 0 {
			sz = 1
		}
		b.WriteString(s[i : i+sz])
		i += sz
		vis++
	}
	return b.String()
}

// splitByVisibleColumns splits a string into left/mid/right by visible column
// positions [start,end). ANSI SGR sequences are treated as zero-width and kept
// in their respective segments. If indexes are out of range, they are clamped.
func splitByVisibleColumns(s string, start, end int) (string, string, string) {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}

	var leftBuf, midBuf, rightBuf strings.Builder
	vis := 0
	i := 0
	for i < len(s) {
		// Handle ANSI SGR sequences: \x1b[ ... m
		if s[i] == '\x1b' {
			j := i + 1
			if j < len(s) && s[j] == '[' {
				j++
				for j < len(s) {
					if s[j] == 'm' {
						j++
						break
					}
					j++
				}
			}
			// Append the entire escape sequence to whichever segment we're in
			if vis < start {
				leftBuf.WriteString(s[i:j])
			} else if vis < end {
				midBuf.WriteString(s[i:j])
			} else {
				rightBuf.WriteString(s[i:j])
			}
			i = j
			continue
		}

		// Decode one rune
		_, size := utf8.DecodeRuneInString(s[i:])

		// Append rune to proper segment
		if vis < start {
			leftBuf.WriteString(s[i : i+size])
		} else if vis < end {
			midBuf.WriteString(s[i : i+size])
		} else {
			rightBuf.WriteString(s[i : i+size])
		}

		vis++
		i += size
	}

	return leftBuf.String(), midBuf.String(), rightBuf.String()
}

func (m *Model) overlayStringCenteredWithClear(canvas []string, str string) {
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return
	}

	// Find the widest line
	maxWidth := 0
	for _, line := range lines {
		width := m.visualWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	startRow := (m.screenHeight - len(lines)) / 2
	if startRow < 0 {
		startRow = 0
	}
	startCol := (m.screenWidth - maxWidth) / 2
	if startCol < 0 {
		startCol = 0
	}

	// Clear a 1-cell border around the modal to provide crisp visual
	// separation from any ANSI background. This clears one row/column
	// outside the modal content on all sides.
	border := 1
	// Strong clear: wipe full screen width across the vertical span covering
	// modal + 1-row border to eliminate any ANSI bleed on the sides.
	totalHeight := len(lines) + 2*border
	m.clearRect(canvas, startRow-border, 0, m.screenWidth, totalHeight)
	// Additionally clear a couple full-width rows just below the modal to
	// avoid any background bleed immediately under the box.
	m.clearRect(canvas, startRow+len(lines), 0, m.screenWidth, 3)

	// Overlay content
	m.overlayString(canvas, str, startRow, startCol)
}

// overlayStringWithBorderClear clears a 1-cell border around the content box
// and then overlays the string at the given row/col. Useful for menus/lists
// drawn over an ANSI background.
// overlayStringWithBorderClear clears a 1-cell border around the content box
// then overlays the content. For finer control, use overlayStringWithClearBorder.
func (m *Model) overlayStringWithBorderClear(canvas []string, str string, startRow, startCol int) {
	m.overlayStringWithClearBorder(canvas, str, startRow, startCol, 1)
}

// overlayStringWithClearBorder clears a configurable border of spaces around
// the content area before drawing. This helps ensure crisp edges over ANSI art.
func (m *Model) overlayStringWithClearBorder(canvas []string, str string, startRow, startCol, border int) {
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return
	}
	maxWidth := 0
	for _, line := range lines {
		if w := m.visualWidth(line); w > maxWidth {
			maxWidth = w
		}
	}
	if border < 0 {
		border = 0
	}
	m.clearRect(canvas, startRow-border, startCol-border, maxWidth+2*border, len(lines)+2*border)
	m.overlayString(canvas, str, startRow, startCol)
}

// clearRect fills a rectangular region with plain spaces to neutralize background
func (m *Model) clearRect(canvas []string, startRow, startCol, width, height int) {
	for r := 0; r < height; r++ {
		row := startRow + r
		if row < 0 || row >= len(canvas) {
			continue
		}
		// Build line with spaces in [startCol, startCol+width)
		left, _, right := splitByVisibleColumns(canvas[row], startCol, startCol+width)
		mid := strings.Repeat(" ", max(0, min(width, m.screenWidth-startCol)))
		canvas[row] = left + mid + right
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// overlayStringCentered places a string in the center of the canvas
func (m *Model) overlayStringCentered(canvas []string, str string) {
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return
	}

	// Find the widest line (accounting for ANSI codes)
	maxWidth := 0
	for _, line := range lines {
		width := m.visualWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	startRow := (m.screenHeight - len(lines)) / 2
	if startRow < 0 {
		startRow = 0
	}
	startCol := (m.screenWidth - maxWidth) / 2
	if startCol < 0 {
		startCol = 0
	}

	m.overlayString(canvas, str, startRow, startCol)
}

// canvasToString converts the canvas back to a string
func (m *Model) canvasToString(canvas []string) string {
	return strings.Join(canvas, "\n")
}

// visualWidth calculates the display width of a string (excluding ANSI codes)
func (m *Model) visualWidth(s string) int {
	// Remove ANSI escape sequences
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	stripped := ansiPattern.ReplaceAllString(s, "")
	return len([]rune(stripped))
}

// ============================================================================
// ANSI Art Loading (CP437 -> UTF-8)
// ============================================================================

// LoadANSIArtCP437 loads an ANSI art file (CP437-encoded) and stores
// it as padded lines on the model for rendering under the menus.
// Lines are padded to 80 columns by visible width; no truncation is performed.
func (m *Model) LoadANSIArtCP437(path string) error {
	// Read raw bytes
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Decode CP437 -> UTF-8
	rdr := transform.NewReader(bytes.NewReader(data), charmap.CodePage437.NewDecoder())
	decoded, err := io.ReadAll(rdr)
	if err != nil {
		return err
	}

	// Normalize newlines
	s := strings.ReplaceAll(string(decoded), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// Rasterize ANSI (supports SGR colors and basic cursor ops) to 80x25
	m.ansiArtLines = rasterizeANSIToLines(s, 80, 25)
	return nil
}

// rasterizeANSIToLines interprets a subset of ANSI (SGR colors, CSI H, J, K)
// and produces a fixed-size array of lines containing only text and SGR.
func rasterizeANSIToLines(s string, width, height int) []string {
	type style struct {
		fg, bg int
		bold   bool
	}
	// 39/49 mean default
	cur := style{fg: 39, bg: 49, bold: false}
	// canvas
	type cell struct {
		ch rune
		st style
	}
	canvas := make([][]cell, height)
	for y := 0; y < height; y++ {
		canvas[y] = make([]cell, width)
		for x := 0; x < width; x++ {
			canvas[y][x] = cell{' ', cur}
		}
	}
	x, y := 0, 0

	writeRune := func(r rune) {
		if r == '\n' {
			y++
			x = 0
			return
		}
		if r == '\r' {
			x = 0
			return
		}
		if y < 0 || y >= height {
			return
		}
		if x < 0 {
			x = 0
		}
		if x >= width {
			// wrap
			y++
			x = 0
			if y >= height {
				return
			}
		}
		canvas[y][x] = cell{r, cur}
		x++
	}

	// helpers
	resetStyle := func() { cur = style{fg: 39, bg: 49, bold: false} }
	setSGR := func(params []int) {
		if len(params) == 0 {
			resetStyle()
			return
		}
		for _, p := range params {
			switch {
			case p == 0:
				resetStyle()
			case p == 1:
				cur.bold = true
			case p == 22:
				cur.bold = false
			case p == 39:
				cur.fg = 39
			case p == 49:
				cur.bg = 49
			case 30 <= p && p <= 37:
				cur.fg = p
			case 90 <= p && p <= 97:
				cur.fg = p
			case 40 <= p && p <= 47:
				cur.bg = p
			case 100 <= p && p <= 107:
				cur.bg = p
			}
		}
	}

	clearScreen := func(mode int) {
		// mode: 2 = entire screen; 0/1 not used here
		if mode == 2 {
			for yy := 0; yy < height; yy++ {
				for xx := 0; xx < width; xx++ {
					canvas[yy][xx] = cell{' ', cur}
				}
			}
			x, y = 0, 0
		}
	}

	clearEOL := func() {
		if y >= 0 && y < height {
			for xx := x; xx < width; xx++ {
				canvas[y][xx] = cell{' ', cur}
			}
		}
	}

	// parse input
	for i := 0; i < len(s); {
		if s[i] != '\x1b' {
			r, sz := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && sz == 1 {
				// treat as raw byte
				writeRune(rune(s[i]))
				i++
				continue
			}
			writeRune(r)
			i += sz
			continue
		}
		// ESC
		j := i + 1
		if j < len(s) && s[j] == '[' { // CSI
			j++
			// collect parameter bytes until final
			start := j
			for j < len(s) && !(s[j] >= '@' && s[j] <= '~') {
				j++
			}
			if j >= len(s) {
				break
			}
			final := s[j]
			paramsStr := s[start:j]
			j++

			// parse params
			var params []int
			if len(paramsStr) > 0 {
				parts := strings.Split(paramsStr, ";")
				for _, part := range parts {
					if part == "" {
						params = append(params, 0)
						continue
					}
					// ignore '?'
					part = strings.TrimPrefix(part, "?")
					if n, err := strconv.Atoi(part); err == nil {
						params = append(params, n)
					}
				}
			}

			switch final {
			case 'm':
				setSGR(params)
			case 'H', 'f':
				// cursor position: row;col (1-based)
				rr, cc := 1, 1
				if len(params) >= 1 {
					rr = params[0]
				}
				if len(params) >= 2 {
					cc = params[1]
				}
				if rr < 1 {
					rr = 1
				}
				if cc < 1 {
					cc = 1
				}
				y = rr - 1
				x = cc - 1
			case 'J':
				mode := 0
				if len(params) >= 1 {
					mode = params[0]
				}
				clearScreen(mode)
			case 'K':
				clearEOL()
			default:
				// ignore other CSI sequences
			}
			i = j
			continue
		}
		// other ESC sequences ignored
		i = j
	}

	// Build output lines with minimal SGR sequences
	lines := make([]string, height)
	resetSGR := "\x1b[0m"
	for yy := 0; yy < height; yy++ {
		var b strings.Builder
		// start reset to avoid bleed
		b.WriteString(resetSGR)
		// current emitted style
		out := style{fg: 39, bg: 49, bold: false}

		emitSGR := func(st style) {
			var params []string
			params = append(params, "0")
			if st.bold {
				params = append(params, "1")
			}
			if st.fg != 39 {
				params = append(params, strconv.Itoa(st.fg))
			}
			if st.bg != 49 {
				params = append(params, strconv.Itoa(st.bg))
			}
			b.WriteString("\x1b[" + strings.Join(params, ";") + "m")
			out = st
		}

		for xx := 0; xx < width; xx++ {
			c := canvas[yy][xx]
			if c.st != out {
				emitSGR(c.st)
			}
			b.WriteRune(c.ch)
		}
		b.WriteString(resetSGR)
		lines[yy] = b.String()
	}
	return lines
}
