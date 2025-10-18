package menu

// registerSysopCommands registers all sysop administration commands
func registerSysopCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Sysop Functions
		{CmdKey: "*B", Name: "Edit Message Bases", Description: "Enter the message base editor", Category: "Sysop"},
		{CmdKey: "*C", Name: "Change User Account", Description: "Switch to another user's account", Category: "Sysop"},
		{CmdKey: "*D", Name: "Mini-DOS Shell", Description: "Enter the Mini-DOS environment", Category: "Sysop"},
		{CmdKey: "*E", Name: "Edit Events", Description: "Enter the event editor", Category: "Sysop"},
		{CmdKey: "*F", Name: "Edit File Bases", Description: "Enter the file base editor", Category: "Sysop"},
		{CmdKey: "*L", Name: "View SysOp Log", Description: "Display the SysOp log for a day", Category: "Sysop"},
		{CmdKey: "*N", Name: "Edit Text File", Description: "Edit a text file", Category: "Sysop"},
		{CmdKey: "*P", Name: "System Configuration", Description: "Enter the system configuration editor", Category: "Sysop"},
		{CmdKey: "*R", Name: "Conference Editor", Description: "Enter the conference editor", Category: "Sysop"},
		{CmdKey: "*U", Name: "User Editor", Description: "Enter the user editor", Category: "Sysop"},
		{CmdKey: "*V", Name: "Voting Editor", Description: "Enter the voting editor", Category: "Sysop"},
		{CmdKey: "*X", Name: "Protocol Editor", Description: "Enter the protocol editor", Category: "Sysop"},
		{CmdKey: "*Z", Name: "Activity Log", Description: "Display the system activity log", Category: "Sysop"},
		{CmdKey: "*1", Name: "Edit Files in Base", Description: "Edit files in the current file base", Category: "Sysop"},
		{CmdKey: "*2", Name: "Sort File Bases", Description: "Sort all file bases by name", Category: "Sysop"},
		{CmdKey: "*3", Name: "Read All Private Mail", Description: "Read every user's private mail", Category: "Sysop"},
		{CmdKey: "*4", Name: "Download Any File", Description: "Download any system file (prompt if unknown)", Category: "Sysop"},
		{CmdKey: "*5", Name: "Recheck Files", Description: "Recheck files for size and online status", Category: "Sysop"},
		{CmdKey: "*6", Name: "Upload Missing Files", Description: "Upload files not already listed", Category: "Sysop"},
		{CmdKey: "*7", Name: "Validate Files", Description: "Validate unvalidated files", Category: "Sysop"},
		{CmdKey: "*8", Name: "Add GIF Specs", Description: "Add resolution specs to GIF files", Category: "Sysop"},
		{CmdKey: "*9", Name: "Pack Message Bases", Description: "Pack the message bases", Category: "Sysop"},
		{CmdKey: "*#", Name: "Menu Editor", Description: "Enter the menu editor", Category: "Sysop"},
		{CmdKey: "*$", Name: "Long DOS Directory", Description: "Show long DOS directory of current file base", Category: "Sysop"},
		{CmdKey: "*%", Name: "Short DOS Directory", Description: "Show condensed DOS directory of current file base", Category: "Sysop"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
