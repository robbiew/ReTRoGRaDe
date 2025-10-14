package tui

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
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
	ColorBlue      = "33"
	ColorGray      = "240"
	ColorLightBlue = "39"
	ColorLightGray = "252"
	ColorDarkGray  = "235"
	ColorDarkGray2 = "243"
)

// GuidedSetupModel represents the guided setup form state
type GuidedSetupModel struct {
	// Form fields
	fields      []SetupField
	fieldIndex  int
	confirmMode bool

	// Text input for editing
	textInput textinput.Model

	// UI state
	screenWidth  int
	screenHeight int

	// Setup data
	rootDir string
	config  *ConfigData
}

// SetupField represents a single field in the setup form
type SetupField struct {
	Label    string
	Value    string
	ReadOnly bool // For database path which is fixed
}

// ConfigData holds the collected configuration
type ConfigData struct {
	Root        string
	Data        string
	Files       string
	Msgs        string
	Logs        string
	Theme       string
	CreateSysop bool
	SysopData   *SysopData
}

// SysopData holds sysop account information
type SysopData struct {
	Username string
	Password string
	RealName string
	Email    string
	Location string
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
		{Label: "Root", Value: rootDir, ReadOnly: false},
		{Label: "Data", Value: rootDir + "/data", ReadOnly: false},
		{Label: "Files", Value: rootDir + "/files", ReadOnly: false},
		{Label: "Msgs", Value: rootDir + "/msgs", ReadOnly: false},
		{Label: "Logs", Value: rootDir + "/logs", ReadOnly: false},
		{Label: "Theme", Value: rootDir + "/theme", ReadOnly: false},
	}

	return GuidedSetupModel{
		fields:      fields,
		fieldIndex:  0,
		confirmMode: false,
		textInput:   ti,
		rootDir:     rootDir,
		config:      &ConfigData{Root: rootDir},
	}
}

// Init implements tea.Model
func (m GuidedSetupModel) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, tea.EnterAltScreen)
}

// Update handles all input events
func (m GuidedSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.confirmMode {
				m.confirmMode = false
				// Stay on last field
			} else if m.fieldIndex > 0 {
				m.fieldIndex--
			}
			return m, nil

		case "down", "j":
			if m.fieldIndex < len(m.fields)-1 {
				m.fieldIndex++
				m.confirmMode = false
			} else {
				m.confirmMode = true
			}
			return m, nil

		case "enter":
			if m.confirmMode {
				// Collect all field values into config
				m.config.Root = m.fields[0].Value
				m.config.Data = m.fields[1].Value
				m.config.Files = m.fields[2].Value
				m.config.Msgs = m.fields[3].Value
				m.config.Logs = m.fields[4].Value
				m.config.Theme = m.fields[5].Value

				// Return completion message
				return m, tea.Quit
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

// View renders the guided setup form
func (m GuidedSetupModel) View() string {
	var content strings.Builder

	// ANSI Art Header (5 rows high, 80 columns wide)
	// Decode CP437 -> UTF-8 and rasterize to 80x5
	rdr := transform.NewReader(bytes.NewReader([]byte(guidedArt)), charmap.CodePage437.NewDecoder())
	decoded, err := io.ReadAll(rdr)
	if err != nil {
		// Fallback if decoding fails
		content.WriteString("Retrograde Setup\n\n")
	} else {
		// Normalize newlines and remove SAUCE metadata
		s := strings.ReplaceAll(string(decoded), "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		s = trimStringFromSauce(s)
		// Rasterize ANSI to 80x5
		artLines := rasterizeANSIToLines(s, 80, 7)
		artHeader := strings.Join(artLines, "\n")
		content.WriteString(artHeader)
		content.WriteString("\n\n")
	}

	// Form fields - align PATHS fields
	for i, field := range m.fields {
		isSelected := i == m.fieldIndex && !m.confirmMode

		// Simple label formatting without padding
		label := field.Label + ": "

		var valuePart string
		if isSelected && m.textInput.Focused() {
			// Show text input for editing
			valuePart = m.textInput.View()
		} else {
			// Show static value
			valuePart = field.Value
			if len(valuePart) > 40 {
				valuePart = valuePart[:37] + "..."
			}
		}

		// Always separate label and value styling
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLightGray)).
			Background(lipgloss.Color(ColorDarkGray))

		var valueStyle lipgloss.Style
		if isSelected {
			// Highlight selected field value - blue background, white text
			valueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorWhite)).
				Background(lipgloss.Color(ColorBlue)).
				Bold(true)
		} else {
			// Normal value styling
			valueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorLightGray)).
				Background(lipgloss.Color(ColorDarkGray))
		}

		styledLine := labelStyle.Render(label) + valueStyle.Render(valuePart)
		content.WriteString(styledLine)
		content.WriteString("\n")
	}

	// Confirm button
	confirmLabel := "[CONFIRM]"
	if m.confirmMode {
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Background(lipgloss.Color(ColorBlue)).
			Bold(true).
			Padding(0, 2)
		content.WriteString("\n")
		content.WriteString(confirmStyle.Render(confirmLabel))
	} else {
		normalStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLightGray)).
			Background(lipgloss.Color(ColorDarkGray)).
			Padding(0, 2)
		content.WriteString("\n")
		content.WriteString(normalStyle.Render(confirmLabel))
	}

	content.WriteString("\n\n")

	// Instructions
	instructions := "Use ↑↓ arrows to navigate, Enter to edit/select, Esc to cancel editing, CTLR+C to quit"
	instStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorDarkGray2)).
		Align(lipgloss.Center).
		Width(60)
	content.WriteString(instStyle.Render(instructions))

	// Center the entire form
	formStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.screenWidth).
		Height(m.screenHeight)

	return formStyle.Render(content.String())
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

	return setupModel.config, nil
}
