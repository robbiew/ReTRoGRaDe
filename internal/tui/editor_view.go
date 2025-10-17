package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/robbiew/retrograde/internal/ui"
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

		// Add color key reference when on Menu Data tab
		if m.currentMenuTab == 0 {
			colorKey := m.renderColorKeyReference()
			m.overlayString(canvas, colorKey, m.screenHeight-2, 0)
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

	// Layer 1.12: SelectingValue mode - show selection list over command modal (NEW)
	if m.navMode == SelectingValue {
		// Show the command edit modal in the background
		menuEditCommandStr := m.renderMenuEditCommand()
		m.overlayStringCenteredWithClear(canvas, menuEditCommandStr)

		// Overlay the selection list on top
		selectionModal := m.renderSelectingValueView()
		m.overlayStringCentered(canvas, selectionModal)

		// Add breadcrumb
		breadcrumb := m.renderCommandEditBreadcrumb()
		m.overlayString(canvas, breadcrumb, m.screenHeight-3, 0)

		// Add footer
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

	// Layer 5: Delete confirmation prompt
	if m.navMode == DeleteConfirmPrompt && m.savePrompt {
		promptStr := m.renderSavePrompt()
		m.overlayStringCentered(canvas, promptStr)

		footer := m.renderFooter()
		m.overlayString(canvas, footer, m.screenHeight-1, 0)

		return m.canvasToString(canvas)
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
		Foreground(lipgloss.Color("7")). // Basic 16: light gray/white
		Bold(true)

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")) // Basic 16: dark gray

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")) // Basic 16: light gray/white

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")). // Basic 16: bright yellow
		Bold(true)

	editingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")). // Basic 16: bright red
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
		Foreground(lipgloss.Color("7")). // Basic 16: light gray/white (changed from 243)
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
	modalWidth := 70 // Slightly wider for modern look

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
				currentValueStr = ui.TruncateWithPipeCodes(currentValueStr, maxValueLen-3)
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
						fieldDisplay = fmt.Sprintf(" %-14s: [Y] Yes  [ ] No", field.EditableItem.Label) // Changed
					} else {
						fieldDisplay = fmt.Sprintf(" %-14s: [ ] Yes  [N] No", field.EditableItem.Label) // Changed
					}

					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color(ColorAccent)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(modalWidth)
					fieldLines = append(fieldLines, fullRowStyle.Render(fieldDisplay))
				} else {
					// Text input field - inline editing
					// Set dynamic width for text input
					availableInputWidth := modalWidth - 17
					m.textInput.Width = availableInputWidth

					label := fmt.Sprintf(" %-14s:", field.EditableItem.Label)

					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color(ColorPrimary)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(modalWidth)

					inlineDisplay := label + " " + m.textInput.View()
					fieldLines = append(fieldLines, fullRowStyle.Render(inlineDisplay))
				}
			} else if isSelected && !isEditing {
				// SELECTION MODE: Split highlighting
				labelText := fmt.Sprintf(" %-14s:", field.EditableItem.Label) // Changed
				valueText := " " + currentValueStr

				availableValueSpace := modalWidth - 16
				if len(valueText) > availableValueSpace {
					valueText = ui.TruncateWithPipeCodes(valueText, availableValueSpace)
				}

				labelStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(ColorAccent)).
					Foreground(lipgloss.Color(ColorTextBright)).
					Bold(true).
					Width(16)

				valueStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(ColorBgMedium)).
					Foreground(lipgloss.Color(ColorTextNormal)).
					Width(modalWidth - 16)

				label := labelStyle.Render(labelText)
				value := valueStyle.Render(valueText)

				fieldLines = append(fieldLines, label+value)
			} else {
				// UNSELECTED: Normal display
				fieldDisplay := fmt.Sprintf(" %-14s: %s", field.EditableItem.Label, currentValueStr) // Changed

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
				line = ui.TruncateWithPipeCodes(line, maxFieldWidth-3)
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

	// Add column headers
	columnHeaders := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(42).
		Render(fmt.Sprintf(" %-15s %-5s %-5s", "Username", "Level", "UID"))

	listView := strings.TrimSpace(m.userListUI.View())

	allLines := []string{header, separator, columnHeaders, separator, listView, separator}

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

	width := 70

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
				Foreground(lipgloss.Color("7")). // Changed: light gray, removed Background
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

// Complete fix for renderMenuDataListWithEditing in internal/tui/editor_view.go
// Replace the entire function with this corrected version

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

			// Track whether we've parsed pipe codes (which converts to ANSI)
			isParsed := false

			// Apply pipe code parsing for Name, Title, and Prompt fields
			if field.EditableItem.ID == "menu-name" || field.EditableItem.ID == "menu-title-1" || field.EditableItem.ID == "menu-title-2" || field.EditableItem.ID == "menu-prompt" {
				valueStr = ui.ParsePipeColorCodes(valueStr)
				isParsed = true
			}

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
						Background(lipgloss.Color(ColorPrimary)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(width - 4)
					dataLines = append(dataLines, fullRowStyle.Render(fieldDisplay))
				} else {
					// Set dynamic width for text input in menu data editing
					availableInputWidth := width - 4 - 16 // width - padding - label width
					// Allow wider input for Prompt field to enable horizontal scrolling
					if field.EditableItem.ID == "menu-prompt" {
						availableInputWidth = width - 4 - 10 // More space for prompt
					}
					m.textInput.Width = availableInputWidth

					label := fmt.Sprintf(" %-14s:", field.EditableItem.Label)

					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color(ColorPrimary)).
						Foreground(lipgloss.Color(ColorTextBright)).
						Bold(true).
						Width(width - 4)

					inlineDisplay := label + " " + m.textInput.View()
					dataLines = append(dataLines, fullRowStyle.Render(inlineDisplay))
				}
			} else if isSelected && !isEditing {
				// SELECTION MODE: Split highlighting - only highlight the label
				labelText := fmt.Sprintf(" %-14s:", field.EditableItem.Label)
				valueText := " " + valueStr

				labelWidth := 16
				availableValueSpace := width - 4 - labelWidth

				// Use ANSI functions if we parsed pipe codes, otherwise use pipe code functions
				if isParsed {
					if len(ui.StripANSI(valueText)) > availableValueSpace {
						valueText = ui.TruncateWithANSICodes(valueText, availableValueSpace)
					}
				} else {
					if len(ui.StripPipeCodes(valueText)) > availableValueSpace {
						valueText = ui.TruncateWithPipeCodes(valueText, availableValueSpace)
					}
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
				// Allow longer display for Prompt field to prevent early truncation
				maxValueLen := 49
				if field.EditableItem.ID == "menu-prompt" {
					maxValueLen = width - 20 // Allow more space for prompt field
				}

				// Use ANSI functions if we parsed pipe codes, otherwise use pipe code functions
				if isParsed {
					if len(ui.StripANSI(valueStr)) > maxValueLen {
						valueStr = ui.TruncateWithANSICodes(valueStr, maxValueLen)
					}
				} else {
					if len(ui.StripPipeCodes(valueStr)) > maxValueLen {
						valueStr = ui.TruncateWithPipeCodes(valueStr, maxValueLen)
					}
				}

				line := fmt.Sprintf(" %-14s: %s", field.EditableItem.Label, valueStr)

				// Check full line and truncate if needed
				if isParsed {
					if len(ui.StripANSI(line)) > width-4 {
						line = ui.TruncateWithANSICodes(line, width-4)
					}
				} else {
					if len(ui.StripPipeCodes(line)) > width-4 {
						line = ui.TruncateWithPipeCodes(line, width-4)
					}
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

	// Add column header with dark grey background
	columnHeader := lipgloss.NewStyle().
		Background(lipgloss.Color("8")).
		Foreground(lipgloss.Color(ColorTextBright)).
		Bold(true).
		Width(width).
		Render(fmt.Sprintf(" %-4s %-4s %-30s %-6s", "#", "Key", "Description", "Active"))
	commandLines = append(commandLines, columnHeader)

	for i, cmd := range m.menuCommandsList {
		// Format: [CommandNumber] [Keys] [ShortDescription] [Active Status]
		activeIndicator := "[✓]"
		if !cmd.Active {
			activeIndicator = "[X]"
		}
		line := fmt.Sprintf(" %-4s %-4s %-30s %-6s", fmt.Sprintf("%d.", cmd.CommandNumber), cmd.Keys, cmd.ShortDescription, activeIndicator)
		if len(line) > width {
			line = ui.TruncateWithPipeCodes(line, width)
		}
		if len(line) < width {
			line += strings.Repeat(" ", width-len(line))
		}

		var style lipgloss.Style
		if i == m.selectedCommandIndex {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorAccent)).
				Bold(true).
				Width(width)
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextNormal)).
				Background(lipgloss.Color(ColorBgMedium)).
				Width(width)
		}
		commandLines = append(commandLines, style.Render(line))
	}

	// If no commands, show message after header
	if len(m.menuCommandsList) == 0 {
		commandLines = append(commandLines, lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim)).
			Background(lipgloss.Color(ColorBgMedium)).
			Width(width).
			Render(" No commands defined - press 'I' to add"))
	}

	// Pad to 14 lines to match menu data list height
	for len(commandLines) < 14 {
		commandLines = append(commandLines, lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgMedium)).
			Width(width).
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
				HelpText: "Key(s) user presses (e.g., R, P, G)",
			},
		},
		{
			ID:       "command-short-description",
			Label:    "Short Description",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-short-description",
				Label:     "Short Desc",
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
				ValueType: SelectValue, // THIS ONE - Keep SelectValue
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.CmdKeys },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.CmdKeys = v.(string)
						return nil
					},
				},
				HelpText:      "Command key for execution handler",
				SelectOptions: getCmdKeySelectOptions(), // Keep this
			},
		},
		// REMOVE THE DUPLICATE CmdKeys FIELD THAT WAS HERE
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
		{
			ID:       "command-active",
			Label:    "Active",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "command-active",
				Label:     "Active",
				ValueType: BoolValue,
				Field: ConfigField{
					GetValue: func() interface{} { return m.editingMenuCommand.Active },
					SetValue: func(v interface{}) error {
						m.editingMenuCommand.Active = v.(bool)
						return nil
					},
				},
				HelpText: "Whether this command is active/available",
			},
		},
	}

	m.modalFieldIndex = 0
	m.modalSectionName = fmt.Sprintf("Edit Command: %s", m.editingMenuCommand.Keys)
}

// renderSelectingValueView renders the selection list modal
func (m Model) renderSelectingValueView() string {
	if m.editingItem == nil || m.editingItem.ValueType != SelectValue {
		return ""
	}

	var b strings.Builder
	options := m.editingItem.SelectOptions

	// Modal dimensions
	modalWidth := 78
	modalHeight := 19
	visibleItems := modalHeight - 6 // Account for header, footer, borders

	// Title
	title := fmt.Sprintf(" Select %s ", m.editingItem.Label)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextBright)).
		Background(lipgloss.Color(ColorPrimary)).
		Bold(true)

	// Build content with fixed line count
	var content strings.Builder
	content.WriteString(titleStyle.Render(title) + "\n\n")

	// Build a list of display lines with their metadata
	type displayLine struct {
		text       string
		isCategory bool
		isSelected bool
		optIndex   int
	}

	var displayLines []displayLine
	lastCategory := ""

	for i := 0; i < len(options); i++ {
		opt := options[i]

		// Add category header if it changed
		if opt.Category != lastCategory {
			if lastCategory != "" {
				displayLines = append(displayLines, displayLine{text: "", isCategory: false}) // blank line
			}
			categoryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorAccent)).
				Bold(true)
			displayLines = append(displayLines, displayLine{
				text:       categoryStyle.Render(fmt.Sprintf("— %s —", opt.Category)),
				isCategory: true,
			})
			lastCategory = opt.Category
		}

		// Add the option line (only implemented commands shown)
		line := fmt.Sprintf(" [%s] %-25s %s", opt.Value, opt.Label, opt.Description)
		if len(line) > modalWidth-4 {
			line = line[:modalWidth-7] + "..."
		}
		displayLines = append(displayLines, displayLine{
			text:       line,
			isCategory: false,
			isSelected: i == m.selectListIndex,
			optIndex:   i,
		})
	}

	// Calculate visible window based on selected item's display line
	selectedDisplayLine := 0
	for i, dl := range displayLines {
		if !dl.isCategory && dl.optIndex == m.selectListIndex {
			selectedDisplayLine = i
			break
		}
	}

	// Calculate scroll window
	displayStart := selectedDisplayLine - (visibleItems / 2)
	if displayStart < 0 {
		displayStart = 0
	}
	displayEnd := displayStart + visibleItems
	if displayEnd > len(displayLines) {
		displayEnd = len(displayLines)
		displayStart = displayEnd - visibleItems
		if displayStart < 0 {
			displayStart = 0
		}
	}

	// Render visible lines and count them
	lineCount := 0
	for i := displayStart; i < displayEnd && i < len(displayLines); i++ {
		dl := displayLines[i]
		if dl.isCategory {
			content.WriteString(dl.text + "\n")
		} else if dl.isSelected {
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextBright)).
				Background(lipgloss.Color(ColorAccent)).
				Bold(true)
			content.WriteString(style.Render(dl.text) + "\n")
		} else {
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextNormal))
			content.WriteString(style.Render(dl.text) + "\n")
		}
		lineCount++
	}

	// Pad with empty lines to maintain consistent height
	for lineCount < visibleItems {
		content.WriteString("\n")
		lineCount++
	}

	// Scrollbar indicator
	if len(options) > visibleItems {
		// Calculate which actual options are visible (not display lines)
		firstVisibleOpt := 0
		lastVisibleOpt := 0
		for i := displayStart; i < displayEnd && i < len(displayLines); i++ {
			if !displayLines[i].isCategory && displayLines[i].text != "" {
				if firstVisibleOpt == 0 {
					firstVisibleOpt = displayLines[i].optIndex + 1
				}
				lastVisibleOpt = displayLines[i].optIndex + 1
			}
		}
		scrollInfo := fmt.Sprintf(" %d-%d of %d ", firstVisibleOpt, lastVisibleOpt, len(options))
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

// setupMenuEditDataModal sets up modal fields for menu data editing
func (m *Model) setupMenuEditDataModal() {
	// Ensure Titles slice has at least 2 elements
	for len(m.editingMenu.Titles) < 2 {
		m.editingMenu.Titles = append(m.editingMenu.Titles, "")
	}
	// Limit to maximum of 2 items
	if len(m.editingMenu.Titles) > 2 {
		m.editingMenu.Titles = m.editingMenu.Titles[:2]
	}

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
			ID:       "menu-title-1",
			Label:    "Title 1",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-title-1",
				Label:     "Title 1",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} {
						if len(m.editingMenu.Titles) > 0 {
							return m.editingMenu.Titles[0]
						}
						return ""
					},
					SetValue: func(v interface{}) error {
						s := v.(string)
						if len(m.editingMenu.Titles) == 0 {
							m.editingMenu.Titles = append(m.editingMenu.Titles, s)
						} else {
							m.editingMenu.Titles[0] = s
						}
						// Ensure at most 2 items
						if len(m.editingMenu.Titles) > 2 {
							m.editingMenu.Titles = m.editingMenu.Titles[:2]
						}
						return nil
					},
				},
				HelpText: "First menu title",
			},
		},
		{
			ID:       "menu-title-2",
			Label:    "Title 2",
			ItemType: EditableField,
			EditableItem: &MenuItem{
				ID:        "menu-title-2",
				Label:     "Title 2",
				ValueType: StringValue,
				Field: ConfigField{
					GetValue: func() interface{} {
						if len(m.editingMenu.Titles) > 1 {
							return m.editingMenu.Titles[1]
						}
						return ""
					},
					SetValue: func(v interface{}) error {
						s := v.(string)
						if len(m.editingMenu.Titles) < 2 {
							m.editingMenu.Titles = append(m.editingMenu.Titles, s)
						} else {
							m.editingMenu.Titles[1] = s
						}
						// Ensure at most 2 items
						if len(m.editingMenu.Titles) > 2 {
							m.editingMenu.Titles = m.editingMenu.Titles[:2]
						}
						return nil
					},
				},
				HelpText: "Second menu title",
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
	case DeleteConfirmPrompt:
		footerText = "  Y Yes   N No   ESC Cancel"
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

// Add this function to internal/tui/editor_view.go

// renderColorKeyReference renders a horizontal color code reference
func (m Model) renderColorKeyReference() string {
	var parts []string

	// Add label
	parts = append(parts, "Color Codes: ")

	// Add each color code (00-15) with the actual color
	for i := 0; i <= 15; i++ {
		code := fmt.Sprintf("%02d", i)
		// Use ui.ColorFromNumber to get the ANSI color
		coloredCode := ui.ColorFromNumber(i) + code + ui.Ansi.Reset
		parts = append(parts, "|")
		parts = append(parts, coloredCode)
		parts = append(parts, " ")
	}

	line := strings.Join(parts, "")

	// Center the line
	centered := lipgloss.NewStyle().
		Width(m.screenWidth).
		Align(lipgloss.Center).
		Render(line)

	return centered
}
