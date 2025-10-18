package menu

// registerTransferCommands registers all batch transfer commands
func registerTransferCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Batch Transfer
		{CmdKey: "BC", Name: "Clear Batch Queue", Description: "Clear the batch transfer queue", Category: "Batch Transfer"},
		{CmdKey: "BD", Name: "Download Batch Queue", Description: "Download the batch queue", Category: "Batch Transfer"},
		{CmdKey: "BL", Name: "List Batch Queue", Description: "List files in the batch queue", Category: "Batch Transfer"},
		{CmdKey: "BR", Name: "Remove Batch Item", Description: "Remove a file from the batch queue", Category: "Batch Transfer"},
		{CmdKey: "BU", Name: "Upload Batch Queue", Description: "Upload the batch queue", Category: "Batch Transfer"},
		{CmdKey: "B?", Name: "Batch Queue Count", Description: "Display number of files in the batch queue", Category: "Batch Transfer"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
