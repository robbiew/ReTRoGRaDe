package tui

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

//go:embed guided.ans
var guidedArt string

// trimStringFromSauce trims SAUCE metadata from a string (if necessary)
func trimStringFromSauce(s string) string {
	return trimMetadata(s, "COMNT", "SAUCE00")
}

// Helper to trim metadata based on delimiters
func trimMetadata(s string, delimiters ...string) string {
	for _, delimiter := range delimiters {
		if idx := strings.Index(s, delimiter); idx != -1 {
			return trimLastChar(s[:idx])
		}
	}
	return s
}

// trimLastChar trims the last character from a string
func trimLastChar(s string) string {
	if len(s) > 0 {
		_, size := utf8.DecodeLastRuneInString(s)
		return s[:len(s)-size]
	}
	return s
}

// Color constants
const (
	ColorWhite     = "15"
	ColorBlue      = "4"
	ColorRed       = "1"
	ColorGray      = "8"
	ColorLightBlue = "12"
	ColorLightGray = "7"
	ColorDarkGray  = "0"
	ColorDarkGray2 = "8"
)

// GuidedSetupModel represents the guided setup form state
type GuidedSetupModel struct {
	// Form fields
	fields      []SetupField
	fieldIndex  int
	confirmMode bool
	buttonIndex int // 0 for CONFIRM, 1 for CANCEL

	// Text input for editing
	textInput textinput.Model

	// UI state
	screenWidth  int
	screenHeight int

	// Setup data
	rootDir   string
	config    *ConfigData
	cancelled bool // Track if user cancelled

	// Message system
	message     string
	messageTime time.Time
	messageType int // 0=Info, 1=Success, 2=Warning, 3=Error
}

// SetupField represents a single field in the setup form
type SetupField struct {
	Label    string
	Value    string
	ReadOnly bool   // For database path which is fixed
	HelpText string // Help text displayed for the field
}

// ConfigData holds the collected configuration
type ConfigData struct {
	Root     string
	Data     string
	Files    string
	Msgs     string
	Logs     string
	Security string
	Theme    string
}

// InitialGuidedSetupModel creates the initial guided setup model
func InitialGuidedSetupModel(rootDir string) GuidedSetupModel {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.CharLimit = 200
	ti.Width = 40
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWhite))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWhite)).Background(lipgloss.Color(ColorBlue))
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWhite)).Background(lipgloss.Color(ColorBlue))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray)).Background(lipgloss.Color(ColorBlue))

	fields := []SetupField{
		{Label: "    Root", Value: rootDir, ReadOnly: false, HelpText: "Base directory for Retrograde installation"},
		{Label: "    Data", Value: rootDir + "/data", ReadOnly: false, HelpText: "Directory for the database(s) and data files"},
		{Label: "   Files", Value: rootDir + "/files", ReadOnly: false, HelpText: "Directory for upload/download area storage"},
		{Label: "    Msgs", Value: rootDir + "/msgs", ReadOnly: false, HelpText: "Directory for message base files"},
		{Label: "    Logs", Value: rootDir + "/logs", ReadOnly: false, HelpText: "Directory for log files"},
		{Label: "Security", Value: rootDir + "/security", ReadOnly: false, HelpText: "Directory for security assets, like blacklists"},
		{Label: "   Theme", Value: rootDir + "/theme", ReadOnly: false, HelpText: "Directory for art and text-based files"},
	}

	return GuidedSetupModel{
		fields:      fields,
		fieldIndex:  1, // Start with Data field selected (index 1)
		confirmMode: false,
		buttonIndex: 0, // Start with CONFIRM selected
		textInput:   ti,
		rootDir:     rootDir,
		config:      &ConfigData{Root: rootDir},
	}
}

// Init implements tea.Model
func (m GuidedSetupModel) Init() tea.Cmd {
	// Focus on first field immediately
	m.textInput.SetValue(m.fields[0].Value)
	m.textInput.Focus()
	return tea.Batch(tea.ClearScreen, tea.EnterAltScreen)
}

func (m GuidedSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.confirmMode {
				// Navigate between buttons
				if m.buttonIndex > 0 {
					m.buttonIndex--
				} else {
					m.confirmMode = false
				}
			} else if m.fieldIndex > 0 {
				m.fieldIndex--
			}
			return m, nil

		case "down", "j":
			if m.confirmMode {
				// Navigate between buttons
				if m.buttonIndex < 1 {
					m.buttonIndex++
				}
			} else if m.fieldIndex < len(m.fields)-1 {
				m.fieldIndex++
			} else {
				m.confirmMode = true
				m.buttonIndex = 0 // Reset to first button
			}
			return m, nil

		case "left", "h":
			if m.confirmMode {
				// Switch between CONFIRM and CANCEL buttons
				if m.buttonIndex > 0 {
					m.buttonIndex--
				} else {
					m.buttonIndex = 1
				}
				return m, nil
			}
		case "right", "l":
			if m.confirmMode {
				// Switch between CONFIRM and CANCEL buttons
				if m.buttonIndex < 1 {
					m.buttonIndex++
				} else {
					m.buttonIndex = 0
				}
				return m, nil
			}
		case "enter":
			if m.confirmMode {
				if m.buttonIndex == 0 {
					// CONFIRM button pressed
					// Validate all fields are not empty
					for _, field := range m.fields {
						if strings.TrimSpace(field.Value) == "" {
							// Show warning message for empty fields
							m.message = "Path cannot be empty."
							m.messageTime = time.Now()
							m.messageType = 3 // Error
							return m, nil
						}
					}

					// Collect all field values into config
					m.config.Root = m.fields[0].Value
					m.config.Data = m.fields[1].Value
					m.config.Files = m.fields[2].Value
					m.config.Msgs = m.fields[3].Value
					m.config.Logs = m.fields[4].Value
					m.config.Security = m.fields[5].Value
					m.config.Theme = m.fields[6].Value

					// Create directories
					dirs := []string{
						m.config.Root,
						m.config.Data,
						m.config.Files,
						m.config.Msgs,
						m.config.Logs,
						m.config.Security,
						m.config.Theme,
					}

					for _, dir := range dirs {
						if err := os.MkdirAll(dir, 0755); err != nil {
							// Store error in config or handle it
							// For now, continue trying to create other dirs
							continue
						}
					}

					// Return completion
					return m, tea.Quit
				} else {
					// CANCEL button pressed
					m.cancelled = true
					return m, tea.Quit
				}
			} else {
				// If currently editing, stop editing and move to next field
				if m.textInput.Focused() {
					m.textInput.Blur()
					// Move to next field
					if m.fieldIndex < len(m.fields)-1 {
						m.fieldIndex++
					} else {
						m.confirmMode = true
					}
				} else {
					// Start editing current field
					currentField := m.fields[m.fieldIndex]
					if !currentField.ReadOnly {
						m.textInput.SetValue(currentField.Value)
						m.textInput.Focus()
					}
				}
			}
			return m, nil

		case "esc":
			// Stop editing if currently editing
			m.textInput.Blur()
			return m, nil
		}

		// Handle text input when editing
		if m.textInput.Focused() {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)

			// Update field value in real-time
			if m.fieldIndex < len(m.fields) {
				m.fields[m.fieldIndex].Value = m.textInput.Value()
			}

			return m, cmd
		}
	}

	return m, nil
}

func (m GuidedSetupModel) View() string {
	var header strings.Builder
	var formFields strings.Builder
	var footer strings.Builder

	// ANSI Art Header
	rdr := transform.NewReader(bytes.NewReader([]byte(guidedArt)), charmap.CodePage437.NewDecoder())
	decoded, err := io.ReadAll(rdr)
	if err != nil {
		header.WriteString("Retrograde Setup\n\n")
	} else {
		s := strings.ReplaceAll(string(decoded), "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		s = trimStringFromSauce(s)
		artLines := rasterizeANSIToLines(s, 80, 7)
		artHeader := strings.Join(artLines, "\n")
		header.WriteString(artHeader)
		header.WriteString("\n")
	}

	// Instructions
	instructions := "Use ↑↓ arrows to navigate fields, Enter to select / edit"
	instStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorDarkGray2)).
		Width(70).
		Align(lipgloss.Center) // Center text within the 70 width
	header.WriteString(instStyle.Render(instructions))
	header.WriteString("\n")
	if m.message != "" {
		header.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorRed)).
			Bold(true).
			Render(m.message))
	} else {
		header.WriteString(" ")
	}

	// Form fields
	const valueWidth = 45 // Fixed width for consistent highlighting

	for i, field := range m.fields {
		isSelected := i == m.fieldIndex && !m.confirmMode

		label := field.Label + ": "

		var valuePart string
		if isSelected && m.textInput.Focused() {
			valuePart = m.textInput.View()
		} else {
			valuePart = field.Value
			if len(valuePart) > 40 {
				valuePart = valuePart[:37] + "..."
			}
		}

		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLightGray)).
			Background(lipgloss.Color(ColorDarkGray))

		var valueStyle lipgloss.Style
		if isSelected {
			// Fixed width for consistent highlighting
			valueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorWhite)).
				Background(lipgloss.Color(ColorBlue)).
				Width(valueWidth)
		} else {
			valueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorLightGray)).
				Background(lipgloss.Color(ColorDarkGray)).
				Width(valueWidth)
		}

		styledLine := labelStyle.Render(label) + valueStyle.Render(valuePart)
		formFields.WriteString(styledLine)
		formFields.WriteString("\n")
	}

	// Confirm/Cancel buttons
	footer.WriteString("\n")

	// CONFIRM button
	confirmLabel := " CONFIRM "
	if m.confirmMode && m.buttonIndex == 0 {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Background(lipgloss.Color(ColorBlue)).
			Bold(true).
			Padding(0, 1).
			MarginRight(2)
		footer.WriteString(confirmStyle.Render(confirmLabel))
	} else {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Background(lipgloss.Color("240")). // Grey background
			Padding(0, 1).
			MarginRight(2)
		footer.WriteString(confirmStyle.Render(confirmLabel))
	}

	// CANCEL button
	cancelLabel := " CANCEL "
	if m.confirmMode && m.buttonIndex == 1 {
		cancelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Background(lipgloss.Color(ColorRed)).
			Bold(true).
			Padding(0, 1)
		footer.WriteString(cancelStyle.Render(cancelLabel))
	} else {
		cancelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Background(lipgloss.Color("240")). // Grey background
			Padding(0, 1)
		footer.WriteString(cancelStyle.Render(cancelLabel))
	}

	footer.WriteString("\n")

	// Apply center alignment to header and footer
	headerCentered := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.screenWidth).
		Render(header.String())

	footerCentered := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.screenWidth).
		Render(footer.String())

	// Move form more to the right for better centering
	formCentered := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Width(m.screenWidth).
		PaddingLeft((m.screenWidth - 50) / 2). // Adjusted from 60 to 50 for more centering
		Render(formFields.String())

	// Help text for selected field
	helpText := ""
	if !m.confirmMode && m.fieldIndex < len(m.fields) {
		helpText = m.fields[m.fieldIndex].HelpText
	}
	helpSection := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorLightGray)).
		Align(lipgloss.Center).
		Width(m.screenWidth).
		Render(helpText)

	// Combine all sections
	finalContent := headerCentered + formCentered + footerCentered + "\n" + helpSection

	// Apply height centering
	finalStyle := lipgloss.NewStyle().
		Height(m.screenHeight)

	return finalStyle.Render(finalContent)
}

// RunGuidedSetupTUI runs the guided setup TUI and returns the configuration
func RunGuidedSetupTUI(rootDir string) (*ConfigData, error) {
	model := InitialGuidedSetupModel(rootDir)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	setupModel, ok := finalModel.(GuidedSetupModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	// Return nil if user cancelled
	if setupModel.cancelled {
		return nil, nil
	}

	return setupModel.config, nil
}
