package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// View Rendering with Layered Approach
// ============================================================================

// View renders the complete UI using a layered canvas approach
func (m Model) View() string {
	// Create a base canvas with texture pattern
	canvas := make([]string, m.screenHeight)

	for i := range canvas {
		// Plain background (no box/pattern fill); ANSI art will overlay
		canvas[i] = strings.Repeat(" ", m.screenWidth)
	}

	// ALWAYS render persistent header at top (row 0)
	persistentHeader := m.renderPersistentHeader()
	m.overlayString(canvas, persistentHeader, 0, 0)

	// Render ANSI art background only on the main menu
	if m.navMode == MainMenuNavigation && len(m.ansiArtLines) > 0 {
		artWidth := 80
		startRow := 2 // just under the persistent header
		startCol := 0
		if m.screenWidth > artWidth {
			startCol = (m.screenWidth - artWidth) / 2
		}
		m.overlayArtBlock(canvas, m.ansiArtLines, startRow, startCol, artWidth)
	}

	// Save prompt overlay - highest priority
	if m.savePrompt {
		promptStr := m.renderSavePrompt()
		m.overlayStringCenteredWithClear(canvas, promptStr)
		return m.canvasToString(canvas)
	}

	if m.quitting {
		if m.message != "" {
			messageBox := lipgloss.NewStyle().
				Padding(1, 2).
				Render(m.message)
			m.overlayStringCenteredWithClear(canvas, messageBox)
		}
		return m.canvasToString(canvas)
	}

	// Determine if we should render as modal
	showAsModal := m.shouldRenderAsModal()

	// Layer 1: Main Menu (only shown when in MainMenuNavigation mode)
	if m.navMode == MainMenuNavigation {
		mainMenuStr := m.renderMainMenu()

		// If main-menu anchor flags are enabled, position bottom/right using offsets; otherwise center
		if MainAnchorBottom || MainAnchorRight {
			lines := strings.Split(mainMenuStr, "\n")
			height := len(lines)
			width := 0
			for _, line := range lines {
				if w := m.visualWidth(line); w > width {
					width = w
				}
			}

			row := (m.screenHeight - height) / 2
			col := (m.screenWidth - width) / 2

			if MainAnchorBottom {
				row = m.screenHeight - height - MainBottomOffset
			}
			if MainAnchorRight {
				col = m.screenWidth - width - MainRightOffset
			}
			if row < 2 {
				row = 2 // keep clear of the header
			}
			if col < 0 {
				col = 0
			}
			m.overlayStringWithBorderClear(canvas, mainMenuStr, row, col)
		} else {
			m.overlayStringCenteredWithClear(canvas, mainMenuStr)
		}
	}

	// Layer 1.5: User Management (full screen mode)
	if m.navMode == UserManagementMode {
		userManagementStr := m.renderUserManagement()
		m.overlayStringCenteredWithClear(canvas, userManagementStr)
		return m.canvasToString(canvas)
	}

	// Layer 1.6: Security Levels Management (full screen mode)
	if m.navMode == SecurityLevelsMode {
		securityLevelsStr := m.renderSecurityLevelsManagement()
		m.overlayStringCenteredWithClear(canvas, securityLevelsStr)
		return m.canvasToString(canvas)
	}

	// Layer 1.7: Menu Management (full screen mode)
	if m.navMode == MenuManagementMode {
		menuManagementStr := m.renderMenuManagement()
		m.overlayStringCentered(canvas, menuManagementStr)

		// Add footer BEFORE returning
		footer := m.renderFooter()
		m.overlayString(canvas, footer, m.screenHeight-1, 0)

		return m.canvasToString(canvas)
	}

	// Layer 1.8: Menu Modify (full screen mode) - but NOT when editing command
	if m.navMode == MenuModifyMode || (m.navMode == EditingValue && m.editingMenu != nil && m.editingMenuCommand == nil) {
		menuModifyStr := m.renderMenuModify()
		m.overlayStringCenteredWithClear(canvas, menuModifyStr)

		// Add breadcrumb when editing in menu modify mode
		if m.navMode == EditingValue && m.editingMenu != nil {
			breadcrumb := m.renderMenuModifyBreadcrumb()
			m.overlayString(canvas, breadcrumb, m.screenHeight-3, 0)
		}

		// Add footer BEFORE returning
		footer := m.renderFooter()
		m.overlayString(canvas, footer, m.screenHeight-1, 0)

		return m.canvasToString(canvas)
	}

	// Layer 1.10: Menu Edit Command (modal overlay)
	if m.navMode == MenuEditCommandMode {
		menuEditCommandStr := m.renderMenuEditCommand()
		m.overlayStringCenteredWithClear(canvas, menuEditCommandStr)

		// Add breadcrumb when editing command
		breadcrumb := m.renderCommandEditBreadcrumb()
		m.overlayString(canvas, breadcrumb, m.screenHeight-3, 0)

		// Add footer BEFORE returning
		footer := m.renderFooter()
		m.overlayString(canvas, footer, m.screenHeight-1, 0)

		return m.canvasToString(canvas)
	}

	// Layer 1.11: Editing a command field
	if m.navMode == EditingValue && m.editingMenuCommand != nil {
		// Still show the modal underneath, but with editing state
		menuEditCommandStr := m.renderMenuEditCommand()
		m.overlayStringCenteredWithClear(canvas, menuEditCommandStr)

		// Add breadcrumb when editing command field
		breadcrumb := m.renderCommandEditBreadcrumb()
		m.overlayString(canvas, breadcrumb, m.screenHeight-3, 0)

		// Add footer BEFORE returning
		footer := m.renderFooter()
		m.overlayString(canvas, footer, m.screenHeight-1, 0)

		return m.canvasToString(canvas)
	}

	// Layer 2: Submenu (centered when visible and not showing modal)
	if m.navMode >= Level2MenuNavigation && !showAsModal && m.navMode != MenuEditCommandMode {
		isDimmed := m.navMode > Level2MenuNavigation
		level2Str := m.renderLevel2Menu(isDimmed)
		m.overlayStringCenteredWithClear(canvas, level2Str)
	}

	// Layer 3: Field list (centered when visible and not showing modal)
	if m.navMode == Level3MenuNavigation && !showAsModal && m.navMode != MenuEditCommandMode {
		level3Str := m.renderLevel3Menu(false)
		m.overlayStringCenteredWithClear(canvas, level3Str)
	}

	// Modal: Show when we have editable fields (Level 3 with fields, or Level 4)
	if showAsModal {
		modalStr := m.renderModalForm()
		m.overlayStringCenteredWithClear(canvas, modalStr)
	}

	// Add breadcrumb near bottom (row screenHeight - 3)
	if m.navMode > MainMenuNavigation {
		breadcrumb := m.renderBreadcrumb()
		m.overlayString(canvas, breadcrumb, m.screenHeight-3, 0)
	}

	// Add status message if present
	if m.message != "" && time.Since(m.messageTime) < 3*time.Second && !m.savePrompt && !m.quitting {
		msgStr := m.renderStatusMessage()
		m.overlayString(canvas, msgStr, m.screenHeight-4, 2)
	}

	// Add footer at bottom (row screenHeight - 1)
	footer := m.renderFooter()
	m.overlayString(canvas, footer, m.screenHeight-1, 0)

	return m.canvasToString(canvas)
}

// Add a breadcrumb function for command editing
func (m Model) renderCommandEditBreadcrumb() string {
	var path strings.Builder

	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Bold(true)

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	editingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	// Show: Editors -> Edit Menu: MAIN -> Menu Commands -> Edit Command: R
	path.WriteString(categoryStyle.Render("Editors"))
	path.WriteString(arrowStyle.Render(" -> "))
	path.WriteString(detailStyle.Render("Edit Menu: " + m.editingMenu.Name))
	path.WriteString(arrowStyle.Render(" -> "))
	path.WriteString(detailStyle.Render("Menu Commands"))

	if m.editingMenuCommand != nil {
		path.WriteString(arrowStyle.Render(" -> "))
		path.WriteString(highlightStyle.Render("Edit Command: " + m.editingMenuCommand.Keys))

		if m.navMode == EditingValue && m.editingItem != nil {
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(highlightStyle.Render(m.editingItem.Label))
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(editingStyle.Render("EDITING"))
		}
	}

	breadcrumbText := " " + path.String()

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render(breadcrumbText)
}

// Update shouldRenderAsModal to include MenuEditCommandMode
func (m *Model) shouldRenderAsModal() bool {
	// Don't render as modal if we're editing menu data inline
	if m.editingMenu != nil && m.navMode == EditingValue {
		return false
	}

	// Render as modal for command editing
	if m.navMode == MenuEditCommandMode {
		return true
	}

	// If we're in Level 3 with modal fields, render as modal
	if m.navMode == Level3MenuNavigation && len(m.modalFields) > 0 {
		return true
	}

	// If we're in Level 4 or editing, always show modal
	if m.navMode >= Level4ModalNavigation {
		return true
	}

	return false
}

// Add this new function for menu modify breadcrumb
func (m Model) renderMenuModifyBreadcrumb() string {
	var path strings.Builder

	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Bold(true)

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	editingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	// Show: Editors -> Edit Menu: MAIN -> Field -> EDITING
	path.WriteString(categoryStyle.Render("Editors"))
	path.WriteString(arrowStyle.Render(" -> "))
	path.WriteString(detailStyle.Render("Edit Menu: " + m.editingMenu.Name))

	if m.currentMenuTab == 0 {
		path.WriteString(arrowStyle.Render(" -> "))
		path.WriteString(detailStyle.Render("Menu Data"))
	} else {
		path.WriteString(arrowStyle.Render(" -> "))
		path.WriteString(detailStyle.Render("Menu Commands"))
	}

	if m.navMode == EditingValue && m.editingItem != nil {
		path.WriteString(arrowStyle.Render(" -> "))
		path.WriteString(highlightStyle.Render(m.editingItem.Label))
		path.WriteString(arrowStyle.Render(" -> "))
		path.WriteString(editingStyle.Render("EDITING"))
	}

	breadcrumbText := " " + path.String()

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render(breadcrumbText)
}

// renderMainMenu renders the centered main menu with left-justified items
func (m Model) renderMainMenu() string {
	var menuItems []string

	// Calculate max width for menu items
	maxItemWidth := 0
	for _, category := range m.menuBar.Items {
		if len(category.Label) > maxItemWidth {
			maxItemWidth = len(category.Label)
		}
	}

	// Menu width - add padding for icon and spacing
	menuWidth := maxItemWidth + 10

	// Create header with gradient-like effect
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPrimary)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(menuWidth).
		Align(lipgloss.Center)

	header := headerStyle.Render("[ Main Menu ]")

	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorBgLight)).
		Width(menuWidth).
		Align(lipgloss.Center)
	separator := separatorStyle.Render(strings.Repeat("-", menuWidth))

	// Render menu items with left justification
	for i, category := range m.menuBar.Items {
		itemText := " * " + category.Label
		// Calculate padding to fill to menuWidth
		padding := ""
		if len(itemText) < menuWidth {
			padding = strings.Repeat(" ", menuWidth-len(itemText))
		}

		var style lipgloss.Style

		if i == m.activeMenu {
			// Active item - bright highlight with accent color, left-aligned
			style = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorAccent)).
				Width(menuWidth)
		} else {
			// Inactive item - subtle styling, left-aligned
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextNormal)).
				Background(lipgloss.Color(ColorBgMedium)).
				Width(menuWidth)
		}

		menuItems = append(menuItems, style.Render(itemText+padding))
	}

	allLines := []string{header, separator}
	allLines = append(allLines, menuItems...)
	allLines = append(allLines, separator)

	menuContent := strings.Join(allLines, "\n")

	menuBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(menuContent)

	return menuBox
}

// renderModalForm renders the modal with all fields for navigation and editing
func (m Model) renderModalForm() string {
	if len(m.modalFields) == 0 {
		return ""
	}

	// Calculate modal width first
	modalWidth := 58 // Slightly wider for modern look

	// Create header style
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorAccent)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(modalWidth).
		Align(lipgloss.Center)

	header := headerStyle.Render("[ " + m.modalSectionName + " ]")

	// Create separator row
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorAccent)).
		Width(modalWidth)
	separator := separatorStyle.Render(strings.Repeat("-", modalWidth))

	// Check if we're actively editing
	isEditing := m.navMode == EditingValue

	var fieldLines []string

	// Display all fields
	for i, field := range m.modalFields {
		if field.ItemType == EditableField && field.EditableItem != nil {
			currentValue := field.EditableItem.Field.GetValue()
			currentValueStr := formatValue(currentValue, field.EditableItem.ValueType)

			// Truncate long values to prevent wrapping
			maxValueLen := 50
			if len(currentValueStr) > maxValueLen {
				currentValueStr = currentValueStr[:maxValueLen-3] + "..."
			}

			// Check if this is the selected field
			isSelected := i == m.modalFieldIndex

			if isSelected && isEditing {
				// EDITING MODE: Full row highlight with inline input
				if field.EditableItem.ValueType == BoolValue {
					// Boolean field - show toggle options
					currentBool := currentValue.(bool)
					var fieldDisplay string
					if currentBool {
						fieldDisplay = fmt.Sprintf(" %-25s [Y] Yes  [ ] No", field.EditableItem.Label+":")
					} else {
						fieldDisplay = fmt.Sprintf(" %-25s [ ] Yes  [N] No", field.EditableItem.Label+":")
					}

					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color(ColorAccent)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(modalWidth)
					fieldLines = append(fieldLines, fullRowStyle.Render(fieldDisplay))
				} else {
					// Text input field - inline editing
					label := fmt.Sprintf(" %-25s", field.EditableItem.Label+":")

					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color(ColorAccent)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(modalWidth)

					inlineDisplay := label + " " + m.textInput.View()
					fieldLines = append(fieldLines, fullRowStyle.Render(inlineDisplay))
				}
			} else if isSelected && !isEditing {
				// SELECTION MODE: Split highlighting
				labelText := fmt.Sprintf(" %-25s", field.EditableItem.Label+":")
				valueText := " " + currentValueStr

				availableValueSpace := modalWidth - 26
				if len(valueText) > availableValueSpace {
					valueText = valueText[:availableValueSpace-3] + "..."
				}

				labelStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(ColorAccent)).
					Foreground(lipgloss.Color(ColorTextBright)).
					Bold(true).
					Width(26)

				valueStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(ColorBgMedium)).
					Foreground(lipgloss.Color(ColorTextNormal)).
					Width(modalWidth - 26)

				label := labelStyle.Render(labelText)
				value := valueStyle.Render(valueText)

				fieldLines = append(fieldLines, label+value)
			} else {
				// UNSELECTED: Normal display
				fieldDisplay := fmt.Sprintf(" %-25s %s", field.EditableItem.Label+":", currentValueStr)

				if len(fieldDisplay) > modalWidth-1 {
					fieldDisplay = fieldDisplay[:modalWidth-4] + "..."
				}

				fieldStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(ColorBgMedium)).
					Foreground(lipgloss.Color(ColorTextNormal)).
					Width(modalWidth)
				fieldLines = append(fieldLines, fieldStyle.Render(fieldDisplay))
			}
		}
	}

	// Show error if present
	if isEditing && m.editingError != "" {
		errorMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Background(lipgloss.Color(ColorBgMedium)).
			Width(modalWidth).
			Render(" Error: " + m.editingError)
		fieldLines = append(fieldLines, errorMsg)
	}

	// Create footer separator
	// footer separator handled in list composition below

	// Build the full content with bottom ASCII separator
	allLines := []string{header, separator}
	allLines = append(allLines, fieldLines...)
	footerSeparator := separatorStyle.Render(strings.Repeat("-", modalWidth))
	allLines = append(allLines, footerSeparator)

	combined := strings.Join(allLines, "\n")

	// Just wrap with background - NO BORDER, NO PADDING
	modalBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return modalBox
}

// ============================================================================
// Individual Component Renderers
// ============================================================================

// renderLevel2Menu renders the Level 2 submenu with modern styling
func (m Model) renderLevel2Menu(dimmed bool) string {
	if len(m.submenuList.Items()) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Italic(true).
			Render("No items available")

		emptyBox := lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Padding(1, 2).
			Render(emptyMsg)

		return emptyBox
	}

	// Get category label for header
	categoryLabel := ""
	if m.activeMenu < len(m.menuBar.Items) {
		categoryLabel = m.menuBar.Items[m.activeMenu].Label
	}

	// Use fixed max width
	maxWidth := 28

	// Create header with icon
	var headerStyle lipgloss.Style
	if dimmed {
		headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgLight)).
			Foreground(lipgloss.Color(ColorTextDim)).
			Bold(false).
			Width(maxWidth).
			Align(lipgloss.Center)
	} else {
		headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorPrimary)).
			Foreground(lipgloss.Color(ColorTextBright)).
			Bold(true).
			Width(maxWidth).
			Align(lipgloss.Center)
	}

	header := headerStyle.Render("[ " + categoryLabel + " ]")

	// Create decorative separator
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Width(maxWidth)

	if dimmed {
		separatorStyle = separatorStyle.Foreground(lipgloss.Color(ColorBgLight))
	} else {
		separatorStyle = separatorStyle.Foreground(lipgloss.Color(ColorPrimary))
	}

	separator := separatorStyle.Render(strings.Repeat("-", maxWidth))

	// Get list view and clean it up
	listView := m.submenuList.View()
	listView = strings.TrimSpace(listView) // Remove leading/trailing whitespace
	listLines := strings.Split(listView, "\n")

	// Create footer separator (same as header separator)
	// footer separator will be computed below

	allLines := []string{header, separator}
	allLines = append(allLines, listLines...)
	allLines = append(allLines, separator)

	// Join with single newlines
	combined := strings.Join(allLines, "\n")

	// Just wrap with background - NO BORDER, NO PADDING
	submenuBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return submenuBox
}

// renderLevel3Menu renders the Level 3 field list as a cascading modal
func (m Model) renderLevel3Menu(dimmed bool) string {
	if len(m.modalFields) == 0 {
		return ""
	}

	// Build field list first to calculate width
	var fieldLines []string
	maxFieldWidth := 0

	// First pass: calculate max width
	for _, field := range m.modalFields {
		if field.ItemType == EditableField && field.EditableItem != nil {
			currentValue := field.EditableItem.Field.GetValue()
			valueStr := formatValue(currentValue, field.EditableItem.ValueType)
			line := fmt.Sprintf(" * %-18s %s", field.EditableItem.Label+":", valueStr)
			if len(line) > maxFieldWidth {
				maxFieldWidth = len(line)
			}
		}
	}

	// Ensure minimum width
	if maxFieldWidth < 20 {
		maxFieldWidth = 20
	}

	// Create header style matching Level 2
	var headerStyle lipgloss.Style
	if dimmed {
		headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgLight)).
			Foreground(lipgloss.Color(ColorTextDim)).
			Bold(false).
			Width(maxFieldWidth).
			Align(lipgloss.Center)
	} else {
		headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorInfo)). // Different color for Level 3
			Foreground(lipgloss.Color(ColorTextBright)).
			Bold(true).
			Width(maxFieldWidth).
			Align(lipgloss.Center)
	}

	header := headerStyle.Render("[ " + m.modalSectionName + " ]")

	// Create separator row
	var separatorStyle lipgloss.Style
	if dimmed {
		separatorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Foreground(lipgloss.Color(ColorBgLight)).
			Width(maxFieldWidth)
	} else {
		separatorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Foreground(lipgloss.Color(ColorInfo)).
			Width(maxFieldWidth)
	}
	separator := separatorStyle.Render(strings.Repeat("-", maxFieldWidth))

	// Second pass: render fields with consistent width
	for i, field := range m.modalFields {
		if field.ItemType == EditableField && field.EditableItem != nil {
			currentValue := field.EditableItem.Field.GetValue()
			valueStr := formatValue(currentValue, field.EditableItem.ValueType)

			// Format with icon
			line := fmt.Sprintf(" * %-18s %s", field.EditableItem.Label+":", valueStr)

			// Truncate if too long
			if len(line) > maxFieldWidth {
				line = line[:maxFieldWidth-3] + "..."
			}

			// Pad to max width for consistent lightbar
			if len(line) < maxFieldWidth {
				line += strings.Repeat(" ", maxFieldWidth-len(line))
			}

			var style lipgloss.Style
			if dimmed {
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorTextDim)).
					Background(lipgloss.Color(ColorBgMedium)).
					Width(maxFieldWidth)
			} else if i == m.modalFieldIndex {
				// Selected field - full width lightbar
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorTextBright)).
					Background(lipgloss.Color(ColorAccent)).
					Bold(true).
					Width(maxFieldWidth)
			} else {
				// Unselected field
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorTextNormal)).
					Background(lipgloss.Color(ColorBgMedium)).
					Width(maxFieldWidth)
			}

			fieldLines = append(fieldLines, style.Render(line))
		}
	}

	combined := header + "\n" + separator + "\n" + strings.Join(fieldLines, "\n")

	// Just wrap with background - NO BORDER, NO PADDING
	modalBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return modalBox
}

// renderBreadcrumb generates enhanced breadcrumb navigation
func (m Model) renderBreadcrumb() string {
	if m.activeMenu >= len(m.menuBar.Items) {
		return ""
	}

	category := m.menuBar.Items[m.activeMenu]
	var path strings.Builder

	// Style for breadcrumb
	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Bold(true)

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	editingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	// Always show category
	path.WriteString(categoryStyle.Render(category.Label))

	// Build path based on navigation mode
	switch m.navMode {
	case MainMenuNavigation:
		// Nothing more to show

	case Level2MenuNavigation:
		// Show current Level 2 submenu item
		selectedItem := m.submenuList.SelectedItem()
		if selectedItem != nil {
			item := selectedItem.(submenuListItem)
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(detailStyle.Render(item.submenuItem.Label))
		}

	case Level3MenuNavigation:
		// Show: Category -> Section -> Field (if modalFields exist)
		if m.modalSectionName != "" {
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(detailStyle.Render(m.modalSectionName))
		}

		if len(m.modalFields) > 0 && m.modalFieldIndex < len(m.modalFields) {
			field := m.modalFields[m.modalFieldIndex]
			if field.ItemType == EditableField && field.EditableItem != nil {
				path.WriteString(arrowStyle.Render(" -> "))
				path.WriteString(highlightStyle.Render(field.EditableItem.Label))
			}
		}

	case Level4ModalNavigation:
		// Show: Category -> Section -> Field (currently selected)
		if m.modalSectionName != "" {
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(detailStyle.Render(m.modalSectionName))
		}

		if len(m.modalFields) > 0 && m.modalFieldIndex < len(m.modalFields) {
			field := m.modalFields[m.modalFieldIndex]
			if field.ItemType == EditableField && field.EditableItem != nil {
				path.WriteString(arrowStyle.Render(" -> "))
				path.WriteString(highlightStyle.Render(field.EditableItem.Label))
			}
		}

	case EditingValue:
		// Show: Category -> Section -> Field -> EDITING
		if m.modalSectionName != "" {
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(detailStyle.Render(m.modalSectionName))
		}

		if m.editingItem != nil {
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(highlightStyle.Render(m.editingItem.Label))
			path.WriteString(arrowStyle.Render(" -> "))
			path.WriteString(editingStyle.Render("EDITING"))
		}
	}

	breadcrumbText := " " + path.String()

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render(breadcrumbText)
}

// renderStatusMessage renders the status message with icon and color
func (m Model) renderStatusMessage() string {
	var icon string
	var color string

	switch m.messageType {
	case SuccessMessage:
		icon = "[OK]"
		color = "46" // Green
	case ErrorMessage:
		icon = "[X]"
		color = "196" // Red
	case WarningMessage:
		icon = "[!]"
		color = "214" // Orange
	default: // InfoMessage
		icon = "[i]"
		color = "39" // Blue
	}

	msg := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Render(fmt.Sprintf("%s %s", icon, m.message))

	return msg
}

// renderUserManagement renders the user management interface
func (m Model) renderUserManagement() string {
	if len(m.userListUI.Items()) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Italic(true).
			Render("No users found")

		emptyBox := lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Padding(2, 4).
			Render(emptyMsg)

		return emptyBox
	}

	// Create header with user count
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPrimary)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(42).
		Align(lipgloss.Center)

	header := headerStyle.Render(fmt.Sprintf("[ User Management (%d users) ]", len(m.userList)))

	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(42)
	separator := separatorStyle.Render(strings.Repeat("-", 42))

	listView := strings.TrimSpace(m.userListUI.View())

	allLines := []string{header, separator, listView, separator}

	combined := strings.Join(allLines, "\n")

	userBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return userBox
}

// renderSecurityLevelsManagement renders the security levels management interface
func (m Model) renderSecurityLevelsManagement() string {
	if len(m.securityLevelsUI.Items()) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Italic(true).
			Render("No security levels found")

		emptyBox := lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Padding(2, 4).
			Render(emptyMsg)

		return emptyBox
	}

	// Create header with security levels count
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPrimary)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(52).
		Align(lipgloss.Center)

	header := headerStyle.Render(fmt.Sprintf("[ Security Levels Management (%d levels) ]", len(m.securityLevelsList)))

	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(52)
	separator := separatorStyle.Render(strings.Repeat("-", 52))

	listView := strings.TrimSpace(m.securityLevelsUI.View())

	allLines := []string{header, separator, listView, separator}

	combined := strings.Join(allLines, "\n")

	securityLevelsBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return securityLevelsBox
}

// renderMenuManagement renders the menu management interface
func (m Model) renderMenuManagement() string {
	if len(m.menuListUI.Items()) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Italic(true).
			Render("No menus found")

		emptyBox := lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Padding(2, 4).
			Render(emptyMsg)

		return emptyBox
	}

	// Create header with menu count
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPrimary)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(42).
		Align(lipgloss.Center)

	header := headerStyle.Render(fmt.Sprintf("[ Menu Management (%d menus) ]", len(m.menuList)))

	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(42)
	separator := separatorStyle.Render(strings.Repeat("-", 42))

	listView := strings.TrimSpace(m.menuListUI.View())

	allLines := []string{header, separator, listView, separator}

	combined := strings.Join(allLines, "\n")

	menuBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return menuBox
}

// Update renderMenuModify to handle editing state
func (m Model) renderMenuModify() string {
	if m.editingMenu == nil {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Italic(true).
			Render("No menu selected for editing")

		emptyBox := lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Padding(2, 4).
			Render(emptyMsg)

		return emptyBox
	}

	width := 60

	// Create header with menu name
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorPrimary)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(width).
		Align(lipgloss.Center)

	header := headerStyle.Render(fmt.Sprintf("[ Modify Menu: %s ]", m.editingMenu.Name))

	// Create tab bar
	tabNames := []string{"Menu Data", "Menu Commands"}
	var tabParts []string

	for i, name := range tabNames {
		var style lipgloss.Style
		if i == m.currentMenuTab {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorAccent)).
				Foreground(lipgloss.Color(ColorTextBright)).
				Bold(true).
				Width(15).
				Align(lipgloss.Center)
		} else {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorBgGrey)).
				Foreground(lipgloss.Color(ColorTextNormal)).
				Width(15).
				Align(lipgloss.Center)
		}
		tabParts = append(tabParts, style.Render(fmt.Sprintf(" %s ", name)))
	}

	tabSeparator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBgLight)).
		Render(" | ")

	tabsLine := strings.Join(tabParts, tabSeparator)

	padding := (width - m.visualWidth(tabsLine)) / 2
	if padding > 0 {
		tabsLine = strings.Repeat(" ", padding) + tabsLine
	}

	tabBar := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Width(width).
		Render(tabsLine)

	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(width)
	separator := separatorStyle.Render(strings.Repeat("-", width))

	var contentLines []string

	// Render content based on tab, handling editing state
	if m.currentMenuTab == 0 {
		// Menu Data tab
		contentLines = m.renderMenuDataListWithEditing(width)
	} else {
		// Commands tab
		contentLines = m.renderCommandList(width)
	}

	allLines := []string{header, tabBar, separator}
	allLines = append(allLines, contentLines...)
	allLines = append(allLines, separator)

	combined := strings.Join(allLines, "\n")

	menuBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return menuBox
}

// Add new function to handle editing in the menu data list
func (m Model) renderMenuDataListWithEditing(width int) []string {
	if len(m.modalFields) == 0 {
		return []string{
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim)).
				Background(lipgloss.Color(ColorBgMedium)).
				Width(width - 4).
				Render(" No menu data fields available"),
		}
	}

	// Check if we're actively editing
	isEditing := m.navMode == EditingValue

	var dataLines []string
	for i, field := range m.modalFields {
		if field.ItemType == EditableField && field.EditableItem != nil {
			currentValue := field.EditableItem.Field.GetValue()
			valueStr := formatValue(currentValue, field.EditableItem.ValueType)

			isSelected := i == m.selectedCommandIndex

			if isSelected && isEditing {
				// EDITING MODE: Full row highlight with inline input
				if field.EditableItem.ValueType == BoolValue {
					currentBool := currentValue.(bool)
					var fieldDisplay string
					if currentBool {
						fieldDisplay = fmt.Sprintf(" %-20s: [Y] Yes  [ ] No", field.EditableItem.Label)
					} else {
						fieldDisplay = fmt.Sprintf(" %-20s: [ ] Yes  [N] No", field.EditableItem.Label)
					}

					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color(ColorAccent)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(width - 4)
					dataLines = append(dataLines, fullRowStyle.Render(fieldDisplay))
				} else {
					label := fmt.Sprintf(" %-20s:", field.EditableItem.Label)

					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color(ColorAccent)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(width - 4)

					inlineDisplay := label + " " + m.textInput.View()
					dataLines = append(dataLines, fullRowStyle.Render(inlineDisplay))
				}
			} else if isSelected && !isEditing {
				// SELECTION MODE: Split highlighting - only highlight the label
				labelText := fmt.Sprintf(" %-20s:", field.EditableItem.Label)
				valueText := " " + valueStr

				labelWidth := 22
				availableValueSpace := width - 4 - labelWidth
				if len(valueText) > availableValueSpace {
					valueText = valueText[:availableValueSpace-3] + "..."
				}

				labelStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(ColorAccent)).
					Foreground(lipgloss.Color(ColorTextBright)).
					Bold(true).
					Width(labelWidth)

				valueStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(ColorBgMedium)).
					Foreground(lipgloss.Color(ColorTextNormal)).
					Width(width - 4 - labelWidth)

				label := labelStyle.Render(labelText)
				value := valueStyle.Render(valueText)

				dataLines = append(dataLines, label+value)
			} else {
				// UNSELECTED: Normal display
				maxValueLen := 30
				if len(valueStr) > maxValueLen {
					valueStr = valueStr[:maxValueLen-3] + "..."
				}

				line := fmt.Sprintf(" %-20s: %s", field.EditableItem.Label, valueStr)
				if len(line) > width-4 {
					line = line[:width-7] + "..."
				}

				style := lipgloss.NewStyle().
					Foreground(lipgloss.Color(ColorTextNormal)).
					Background(lipgloss.Color(ColorBgMedium)).
					Width(width - 4)

				dataLines = append(dataLines, style.Render(line))
			}
		}
	}

	// Pad to 14 lines
	for len(dataLines) < 14 {
		dataLines = append(dataLines, lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Width(width-4).
			Render(""))
	}

	return dataLines
}

// renderCommandList renders the command list for the commands tab
func (m Model) renderCommandList(width int) []string {
	var commandLines []string
	for i, cmd := range m.menuCommandsList {
		// Format: [CommandNumber] [Keys] [ShortDescription]
		line := fmt.Sprintf(" %d. %-3s %s", cmd.CommandNumber, cmd.Keys, cmd.ShortDescription)
		if len(line) > width-4 {
			line = line[:width-7] + "..."
		}
		if len(line) < width-4 {
			line += strings.Repeat(" ", width-4-len(line))
		}

		var style lipgloss.Style
		if i == m.selectedCommandIndex {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorAccent)).
				Bold(true).
				Width(width - 4)
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextNormal)).
				Background(lipgloss.Color(ColorBgMedium)).
				Width(width - 4)
		}
		commandLines = append(commandLines, style.Render(line))
	}

	// If no commands, show message
	if len(commandLines) == 0 {
		commandLines = []string{
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim)).
				Background(lipgloss.Color(ColorBgMedium)).
				Width(width - 4).
				Render(" No commands defined - press 'I' to add"),
		}
	}

	// Pad to 14 lines to match menu data list height
	for len(commandLines) < 14 {
		commandLines = append(commandLines, lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Width(width-4).
			Render(""))
	}

	return commandLines
}

// setupMenuEditCommandModal sets up modal fields for command editing
func (m *Model) setupMenuEditCommandModal() {
	if m.editingMenuCommand == nil {
		return
	}

	// Create modal fields for command editing
	m.modalFields = []SubmenuItem{
		{
			ID:       "command-number",
			Label:    "Command Number",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-number",
				Label:     "Command Number",
				ValueType: IntValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.CommandNumber },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.CommandNumber = v.(int)
						return nil
					},
				},
				HelpText: "Command number in menu",
			},
		},
		{
			ID:       "command-keys",
			Label:    "Keys",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-keys",
				Label:     "Keys",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.Keys },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.Keys = v.(string)
						return nil
					},
				},
				HelpText: "Hotkeys for this command",
			},
		},
		{
			ID:       "command-short-description",
			Label:    "Short Description",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-short-description",
				Label:     "Short Description",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.ShortDescription },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.ShortDescription = v.(string)
						return nil
					},
				},
				HelpText: "Short description for menu display",
			},
		},
		{
			ID:       "command-acs-required",
			Label:    "ACS Required",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-acs-required",
				Label:     "ACS Required",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.ACSRequired },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.ACSRequired = v.(string)
						return nil
					},
				},
				HelpText: "Access Control String required",
			},
		},
		{
			ID:       "command-cmdkeys",
			Label:    "CmdKeys",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-cmdkeys",
				Label:     "CmdKeys",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.CmdKeys },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.CmdKeys = v.(string)
						return nil
					},
				},
				HelpText: "Command key for execution handler",
			},
		},
		{
			ID:       "command-options",
			Label:    "Options",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-options",
				Label:     "Options",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.Options },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.Options = v.(string)
						return nil
					},
				},
				HelpText: "Command options/parameters",
			},
		},
	}

	m.modalFieldIndex = 0
	m.modalSectionName = fmt.Sprintf("Edit Command: %s", m.editingMenuCommand.Keys)
}

// setupMenuEditDataModal sets up modal fields for menu data editing
func (m *Model) setupMenuEditDataModal() {
	// Create modal fields for menu data editing
	m.modalFields = []SubmenuItem{
		{
			ID:       "menu-name",
			Label:    "Name",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-name",
				Label:     "Name",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.Name },
					SetValue: func(v interface{}) error {
						m.editingMenu.Name = v.(string)
						return nil
					},
				},
				HelpText: "Menu name (unique identifier)",
			},
		},
		{
			ID:       "menu-titles",
			Label:    "Titles",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-titles",
				Label:     "Titles",
				ValueType: ListValue,
				Field: ConfigField{
					GetValue: func() interface{} { return strings.Join(m.editingMenu.Titles, ", ") },
					SetValue: func(v interface{}) error {
						s := v.(string)
						m.editingMenu.Titles = nil
						for _, title := range strings.Split(s, ",") {
							title = strings.TrimSpace(title)
							if title != "" {
								m.editingMenu.Titles = append(m.editingMenu.Titles, title)
							}
						}
						return nil
					},
				},
				HelpText: "Menu titles (comma-separated)",
			},
		},
		{
			ID:       "menu-prompt",
			Label:    "Prompt",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-prompt",
				Label:     "Prompt",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.Prompt },
					SetValue: func(v interface{}) error {
						m.editingMenu.Prompt = v.(string)
						return nil
					},
				},
				HelpText: "Menu prompt text",
			},
		},
		{
			ID:       "menu-acs-required",
			Label:    "ACS Required",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-acs-required",
				Label:     "ACS Required",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.ACSRequired },
					SetValue: func(v interface{}) error {
						m.editingMenu.ACSRequired = v.(string)
						return nil
					},
				},
				HelpText: "Access Control String required",
			},
		},
		{
			ID:       "menu-generic-columns",
			Label:    "Generic Columns",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-generic-columns",
				Label:     "Columns",
				ValueType: IntValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.GenericColumns },
					SetValue: func(v interface{}) error {
						m.editingMenu.GenericColumns = v.(int)
						return nil
					},
				},
				HelpText: "Number of columns for generic display",
			},
		},
		{
			ID:       "menu-generic-bracket-color",
			Label:    "Bracket Color",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-generic-bracket-color",
				Label:     "Bracket Color",
				ValueType: IntValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.GenericBracketColor },
					SetValue: func(v interface{}) error {
						m.editingMenu.GenericBracketColor = v.(int)
						return nil
					},
				},
				HelpText: "Color for brackets in generic display",
			},
		},
		{
			ID:       "menu-generic-command-color",
			Label:    "Command Color",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-generic-command-color",
				Label:     "Command Color",
				ValueType: IntValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.GenericCommandColor },
					SetValue: func(v interface{}) error {
						m.editingMenu.GenericCommandColor = v.(int)
						return nil
					},
				},
				HelpText: "Color for commands in generic display",
			},
		},
		{
			ID:       "menu-generic-desc-color",
			Label:    "Desc Color",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-generic-desc-color",
				Label:     "Desc Color",
				ValueType: IntValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.GenericDescColor },
					SetValue: func(v interface{}) error {
						m.editingMenu.GenericDescColor = v.(int)
						return nil
					},
				},
				HelpText: "Color for descriptions in generic display",
			},
		},
		{
			ID:       "menu-clear-screen",
			Label:    "Clear Screen",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-clear-screen",
				Label:     "Clear Screen",
				ValueType: BoolValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenu.ClearScreen },
					SetValue: func(v interface{}) error {
						m.editingMenu.ClearScreen = v.(bool)
						return nil
					},
				},
				HelpText: "Clear screen before menu display",
			},
		},
	}

	m.modalFieldIndex = 0
	m.modalSectionName = fmt.Sprintf("Edit Menu: %s", m.editingMenu.Name)
}

// renderMenuEditCommand renders the menu edit command modal
func (m Model) renderMenuEditCommand() string {
	if m.editingMenuCommand == nil {
		return "No command selected for editing"
	}

	// Use the existing modal rendering
	return m.renderModalForm()
}

// Update renderFooter to show correct footer when editing in MenuModifyMode
func (m Model) renderFooter() string {
	var footerText string

	switch m.navMode {
	case MainMenuNavigation:
		footerText = "  Up/Down Navigate   ENTER Select   ESC Exit"
	case Level2MenuNavigation:
		footerText = "  Up/Down Navigate   ENTER Select   ESC Back"
	case Level3MenuNavigation:
		footerText = "  Up/Down Navigate   ENTER Edit   ESC Back"
	case Level4ModalNavigation:
		footerText = "  Up/Down Navigate   ENTER Edit   ESC Back"
	case EditingValue:
		footerText = "  ENTER Save   ESC Cancel"
	case UserManagementMode:
		footerText = "  Up/Down Navigate   ENTER Edit   ESC Back"
	case SecurityLevelsMode:
		footerText = "  Up/Down Navigate   ENTER Edit   ESC Back"
	case MenuManagementMode:
		footerText = "  Up/Down Navigate   ENTER/M Modify   I Insert   D Delete   ESC Back"
	case MenuModifyMode:
		if m.currentMenuTab == 0 {
			footerText = "  Up/Down Navigate   ENTER Edit   TAB Switch   ESC Back"
		} else {
			footerText = "  Up/Down Navigate   ENTER Edit   I Insert   D Delete   TAB Switch   ESC Back"
		}
	case MenuEditDataMode:
		footerText = "  Up/Down Navigate   ENTER Edit   ESC Back"
	case MenuEditCommandMode:
		footerText = "  Up/Down Navigate   ENTER Edit   ESC Back to Commands"
	case SavePrompt:
		footerText = "  Y Yes   N No   ESC Cancel"
	case SaveChangesPrompt:
		footerText = "  Y Yes   N No   ESC Cancel"
	default:
		footerText = "  F1 Help   ESC Exit"
	}

	// Style the footer with full width
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextNormal)).
		Background(lipgloss.Color("8")).
		Width(m.screenWidth).
		Render(footerText)

	return footer
}
