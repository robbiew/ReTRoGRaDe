package menu

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/robbiew/retrograde/internal/database"
	"github.com/robbiew/retrograde/internal/telnet"
	"github.com/robbiew/retrograde/internal/ui"
)

// MenuExecutor handles the execution of menus
type MenuExecutor struct {
	db         database.Database
	registry   *CmdKeyRegistry
	io         *telnet.TelnetIO
	currentRow int
}

// NewMenuExecutor creates a new menu executor
func NewMenuExecutor(db database.Database, io *telnet.TelnetIO) *MenuExecutor {
	return &MenuExecutor{
		db:         db,
		registry:   NewCmdKeyRegistry(),
		io:         io,
		currentRow: 1,
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

	// Display generic menu if applicable
	e.displayGenericMenu(menu, commands, ctx)

	// Main menu loop
	for {
		// Position prompt at next available row after menu display
		height := 24
		if ctx != nil && ctx.Session != nil && ctx.Session.Height > 0 {
			height = ctx.Session.Height
		}
		promptRow := min(height, e.currentRow+1)
		e.io.Print(ui.MoveCursorSequence(1, promptRow))
		parsedPrompt := ui.ParsePipeColorCodes(menu.Prompt)
		e.io.Print(parsedPrompt)

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
		parsedTitle := ui.ParsePipeColorCodes(title)
		e.io.Printf("\r\n%s\r\n", parsedTitle)
		e.currentRow += 2
	}
}

// displayGenericMenu displays the generic menu if applicable
func (e *MenuExecutor) displayGenericMenu(menu *database.Menu, commands []database.MenuCommand, ctx *ExecutionContext) {
	// Check ACS for menu access
	if !e.checkACS(menu.ACSRequired, ctx) {
		e.io.Print("Access denied.\r\n")
		e.currentRow += 1
		return
	}

	// Clear screen if required
	e.clearScreen(menu.ClearScreen)
	if menu.ClearScreen {
		e.currentRow = 1 // Reset to top after clear screen
	}

	// Display titles (centered)
	for _, title := range menu.Titles {
		parsedTitle := ui.ParsePipeColorCodes(title)     // Parse pipe codes first
		centeredTitle := e.centerTitle(parsedTitle, ctx) // Then center
		e.io.Printf("\r\n%s\r\n", centeredTitle)
		e.currentRow += 2
	}
	// Add row spacing after titles
	e.io.Printf("\r\n")
	e.currentRow += 1

	// Display commands in columns with colors
	e.displayCommandsInColumns(commands, menu, ctx)
}

// findCommand finds a command matching the input (only active commands)
func (e *MenuExecutor) findCommand(commands []database.MenuCommand, input string) *database.MenuCommand {
	for _, cmd := range commands {
		if strings.EqualFold(cmd.Keys, input) && cmd.Active {
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

	if ctx == nil || ctx.Session == nil {
		return false // Deny access if context or session is nil
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
		e.io.Print(ui.ClearScreenSequence()) // ANSI clear screen and move cursor to top-left
		e.currentRow = 1                     // Reset cursor row after clear screen
	}
}

// centerTitle centers the title text on screen
func (e *MenuExecutor) centerTitle(title string, ctx *ExecutionContext) string {
	width := 80
	if ctx != nil && ctx.Session != nil && ctx.Session.Width > 0 {
		width = ctx.Session.Width
	}
	// Strip both pipe codes and ANSI to get visible length
	visible := ui.StripANSI(ui.StripPipeCodes(title))
	if len(visible) >= width {
		return title
	}
	padding := (width - len(visible)) / 2
	return strings.Repeat(" ", padding) + title
}

// displayCommandsInColumns displays commands in columns with colors
func (e *MenuExecutor) displayCommandsInColumns(commands []database.MenuCommand, menu *database.Menu, ctx *ExecutionContext) {
	if len(commands) == 0 {
		return
	}

	// Filter commands that have short descriptions and are active
	var displayCommands []database.MenuCommand
	for _, cmd := range commands {
		if cmd.ShortDescription != "" && cmd.Active {
			displayCommands = append(displayCommands, cmd)
		}
	}

	if len(displayCommands) == 0 {
		return
	}

	// Use configured columns for the layout
	columns := menu.GenericColumns

	// ANSI color codes
	bracketColor := ui.ColorFromNumber(menu.GenericBracketColor)
	commandColor := ui.ColorFromNumber(menu.GenericCommandColor)
	descColor := ui.ColorFromNumber(menu.GenericDescColor)
	resetColor := ui.Ansi.Reset

	// Calculate items per column
	itemsPerColumn := (len(displayCommands) + columns - 1) / columns
	screenWidth := 80
	if ctx != nil && ctx.Session != nil && ctx.Session.Width > 0 {
		screenWidth = ctx.Session.Width
	}
	const margin = 2
	const interColumnPadding = 2

	// Calculate the maximum column width across all columns
	maxWidth := 0
	for _, cmd := range displayCommands {
		formatted := fmt.Sprintf("%s[%s%s%s%s%s]%s %s%s%s",
			bracketColor, resetColor, commandColor, cmd.Keys, resetColor, bracketColor, resetColor,
			descColor, cmd.ShortDescription, resetColor)
		visibleLen := len(ui.StripANSI(formatted))
		if visibleLen > maxWidth {
			maxWidth = visibleLen
		}
	}
	colWidths := make([]int, columns)
	for i := range colWidths {
		colWidths[i] = maxWidth
	}

	// Calculate total menu width and centering padding
	totalMenuWidth := 0
	for _, w := range colWidths {
		totalMenuWidth += w
	}
	totalMenuWidth += margin + margin + (columns-1)*interColumnPadding
	padding := (screenWidth - totalMenuWidth) / 2
	if padding < 0 {
		padding = 0
	}

	for row := 0; row < itemsPerColumn; row++ {
		line := ""
		for col := 0; col < columns; col++ {
			idx := col*itemsPerColumn + row
			if idx < len(displayCommands) {
				cmd := displayCommands[idx]
				// Format: [keys] Short Description
				formatted := fmt.Sprintf("%s[%s%s%s%s%s]%s %s%s%s",
					bracketColor, resetColor, commandColor, cmd.Keys, resetColor, bracketColor, resetColor,
					descColor, cmd.ShortDescription, resetColor)

				visibleLen := len(ui.StripANSI(formatted))
				if visibleLen < colWidths[col] {
					formatted += strings.Repeat(" ", colWidths[col]-visibleLen)
				}
				// Add left margin for first column, inter-column padding and right margin for others
				if col == 0 {
					line += strings.Repeat(" ", margin) + formatted
				} else {
					line += strings.Repeat(" ", interColumnPadding) + formatted + strings.Repeat(" ", margin)
				}
			}
		}
		// Apply centering padding to the beginning of each row
		line = strings.Repeat(" ", padding) + line
		e.io.Printf("%s\r\n", strings.TrimRight(line, " "))
		e.currentRow += 1 // Each command row adds 1 row
	}
}
