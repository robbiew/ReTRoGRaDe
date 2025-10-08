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

// CmdKeyRegistry holds the registered command key handlers
type CmdKeyRegistry struct {
	handlers map[string]CmdKeyHandler
}

// NewCmdKeyRegistry creates a new command key registry
func NewCmdKeyRegistry() *CmdKeyRegistry {
	r := &CmdKeyRegistry{
		handlers: make(map[string]CmdKeyHandler),
	}
	r.registerDefaults()
	return r
}

// Register registers a handler for a command key
func (r *CmdKeyRegistry) Register(cmdKey string, handler CmdKeyHandler) {
	r.handlers[strings.ToUpper(cmdKey)] = handler
}

// Execute executes a command key with the given context and options
func (r *CmdKeyRegistry) Execute(cmdKey string, ctx *ExecutionContext, options string) error {
	handler, exists := r.handlers[strings.ToUpper(cmdKey)]
	if !exists {
		return fmt.Errorf("unknown command key: %s", cmdKey)
	}
	return handler(ctx, options)
}

// registerDefaults registers the default command key handlers
func (r *CmdKeyRegistry) registerDefaults() {
	r.Register("MM", handleReadMail)
	r.Register("MP", handlePostMessage)
	r.Register("G", handleGoodbye)
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
