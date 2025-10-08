package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func setupTestSQLiteDB(t *testing.T) *SQLiteDB {
	t.Helper()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, fmt.Sprintf("test_%s.db", t.Name()))

	db, err := OpenSQLite(ConnectionConfig{
		Path:    path,
		Timeout: 5,
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	if err := db.InitializeSchema(); err != nil {
		db.Close()
		t.Fatalf("failed to initialize schema: %v", err)
	}

	return db
}

func createTestUser(t *testing.T, db *SQLiteDB, username string) int64 {
	t.Helper()

	now := time.Now().UTC().Format(sqliteTimeFormat)
	user := &UserRecord{
		Username:          username,
		PasswordHash:      "hash",
		PasswordSalt:      NullString("salt"),
		PasswordAlgo:      NullString("sha256"),
		PasswordUpdatedAt: NullString(now),
		SecurityLevel:     10,
		CreatedDate:       now,
		LastLogin:         sql.NullString{},
		Email:             NullString(username + "@example.com"),
	}

	id, err := db.CreateUser(user)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return id
}

func fetchFailedAttempts(t *testing.T, db *SQLiteDB, userID int64) (int, sql.NullString) {
	t.Helper()

	var attempts int
	var locked sql.NullString
	if err := db.db.QueryRow(`SELECT failed_attempts, locked_until FROM users WHERE id = ?`, userID).Scan(&attempts, &locked); err != nil {
		t.Fatalf("failed to fetch failed attempts: %v", err)
	}
	return attempts, locked
}

func TestIncrementAndResetFailedAttempts(t *testing.T) {
	db := setupTestSQLiteDB(t)
	defer db.Close()

	userID := createTestUser(t, db, "alice")

	now := time.Now().UTC()
	for i := 1; i <= 2; i++ {
		count, lock, err := db.IncrementFailedAttempts(userID, now, 3, 30)
		if err != nil {
			t.Fatalf("IncrementFailedAttempts iteration %d returned error: %v", i, err)
		}
		if count != i {
			t.Fatalf("expected count %d, got %d", i, count)
		}
		if lock != nil {
			t.Fatalf("expected no lock before threshold, got %v", lock)
		}
	}

	count, lock, err := db.IncrementFailedAttempts(userID, now, 3, 30)
	if err != nil {
		t.Fatalf("IncrementFailedAttempts threshold returned error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3 after third failure, got %d", count)
	}
	if lock == nil {
		t.Fatalf("expected lock pointer after threshold reached")
	} else if lock.Sub(now) < 25*time.Minute || lock.Sub(now) > 35*time.Minute {
		t.Fatalf("expected lock approximately 30 minutes from now, got %v", lock.Sub(now))
	}

	attempts, lockedUntil := fetchFailedAttempts(t, db, userID)
	if attempts != 3 {
		t.Fatalf("expected stored attempts 3, got %d", attempts)
	}
	if !lockedUntil.Valid {
		t.Fatalf("expected locked_until stored value")
	}

	if err := db.ResetFailedAttempts(userID, now); err != nil {
		t.Fatalf("reset failed attempts returned error: %v", err)
	}

	attempts, lockedUntil = fetchFailedAttempts(t, db, userID)
	if attempts != 0 {
		t.Fatalf("expected attempts reset to 0, got %d", attempts)
	}
	if lockedUntil.Valid {
		t.Fatalf("expected locked_until cleared, got %v", lockedUntil.String)
	}
}

func TestUpdatePasswordResetsState(t *testing.T) {
	db := setupTestSQLiteDB(t)
	defer db.Close()

	userID := createTestUser(t, db, "bob")

	lockedTime := time.Now().UTC().Add(30 * time.Minute).Format(sqliteTimeFormat)
	if _, err := db.db.Exec(`UPDATE users SET failed_attempts = 5, locked_until = ? WHERE id = ?`, lockedTime, userID); err != nil {
		t.Fatalf("failed to seed failed attempts: %v", err)
	}

	newHash := "newhash"
	newAlgo := "bcrypt"
	newSalt := "pepper"
	now := time.Now().UTC()

	if err := db.UpdatePassword(userID, newHash, newAlgo, newSalt, now); err != nil {
		t.Fatalf("UpdatePassword returned error: %v", err)
	}

	user, err := db.GetUserByUsername("bob")
	if err != nil {
		t.Fatalf("GetUserByUsername returned error: %v", err)
	}
	if user == nil {
		t.Fatalf("expected user record")
	}
	if user.PasswordHash != newHash {
		t.Fatalf("expected password hash updated")
	}
	if user.PasswordAlgo.String != newAlgo {
		t.Fatalf("expected password algo %q, got %q", newAlgo, user.PasswordAlgo.String)
	}
	if user.PasswordSalt.String != newSalt {
		t.Fatalf("expected password salt %q, got %q", newSalt, user.PasswordSalt.String)
	}
	if !user.PasswordUpdatedAt.Valid {
		t.Fatalf("expected password updated timestamp to be set")
	} else {
		updated, err := time.Parse(sqliteTimeFormat, user.PasswordUpdatedAt.String)
		if err != nil {
			t.Fatalf("failed to parse password_updated_at: %v", err)
		}
		if updated.Before(now.Add(-time.Minute)) || updated.After(now.Add(time.Minute)) {
			t.Fatalf("expected password_updated_at close to now, got %v", updated)
		}
	}
	if user.FailedAttempts != 0 {
		t.Fatalf("expected failed attempts reset, got %d", user.FailedAttempts)
	}
	if user.LockedUntil.Valid {
		t.Fatalf("expected lock cleared, got %v", user.LockedUntil.String)
	}
}

func TestAuthAuditOperations(t *testing.T) {
	db := setupTestSQLiteDB(t)
	defer db.Close()

	userID := createTestUser(t, db, "carol")

	entry := &AuthAuditEntry{
		UserID:    sql.NullInt64{Int64: userID, Valid: true},
		Username:  NullString("carol"),
		EventType: "LOGIN_FAILED",
		IPAddress: NullString("127.0.0.1"),
		Metadata:  NullString(`{"reason":"bad password"}`),
		Context:   NullString("telnet"),
		CreatedAt: time.Now().UTC(),
	}
	if err := db.InsertAuthAudit(entry); err != nil {
		t.Fatalf("InsertAuthAudit returned error: %v", err)
	}

	var auditCount int
	if err := db.db.QueryRow(`SELECT COUNT(*) FROM auth_audit WHERE user_id = ?`, userID).Scan(&auditCount); err != nil {
		t.Fatalf("failed to count auth audit rows: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected 1 auth audit row, got %d", auditCount)
	}
}
