[2025-10-13 14:15:00] - The ASCII normalization of mojibake glyphs throughout the TUI rendering code
[2025-10-13 10:30:00] - The further splitting of menu structures into separate files (editor_menu_configuration.go, editor_menu_servers.go, editor_menu_editors.go, editor_menu_other.go)
[2025-10-13 09:00:00] - The initial split of the 6447-line editor.go into multiple focused files (editor.go, editor_menu_structure.go, editor_update.go, editor_view.go, editor_canvas.go, editor_data.go)
[2025-10-13 16:07:00] - Updated TUI editor to print ANSI background and removed the existing background pattern (shaded boxes)
[2025-10-08 20:47:00] - Fixed unused field 'selectedMenuDataIndex' in internal/tui/editor.go by removing the unused struct field (U1000)
[2025-10-08 20:40:00] - Modified menu modify TUI to default to "Menu Data" tab instead of "Commands", made Menu Data tab directly show edit modal instead of field list, fixed inactive tab visibility by changing background color from white to dark, and updated footer text for modal navigation behavior
[2025-10-08 20:03:00] - Fixed index out of range panic when editing menu Flags field by properly initializing modalFields in setupMenuEditDataModal() before entering MenuEditDataMode
[2025-10-08 20:00:00] - Enhanced menu management TUI with tabbed interface for editing menu fields and commands, allowing navigation between Commands and Menu Data tabs with TAB/1/2 keys, and direct editing of all menu properties and command details
[2025-10-08 11:40:42] - Fixed unused field 'editingMenu' in internal/tui/editor.go by removing the unused struct field (U1000)

[2025-10-07 11:40:27] - Fixed unused variable issue in internal/auth/auth.go by removing unused global variable 'db'
[2025-10-07 11:57:00] - Implemented basic user management interface in TUI config editor with user listing functionality
[2025-10-07 11:48:29] - Fixed performance issue in internal/auth/auth.go by compiling username validation regex once instead of in a loop (SA6000)

[2025-10-07 18:59:01] - Fixed unused function parseInt in internal/config/config.go by removing it

[2025-10-07 19:02:12] - Fixed User Editor visibility issue in config TUI by updating buildListItems function to include ActionItem types in submenu display

[2025-10-07 19:07:28] - Fixed User Editor selection issue by setting messageTime when displaying error messages in TUI

[2025-10-07 19:11:35] - Fixed database initialization in config TUI by opening database lazily when User Editor is selected

[2025-10-07 19:13:59] - Added database path validation and schema initialization for User Editor in config TUI

[2025-10-07 19:52:32] - Implemented guided first-time user setup in cmd/server/main.go with console prompts for paths, sysop account, and theme copying

[2025-10-07 20:01:30] - Fixed auth initialization in guided setup by calling auth.Init before CreateUser

[2025-10-07 20:09:27] - Cleaned up guided setup output by removing debug messages and adding exit instructions after setup completion

[2025-10-08 15:25:00] - Completed user management interface implementation in TUI config editor with full CRUD operations

[2025-10-08 15:25:00] - Updated all documentation files to reflect current application state including JAM message base support, SQLite storage, and guided setup

[2025-10-08 18:53:00] - Fixed Config TUI menu command count display issue by modifying menuListItem struct to include commandCount field, updating loadMenus to fetch actual command counts from database, and fixing menuDelegate Render function to use the stored count instead of hardcoded 0

[2025-10-08 18:57:00] - Fixed menu execution to use TelnetIO instead of standard I/O by modifying MenuExecutor to accept TelnetIO, updating all print/scan operations to use TelnetIO methods, and adding IO to ExecutionContext for cmdkey handlers

[2025-10-08 19:04:00] - Fixed menu logout behavior to properly disconnect users instead of returning to pre-login menu when pressing 'G' (goodbye) from menu system

[2025-10-08 20:58:39] - Fixed QF1003 linter issue by replacing if-else if chain with tagged switch on item.submenuItem.ID in handleLevel2MenuNavigation function

[2025-10-08 21:00:48] - Fixed menu tabs display issue by changing inactive tab background from ColorBgMedium to ColorBgGrey in renderMenuTabs function

[2025-10-08 21:05:35] - Fixed menu modify interface to show tab bar consistently by changing Menu Data tab to display field list instead of directly showing modal

[2025-10-08 21:54:10] - Fixed menu modify interface height consistency by padding command list to 13 lines, added F1 help to footer and handler

[2025-10-08 21:57:32] - Fixed tab navigation to correctly handle separate indices for Menu Data and Menu Commands tabs

[2025-10-08 21:58:54] - Fixed enter key handling to correctly edit menu data fields in Menu Data tab instead of menu commands

[2025-10-08 22:03:39] - Fixed breadcrumb and footer display in menu edit modal screens by changing from return to overlay

[2025-10-08 22:04:48] - Fixed Menu Data tab display issue by using overlayStringCenteredWithClear to clear texture before overlaying modal

[2025-10-08 22:07:06] - Changed Menu Data tab to show field list with visible tab bar, enter opens edit modal

[2025-10-08 22:11:19] - Changed Menu Data tab to show edit modal directly instead of field list
