# Retrograde BBS

Retrograde is a retro-style Bulletin Board System (BBS) written in Go. This project is experimental and a work in progress. It is not recommended for production use unless you are contributing to development.

## Learning Goals

This project serves as a platform to deepen Go programming expertise. As an experienced BBS sysop and user, I bring knowledge of legacy and modern BBS architectures, while exploring new concepts in development.

As a hobbyist Go developer, I'm leveraging AI tools like Roo Code to accelerate learning and development. Rather than "vibe coding," I use LLMs strategically to guide implementation, debug issues, and explore patterns that would otherwise take months of trial-and-error in my limited free time. This approach allows me to make meaningful progress while maintaining code quality and understanding.

## BBS Goals

### Core Objectives

- Open and transparent development
- Cross-platform support: Windows, Linux (including Raspberry Pi), macOS
- Classic BBS experience: Maintain the spirit and look & feel of software like Telegard and Renegade

### Networking and Messaging

- Telnet and SSH servers
- Message areas supporting Local, FTN echomail, Local Email, Echomail, Netmail, QWK networking, and WWIVnet-style message bases (JAM)
- Features: Scan for new messages, managed scanned areas, set newscan dates, Full Screen Editor, Full Screen Reader, automated FTN setup, import/export echomail

### Modern Features

- Optional Web APIs
- Discord notifications
- ActivityPub integration

### Ease of Use

- Easy installation, running, and maintenance with config editor
- Editors for: Servers, Security, Users, Security Levels, Menus, Events, Message Bases, File Bases, Door Programs

### Extensibility

- API for building external apps: Web frontends and native mobile apps
- Launch external programs (doors): Native console apps and DOS-based doors using Dosemu2
- Internal MCI codes
- Password reset (SMTP)
- Web front-end

### Core Menus

- Main, Message, Files, Doors, Private E-mail
- Node-to-Node Chat

## Features so Far

- **Telnet Server**: Direct telnet access on port 2323 with character-mode negotiation
- **Security System**: IP filtering, rate limiting, geographic blocking, and threat intelligence
- **User Authentication**: Account creation, login, and session management with SQLite storage
- **SQLite Database**: All BBS data stored in SQLite with comprehensive schema
- **TUI Configuration Editor**: Modern BubbleTea-based terminal interface for server configuration
- **Guided First-Time Setup**: Interactive setup wizard for initial configuration
- **ANSI Art Support**: Display of ANSI artwork (SAUCE parsing too)
- **Session Management**: Automatic timeout handling with sysop exemptions
- **Node Management**: Multi-user support with configurable node limits

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
| Menu Construction System | 0% | Renegade-style system for constructing menus and prompts |
| Message Base Configuration & UI | 0% | JAM format: local, echomail (FTN), netmail, User-to-User |
| Message Editor (basic) | 0% | Simple FSE |
| Message Reader (basic) | 0% | Simple FSR |
| Native Door Support | 0% | Linux native door launcher (menu action) |
| DOS Door Support | 0% | Dosemu2 launch door (menu action) |

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

Part of the _Retrograde BBS_ nproject.
