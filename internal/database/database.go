package database

import (
	"database/sql"
	"time"
)

// ConfigValue represents a configuration setting
type ConfigValue struct {
	ID           int
	Section      string
	Subsection   string // Can be empty
	Key          string
	Value        string
	ValueType    string // 'string', 'int', 'bool', 'list', 'path'
	DefaultValue string
	Description  string
	CreatedAt    time.Time
	ModifiedAt   time.Time
	ModifiedBy   string
}

// AuthAuditEntry represents a row in the auth_audit table.
type AuthAuditEntry struct {
	ID        int64
	UserID    sql.NullInt64
	Username  sql.NullString
	EventType string
	IPAddress sql.NullString
	Metadata  sql.NullString
	CreatedAt time.Time
	Context   sql.NullString
}

// Menu represents a menu in the BBS system
type Menu struct {
	ID                  int
	Name                string
	Titles              []string // JSON
	Prompt              string
	ACSRequired         string
	Password            string
	GenericColumns      int
	GenericBracketColor int
	GenericCommandColor int
	GenericDescColor    int
	ClearScreen         bool
}

// MenuCommand represents a command in a menu
type MenuCommand struct {
	ID               int
	MenuID           int
	CommandNumber    int
	Keys             string
	ShortDescription string
	ACSRequired      string
	CmdKeys          string
	Options          string
}

// Database interface defines all database operations
type Database interface {
	// Configuration operations
	GetConfig(section, subsection, key string) (string, error)
	GetConfigInt(section, subsection, key string) (int, error)
	GetConfigBool(section, subsection, key string) (bool, error)
	GetConfigList(section, subsection, key string) ([]string, error)
	SetConfig(section, subsection, key, value, valueType, modifiedBy string) error
	GetAllConfigValues() ([]ConfigValue, error)

	// User operations
	CreateUser(user *UserRecord) (int64, error)
	GetUserByUsername(username string) (*UserRecord, error)
	GetUserByID(userID int64) (*UserRecord, error)
	GetAllUsers() ([]UserRecord, error)
	UpdateUser(user *UserRecord) error
	WithTransaction(fn func(*sql.Tx) error) error
	UpsertUserDetail(userID int64, attrib, value string) error
	GetUserDetails(userID int64) (map[string]string, error)
	IncrementFailedAttempts(userID int64, now time.Time, maxAttempts int, lockMinutes int) (int, *time.Time, error)
	ResetFailedAttempts(userID int64, now time.Time) error
	UpdatePassword(userID int64, hash, algo, salt string, now time.Time) error
	InsertAuthAudit(entry *AuthAuditEntry) error

	// Security level operations
	CreateSecurityLevel(level *SecurityLevelRecord) (int64, error)
	GetSecurityLevelByID(id int64) (*SecurityLevelRecord, error)
	GetSecurityLevelByLevel(secLevel int) (*SecurityLevelRecord, error)
	GetAllSecurityLevels() ([]SecurityLevelRecord, error)
	UpdateSecurityLevel(level *SecurityLevelRecord) error
	DeleteSecurityLevel(id int64) error

	// Menu operations
	CreateMenu(menu *Menu) (int64, error)
	GetMenuByID(id int64) (*Menu, error)
	GetMenuByName(name string) (*Menu, error)
	GetAllMenus() ([]Menu, error)
	UpdateMenu(menu *Menu) error
	DeleteMenu(id int64) error
	CreateMenuCommand(cmd *MenuCommand) (int64, error)
	GetMenuCommands(menuID int) ([]MenuCommand, error)
	UpdateMenuCommand(cmd *MenuCommand) error
	DeleteMenuCommand(id int64) error

	// Database management
	InitializeSchema() error
	Close() error
}

// ConnectionConfig holds database connection settings
type ConnectionConfig struct {
	Path    string
	Timeout int
}
