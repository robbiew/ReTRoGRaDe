# Retrograde BBS

A retro-style BBS written in Go. This is experimental and Work In Progress. I'd recommend against using unless you are developing/contributing.

## Features So Far

- **Telnet Server**: Direct telnet access with terminal detection
- **Security System**: IP filtering, rate limiting, geographic blocking, and threat intelligence
- **User Authentication**: Basic Account creation and Login
- **Sqlite Storage**: BBS data kept in SQLite databases
- **TUI Configuration Program**: For graphical BBS configuration

## Quick Start

### Build
```bash
# Build for current platform
go build -o retrograde

# Or use the build script for production binaries (builds server + config-editor)
chmod +x build.sh
./build.sh
```

The build script creates:
- `release/retrograde` / `releae/retrigrade.exe` - Server binaries for Windows and Linux


### Configure
```bash
./retrograde config
```


### Run Server
```bash
./retrograde
```

The server will start.

## Configuration

### Configuration Editor TUI

A BBS-style terminal UI is available for editing configuration:

```bash
# Using standalone binary
./config-editor        # Linux/macOS
.\config-editor.exe    # Windows

# Or integrated with server
.\retrograde.exe config
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

User accounts and applications are stored in JSON format:

**Security Logs** (`logs/` directory):
- Security events logged to `logs/security.log`
- Connection attempts, blocks, and system events

## Usage

1. **Users**: Connect via telnet to register, login, and submit network applications

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```
retrograde/
├── cmd/
│   ├── server/         # Main BBS server binary
│   └── config-editor/  # Standalone configuration editor
├── internal/           # Private application packages
│   ├── auth/           # User authentication
│   ├── config/         # Configuration management
│   ├── discord/        # Discord integration
│   ├── logging/        # Logging utilities
│   ├── security/       # Security features
│   ├── telnet/         # Telnet I/O
│   ├── tui/            # Configuration TUI
│   └── ui/             # UI utilities
├── theme/              # ANSI art assets
└── docs/               # Documentation
```

## Requirements

- Go 1.22+
- Terminal/telnet client with ANSI support

## Dependencies

- `github.com/gtuk/discordwebhook` - Discord notifications

---

*Part of the Retrograde BBS network project*