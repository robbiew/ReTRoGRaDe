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
} // ColorPrint sends colored text to the telnet client
// fgColor: foreground color (e.g., ui.Ansi.Cyan)
// bgColor: background color (e.g., ui.Ansi.BgBlue) - use "" for no background
func (t *TelnetIO) ColorPrint(text string, fgColor, bgColor string) error {
	// Build the colored output
	output := ""
	if fgColor != "" {
		output += fgColor
	}
	if bgColor != "" {
		output += bgColor
	}
	output += text
	if fgColor != "" || bgColor != "" {
		output += ui.Ansi.Reset
	}

	return t.Print(output)
}

// ColorPrintf sends formatted colored text to the telnet client
// fgColor: foreground color (e.g., ui.Ansi.Cyan)
// bgColor: background color (e.g., ui.Ansi.BgBlue) - use "" for no background
func (t *TelnetIO) ColorPrintf(format string, fgColor, bgColor string, args ...interface{}) error {
	text := fmt.Sprintf(format, args...)
	return t.ColorPrint(text, fgColor, bgColor)
}

// PrintSuccess prints a success message with default success colors
func (t *TelnetIO) PrintSuccess(text string) error {
	return t.ColorPrint(text, ui.Ansi.GreenHi, "")
}

// PrintError prints an error message with default error colors
func (t *TelnetIO) PrintError(text string) error {
	return t.ColorPrint(text, ui.Ansi.RedHi, "")
}

// PrintWarning prints a warning message with default warning colors
func (t *TelnetIO) PrintWarning(text string) error {
	return t.ColorPrint(text, ui.Ansi.YellowHi, "")
}

// PrintInfo prints an info message with default info colors
func (t *TelnetIO) PrintInfo(text string) error {
	return t.ColorPrint(text, ui.Ansi.Cyan, "")
}

// ClearScreen clears the telnet client screen
func (t *TelnetIO) ClearScreen() error {
	return t.Print(ui.ClearScreenSequence())
}

// MoveCursor moves cursor to specific position
func (t *TelnetIO) MoveCursor(x, y int) error {
	return t.Printf("%s", ui.MoveCursorSequence(x, y))
}

// Pause waits for any key press and shows centered message
func (t *TelnetIO) Pause() error {
	return ui.Pause(t)
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
