// maintenance.go - Message base maintenance utilities
package jam

import (
	"fmt"
	"io"
	"os"
)

// Maintenance provides message base maintenance functions
type Maintenance struct {
	base *JAMBase
}

// NewMaintenance creates a maintenance helper
func NewMaintenance(base *JAMBase) *Maintenance {
	return &Maintenance{base: base}
}

// Pack removes deleted messages and compacts the base
func (m *Maintenance) Pack() error {
	err := m.base.Lock()
	if err != nil {
		return err
	}
	defer m.base.Unlock()

	// Create temporary files
	tmpBase := m.base.BasePath + ".tmp"
	tmpJAM, err := Open(tmpBase)
	if err != nil {
		return err
	}
	defer tmpJAM.Close()

	// Copy non-deleted messages
	count, err := m.base.GetMessageCount()
	if err != nil {
		return err
	}

	packed := 0
	for i := 1; i <= count; i++ {
		msg, err := m.base.ReadMessage(i)
		if err != nil {
			continue
		}

		if !msg.IsDeleted() {
			_, err = tmpJAM.WriteMessage(msg)
			if err != nil {
				return err
			}
			packed++
		}
	}

	// Close both bases
	m.base.Close()
	tmpJAM.Close()

	// Replace original files with packed versions
	for _, ext := range []string{".jhr", ".jdt", ".jdx"} {
		err = os.Rename(tmpBase+ext, m.base.BasePath+ext)
		if err != nil {
			return err
		}
	}

	// Reopen the base
	m.base, err = Open(m.base.BasePath)
	if err != nil {
		return err
	}

	fmt.Printf("Packed %d messages\n", packed)
	return nil
}

// Renumber renumbers the message base starting from 1
func (m *Maintenance) Renumber() error {
	err := m.base.Lock()
	if err != nil {
		return err
	}
	defer m.base.Unlock()

	// Read fixed header
	err = m.base.readFixedHeader()
	if err != nil {
		return err
	}

	// Set base message number to 1
	m.base.fixedHeader.BaseMsgNum = 1

	// Write updated header
	return m.base.writeFixedHeader()
}

// GetStatistics returns base statistics
func (m *Maintenance) GetStatistics() (*BaseStats, error) {
	stats := &BaseStats{}

	count, err := m.base.GetMessageCount()
	if err != nil {
		return nil, err
	}

	stats.TotalMessages = count
	stats.ActiveMessages = m.base.GetActiveMessageCount()
	stats.DeletedMessages = stats.TotalMessages - stats.ActiveMessages

	// Get file sizes
	jhrInfo, _ := m.base.jhrFile.Stat()
	jdtInfo, _ := m.base.jdtFile.Stat()
	jdxInfo, _ := m.base.jdxFile.Stat()
	jlrInfo, _ := m.base.jlrFile.Stat()

	stats.HeaderSize = jhrInfo.Size()
	stats.TextSize = jdtInfo.Size()
	stats.IndexSize = jdxInfo.Size()
	stats.LastReadSize = jlrInfo.Size()
	stats.TotalSize = stats.HeaderSize + stats.TextSize + stats.IndexSize + stats.LastReadSize

	// Scan messages for additional stats
	for i := 1; i <= count; i++ {
		msg, err := m.base.ReadMessage(i)
		if err != nil {
			continue
		}

		if msg.IsPrivate() {
			stats.PrivateMessages++
		}
		if msg.IsEcho() {
			stats.EchoMessages++
		}
		if msg.IsNetmail() {
			stats.NetmailMessages++
		}
		if msg.IsLocal() {
			stats.LocalMessages++
		}
	}

	return stats, nil
}

// BaseStats contains message base statistics
type BaseStats struct {
	TotalMessages   int
	ActiveMessages  int
	DeletedMessages int
	PrivateMessages int
	EchoMessages    int
	NetmailMessages int
	LocalMessages   int
	HeaderSize      int64
	TextSize        int64
	IndexSize       int64
	LastReadSize    int64
	TotalSize       int64
}

// String returns a formatted string of statistics
func (bs *BaseStats) String() string {
	return fmt.Sprintf(`Message Base Statistics
======================
Total Messages:    %d
Active Messages:   %d
Deleted Messages:  %d
Private Messages:  %d
Echo Messages:     %d
Netmail Messages:  %d
Local Messages:    %d

File Sizes:
Header File:       %d bytes
Text File:         %d bytes
Index File:        %d bytes
LastRead File:     %d bytes
Total Size:        %d bytes`,
		bs.TotalMessages,
		bs.ActiveMessages,
		bs.DeletedMessages,
		bs.PrivateMessages,
		bs.EchoMessages,
		bs.NetmailMessages,
		bs.LocalMessages,
		bs.HeaderSize,
		bs.TextSize,
		bs.IndexSize,
		bs.LastReadSize,
		bs.TotalSize)
}

// Backup creates a backup of the message base
func (m *Maintenance) Backup(destPath string) error {
	for _, ext := range []string{".jhr", ".jdt", ".jdx", ".jlr"} {
		src := m.base.BasePath + ext
		dst := destPath + ext

		err := copyFile(src, dst)
		if err != nil {
			return fmt.Errorf("backup failed: %v", err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
