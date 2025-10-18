package config

import (
	"fmt"

	"github.com/robbiew/retrograde/internal/database"
)

// SetDefaultMessageArea sets the user's current message area to the first available area
func (session *TelnetSession) SetDefaultMessageArea(db database.Database) error {
	if db == nil {
		return fmt.Errorf("database not available")
	}

	// Get all message areas
	areas, err := db.GetAllMessageAreas()
	if err != nil {
		return fmt.Errorf("failed to get message areas: %w", err)
	}

	if len(areas) == 0 {
		// No message areas available, leave current area as nil
		session.CurrentMessageArea = nil
		return nil
	}

	// Set to the first available area (usually "General Chatter" in "Local Areas")
	session.CurrentMessageArea = &areas[0]
	return nil
}

// GetCurrentMessageAreaPath returns the file path for the current message area
func (session *TelnetSession) GetCurrentMessageAreaPath() string {
	if session.CurrentMessageArea == nil {
		return ""
	}

	// Construct the full path to the JAM base files
	// The path stored in the database is the directory, we need to append the filename
	path := session.CurrentMessageArea.Path
	if path == "" {
		return ""
	}

	// Add separator if needed
	if path[len(path)-1] != '/' && path[len(path)-1] != '\\' {
		path += "/"
	}

	// Append the base filename
	return path + session.CurrentMessageArea.File
}

// GetCurrentMessageAreaName returns the display name of the current message area
func (session *TelnetSession) GetCurrentMessageAreaName() string {
	if session.CurrentMessageArea == nil {
		return "No Area Selected"
	}
	return session.CurrentMessageArea.Name
}
