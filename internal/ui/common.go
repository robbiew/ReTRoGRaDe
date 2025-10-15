package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/robbiew/retrograde/internal/util"
)

// Struct for organizing all ANSI escape sequences
type AnsiEscapes struct {
	EraseScreen        string
	CursorHide         string
	CursorShow         string
	CursorBackward     string
	CursorPrevLine     string
	CursorLeft         string
	CursorTop          string
	CursorTopLeft      string
	CursorBlinkEnable  string
	CursorBlinkDisable string
	ScrollUp           string
	ScrollDown         string
	TextInsertChar     string
	TextDeleteChar     string
	TextEraseChar      string
	TextInsertLine     string
	TextDeleteLine     string
	EraseRight         string
	EraseLeft          string
	EraseLine          string
	EraseDown          string
	EraseUp            string
	Black              string
	Red                string
	Green              string
	Yellow             string
	Blue               string
	Magenta            string
	Cyan               string
	White              string
	BlackHi            string
	RedHi              string
	GreenHi            string
	YellowHi           string
	BlueHi             string
	MagentaHi          string
	CyanHi             string
	WhiteHi            string
	BgBlack            string
	BgRed              string
	BgGreen            string
	BgYellow           string
	BgBlue             string
	BgMagenta          string
	BgCyan             string
	BgWhite            string
	BgBlackHi          string
	BgRedHi            string
	BgGreenHi          string
	BgYellowHi         string
	BgBlueHi           string
	BgMagentaHi        string
	BgCyanHi           string
	BgWhiteHi          string
	Reset              string
}

var Ansi = AnsiEscapes{
	EraseScreen:        Esc + "2J",
	CursorHide:         Esc + "?25l",
	CursorShow:         Esc + "?25h",
	CursorBackward:     Esc + "D",
	CursorPrevLine:     Esc + "F",
	CursorLeft:         Esc + "G",
	CursorTop:          Esc + "d",
	CursorTopLeft:      Esc + "H",
	CursorBlinkEnable:  Esc + "?12h",
	CursorBlinkDisable: Esc + "?12l",
	ScrollUp:           Esc + "S",
	ScrollDown:         Esc + "T",
	TextInsertChar:     Esc + "@",
	TextDeleteChar:     Esc + "P",
	TextEraseChar:      Esc + "X",
	TextInsertLine:     Esc + "L",
	TextDeleteLine:     Esc + "M",
	EraseRight:         Esc + "K",
	EraseLeft:          Esc + "1K",
	EraseLine:          Esc + "2K",
	EraseDown:          Esc + "J",
	EraseUp:            Esc + "1J",
	Black:              Esc + "30m",
	Red:                Esc + "31m",
	Green:              Esc + "32m",
	Yellow:             Esc + "33m",
	Blue:               Esc + "34m",
	Magenta:            Esc + "35m",
	Cyan:               Esc + "36m",
	White:              Esc + "37m",
	BlackHi:            Esc + "30;1m",
	RedHi:              Esc + "31;1m",
	GreenHi:            Esc + "32;1m",
	YellowHi:           Esc + "33;1m",
	BlueHi:             Esc + "34;1m",
	MagentaHi:          Esc + "35;1m",
	CyanHi:             Esc + "36;1m",
	WhiteHi:            Esc + "37;1m",
	BgBlack:            Esc + "40m",
	BgRed:              Esc + "41m",
	BgGreen:            Esc + "42m",
	BgYellow:           Esc + "43m",
	BgBlue:             Esc + "44m",
	BgMagenta:          Esc + "45m",
	BgCyan:             Esc + "46m",
	BgWhite:            Esc + "47m",
	BgBlackHi:          Esc + "40;1m",
	BgRedHi:            Esc + "41;1m",
	BgGreenHi:          Esc + "42;1m",
	BgYellowHi:         Esc + "43;1m",
	BgBlueHi:           Esc + "44;1m",
	BgMagentaHi:        Esc + "45;1m",
	BgCyanHi:           Esc + "46;1m",
	BgWhiteHi:          Esc + "47;1m",
	Reset:              Esc + "0m",
}

const (
	Esc = "\u001B[" // ANSI escape sequence prefix
	Osc = "\u001B]" // Operating System Command prefix
	Bel = "\u0007"  // Bell character

	ReverseText = "\033[44;37m" // Blue background, white text
	ResetText   = "\033[0m"     // Reset to normal
)

var (
	artMu    sync.RWMutex
	artDirs  []string
	artCache = make(map[string]string)
)

// SetArtDirectories configures the lookup order for ANSI art files. Empty values are ignored.
func SetArtDirectories(dirs ...string) {
	artMu.Lock()
	defer artMu.Unlock()

	seen := make(map[string]struct{})
	artDirs = artDirs[:0]

	appendDir := func(dir string) {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return
		}
		clean := filepath.Clean(dir)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		artDirs = append(artDirs, clean)
	}

	for _, dir := range dirs {
		appendDir(dir)
	}

	artCache = make(map[string]string)
}

func artDirectories() []string {
	artMu.RLock()
	defer artMu.RUnlock()
	return append([]string(nil), artDirs...)
}

func cacheAnsiArt(name, content string) {
	artMu.Lock()
	defer artMu.Unlock()
	artCache[name] = content
}

func cachedAnsiArt(name string) (string, bool) {
	artMu.RLock()
	defer artMu.RUnlock()
	content, ok := artCache[name]
	return content, ok
}

func artCandidates(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	if filepath.Ext(name) != "" {
		return []string{name}
	}

	return []string{name, name + ".ans", name + ".asc"}
}

// LoadAnsiArt reads the ANSI art file content for the provided name, searching configured directories.
func LoadAnsiArt(artName string) (string, error) {
	if artName == "" {
		return "", fmt.Errorf("ANSI art name cannot be empty")
	}

	if content, ok := cachedAnsiArt(artName); ok {
		return content, nil
	}

	candidates := artCandidates(artName)
	if len(candidates) == 0 {
		return "", fmt.Errorf("ANSI art %q not found", artName)
	}

	for _, dir := range artDirectories() {
		for _, candidate := range candidates {
			path := candidate
			if dir != "" {
				path = filepath.Join(dir, candidate)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			content := string(data)
			cacheAnsiArt(artName, content)
			return content, nil
		}
	}

	return "", fmt.Errorf("ANSI art %q not found", artName)
}

// LoadAnsiLines returns ANSI art content split into lines (without trailing carriage returns).
// func LoadAnsiLines(artName string) ([]string, error) {
// 	content, err := LoadAnsiArt(artName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	trimmedContent := TrimStringFromSauce(content)
// 	rawLines := strings.Split(trimmedContent, "\n")
// 	lines := make([]string, 0, len(rawLines))
// 	for _, line := range rawLines {
// 		lines = append(lines, strings.TrimRight(line, "\r"))
// 	}
// 	return lines, nil
// }

// ClearScreen clears the terminal screen.
func ClearScreen() {
	fmt.Print(ClearScreenSequence())
}

// ClearScreenSequence returns the escape sequence to clear the screen and move the cursor to the top-left.
func ClearScreenSequence() string {
	return Ansi.EraseScreen + Ansi.CursorTopLeft
}

// MoveCursorSequence returns the escape sequence to move the cursor to the specified coordinates.
func MoveCursorSequence(x, y int) string {
	return fmt.Sprintf(Esc+"%d;%df", y, x)
}

// StripANSI removes ANSI escape codes to calculate the visible length of a string.
func StripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// GetTermSize retrieves the terminal's current height and width - for telnet we'll use defaults.
func GetTermSize() (int, int) {
	// For telnet connections, we'll use standard terminal size
	// This can be enhanced later to negotiate terminal size via telnet options
	return 24, 80 // height, width
}

// DrawInputField draws a labeled input field at a fixed width with no blue background.
func DrawInputField(label string, x, y, width int) {
	PrintStringLoc(label+":", x, y)
	PrintStringLoc(PadRight(" ", width), x+len(label)+2, y) // Non-highlighted background for inactive fields
}

// PadRight pads a string with spaces to a specified width.
func PadRight(str string, width int) string {
	for len(str) < width {
		str += " "
	}
	return str
}

// PrintStringLoc prints text at a specific X, Y location.
func PrintStringLoc(text string, x int, y int) {
	WriteAt(os.Stdout, text, x, y) // ignore error for stdout printing
}

// WriteAt writes text to the provided writer after moving the cursor to X, Y.
func WriteAt(w io.Writer, text string, x int, y int) (int, error) {
	if w == nil {
		return 0, fmt.Errorf("nil writer")
	}
	return fmt.Fprint(w, MoveCursorSequence(x, y)+text)
}

// PrintStyledText prints styled text at a specific location using the provided font escape code.
func PrintStyledText(text, font string, x, y int) {
	styledText := font + text + Ansi.Reset
	PrintStringLoc(styledText, x, y)
}

// PrintAnsi displays ANSI art by name, removes SAUCE metadata, and prints it line by line with an optional delay and height limit.
func PrintAnsi(artName string, delay int, height int) {
	lines, err := LoadAnsiLines(artName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	delayDuration := time.Duration(delay) * time.Millisecond
	printed := 0
	for _, line := range lines {
		if height > 0 && printed >= height {
			break
		}
		fmt.Print(line + "\r\n")
		if delay > 0 {
			time.Sleep(delayDuration)
		}
		printed++
	}
}

// TrimStringFromSauce trims SAUCE metadata from a string (if necessary).
func TrimStringFromSauce(s string) string {
	return trimMetadata(s, "SAUCE00", "COMNT")
}

// Helper to trim metadata based on delimiters.
func trimMetadata(s string, delimiters ...string) string {
	for _, delimiter := range delimiters {
		if idx := strings.Index(s, delimiter); idx != -1 {
			return trimLastChar(s[:idx])
		}
	}
	return s
}

// trimLastChar trims the last character from a string.
// func trimLastChar(s string) string {
// 	if len(s) > 0 {
// 		_, size := utf8.DecodeLastRuneInString(s)
// 		return s[:len(s)-size]
// 	}
// 	return s
// }

// MoveCursor moves the cursor to X, Y location.
func MoveCursor(x int, y int) {
	fmt.Print(MoveCursorSequence(x, y))
}

// SanitizeFilename removes or replaces unsafe characters from filenames.
func SanitizeFilename(filename string) string {
	return util.SanitizeFilename(filename)
}

var ansiColorTable = []string{
	Ansi.Black,
	Ansi.Red,
	Ansi.Green,
	Ansi.Yellow,
	Ansi.Blue,
	Ansi.Magenta,
	Ansi.Cyan,
	Ansi.White,
	Ansi.BlackHi,
	Ansi.RedHi,
	Ansi.GreenHi,
	Ansi.YellowHi,
	Ansi.BlueHi,
	Ansi.MagentaHi,
	Ansi.CyanHi,
	Ansi.WhiteHi,
}

// ColorFromNumber maps a 0-15 color code to the corresponding ANSI escape sequence.
func ColorFromNumber(code int) string {
	if code >= 0 && code < len(ansiColorTable) {
		return ansiColorTable[code]
	}
	return Ansi.White
}
