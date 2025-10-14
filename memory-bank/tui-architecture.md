# TUI Architecture - Actual Implementation

## Retrograde Configuration Editor - BubbleTea TUI

**Version**: 1.0
**Date**: 2025-10-13
**Author**: Code Analysis
**Status**: Implemented

---

## Table of Contents

- [TUI Architecture - Actual Implementation](#tui-architecture---actual-implementation)
  - [Retrograde Configuration Editor - BubbleTea TUI](#retrograde-configuration-editor---bubbletea-tui)
  - [Table of Contents](#table-of-contents)
  - [Executive Summary](#executive-summary)
  - [Core Architecture](#core-architecture)
    - [Technology Stack](#technology-stack)
    - [Architecture Pattern](#architecture-pattern)
    - [Key Files](#key-files)
    - [File Organization](#file-organization)
  - [Navigation Modes](#navigation-modes)
    - [1. MainMenuNavigation (Level 1)](#1-mainmenunavigation-level-1)
    - [2. Level2MenuNavigation (Level 2)](#2-level2menunavigation-level-2)
    - [3. Level3MenuNavigation (Level 3)](#3-level3menunavigation-level-3)
    - [4. Level4ModalNavigation \& EditingValue (Level 4)](#4-level4modalnavigation--editingvalue-level-4)
  - [Component Structure](#component-structure)
    - [Core Data Structures](#core-data-structures)
    - [Menu Hierarchy](#menu-hierarchy)
  - [Data Flow](#data-flow)
    - [Configuration Loading](#configuration-loading)
    - [State Transitions](#state-transitions)
    - [Data Persistence](#data-persistence)
  - [Rendering System](#rendering-system)
    - [Layered Canvas Architecture](#layered-canvas-architecture)
    - [Visual Styling](#visual-styling)
    - [Responsive Design](#responsive-design)
  - [Configuration Mapping](#configuration-mapping)
    - [Data Types Supported](#data-types-supported)
    - [Validation System](#validation-system)
    - [Change Tracking](#change-tracking)

---

## Executive Summary

The Retrograde Configuration Editor is implemented as a BubbleTea-based TUI that provides a hierarchical, Mystic BBS-style interface for system configuration. The actual implementation features a four-level navigation system with layered rendering and type-safe value editing.

---

## Core Architecture

### Technology Stack

- **Framework**: [BubbleTea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework for Go
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling library
- **Lists**: [Bubbles](https://github.com/charmbracelet/bubbles) - BubbleTea components
- **Language**: Go 1.22.2

### Architecture Pattern

The TUI follows the **Model-View-Update (MVU)** pattern enforced by BubbleTea:

- **Model**: Application state and data structures
- **View**: Rendering logic using layered canvas approach
- **Update**: Event handling and state transitions

### Key Files

- `internal/tui/editor.go` - Core model structures and initialization
- `internal/tui/editor_data.go` - Data structures and helper functions
- `internal/tui/editor_menu_structure.go` - Main menu building logic
- `internal/tui/editor_menu_configuration.go` - Configuration menu definitions
- `internal/tui/editor_menu_servers.go` - Server menu definitions
- `internal/tui/editor_menu_editors.go` - Editor menu definitions
- `internal/tui/editor_menu_other.go` - Other menu definitions
- `internal/tui/editor_update.go` - BubbleTea update function
- `internal/tui/editor_view.go` - BubbleTea view function
- `internal/tui/editor_canvas.go` - Canvas rendering utilities
- `internal/config/` - Configuration data structures
- `internal/database/` - SQLite backend for persistence

### File Organization

Following a recent refactoring, the original monolithic `editor.go` (6447 lines) has been split into multiple focused files to improve maintainability and organization:

#### Core Components

- `editor.go`: Core model structures and initialization logic
- `editor_data.go`: Data structures, types, and helper functions

#### Menu System

- `editor_menu_structure.go`: Main menu building and hierarchy logic
- `editor_menu_configuration.go`: Configuration-specific menu definitions
- `editor_menu_servers.go`: Server-related menu definitions
- `editor_menu_editors.go`: Editor-related menu definitions
- `editor_menu_other.go`: Other/miscellaneous menu definitions

#### BubbleTea Implementation

- `editor_update.go`: Update function handling user input and state transitions
- `editor_view.go`: View function responsible for rendering the UI
- `editor_canvas.go`: Canvas-based rendering utilities and positioning logic

This modular structure allows for better separation of concerns and easier maintenance of individual components.

---

## Navigation Modes

The TUI implements **4 navigation modes** with hierarchical progression:

### 1. MainMenuNavigation (Level 1)

- **Purpose**: Top-level category selection
- **Display**: Centered vertical menu with categories
- **Controls**: ↑↓/hjkl navigation, Enter to select, Q to quit
- **Visual**: Blue-bordered box with white/blue highlighting

### 2. Level2MenuNavigation (Level 2)

- **Purpose**: Section selection within category
- **Display**: Left-anchored submenu with section headers
- **Controls**: ↑↓ navigation, Enter to drill down, Esc to go back
- **Visual**: Cyan-bordered overlay with lightbar selection

### 3. Level3MenuNavigation (Level 3)

- **Purpose**: Field selection within section
- **Display**: Bottom-right anchored field list
- **Controls**: ↑↓ navigation, Enter to edit, Esc to go back
- **Visual**: Blue-bordered modal with field highlighting

### 4. Level4ModalNavigation & EditingValue (Level 4)

- **Purpose**: Value editing with type-specific input
- **Display**: Centered modal with inline editing
- **Controls**: Type-dependent (Y/N for bools, text input for others)
- **Visual**: Full-width modal with blue highlighting

---

## Component Structure

### Core Data Structures

```go
type Model struct {
    config       *config.Config        // Configuration data
    navMode      NavigationMode        // Current navigation state
    activeMenu   int                   // Selected main menu index
    submenuList  list.Model            // BubbleTea list for Level 2
    modalFields  []SubmenuItem         // Fields in current section
    modalFieldIndex int                // Selected field in modal
    editingItem  *MenuItem             // Currently editing item
    textInput    textinput.Model       // BubbleTea text input
    screenWidth/Height int             // Terminal dimensions
    message      string                // Status messages
    modifiedCount int                  // Change tracking
}
```

### Menu Hierarchy

```
Configuration
├── Paths
│   ├── Database (path)
│   ├── File Base (path)
│   ├── Logs (path)
│   ├── Message Base (path)
│   ├── System (path)
│   └── Themes (path)
├── General
│   ├── BBS Name (string)
│   ├── BBS Location (string)
│   ├── SysOp Name (string)
│   ├── SysOp Timeout Exempt (bool)
│   ├── System Password (string)
│   ├── Timeout Minutes (int)
│   ├── Default Theme (string)
│   └── Start Menu (string)
├── New Users
│   ├── Allow New Users (bool)
│   ├── Ask Real Name (bool)
│   └── Ask Location (bool)
├── Auth Persistence
│   ├── Max Failed Attempts (int)
│   ├── Account Lock Minutes (int)
│   └── Password Algorithm (string)
└── ...

Servers
├── General Settings
│   ├── Max Nodes (int)
│   └── Max Connections/IP (int)
├── Telnet Server
│   ├── Active (bool)
│   ├── Port (port)
│   └── ...
└── ...

Networking, Editors, Other
```

---

## Data Flow

### Configuration Loading

1. `RunConfigEditorTUI()` called from main
2. `InitialModelV2()` creates model with config data
3. `buildMenuStructure()` maps config to menu hierarchy
4. BubbleTea program starts with `tea.NewProgram()`

### State Transitions

```
MainMenuNavigation → Level2MenuNavigation → Level3MenuNavigation → Level4ModalNavigation → EditingValue
      ↑                       ↑                        ↑                        ↑
      └───────────────────────┴────────────────────────┴────────────────────────┘
                              Esc returns to previous level
```

### Data Persistence

- Changes tracked via `modifiedCount`
- Save prompt on quit if `modifiedCount > 0`
- `config.SaveConfig()` writes to SQLite database
- Real-time validation prevents invalid data

---

## Rendering System

### Layered Canvas Architecture

The rendering uses a **canvas-based approach** with absolute positioning:

```go
func (m Model) View() string {
    canvas := make([]string, m.screenHeight)  // Array of screen lines
    for i := range canvas {
        canvas[i] = strings.Repeat(" ", m.screenWidth)  // Initialize empty
    }

    // Layer components onto canvas
    m.overlayStringCentered(canvas, mainMenuStr)     // Level 1
    m.overlayString(canvas, level2Str, row, col)     // Level 2
    m.overlayStringCenteredWithClear(canvas, modalStr) // Level 3/4

    return strings.Join(canvas, "\n")
}
```

### Visual Styling

- **Colors**: Blue (#33) primary, Cyan (#51) secondary, Gray (#250) text
- **Borders**: Rounded corners with `lipgloss.RoundedBorder()`
- **Highlighting**: Full-width lightbar selection
- **Typography**: Bold for active items, normal for inactive

### Responsive Design

- Dynamic sizing based on `screenWidth`/`screenHeight`
- Centered modals that adapt to terminal size
- Scrollable content for long lists
- ANSI escape sequences for positioning

---

## Configuration Mapping

### Data Types Supported

- **StringValue**: Text input with validation
- **IntValue**: Numeric input with range checking
- **BoolValue**: Y/N toggle with visual indicators
- **ListValue**: Comma-separated values
- **PortValue**: Integer with port range (1-65535)
- **PathValue**: File/directory paths

### Validation System

- **Built-in**: Type-specific validation (port ranges, required fields)
- **Custom**: Per-field validation functions
- **Real-time**: Immediate feedback on invalid input
- **Save Prevention**: Invalid data blocks saving

### Change Tracking

- `modifiedCount` increments on successful edits
- Visual indicator in footer when changes exist
- Save confirmation dialog on exit with unsaved changes
- Automatic backup prevention for invalid configurations
