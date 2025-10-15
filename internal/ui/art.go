package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// AnsiArtConfig holds configuration for ANSI art loading
type AnsiArtConfig struct {
	mu             sync.RWMutex
	ThemeDirectory string // Path to the theme directory from SQLite
}

var config AnsiArtConfig

// SetThemeDirectory sets the theme directory for ANSI art files
func SetThemeDirectory(dir string) {
	config.mu.Lock()
	defer config.mu.Unlock()
	config.ThemeDirectory = filepath.Clean(dir)
}

// getThemeDirectory safely retrieves the theme directory
func getThemeDirectory() string {
	config.mu.RLock()
	defer config.mu.RUnlock()
	return config.ThemeDirectory
}

// PrintAnsiTerminal displays ANSI art content with optional line delay and height limit
func PrintAnsiTerminal(term InteractiveTerminal, artName string, delay, height int) error {
	lines, err := LoadAnsiLines(artName)
	if err != nil {
		return fmt.Errorf("error: ANSI art '%s' not found", artName)
	}

	delayDuration := time.Duration(delay) * time.Millisecond
	printed := 0

	for _, line := range lines {
		if height > 0 && printed >= height {
			break
		}
		if err := term.Print(line + "\r\n"); err != nil {
			return err
		}
		if delay > 0 {
			time.Sleep(delayDuration)
		}
		printed++
	}
	return nil
}

// PrintAnsiArt loads and displays an ANSI art file to the terminal
func PrintAnsiArt(term InteractiveTerminal, filename string) error {
	themeDir := getThemeDirectory()

	if themeDir == "" {
		return fmt.Errorf("theme directory not configured")
	}

	// Try multiple file extensions if no extension provided
	candidates := buildFileCandidates(filename)

	var lastErr error
	for _, candidate := range candidates {
		artPath := filepath.Join(themeDir, candidate)

		// Read the file
		data, err := os.ReadFile(artPath)
		if err != nil {
			lastErr = err
			continue // Try next candidate
		}

		// Remove SAUCE metadata
		content := stripSauce(string(data))

		// Print to terminal
		if err := term.Print(content); err != nil {
			return fmt.Errorf("failed to print ANSI art: %w", err)
		}

		return nil // Success!
	}

	return fmt.Errorf("failed to load ANSI art '%s' from '%s': %w", filename, themeDir, lastErr)
}

// buildFileCandidates returns a list of filenames to try
func buildFileCandidates(name string) []string {
	name = strings.TrimSpace(name)

	// If filename already has an extension, just use it
	if filepath.Ext(name) != "" {
		return []string{name}
	}

	// Otherwise, try common ANSI art extensions
	return []string{
		name,
		name + ".ans",
		name + ".asc",
	}
}

// LoadAnsiLines returns ANSI art content split into lines (without trailing carriage returns)
func LoadAnsiLines(artName string) ([]string, error) {
	themeDir := getThemeDirectory()

	if themeDir == "" {
		return nil, fmt.Errorf("theme directory not configured")
	}

	// Build full path
	artPath := filepath.Join(themeDir, artName)

	// Read file
	data, err := os.ReadFile(artPath)
	if err != nil {
		return nil, fmt.Errorf("ANSI art '%s' not found", artName)
	}

	// Strip SAUCE and split into lines
	content := stripSauce(string(data))
	rawLines := strings.Split(content, "\n")

	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		lines = append(lines, strings.TrimRight(line, "\r"))
	}

	return lines, nil
}

// stripSauce removes SAUCE metadata from ANSI art content
func stripSauce(content string) string {
	// SAUCE records start with "SAUCE00" marker
	if idx := strings.Index(content, "SAUCE00"); idx != -1 {
		// Trim the last character before SAUCE (usually EOF marker)
		return trimLastChar(content[:idx])
	}

	// Also check for COMNT blocks
	if idx := strings.Index(content, "COMNT"); idx != -1 {
		return trimLastChar(content[:idx])
	}

	return content
}

// trimLastChar removes the last character from a string
func trimLastChar(s string) string {
	if len(s) > 0 {
		_, size := utf8.DecodeLastRuneInString(s)
		return s[:len(s)-size]
	}
	return s
}
