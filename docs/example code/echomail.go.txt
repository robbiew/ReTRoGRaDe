// echomail.go - Echomail specific functions
package jam

import (
	"fmt"
	"strings"
	"time"
)

// EchomailBuilder helps build echomail messages
type EchomailBuilder struct {
	mb      *MessageBuilder
	seenBy  []string
	path    []string
	echoTag string
}

// NewEchomail creates a new echomail message builder
func NewEchomail(echoTag string) *EchomailBuilder {
	return &EchomailBuilder{
		mb:      NewMessage().Echo(),
		seenBy:  []string{},
		path:    []string{},
		echoTag: echoTag,
	}
}

// From sets the sender name and address
func (eb *EchomailBuilder) From(name, address string) *EchomailBuilder {
	eb.mb.From(name).OrigAddr(address)
	return eb
}

// To sets the recipient name
func (eb *EchomailBuilder) To(name string) *EchomailBuilder {
	eb.mb.To(name)
	return eb
}

// Subject sets the subject
func (eb *EchomailBuilder) Subject(subject string) *EchomailBuilder {
	eb.mb.Subject(subject)
	return eb
}

// Text sets the message text
func (eb *EchomailBuilder) Text(text string) *EchomailBuilder {
	eb.mb.Text(text)
	return eb
}

// AddSeenBy adds a node to SEEN-BY
func (eb *EchomailBuilder) AddSeenBy(address string) *EchomailBuilder {
	// Convert to 2D format (net/node)
	_, net, node, _, _, _ := ParseAddress(address)
	addr2D := fmt.Sprintf("%d/%d", net, node)

	// Check if already in list
	for _, sb := range eb.seenBy {
		if sb == addr2D {
			return eb
		}
	}

	eb.seenBy = append(eb.seenBy, addr2D)
	return eb
}

// AddPath adds a node to PATH
func (eb *EchomailBuilder) AddPath(address string) *EchomailBuilder {
	// Convert to 2D format (net/node)
	_, net, node, _, _, _ := ParseAddress(address)
	addr2D := fmt.Sprintf("%d/%d", net, node)

	eb.path = append(eb.path, addr2D)
	return eb
}

// AddVia adds a Via line with timestamp
func (eb *EchomailBuilder) AddVia(address, program string) *EchomailBuilder {
	timestamp := time.Now().Format("20060102150405")
	via := fmt.Sprintf("Via %s @%s %s", address, timestamp, program)
	eb.mb.AddKludge(via)
	return eb
}

// AddArea adds AREA kludge
func (eb *EchomailBuilder) AddArea() *EchomailBuilder {
	eb.mb.AddKludge("AREA:" + eb.echoTag)
	return eb
}

// Build returns the constructed message
func (eb *EchomailBuilder) Build() *Message {
	// Add SEEN-BY
	if len(eb.seenBy) > 0 {
		seenByStr := strings.Join(eb.seenBy, " ")
		eb.mb.msg.SeenBy = seenByStr
	}

	// Add PATH
	if len(eb.path) > 0 {
		pathStr := strings.Join(eb.path, " ")
		eb.mb.msg.Path = pathStr
	}

	return eb.mb.Build()
}
