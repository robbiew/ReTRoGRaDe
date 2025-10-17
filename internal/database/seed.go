package database

import "fmt"

// SeedDefaultMainMenu seeds the MAIN menu with default structure and commands
func SeedDefaultMainMenu(db Database) error {
	// Check if MAIN menu already exists
	menu, err := db.GetMenuByName("MAIN")
	if err == nil {
		// Menu exists, check if it has commands
		commands, err := db.GetMenuCommands(menu.ID)
		if err != nil {
			return fmt.Errorf("failed to get menu commands: %w", err)
		}
		if len(commands) > 0 {
			// Already has commands, skip
			return nil
		}
		// Add default commands
		defaultCommands := []MenuCommand{
			{
				MenuID:           menu.ID,
				CommandNumber:    3,
				Keys:             "G",
				ShortDescription: "Goodbye",
				ACSRequired:      "",
				CmdKeys:          "G",
				Options:          "",
				Active:           true,
			},
		}
		for _, cmd := range defaultCommands {
			_, err := db.CreateMenuCommand(&cmd)
			if err != nil {
				return fmt.Errorf("failed to create menu command %s: %w", cmd.Keys, err)
			}
		}
		return nil
	}

	// Menu doesn't exist, create it with commands
	menu = &Menu{
		Name:                "MAIN",
		Titles:              []string{"|05-= |13Retrograde BBS |05=-", "|07-|06- |14Main Menu |06-|07-"},
		Prompt:              "|08[ |14M|06ain |14M|06enu |08] |05CMD|13? :",
		ACSRequired:         "",
		GenericColumns:      3,
		GenericBracketColor: 3,
		GenericCommandColor: 11,
		GenericDescColor:    15,
		ClearScreen:         true,
	}

	menuID, err := db.CreateMenu(menu)
	if err != nil {
		return fmt.Errorf("failed to create MAIN menu: %w", err)
	}

	// Create menu commands
	commands := []MenuCommand{
		{
			MenuID:           int(menuID),
			CommandNumber:    3,
			Keys:             "G",
			ShortDescription: "Goodbye",
			ACSRequired:      "",
			CmdKeys:          "G",
			Options:          "",
			Active:           true,
		},
	}

	for _, cmd := range commands {
		_, err := db.CreateMenuCommand(&cmd)
		if err != nil {
			return fmt.Errorf("failed to create menu command %s: %w", cmd.Keys, err)
		}
	}

	return nil
}
