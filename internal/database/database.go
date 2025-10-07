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

// UserMFARecord represents a multi-factor authentication enrollment.
type UserMFARecord struct {
	ID          int64
	UserID      int64
	Method      string
	Secret      sql.NullString
	Config      sql.NullString
	IsEnabled   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastUsedAt  sql.NullTime
	BackupCodes sql.NullString
	DisplayName sql.NullString
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
	UpsertUserMFA(record *UserMFARecord) error
	GetMFAForUser(userID int64) ([]UserMFARecord, error)
	DeleteMFAForUser(userID int64, method string) error

	// Database management
	InitializeSchema() error
	Close() error
}

// ConnectionConfig holds database connection settings
type ConnectionConfig struct {
	Path    string
	Timeout int
}
