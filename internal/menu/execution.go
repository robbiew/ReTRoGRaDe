package menu

import (
	"fmt"
	"strings"

	"github.com/robbiew/retrograde/internal/database"
	"github.com/robbiew/retrograde/internal/telnet"
)

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
	for _, cmd := range commands {
		if strings.Contains(cmd.Flags, "EVERYTIME") {
			if err := e.executeCommand(cmd, ctx); err != nil {
				return err
			}
		}
	}

	// Display generic menu if applicable
	e.displayGenericMenu(menu, commands)

	// Main menu loop
	for {
		// Display prompt
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
			e.io.Print("Invalid command. Try again.\r\n")
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
func (e *MenuExecutor) displayGenericMenu(menu *database.Menu, commands []database.MenuCommand) {
	// Simplified: just list commands
	for _, cmd := range commands {
		if cmd.ShortDescription != "" {
			e.io.Printf("%s - %s\r\n", cmd.Keys, cmd.ShortDescription)
		}
	}
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
	// Simplified: always allow for now
	return true
}

// executeCommand executes a menu command
func (e *MenuExecutor) executeCommand(cmd database.MenuCommand, ctx *ExecutionContext) error {
	return e.registry.Execute(cmd.CmdKeys, ctx, cmd.Options)
}
