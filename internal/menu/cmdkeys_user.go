package menu

// registerUserCommands registers all user-related commands
func registerUserCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// User / System Operations (O*)
		{CmdKey: "O1", Name: "Logon (Shuttle)", Description: "Log on to the BBS when using the shuttle menu", Category: "User"},
		{CmdKey: "O2", Name: "Apply as New User", Description: "Apply for access using the shuttle menu", Category: "User"},
		{CmdKey: "OA", Name: "Auto-Validate User", Description: "Allow auto-validation with password and level", Category: "User"},
		{CmdKey: "OB", Name: "User Statistics", Description: "View Top 10 user statistics", Category: "User"},
		{CmdKey: "OC", Name: "Page the SysOp", Description: "Page the SysOp or leave a message", Category: "User"},
		{CmdKey: "OE", Name: "Pause Screen", Description: "Toggle or force a pause in output", Category: "User", Implemented: true, Handler: handlePauseScreen},
		{CmdKey: "OF", Name: "Modify AR Flags", Description: "Set, reset, or toggle AR flags", Category: "User"},
		{CmdKey: "OG", Name: "Modify AC Flags", Description: "Set, reset, or toggle AC flags", Category: "User"},
		{CmdKey: "OL", Name: "List Today's Callers", Description: "Display today's caller list", Category: "User"},
		{CmdKey: "ON", Name: "Clear Screen", Description: "Clear the caller's screen", Category: "User"},
		{CmdKey: "OP", Name: "Modify User Information", Description: "Modify specific user information fields", Category: "User"},
		{CmdKey: "OR", Name: "Change Conference", Description: "Switch to a different conference", Category: "User"},
		{CmdKey: "OS", Name: "Bulletins Menu", Description: "Go to the bulletins menu", Category: "User"},
		{CmdKey: "OU", Name: "User Listing", Description: "Display the user listing", Category: "User"},
		{CmdKey: "OV", Name: "BBS Listing", Description: "Display the BBS list", Category: "User"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
