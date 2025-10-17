package tui

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/database"
)

// ============================================================================
// Helper Functions
// ============================================================================

// setTextInputValueWithCursor sets the text input value and positions cursor at the end
func (m *Model) setTextInputValueWithCursor(value string) {
	m.textInput.SetValue(value)
	m.textInput.SetCursor(len(value))
}

// ============================================================================
// Update Logic
// ============================================================================

// Update handles all input events
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit handling
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// For EditingValue mode, handle text input first (except for control keys)
		if m.navMode == EditingValue {
			// Check if editing a boolean field
			isBoolField := m.editingItem != nil && m.editingItem.ValueType == BoolValue

			// For boolean fields, route Y/N/Space/Tab to handler
			if isBoolField {
				if msg.String() == "enter" || msg.String() == "esc" ||
					msg.String() == "y" || msg.String() == "Y" ||
					msg.String() == "n" || msg.String() == "N" ||
					msg.String() == " " || msg.String() == "tab" {
					return m.handleEditingValue(msg)
				}
			} else {
				// For text fields, only route Enter and Esc
				if msg.String() == "enter" || msg.String() == "esc" {
					return m.handleEditingValue(msg)
				}
			}

			// For all other keys (typing), update text input
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// State-specific handling for other modes
		switch m.navMode {
		case MainMenuNavigation:
			return m.handleMainMenuNavigation(msg)
		case Level2MenuNavigation:
			return m.handleLevel2MenuNavigation(msg)
		case Level3MenuNavigation:
			return m.handleLevel3MenuNavigation(msg)
		case Level4ModalNavigation:
			return m.handleLevel4ModalNavigation(msg)
		case UserManagementMode:
			return m.handleUserManagement(msg)
		case SecurityLevelsMode:
			return m.handleSecurityLevelsManagement(msg)
		case MenuManagementMode:
			return m.handleMenuManagement(msg)
		case MenuModifyMode:
			return m.handleMenuModify(msg)
		case MenuEditDataMode:
			return m.handleMenuEditData(msg)
		case MenuEditCommandMode:
			return m.handleMenuEditCommand(msg)
		case SavePrompt:
			return m.handleSavePrompt(msg)
		case SaveChangesPrompt: // ADD THIS
			return m.handleSaveChangesPrompt(msg)

		}
	}

	// Update submenu list if in Level 2 menu navigation mode
	if m.navMode == Level2MenuNavigation {
		m.submenuList, cmd = m.submenuList.Update(msg)
		return m, cmd
	}

	// Update user list if in user management mode
	if m.navMode == UserManagementMode {
		m.userListUI, cmd = m.userListUI.Update(msg)
		return m, cmd
	}

	// Update security levels list if in security levels management mode
	if m.navMode == SecurityLevelsMode {
		m.securityLevelsUI, cmd = m.securityLevelsUI.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleLevel4ModalNavigation processes input while in Level 4 modal navigation mode
func (m Model) handleLevel4ModalNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Navigate up in field list
		if m.modalFieldIndex > 0 {
			m.modalFieldIndex--
		}
		m.message = ""
		return m, nil
	case "down", "j":
		// Navigate down in field list
		if m.modalFieldIndex < len(m.modalFields)-1 {
			m.modalFieldIndex++
		}
		m.message = ""
		return m, nil
	case "enter":
		// Edit selected field
		selectedField := m.modalFields[m.modalFieldIndex]
		if selectedField.ItemType == EditableField && selectedField.EditableItem != nil {
			m.editingItem = selectedField.EditableItem
			m.originalValue = m.editingItem.Field.GetValue()
			m.editingError = ""

			// Initialize text input based on value type
			if m.editingItem.ValueType == BoolValue {
				// For bool, we don't use text input, just toggle
				m.textInput.SetValue("")
			} else {
				// Set current value as text
				currentValue := m.editingItem.Field.GetValue()
				m.setTextInputValueWithCursor(formatValue(currentValue, m.editingItem.ValueType))

				// Set placeholder, char limit and width based on type
				switch m.editingItem.ValueType {
				case PortValue:
					m.textInput.Placeholder = "1-65535"
					m.textInput.CharLimit = 5
					m.textInput.Width = 10
				case IntValue:
					m.textInput.Placeholder = "Enter number"
					m.textInput.CharLimit = 10
					m.textInput.Width = 15
				case PathValue:
					m.textInput.Placeholder = "Enter path"
					m.textInput.CharLimit = 200
					m.textInput.Width = 53 // Fixed width for scrolling
				case ListValue:
					m.textInput.Placeholder = "comma,separated,values"
					m.textInput.CharLimit = 200
					m.textInput.Width = 53 // Fixed width for scrolling
				default: // StringValue
					m.textInput.Placeholder = "Enter value"
					m.textInput.CharLimit = 200
					m.textInput.Width = 53 // Fixed width for scrolling
				}
			}

			m.textInput.Focus()
			m.navMode = EditingValue
			m.textInput.TextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.Cursor.Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright))
			m.textInput.PromptStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.PlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim)).
				Background(lipgloss.Color(ColorPrimary))
			m.message = ""
		}
		return m, nil
	case "esc":
		// Check if there are unsaved changes in this modal
		hasUnsavedChanges := false
		for _, field := range m.modalFields {
			if field.ItemType == EditableField && field.EditableItem != nil {
				// You could track which fields were modified, or just check modifiedCount
				if m.modifiedCount > 0 {
					hasUnsavedChanges = true
					break
				}
			}
		}

		if hasUnsavedChanges {
			// Show save prompt
			m.savePrompt = true
			m.savePromptSelection = 1 // Default to Yes
			m.navMode = SaveChangesPrompt

			// Remember where to go after save/cancel
			if m.editingUser != nil {
				m.returnToMode = UserManagementMode
			} else if m.editingSecurityLevel != nil {
				m.returnToMode = SecurityLevelsMode
			} else {
				hasSubSections := false
				for _, field := range m.modalFields {
					if field.ItemType == SectionHeader {
						hasSubSections = true
						break
					}
				}

				if hasSubSections {
					m.returnToMode = Level3MenuNavigation
				} else {
					m.returnToMode = Level2MenuNavigation
				}
			}

			return m, nil
		}

		// No unsaved changes, exit normally
		if m.editingUser != nil {
			// Return to user management list
			m.navMode = UserManagementMode
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
			m.editingUser = nil
		} else if m.editingSecurityLevel != nil {
			// Return to security levels list
			m.navMode = SecurityLevelsMode
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
			m.editingSecurityLevel = nil
		} else {
			hasSubSections := false
			for _, field := range m.modalFields {
				if field.ItemType == SectionHeader {
					hasSubSections = true
					break
				}
			}

			if hasSubSections {
				m.navMode = Level3MenuNavigation
			} else {
				m.navMode = Level2MenuNavigation
				m.modalFields = nil
				m.modalFieldIndex = 0
				m.modalSectionName = ""
			}
		}
		m.message = ""
		return m, nil
	}
	return m, nil
}

// Update handleSaveChangesPrompt to handle menu saving
func (m Model) handleSaveChangesPrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h", "right", "l", "tab":
		// Toggle between Yes and No
		m.savePromptSelection = (m.savePromptSelection + 1) % 2
		return m, nil
	case "up", "k":
		m.savePromptSelection = 1
		return m, nil
	case "down", "j":
		m.savePromptSelection = 0
		return m, nil
	case "enter":
		if m.savePromptSelection == 1 {
			// Yes - Save changes
			var err error

			if m.editingMenu != nil {
				// Save menu changes
				err = m.db.UpdateMenu(m.editingMenu)
				if err != nil {
					m.message = fmt.Sprintf("Error saving menu: %v", err)
					m.messageTime = time.Now()
					m.messageType = ErrorMessage
					m.savePrompt = false
					m.navMode = m.returnToMode
					return m, nil
				}

				// Save all command changes
				for _, cmd := range m.menuCommandsList {
					if cmd.ID == 0 {
						// New command - insert
						_, err = m.db.CreateMenuCommand(&cmd)
					} else {
						// Existing command - update
						err = m.db.UpdateMenuCommand(&cmd)
					}
					if err != nil {
						m.message = fmt.Sprintf("Error saving command: %v", err)
						m.messageTime = time.Now()
						m.messageType = ErrorMessage
						m.savePrompt = false
						m.navMode = m.returnToMode
						return m, nil
					}
				}

				// Reload menus list to reflect changes
				if reloadErr := m.loadMenus(); reloadErr != nil {
					m.message = fmt.Sprintf("Error reloading menus: %v", reloadErr)
					m.messageTime = time.Now()
					m.messageType = ErrorMessage
				}

				m.menuModified = false
				m.editingMenu = nil
				m.originalMenu = nil
				m.menuCommandsList = nil
				m.originalMenuCommands = nil
				m.menuDataFields = nil
			} else if m.editingSecurityLevel != nil {
				// Save security level changes
				err = m.db.UpdateSecurityLevel(m.editingSecurityLevel)
				if err != nil {
					m.message = fmt.Sprintf("Error saving security level: %v", err)
					m.messageTime = time.Now()
					m.messageType = ErrorMessage
					m.savePrompt = false
					m.navMode = m.returnToMode
					return m, nil
				}
				// Reload security levels list to reflect changes
				if reloadErr := m.loadSecurityLevels(); reloadErr != nil {
					m.message = fmt.Sprintf("Error reloading security levels: %v", reloadErr)
					m.messageTime = time.Now()
					m.messageType = ErrorMessage
				}
				m.editingSecurityLevel = nil
			} else if m.editingUser != nil {
				// Save user changes
				err = m.db.UpdateUser(m.editingUser)
				if err != nil {
					m.message = fmt.Sprintf("Error saving user: %v", err)
					m.messageTime = time.Now()
					m.messageType = ErrorMessage
					m.savePrompt = false
					m.navMode = m.returnToMode
					return m, nil
				}
				// Reload users list to reflect changes
				if reloadErr := m.loadUsers(); reloadErr != nil {
					m.message = fmt.Sprintf("Error reloading users: %v", reloadErr)
					m.messageTime = time.Now()
					m.messageType = ErrorMessage
				}
				m.editingUser = nil
			} else {
				// Save config changes
				err = config.SaveConfig(m.config, "")
				if err != nil {
					m.message = fmt.Sprintf("Error saving: %v", err)
					m.messageTime = time.Now()
					m.messageType = ErrorMessage
					m.savePrompt = false
					m.editingSecurityLevel = nil
					m.editingUser = nil
					m.navMode = m.returnToMode
					return m, nil
				}
			}
			// Reset counter after successful save
			m.modifiedCount = 0
		} else {
			// No - Discard changes
			if m.editingMenu != nil {
				// Restore original menu and commands (not needed, we're discarding)
				m.menuModified = false
				m.editingMenu = nil
				m.originalMenu = nil
				m.menuCommandsList = nil
				m.originalMenuCommands = nil
				m.menuDataFields = nil
			}
			// CRITICAL: Reset modifiedCount when discarding changes
			m.modifiedCount = 0
		}

		// Either way, return to previous mode
		m.savePrompt = false
		m.navMode = m.returnToMode
		m.editingSecurityLevel = nil
		m.editingUser = nil

		// Clean up modal if returning to Level 2
		if m.returnToMode == Level2MenuNavigation {
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
		}
		return m, nil
	case "y", "Y":
		// Quick Yes - same as enter with selection 1
		m.savePromptSelection = 1
		return m.handleSaveChangesPrompt(tea.KeyMsg{Type: tea.KeyEnter})
	case "n", "N":
		// Quick No - same as enter with selection 0
		m.savePromptSelection = 0
		return m.handleSaveChangesPrompt(tea.KeyMsg{Type: tea.KeyEnter})
	case "esc":
		// Cancel - return to menu modify mode (don't exit)
		m.savePrompt = false
		m.navMode = MenuModifyMode
	}
	return m, nil
}

// handleMainMenuNavigation processes input while in main menu navigation mode (Level 1)
func (m Model) handleMainMenuNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		m.activeMenu = (m.activeMenu - 1 + len(m.menuBar.Items)) % len(m.menuBar.Items)
		m.message = ""
		m.messageType = InfoMessage
	case "right", "l":
		m.activeMenu = (m.activeMenu + 1) % len(m.menuBar.Items)
		m.message = ""
		m.messageType = InfoMessage
	case "up", "k":
		m.activeMenu = (m.activeMenu - 1 + len(m.menuBar.Items)) % len(m.menuBar.Items)
		m.message = ""
		m.messageType = InfoMessage
	case "down", "j":
		m.activeMenu = (m.activeMenu + 1) % len(m.menuBar.Items)
		m.message = ""
		m.messageType = InfoMessage
	case "enter":
		// Enter Level 2 menu navigation mode
		m.navMode = Level2MenuNavigation
		// Update list with current menu's items
		listItems := buildListItems(m.menuBar.Items[m.activeMenu])

		// Use narrow width for level 2 menus (max 25 columns)
		maxWidth := 25
		m.submenuList.SetDelegate(submenuDelegate{maxWidth: maxWidth})
		m.submenuList.SetItems(listItems)
		m.submenuList.Select(0)
		m.message = ""
	case "esc":
		// Show exit confirmation modal
		m.savePrompt = true
		m.savePromptSelection = 0 // Default to "No" (index 0)
		m.navMode = SavePrompt
	case "1", "2", "3", "4", "5":
		// Direct menu access
		idx := int(msg.String()[0] - '1')
		if idx >= 0 && idx < len(m.menuBar.Items) {
			m.activeMenu = idx
			m.message = ""
		}
	}
	return m, nil
}

// handleLevel2MenuNavigation processes input while in Level 2 menu navigation mode
func (m Model) handleLevel2MenuNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Navigate up in list
		currentIdx := m.submenuList.Index()
		if currentIdx > 0 {
			m.submenuList.Select(currentIdx - 1)
		}
		m.message = ""
		return m, nil
	case "down", "j":
		// Navigate down in list
		currentIdx := m.submenuList.Index()
		items := m.submenuList.Items()
		if currentIdx < len(items)-1 {
			m.submenuList.Select(currentIdx + 1)
		}
		m.message = ""
		return m, nil
	case "home":
		// Jump to first item
		m.submenuList.Select(0)
		m.message = ""
	case "end":
		// Jump to last item
		items := m.submenuList.Items()
		if len(items) > 0 {
			m.submenuList.Select(len(items) - 1)
		}
		m.message = ""
	case "enter":
		// Select section - check what type of content it has
		selectedItem := m.submenuList.SelectedItem()
		if selectedItem != nil {
			item := selectedItem.(submenuListItem)
			if item.submenuItem.ItemType == SectionHeader {
				// Section selected - check if it has sub-items
				if len(item.submenuItem.SubItems) > 0 {
					// Check if sub-items are all editable fields or contain sub-sections
					hasSubSections := false
					for _, subItem := range item.submenuItem.SubItems {
						if subItem.ItemType == SectionHeader {
							hasSubSections = true
							break
						}
					}

					// Set up modal fields
					m.modalFields = item.submenuItem.SubItems
					m.modalFieldIndex = 0
					m.modalSectionName = item.submenuItem.Label

					if hasSubSections {
						// Has sub-sections, go to Level 3 navigation
						m.navMode = Level3MenuNavigation
					} else {
						// Only editable fields, go directly to Level 4 modal
						m.navMode = Level4ModalNavigation
					}
					m.message = ""
				} else {
					m.message = "This section has no sub-items"
					m.messageTime = time.Now()
				}
			} else if item.submenuItem.ItemType == ActionItem {
				// Handle action items
				switch item.submenuItem.ID {
				case "user-editor":
					// Launch user management interface
					// Check if database path is configured
					if m.config.Configuration.Paths.Database == "" {
						m.message = "Database path not configured. Please set it under Configuration > Paths > Database first."
						m.messageTime = time.Now()
						return m, nil
					}

					// Try to get database connection if not already available
					if m.db == nil {
						if existingDB := config.GetDatabase(); existingDB != nil {
							if sqliteDB, ok := existingDB.(*database.SQLiteDB); ok {
								m.db = sqliteDB
								// Ensure user schema is initialized
								if err := m.db.InitializeSchema(); err != nil {
									m.message = fmt.Sprintf("Failed to initialize database schema: %v", err)
									m.messageTime = time.Now()
									return m, nil
								}
							} else {
								m.message = "Database connection type mismatch"
								m.messageTime = time.Now()
								return m, nil
							}
						} else {
							m.message = "No database connection available"
							m.messageTime = time.Now()
							return m, nil
						}
					}

					// Load users
					if err := m.loadUsers(); err != nil {
						m.message = fmt.Sprintf("Error loading users: %v", err)
						m.messageTime = time.Now()
					} else {
						m.navMode = UserManagementMode
						m.message = ""
					}
				case "security-levels-editor":
					// Launch security levels management interface
					// Check if database path is configured
					if m.config.Configuration.Paths.Database == "" {
						m.message = "Database path not configured. Please set it under Configuration > Paths > Database first."
						m.messageTime = time.Now()
						return m, nil
					}

					// Try to get database connection if not already available
					if m.db == nil {
						if existingDB := config.GetDatabase(); existingDB != nil {
							if sqliteDB, ok := existingDB.(*database.SQLiteDB); ok {
								m.db = sqliteDB
								// Ensure user schema is initialized
								if err := m.db.InitializeSchema(); err != nil {
									m.message = fmt.Sprintf("Failed to initialize database schema: %v", err)
									m.messageTime = time.Now()
									return m, nil
								}
							} else {
								m.message = "Database connection type mismatch"
								m.messageTime = time.Now()
								return m, nil
							}
						} else {
							m.message = "No database connection available"
							m.messageTime = time.Now()
							return m, nil
						}
					}

					// Load security levels
					if err := m.loadSecurityLevels(); err != nil {
						m.message = fmt.Sprintf("Error loading security levels: %v", err)
						m.messageTime = time.Now()
					} else {
						m.navMode = SecurityLevelsMode
						m.message = ""
					}
				case "menu-editor":
					// Launch menu management interface
					// Check if database path is configured
					if m.config.Configuration.Paths.Database == "" {
						m.message = "Database path not configured. Please set it under Configuration > Paths > Database first."
						m.messageTime = time.Now()
						return m, nil
					}

					// Try to get database connection if not already available
					if m.db == nil {
						if existingDB := config.GetDatabase(); existingDB != nil {
							if sqliteDB, ok := existingDB.(*database.SQLiteDB); ok {
								m.db = sqliteDB
								// Ensure menu schema is initialized
								if err := m.db.InitializeSchema(); err != nil {
									m.message = fmt.Sprintf("Failed to initialize database schema: %v", err)
									m.messageTime = time.Now()
									return m, nil
								}
							} else {
								m.message = "Database connection type mismatch"
								m.messageTime = time.Now()
								return m, nil
							}
						} else {
							m.message = "No database connection available"
							m.messageTime = time.Now()
							return m, nil
						}
					}

					// Load menus
					if err := m.loadMenus(); err != nil {
						m.message = fmt.Sprintf("Error loading menus: %v", err)
						m.messageTime = time.Now()
					} else {
						m.navMode = MenuManagementMode
						m.message = ""
					}
				default:
					m.message = fmt.Sprintf("Action '%s' not implemented yet", item.submenuItem.Label)
					m.messageTime = time.Now()
				}
			}
		}
	case "esc":
		// Return to main menu navigation
		m.navMode = MainMenuNavigation
		m.message = ""
	}
	return m, nil
}

// handleLevel3MenuNavigation processes input while in Level 3 menu navigation mode
func (m Model) handleLevel3MenuNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Navigate up in field list
		if m.modalFieldIndex > 0 {
			m.modalFieldIndex--
		}
		m.message = ""
		return m, nil
	case "down", "j":
		// Navigate down in field list
		if m.modalFieldIndex < len(m.modalFields)-1 {
			m.modalFieldIndex++
		}
		m.message = ""
		return m, nil
	case "home":
		// Jump to first field
		m.modalFieldIndex = 0
		m.message = ""
		return m, nil
	case "end":
		// Jump to last field
		if len(m.modalFields) > 0 {
			m.modalFieldIndex = len(m.modalFields) - 1
		}
		m.message = ""
		return m, nil
	case "enter":
		// Edit selected field - go to Level 4 modal
		selectedField := m.modalFields[m.modalFieldIndex]
		if selectedField.ItemType == EditableField && selectedField.EditableItem != nil {
			m.editingItem = selectedField.EditableItem
			m.originalValue = m.editingItem.Field.GetValue()
			m.editingError = ""

			// Initialize text input based on value type
			if m.editingItem.ValueType == BoolValue {
				// For bool, we don't use text input, just toggle
				m.textInput.SetValue("")
			} else {
				// Set current value as text
				currentValue := m.editingItem.Field.GetValue()
				m.setTextInputValueWithCursor(formatValue(currentValue, m.editingItem.ValueType))

				// Set placeholder, char limit and width based on type
				switch m.editingItem.ValueType {
				case PortValue:
					m.textInput.Placeholder = "1-65535"
					m.textInput.CharLimit = 5
					m.textInput.Width = 10
				case IntValue:
					m.textInput.Placeholder = "Enter number"
					m.textInput.CharLimit = 10
					m.textInput.Width = 15
				case PathValue:
					m.textInput.Placeholder = "Enter path"
					m.textInput.CharLimit = 200
					m.textInput.Width = 49 // Fixed width for scrolling
				case ListValue:
					m.textInput.Placeholder = "comma,separated,values"
					m.textInput.CharLimit = 200
					m.textInput.Width = 49 // Fixed width for scrolling
				default: // StringValue
					m.textInput.Placeholder = "Enter value"
					m.textInput.CharLimit = 200
					m.textInput.Width = 49 // Fixed width for scrolling
				}
			}

			m.textInput.Focus()
			// Add this in each place where you enter EditingValue mode:
			m.textInput.TextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.Cursor.Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright))
			m.textInput.PromptStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.PlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim)).
				Background(lipgloss.Color(ColorPrimary))
			m.navMode = Level4ModalNavigation
			m.message = ""
		}
		return m, nil
	case "esc":
		// Return to main menu navigation
		m.navMode = MainMenuNavigation
		m.message = ""
		m.modalFields = nil
		m.modalFieldIndex = 0
		return m, nil
	}
	return m, nil
}

// handleSavePrompt processes input during save confirmation
func (m Model) handleSavePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h", "right", "l", "tab":
		// Toggle between Yes and No
		m.savePromptSelection = (m.savePromptSelection + 1) % 2
		return m, nil
	case "up", "k":
		// Move to Yes (1)
		m.savePromptSelection = 1
		return m, nil
	case "down", "j":
		// Move to No (0)
		m.savePromptSelection = 0
		return m, nil
	case "enter":
		// Execute based on selection
		if m.savePromptSelection == 1 {
			// Yes - Save and quit
			if err := config.SaveConfig(m.config, ""); err != nil {
				m.message = fmt.Sprintf("Error saving: %v", err)
				m.savePrompt = false
				return m, nil
			}
			m.message = "Configuration saved to database" // Show ONLY this message
			m.quitting = true
			return m, tea.Quit
		} else {
			// No - Don't exit, return to main menu
			m.savePrompt = false
			m.navMode = MainMenuNavigation
			m.message = ""
			return m, nil
		}
	case "y", "Y":
		// Shortcut for Yes
		if err := config.SaveConfig(m.config, ""); err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
			m.savePrompt = false
			return m, nil
		}
		m.message = "Configuration saved to database" // Show ONLY this message
		m.quitting = true
		return m, tea.Quit
	case "n", "N":
		// Shortcut for No - don't exit, return to main menu
		m.savePrompt = false
		m.navMode = MainMenuNavigation
		m.message = ""
		return m, nil
	case "esc":
		// Cancel - return to main menu
		m.savePrompt = false
		m.navMode = MainMenuNavigation
		m.message = ""
	}
	return m, nil
}

// Update renderSavePrompt to show appropriate warning
func (m Model) renderSavePrompt() string {
	modalWidth := 50

	// Create header
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPrimary)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(modalWidth).
		Align(lipgloss.Center)

	var header string
	var promptMsg string

	if m.navMode == SaveChangesPrompt {
		if m.editingMenu != nil {
			// Saving menu changes
			header = headerStyle.Render("[ Save Menu Changes? ]")
			if m.menuModified {
				promptMsg = "Save changes to menu and commands?"
			} else {
				promptMsg = "Exit without changes?"
			}
		} else {
			// Saving other changes (user, security level, etc)
			header = headerStyle.Render(fmt.Sprintf("[ Save Changes? (%d) ]", m.modifiedCount))
			promptMsg = "Save changes before exiting?"
		}
	} else {
		// Exiting application
		if m.modifiedCount > 0 {
			header = headerStyle.Render(fmt.Sprintf("[ Exit Config? (%d changes) ]", m.modifiedCount))
			promptMsg = "You have unsaved changes. Save before exiting?"
		} else {
			header = headerStyle.Render("[ Exit Config? ]")
			promptMsg = "Exit configuration editor?"
		}
	}

	// Create separator style
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(modalWidth)

	separator := separatorStyle.Render(strings.Repeat("-", modalWidth))

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextNormal)).
		Background(lipgloss.Color(ColorBgMedium)).
		Width(modalWidth).
		Align(lipgloss.Center).
		Padding(1, 0)

	promptLine := promptStyle.Render(promptMsg)

	// Create Yes/No options with lightbar
	var yesOption, noOption string

	if m.savePromptSelection == 1 {
		// Yes is selected
		yesOption = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextBright)).
			Background(lipgloss.Color(ColorAccent)).
			Bold(true).
			Padding(0, 2).
			Render("Yes")
		noOption = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextNormal)).
			Background(lipgloss.Color(ColorBgMedium)).
			Padding(0, 2).
			Render("No")
	} else {
		// No is selected
		yesOption = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextNormal)).
			Background(lipgloss.Color(ColorBgMedium)).
			Padding(0, 2).
			Render("Yes")
		noOption = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextBright)).
			Background(lipgloss.Color(ColorAccent)).
			Bold(true).
			Padding(0, 2).
			Render("No")
	}

	// Center the options
	optionsLine := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Width(modalWidth).
		Align(lipgloss.Center).
		Render(yesOption + noOption)

	// Create footer separator
	// footer separator defined below with correct width variable

	// ASCII-safe header and separators
	headerASCII := headerStyle.Render("[ Exit Config? ]")
	if m.modifiedCount > 0 {
		headerASCII = headerStyle.Render(fmt.Sprintf("[ Exit Config? (%d changes) ]", m.modifiedCount))
	}
	separatorASCII := separatorStyle.Render(strings.Repeat("-", modalWidth))
	footerSeparatorASCII := separatorStyle.Render(strings.Repeat("-", modalWidth))
	// Avoid unused original vars
	_ = header
	_ = separator
	// no footerSeparator in this variant

	// Build the modal (ASCII-safe)
	allLines := []string{headerASCII, separatorASCII, promptLine, optionsLine, footerSeparatorASCII}
	combined := strings.Join(allLines, "\n")

	// Wrap with background
	modalBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return modalBox
}

// handleUserManagement processes input in user management mode
func (m Model) handleUserManagement(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "up", "k":
		// Navigate up in list
		currentIdx := m.userListUI.Index()
		if currentIdx > 0 {
			m.userListUI.Select(currentIdx - 1)
		}
		return m, nil
	case "down", "j":
		// Navigate down in list
		currentIdx := m.userListUI.Index()
		items := m.userListUI.Items()
		if currentIdx < len(items)-1 {
			m.userListUI.Select(currentIdx + 1)
		}
		return m, nil
	case "home":
		// Jump to first item
		m.userListUI.Select(0)
		return m, nil
	case "end":
		// Jump to last item
		items := m.userListUI.Items()
		if len(items) > 0 {
			m.userListUI.Select(len(items) - 1)
		}
		return m, nil
	case "enter":
		// Select user - open edit modal
		selectedItem := m.userListUI.SelectedItem()
		if selectedItem != nil {
			user := selectedItem.(userListItem)
			// Create modal fields for editing user
			m.modalFields = []SubmenuItem{
				{
					ID:       "user-username",
					Label:    "Username",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "user-username",
						Label:     "Username",
						ValueType: StringValue,
						Field: ConfigField{
							GetValue: func() interface{} { return user.user.Username },
							SetValue: func(v interface{}) error {
								user.user.Username = v.(string)
								return nil
							},
						},
						HelpText: "User login name",
					},
				},
				{
					ID:       "user-first-name",
					Label:    "First Name",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "user-first-name",
						Label:     "First Name",
						ValueType: StringValue,
						Field: ConfigField{
							GetValue: func() interface{} {
								if user.user.FirstName.Valid {
									return user.user.FirstName.String
								}
								return ""
							},
							SetValue: func(v interface{}) error {
								s := v.(string)
								if s == "" {
									user.user.FirstName = sql.NullString{Valid: false}
								} else {
									user.user.FirstName = sql.NullString{String: s, Valid: true}
								}
								return nil
							},
						},
						HelpText: "User's first name",
					},
				},
				{
					ID:       "user-last-name",
					Label:    "Last Name",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "user-last-name",
						Label:     "Last Name",
						ValueType: StringValue,
						Field: ConfigField{
							GetValue: func() interface{} {
								if user.user.LastName.Valid {
									return user.user.LastName.String
								}
								return ""
							},
							SetValue: func(v interface{}) error {
								s := v.(string)
								if s == "" {
									user.user.LastName = sql.NullString{Valid: false}
								} else {
									user.user.LastName = sql.NullString{String: s, Valid: true}
								}
								return nil
							},
						},
						HelpText: "User's last name",
					},
				},
				{
					ID:       "user-security-level",
					Label:    "Security Level",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "user-security-level",
						Label:     "Security Level",
						ValueType: IntValue,
						Field: ConfigField{
							GetValue: func() interface{} { return user.user.SecurityLevel },
							SetValue: func(v interface{}) error {
								user.user.SecurityLevel = v.(int)
								return nil
							},
						},
						HelpText: "User's security level (0-255)",
						Validation: func(v interface{}) error {
							level := v.(int)
							if level < 0 || level > 255 {
								return fmt.Errorf("security level must be between 0 and 255")
							}
							return nil
						},
					},
				},
				{
					ID:       "user-email",
					Label:    "Email",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "user-email",
						Label:     "Email",
						ValueType: StringValue,
						Field: ConfigField{
							GetValue: func() interface{} {
								if user.user.Email.Valid {
									return user.user.Email.String
								}
								return ""
							},
							SetValue: func(v interface{}) error {
								s := v.(string)
								if s == "" {
									user.user.Email = sql.NullString{Valid: false}
								} else {
									user.user.Email = sql.NullString{String: s, Valid: true}
								}
								return nil
							},
						},
						HelpText: "User's email address (optional)",
					},
				},
				{
					ID:       "user-locations",
					Label:    "Locations",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "user-locations",
						Label:     "Locations",
						ValueType: StringValue,
						Field: ConfigField{
							GetValue: func() interface{} {
								if user.user.Locations.Valid {
									return user.user.Locations.String
								}
								return ""
							},
							SetValue: func(v interface{}) error {
								s := v.(string)
								if s == "" {
									user.user.Locations = sql.NullString{Valid: false}
								} else {
									user.user.Locations = sql.NullString{String: s, Valid: true}
								}
								return nil
							},
						},
						HelpText: "User's locations (optional)",
					},
				},
			}
			m.modalFieldIndex = 0
			m.modalSectionName = fmt.Sprintf("Edit User: %s (%d)", user.user.Username, user.user.ID)
			m.editingUser = &user.user // Store reference for saving
			m.navMode = Level4ModalNavigation
			m.message = ""
		}
		return m, nil
	case "esc":
		// Return to Level 2 menu navigation (Editors menu)
		m.navMode = Level2MenuNavigation
		m.message = ""
	}

	// Update the list for any other input (like filtering)
	// Only update if not handled above and not ESC (to prevent quit command)
	if msg.String() != "esc" {
		m.userListUI, cmd = m.userListUI.Update(msg)
	}
	return m, cmd
}

// In handleMenuModify - fix the ENTER, I, and D key handlers
func (m Model) handleMenuModify(msg tea.KeyMsg) (tea.Model, tea.Cmd) {

	switch msg.String() {
	case "tab":
		// Switch between Commands and Menu Data tabs
		m.currentMenuTab = (m.currentMenuTab + 1) % 2
		// Reset selection when switching tabs
		m.selectedCommandIndex = 0
		return m, nil
	case "1":
		// Switch to Menu Data tab
		m.currentMenuTab = 0
		m.selectedCommandIndex = 0
		return m, nil
	case "2":
		// Switch to Menu Commands tab
		m.currentMenuTab = 1
		m.selectedCommandIndex = 0
		return m, nil

	case "up", "k":
		// Navigate up in current tab
		if m.currentMenuTab == 0 {
			// Menu Data tab - navigate through menu data fields
			if len(m.modalFields) > 0 {
				m.selectedCommandIndex = (m.selectedCommandIndex - 1 + len(m.modalFields)) % len(m.modalFields)
			}
		} else {
			// Commands tab
			if len(m.menuCommandsList) > 0 {
				m.selectedCommandIndex = (m.selectedCommandIndex - 1 + len(m.menuCommandsList)) % len(m.menuCommandsList)
			}
		}
		return m, nil
	case "down", "j":
		// Navigate down in current tab
		if m.currentMenuTab == 0 {
			// Menu Data tab - navigate through menu data fields
			if len(m.modalFields) > 0 {
				m.selectedCommandIndex = (m.selectedCommandIndex + 1) % len(m.modalFields)
			}
		} else {
			// Commands tab
			if len(m.menuCommandsList) > 0 {
				m.selectedCommandIndex = (m.selectedCommandIndex + 1) % len(m.menuCommandsList)
			}
		}
		return m, nil
	case "home":
		// Jump to first item in current tab
		m.selectedCommandIndex = 0
		return m, nil
	case "end":
		// Jump to last item in current tab
		if m.currentMenuTab == 0 {
			// Menu Data tab
			if len(m.modalFields) > 0 {
				m.selectedCommandIndex = len(m.modalFields) - 1
			}
		} else {
			// Commands tab
			if len(m.menuCommandsList) > 0 {
				m.selectedCommandIndex = len(m.menuCommandsList) - 1
			}
		}
		return m, nil
		// In handleMenuModify - when entering command edit, don't overwrite modalFields
	case "enter":
		// Edit selected item based on current tab
		if m.currentMenuTab == 0 {
			// Menu Data tab - edit field INLINE
			if len(m.modalFields) > 0 && m.selectedCommandIndex < len(m.modalFields) {
				selectedField := m.modalFields[m.selectedCommandIndex]
				if selectedField.ItemType == EditableField && selectedField.EditableItem != nil {
					m.editingItem = selectedField.EditableItem
					m.originalValue = m.editingItem.Field.GetValue()
					m.editingError = ""
					m.modalFieldIndex = m.selectedCommandIndex // Sync the index

					// Initialize text input based on value type
					if m.editingItem.ValueType == BoolValue {
						m.textInput.SetValue("")
					} else {
						currentValue := m.editingItem.Field.GetValue()
						m.setTextInputValueWithCursor(formatValue(currentValue, m.editingItem.ValueType))

						// Set placeholder and limits based on type
						switch m.editingItem.ValueType {
						case IntValue:
							m.textInput.Placeholder = "Enter number"
							m.textInput.CharLimit = 10
							m.textInput.Width = 15
						case ListValue:
							m.textInput.Placeholder = "comma,separated,values"
							m.textInput.CharLimit = 200
							m.textInput.Width = 50 // Fixed width for scrolling
						default: // StringValue
							m.textInput.Placeholder = "Enter value"
							m.textInput.CharLimit = 200
							m.textInput.Width = 50
						}
					}

					m.textInput.Focus()
					m.navMode = EditingValue
					m.textInput.CursorEnd() // Move cursor to end
					m.textInput.Focus()     // Ensure it's focused
					// Add this in each place where you enter EditingValue mode:
					m.textInput.TextStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color(ColorTextBright)).
						Background(lipgloss.Color(ColorPrimary))
					m.textInput.Cursor.Style = lipgloss.NewStyle().
						Foreground(lipgloss.Color(ColorTextBright))
					m.textInput.PromptStyle = lipgloss.NewStyle().
						Background(lipgloss.Color(ColorPrimary))
					m.textInput.PlaceholderStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color(ColorTextDim)).
						Background(lipgloss.Color(ColorPrimary))
					m.message = ""
				}
			}
		} else {
			// Commands tab - edit selected command in MODAL
			if len(m.menuCommandsList) > 0 && m.selectedCommandIndex < len(m.menuCommandsList) {
				// Create a copy of the command for editing
				selectedCmd := m.menuCommandsList[m.selectedCommandIndex]
				m.editingMenuCommand = &database.MenuCommand{
					ID:               selectedCmd.ID,
					MenuID:           selectedCmd.MenuID,
					CommandNumber:    selectedCmd.CommandNumber,
					Keys:             selectedCmd.Keys,
					ShortDescription: selectedCmd.ShortDescription,
					ACSRequired:      selectedCmd.ACSRequired,
					CmdKeys:          selectedCmd.CmdKeys,
					Options:          selectedCmd.Options,
				}
				// Set up modal fields for command editing
				m.setupMenuEditCommandModal()
				m.navMode = MenuEditCommandMode
			}
		}
		return m, nil
	case "i", "I":
		// Insert new command (only works in Commands tab)
		if m.currentMenuTab == 1 { // FIXED: Was 0, should be 1 for commands tab
			m.editingMenuCommand = &database.MenuCommand{
				MenuID:           m.editingMenu.ID,
				CommandNumber:    len(m.menuCommandsList) + 1,
				Keys:             "",
				ShortDescription: "",
				ACSRequired:      "",
				CmdKeys:          "",
				Options:          "",
			}
			// Set up modal fields for command editing
			m.setupMenuEditCommandModal()
			m.navMode = MenuEditCommandMode
		}
		return m, nil
	case "d", "D":
		// Delete selected command (only works in Commands tab)
		if m.currentMenuTab == 1 && len(m.menuCommandsList) > 0 && m.selectedCommandIndex < len(m.menuCommandsList) { // FIXED: Was 0, should be 1
			selectedCmd := m.menuCommandsList[m.selectedCommandIndex]
			m.message = fmt.Sprintf("Delete command '%s'? (Y/N)", selectedCmd.Keys)
			m.messageTime = time.Now()
			m.messageType = WarningMessage
			m.confirmAction = "delete_command"
			m.confirmMenuID = int64(selectedCmd.ID)
		}
		return m, nil
	case "x", "X":
		// Switch to Menu Data tab
		m.currentMenuTab = 0
		m.selectedCommandIndex = 0
		return m, nil
	case "f1":
		// Show help
		m.message = "Help: Arrow keys Navigate | ENTER Edit | TAB Switch | I Insert | D Delete | ESC Back"
		m.messageTime = time.Now()
		m.messageType = InfoMessage
		return m, nil
		// Update handleMenuModify ESC handler to check for unsaved changes
	case "esc":
		// Check if there are unsaved changes
		if m.menuModified {
			m.savePrompt = true
			m.savePromptSelection = 1 // Default to Yes
			m.returnToMode = MenuManagementMode
			m.navMode = SaveChangesPrompt
			return m, nil
		}

		// No unsaved changes, exit normally
		m.navMode = MenuManagementMode
		m.editingMenu = nil
		m.originalMenu = nil
		m.menuCommandsList = nil
		m.originalMenuCommands = nil
		m.selectedCommandIndex = 0
		m.currentMenuTab = 0
		m.menuDataFields = nil
		return m, nil
	}

	// Handle confirmation for delete
	if m.confirmAction == "delete_command" && (msg.String() == "y" || msg.String() == "Y") {
		if err := m.db.DeleteMenuCommand(m.confirmMenuID); err != nil {
			m.message = fmt.Sprintf("Error deleting command: %v", err)
			m.messageTime = time.Now()
			m.messageType = ErrorMessage
		} else {
			// Reload commands
			if err := m.loadMenuCommandsForEditing(); err != nil {
				m.message = fmt.Sprintf("Error reloading commands: %v", err)
				m.messageTime = time.Now()
				m.messageType = ErrorMessage
			} else {
				m.message = "Command deleted successfully"
				m.messageTime = time.Now()
				m.messageType = SuccessMessage
				// Adjust selection if necessary
				if m.selectedCommandIndex >= len(m.menuCommandsList) && len(m.menuCommandsList) > 0 {
					m.selectedCommandIndex = len(m.menuCommandsList) - 1
				}
			}
		}
		m.confirmAction = ""
		m.confirmMenuID = 0
		return m, nil
	} else if m.confirmAction == "delete_command" && (msg.String() == "n" || msg.String() == "N" || msg.String() == "esc") {
		m.message = ""
		m.confirmAction = ""
		m.confirmMenuID = 0
		return m, nil
	}

	return m, nil
}

// handleMenuEditData processes input in menu edit data mode
func (m Model) handleMenuEditData(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Safety check at the top - if modalFields is empty, return to menu management
	if len(m.modalFields) == 0 {
		m.message = "Error: No fields to edit. Returning to menu list."
		m.messageTime = time.Now()
		m.messageType = ErrorMessage
		m.navMode = MenuManagementMode
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		// Navigate up in field list
		if m.modalFieldIndex > 0 {
			m.modalFieldIndex--
		}
		m.message = ""
		return m, nil
	case "down", "j":
		// Navigate down in field list
		if m.modalFieldIndex < len(m.modalFields)-1 {
			m.modalFieldIndex++
		}
		m.message = ""
		return m, nil
	case "enter":
		// Bounds check for modalFieldIndex
		if m.modalFieldIndex >= len(m.modalFields) {
			m.modalFieldIndex = len(m.modalFields) - 1
			return m, nil
		}

		// Edit selected field
		selectedField := m.modalFields[m.modalFieldIndex]
		if selectedField.ItemType == EditableField && selectedField.EditableItem != nil {
			m.editingItem = selectedField.EditableItem
			m.originalValue = m.editingItem.Field.GetValue()
			m.editingError = ""

			// Initialize text input based on value type
			if m.editingItem.ValueType == BoolValue {
				m.textInput.SetValue("")
			} else {
				currentValue := m.editingItem.Field.GetValue()
				m.setTextInputValueWithCursor(formatValue(currentValue, m.editingItem.ValueType))

				// Set placeholder and limits based on type
				switch m.editingItem.ValueType {
				case IntValue:
					m.textInput.Placeholder = "Enter number"
					m.textInput.CharLimit = 10
					m.textInput.Width = 15
				case ListValue:
					m.textInput.Placeholder = "comma,separated,values"
					m.textInput.CharLimit = 200
					m.textInput.Width = 49 // Fixed width for scrolling
				default: // StringValue
					m.textInput.Placeholder = "Enter value"
					m.textInput.CharLimit = 200
					m.textInput.Width = 49 // Fixed width for scrolling
				}
			}

			m.textInput.Focus()
			m.textInput.CursorEnd() // Move cursor to end
			m.textInput.Focus()     // Ensure it's focused
			// Add this in each place where you enter EditingValue mode:
			m.textInput.TextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.Cursor.Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright))
			m.textInput.PromptStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.PlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim)).
				Background(lipgloss.Color(ColorPrimary))
			m.navMode = EditingValue
			m.message = ""
		}
		return m, nil
	case "esc":
		// Check if there are unsaved changes
		hasUnsavedChanges := false
		// TODO: Implement change detection for menu data
		if hasUnsavedChanges {
			m.savePrompt = true
			m.savePromptSelection = 1
			m.returnToMode = MenuModifyMode
			m.navMode = SaveChangesPrompt
		} else {
			m.navMode = MenuModifyMode
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
		}
		m.message = ""
		return m, nil
	}
	return m, nil
}

// handleMenuEditCommand processes input in menu edit command mode
func (m Model) handleMenuEditCommand(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Use the same logic as Level4ModalNavigation for field editing
	switch msg.String() {
	case "up", "k":
		// Navigate up in field list
		if m.modalFieldIndex > 0 {
			m.modalFieldIndex--
		}
		m.message = ""
		return m, nil
	case "down", "j":
		// Navigate down in field list
		if m.modalFieldIndex < len(m.modalFields)-1 {
			m.modalFieldIndex++
		}
		m.message = ""
		return m, nil
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
			} else {
				currentValue := m.editingItem.Field.GetValue()
				m.setTextInputValueWithCursor(formatValue(currentValue, m.editingItem.ValueType))

				// Set placeholder and limits based on type
				switch m.editingItem.ValueType {
				case IntValue:
					m.textInput.Placeholder = "Enter number"
					m.textInput.CharLimit = 10
					m.textInput.Width = 15
				case ListValue:
					m.textInput.Placeholder = "comma,separated,values"
					m.textInput.CharLimit = 200
					m.textInput.Width = 49 // Fixed width for scrolling
				default: // StringValue
					m.textInput.Placeholder = "Enter value"
					m.textInput.CharLimit = 200
					m.textInput.Width = 49 // Fixed width for scrolling
				}
			}

			m.textInput.Focus()
			// Add this in each place where you enter EditingValue mode:
			m.textInput.TextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.Cursor.Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright))
			m.textInput.PromptStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorPrimary))
			m.textInput.PlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim)).
				Background(lipgloss.Color(ColorPrimary))
			m.navMode = EditingValue
			m.textInput.CursorEnd() // Move cursor to end
			m.textInput.Focus()     // Ensure it's focused
			m.message = ""
		}
		return m, nil
	case "esc":
		// Restore menu data fields from backup
		if len(m.menuDataFields) > 0 {
			m.modalFields = make([]SubmenuItem, len(m.menuDataFields))
			copy(m.modalFields, m.menuDataFields)
		}

		// Return to MenuModifyMode
		m.navMode = MenuModifyMode
		m.editingMenuCommand = nil
		m.modalFieldIndex = 0
		m.modalSectionName = ""
		m.message = ""
		return m, nil
	}
	return m, nil
}

// handleSecurityLevelsManagement processes input in security levels management mode
func (m Model) handleSecurityLevelsManagement(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "up", "k":
		// Navigate up in list
		currentIdx := m.securityLevelsUI.Index()
		if currentIdx > 0 {
			m.securityLevelsUI.Select(currentIdx - 1)
		}
		return m, nil
	case "down", "j":
		// Navigate down in list
		currentIdx := m.securityLevelsUI.Index()
		items := m.securityLevelsUI.Items()
		if currentIdx < len(items)-1 {
			m.securityLevelsUI.Select(currentIdx + 1)
		}
		return m, nil
	case "home":
		// Jump to first item
		m.securityLevelsUI.Select(0)
		return m, nil
	case "end":
		// Jump to last item
		items := m.securityLevelsUI.Items()
		if len(items) > 0 {
			m.securityLevelsUI.Select(len(items) - 1)
		}
		return m, nil
	case "enter":
		// Select security level - open edit modal
		selectedItem := m.securityLevelsUI.SelectedItem()
		if selectedItem != nil {
			level := selectedItem.(securityLevelListItem)
			// Create modal fields for editing security level
			m.modalFields = []SubmenuItem{
				{
					ID:       "security-level-name",
					Label:    "Name",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-name",
						Label:     "Name",
						ValueType: StringValue,
						Field: ConfigField{
							GetValue: func() interface{} { return level.securityLevel.Name },
							SetValue: func(v interface{}) error {
								level.securityLevel.Name = v.(string)
								return nil
							},
						},
						HelpText: "Display name for this security level",
					},
				},
				{
					ID:       "security-level-mins-per-day",
					Label:    "Minutes per Day",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-mins-per-day",
						Label:     "Mins per Day",
						ValueType: IntValue,
						Field: ConfigField{
							GetValue: func() interface{} { return level.securityLevel.MinsPerDay },
							SetValue: func(v interface{}) error {
								level.securityLevel.MinsPerDay = v.(int)
								return nil
							},
						},
						HelpText: "Maximum minutes per day (0 = unlimited)",
						Validation: func(v interface{}) error {
							mins := v.(int)
							if mins < 0 {
								return fmt.Errorf("minutes per day must be non-negative")
							}
							return nil
						},
					},
				},
				{
					ID:       "security-level-timeout-mins",
					Label:    "Timeout Minutes",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-timeout-mins",
						Label:     "Timeout Mins",
						ValueType: IntValue,
						Field: ConfigField{
							GetValue: func() interface{} { return level.securityLevel.TimeoutMins },
							SetValue: func(v interface{}) error {
								level.securityLevel.TimeoutMins = v.(int)
								return nil
							},
						},
						HelpText: "Idle timeout in minutes",
						Validation: func(v interface{}) error {
							mins := v.(int)
							if mins < 0 {
								return fmt.Errorf("timeout minutes must be non-negative")
							}
							return nil
						},
					},
				},
				{
					ID:       "security-level-can-delete-own-msgs",
					Label:    "Delete Own Msgs",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-can-delete-own-msgs",
						Label:     "Delete Own Msg",
						ValueType: BoolValue,
						Field: ConfigField{
							GetValue: func() interface{} { return level.securityLevel.CanDeleteOwnMsgs },
							SetValue: func(v interface{}) error {
								level.securityLevel.CanDeleteOwnMsgs = v.(bool)
								return nil
							},
						},
						HelpText: "Allow users to delete their own messages",
					},
				},
				{
					ID:       "security-level-can-delete-msgs",
					Label:    "Delete Any Msgs",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-can-delete-msgs",
						Label:     "Delete Any Msg",
						ValueType: BoolValue,
						Field: ConfigField{
							GetValue: func() interface{} { return level.securityLevel.CanDeleteMsgs },
							SetValue: func(v interface{}) error {
								level.securityLevel.CanDeleteMsgs = v.(bool)
								return nil
							},
						},
						HelpText: "Allow users to delete any messages",
					},
				},
				{
					ID:       "security-level-invisible",
					Label:    "Invisible",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-invisible",
						Label:     "Invisible",
						ValueType: BoolValue,
						Field: ConfigField{
							GetValue: func() interface{} { return level.securityLevel.Invisible },
							SetValue: func(v interface{}) error {
								level.securityLevel.Invisible = v.(bool)
								return nil
							},
						},
						HelpText: "Hide user from user lists",
					},
				},
			}
			m.modalFieldIndex = 0
			m.modalSectionName = fmt.Sprintf("Security Level %d", level.securityLevel.SecLevel)
			m.editingSecurityLevel = &level.securityLevel // Store reference for saving
			m.navMode = Level4ModalNavigation
			m.message = ""
		}
		return m, nil
	case "esc":
		// Return to Level 2 menu navigation (Editors menu)
		m.navMode = Level2MenuNavigation
		m.message = ""
	}

	// Update the list for any other input (like filtering)
	// Only update if not handled above and not ESC (to prevent quit command)
	if msg.String() != "esc" {
		m.securityLevelsUI, cmd = m.securityLevelsUI.Update(msg)
	}
	return m, cmd
}

// handleMenuManagement processes input in menu management mode
func (m Model) handleMenuManagement(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "up", "k":
		// Navigate up in list
		currentIdx := m.menuListUI.Index()
		if currentIdx > 0 {
			m.menuListUI.Select(currentIdx - 1)
		}
		return m, nil
	case "down", "j":
		// Navigate down in list
		currentIdx := m.menuListUI.Index()
		items := m.menuListUI.Items()
		if currentIdx < len(items)-1 {
			m.menuListUI.Select(currentIdx + 1)
		}
		return m, nil
	case "home":
		// Jump to first item
		m.menuListUI.Select(0)
		return m, nil
	case "end":
		// Jump to last item
		items := m.menuListUI.Items()
		if len(items) > 0 {
			m.menuListUI.Select(len(items) - 1)
		}
		return m, nil
		// In handleMenuManagement - when entering MenuModifyMode
		// In handleMenuManagement - when entering MenuModifyMode, save the menu data fields
	case "enter", "m", "M":
		// Modify menu - open modify menu interface
		selectedItem := m.menuListUI.SelectedItem()
		if selectedItem != nil {
			menuItem := selectedItem.(menuListItem)

			// Make a DEEP COPY of the original menu
			originalMenu := menuItem.menu
			m.originalMenu = &database.Menu{
				ID:                  originalMenu.ID,
				Name:                originalMenu.Name,
				Titles:              append([]string{}, originalMenu.Titles...),
				Prompt:              originalMenu.Prompt,
				ACSRequired:         originalMenu.ACSRequired,
				GenericColumns:      originalMenu.GenericColumns,
				GenericBracketColor: originalMenu.GenericBracketColor,
				GenericCommandColor: originalMenu.GenericCommandColor,
				GenericDescColor:    originalMenu.GenericDescColor,
				ClearScreen:         originalMenu.ClearScreen,
			}

			// Make a working copy
			m.editingMenu = &database.Menu{
				ID:                  originalMenu.ID,
				Name:                originalMenu.Name,
				Titles:              append([]string{}, originalMenu.Titles...),
				Prompt:              originalMenu.Prompt,
				ACSRequired:         originalMenu.ACSRequired,
				GenericColumns:      originalMenu.GenericColumns,
				GenericBracketColor: originalMenu.GenericBracketColor,
				GenericCommandColor: originalMenu.GenericCommandColor,
				GenericDescColor:    originalMenu.GenericDescColor,
				ClearScreen:         originalMenu.ClearScreen,
			}

			if err := m.loadMenuCommandsForEditing(); err != nil {
				m.message = fmt.Sprintf("Error loading menu commands: %v", err)
				m.messageTime = time.Now()
				m.messageType = ErrorMessage
				return m, nil
			}

			// Make a DEEP COPY of original commands
			m.originalMenuCommands = make([]database.MenuCommand, len(m.menuCommandsList))
			for i, cmd := range m.menuCommandsList {
				m.originalMenuCommands[i] = database.MenuCommand{
					ID:               cmd.ID,
					MenuID:           cmd.MenuID,
					CommandNumber:    cmd.CommandNumber,
					Keys:             cmd.Keys,
					ShortDescription: cmd.ShortDescription,
					ACSRequired:      cmd.ACSRequired,
					CmdKeys:          cmd.CmdKeys,
					Options:          cmd.Options,
				}
			}

			// Set up modal fields for menu data editing
			m.setupMenuEditDataModal()
			// SAVE the menu data fields for later restoration
			m.menuDataFields = make([]SubmenuItem, len(m.modalFields))
			copy(m.menuDataFields, m.modalFields)

			m.menuModified = false // Reset modification flag
			m.navMode = MenuModifyMode
			m.selectedCommandIndex = 0
			m.currentMenuTab = 0
			m.message = ""
		}
		return m, nil
	case "i", "I":
		// Insert new menu
		m.editingMenu = &database.Menu{
			Name:                "",
			Titles:              []string{"", ""},
			Prompt:              "",
			ACSRequired:         "",
			GenericColumns:      4,
			GenericBracketColor: 1,
			GenericCommandColor: 9,
			GenericDescColor:    1,
			ClearScreen:         false,
		}
		// CRITICAL: Set up modal fields for new menu editing
		// ADD THIS: Initialize modal fields for the new menu
		m.setupMenuEditDataModal()
		m.modalFieldIndex = 0

		m.navMode = MenuEditDataMode
		m.message = "Creating new menu"
		m.messageTime = time.Now()
		m.messageType = InfoMessage
		return m, nil
	case "d", "D":
		// Delete menu - confirm first
		selectedItem := m.menuListUI.SelectedItem()
		if selectedItem != nil {
			menuItem := selectedItem.(menuListItem)
			m.message = fmt.Sprintf("Delete menu '%s'? (Y/N)", menuItem.menu.Name)
			m.messageTime = time.Now()
			m.messageType = WarningMessage
			m.confirmAction = "delete_menu"
			m.confirmMenuID = int64(menuItem.menu.ID)
		}
		return m, nil
	case "q", "Q", "esc":
		// Return to Level 2 menu navigation (Editors menu)
		m.navMode = Level2MenuNavigation
		m.message = ""
	}

	// Handle confirmation for delete
	if m.confirmAction == "delete_menu" && (msg.String() == "y" || msg.String() == "Y") {
		if err := m.db.DeleteMenu(m.confirmMenuID); err != nil {
			m.message = fmt.Sprintf("Error deleting menu: %v", err)
			m.messageTime = time.Now()
			m.messageType = ErrorMessage
		} else {
			// Reload menu list
			if err := m.loadMenus(); err != nil {
				m.message = fmt.Sprintf("Error reloading menus: %v", err)
				m.messageTime = time.Now()
				m.messageType = ErrorMessage
			} else {
				m.message = "Menu deleted successfully"
				m.messageTime = time.Now()
				m.messageType = SuccessMessage
			}
		}
		m.confirmAction = ""
		m.confirmMenuID = 0
		return m, nil
	} else if m.confirmAction == "delete_menu" && (msg.String() == "n" || msg.String() == "N" || msg.String() == "esc") {
		m.message = ""
		m.confirmAction = ""
		m.confirmMenuID = 0
		return m, nil
	}

	// Update the list for any other input (like filtering)
	// Only update if not handled above and not ESC (to prevent quit command)
	if msg.String() != "esc" && msg.String() != "q" && msg.String() != "Q" {
		m.menuListUI, cmd = m.menuListUI.Update(msg)
	}
	return m, cmd
}

func (m Model) handleEditingValue(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle bool toggle separately
	if m.editingItem.ValueType == BoolValue {
		switch msg.String() {
		case "y", "Y":
			newValue := true
			if newValue != m.originalValue {
				if err := m.editingItem.Field.SetValue(newValue); err != nil {
					m.editingError = err.Error()
				} else {
					m.modifiedCount++
					if m.editingMenu != nil {
						m.menuModified = true // Mark menu as modified
					}
					m.editingError = ""
					m.message = ""
					m.navMode = m.returnToMenuModifyOrModal()
				}
			} else {
				// No change - just exit without incrementing counter
				m.navMode = m.returnToMenuModifyOrModal()
			}
			return m, nil
		case "n", "N":
			newValue := false
			if newValue != m.originalValue {
				if err := m.editingItem.Field.SetValue(newValue); err != nil {
					m.editingError = err.Error()
				} else {
					m.modifiedCount++
					if m.editingMenu != nil {
						m.menuModified = true // Mark menu as modified
					}
					m.editingError = ""
					m.message = ""
					m.navMode = m.returnToMenuModifyOrModal()
				}
			} else {
				// No change - just exit without incrementing counter
				m.navMode = m.returnToMenuModifyOrModal()
			}
			return m, nil
		case "enter", " ", "tab":
			currentValue := m.editingItem.Field.GetValue().(bool)
			if err := m.editingItem.Field.SetValue(!currentValue); err != nil {
				m.editingError = err.Error()
			} else {
				// This is a toggle, so it's always a change
				m.modifiedCount++
				if m.editingMenu != nil {
					m.menuModified = true
				}
				m.editingError = ""
			}
			return m, nil
		case "esc":
			m.editingItem.Field.SetValue(m.originalValue)
			m.navMode = m.returnToMenuModifyOrModal()
			m.editingError = ""
			m.message = ""
			return m, nil
		}
		return m, nil
	}

	// Handle text input for other types
	switch msg.String() {
	case "enter":
		// Validate and save
		input := m.textInput.Value()

		// Parse value based on type
		parsedValue, err := parseValue(input, m.editingItem.ValueType)
		if err != nil {
			m.editingError = fmt.Sprintf("Invalid input: %v", err)
			return m, nil
		}

		// Run type-specific validation
		if m.editingItem.Validation != nil {
			if err := m.editingItem.Validation(parsedValue); err != nil {
				m.editingError = err.Error()
				return m, nil
			}
		}

		// Additional built-in validations
		switch m.editingItem.ValueType {
		case PortValue:
			port := parsedValue.(int)
			if port < 1 || port > 65535 {
				m.editingError = "Port must be between 1 and 65535"
				return m, nil
			}
		case IntValue:
			num := parsedValue.(int)
			if num < 0 {
				m.editingError = "Value must be positive"
				return m, nil
			}
		case StringValue, PathValue:
			str := parsedValue.(string)
			if strings.TrimSpace(str) == "" {
				m.editingError = "Value cannot be empty"
				return m, nil
			}
		}

		// Check if value actually changed before saving
		valueChanged := false
		switch m.editingItem.ValueType {
		case ListValue:
			// For list values, compare the formatted strings
			originalStr := formatValue(m.originalValue, m.editingItem.ValueType)
			newStr := formatValue(parsedValue, m.editingItem.ValueType)
			valueChanged = originalStr != newStr
		default:
			// For other types, direct comparison
			valueChanged = parsedValue != m.originalValue
		}

		// Only save and increment counter if value actually changed
		if valueChanged {
			if err := m.editingItem.Field.SetValue(parsedValue); err != nil {
				m.editingError = fmt.Sprintf("Error saving: %v", err)
				return m, nil
			}

			// Success - increment counters
			m.modifiedCount++
			if m.editingMenu != nil {
				m.menuModified = true // Mark menu as modified
			}
		}

		// Either way, exit editing mode
		m.editingError = ""
		m.message = ""
		m.navMode = m.returnToMenuModifyOrModal()
		m.textInput.Blur()

	case "esc":
		// Cancel editing - restore original value
		m.editingItem.Field.SetValue(m.originalValue)
		m.navMode = m.returnToMenuModifyOrModal()
		m.editingError = ""
		m.message = ""
		m.textInput.Blur()
	}

	return m, nil
}

// Update this helper function
func (m Model) returnToMenuModifyOrModal() NavigationMode {
	// If we're editing a menu command, return to command edit mode
	if m.editingMenuCommand != nil {
		return MenuEditCommandMode
	}
	// If we have editingMenu set, we're in MenuModifyMode
	if m.editingMenu != nil {
		return MenuModifyMode
	}
	// Otherwise return to Level4ModalNavigation
	return Level4ModalNavigation
}
