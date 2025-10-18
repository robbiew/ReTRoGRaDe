package database

import "fmt"

// SeedDefaultMainMenu seeds the MainMenu with default structure and commands,
// and ensures related default menus exist.
func SeedDefaultMainMenu(db Database) error {
	menu, err := db.GetMenuByName("MainMenu")
	if err == nil {
		commands, err := db.GetMenuCommands(menu.ID)
		if err != nil {
			return fmt.Errorf("failed to get menu commands: %w", err)
		}
		if len(commands) == 0 {
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
				{
					MenuID:           menu.ID,
					PositionNumber:   2,
					Keys:             "M",
					ShortDescription: "Message Menu",
					LongDescription:  "View and manage messages",
					ACSRequired:      "",
					CmdKeys:          "-^",
					Options:          "MsgMenu",
					NodeActivity:     "Browsing the message menu.",
					Active:           true,
					Hidden:           false,
				},
			}
			for _, cmd := range defaultCommands {
				cmd := cmd
				if _, err := db.CreateMenuCommand(&cmd); err != nil {
					return fmt.Errorf("failed to create menu command %s: %w", cmd.Keys, err)
				}
			}
		}
		return seedDefaultMsgMenu(db)
	}

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
			CmdKeys:          "-^",
			Options:          "MsgMenu",
			NodeActivity:     "Browsing the message menu.",
			Active:           true,
			Hidden:           false,
		},
	}

	for _, cmd := range commands {
		cmd := cmd
		if _, err := db.CreateMenuCommand(&cmd); err != nil {
			return fmt.Errorf("failed to create menu command %s: %w", cmd.Keys, err)
		}
	}

	return seedDefaultMsgMenu(db)
}

func seedDefaultMsgMenu(db Database) error {
	menu, err := db.GetMenuByName("MsgMenu")
	if err == nil {
		commands, err := db.GetMenuCommands(menu.ID)
		if err != nil {
			return fmt.Errorf("failed to get MsgMenu commands: %w", err)
		}
		if len(commands) == 0 {
			defaultCommands := []MenuCommand{
				{
					MenuID:           menu.ID,
					PositionNumber:   1,
					Keys:             "Q",
					ShortDescription: "Return to Main Menu",
					LongDescription:  "Return to the main menu",
					ACSRequired:      "",
					CmdKeys:          "-^",
					Options:          "MainMenu",
					NodeActivity:     "Returning to the main menu.",
					Active:           true,
					Hidden:           false,
				},
				{
					MenuID:           menu.ID,
					PositionNumber:   2,
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
				cmd := cmd
				if _, err := db.CreateMenuCommand(&cmd); err != nil {
					return fmt.Errorf("failed to create MsgMenu command %s: %w", cmd.Keys, err)
				}
			}
		}
		return nil
	}

	menu = &Menu{
		Name:                "MsgMenu",
		Titles:              []string{"|05-= |13Retrograde BBS |05=-", "|07-|06- |14Message Menu |06-|07-"},
		DisplayMode:         DisplayModeTitlesGenerated,
		Prompt:              " |08[ |14M|06essage |14M|06enu |08] |05CMD|13?: ",
		ACSRequired:         "",
		GenericColumns:      2,
		GenericBracketColor: 3,
		GenericCommandColor: 11,
		GenericDescColor:    15,
		ClearScreen:         true,
		LeftBracket:         "[",
		RightBracket:        "]",
		NodeActivity:        "Browsing the message menu.",
	}

	menuID, err := db.CreateMenu(menu)
	if err != nil {
		return fmt.Errorf("failed to create MsgMenu: %w", err)
	}

	commands := []MenuCommand{
		{
			MenuID:           int(menuID),
			PositionNumber:   1,
			Keys:             "Q",
			ShortDescription: "Return to Main Menu",
			LongDescription:  "Return to the main menu",
			ACSRequired:      "",
			CmdKeys:          "-^",
			Options:          "MainMenu",
			NodeActivity:     "Returning to the main menu.",
			Active:           true,
			Hidden:           false,
		},
		{
			MenuID:           int(menuID),
			PositionNumber:   2,
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

	for _, cmd := range commands {
		cmd := cmd
		if _, err := db.CreateMenuCommand(&cmd); err != nil {
			return fmt.Errorf("failed to create MsgMenu command %s: %w", cmd.Keys, err)
		}
	}

	return nil
}
