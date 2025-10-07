package ui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// ANSI art files are read from disk at runtime
// Paths are relative to the server executable location
var welcomeArt string
var adminArt string
var ghostnetArt string
var mainArt string
var statusArt string
var wwivnetArt string

func init() {
	// Load art files at startup
	if data, err := os.ReadFile("artfiles/welcome.ans"); err == nil {
		welcomeArt = string(data)
	}
	if data, err := os.ReadFile("artfiles/admin.ans"); err == nil {
		adminArt = string(data)
	}
	if data, err := os.ReadFile("artfiles/ghostnet.ans"); err == nil {
		ghostnetArt = string(data)
	}
	if data, err := os.ReadFile("artfiles/main.ans"); err == nil {
		mainArt = string(data)
	}
	if data, err := os.ReadFile("artfiles/status.ans"); err == nil {
		statusArt = string(data)
	}
	if data, err := os.ReadFile("artfiles/wwivnet.ans"); err == nil {
		wwivnetArt = string(data)
	}
}

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
	// Custom handle colors in README
	H1   string
	H2   string
	H3   string
	H4   string
	Bold string
	// Add more as needed
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
	H1:                 Esc + "32;1m", // Bright Green
	H2:                 Esc + "34;1m", // Bright Blue
	H3:                 Esc + "36;1m", // Bright Cyan
	H4:                 Esc + "35;1m", // Bright Magenta
	Bold:               Esc + "1m",
}

const (
	Esc = "\u001B[" // ANSI escape sequence prefix
	Osc = "\u001B]" // Operating System Command prefix
	Bel = "\u0007"  // Bell character

	ReverseText = "\033[44;37m" // Blue background, white text
	ResetText   = "\033[0m"     // Reset to normal
)

// ClearScreen clears the terminal screen
func ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

// StripANSI removes ANSI escape codes to calculate the visible length of a string
func StripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// GetTermSize retrieves the terminal's current height and width - for telnet we'll use defaults
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

// PadRight pads a string with spaces to a specified width
func PadRight(str string, width int) string {
	for len(str) < width {
		str += " "
	}
	return str
}

// Print text at a specific X, Y location (optional)
func PrintStringLoc(text string, x int, y int) {
	fmt.Fprintf(os.Stdout, "\033[%d;%df%s", y, x, text)
}

func PrintStyledText(text, font string, x, y int) {
	styledText := font + text + Ansi.Reset
	PrintStringLoc(styledText, x, y)
}

// PrintAnsi displays embedded ANSI art by name, removes SAUCE metadata, and prints it line by line with an optional delay and height limit
func PrintAnsi(artName string, delay int, height int) {
	// Get the embedded art content
	content := GetArtContent(artName)
	if content == "" {
		fmt.Printf("Error: ANSI art '%s' not found\n", artName)
		return
	}

	// Remove SAUCE metadata
	trimmedContent := TrimStringFromSauce(content)

	// Create a scanner to read the trimmed content line by line
	scanner := bufio.NewScanner(strings.NewReader(trimmedContent))
	lineCount := 0

	// Print each line with a delay and stop after reaching the specified height
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Print(line + "\r\n") // Print the current line
		time.Sleep(time.Duration(delay) * time.Millisecond)

		lineCount++
		if lineCount >= height {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading ANSI art content: %v\n", err)
	}
}

// GetArtContent returns the embedded art content by name
func GetArtContent(artName string) string {
	switch artName {
	case "welcome":
		return welcomeArt
	case "admin":
		return adminArt
	case "ghostnet":
		return ghostnetArt
	case "main":
		return mainArt
	case "status":
		return statusArt
	case "wwivnet":
		return wwivnetArt
	default:
		return ""
	}
}

// Trims SAUCE metadata from a string (if necessary)
func TrimStringFromSauce(s string) string {
	return trimMetadata(s, "COMNT", "SAUCE00")
}

// Helper to trim metadata based on delimiters
func trimMetadata(s string, delimiters ...string) string {
	for _, delimiter := range delimiters {
		if idx := strings.Index(s, delimiter); idx != -1 {
			return trimLastChar(s[:idx])
		}
	}
	return s
}

// trimLastChar trims the last character from a string
func trimLastChar(s string) string {
	if len(s) > 0 {
		_, size := utf8.DecodeLastRuneInString(s)
		return s[:len(s)-size]
	}
	return s
}

// Move cursor to X, Y location
func MoveCursor(x int, y int) {
	fmt.Printf(Esc+"%d;%df", y, x)
}

// SanitizeFilename removes or replaces unsafe characters from filenames
func SanitizeFilename(filename string) string {
	// Replace unsafe characters with underscores
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := filename
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
