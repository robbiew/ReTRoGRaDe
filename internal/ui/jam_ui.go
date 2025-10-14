package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/robbiew/retrograde/internal/jam"
)

type MessageArea struct {
	Conf string
	Area string
	Path string
}

type MessageMenu struct {
	area          MessageArea
	reader        *bufio.Reader
	currentMsg    int
	totalMessages int
	jamBase       *jam.JAMBase
	username      string
	userAddress   string
}

func NewMessageMenu(conf, area, basePath, username, userAddress string) (*MessageMenu, error) {
	jamBase, err := jam.Open(basePath)
	if err != nil {
		return nil, err
	}

	count, err := jamBase.GetMessageCount()
	if err != nil {
		count = 0
	}

	return &MessageMenu{
		area: MessageArea{
			Conf: conf,
			Area: area,
			Path: basePath,
		},
		reader:        bufio.NewReader(os.Stdin),
		currentMsg:    1,
		totalMessages: count,
		jamBase:       jamBase,
		username:      username,
		userAddress:   userAddress,
	}, nil
}

func (m *MessageMenu) Close() error {
	if m.jamBase != nil {
		return m.jamBase.Close()
	}
	return nil
}

func (m *MessageMenu) clearScreen() {
	fmt.Print(ClearScreenSequence())
}

func (m *MessageMenu) displayHeader() {
	m.clearScreen()
	fmt.Println("╔════════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ Message Base - Conf: %-20s Area: %-18s ║\n", m.area.Conf, m.area.Area)
	fmt.Println("╚════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func (m *MessageMenu) displayMainMenu() {
	m.displayHeader()
	fmt.Printf("  Total Messages: %d  Active: %d\n\n",
		m.totalMessages, m.jamBase.GetActiveMessageCount())
	fmt.Println("  [L] List messages")
	fmt.Println("  [R] Read messages")
	fmt.Println("  [P] Post message")
	fmt.Println("  [A] Change area")
	fmt.Println("  [Q] Quit to main menu")
	fmt.Println()
	fmt.Print("Select option: ")
}

func (m *MessageMenu) displayMessageReader(msg *jam.Message, msgNum int) {
	m.displayHeader()
	fmt.Printf("Message %d of %d", msgNum, m.totalMessages)

	if msg.IsPrivate() {
		fmt.Print(" [PRIVATE]")
	}
	if msg.IsDeleted() {
		fmt.Print(" [DELETED]")
	}
	fmt.Println()

	fmt.Println("┌────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ From    : %-62s │\n", truncate(msg.From, 62))
	fmt.Printf("│ To      : %-62s │\n", truncate(msg.To, 62))
	fmt.Printf("│ Subject : %-62s │\n", truncate(msg.Subject, 62))
	fmt.Printf("│ Date    : %-62s │\n", truncate(msg.DateTime.Format("2006-01-02 15:04:05"), 62))

	if msg.OrigAddr != "" {
		fmt.Printf("│ Origin  : %-62s │\n", truncate(msg.OrigAddr, 62))
	}
	if msg.DestAddr != "" {
		fmt.Printf("│ Dest    : %-62s │\n", truncate(msg.DestAddr, 62))
	}

	fmt.Println("├────────────────────────────────────────────────────────────────────────┤")

	// Display message body with word wrapping
	lines := wrapText(msg.Text, 72)
	for _, line := range lines {
		fmt.Printf("│ %-72s │\n", line)
	}

	fmt.Println("└────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Navigation options
	navOptions := []string{}
	if msgNum > 1 {
		navOptions = append(navOptions, "[P] Previous")
	}
	if msgNum < m.totalMessages {
		navOptions = append(navOptions, "[N] Next")
	}
	navOptions = append(navOptions, "[R] Reply", "[D] Delete", "[Q] Quit to menu")

	fmt.Printf("  %s\n", strings.Join(navOptions, "  "))
	fmt.Println()
	fmt.Print("Select option: ")
}

func (m *MessageMenu) displayMessageList(messages []*jam.Message) {
	m.displayHeader()
	fmt.Println("  #    From                 To                   Subject              Date      ")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────")

	for i, msg := range messages {
		flags := ""
		if msg.IsPrivate() {
			flags += "P"
		}
		if msg.IsDeleted() {
			flags += "D"
		}

		fmt.Printf(" %4d %-20s %-20s %-20s %s %2s\n",
			i+1,
			truncate(msg.From, 20),
			truncate(msg.To, 20),
			truncate(msg.Subject, 20),
			msg.DateTime.Format("01-02"),
			flags)
	}

	fmt.Println()
	fmt.Print("Press [Enter] to return to menu: ")
}

func (m *MessageMenu) displayPostEditor() {
	m.displayHeader()
	fmt.Println("Post New Message")
	fmt.Println("────────────────────────────────────────────────────────────────────────")
	fmt.Println()
}

func (m *MessageMenu) getInput() string {
	input, _ := m.reader.ReadString('\n')
	return strings.TrimSpace(strings.ToUpper(input))
}

func (m *MessageMenu) getInputPreserveCase() string {
	input, _ := m.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (m *MessageMenu) Run() {
	defer m.Close()

	for {
		// Refresh message count
		count, _ := m.jamBase.GetMessageCount()
		m.totalMessages = count

		m.displayMainMenu()
		choice := m.getInput()

		switch choice {
		case "L":
			m.handleList()
		case "R":
			m.handleRead()
		case "P":
			m.handlePost()
		case "A":
			m.handleChangeArea()
		case "Q":
			return
		default:
			continue
		}
	}
}

func (m *MessageMenu) handleList() {
	var messages []*jam.Message

	for i := 1; i <= m.totalMessages; i++ {
		msg, err := m.jamBase.ReadMessage(i)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	m.displayMessageList(messages)
	m.reader.ReadString('\n')
}

func (m *MessageMenu) handleRead() {
	// Get last read position
	lr, err := m.jamBase.GetLastRead(m.username)
	if err == nil && lr.LastReadMsg > 0 {
		// Convert message number to index
		for i := 1; i <= m.totalMessages; i++ {
			idx, _ := m.jamBase.ReadIndexRecord(i)
			if idx != nil && idx.MessageNum == lr.LastReadMsg {
				m.currentMsg = i
				break
			}
		}
	} else {
		m.currentMsg = 1
	}

	for {
		if m.currentMsg < 1 {
			m.currentMsg = 1
		}
		if m.currentMsg > m.totalMessages {
			m.currentMsg = m.totalMessages
		}

		msg, err := m.jamBase.ReadMessage(m.currentMsg)
		if err != nil {
			fmt.Printf("Error reading message: %v\n", err)
			m.reader.ReadString('\n')
			return
		}

		m.displayMessageReader(msg, m.currentMsg)

		// Update last read
		m.jamBase.SetLastRead(m.username, msg.Header.MessageNumber, msg.Header.MessageNumber)

		choice := m.getInput()

		switch choice {
		case "N":
			if m.currentMsg < m.totalMessages {
				m.currentMsg++
			}
		case "P":
			if m.currentMsg > 1 {
				m.currentMsg--
			}
		case "R":
			m.handleReply(msg)
		case "D":
			m.handleDelete(m.currentMsg)
			if m.currentMsg < m.totalMessages {
				m.currentMsg++
			}
		case "Q":
			return
		}
	}
}

func (m *MessageMenu) handlePost() {
	m.displayPostEditor()

	fmt.Print("To: ")
	to := m.getInputPreserveCase()
	if to == "" {
		return
	}

	fmt.Print("Subject: ")
	subject := m.getInputPreserveCase()
	if subject == "" {
		return
	}

	fmt.Println("\nEnter message body (type /SAVE to save, /ABORT to cancel):")
	fmt.Println()

	var body strings.Builder
	for {
		line := m.getInputPreserveCase()

		if line == "/SAVE" {
			break
		} else if line == "/ABORT" {
			fmt.Println("\nMessage cancelled.")
			m.reader.ReadString('\n')
			return
		}

		body.WriteString(line)
		body.WriteString("\n")
	}

	// Build message
	text := body.String()
	text = jam.AddTearline(text, "GoBBS v1.0")
	text = jam.AddOriginLine(text, "My BBS", m.userAddress)

	msg := jam.NewMessage().
		From(m.username).
		To(to).
		Subject(subject).
		Text(text).
		OrigAddr(m.userAddress).
		Local().
		Build()

	_, err := m.jamBase.WriteMessage(msg)
	if err != nil {
		fmt.Printf("\nError saving message: %v\n", err)
	} else {
		fmt.Println("\nMessage saved!")
	}

	m.reader.ReadString('\n')
}

func (m *MessageMenu) handleReply(original *jam.Message) {
	m.displayPostEditor()

	fmt.Printf("Replying to: %s\n", original.From)

	subject := original.Subject
	if !strings.HasPrefix(strings.ToUpper(subject), "RE:") {
		subject = "Re: " + subject
	}

	fmt.Printf("Subject: %s\n\n", subject)

	// Quote original message
	initials := jam.GetInitials(original.From)
	quoted := jam.QuoteMessage(original.Text, initials)

	fmt.Println("Quoted text:")
	fmt.Println(quoted)
	fmt.Println()
	fmt.Println("Enter your reply (type /SAVE to save, /ABORT to cancel):")
	fmt.Println()

	var body strings.Builder
	body.WriteString(quoted)
	body.WriteString("\n\n")

	for {
		line := m.getInputPreserveCase()

		if line == "/SAVE" {
			break
		} else if line == "/ABORT" {
			fmt.Println("\nReply cancelled.")
			m.reader.ReadString('\n')
			return
		}

		body.WriteString(line)
		body.WriteString("\n")
	}

	// Build reply
	text := body.String()
	text = jam.AddTearline(text, "GoBBS v1.0")
	text = jam.AddOriginLine(text, "My BBS", m.userAddress)

	msg := jam.NewMessage().
		From(m.username).
		To(original.From).
		Subject(subject).
		Text(text).
		OrigAddr(m.userAddress).
		ReplyID(original.MsgID).
		Local().
		Build()

	_, err := m.jamBase.WriteMessage(msg)
	if err != nil {
		fmt.Printf("\nError saving reply: %v\n", err)
	} else {
		fmt.Println("\nReply saved!")
	}

	m.reader.ReadString('\n')
}

func (m *MessageMenu) handleDelete(msgNum int) {
	fmt.Print("Delete this message? (Y/N): ")
	confirm := m.getInput()

	if confirm == "Y" {
		err := m.jamBase.DeleteMessage(msgNum)
		if err != nil {
			fmt.Printf("Error deleting message: %v\n", err)
		} else {
			fmt.Println("Message deleted.")
		}
		time.Sleep(1 * time.Second)
	}
}

func (m *MessageMenu) handleChangeArea() {
	m.displayHeader()
	fmt.Println("Available Areas:")
	fmt.Println("  1. General")
	fmt.Println("  2. Testing")
	fmt.Println("  3. Support")
	fmt.Println()
	fmt.Print("Select area (or Q to cancel): ")

	choice := m.getInput()

	var newArea, newPath string

	switch choice {
	case "1":
		newArea = "General"
		newPath = "/path/to/general"
	case "2":
		newArea = "Testing"
		newPath = "/path/to/testing"
	case "3":
		newArea = "Support"
		newPath = "/path/to/support"
	case "Q":
		return
	default:
		return
	}

	// Close current base
	m.jamBase.Close()

	// Open new base
	jamBase, err := jam.Open(newPath)
	if err != nil {
		fmt.Printf("Error opening area: %v\n", err)
		m.reader.ReadString('\n')
		// Reopen original
		m.jamBase, _ = jam.Open(m.area.Path)
		return
	}

	m.jamBase = jamBase
	m.area.Area = newArea
	m.area.Path = newPath

	count, _ := m.jamBase.GetMessageCount()
	m.totalMessages = count
	m.currentMsg = 1
}

// Helper functions
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, width int) []string {
	var lines []string

	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		var currentLine strings.Builder

		for _, word := range words {
			if currentLine.Len() == 0 {
				currentLine.WriteString(word)
			} else if currentLine.Len()+1+len(word) <= width {
				currentLine.WriteString(" ")
				currentLine.WriteString(word)
			} else {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
				currentLine.WriteString(word)
			}
		}

		if currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
		}
	}

	return lines
}

// func main() {
// 	// Example usage
// 	menu, err := NewMessageMenu(
// 		"Local",                  // Conference
// 		"General",                // Area
// 		"/home/bbs/msgs/general", // JAM base path
// 		"SysOp",                  // Username
// 		"1:123/456",              // User's Fidonet address
// 	)

// 	if err != nil {
// 		fmt.Printf("Error opening message base: %v\n", err)
// 		return
// 	}

// 	menu.Run()
// }
