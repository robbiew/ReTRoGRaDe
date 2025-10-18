package telnet

import (
	"bufio"
	"fmt"
	"net"
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
	seq, err := t.ReadKeySequence(0)
	if err != nil {
		return 0, err
	}
	if len(seq) == 0 {
		return 0, nil
	}
	key := seq[0]
	if key >= 'a' && key <= 'z' {
		key -= 32
	}
	return key, nil
}

// GetKeyPressUpperWithTimeout reads a key press with a timeout, returning an error if the deadline expires.
func (t *TelnetIO) GetKeyPressUpperWithTimeout(timeout time.Duration) (byte, error) {
	seq, err := t.ReadKeySequence(timeout)
	if err != nil {
		return 0, err
	}
	if len(seq) == 0 {
		return 0, nil
	}
	key := seq[0]
	if key >= 'a' && key <= 'z' {
		key -= 32
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

// MoveCursor moves cursor to specific position
func (t *TelnetIO) MoveCursor(x, y int) error {
	return t.Printf("%s", ui.MoveCursorSequence(x, y))
}

// Pause waits for any key press and shows centered message
func (t *TelnetIO) Pause() error {
	width := 0
	if t.Session != nil && t.Session.Width > 0 {
		width = t.Session.Width
	}
	return ui.PauseWithText(t, "", width)
}

// ReadKeySequence reads a key or escape sequence, honoring an optional timeout.
func (t *TelnetIO) ReadKeySequence(timeout time.Duration) (string, error) {
	if t.Reader == nil {
		return "", fmt.Errorf("telnet reader is not initialized")
	}

	var conn net.Conn
	if t.Session != nil {
		conn = t.Session.Conn
	}

	if timeout > 0 && conn != nil {
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			conn = nil
		} else {
			defer conn.SetReadDeadline(time.Time{})
		}
	}

	b, err := t.Reader.ReadByte()
	if err != nil {
		return "", err
	}

	// Handle telnet command sequences (IAC = 255)
	if b == 255 {
		t.handleTelnetCommand()
		return t.ReadKeySequence(timeout)
	}

	if t.Session != nil {
		t.Session.LastActivity = time.Now()
	}

	seq := []byte{b}

	switch b {
	case 27: // ESC
		seq = append(seq, t.readEscapeSequence(conn)...)
	case '\r':
		if t.Reader.Buffered() > 0 {
			if next, err := t.Reader.Peek(1); err == nil && len(next) == 1 && next[0] == '\n' {
				t.Reader.ReadByte() // consume LF
				seq = append(seq, '\n')
			}
		}
	}

	return string(seq), nil
}

func (t *TelnetIO) readEscapeSequence(conn net.Conn) []byte {
	var seq []byte
	for i := 0; i < 5; i++ {
		if conn == nil && t.Reader.Buffered() == 0 {
			break
		}
		if conn != nil && t.Reader.Buffered() == 0 {
			if err := conn.SetReadDeadline(time.Now().Add(5 * time.Millisecond)); err != nil {
				conn = nil
			}
		}

		b, err := t.Reader.ReadByte()
		if conn != nil {
			conn.SetReadDeadline(time.Time{})
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			break
		}

		seq = append(seq, b)

		if len(seq) == 1 && seq[0] != '[' && seq[0] != 'O' {
			break
		}

		if len(seq) >= 2 && seq[0] == 'O' {
			break
		}

		if len(seq) >= 2 && seq[0] == '[' {
			last := seq[len(seq)-1]
			if last == '~' || (last >= 'A' && last <= 'Z') {
				break
			}
		}
	}
	return seq
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

func (t *TelnetIO) handleNAWSSubnegotiation() {
	// Read the subnegotiation data: IAC SB NAWS <width-high> <width-low> <height-high> <height-low> IAC SE

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

	fmt.Printf("DEBUG: NAWS received - Width: %d, Height: %d\n", width, height)

	// If NAWS returned invalid values, use defaults
	if width <= 0 || height <= 0 || width > 255 || height > 255 {
		height, width = getDefaultTermSize()
		fmt.Printf("DEBUG: NAWS invalid, using defaults - Width: %d, Height: %d\n", width, height)
	}

	// Cap at reasonable maximums
	if width > 80 {
		width = 80
	}
	if height > 24 {
		height = 24
	}

	// Store dimensions in session
	if t.Session != nil {
		t.Session.Width = width
		t.Session.Height = height
	}
}

// getDefaultTermSize returns safe default terminal dimensions
func getDefaultTermSize() (height, width int) {
	return 25, 80 // Standard BBS terminal size
}
