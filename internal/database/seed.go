package database

import (
	"fmt"
	"strings"
)

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
		if err := seedDefaultMsgMenu(db); err != nil {
			return err
		}
		return seedDefaultMessageStructure(db)
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

	if err := seedDefaultMsgMenu(db); err != nil {
		return err
	}
	return seedDefaultMessageStructure(db)
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

func seedDefaultMessageStructure(db Database) error {
	const conferenceName = "Local Areas"
	const areaName = "General Chatter"

	conferenceID, err := ensureConference(db, conferenceName)
	if err != nil {
		return err
	}

	if err := ensureMessageArea(db, areaName, conferenceID, conferenceName); err != nil {
		return err
	}

	return nil
}

func ensureConference(db Database, name string) (int, error) {
	conferences, err := db.GetAllConferences()
	if err != nil {
		return 0, fmt.Errorf("failed to get conferences: %w", err)
	}

	for _, conf := range conferences {
		if strings.EqualFold(conf.Name, name) {
			return conf.ID, nil
		}
	}

	conf := &Conference{
		Name:     name,
		SecLevel: "public",
		Tagline:  "",
		Hidden:   false,
	}

	id, err := db.CreateConference(conf)
	if err != nil {
		return 0, fmt.Errorf("failed to create conference %s: %w", name, err)
	}

	return int(id), nil
}

func ensureMessageArea(db Database, name string, conferenceID int, conferenceName string) error {
	areas, err := db.GetAllMessageAreas()
	if err != nil {
		return fmt.Errorf("failed to get message areas: %w", err)
	}

	for _, area := range areas {
		if strings.EqualFold(area.Name, name) {
			if area.ConferenceID != conferenceID {
				area.ConferenceID = conferenceID
				area.ConferenceName = conferenceName
				if err := db.UpdateMessageArea(&area); err != nil {
					return fmt.Errorf("failed to update existing message area %s: %w", name, err)
				}
			}
			return nil
		}
	}

	basePath := "messages"
	if value, err := db.GetConfig("Configuration.Paths", "", "Message_Base"); err == nil {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			basePath = trimmed
		}
	}
	basePath = strings.TrimRight(basePath, "/\\")

	areaPath := basePath
	if areaPath == "" {
		areaPath = "messages"
	}
	areaPath = areaPath + "/general"

	area := &MessageArea{
		Name:           name,
		File:           "general",
		Path:           areaPath,
		ReadSecLevel:   "public",
		WriteSecLevel:  "public",
		AreaType:       "local",
		EchoTag:        "",
		RealNames:      false,
		Address:        "0:0/0 - Local",
		ConferenceID:   conferenceID,
		ConferenceName: conferenceName,
	}

	if _, err := db.CreateMessageArea(area); err != nil {
		return fmt.Errorf("failed to create message area %s: %w", name, err)
	}

	return nil
}
