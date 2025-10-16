package tui

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/database"
)

//go:embed config.ans
var configArt string

// ============================================================================
// Color Scheme & Styling
// ============================================================================

// Color palette for modern theme
const (
	// Primary colors
	ColorPrimary     = "4"  // Blue
	ColorPrimaryDark = "4"  // Blue
	ColorAccent      = "5"  // Magenta
	ColorAccentLight = "13" // Bright Magenta

	// Backgrounds
	ColorBgGrey   = "8" // Black
	ColorBgDark   = "0" // Black
	ColorBgMedium = "0" // Black
	ColorBgLight  = "7" // White

	// Text colors
	ColorTextBright = "15" // Bright White
	ColorTextNormal = "7"  // White
	ColorTextDim    = "8"  // Bright Black/Gray
	ColorTextAccent = "13" // Bright Magenta

	// Status colors
	ColorSuccess = "2" // Green
	ColorWarning = "3" // Yellow
	ColorError   = "1" // Red
	ColorInfo    = "6" // Cyan
)

// ============================================================================
// UI Layout Constants
// ============================================================================

const (
	Level3AnchorBottom = true // Set to true to anchor to bottom
	Level3AnchorRight  = true // Set to true to anchor to right
	Level3BottomOffset = 3    // Rows from bottom (for footer/breadcrumb space)
	Level3RightOffset  = 2    // Columns from right edge

	// When enabled, the main menu is positioned relative to bottom/right
	// of the screen using the offsets below. When disabled, it is centered.
	MainAnchorBottom = true
	MainAnchorRight  = true
	MainBottomOffset = 2
	MainRightOffset  = 1
)

// NavigationMode represents the current UI state
type NavigationMode int

const (
	MainMenuNavigation    NavigationMode = iota // Level 1: Main menu centered H/V
	Level2MenuNavigation                        // Level 2: Submenu anchored top left
	Level3MenuNavigation                        // Level 3: Submenu offset lower left from Level 2
	Level4ModalNavigation                       // Level 4: Centered modal for final menus
	EditingValue                                // Editing a value in modal form
	UserManagementMode                          // User management interface
	SecurityLevelsMode                          // Security levels management interface
	MenuManagementMode                          // Menu management interface
	MenuModifyMode                              // Menu modification interface (command list)
	MenuEditDataMode                            // Edit menu data modal
	MenuEditCommandMode                         // Edit command modal
	SavePrompt                                  // Confirming save on exit
	SaveChangesPrompt                           // NEW: Prompt to save changes when exiting edit modal

)

// Model represents the complete application state
type Model struct {
	// Configuration data
	config *config.Config
	db     *database.SQLiteDB // Database connection for user management

	// Navigation state
	navMode    NavigationMode
	activeMenu int // Current horizontal menu index

	// Menu structure
	menuBar MenuBar

	// Submenu list
	submenuList list.Model

	// User management list
	userListUI list.Model

	// Security levels management list
	securityLevelsUI list.Model

	// Menu management list
	menuListUI list.Model

	// Modal form state
	modalFields      []SubmenuItem // All fields in the current section
	modalFieldIndex  int           // Currently selected field in modal
	modalSectionName string        // Section name for modal header

	// Editing state
	editingItem   *MenuItem
	textInput     textinput.Model
	editingError  string
	originalValue interface{}

	// User management state
	userList []database.UserRecord // List of users for management

	// Security levels management state
	securityLevelsList   []database.SecurityLevelRecord // List of security levels for management
	editingSecurityLevel *database.SecurityLevelRecord  // Currently editing security level

	// Menu management state
	menuList         []database.Menu        // List of menus for management
	menuCommandsList []database.MenuCommand // List of commands for current menu

	// User management state
	editingUser *database.UserRecord // Currently editing user

	// Confirmation state
	confirmAction string // Current confirmation action
	confirmMenuID int64  // Menu ID for confirmation

	// UI state
	screenWidth   int
	screenHeight  int
	message       string
	messageTime   time.Time
	messageType   MessageType
	savePrompt    bool
	quitting      bool
	modifiedCount int

	savePromptSelection int            // 0 = No, 1 = Yes
	returnToMode        NavigationMode // Where to return after save prompt

	// Menu editing state
	editingMenu          *database.Menu        // Currently editing menu
	editingMenuCommand   *database.MenuCommand // Currently editing menu command
	selectedCommandIndex int                   // Selected command in modify mode
	currentMenuTab       int                   // Current tab in menu modify mode (0=Menu Data, 1=Menu Commands)
	menuDataFields       []SubmenuItem         // Preserve menu data fields when editing commands

	// racking original state:
	originalMenu         *database.Menu         // Original menu before editing
	originalMenuCommands []database.MenuCommand // Original commands before editing
	menuModified         bool                   // Track if menu has been modified

	// ANSI art (CP437) rendering state
	ansiArtLines []string // pre-split, padded lines (expected 80x25)
}

// MessageType defines the type of status message
type MessageType int

const (
	InfoMessage MessageType = iota
	SuccessMessage
	WarningMessage
	ErrorMessage
)

// MenuBar represents the horizontal top-level navigation
type MenuBar struct {
	Items []MenuCategory
}

// MenuCategory represents a top-level menu item with submenu
type MenuCategory struct {
	ID       string        // Unique identifier (e.g., "configuration")
	Label    string        // Display name (e.g., "Configuration")
	HotKey   rune          // Keyboard shortcut (e.g., 'C')
	SubItems []SubmenuItem // Items in the submenu
}

// SubmenuItem represents an item in the vertical dropdown menu
type SubmenuItem struct {
	ID           string          // Unique identifier
	Label        string          // Display label
	ItemType     SubmenuItemType // Type of item
	EditableItem *MenuItem       // If editable, links to MenuItem
	SubItems     []SubmenuItem   // If section header, nested items
}

// SubmenuItemType defines the type of submenu item
type SubmenuItemType int

const (
	SectionHeader SubmenuItemType = iota // Non-editable section divider
	EditableField                        // Editable configuration value
	ActionItem                           // Triggers an action (e.g., "View Logs")
)

// MenuItem represents an editable configuration value
type MenuItem struct {
	ID         string         // Unique identifier
	Label      string         // Display label
	Field      ConfigField    // Link to configuration field
	ValueType  ValueType      // Data type
	Validation ValidationFunc // Validation function
	HelpText   string         // Help text for editing
}

// ConfigField provides access to configuration values
type ConfigField struct {
	GetValue func() interface{}      // Getter function
	SetValue func(interface{}) error // Setter function with validation
}

// ValueType defines supported data types
type ValueType int

const (
	StringValue ValueType = iota
	IntValue
	BoolValue
	ListValue // Comma-separated list
	PortValue // Integer with port range validation
	PathValue // File/directory path
)

// ValidationFunc validates input based on value type
type ValidationFunc func(interface{}) error

// ============================================================================
// Custom List Types for bubbles/list
// ============================================================================

// submenuListItem implements list.Item interface for submenu items
type submenuListItem struct {
	submenuItem SubmenuItem
}

func (i submenuListItem) FilterValue() string {
	return i.submenuItem.Label
}

// submenuDelegate implements list.ItemDelegate for custom rendering
type submenuDelegate struct {
	maxWidth int
}

func (d submenuDelegate) Height() int                             { return 1 }
func (d submenuDelegate) Spacing() int                            { return 0 }
func (d submenuDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d submenuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(submenuListItem)
	if !ok {
		return
	}

	var str string
	isSelected := index == m.Index()

	// Build item label with a simple bullet prefix
	itemText := " * " + item.submenuItem.Label
	// Truncate if text is too long
	if len(itemText) > d.maxWidth {
		itemText = itemText[:d.maxWidth-3] + "..."
	}

	// Calculate padding to fill to maxWidth
	padding := ""
	if len(itemText) < d.maxWidth {
		padding = strings.Repeat(" ", d.maxWidth-len(itemText))
	}

	// Render with modern colors
	if isSelected {
		// Selected: bright accent color
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextBright)).
			Background(lipgloss.Color(ColorAccent)).
			Bold(true).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	} else {
		// Unselected: subtle
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextNormal)).
			Background(lipgloss.Color(ColorBgMedium)).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	}

	fmt.Fprint(w, str)
}

// userListItem implements list.Item interface for user items
type userListItem struct {
	user database.UserRecord
}

func (i userListItem) FilterValue() string {
	return i.user.Username
}

// userDelegate implements list.ItemDelegate for custom user rendering
type userDelegate struct {
	maxWidth int
}

func (d userDelegate) Height() int                             { return 1 }
func (d userDelegate) Spacing() int                            { return 0 }
func (d userDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d userDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(userListItem)
	if !ok {
		return
	}

	var str string
	isSelected := index == m.Index()

	// Format with fixed-width columns: Username (15 chars), Level (5 chars), UID (5 chars)
	itemText := fmt.Sprintf(" %-15s %-5d %-5d", item.user.Username, item.user.SecurityLevel, item.user.ID)

	// Truncate if text is too long
	if len(itemText) > d.maxWidth {
		itemText = itemText[:d.maxWidth-3] + "..."
	}

	// Calculate padding to fill to maxWidth
	padding := ""
	if len(itemText) < d.maxWidth {
		padding = strings.Repeat(" ", d.maxWidth-len(itemText))
	}

	// Render with modern colors
	if isSelected {
		// Selected: bright accent color
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextBright)).
			Background(lipgloss.Color(ColorAccent)).
			Bold(true).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	} else {
		// Unselected: subtle
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextNormal)).
			Background(lipgloss.Color(ColorBgMedium)).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	}

	fmt.Fprint(w, str)
}

// securityLevelListItem implements list.Item interface for security level items
type securityLevelListItem struct {
	securityLevel database.SecurityLevelRecord
}

func (i securityLevelListItem) FilterValue() string {
	return i.securityLevel.Name
}

// menuListItem implements list.Item interface for menu items
type menuListItem struct {
	menu         database.Menu
	commandCount int
}

func (i menuListItem) FilterValue() string {
	return i.menu.Name
}

// menuDelegate implements list.ItemDelegate for custom menu rendering
type menuDelegate struct {
	maxWidth int
}

func (d menuDelegate) Height() int                             { return 1 }
func (d menuDelegate) Spacing() int                            { return 0 }
func (d menuDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d menuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(menuListItem)
	if !ok {
		return
	}

	var str string
	isSelected := index == m.Index()

	// Format: [MenuName] [CommandCount]
	itemText := fmt.Sprintf(" %s (%d commands)", item.menu.Name, item.commandCount)

	// Truncate if text is too long
	if len(itemText) > d.maxWidth {
		itemText = itemText[:d.maxWidth-3] + "..."
	}

	// Calculate padding to fill to maxWidth
	padding := ""
	if len(itemText) < d.maxWidth {
		padding = strings.Repeat(" ", d.maxWidth-len(itemText))
	}

	// Render with modern colors
	if isSelected {
		// Selected: bright accent color
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextBright)).
			Background(lipgloss.Color(ColorAccent)).
			Bold(true).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	} else {
		// Unselected: subtle
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextNormal)).
			Background(lipgloss.Color(ColorBgMedium)).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	}

	fmt.Fprint(w, str)
}

// securityLevelDelegate implements list.ItemDelegate for custom security level rendering
type securityLevelDelegate struct {
	maxWidth int
}

func (d securityLevelDelegate) Height() int                             { return 1 }
func (d securityLevelDelegate) Spacing() int                            { return 0 }
func (d securityLevelDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d securityLevelDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(securityLevelListItem)
	if !ok {
		return
	}

	var str string
	isSelected := index == m.Index()

	// Format: [SecLevel] [Label]
	itemText := fmt.Sprintf(" [%d] %s", item.securityLevel.SecLevel, item.securityLevel.Name)

	// Truncate if text is too long
	if len(itemText) > d.maxWidth {
		itemText = itemText[:d.maxWidth-3] + "..."
	}

	// Calculate padding to fill to maxWidth
	padding := ""
	if len(itemText) < d.maxWidth {
		padding = strings.Repeat(" ", d.maxWidth-len(itemText))
	}

	// Render with modern colors
	if isSelected {
		// Selected: bright accent color
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextBright)).
			Background(lipgloss.Color(ColorAccent)).
			Bold(true).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	} else {
		// Unselected: subtle
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextNormal)).
			Background(lipgloss.Color(ColorBgMedium)).
			Width(d.maxWidth)
		str = style.Render(itemText + padding)
	}

	fmt.Fprint(w, str)
}

// ============================================================================
// Menu Bar Component
// ============================================================================

// MenuBarComponent handles horizontal top-level navigation
type MenuBarComponent struct {
	items       []MenuCategory
	activeIndex int
	width       int
}

// Render generates the menu bar display
func (c *MenuBarComponent) Render() string {
	var items []string

	for i, item := range c.items {
		style := c.getMenuItemStyle(i == c.activeIndex)
		// Using clean format without hotkey to prevent stray characters
		label := item.Label
		items = append(items, style.Render(" "+label+" "))
	}

	// Join with decorative separator and center in available width
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(" | ")

	return lipgloss.NewStyle().
		Width(c.width).
		Align(lipgloss.Center).
		Render(strings.Join(items, separator))
}

// getMenuItemStyle returns the style for menu items based on active state
func (c *MenuBarComponent) getMenuItemStyle(active bool) lipgloss.Style {
	if active {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")). // White text
			Background(lipgloss.Color("4"))   // Blue background
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")). // Gray text
		Background(lipgloss.Color("8"))  // Dark gray background
}

// ============================================================================
// Configuration Mapping
// ============================================================================

func InitialModelV2(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "" // Remove prompt to prevent shifting
	ti.Placeholder = "Enter value"
	ti.CharLimit = 200
	// Width will be set dynamically in render functions
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	ti.Cursor.Blink = true

	// Set text input styling to match blue background (remove conflicting background)
	ti.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))
	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	menuBar := buildMenuStructure(cfg)

	// Initialize the submenu list with items from the first menu category
	listItems := buildListItems(menuBar.Items[0])

	// Calculate max width for light bar - narrow width for level 2 menus
	maxWidth := 25

	submenuList := list.New(listItems, submenuDelegate{maxWidth: maxWidth}, maxWidth, 10)
	submenuList.Title = ""
	submenuList.SetShowStatusBar(false)
	submenuList.SetFilteringEnabled(false)
	submenuList.SetShowHelp(false)
	submenuList.SetShowPagination(false)

	// Remove any default list styling/padding
	submenuList.Styles.Title = lipgloss.NewStyle()
	submenuList.Styles.PaginationStyle = lipgloss.NewStyle()
	submenuList.Styles.HelpStyle = lipgloss.NewStyle()

	// Use existing database connection from config system
	var db *database.SQLiteDB
	if existingDB := config.GetDatabase(); existingDB != nil {
		if sqliteDB, ok := existingDB.(*database.SQLiteDB); ok {
			db = sqliteDB
			// Ensure user schema is initialized
			if err := db.InitializeSchema(); err != nil {
				// If schema initialization fails, set db to nil
				db = nil
			}
		}
	}

	m := Model{
		config:         cfg,
		db:             db,
		navMode:        MainMenuNavigation,
		activeMenu:     0,
		menuBar:        menuBar,
		submenuList:    submenuList,
		textInput:      ti,
		screenWidth:    80,
		screenHeight:   24,
		currentMenuTab: 0, // Default to Menu Data tab
	}

	// Try to load a default ANSI art if present (CP437 encoded). Failures are ignored.
	_ = m.LoadANSIArtFromContent(configArt)

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	// Alt screen is enabled via tea.WithAltScreen in RunConfigEditorTUI.
	// Avoid duplicating screen/raw mode transitions on Windows.
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseValue converts string input to appropriate type
func parseValue(input string, valueType ValueType) (interface{}, error) {
	switch valueType {
	case StringValue:
		return input, nil
	case IntValue, PortValue:
		return strconv.Atoi(input)
	case BoolValue:
		switch strings.ToLower(input) {
		case "true", "yes", "y", "1":
			return true, nil
		case "false", "no", "n", "0":
			return false, nil
		default:
			return false, fmt.Errorf("invalid boolean value")
		}
	case ListValue:
		// Comma-separated list
		items := strings.Split(input, ",")
		for i := range items {
			items[i] = strings.TrimSpace(items[i])
		}
		return strings.Join(items, ", "), nil
	case PathValue:
		if input == "" {
			return "", fmt.Errorf("path cannot be empty")
		}
		return input, nil
	default:
		return nil, fmt.Errorf("unsupported value type")
	}
}

// formatValue converts a value to display string
func formatValue(value interface{}, valueType ValueType) string {
	switch valueType {
	case BoolValue:
		if value.(bool) {
			return "Yes"
		}
		return "No"
	default:
		return fmt.Sprintf("%v", value)
	}
}

// renderPersistentHeader renders the top header bar
func (m Model) renderPersistentHeader() string {

	// Right side: Version with decorative elements
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBgLight))

	versionAccent := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBgLight)).
		Bold(true)

	rightText := versionStyle.Render(" v") + versionAccent.Render("0.01") + versionStyle.Render(" ")

	// Calculate spacing with texture
	rightWidth := m.visualWidth(rightText)
	spacingNeeded := m.screenWidth - rightWidth

	// Plain spacing (no texture)
	spacing := ""
	if spacingNeeded > 0 {
		spacing = strings.Repeat(" ", spacingNeeded)
	}

	// Combine all parts
	header := spacing + rightText

	return header
}

// ============================================================================
// Entry Point
// ============================================================================

// RunConfigEditorTUI starts the configuration editor TUI
func RunConfigEditorTUI(cfg *config.Config) error {
	m := InitialModelV2(cfg)

	// Use alt screen only when running on an interactive TTY to avoid
	// Windows "making raw" errors when stdin/stdout are not consoles.
	var p *tea.Program
	if isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd()) {
		p = tea.NewProgram(m, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(m)
	}

	_, err := p.Run()
	return err
}
