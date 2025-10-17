package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/robbiew/retrograde/internal/database"
)

var db database.Database

// LoadConfig loads configuration from SQLite database
// If database doesn't exist, creates it with default values
func LoadConfig(filePath string) (*Config, error) {
	// Determine database path
	dbPath := getDBPath()

	// Check if database exists
	if fileExists(dbPath) {
		fmt.Println("Loading configuration from SQLite database...")
		return loadFromDatabase(dbPath)
	}

	// Database doesn't exist - create with defaults
	fmt.Println("Database not found, creating new database with defaults...")

	cfg := GetDefaultConfig()
	if err := initializeDatabaseWithConfig(dbPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	fmt.Println("Database created with default configuration")
	return cfg, nil
}

// loadFromDatabase loads configuration from existing database
func loadFromDatabase(dbPath string) (*Config, error) {
	dbConfig := database.ConnectionConfig{
		Path:    dbPath,
		Timeout: 5,
	}

	var err error
	db, err = database.OpenSQLite(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ensure schema is initialized
	if err := db.InitializeSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Load from database
	cfg, err := LoadConfigFromDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to load from database: %w", err)
	}

	return cfg, nil
}

// initializeDatabaseWithConfig creates a new database and populates it with the given config
func initializeDatabaseWithConfig(dbPath string, cfg *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	dbConfig := database.ConnectionConfig{
		Path:    dbPath,
		Timeout: 5,
	}

	var err error
	db, err = database.OpenSQLite(dbConfig)
	if err != nil {
		return err
	}

	if err := db.InitializeSchema(); err != nil {
		return err
	}

	// Create required directories
	if err := EnsureRequiredPaths(cfg); err != nil {
		return fmt.Errorf("failed to create required directories: %w", err)
	}

	// Save configuration to database
	if err := SaveConfigToDB(db, cfg, "system"); err != nil {
		return err
	}

	return nil
}

// SaveConfig saves configuration to SQLite database
func SaveConfig(config *Config, filePath string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create required directories before saving
	if err := EnsureRequiredPaths(config); err != nil {
		return fmt.Errorf("failed to create required directories: %w", err)
	}

	if err := SaveConfigToDB(db, config, "user"); err != nil {
		return fmt.Errorf("failed to save to database: %w", err)
	}

	fmt.Println(" Configuration saved to database")
	return nil
}

// Database mapping functions for converting between Config struct and database ConfigValue

// LoadConfigFromDB loads configuration from database
func LoadConfigFromDB(db database.Database) (*Config, error) {
	cfg := GetDefaultConfig()

	// Load all values from database
	values, err := db.GetAllConfigValues()
	if err != nil {
		return nil, err
	}

	// Map each value to the appropriate field in config struct
	for _, v := range values {
		mapValueToConfig(cfg, v)
	}

	return cfg, nil
}

// SaveConfigToDB saves configuration to database
func SaveConfigToDB(db database.Database, cfg *Config, modifiedBy string) error {
	// Convert config struct to ConfigValue slice
	values := configToValues(cfg)

	// Insert/update each value
	for _, v := range values {
		if err := db.SetConfig(v.Section, v.Subsection, v.Key,
			v.Value, v.ValueType, modifiedBy); err != nil {
			return err
		}
	}

	return nil
}

// mapValueToConfig maps a single database.ConfigValue to the appropriate field in Config struct
func mapValueToConfig(cfg *Config, v database.ConfigValue) {
	section := v.Section
	subsection := v.Subsection
	key := v.Key
	value := v.Value

	// Configuration.Paths
	if section == "Configuration.Paths" {
		switch key {
		case "Database":
			cfg.Configuration.Paths.Database = value
		case "File_Base":
			cfg.Configuration.Paths.FileBase = value
		case "Logs":
			cfg.Configuration.Paths.Logs = value
		case "Message_Base":
			cfg.Configuration.Paths.MessageBase = value
		case "System":
			cfg.Configuration.Paths.System = value
		case "Themes":
			cfg.Configuration.Paths.Themes = value
		case "Security":
			cfg.Configuration.Paths.Security = value
		}
		return
	}

	// Configuration.General
	if section == "Configuration.General" {
		switch key {
		case "BBS_Location":
			cfg.Configuration.General.BBSLocation = value
		case "BBS_Name":
			cfg.Configuration.General.BBSName = value
		case "Start_Menu":
			cfg.Configuration.General.StartMenu = value
		case "SysOp_Name":
			cfg.Configuration.General.SysOpName = value
		case "Timeout_Minutes":
			cfg.Configuration.General.TimeoutMinutes = parseIntValue(value)
		}
		return
	}

	// Configuration.New_Users
	if section == "Configuration.New_Users" {
		if cfg.Configuration.NewUsers.RegistrationFields == nil {
			cfg.Configuration.NewUsers.RegistrationFields = make(map[string]RegistrationFieldConfig)
		}
		if cfg.Configuration.NewUsers.FormLayout == nil {
			cfg.Configuration.NewUsers.FormLayout = make(map[string]FormLayoutConfig)
		}

		switch {
		case key == "Allow_New":
			cfg.Configuration.NewUsers.AllowNew = parseBoolValue(value)
		case key == "Ask_Location":
			cfg.Configuration.NewUsers.AskLocation = parseBoolValue(value)
		case key == "Ask_First_Name":
			cfg.Configuration.NewUsers.AskFirstName = parseBoolValue(value)
		case key == "Ask_Last_Name":
			cfg.Configuration.NewUsers.AskLastName = parseBoolValue(value)
		case key == "RegistrationFormEnabledFields":
			cfg.Configuration.NewUsers.RegistrationFormEnabledFields = parseListValue(value)
		case strings.HasPrefix(key, "RegistrationField."):
			parts := strings.Split(key, ".")
			if len(parts) == 3 {
				fieldID := parts[1]
				attr := parts[2]
				field := cfg.Configuration.NewUsers.RegistrationFields[fieldID]
				switch attr {
				case "Enabled":
					field.Enabled = parseBoolValue(value)
				case "Required":
					field.Required = parseBoolValue(value)
				}
				cfg.Configuration.NewUsers.RegistrationFields[fieldID] = field
			}
		case strings.HasPrefix(key, "FormLayout."):
			parts := strings.Split(key, ".")
			if len(parts) == 3 {
				fieldID := parts[1]
				attr := parts[2]
				layout := cfg.Configuration.NewUsers.FormLayout[fieldID]
				switch attr {
				case "Row":
					layout.Row = parseIntValue(value)
				case "Col":
					layout.Col = parseIntValue(value)
				}
				cfg.Configuration.NewUsers.FormLayout[fieldID] = layout
			}
		}
		return
	}

	// Configuration.Auth
	if section == "Configuration.Auth" {
		switch key {
		case "MaxFailedAttempts":
			cfg.Configuration.Auth.MaxFailedAttempts = parseIntValue(value)
		case "AccountLockMinutes":
			cfg.Configuration.Auth.AccountLockMinutes = parseIntValue(value)
		case "PasswordAlgorithm":
			cfg.Configuration.Auth.PasswordAlgorithm = value
		}
		return
	}

	// Servers.General_Settings
	if section == "Servers.General_Settings" {
		switch key {
		case "Max_Connections_Per_IP":
			cfg.Servers.GeneralSettings.MaxConnectionsPerIP = parseIntValue(value)
		case "Max_Nodes":
			cfg.Servers.GeneralSettings.MaxNodes = parseIntValue(value)
		}
		return
	}

	// Servers.Telnet
	if section == "Servers.Telnet" {
		switch key {
		case "Active":
			cfg.Servers.Telnet.Active = parseBoolValue(value)
		case "Port":
			cfg.Servers.Telnet.Port = parseIntValue(value)
		}
		return
	}

	// Servers.Security with subsections
	if section == "Servers.Security" {
		switch subsection {
		case "Rate_Limits":
			switch key {
			case "RateLimitEnabled":
				cfg.Servers.Security.RateLimits.Enabled = parseBoolValue(value)
			case "RateLimitWindowMinutes":
				cfg.Servers.Security.RateLimits.WindowMinutes = parseIntValue(value)
			}
		case "Local_Lists":
			switch key {
			case "BlacklistEnabled":
				cfg.Servers.Security.LocalLists.BlacklistEnabled = parseBoolValue(value)
			case "BlacklistFile":
				cfg.Servers.Security.LocalLists.BlacklistFile = value
			case "WhitelistEnabled":
				cfg.Servers.Security.LocalLists.WhitelistEnabled = parseBoolValue(value)
			case "WhitelistFile":
				cfg.Servers.Security.LocalLists.WhitelistFile = value
			}
		case "External_Lists":
			switch key {
			case "External_Block_Enabled":
				cfg.Servers.Security.ExternalLists.Enabled = parseBoolValue(value)
			case "ExternalBlocklistURLs":
				cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs = parseListValue(value)
			}
		case "Geo_Block":
			switch key {
			case "AllowedCountries":
				cfg.Servers.Security.GeoBlock.AllowedCountries = parseListValue(value)
			case "BlockedCountries":
				cfg.Servers.Security.GeoBlock.BlockedCountries = parseListValue(value)
			case "BlocklistUpdateHours":
				cfg.Servers.Security.GeoBlock.BlocklistUpdateHours = parseIntValue(value)
			case "GeoAPIKey":
				cfg.Servers.Security.GeoBlock.GeoAPIKey = value
			case "GeoAPIProvider":
				cfg.Servers.Security.GeoBlock.GeoAPIProvider = value
			case "GeoBlockEnabled":
				cfg.Servers.Security.GeoBlock.GeoBlockEnabled = parseBoolValue(value)
			case "ThreatIntelEnabled":
				cfg.Servers.Security.GeoBlock.ThreatIntelEnabled = parseBoolValue(value)
			}
		case "Logs":
			switch key {
			case "LogBlockedAttempts":
				cfg.Servers.Security.Logs.LogBlockedAttempts = parseBoolValue(value)
			case "LogSecurityEvents":
				cfg.Servers.Security.Logs.LogSecurityEvents = parseBoolValue(value)
			case "SecurityLogFile":
				cfg.Servers.Security.Logs.SecurityLogFile = value
			}
		}
		return
	}

	// Other.Discord
	if section == "Other.Discord" {
		switch key {
		case "Discord":
			cfg.Other.Discord.Enabled = parseBoolValue(value)
		case "DiscordInviteURL":
			cfg.Other.Discord.InviteURL = value
		case "DiscordTitle":
			cfg.Other.Discord.Title = value
		case "DiscordUsername":
			cfg.Other.Discord.Username = value
		case "DiscordWebhookURL":
			cfg.Other.Discord.WebhookURL = value
		}
		return
	}
}

// configToValues converts Config struct to slice of database.ConfigValue
func configToValues(cfg *Config) []database.ConfigValue {
	var values []database.ConfigValue

	// Configuration.Paths
	values = append(values,
		database.ConfigValue{Section: "Configuration.Paths", Key: "Database", Value: cfg.Configuration.Paths.Database, ValueType: "path"},
		database.ConfigValue{Section: "Configuration.Paths", Key: "File_Base", Value: cfg.Configuration.Paths.FileBase, ValueType: "path"},
		database.ConfigValue{Section: "Configuration.Paths", Key: "Logs", Value: cfg.Configuration.Paths.Logs, ValueType: "path"},
		database.ConfigValue{Section: "Configuration.Paths", Key: "Message_Base", Value: cfg.Configuration.Paths.MessageBase, ValueType: "path"},
		database.ConfigValue{Section: "Configuration.Paths", Key: "System", Value: cfg.Configuration.Paths.System, ValueType: "path"},
		database.ConfigValue{Section: "Configuration.Paths", Key: "Themes", Value: cfg.Configuration.Paths.Themes, ValueType: "path"},
	)

	// Configuration.General
	values = append(values,
		database.ConfigValue{Section: "Configuration.General", Key: "BBS_Location", Value: cfg.Configuration.General.BBSLocation, ValueType: "string"},
		database.ConfigValue{Section: "Configuration.General", Key: "BBS_Name", Value: cfg.Configuration.General.BBSName, ValueType: "string"},
		database.ConfigValue{Section: "Configuration.General", Key: "Start_Menu", Value: cfg.Configuration.General.StartMenu, ValueType: "string"},
		database.ConfigValue{Section: "Configuration.General", Key: "SysOp_Name", Value: cfg.Configuration.General.SysOpName, ValueType: "string"},
		database.ConfigValue{Section: "Configuration.General", Key: "Timeout_Minutes", Value: strconv.Itoa(cfg.Configuration.General.TimeoutMinutes), ValueType: "int"},
	)

	// Configuration.New_Users
	values = append(values,
		database.ConfigValue{Section: "Configuration.New_Users", Key: "Allow_New", Value: formatBoolValue(cfg.Configuration.NewUsers.AllowNew), ValueType: "bool"},
		database.ConfigValue{Section: "Configuration.New_Users", Key: "Ask_Location", Value: formatBoolValue(cfg.Configuration.NewUsers.AskLocation), ValueType: "bool"},
		database.ConfigValue{Section: "Configuration.New_Users", Key: "Ask_First_Name", Value: formatBoolValue(cfg.Configuration.NewUsers.AskFirstName), ValueType: "bool"},
		database.ConfigValue{Section: "Configuration.New_Users", Key: "Ask_Last_Name", Value: formatBoolValue(cfg.Configuration.NewUsers.AskLastName), ValueType: "bool"},
		database.ConfigValue{Section: "Configuration.New_Users", Key: "RegistrationFormEnabledFields", Value: formatListValue(cfg.Configuration.NewUsers.RegistrationFormEnabledFields), ValueType: "list"},
	)

	// Configuration.Auth
	values = append(values,
		database.ConfigValue{Section: "Configuration.Auth", Key: "MaxFailedAttempts", Value: strconv.Itoa(cfg.Configuration.Auth.MaxFailedAttempts), ValueType: "int"},
		database.ConfigValue{Section: "Configuration.Auth", Key: "AccountLockMinutes", Value: strconv.Itoa(cfg.Configuration.Auth.AccountLockMinutes), ValueType: "int"},
		database.ConfigValue{Section: "Configuration.Auth", Key: "PasswordAlgorithm", Value: cfg.Configuration.Auth.PasswordAlgorithm, ValueType: "string"},
	)

	if cfg.Configuration.NewUsers.RegistrationFields != nil {
		fieldIDs := make([]string, 0, len(cfg.Configuration.NewUsers.RegistrationFields))
		for fieldID := range cfg.Configuration.NewUsers.RegistrationFields {
			fieldIDs = append(fieldIDs, fieldID)
		}
		sort.Strings(fieldIDs)
		for _, fieldID := range fieldIDs {
			field := cfg.Configuration.NewUsers.RegistrationFields[fieldID]
			values = append(values,
				database.ConfigValue{
					Section:   "Configuration.New_Users",
					Key:       fmt.Sprintf("RegistrationField.%s.Enabled", fieldID),
					Value:     formatBoolValue(field.Enabled),
					ValueType: "bool",
				},
				database.ConfigValue{
					Section:   "Configuration.New_Users",
					Key:       fmt.Sprintf("RegistrationField.%s.Required", fieldID),
					Value:     formatBoolValue(field.Required),
					ValueType: "bool",
				},
			)
		}
	}

	if cfg.Configuration.NewUsers.FormLayout != nil {
		fieldIDs := make([]string, 0, len(cfg.Configuration.NewUsers.FormLayout))
		for fieldID := range cfg.Configuration.NewUsers.FormLayout {
			fieldIDs = append(fieldIDs, fieldID)
		}
		sort.Strings(fieldIDs)
		for _, fieldID := range fieldIDs {
			layout := cfg.Configuration.NewUsers.FormLayout[fieldID]
			values = append(values,
				database.ConfigValue{
					Section:   "Configuration.New_Users",
					Key:       fmt.Sprintf("FormLayout.%s.Row", fieldID),
					Value:     strconv.Itoa(layout.Row),
					ValueType: "int",
				},
				database.ConfigValue{
					Section:   "Configuration.New_Users",
					Key:       fmt.Sprintf("FormLayout.%s.Col", fieldID),
					Value:     strconv.Itoa(layout.Col),
					ValueType: "int",
				},
			)
		}
	}

	// Servers.General_Settings
	values = append(values,
		database.ConfigValue{Section: "Servers.General_Settings", Key: "Max_Connections_Per_IP", Value: strconv.Itoa(cfg.Servers.GeneralSettings.MaxConnectionsPerIP), ValueType: "int"},
		database.ConfigValue{Section: "Servers.General_Settings", Key: "Max_Nodes", Value: strconv.Itoa(cfg.Servers.GeneralSettings.MaxNodes), ValueType: "int"},
	)

	// Servers.Telnet
	values = append(values,
		database.ConfigValue{Section: "Servers.Telnet", Key: "Active", Value: formatBoolValue(cfg.Servers.Telnet.Active), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Telnet", Key: "Port", Value: strconv.Itoa(cfg.Servers.Telnet.Port), ValueType: "int"},
	)

	// Servers.Security.Rate_Limits
	values = append(values,
		database.ConfigValue{Section: "Servers.Security", Subsection: "Rate_Limits", Key: "RateLimitEnabled", Value: formatBoolValue(cfg.Servers.Security.RateLimits.Enabled), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Rate_Limits", Key: "RateLimitWindowMinutes", Value: strconv.Itoa(cfg.Servers.Security.RateLimits.WindowMinutes), ValueType: "int"},
	)

	// Servers.Security.Local_Lists
	values = append(values,
		database.ConfigValue{Section: "Servers.Security", Subsection: "Local_Lists", Key: "BlacklistEnabled", Value: formatBoolValue(cfg.Servers.Security.LocalLists.BlacklistEnabled), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Local_Lists", Key: "BlacklistFile", Value: cfg.Servers.Security.LocalLists.BlacklistFile, ValueType: "path"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Local_Lists", Key: "WhitelistEnabled", Value: formatBoolValue(cfg.Servers.Security.LocalLists.WhitelistEnabled), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Local_Lists", Key: "WhitelistFile", Value: cfg.Servers.Security.LocalLists.WhitelistFile, ValueType: "path"},
	)

	// Servers.Security.External_Lists
	values = append(values,
		database.ConfigValue{Section: "Servers.Security", Subsection: "External_Lists", Key: "External_Block_Enabled", Value: formatBoolValue(cfg.Servers.Security.ExternalLists.Enabled), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "External_Lists", Key: "ExternalBlocklistURLs", Value: formatListValue(cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs), ValueType: "list"},
	)

	// Servers.Security.Geo_Block
	values = append(values,
		database.ConfigValue{Section: "Servers.Security", Subsection: "Geo_Block", Key: "AllowedCountries", Value: formatListValue(cfg.Servers.Security.GeoBlock.AllowedCountries), ValueType: "list"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Geo_Block", Key: "BlockedCountries", Value: formatListValue(cfg.Servers.Security.GeoBlock.BlockedCountries), ValueType: "list"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Geo_Block", Key: "BlocklistUpdateHours", Value: strconv.Itoa(cfg.Servers.Security.GeoBlock.BlocklistUpdateHours), ValueType: "int"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Geo_Block", Key: "GeoAPIKey", Value: cfg.Servers.Security.GeoBlock.GeoAPIKey, ValueType: "string"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Geo_Block", Key: "GeoAPIProvider", Value: cfg.Servers.Security.GeoBlock.GeoAPIProvider, ValueType: "string"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Geo_Block", Key: "GeoBlockEnabled", Value: formatBoolValue(cfg.Servers.Security.GeoBlock.GeoBlockEnabled), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Geo_Block", Key: "ThreatIntelEnabled", Value: formatBoolValue(cfg.Servers.Security.GeoBlock.ThreatIntelEnabled), ValueType: "bool"},
	)

	// Servers.Security.Logs
	values = append(values,
		database.ConfigValue{Section: "Servers.Security", Subsection: "Logs", Key: "LogBlockedAttempts", Value: formatBoolValue(cfg.Servers.Security.Logs.LogBlockedAttempts), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Logs", Key: "LogSecurityEvents", Value: formatBoolValue(cfg.Servers.Security.Logs.LogSecurityEvents), ValueType: "bool"},
		database.ConfigValue{Section: "Servers.Security", Subsection: "Logs", Key: "SecurityLogFile", Value: cfg.Servers.Security.Logs.SecurityLogFile, ValueType: "path"},
	)

	// Other.Discord
	values = append(values,
		database.ConfigValue{Section: "Other.Discord", Key: "Discord", Value: formatBoolValue(cfg.Other.Discord.Enabled), ValueType: "bool"},
		database.ConfigValue{Section: "Other.Discord", Key: "DiscordInviteURL", Value: cfg.Other.Discord.InviteURL, ValueType: "string"},
		database.ConfigValue{Section: "Other.Discord", Key: "DiscordTitle", Value: cfg.Other.Discord.Title, ValueType: "string"},
		database.ConfigValue{Section: "Other.Discord", Key: "DiscordUsername", Value: cfg.Other.Discord.Username, ValueType: "string"},
		database.ConfigValue{Section: "Other.Discord", Key: "DiscordWebhookURL", Value: cfg.Other.Discord.WebhookURL, ValueType: "string"},
	)

	return values
}

// Helper functions for parsing database values
func parseBoolValue(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	return v == "true" || v == "yes" || v == "1"
}

func parseIntValue(value string) int {
	i, _ := strconv.Atoi(value)
	return i
}

func parseListValue(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func formatBoolValue(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func formatListValue(values []string) string {
	return strings.Join(values, ", ")
}

// getDBPath returns the database path (uses configured database path or default)
func getDBPath() string {
	// For initial database creation, use default location
	// After config is loaded, this path will be used from the config
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	defaultPath := filepath.Join(cwd, "data", "retrograde.db")

	// Ensure the data directory exists
	dir := filepath.Dir(defaultPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// If we can't create the default directory, fall back to current directory
		return filepath.Join(cwd, "retrograde.db")
	}

	return defaultPath
}

// GetDatabase returns the active database handle, or nil if uninitialized.
func GetDatabase() database.Database {
	return db
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		err := db.Close()
		db = nil // Clear the reference after closing
		return err
	}
	return nil
}

// CheckRequiredPathsExist checks if the required directories from configuration exist
// Returns true if all paths exist, false if any are missing
func CheckRequiredPathsExist(cfg *Config) bool {
	pathsToCheck := []string{
		cfg.Configuration.Paths.Logs,
		cfg.Configuration.Paths.MessageBase,
		cfg.Configuration.Paths.FileBase,
		// Skip System path as it's always the current directory
		cfg.Configuration.Paths.Themes,
		filepath.Dir(cfg.Servers.Security.LocalLists.BlacklistFile), // security directory
	}

	for _, path := range pathsToCheck {
		if path == "" {
			continue // Skip empty paths
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// EnsureRequiredPaths creates required directories if they don't exist
func EnsureRequiredPaths(cfg *Config) error {
	pathsToCreate := []string{
		cfg.Configuration.Paths.Logs,
		cfg.Configuration.Paths.MessageBase,
		cfg.Configuration.Paths.FileBase,
		// Skip System path as it's always the current directory
		cfg.Configuration.Paths.Themes,
		filepath.Dir(cfg.Servers.Security.LocalLists.BlacklistFile), // security directory
	}

	for _, path := range pathsToCreate {
		if path == "" {
			continue // Skip empty paths
		}
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// Note: Application management functions moved to auth package to avoid circular imports
