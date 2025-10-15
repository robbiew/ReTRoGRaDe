// jam.go - Main JAM package file
package jam

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strings"
	"time"
)

const (
	JAMSignature    = "JAM\x00"
	HeaderSize      = 1024
	FixedHeaderSize = 76
	SubfieldHdrSize = 8
	IndexRecordSize = 8
	LastReadSize    = 16
)

// Message attributes
const (
	MSG_LOCAL       = 0x00000001 // Msg created locally
	MSG_INTRANSIT   = 0x00000002 // Msg is in-transit
	MSG_PRIVATE     = 0x00000004 // Private
	MSG_READ        = 0x00000008 // Read by addressee
	MSG_SENT        = 0x00000010 // Sent to remote
	MSG_KILLSENT    = 0x00000020 // Kill when sent
	MSG_ARCHIVESENT = 0x00000040 // Archive when sent
	MSG_HOLD        = 0x00000080 // Hold for pick-up
	MSG_CRASH       = 0x00000100 // Crash
	MSG_IMMEDIATE   = 0x00000200 // Send now
	MSG_DIRECT      = 0x00000400 // Send directly
	MSG_GATE        = 0x00000800 // Send via gateway
	MSG_FILEREQUEST = 0x00001000 // File request
	MSG_FILEATTACH  = 0x00002000 // File(s) attached
	MSG_TRUNCFILE   = 0x00004000 // Truncate file(s)
	MSG_KILLFILE    = 0x00008000 // Delete file(s)
	MSG_RECEIPTREQ  = 0x00010000 // Return receipt
	MSG_CONFIRMREQ  = 0x00020000 // Confirmation receipt
	MSG_ORPHAN      = 0x00040000 // Unknown destination
	MSG_ENCRYPT     = 0x00080000 // Encrypted
	MSG_COMPRESS    = 0x00100000 // Compressed
	MSG_ESCAPED     = 0x00200000 // Seven bit ASCII
	MSG_FPU         = 0x00400000 // Force pickup
	MSG_TYPELOCAL   = 0x00800000 // Local use only
	MSG_TYPEECHO    = 0x01000000 // Conference/echo
	MSG_TYPENET     = 0x02000000 // Direct network mail
	MSG_NODISP      = 0x20000000 // May not be displayed
	MSG_LOCKED      = 0x40000000 // Locked
	MSG_DELETED     = 0x80000000 // Deleted
)

// Subfield types
const (
	JAMSFLD_OADDRESS             = 0
	JAMSFLD_DADDRESS             = 1
	JAMSFLD_SENDERNAME           = 2
	JAMSFLD_RECEIVERNAME         = 3
	JAMSFLD_MSGID                = 4
	JAMSFLD_REPLYID              = 5
	JAMSFLD_SUBJECT              = 6
	JAMSFLD_PID                  = 7
	JAMSFLD_TRACE                = 8
	JAMSFLD_ENCLOSEDFILE         = 9
	JAMSFLD_ENCLOSEDFILEWALIAS   = 10
	JAMSFLD_ENCLOSEDFREQ         = 11
	JAMSFLD_ENCLOSEDFILEWCARD    = 12
	JAMSFLD_ENCLOSEDINDIRECTFILE = 13
	JAMSFLD_EMBINDAT             = 1000
	JAMSFLD_FTSKLUDGE            = 2000
	JAMSFLD_SEENBY2D             = 2001
	JAMSFLD_PATH2D               = 2002
	JAMSFLD_FLAGS                = 2003
	JAMSFLD_TZUTCINFO            = 2004
)

var (
	ErrInvalidSignature = errors.New("invalid JAM signature")
	ErrInvalidMessage   = errors.New("invalid message number")
	ErrBaseLocked       = errors.New("message base is locked")
	ErrNotFound         = errors.New("not found")
)

// JAMBase represents a JAM message base
type JAMBase struct {
	BasePath    string
	fixedHeader *FixedHeaderInfo
	jhrFile     *os.File
	jdtFile     *os.File
	jdxFile     *os.File
	jlrFile     *os.File
}

// FixedHeaderInfo - JAM base header (1024 bytes)
type FixedHeaderInfo struct {
	Signature   [4]byte
	DateCreated uint32
	ModCounter  uint32
	ActiveMsgs  uint32
	PasswordCRC uint32
	BaseMsgNum  uint32
	Reserved    [1000]byte
}

// MessageHeader - JAM message header
type MessageHeader struct {
	Signature     [4]byte
	Revision      uint16
	ReservedWord  uint16
	SubfieldLen   uint32
	TimesRead     uint32
	MSGIDcrc      uint32
	REPLYcrc      uint32
	ReplyTo       uint32
	Reply1st      uint32
	ReplyNext     uint32
	DateWritten   uint32
	DateReceived  uint32
	DateProcessed uint32
	MessageNumber uint32
	Attribute     uint32
	Attribute2    uint32
	Offset        uint32
	TxtLen        uint32
	PasswordCRC   uint32
	Cost          uint32
	Subfields     []Subfield
}

// Subfield - message subfield
type Subfield struct {
	LoID   uint16
	HiID   uint16
	DatLen uint32
	Buffer []byte
}

// IndexRecord - JAM index record
type IndexRecord struct {
	ToCRC      uint32
	HdrOffset  uint32
	MessageNum uint32
}

// LastRead - lastread record
type LastRead struct {
	UserCRC     uint32
	UserID      uint32
	LastReadMsg uint32
	HighReadMsg uint32
}

// Message - high-level message structure
type Message struct {
	Header    *MessageHeader
	From      string
	To        string
	Subject   string
	DateTime  time.Time
	Text      string
	OrigAddr  string
	DestAddr  string
	MsgID     string
	ReplyID   string
	PID       string
	Flags     string
	SeenBy    string
	Path      string
	TZUTCInfo string
	Kludges   []string
}

// Open opens or creates a JAM message base
func Open(basePath string) (*JAMBase, error) {
	base := &JAMBase{
		BasePath: basePath,
	}

	// Try to open existing base
	jhrPath := basePath + ".jhr"
	jdtPath := basePath + ".jdt"
	jdxPath := basePath + ".jdx"
	jlrPath := basePath + ".jlr"

	// Check if base exists
	_, err := os.Stat(jhrPath)
	if os.IsNotExist(err) {
		// Create new base
		return base, base.Create()
	}

	// Open existing files
	base.jhrFile, err = os.OpenFile(jhrPath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	base.jdtFile, err = os.OpenFile(jdtPath, os.O_RDWR, 0644)
	if err != nil {
		base.jhrFile.Close()
		return nil, err
	}

	base.jdxFile, err = os.OpenFile(jdxPath, os.O_RDWR, 0644)
	if err != nil {
		base.jhrFile.Close()
		base.jdtFile.Close()
		return nil, err
	}

	base.jlrFile, err = os.OpenFile(jlrPath, os.O_RDWR, 0644)
	if err != nil {
		base.jhrFile.Close()
		base.jdtFile.Close()
		base.jdxFile.Close()
		return nil, err
	}

	// Read fixed header
	err = base.readFixedHeader()
	if err != nil {
		base.Close()
		return nil, err
	}

	return base, nil
}

// Create creates a new JAM message base
func (j *JAMBase) Create() error {
	var err error

	// Create files
	j.jhrFile, err = os.Create(j.BasePath + ".jhr")
	if err != nil {
		return err
	}

	j.jdtFile, err = os.Create(j.BasePath + ".jdt")
	if err != nil {
		j.jhrFile.Close()
		return err
	}

	j.jdxFile, err = os.Create(j.BasePath + ".jdx")
	if err != nil {
		j.jhrFile.Close()
		j.jdtFile.Close()
		return err
	}

	j.jlrFile, err = os.Create(j.BasePath + ".jlr")
	if err != nil {
		j.jhrFile.Close()
		j.jdtFile.Close()
		j.jdxFile.Close()
		return err
	}

	// Initialize fixed header
	j.fixedHeader = &FixedHeaderInfo{
		DateCreated: uint32(time.Now().Unix()),
		ModCounter:  0,
		ActiveMsgs:  0,
		PasswordCRC: 0,
		BaseMsgNum:  1,
	}
	copy(j.fixedHeader.Signature[:], JAMSignature)

	// Write fixed header
	return j.writeFixedHeader()
}

// Close closes the JAM message base
func (j *JAMBase) Close() error {
	var errs []error

	if j.jhrFile != nil {
		if err := j.jhrFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if j.jdtFile != nil {
		if err := j.jdtFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if j.jdxFile != nil {
		if err := j.jdxFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if j.jlrFile != nil {
		if err := j.jlrFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing JAM base: %v", errs)
	}
	return nil
}

// readFixedHeader reads the fixed header from .jhr file
func (j *JAMBase) readFixedHeader() error {
	j.jhrFile.Seek(0, 0)
	j.fixedHeader = &FixedHeaderInfo{}

	err := binary.Read(j.jhrFile, binary.LittleEndian, j.fixedHeader)
	if err != nil {
		return err
	}

	if string(j.fixedHeader.Signature[:]) != JAMSignature {
		return ErrInvalidSignature
	}

	return nil
}

// writeFixedHeader writes the fixed header to .jhr file
func (j *JAMBase) writeFixedHeader() error {
	j.jhrFile.Seek(0, 0)
	return binary.Write(j.jhrFile, binary.LittleEndian, j.fixedHeader)
}

// Lock locks the message base for writing
func (j *JAMBase) Lock() error {
	// In a real implementation, use file locking
	// For now, we'll use a simple approach
	return nil
}

// Unlock unlocks the message base
func (j *JAMBase) Unlock() error {
	return nil
}

// GetMessageCount returns the number of messages in the base
func (j *JAMBase) GetMessageCount() (int, error) {
	info, err := j.jdxFile.Stat()
	if err != nil {
		return 0, err
	}

	count := info.Size() / IndexRecordSize
	return int(count), nil
}

// GetActiveMessageCount returns the number of active (non-deleted) messages
func (j *JAMBase) GetActiveMessageCount() int {
	return int(j.fixedHeader.ActiveMsgs)
}

// ReadIndexRecord reads an index record
func (j *JAMBase) ReadIndexRecord(msgNum int) (*IndexRecord, error) {
	count, err := j.GetMessageCount()
	if err != nil {
		return nil, err
	}

	if msgNum < 1 || msgNum > count {
		return nil, ErrInvalidMessage
	}

	offset := int64((msgNum - 1) * IndexRecordSize)
	j.jdxFile.Seek(offset, 0)

	var toCRC, hdrOffset uint32
	binary.Read(j.jdxFile, binary.LittleEndian, &toCRC)
	binary.Read(j.jdxFile, binary.LittleEndian, &hdrOffset)

	// Check for deleted message
	if toCRC == 0xFFFFFFFF && hdrOffset == 0xFFFFFFFF {
		return nil, ErrNotFound
	}

	return &IndexRecord{
		ToCRC:      toCRC,
		HdrOffset:  hdrOffset,
		MessageNum: uint32(msgNum) + j.fixedHeader.BaseMsgNum - 1,
	}, nil
}

// WriteIndexRecord writes an index record
func (j *JAMBase) WriteIndexRecord(msgNum int, rec *IndexRecord) error {
	offset := int64((msgNum - 1) * IndexRecordSize)
	j.jdxFile.Seek(offset, 0)

	binary.Write(j.jdxFile, binary.LittleEndian, rec.ToCRC)
	binary.Write(j.jdxFile, binary.LittleEndian, rec.HdrOffset)

	return nil
}

// ReadMessageHeader reads a message header
func (j *JAMBase) ReadMessageHeader(msgNum int) (*MessageHeader, error) {
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

// WriteMessageHeader writes a message header
func (j *JAMBase) WriteMessageHeader(hdr *MessageHeader) (uint32, error) {
	// Get current position (end of file)
	pos, err := j.jhrFile.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
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

// ReadMessageText reads message text
func (j *JAMBase) ReadMessageText(hdr *MessageHeader) (string, error) {
	if hdr.TxtLen == 0 {
		return "", nil
	}

	j.jdtFile.Seek(int64(hdr.Offset), 0)

	buf := make([]byte, hdr.TxtLen)
	_, err := j.jdtFile.Read(buf)
	if err != nil {
		return "", err
	}

	// Convert CR to LF
	text := strings.ReplaceAll(string(buf), "\r", "\n")
	return text, nil
}

// WriteMessageText writes message text
func (j *JAMBase) WriteMessageText(text string) (uint32, uint32, error) {
	// Convert LF to CR (JAM uses CR as line separator)
	text = strings.ReplaceAll(text, "\n", "\r")

	// Get current position (end of file)
	pos, err := j.jdtFile.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, 0, err
	}

	// Write text
	buf := []byte(text)
	_, err = j.jdtFile.Write(buf)
	if err != nil {
		return 0, 0, err
	}

	return uint32(pos), uint32(len(buf)), nil
}

// ReadMessage reads a complete message
func (j *JAMBase) ReadMessage(msgNum int) (*Message, error) {
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
	err := j.Lock()
	if err != nil {
		return 0, err
	}
	defer j.Unlock()

	// Reload fixed header
	err = j.readFixedHeader()
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
		DateWritten:   uint32(time.Now().Unix()),
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
	if msg.SeenBy != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_SEENBY2D, msg.SeenBy))
	}
	if msg.Path != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_PATH2D, msg.Path))
	}
	if msg.Flags != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_FLAGS, msg.Flags))
	}
	if msg.TZUTCInfo != "" {
		hdr.Subfields = append(hdr.Subfields, CreateSubfield(JAMSFLD_TZUTCINFO, msg.TZUTCInfo))
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
		ToCRC:      toCRC,
		HdrOffset:  hdrOffset,
		MessageNum: hdr.MessageNumber,
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
	err := j.Lock()
	if err != nil {
		return err
	}
	defer j.Unlock()

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

// GetLastRead gets the last read message for a user
func (j *JAMBase) GetLastRead(username string) (*LastRead, error) {
	userCRC := CRC32String(strings.ToLower(username))

	info, err := j.jlrFile.Stat()
	if err != nil {
		return nil, err
	}

	recordCount := info.Size() / LastReadSize

	j.jlrFile.Seek(0, 0)

	for i := int64(0); i < recordCount; i++ {
		lr := &LastRead{}
		binary.Read(j.jlrFile, binary.LittleEndian, &lr.UserCRC)
		binary.Read(j.jlrFile, binary.LittleEndian, &lr.UserID)
		binary.Read(j.jlrFile, binary.LittleEndian, &lr.LastReadMsg)
		binary.Read(j.jlrFile, binary.LittleEndian, &lr.HighReadMsg)

		if lr.UserCRC == userCRC {
			return lr, nil
		}
	}

	return nil, ErrNotFound
}

// SetLastRead sets the last read message for a user
func (j *JAMBase) SetLastRead(username string, lastRead, highRead uint32) error {
	userCRC := CRC32String(strings.ToLower(username))

	info, err := j.jlrFile.Stat()
	if err != nil {
		return err
	}

	recordCount := info.Size() / LastReadSize

	j.jlrFile.Seek(0, 0)

	// Try to find existing record
	for i := int64(0); i < recordCount; i++ {
		pos := i * LastReadSize
		j.jlrFile.Seek(pos, 0)

		var readUserCRC uint32
		binary.Read(j.jlrFile, binary.LittleEndian, &readUserCRC)

		if readUserCRC == userCRC {
			// Update existing record
			j.jlrFile.Seek(pos, 0)
			binary.Write(j.jlrFile, binary.LittleEndian, userCRC)
			binary.Write(j.jlrFile, binary.LittleEndian, userCRC) // UserID = UserCRC
			binary.Write(j.jlrFile, binary.LittleEndian, lastRead)
			binary.Write(j.jlrFile, binary.LittleEndian, highRead)
			return nil
		}
	}

	// Add new record
	j.jlrFile.Seek(0, io.SeekEnd)
	binary.Write(j.jlrFile, binary.LittleEndian, userCRC)
	binary.Write(j.jlrFile, binary.LittleEndian, userCRC) // UserID = UserCRC
	binary.Write(j.jlrFile, binary.LittleEndian, lastRead)
	binary.Write(j.jlrFile, binary.LittleEndian, highRead)

	return nil
}

// CreateSubfield creates a subfield
func CreateSubfield(fieldType uint16, data string) Subfield {
	buf := []byte(data)
	return Subfield{
		LoID:   fieldType,
		HiID:   0,
		DatLen: uint32(len(buf)),
		Buffer: buf,
	}
}

// CRC32String calculates JAM-style CRC32 of a string
func CRC32String(s string) uint32 {
	// Convert to lowercase (A-Z only)
	lower := strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return r
	}, s)

	// Calculate CRC32 and invert
	table := crc32.MakeTable(crc32.IEEE)
	crc := crc32.Checksum([]byte(lower), table)
	return ^crc
}

// GetAttribute returns the message attribute based on message type
func (m *Message) GetAttribute() uint32 {
	attr := uint32(MSG_LOCAL)

	if m.Header != nil {
		return m.Header.Attribute
	}

	// Determine message type from addresses
	if m.DestAddr == "" {
		// Local message or echomail
		attr |= MSG_TYPELOCAL
	} else if strings.Contains(m.Subject, "file attach") {
		attr |= MSG_FILEATTACH
	}

	return attr
}

// IsPrivate checks if message is private
func (m *Message) IsPrivate() bool {
	return m.Header.Attribute&MSG_PRIVATE != 0
}

// IsDeleted checks if message is deleted
func (m *Message) IsDeleted() bool {
	return m.Header.Attribute&MSG_DELETED != 0
}

// IsLocal checks if message is local
func (m *Message) IsLocal() bool {
	return m.Header.Attribute&MSG_TYPELOCAL != 0
}

// IsEcho checks if message is echomail
func (m *Message) IsEcho() bool {
	return m.Header.Attribute&MSG_TYPEECHO != 0
}

// IsNetmail checks if message is netmail
func (m *Message) IsNetmail() bool {
	return m.Header.Attribute&MSG_TYPENET != 0
}

// GetSubfieldByType gets the first subfield of a specific type
func (h *MessageHeader) GetSubfieldByType(fieldType uint16) *Subfield {
	for i := range h.Subfields {
		if h.Subfields[i].LoID == fieldType {
			return &h.Subfields[i]
		}
	}
	return nil
}

// GetAllSubfieldsByType gets all subfields of a specific type
func (h *MessageHeader) GetAllSubfieldsByType(fieldType uint16) []Subfield {
	var fields []Subfield
	for _, sf := range h.Subfields {
		if sf.LoID == fieldType {
			fields = append(fields, sf)
		}
	}
	return fields
}
