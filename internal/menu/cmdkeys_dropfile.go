package menu

// registerDropfileCommands registers all door/dropfile commands
func registerDropfileCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Dropfile / Door Launch
		{CmdKey: "DC", Name: "Create CHAIN.TXT", Description: "Create CHAIN.TXT (WWIV) and execute command", Category: "Dropfile"},
		{CmdKey: "DD", Name: "Create DORINFO1.DEF", Description: "Create DORINFO1.DEF (RBBS) and execute command", Category: "Dropfile"},
		{CmdKey: "DG", Name: "Create DOOR.SYS", Description: "Create DOOR.SYS (GAP) and execute command", Category: "Dropfile"},
		{CmdKey: "DP", Name: "Create PCBOARD.SYS", Description: "Create PCBOARD.SYS and execute command", Category: "Dropfile"},
		{CmdKey: "DS", Name: "Create SFDOORS.DAT", Description: "Create SFDOORS.DAT (Spitfire) and execute command", Category: "Dropfile"},
		{CmdKey: "DW", Name: "Create CALLINFO.BBS", Description: "Create CALLINFO.BBS (Wildcat!) and execute command", Category: "Dropfile"},
		{CmdKey: "D-", Name: "Execute Without Dropfile", Description: "Execute command without creating a dropfile", Category: "Dropfile"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
