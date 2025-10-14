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
		return fmt.Errorf("Error: ANSI art '%s' not found", artName)
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
