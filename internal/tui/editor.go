package tui

import (
	"database/sql"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/database"
)

// ============================================================================
// Color Scheme & Styling
// ============================================================================

// Color palette for modern theme
const (
	// Primary colors
	ColorPrimary     = "4"  // Blue (was "39")
	ColorPrimaryDark = "4"  // Blue (was "33")
	ColorAccent      = "5"  // Magenta (was "170")
	ColorAccentLight = "13" // Bright Magenta (was "177")

	// Backgrounds
	ColorBgGrey   = "8" // Black (unchanged)
	ColorBgDark   = "0" // Black (unchanged)
	ColorBgMedium = "0" // Black (was "8")
	ColorBgLight  = "7" // White (unchanged)

	// Text colors
	ColorTextBright = "15" // Bright White (unchanged)
	ColorTextNormal = "7"  // White (was "252")
	ColorTextDim    = "8"  // Bright Black/Gray (was "243")
	ColorTextAccent = "13" // Bright Magenta (was "213")

	// Status colors
	ColorSuccess = "2" // Green (was "42")
	ColorWarning = "3" // Yellow (was "214")
	ColorError   = "1" // Red (was "196")
	ColorInfo    = "6" // Cyan (was "117")
)

// Background texture patterns
var (
	TexturePattern1 = "░" // Light shade
	TexturePattern2 = "▒" // Medium shade
	TexturePattern3 = "▓" // Dark shade
	TextureDot      = "·"
	TextureBox      = "▪"
)

// ============================================================================
// Core Data Structures & Menu Bar
// ============================================================================

// ============================================================================
// UI Layout Constants
// ============================================================================

const (
	// Level 2 menu positioning (upper-left)
	Level2StartRow = 2
	Level2StartCol = 2

	// Level 3 menu positioning (cascading from Level 2)
	Level3StartRow = 4
	Level3StartCol = 25

	// Alternatively, for bottom-right anchoring:
	// Set these to negative values to anchor from bottom-right
	// Level3AnchorBottom will position from bottom of screen
	// Level3AnchorRight will position from right of screen
	Level3AnchorBottom = true // Set to true to anchor to bottom
	Level3AnchorRight  = true // Set to true to anchor to right
	Level3BottomOffset = 3    // Rows from bottom (for footer/breadcrumb space)
	Level3RightOffset  = 2    // Columns from right edge
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
	menuList []database.Menu // List of menus for management

	// User management state
	editingUser *database.UserRecord // Currently editing user

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

	// Texture configuration
	texturePatterns []string       // ADD THIS
	textureStyle    lipgloss.Style // ADD THIS
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

	// Get item text with icon
	itemText := " ▪ " + item.submenuItem.Label

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

	// Format: [UserName] [SecLevel] [UserID] - tab separated for alignment
	itemText := fmt.Sprintf(" %s\t%d\t%d", item.user.Username, item.user.SecurityLevel, item.user.ID)

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

// buildMenuStructure constructs the complete menu hierarchy from existing config
func buildMenuStructure(cfg *config.Config) MenuBar {
	return MenuBar{
		Items: []MenuCategory{
			{
				ID:     "configuration",
				Label:  "Configuration",
				HotKey: 'C',
				SubItems: []SubmenuItem{
					{
						ID:       "paths",
						Label:    "Paths",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "database-path",
								Label:    "Database",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.paths.database",
									Label:     "Database",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Paths.Database },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Paths.Database = v.(string)
											return nil
										},
									},
									HelpText: "Path to database directory",
								},
							},
							{
								ID:       "file-base-path",
								Label:    "File Base",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.paths.file_base",
									Label:     "File Base",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Paths.FileBase },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Paths.FileBase = v.(string)
											return nil
										},
									},
									HelpText: "Path to file base directory",
								},
							},
							{
								ID:       "logs-path",
								Label:    "Logs",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.paths.logs",
									Label:     "Logs",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Paths.Logs },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Paths.Logs = v.(string)
											return nil
										},
									},
									HelpText: "Path to logs directory",
								},
							},
							{
								ID:       "message-base-path",
								Label:    "Message Base",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.paths.message_base",
									Label:     "Message Base",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Paths.MessageBase },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Paths.MessageBase = v.(string)
											return nil
										},
									},
									HelpText: "Path to message base directory",
								},
							},
							{
								ID:       "system-path",
								Label:    "System",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.paths.system",
									Label:     "System",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Paths.System },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Paths.System = v.(string)
											return nil
										},
									},
									HelpText: "Path to system directory",
								},
							},
							{
								ID:       "themes-path",
								Label:    "Themes",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.paths.themes",
									Label:     "Themes",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Paths.Themes },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Paths.Themes = v.(string)
											return nil
										},
									},
									HelpText: "Path to themes directory",
								},
							},
						},
					},
					{
						ID:       "general-settings",
						Label:    "General",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "bbs-name",
								Label:    "BBS Name",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.bbs_name",
									Label:     "BBS Name",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.BBSName },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.BBSName = v.(string)
											return nil
										},
									},
									HelpText: "Name of your BBS system",
								},
							},
							{
								ID:       "bbs-location",
								Label:    "BBS Location",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.bbs_location",
									Label:     "BBS Location",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.BBSLocation },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.BBSLocation = v.(string)
											return nil
										},
									},
									HelpText: "Location of your BBS (e.g., Berkeley, CA)",
								},
							},
							{
								ID:       "sysop-name",
								Label:    "SysOp Name",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.sysop_name",
									Label:     "SysOp Name",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.SysOpName },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.SysOpName = v.(string)
											return nil
										},
									},
									HelpText: "System operator name",
								},
							},
							{
								ID:       "sysop-timeout-exempt",
								Label:    "SysOp Timeout Exempt",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.sysop_timeout_exempt",
									Label:     "SysOp Timeout Exempt",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.SysOpTimeoutExempt },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.SysOpTimeoutExempt = v.(bool)
											return nil
										},
									},
									HelpText: "Exempt sysop from idle timeout",
								},
							},
							{
								ID:       "system-password",
								Label:    "System Password",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.system_password",
									Label:     "System Password",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.SystemPassword },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.SystemPassword = v.(string)
											return nil
										},
									},
									HelpText: "System password for protected areas",
								},
							},
							{
								ID:       "timeout-minutes",
								Label:    "Timeout Minutes",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.timeout_minutes",
									Label:     "Timeout Minutes",
									ValueType: IntValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.TimeoutMinutes },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.TimeoutMinutes = v.(int)
											return nil
										},
									},
									HelpText: "Idle timeout in minutes",
									Validation: func(v interface{}) error {
										timeout := v.(int)
										if timeout <= 0 {
											return fmt.Errorf("timeout must be positive")
										}
										return nil
									},
								},
							},
							{
								ID:       "default-theme",
								Label:    "Default Theme",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.default_theme",
									Label:     "Default Theme",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.DefaultTheme },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.DefaultTheme = v.(string)
											return nil
										},
									},
									HelpText: "Default theme name",
								},
							},
							{
								ID:       "start-menu",
								Label:    "Start Menu",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.general.start_menu",
									Label:     "Start Menu",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.General.StartMenu },
										SetValue: func(v interface{}) error {
											cfg.Configuration.General.StartMenu = v.(string)
											return nil
										},
									},
									HelpText: "Initial menu to display (e.g., prelogin)",
								},
							},
						},
					},
					{
						ID:       "new-users",
						Label:    "New Users",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "allow-new",
								Label:    "Allow New Users",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.new_users.allow_new",
									Label:     "Allow New Users",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.NewUsers.AllowNew },
										SetValue: func(v interface{}) error {
											cfg.Configuration.NewUsers.AllowNew = v.(bool)
											return nil
										},
									},
									HelpText: "Allow new user registration",
								},
							},
							{
								ID:       "ask-real-name",
								Label:    "Ask Real Name",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.new_users.ask_real_name",
									Label:     "Ask Real Name",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.NewUsers.AskRealName },
										SetValue: func(v interface{}) error {
											cfg.Configuration.NewUsers.AskRealName = v.(bool)
											return nil
										},
									},
									HelpText: "Ask for real name during registration",
								},
							},
							{
								ID:       "ask-location",
								Label:    "Ask Location",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.new_users.ask_location",
									Label:     "Ask Location",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.NewUsers.AskLocation },
										SetValue: func(v interface{}) error {
											cfg.Configuration.NewUsers.AskLocation = v.(bool)
											return nil
										},
									},
									HelpText: "Ask for location during registration",
								},
							},
						},
					},
					{
						ID:       "auth-settings",
						Label:    "Auth Persistence",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "auth-max-failed-attempts",
								Label:    "Max Failed Attempts",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.auth.max_failed_attempts",
									Label:     "Max Failed Attempts",
									ValueType: IntValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Auth.MaxFailedAttempts },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Auth.MaxFailedAttempts = v.(int)
											return nil
										},
									},
									HelpText: "Failed login attempts before account lockout",
								},
							},
							{
								ID:       "auth-account-lock-minutes",
								Label:    "Account Lock Minutes",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.auth.account_lock_minutes",
									Label:     "Account Lock Minutes",
									ValueType: IntValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Auth.AccountLockMinutes },
										SetValue: func(v interface{}) error {
											cfg.Configuration.Auth.AccountLockMinutes = v.(int)
											return nil
										},
									},
									HelpText: "Duration (minutes) to lock account after reaching the max failed attempts",
								},
							},
							{
								ID:       "auth-password-algorithm",
								Label:    "Password Algorithm",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "config.auth.password_algorithm",
									Label:     "Password Algorithm",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Configuration.Auth.PasswordAlgorithm },
										SetValue: func(v interface{}) error {
											algo := strings.TrimSpace(v.(string))
											if algo == "" {
												return fmt.Errorf("password algorithm cannot be empty")
											}
											cfg.Configuration.Auth.PasswordAlgorithm = algo
											return nil
										},
									},
									HelpText: "Default hashing algorithm for persisted passwords",
								},
							},
						},
					},
				},
			},
			{
				ID:     "servers",
				Label:  "Servers",
				HotKey: 'S',
				SubItems: []SubmenuItem{
					{
						ID:       "general-settings",
						Label:    "General Settings",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "max-nodes",
								Label:    "Max Nodes",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.general.max_nodes",
									Label:     "Max Nodes",
									ValueType: IntValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.GeneralSettings.MaxNodes },
										SetValue: func(v interface{}) error {
											cfg.Servers.GeneralSettings.MaxNodes = v.(int)
											return nil
										},
									},
									HelpText: "Maximum number of concurrent connections",
									Validation: func(v interface{}) error {
										nodes := v.(int)
										if nodes <= 0 {
											return fmt.Errorf("max nodes must be positive")
										}
										return nil
									},
								},
							},
							{
								ID:       "max-connections-per-ip",
								Label:    "Max Connections/IP",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.general.max_connections_per_ip",
									Label:     "Max Connections per IP",
									ValueType: IntValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.GeneralSettings.MaxConnectionsPerIP },
										SetValue: func(v interface{}) error {
											cfg.Servers.GeneralSettings.MaxConnectionsPerIP = v.(int)
											return nil
										},
									},
									HelpText: "Maximum connections allowed per IP address",
								},
							},
						},
					},
					{
						ID:       "telnet-server",
						Label:    "Telnet Server",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "telnet-active",
								Label:    "Active",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.telnet.active",
									Label:     "Active",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Telnet.Active },
										SetValue: func(v interface{}) error {
											cfg.Servers.Telnet.Active = v.(bool)
											return nil
										},
									},
									HelpText: "Enable/disable telnet server",
								},
							},
							{
								ID:       "telnet-port",
								Label:    "Port",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.telnet.port",
									Label:     "Telnet Port",
									ValueType: PortValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Telnet.Port },
										SetValue: func(v interface{}) error {
											cfg.Servers.Telnet.Port = v.(int)
											return nil
										},
									},
									HelpText: "TCP port for telnet connections (1-65535)",
									Validation: func(v interface{}) error {
										port := v.(int)
										if port < 1 || port > 65535 {
											return fmt.Errorf("port must be between 1 and 65535")
										}
										return nil
									},
								},
							},
						},
					},
					{
						ID:       "security-rate-limits",
						Label:    "Rate Limits",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "rate-limit-enabled",
								Label:    "Rate Limiting Enabled",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.rate_limits.enabled",
									Label:     "Rate Limiting Enabled",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.RateLimits.Enabled },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.RateLimits.Enabled = v.(bool)
											return nil
										},
									},
									HelpText: "Enable connection rate limiting",
								},
							},
							{
								ID:       "rate-limit-window",
								Label:    "Window Minutes",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.rate_limits.window_minutes",
									Label:     "Window Minutes",
									ValueType: IntValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.RateLimits.WindowMinutes },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.RateLimits.WindowMinutes = v.(int)
											return nil
										},
									},
									HelpText: "Rate limit time window in minutes",
								},
							},
						},
					},
					{
						ID:       "security-local-lists",
						Label:    "Local Lists",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "blacklist-enabled",
								Label:    "Blacklist Enabled",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.local_lists.blacklist_enabled",
									Label:     "Blacklist Enabled",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.BlacklistEnabled },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.LocalLists.BlacklistEnabled = v.(bool)
											return nil
										},
									},
									HelpText: "Enable IP blacklisting",
								},
							},
							{
								ID:       "blacklist-file",
								Label:    "Blacklist File",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.local_lists.blacklist_file",
									Label:     "Blacklist File",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.BlacklistFile },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.LocalLists.BlacklistFile = v.(string)
											return nil
										},
									},
									HelpText: "Path to blacklist file",
								},
							},
							{
								ID:       "whitelist-enabled",
								Label:    "Whitelist Enabled",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.local_lists.whitelist_enabled",
									Label:     "Whitelist Enabled",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.WhitelistEnabled },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.LocalLists.WhitelistEnabled = v.(bool)
											return nil
										},
									},
									HelpText: "Enable IP whitelisting",
								},
							},
							{
								ID:       "whitelist-file",
								Label:    "Whitelist File",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.local_lists.whitelist_file",
									Label:     "Whitelist File",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.WhitelistFile },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.LocalLists.WhitelistFile = v.(string)
											return nil
										},
									},
									HelpText: "Path to whitelist file",
								},
							},
						},
					},
					{
						ID:       "security-external-lists",
						Label:    "External Lists",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "external-block-enabled",
								Label:    "External Block Enabled",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.external_lists.enabled",
									Label:     "External Block Enabled",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.ExternalLists.Enabled },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.ExternalLists.Enabled = v.(bool)
											return nil
										},
									},
									HelpText: "Enable external blocklist checking",
								},
							},
							{
								ID:       "external-blocklist-urls",
								Label:    "Blocklist URLs",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.external_lists.urls",
									Label:     "Blocklist URLs",
									ValueType: ListValue,
									Field: ConfigField{
										GetValue: func() interface{} {
											return strings.Join(cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs, ", ")
										},
										SetValue: func(v interface{}) error {
											s := v.(string)
											cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs = nil
											for _, u := range strings.Split(s, ",") {
												url := strings.TrimSpace(u)
												if url != "" {
													cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs = append(cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs, url)
												}
											}
											return nil
										},
									},
									HelpText: "Comma-separated list of blocklist URLs",
								},
							},
						},
					},
					{
						ID:       "security-geo-block",
						Label:    "Geo Blocking",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "geo-block-enabled",
								Label:    "Geo Block Enabled",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.geo_block.enabled",
									Label:     "Geo Block Enabled",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.GeoBlockEnabled },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.GeoBlock.GeoBlockEnabled = v.(bool)
											return nil
										},
									},
									HelpText: "Enable geographic IP blocking",
								},
							},
							{
								ID:       "blocked-countries",
								Label:    "Blocked Countries",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.geo_block.blocked_countries",
									Label:     "Blocked Countries",
									ValueType: ListValue,
									Field: ConfigField{
										GetValue: func() interface{} { return strings.Join(cfg.Servers.Security.GeoBlock.BlockedCountries, ", ") },
										SetValue: func(v interface{}) error {
											s := v.(string)
											cfg.Servers.Security.GeoBlock.BlockedCountries = nil
											for _, country := range strings.Split(s, ",") {
												country = strings.TrimSpace(country)
												if country != "" {
													cfg.Servers.Security.GeoBlock.BlockedCountries = append(cfg.Servers.Security.GeoBlock.BlockedCountries, country)
												}
											}
											return nil
										},
									},
									HelpText: "Comma-separated country codes to block",
								},
							},
							{
								ID:       "allowed-countries",
								Label:    "Allowed Countries",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.geo_block.allowed_countries",
									Label:     "Allowed Countries",
									ValueType: ListValue,
									Field: ConfigField{
										GetValue: func() interface{} { return strings.Join(cfg.Servers.Security.GeoBlock.AllowedCountries, ", ") },
										SetValue: func(v interface{}) error {
											s := v.(string)
											cfg.Servers.Security.GeoBlock.AllowedCountries = nil
											for _, country := range strings.Split(s, ",") {
												country = strings.TrimSpace(country)
												if country != "" {
													cfg.Servers.Security.GeoBlock.AllowedCountries = append(cfg.Servers.Security.GeoBlock.AllowedCountries, country)
												}
											}
											return nil
										},
									},
									HelpText: "Comma-separated country codes to allow (leave empty for all)",
								},
							},
							{
								ID:       "geo-api-provider",
								Label:    "Geo API Provider",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.geo_block.api_provider",
									Label:     "Geo API Provider",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.GeoAPIProvider },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.GeoBlock.GeoAPIProvider = v.(string)
											return nil
										},
									},
									HelpText: "Geolocation API provider (e.g., ipapi)",
								},
							},
							{
								ID:       "geo-api-key",
								Label:    "Geo API Key",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.geo_block.api_key",
									Label:     "Geo API Key",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.GeoAPIKey },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.GeoBlock.GeoAPIKey = v.(string)
											return nil
										},
									},
									HelpText: "API key for geolocation service",
								},
							},
							{
								ID:       "threat-intel-enabled",
								Label:    "Threat Intel Enabled",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.geo_block.threat_intel",
									Label:     "Threat Intel Enabled",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.ThreatIntelEnabled },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.GeoBlock.ThreatIntelEnabled = v.(bool)
											return nil
										},
									},
									HelpText: "Enable threat intelligence checking",
								},
							},
							{
								ID:       "blocklist-update-hours",
								Label:    "Blocklist Update Hours",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.geo_block.update_hours",
									Label:     "Blocklist Update Hours",
									ValueType: IntValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.BlocklistUpdateHours },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.GeoBlock.BlocklistUpdateHours = v.(int)
											return nil
										},
									},
									HelpText: "Hours between blocklist updates",
								},
							},
						},
					},
					{
						ID:       "security-logs",
						Label:    "Security Logs",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "log-security-events",
								Label:    "Log Security Events",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.logs.log_events",
									Label:     "Log Security Events",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.Logs.LogSecurityEvents },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.Logs.LogSecurityEvents = v.(bool)
											return nil
										},
									},
									HelpText: "Log security-related events",
								},
							},
							{
								ID:       "log-blocked-attempts",
								Label:    "Log Blocked Attempts",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.logs.log_blocked",
									Label:     "Log Blocked Attempts",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.Logs.LogBlockedAttempts },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.Logs.LogBlockedAttempts = v.(bool)
											return nil
										},
									},
									HelpText: "Log blocked connection attempts",
								},
							},
							{
								ID:       "security-log-file",
								Label:    "Security Log File",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "servers.security.logs.log_file",
									Label:     "Security Log File",
									ValueType: PathValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Servers.Security.Logs.SecurityLogFile },
										SetValue: func(v interface{}) error {
											cfg.Servers.Security.Logs.SecurityLogFile = v.(string)
											return nil
										},
									},
									HelpText: "Path to security log file",
								},
							},
						},
					},
				},
			},
			{
				ID:       "networking",
				Label:    "Networking",
				HotKey:   'N',
				SubItems: []SubmenuItem{
					// TODO: Add network configuration when needed
				},
			},
			{
				ID:     "editors",
				Label:  "Editors",
				HotKey: 'E',
				SubItems: []SubmenuItem{
					{
						ID:       "user-editor",
						Label:    "Users",
						ItemType: ActionItem,
					},
					{
						ID:       "security-levels-editor",
						Label:    "Security Levels",
						ItemType: ActionItem,
					},
					{
						ID:       "menu-editor",
						Label:    "Menus",
						ItemType: ActionItem,
					},
				},
			},
			{
				ID:     "other",
				Label:  "Other",
				HotKey: 'O',
				SubItems: []SubmenuItem{
					{
						ID:       "discord-integration",
						Label:    "Discord Integration",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							{
								ID:       "discord-enabled",
								Label:    "Discord Enabled",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "other.discord.enabled",
									Label:     "Discord Enabled",
									ValueType: BoolValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Other.Discord.Enabled },
										SetValue: func(v interface{}) error {
											cfg.Other.Discord.Enabled = v.(bool)
											return nil
										},
									},
									HelpText: "Enable Discord webhook integration",
								},
							},
							{
								ID:       "discord-webhook-url",
								Label:    "Discord Webhook URL",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "other.discord.webhook_url",
									Label:     "Discord Webhook URL",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Other.Discord.WebhookURL },
										SetValue: func(v interface{}) error {
											cfg.Other.Discord.WebhookURL = v.(string)
											return nil
										},
									},
									HelpText: "Discord webhook URL for notifications",
								},
							},
							{
								ID:       "discord-username",
								Label:    "Discord Username",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "other.discord.username",
									Label:     "Discord Username",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Other.Discord.Username },
										SetValue: func(v interface{}) error {
											cfg.Other.Discord.Username = v.(string)
											return nil
										},
									},
									HelpText: "Bot username for Discord notifications",
								},
							},
							{
								ID:       "discord-title",
								Label:    "Discord Title",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "other.discord.title",
									Label:     "Discord Title",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Other.Discord.Title },
										SetValue: func(v interface{}) error {
											cfg.Other.Discord.Title = v.(string)
											return nil
										},
									},
									HelpText: "Title for Discord notifications",
								},
							},
							{
								ID:       "discord-invite-url",
								Label:    "Discord Invite URL",
								ItemType: EditableField,
								EditableItem: &MenuItem{
									ID:        "other.discord.invite_url",
									Label:     "Discord Invite URL",
									ValueType: StringValue,
									Field: ConfigField{
										GetValue: func() interface{} { return cfg.Other.Discord.InviteURL },
										SetValue: func(v interface{}) error {
											cfg.Other.Discord.InviteURL = v.(string)
											return nil
										},
									},
									HelpText: "Discord server invite URL",
								},
							},
						},
					},
				},
			},
		},
	}
}

// ============================================================================
// Initialization
// ============================================================================

// buildListItems creates list items from a menu category's submenu items
// Sub-menu should contain section headers and action items, not individual fields
func buildListItems(category MenuCategory) []list.Item {
	var items []list.Item

	for _, section := range category.SubItems {
		// Add section headers and action items to the submenu
		// Individual fields will be shown in a modal after selecting a section
		if section.ItemType == SectionHeader || section.ItemType == ActionItem {
			items = append(items, submenuListItem{submenuItem: section})
		}
	}

	return items
}

func InitialModelV2(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "" // Remove prompt to prevent shifting
	ti.Placeholder = "Enter value"
	ti.CharLimit = 200
	ti.Width = 25 // Fixed width to prevent wrapping, allows scrolling
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	ti.Cursor.Blink = true

	// Set text input styling to match blue background
	ti.TextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33"))
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33"))
	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("33"))

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

	// Initialize texture configuration
	texturePatterns := []string{TexturePattern2, TexturePattern2, TexturePattern2, TexturePattern2}
	textureStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("0"))

	return Model{
		config:          cfg,
		db:              db,
		navMode:         MainMenuNavigation,
		activeMenu:      0,
		menuBar:         menuBar,
		submenuList:     submenuList,
		textInput:       ti,
		screenWidth:     80,
		screenHeight:    24,
		texturePatterns: texturePatterns, // ADD THIS
		textureStyle:    textureStyle,    // ADD THIS
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, tea.EnterAltScreen)
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
				m.textInput.SetValue(formatValue(currentValue, m.editingItem.ValueType))

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
					m.textInput.Width = 25 // Fixed width for scrolling
				case ListValue:
					m.textInput.Placeholder = "comma,separated,values"
					m.textInput.CharLimit = 200
					m.textInput.Width = 25 // Fixed width for scrolling
				default: // StringValue
					m.textInput.Placeholder = "Enter value"
					m.textInput.CharLimit = 200
					m.textInput.Width = 25 // Fixed width for scrolling
				}
			}

			m.textInput.Focus()
			m.navMode = EditingValue
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

			return m, nil
		}

		// No unsaved changes, exit normally
		if m.editingSecurityLevel != nil {
			// Return to security levels list
			m.navMode = SecurityLevelsMode
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
			m.editingSecurityLevel = nil
			m.editingUser = nil
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

// handleSaveChangesPrompt processes input during save changes confirmation
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
			if m.editingSecurityLevel != nil {
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
				m.editingSecurityLevel = nil // Clear editing state
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
				m.editingUser = nil // Clear editing state
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
			// DON'T show success message - just reset counter
			m.modifiedCount = 0 // Reset counter
		}
		// Either way, return to previous mode
		m.savePrompt = false
		m.navMode = m.returnToMode
		m.editingSecurityLevel = nil // Clear editing state
		m.editingUser = nil          // Clear editing state

		// Clean up modal if returning to Level 2
		if m.returnToMode == Level2MenuNavigation {
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
		}
		return m, nil
	case "y", "Y":
		// Save and return
		if err := config.SaveConfig(m.config, ""); err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
			m.messageTime = time.Now()
			m.messageType = ErrorMessage
		} else {
			// DON'T show success message - just reset counter
			m.modifiedCount = 0
		}
		m.savePrompt = false
		m.editingSecurityLevel = nil
		m.editingUser = nil
		m.navMode = m.returnToMode
		if m.returnToMode == Level2MenuNavigation {
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
		}
		return m, nil
	case "n", "N":
		// Don't save, just return
		m.savePrompt = false
		m.editingSecurityLevel = nil
		m.editingUser = nil
		m.navMode = m.returnToMode
		if m.returnToMode == Level2MenuNavigation {
			m.modalFields = nil
			m.modalFieldIndex = 0
			m.modalSectionName = ""
		}
		return m, nil
	case "esc":
		// Cancel - return to modal
		m.savePrompt = false
		m.editingSecurityLevel = nil
		m.editingUser = nil
		m.navMode = Level4ModalNavigation
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
				if item.submenuItem.ID == "user-editor" {
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
				} else if item.submenuItem.ID == "security-levels-editor" {
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
				} else if item.submenuItem.ID == "menu-editor" {
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
				} else {
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
				m.textInput.SetValue(formatValue(currentValue, m.editingItem.ValueType))

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
					m.textInput.Width = 25 // Fixed width for scrolling
				case ListValue:
					m.textInput.Placeholder = "comma,separated,values"
					m.textInput.CharLimit = 200
					m.textInput.Width = 25 // Fixed width for scrolling
				default: // StringValue
					m.textInput.Placeholder = "Enter value"
					m.textInput.CharLimit = 200
					m.textInput.Width = 25 // Fixed width for scrolling
				}
			}

			m.textInput.Focus()
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

// renderSavePrompt renders the save confirmation dialog with lightbar selection
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
		// Saving changes before exiting modal
		header = headerStyle.Render(fmt.Sprintf("▸ Save Changes? (%d) ◂", m.modifiedCount))
		promptMsg = "Save changes before exiting?"
	} else {
		// Exiting application
		if m.modifiedCount > 0 {
			header = headerStyle.Render(fmt.Sprintf("▸ Exit Config? (%d changes) ◂", m.modifiedCount))
			promptMsg = "You have unsaved changes. Save before exiting?"
		} else {
			header = headerStyle.Render("▸ Exit Config? ◂")
			promptMsg = "Exit configuration editor?"
		}
	}

	// Create separator style - ADD THIS
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(modalWidth)

	separator := separatorStyle.Render(strings.Repeat("─", modalWidth))

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
	footerSeparator := separatorStyle.Render(strings.Repeat("─", modalWidth))

	// Build the modal
	allLines := []string{header, separator, promptLine, optionsLine, footerSeparator}
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
						HelpText: "User's first name (optional)",
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
						HelpText: "User's last name (optional)",
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
					ID:       "user-country",
					Label:    "Country",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "user-country",
						Label:     "Country",
						ValueType: StringValue,
						Field: ConfigField{
							GetValue: func() interface{} {
								if user.user.Country.Valid {
									return user.user.Country.String
								}
								return ""
							},
							SetValue: func(v interface{}) error {
								s := v.(string)
								if s == "" {
									user.user.Country = sql.NullString{Valid: false}
								} else {
									user.user.Country = sql.NullString{String: s, Valid: true}
								}
								return nil
							},
						},
						HelpText: "User's country (optional)",
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
						Label:     "Minutes per Day",
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
						Label:     "Timeout Minutes",
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
					Label:    "Can Delete Own Messages",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-can-delete-own-msgs",
						Label:     "Can Delete Own Messages",
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
					Label:    "Can Delete Any Messages",
					ItemType: EditableField,
					EditableItem: &MenuItem{
						ID:        "security-level-can-delete-msgs",
						Label:     "Can Delete Any Messages",
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
	case "enter":
		// Select menu - show menu details
		selectedItem := m.menuListUI.SelectedItem()
		if selectedItem != nil {
			menuItem := selectedItem.(menuListItem)
			// Get command count
			commands, err := m.db.GetMenuCommands(menuItem.menu.ID)
			commandCount := 0
			if err == nil {
				commandCount = len(commands)
			}
			m.message = fmt.Sprintf("Menu '%s': %d commands, prompt: %s", menuItem.menu.Name, commandCount, menuItem.menu.Prompt)
			m.messageTime = time.Now()
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
		m.menuListUI, cmd = m.menuListUI.Update(msg)
	}
	return m, cmd
}

// handleEditingValue processes input while editing a value in the modal
func (m Model) handleEditingValue(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle bool toggle separately
	if m.editingItem.ValueType == BoolValue {
		switch msg.String() {
		case "y", "Y":
			// Set to true
			newValue := true
			if newValue != m.originalValue {
				if err := m.editingItem.Field.SetValue(newValue); err != nil {
					m.editingError = err.Error()
				} else {
					m.modifiedCount++
					m.editingError = ""
					m.message = ""
					// Return to modal field selection
					m.navMode = Level4ModalNavigation
				}
			} else {
				// No change, just return
				m.navMode = Level4ModalNavigation
			}
			return m, nil
		case "n", "N":
			// Set to false
			newValue := false
			if newValue != m.originalValue {
				if err := m.editingItem.Field.SetValue(newValue); err != nil {
					m.editingError = err.Error()
				} else {
					m.modifiedCount++
					m.editingError = ""
					m.message = ""
					// Return to modal field selection
					m.navMode = Level4ModalNavigation
				}
			} else {
				// No change, just return
				m.navMode = Level4ModalNavigation
			}
			return m, nil
		case "enter", " ", "tab":
			// Toggle current value
			currentValue := m.editingItem.Field.GetValue().(bool)
			if err := m.editingItem.Field.SetValue(!currentValue); err != nil {
				m.editingError = err.Error()
			} else {
				m.editingError = ""
				// Don't exit, just update display to show toggled value
			}
			return m, nil
		case "esc":
			// Cancel - restore original value
			m.editingItem.Field.SetValue(m.originalValue)
			m.navMode = Level4ModalNavigation
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

		// Save the value
		if err := m.editingItem.Field.SetValue(parsedValue); err != nil {
			m.editingError = fmt.Sprintf("Error saving: %v", err)
			return m, nil
		}

		// Success - silently save and return
		m.modifiedCount++
		m.editingError = ""
		m.message = ""

		// Return to modal field selection
		m.navMode = Level4ModalNavigation
		m.textInput.Blur()

	case "esc":
		// Cancel editing - restore original value
		m.editingItem.Field.SetValue(m.originalValue)
		m.navMode = Level4ModalNavigation
		m.editingError = ""
		m.message = ""
		m.textInput.Blur()
	}

	return m, nil
}

// ============================================================================
// Canvas-Based Rendering System
// ============================================================================

// overlayString places a string onto the canvas at the given position
func (m *Model) overlayString(canvas []string, str string, startRow, startCol int) {
	lines := strings.Split(str, "\n")

	for i, line := range lines {
		row := startRow + i
		if row < 0 || row >= len(canvas) {
			continue
		}

		if startCol <= 0 {
			// Simple case: replace entire line
			canvas[row] = line
		} else {
			// Complex case: rebuild line with texture on sides
			// Get visual width of the line we're adding
			lineVisualWidth := m.visualWidth(line)

			// Build left texture
			leftTexture := ""
			if startCol > 0 {
				pattern := m.texturePatterns[row%len(m.texturePatterns)]
				leftTexture = m.textureStyle.Render(strings.Repeat(pattern, startCol))
			}

			// Build right texture
			rightTexture := ""
			endCol := startCol + lineVisualWidth
			if endCol < m.screenWidth {
				remainingWidth := m.screenWidth - endCol
				pattern := m.texturePatterns[row%len(m.texturePatterns)]
				rightTexture = m.textureStyle.Render(strings.Repeat(pattern, remainingWidth))
			}

			// Combine: left texture + content + right texture
			canvas[row] = leftTexture + line + rightTexture
		}
	}
}

func (m *Model) overlayStringCenteredWithClear(canvas []string, str string) {
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return
	}

	// Find the widest line
	maxWidth := 0
	for _, line := range lines {
		width := m.visualWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	startRow := (m.screenHeight - len(lines)) / 2
	if startRow < 0 {
		startRow = 0
	}
	startCol := (m.screenWidth - maxWidth) / 2
	if startCol < 0 {
		startCol = 0
	}

	// Just overlay directly - the texture is already in the canvas
	// No need to clear/recreate it
	m.overlayString(canvas, str, startRow, startCol)
}

// overlayStringCentered places a string in the center of the canvas
func (m *Model) overlayStringCentered(canvas []string, str string) {
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return
	}

	// Find the widest line (accounting for ANSI codes)
	maxWidth := 0
	for _, line := range lines {
		width := m.visualWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	startRow := (m.screenHeight - len(lines)) / 2
	if startRow < 0 {
		startRow = 0
	}
	startCol := (m.screenWidth - maxWidth) / 2
	if startCol < 0 {
		startCol = 0
	}

	m.overlayString(canvas, str, startRow, startCol)
}

// canvasToString converts the canvas back to a string
func (m *Model) canvasToString(canvas []string) string {
	return strings.Join(canvas, "\n")
}

// visualWidth calculates the display width of a string (excluding ANSI codes)
func (m *Model) visualWidth(s string) int {
	// Remove ANSI escape sequences
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	stripped := ansiPattern.ReplaceAllString(s, "")
	return len([]rune(stripped))
}

// ============================================================================
// View Rendering with Layered Approach
// ============================================================================

// View renders the complete UI using a layered canvas approach
func (m Model) View() string {
	// Create a base canvas with texture pattern
	canvas := make([]string, m.screenHeight)

	for i := range canvas {
		// Use texture pattern from struct
		pattern := m.texturePatterns[i%len(m.texturePatterns)]
		canvas[i] = m.textureStyle.Render(strings.Repeat(pattern, m.screenWidth))
	}

	// ALWAYS render persistent header at top (row 0)
	persistentHeader := m.renderPersistentHeader()
	m.overlayString(canvas, persistentHeader, 0, 0)

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

	// Layer 1: Main Menu (only shown when in MainMenuNavigation mode, centered)
	if m.navMode == MainMenuNavigation {
		mainMenuStr := m.renderMainMenu()
		m.overlayStringCentered(canvas, mainMenuStr)
	}

	// Layer 1.5: User Management (full screen mode)
	if m.navMode == UserManagementMode {
		userManagementStr := m.renderUserManagement()
		m.overlayStringCentered(canvas, userManagementStr)
		return m.canvasToString(canvas)
	}

	// Layer 1.6: Security Levels Management (full screen mode)
	if m.navMode == SecurityLevelsMode {
		securityLevelsStr := m.renderSecurityLevelsManagement()
		m.overlayStringCentered(canvas, securityLevelsStr)
		return m.canvasToString(canvas)
	}

	// Layer 1.7: Menu Management (full screen mode)
	if m.navMode == MenuManagementMode {
		menuManagementStr := m.renderMenuManagement()
		m.overlayStringCentered(canvas, menuManagementStr)
		return m.canvasToString(canvas)
	}

	// Layer 2: Submenu (visible from Level2 onwards, but not if showing modal)
	if m.navMode >= Level2MenuNavigation && !showAsModal {
		isDimmed := m.navMode > Level2MenuNavigation
		level2Str := m.renderLevel2Menu(isDimmed)
		// Start at row 2 to leave space for persistent header
		m.overlayString(canvas, level2Str, 2, Level2StartCol)
	}

	// Layer 3: Field list (visible from Level3 onwards, but only if NOT showing as modal)
	if m.navMode == Level3MenuNavigation && !showAsModal {
		// This means Level 3 has sub-sections to navigate
		level3Str := m.renderLevel3Menu(false)

		// Calculate bottom-right position
		lines := strings.Split(level3Str, "\n")
		height := len(lines)

		// Get width of the menu
		width := 0
		for _, line := range lines {
			lineWidth := m.visualWidth(line)
			if lineWidth > width {
				width = lineWidth
			}
		}

		// Position at bottom-right with offsets
		row := m.screenHeight - height - 4
		col := m.screenWidth - width - 2

		if row < 0 {
			row = 0
		}
		if col < 0 {
			col = 0
		}

		m.overlayString(canvas, level3Str, row, col)
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

// shouldRenderAsModal determines if current state should show a modal
func (m *Model) shouldRenderAsModal() bool {
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

	header := headerStyle.Render("⚡ Main Menu ⚡")

	// Create decorative separator with pattern
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorBgLight)).
		Width(menuWidth).
		Align(lipgloss.Center)
	separator := separatorStyle.Render(strings.Repeat("─", menuWidth))

	// Render menu items with left justification
	for i, category := range m.menuBar.Items {
		// Add icon prefix like other menus
		itemText := " ▪ " + category.Label

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

	// Create footer separator
	footerSeparator := separatorStyle.Render(strings.Repeat("─", menuWidth))

	// Build the full menu
	allLines := []string{header, separator}
	allLines = append(allLines, menuItems...)
	allLines = append(allLines, footerSeparator)

	menuContent := strings.Join(allLines, "\n")

	// Just wrap with background - NO BORDER, NO PADDING
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

	header := headerStyle.Render("▸ " + m.modalSectionName + " ◂")

	// Create separator row
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorAccent)).
		Width(modalWidth)
	separator := separatorStyle.Render(strings.Repeat("─", modalWidth))

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

				availableValueSpace := modalWidth - 26 - 1
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
	footerSeparator := separatorStyle.Render(strings.Repeat("─", modalWidth))

	// Build the full content
	allLines := []string{header, separator}
	allLines = append(allLines, fieldLines...)
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

	header := headerStyle.Render("▸ " + categoryLabel + " ◂")

	// Create decorative separator
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Width(maxWidth)

	if dimmed {
		separatorStyle = separatorStyle.Foreground(lipgloss.Color(ColorBgLight))
	} else {
		separatorStyle = separatorStyle.Foreground(lipgloss.Color(ColorPrimary))
	}

	separator := separatorStyle.Render(strings.Repeat("─", maxWidth))

	// Get list view and clean it up
	listView := m.submenuList.View()
	listView = strings.TrimSpace(listView) // Remove leading/trailing whitespace
	listLines := strings.Split(listView, "\n")

	// Create footer separator (same as header separator)
	footerSeparator := separatorStyle.Render(strings.Repeat("─", maxWidth))

	// Build as array like Main Menu does
	allLines := []string{header, separator}
	allLines = append(allLines, listLines...)
	allLines = append(allLines, footerSeparator)

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
			line := fmt.Sprintf(" ▪ %-18s %s", field.EditableItem.Label+":", valueStr)
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

	header := headerStyle.Render("▸ " + m.modalSectionName + " ◂")

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
	separator := separatorStyle.Render(strings.Repeat("─", maxFieldWidth))

	// Second pass: render fields with consistent width
	for i, field := range m.modalFields {
		if field.ItemType == EditableField && field.EditableItem != nil {
			currentValue := field.EditableItem.Field.GetValue()
			valueStr := formatValue(currentValue, field.EditableItem.ValueType)

			// Format with icon
			line := fmt.Sprintf(" ▪ %-18s %s", field.EditableItem.Label+":", valueStr)

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

	// Combine header, separator, and fields
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
		if err := m.seedDefaultMenu(); err != nil {
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

// seedDefaultMenu creates a basic MAIN menu for development
func (m *Model) seedDefaultMenu() error {
	// Create MAIN menu
	menu := &database.Menu{
		Name:                "MAIN",
		Titles:              []string{"-= Retrograde BBS =-", "-- Main Menu --"},
		HelpFile:            "",
		LongHelpFile:        "",
		Prompt:              "[@1 - @2]@MTime Left: [@V] (?=Help)@MMain Menu :",
		ACSRequired:         "",
		Password:            "",
		FallbackMenu:        "MAIN",
		ForcedHelpLevel:     0,
		GenericColumns:      4,
		GenericBracketColor: 1,
		GenericCommandColor: 9,
		GenericDescColor:    1,
		Flags:               "C---T-----",
	}

	menuID, err := m.db.CreateMenu(menu)
	if err != nil {
		return fmt.Errorf("failed to create MAIN menu: %w", err)
	}

	// Create menu commands
	commands := []database.MenuCommand{
		{
			MenuID:           int(menuID),
			CommandNumber:    1,
			Keys:             "R",
			LongDescription:  "(R)ead Mail - Read private Electronic mail",
			ShortDescription: "(R)ead Mail",
			ACSRequired:      "",
			CmdKeys:          "MM",
			Options:          "",
			Flags:            "",
		},
		{
			MenuID:           int(menuID),
			CommandNumber:    2,
			Keys:             "P",
			LongDescription:  "(P)ost Message - Post a message",
			ShortDescription: "(P)ost Message",
			ACSRequired:      "",
			CmdKeys:          "MP",
			Options:          "",
			Flags:            "",
		},
		{
			MenuID:           int(menuID),
			CommandNumber:    3,
			Keys:             "G",
			LongDescription:  "(G)oodbye - Logout and disconnect",
			ShortDescription: "(G)oodbye",
			ACSRequired:      "",
			CmdKeys:          "G",
			Options:          "",
			Flags:            "",
		},
	}

	for _, cmd := range commands {
		_, err := m.db.CreateMenuCommand(&cmd)
		if err != nil {
			return fmt.Errorf("failed to create menu command %s: %w", cmd.Keys, err)
		}
	}

	return nil
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

	header := headerStyle.Render(fmt.Sprintf("▸ User Management (%d users) ◂", len(m.userList)))

	// Create separator
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(42)
	separator := separatorStyle.Render(strings.Repeat("─", 42))

	// Get list view
	listView := m.userListUI.View()
	listView = strings.TrimSpace(listView)

	// Create footer separator
	footerSeparator := separatorStyle.Render(strings.Repeat("─", 42))

	// Build the full content
	allLines := []string{header, separator, listView, footerSeparator}

	combined := strings.Join(allLines, "\n")

	// Just wrap with background - NO BORDER, NO PADDING
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

	header := headerStyle.Render(fmt.Sprintf("▸ Security Levels Management (%d levels) ◂", len(m.securityLevelsList)))

	// Create separator
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(52)
	separator := separatorStyle.Render(strings.Repeat("─", 52))

	// Get list view
	listView := m.securityLevelsUI.View()
	listView = strings.TrimSpace(listView)

	// Create footer separator
	footerSeparator := separatorStyle.Render(strings.Repeat("─", 52))

	// Build the full content
	allLines := []string{header, separator, listView, footerSeparator}

	combined := strings.Join(allLines, "\n")

	// Just wrap with background - NO BORDER, NO PADDING
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

	header := headerStyle.Render(fmt.Sprintf("▸ Menu Management (%d menus) ◂", len(m.menuList)))

	// Create separator
	separatorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Foreground(lipgloss.Color(ColorPrimary)).
		Width(42)
	separator := separatorStyle.Render(strings.Repeat("─", 42))

	// Get list view
	listView := m.menuListUI.View()
	listView = strings.TrimSpace(listView)

	// Create footer separator
	footerSeparator := separatorStyle.Render(strings.Repeat("─", 42))

	// Build the full content
	allLines := []string{header, separator, listView, footerSeparator}

	combined := strings.Join(allLines, "\n")

	// Just wrap with background - NO BORDER, NO PADDING
	menuBox := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgMedium)).
		Render(combined)

	return menuBox
}

// renderFooter generates simple footer
func (m Model) renderFooter() string {
	footerText := "  F1  Help   ESC Exit"

	// Style the footer with full width
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextNormal)).
		Background(lipgloss.Color("8")).
		Width(m.screenWidth).
		Render(footerText)

	return footer
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

// renderPersistentHeader renders the top header bar with gradient and style
func (m Model) renderPersistentHeader() string {
	// Left side: Application name with gradient-like effect using different shades
	appName := "ReTRoGRaDe"
	bbsConfig := " BBS Config"

	leftStyle1 := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccent)).
		Background(lipgloss.Color(ColorBgGrey))

	leftStyle2 := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccentLight)).
		Background(lipgloss.Color(ColorBgGrey))

	leftText := leftStyle1.Render(" "+appName) + leftStyle2.Render(bbsConfig+" ")

	// Right side: Version with decorative elements
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Background(lipgloss.Color(ColorBgGrey))

	versionAccent := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBgGrey)).
		Bold(true)

	rightText := versionStyle.Render(" v") + versionAccent.Render("0.01") + versionStyle.Render(" ")

	// Calculate spacing with texture
	leftWidth := m.visualWidth(leftText)
	rightWidth := m.visualWidth(rightText)
	spacingNeeded := m.screenWidth - leftWidth - rightWidth

	// Create textured spacing
	spacing := ""
	if spacingNeeded > 0 {
		// Use subtle texture pattern
		textureStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBgGrey)).
			Foreground(lipgloss.Color("7")) // Very subtle
		spacing = textureStyle.Render(strings.Repeat(TextureDot, spacingNeeded))
	}

	// Combine all parts
	header := leftText + spacing + rightText

	return header
}

// ============================================================================
// Entry Point
// ============================================================================

// RunConfigEditorTUI starts the configuration editor TUI
func RunConfigEditorTUI(cfg *config.Config) error {
	p := tea.NewProgram(InitialModelV2(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
