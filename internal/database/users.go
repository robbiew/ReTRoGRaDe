package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// execer allows shared logic between *sql.DB and *sql.Tx.
type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// NullString returns a sql.NullString helper.
func NullString(value string) sql.NullString {
	if strings.TrimSpace(value) == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: value, Valid: true}
}

// NullInt returns a sql.NullInt64 helper.
func NullInt(value int64) sql.NullInt64 {
	if value == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: value, Valid: true}
}

// WithTransaction executes fn inside a transaction.
func (s *SQLiteDB) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UserRecord represents a row in the users table.
type UserRecord struct {
	ID                int64
	Username          string
	FirstName         sql.NullString
	LastName          sql.NullString
	PasswordHash      string
	PasswordSalt      sql.NullString
	PasswordAlgo      sql.NullString
	PasswordUpdatedAt sql.NullString
	FailedAttempts    int
	LockedUntil       sql.NullString
	SecurityLevel     int
	CreatedDate       string
	LastLogin         sql.NullString
	Email             sql.NullString
	Locations         sql.NullString
}

// BBSSessionRecord represents a row in the bbs_sessions table.
type BBSSessionRecord struct {
	ID             int64
	UserID         int64
	NodeNumber     int
	SessionStart   string
	LastActivity   string
	TimeLeft       int
	CallsToday     int
	Status         string
	IPAddress      sql.NullString
	ConnectionType sql.NullString
	CurrentArea    sql.NullString
	CurrentMenu    sql.NullString
}

// BBSConfigRecord represents a row in the bbs_config table.
type BBSConfigRecord struct {
	ID          int64
	Section     string
	Subsection  sql.NullString
	Key         string
	Value       string
	ValueType   string
	Description sql.NullString
	ModifiedBy  sql.NullString
	ModifiedAt  string
}

// UserPreferenceRecord represents a row in the user_preferences table.
type UserPreferenceRecord struct {
	UserID          int64
	PreferenceKey   string
	PreferenceValue string
	Category        sql.NullString
}

// SecurityLevelRecord represents a row in the security_levels table.
type SecurityLevelRecord struct {
	ID               int64
	Name             string
	SecLevel         int
	MinsPerDay       int
	TimeoutMins      int
	CanDeleteOwnMsgs bool
	CanDeleteMsgs    bool
	Invisible        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// InitializeUserSchema ensures the user-related tables exist.
const sqliteTimeFormat = time.RFC3339Nano

func (s *SQLiteDB) InitializeUserSchema() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	createStatements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE COLLATE NOCASE,
			first_name TEXT,
			last_name TEXT,
			password_hash TEXT NOT NULL,
			password_salt TEXT,
			password_algo TEXT,
			password_updated_at TEXT,
			failed_attempts INTEGER NOT NULL DEFAULT 0,
			locked_until TEXT,
			security_level INTEGER NOT NULL DEFAULT 0,
			created_date TEXT NOT NULL,
			last_login TEXT,
			email TEXT UNIQUE,
			locations TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS bbs_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			node_number INTEGER NOT NULL,
			session_start TEXT NOT NULL,
			last_activity TEXT NOT NULL,
			time_left INTEGER NOT NULL DEFAULT 0,
			calls_today INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'active',
			ip_address TEXT,
			connection_type TEXT,
			current_area TEXT,
			current_menu TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS bbs_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			section TEXT NOT NULL,
			subsection TEXT,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			value_type TEXT NOT NULL DEFAULT 'string',
			description TEXT,
			modified_by TEXT,
			modified_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (section, subsection, key)
		)`,
		`CREATE TABLE IF NOT EXISTS user_preferences (
			user_id INTEGER NOT NULL,
			preference_key TEXT NOT NULL,
			preference_value TEXT NOT NULL,
			category TEXT,
			PRIMARY KEY (user_id, preference_key),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS user_details (
			user_id INTEGER NOT NULL,
			attrib TEXT NOT NULL COLLATE NOCASE,
			value TEXT,
			PRIMARY KEY (user_id, attrib),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS user_subscriptions (
			user_id INTEGER NOT NULL,
			msgbase TEXT NOT NULL,
			PRIMARY KEY (user_id, msgbase),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS user_lastread (
			user_id INTEGER NOT NULL,
			msgbase TEXT NOT NULL,
			last_message_id INTEGER,
			PRIMARY KEY (user_id, msgbase),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS auth_audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			username TEXT,
			event_type TEXT NOT NULL,
			ip_address TEXT,
			metadata TEXT,
			context TEXT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS security_levels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT 'Unknown Name',
			sec_level INTEGER NOT NULL UNIQUE,
			mins_per_day INTEGER NOT NULL,
			timeout_mins INTEGER NOT NULL,
			can_delete_own_msgs INTEGER NOT NULL DEFAULT 0,
			can_delete_msgs INTEGER NOT NULL DEFAULT 0,
			invisible INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range createStatements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute schema statement: %w", err)
		}
	}

	indexStatements := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_user_details_user ON user_details(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user ON user_subscriptions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_lastread_user ON user_lastread(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_audit_user ON auth_audit(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_audit_created ON auth_audit(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_audit_event ON auth_audit(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_bbs_sessions_user ON bbs_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bbs_sessions_node ON bbs_sessions(node_number)`,
		`CREATE INDEX IF NOT EXISTS idx_bbs_sessions_status ON bbs_sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_user_preferences_category ON user_preferences(category)`,
	}

	for _, stmt := range indexStatements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute index statement: %w", err)
		}
	}

	// Seed default security levels
	seedStatements := []struct {
		name             string
		secLevel         int
		minsPerDay       int
		timeoutMins      int
		canDeleteOwnMsgs bool
		canDeleteMsgs    bool
		invisible        bool
	}{
		{"Users", 10, 60, 5, false, false, false},
		{"SysOps", 100, 250, 20, true, true, false},
	}

	now := time.Now().Format(sqliteTimeFormat)
	for _, seed := range seedStatements {
		// Check if security level already exists
		var count int
		err := tx.QueryRow(`SELECT COUNT(*) FROM security_levels WHERE sec_level = ?`, seed.secLevel).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check existing security level %d: %w", seed.secLevel, err)
		}

		if count == 0 {
			// Insert the security level
			_, err = tx.Exec(`
				INSERT INTO security_levels (name, sec_level, mins_per_day, timeout_mins, can_delete_own_msgs, can_delete_msgs, invisible, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				seed.name,
				seed.secLevel,
				seed.minsPerDay,
				seed.timeoutMins,
				boolToInt(seed.canDeleteOwnMsgs),
				boolToInt(seed.canDeleteMsgs),
				boolToInt(seed.invisible),
				now,
				now,
			)
			if err != nil {
				return fmt.Errorf("failed to seed security level %d: %w", seed.secLevel, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit user schema transaction: %w", err)
	}

	return nil
}

// CreateUser inserts a new user row and returns the generated ID.
func (s *SQLiteDB) CreateUser(user *UserRecord) (int64, error) {
	return createUserExec(s.db, user)
}

// CreateUserTx inserts a user inside an existing transaction.
func (s *SQLiteDB) CreateUserTx(tx *sql.Tx, user *UserRecord) (int64, error) {
	return createUserExec(tx, user)
}

// UpdateUserTx updates a user inside an existing transaction.
func (s *SQLiteDB) UpdateUserTx(tx *sql.Tx, user *UserRecord) error {
	return updateUserExec(tx, user)
}

func createUserExec(ex execer, user *UserRecord) (int64, error) {
	result, err := ex.Exec(`
		INSERT INTO users (
			username, first_name, last_name, password_hash, password_salt, password_algo, password_updated_at,
			failed_attempts, locked_until, security_level, created_date, last_login, email, locations
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username,
		user.FirstName,
		user.LastName,
		user.PasswordHash,
		user.PasswordSalt,
		user.PasswordAlgo,
		user.PasswordUpdatedAt,
		user.FailedAttempts,
		nullOrString(user.LockedUntil),
		user.SecurityLevel,
		user.CreatedDate,
		user.LastLogin,
		user.Email,
		user.Locations,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve user ID: %w", err)
	}

	return id, nil
}

func updateUserExec(ex execer, user *UserRecord) error {
	_, err := ex.Exec(`
		UPDATE users
		SET password_hash = ?, password_salt = ?, password_algo = ?, password_updated_at = ?, failed_attempts = ?, locked_until = ?, security_level = ?, created_date = ?, last_login = ?, email = ?, first_name = ?, last_name = ?, locations = ?
		WHERE id = ?`,
		user.PasswordHash,
		user.PasswordSalt,
		user.PasswordAlgo,
		user.PasswordUpdatedAt,
		user.FailedAttempts,
		nullOrString(user.LockedUntil),
		user.SecurityLevel,
		user.CreatedDate,
		user.LastLogin,
		user.Email,
		user.Locations,
		user.ID,
		user.FirstName,
		user.LastName,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeleteUserTx deletes a user and dependent rows inside a transaction.
func (s *SQLiteDB) DeleteUserTx(tx *sql.Tx, userID int64) error {
	return deleteUserExec(tx, userID)
}

func deleteUserExec(ex execer, userID int64) error {
	statements := []string{
		`DELETE FROM user_details WHERE user_id = ?`,
		`DELETE FROM user_subscriptions WHERE user_id = ?`,
		`DELETE FROM user_lastread WHERE user_id = ?`,
		`DELETE FROM bbs_sessions WHERE user_id = ?`,
		`DELETE FROM user_preferences WHERE user_id = ?`,
		`DELETE FROM users WHERE id = ?`,
	}

	for _, stmt := range statements {
		if _, err := ex.Exec(stmt, userID); err != nil {
			return fmt.Errorf("failed to delete user data: %w", err)
		}
	}

	return nil
}

// GetUserByUsername retrieves a user row by username (case-insensitive).
func (s *SQLiteDB) GetUserByUsername(username string) (*UserRecord, error) {
	fmt.Printf("DEBUG: GetUserByUsername called with username: %s\n", username)
	row := s.db.QueryRow(`
		SELECT id, username, first_name, last_name, password_hash, password_salt, password_algo, password_updated_at, failed_attempts, locked_until, security_level, created_date, last_login, email, locations
		FROM users
		WHERE LOWER(username) = LOWER(?)`,
		username,
	)

	var user UserRecord
	if err := row.Scan(
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
	); err != nil {
		fmt.Printf("DEBUG: GetUserByUsername scan error: %v\n", err)
		if err == sql.ErrNoRows {
			fmt.Printf("DEBUG: User not found\n")
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	fmt.Printf("DEBUG: User found: ID=%d, Username=%s\n", user.ID, user.Username)
	return &user, nil
}

// GetUserByEmail retrieves a user row by email address (case-insensitive).
func (s *SQLiteDB) GetUserByEmail(email string) (*UserRecord, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	row := s.db.QueryRow(`
		SELECT id, username, first_name, last_name, password_hash, password_salt, password_algo, password_updated_at, failed_attempts, locked_until, security_level, created_date, last_login, email, locations
		FROM users
		WHERE LOWER(email) = LOWER(?)`,
		email,
	)

	var user UserRecord
	if err := row.Scan(
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
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user row by ID.
func (s *SQLiteDB) GetUserByID(userID int64) (*UserRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, username, first_name, last_name, password_hash, password_salt, password_algo, password_updated_at, failed_attempts, locked_until, security_level, created_date, last_login, email, locations
		FROM users
		WHERE id = ?`,
		userID,
	)

	var user UserRecord
	if err := row.Scan(
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
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates mutable fields for an existing user.
func (s *SQLiteDB) UpdateUser(user *UserRecord) error {
	_, err := s.db.Exec(`
		UPDATE users
		SET first_name = ?, last_name = ?, password_hash = ?, password_salt = ?, password_algo = ?, password_updated_at = ?, failed_attempts = ?, locked_until = ?, security_level = ?, created_date = ?, last_login = ?, email = ?, locations = ?
		WHERE id = ?`,
		user.FirstName,
		user.LastName,
		user.PasswordHash,
		user.PasswordSalt,
		user.PasswordAlgo,
		user.PasswordUpdatedAt,
		user.FailedAttempts,
		nullOrString(user.LockedUntil),
		user.SecurityLevel,
		user.CreatedDate,
		user.LastLogin,
		user.Email,
		user.Locations,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpsertUserDetail stores a key/value attribute for a user.
func (s *SQLiteDB) UpsertUserDetail(userID int64, attrib, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO user_details (user_id, attrib, value)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, attrib) DO UPDATE SET value = excluded.value`,
		userID,
		attrib,
		value,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert user detail: %w", err)
	}
	return nil
}

// GetUserDetails returns all key/value attributes for a user.
func (s *SQLiteDB) GetUserDetails(userID int64) (map[string]string, error) {
	rows, err := s.db.Query(`
		SELECT attrib, value
		FROM user_details
		WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query user details: %w", err)
	}
	defer rows.Close()

	details := make(map[string]string)
	for rows.Next() {
		var attrib string
		var value sql.NullString
		if err := rows.Scan(&attrib, &value); err != nil {
			return nil, fmt.Errorf("failed to scan user detail: %w", err)
		}
		details[attrib] = value.String
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user details: %w", err)
	}

	return details, nil
}

// IncrementFailedAttempts increases the failed login count and optionally sets a lockout.
func (s *SQLiteDB) IncrementFailedAttempts(userID int64, now time.Time, maxAttempts int, lockMinutes int) (int, *time.Time, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	var currentCount int
	var lockedUntil sql.NullString
	err = tx.QueryRow(`SELECT failed_attempts, locked_until FROM users WHERE id = ?`, userID).Scan(&currentCount, &lockedUntil)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, fmt.Errorf("user %d not found", userID)
		}
		return 0, nil, err
	}

	currentCount++
	var lockPtr *time.Time
	var lockValue interface{}

	if maxAttempts > 0 && currentCount >= maxAttempts {
		lockTime := now.Add(time.Duration(lockMinutes) * time.Minute)
		lockPtr = &lockTime
		lockValue = lockTime.Format(sqliteTimeFormat)
	} else if lockedUntil.Valid {
		if parsed, err := time.Parse(sqliteTimeFormat, lockedUntil.String); err == nil && parsed.After(now) {
			lockPtr = &parsed
			lockValue = lockedUntil.String
		} else {
			lockValue = nil
		}
	} else {
		lockValue = nil
	}

	if _, err := tx.Exec(`
		UPDATE users
		SET failed_attempts = ?, locked_until = ?
		WHERE id = ?`, currentCount, lockValue, userID); err != nil {
		return 0, nil, err
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}

	return currentCount, lockPtr, nil
}

// ResetFailedAttempts clears the failed attempt counter and lockout.
func (s *SQLiteDB) ResetFailedAttempts(userID int64, _ time.Time) error {
	_, err := s.db.Exec(`UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE id = ?`, userID)
	return err
}

// UpdatePassword updates hash metadata, resets failures, and records algorithm/salt information.
func (s *SQLiteDB) UpdatePassword(userID int64, hash, algo, salt string, now time.Time) error {
	_, err := s.db.Exec(`
		UPDATE users
		SET password_hash = ?, password_algo = ?, password_salt = ?, password_updated_at = ?, failed_attempts = 0, locked_until = NULL
		WHERE id = ?`,
		hash,
		NullString(algo),
		NullString(salt),
		now.Format(sqliteTimeFormat),
		userID,
	)
	return err
}

// InsertAuthAudit inserts an authentication audit trail entry.
func (s *SQLiteDB) InsertAuthAudit(entry *AuthAuditEntry) error {
	if entry == nil {
		return fmt.Errorf("auth audit entry is nil")
	}

	created := entry.CreatedAt
	if created.IsZero() {
		created = time.Now()
	}

	_, err := s.db.Exec(`
		INSERT INTO auth_audit (user_id, username, event_type, ip_address, metadata, context, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		nullOrInt64(entry.UserID),
		entry.Username,
		entry.EventType,
		entry.IPAddress,
		entry.Metadata,
		entry.Context,
		created.Format(sqliteTimeFormat),
	)
	return err
}

func parseSQLiteTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}

	layouts := []string{
		sqliteTimeFormat,
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	var lastErr error
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, lastErr
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func nullOrString(value sql.NullString) interface{} {
	if value.Valid {
		return value.String
	}
	return nil
}

func nullOrInt64(value sql.NullInt64) interface{} {
	if value.Valid {
		return value.Int64
	}
	return nil
}

// BBSSessionRecord DAL functions

// CreateBBSSession creates a new BBS session record.
func (s *SQLiteDB) CreateBBSSession(session *BBSSessionRecord) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO bbs_sessions (user_id, node_number, session_start, last_activity, time_left, calls_today, status, ip_address, connection_type, current_area, current_menu)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.UserID,
		session.NodeNumber,
		session.SessionStart,
		session.LastActivity,
		session.TimeLeft,
		session.CallsToday,
		session.Status,
		nullOrString(session.IPAddress),
		nullOrString(session.ConnectionType),
		nullOrString(session.CurrentArea),
		nullOrString(session.CurrentMenu),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create BBS session: %w", err)
	}
	return result.LastInsertId()
}

// GetBBSSessionByUser retrieves the active BBS session for a user.
func (s *SQLiteDB) GetBBSSessionByUser(userID int64) (*BBSSessionRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, user_id, node_number, session_start, last_activity, time_left, calls_today, status, ip_address, connection_type, current_area, current_menu
		FROM bbs_sessions
		WHERE user_id = ? AND status = 'active'`,
		userID,
	)
	var session BBSSessionRecord
	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.NodeNumber,
		&session.SessionStart,
		&session.LastActivity,
		&session.TimeLeft,
		&session.CallsToday,
		&session.Status,
		&session.IPAddress,
		&session.ConnectionType,
		&session.CurrentArea,
		&session.CurrentMenu,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get BBS session: %w", err)
	}
	return &session, nil
}

// UpdateBBSSession updates a BBS session record.
func (s *SQLiteDB) UpdateBBSSession(session *BBSSessionRecord) error {
	_, err := s.db.Exec(`
		UPDATE bbs_sessions
		SET last_activity = ?, time_left = ?, calls_today = ?, status = ?, current_area = ?, current_menu = ?
		WHERE id = ?`,
		session.LastActivity,
		session.TimeLeft,
		session.CallsToday,
		session.Status,
		nullOrString(session.CurrentArea),
		nullOrString(session.CurrentMenu),
		session.ID,
	)
	return err
}

// DeleteBBSSession deletes a BBS session record.
func (s *SQLiteDB) DeleteBBSSession(sessionID int64) error {
	_, err := s.db.Exec(`DELETE FROM bbs_sessions WHERE id = ?`, sessionID)
	return err
}

// BBSConfigRecord DAL functions

// GetBBSConfig retrieves a BBS config value.
func (s *SQLiteDB) GetBBSConfig(section, subsection, key string) (*BBSConfigRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, section, subsection, key, value, value_type, description, modified_by, modified_at
		FROM bbs_config
		WHERE section = ? AND subsection = ? AND key = ?`,
		section, nullOrString(sql.NullString{String: subsection, Valid: subsection != ""}), key,
	)
	var config BBSConfigRecord
	if err := row.Scan(
		&config.ID,
		&config.Section,
		&config.Subsection,
		&config.Key,
		&config.Value,
		&config.ValueType,
		&config.Description,
		&config.ModifiedBy,
		&config.ModifiedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get BBS config: %w", err)
	}
	return &config, nil
}

// SetBBSConfig sets or updates a BBS config value.
func (s *SQLiteDB) SetBBSConfig(config *BBSConfigRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO bbs_config (section, subsection, key, value, value_type, description, modified_by, modified_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(section, subsection, key) DO UPDATE SET
			value = excluded.value,
			value_type = excluded.value_type,
			description = excluded.description,
			modified_by = excluded.modified_by,
			modified_at = excluded.modified_at`,
		config.Section,
		nullOrString(config.Subsection),
		config.Key,
		config.Value,
		config.ValueType,
		nullOrString(config.Description),
		nullOrString(config.ModifiedBy),
		config.ModifiedAt,
	)
	return err
}

// GetAllBBSConfig retrieves all BBS config records.
func (s *SQLiteDB) GetAllBBSConfig() ([]BBSConfigRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, section, subsection, key, value, value_type, description, modified_by, modified_at
		FROM bbs_config
		ORDER BY section, subsection, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var configs []BBSConfigRecord
	for rows.Next() {
		var config BBSConfigRecord
		if err := rows.Scan(
			&config.ID,
			&config.Section,
			&config.Subsection,
			&config.Key,
			&config.Value,
			&config.ValueType,
			&config.Description,
			&config.ModifiedBy,
			&config.ModifiedAt,
		); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, rows.Err()
}

// UserPreferenceRecord DAL functions

// GetUserPreference retrieves a user preference.
func (s *SQLiteDB) GetUserPreference(userID int64, key string) (*UserPreferenceRecord, error) {
	row := s.db.QueryRow(`
		SELECT user_id, preference_key, preference_value, category
		FROM user_preferences
		WHERE user_id = ? AND preference_key = ?`,
		userID, key,
	)
	var pref UserPreferenceRecord
	if err := row.Scan(
		&pref.UserID,
		&pref.PreferenceKey,
		&pref.PreferenceValue,
		&pref.Category,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user preference: %w", err)
	}
	return &pref, nil
}

// SetUserPreference sets or updates a user preference.
func (s *SQLiteDB) SetUserPreference(pref *UserPreferenceRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO user_preferences (user_id, preference_key, preference_value, category)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, preference_key) DO UPDATE SET
			preference_value = excluded.preference_value,
			category = excluded.category`,
		pref.UserID,
		pref.PreferenceKey,
		pref.PreferenceValue,
		nullOrString(pref.Category),
	)
	return err
}

// GetUserPreferences retrieves all preferences for a user.
func (s *SQLiteDB) GetUserPreferences(userID int64) (map[string]UserPreferenceRecord, error) {
	rows, err := s.db.Query(`
		SELECT user_id, preference_key, preference_value, category
		FROM user_preferences
		WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	prefs := make(map[string]UserPreferenceRecord)
	for rows.Next() {
		var pref UserPreferenceRecord
		if err := rows.Scan(
			&pref.UserID,
			&pref.PreferenceKey,
			&pref.PreferenceValue,
			&pref.Category,
		); err != nil {
			return nil, err
		}
		prefs[pref.PreferenceKey] = pref
	}
	return prefs, rows.Err()
}

// DeleteUserPreference deletes a user preference.
func (s *SQLiteDB) DeleteUserPreference(userID int64, key string) error {
	_, err := s.db.Exec(`DELETE FROM user_preferences WHERE user_id = ? AND preference_key = ?`, userID, key)
	return err
}

// SecurityLevelRecord DAL functions

// CreateSecurityLevel creates a new security level record.
func (s *SQLiteDB) CreateSecurityLevel(level *SecurityLevelRecord) (int64, error) {
	now := time.Now().Format(sqliteTimeFormat)
	result, err := s.db.Exec(`
		INSERT INTO security_levels (name, sec_level, mins_per_day, timeout_mins, can_delete_own_msgs, can_delete_msgs, invisible, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		level.Name,
		level.SecLevel,
		level.MinsPerDay,
		level.TimeoutMins,
		boolToInt(level.CanDeleteOwnMsgs),
		boolToInt(level.CanDeleteMsgs),
		boolToInt(level.Invisible),
		now,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create security level: %w", err)
	}
	return result.LastInsertId()
}

// GetSecurityLevelByID retrieves a security level by ID.
func (s *SQLiteDB) GetSecurityLevelByID(id int64) (*SecurityLevelRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, name, sec_level, mins_per_day, timeout_mins, can_delete_own_msgs, can_delete_msgs, invisible, created_at, updated_at
		FROM security_levels
		WHERE id = ?`,
		id,
	)
	var level SecurityLevelRecord
	var createdStr, updatedStr string
	var canDeleteOwnMsgs, canDeleteMsgs, invisible int

	if err := row.Scan(
		&level.ID,
		&level.Name,
		&level.SecLevel,
		&level.MinsPerDay,
		&level.TimeoutMins,
		&canDeleteOwnMsgs,
		&canDeleteMsgs,
		&invisible,
		&createdStr,
		&updatedStr,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get security level: %w", err)
	}

	level.CanDeleteOwnMsgs = canDeleteOwnMsgs != 0
	level.CanDeleteMsgs = canDeleteMsgs != 0
	level.Invisible = invisible != 0

	if created, err := parseSQLiteTime(createdStr); err == nil {
		level.CreatedAt = created
	}
	if updated, err := parseSQLiteTime(updatedStr); err == nil {
		level.UpdatedAt = updated
	}

	return &level, nil
}

// GetSecurityLevelByLevel retrieves a security level by security level number.
func (s *SQLiteDB) GetSecurityLevelByLevel(secLevel int) (*SecurityLevelRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, name, sec_level, mins_per_day, timeout_mins, can_delete_own_msgs, can_delete_msgs, invisible, created_at, updated_at
		FROM security_levels
		WHERE sec_level = ?`,
		secLevel,
	)
	var level SecurityLevelRecord
	var createdStr, updatedStr string
	var canDeleteOwnMsgs, canDeleteMsgs, invisible int

	if err := row.Scan(
		&level.ID,
		&level.Name,
		&level.SecLevel,
		&level.MinsPerDay,
		&level.TimeoutMins,
		&canDeleteOwnMsgs,
		&canDeleteMsgs,
		&invisible,
		&createdStr,
		&updatedStr,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get security level: %w", err)
	}

	level.CanDeleteOwnMsgs = canDeleteOwnMsgs != 0
	level.CanDeleteMsgs = canDeleteMsgs != 0
	level.Invisible = invisible != 0

	if created, err := parseSQLiteTime(createdStr); err == nil {
		level.CreatedAt = created
	}
	if updated, err := parseSQLiteTime(updatedStr); err == nil {
		level.UpdatedAt = updated
	}

	return &level, nil
}

// GetAllSecurityLevels retrieves all security levels ordered by sec_level.
func (s *SQLiteDB) GetAllSecurityLevels() ([]SecurityLevelRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, name, sec_level, mins_per_day, timeout_mins, can_delete_own_msgs, can_delete_msgs, invisible, created_at, updated_at
		FROM security_levels
		ORDER BY sec_level`)
	if err != nil {
		return nil, fmt.Errorf("failed to query security levels: %w", err)
	}
	defer rows.Close()

	var levels []SecurityLevelRecord
	for rows.Next() {
		var level SecurityLevelRecord
		var createdStr, updatedStr string
		var canDeleteOwnMsgs, canDeleteMsgs, invisible int

		if err := rows.Scan(
			&level.ID,
			&level.Name,
			&level.SecLevel,
			&level.MinsPerDay,
			&level.TimeoutMins,
			&canDeleteOwnMsgs,
			&canDeleteMsgs,
			&invisible,
			&createdStr,
			&updatedStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan security level: %w", err)
		}

		level.CanDeleteOwnMsgs = canDeleteOwnMsgs != 0
		level.CanDeleteMsgs = canDeleteMsgs != 0
		level.Invisible = invisible != 0

		if created, err := parseSQLiteTime(createdStr); err == nil {
			level.CreatedAt = created
		}
		if updated, err := parseSQLiteTime(updatedStr); err == nil {
			level.UpdatedAt = updated
		}

		levels = append(levels, level)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating security levels: %w", err)
	}

	return levels, nil
}

// UpdateSecurityLevel updates an existing security level.
func (s *SQLiteDB) UpdateSecurityLevel(level *SecurityLevelRecord) error {
	now := time.Now().Format(sqliteTimeFormat)
	_, err := s.db.Exec(`
		UPDATE security_levels
		SET name = ?, mins_per_day = ?, timeout_mins = ?, can_delete_own_msgs = ?, can_delete_msgs = ?, invisible = ?, updated_at = ?
		WHERE id = ?`,
		level.Name,
		level.MinsPerDay,
		level.TimeoutMins,
		boolToInt(level.CanDeleteOwnMsgs),
		boolToInt(level.CanDeleteMsgs),
		boolToInt(level.Invisible),
		now,
		level.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update security level: %w", err)
	}
	return nil
}

// DeleteSecurityLevel deletes a security level by ID.
func (s *SQLiteDB) DeleteSecurityLevel(id int64) error {
	_, err := s.db.Exec(`DELETE FROM security_levels WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete security level: %w", err)
	}
	return nil
}
