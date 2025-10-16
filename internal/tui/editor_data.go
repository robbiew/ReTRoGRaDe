package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/robbiew/retrograde/internal/database"
)

// loadUsers loads all users from the database
func (m *Model) loadUsers() error {
	if m.db == nil {
		return fmt.Errorf("database not available")
	}

	users, err := m.db.GetAllUsers()
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	m.userList = users

	// Initialize user list UI
	var userItems []list.Item
	for _, user := range users {
		userItems = append(userItems, userListItem{user: user})
	}

	// Calculate max width for user list
	maxWidth := 40 // Narrower for user info

	userList := list.New(userItems, userDelegate{maxWidth: maxWidth}, maxWidth, 15)
	userList.Title = ""
	userList.SetShowStatusBar(false)
	userList.SetFilteringEnabled(true)
	userList.SetShowHelp(false)
	userList.SetShowPagination(true)

	// Remove any default list styling/padding
	userList.Styles.Title = lipgloss.NewStyle()
	userList.Styles.PaginationStyle = lipgloss.NewStyle()
	userList.Styles.HelpStyle = lipgloss.NewStyle()

	m.userListUI = userList

	return nil
}

// loadSecurityLevels loads all security levels from the database
func (m *Model) loadSecurityLevels() error {
	if m.db == nil {
		return fmt.Errorf("database not available")
	}

	securityLevels, err := m.db.GetAllSecurityLevels()
	if err != nil {
		return fmt.Errorf("failed to get security levels: %w", err)
	}

	m.securityLevelsList = securityLevels

	// Initialize security levels list UI
	var securityLevelItems []list.Item
	for _, level := range securityLevels {
		securityLevelItems = append(securityLevelItems, securityLevelListItem{securityLevel: level})
	}

	// Calculate max width for security levels list
	maxWidth := 50 // Narrower for security level info

	securityLevelsList := list.New(securityLevelItems, securityLevelDelegate{maxWidth: maxWidth}, maxWidth, 15)
	securityLevelsList.Title = ""
	securityLevelsList.SetShowStatusBar(false)
	securityLevelsList.SetFilteringEnabled(true)
	securityLevelsList.SetShowHelp(false)
	securityLevelsList.SetShowPagination(true)

	// Remove any default list styling/padding
	securityLevelsList.Styles.Title = lipgloss.NewStyle()
	securityLevelsList.Styles.PaginationStyle = lipgloss.NewStyle()
	securityLevelsList.Styles.HelpStyle = lipgloss.NewStyle()

	m.securityLevelsUI = securityLevelsList

	return nil
}

// loadMenus loads all menus from the database
func (m *Model) loadMenus() error {
	if m.db == nil {
		return fmt.Errorf("database not available")
	}

	menus, err := m.db.GetAllMenus()
	if err != nil {
		return fmt.Errorf("failed to get menus: %w", err)
	}

	// If no menus exist, seed a default menu
	if len(menus) == 0 {
		if err := database.SeedDefaultMainMenu(m.db); err != nil {
			return fmt.Errorf("failed to seed default menu: %w", err)
		}
		// Reload menus after seeding
		menus, err = m.db.GetAllMenus()
		if err != nil {
			return fmt.Errorf("failed to reload menus after seeding: %w", err)
		}
	}

	m.menuList = menus

	// Initialize menu list UI
	var menuItems []list.Item
	for _, menu := range menus {
		// Get command count for this menu
		commands, err := m.db.GetMenuCommands(menu.ID)
		commandCount := 0
		if err == nil {
			commandCount = len(commands)
		}
		menuItems = append(menuItems, menuListItem{menu: menu, commandCount: commandCount})
	}

	// Calculate max width for menu list
	maxWidth := 40 // For menu info

	menuList := list.New(menuItems, menuDelegate{maxWidth: maxWidth}, maxWidth, 15)
	menuList.Title = ""
	menuList.SetShowStatusBar(false)
	menuList.SetFilteringEnabled(true)
	menuList.SetShowHelp(false)
	menuList.SetShowPagination(true)

	// Remove any default list styling/padding
	menuList.Styles.Title = lipgloss.NewStyle()
	menuList.Styles.PaginationStyle = lipgloss.NewStyle()
	menuList.Styles.HelpStyle = lipgloss.NewStyle()

	m.menuListUI = menuList

	return nil
}

// loadMenuCommandsForEditing loads commands for the currently editing menu
func (m *Model) loadMenuCommandsForEditing() error {
	if m.editingMenu == nil {
		return fmt.Errorf("no menu being edited")
	}

	commands, err := m.db.GetMenuCommands(m.editingMenu.ID)
	if err != nil {
		return fmt.Errorf("failed to get menu commands: %w", err)
	}

	m.menuCommandsList = commands
	return nil
}
