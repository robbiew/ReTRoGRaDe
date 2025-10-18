package jam

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

// ReadMessageHeader reads a message header from the specified message number
func (j *JAMBase) ReadMessageHeader(msgNum int) (*MessageHeader, error) {
	if !j.isOpen {
		return nil, ErrBaseNotOpen
	}

	idx, err := j.ReadIndexRecord(msgNum)
	if err != nil {
		return nil, err
	}

	j.jhrFile.Seek(int64(idx.HdrOffset), 0)

	hdr := &MessageHeader{}

	// Read fixed part of header
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.Signature)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.Revision)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.ReservedWord)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.SubfieldLen)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.TimesRead)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.MSGIDcrc)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.REPLYcrc)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.ReplyTo)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.Reply1st)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.ReplyNext)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.DateWritten)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.DateReceived)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.DateProcessed)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.MessageNumber)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.Attribute)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.Attribute2)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.Offset)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.TxtLen)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.PasswordCRC)
	binary.Read(j.jhrFile, binary.LittleEndian, &hdr.Cost)

	if string(hdr.Signature[:]) != JAMSignature {
		return nil, ErrInvalidSignature
	}

	// Read subfields
	bytesRead := uint32(0)
	for bytesRead < hdr.SubfieldLen {
		subfield := Subfield{}
		binary.Read(j.jhrFile, binary.LittleEndian, &subfield.LoID)
		binary.Read(j.jhrFile, binary.LittleEndian, &subfield.HiID)
		binary.Read(j.jhrFile, binary.LittleEndian, &subfield.DatLen)

		subfield.Buffer = make([]byte, subfield.DatLen)
		j.jhrFile.Read(subfield.Buffer)

		hdr.Subfields = append(hdr.Subfields, subfield)
		bytesRead += SubfieldHdrSize + subfield.DatLen
	}

	return hdr, nil
}

// WriteMessageHeader writes a message header and returns its offset
func (j *JAMBase) WriteMessageHeader(hdr *MessageHeader) (uint32, error) {
	if !j.isOpen {
		return 0, ErrBaseNotOpen
	}

	// Get current position (end of file)
	pos, err := j.jhrFile.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to seek to end of header file: %w", err)
	}

	// If this is the first message, skip the fixed header
	if pos == 0 {
		pos = HeaderSize
		j.jhrFile.Seek(pos, 0)
	}

	// Write fixed part
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Signature)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Revision)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.ReservedWord)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.SubfieldLen)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.TimesRead)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.MSGIDcrc)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.REPLYcrc)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.ReplyTo)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Reply1st)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.ReplyNext)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.DateWritten)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.DateReceived)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.DateProcessed)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.MessageNumber)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Attribute)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Attribute2)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Offset)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.TxtLen)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.PasswordCRC)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Cost)

	// Write subfields
	for _, subfield := range hdr.Subfields {
		binary.Write(j.jhrFile, binary.LittleEndian, subfield.LoID)
		binary.Write(j.jhrFile, binary.LittleEndian, subfield.HiID)
		binary.Write(j.jhrFile, binary.LittleEndian, subfield.DatLen)
		j.jhrFile.Write(subfield.Buffer)
	}

	return uint32(pos), nil
}

// ReadMessageText reads message text from the specified header
func (j *JAMBase) ReadMessageText(hdr *MessageHeader) (string, error) {
	if !j.isOpen {
		return "", ErrBaseNotOpen
	}

	if hdr.TxtLen == 0 {
		return "", nil
	}

	j.jdtFile.Seek(int64(hdr.Offset), 0)

	buf := make([]byte, hdr.TxtLen)
	_, err := j.jdtFile.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read message text: %w", err)
	}

	// Convert CR to LF for display
	text := strings.ReplaceAll(string(buf), "\r", "\n")
	return text, nil
}

// WriteMessageText writes message text and returns offset and length
func (j *JAMBase) WriteMessageText(text string) (uint32, uint32, error) {
	if !j.isOpen {
		return 0, 0, ErrBaseNotOpen
	}

	// Convert LF to CR (JAM uses CR as line separator)
	text = strings.ReplaceAll(text, "\n", "\r")

	// Get current position (end of file)
	pos, err := j.jdtFile.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to seek to end of text file: %w", err)
	}

	// Write text
	buf := []byte(text)
	_, err = j.jdtFile.Write(buf)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to write message text: %w", err)
	}

	return uint32(pos), uint32(len(buf)), nil
}

// ReadMessage reads a complete message
func (j *JAMBase) ReadMessage(msgNum int) (*Message, error) {
	if !j.isOpen {
		return nil, ErrBaseNotOpen
	}

	hdr, err := j.ReadMessageHeader(msgNum)
	if err != nil {
		return nil, err
	}

	text, err := j.ReadMessageText(hdr)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Header:   hdr,
		Text:     text,
		DateTime: time.Unix(int64(hdr.DateWritten), 0),
	}

	// Parse subfields
	for _, sf := range hdr.Subfields {
		value := string(sf.Buffer)

		switch sf.LoID {
		case JAMSFLD_OADDRESS:
			msg.OrigAddr = value
		case JAMSFLD_DADDRESS:
			msg.DestAddr = value
		case JAMSFLD_SENDERNAME:
			msg.From = value
		case JAMSFLD_RECEIVERNAME:
			msg.To = value
		case JAMSFLD_MSGID:
			msg.MsgID = value
		case JAMSFLD_REPLYID:
			msg.ReplyID = value
		case JAMSFLD_SUBJECT:
			msg.Subject = value
		case JAMSFLD_PID:
			msg.PID = value
		case JAMSFLD_FTSKLUDGE:
			msg.Kludges = append(msg.Kludges, value)
		case JAMSFLD_SEENBY2D:
			msg.SeenBy = value
		case JAMSFLD_PATH2D:
			msg.Path = value
		case JAMSFLD_FLAGS:
			msg.Flags = value
		case JAMSFLD_TZUTCINFO:
			msg.TZUTCInfo = value
		}
	}

	return msg, nil
}

// WriteMessage writes a complete message
func (j *JAMBase) WriteMessage(msg *Message) (int, error) {
	if !j.isOpen {
		return 0, ErrBaseNotOpen
	}

	// Reload fixed header to get current state
	err := j.readFixedHeader()
	if err != nil {
		return 0, err
	}

	// Create header
	hdr := &MessageHeader{
		Revision:      1,
		TimesRead:     0,
		ReplyTo:       0,
		Reply1st:      0,
		ReplyNext:     0,
		DateWritten:   uint32(msg.DateTime.Unix()),
		DateReceived:  0,
		DateProcessed: uint32(time.Now().Unix()),
		Attribute:     msg.GetAttribute(),
		Attribute2:    0,
		PasswordCRC:   0,
		Cost:          0,
	}
	copy(hdr.Signature[:], JAMSignature)

	// Add subfields
	hdr.Subfields = []Subfield{}

	if msg.OrigAddr != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_OADDRESS, msg.OrigAddr))
	}
	if msg.DestAddr != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_DADDRESS, msg.DestAddr))
	}
	if msg.From != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_SENDERNAME, msg.From))
	}
	if msg.To != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_RECEIVERNAME, msg.To))
	}
	if msg.Subject != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_SUBJECT, msg.Subject))
	}
	if msg.MsgID != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_MSGID, msg.MsgID))
		hdr.MSGIDcrc = CRC32String(msg.MsgID)
	}
	if msg.ReplyID != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_REPLYID, msg.ReplyID))
		hdr.REPLYcrc = CRC32String(msg.ReplyID)
	}
	if msg.PID != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_PID, msg.PID))
	}

	for _, kludge := range msg.Kludges {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_FTSKLUDGE, kludge))
	}

	// Calculate subfield length
	hdr.SubfieldLen = 0
	for _, sf := range hdr.Subfields {
		hdr.SubfieldLen += SubfieldHdrSize + sf.DatLen
	}

	// Write message text
	offset, txtLen, err := j.WriteMessageText(msg.Text)
	if err != nil {
		return 0, err
	}
	hdr.Offset = offset
	hdr.TxtLen = txtLen

	// Get next message number
	count, err := j.GetMessageCount()
	if err != nil {
		return 0, err
	}
	msgNum := count + 1
	hdr.MessageNumber = uint32(msgNum) + j.fixedHeader.BaseMsgNum - 1

	// Write header
	hdrOffset, err := j.WriteMessageHeader(hdr)
	if err != nil {
		return 0, err
	}

	// Write index
	toCRC := CRC32String(strings.ToLower(msg.To))
	idx := &IndexRecord{
		ToCRC:     toCRC,
		HdrOffset: hdrOffset,
	}
	err = j.WriteIndexRecord(msgNum, idx)
	if err != nil {
		return 0, err
	}

	// Update fixed header
	j.fixedHeader.ActiveMsgs++
	j.fixedHeader.ModCounter++
	err = j.writeFixedHeader()
	if err != nil {
		return 0, err
	}

	return msgNum, nil
}

// DeleteMessage marks a message as deleted
func (j *JAMBase) DeleteMessage(msgNum int) error {
	if !j.isOpen {
		return ErrBaseNotOpen
	}

	hdr, err := j.ReadMessageHeader(msgNum)
	if err != nil {
		return err
	}

	// Mark as deleted
	hdr.Attribute |= MSG_DELETED
	hdr.TxtLen = 0

	// Update header
	idx, err := j.ReadIndexRecord(msgNum)
	if err != nil {
		return err
	}

	j.jhrFile.Seek(int64(idx.HdrOffset), 0)

	// Write fixed part only (to update attributes)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Signature)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Revision)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.ReservedWord)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.SubfieldLen)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.TimesRead)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.MSGIDcrc)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.REPLYcrc)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.ReplyTo)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Reply1st)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.ReplyNext)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.DateWritten)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.DateReceived)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.DateProcessed)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.MessageNumber)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Attribute)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Attribute2)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Offset)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.TxtLen)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.PasswordCRC)
	binary.Write(j.jhrFile, binary.LittleEndian, hdr.Cost)

	// Update fixed header
	j.fixedHeader.ActiveMsgs--
	j.fixedHeader.ModCounter++
	return j.writeFixedHeader()
}
