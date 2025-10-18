# ReTRoGRaDe BBS

Retrograde is a retro-style Bulletin Board System (BBS) implemented in Go. This project is experimental and actively under development. It is not recommended for production use unless you are contributing to its development. Please note that this is not a complete BBS implementation -- not even close! -- but rather the initial foundation for one.

## Learning Goals

This project serves as a platform to deepen Go programming expertise. As an experienced BBS sysop and user, I bring knowledge of legacy and modern BBS architectures, while exploring new concepts in development.

### AI Trigger warning

As a hobbyist Go developer, I'm leveraging AI tools like Roo Code, Codex and Claude to accelerate learning and development. Rather than "vibe coding," I use LLMs strategically to guide implementation, debug issues, and explore patterns that would otherwise take months of trial-and-error in my limited free time. This approach allows me to make meaningful progress while maintaining code quality and understanding.

## Core Objectives

- Open and transparent development
- Cross-platform support: Windows, Linux (including Raspberry Pi), macOS
- Classic BBS experience: Maintain the spirit and look & feel of software like Telegard, Renegade, Iniquity, ViSiON-X

## Feature Status

| Feature                         | Progress | Notes                                                                              |
| ------------------------------- | -------- | ---------------------------------------------------------------------------------- |
| Telnet Server                   | 100%     | Completed                                                                          |
| Security System                 | 100%     | White/Blocklist support, GeoIP filtering, Rate Limiting                            |
| SQLite Database                 | 100%     | Scaffolds sensible defaults on initialization                                      |
| TUI Configuration Editor        | 100%     | View and edit configuration files                                                  |
| Guided First-Time Setup         | 100%     | Ensures paths are set correctly                                                    |
| ANSI Art Support                | 100%     | SAUCE strip                                                                        |
| Session Management              | 100%     | Idle timeout, disconnection                                                        |
| Node Management                 | 100%     | Max nodes, per-user limits, logging                                                |
| Auth /Login UI                  | 100%     | Create New User, Login                                                             |
| Times Event System              | 0%       | Do things on a schedule                                                            |
| Menu Construction System        | 100%     | Renegade-style system (TUI) for constructing menus and prompts.                    |
| Menu Commands                   | 1%       | [List](docs/command-key-reference.md)                                              |
| Menu Execution                  | 50%      | Execute Menu Command Logic (stacking, first run, etc)                              |
| Message Base Configuration & UI | 0%       | Local message base configuration                                                   |
| Message Base (FTN) Support      | 0%       | Read/write support for FTN message bases for echomail                              |
| Netmail Support                 | 0%       | Read/write support for private Netmail                                             |
| Private Email Support           | 0%       | Read/write support for private Email                                               |
| Message Editor (basic)          | 0%       | Simple Full Screen Editor                                                          |
| Message Reader (basic)          | 0%       | Simple Full Screen Reader                                                          |
| Native Door Support             | 0%       | Linux native door launcher (menu action)                                           |
| DOS Door Support                | 0%       | Dosemu2 launch door (menu action)                                                  |
| MCI Codes                       | 0%       | Support for MCI codes                                                              |
| Pipe Colors                     | 100%     | Support for Renegade-style pipe colors                                             |
| Upload/Download Functions       | 0%       | SexyZ file transfer, Up/Down, DIZ extraction, File search                          |
| Archivers                       | 0%       | zip, arj, lzh                                                                      |
| Achievements                    | 0%       | Implement achievement tracking and rewards                                         |

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

The script creates `release/retrograde-darwin-arm64`, `release/retrograde-linux-amd64`, etc. for Linux/macOS/Windows in the `release/` directory.

### First-Time Setup

1. Run the binary: `./retrograde` (first run triggers setup)
2. Follow the guided setup wizard to configure paths and initialize the database.
3. Copy ANSI art files from the repo's `theme/` directory to your configured theme directory.

### Configure

Launch the configuration editor for more settings:

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

The server uses IP allowlist and blocklist files for access control. These files don't exist, create them if you want to se them:

**Blocklist** (`security/blocklist.txt`):

```text
# Format: IP_ADDRESS REASON
1.2.3.4 Known spam source
192.168.100.0/24 Blocked subnet range
```

**Allowlist** (`security/allowlist.txt`):

```text
# Format: IP_ADDRESS REASON
127.0.0.1 Localhost
192.168.1.100 Admin home IP
203.0.113.0/24 Trusted ISP range
```

**Features:**

- Comments start with `#`
- Supports individual IPs and CIDR ranges
- Allowlist takes priority over blocklist
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
├── content/            # Content assets for the BBS (e.g. ANSI art)
├── docs/               # Documentation, design notes, and research
├── internal/           # Private application packages
│   ├── auth/           # User authentication, registration, and session management
│   ├── config/         # Configuration management
│   ├── database/       # SQLite database layer
│   ├── filesystem/     # Filesystem operations
│   ├── logging/        # Logging utilities
│   ├── menu/           # Menu construction system, rendering, and navigation
│   ├── security/       # Security features
│   ├── telnet/         # Telnet I/O
│   ├── tui/            # Configuration TUI
│   └── ui/             # UI utilities (e.g. ANSI art, terminal handling)
├── memory-bank/        # Development documentation (AI notes, design docs, etc.)
└── release/            # Build artifacts and release binaries

```

## Requirements

- Go 1.24+
- Modern terminal/telnet client with ANSI support (like Syncterm, Netrunner, Icy_Term or MagiTerm)

## Dependencies

- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `modernc.org/sqlite` - SQLite database
- `github.com/mattn/go-isatty` - Terminal detection
- `golang.org/x/text` - Text processing utilities
- `github.com/google/uuid` - UUID generation

---

Part of the _Retrograde BBS_ nproject.
