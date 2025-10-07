// netmail.go - Netmail specific functions
package jam

import (
	"fmt"
)

// NetmailBuilder helps build netmail messages
type NetmailBuilder struct {
	mb *MessageBuilder
}

// NewNetmail creates a new netmail message builder
func NewNetmail() *NetmailBuilder {
	return &NetmailBuilder{
		mb: NewMessage().Netmail(),
	}
}

// From sets the sender name and address
func (nb *NetmailBuilder) From(name, address string) *NetmailBuilder {
	nb.mb.From(name).OrigAddr(address)
	return nb
}

// To sets the recipient name and address
func (nb *NetmailBuilder) To(name, address string) *NetmailBuilder {
	nb.mb.To(name).DestAddr(address)
	return nb
}

// Subject sets the subject
func (nb *NetmailBuilder) Subject(subject string) *NetmailBuilder {
	nb.mb.Subject(subject)
	return nb
}

// Text sets the message text
func (nb *NetmailBuilder) Text(text string) *NetmailBuilder {
	nb.mb.Text(text)
	return nb
}

// Private marks as private
func (nb *NetmailBuilder) Private() *NetmailBuilder {
	nb.mb.Private()
	return nb
}

// Crash marks as crash priority
func (nb *NetmailBuilder) Crash() *NetmailBuilder {
	nb.mb.Crash()
	return nb
}

// Hold marks as hold
func (nb *NetmailBuilder) Hold() *NetmailBuilder {
	nb.mb.Hold()
	return nb
}

// FileAttach adds a file attachment
func (nb *NetmailBuilder) FileAttach(filename string) *NetmailBuilder {
	nb.mb.Subject("^" + filename)
	if nb.mb.msg.Header == nil {
		nb.mb.msg.Header = &MessageHeader{}
	}
	nb.mb.msg.Header.Attribute |= MSG_FILEATTACH

	// Add subfield
	if nb.mb.msg.Header.Subfields == nil {
		nb.mb.msg.Header.Subfields = []Subfield{}
	}
	nb.mb.msg.Header.Subfields = append(
		nb.mb.msg.Header.Subfields,
		CreateSubfield(JAMSFLD_ENCLOSEDFILE, filename),
	)
	return nb
}

// FileRequest adds a file request
func (nb *NetmailBuilder) FileRequest(filemask, password string) *NetmailBuilder {
	nb.mb.Subject(filemask)
	if nb.mb.msg.Header == nil {
		nb.mb.msg.Header = &MessageHeader{}
	}
	nb.mb.msg.Header.Attribute |= MSG_FILEREQUEST

	// Add subfield with optional password
	data := filemask
	if password != "" {
		data += "\x00" + password
	}

	if nb.mb.msg.Header.Subfields == nil {
		nb.mb.msg.Header.Subfields = []Subfield{}
	}
	nb.mb.msg.Header.Subfields = append(
		nb.mb.msg.Header.Subfields,
		CreateSubfield(JAMSFLD_ENCLOSEDFREQ, data),
	)
	return nb
}

// AddINTLKludge adds INTL kludge for zone routing
func (nb *NetmailBuilder) AddINTLKludge() *NetmailBuilder {
	// Parse addresses
	destZone, destNet, destNode, destPoint, _, _ := ParseAddress(nb.mb.msg.DestAddr)
	origZone, origNet, origNode, origPoint, _, _ := ParseAddress(nb.mb.msg.OrigAddr)

	// INTL format: destZone:destNet/destNode origZone:origNet/origNode
	intl := fmt.Sprintf("INTL %d:%d/%d %d:%d/%d",
		destZone, destNet, destNode,
		origZone, origNet, origNode)

	nb.mb.AddKludge(intl)

	// Add TOPT and FMPT if points are involved
	if destPoint > 0 {
		nb.mb.AddKludge(fmt.Sprintf("TOPT %d", destPoint))
	}
	if origPoint > 0 {
		nb.mb.AddKludge(fmt.Sprintf("FMPT %d", origPoint))
	}

	return nb
}

// Build returns the constructed message
func (nb *NetmailBuilder) Build() *Message {
	return nb.mb.Build()
}
