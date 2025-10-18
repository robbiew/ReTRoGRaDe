package menu

// registerFileCommands registers all file-related commands (File + File Scanning)
func registerFileCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// File System
		{CmdKey: "FA", Name: "Change File Base", Description: "Change to a different file base", Category: "File"},
		{CmdKey: "FB", Name: "Add to Batch Download", Description: "Add a file to the batch download list", Category: "File"},
		{CmdKey: "FD", Name: "Download File", Description: "Download a file from the BBS", Category: "File"},
		{CmdKey: "FF", Name: "Search Descriptions", Description: "Search all file bases for a description", Category: "File"},
		{CmdKey: "FL", Name: "List Filespec", Description: "List a filespec in the current file base", Category: "File"},
		{CmdKey: "FN", Name: "New File Scan", Description: "Scan file bases for new files", Category: "File"},
		{CmdKey: "FP", Name: "Set File Pointer Date", Description: "Change the pointer date used for new files", Category: "File"},
		{CmdKey: "FS", Name: "Search Filespec", Description: "Search file bases for a filespec", Category: "File"},
		{CmdKey: "FU", Name: "Upload File", Description: "Upload a file to the BBS", Category: "File"},
		{CmdKey: "FV", Name: "View Archive Contents", Description: "List contents of an archive file", Category: "File"},
		{CmdKey: "FZ", Name: "Set File NewScan List", Description: "Select file bases to include in new scan", Category: "File"},
		{CmdKey: "F@", Name: "Create Temporary Base", Description: "Create a temporary file base", Category: "File"},
		{CmdKey: "F#", Name: "Quick File Base Change", Description: "Prompt for a file base to change to", Category: "File"},

		// File Scanning (FILEP.MNU)
		{CmdKey: "L1", Name: "Continue Listing", Description: "Continue listing during file scan", Category: "File Scanning"},
		{CmdKey: "L2", Name: "Quit Listing", Description: "Quit listing during file scan", Category: "File Scanning"},
		{CmdKey: "L3", Name: "Next File Base", Description: "Move to the next file base", Category: "File Scanning"},
		{CmdKey: "L4", Name: "Toggle NewScan", Description: "Toggle newscan for the current file base", Category: "File Scanning"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
