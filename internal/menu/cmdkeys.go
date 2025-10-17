package menu

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/logging"
	"github.com/robbiew/retrograde/internal/telnet"
)

// CmdKeyHandler is the function signature for command key handlers
type CmdKeyHandler func(ctx *ExecutionContext, options string) error

// ExecutionContext holds the context for executing a command
type ExecutionContext struct {
	UserID   int64
	Username string
	IO       *telnet.TelnetIO
	Session  *config.TelnetSession
	// Add more context as needed: session, database, etc.
}

// CmdKeyDefinition describes a command key with its metadata
type CmdKeyDefinition struct {
	CmdKey       string // The 2-letter command key (e.g., "MM", "MP", "G")
	Name         string // Human-readable name (e.g., "Read Mail", "Post Message")
	Description  string // Detailed description of what the command does
	Category     string // Category for grouping (e.g., "Message", "File", "System")
	NodeActivity string // Default node activity text shown to other users
	Implemented  bool   // Whether this command is fully implemented
	Handler      CmdKeyHandler
}

// CmdKeyRegistry holds the registered command key handlers
type CmdKeyRegistry struct {
	handlers    map[string]CmdKeyHandler
	definitions map[string]*CmdKeyDefinition
}

// NewCmdKeyRegistry creates a new command key registry
func NewCmdKeyRegistry() *CmdKeyRegistry {
	r := &CmdKeyRegistry{
		handlers:    make(map[string]CmdKeyHandler),
		definitions: make(map[string]*CmdKeyDefinition),
	}
	r.registerDefaults()
	return r
}

// Register registers a handler for a command key with its definition
func (r *CmdKeyRegistry) Register(def *CmdKeyDefinition) {
	key := strings.ToUpper(def.CmdKey)
	def.NodeActivity = computeNodeActivity(def)
	r.handlers[key] = def.Handler
	r.definitions[key] = def
}

// Execute executes a command key with the given context and options
func (r *CmdKeyRegistry) Execute(cmdKey string, ctx *ExecutionContext, options string) error {
	handler, exists := r.handlers[strings.ToUpper(cmdKey)]
	if !exists {
		return fmt.Errorf("unknown command key: %s", cmdKey)
	}
	return handler(ctx, options)
}

// GetDefinition returns the definition for a command key
func (r *CmdKeyRegistry) GetDefinition(cmdKey string) *CmdKeyDefinition {
	return r.definitions[strings.ToUpper(cmdKey)]
}

// GetAllDefinitions returns all registered command key definitions
func (r *CmdKeyRegistry) GetAllDefinitions() []*CmdKeyDefinition {
	defs := make([]*CmdKeyDefinition, 0, len(r.definitions))
	for _, def := range r.definitions {
		defs = append(defs, def)
	}
	return defs
}

// GetDefinitionsByCategory returns all command key definitions for a specific category
func (r *CmdKeyRegistry) GetDefinitionsByCategory(category string) []*CmdKeyDefinition {
	defs := make([]*CmdKeyDefinition, 0)
	for _, def := range r.definitions {
		if def.Category == category {
			defs = append(defs, def)
		}
	}
	return defs
}

// registerDefaults registers the default command key handlers with their definitions
func (r *CmdKeyRegistry) registerDefaults() {
	defs := []CmdKeyDefinition{

		// These are from Chapter 11 of the Renegade BBS Sysop Manual. They are placeholders for now.

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

		// Navigation / Display & Flow
		{CmdKey: "-C", Name: "SysOp Window Message", Description: "Display a message on the SysOp window", Category: "Navigation/Display"},
		{CmdKey: "-F", Name: "Display File (MCI)", Description: "Display a text file (MCI codes enabled)", Category: "Navigation/Display"},
		{CmdKey: "/F", Name: "Display File (Literal)", Description: "Display a text file without MCI expansion", Category: "Navigation/Display"},
		{CmdKey: "-L", Name: "Display Line", Description: "Display a single line of text", Category: "Navigation/Display"},
		{CmdKey: "-N", Name: "Prompt: Yes Shows Quote", Description: "Prompt the user; show quote if they answer Yes", Category: "Navigation/Display"},
		{CmdKey: "-Q", Name: "Read Infoform", Description: "Read an Infoform questionnaire", Category: "Navigation/Display"},
		{CmdKey: "-R", Name: "Read Infoform Answers", Description: "Display answers to an Infoform questionnaire", Category: "Navigation/Display"},
		{CmdKey: "-S", Name: "Append SysOp Log", Description: "Append a line to the SysOp log", Category: "Navigation/Display"},
		{CmdKey: "-Y", Name: "Prompt: No Shows Quote", Description: "Prompt the user; show quote if they answer No", Category: "Navigation/Display"},
		{CmdKey: "-;", Name: "Execute Macro", Description: "Execute a macro string (substitutes ';' with <CR>)", Category: "Navigation/Display"},
		{CmdKey: "-$", Name: "Prompt for Password", Description: "Prompt the user for a password", Category: "Navigation/Display"},
		{CmdKey: "-^", Name: "Go To Menu", Description: "Jump to another menu", Category: "Navigation/Display"},
		{CmdKey: "-/", Name: "Gosub Menu", Description: "Jump to a menu and return", Category: "Navigation/Display"},
		{CmdKey: "-\\", Name: "Return from Menu", Description: "Return to the previous menu", Category: "Navigation/Display"},
		{CmdKey: "-\"", Name: "Return from Menu (Legacy)", Description: "Legacy alias for returning to the previous menu", Category: "Navigation/Display"},

		// Archive Management
		{CmdKey: "AA", Name: "Add to Archive", Description: "Add files to an archive", Category: "Archive"},
		{CmdKey: "AC", Name: "Convert Archive", Description: "Convert between archive formats", Category: "Archive"},
		{CmdKey: "AE", Name: "Extract Archive", Description: "Extract files from an archive", Category: "Archive"},
		{CmdKey: "AG", Name: "Manage Extracted Files", Description: "Manipulate files extracted from archives", Category: "Archive"},
		{CmdKey: "AM", Name: "Modify Archive Comments", Description: "Edit comment fields within an archive", Category: "Archive"},
		{CmdKey: "AR", Name: "Re-Archive", Description: "Re-archive files using the same format", Category: "Archive"},
		{CmdKey: "AT", Name: "Test Archive", Description: "Run an integrity test on an archive", Category: "Archive"},

		// Batch Transfer
		{CmdKey: "BC", Name: "Clear Batch Queue", Description: "Clear the batch transfer queue", Category: "Batch Transfer"},
		{CmdKey: "BD", Name: "Download Batch Queue", Description: "Download the batch queue", Category: "Batch Transfer"},
		{CmdKey: "BL", Name: "List Batch Queue", Description: "List files in the batch queue", Category: "Batch Transfer"},
		{CmdKey: "BR", Name: "Remove Batch Item", Description: "Remove a file from the batch queue", Category: "Batch Transfer"},
		{CmdKey: "BU", Name: "Upload Batch Queue", Description: "Upload the batch queue", Category: "Batch Transfer"},
		{CmdKey: "B?", Name: "Batch Queue Count", Description: "Display number of files in the batch queue", Category: "Batch Transfer"},

		// Dropfile / Door Launch
		{CmdKey: "DC", Name: "Create CHAIN.TXT", Description: "Create CHAIN.TXT (WWIV) and execute command", Category: "Dropfile"},
		{CmdKey: "DD", Name: "Create DORINFO1.DEF", Description: "Create DORINFO1.DEF (RBBS) and execute command", Category: "Dropfile"},
		{CmdKey: "DG", Name: "Create DOOR.SYS", Description: "Create DOOR.SYS (GAP) and execute command", Category: "Dropfile"},
		{CmdKey: "DP", Name: "Create PCBOARD.SYS", Description: "Create PCBOARD.SYS and execute command", Category: "Dropfile"},
		{CmdKey: "DS", Name: "Create SFDOORS.DAT", Description: "Create SFDOORS.DAT (Spitfire) and execute command", Category: "Dropfile"},
		{CmdKey: "DW", Name: "Create CALLINFO.BBS", Description: "Create CALLINFO.BBS (Wildcat!) and execute command", Category: "Dropfile"},
		{CmdKey: "D-", Name: "Execute Without Dropfile", Description: "Execute command without creating a dropfile", Category: "Dropfile"},

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

		// Hangup / Logoff
		{CmdKey: "HC", Name: "Careful Logoff", Description: "Prompt and then log off if confirmed", Category: "Hangup"},
		{CmdKey: "HI", Name: "Immediate Logoff", Description: "Log off immediately", Category: "Hangup"},
		{CmdKey: "HM", Name: "Display & Logoff", Description: "Display a string and log off the user", Category: "Hangup"},

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

		// Multinode
		{CmdKey: "NA", Name: "Toggle Page Availability", Description: "Toggle whether this node can be paged", Category: "Multinode"},
		{CmdKey: "ND", Name: "Hangup Node", Description: "Disconnect another node", Category: "Multinode"},
		{CmdKey: "NG", Name: "Join Group Chat", Description: "Join the multi-node group chat", Category: "Multinode"},
		{CmdKey: "NO", Name: "View All Nodes", Description: "Display users on all nodes", Category: "Multinode"},
		{CmdKey: "NP", Name: "Page Node", Description: "Page another node for chat", Category: "Multinode"},
		{CmdKey: "NS", Name: "Send Node Message", Description: "Send a message to another node", Category: "Multinode"},
		{CmdKey: "NT", Name: "Toggle Stealth Mode", Description: "Toggle stealth mode on or off", Category: "Multinode"},
		{CmdKey: "NW", Name: "Set Activity String", Description: "Display a string under node activity", Category: "Multinode"},

		// User / System Operations (O*)
		{CmdKey: "O1", Name: "Logon (Shuttle)", Description: "Log on to the BBS when using the shuttle menu", Category: "User"},
		{CmdKey: "O2", Name: "Apply as New User", Description: "Apply for access using the shuttle menu", Category: "User"},
		{CmdKey: "OA", Name: "Auto-Validate User", Description: "Allow auto-validation with password and level", Category: "User"},
		{CmdKey: "OB", Name: "User Statistics", Description: "View Top 10 user statistics", Category: "User"},
		{CmdKey: "OC", Name: "Page the SysOp", Description: "Page the SysOp or leave a message", Category: "User"},
		{CmdKey: "OE", Name: "Pause Screen", Description: "Toggle or force a pause in output", Category: "User"},
		{CmdKey: "OF", Name: "Modify AR Flags", Description: "Set, reset, or toggle AR flags", Category: "User"},
		{CmdKey: "OG", Name: "Modify AC Flags", Description: "Set, reset, or toggle AC flags", Category: "User"},
		{CmdKey: "OL", Name: "List Today's Callers", Description: "Display today's caller list", Category: "User"},
		{CmdKey: "ON", Name: "Clear Screen", Description: "Clear the caller's screen", Category: "User"},
		{CmdKey: "OP", Name: "Modify User Information", Description: "Modify specific user information fields", Category: "User"},
		{CmdKey: "OR", Name: "Change Conference", Description: "Switch to a different conference", Category: "User"},
		{CmdKey: "OS", Name: "Bulletins Menu", Description: "Go to the bulletins menu", Category: "User"},
		{CmdKey: "OU", Name: "User Listing", Description: "Display the user listing", Category: "User"},
		{CmdKey: "OV", Name: "BBS Listing", Description: "Display the BBS list", Category: "User"},

		// Automessage
		{CmdKey: "UA", Name: "Reply to Automessage", Description: "Reply to the current automessage author", Category: "Automessage"},
		{CmdKey: "UR", Name: "Display Automessage", Description: "Display the current automessage", Category: "Automessage"},
		{CmdKey: "UW", Name: "Write Automessage", Description: "Write a new automessage", Category: "Automessage"},

		// Voting
		{CmdKey: "VA", Name: "Add Voting Topic", Description: "Add a new voting topic", Category: "Voting"},
		{CmdKey: "VL", Name: "List Voting Topics", Description: "List available voting topics", Category: "Voting"},
		{CmdKey: "VR", Name: "View Voting Results", Description: "View results for a voting topic", Category: "Voting"},
		{CmdKey: "VT", Name: "Track User Vote", Description: "Track how a user voted", Category: "Voting"},
		{CmdKey: "VU", Name: "View Topic Voters", Description: "View users who voted on a topic", Category: "Voting"},
		{CmdKey: "VV", Name: "Vote on All Topics", Description: "Vote on all un-voted topics", Category: "Voting"},
		{CmdKey: "V#", Name: "Vote on Topic", Description: "Vote on a specific topic number", Category: "Voting"},

		// File Scanning (FILEP.MNU)
		{CmdKey: "L1", Name: "Continue Listing", Description: "Continue listing during file scan", Category: "File Scanning"},
		{CmdKey: "L2", Name: "Quit Listing", Description: "Quit listing during file scan", Category: "File Scanning"},
		{CmdKey: "L3", Name: "Next File Base", Description: "Move to the next file base", Category: "File Scanning"},
		{CmdKey: "L4", Name: "Toggle NewScan", Description: "Toggle newscan for the current file base", Category: "File Scanning"},

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

// handleGoodbye handles the G command (logout)
func handleGoodbye(ctx *ExecutionContext, options string) error {
	// Log the logout
	logging.LogLogout(ctx.Session.NodeNumber, ctx.Username, ctx.Session.IPAddress)

	// Display goodbye message
	ctx.IO.Printf("Goodbye, %s!\r\n", ctx.Username)
	time.Sleep(2 * time.Second) // Give time to read message

	// Disconnect the user
	ctx.Session.Connected = false
	if ctx.Session.Conn != nil {
		ctx.Session.Conn.Close()
	}

	return nil
}

// handleNotImplemented is a placeholder for commands not yet implemented
func handleNotImplemented(ctx *ExecutionContext, options string) error {
	ctx.IO.Print("This command is not yet implemented.\r\n")
	return nil
}

func computeNodeActivity(def *CmdKeyDefinition) string {
	if def == nil {
		return ""
	}

	if activity := strings.TrimSpace(def.NodeActivity); activity != "" {
		return ensureSentence(activity)
	}

	if desc := strings.TrimSpace(def.Description); desc != "" {
		return ensureSentence(toPresentProgressive(desc))
	}

	if name := strings.TrimSpace(def.Name); name != "" {
		return ensureSentence(fmt.Sprintf("Using %s", name))
	}

	if def.CmdKey != "" {
		return ensureSentence(fmt.Sprintf("Using %s command", strings.ToUpper(def.CmdKey)))
	}

	return "Performing command."
}

func ensureSentence(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return trimmed
	}
	runes := []rune(trimmed)
	last := runes[len(runes)-1]
	if last == '.' || last == '!' || last == '?' {
		return trimmed
	}
	return trimmed + "."
}

func toPresentProgressive(phrase string) string {
	trimmed := strings.TrimSpace(phrase)
	if trimmed == "" {
		return trimmed
	}

	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return trimmed
	}

	first := words[0]
	if !isAlphabeticWord(first) {
		return trimmed
	}

	lower := strings.ToLower(first)
	gerund := irregularGerunds[lower]
	if gerund == "" {
		gerund = makeGerund(lower)
	}

	switch {
	case isAllUpper(first):
		gerund = strings.ToUpper(gerund)
	case unicode.IsUpper([]rune(first)[0]):
		gerund = capitalize(gerund)
	}

	words[0] = gerund
	return strings.Join(words, " ")
}

func makeGerund(word string) string {
	if word == "" {
		return word
	}
	if strings.HasSuffix(word, "ing") {
		return word
	}
	if strings.HasSuffix(word, "ie") && len(word) > 2 {
		return word[:len(word)-2] + "ying"
	}
	if strings.HasSuffix(word, "e") && len(word) > 1 && word != "be" {
		return word[:len(word)-1] + "ing"
	}
	if isConsonantVowelConsonant(word) {
		return word + word[len(word)-1:] + "ing"
	}
	return word + "ing"
}

func isConsonantVowelConsonant(word string) bool {
	if len(word) < 3 {
		return false
	}
	runes := []rune(word)
	last := runes[len(runes)-1]
	middle := runes[len(runes)-2]
	first := runes[len(runes)-3]
	return !isVowel(last) && shouldDouble(last) && isVowel(middle) && !isVowel(first)
}

func shouldDouble(r rune) bool {
	switch r {
	case 'b', 'd', 'g', 'm', 'n', 'p', 'r', 't':
		return true
	}
	return false
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

func isAlphabeticWord(word string) bool {
	for _, r := range word {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isAllUpper(word string) bool {
	hasLetter := false
	for _, r := range word {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func capitalize(word string) string {
	if word == "" {
		return ""
	}
	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	return string(runes)
}

var irregularGerunds = map[string]string{
	"be":   "being",
	"die":  "dying",
	"see":  "seeing",
	"flee": "fleeing",
	"lie":  "lying",
	"tie":  "tying",
	"quit": "quitting",
}
