// scanner.go - Message scanning and processing
package jam

import (
	"fmt"
	"strings"
	"time"
)

// Scanner scans messages in a JAM base
type Scanner struct {
	base *JAMBase
}

// NewScanner creates a new message scanner
func NewScanner(base *JAMBase) *Scanner {
	return &Scanner{base: base}
}

// ScanMessages scans messages with a filter function
func (s *Scanner) ScanMessages(filter func(*Message) bool) ([]*Message, error) {
	count, err := s.base.GetMessageCount()
	if err != nil {
		return nil, err
	}

	var messages []*Message

	for i := 1; i <= count; i++ {
		msg, err := s.base.ReadMessage(i)
		if err != nil {
			continue
		}

		if filter(msg) {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// ScanUnread scans unread messages for a user
func (s *Scanner) ScanUnread(username string) ([]*Message, error) {
	lr, err := s.base.GetLastRead(username)
	if err != nil {
		// No last read record, all messages are unread
		lr = &LastRead{LastReadMsg: 0}
	}

	return s.ScanMessages(func(msg *Message) bool {
		return msg.Header.MessageNumber > lr.LastReadMsg && !msg.IsDeleted()
	})
}

// ScanByDate scans messages after a specific date
func (s *Scanner) ScanByDate(after time.Time) ([]*Message, error) {
	afterUnix := uint32(after.Unix())

	return s.ScanMessages(func(msg *Message) bool {
		return msg.Header.DateWritten >= afterUnix && !msg.IsDeleted()
	})
}

// ScanByFrom scans messages from a specific sender
func (s *Scanner) ScanByFrom(from string) ([]*Message, error) {
	fromLower := strings.ToLower(from)

	return s.ScanMessages(func(msg *Message) bool {
		return strings.Contains(strings.ToLower(msg.From), fromLower) && !msg.IsDeleted()
	})
}

// ScanByTo scans messages to a specific recipient
func (s *Scanner) ScanByTo(to string) ([]*Message, error) {
	toLower := strings.ToLower(to)

	return s.ScanMessages(func(msg *Message) bool {
		return strings.Contains(strings.ToLower(msg.To), toLower) && !msg.IsDeleted()
	})
}

// ScanBySubject scans messages by subject
func (s *Scanner) ScanBySubject(subject string) ([]*Message, error) {
	subjectLower := strings.ToLower(subject)

	return s.ScanMessages(func(msg *Message) bool {
		return strings.Contains(strings.ToLower(msg.Subject), subjectLower) && !msg.IsDeleted()
	})
}

// ScanPrivate scans private messages for a user
func (s *Scanner) ScanPrivate(username string) ([]*Message, error) {
	usernameLower := strings.ToLower(username)

	return s.ScanMessages(func(msg *Message) bool {
		return msg.IsPrivate() &&
			strings.ToLower(msg.To) == usernameLower &&
			!msg.IsDeleted()
	})
}

// ScanReplies finds all replies to a message
func (s *Scanner) ScanReplies(msgNum int) ([]*Message, error) {
	originalMsg, err := s.base.ReadMessage(msgNum)
	if err != nil {
		return nil, err
	}

	return s.ScanMessages(func(msg *Message) bool {
		return msg.ReplyID == originalMsg.MsgID && !msg.IsDeleted()
	})
}

// ExportToText exports messages to text format
func (s *Scanner) ExportToText(messages []*Message) string {
	var output strings.Builder

	for i, msg := range messages {
		output.WriteString(fmt.Sprintf("Message %d\n", i+1))
		output.WriteString(strings.Repeat("=", 76))
		output.WriteString("\n")
		output.WriteString(fmt.Sprintf("From: %s (%s)\n", msg.From, msg.OrigAddr))
		output.WriteString(fmt.Sprintf("To: %s\n", msg.To))
		output.WriteString(fmt.Sprintf("Subject: %s\n", msg.Subject))
		output.WriteString(fmt.Sprintf("Date: %s\n", msg.DateTime.Format("2006-01-02 15:04:05")))
		output.WriteString("\n")
		output.WriteString(msg.Text)
		output.WriteString("\n\n")
	}

	return output.String()
}
