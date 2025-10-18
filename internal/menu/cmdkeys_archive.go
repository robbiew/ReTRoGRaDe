package menu

// registerArchiveCommands registers all archive management commands
func registerArchiveCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Archive Management
		{CmdKey: "AA", Name: "Add to Archive", Description: "Add files to an archive", Category: "Archive"},
		{CmdKey: "AC", Name: "Convert Archive", Description: "Convert between archive formats", Category: "Archive"},
		{CmdKey: "AE", Name: "Extract Archive", Description: "Extract files from an archive", Category: "Archive"},
		{CmdKey: "AG", Name: "Manage Extracted Files", Description: "Manipulate files extracted from archives", Category: "Archive"},
		{CmdKey: "AM", Name: "Modify Archive Comments", Description: "Edit comment fields within an archive", Category: "Archive"},
		{CmdKey: "AR", Name: "Re-Archive", Description: "Re-archive files using the same format", Category: "Archive"},
		{CmdKey: "AT", Name: "Test Archive", Description: "Run an integrity test on an archive", Category: "Archive"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
