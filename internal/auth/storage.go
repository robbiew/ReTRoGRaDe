package auth

import (
	"errors"
	"time"
)

// Storage exposes the persistence contract for all authentication flows.

type Storage interface {
	// CreateUser persist a new account using the provided seed data.
	// Returns ErrUserExists when the username already exists.
	CreateUser(params CreateUserParams) (*UserRecord, error)

	// GetUserByEmail fetches a user by their canonical email address (case insensitive).
	GetUserByEmail(email string) (*UserRecord, error)
	// GetUserByUsername fetches a user by their canonical username (case insensitive).
	GetUserByUsername(username string) (*UserRecord, error)

	// UpdateUserMetadata writes mutable profile data (security level, email, timestamps, etc.).
	UpdateUserMetadata(user *UserRecord) error

	// DeleteUser removes a user and all dependent records.
	DeleteUser(userID int64) error

	// UpdateLastLogin records the user’s last successful login timestamp.
	UpdateLastLogin(userID int64, when time.Time) error

	// IncrementFailedLogin increases the failed-attempt counter and enforces lockouts.
	// Returns the updated failure count and any active lockout expiration.
	IncrementFailedLogin(userID int64, when time.Time, lockThreshold int) (failedAttempts int, lockedUntil *time.Time, err error)

	// ResetFailedLogin clears the failure counter and lockout metadata.
	ResetFailedLogin(userID int64, when time.Time) error

	// UpdatePassword replaces the stored password digest and resets failure tracking.
	UpdatePassword(userID int64, digest PasswordDigest) error

	// UpsertUserDetail stores or updates a user detail field.
	UpsertUserDetail(userID int64, attrib, value string) error

	// DeleteApplication removes the application for the user.
	DeleteApplication(userID int64) error
}

var (
	// ErrUserExists is returned when attempting to create a duplicate username.
	ErrUserExists = errors.New("auth: user already exists")

	// ErrEmailExists is returned when attempting to create a user with an email that already exists.
	ErrEmailExists = errors.New("auth: email already exists")

	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound = errors.New("auth: user not found")

	// ErrInvalidCredentials signals that provided credentials cannot be verified.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
)

// CreateUserParams captures all data required to create a new account record.
type CreateUserParams struct {
	Username      string
	Password      PasswordDigest
	SecurityLevel int
	Email         string
	CreatedAt     time.Time
	FirstName     string
	LastName      string
	Location      string
}

// PasswordDigest holds password hash metadata.
type PasswordDigest struct {
	Hash      string
	Algorithm string
	Salt      string
	UpdatedAt time.Time
}

// UserRecord represents the canonical view of an account as stored in the backend.
type UserRecord struct {
	ID             int64
	Username       string
	Password       PasswordDigest
	SecurityLevel  int
	CreatedAt      time.Time
	LastLogin      *time.Time
	Email          string
	FailedAttempts int
	LockedUntil    *time.Time
}
