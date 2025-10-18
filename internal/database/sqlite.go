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

func clampBracket(value, fallback string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return fallback
	}
	if len(runes) > 2 {
		runes = runes[:2]
	}
	return string(runes)
}

func sanitizeDisplayMode(value string) string {
	switch value {
	case DisplayModeHeaderGenerated, DisplayModeThemeOnly:
		return value
	case DisplayModeTitlesGenerated:
		return value
	default:
		return DisplayModeTitlesGenerated
	}
}

func normalizeMenuDefaults(menu *Menu) {
	if menu == nil {
		return
	}
	menu.LeftBracket = clampBracket(menu.LeftBracket, "[")
	menu.RightBracket = clampBracket(menu.RightBracket, "]")
	menu.DisplayMode = sanitizeDisplayMode(menu.DisplayMode)
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
			generic_columns INTEGER DEFAULT 4,
			generic_bracket_color INTEGER DEFAULT 1,
			generic_command_color INTEGER DEFAULT 9,
			generic_desc_color INTEGER DEFAULT 1,
			clear_screen BOOLEAN DEFAULT 0,
			left_bracket TEXT DEFAULT '[',
			right_bracket TEXT DEFAULT ']',
			display_mode TEXT DEFAULT 'titles_generated',
			node_activity TEXT DEFAULT ''
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
			position_number INTEGER NOT NULL DEFAULT 0,
			keys TEXT,
			short_description TEXT,
			long_description TEXT,
			acs_required TEXT,
			cmdkeys TEXT,
			options TEXT,
			node_activity TEXT,
			active BOOLEAN DEFAULT 1,
			hidden BOOLEAN DEFAULT 0,
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

	// Create conferences table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS conferences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			sec_level TEXT NOT NULL,
			tagline TEXT,
			hidden BOOLEAN NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create conferences: %w", err)
	}

	// Create message_areas table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS message_areas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			file TEXT NOT NULL UNIQUE,
			path TEXT NOT NULL,
			read_sec_level TEXT NOT NULL,
			write_sec_level TEXT NOT NULL,
			area_type TEXT NOT NULL,
			echo_tag TEXT,
			real_names BOOLEAN NOT NULL DEFAULT 0,
			address TEXT,
			conference_id INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (conference_id) REFERENCES conferences(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create message_areas: %w", err)
	}

	if _, err := tx.Exec(`ALTER TABLE message_areas ADD COLUMN conference_id INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add conference_id column to message_areas: %w", err)
		}
	}

	// Ensure new columns exist for legacy databases
	if _, err := tx.Exec(`ALTER TABLE menus ADD COLUMN left_bracket TEXT DEFAULT '['`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add left_bracket column: %w", err)
		}
	}
	if _, err := tx.Exec(`ALTER TABLE menus ADD COLUMN right_bracket TEXT DEFAULT ']'`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add right_bracket column: %w", err)
		}
	}
	if _, err := tx.Exec(`ALTER TABLE menus ADD COLUMN display_mode TEXT DEFAULT 'titles_generated'`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add display_mode column: %w", err)
		}
	}
	if _, err := tx.Exec(`ALTER TABLE menus ADD COLUMN node_activity TEXT DEFAULT ''`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add node_activity column: %w", err)
		}
	}
	if _, err := tx.Exec(`ALTER TABLE menu_commands ADD COLUMN long_description TEXT`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add long_description column: %w", err)
		}
	}
	if _, err := tx.Exec(`ALTER TABLE menu_commands ADD COLUMN node_activity TEXT`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add node_activity column: %w", err)
		}
	}
	if _, err := tx.Exec(`ALTER TABLE menu_commands ADD COLUMN hidden BOOLEAN DEFAULT 0`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("failed to add hidden column: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Ensure user-related schema is initialized as well.
	if err := s.InitializeUserSchema(); err != nil {
		return err
	}

	// Seed default menu structure and message areas
	return SeedDefaultMainMenu(s)
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
	normalizeMenuDefaults(menu)

	result, err := s.db.Exec(`
		INSERT INTO menus (name, titles, prompt, acs_required, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen, left_bracket, right_bracket, display_mode, node_activity)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		menu.Name, string(titlesJSON), menu.Prompt, menu.ACSRequired, menu.GenericColumns, menu.GenericBracketColor, menu.GenericCommandColor, menu.GenericDescColor, menu.ClearScreen, menu.LeftBracket, menu.RightBracket, menu.DisplayMode, menu.NodeActivity)
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
		SELECT id, name, titles, prompt, acs_required, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen, left_bracket, right_bracket, display_mode, node_activity
		FROM menus WHERE name = ?`, name).Scan(
		&menu.ID, &menu.Name, &titlesJSON, &menu.Prompt, &menu.ACSRequired, &menu.GenericColumns, &menu.GenericBracketColor, &menu.GenericCommandColor, &menu.GenericDescColor, &menu.ClearScreen, &menu.LeftBracket, &menu.RightBracket, &menu.DisplayMode, &menu.NodeActivity)
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
	normalizeMenuDefaults(&menu)

	return &menu, nil
}

// GetMenuByID retrieves a menu by ID
func (s *SQLiteDB) GetMenuByID(id int64) (*Menu, error) {
	var menu Menu
	var titlesJSON string

	err := s.db.QueryRow(`
		SELECT id, name, titles, prompt, acs_required, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen, left_bracket, right_bracket, display_mode, node_activity
		FROM menus WHERE id = ?`, id).Scan(
		&menu.ID, &menu.Name, &titlesJSON, &menu.Prompt, &menu.ACSRequired, &menu.GenericColumns, &menu.GenericBracketColor, &menu.GenericCommandColor, &menu.GenericDescColor, &menu.ClearScreen, &menu.LeftBracket, &menu.RightBracket, &menu.DisplayMode, &menu.NodeActivity)
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
	normalizeMenuDefaults(&menu)

	return &menu, nil
}

// GetAllMenus retrieves all menus
func (s *SQLiteDB) GetAllMenus() ([]Menu, error) {
	rows, err := s.db.Query(`
		SELECT id, name, titles, prompt, acs_required, generic_columns, generic_bracket_color, generic_command_color, generic_desc_color, clear_screen, left_bracket, right_bracket, display_mode, node_activity
		FROM menus ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query menus: %w", err)
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var menu Menu
		var titlesJSON string
		err := rows.Scan(&menu.ID, &menu.Name, &titlesJSON, &menu.Prompt, &menu.ACSRequired, &menu.GenericColumns, &menu.GenericBracketColor, &menu.GenericCommandColor, &menu.GenericDescColor, &menu.ClearScreen, &menu.LeftBracket, &menu.RightBracket, &menu.DisplayMode, &menu.NodeActivity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu: %w", err)
		}

		err = json.Unmarshal([]byte(titlesJSON), &menu.Titles)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal titles: %w", err)
		}
		normalizeMenuDefaults(&menu)

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
	normalizeMenuDefaults(menu)

	_, err = s.db.Exec(`
		UPDATE menus SET name = ?, titles = ?, prompt = ?, acs_required = ?, generic_columns = ?, generic_bracket_color = ?, generic_command_color = ?, generic_desc_color = ?, clear_screen = ?, left_bracket = ?, right_bracket = ?, display_mode = ?, node_activity = ?
		WHERE id = ?`,
		menu.Name, string(titlesJSON), menu.Prompt, menu.ACSRequired, menu.GenericColumns, menu.GenericBracketColor, menu.GenericCommandColor, menu.GenericDescColor, menu.ClearScreen, menu.LeftBracket, menu.RightBracket, menu.DisplayMode, menu.NodeActivity, menu.ID)
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
	if cmd.PositionNumber <= 0 {
		var nextPosition int
		if err := s.db.QueryRow(`SELECT COALESCE(MAX(position_number), 0) + 1 FROM menu_commands WHERE menu_id = ?`, cmd.MenuID).
			Scan(&nextPosition); err != nil {
			return 0, fmt.Errorf("failed to determine next command position: %w", err)
		}
		cmd.PositionNumber = nextPosition
	}

	result, err := s.db.Exec(`
		INSERT INTO menu_commands (menu_id, position_number, keys, short_description, long_description, acs_required, cmdkeys, options, node_activity, active, hidden)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cmd.MenuID, cmd.PositionNumber, cmd.Keys, cmd.ShortDescription, cmd.LongDescription, cmd.ACSRequired, cmd.CmdKeys, cmd.Options, cmd.NodeActivity, cmd.Active, cmd.Hidden)
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
		SELECT id, menu_id, position_number, keys, short_description, long_description, acs_required, cmdkeys, options, node_activity, active, hidden
		FROM menu_commands WHERE menu_id = ? ORDER BY position_number, id`, menuID)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu commands: %w", err)
	}
	defer rows.Close()

	var commands []MenuCommand
	for rows.Next() {
		var cmd MenuCommand
		err := rows.Scan(&cmd.ID, &cmd.MenuID, &cmd.PositionNumber, &cmd.Keys, &cmd.ShortDescription, &cmd.LongDescription, &cmd.ACSRequired, &cmd.CmdKeys, &cmd.Options, &cmd.NodeActivity, &cmd.Active, &cmd.Hidden)
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
		UPDATE menu_commands SET menu_id = ?, position_number = ?, keys = ?, short_description = ?, long_description = ?, acs_required = ?, cmdkeys = ?, options = ?, node_activity = ?, active = ?, hidden = ?
		WHERE id = ?`,
		cmd.MenuID, cmd.PositionNumber, cmd.Keys, cmd.ShortDescription, cmd.LongDescription, cmd.ACSRequired, cmd.CmdKeys, cmd.Options, cmd.NodeActivity, cmd.Active, cmd.Hidden, cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to update menu command: %w", err)
	}
	return nil
}

// DeleteMenuCommand deletes a menu command
func (s *SQLiteDB) DeleteMenuCommand(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var menuID int
	err = tx.QueryRow(`SELECT menu_id FROM menu_commands WHERE id = ?`, id).Scan(&menuID)
	if err == sql.ErrNoRows {
		return tx.Commit()
	}
	if err != nil {
		return fmt.Errorf("failed to find menu command: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM menu_commands WHERE id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete menu command: %w", err)
	}

	rows, err := tx.Query(`SELECT id FROM menu_commands WHERE menu_id = ? ORDER BY position_number, id`, menuID)
	if err != nil {
		return fmt.Errorf("failed to query remaining menu commands: %w", err)
	}
	defer rows.Close()

	position := 1
	for rows.Next() {
		var cmdID int64
		if err := rows.Scan(&cmdID); err != nil {
			return fmt.Errorf("failed to scan menu command id: %w", err)
		}
		if _, err := tx.Exec(`UPDATE menu_commands SET position_number = ? WHERE id = ?`, position, cmdID); err != nil {
			return fmt.Errorf("failed to renumber menu commands: %w", err)
		}
		position++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate menu commands: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// CreateConference inserts a new conference record
func (s *SQLiteDB) CreateConference(conf *Conference) (int64, error) {
	if conf == nil {
		return 0, fmt.Errorf("conference cannot be nil")
	}

	result, err := s.db.Exec(`
		INSERT INTO conferences (name, description, sec_level, tagline, hidden)
		VALUES (?, ?, ?, ?, ?)
	`, conf.Name, conf.Description, conf.SecLevel, conf.Tagline, conf.Hidden)
	if err != nil {
		return 0, fmt.Errorf("failed to create conference: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get conference ID: %w", err)
	}

	conf.ID = int(id)
	return id, nil
}

// GetConferenceByID retrieves a conference by its ID
func (s *SQLiteDB) GetConferenceByID(id int64) (*Conference, error) {
	var conf Conference
	var hiddenInt int

	err := s.db.QueryRow(`
		SELECT id, name, description, sec_level, tagline, hidden
		FROM conferences WHERE id = ?
	`, id).Scan(&conf.ID, &conf.Name, &conf.Description, &conf.SecLevel, &conf.Tagline, &hiddenInt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conference not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conference: %w", err)
	}

	conf.Hidden = hiddenInt != 0
	return &conf, nil
}

// GetAllConferences returns all conferences ordered by name
func (s *SQLiteDB) GetAllConferences() ([]Conference, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, sec_level, tagline, hidden
		FROM conferences
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query conferences: %w", err)
	}
	defer rows.Close()

	var conferences []Conference
	for rows.Next() {
		var conf Conference
		var hiddenInt int
		if err := rows.Scan(&conf.ID, &conf.Name, &conf.Description, &conf.SecLevel, &conf.Tagline, &hiddenInt); err != nil {
			return nil, fmt.Errorf("failed to scan conference: %w", err)
		}
		conf.Hidden = hiddenInt != 0
		conferences = append(conferences, conf)
	}

	return conferences, rows.Err()
}

// UpdateConference updates an existing conference
func (s *SQLiteDB) UpdateConference(conf *Conference) error {
	if conf == nil {
		return fmt.Errorf("conference cannot be nil")
	}

	_, err := s.db.Exec(`
		UPDATE conferences
		SET name = ?, description = ?, sec_level = ?, tagline = ?, hidden = ?
		WHERE id = ?
	`, conf.Name, conf.Description, conf.SecLevel, conf.Tagline, conf.Hidden, conf.ID)
	if err != nil {
		return fmt.Errorf("failed to update conference: %w", err)
	}

	return nil
}

// DeleteConference removes a conference by ID
func (s *SQLiteDB) DeleteConference(id int64) error {
	_, err := s.db.Exec(`DELETE FROM conferences WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete conference: %w", err)
	}
	return nil
}

// CreateMessageArea inserts a new message area record
func (s *SQLiteDB) CreateMessageArea(area *MessageArea) (int64, error) {
	if area == nil {
		return 0, fmt.Errorf("message area cannot be nil")
	}
	if area.ConferenceID <= 0 {
		return 0, fmt.Errorf("message area must be assigned to a conference")
	}

	result, err := s.db.Exec(`
		INSERT INTO message_areas (name, file, path, read_sec_level, write_sec_level, area_type, echo_tag, real_names, address, conference_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, area.Name, area.File, area.Path, area.ReadSecLevel, area.WriteSecLevel, area.AreaType, area.EchoTag, area.RealNames, area.Address, area.ConferenceID)
	if err != nil {
		return 0, fmt.Errorf("failed to create message area: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get message area ID: %w", err)
	}

	area.ID = int(id)
	return id, nil
}

// GetMessageAreaByID retrieves a message area by ID
func (s *SQLiteDB) GetMessageAreaByID(id int64) (*MessageArea, error) {
	var area MessageArea
	var realNamesInt int
	var conferenceName sql.NullString

	err := s.db.QueryRow(`
		SELECT ma.id, ma.name, ma.file, ma.path, ma.read_sec_level, ma.write_sec_level, ma.area_type, ma.echo_tag, ma.real_names, ma.address, ma.conference_id, c.name
		FROM message_areas ma
		LEFT JOIN conferences c ON c.id = ma.conference_id
		WHERE ma.id = ?
	`, id).Scan(&area.ID, &area.Name, &area.File, &area.Path, &area.ReadSecLevel, &area.WriteSecLevel, &area.AreaType, &area.EchoTag, &realNamesInt, &area.Address, &area.ConferenceID, &conferenceName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message area not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message area: %w", err)
	}

	area.RealNames = realNamesInt != 0
	area.ConferenceName = conferenceName.String
	return &area, nil
}

// GetAllMessageAreas returns all message areas ordered by name
func (s *SQLiteDB) GetAllMessageAreas() ([]MessageArea, error) {
	rows, err := s.db.Query(`
		SELECT ma.id, ma.name, ma.file, ma.path, ma.read_sec_level, ma.write_sec_level, ma.area_type, ma.echo_tag, ma.real_names, ma.address, ma.conference_id, COALESCE(c.name, '')
		FROM message_areas ma
		LEFT JOIN conferences c ON c.id = ma.conference_id
		ORDER BY ma.name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query message areas: %w", err)
	}
	defer rows.Close()

	var areas []MessageArea
	for rows.Next() {
		var area MessageArea
		var realNamesInt int
		var conferenceName string
		if err := rows.Scan(&area.ID, &area.Name, &area.File, &area.Path, &area.ReadSecLevel, &area.WriteSecLevel, &area.AreaType, &area.EchoTag, &realNamesInt, &area.Address, &area.ConferenceID, &conferenceName); err != nil {
			return nil, fmt.Errorf("failed to scan message area: %w", err)
		}
		area.RealNames = realNamesInt != 0
		area.ConferenceName = conferenceName
		areas = append(areas, area)
	}

	return areas, rows.Err()
}

// UpdateMessageArea updates an existing message area record
func (s *SQLiteDB) UpdateMessageArea(area *MessageArea) error {
	if area == nil {
		return fmt.Errorf("message area cannot be nil")
	}

	if area.ConferenceID <= 0 {
		return fmt.Errorf("message area must be assigned to a conference")
	}

	_, err := s.db.Exec(`
		UPDATE message_areas
		SET name = ?, file = ?, path = ?, read_sec_level = ?, write_sec_level = ?, area_type = ?, echo_tag = ?, real_names = ?, address = ?, conference_id = ?
		WHERE id = ?
	`, area.Name, area.File, area.Path, area.ReadSecLevel, area.WriteSecLevel, area.AreaType, area.EchoTag, area.RealNames, area.Address, area.ConferenceID, area.ID)
	if err != nil {
		return fmt.Errorf("failed to update message area: %w", err)
	}

	return nil
}

// DeleteMessageArea removes a message area by ID
func (s *SQLiteDB) DeleteMessageArea(id int64) error {
	_, err := s.db.Exec(`DELETE FROM message_areas WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete message area: %w", err)
	}
	return nil
}
