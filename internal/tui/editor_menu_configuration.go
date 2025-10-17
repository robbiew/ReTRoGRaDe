package tui

import (
	"fmt"
	"strings"

	"github.com/robbiew/retrograde/internal/config"
)

func configurationMenu(cfg *config.Config) MenuCategory {
	return MenuCategory{
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
					{
						ID:       "security-path",
						Label:    "Security",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.paths.security",
							Label:     "Security",
							ValueType: PathValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Configuration.Paths.Security },
								SetValue: func(v interface{}) error {
									cfg.Configuration.Paths.Security = v.(string)
									return nil
								},
							},
							HelpText: "Path to security lists directory",
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
						ID:       "timeout-minutes",
						Label:    "Timeout Mins",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.general.timeout_minutes",
							Label:     "Timeout Mins",
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
						Label:    "Allow New",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.new_users.allow_new",
							Label:     "Allow New",
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
						ID:       "ask-first-name",
						Label:    "Ask First Name",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.new_users.ask_first_name",
							Label:     "Ask First Name",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Configuration.NewUsers.AskFirstName },
								SetValue: func(v interface{}) error {
									cfg.Configuration.NewUsers.AskFirstName = v.(bool)
									return nil
								},
							},
							HelpText: "Ask for first name during registration",
						},
					},
					{
						ID:       "ask-last-name",
						Label:    "Ask Last Name",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.new_users.ask_last_name",
							Label:     "Ask Last Name",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Configuration.NewUsers.AskLastName },
								SetValue: func(v interface{}) error {
									cfg.Configuration.NewUsers.AskLastName = v.(bool)
									return nil
								},
							},
							HelpText: "Ask for last name during registration",
						},
					},
					{
						ID:       "ask-email",
						Label:    "Ask Email",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.new_users.ask_email",
							Label:     "Ask Email",
							ValueType: BoolValue,
							Field: ConfigField{GetValue: func() interface{} { return cfg.Configuration.NewUsers.AskEmail },
								SetValue: func(v interface{}) error {
									cfg.Configuration.NewUsers.AskEmail = v.(bool)
									return nil
								},
							},
							HelpText: "Ask for email during registration",
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
						Label:    "Failed Logins",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.auth.max_failed_attempts",
							Label:     "Failed Logins",
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
						Label:    "Lock Minutes",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.auth.account_lock_minutes",
							Label:     "Lock Minutes",
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
						Label:    "Pwd Algorithm",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "config.auth.password_algorithm",
							Label:     "Pwd Algorithm",
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
	}
}
