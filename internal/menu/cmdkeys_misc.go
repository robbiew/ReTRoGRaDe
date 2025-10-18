package menu

// registerMiscCommands registers miscellaneous commands (smaller categories)
func registerMiscCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Offline Mail
		{CmdKey: "!D", Name: "Download QWK Packet", Description: "Download offline mail in .QWK format", Category: "Offline Mail"},
		{CmdKey: "!P", Name: "Set Message Pointers", Description: "Set offline message pointers", Category: "Offline Mail"},
		{CmdKey: "!U", Name: "Upload REP Packet", Description: "Upload offline replies in .REP format", Category: "Offline Mail"},

		// Timebank
		{CmdKey: "$D", Name: "Deposit Time", Description: "Deposit time into the timebank", Category: "Timebank"},
		{CmdKey: "$W", Name: "Withdraw Time", Description: "Withdraw time from the timebank", Category: "Timebank"},

		// Credit System
		{CmdKey: "$+", Name: "Increase Credit", Description: "Increase a user's credit balance", Category: "Credit"},
		{CmdKey: "$-", Name: "Increase Debit", Description: "Increase a user's debit balance", Category: "Credit"},

		// Hangup / Logoff
		{CmdKey: "HC", Name: "Careful Logoff", Description: "Prompt and then log off if confirmed", Category: "Hangup"},
		{CmdKey: "HI", Name: "Immediate Logoff", Description: "Log off immediately", Category: "Hangup"},
		{CmdKey: "HM", Name: "Display & Logoff", Description: "Display a string and log off the user", Category: "Hangup"},

		// Automessage
		{CmdKey: "UA", Name: "Reply to Automessage", Description: "Reply to the current automessage author", Category: "Automessage"},
		{CmdKey: "UR", Name: "Display Automessage", Description: "Display the current automessage", Category: "Automessage"},
		{CmdKey: "UW", Name: "Write Automessage", Description: "Write a new automessage", Category: "Automessage"},

		// System
		{CmdKey: "G", Name: "Goodbye / Logoff", Description: "Log off the BBS", Category: "System", NodeActivity: "Logging off.", Implemented: true, Handler: handleGoodbye},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
