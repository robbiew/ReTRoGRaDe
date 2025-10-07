# Retrograde BBS Configuration Editor

A modern Mystic BBS-style terminal user interface (TUI) for editing Retrograde BBS server configuration using BubbleTea.

## Building

### Using build script (recommended):
```bash
# Windows
.\build.bat

# Linux/Mac
./build.sh
```

This builds a single binary that serves both as the BBS server and configuration editor.

## Running the Config Editor

The config editor is a terminal user interface (TUI) application that **requires a proper terminal environment** to run correctly.

### Launch the Config Editor

Run the main binary with the `config` argument:

```bash
# Windows
.\retrograde.exe config

# Linux/Mac
./retrograde config
```

You can also use the `edit` alias:

```bash
.\retrograde.exe edit
```

### Important: Terminal Requirement

⚠️ **Always run from a terminal** - The config editor requires a fully-initialized terminal to function properly. Open Command Prompt, PowerShell, or a terminal before running the command.

## Usage

Navigate through the interface to configure your Retrograde BBS server. The same binary handles both server operation and configuration:

```bash
.\retrograde.exe          # Start BBS server
.\retrograde.exe config   # Launch config editor
.\retrograde.exe edit     # Launch config editor (alias)
```

## Interface Overview

The TUI features a Mystic BBS-style horizontal menu bar with 5 main categories:

### 🏗️ Configuration
- **System Paths** - Data, config, log, and temp directories
- **General Settings** - BBS name, admin security level, network enables, admin users
- **New User Settings** - Default security, auto-validation, welcome messages

### 🌐 Networking
- **Echomail Addresses** - WWIVnet and FTN network addresses
- **Echomail Nodes** - Connected nodes, hub configuration, routing
- **Echomail Groups** - Message areas, moderation settings

### 🖥️ Servers
- **General Settings** - Server name, location, sysop info, user limits
- **Telnet Server** - Port, max nodes, timeouts, admin exemptions
- **SSH Server** - SSH access configuration
- **RLOGIN Server** - RLOGIN protocol settings
- **Security Settings** - Rate limiting, IP management, geo blocking

### ✏️ Editors
- **User Editor** - Manage user accounts and permissions
- **Menu Editor** - Configure BBS menus and navigation
- **Message Base Editor** - Set up message areas and conferences
- **File Base Editor** - Configure file areas and downloads
- **Event Editor** - Schedule automated events and maintenance
- **External/Door Program Editor** - Configure external programs and doors

### 🔧 Other
- **Log Viewer** - View server and security logs
- **Version Info** - System information and version details

## Navigation

### Menu Bar
- **Left/Right Arrows** or **H/L** - Switch between menu categories
- **1-5** - Jump directly to specific categories
- **Down Arrow** or **Enter** - Open submenu

### Submenu
- **Up/Down Arrows** or **J/K** - Navigate items
- **Left/Right Arrows** - Switch menu categories
- **Enter** - Edit selected item
- **Esc** - Return to menu bar
- **Home/End** - Jump to first/last item

### Editing
- **Type** - Enter new value (text/number fields)
- **Y/N** - Toggle boolean values
- **Space/Tab** - Toggle boolean without saving
- **Enter** - Save changes
- **Esc** - Cancel and restore original value

### Global
- **Q** - Quit (will prompt to save changes)
- **Ctrl+C** - Force quit

## Saving Changes

- Changes are kept in memory until you exit
- When quitting with **Q**, you'll be prompted: "Save X change(s) to database? (y/n)"
- Answer **Y** to save to the SQLite database, **N** to discard changes
- Configuration is stored in `C:\retrograde\data\retrograde.db`

## Features

- **Mystic BBS-Style Interface** - Familiar horizontal menu bar navigation
- **Modal Value Editor** - Focused editing experience with type validation
- **Real-time Display** - See current configuration values inline
- **Change Tracking** - Visual indicator of modified fields
- **Breadcrumb Navigation** - Always know your location in the menus
- **Contextual Help** - Field-specific help text and format hints
- **Safe Editing** - Changes kept in memory until explicitly saved
- **Comprehensive Coverage** - All server settings in one interface
- **Enhanced Visual Polish** - Color-coded sections, icons, and status messages

## Example Session

```bash
# Start the editor
.\retrograde.exe config

# Press 1 to jump to Configuration
# Press Down or Enter to open submenu
# Use Up/Down to navigate to "BBS Name"
# Press Enter to edit
# Type your BBS name
# Press Enter to save
# Press Q when done
# Answer 'y' to save all changes to database
```

## Testing

See [TESTING-TUI-v2.md](TESTING-TUI-v2.md) for comprehensive testing instructions and checklist.

The TUI provides a professional, Mystic BBS-style configuration experience with modern enhancements.