# TUI Implementation: Selectable CmdKeys Field

## Overview
This document describes the changes needed to replace the text input for CmdKeys with a selectable list of defined command keys from `internal/menu/cmdkeys.go`.

## Changes Required

### 1. Add SelectValue Type to `internal/tui/editor.go`

Add a new value type to the ValueType constants:

```go
// ValueType defines supported data types
type ValueType int

const (
	StringValue ValueType = iota
	IntValue
	BoolValue
	ListValue  // Comma-separated list
	PortValue  // Integer with port range validation
	PathValue  // File/directory path
	SelectValue // Selectable list of options
)
```

### 2. Extend MenuItem with SelectOptions

Add a field to MenuItem to hold selectable options:

```go
// MenuItem represents an editable configuration value
type MenuItem struct {
	ID            string         // Unique identifier
	Label         string         // Display label
	Field         ConfigField    // Link to configuration field
	ValueType     ValueType      // Data type
	Validation    ValidationFunc // Validation function
	HelpText      string         // Help text for editing
	SelectOptions []SelectOption // Options for SelectValue type (NEW)
}

// SelectOption represents a selectable option (NEW)
type SelectOption struct {
	Value       string // The actual value to store (e.g., "MM", "MP", "G")
	Label       string // Display name (e.g., "Read Mail", "Post Message")
	Description string // Additional info shown in selection list
	Category    string // Optional category for grouping
}
```

### 3. Create Helper Function to Load CmdKeys

Add a helper function in `internal/tui/editor_menu_helpers.go` (or create it):

```go
package tui

import (
	"sort"
	"github.com/robbiew/retrograde/internal/menu"
)

// getCmdKeySelectOptions returns SelectOptions for all registered command keys
func getCmdKeySelectOptions() []SelectOption {
	registry := menu.NewCmdKeyRegistry()
	definitions := registry.GetAllDefinitions()
	
	// Sort by category, then by name
	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].Category == definitions[j].Category {
			return definitions[i].Name < definitions[j].Name
		}
		return definitions[i].Category < definitions[j].Category
	})
	
	options := make([]SelectOption, 0, len(definitions))
	for _, def := range definitions {
		options = append(options, SelectOption{
			Value:       def.CmdKey,
			Label:       def.Name,
			Description: def.Description,
			Category:    def.Category,
		})
	}
	
	return options
}
```

### 4. Update `setupMenuEditCommandModal` in `internal/tui/editor_view.go`

Replace the CmdKeys field definition with SelectValue type:

```go
{
	ID:       "command-cmdkeys",
	Label:    "CmdKeys",
	ItemType: EditableField,
	EditableItem: &MenuItem{
		ID:        "command-cmdkeys",
		Label:     "CmdKeys",
		ValueType: SelectValue, // CHANGED from StringValue
		Field: ConfigField{
			GetValue: func() interface{} { return m.editingMenuCommand.CmdKeys },
			SetValue: func(v interface{}) error {
				m.editingMenuCommand.CmdKeys = v.(string)
				return nil
			},
		},
		HelpText:      "Command key for execution handler",
		SelectOptions: getCmdKeySelectOptions(), // NEW
	},
},
```

### 5. Add Selection UI to `internal/tui/editor_update.go`

In the `handleMenuEditCommand` function, add special handling for SelectValue when user presses Enter:

```go
case "enter":
	// Edit selected field
	selectedField := m.modalFields[m.modalFieldIndex]
	if selectedField.ItemType == EditableField && selectedField.EditableItem != nil {
		m.editingItem = selectedField.EditableItem
		m.originalValue = m.editingItem.Field.GetValue()
		m.editingError = ""

		// Initialize text input based on value type
		if m.editingItem.ValueType == BoolValue {
			m.textInput.SetValue("")
		} else if m.editingItem.ValueType == SelectValue {
			// NEW: Handle SelectValue - create a selection list
			m.navMode = SelectingValue // NEW navigation mode
			m.selectListIndex = 0 // Initialize selection
			// Find current value in options
			currentValue := m.editingItem.Field.GetValue().(string)
			for i, opt := range m.editingItem.SelectOptions {
				if opt.Value == currentValue {
					m.selectListIndex = i
					break
				}
			}
			m.message = ""
		} else {
			// Existing text input handling...
			currentValue := m.editingItem.Field.GetValue()
			m.setTextInputValueWithCursor(formatValue(currentValue, m.editingItem.ValueType))
			// ... rest of existing code
		}
	}
	return m, nil
```

### 6. Add SelectingValue Navigation Mode

Add a new navigation mode constant in `internal/tui/editor.go`:

```go
const (
	MainMenuNavigation NavigationMode = iota
	Level2MenuNavigation
	Level3MenuNavigation
	Level4ModalNavigation
	EditingValue
	SelectingValue // NEW
	SaveChangesPrompt
	// ... other modes
)
```

Add fields to Model struct to track selection state:

```go
type Model struct {
	// ... existing fields ...
	
	// Selection state (NEW)
	selectListIndex    int            // Currently selected option index
	selectFilterText   string         // Optional filter text
	// ... rest of fields
}
```

### 7. Implement SelectingValue Input Handler

Add a new handler function in `internal/tui/editor_update.go`:

```go
// handleSelectingValue processes input when selecting from a list
func (m Model) handleSelectingValue(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingItem == nil || m.editingItem.ValueType != SelectValue {
		m.navMode = m.returnToMenuModifyOrModal()
		return m, nil
	}

	options := m.editingItem.SelectOptions

	switch msg.String() {
	case "up", "k":
		if m.selectListIndex > 0 {
			m.selectListIndex--
		}
		m.message = ""
		return m, nil
		
	case "down", "j":
		if m.selectListIndex < len(options)-1 {
			m.selectListIndex++
		}
		m.message = ""
		return m, nil
		
	case "pgup":
		m.selectListIndex -= 10
		if m.selectListIndex < 0 {
			m.selectListIndex = 0
		}
		m.message = ""
		return m, nil
		
	case "pgdown":
		m.selectListIndex += 10
		if m.selectListIndex >= len(options) {
			m.selectListIndex = len(options) - 1
		}
		m.message = ""
		return m, nil
		
	case "home":
		m.selectListIndex = 0
		m.message = ""
		return m, nil
		
	case "end":
		m.selectListIndex = len(options) - 1
		m.message = ""
		return m, nil
		
	case "enter":
		// Select the current option
		selectedOption := options[m.selectListIndex]
		if err := m.editingItem.Field.SetValue(selectedOption.Value); err != nil {
			m.editingError = err.Error()
			return m, nil
		}
		
		// Mark as modified
		m.modifiedCount++
		if m.editingMenu != nil {
			m.menuModified = true
		}
		
		// Return to previous mode
		m.navMode = m.returnToMenuModifyOrModal()
		m.editingItem = nil
		m.editingError = ""
		m.message = fmt.Sprintf("CmdKey set to: %s (%s)", selectedOption.Value, selectedOption.Label)
		m.messageTime = time.Now()
		m.messageType = SuccessMessage
		return m, nil
		
	case "esc":
		// Cancel selection
		m.navMode = m.returnToMenuModifyOrModal()
		m.editingItem = nil
		m.editingError = ""
		m.message = ""
		return m, nil
		
	case "/":
		// Future: implement filtering
		m.message = "Filtering not yet implemented"
		m.messageTime = time.Now()
		m.messageType = InfoMessage
		return m, nil
	}

	return m, nil
}
```

### 8. Add Case to Main Update Handler

In `internal/tui/editor_update.go`, add the new case to the Update method:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// ... existing code ...
		
		switch m.navMode {
		case MainMenuNavigation:
			return m.handleMainMenu(msg)
		// ... other cases ...
		case SelectingValue:
			return m.handleSelectingValue(msg) // NEW
		// ... rest of cases
		}
	// ... rest of Update method
}
```

### 9. Implement Selection View Rendering

Add a view function in `internal/tui/editor_view.go`:

```go
// renderSelectingValueView renders the selection list modal
func (m Model) renderSelectingValueView() string {
	if m.editingItem == nil || m.editingItem.ValueType != SelectValue {
		return ""
	}

	var b strings.Builder
	options := m.editingItem.SelectOptions
	
	// Modal dimensions
	modalWidth := 70
	modalHeight := 20
	visibleItems := modalHeight - 6 // Account for header, footer, borders
	
	// Title
	title := fmt.Sprintf(" Select %s ", m.editingItem.Label)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextBright)).
		Background(lipgloss.Color(ColorPrimary)).
		Bold(true)
	
	// Calculate scroll window
	startIdx := 0
	endIdx := len(options)
	if len(options) > visibleItems {
		// Center the selection in the window
		startIdx = m.selectListIndex - (visibleItems / 2)
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + visibleItems
		if endIdx > len(options) {
			endIdx = len(options)
			startIdx = endIdx - visibleItems
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}
	
	// Build content
	var content strings.Builder
	content.WriteString(titleStyle.Render(title) + "\n\n")
	
	lastCategory := ""
	for i := startIdx; i < endIdx; i++ {
		opt := options[i]
		
		// Show category header if it changed
		if opt.Category != lastCategory {
			if lastCategory != "" {
				content.WriteString("\n")
			}
			categoryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorAccent)).
				Bold(true)
			content.WriteString(categoryStyle.Render(fmt.Sprintf("— %s —", opt.Category)) + "\n")
			lastCategory = opt.Category
		}
		
		// Render option
		isSelected := i == m.selectListIndex
		
		// Format: "[CmdKey] Name - Description"
		line := fmt.Sprintf(" [%s] %-20s %s", opt.Value, opt.Label, opt.Description)
		if len(line) > modalWidth-4 {
			line = line[:modalWidth-7] + "..."
		}
		
		if isSelected {
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorAccent)).
				Bold(true)
			content.WriteString(style.Render(line) + "\n")
		} else {
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextNormal))
			content.WriteString(style.Render(line) + "\n")
		}
	}
	
	// Scrollbar indicator
	if len(options) > visibleItems {
		scrollInfo := fmt.Sprintf(" %d-%d of %d ", startIdx+1, endIdx, len(options))
		content.WriteString("\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Render(scrollInfo))
	}
	
	// Help text
	help := " ↑↓ Navigate | Enter Select | Esc Cancel "
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Italic(true)
	content.WriteString("\n" + helpStyle.Render(help))
	
	// Create modal box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		Padding(1, 2).
		Width(modalWidth)
	
	b.WriteString(boxStyle.Render(content.String()))
	
	return b.String()
}
```

### 10. Update View Method to Show Selection Modal

In `internal/tui/editor_view.go`, modify the main View method:

```go
func (m Model) View() string {
	// ... existing code ...
	
	switch m.navMode {
	// ... other cases ...
	case SelectingValue:
		// Render the selection modal over the main view
		mainView := m.renderMenuEditCommandView() // or appropriate background
		selectionModal := m.renderSelectingValueView()
		
		// Center the modal
		centered := lipgloss.Place(
			m.screenWidth,
			m.screenHeight,
			lipgloss.Center,
			lipgloss.Center,
			selectionModal,
		)
		
		return mainView + "\n" + centered
	// ... other cases ...
	}
}
```

## Testing

1. Start the config editor: `./retrograde config`
2. Navigate to: Editors → Menus → Select a menu → Commands tab → Edit or Insert command
3. Select the CmdKeys field and press Enter
4. You should see a selectable list of all available command keys
5. Use arrow keys to navigate, Enter to select, Esc to cancel

## Benefits

1. **User-friendly**: No need to memorize or type command key codes
2. **Self-documenting**: Each option shows its name and description
3. **Organized**: Commands grouped by category
4. **Error-proof**: Can't enter invalid command keys
5. **Discoverable**: Users can browse available commands

## Future Enhancements

- Add filtering by typing (filter as you type)
- Add search functionality (press '/' to search)
- Show which commands are already used in the menu
- Color-code commands by implementation status (implemented vs. placeholder)