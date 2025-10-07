// message.go - Message building and utilities
package jam

import (
	"fmt"
	"strings"
	"time"
)

// MessageBuilder helps build messages
type MessageBuilder struct {
	msg *Message
}

// NewMessage creates a new message builder
func NewMessage() *MessageBuilder {
	return &MessageBuilder{
		msg: &Message{
			DateTime: time.Now(),
			Kludges:  []string{},
		},
	}
}

// From sets the sender
func (mb *MessageBuilder) From(from string) *MessageBuilder {
	mb.msg.From = from
	return mb
}

// To sets the recipient
func (mb *MessageBuilder) To(to string) *MessageBuilder {
	mb.msg.To = to
	return mb
}

// Subject sets the subject
func (mb *MessageBuilder) Subject(subject string) *MessageBuilder {
	mb.msg.Subject = subject
	return mb
}

// Text sets the message text
func (mb *MessageBuilder) Text(text string) *MessageBuilder {
	mb.msg.Text = text
	return mb
}

// OrigAddr sets the origin address
func (mb *MessageBuilder) OrigAddr(addr string) *MessageBuilder {
	mb.msg.OrigAddr = addr
	return mb
}

// DestAddr sets the destination address
func (mb *MessageBuilder) DestAddr(addr string) *MessageBuilder {
	mb.msg.DestAddr = addr
	return mb
}

// MsgID sets the message ID
func (mb *MessageBuilder) MsgID(msgid string) *MessageBuilder {
	mb.msg.MsgID = msgid
	return mb
}

// ReplyID sets the reply ID
func (mb *MessageBuilder) ReplyID(replyid string) *MessageBuilder {
	mb.msg.ReplyID = replyid
	return mb
}

// PID sets the PID
func (mb *MessageBuilder) PID(pid string) *MessageBuilder {
	mb.msg.PID = pid
	return mb
}

// AddKludge adds a kludge line
func (mb *MessageBuilder) AddKludge(kludge string) *MessageBuilder {
	mb.msg.Kludges = append(mb.msg.Kludges, kludge)
	return mb
}

// Private marks the message as private
func (mb *MessageBuilder) Private() *MessageBuilder {
	if mb.msg.Header == nil {
		mb.msg.Header = &MessageHeader{}
	}
	mb.msg.Header.Attribute |= MSG_PRIVATE
	return mb
}

// Local marks the message as local
func (mb *MessageBuilder) Local() *MessageBuilder {
	if mb.msg.Header == nil {
		mb.msg.Header = &MessageHeader{}
	}
	mb.msg.Header.Attribute |= MSG_TYPELOCAL
	return mb
}

// Echo marks the message as echomail
func (mb *MessageBuilder) Echo() *MessageBuilder {
	if mb.msg.Header == nil {
		mb.msg.Header = &MessageHeader{}
	}
	mb.msg.Header.Attribute |= MSG_TYPEECHO
	return mb
}

// Netmail marks the message as netmail
func (mb *MessageBuilder) Netmail() *MessageBuilder {
	if mb.msg.Header == nil {
		mb.msg.Header = &MessageHeader{}
	}
	mb.msg.Header.Attribute |= MSG_TYPENET
	return mb
}

// Crash marks the message as crash priority
func (mb *MessageBuilder) Crash() *MessageBuilder {
	if mb.msg.Header == nil {
		mb.msg.Header = &MessageHeader{}
	}
	mb.msg.Header.Attribute |= MSG_CRASH
	return mb
}

// Hold marks the message as hold
func (mb *MessageBuilder) Hold() *MessageBuilder {
	if mb.msg.Header == nil {
		mb.msg.Header = &MessageHeader{}
	}
	mb.msg.Header.Attribute |= MSG_HOLD
	return mb
}

// Build returns the constructed message
func (mb *MessageBuilder) Build() *Message {
	return mb.msg
}

// GenerateMsgID generates a FidoNet-style MSGID
func GenerateMsgID(address string, serial uint32) string {
	return fmt.Sprintf("%s %08x", address, serial)
}

// ParseAddress parses a FidoNet address
func ParseAddress(addr string) (zone, net, node, point int, domain string, err error) {
	// Format: zone:net/node.point@domain
	parts := strings.Split(addr, "@")
	if len(parts) > 1 {
		domain = parts[1]
		addr = parts[0]
	}

	// Parse zone:net/node.point
	var zoneNet, nodePoint string
	if idx := strings.Index(addr, ":"); idx >= 0 {
		fmt.Sscanf(addr[:idx], "%d", &zone)
		zoneNet = addr[idx+1:]
	} else {
		zoneNet = addr
	}

	if idx := strings.Index(zoneNet, "/"); idx >= 0 {
		fmt.Sscanf(zoneNet[:idx], "%d", &net)
		nodePoint = zoneNet[idx+1:]
	} else {
		return 0, 0, 0, 0, "", fmt.Errorf("invalid address format")
	}

	if idx := strings.Index(nodePoint, "."); idx >= 0 {
		fmt.Sscanf(nodePoint[:idx], "%d", &node)
		fmt.Sscanf(nodePoint[idx+1:], "%d", &point)
	} else {
		fmt.Sscanf(nodePoint, "%d", &node)
		point = 0
	}

	return
}

// FormatAddress formats a FidoNet address
func FormatAddress(zone, net, node, point int, domain string) string {
	addr := fmt.Sprintf("%d:%d/%d", zone, net, node)
	if point > 0 {
		addr += fmt.Sprintf(".%d", point)
	}
	if domain != "" {
		addr += "@" + domain
	}
	return addr
}

// QuoteMessage quotes a message for replying
func QuoteMessage(text, initials string) string {
	lines := strings.Split(text, "\n")
	var quoted []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			quoted = append(quoted, fmt.Sprintf(" %s> %s", initials, line))
		}
	}

	return strings.Join(quoted, "\n")
}

// GetInitials extracts initials from a name
func GetInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return ""
	}

	var initials strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			initials.WriteByte(part[0])
		}
	}

	return strings.ToUpper(initials.String())
}

// WrapText wraps text to a specific line width
func WrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
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

	return strings.Join(lines, "\n")
}

// AddOriginLine adds a Fidonet origin line to message text
func AddOriginLine(text, systemName, address string) string {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += fmt.Sprintf("--- \n * Origin: %s (%s)\n", systemName, address)
	return text
}

// AddTearline adds a tearline to message text
func AddTearline(text, program string) string {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += fmt.Sprintf("--- %s\n", program)
	return text
}
