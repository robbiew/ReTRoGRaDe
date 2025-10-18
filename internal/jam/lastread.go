package jam

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// GetLastRead gets the last read message for a user
func (j *JAMBase) GetLastRead(username string) (*LastRead, error) {
	if !j.isOpen {
		return nil, ErrBaseNotOpen
	}

	userCRC := CRC32String(strings.ToLower(username))

	info, err := j.jlrFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat lastread file: %w", err)
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
	if !j.isOpen {
		return ErrBaseNotOpen
	}

	userCRC := CRC32String(strings.ToLower(username))

	info, err := j.jlrFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat lastread file: %w", err)
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

// GetNextUnreadMessage returns the next unread message number for a user
func (j *JAMBase) GetNextUnreadMessage(username string) (int, error) {
	if !j.isOpen {
		return 0, ErrBaseNotOpen
	}

	lr, err := j.GetLastRead(username)
	if err != nil {
		if err == ErrNotFound {
			// User has never read any messages, start from message 1
			count, err := j.GetMessageCount()
			if err != nil {
				return 0, err
			}
			if count > 0 {
				return 1, nil
			}
			return 0, ErrNotFound
		}
		return 0, err
	}

	// Return the next message after the last read
	nextMsg := int(lr.LastReadMsg) + 1
	count, err := j.GetMessageCount()
	if err != nil {
		return 0, err
	}

	if nextMsg <= count {
		return nextMsg, nil
	}

	return 0, ErrNotFound
}

// MarkMessageRead marks a message as read by a user
func (j *JAMBase) MarkMessageRead(username string, msgNum int) error {
	if !j.isOpen {
		return ErrBaseNotOpen
	}

	lr, err := j.GetLastRead(username)
	if err != nil {
		if err == ErrNotFound {
			// Create new lastread record
			return j.SetLastRead(username, uint32(msgNum), uint32(msgNum))
		}
		return err
	}

	// Update lastread pointer
	newLastRead := uint32(msgNum)
	newHighRead := lr.HighReadMsg
	if newLastRead > lr.HighReadMsg {
		newHighRead = newLastRead
	}

	return j.SetLastRead(username, newLastRead, newHighRead)
}

// GetUnreadCount returns the number of unread messages for a user
func (j *JAMBase) GetUnreadCount(username string) (int, error) {
	if !j.isOpen {
		return 0, ErrBaseNotOpen
	}

	count, err := j.GetMessageCount()
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	lr, err := j.GetLastRead(username)
	if err != nil {
		if err == ErrNotFound {
			// User has never read any messages, all are unread
			return count, nil
		}
		return 0, err
	}

	unread := count - int(lr.LastReadMsg)
	if unread < 0 {
		unread = 0
	}

	return unread, nil
}

// ScanMessages scans messages in the base for display/reading
func (j *JAMBase) ScanMessages(startMsg int, maxMessages int) ([]*Message, error) {
	if !j.isOpen {
		return nil, ErrBaseNotOpen
	}

	count, err := j.GetMessageCount()
	if err != nil {
		return nil, err
	}

	if startMsg < 1 {
		startMsg = 1
	}

	var messages []*Message
	messagesRead := 0

	for msgNum := startMsg; msgNum <= count && (maxMessages == 0 || messagesRead < maxMessages); msgNum++ {
		msg, err := j.ReadMessage(msgNum)
		if err != nil {
			// Skip messages that can't be read (possibly deleted)
			continue
		}

		if !msg.IsDeleted() {
			messages = append(messages, msg)
			messagesRead++
		}
	}

	return messages, nil
}
