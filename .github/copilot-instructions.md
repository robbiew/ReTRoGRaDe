# Retrograde BBS Development Guide

## Project Overview

Retrograde is a telnet-based Bulletin Board System (BBS) written in Go, implementing classic BBS functionality with modern patterns. It serves multiple concurrent telnet connections on port 2323, providing user authentication, menu systems, and extensible command processing.

## Architecture Patterns

### Core Components
- **`cmd/server/main.go`**: Entry point handling server lifecycle, telnet negotiation, and connection routing
- **`internal/config/`**: SQLite-backed configuration with TUI editor using BubbleTea
- **`internal/menu/`**: Renegade-style menu system with command key (cmdkey) execution
- **`internal/auth/`**: User authentication with security levels and session management
- **`internal/telnet/`**: Raw telnet I/O handling with session state tracking
- **`internal/database/`**: SQLite abstraction with schema migration and seeding

### Key Design Principles
1. **Modular separation**: Each `internal/` package handles specific concerns
2. **SQLite-first**: All persistent data (config, users, menus) stored in single database
3. **Session-centric**: `TelnetSession` struct tracks user state across components
4. **Command registry pattern**: Menu commands map to handlers via `CmdKeyRegistry`
5. **Direct authentication**: Bypass traditional main menu, require immediate login
6. **Canvas-based TUI**: Layered rendering with absolute positioning for configuration

## Message System Architecture

### JAM Message Base Format
Three message types supported with type-specific handling:
- **Local BBS**: Simple user-to-user messages (`MSG_TYPELOCAL`)
- **FidoNet Echomail**: Public conferences with routing (`MSG_TYPEECHO`) 
- **FidoNet Netmail**: Private point-to-point (`MSG_TYPENET | MSG_PRIVATE`)

### Message Base Structure
- SQLite conferences and areas configuration via TUI
- JAM format files for message storage
- Default "Local Areas" conference seeded on install
- Automatic kludge/routing generation for FidoNet messages

## BubbleTea TUI System

### Configuration Editor Architecture
- **MVU Pattern**: Model-View-Update with layered canvas rendering
- **4-Level Navigation**: MainMenu → Level2 → Level3 → Level4Modal/EditingValue
- **Type-Safe Editing**: String/Int/Bool/Port/Path with validation
- **Modular Files**: Split from monolithic editor.go into focused components

### TUI Navigation Flow
```bash
Configuration/Servers/Other → Sections → Fields → Value Editing
```

### Canvas-Based Rendering
- Absolute positioning with `overlayString()` functions
- Dynamic sizing based on terminal dimensions
- Blue/cyan color scheme matching classic BBS aesthetics

## Session & Security Architecture

### Connection Flow Sequence
1. **Security Layer**: IP filtering, rate limiting, geo-blocking before BBS access
2. **Node Assignment**: Max concurrent users with node tracking
3. **Telnet Negotiation**: Character mode, echo suppression
4. **Direct Authentication**: Bypass main menu, immediate login/register
5. **Menu System**: Post-auth cmdkey execution

### Security Patterns
- **Three-Strike Rule**: Permanent IP blocklist after 3 failed logins
- **Comprehensive Logging**: All events to daily logs and security.log
- **Session Monitoring**: Timeout warnings, activity tracking
- **Multiple Protection Layers**: Allowlist → Blocklist → Rate limiting → Geo → Threat intel

## First-Time Setup System

### Guided Setup TUI
Run when `data/retrograde.db` doesn't exist:
- Interactive directory configuration
- Real-time validation and error display
- Automatic directory creation with proper permissions
- Database initialization with default menu structure

### Setup Commands
```bash
./retrograde setup  # Force guided setup
./retrograde config # TUI configuration editor  
./retrograde        # Start server (auto-setup if needed)
```

## Development Workflows

### Build Process
```bash
# Always use production build for testing
./build.sh  # Creates release/retrograde-{os}-{arch} binaries
```

### First-Time Setup
```bash
./retrograde setup  # Guided TUI setup wizard
./retrograde config # BubbleTea configuration editor
./retrograde        # Start server
```

### Testing Patterns
- Connect via `telnet localhost 2323`
- Use `security/allowlist.txt` and `security/blocklist.txt` for IP filtering
- Monitor logs in `logs/` directory

## Menu System Architecture

### Command Key (CmdKey) Pattern
The menu system uses 2-letter command codes from classic BBS systems:
- **`MM`**: Read mail, **`MP`**: Post message, **`G`**: Goodbye/logout
- **`-^`**: Go to menu (implemented), **`-/`**: Gosub menu
- **Special keys**: `FIRSTCMD`, `ANYKEY`, `NOKEY`, `ENTER`, `ESC`

### Menu Execution Flow
1. Load menu from database by name
2. Execute `FIRSTCMD` commands
3. Display generic menu with color codes
4. Read user input (single keypress)
5. Match to commands via `findCommands()`
6. Execute via `CmdKeyRegistry.Execute()`

### Adding New Commands
Register in `internal/menu/cmdkeys.go`:
```go
{CmdKey: "XY", Name: "Your Command", Description: "What it does", 
 Category: "User", Handler: handleYourCommand, Implemented: true}
```

## Configuration System

### Config Structure
- **Hierarchical**: `Config.Configuration.General.BBSName`
- **Database-backed**: Stored in SQLite, not files
- **TUI editor**: BubbleTea interface for sysop configuration
- **Path management**: All directories configurable via `PathsConfig`

### Loading Pattern
```go
cfg, err := config.LoadConfig("")  // Loads from default DB path
defer config.CloseDatabase()      // Always defer close
```

## Security & Session Management

### Authentication Flow
1. Telnet connection established
2. Security checks (IP filtering, rate limiting)
3. Node allocation (max concurrent users)
4. Login/registration prompt
5. Session creation with security level

### Session State
`TelnetSession` tracks:
- User identity and security level
- Connection metadata (IP, node number)
- Activity timestamps for timeout detection
- Terminal dimensions (width/height)

## Database Patterns

### Schema Evolution
- Migrations in `database.InitializeSchema()`
- Default data seeding in `database.SeedDefaultMainMenu()`
- Connection pooling via `database.OpenSQLite()`

### Data Access
- Repository pattern: `db.GetMenuCommands(menuID)`
- Transaction handling for consistency
- Error wrapping with context

## ANSI Art & Terminal Handling

### Theme System
- ANSI files in configurable `Themes` directory
- SAUCE metadata stripping via `ui.LoadANSIFile()`
- Pipe color codes: `|15Hello |12World` → colored output
- Cursor positioning: `ui.MoveCursorSequence(col, row)`

### Telnet Protocol
- Character mode negotiation in `negotiateTelnetOptions()`
- Raw keypress handling via `TelnetIO.GetKeyPress()`
- Session timeout monitoring with warnings

### UI Package Organization
- **`internal/ui/colors.go`**: Pure ANSI color definitions and utilities
- **`internal/ui/terminal.go`**: Input prompts and cursor movement
- **`internal/ui/art.go`**: ANSI art file loading and SAUCE stripping
- **`internal/filesystem/sanitize.go`**: Safe filename utilities

## Development Debugging Patterns

### Connection Testing
- Use Syncterm, Netrunner, or basic telnet clients
- Monitor `logs/YYYY-MM-DD.log` and `logs/security.log`
- Check `security/blocklist.txt` for IP blocks
- Verify `data/retrograde.db` permissions

### Common Development Issues
- **Missing directories**: Run guided setup or check path configuration
- **Database corruption**: Delete retrograde.db to force reinit
- **Permission issues**: Ensure write access to configured directories
- **Telnet negotiation**: Some clients need specific terminal settings

## Testing & Debugging

### Development Server
```bash
./retrograde config  # Edit config if paths missing
./retrograde         # Start with debug output
```

### Common Issues
- **Missing directories**: Run guided setup or create paths manually
- **Database issues**: Check `data/retrograde.db` permissions
- **Telnet client**: Use Syncterm, Netrunner, or basic telnet

## Project-Specific Conventions

### Error Handling
- Always wrap errors with context: `fmt.Errorf("failed to load menu %s: %w", name, err)`
- Use specific error types for flow control (e.g., `errNoKeyTimeout`)

### Package Organization
- `internal/` for private application code
- Single responsibility per package
- Interface definitions where abstractions needed

## Message System Architecture

### JAM Message Base Format
Three message types supported with type-specific handling:
- **Local BBS**: Simple user-to-user messages (`MSG_TYPELOCAL`)
- **FidoNet Echomail**: Public conferences with routing (`MSG_TYPEECHO`) 
- **FidoNet Netmail**: Private point-to-point (`MSG_TYPENET | MSG_PRIVATE`)

### Message Base Structure
- SQLite conferences and areas configuration via TUI
- JAM format files for message storage
- Default "Local Areas" conference seeded on install
- Automatic kludge/routing generation for FidoNet messages

## BubbleTea TUI System

### Configuration Editor Architecture
- **MVU Pattern**: Model-View-Update with layered canvas rendering
- **4-Level Navigation**: MainMenu → Level2 → Level3 → Level4Modal/EditingValue
- **Type-Safe Editing**: String/Int/Bool/Port/Path with validation
- **Modular Files**: Split from monolithic editor.go into focused components

### TUI Navigation Flow
```bash
Configuration/Servers/Other → Sections → Fields → Value Editing
```

### Canvas-Based Rendering
- Absolute positioning with `overlayString()` functions
- Dynamic sizing based on terminal dimensions
- Blue/cyan color scheme matching classic BBS aesthetics

## Session & Security Architecture

### Connection Flow Sequence
1. **Security Layer**: IP filtering, rate limiting, geo-blocking before BBS access
2. **Node Assignment**: Max concurrent users with node tracking
3. **Telnet Negotiation**: Character mode, echo suppression
4. **Direct Authentication**: Bypass main menu, immediate login/register
5. **Menu System**: Post-auth cmdkey execution

### Security Patterns
- **Three-Strike Rule**: Permanent IP blocklist after 3 failed logins
- **Comprehensive Logging**: All events to daily logs and security.log
- **Session Monitoring**: Timeout warnings, activity tracking
- **Multiple Protection Layers**: Allowlist → Blocklist → Rate limiting → Geo → Threat intel

## First-Time Setup System

### Guided Setup TUI
Run when `data/retrograde.db` doesn't exist:
- Interactive directory configuration
- Real-time validation and error display
- Automatic directory creation with proper permissions
- Database initialization with default menu structure

### Setup Commands
```bash
./retrograde setup  # Force guided setup
./retrograde config # TUI configuration editor  
./retrograde        # Start server (auto-setup if needed)
```

## Development Debugging Patterns

### Connection Testing
- Use Syncterm, Netrunner, or basic telnet clients
- Monitor `logs/YYYY-MM-DD.log` and `logs/security.log`
- Check `security/blocklist.txt` for IP blocks
- Verify `data/retrograde.db` permissions

### Common Development Issues
- **Missing directories**: Run guided setup or check path configuration
- **Database corruption**: Delete retrograde.db to force reinit
- **Permission issues**: Ensure write access to configured directories
- **Telnet negotiation**: Some clients need specific terminal settings

## Project-Specific Conventions

### Error Handling
- Always wrap errors with context: `fmt.Errorf("failed to load menu %s: %w", name, err)`
- Use specific error types for flow control (e.g., `errNoKeyTimeout`)

### Package Organization
- `internal/` for private application code
- Single responsibility per package
- Interface definitions where abstractions needed

### Memory Bank Documentation
The `memory-bank/` directory contains AI-generated design docs and decision logs:
- `decisionLog.md`: Recent development changes and fixes
- `applicationFlow.md`: Complete authentication and connection flow
- `tui-architecture.md`: BubbleTea implementation details
- `messageBase-format.md`: JAM message format specifications
- Reference for understanding architectural decisions and patterns