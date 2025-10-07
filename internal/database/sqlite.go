package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// SQLiteDB implements the Database interface using SQLite
type SQLiteDB struct {
	db   *sql.DB
	path string
}

// OpenSQLite opens or creates a SQLite database
func OpenSQLite(config ConnectionConfig) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", config.Path+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database at path '%s': %w", config.Path, err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database at path '%s': %w", config.Path, err)
	}

	sqliteDB := &SQLiteDB{
		db:   db,
		path: config.Path,
	}

	return sqliteDB, nil
}

// InitializeSchema creates all necessary tables
func (s *SQLiteDB) InitializeSchema() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create schema_version table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			version INTEGER NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_version: %w", err)
	}

	// Initialize version if empty
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO schema_version (id, version, description)
		VALUES (1, 1, 'Initial schema for configuration')
	`)
	if err != nil {
		return err
	}

	// Create config_settings table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS config_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			section TEXT NOT NULL,
			subsection TEXT,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			value_type TEXT NOT NULL,
			default_value TEXT,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			modified_by TEXT DEFAULT 'system',
			UNIQUE(section, subsection, key)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create config_settings: %w", err)
	}

	// Create indexes
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_config_section 
					  ON config_settings(section)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_config_subsection 
					  ON config_settings(section, subsection)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_config_key 
					  ON config_settings(section, subsection, key)`)
	if err != nil {
		return err
	}

	// Create update trigger
	_, err = tx.Exec(`
		CREATE TRIGGER IF NOT EXISTS config_update_timestamp 
		AFTER UPDATE ON config_settings
		BEGIN
			UPDATE config_settings 
			SET modified_at = CURRENT_TIMESTAMP 
			WHERE id = NEW.id;
		END
	`)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Ensure user-related schema is initialized as well.
	return s.InitializeUserSchema()
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// GetConfig retrieves a configuration value as string
func (s *SQLiteDB) GetConfig(section, subsection, key string) (string, error) {
	var value string

	query := `SELECT value FROM config_settings 
			  WHERE section = ? AND key = ?`
	args := []interface{}{section, key}

	if subsection != "" {
		query += ` AND subsection = ?`
		args = append(args, subsection)
	} else {
		query += ` AND subsection IS NULL`
	}

	err := s.db.QueryRow(query, args...).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("config not found: %s.%s.%s", section, subsection, key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	return value, nil
}

// GetConfigInt retrieves a configuration value as integer
func (s *SQLiteDB) GetConfigInt(section, subsection, key string) (int, error) {
	value, err := s.GetConfig(section, subsection, key)
	if err != nil {
		return 0, err
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %w", err)
	}

	return intVal, nil
}

// GetConfigBool retrieves a configuration value as boolean
func (s *SQLiteDB) GetConfigBool(section, subsection, key string) (bool, error) {
	value, err := s.GetConfig(section, subsection, key)
	if err != nil {
		return false, err
	}

	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "yes" || value == "1", nil
}

// GetConfigList retrieves a configuration value as string slice
func (s *SQLiteDB) GetConfigList(section, subsection, key string) ([]string, error) {
	value, err := s.GetConfig(section, subsection, key)
	if err != nil {
		return nil, err
	}

	if value == "" {
		return []string{}, nil
	}

	// Split by comma and trim spaces
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// SetConfig sets a configuration value
func (s *SQLiteDB) SetConfig(section, subsection, key, value, valueType, modifiedBy string) error {
	// Use INSERT OR REPLACE to handle both insert and update
	query := `
		INSERT INTO config_settings (section, subsection, key, value, value_type, modified_by)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(section, subsection, key) 
		DO UPDATE SET 
			value = excluded.value,
			modified_by = excluded.modified_by,
			modified_at = CURRENT_TIMESTAMP
	`

	var subsectionVal interface{}
	if subsection == "" {
		subsectionVal = nil
	} else {
		subsectionVal = subsection
	}

	_, err := s.db.Exec(query, section, subsectionVal, key, value, valueType, modifiedBy)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	return nil
}

// GetAllUsers retrieves all user records
func (s *SQLiteDB) GetAllUsers() ([]UserRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, username, first_name, last_name, password_hash, password_salt, password_algo, password_updated_at, failed_attempts, locked_until, security_level, created_date, last_login, email, country, locations
		FROM users
		ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []UserRecord
	for rows.Next() {
		var user UserRecord
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.PasswordHash,
			&user.PasswordSalt,
			&user.PasswordAlgo,
			&user.PasswordUpdatedAt,
			&user.FailedAttempts,
			&user.LockedUntil,
			&user.SecurityLevel,
			&user.CreatedDate,
			&user.LastLogin,
			&user.Email,
			&user.Country,
			&user.Locations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// GetAllConfigValues retrieves all configuration values
func (s *SQLiteDB) GetAllConfigValues() ([]ConfigValue, error) {
	rows, err := s.db.Query(`
		SELECT id, section, COALESCE(subsection, ''), key, value, value_type,
			   COALESCE(default_value, ''), COALESCE(description, ''),
			   created_at, modified_at, modified_by
		FROM config_settings
		ORDER BY section, subsection, key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []ConfigValue
	for rows.Next() {
		var v ConfigValue
		err := rows.Scan(&v.ID, &v.Section, &v.Subsection, &v.Key, &v.Value,
			&v.ValueType, &v.DefaultValue, &v.Description,
			&v.CreatedAt, &v.ModifiedAt, &v.ModifiedBy)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, rows.Err()
}
