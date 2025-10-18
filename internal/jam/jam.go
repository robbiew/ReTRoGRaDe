// Package jam implements the JAM message base format for Retrograde BBS.
// This package follows the JAM(mbp) specification as documented in JAM-001.
package jam

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
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

// Message attributes from JAM specification
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

// Subfield types from JAM specification
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
	ErrBaseNotOpen      = errors.New("message base not open")
)

// JAMBase represents a JAM message base
type JAMBase struct {
	BasePath    string
	fixedHeader *FixedHeaderInfo
	jhrFile     *os.File
	jdtFile     *os.File
	jdxFile     *os.File
	jlrFile     *os.File
	isOpen      bool
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
	ToCRC     uint32
	HdrOffset uint32
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
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(basePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create message base directory: %w", err)
	}

	base := &JAMBase{
		BasePath: basePath,
	}

	// Try to open existing base
	jhrPath := basePath + ".jhr"
	jdtPath := basePath + ".jdt"
	jdxPath := basePath + ".jdx"
	jlrPath := basePath + ".jlr"

	// Check if base exists and is valid
	stat, err := os.Stat(jhrPath)
	if os.IsNotExist(err) {
		// Create new base
		return base, base.Create()
	}

	// Check if header file is too small (corrupted)
	if stat.Size() < 1024 {
		// Header file is corrupted, recreate the entire base
		os.Remove(jhrPath)
		os.Remove(jdtPath)
		os.Remove(jdxPath)
		os.Remove(jlrPath)
		return base, base.Create()
	}

	// Open existing files
	base.jhrFile, err = os.OpenFile(jhrPath, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open .jhr file: %w", err)
	}

	base.jdtFile, err = os.OpenFile(jdtPath, os.O_RDWR, 0644)
	if err != nil {
		base.jhrFile.Close()
		return nil, fmt.Errorf("failed to open .jdt file: %w", err)
	}

	base.jdxFile, err = os.OpenFile(jdxPath, os.O_RDWR, 0644)
	if err != nil {
		base.jhrFile.Close()
		base.jdtFile.Close()
		return nil, fmt.Errorf("failed to open .jdx file: %w", err)
	}

	base.jlrFile, err = os.OpenFile(jlrPath, os.O_RDWR, 0644)
	if err != nil {
		base.jhrFile.Close()
		base.jdtFile.Close()
		base.jdxFile.Close()
		return nil, fmt.Errorf("failed to open .jlr file: %w", err)
	}

	// Set as open before reading header
	base.isOpen = true

	// Read fixed header
	err = base.readFixedHeader()
	if err != nil {
		base.Close()
		// If header is corrupted, try to recreate the base
		if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "invalid signature") {
			os.Remove(jhrPath)
			os.Remove(jdtPath)
			os.Remove(jdxPath)
			os.Remove(jlrPath)
			return base, base.Create()
		}
		return nil, fmt.Errorf("failed to read fixed header: %w", err)
	}

	return base, nil
}

// Create creates a new JAM message base
func (j *JAMBase) Create() error {
	var err error

	// Create files
	j.jhrFile, err = os.Create(j.BasePath + ".jhr")
	if err != nil {
		return fmt.Errorf("failed to create .jhr file: %w", err)
	}

	j.jdtFile, err = os.Create(j.BasePath + ".jdt")
	if err != nil {
		j.jhrFile.Close()
		return fmt.Errorf("failed to create .jdt file: %w", err)
	}

	j.jdxFile, err = os.Create(j.BasePath + ".jdx")
	if err != nil {
		j.jhrFile.Close()
		j.jdtFile.Close()
		return fmt.Errorf("failed to create .jdx file: %w", err)
	}

	j.jlrFile, err = os.Create(j.BasePath + ".jlr")
	if err != nil {
		j.jhrFile.Close()
		j.jdtFile.Close()
		j.jdxFile.Close()
		return fmt.Errorf("failed to create .jlr file: %w", err)
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

	// Set as open before writing header
	j.isOpen = true

	// Write fixed header
	err = j.writeFixedHeader()
	if err != nil {
		j.Close()
		return fmt.Errorf("failed to write fixed header: %w", err)
	}

	return nil
}

// Close closes the JAM message base
func (j *JAMBase) Close() error {
	var errs []error

	if j.jhrFile != nil {
		if err := j.jhrFile.Close(); err != nil {
			errs = append(errs, err)
		}
		j.jhrFile = nil
	}
	if j.jdtFile != nil {
		if err := j.jdtFile.Close(); err != nil {
			errs = append(errs, err)
		}
		j.jdtFile = nil
	}
	if j.jdxFile != nil {
		if err := j.jdxFile.Close(); err != nil {
			errs = append(errs, err)
		}
		j.jdxFile = nil
	}
	if j.jlrFile != nil {
		if err := j.jlrFile.Close(); err != nil {
			errs = append(errs, err)
		}
		j.jlrFile = nil
	}

	j.isOpen = false

	if len(errs) > 0 {
		return fmt.Errorf("errors closing JAM base: %v", errs)
	}
	return nil
}

// IsOpen returns whether the message base is currently open
func (j *JAMBase) IsOpen() bool {
	return j.isOpen
}

// readFixedHeader reads the fixed header from .jhr file
func (j *JAMBase) readFixedHeader() error {
	if !j.isOpen {
		return ErrBaseNotOpen
	}

	j.jhrFile.Seek(0, 0)
	j.fixedHeader = &FixedHeaderInfo{}

	err := binary.Read(j.jhrFile, binary.LittleEndian, j.fixedHeader)
	if err != nil {
		return fmt.Errorf("failed to read fixed header: %w", err)
	}

	if string(j.fixedHeader.Signature[:]) != JAMSignature {
		return ErrInvalidSignature
	}

	return nil
}

// writeFixedHeader writes the fixed header to .jhr file
func (j *JAMBase) writeFixedHeader() error {
	if !j.isOpen {
		return ErrBaseNotOpen
	}

	j.jhrFile.Seek(0, 0)
	return binary.Write(j.jhrFile, binary.LittleEndian, j.fixedHeader)
}

// GetMessageCount returns the number of messages in the base
func (j *JAMBase) GetMessageCount() (int, error) {
	if !j.isOpen {
		return 0, ErrBaseNotOpen
	}

	info, err := j.jdxFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat index file: %w", err)
	}

	count := info.Size() / IndexRecordSize
	return int(count), nil
}

// GetActiveMessageCount returns the number of active (non-deleted) messages
func (j *JAMBase) GetActiveMessageCount() int {
	if j.fixedHeader == nil {
		return 0
	}
	return int(j.fixedHeader.ActiveMsgs)
}

// ReadIndexRecord reads an index record
func (j *JAMBase) ReadIndexRecord(msgNum int) (*IndexRecord, error) {
	if !j.isOpen {
		return nil, ErrBaseNotOpen
	}

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
		ToCRC:     toCRC,
		HdrOffset: hdrOffset,
	}, nil
}

// WriteIndexRecord writes an index record
func (j *JAMBase) WriteIndexRecord(msgNum int, rec *IndexRecord) error {
	if !j.isOpen {
		return ErrBaseNotOpen
	}

	offset := int64((msgNum - 1) * IndexRecordSize)
	j.jdxFile.Seek(offset, 0)

	binary.Write(j.jdxFile, binary.LittleEndian, rec.ToCRC)
	binary.Write(j.jdxFile, binary.LittleEndian, rec.HdrOffset)

	return nil
}

// NewMessage creates a new Message with default values for local messages
func NewMessage() *Message {
	return &Message{
		DateTime: time.Now(),
	}
}

// IsLocal checks if message is local
func (m *Message) IsLocal() bool {
	if m.Header == nil {
		return true // Default for new messages
	}
	return m.Header.Attribute&MSG_TYPELOCAL != 0
}

// IsDeleted checks if message is deleted
func (m *Message) IsDeleted() bool {
	if m.Header == nil {
		return false
	}
	return m.Header.Attribute&MSG_DELETED != 0
}

// IsPrivate checks if message is private
func (m *Message) IsPrivate() bool {
	if m.Header == nil {
		return false
	}
	return m.Header.Attribute&MSG_PRIVATE != 0
}

// GetAttribute returns the message attribute based on message type
func (m *Message) GetAttribute() uint32 {
	if m.Header != nil {
		return m.Header.Attribute
	}

	// Default attribute for local messages
	return MSG_LOCAL | MSG_TYPELOCAL
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
	// Convert to lowercase (A-Z only) according to JAM spec
	lower := strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return r
	}, s)

	// Calculate CRC32 and invert per JAM spec
	table := crc32.MakeTable(crc32.IEEE)
	crc := crc32.Checksum([]byte(lower), table)
	return ^crc
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
