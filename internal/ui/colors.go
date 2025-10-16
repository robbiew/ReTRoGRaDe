package ui

import (
	"regexp"
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

var ansiColorTable = []string{
	Ansi.Black,     // 0
	Ansi.Red,       // 1
	Ansi.Green,     // 2
	Ansi.Yellow,    // 3
	Ansi.Blue,      // 4
	Ansi.Magenta,   // 5
	Ansi.Cyan,      // 6
	Ansi.White,     // 7
	Ansi.BlackHi,   // 8
	Ansi.RedHi,     // 9
	Ansi.GreenHi,   // 10
	Ansi.YellowHi,  // 11
	Ansi.BlueHi,    // 12
	Ansi.MagentaHi, // 13
	Ansi.CyanHi,    // 14
	Ansi.WhiteHi,   // 15
}

const (
	Esc = "\u001B[" // ANSI escape sequence prefix
	Osc = "\u001B]" // Operating System Command prefix
	Bel = "\u0007"  // Bell character

	ReverseText = "\033[44;37m" // Blue background, white text
	ResetText   = "\033[0m"     // Reset to normal
)

// StripANSI removes ANSI escape codes to calculate the visible length of a string.
func StripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// ColorFromNumber maps a 0-15 color code to the corresponding ANSI escape sequence.
func ColorFromNumber(code int) string {
	if code >= 0 && code < len(ansiColorTable) {
		return ansiColorTable[code]
	}
	return Ansi.White
}
