package telnet

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/robbiew/retrograde/internal/config"
)

const (
	ReverseText = "\033[44;37m" // Blue background, white text
	ResetText   = "\033[0m"     // Reset to normal
)

// ANSI color codes
var Ansi = struct {
	Cyan     string
	Reset    string
	RedHi    string
	YellowHi string
	GreenHi  string
	BlackHi  string
	Green    string
	BgCyan   string
	WhiteHi  string
}{
	Cyan:     "\033[36m",
	Reset:    "\033[0m",
	RedHi:    "\033[31;1m",
	YellowHi: "\033[33;1m",
	GreenHi:  "\033[32;1m",
	BlackHi:  "\033[30;1m",
	Green:    "\033[32m",
	BgCyan:   "\033[46m",
	WhiteHi:  "\033[37;1m",
}

// PadRight pads a string with spaces to a specified width
func PadRight(str string, width int) string {
	for len(str) < width {
		str += " "
	}
	return str
}

// TelnetIO handles input/output for telnet connections
type TelnetIO struct {
	Reader  *bufio.Reader
	Writer  *bufio.Writer
	Session *config.TelnetSession // Reference to session for activity tracking
}

// GetKeyPress reads a single key press from telnet connection
func (t *TelnetIO) GetKeyPress() (byte, error) {
	b, err := t.Reader.ReadByte()
	if err != nil {
		return 0, err
	}

	// Handle telnet command sequences (IAC = 255)
	if b == 255 {
		t.handleTelnetCommand()
		// Recursively call to get the actual key press
		return t.GetKeyPress()
	}

	// Update activity time if session is available
	if t.Session != nil {
		t.Session.LastActivity = time.Now()
	}

	return b, nil
}

// GetKeyPressUpper reads a key and converts to uppercase
func (t *TelnetIO) GetKeyPressUpper() (byte, error) {
	key, err := t.GetKeyPress()
	if err != nil {
		return 0, err
	}

	// Convert to uppercase
	if key >= 'a' && key <= 'z' {
		key = key - 32
	}

	return key, nil
}

// Print sends text to the telnet client
func (t *TelnetIO) Print(text string) error {
	_, err := t.Writer.WriteString(text)
	if err != nil {
		return err
	}
	return t.Writer.Flush()
}

// Printf sends formatted text to the telnet client
func (t *TelnetIO) Printf(format string, args ...interface{}) error {
	text := fmt.Sprintf(format, args...)
	return t.Print(text)
}

// PrintAt sends text at a specific cursor position
func (t *TelnetIO) PrintAt(text string, x, y int) error {
	return t.Printf("\033[%d;%dH%s", y, x, text)
}

// ClearScreen clears the telnet client screen
func (t *TelnetIO) ClearScreen() error {
	return t.Print("\033[2J\033[H")
}

// ShowCursor shows the cursor
func (t *TelnetIO) ShowCursor() error {
	return t.Print("\033[?25h")
}

// HideCursor hides the cursor
func (t *TelnetIO) HideCursor() error {
	return t.Print("\033[?25l")
}

// MoveCursor moves cursor to specific position
func (t *TelnetIO) MoveCursor(x, y int) error {
	return t.Printf("\033[%d;%dH", y, x)
}

// Prompt collects string input within a defined width for telnet with ESC key detection
func (t *TelnetIO) Prompt(label string, x, y, width int) (string, error) {
	var input strings.Builder
	labelX := x
	inputFieldX := labelX + len(label)

	// Draw the label and initialize the input field with reverse background
	t.PrintAt(label, labelX, y)
	t.PrintAt(ReverseText+PadRight("", width)+ResetText, inputFieldX, y)
	t.MoveCursor(inputFieldX, y)

	for {
		key, err := t.GetKeyPress()
		if err != nil {
			return "", err
		}

		switch key {
		case 13: // Enter key
			// Clear input area and redraw content with standard background
			t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
			t.PrintAt(Ansi.Cyan+label+Ansi.Reset, labelX, y)
			t.PrintAt(PadRight(input.String(), width), inputFieldX, y)
			return strings.TrimSpace(input.String()), nil

		case 27: // ESC key
			return "", fmt.Errorf("ESC_PRESSED")

		case 8, 127: // Backspace
			if input.Len() > 0 {
				// Remove last character and update display
				inputStr := input.String()
				input.Reset()
				input.WriteString(inputStr[:len(inputStr)-1])
				// Clear input area and redraw with updated content
				t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
				t.PrintAt(ReverseText+PadRight(input.String(), width)+ResetText, inputFieldX, y)
				t.MoveCursor(inputFieldX+input.Len(), y)
			}

		case 32: // Space
			if input.Len() < width {
				input.WriteString(" ")
				t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
				t.PrintAt(ReverseText+PadRight(input.String(), width)+ResetText, inputFieldX, y)
				t.MoveCursor(inputFieldX+input.Len(), y)
			}

		default:
			// Add character if within width limit and is printable
			if input.Len() < width && key >= 32 && key <= 126 {
				input.WriteByte(key)
				t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
				t.PrintAt(ReverseText+PadRight(input.String(), width)+ResetText, inputFieldX, y)
				t.MoveCursor(inputFieldX+input.Len(), y)
			}
		}
	}
}

// PromptPassword collects password input with asterisk masking for telnet with ESC key detection
func (t *TelnetIO) PromptPassword(label string, x, y, width int) (string, error) {
	var input strings.Builder
	labelX := x
	inputFieldX := labelX + len(label)

	// Draw the label and initialize the input field with reverse background
	t.PrintAt(label, labelX, y)
	t.PrintAt(ReverseText+PadRight("", width)+ResetText, inputFieldX, y)
	t.MoveCursor(inputFieldX, y)

	for {
		key, err := t.GetKeyPress()
		if err != nil {
			return "", err
		}

		switch key {
		case 13: // Enter key
			// Clear input area and redraw content with standard background
			t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
			t.PrintAt(Ansi.Cyan+label+Ansi.Reset, labelX, y)
			// Show asterisks for the final display
			asterisks := strings.Repeat("*", input.Len())
			t.PrintAt(PadRight(asterisks, width), inputFieldX, y)
			return strings.TrimSpace(input.String()), nil

		case 27: // ESC key
			return "", fmt.Errorf("ESC_PRESSED")

		case 8, 127: // Backspace
			if input.Len() > 0 {
				// Remove last character and update display
				inputStr := input.String()
				input.Reset()
				input.WriteString(inputStr[:len(inputStr)-1])
				// Clear input area and redraw with asterisks
				t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
				asterisks := strings.Repeat("*", input.Len())
				t.PrintAt(ReverseText+PadRight(asterisks, width)+ResetText, inputFieldX, y)
				t.MoveCursor(inputFieldX+input.Len(), y)
			}

		case 32: // Space
			if input.Len() < width {
				input.WriteString(" ")
				t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
				asterisks := strings.Repeat("*", input.Len())
				t.PrintAt(ReverseText+PadRight(asterisks, width)+ResetText, inputFieldX, y)
				t.MoveCursor(inputFieldX+input.Len(), y)
			}

		default:
			// Add character if within width limit and is printable
			if input.Len() < width && key >= 32 && key <= 126 {
				input.WriteByte(key)
				t.PrintAt(" "+PadRight("", width)+" ", inputFieldX, y)
				asterisks := strings.Repeat("*", input.Len())
				t.PrintAt(ReverseText+PadRight(asterisks, width)+ResetText, inputFieldX, y)
				t.MoveCursor(inputFieldX+input.Len(), y)
			}
		}
	}
}

// ShowTimedError displays an error message for 5 seconds then clears it
func (t *TelnetIO) ShowTimedError(message string, x, y int) {
	// Display error message
	t.PrintAt(Ansi.RedHi+message+Ansi.Reset, x, y)

	// Wait 5 seconds then clear the line
	go func() {
		time.Sleep(2 * time.Second)
		// Clear the error message line by overwriting with spaces
		clearLine := strings.Repeat(" ", len(message)+10) // Extra spaces to ensure full clear
		t.PrintAt(clearLine, x, y)
	}()
}

// HandleEscQuit shows quit confirmation and returns true if user wants to quit
func (t *TelnetIO) HandleEscQuit() bool {
	t.Print(Ansi.YellowHi + "\r\n\r\n Do you really want to quit? [Y/N]: " + Ansi.Reset)

	for {
		key, err := t.GetKeyPressUpper()
		if err != nil {
			return true // Default to quit on error
		}

		switch key {
		case 'Y':
			return true
		case 'N':
			// Clear the quit confirmation message
			t.PrintAt(strings.Repeat(" ", 40), 1, 0) // Clear confirmation line
			return false
		}
	}
}

// ClearField clears a form field and resets cursor position
func (t *TelnetIO) ClearField(label string, x, y, width int) {
	labelX := x
	inputFieldX := labelX + len(label)

	// Clear and redraw the field
	t.PrintAt(" "+PadRight("", width+len(label)+2)+" ", labelX, y)
	t.PrintAt(label, labelX, y)
	t.PrintAt(ReverseText+PadRight("", width)+ResetText, inputFieldX, y)
	t.MoveCursor(inputFieldX, y)
}

// ShowPersistentEscIndicator displays persistent ESC quit option
func (t *TelnetIO) ShowPersistentEscIndicator(x, y int) {
	t.PrintAt(Ansi.BlackHi+" [ESC] Quit/Cancel "+Ansi.Reset, x, y)
}

// Pause waits for any key press and shows centered message
func (t *TelnetIO) Pause() error {
	const message = " [ PRESS A KEY TO CONTINUE ] "
	width := 80
	padWidth := (width - len(message)) / 2

	t.Print("\r\n" + strings.Repeat(" ", padWidth) + Ansi.RedHi + message + Ansi.Reset + "\r\n")
	_, err := t.GetKeyPress()
	return err
}

// PrintAnsi displays embedded ANSI art content to telnet client
func (t *TelnetIO) PrintAnsi(artName string, delay int, height int) error {
	// Get the embedded art content
	content := GetArtContent(artName)
	if content == "" {
		return t.Printf("Error: ANSI art '%s' not found\r\n", artName)
	}

	// Remove SAUCE metadata
	trimmedContent := TrimStringFromSauce(content)

	lines := strings.Split(trimmedContent, "\n")
	lineCount := 0

	// Print each line with delay and stop after reaching height
	for _, line := range lines {
		if lineCount >= height && height > 0 {
			break
		}
		t.Print(line + "\r\n")
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		lineCount++
	}

	return nil
}

// Handle telnet command sequences
func (t *TelnetIO) handleTelnetCommand() {
	// Read the telnet command
	cmd, err := t.Reader.ReadByte()
	if err != nil {
		return
	}

	// Handle different telnet commands
	switch cmd {
	case 251, 252, 253, 254: // WILL, WONT, DO, DONT
		// Read the option byte
		t.Reader.ReadByte()
		// For now, just consume the option - could respond appropriately later
	}
}

// ANSI art files - these should be loaded from disk
// For now, return empty string as placeholder
func GetArtContent(artName string) string {
	// This is a placeholder - in practice, this would load from artfiles directory
	// The actual loading happens in cmd/server/main.go
	return ""
}

// TrimStringFromSauce removes SAUCE metadata from a string
func TrimStringFromSauce(s string) string {
	return trimMetadata(s, "COMNT", "SAUCE00")
}

// trimMetadata trims metadata based on delimiters
func trimMetadata(s string, delimiters ...string) string {
	for _, delimiter := range delimiters {
		if idx := strings.Index(s, delimiter); idx != -1 {
			return s[:idx]
		}
	}
	return s
}

// Helper function to read embedded file (abstracted for easier testing)
var ReadFile = func(filename string) ([]byte, error) {
	// Extract art name from filename path
	artName := strings.TrimPrefix(filename, "artfiles/")
	artName = strings.TrimSuffix(artName, ".ans")
	content := GetArtContent(artName)
	if content == "" {
		return nil, fmt.Errorf("art file not found: %s", filename)
	}
	return []byte(content), nil
}
