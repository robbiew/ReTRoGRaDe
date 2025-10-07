# Decision Log

[2025-10-07 11:57:00] - Added database connection to TUI editor Model for user management functionality, enabling direct database queries from the configuration interface.
[2025-10-07 11:40:27] - Removed unused global variable 'db' from internal/auth/auth.go: All database operations use getStorage() abstraction instead, eliminating unnecessary global state.
