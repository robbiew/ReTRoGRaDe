package menu

// registerMessageCommands registers all message-related commands (Message + Message Scanning)
func registerMessageCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Message System
		{CmdKey: "MA", Name: "Change Message Base", Description: "Change to another message base", Category: "Message"},
		{CmdKey: "ME", Name: "Send Private Mail", Description: "Send private mail to a user", Category: "Message"},
		{CmdKey: "MK", Name: "Edit Outgoing Mail", Description: "Edit or delete outgoing private mail", Category: "Message"},
		{CmdKey: "ML", Name: "Send Mass Mail", Description: "Send private mail to multiple users", Category: "Message"},
		{CmdKey: "MM", Name: "Read Private Mail", Description: "Read your private mail", Category: "Message"},
		{CmdKey: "MN", Name: "New Message Scan", Description: "Scan for new messages", Category: "Message"},
		{CmdKey: "MP", Name: "Post Message", Description: "Post a message in the current base", Category: "Message"},
		{CmdKey: "MR", Name: "Read Messages", Description: "Read messages in the current base", Category: "Message"},
		{CmdKey: "MS", Name: "Scan Current Base", Description: "Scan the current message base", Category: "Message"},
		{CmdKey: "MU", Name: "List Base Access", Description: "List users with access to the current base", Category: "Message"},
		{CmdKey: "MY", Name: "Scan for Personal Mail", Description: "Scan message bases for personal messages", Category: "Message"},
		{CmdKey: "MZ", Name: "Set Message NewScan List", Description: "Select message bases to include in new scan", Category: "Message"},
		{CmdKey: "M#", Name: "Quick Message Base Change", Description: "Prompt for a message base to change to", Category: "Message"},

		// Message Scanning (READP.MNU)
		{CmdKey: "RA", Name: "Read Again", Description: "Re-read the current message", Category: "Message Scanning"},
		{CmdKey: "RB", Name: "Back in Thread", Description: "Move backward in the message thread", Category: "Message Scanning"},
		{CmdKey: "RC", Name: "Continuous Reading", Description: "Toggle continuous message reading", Category: "Message Scanning"},
		{CmdKey: "RD", Name: "Delete Message", Description: "Delete the current message", Category: "Message Scanning"},
		{CmdKey: "RE", Name: "Edit Message", Description: "Edit the current message", Category: "Message Scanning"},
		{CmdKey: "RF", Name: "Forward in Thread", Description: "Move forward in the message thread", Category: "Message Scanning"},
		{CmdKey: "RG", Name: "Next Message Base", Description: "Go to the next message base", Category: "Message Scanning"},
		{CmdKey: "RH", Name: "Set High-Read Pointer", Description: "Set the high-read pointer", Category: "Message Scanning"},
		{CmdKey: "RI", Name: "Ignore Remaining Messages", Description: "Ignore remaining messages and set pointer", Category: "Message Scanning"},
		{CmdKey: "RL", Name: "List Messages", Description: "List messages in the current base", Category: "Message Scanning"},
		{CmdKey: "RM", Name: "Move Message", Description: "Move the current message", Category: "Message Scanning"},
		{CmdKey: "RN", Name: "Next Message", Description: "Read the next message", Category: "Message Scanning"},
		{CmdKey: "RQ", Name: "Quit Reading", Description: "Quit the message reader", Category: "Message Scanning"},
		{CmdKey: "RR", Name: "Reply to Message", Description: "Reply to the current message", Category: "Message Scanning"},
		{CmdKey: "RT", Name: "Toggle Base NewScan", Description: "Toggle newscan for the message base", Category: "Message Scanning"},
		{CmdKey: "RU", Name: "Edit Message Author", Description: "Edit the user associated with the message", Category: "Message Scanning"},
		{CmdKey: "RX", Name: "Extract Message", Description: "Extract the message to a file", Category: "Message Scanning"},
		{CmdKey: "R#", Name: "Jump to Message", Description: "Jump directly to a message number", Category: "Message Scanning"},
		{CmdKey: "R-", Name: "Previous Message", Description: "Read the previous message", Category: "Message Scanning"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
