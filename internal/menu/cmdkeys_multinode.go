package menu

// registerMultinodeCommands registers all multinode/chat commands
func registerMultinodeCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Multinode
		{CmdKey: "NA", Name: "Toggle Page Availability", Description: "Toggle whether this node can be paged", Category: "Multinode"},
		{CmdKey: "ND", Name: "Hangup Node", Description: "Disconnect another node", Category: "Multinode"},
		{CmdKey: "NG", Name: "Join Group Chat", Description: "Join the multi-node group chat", Category: "Multinode"},
		{CmdKey: "NO", Name: "View All Nodes", Description: "Display users on all nodes", Category: "Multinode"},
		{CmdKey: "NP", Name: "Page Node", Description: "Page another node for chat", Category: "Multinode"},
		{CmdKey: "NS", Name: "Send Node Message", Description: "Send a message to another node", Category: "Multinode"},
		{CmdKey: "NT", Name: "Toggle Stealth Mode", Description: "Toggle stealth mode on or off", Category: "Multinode"},
		{CmdKey: "NW", Name: "Set Activity String", Description: "Display a string under node activity", Category: "Multinode"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
