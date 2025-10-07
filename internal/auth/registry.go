package auth

import (
	"fmt"
	"sync"

	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/database"
)

var (
	storage Storage
	mu      sync.RWMutex
)

// Init initializes the auth storage system
func Init(cfg *config.AuthConfig, db database.Database) error {
	mu.Lock()
	defer mu.Unlock()

	if db == nil {
		return fmt.Errorf("auth: database handle is required for SQLite storage")
	}

	storage = &sqliteStorage{
		db:                db,
		maxFailedAttempts: cfg.MaxFailedAttempts,
		lockMinutes:       cfg.AccountLockMinutes,
	}

	return nil
}

// getStorage returns the current storage implementation
func getStorage() Storage {
	mu.RLock()
	defer mu.RUnlock()

	if storage == nil {
		panic("auth: storage not initialized - call Init() first")
	}

	return storage
}
