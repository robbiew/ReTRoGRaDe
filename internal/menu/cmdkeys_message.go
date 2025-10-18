package menu

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robbiew/retrograde/internal/jam"
	"github.com/robbiew/retrograde/internal/ui"
)

// registerMessageCommands registers all message-related commands (Message + Message Scanning)
func registerMessageCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Message System
		{CmdKey: "MA", Name: "Change Message Base", Description: "Change to another message base", Category: "Message"},
		{CmdKey: "ME", Name: "Send Private Mail", Description: "Send private mail to a user", Category: "Message"},
		{CmdKey: "MK", Name: "Edit Outgoing Mail", Description: "Edit or delete outgoing private mail", Category: "Message"},
		{CmdKey: "ML", Name: "Send Mass Mail", Description: "Send private mail to multiple users", Category: "Message"},
		{CmdKey: "MM", Name: "Read Private Mail", Description: "Read your private mail", Category: "Message"},
		{CmdKey: "MN", Name: "New Message Scan", Description: "Scan for new messages", Category: "Message"},
		{CmdKey: "MP", Name: "Post Message", Description: "Post a message in the current base", Category: "Message", Handler: handlePostMessage, Implemented: true},
		{CmdKey: "MR", Name: "Read Messages", Description: "Read messages in the current base", Category: "Message", Handler: handleReadMessages, Implemented: true},
		{CmdKey: "MS", Name: "Scan Current Base", Description: "Scan the current message base", Category: "Message"},
		{CmdKey: "MU", Name: "List Base Access", Description: "List users with access to the current base", Category: "Message"},
		{CmdKey: "MY", Name: "Scan for Personal Mail", Description: "Scan message bases for personal messages", Category: "Message"},
		{CmdKey: "MZ", Name: "Set Message NewScan List", Description: "Select message bases to include in new scan", Category: "Message"},
		{CmdKey: "M#", Name: "Quick Message Base Change", Description: "Prompt for a message base to change to", Category: "Message"},

		// Message Scanning (READP.MNU)
		{CmdKey: "RA", Name: "Read Again", Description: "Re-read the current message", Category: "Message Scanning"},
		{CmdKey: "RB", Name: "Back in Thread", Description: "Move backward in the message thread", Category: "Message Scanning"},
		{CmdKey: "RC", Name: "Continuous Reading", Description: "Toggle continuous message reading", Category: "Message Scanning"},
		{CmdKey: "RD", Name: "Delete Message", Description: "Delete the current message", Category: "Message Scanning"},
		{CmdKey: "RE", Name: "Edit Message", Description: "Edit the current message", Category: "Message Scanning"},
		{CmdKey: "RF", Name: "Forward in Thread", Description: "Move forward in the message thread", Category: "Message Scanning"},
		{CmdKey: "RG", Name: "Next Message Base", Description: "Go to the next message base", Category: "Message Scanning"},
		{CmdKey: "RH", Name: "Set High-Read Pointer", Description: "Set the high-read pointer", Category: "Message Scanning"},
		{CmdKey: "RI", Name: "Ignore Remaining Messages", Description: "Ignore remaining messages and set pointer", Category: "Message Scanning"},
		{CmdKey: "RL", Name: "List Messages", Description: "List messages in the current base", Category: "Message Scanning"},
		{CmdKey: "RM", Name: "Move Message", Description: "Move the current message", Category: "Message Scanning"},
		{CmdKey: "RN", Name: "Next Message", Description: "Read the next message", Category: "Message Scanning"},
		{CmdKey: "RQ", Name: "Quit Reading", Description: "Quit the message reader", Category: "Message Scanning"},
		{CmdKey: "RR", Name: "Reply to Message", Description: "Reply to the current message", Category: "Message Scanning"},
		{CmdKey: "RT", Name: "Toggle Base NewScan", Description: "Toggle newscan for the message base", Category: "Message Scanning"},
		{CmdKey: "RU", Name: "Edit Message Author", Description: "Edit the user associated with the message", Category: "Message Scanning"},
		{CmdKey: "RX", Name: "Extract Message", Description: "Extract the message to a file", Category: "Message Scanning"},
		{CmdKey: "R#", Name: "Jump to Message", Description: "Jump directly to a message number", Category: "Message Scanning"},
		{CmdKey: "R-", Name: "Previous Message", Description: "Read the previous message", Category: "Message Scanning"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}

// handlePostMessage handles the MP (Post Message) command
func handlePostMessage(ctx *ExecutionContext, options string) error {
	io := ctx.IO
	session := ctx.Session

	// Check if user has a current message area
	if session.CurrentMessageArea == nil {
		io.Print(ui.Ansi.RedHi + "\r\n No message area selected.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		return nil
	}

	// Check write access level
	userSecLevel := session.SecurityLevel
	requiredLevel := 10 // Default public level
	if session.CurrentMessageArea.WriteSecLevel != "public" {
		// TODO: Parse security level from WriteSecLevel when we implement proper security levels
	}

	if userSecLevel < requiredLevel {
		io.Print(ui.Ansi.RedHi + "\r\n You don't have permission to post in this area.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		return nil
	}

	io.ClearScreen()
	io.Printf(ui.Ansi.CyanHi+"\r\n Post Message in: %s\r\n"+ui.Ansi.Reset, session.CurrentMessageArea.Name)
	io.Print("\r\n")

	// Get To field
	to, err := ui.PromptSimple(io, " To: ", 50, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgBlack, "All")
	if err != nil {
		if err.Error() == "ESC_PRESSED" {
			io.ClearScreen()
			return nil
		}
		return err
	}
	if strings.TrimSpace(to) == "" {
		to = "All"
	}

	// Get Subject field
	subject, err := ui.PromptSimple(io, " Subject: ", 60, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgBlue, "")
	if err != nil {
		if err.Error() == "ESC_PRESSED" {
			io.ClearScreen()
			return nil
		}
		return err
	}
	if strings.TrimSpace(subject) == "" {
		io.Print(ui.Ansi.RedHi + "\r\n Subject is required.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		io.ClearScreen()
		return nil
	}

	// Get message text using the simple text editor
	io.Print("\r\n" + ui.Ansi.Yellow + " Enter your message (empty line to end):\r\n\r\n" + ui.Ansi.Reset)

	var lines []string
	lineNumber := 1

	for {
		io.Printf("%2d: ", lineNumber)
		line, err := ui.PromptSimple(io, "", 75, ui.Ansi.Green, ui.Ansi.White, ui.Ansi.BgBlack, "")
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				io.ClearScreen()
				return nil
			}
			return err
		}

		if strings.TrimSpace(line) == "" {
			break
		}

		lines = append(lines, line)
		lineNumber++

		// Limit message length
		if len(lines) >= 100 {
			io.Print(ui.Ansi.Yellow + "\r\n Maximum message length reached.\r\n" + ui.Ansi.Reset)
			break
		}
	}

	if len(lines) == 0 {
		io.Print(ui.Ansi.RedHi + "\r\n Empty message not posted.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		io.ClearScreen()
		return nil
	}

	text := strings.Join(lines, "\n")

	// Confirm posting
	io.Print("\r\n" + ui.Ansi.Yellow + " Post this message? (Y/N): " + ui.Ansi.Reset)
	key, err := io.GetKeyPress()
	if err != nil {
		return err
	}

	io.Printf("%c\r\n", key)
	if key != 'Y' && key != 'y' {
		io.Print(ui.Ansi.Yellow + "\r\n Message not posted.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		io.ClearScreen()
		return nil
	}

	// Create JAM message base path
	jamPath := session.GetCurrentMessageAreaPath()
	if jamPath == "" {
		io.Print(ui.Ansi.RedHi + "\r\n Invalid message area path.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(jamPath), 0755); err != nil {
		io.Printf(ui.Ansi.RedHi+"\r\n Error creating message directory: %v\r\n"+ui.Ansi.Reset, err)
		ui.Pause(io)
		io.ClearScreen()
		return nil
	}

	// Open JAM message base
	jamBase, err := jam.Open(jamPath)
	if err != nil {
		io.Printf(ui.Ansi.RedHi+"\r\n Error opening message base: %v\r\n"+ui.Ansi.Reset, err)
		ui.Pause(io)
		io.ClearScreen()
		return nil
	}
	defer jamBase.Close()

	// Create message
	message := jam.NewMessage()
	message.From = ctx.Username
	message.To = to
	message.Subject = subject
	message.Text = text
	message.DateTime = time.Now()

	// Write message to JAM base
	msgNum, err := jamBase.WriteMessage(message)
	if err != nil {
		io.Printf(ui.Ansi.RedHi+"\r\n Error saving message: %v\r\n"+ui.Ansi.Reset, err)
		ui.Pause(io)
		io.ClearScreen()
		return nil
	}

	io.Printf(ui.Ansi.GreenHi+"\r\n Message #%d posted successfully!\r\n"+ui.Ansi.Reset, msgNum)
	ui.Pause(io)
	io.ClearScreen()
	return nil
}

// handleReadMessages handles the MR (Read Messages) command
func handleReadMessages(ctx *ExecutionContext, options string) error {
	io := ctx.IO
	session := ctx.Session

	// Check if user has a current message area
	if session.CurrentMessageArea == nil {
		io.Print(ui.Ansi.RedHi + "\r\n No message area selected.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		return nil
	}

	// Check read access level
	userSecLevel := session.SecurityLevel
	requiredLevel := 10 // Default public level
	if session.CurrentMessageArea.ReadSecLevel != "public" {
		// TODO: Parse security level from ReadSecLevel when we implement proper security levels
	}

	if userSecLevel < requiredLevel {
		io.Print(ui.Ansi.RedHi + "\r\n You don't have permission to read messages in this area.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		return nil
	}

	// Create JAM message base path
	jamPath := session.GetCurrentMessageAreaPath()
	if jamPath == "" {
		io.Print(ui.Ansi.RedHi + "\r\n Invalid message area path.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		return nil
	}

	// Open JAM message base
	jamBase, err := jam.Open(jamPath)
	if err != nil {
		io.Printf(ui.Ansi.RedHi+"\r\n Error opening message base: %v\r\n"+ui.Ansi.Reset, err)
		ui.Pause(io)
		io.ClearScreen()
		return nil
	}
	defer jamBase.Close()

	// Get message count
	count, err := jamBase.GetMessageCount()
	if err != nil {
		io.Printf(ui.Ansi.RedHi+"\r\n Error getting message count: %v\r\n"+ui.Ansi.Reset, err)
		ui.Pause(io)
		return nil
	}

	if count == 0 {
		io.Print(ui.Ansi.Yellow + "\r\n No messages in this area.\r\n" + ui.Ansi.Reset)
		ui.Pause(io)
		return nil
	}

	// Get next unread message or start from message 1
	currentMsg, err := jamBase.GetNextUnreadMessage(ctx.Username)
	if err != nil {
		// If no unread messages or error, start from message 1
		currentMsg = 1
	}

	// Message reading loop
	for {
		if currentMsg < 1 {
			currentMsg = 1
		}
		if currentMsg > count {
			currentMsg = count
		}

		// Read the message
		message, err := jamBase.ReadMessage(currentMsg)
		if err != nil {
			io.Printf(ui.Ansi.RedHi+"\r\n Error reading message %d: %v\r\n"+ui.Ansi.Reset, currentMsg, err)
			break
		}

		// Skip deleted messages
		if message.IsDeleted() {
			currentMsg++
			if currentMsg > count {
				io.Print(ui.Ansi.Yellow + "\r\n No more messages.\r\n" + ui.Ansi.Reset)
				break
			}
			continue
		}

		// Display message
		io.ClearScreen()
		io.Printf(ui.Ansi.CyanHi+"\r\n Message %d of %d in %s\r\n"+ui.Ansi.Reset,
			currentMsg, count, session.CurrentMessageArea.Name)
		io.Print(ui.Ansi.Cyan + strings.Repeat("-", 70) + ui.Ansi.Reset + "\r\n")
		io.Printf(ui.Ansi.WhiteHi+"From: "+ui.Ansi.Yellow+"%s\r\n"+ui.Ansi.Reset, message.From)
		io.Printf(ui.Ansi.WhiteHi+"To: "+ui.Ansi.Yellow+"%s\r\n"+ui.Ansi.Reset, message.To)
		io.Printf(ui.Ansi.WhiteHi+"Subject: "+ui.Ansi.Yellow+"%s\r\n"+ui.Ansi.Reset, message.Subject)
		io.Printf(ui.Ansi.WhiteHi+"Date: "+ui.Ansi.Yellow+"%s\r\n"+ui.Ansi.Reset,
			message.DateTime.Format("2006-01-02 15:04:05"))
		io.Print(ui.Ansi.Cyan + strings.Repeat("-", 70) + ui.Ansi.Reset + "\r\n\r\n")

		// Display message text
		io.Print(ui.Ansi.Green + message.Text + ui.Ansi.Reset + "\r\n\r\n")

		// Mark message as read
		jamBase.MarkMessageRead(ctx.Username, currentMsg)

		// Show prompt
		io.Print(ui.Ansi.Cyan + " (N)ext, (P)revious, (Q)uit: " + ui.Ansi.Reset)
		key, err := io.GetKeyPress()
		if err != nil {
			break
		}

		io.Printf("%c\r\n", key)

		switch key {
		case 'N', 'n', ' ', '\r':
			currentMsg++
			if currentMsg > count {
				io.Print(ui.Ansi.Yellow + "\r\n No more messages.\r\n" + ui.Ansi.Reset)
				ui.Pause(io)
				io.ClearScreen()
				return nil
			}
		case 'P', 'p':
			currentMsg--
			if currentMsg < 1 {
				io.Print(ui.Ansi.Yellow + "\r\n At first message.\r\n" + ui.Ansi.Reset)
				currentMsg = 1
			}
		case 'Q', 'q', 27: // ESC
			io.ClearScreen()
			return nil
		default:
			// Ignore other keys
		}
	}

	return nil
}
