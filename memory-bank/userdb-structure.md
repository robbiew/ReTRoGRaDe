# User Database Structure

This document outlines the comprehensive user database schema for the Retrograde BBS system. It covers the implemented database tables, their relationships, and recommended extensions for traditional BBS functionality.

The schema is designed to support modern authentication features while maintaining compatibility with classic BBS operations including user sessions, message base subscriptions, and system configuration.

## Database Tables

### Core User Tables

#### 1. users
Primary user account table containing authentication and profile information.

**Columns:**
- `id` (INTEGER PRIMARY KEY) - Unique user identifier
- `username` (TEXT UNIQUE) - Login username (case-insensitive)
- `first_name` (TEXT) - User's first name
- `last_name` (TEXT) - User's last name
- `password_hash` (TEXT) - Hashed password
- `password_salt` (TEXT) - Password salt for hashing
- `password_algo` (TEXT) - Password hashing algorithm used
- `password_updated_at` (TEXT) - Timestamp of last password change
- `failed_attempts` (INTEGER) - Count of consecutive failed login attempts
- `locked_until` (TEXT) - Timestamp until account is locked (null if not locked)
- `security_level` (INTEGER) - User's security/access level
- `created_date` (TEXT) - Account creation timestamp
- `last_login` (TEXT) - Last successful login timestamp
- `email` (TEXT UNIQUE) - User's email address
- `country` (TEXT) - User's country
- `locations` (TEXT) - User's city and state/province or other

#### 2. user_details
Key-value storage for extensible user attributes and preferences.

**Columns:**
- `user_id` (INTEGER) - Foreign key to users.id
- `attrib` (TEXT) - Attribute name/key
- `value` (TEXT) - Attribute value

**Constraints:**
- PRIMARY KEY (user_id, attrib)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

#### 3. user_subscriptions
Tracks message base subscriptions for new message scanning.

**Columns:**
- `user_id` (INTEGER) - Foreign key to users.id
- `msgbase` (TEXT) - Message base identifier

**Constraints:**
- PRIMARY KEY (user_id, msgbase)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

#### 4. user_lastread
Stores last read message positions for continuing reading sessions.

**Columns:**
- `user_id` (INTEGER) - Foreign key to users.id
- `msgbase` (TEXT) - Message base identifier
- `last_message_id` (INTEGER) - ID of last read message (0 if none read)

**Constraints:**
- PRIMARY KEY (user_id, msgbase)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

### Security Tables

#### 5. auth_audit
Comprehensive audit trail for authentication events.

**Columns:**
- `id` (INTEGER PRIMARY KEY) - Unique audit entry identifier
- `user_id` (INTEGER) - Foreign key to users.id (null for failed anonymous attempts)
- `username` (TEXT) - Username associated with the event
- `event_type` (TEXT) - Type of authentication event
- `ip_address` (TEXT) - Client IP address
- `metadata` (TEXT) - Additional event metadata (JSON)
- `context` (TEXT) - Additional context information
- `created_at` (TEXT) - Event timestamp

**Constraints:**
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL

#### 6. user_mfa
Multi-factor authentication settings per user.

**Columns:**
- `id` (INTEGER PRIMARY KEY) - Unique MFA record identifier
- `user_id` (INTEGER) - Foreign key to users.id
- `method` (TEXT) - MFA method (e.g., 'totp', 'sms', 'email')
- `secret` (TEXT) - MFA secret key
- `config` (TEXT) - Method-specific configuration (JSON)
- `is_enabled` (INTEGER) - Whether MFA method is active (1/0)
- `created_at` (TEXT) - MFA method creation timestamp
- `updated_at` (TEXT) - Last update timestamp
- `last_used_at` (TEXT) - Last successful MFA use timestamp
- `backup_codes` (TEXT) - Encrypted backup recovery codes
- `display_name` (TEXT) - User-friendly method name

**Constraints:**
- UNIQUE (user_id, method)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

### Recommended BBS Extension Tables

#### 7. bbs_sessions (Recommended)
Session state management for active BBS connections.

**Columns:**
- `id` (INTEGER PRIMARY KEY) - Unique session identifier
- `user_id` (INTEGER) - Foreign key to users.id
- `node_number` (INTEGER) - BBS node number
- `session_start` (TEXT) - Session start timestamp
- `last_activity` (TEXT) - Last user activity timestamp
- `time_left` (INTEGER) - Minutes remaining in session
- `calls_today` (INTEGER) - Number of calls made today
- `status` (TEXT) - Session status ('active', 'idle', 'timed_out')
- `ip_address` (TEXT) - Client IP address
- `connection_type` (TEXT) - Connection protocol ('telnet', 'ssh', etc.)
- `current_area` (TEXT) - Current message/file area
- `current_menu` (TEXT) - Current menu context

**Constraints:**
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

#### 8. bbs_config (Recommended)
System-wide BBS configuration settings.

**Columns:**
- `id` (INTEGER PRIMARY KEY) - Unique config entry identifier
- `section` (TEXT) - Configuration section
- `subsection` (TEXT) - Configuration subsection
- `key` (TEXT) - Configuration key
- `value` (TEXT) - Configuration value
- `value_type` (TEXT) - Data type ('string', 'int', 'bool', 'list')
- `description` (TEXT) - Human-readable description
- `modified_by` (TEXT) - User/system that last modified
- `modified_at` (TEXT) - Last modification timestamp

**Constraints:**
- UNIQUE (section, subsection, key)

#### 9. user_preferences (Recommended)
User-specific BBS preferences and settings.

**Columns:**
- `user_id` (INTEGER) - Foreign key to users.id
- `preference_key` (TEXT) - Preference identifier
- `preference_value` (TEXT) - Preference value
- `category` (TEXT) - Preference category ('display', 'behavior', 'notifications')

**Constraints:**
- PRIMARY KEY (user_id, preference_key)
- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

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
- `idx_user_mfa_user` on user_mfa(user_id)

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
- Multi-factor authentication support

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

   
