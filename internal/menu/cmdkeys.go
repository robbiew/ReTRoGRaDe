package menu

import (
	"fmt"
	"strings"
	"time"

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
	CmdKey      string // The 2-letter command key (e.g., "MM", "MP", "G")
	Name        string // Human-readable name (e.g., "Read Mail", "Post Message")
	Description string // Detailed description of what the command does
	Category    string // Category for grouping (e.g., "Message", "File", "System")
	Implemented bool   // Whether this command is fully implemented
	Handler     CmdKeyHandler
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
	// Offline Mail Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "!D",
		Name:        "Download QWK Packet",
		Description: "Download offline mail in .QWK",
		Category:    "Offline Mail",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "!P",
		Name:        "Set Message Pointers",
		Description: "Set pointers for offline mail",
		Category:    "Offline Mail",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "!U",
		Name:        "Upload REP Packet",
		Description: "Upload offline mail in .REP",
		Category:    "Offline Mail",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// Message System Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MA",
		Name:        "Change Message Base",
		Description: "Change to a different message base",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "ME",
		Name:        "Enter Message",
		Description: "Enter a new message to current base",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MK",
		Name:        "Kill Message Scan",
		Description: "Stop scanning messages",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "ML",
		Name:        "List Messages",
		Description: "List message titles in current base",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MM",
		Name:        "Read Mail",
		Description: "Read messages addressed to you",
		Category:    "Message",
		Implemented: false,
		Handler:     handleReadMail,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MN",
		Name:        "New Message Scan",
		Description: "Scan all new messages",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MP",
		Name:        "Post Message",
		Description: "Post a new message to current base",
		Category:    "Message",
		Implemented: false,
		Handler:     handlePostMessage,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MR",
		Name:        "Read Messages",
		Description: "Read messages in current base",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MS",
		Name:        "Scan Messages",
		Description: "Quick scan of message subjects",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MU",
		Name:        "Your Messages",
		Description: "Read messages you've posted",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MY",
		Name:        "Your Mail Scan",
		Description: "Scan your personal mail",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "MZ",
		Name:        "Global New Scan",
		Description: "Scan new messages across all bases",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "M#",
		Name:        "Read Message Number",
		Description: "Read a specific message by number",
		Category:    "Message",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// File System Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FA",
		Name:        "Change File Base",
		Description: "Change to a different file area",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FB",
		Name:        "Batch Download",
		Description: "Add files to batch download queue",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FD",
		Name:        "Download File",
		Description: "Download a file from current area",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FF",
		Name:        "Find File",
		Description: "Search for files across all areas",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FL",
		Name:        "List Files",
		Description: "List files in current area",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FN",
		Name:        "New Files Scan",
		Description: "Scan for new files",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FP",
		Name:        "File Points",
		Description: "Display your file points/ratio",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FS",
		Name:        "Search Files",
		Description: "Search for files by keyword",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FU",
		Name:        "Upload File",
		Description: "Upload a file to current area",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FV",
		Name:        "View File",
		Description: "View a file's description or contents",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "FZ",
		Name:        "Global File Search",
		Description: "Search all file areas",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "F@",
		Name:        "Your Uploads",
		Description: "List files you've uploaded",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "F#",
		Name:        "Contents",
		Description: "View archive contents",
		Category:    "File",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// Batch Transfer Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "BC",
		Name:        "Clear Batch Queue",
		Description: "Clear your batch transfer queue",
		Category:    "Batch",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "BD",
		Name:        "Batch Download",
		Description: "Download all files in batch queue",
		Category:    "Batch",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "BL",
		Name:        "List Batch Queue",
		Description: "List files in your batch queue",
		Category:    "Batch",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "BR",
		Name:        "Remove from Batch",
		Description: "Remove a file from batch queue",
		Category:    "Batch",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "BU",
		Name:        "Batch Upload",
		Description: "Upload multiple files at once",
		Category:    "Batch",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "B?",
		Name:        "Batch Queue Status",
		Description: "Display number of files in batch queue",
		Category:    "Batch",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// System/User Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "G",
		Name:        "Goodbye / Logoff",
		Description: "Log off the BBS",
		Category:    "System",
		Implemented: true,
		Handler:     handleGoodbye,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "HC",
		Name:        "Careful Logoff",
		Description: "Prompt before logging off",
		Category:    "System",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "HI",
		Name:        "Instant Logoff",
		Description: "Immediate logoff without prompt",
		Category:    "System",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "HM",
		Name:        "Main Menu",
		Description: "Return to main menu",
		Category:    "System",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// User Information Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "O1",
		Name:        "Logon to BBS",
		Description: "Login (shuttle mode)",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OA",
		Name:        "Apply for Access",
		Description: "New user application",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OB",
		Name:        "Bulletins",
		Description: "Read system bulletins",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OC",
		Name:        "Page Sysop",
		Description: "Page the system operator",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OE",
		Name:        "User Editor",
		Description: "Edit your user settings",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OF",
		Name:        "Feedback",
		Description: "Send feedback to sysop",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OG",
		Name:        "Goodbye Script",
		Description: "Display goodbye and logoff",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OL",
		Name:        "Last Callers",
		Description: "View list of recent callers",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "ON",
		Name:        "Node List",
		Description: "View active nodes/users",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OP",
		Name:        "Page User",
		Description: "Page another user",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OR",
		Name:        "Your Stats",
		Description: "View your usage statistics",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OS",
		Name:        "System Information",
		Description: "View BBS system information",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OU",
		Name:        "User List",
		Description: "View list of users",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "OV",
		Name:        "Version Info",
		Description: "View BBS software version",
		Category:    "User",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// Voting Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "VA",
		Name:        "Add Voting Question",
		Description: "Add a new voting question",
		Category:    "Voting",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "VL",
		Name:        "List Voting Questions",
		Description: "List all voting questions",
		Category:    "Voting",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "VR",
		Name:        "View Voting Results",
		Description: "View results of voting questions",
		Category:    "Voting",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "VV",
		Name:        "Vote on All",
		Description: "Vote on all unvoted questions",
		Category:    "Voting",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "V#",
		Name:        "Vote on Question",
		Description: "Vote on a specific question",
		Category:    "Voting",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// Menu Navigation Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "-^",
		Name:        "Go to Menu",
		Description: "Navigate to a different menu",
		Category:    "Navigation",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "-/",
		Name:        "Gosub Menu",
		Description: "Go to menu and return",
		Category:    "Navigation",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "-\"",
		Name:        "Return from Menu",
		Description: "Return to previous menu",
		Category:    "Navigation",
		Implemented: false,
		Handler:     handleNotImplemented,
	})

	// Miscellaneous Commands
	r.Register(&CmdKeyDefinition{
		CmdKey:      "-!",
		Name:        "Execute Program",
		Description: "Execute an external program",
		Category:    "Misc",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "-&",
		Name:        "Display File",
		Description: "Display a text file",
		Category:    "Misc",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "-%",
		Name:        "Display String",
		Description: "Display a text string",
		Category:    "Misc",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
	r.Register(&CmdKeyDefinition{
		CmdKey:      "-$",
		Name:        "Prompt for Password",
		Description: "Prompt user for password",
		Category:    "Misc",
		Implemented: false,
		Handler:     handleNotImplemented,
	})
}

// handleReadMail handles the MM command (read mail)
func handleReadMail(ctx *ExecutionContext, options string) error {
	// TODO: Implement mail reading logic
	ctx.IO.Printf("User %s is reading mail\r\n", ctx.Username)
	return nil
}

// handlePostMessage handles the MP command (post message)
func handlePostMessage(ctx *ExecutionContext, options string) error {
	// TODO: Implement message posting logic
	ctx.IO.Printf("User %s is posting a message\r\n", ctx.Username)
	return nil
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
