# User Database Structure - Actual Implementation

This document describes the actual SQLite database schema implemented in Retrograde BBS. The schema supports modern authentication, user management, and BBS-specific functionality.

## Database Overview

- **Engine**: SQLite with WAL mode enabled
- **Location**: `data/retrograde.db`
- **Schema Version**: Managed via `schema_version` table
- **Initialization**: Automatic schema creation on first run

## Implemented Tables

### 1. schema_version

Tracks database schema version for migrations.

**Columns:**

- `id` (INTEGER PRIMARY KEY CHECK (id = 1)) - Singleton constraint
- `version` (INTEGER NOT NULL) - Current schema version
- `applied_at` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP) - When version was applied
- `description` (TEXT) - Version description

### 2. config_settings

SQLite-backed configuration storage (alternative to INI files).

**Columns:**

- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `section` (TEXT NOT NULL)
- `subsection` (TEXT) - NULL for top-level sections
- `key` (TEXT NOT NULL)
- `value` (TEXT NOT NULL)
- `value_type` (TEXT NOT NULL) - 'string', 'int', 'bool', 'list'
- `default_value` (TEXT)
- `description` (TEXT)
- `created_at` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP)
- `modified_at` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP)
- `modified_by` (TEXT DEFAULT 'system')

**Constraints:**

- UNIQUE(section, subsection, key)
- INDEX idx_config_section(section)
- INDEX idx_config_subsection(section, subsection)
- INDEX idx_config_key(section, subsection, key)

### 3. users

Primary user account table with authentication and profile data.

**Columns:**

- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `username` (TEXT NOT NULL UNIQUE COLLATE NOCASE)
- `first_name` (TEXT)
- `last_name` (TEXT)
- `password_hash` (TEXT NOT NULL)
- `password_salt` (TEXT)
- `password_algo` (TEXT)
- `password_updated_at` (TEXT)
- `failed_attempts` (INTEGER NOT NULL DEFAULT 0)
- `locked_until` (TEXT)
- `security_level` (INTEGER NOT NULL DEFAULT 0)
- `created_date` (TEXT NOT NULL)
- `last_login` (TEXT)
- `email` (TEXT UNIQUE)
- `locations` (TEXT)

**Indexes:**

- `idx_users_username` on users(username)
- `idx_users_email` on users(email)

### 4. bbs_sessions

Active BBS session tracking for connected users.

**Columns:**

- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `user_id` (INTEGER NOT NULL)
- `node_number` (INTEGER NOT NULL)
- `session_start` (TEXT NOT NULL)
- `last_activity` (TEXT NOT NULL)
- `time_left` (INTEGER NOT NULL DEFAULT 0)
- `calls_today` (INTEGER NOT NULL DEFAULT 0)
- `status` (TEXT NOT NULL DEFAULT 'active')
- `ip_address` (TEXT)
- `connection_type` (TEXT)
- `current_area` (TEXT)
- `current_menu` (TEXT)

**Constraints:**

- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

**Indexes:**

- `idx_bbs_sessions_user` on bbs_sessions(user_id)
- `idx_bbs_sessions_node` on bbs_sessions(node_number)
- `idx_bbs_sessions_status` on bbs_sessions(status)

### 5. bbs_config

BBS-specific configuration settings.

**Columns:**

- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `section` (TEXT NOT NULL)
- `subsection` (TEXT)
- `key` (TEXT NOT NULL)
- `value` (TEXT NOT NULL)
- `value_type` (TEXT NOT NULL DEFAULT 'string')
- `description` (TEXT)
- `modified_by` (TEXT)
- `modified_at` (TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)

**Constraints:**

- UNIQUE (section, subsection, key)

### 6. user_preferences

User-specific preferences and settings.

**Columns:**

- `user_id` (INTEGER NOT NULL)
- `preference_key` (TEXT NOT NULL)
- `preference_value` (TEXT NOT NULL)
- `category` (TEXT)

**Constraints:**

- PRIMARY KEY (user_id, preference_key)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

**Indexes:**

- `idx_user_preferences_category` on user_preferences(category)

### 7. user_details

Extensible key-value storage for additional user attributes.

**Columns:**

- `user_id` (INTEGER NOT NULL)
- `attrib` (TEXT NOT NULL COLLATE NOCASE)
- `value` (TEXT)

**Constraints:**

- PRIMARY KEY (user_id, attrib)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

**Indexes:**

- `idx_user_details_user` on user_details(user_id)

### 8. user_subscriptions

Message base subscriptions for new message notifications.

**Columns:**

- `user_id` (INTEGER NOT NULL)
- `msgbase` (TEXT NOT NULL)

**Constraints:**

- PRIMARY KEY (user_id, msgbase)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

**Indexes:**

- `idx_user_subscriptions_user` on user_subscriptions(user_id)

### 9. user_lastread

Last read message positions for continuing reading sessions.

**Columns:**

- `user_id` (INTEGER NOT NULL)
- `msgbase` (TEXT NOT NULL)
- `last_message_id` (INTEGER)

**Constraints:**

- PRIMARY KEY (user_id, msgbase)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

**Indexes:**

- `idx_user_lastread_user` on user_lastread(user_id)

### 10. auth_audit

Comprehensive authentication audit trail.

**Columns:**

- `id` (INTEGER PRIMARY KEY AUTOINCREMENT)
- `user_id` (INTEGER)
- `username` (TEXT)
- `event_type` (TEXT NOT NULL)
- `ip_address` (TEXT)
- `metadata` (TEXT)
- `context` (TEXT)
- `created_at` (TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)

**Constraints:**

- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL

**Indexes:**

- `idx_auth_audit_user` on auth_audit(user_id)
- `idx_auth_audit_created` on auth_audit(created_at)
- `idx_auth_audit_event` on auth_audit(event_type)

## Indexing Strategy

### Implemented Indexes

- `idx_users_username` on users(username)
- `idx_users_email` on users(email)
- `idx_user_details_user` on user_details(user_id)
- `idx_user_subscriptions_user` on user_subscriptions(user_id)
- `idx_user_lastread_user` on user_lastread(user_id)
- `idx_auth_audit_user` on auth_audit(user_id)
- `idx_auth_audit_created` on auth_audit(created_at)
- `idx_auth_audit_event` on auth_audit(event_type)

### Recommended Additional Indexes

- `idx_users_security_level` on users(security_level)
- `idx_users_last_login` on users(last_login)
- `idx_users_created_date` on users(created_date)
- `idx_bbs_sessions_user` on bbs_sessions(user_id)
- `idx_bbs_sessions_node` on bbs_sessions(node_number)
- `idx_bbs_sessions_status` on bbs_sessions(status)
- `idx_user_preferences_category` on user_preferences(category)

## Security Features

### Authentication Security

- Password hashing with configurable algorithms
- Account lockout after failed attempts
- Comprehensive audit logging

### Session Management

- Session timeout handling
- Activity tracking
- Node-based connection management
- IP-based security controls

### Data Protection

- Foreign key constraints with cascade deletes
- Input validation and sanitization
- Secure password storage practices
- Audit trail for security events
