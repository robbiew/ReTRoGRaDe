package ui

import (
	"fmt"
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

// Prompt collects string input within a defined width with ESC key detection.
func Prompt(term InteractiveTerminal, label string, x, y, width int) (string, error) {
	var input strings.Builder
	labelX := x
	inputFieldX := labelX + len(label)

	// Draw the label and initialize the input field with reverse background.
	if err := term.PrintAt(label, labelX, y); err != nil {
		return "", err
	}
	if err := term.PrintAt(ReverseText+PadRight("", width)+ResetText, inputFieldX, y); err != nil {
		return "", err
	}
	if err := term.MoveCursor(inputFieldX, y); err != nil {
		return "", err
	}

	for {
		key, err := term.GetKeyPress()
		if err != nil {
			return "", err
		}

		switch key {
		case 13: // Enter key
			if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
				return "", err
			}
			if err := term.PrintAt(Ansi.Cyan+label+Ansi.Reset, labelX, y); err != nil {
				return "", err
			}
			if err := term.PrintAt(PadRight(input.String(), width), inputFieldX, y); err != nil {
				return "", err
			}
			return strings.TrimSpace(input.String()), nil

		case 27: // ESC key
			return "", fmt.Errorf("ESC_PRESSED")

		case 8, 127: // Backspace
			if input.Len() > 0 {
				inputStr := input.String()
				input.Reset()
				input.WriteString(inputStr[:len(inputStr)-1])
				if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.PrintAt(ReverseText+PadRight(input.String(), width)+ResetText, inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.MoveCursor(inputFieldX+input.Len(), y); err != nil {
					return "", err
				}
			}

		case 32: // Space
			if input.Len() < width {
				input.WriteString(" ")
				if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.PrintAt(ReverseText+PadRight(input.String(), width)+ResetText, inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.MoveCursor(inputFieldX+input.Len(), y); err != nil {
					return "", err
				}
			}

		default:
			if input.Len() < width && key >= 32 && key <= 126 {
				input.WriteByte(key)
				if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.PrintAt(ReverseText+PadRight(input.String(), width)+ResetText, inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.MoveCursor(inputFieldX+input.Len(), y); err != nil {
					return "", err
				}
			}
		}
	}
}

// PromptPassword collects password input with asterisk masking and ESC detection.
func PromptPassword(term InteractiveTerminal, label string, x, y, width int) (string, error) {
	var input strings.Builder
	labelX := x
	inputFieldX := labelX + len(label)

	if err := term.PrintAt(label, labelX, y); err != nil {
		return "", err
	}
	if err := term.PrintAt(ReverseText+PadRight("", width)+ResetText, inputFieldX, y); err != nil {
		return "", err
	}
	if err := term.MoveCursor(inputFieldX, y); err != nil {
		return "", err
	}

	for {
		key, err := term.GetKeyPress()
		if err != nil {
			return "", err
		}

		switch key {
		case 13: // Enter key
			if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
				return "", err
			}
			if err := term.PrintAt(Ansi.Cyan+label+Ansi.Reset, labelX, y); err != nil {
				return "", err
			}
			asterisks := strings.Repeat("*", input.Len())
			if err := term.PrintAt(PadRight(asterisks, width), inputFieldX, y); err != nil {
				return "", err
			}
			return strings.TrimSpace(input.String()), nil

		case 27: // ESC key
			return "", fmt.Errorf("ESC_PRESSED")

		case 8, 127: // Backspace
			if input.Len() > 0 {
				inputStr := input.String()
				input.Reset()
				input.WriteString(inputStr[:len(inputStr)-1])
				if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
					return "", err
				}
				asterisks := strings.Repeat("*", input.Len())
				if err := term.PrintAt(ReverseText+PadRight(asterisks, width)+ResetText, inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.MoveCursor(inputFieldX+input.Len(), y); err != nil {
					return "", err
				}
			}

		case 32: // Space
			if input.Len() < width {
				input.WriteString(" ")
				if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
					return "", err
				}
				asterisks := strings.Repeat("*", input.Len())
				if err := term.PrintAt(ReverseText+PadRight(asterisks, width)+ResetText, inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.MoveCursor(inputFieldX+input.Len(), y); err != nil {
					return "", err
				}
			}

		default:
			if input.Len() < width && key >= 32 && key <= 126 {
				input.WriteByte(key)
				if err := term.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y); err != nil {
					return "", err
				}
				asterisks := strings.Repeat("*", input.Len())
				if err := term.PrintAt(ReverseText+PadRight(asterisks, width)+ResetText, inputFieldX, y); err != nil {
					return "", err
				}
				if err := term.MoveCursor(inputFieldX+input.Len(), y); err != nil {
					return "", err
				}
			}
		}
	}
}

// ShowTimedError displays an error message for a short duration then clears it.
func ShowTimedError(term InteractiveTerminal, message string, x, y int) {
	if err := term.PrintAt(Ansi.RedHi+message+Ansi.Reset, x, y); err != nil {
		return
	}

	go func() {
		time.Sleep(2 * time.Second)
		clearLine := strings.Repeat(" ", len(message)+10)
		_ = term.PrintAt(clearLine, x, y)
	}()
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

// ClearField clears a form field and resets cursor position.
func ClearField(term InteractiveTerminal, label string, x, y, width int) {
	labelX := x
	inputFieldX := labelX + len(label)

	_ = term.PrintAt(" "+PadRight("", width+len(label)+2)+" ", labelX, y)
	_ = term.PrintAt(label, labelX, y)
	_ = term.PrintAt(ReverseText+PadRight("", width)+ResetText, inputFieldX, y)
	_ = term.MoveCursor(inputFieldX, y)
}

// ShowPersistentEscIndicator displays a persistent ESC quit option.
func ShowPersistentEscIndicator(term InteractiveTerminal, x, y int) {
	_ = term.PrintAt(Ansi.BlackHi+" [ESC] Quit/Cancel "+Ansi.Reset, x, y)
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

// PrintAnsiTerminal displays ANSI art content on the provided terminal.
func PrintAnsiTerminal(term InteractiveTerminal, artName string, delay, height int) error {
	lines, err := LoadAnsiLines(artName)
	if err != nil {
		return fmt.Errorf("error: ANSI art '%s' not found", artName)
	}

	delayDuration := time.Duration(delay) * time.Millisecond
	printed := 0

	for _, line := range lines {
		if height > 0 && printed >= height {
			break
		}
		if err := term.Print(line + "\r\n"); err != nil {
			return err
		}
		if delay > 0 {
			time.Sleep(delayDuration)
		}
		printed++
	}
	return nil
}

func toUpperASCII(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}
