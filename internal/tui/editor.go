package tui

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/robbiew/retrograde/internal/config"
)

// ============================================================================
// Core Data Structures & Menu Bar
// ============================================================================

// NavigationMode represents the current UI state
type NavigationMode int

const (
	MenuNavigation    NavigationMode = iota // Navigating horizontal menu bar
	SubmenuNavigation                       // Navigating vertical submenu overlay
	ModalFieldSelect                        // Selecting field in modal form
	EditingValue                            // Editing a value in modal form
	SavePrompt                              // Confirming save on exit
)

// Model represents the complete application state
type Model struct {
	// Configuration data
	config *config.Config

	// Navigation state
	navMode       NavigationMode
	activeMenu    int // Current horizontal menu index
	activeSubmenu int // Current submenu item index

	// Menu structure
	menuBar MenuBar

	// Submenu list
	submenuList list.Model

	// Modal form state
	modalFields      []SubmenuItem // All fields in the current section
	modalFieldIndex  int           // Currently selected field in modal
	modalSectionName string        // Section name for modal header

	// Editing state
	editing       bool
	editingItem   *MenuItem
	textInput     textinput.Model
	editingError  string
	originalValue interface{}

	// UI state
	screenWidth   int
	screenHeight  int
	message       string
	messageTime   time.Time
	messageType   MessageType
	savePrompt    bool
	quitting      bool
	modifiedCount int
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

	// Sub-menu displays ONLY section names as navigation items
	var str string
	isSelected := index == m.Index()

	// Calculate the width needed (including prefix)
	itemText := item.submenuItem.Label
	prefix := "- "
	if isSelected {
		prefix = "> "
	}
	contentWidth := len(prefix) + len(itemText)

	// Pad to max width so background extends full width
	padding := ""
	if d.maxWidth > contentWidth {
		padding = strings.Repeat(" ", d.maxWidth-contentWidth)
	}

	// All items in the sub-menu are section headers (navigation items)
	if isSelected {
		// Selected section: white on blue with indicator, full width
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("33")).
			Bold(true)
		str = style.Render(prefix + itemText + padding)
	} else {
		// Unselected section: gray text with dash, full width with background
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("235"))
		str = style.Render(prefix + itemText + padding)
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
		label := fmt.Sprintf("%s", item.Label)
		items = append(items, style.Render(" "+label+" "))
	}

	// Join with decorative separator and center in available width
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
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
			Background(lipgloss.Color("33"))  // Blue background
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")). // Gray text
		Background(lipgloss.Color("235"))  // Dark gray background
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
						ID:       "system-editors",
						Label:    "System Editors",
						ItemType: SectionHeader,
						SubItems: []SubmenuItem{
							// TODO: Phase 2 - Add editor submenu items
						},
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
// Sub-menu should ONLY contain section headers, not individual fields
func buildListItems(category MenuCategory) []list.Item {
	var items []list.Item

	for _, section := range category.SubItems {
		// Add ONLY section headers to the submenu
		// Individual fields will be shown in a modal after selecting a section
		if section.ItemType == SectionHeader {
			items = append(items, submenuListItem{submenuItem: section})
		}
	}

	return items
}

// InitialModelV2 creates the initial model state for the v2 TUI
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
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("33"))

	menuBar := buildMenuStructure(cfg)

	// Initialize the submenu list with items from the first menu category
	listItems := buildListItems(menuBar.Items[0])

	// Calculate max width for light bar
	maxWidth := 0
	for _, item := range listItems {
		if sli, ok := item.(submenuListItem); ok {
			itemWidth := len(sli.submenuItem.Label) + 2 // +2 for prefix ("- " or "> ")
			if itemWidth > maxWidth {
				maxWidth = itemWidth
			}
		}
	}

	submenuList := list.New(listItems, submenuDelegate{maxWidth: maxWidth}, 50, 10)
	submenuList.Title = ""
	submenuList.SetShowStatusBar(false)
	submenuList.SetFilteringEnabled(false)
	submenuList.SetShowHelp(false)
	submenuList.SetShowPagination(false)

	// Apply Mystic BBS styling to the list
	submenuList.Styles.Title = lipgloss.NewStyle()
	submenuList.Styles.PaginationStyle = lipgloss.NewStyle()
	submenuList.Styles.HelpStyle = lipgloss.NewStyle()

	return Model{
		config:       cfg,
		navMode:      MenuNavigation,
		activeMenu:   0,
		menuBar:      menuBar,
		submenuList:  submenuList,
		textInput:    ti,
		screenWidth:  80,
		screenHeight: 24,
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
		case MenuNavigation:
			return m.handleMenuNavigation(msg)
		case SubmenuNavigation:
			return m.handleSubmenuNavigation(msg)
		case ModalFieldSelect:
			return m.handleModalFieldSelect(msg)
		case SavePrompt:
			return m.handleSavePrompt(msg)
		}
	}

	// Update submenu list if in submenu navigation mode
	if m.navMode == SubmenuNavigation {
		m.submenuList, cmd = m.submenuList.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleMenuNavigation processes input while in menu navigation mode
func (m Model) handleMenuNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		m.activeMenu = (m.activeMenu - 1 + len(m.menuBar.Items)) % len(m.menuBar.Items)
		m.message = ""
		m.messageType = InfoMessage
	case "right", "l":
		m.activeMenu = (m.activeMenu + 1) % len(m.menuBar.Items)
		m.message = ""
		m.messageType = InfoMessage
	case "down", "j", "enter":
		// Enter submenu navigation mode
		m.navMode = SubmenuNavigation
		// Update list with current menu's items
		listItems := buildListItems(m.menuBar.Items[m.activeMenu])

		// Recalculate max width for the new menu
		maxWidth := 0
		for _, item := range listItems {
			if sli, ok := item.(submenuListItem); ok {
				itemWidth := len(sli.submenuItem.Label) + 2
				if itemWidth > maxWidth {
					maxWidth = itemWidth
				}
			}
		}
		m.submenuList.SetDelegate(submenuDelegate{maxWidth: maxWidth})
		m.submenuList.SetItems(listItems)
		m.submenuList.Select(0)
		m.message = ""
	case "q":
		m.savePrompt = true
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

// handleSubmenuNavigation processes input while in submenu navigation mode
func (m Model) handleSubmenuNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "left", "h":
		// Switch to previous menu category
		m.activeMenu = (m.activeMenu - 1 + len(m.menuBar.Items)) % len(m.menuBar.Items)
		listItems := buildListItems(m.menuBar.Items[m.activeMenu])

		// Recalculate max width
		maxWidth := 0
		for _, item := range listItems {
			if sli, ok := item.(submenuListItem); ok {
				itemWidth := len(sli.submenuItem.Label) + 2
				if itemWidth > maxWidth {
					maxWidth = itemWidth
				}
			}
		}
		m.submenuList.SetDelegate(submenuDelegate{maxWidth: maxWidth})
		m.submenuList.SetItems(listItems)
		m.submenuList.Select(0)
		m.message = ""
	case "right", "l":
		// Switch to next menu category
		m.activeMenu = (m.activeMenu + 1) % len(m.menuBar.Items)
		listItems := buildListItems(m.menuBar.Items[m.activeMenu])

		// Recalculate max width
		maxWidth := 0
		for _, item := range listItems {
			if sli, ok := item.(submenuListItem); ok {
				itemWidth := len(sli.submenuItem.Label) + 2
				if itemWidth > maxWidth {
					maxWidth = itemWidth
				}
			}
		}
		m.submenuList.SetDelegate(submenuDelegate{maxWidth: maxWidth})
		m.submenuList.SetItems(listItems)
		m.submenuList.Select(0)
		m.message = ""
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
		// Select section - show all fields in modal for navigation
		selectedItem := m.submenuList.SelectedItem()
		if selectedItem != nil {
			item := selectedItem.(submenuListItem)
			if item.submenuItem.ItemType == SectionHeader {
				// Section selected - check if it has editable fields
				if len(item.submenuItem.SubItems) > 0 {
					// Enter modal field selection mode
					m.modalFields = item.submenuItem.SubItems
					m.modalFieldIndex = 0
					m.modalSectionName = item.submenuItem.Label
					m.navMode = ModalFieldSelect
					m.message = ""
				} else {
					m.message = "This section has no editable fields"
				}
			}
		}
	case "esc":
		// Return to menu navigation
		m.navMode = MenuNavigation
		m.message = ""
	}
	return m, nil
}

// handleModalFieldSelect processes input while selecting fields in modal
func (m Model) handleModalFieldSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		// Return to submenu navigation
		m.navMode = SubmenuNavigation
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
	case "y", "Y":
		// Save and quit
		if err := config.SaveConfig(m.config, ""); err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
			m.savePrompt = false
			return m, nil
		}
		m.message = "Configuration saved successfully!"
		m.quitting = true
		return m, tea.Quit
	case "n", "N":
		// Quit without saving
		m.quitting = true
		return m, tea.Quit
	case "esc":
		// Cancel quit
		m.savePrompt = false
		m.message = ""
	}
	return m, nil
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
					m.navMode = ModalFieldSelect
				}
			} else {
				// No change, just return
				m.navMode = ModalFieldSelect
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
					m.navMode = ModalFieldSelect
				}
			} else {
				// No change, just return
				m.navMode = ModalFieldSelect
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
			m.navMode = ModalFieldSelect
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
		m.navMode = ModalFieldSelect
		m.textInput.Blur()

	case "esc":
		// Cancel editing - restore original value
		m.editingItem.Field.SetValue(m.originalValue)
		m.navMode = ModalFieldSelect
		m.editingError = ""
		m.message = ""
		m.textInput.Blur()
	}

	return m, nil
}

// ============================================================================
// View Rendering
// ============================================================================

// View renders the complete UI
func (m Model) View() string {
	var doc strings.Builder

	// Menu bar
	menuBarComponent := MenuBarComponent{
		items:       m.menuBar.Items,
		activeIndex: m.activeMenu,
		width:       m.screenWidth,
	}
	doc.WriteString(menuBarComponent.Render() + "\n\n")

	// Save prompt overlay - Enhanced
	if m.savePrompt {
		var promptText string
		if m.modifiedCount > 0 {
			promptText = fmt.Sprintf("[!]  Save %d change(s) to database?\n\n", m.modifiedCount)
			promptText += "    [Y] Yes, save changes\n"
			promptText += "    [N] No, discard changes\n"
			promptText += "    [Esc] Cancel and continue editing"
		} else {
			promptText = "Exit configuration editor?\n\n"
			promptText += "    [Y] Yes\n"
			promptText += "    [N] No\n"
			promptText += "    [Esc] Cancel"
		}

		promptBox := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("214")). // Orange/yellow
			Background(lipgloss.Color("235")).
			Padding(1, 3).
			Align(lipgloss.Center).
			Width(60).
			Render(promptText)

		centeredPrompt := lipgloss.Place(
			m.screenWidth,
			10,
			lipgloss.Center,
			lipgloss.Center,
			promptBox,
		)
		doc.WriteString(centeredPrompt + "\n")
	} else if m.quitting {
		// Quitting message
		if m.message != "" {
			messageBox := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("46")).
				Padding(1, 2).
				Render(m.message)
			doc.WriteString(messageBox + "\n")
		}
	} else {
		// Always render content for each mode (don't skip menu bar)
		switch m.navMode {
		case SubmenuNavigation:
			// Render submenu overlay
			submenuOverlay := m.renderSubmenuOverlay()
			doc.WriteString(submenuOverlay + "\n")
		case ModalFieldSelect, EditingValue:
			// Render modal (handles both selection and inline editing)
			modalForm := m.renderModalForm()
			doc.WriteString(modalForm + "\n")
		default:
			// Content area placeholder for menu navigation
			content := m.renderContentArea()
			doc.WriteString(content + "\n")
		}
	}

	// Enhanced message display with icons and colors based on type
	if m.message != "" && !m.savePrompt && !m.quitting {
		// Check if message should fade (3 seconds)
		if time.Since(m.messageTime) < 3*time.Second {
			var icon string
			var color string

			switch m.messageType {
			case SuccessMessage:
				icon = "[OK]"
				color = "46" // Green
			case ErrorMessage:
				icon = "[√]"
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
			doc.WriteString("\n" + msg + "\n")
		}
	}

	// Get the content so far
	contentStr := doc.String()

	// Enhanced footer with sections - anchored to bottom row
	if !m.quitting {
		footer := m.renderFooter()

		// Calculate how many lines the content uses
		contentLines := strings.Count(contentStr, "\n")

		// Add padding to push footer to row 25 (fixed position)
		// After contentLines newlines, we're at row (contentLines + 1)
		// We want the footer to start at row 25, so we need (25 - (contentLines + 1)) more newlines
		targetRow := 24 // Row 25 in 0-indexed terms
		paddingNeeded := targetRow - contentLines

		if paddingNeeded > 0 {
			doc.WriteString(strings.Repeat("\n", paddingNeeded))
		}

		doc.WriteString(footer)
	}

	return doc.String()
}

// renderContentArea renders the main content area
func (m Model) renderContentArea() string {
	if m.activeMenu >= len(m.menuBar.Items) {
		return ""
	}

	var content strings.Builder

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render("Press v or Enter to view items in this menu"))

	return content.String()
}

// renderSubmenuOverlay renders the submenu dropdown overlay
func (m Model) renderSubmenuOverlay() string {
	// Check if submenu is empty
	if len(m.submenuList.Items()) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true).
			Render("No items available in this category")

		emptyBox := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("33")).
			Background(lipgloss.Color("235")).
			Padding(1, 2).
			Render(emptyMsg)

		return emptyBox
	}

	// Render the list - compact dropdown style with just the items
	listView := m.submenuList.View()

	// Calculate submenu width (get width from the rendered content)
	listLines := strings.Split(listView, "\n")
	maxListWidth := 0
	for _, line := range listLines {
		if len(line) > maxListWidth {
			maxListWidth = len(line)
		}
	}
	submenuWidth := maxListWidth + 4 // +4 for border and padding

	// Calculate position to align under the active menu item
	separator := " | " // 3 chars

	// Calculate total width of all menu items plus separators
	totalMenuWidth := 0
	for i, item := range m.menuBar.Items {
		itemWidth := len(item.Label) + 2 // +2 for padding spaces
		totalMenuWidth += itemWidth
		if i < len(m.menuBar.Items)-1 {
			totalMenuWidth += len(separator)
		}
	}

	// Calculate starting position (centered menu bar)
	menuStartPos := (m.screenWidth - totalMenuWidth) / 2

	// Calculate position of active menu item
	activeMenuPos := menuStartPos
	for i := 0; i < m.activeMenu; i++ {
		activeMenuPos += len(m.menuBar.Items[i].Label) + 2 // +2 for padding
		activeMenuPos += len(separator)
	}

	// Check if submenu would extend beyond screen width
	if activeMenuPos+submenuWidth > m.screenWidth {
		// Adjust position to keep within bounds
		activeMenuPos = m.screenWidth - submenuWidth
		if activeMenuPos < 0 {
			activeMenuPos = 0
		}
	}

	// Create a compact dropdown-style box
	submenuBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("33")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Render(listView)

	// Add left padding to align under menu item
	lines := strings.Split(submenuBox, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat(" ", activeMenuPos) + line
		}
	}
	alignedSubmenu := strings.Join(lines, "\n")

	// Calculate how many lines the submenu takes
	submenuLines := strings.Count(alignedSubmenu, "\n")

	// Breadcrumb should be at row 23 (0-indexed, so position 22)
	targetRow := 23
	// Current position after menu bar (line 2) and submenu
	currentRow := 2 + submenuLines

	// Calculate padding needed to reach row 23
	paddingNeeded := targetRow - currentRow
	if paddingNeeded < 0 {
		paddingNeeded = 0
	}

	// Enhanced breadcrumb with full path
	breadcrumb := m.renderBreadcrumb()

	var result strings.Builder
	result.WriteString(alignedSubmenu)
	if paddingNeeded > 0 {
		result.WriteString(strings.Repeat("\n", paddingNeeded))
	}
	result.WriteString(breadcrumb)

	return result.String()
}

// renderModalForm renders the modal with all fields for navigation
func (m Model) renderModalForm() string {
	if len(m.modalFields) == 0 {
		return ""
	}

	var modalContent strings.Builder

	// Header with section name on blue background
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("33")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Width(56).
		Align(lipgloss.Center)

	modalContent.WriteString(headerStyle.Render(m.modalSectionName) + "\n\n")

	// Check if we're actively editing
	isEditing := m.navMode == EditingValue

	// Display all fields
	for i, field := range m.modalFields {
		if field.ItemType == EditableField && field.EditableItem != nil {
			currentValue := field.EditableItem.Field.GetValue()
			currentValueStr := formatValue(currentValue, field.EditableItem.ValueType)

			// Truncate long values to prevent wrapping (max 50 chars for value)
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
						fieldDisplay = fmt.Sprintf("%-25s [Y] Yes  [ ] No", field.EditableItem.Label+":")
					} else {
						fieldDisplay = fmt.Sprintf("%-25s [ ] Yes  [N] No", field.EditableItem.Label+":")
					}
					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color("33")).
						Foreground(lipgloss.Color("15")).
						Bold(true).
						Width(56)
					modalContent.WriteString(fullRowStyle.Render(" "+fieldDisplay) + "\n")
				} else {
					// Text input field - inline editing with cursor at value position
					label := fmt.Sprintf("%-25s", field.EditableItem.Label+":")

					// Create inline display with text input at value position
					fullRowStyle := lipgloss.NewStyle().
						Background(lipgloss.Color("33")).
						Foreground(lipgloss.Color("15")).
						Bold(true).
						Width(56)

					// Combine label and input inline
					inlineDisplay := label + " " + m.textInput.View()
					modalContent.WriteString(fullRowStyle.Render(" "+inlineDisplay) + "\n")
				}
			} else if isSelected && !isEditing {
				// SELECTION MODE: Only label portion highlighted in blue, rest has dark background
				labelText := fmt.Sprintf("%-25s", field.EditableItem.Label+":")
				valueText := " " + currentValueStr

				// Calculate available space for value (56 total - 1 space - 25 label - 1 space)
				availableValueSpace := 56 - 1 - 25 - 1
				if len(valueText) > availableValueSpace {
					valueText = valueText[:availableValueSpace-3] + "..."
				}

				labelStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("33")).
					Foreground(lipgloss.Color("15")).
					Bold(true).
					Width(26) // 25 + 1 for leading space

				valueStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("235")).
					Foreground(lipgloss.Color("250")).
					Width(30)

				label := labelStyle.Render(" " + labelText)
				value := valueStyle.Render(valueText)

				modalContent.WriteString(label + value + "\n")
			} else {
				// UNSELECTED: Normal display with dark background
				fieldDisplay := fmt.Sprintf("%-25s %s", field.EditableItem.Label+":", currentValueStr)

				// Truncate if too long
				if len(fieldDisplay) > 55 {
					fieldDisplay = fieldDisplay[:52] + "..."
				}

				fieldStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("235")).
					Foreground(lipgloss.Color("250")).
					Width(56)
				modalContent.WriteString(fieldStyle.Render(" "+fieldDisplay) + "\n")
			}
		}
	}

	// Show error if present
	if isEditing && m.editingError != "" {
		modalContent.WriteString("\n")
		errorMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("Error: " + m.editingError)
		modalContent.WriteString(errorMsg + "\n")
	}

	// Create compact modal box
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("33")).
		Background(lipgloss.Color("235")).
		Padding(1, 2).
		Render(modalContent.String())

	// Center the modal horizontally
	centeredModal := lipgloss.Place(
		m.screenWidth,
		m.screenHeight-10, // Leave room for menu, breadcrumb, and footer
		lipgloss.Center,
		lipgloss.Top,
		modalBox,
		lipgloss.WithWhitespaceChars(" "),
	)

	// Add breadcrumb at bottom
	breadcrumb := m.renderBreadcrumb()

	var result strings.Builder
	result.WriteString(centeredModal)
	result.WriteString("\n" + breadcrumb)

	return result.String()
}

// renderModalEditor renders the modal editing dialog
func (m Model) renderModalEditor() string {
	if m.editingItem == nil {
		return ""
	}

	var modalContent strings.Builder

	// Compact header with section name - Mystic style
	// Get the parent section name
	var sectionName string
	category := m.menuBar.Items[m.activeMenu]
	for _, section := range category.SubItems {
		if section.ItemType == SectionHeader {
			for _, subItem := range section.SubItems {
				if subItem.EditableItem != nil && subItem.EditableItem.ID == m.editingItem.ID {
					sectionName = section.Label
					break
				}
			}
		}
		if sectionName != "" {
			break
		}
	}

	// Header with section name on blue background
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("33")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Width(56).
		Align(lipgloss.Center)

	if sectionName != "" {
		modalContent.WriteString(headerStyle.Render(sectionName) + "\n\n")
	}

	// Show field in compact format: Label + Value
	currentValue := m.editingItem.Field.GetValue()

	// Selected field style (blue highlight like Mystic)
	fieldStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("33")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Width(54)

	// Format: "Label: Value" or editing interface
	var fieldDisplay string
	switch m.editingItem.ValueType {
	case BoolValue:
		// Show label and toggle options
		currentBool := currentValue.(bool)
		if currentBool {
			fieldDisplay = fmt.Sprintf("%s: [Y] Yes  [ ] No", m.editingItem.Label)
		} else {
			fieldDisplay = fmt.Sprintf("%s: [ ] Yes  [N] No", m.editingItem.Label)
		}
		modalContent.WriteString(fieldStyle.Render(fieldDisplay) + "\n")
	default:
		// Text input field
		fieldLabel := fmt.Sprintf("%s:", m.editingItem.Label)
		modalContent.WriteString(fieldStyle.Render(fieldLabel) + "\n")
		modalContent.WriteString(m.textInput.View() + "\n")
	}

	// Show error if present (compact)
	if m.editingError != "" {
		modalContent.WriteString("\n")
		errorMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("Error: " + m.editingError)
		modalContent.WriteString(errorMsg)
	}

	// Create compact modal box
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("33")).
		Background(lipgloss.Color("235")).
		Padding(1, 2).
		Render(modalContent.String())

	// Position modal below menu bar, no dimmed background
	var result strings.Builder
	result.WriteString("\n")
	result.WriteString(modalBox)

	// Add breadcrumb
	breadcrumb := m.renderBreadcrumb()
	result.WriteString("\n\n" + breadcrumb)

	return result.String()
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

	path.WriteString(categoryStyle.Render(category.Label))

	if m.navMode == SubmenuNavigation {
		// Show current submenu item
		selectedItem := m.submenuList.SelectedItem()
		if selectedItem != nil {
			item := selectedItem.(submenuListItem)
			path.WriteString(arrowStyle.Render(" -> "))

			// Find section name if item is in a section
			var sectionName string
			for _, section := range category.SubItems {
				if section.ItemType == SectionHeader {
					for _, subItem := range section.SubItems {
						if subItem.ID == item.submenuItem.ID {
							sectionName = section.Label
							break
						}
					}
				}
				if sectionName != "" {
					break
				}
			}

			if sectionName != "" {
				path.WriteString(detailStyle.Render(sectionName))
				path.WriteString(arrowStyle.Render(" -> "))
			}

			path.WriteString(detailStyle.Render(item.submenuItem.Label))
		}
	} else if m.navMode == EditingValue && m.editingItem != nil {
		// Show editing state
		path.WriteString(arrowStyle.Render(" -> "))
		path.WriteString(detailStyle.Render(m.editingItem.Label))
		path.WriteString(arrowStyle.Render(" -> "))
		path.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Render("EDITING"))
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Render(" " + path.String())
}

// renderFooter generates enhanced footer with sections
func (m Model) renderFooter() string {
	var sections []string

	// Breadcrumb section (already rendered elsewhere for some modes)
	// Help section based on current mode
	var helpText string
	switch m.navMode {
	case MenuNavigation:
		helpText = "◄► Menu ↑↓ Move [ESC] Back [Q] Quit"
	case SubmenuNavigation:
		helpText = "◄► Menu ↑↓ Move [ESC] Back [Q] Quit"
	case ModalFieldSelect:
		helpText = "◄► Menu ↑↓ Move [ESC] Back [Q] Quit"
	case EditingValue:
		if m.editingItem != nil && m.editingItem.ValueType == BoolValue {
			helpText = "[Key]  Y/N:Select | Space/Tab:Toggle | Enter:Save | Esc:Cancel"
		} else {
			helpText = "[Key]  Type value | Enter:Save | Esc:Cancel"
		}
	default:
		helpText = "[Key]  Q:Quit"
	}

	helpSection := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Render(helpText)
	sections = append(sections, helpSection)

	// Status section showing modified count
	if m.modifiedCount > 0 {
		statusText := fmt.Sprintf("[#] %d change(s)", m.modifiedCount)
		statusSection := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Bold(true).
			Render(statusText)
		sections = append(sections, statusSection)
	}

	// Join sections
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Render("|")

	footerContent := strings.Join(sections, separator)

	return lipgloss.NewStyle().
		Width(m.screenWidth).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Render(footerContent)
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

// ============================================================================
// Entry Point
// ============================================================================

// RunConfigEditorTUI starts the configuration editor TUI
func RunConfigEditorTUI(cfg *config.Config) error {
	p := tea.NewProgram(InitialModelV2(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
