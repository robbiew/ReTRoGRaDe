package menu

// registerVotingCommands registers all voting system commands
func registerVotingCommands(r *CmdKeyRegistry) {
	defs := []CmdKeyDefinition{
		// Voting
		{CmdKey: "VA", Name: "Add Voting Topic", Description: "Add a new voting topic", Category: "Voting"},
		{CmdKey: "VL", Name: "List Voting Topics", Description: "List available voting topics", Category: "Voting"},
		{CmdKey: "VR", Name: "View Voting Results", Description: "View results for a voting topic", Category: "Voting"},
		{CmdKey: "VT", Name: "Track User Vote", Description: "Track how a user voted", Category: "Voting"},
		{CmdKey: "VU", Name: "View Topic Voters", Description: "View users who voted on a topic", Category: "Voting"},
		{CmdKey: "VV", Name: "Vote on All Topics", Description: "Vote on all un-voted topics", Category: "Voting"},
		{CmdKey: "V#", Name: "Vote on Topic", Description: "Vote on a specific topic number", Category: "Voting"},
	}

	for _, def := range defs {
		d := def
		if d.Handler == nil {
			d.Handler = handleNotImplemented
		}
		r.Register(&d)
	}
}
