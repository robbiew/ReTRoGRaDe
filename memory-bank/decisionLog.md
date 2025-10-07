# Decision Log

[2025-10-07 11:40:27] - Removed unused global variable 'db' from internal/auth/auth.go: All database operations use getStorage() abstraction instead, eliminating unnecessary global state.