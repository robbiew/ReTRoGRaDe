package menu

// registerNavigationCommands registers all navigation/display commands
func registerNavigationCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Navigation / Display & Flow
		{CmdKey: "-C", Name: "SysOp Window Message", Description: "Display a message on the SysOp window", Category: "Navigation/Display"},
		{CmdKey: "-F", Name: "Display File (MCI)", Description: "Display a text file (MCI codes enabled)", Category: "Navigation/Display"},
		{CmdKey: "/F", Name: "Display File (Literal)", Description: "Display a text file without MCI expansion", Category: "Navigation/Display"},
		{CmdKey: "-L", Name: "Display Line", Description: "Display a single line of text", Category: "Navigation/Display", Implemented: true, Handler: handleDisplayLine},
		{CmdKey: "-N", Name: "Prompt: Yes Shows Quote", Description: "Prompt the user; show quote if they answer Yes", Category: "Navigation/Display"},
		{CmdKey: "-Q", Name: "Read Infoform", Description: "Read an Infoform questionnaire", Category: "Navigation/Display"},
		{CmdKey: "-R", Name: "Read Infoform Answers", Description: "Display answers to an Infoform questionnaire", Category: "Navigation/Display"},
		{CmdKey: "-S", Name: "Append SysOp Log", Description: "Append a line to the SysOp log", Category: "Navigation/Display"},
		{CmdKey: "-Y", Name: "Prompt: No Shows Quote", Description: "Prompt the user; show quote if they answer No", Category: "Navigation/Display"},
		{CmdKey: "-;", Name: "Execute Macro", Description: "Execute a macro string (substitutes ';' with <CR>)", Category: "Navigation/Display"},
		{CmdKey: "-$", Name: "Prompt for Password", Description: "Prompt the user for a password", Category: "Navigation/Display"},
		{CmdKey: "-^", Name: "Go To Menu", Description: "Jump to another menu", Category: "Navigation/Display", Implemented: true, Handler: handleGoToMenu},
		{CmdKey: "-/", Name: "Gosub Menu", Description: "Jump to a menu and return", Category: "Navigation/Display"},
		{CmdKey: "-\\", Name: "Return from Menu", Description: "Return to the previous menu", Category: "Navigation/Display"},
		{CmdKey: "-\"", Name: "Return from Menu (Legacy)", Description: "Legacy alias for returning to the previous menu", Category: "Navigation/Display"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
