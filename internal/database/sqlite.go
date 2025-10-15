package database

import (
	"database/sql"
	"encoding/json"
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

	// Check if we need to migrate from old menu schema
	var currentVersion int
	err = tx.QueryRow(`SELECT version FROM schema_version WHERE id = 1`).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// If version is 1, check if menus table has obsolete columns and migrate
	if currentVersion == 1 {
		// Check if obsolete columns exist
		var count int
		err = tx.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('menus')
			WHERE name IN ('help_file', 'long_help_file', 'fallback_menu', 'forced_help_level')
		`).Scan(&count)
		if err == nil && count > 0 {
			// Migrate by recreating table without obsolete columns
			_, err = tx.Exec(`
				ALTER TABLE menus RENAME TO menus_old
			`)
			if err != nil {
				return fmt.Errorf("failed to rename menus table: %w", err)
			}

			// Create new menus table without obsolete columns
			_, err = tx.Exec(`
				CREATE TABLE menus (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT UNIQUE NOT NULL,
					titles TEXT, -- JSON array
					prompt TEXT,
					acs_required TEXT,
					password TEXT,
					generic_columns INTEGER DEFAULT 4,
					generic_bracket_color INTEGER DEFAULT 1,
					generic_command_color INTEGER DEFAULT 9,
					generic_desc_color INTEGER DEFAULT 1,
					clear_screen BOOLEAN DEFAULT 0
				)
			`)
			if err != nil {
				return fmt.Errorf("failed to create new menus table: %w", err)
			}

			// Copy data from old table to new table, excluding obsolete columns and converting flags
			_, err = tx.Exec(`
				INSERT INTO menus (id, name, titles, prompt, acs_required, password, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen)
				SELECT id, name, titles, prompt, acs_required, password, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color,
					   CASE WHEN substr(flags, 1, 1) = 'C' THEN 1 ELSE 0 END
				FROM menus_old
			`)
			if err != nil {
				return fmt.Errorf("failed to copy data to new menus table: %w", err)
			}

			// Drop old table
			_, err = tx.Exec(`DROP TABLE menus_old`)
			if err != nil {
				return fmt.Errorf("failed to drop old menus table: %w", err)
			}

			// Update schema version
			_, err = tx.Exec(`
				UPDATE schema_version SET version = 2, description = 'Removed obsolete menu fields: help_file, long_help_file, fallback_menu, forced_help_level'
				WHERE id = 1
			`)
			if err != nil {
				return fmt.Errorf("failed to update schema version: %w", err)
			}
		} else if currentVersion == 2 {
			// Check if menu_commands table has obsolete columns and migrate
			var count int
			err = tx.QueryRow(`
				SELECT COUNT(*) FROM pragma_table_info('menu_commands')
				WHERE name IN ('long_description', 'flags')
			`).Scan(&count)
			if err == nil && count > 0 {
				// Migrate by recreating table without obsolete columns
				_, err = tx.Exec(`
					ALTER TABLE menu_commands RENAME TO menu_commands_old
				`)
				if err != nil {
					return fmt.Errorf("failed to rename menu_commands table: %w", err)
				}

				// Create new menu_commands table without obsolete columns
				_, err = tx.Exec(`
					CREATE TABLE menu_commands (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						menu_id INTEGER NOT NULL,
						command_number INTEGER NOT NULL,
						keys TEXT,
						short_description TEXT,
						acs_required TEXT,
						cmdkeys TEXT,
						options TEXT,
						FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
					)
				`)
				if err != nil {
					return fmt.Errorf("failed to create new menu_commands table: %w", err)
				}

				// Copy data from old table to new table, excluding obsolete columns
				_, err = tx.Exec(`
					INSERT INTO menu_commands (id, menu_id, command_number, keys, short_description, acs_required, cmdkeys, options)
					SELECT id, menu_id, command_number, keys, short_description, acs_required, cmdkeys, options
					FROM menu_commands_old
				`)
				if err != nil {
					return fmt.Errorf("failed to copy data to new menu_commands table: %w", err)
				}

				// Drop old table
				_, err = tx.Exec(`DROP TABLE menu_commands_old`)
				if err != nil {
					return fmt.Errorf("failed to drop old menu_commands table: %w", err)
				}

				// Update schema version
				_, err = tx.Exec(`
					UPDATE schema_version SET version = 3, description = 'Removed obsolete menu_commands fields: long_description, flags'
					WHERE id = 1
				`)
				if err != nil {
					return fmt.Errorf("failed to update schema version: %w", err)
				}
			}
		}
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
		return fmt.Errorf("failed to create trigger: %w", err)
	}

	// Create menus table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS menus (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			titles TEXT, -- JSON array
			prompt TEXT,
			acs_required TEXT,
			password TEXT,
			generic_columns INTEGER DEFAULT 4,
			generic_bracket_color INTEGER DEFAULT 1,
			generic_command_color INTEGER DEFAULT 9,
			generic_desc_color INTEGER DEFAULT 1,
			clear_screen BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create menus: %w", err)
	}

	// Create menu_commands table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS menu_commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			menu_id INTEGER NOT NULL,
			command_number INTEGER NOT NULL,
			keys TEXT,
			short_description TEXT,
			acs_required TEXT,
			cmdkeys TEXT,
			options TEXT,
			FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create menu_commands: %w", err)
	}

	// Create index on menu_commands
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_menu_commands_menu_id
					  ON menu_commands(menu_id)`)
	if err != nil {
		return fmt.Errorf("failed to create menu_commands index: %w", err)
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
		SELECT id, username, first_name, last_name, password_hash, password_salt, password_algo, password_updated_at, failed_attempts, locked_until, security_level, created_date, last_login, email, locations
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

// CreateMenu creates a new menu
func (s *SQLiteDB) CreateMenu(menu *Menu) (int64, error) {
	titlesJSON, err := json.Marshal(menu.Titles)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal titles: %w", err)
	}

	result, err := s.db.Exec(`
		INSERT INTO menus (name, titles, prompt, acs_required, password, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		menu.Name, string(titlesJSON), menu.Prompt, menu.ACSRequired, menu.Password, menu.GenericColumns, menu.GenericBracketColor, menu.GenericCommandColor, menu.GenericDescColor, menu.ClearScreen)
	if err != nil {
		return 0, fmt.Errorf("failed to create menu: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get menu ID: %w", err)
	}

	return id, nil
}

// GetMenuByName retrieves a menu by name
func (s *SQLiteDB) GetMenuByName(name string) (*Menu, error) {
	var menu Menu
	var titlesJSON string

	err := s.db.QueryRow(`
		SELECT id, name, titles, prompt, acs_required, password, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen
		FROM menus WHERE name = ?`, name).Scan(
		&menu.ID, &menu.Name, &titlesJSON, &menu.Prompt, &menu.ACSRequired, &menu.Password, &menu.GenericColumns, &menu.GenericBracketColor, &menu.GenericCommandColor, &menu.GenericDescColor, &menu.ClearScreen)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("menu not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	err = json.Unmarshal([]byte(titlesJSON), &menu.Titles)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal titles: %w", err)
	}

	return &menu, nil
}

// GetMenuByID retrieves a menu by ID
func (s *SQLiteDB) GetMenuByID(id int64) (*Menu, error) {
	var menu Menu
	var titlesJSON string

	err := s.db.QueryRow(`
		SELECT id, name, titles, prompt, acs_required, password, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen
		FROM menus WHERE id = ?`, id).Scan(
		&menu.ID, &menu.Name, &titlesJSON, &menu.Prompt, &menu.ACSRequired, &menu.Password, &menu.GenericColumns, &menu.GenericBracketColor, &menu.GenericCommandColor, &menu.GenericDescColor, &menu.ClearScreen)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("menu not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	err = json.Unmarshal([]byte(titlesJSON), &menu.Titles)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal titles: %w", err)
	}

	return &menu, nil
}

// GetAllMenus retrieves all menus
func (s *SQLiteDB) GetAllMenus() ([]Menu, error) {
	rows, err := s.db.Query(`
		SELECT id, name, titles, prompt, acs_required, password, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen
		FROM menus ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query menus: %w", err)
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var menu Menu
		var titlesJSON string
		err := rows.Scan(&menu.ID, &menu.Name, &titlesJSON, &menu.Prompt, &menu.ACSRequired, &menu.Password, &menu.GenericColumns, &menu.GenericBracketColor, &menu.GenericCommandColor, &menu.GenericDescColor, &menu.ClearScreen)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu: %w", err)
		}

		err = json.Unmarshal([]byte(titlesJSON), &menu.Titles)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal titles: %w", err)
		}

		menus = append(menus, menu)
	}

	return menus, rows.Err()
}

// UpdateMenu updates a menu
func (s *SQLiteDB) UpdateMenu(menu *Menu) error {
	titlesJSON, err := json.Marshal(menu.Titles)
	if err != nil {
		return fmt.Errorf("failed to marshal titles: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE menus SET name = ?, titles = ?, prompt = ?, acs_required = ?, password = ?, generic_columns = ?, generic_bracket_color = ?, generic_command_color = ?, generic_desc_color = ?, clear_screen = ?
		WHERE id = ?`,
		menu.Name, string(titlesJSON), menu.Prompt, menu.ACSRequired, menu.Password, menu.GenericColumns, menu.GenericBracketColor, menu.GenericCommandColor, menu.GenericDescColor, menu.ClearScreen, menu.ID)
	if err != nil {
		return fmt.Errorf("failed to update menu: %w", err)
	}

	return nil
}

// DeleteMenu deletes a menu
func (s *SQLiteDB) DeleteMenu(id int64) error {
	_, err := s.db.Exec(`DELETE FROM menus WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete menu: %w", err)
	}
	return nil
}

// CreateMenuCommand creates a new menu command
func (s *SQLiteDB) CreateMenuCommand(cmd *MenuCommand) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO menu_commands (menu_id, command_number, keys, short_description, acs_required, cmdkeys, options)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		cmd.MenuID, cmd.CommandNumber, cmd.Keys, cmd.ShortDescription, cmd.ACSRequired, cmd.CmdKeys, cmd.Options)
	if err != nil {
		return 0, fmt.Errorf("failed to create menu command: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get menu command ID: %w", err)
	}

	return id, nil
}

// GetMenuCommands retrieves all commands for a menu
func (s *SQLiteDB) GetMenuCommands(menuID int) ([]MenuCommand, error) {
	rows, err := s.db.Query(`
		SELECT id, menu_id, command_number, keys, short_description, acs_required, cmdkeys, options
		FROM menu_commands WHERE menu_id = ? ORDER BY command_number`, menuID)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu commands: %w", err)
	}
	defer rows.Close()

	var commands []MenuCommand
	for rows.Next() {
		var cmd MenuCommand
		err := rows.Scan(&cmd.ID, &cmd.MenuID, &cmd.CommandNumber, &cmd.Keys, &cmd.ShortDescription, &cmd.ACSRequired, &cmd.CmdKeys, &cmd.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu command: %w", err)
		}
		commands = append(commands, cmd)
	}

	return commands, rows.Err()
}

// UpdateMenuCommand updates a menu command
func (s *SQLiteDB) UpdateMenuCommand(cmd *MenuCommand) error {
	_, err := s.db.Exec(`
		UPDATE menu_commands SET menu_id = ?, command_number = ?, keys = ?, short_description = ?, acs_required = ?, cmdkeys = ?, options = ?
		WHERE id = ?`,
		cmd.MenuID, cmd.CommandNumber, cmd.Keys, cmd.ShortDescription, cmd.ACSRequired, cmd.CmdKeys, cmd.Options, cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to update menu command: %w", err)
	}
	return nil
}

// DeleteMenuCommand deletes a menu command
func (s *SQLiteDB) DeleteMenuCommand(id int64) error {
	_, err := s.db.Exec(`DELETE FROM menu_commands WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete menu command: %w", err)
	}
	return nil
}
