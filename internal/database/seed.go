package database

import "fmt"

// SeedDefaultMainMenu seeds the MainMenu with default structure and commands
func SeedDefaultMainMenu(db Database) error {
	// Check if MainMenu already exists
	menu, err := db.GetMenuByName("MainMenu")
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
				PositionNumber:   1,
				Keys:             "G",
				ShortDescription: "Goodbye",
				LongDescription:  "Disconnect from the BBS",
				ACSRequired:      "",
				CmdKeys:          "G",
				Options:          "",
				NodeActivity:     "Logging off.",
				Active:           true,
				Hidden:           false,
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

	// Menus don't exist, create it with commands
	menu = &Menu{

		Name:                "MainMenu",
		Titles:              []string{"|05-= |13Retrograde BBS |05=-", "|07-|06- |14Main Menu |06-|07-"},
		DisplayMode:         DisplayModeTitlesGenerated,
		Prompt:              " |08[ |14M|06ain |14M|06enu |08] |05CMD|13?: ",
		ACSRequired:         "",
		GenericColumns:      3,
		GenericBracketColor: 3,
		GenericCommandColor: 11,
		GenericDescColor:    15,
		ClearScreen:         true,
		LeftBracket:         "[",
		RightBracket:        "]",
		NodeActivity:        "Browsing the main menu.",
	}

	menuID, err := db.CreateMenu(menu)
	if err != nil {
		return fmt.Errorf("failed to create menu: %w", err)
	}

	// Create menu commands
	commands := []MenuCommand{
		{
			MenuID:           int(menuID),
			PositionNumber:   1,
			Keys:             "G",
			ShortDescription: "Goodbye",
			LongDescription:  "Disconnect from the BBS",
			ACSRequired:      "",
			CmdKeys:          "G",
			Options:          "",
			NodeActivity:     "Logging off.",
			Active:           true,
			Hidden:           false,
		},
		{
			MenuID:           int(menuID),
			PositionNumber:   2,
			Keys:             "M",
			ShortDescription: "Message Menu",
			LongDescription:  "View and manage messages",
			ACSRequired:      "",
			CmdKeys:          "M",
			Options:          "",
			NodeActivity:     "Browsing the message menu.",
			Active:           true,
			Hidden:           false,
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
