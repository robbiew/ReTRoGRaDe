package menu

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/robbiew/retrograde/internal/database"
	"github.com/robbiew/retrograde/internal/telnet"
)

// getTerminalDimensions returns the terminal width and height from the session,
// with fallback to 80x24 if dimensions aren't available
func getTerminalDimensions(ctx *ExecutionContext) (width, height int) {
	if ctx != nil && ctx.Session != nil {
		if ctx.Session.Width > 0 {
			width = ctx.Session.Width
		} else {
			width = 80
		}
		if ctx.Session.Height > 0 {
			height = ctx.Session.Height
		} else {
			height = 24
		}
	} else {
		width = 80
		height = 24
	}
	return
}

// MenuExecutor handles the execution of menus
type MenuExecutor struct {
	db       database.Database
	registry *CmdKeyRegistry
	io       *telnet.TelnetIO
}

// NewMenuExecutor creates a new menu executor
func NewMenuExecutor(db database.Database, io *telnet.TelnetIO) *MenuExecutor {
	return &MenuExecutor{
		db:       db,
		registry: NewCmdKeyRegistry(),
		io:       io,
	}
}

// ExecuteMenu executes a menu by name
func (e *MenuExecutor) ExecuteMenu(menuName string, ctx *ExecutionContext) error {
	// Check for theme file first
	themeFile := e.checkThemeFile(menuName)
	if themeFile != "" {
		return e.serveThemeFile(themeFile)
	}

	menu, err := e.db.GetMenuByName(menuName)
	if err != nil {
		return fmt.Errorf("failed to load menu %s: %w", menuName, err)
	}

	commands, err := e.db.GetMenuCommands(menu.ID)
	if err != nil {
		return fmt.Errorf("failed to load menu commands for %s: %w", menuName, err)
	}

	// Display menu titles
	e.displayTitles(menu)

	// Execute EVERYTIME commands
	// Note: EVERYTIME functionality removed as Flags field was removed

	// Display generic menu if applicable
	e.displayGenericMenu(menu, commands, ctx)

	// Main menu loop
	for {
		// Position prompt at bottom of terminal
		_, height := getTerminalDimensions(ctx)
		e.io.Printf("\033[%d;1H", height)
		e.io.Print(menu.Prompt)

		// Read input (single key for now, can be expanded)
		key, err := e.io.GetKeyPressUpper()
		if err != nil {
			return err
		}

		input := string(key)
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		// Find matching command
		cmd := e.findCommand(commands, input)
		if cmd == nil {
			// Suppress invalid command messages
			continue
		}

		// Check ACS
		if !e.checkACS(cmd.ACSRequired, ctx) {
			e.io.Print("Access denied.\r\n")
			continue
		}

		// Execute command
		if err := e.executeCommand(*cmd, ctx); err != nil {
			// Check if this is a user logout
			if err.Error() == "user_logout" {
				return err
			}
			return err
		}

		// Check for quit/logout
		if cmd.CmdKeys == "G" {
			break
		}
	}

	return nil
}

// displayTitles displays the menu titles
func (e *MenuExecutor) displayTitles(menu *database.Menu) {
	for _, title := range menu.Titles {
		e.io.Printf("%s\r\n", title)
	}
}

// displayGenericMenu displays the generic menu if applicable
func (e *MenuExecutor) displayGenericMenu(menu *database.Menu, commands []database.MenuCommand, ctx *ExecutionContext) {
	// Check ACS for menu access
	if !e.checkACS(menu.ACSRequired, ctx) {
		e.io.Print("Access denied.\r\n")
		return
	}

	// Clear screen if required
	e.clearScreen(menu.ClearScreen)

	// Display titles (centered)
	for _, title := range menu.Titles {
		centeredTitle := e.centerTitle(title, ctx)
		e.io.Printf("%s\r\n", centeredTitle)
	}

	// Add row spacing after titles
	e.io.Printf("\r\n")

	// Display commands in columns with colors
	e.displayCommandsInColumns(commands, menu, ctx)
}

// findCommand finds a command matching the input
func (e *MenuExecutor) findCommand(commands []database.MenuCommand, input string) *database.MenuCommand {
	for _, cmd := range commands {
		if strings.EqualFold(cmd.Keys, input) {
			return &cmd
		}
	}
	return nil
}

// checkACS checks if the user has access based on ACS
func (e *MenuExecutor) checkACS(acs string, ctx *ExecutionContext) bool {
	if acs == "" {
		return true // No ACS requirement means allow access
	}

	// Get user's security level
	userSecLevel := ctx.Session.SecurityLevel

	// Parse ACS string - it can contain multiple conditions separated by operators
	// For now, support basic level comparison (e.g., "10", ">5", "<100")
	acs = strings.TrimSpace(acs)

	// Check for comparison operators
	if strings.HasPrefix(acs, ">=") {
		if level, err := strconv.Atoi(acs[2:]); err == nil {
			return userSecLevel >= level
		}
	} else if strings.HasPrefix(acs, "<=") {
		if level, err := strconv.Atoi(acs[2:]); err == nil {
			return userSecLevel <= level
		}
	} else if strings.HasPrefix(acs, ">") {
		if level, err := strconv.Atoi(acs[1:]); err == nil {
			return userSecLevel > level
		}
	} else if strings.HasPrefix(acs, "<") {
		if level, err := strconv.Atoi(acs[1:]); err == nil {
			return userSecLevel < level
		}
	} else if strings.HasPrefix(acs, "=") || strings.HasPrefix(acs, "==") {
		levelStr := acs
		if strings.HasPrefix(acs, "==") {
			levelStr = acs[2:]
		} else {
			levelStr = acs[1:]
		}
		if level, err := strconv.Atoi(levelStr); err == nil {
			return userSecLevel == level
		}
	} else {
		// Direct level number
		if level, err := strconv.Atoi(acs); err == nil {
			return userSecLevel >= level
		}
	}

	// If we can't parse the ACS, deny access for security
	return false
}

// executeCommand executes a menu command
func (e *MenuExecutor) executeCommand(cmd database.MenuCommand, ctx *ExecutionContext) error {
	return e.registry.Execute(cmd.CmdKeys, ctx, cmd.Options)
}

// checkThemeFile checks if a theme file exists for the given menu name
func (e *MenuExecutor) checkThemeFile(menuName string) string {
	extensions := []string{".ans", ".asc"}
	for _, ext := range extensions {
		themePath := filepath.Join("theme", menuName+ext)
		if _, err := os.Stat(themePath); err == nil {
			return themePath
		}
	}
	return ""
}

// serveThemeFile serves the content of a theme file
func (e *MenuExecutor) serveThemeFile(themePath string) error {
	content, err := os.ReadFile(themePath)
	if err != nil {
		return fmt.Errorf("failed to read theme file %s: %w", themePath, err)
	}
	e.io.Print(string(content))
	return nil
}

// clearScreen clears the screen if ClearScreen is true
func (e *MenuExecutor) clearScreen(clear bool) {
	if clear {
		e.io.Print("\x1b[2J\x1b[H") // ANSI clear screen and move cursor to top-left
	}
}

// centerTitle centers the title text on screen
func (e *MenuExecutor) centerTitle(title string, ctx *ExecutionContext) string {
	width, _ := getTerminalDimensions(ctx)
	if len(title) >= width {
		return title
	}
	padding := (width - len(title)) / 2
	return strings.Repeat(" ", padding) + title
}

// displayCommandsInColumns displays commands in columns with colors
func (e *MenuExecutor) displayCommandsInColumns(commands []database.MenuCommand, menu *database.Menu, ctx *ExecutionContext) {
	if len(commands) == 0 {
		return
	}

	// Filter commands that have short descriptions
	var displayCommands []database.MenuCommand
	for _, cmd := range commands {
		if cmd.ShortDescription != "" {
			displayCommands = append(displayCommands, cmd)
		}
	}

	if len(displayCommands) == 0 {
		return
	}

	// Use 4 columns for the new layout
	columns := 4

	// ANSI color codes
	bracketColor := e.getANSIColor(menu.GenericBracketColor)
	commandColor := e.getANSIColor(menu.GenericCommandColor)
	descColor := e.getANSIColor(menu.GenericDescColor)
	resetColor := "\x1b[0m"

	// Calculate items per column
	itemsPerColumn := (len(displayCommands) + columns - 1) / columns

	for row := 0; row < itemsPerColumn; row++ {
		line := ""
		for col := 0; col < columns; col++ {
			idx := col*itemsPerColumn + row
			if idx < len(displayCommands) {
				cmd := displayCommands[idx]
				// Format: [keys] Short Description
				formatted := fmt.Sprintf("%s[%s%s%s]%s %s%s%s",
					bracketColor, resetColor,
					commandColor, cmd.Keys, resetColor,
					descColor, cmd.ShortDescription, resetColor)

				// Calculate column width with margins (width - 2*2) / 4
				width, _ := getTerminalDimensions(ctx)
				colWidth := (width - 4) / 4
				const margin = 2
				if len(formatted) < colWidth {
					formatted += strings.Repeat(" ", colWidth-len(formatted))
				}
				// Add left margin for first column, right margin for others
				if col == 0 {
					line += strings.Repeat(" ", margin) + formatted
				} else {
					line += formatted + strings.Repeat(" ", margin)
				}
			}
		}
		e.io.Printf("%s\r\n", strings.TrimRight(line, " "))
	}
}

// getANSIColor converts color number to ANSI color code
func (e *MenuExecutor) getANSIColor(colorNum int) string {
	switch colorNum {
	case 0:
		return "\x1b[30m" // Black
	case 1:
		return "\x1b[31m" // Red
	case 2:
		return "\x1b[32m" // Green
	case 3:
		return "\x1b[33m" // Yellow
	case 4:
		return "\x1b[34m" // Blue
	case 5:
		return "\x1b[35m" // Magenta
	case 6:
		return "\x1b[36m" // Cyan
	case 7:
		return "\x1b[37m" // White
	case 8:
		return "\x1b[90m" // Bright Black
	case 9:
		return "\x1b[91m" // Bright Red
	case 10:
		return "\x1b[92m" // Bright Green
	case 11:
		return "\x1b[93m" // Bright Yellow
	case 12:
		return "\x1b[94m" // Bright Blue
	case 13:
		return "\x1b[95m" // Bright Magenta
	case 14:
		return "\x1b[96m" // Bright Cyan
	case 15:
		return "\x1b[97m" // Bright White
	default:
		return "\x1b[37m" // Default to white
	}
}
