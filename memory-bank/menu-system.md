# Menu System Architecture

## Overview

Implement a menu system inspired by Renegade BBS, retaining its configuration approach while modernizing with database storage, improved TUI editing, and Go-based execution. The system will support menu files with commands, ACS requirements, cmdkeys for actions, and generic menu display. Focus on basic functionality for testing: configure menus via TUI, execute a starting main menu post-login.

## Database Schema

- **menus table**:
  - id (INTEGER PRIMARY KEY)
  - name (TEXT UNIQUE, e.g., 'MAIN')
  - titles (TEXT, JSON array of strings)
  - help_file (TEXT)
  - long_help_file (TEXT)
  - prompt (TEXT)
  - acs_required (TEXT)
  - password (TEXT, hashed)
  - fallback_menu (TEXT)
  - forced_help_level (INTEGER, 0-3)
  - generic_columns (INTEGER)
  - generic_bracket_color (INTEGER)
  - generic_command_color (INTEGER)
  - generic_desc_color (INTEGER)
  - flags (TEXT, e.g., 'C---T-----')

- **menu_commands table**:
  - id (INTEGER PRIMARY KEY)
  - menu_id (INTEGER, FOREIGN KEY to menus.id)
  - command_number (INTEGER)
  - keys (TEXT, e.g., 'R')
  - long_description (TEXT)
  - short_description (TEXT)
  - acs_required (TEXT)
  - cmdkeys (TEXT, e.g., 'MM')
  - options (TEXT)
  - flags (TEXT)

## Data Models

- **Menu struct**: Fields matching DB schema, with Titles as []string (unmarshaled from JSON).
- **MenuCommand struct**: Fields matching DB schema.
- Repository interfaces for CRUD operations on menus and commands.

## Menu Execution Logic

1. Load menu by name; check ACS/password; set help level.
2. Clear screen/display titles based on flags.
3. Execute EVERYTIME commands.
4. Display generic menu (columns/colors based on help level).
5. Display prompt; read user input.
6. Match input to command keys (support hotkeys, special keys like ENTER).
7. Check command ACS; execute cmdkeys via registry map (e.g., 'MM' -> readMail function).
8. Handle command linking, goto/gosub menus.
9. Loop until quit or logout.

Cmdkeys registry: Map strings to functions taking user context and options. Implement core cmdkeys (MM, MP, G, etc.) initially.

## TUI Menu Editor

- **Main Editor**: List menus; (D)elete, (I)nsert, (M)odify, (Q)uit.
- **Modify Menu**: Display commands; (D)elete, (I)nsert, (M)odify, (P)osition, (S/L) generic display, (T)oggle format, (X) menu data.
- **Edit Command**: Form for keys, descriptions, ACS, cmdkeys, options.
- **Edit Menu Data**: Form for titles, prompt, ACS, etc.
- Use BubbleTea, follow existing TUI patterns (e.g., user editor).

## Integration with BBS Flow

- Add "Start Menu" to configuration (default 'MAIN').
- Post-login, load start menu and enter execution loop.
- Add "Menu Editor" to TUI config under Editors.
- For pre-login, consider a separate menu or integrate with login UI.

## Testing Approach

1. Create DB migration for schema.
2. Implement Menu/MenuCommand models and basic DB functions.
3. Implement cmdkey registry with handlers for MM (read mail), MP (post), G (goodbye).
4. Seed a basic MAIN menu via DB or TUI editor: commands for mail, post, logout.
5. Integrate: After login, load MAIN menu.
6. Test via telnet: Login, view menu, execute commands, verify execution and navigation.

## Modern Improvements over Renegade

- **Database Storage**: Easier backup, search, no file corruption; atomic updates.
- **Better Validation**: Stronger ACS checks, input sanitization, error handling.
- **Extensibility**: JSON for complex fields; potential for custom cmdkeys or plugins.
- **Improved UI**: Modern TUI editor with forms; future web interface.
- **Security**: Hashed passwords, integrated with Retrograde security.
- **Performance**: DB queries vs. file parsing; cached menus.
- **Maintenance**: No manual file editing; version control friendly.
- **Cross-Platform**: No DOS limitations; Go concurrency.
- **Integration**: Seamless with existing Retrograde features (users, messages, etc.).

## Phase 1: Basic Menu System Implementation

- Create DB schema migration for menus and menu_commands tables
- Implement Menu and MenuCommand structs with JSON marshaling
- Create repository interfaces and implementations for menu CRUD
- Implement cmdkey registry and basic handlers (MM, MP, G)
- Build TUI menu editor with main list and command editing
- Implement menu execution engine with prompt/input handling
- Add start menu config option
- Integrate menu loading post-login
- Test basic MAIN menu creation and execution
