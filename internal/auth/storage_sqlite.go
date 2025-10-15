package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/robbiew/retrograde/internal/database"
)

const (
	sqliteTimeLayout = time.RFC3339Nano
)

var timeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	userTimestampLayout,
	"2006-01-02 15:04:05",
}

// NewSQLiteStorage returns a Storage implementation backed by the shared database layer.
func NewSQLiteStorage(db database.Database, maxFailedAttempts, lockMinutes int) Storage {
	if db == nil {
		panic("auth: storage database handle is nil")
	}
	return &sqliteStorage{
		db:                db,
		maxFailedAttempts: maxFailedAttempts,
		lockMinutes:       lockMinutes,
	}
}

type sqliteStorage struct {
	db                database.Database
	maxFailedAttempts int
	lockMinutes       int
}

func (s *sqliteStorage) CreateUser(params CreateUserParams) (*UserRecord, error) {
	username := strings.TrimSpace(params.Username)
	if username == "" {
		return nil, fmt.Errorf("auth: username is required")
	}

	now := time.Now().UTC()
	createdAt := params.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}

	digest := params.Password
	if digest.UpdatedAt.IsZero() {
		digest.UpdatedAt = createdAt
	}

	// User creation logging removed for production

	dbUser := database.UserRecord{
		Username:          username,
		PasswordHash:      digest.Hash,
		PasswordSalt:      database.NullString(digest.Salt),
		PasswordAlgo:      database.NullString(digest.Algorithm),
		PasswordUpdatedAt: database.NullString(formatTimestamp(digest.UpdatedAt)),
		SecurityLevel:     params.SecurityLevel,
		CreatedDate:       formatTimestamp(createdAt),
		LastLogin:         sql.NullString{},
		Email:             database.NullString(strings.TrimSpace(params.Email)),
		FailedAttempts:    0,
		LockedUntil:       sql.NullString{},
		FirstName:         database.NullString(strings.TrimSpace(params.FirstName)),
		LastName:          database.NullString(strings.TrimSpace(params.LastName)),
		Locations:         database.NullString(strings.TrimSpace(params.Location)),
	}

	id, err := s.db.CreateUser(&dbUser)
	if err != nil {
		// Check for unique constraint violations
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "unique") {
			if strings.Contains(errMsg, "username") {
				return nil, ErrUserExists
			} else if strings.Contains(errMsg, "email") {
				return nil, ErrEmailExists
			}
			// Fallback to generic user exists error
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("auth: create user failed: %w", err)
	}
	dbUser.ID = id

	user, err := toDomainUserRecord(&dbUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *sqliteStorage) GetUserByUsername(username string) (*UserRecord, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("auth: username is required")
	}

	dbUser, err := s.db.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("auth: lookup user failed: %w", err)
	}
	if dbUser == nil {
		return nil, ErrUserNotFound
	}

	user, err := toDomainUserRecord(dbUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *sqliteStorage) GetUserByEmail(email string) (*UserRecord, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, fmt.Errorf("auth: email is required")
	}

	dbUser, err := s.db.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("auth: lookup user failed: %w", err)
	}
	if dbUser == nil {
		return nil, ErrUserNotFound
	}

	user, err := toDomainUserRecord(dbUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *sqliteStorage) UpdateUserMetadata(user *UserRecord) error {
	if user == nil {
		return fmt.Errorf("auth: user record is nil")
	}
	dbRec, err := toDatabaseUserRecord(user)
	if err != nil {
		return err
	}

	if err := s.db.UpdateUser(dbRec); err != nil {
		return fmt.Errorf("auth: update user failed: %w", err)
	}
	return nil
}

func (s *sqliteStorage) DeleteUser(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("auth: invalid user id")
	}

	statements := []string{
		`DELETE FROM user_details WHERE user_id = ?`,
		`DELETE FROM user_applications WHERE user_id = ?`,
		`DELETE FROM user_subscriptions WHERE user_id = ?`,
		`DELETE FROM user_lastread WHERE user_id = ?`,
		`DELETE FROM auth_audit WHERE user_id = ?`,
		`DELETE FROM users WHERE id = ?`,
	}

	err := s.db.WithTransaction(func(tx *sql.Tx) error {
		for _, stmt := range statements {
			if _, execErr := tx.Exec(stmt, userID); execErr != nil {
				return fmt.Errorf("auth: delete user statement failed: %w", execErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStorage) UpdateLastLogin(userID int64, when time.Time) error {
	if userID <= 0 {
		return fmt.Errorf("auth: invalid user id")
	}
	if when.IsZero() {
		when = time.Now().UTC()
	}

	err := s.db.WithTransaction(func(tx *sql.Tx) error {
		result, execErr := tx.Exec(`UPDATE users SET last_login = ? WHERE id = ?`, formatTimestamp(when), userID)
		if execErr != nil {
			return fmt.Errorf("auth: update last_login failed: %w", execErr)
		}
		if rows, _ := result.RowsAffected(); rows == 0 {
			return ErrUserNotFound
		}
		return nil
	})
	if errors.Is(err, ErrUserNotFound) {
		return err
	}
	return err
}

func (s *sqliteStorage) IncrementFailedLogin(userID int64, when time.Time, lockThreshold int) (int, *time.Time, error) {
	failed, lockedUntil, err := s.db.IncrementFailedAttempts(userID, when.UTC(), s.maxFailedAttempts, s.lockMinutes)
	if err != nil {
		return 0, nil, fmt.Errorf("auth: increment failed attempts: %w", err)
	}
	return failed, lockedUntil, nil
}

func (s *sqliteStorage) ResetFailedLogin(userID int64, when time.Time) error {
	if err := s.db.ResetFailedAttempts(userID, when.UTC()); err != nil {
		return fmt.Errorf("auth: reset failed attempts: %w", err)
	}
	return nil
}

func (s *sqliteStorage) UpdatePassword(userID int64, digest PasswordDigest) error {
	if userID <= 0 {
		return fmt.Errorf("auth: invalid user id")
	}
	if digest.UpdatedAt.IsZero() {
		digest.UpdatedAt = time.Now().UTC()
	}
	if err := s.db.UpdatePassword(userID, digest.Hash, digest.Algorithm, digest.Salt, digest.UpdatedAt.UTC()); err != nil {
		return fmt.Errorf("auth: update password failed: %w", err)
	}
	return nil
}

func (s *sqliteStorage) UpsertUserDetail(userID int64, attrib, value string) error {
	if userID <= 0 {
		return fmt.Errorf("auth: invalid user id")
	}
	attrib = strings.TrimSpace(attrib)
	if attrib == "" {
		return fmt.Errorf("auth: attribute name is required")
	}

	if err := s.db.UpsertUserDetail(userID, attrib, strings.TrimSpace(value)); err != nil {
		return fmt.Errorf("auth: upsert user detail failed: %w", err)
	}
	return nil
}

func (s *sqliteStorage) DeleteApplication(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("auth: invalid user id")
	}

	// Get the current user record
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("auth: failed to get user for application delete: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Clear the application fields (none in current schema)
	// Application fields have been removed from the user table

	// Update the user record
	if err := s.db.UpdateUser(user); err != nil {
		return fmt.Errorf("auth: failed to update user with cleared application data: %w", err)
	}
	return nil
}

func toDomainUserRecord(dbUser *database.UserRecord) (*UserRecord, error) {
	if dbUser == nil {
		return nil, fmt.Errorf("auth: database user is nil")
	}

	createdAt, _ := parseTimestampString(dbUser.CreatedDate)
	var lastLogin *time.Time
	if dbUser.LastLogin.Valid {
		if ts, ok := parseTimestampString(dbUser.LastLogin.String); ok {
			lastLogin = &ts
		}
	}
	var lockedUntil *time.Time
	if dbUser.LockedUntil.Valid {
		if ts, ok := parseTimestampString(dbUser.LockedUntil.String); ok {
			lockedUntil = &ts
		}
	}
	var passwordUpdatedAt time.Time
	if dbUser.PasswordUpdatedAt.Valid {
		if ts, ok := parseTimestampString(dbUser.PasswordUpdatedAt.String); ok {
			passwordUpdatedAt = ts
		}
	}

	user := &UserRecord{
		ID:             dbUser.ID,
		Username:       dbUser.Username,
		SecurityLevel:  dbUser.SecurityLevel,
		CreatedAt:      createdAt,
		LastLogin:      lastLogin,
		Email:          dbUser.Email.String,
		FailedAttempts: dbUser.FailedAttempts,
		LockedUntil:    lockedUntil,
		Password: PasswordDigest{
			Hash:      dbUser.PasswordHash,
			Algorithm: dbUser.PasswordAlgo.String,
			Salt:      dbUser.PasswordSalt.String,
			UpdatedAt: passwordUpdatedAt,
		},
	}
	return user, nil
}

func toDatabaseUserRecord(user *UserRecord) (*database.UserRecord, error) {
	if user == nil {
		return nil, fmt.Errorf("auth: user is nil")
	}
	if user.Username == "" {
		return nil, fmt.Errorf("auth: username is required")
	}

	createdAt := user.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	var lastLogin sql.NullString
	if user.LastLogin != nil {
		lastLogin = sql.NullString{String: formatTimestamp(user.LastLogin.UTC()), Valid: true}
	}

	var lockedUntil sql.NullString
	if user.LockedUntil != nil {
		lockedUntil = sql.NullString{String: formatTimestamp(user.LockedUntil.UTC()), Valid: true}
	}

	passwordUpdated := user.Password.UpdatedAt
	if passwordUpdated.IsZero() {
		passwordUpdated = time.Now().UTC()
	}

	dbUser := &database.UserRecord{
		ID:                user.ID,
		Username:          user.Username,
		PasswordHash:      user.Password.Hash,
		PasswordSalt:      database.NullString(user.Password.Salt),
		PasswordAlgo:      database.NullString(user.Password.Algorithm),
		PasswordUpdatedAt: database.NullString(formatTimestamp(passwordUpdated)),
		SecurityLevel:     user.SecurityLevel,
		CreatedDate:       formatTimestamp(createdAt),
		LastLogin:         lastLogin,
		Email:             database.NullString(strings.TrimSpace(user.Email)),
		FailedAttempts:    user.FailedAttempts,
		LockedUntil:       lockedUntil,
	}
	return dbUser, nil
}

func formatTimestamp(t time.Time) string {
	return t.UTC().Format(sqliteTimeLayout)
}

func parseTimestampString(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range timeLayouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts.UTC(), true
		}
	}
	return time.Time{}, false
}

// ErrNotFound returns a typed error for missing entities.
func ErrNotFound(entity string) error {
	return fmt.Errorf("auth: %s not found", entity)
}
