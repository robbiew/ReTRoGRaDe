package telnet

import (
	"bufio"
	"fmt"
	"time"

	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/ui"
)

// TelnetIO handles input/output for telnet connections
type TelnetIO struct {
	Reader  *bufio.Reader
	Writer  *bufio.Writer
	Session *config.TelnetSession // Reference to session for activity tracking
}

// FlushInput clears any buffered input from the reader
func (t *TelnetIO) FlushInput() {
	// Read and discard any available bytes
	if t.Reader != nil {
		// Set a very short read deadline to avoid blocking
		if conn, ok := t.Session.Conn.(interface{ SetReadDeadline(time.Time) error }); ok {
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			// Read and discard bytes
			buf := make([]byte, 1024)
			for {
				_, err := t.Reader.Read(buf)
				if err != nil {
					break
				}
			}

			// Reset deadline to no timeout
			conn.SetReadDeadline(time.Time{})
		}
	}
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
	if _, err := ui.WriteAt(t.Writer, text, x, y); err != nil {
		return err
	}
	return t.Writer.Flush()
}

// ClearScreen clears the telnet client screen
func (t *TelnetIO) ClearScreen() error {
	return t.Print(ui.ClearScreenSequence())
}

// ShowCursor shows the cursor
func (t *TelnetIO) ShowCursor() error {
	return t.Print(ui.Ansi.CursorShow)
}

// HideCursor hides the cursor
func (t *TelnetIO) HideCursor() error {
	return t.Print(ui.Ansi.CursorHide)
}

// MoveCursor moves cursor to specific position
func (t *TelnetIO) MoveCursor(x, y int) error {
	return t.Printf("%s", ui.MoveCursorSequence(x, y))
}

// Prompt collects string input within a defined width for telnet with ESC key detection
func (t *TelnetIO) Prompt(label string, x, y, width int) (string, error) {
	return ui.Prompt(t, label, x, y, width)
}

// PromptPassword collects password input with asterisk masking for telnet with ESC key detection
func (t *TelnetIO) PromptPassword(label string, x, y, width int) (string, error) {
	return ui.PromptPassword(t, label, x, y, width)
}

// ShowTimedError displays an error message for 5 seconds then clears it
func (t *TelnetIO) ShowTimedError(message string, x, y int) {
	ui.ShowTimedError(t, message, x, y)
}

// HandleEscQuit shows quit confirmation and returns true if user wants to quit
func (t *TelnetIO) HandleEscQuit() bool {
	return ui.HandleEscQuit(t)
}

// ClearField clears a form field and resets cursor position
func (t *TelnetIO) ClearField(label string, x, y, width int) {
	ui.ClearField(t, label, x, y, width)
}

// ShowPersistentEscIndicator displays persistent ESC quit option
func (t *TelnetIO) ShowPersistentEscIndicator(x, y int) {
	ui.ShowPersistentEscIndicator(t, x, y)
}

// Pause waits for any key press and shows centered message
func (t *TelnetIO) Pause() error {
	return ui.Pause(t)
}

// PrintAnsi displays embedded ANSI art content to telnet client
func (t *TelnetIO) PrintAnsi(artName string, delay int, height int) error {
	if err := ui.PrintAnsiTerminal(t, artName, delay, height); err != nil {
		return t.Printf("%s\r\n", err.Error())
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
		option, err := t.Reader.ReadByte()
		if err != nil {
			return
		}
		// Handle NAWS subnegotiation
		if option == 31 { // NAWS option
			t.handleNAWSSubnegotiation()
		}
		// For other options, just consume the option - could respond appropriately later
	}
}

// handleNAWSSubnegotiation parses NAWS (Negotiate About Window Size) responses
func (t *TelnetIO) handleNAWSSubnegotiation() {
	// Read the subnegotiation data: IAC SB NAWS <width-high> <width-low> <height-high> <height-low> IAC SE
	// Skip the IAC and SB bytes that were already consumed

	// Read width high byte
	widthHigh, err := t.Reader.ReadByte()
	if err != nil {
		return
	}

	// Read width low byte
	widthLow, err := t.Reader.ReadByte()
	if err != nil {
		return
	}

	// Read height high byte
	heightHigh, err := t.Reader.ReadByte()
	if err != nil {
		return
	}

	// Read height low byte
	heightLow, err := t.Reader.ReadByte()
	if err != nil {
		return
	}

	// Calculate width and height (big-endian)
	width := int(widthHigh)<<8 | int(widthLow)
	height := int(heightHigh)<<8 | int(heightLow)

	// Cap at 80x25 maximum as specified
	if width > 80 {
		width = 80
	}
	if height > 25 {
		height = 25
	}

	// Store dimensions in session
	if t.Session != nil {
		t.Session.Width = width
		t.Session.Height = height
	}
}
