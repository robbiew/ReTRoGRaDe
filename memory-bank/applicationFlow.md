# GHOSTnet Authentication & Connection Flow

## Overview

This document describes the complete authentication and connection flow for the Retrograde Server, including pre-authentication security checks, user login/registration, session management, and post-authentication procedures.

**Terminology Note:**

- **Login**: The process of authenticating a user (verifying credentials)
- **Logon**: Post-authentication initialization and welcome sequence (planned feature)

---

## Complete Connection Flow Diagram

```mermaid
graph TD
    A[TCP Connection Attempt] --> B{Security Check Layer}
    B -->|Blocked| C[Connection Rejected<br/>Generic Error Message]
    B -->|Allowed| D[Connection Accepted<br/>Clear Security Message]

    D --> E[Node Assignment]
    E --> F[Telnet Options Negotiation]
    F --> G[Session Created as Guest]
    G --> H[Main Menu Display]

    H --> I{User Action}

    I -->|Login L| J[Login Process]
    I -->|Register R| K[Registration Process]
    I -->|Other Options| H
    I -->|Quit Q| Z[Disconnect]

    J -->|Success| L[Update Session<br/>User Authenticated]
    J -->|Failure| M{Max Attempts?}
    M -->|Yes| N[Blacklist IP<br/>Force Disconnect]
    M -->|No| J
    J -->|Cancelled ESC| H

    K -->|Success| L
    K -->|Failure| H
    K -->|Cancelled ESC| H

    L --> O[Return to Main Menu<br/>Authenticated State]
    O --> P[Post-Login Features Available]

    P -.->|Future| Q[Logon Sequence*<br/>Welcome Screens]
    Q -.-> R[Main Menu]

    style B fill:#ff6b6b
    style C fill:#ff6b6b
    style N fill:#ff6b6b
    style Q fill:#95e1d3,stroke-dasharray: 5 5
```

\*Note: Logon sequence (welcome screens) is a planned future enhancement

---

## Phase 1: Pre-Authentication Security Layer

### Security Checks (Before Any Content Displayed)

The security layer validates connections **before** any BBS content is shown to the user.

**Implementation:** [`CheckConnectionSecurity()`](security.go:68) in [`main.go`](main.go:112)

#### 1.1 Whitelist Check (Optional)

- **When Enabled:** in TUI: `[security]` → `whitelist_enabled=true`
- **Action:** If IP is in whitelist, immediately allow connection
- **File:** Configured by `whitelist_file` setting
- **Logging:** `WHITELIST_ALLOW` event logged

#### 1.2 Blacklist Check

- **When Enabled:** Config TUI → `[security]` → `blacklist_enabled=true`
- **Sources:**
  - Manual entries in blacklist file
  - Auto-added from failed login attempts (3 strikes)
  - Temporary blocks from rate limiting
  - External threat intelligence feeds
- **File:** Configured by `blacklist_file` setting (default: `security/blacklist.txt`)
- **Action:** Reject connection with generic error message
- **Logging:** `BLACKLIST_BLOCK` event logged

#### 1.3 Rate Limiting Check

- **When Enabled:** Config TUI → `[security]` → `rate_limit_enabled=true`
- **Parameters:**
  - `max_connections_per_ip`: Maximum connections per IP in time window
  - `rate_limit_window_minutes`: Time window for counting attempts
- **Behavior:**
  - Tracks connection attempts per IP address
  - Exceeding limit triggers temporary block
  - Auto-adds IP to blacklist during block period
  - Block expires after rate limit window
- **Action:** Reject with reason "Rate limit exceeded"
- **Logging:** `RATE_LIMITED` event logged

#### 1.4 Geographic Blocking Check (Optional)

- **When Enabled:** Config TUI → `[security]` → `geo_block_enabled=true`
- **API Providers:**
  - `ipapi`: Free, no API key required (default)
  - `ipstack`: Paid, requires API key
  - `nerds`: Alternative provider
- **Configuration:**
  - `blocked_countries`: List of country codes to block
  - `allowed_countries`: Whitelist mode (only these allowed)
- **Caching:** Geo lookups cached for 24 hours
- **Private IPs:** Automatically allowed (no lookup)
- **Action:** Reject with "Connection from blocked/non-allowed country"
- **Logging:** `GEO_BLOCKED` event logged

#### 1.5 Threat Intelligence Check (Optional)

- **When Enabled:** Config TUI → `[security]` → `threat_intel_enabled=true`
- **Sources:** External blocklist URLs (configurable)
- **Update Frequency:** `blocklist_update_hours` setting
- **Action:** Reject with "IP found in threat intelligence feeds"
- **Logging:** `THREAT_BLOCKED` event logged

#### Security Layer Result

- **Allowed:** Connection proceeds to node assignment
- **Blocked:** Generic error message displayed, connection closed, no BBS access

**Security Principle:** Failed connections see generic messages only. Detailed rejection reasons logged server-side for security analysis.

---

## Phase 2: Connection Establishment

### 2.1 Node Assignment

**Implementation:** [`GetAvailableNode()`](main.go:127) in [`main.go`](main.go:127)

- Check for available nodes (max nodes configured in `config.ini`)
- If no nodes available: Display "all nodes in use" message, reject connection
- If node available: Assign node number, proceed to telnet negotiation
- display `[theme dir]\connect.ans` and prompt for login or 'new'

### 2.2 Telnet Options Negotiation

**Implementation:** [`negotiateTelnetOptions()`](main.go:491)

Configures telnet client for character-mode operation:

- Enable server echo (server echoes characters back)
- Suppress Go-Ahead (disable line buffering)
- Disable line mode (enable character-by-character input)
- 100ms delay to allow client processing

### 2.3 Session Initialization

**Implementation:** [`TelnetSession`](types.go:72) struct creation

Initial session state:

```go
{
    Alias:         "Guest",      // Unauthenticated state
    SecurityLevel: 0,            // No privileges
    TimeLeft:      10,           // 10 minutes default
    NodeNumber:    <assigned>,
    IPAddress:     <client IP>,
    Connected:     true,
    Conn:          <connection>
}
```

### 2.4 Timeout Monitoring

**Implementation:** [`monitorSessionTimeout()`](main.go:561)

Background goroutine monitors session activity:

- **Timeout Duration:** Configured in Config TUI → `timeout_minutes`
- **Admin Exemption:** Configurable via `admin_timeout_exempt`
- **Warning:** Displayed 30 seconds before timeout
- **Action on Timeout:** Session disconnected, connection closed
- **Activity Tracking:** Every keypress updates `LastActivity` timestamp

---

## Phase 3: Authentication - Login Flow

**Entry Point:** User selects [L]ogin from main menu

**Implementation:** [`LoginPrompt()`](auth.go:154) in [`auth.go`](auth.go:154)

### 3.1 Login Screen Display

- Clear screen
- Display GHOSTnet header ANSI art ([`connect.ans`](theme/connect.ans))
- Display instruction text

### 3.2 Username Input

**Field Specifications:**

- Label: "Username: "
- Position: Column 2, Row 10
- Max Length: 20 characters
- Input Method: Character-by-character with reverse video background

**Validation:**

1. **Empty Check:** Username cannot be empty
   - Error: "Username cannot be empty." (displayed for 2 seconds)
   - Action: Clear field, retry input

**ESC Handling:**

- Press ESC → Confirmation prompt: "Do you really want to quit? [Y/N]"
- Y → Cancel login, log event, return to main menu
- N → Clear confirmation, continue input

### 3.3 Password Input

**Field Specifications:**

- Label: "Password: "
- Position: Column 2, Row 11
- Max Length: 20 characters
- Display: Asterisks (\*) for each character
- Input Method: Character-by-character with masking

**Validation:**

1. **Empty Check:** Password cannot be empty
   - Error: "Password cannot be empty." (displayed for 2 seconds)
   - Action: Clear field, retry input

**ESC Handling:** Same as username (quit confirmation)

### 3.4 Authentication Process

**Implementation:** [`AuthenticateUser()`](auth.go:72)

#### Step 1: Locate User File

- Username sanitized and converted to lowercase
- File path: `users/<sanitized_username>.json`
- If file not found → Authentication fails

#### Step 2: Load User Data

- Query SQLite database for user record
- Parse into [`User`](auth.go:15) struct

#### Step 3: Verify Password

**Implementation:** [`VerifyPassword()`](auth.go:123)

- Hash submitted password with salt: `ghostnet_salt_2025`
- Compare with stored hash
- **Security Note:** Currently uses SHA-256; production should use bcrypt

#### Step 4: Update Last Login

- Set `LastLogin` to current timestamp
- Save updated user record to file

### 3.5 Failed Authentication Handling

**Maximum Attempts:** 3 (hardcoded in [`LoginPrompt()`](auth.go:228))

#### After Each Failed Attempt

1. **Increment Counter:** Track failed attempts for this session
2. **Log Event:** Record failed login with details
3. **Generic Error:** Display "Invalid login, try again." (2 seconds)
4. **Security Principle:** Don't reveal whether username or password was wrong
5. **Clear Form:** Erase both username and password fields
6. **Restart Input:** Begin with username input again

#### After 3rd Failed Attempt

1. **Display Warning:** "Too many login tries, hacker -- see ya!"
2. **Log Event:** Record disconnection reason
3. **Permanent Blacklist:** Add IP to blacklist file
   - Reason: "Failed login attempts exceeded"
   - Source: "login_security"
   - **Critical:** This is a permanent ban, not temporary
4. **Force Disconnection:**
   - Set `session.Connected = false`
   - Close TCP connection
   - User cannot retry

**Security Files Updated:**

- Blacklist file: `security/blacklist.txt` (or configured location)
- Security log: `logs/security.log` (if logging enabled)

### 3.6 Successful Authentication

#### Actions on Login Success

1. **Log Success:** Record successful login event
2. **Welcome Message:** Display "Welcome back, {username}!"
3. **Update Session:**

   ```go
   session.Alias = user.Username
   session.SecurityLevel = user.SecurityLevel
   // Special case: username "admin" gets level 255
   if strings.ToLower(username) == "admin" {
       session.SecurityLevel = 255
   }
   ```

4. **Update Node Manager:** Change node display from "Guest" to username
5. **Return to Main Menu:** User now sees authenticated menu options

**State Transition:** Guest → Authenticated User

---

## Phase 4: Authentication - Registration Flow

**Entry Point:** User selects [R]egister from main menu

**Implementation:** [`RegisterPrompt()`](auth.go:343) in [`auth.go`](auth.go:343)

### 4.1 Registration Screen Display

- Clear screen
- Display header ANSI art from theme/
- Show header bar: "New User Registration" with "[ESC] Quit/Cancel"
- Display instruction text about account creation

### 4.2 Username Input & Validation

**Field Specifications:**

- Label: "Username: "
- Position: Column 2, Row 10
- Max Length: 20 characters

**Validation Rules:**

1. **Empty Check:** Username cannot be empty

   - Error: "Username cannot be empty." (2 seconds)

2. **Minimum Length:** Must be at least 3 characters

   - Error: "Username must be at least 3 characters." (2 seconds)
   - Logged as: `REGISTER_FAILED: username too short`

3. **Uniqueness:** Username must not already exist
   - Check via [`UserExists()`](auth.go:146)
   - Error: "Username {name} already exists." (2 seconds)
   - Logged as: `REGISTER_FAILED: username already exists`

**ESC Handling:** Confirmation prompt (same as login)

### 4.3 Password Input & Validation

**Field Specifications:**

- Label: "Password: "
- Position: Column 2, Row 11
- Max Length: 20 characters
- Display: Asterisks for masking

**Validation Rules:**

1. **Empty Check:** Password cannot be empty

   - Error: "Password cannot be empty." (2 seconds)

2. **Minimum Length:** Must be at least 4 characters
   - Error: "Password must be at least 4 characters." (2 seconds)
   - Logged as: `REGISTER_FAILED: password too short`

**ESC Handling:** Confirmation prompt

### 4.4 Email Input (Optional)

**Field Specifications:**

- Label: " Email: " (3 spaces for alignment)
- Position: Column 2, Row 12
- Max Length: 30 characters
- **Optional:** Can be left empty

**Note:** No email format validation currently implemented

### 4.5 Account Creation

**Implementation:** [`CreateUser()`](auth.go:26)

#### Step 1: Create User Directory

- Directory: `users/` (created if doesn't exist)

#### Step 2: Generate Password Hash

- **Algorithm:** SHA-256 with static salt
- **Salt:** `retrograde_salt_2025`
- **Implementation:** [`HashPassword()`](auth.go:115)
- **Production Note:** Should migrate to bcrypt

#### Step 3: Create User Record

```go
User{
     Username:      <entered username>,
     PasswordHash:  <hashed password>,
     SecurityLevel: SecurityLevelRegular, // Default level (10)
     CreatedDate:   <current timestamp>,
     LastLogin:     "",                   // Empty until first login
     Email:         <entered email>,
     Applications:  make(map[string]Application)
}
```

#### Step 4: Save to File

- Filename: `users/<sanitized_lowercase_username>.json`
- Format: Pretty-printed JSON with 2-space indentation

### 4.6 Successful Registration

#### Actions on Registration Success

1. **Log Success:** Record `REGISTER_SUCCESS: new account created`
2. **Auto-Login:** Automatically log in the new user
   - Log login event
   - Update session with user data
3. **Welcome Message:** "Account created successfully. Welcome, {username}!"
4. **Update Node Manager:** Change node from "Guest" to username
5. **Return to Main Menu:** User sees authenticated menu options

**State Transition:** Guest → Authenticated New User

**Security Note:** Registration requires same 3-attempt policy as login (implemented at connection level, not registration level)

---

## Phase 5: Post-Authentication State

### 5.1 Current Implementation

After successful login/registration, user returns to main menu with authenticated privileges:

**Available Options:**

- GHOSTnet information
- Discord invitation (if configured)
- Admin menu (if admin privileges)
- Logout option
- Quit option

**Session State:**

- `session.Alias`: User's username (no longer "Guest")
- `session.SecurityLevel`: User's privilege level (0-255)
- `session.LastActivity`: Continuously updated
- Activity timeout monitoring continues

### 5.2 Planned Feature: Logon Sequence

**Status:** Not currently implemented (legacy documentation reference)

**Planned Flow:**

1. `text/menu/logon.*` - Default logon welcome screen
2. `text/menu/logon1.*` through `logon99.*` - Numbered sequence screens
3. `text/menu/random*.*` - Random welcome screen selection
4. Logon events execution (configurable)
5. Return to main menu

**File Format:** `.ans` or `.asc` files (depending on terminal type detection)

**Implementation Notes:**

- Would require screen file management system
- Terminal type detection already exists
- Event system would need development
- Numbered sequence implies ordered presentation

---

## Phase 6: Session Management

### 6.1 Session Timeout Monitoring

**Implementation:** [`monitorSessionTimeout()`](main.go:561)

#### Configuration

- **Setting:** Config TUI → `timeout_minutes`
- **Admin Exemption:** `admin_timeout_exempt` (true/false)
- **Check Interval:** Every 10 seconds

#### Timeout Behavior

**Normal Users:**

1. **Activity Tracking:** Every keypress updates `session.LastActivity`
2. **Warning Phase:** 30 seconds before timeout
   - Display: "WARNING: You will be disconnected in 30 seconds due to inactivity!"
   - Prompt: "Press any key to remain connected..."
   - If key pressed: Reset warning, continue session
3. **Timeout Phase:** After full timeout period
   - Display: "Session timeout: Disconnected due to {N} minutes of inactivity."
   - Set: `session.Connected = false`
   - Close TCP connection
   - 2-second delay to read message

**Admin Users (when exemption enabled):**

- No timeout monitoring
- Can remain connected indefinitely
- Detected via: [`IsAdmin()`](auth.go:461) check or `SecurityLevel >= AdminSecurityLevel`

### 6.2 Manual Logout

**Entry Point:** User selects [L]ogout from main menu (when authenticated)

#### Logout Process

1. **Log Event:** Record logout action
2. **Display Message:** "Logging out {username}..."
3. **Reset Session:**

   ```go
   session.Alias = "Guest"
   session.SecurityLevel = 0
   ```

4. **Update Node Manager:** Change display back to "Guest"
5. **Return to Main Menu:** Guest-level options shown

**State Transition:** Authenticated User → Guest

**Note:** User remains connected but loses authentication state

### 6.3 Connection Termination

**Quit Option:** User selects [Q]uit from main menu

#### Termination Process

1. Display: "Goodbye! Thanks for visiting GHOSTnet."
2. Return from [`mainMenu()`](main.go:191) function
3. Deferred cleanup executes:
   - Close TCP connection
   - Release node via [`ReleaseNode()`](main.go:143)
4. Session ends

---

## Security Considerations

### 7.1 Password Security

**Current Implementation:**

- Algorithm: SHA-256 with static salt
- Salt Value: `ghostnet_salt_2025` (hardcoded)
- **Risk Level:** Medium

**Recommendations for Production:**

- Migrate to bcrypt or Argon2
- Use per-user salts
- Implement password complexity requirements
- Add password history to prevent reuse

### 7.2 Brute Force Protection

**Current Protections:**

1. **3-Attempt Limit:** Automatic blacklist after 3 failed attempts
2. **Permanent Ban:** IP permanently blocked (not temporary)
3. **Generic Errors:** Don't reveal if username or password was wrong
4. **Rate Limiting:** Pre-authentication connection rate limits
5. **Security Logging:** All failed attempts logged with details

**Effectiveness:** Strong protection against brute force attacks

### 7.3 Session Security

**Protections:**

1. **Timeout Monitoring:** Automatic disconnection on inactivity
2. **Activity Tracking:** Every action updates last activity time
3. **Connection Reference:** Session holds TCP connection for forced closure
4. **Admin Exemption:** Configurable for trusted users

### 7.4 IP-Based Security

**Blacklist Sources:**

1. Manual entries (permanent)
2. Failed login attempts (permanent)
3. Rate limiting violations (temporary → permanent)
4. External threat feeds (cached)

**Whitelist Priority:** Whitelist checked before any other security measure

### 7.5 Information Disclosure Prevention

**Security Principles:**

1. **Generic Rejection Messages:** Blocked users see "Connection temporarily unavailable"
2. **Server-Side Logging:** Detailed reasons logged only on server
3. **Authentication Errors:** Don't reveal username existence
4. **No Enumeration:** Can't determine valid usernames through timing

---

## Integration Points

### 8.1 Configuration System

Config TUI

**Authentication Settings:**

### 8.2 Logging System

**Implementation:** [`logging.go`](logging.go)

**Log Types:**

1. **Login Events:** [`LogLogin()`](logging.go) - Successful authentication
2. **Failed Login Events:** [`LogLoginFailed()`](logging.go) - Failed authentication attempts
3. **Registration Events:** [`LogEvent()`](logging.go) - Account creation and failures
4. **Logout Events:** [`LogLogout()`](logging.go) - User logout actions
5. **Security Events:** [`logSecurityEvent()`](security.go:595) - Security layer events

**Log Files:**

- Primary log: `logs/YYYY-MM-DD.log` (daily rotation)
- Security log: `logs/security.log` (if security logging enabled)

### 8.3 User Storage System

**Database Tables:**

- `users` - Primary user account table
- `user_details` - Extensible user attributes
- `user_preferences` - User-specific settings
- `auth_audit` - Authentication audit trail

**Storage:** SQLite database with comprehensive schema
**Access Pattern:** SQL queries with connection pooling

### 8.4 Node Management System

**Implementation:** [`NodeManager`](types.go:95)

**Responsibilities:**

1. Track active connections per node
2. Display username for each node
3. Enforce maximum node limit
4. Release nodes on disconnection

**Integration Points:**

- Updated on login/registration (Guest → Username)
- Updated on logout (Username → Guest)
- Checked before accepting connections

### 8.5 Security System

**Implementation:** [`SecurityManager`](types.go:176)

**Components:**

1. **Connection Tracker:** Per-IP attempt tracking
2. **Blacklist Manager:** Permanent and temporary bans
3. **Whitelist Manager:** Trusted IP addresses
4. **Geo Cache:** Geographic location data
5. **Threat Intel Cache:** External blocklist IPs

**Initialization:** [`InitializeSecurity()`](security.go:23) called at server startup

---

## Error Handling & Edge Cases

### 9.1 Connection Errors

**Scenario:** Network disruption during session

**Handling:**

- IO errors break read loops
- Deferred cleanup releases resources
- Node automatically freed
- No hanging connections

### 9.2 File System Errors

**User File Not Found:**

- During login: Authentication fails (generic error)
- During registration: Proceed with creation

**Directory Creation Failure:**

- Registration blocked with error message
- Logged for administrator attention

**JSON Parse Errors:**

- User file corrupted: Authentication fails
- Logged as file system error
- Administrator must fix or delete file

### 9.3 Concurrent Access

**Multiple Logins (Same User):**

- Currently allowed (no single-session enforcement)
- Each gets independent session
- Node manager tracks all separately

**File Write Conflicts:**

- Last write wins (no locking)
- User data updates atomic (full file write)
- Minimal risk: updates are infrequent

### 9.4 Timeout Edge Cases

**Activity During Warning:**

- Warning flag reset
- Timer continues from last activity
- User can continue session

**Admin Mid-Session Privilege Change:**

- Timeout exemption checked each cycle
- Admin promotion: Immediate exemption
- Admin demotion: Immediate timeout enforcement

### 9.5 Security Layer Failures

**Geo API Unavailable:**

- Connection allowed (fail-open for availability)
- Error logged for administrator
- Cached data used if available

**Threat Intel Update Failure:**

- Previous data retained
- Error logged with URL and reason
- Next update attempted per schedule

**Blacklist File Errors:**

- File missing: Empty blacklist (connections allowed)
- File corrupted: Skip invalid entries
- Continue with valid entries

---

## State Transitions Summary

```text
┌─────────────────────────────────────────────────────────────┐
│                    Connection Lifecycle                      │
└─────────────────────────────────────────────────────────────┘

State 1: NO CONNECTION
    ↓ [TCP SYN]
State 2: SECURITY CHECK
    ↓ [Pass] / → [Fail: Disconnect]
State 3: NODE ASSIGNED (Guest)
    ↓ [Telnet Negotiation]
State 4: MAIN MENU (Guest)
    ↓ [Login/Register]
State 5: AUTHENTICATION IN PROGRESS
    ↓ [Success] / → [Fail: Retry or Disconnect]
State 6: MAIN MENU (Authenticated)
    ↓ [Use Features]
State 7a: LOGOUT → State 4 (Guest)
State 7b: QUIT → State 8
State 7c: TIMEOUT → State 8
State 8: DISCONNECTED
```

**Critical Points:**

- States 1-2: Pre-BBS security layer
- States 3-4: Guest access (limited features)
- States 5-6: User authentication and full access
- State 7: Exit paths (graceful or forced)

---

## Performance Considerations

### 10.1 File I/O Operations

**User Authentication:**

- Single file read per login attempt
- File size: ~1-2KB per user
- Performance: Acceptable for <1000 concurrent users

**Optimization Opportunities:**

- Implement user data caching
- Use database instead of flat files
- Add connection pooling

### 10.2 Security Checks

**Geo Lookup:**

- First lookup: 100-500ms (API call)
- Cached lookup: <1ms
- Cache duration: 24 hours
- Impact: Minimal (cached after first check)

**Threat Intelligence:**

- Memory-based map lookup
- O(1) complexity
- Pre-loaded at startup
- Background updates (non-blocking)

### 10.3 Concurrent Connections

**Current Limits:**

- Configurable max nodes (default: 10)
- One goroutine per connection
- One goroutine per timeout monitor
- Memory usage: ~2-3MB per connection

**Scalability:**

- Proven stable up to configured node limit
- No connection pooling required
- Telnet protocol naturally low-bandwidth

---

## Future Enhancements

### 11.1 Planned Features

**Logon Sequence (Priority: Medium):**

- Welcome screen display system
- Numbered sequence support (logon1-99)
- Random screen selection
- Event system for custom actions
- ANSI/ASCII art support

**Multi-Factor Authentication (Priority: Low):**

- Email verification codes
- Time-based OTP support
- Recovery codes
- Security questions

**Enhanced Password Security (Priority: High):**

- Migration to bcrypt/Argon2
- Per-user salts
- Password complexity requirements
- Password strength meter
- Password expiration policy

**Session Management (Priority: Medium):**

- Single-session enforcement option
- Session token generation
- Concurrent login limits
- Session transfer between nodes

### 11.2 Potential Improvements

**User Experience:**

- Password reset via email
- "Remember me" functionality
- Last login notification
- Failed attempt notification

**Security:**

- CAPTCHA for registration
- Progressive delays on failed attempts
- Account lockout after X failures
- Two-factor authentication
- IP reputation scoring

**Performance:**

- Database backend option
- User data caching layer
- Session state persistence
- Load balancing support

**Monitoring:**

- Real-time security dashboard
- Login pattern analysis
- Anomaly detection
- Geographic login reports

---

## Troubleshooting Guide

### 12.1 Common Issues

**Problem:** User cannot login despite correct password

**Possible Causes:**

1. Username case mismatch in file
2. Corrupted JSON file
3. File permissions error
4. IP blacklisted from previous attempts

**Resolution:**

1. Check `users/` directory for exact filename
2. Validate JSON syntax in user file
3. Check file permissions (readable by server)
4. Review `security/blacklist.txt` and `logs/security.log`

---

**Problem:** Connection blocked at security layer

**Possible Causes:**

1. IP in blacklist
2. Rate limit exceeded
3. Geographic location blocked
4. Threat intelligence match

**Resolution:**

1. Check `logs/security.log` for specific reason
2. Review `security/blacklist.txt` for IP entry
3. Verify Config TUI security settings
4. Consider whitelist addition if legitimate

---

**Problem:** Session timeout despite activity

**Possible Causes:**

1. Activity tracking not updating
2. Timeout configured too short
3. Admin exemption not working

**Resolution:**

1. Verify `timeout_minutes` setting in config
2. Check if admin exemption is enabled
3. Verify user is in admin list
4. Check session activity tracking in code

---

**Problem:** Registration fails with unclear error

**Possible Causes:**

1. Username already exists (different case)
2. File system permissions
3. Users directory doesn't exist

**Resolution:**

1. Check `users/` directory listings
2. Verify directory write permissions
3. Review registration logs in daily log file
4. Check for sanitization issues with special characters

---

## Documentation Maintenance

**Last Updated:** 2025-10-04
**Code Version:** Current implementation as of this date
**Author:** System Documentation

**Update Triggers:**

- Authentication flow changes
- New security features
- Configuration changes
- Bug fixes affecting login/logout

**Validation:**

- Cross-referenced with source code in `auth.go`, `security.go`, `main.go`
- Tested against actual system behavior
- Line numbers verified for accuracy

---

## Quick Reference

### Key Files

- **Authentication:** [`auth.go`](auth.go)
- **Security Layer:** [`security.go`](security.go)
- **Main Server:** [`main.go`](main.go)
- **Session Types:** [`types.go`](types.go)
- **Telnet I/O:** [`telnet_io.go`](telnet_io.go)
- **Logging:** [`logging.go`](logging.go)

### Configuration Quick Reference

Config TUI

- **Sections:** `[server]`, `[security]`
- **Critical Settings:** `timeout_minutes`, `max_nodes`, `admin_users`, security options

### Security Files

- **Blacklist:** `security/blacklist.txt`
- **Whitelist:** `security/whitelist.txt`
- **Security Log:** `logs/security.log`

### User Data

- **Directory:** `users/`
- **Format:** JSON files
- **Naming:** `<lowercase_username>.json`

---

## End of Authentication Flow Documentation
