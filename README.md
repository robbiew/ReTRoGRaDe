# Retrograde BBS

A retro-style Bulletin Board System (BBS) written in Go. This is experimental and Work In Progress. I'd recommend against using unless you are developing/contributing.

## Features

- **Telnet Server**: Direct telnet access on port 2323 with character-mode negotiation
- **Security System**: IP filtering, rate limiting, geographic blocking, and threat intelligence
- **User Authentication**: Account creation, login, and session management with SQLite storage
- **SQLite Database**: All BBS data stored in SQLite with comprehensive schema
- **TUI Configuration Editor**: Modern BubbleTea-based terminal interface for server configuration
- **JAM Message Base Format**: Full support for JAM message bases with local messages, echomail, and netmail
- **FidoNet Echomail & Netmail**: Echomail network support with SEEN-BY/PATH routing and netmail point-to-point messaging
- **Guided First-Time Setup**: Interactive setup wizard for initial configuration
- **ANSI Art Support**: Display of ANSI artwork for enhanced user experience
- **Session Management**: Automatic timeout handling with sysop exemptions
- **Node Management**: Multi-user support with configurable node limits

## Quick Start

### Build

```bash
# Build for current platform
go build -o retrograde ./cmd/server

# Or use the build script for production binaries
chmod +x build.sh
./build.sh
```

The build script creates:

- `release/retrograde` / `release/retrograde.exe` - Server binaries for Linux/macOS and Windows

### First-Time Setup

```bash
./retrograde
```

On first run, Retrograde will automatically launch a guided setup wizard to configure:

- Directory paths for data, logs, messages, themes, etc.
- Database initialization
- Optional sysop account creation

### Configure

```bash
./retrograde config
```

Launch the TUI configuration editor to customize server settings, security options, and BBS configuration.

### Run Server

```bash
./retrograde
```

The server will start on the configured telnet port (default: 2323).

## Configuration

### Configuration Editor TUI

A BBS-style terminal UI is available for editing configuration:

```bash
# Launch configuration editor
./retrograde config
```

See [README-config-editor.md](README-config-editor.md) for full configuration editor documentation.

### Security File Management

The server uses IP whitelist and blacklist files for access control:

**Blacklist** (`security/blacklist.txt`):

```
# Format: IP_ADDRESS REASON
1.2.3.4 Known spam source
192.168.100.0/24 Blocked subnet range
```

**Whitelist** (`security/whitelist.txt`):

```
# Format: IP_ADDRESS REASON
127.0.0.1 Localhost
192.168.1.100 Admin home IP
203.0.113.0/24 Trusted ISP range
```

**Features:**

- Comments start with `#`
- Supports individual IPs and CIDR ranges
- Whitelist takes priority over blacklist
- Files are loaded at server startup
- Auto-populated by rate limiting system

### Data Storage

All BBS data is stored in SQLite database (`data/retrograde.db`):

- **User accounts**: Authentication, profiles, and preferences
- **Configuration**: Server settings and BBS configuration
- **Sessions**: Active user sessions and node management
- **Messages**: JAM message base metadata and indexing
- **Security**: Audit logs and threat intelligence data

**Security Logs** (`logs/` directory):

- Security events logged to `logs/security.log`
- Connection attempts, blocks, and system events
- Daily logs in `logs/YYYY-MM-DD.log` format

## Usage

1. **Users**: Connect via telnet to register accounts, login, and access BBS features
2. **SysOps**: Use the TUI configuration editor to manage users, security settings, and message areas
3. **Network**: Configure echomail areas for FidoNet-style message networking

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```
retrograde/
├── cmd/
│   └── server/         # Main BBS server binary
├── internal/           # Private application packages
│   ├── auth/           # User authentication
│   ├── config/         # Configuration management
│   ├── database/       # SQLite database layer
│   ├── jam/            # JAM message base implementation
│   ├── logging/        # Logging utilities
│   ├── security/       # Security features
│   ├── telnet/         # Telnet I/O
│   ├── tui/            # Configuration TUI
│   └── ui/             # UI utilities
├── memory-bank/        # Development documentation
├── theme/              # ANSI art assets
├── release/            # Build artifacts
└── docs/               # Documentation
```

## Requirements

- Go 1.22+
- Terminal/telnet client with ANSI support

## Dependencies

- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling

---

_Part of the Retrograde BBS network project_
