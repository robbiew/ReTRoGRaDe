# Retrograde BBS

Retrograde is a retro-style Bulletin Board System (BBS) implemented in Go. This project is experimental and actively under development. It is not recommended for production use unless you are contributing to its development. Please note that this is not a complete BBS implementation -- not even close! -- but rather the initial foundation for one.

## Learning Goals

This project serves as a platform to deepen Go programming expertise. As an experienced BBS sysop and user, I bring knowledge of legacy and modern BBS architectures, while exploring new concepts in development.

### AI Trigger warning

As a hobbyist Go developer, I'm leveraging AI tools like Roo Code to accelerate learning and development. Rather than "vibe coding," I use LLMs strategically to guide implementation, debug issues, and explore patterns that would otherwise take months of trial-and-error in my limited free time. This approach allows me to make meaningful progress while maintaining code quality and understanding.

## Core Objectives

- Open and transparent development
- Cross-platform support: Windows, Linux (including Raspberry Pi), macOS
- Classic BBS experience: Maintain the spirit and look & feel of software like Telegard, Renegade, Iniquity, ViSiON-X

## Feature Status

| Feature | Progress | Notes |
|---------|----------|-------|
| Telnet Server | 100% | Completed |
| Security System | 100% | Completed |
| User Authentication | 100% | Completed |
| SQLite Database | 100% | Completed |
| TUI Configuration Editor | 100% | Completed |
| Guided First-Time Setup | 100% | Completed |
| ANSI Art Support | 100% | Completed |
| Session Management | 100% | Completed |
| Node Management | 100% | Completed |
| Login UI | 50% | Needs layout work |
| Event System | 0% | e.g. Logon Event List -> main menu |
| Menu Construction System | 25% | Renegade-style system (TUI) for constructing menus and prompts |
| Message Base Configuration & UI | 0% | JAM format: local, echomail (FTN), netmail, User-to-User |
| Message Editor (basic) | 0% | Simple FSE |
| Message Reader (basic) | 0% | Simple FSR |
| Native Door Support | 0% | Linux native door launcher (menu action) |
| DOS Door Support | 0% | Dosemu2 launch door (menu action) |

## Quick Start

### Build

Build for your platform:

```bash
go build -o retrograde ./cmd/server
```

Or use the build script for optimized binaries:

```bash
chmod +x build.sh
./build.sh
```

On Windows, use the batch script:

```batch
build.bat
```

The script creates `release/retrograde-darwin-arm64`, `release/retrograde-linux-amd64`, etc. for Linux/macOS/Windows.

### First-Time Setup

1. Run the binary: `./retrograde`
2. Follow the guided setup wizard to configure paths and initialize the database.
3. Copy ANSI art files from the repo's `theme/` directory to your configured theme directory.

### Configure

Launch the configuration editor:

```bash
./retrograde config
```

Use ESC to exit and save changes.

### Run Server

Start the server:

```bash
./retrograde
```

It runs on the configured telnet port (default: 2323).

## Command Line Options

- `./retrograde` - Run the server
- `./retrograde config` (or -config, --config, /config) - Launch configuration editor
- `./retrograde setup` (or install, -setup, --setup, -install, --install) - Run guided setup

## Configuration

### Configuration Editor TUI

A BBS-style terminal UI is available for editing configuration:

```bash
# Launch configuration editor
./retrograde config
```

### Security File Management

The server uses IP whitelist and blacklist files for access control. These files don't exist, create them if you want to se them:

**Blacklist** (`security/blacklist.txt`):

```text
# Format: IP_ADDRESS REASON
1.2.3.4 Known spam source
192.168.100.0/24 Blocked subnet range
```

**Whitelist** (`security/whitelist.txt`):

```text
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

All BBS data is stored in SQLite database, e.g. default (`data/retrograde.db`):

- **User accounts**: Authentication, profiles, and preferences
- **Configuration**: Server settings and BBS configuration
- **Sessions**: Active user sessions and node management
- **Security**: Audit logs and threat intelligence data

**Security Logs** (`logs/` directory):

- Security events logged to `logs/security.log`
- Connection attempts, blocks, and system events
- Daily logs in `logs/YYYY-MM-DD.log` format

## Usage

1. **Users**: Connect via telnet to register accounts, login, and access BBS features
2. **SysOps**: Use the TUI configuration editor to manage users, security settings, and message areas

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```text
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
│   ├── ui/             # UI utilities
│   ├── menu/           # Menu construction system
│   └── util/           # Utility functions
├── memory-bank/        # Development documentation
├── content/            # Content assets for the BBS
├── security/           # Security configuration files
├── release/            # Build artifacts
└── docs/               # Documentation
```

## Requirements

- Go 1.24+
- Terminal/telnet client with ANSI support (like Syncterm, Netrunner or Icy_Term)

## Dependencies

- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `modernc.org/sqlite` - SQLite database
- `github.com/mattn/go-isatty` - Terminal detection
- `golang.org/x/text` - Text processing utilities
- `github.com/google/uuid` - UUID generation

---

Part of the _Retrograde BBS_ nproject.
