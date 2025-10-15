package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// InteractiveTerminal describes the primitives required by the high-level UI helpers.
type InteractiveTerminal interface {
	Print(text string) error
	PrintAt(text string, x, y int) error
	MoveCursor(x, y int) error
	GetKeyPress() (byte, error)
}

// PromptSimple collects string input without specific positioning
// labelColor: color for the label text
// fgColor: color for the input text
// bgColor: background color for the input field
func PromptSimple(term InteractiveTerminal, label string, width int, labelColor, fgColor, bgColor string) (string, error) {
	var input strings.Builder

	// Set defaults if not provided
	if labelColor == "" {
		labelColor = Ansi.Cyan
	}
	if fgColor == "" {
		fgColor = Ansi.Reset
	}
	if bgColor == "" {
		bgColor = Ansi.BgBlue
	}

	// Print the label with color
	if err := term.Print(labelColor + label + Ansi.Reset); err != nil {
		return "", err
	}

	// Start with background and empty field
	if err := term.Print(bgColor + strings.Repeat(" ", width) + Ansi.Reset); err != nil {
		return "", err
	}
	// Move cursor back to start of input field
	if err := term.Print(strings.Repeat("\x08", width)); err != nil {
		return "", err
	}

	for {
		key, err := term.GetKeyPress()
		if err != nil {
			return "", err
		}

		switch key {
		case 13: // Enter key
			// Clear the background, show final input
			if err := term.Print(Ansi.Reset + "\r\n"); err != nil {
				return "", err
			}
			return strings.TrimSpace(input.String()), nil

		case 27: // ESC key
			if err := term.Print(Ansi.Reset); err != nil {
				return "", err
			}
			return "", fmt.Errorf("ESC_PRESSED")

		case 8, 127: // Backspace
			if input.Len() > 0 {
				inputStr := input.String()
				input.Reset()
				input.WriteString(inputStr[:len(inputStr)-1])

				// Move back, print space with bg color, move back again
				if err := term.Print("\x08" + bgColor + " " + Ansi.Reset + "\x08"); err != nil {
					return "", err
				}
			}

		case 32: // Space
			if input.Len() < width {
				input.WriteString(" ")
				// Print space with foreground and background colors
				if err := term.Print(fgColor + bgColor + " " + Ansi.Reset); err != nil {
					return "", err
				}
			}

		default:
			if input.Len() < width && key >= 32 && key <= 126 {
				input.WriteByte(key)
				// Print character with foreground and background colors
				if err := term.Print(fgColor + bgColor + string(key) + Ansi.Reset); err != nil {
					return "", err
				}
			}
		}
	}
}

// PromptPasswordSimple collects password input with asterisk masking and configurable colors
// labelColor: color for the label text
// fgColor: color for the asterisks
// bgColor: background color for the input field
func PromptPasswordSimple(term InteractiveTerminal, label string, width int, labelColor, fgColor, bgColor string) (string, error) {
	var input strings.Builder

	// Set defaults if not provided
	if labelColor == "" {
		labelColor = Ansi.Cyan
	}
	if fgColor == "" {
		fgColor = Ansi.Reset
	}
	if bgColor == "" {
		bgColor = Ansi.BgBlue
	}

	// Print the label with color
	if err := term.Print(labelColor + label + Ansi.Reset); err != nil {
		return "", err
	}

	// Start with background and empty field
	if err := term.Print(bgColor + strings.Repeat(" ", width) + Ansi.Reset); err != nil {
		return "", err
	}
	// Move cursor back to start of input field
	if err := term.Print(strings.Repeat("\x08", width)); err != nil {
		return "", err
	}

	for {
		key, err := term.GetKeyPress()
		if err != nil {
			return "", err
		}

		switch key {
		case 13: // Enter key
			if err := term.Print(Ansi.Reset + "\r\n"); err != nil {
				return "", err
			}
			return strings.TrimSpace(input.String()), nil

		case 27: // ESC key
			if err := term.Print(Ansi.Reset); err != nil {
				return "", err
			}
			return "", fmt.Errorf("ESC_PRESSED")

		case 8, 127: // Backspace
			if input.Len() > 0 {
				inputStr := input.String()
				input.Reset()
				input.WriteString(inputStr[:len(inputStr)-1])

				// Move back, print space with bg color, move back again
				if err := term.Print("\x08" + bgColor + " " + Ansi.Reset + "\x08"); err != nil {
					return "", err
				}
			}

		case 32: // Space
			if input.Len() < width {
				input.WriteString(" ")
				// Print asterisk with foreground and background colors
				if err := term.Print(fgColor + bgColor + "*" + Ansi.Reset); err != nil {
					return "", err
				}
			}

		default:
			if input.Len() < width && key >= 32 && key <= 126 {
				input.WriteByte(key)
				// Print asterisk with foreground and background colors
				if err := term.Print(fgColor + bgColor + "*" + Ansi.Reset); err != nil {
					return "", err
				}
			}
		}
	}
}

// ClearLine clears the current line
func ClearLine(term InteractiveTerminal, length int) error {
	// Move to start of line, print spaces, move back to start
	clearStr := "\r" + strings.Repeat(" ", length) + "\r"
	return term.Print(clearStr)
}

// ClearFieldSimple clears the input area and reprints the label
func ClearFieldSimple(term InteractiveTerminal, label string, width int) error {
	// Calculate total length (label + input area)
	totalLength := len(label) + width

	// Clear the line
	if err := ClearLine(term, totalLength); err != nil {
		return err
	}

	// Reprint the label
	return term.Print(label)
}

// ShowTimedErrorSimple displays an error message for a short duration (simpler version)
func ShowTimedErrorSimple(term InteractiveTerminal, message string) {
	if err := term.Print(Ansi.RedHi + message + Ansi.Reset + "\r\n"); err != nil {
		return
	}
	time.Sleep(2 * time.Second)
}

// FlushInput attempts to clear any buffered input
func FlushInput(term InteractiveTerminal) {
	// This is a marker interface method that should be implemented by TelnetIO
	// For now, we'll just add a small delay to let any stray bytes settle
	time.Sleep(50 * time.Millisecond)
}

// HandleEscQuit shows quit confirmation and returns true if the user confirms.
func HandleEscQuit(term InteractiveTerminal) bool {
	if err := term.Print(Ansi.YellowHi + "\r\n\r\n Do you really want to quit? [Y/N]: " + Ansi.Reset); err != nil {
		return true
	}

	for {
		key, err := term.GetKeyPress()
		if err != nil {
			return true
		}

		switch toUpperASCII(key) {
		case 'Y':
			return true
		case 'N':
			_ = term.PrintAt(strings.Repeat(" ", 40), 1, 0)
			return false
		}
	}
}

// Pause waits for any key press and shows centered message.
func Pause(term InteractiveTerminal) error {
	const message = " [ PRESS A KEY TO CONTINUE ] "
	width := 80
	padWidth := (width - len(message)) / 2

	if err := term.Print("\r\n" + strings.Repeat(" ", padWidth) + Ansi.RedHi + message + Ansi.Reset + "\r\n"); err != nil {
		return err
	}
	_, err := term.GetKeyPress()
	return err
}

func toUpperASCII(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

// ClearScreenSequence returns the escape sequence to clear the screen and move the cursor to the top-left.
func ClearScreenSequence() string {
	return Ansi.EraseScreen + Ansi.CursorTopLeft
}

// MoveCursorSequence returns the escape sequence to move the cursor to the specified coordinates.
func MoveCursorSequence(x, y int) string {
	return fmt.Sprintf(Esc+"%d;%df", y, x)
}

// GetTermSize retrieves the terminal's current height and width - for telnet we'll use defaults.
func GetTermSize() (int, int) {
	// For telnet connections, we'll use standard terminal size
	// This can be enhanced later to negotiate terminal size via telnet options
	return 24, 80 // height, width
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
