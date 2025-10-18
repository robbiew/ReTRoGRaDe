package menu

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/logging"
	"github.com/robbiew/retrograde/internal/telnet"
	"github.com/robbiew/retrograde/internal/ui"
)

// CmdKeyHandler is the function signature for command key handlers
type CmdKeyHandler func(ctx *ExecutionContext, options string) error

// ExecutionContext holds the context for executing a command
type ExecutionContext struct {
	UserID      int64
	Username    string
	IO          *telnet.TelnetIO
	Session     *config.TelnetSession
	Executor    *MenuExecutor
	AdvanceRows func(lines int)
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
	// Register commands from all categories
	registerSysopCommands(r)
	registerMessageCommands(r)
	registerNavigationCommands(r)
	registerUserCommands(r)
	registerFileCommands(r)
	registerMultinodeCommands(r)
	registerArchiveCommands(r)
	registerVotingCommands(r)
	registerDropfileCommands(r)
	registerTransferCommands(r)
	registerMiscCommands(r)
}

func handleGoToMenu(ctx *ExecutionContext, options string) error {
	if ctx == nil {
		return fmt.Errorf("go to menu command requires an execution context")
	}

	target := strings.TrimSpace(options)
	if target == "" {
		return fmt.Errorf("go to menu command requires a target menu name in options")
	}

	if ctx.Executor == nil {
		return fmt.Errorf("go to menu command cannot run without a menu executor")
	}

	return ctx.Executor.ExecuteMenu(target, ctx)
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

// handleDisplayLine renders a single line of text honouring pipe color codes.
func handleDisplayLine(ctx *ExecutionContext, options string) error {
	if ctx == nil || ctx.IO == nil {
		return fmt.Errorf("display line command requires an execution context with IO")
	}

	normalized := strings.ReplaceAll(options, "~", "\r\n")
	colored := ui.ParsePipeColorCodes(normalized)

	var out strings.Builder
	out.WriteString("\r\n") // move below the menu prompt
	if colored != "" {
		out.WriteString(colored)
	}
	out.WriteString("\r\n")

	result := out.String()
	if err := ctx.IO.Print(result); err != nil {
		return err
	}

	if ctx.AdvanceRows != nil {
		lines := strings.Count(result, "\n")
		if lines == 0 {
			lines = 1
		}
		ctx.AdvanceRows(lines)
	}

	return nil
}

// handlePauseScreen renders a pause prompt, allowing optional custom text.
func handlePauseScreen(ctx *ExecutionContext, options string) error {
	if ctx == nil || ctx.IO == nil {
		return fmt.Errorf("pause screen command requires an execution context with IO")
	}

	width := 0
	if ctx.Session != nil && ctx.Session.Width > 0 {
		width = ctx.Session.Width
	}

	if err := ui.PauseWithText(ctx.IO, options, width); err != nil {
		return err
	}

	if ctx.AdvanceRows != nil {
		ctx.AdvanceRows(2)
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
