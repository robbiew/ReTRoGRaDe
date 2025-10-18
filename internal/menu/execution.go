package menu

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/robbiew/retrograde/internal/database"
	"github.com/robbiew/retrograde/internal/telnet"
	"github.com/robbiew/retrograde/internal/ui"
)

// MenuExecutor handles the execution of menus
type MenuExecutor struct {
	db         database.Database
	registry   *CmdKeyRegistry
	io         *telnet.TelnetIO
	currentRow int
}

// NewMenuExecutor creates a new menu executor
func NewMenuExecutor(db database.Database, io *telnet.TelnetIO) *MenuExecutor {
	return &MenuExecutor{
		db:         db,
		registry:   NewCmdKeyRegistry(),
		io:         io,
		currentRow: 1,
	}
}

// ExecuteMenu executes a menu by name

var errNoKeyTimeout = errors.New("menu_no_key_timeout")

var specialKeyLiterals = func() map[string]struct{} {
	keys := map[string]struct{}{
		"FIRSTCMD": {},
		"ANYKEY":   {},
		"NOKEY":    {},
		"ENTER":    {},
		"ESC":      {},
		"TAB":      {},
	}
	return keys
}()

func (e *MenuExecutor) ExecuteMenu(menuName string, ctx *ExecutionContext) error {
	if ctx == nil {
		ctx = &ExecutionContext{}
	}
	ctx.Executor = e
	e.currentRow = 1
	if ctx.IO == nil {
		ctx.IO = e.io
	}
	if ctx.Session == nil && ctx.IO != nil {
		ctx.Session = ctx.IO.Session
	}
	if ctx.AdvanceRows == nil {
		ctx.AdvanceRows = func(lines int) {
			if lines <= 0 {
				return
			}
			e.currentRow += lines
		}
	}

	menu, err := e.lookupMenuByName(menuName)
	if err != nil {
		return err
	}

	commands, err := e.db.GetMenuCommands(menu.ID)
	if err != nil {
		return fmt.Errorf("failed to load menu commands for %s: %w", menuName, err)
	}

	// Execute FIRSTCMD commands before displaying the menu
	firstCommands := e.findCommands(commands, "FIRSTCMD")
	if len(firstCommands) > 0 {
		exitMenu, execErr := e.runCommands(firstCommands, ctx)
		if execErr != nil {
			return execErr
		}
		if exitMenu {
			return nil
		}
	}

	// Pre-calculate special command groups
	anyKeyCommands := e.findCommands(commands, "ANYKEY")
	noKeyCommands := e.findCommands(commands, "NOKEY")
	noKeyTimeout := e.resolveNoKeyTimeout(noKeyCommands)

	// Display generic menu if applicable
	e.displayGenericMenu(menu, commands, ctx)

	// Main menu loop
	for {
		// Position prompt at next available row after menu display
		height := 24
		if ctx != nil && ctx.Session != nil && ctx.Session.Height > 0 {
			height = ctx.Session.Height
		}
		promptRow := min(height, e.currentRow+1)
		e.io.Print(ui.MoveCursorSequence(1, promptRow))
		parsedPrompt := ui.ParsePipeColorCodes(menu.Prompt)
		e.io.Print(parsedPrompt)

		// Read input (single key press with optional timeout)
		input, err := e.readKeyPress(noKeyCommands, noKeyTimeout)
		if err != nil {
			// Timeout reached - execute NOKEY commands
			if errors.Is(err, errNoKeyTimeout) {
				exitMenu, execErr := e.runCommands(noKeyCommands, ctx)
				if execErr != nil {
					return execErr
				}
				if exitMenu {
					return nil
				}
				continue
			}
			return err
		}

		if input == "" {
			continue
		}

		// Find matching commands (supports linked commands with same key)
		matchingCommands := e.findCommands(commands, input)
		if len(anyKeyCommands) > 0 {
			matchingCommands = append(matchingCommands, anyKeyCommands...)
		}

		if len(matchingCommands) == 0 {
			// Suppress invalid command messages
			continue
		}

		exitMenu, execErr := e.runCommands(matchingCommands, ctx)
		if execErr != nil {
			return execErr
		}
		if exitMenu {
			break
		}
	}

	return nil
}

// displayGenericMenu displays the generic menu if applicable
func (e *MenuExecutor) displayGenericMenu(menu *database.Menu, commands []database.MenuCommand, ctx *ExecutionContext) {
	// Check ACS for menu access
	if !e.checkACS(menu.ACSRequired, ctx) {
		e.io.Print("Access denied.\r\n")
		e.currentRow += 1
		return
	}

	// Clear screen if required
	e.clearScreen(menu.ClearScreen)
	if menu.ClearScreen {
		e.currentRow = 1 // Reset to top after clear screen
	}

	displayMode := sanitizeDisplayMode(menu.DisplayMode)
	headerDisplayed := false

	switch displayMode {
	case database.DisplayModeThemeOnly:
		if art := e.findThemeFile(menu.Name); art != "" {
			if lines, err := e.serveThemeFile(art); err == nil && lines > 0 {
				headerDisplayed = true
			}
		}
		if !headerDisplayed {
			displayMode = database.DisplayModeTitlesGenerated
		}
	case database.DisplayModeHeaderGenerated:
		if art := e.findThemeFile(menu.Name + ".hdr"); art != "" {
			if lines, err := e.serveThemeFile(art); err == nil && lines > 0 {
				headerDisplayed = true
			}
		}
		if !headerDisplayed {
			displayMode = database.DisplayModeTitlesGenerated
		}
	}

	if displayMode == database.DisplayModeTitlesGenerated {
		headerDisplayed = e.displayCenteredTitles(menu, ctx) || headerDisplayed
	}

	if headerDisplayed && displayMode != database.DisplayModeThemeOnly {
		e.io.Printf("\r\n")
		e.currentRow += 1
	}

	menu.DisplayMode = displayMode

	if displayMode == database.DisplayModeThemeOnly {
		return
	}

	// Display commands in columns with colors
	e.displayCommandsInColumns(commands, menu, ctx)
}

func sanitizeDisplayMode(value string) string {
	switch value {
	case database.DisplayModeHeaderGenerated, database.DisplayModeThemeOnly:
		return value
	case database.DisplayModeTitlesGenerated:
		return value
	default:
		return database.DisplayModeTitlesGenerated
	}
}

func (e *MenuExecutor) displayCenteredTitles(menu *database.Menu, ctx *ExecutionContext) bool {
	if len(menu.Titles) == 0 {
		return false
	}
	for _, title := range menu.Titles {
		parsedTitle := ui.ParsePipeColorCodes(title)
		centeredTitle := e.centerTitle(parsedTitle, ctx)
		e.io.Printf("\r\n%s\r\n", centeredTitle)
		e.currentRow += 2
	}
	return true
}

func (e *MenuExecutor) findThemeFile(base string) string {
	themeDir := e.getThemeBaseDir()

	candidates := []string{}
	for _, ext := range []string{".ans", ".asc"} {
		name := base + ext
		candidates = append(candidates,
			filepath.Join(themeDir, name),
			filepath.Join(themeDir, strings.ToLower(name)),
			filepath.Join(themeDir, strings.ToUpper(name)))
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	entries, err := os.ReadDir(themeDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		for _, ext := range []string{".ans", ".asc"} {
			if strings.EqualFold(name, base+ext) {
				return filepath.Join(themeDir, name)
			}
		}
	}
	return ""
}

// findCommands returns all active commands matching the input (supports linked commands)
func (e *MenuExecutor) findCommands(commands []database.MenuCommand, input string) []database.MenuCommand {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	var matches []database.MenuCommand
	for _, cmd := range commands {
		if !cmd.Active {
			continue
		}
		if matchesMenuKey(cmd.Keys, input) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

func matchesMenuKey(keyDef, input string) bool {
	keyDef = strings.TrimSpace(keyDef)
	input = strings.TrimSpace(input)
	if keyDef == "" || input == "" {
		return false
	}

	tokens := splitMenuKeyVariants(keyDef)
	if len(tokens) == 0 {
		tokens = []string{keyDef}
	}

	inputRunes := []rune(input)
	inputLen := len(inputRunes)

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if strings.EqualFold(token, input) {
			return true
		}

		upperToken := strings.ToUpper(token)
		if _, special := specialKeyLiterals[upperToken]; special {
			continue
		}

		tokenRunes := []rune(token)
		if len(tokenRunes) > 1 && inputLen == 1 {
			for _, r := range tokenRunes {
				if strings.EqualFold(string(r), input) {
					return true
				}
			}
		}
	}
	return false
}

func splitMenuKeyVariants(keyDef string) []string {
	return strings.FieldsFunc(keyDef, func(r rune) bool {
		switch r {
		case ',', ';', '|', '/', '+':
			return true
		default:
			return unicode.IsSpace(r)
		}
	})
}

// runCommands executes the provided commands sequentially.
// Returns true if menu execution should exit.
func (e *MenuExecutor) runCommands(commands []database.MenuCommand, ctx *ExecutionContext) (bool, error) {
	for _, cmd := range commands {
		// Check ACS before executing each command
		if !e.checkACS(cmd.ACSRequired, ctx) {
			e.io.Print("Access denied.\r\n")
			continue
		}

		if err := e.executeCommand(cmd, ctx); err != nil {
			if err.Error() == "user_logout" {
				return true, err
			}
			return false, err
		}

		if strings.EqualFold(cmd.CmdKeys, "G") {
			return true, nil
		}
	}

	return false, nil
}

func (e *MenuExecutor) readKeyPress(noKeyCommands []database.MenuCommand, timeout time.Duration) (string, error) {
	var seq string
	var err error
	if len(noKeyCommands) == 0 || timeout <= 0 {
		seq, err = e.io.ReadKeySequence(0)
	} else {
		seq, err = e.io.ReadKeySequence(timeout)
	}
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return "", errNoKeyTimeout
		}
		return "", err
	}
	return normalizeInputKey(seq), nil
}

func (e *MenuExecutor) resolveNoKeyTimeout(commands []database.MenuCommand) time.Duration {
	if len(commands) == 0 {
		return 0
	}

	const defaultTimeout = 5 * time.Second
	for _, cmd := range commands {
		value := strings.TrimSpace(cmd.Options)
		if value == "" {
			continue
		}
		if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultTimeout
}

func normalizeInputKey(raw string) string {
	switch raw {
	case "":
		return ""
	case "\r", "\n", "\r\n":
		return "ENTER"
	case "\t":
		return "TAB"
	}

	if raw == "\x1b" {
		return "ESC"
	}

	if strings.HasPrefix(raw, "\x1b") {
		return "ESC"
	}

	if len(raw) == 1 {
		ch := raw[0]
		if ch >= 'a' && ch <= 'z' {
			ch -= 32
		}
		return string(ch)
	}

	return strings.ToUpper(raw)
}

// checkACS checks if the user has access based on ACS
func (e *MenuExecutor) checkACS(acs string, ctx *ExecutionContext) bool {
	if acs == "" {
		return true // No ACS requirement means allow access
	}

	if ctx == nil || ctx.Session == nil {
		return false // Deny access if context or session is nil
	}

	// Get user's security level
	userSecLevel := ctx.Session.SecurityLevel

	// Parse ACS string - it can contain multiple conditions separated by operators
	// For now, support basic level comparison (e.g., "10", ">5", "<100")
	acs = strings.TrimSpace(acs)

	// Check for comparison operators
	if strings.HasPrefix(acs, ">=") {
		if level, err := strconv.Atoi(acs[2:]); err == nil {
			return userSecLevel >= level
		}
	} else if strings.HasPrefix(acs, "<=") {
		if level, err := strconv.Atoi(acs[2:]); err == nil {
			return userSecLevel <= level
		}
	} else if strings.HasPrefix(acs, ">") {
		if level, err := strconv.Atoi(acs[1:]); err == nil {
			return userSecLevel > level
		}
	} else if strings.HasPrefix(acs, "<") {
		if level, err := strconv.Atoi(acs[1:]); err == nil {
			return userSecLevel < level
		}
	} else if strings.HasPrefix(acs, "=") || strings.HasPrefix(acs, "==") {
		levelStr := acs
		if strings.HasPrefix(acs, "==") {
			levelStr = acs[2:]
		} else {
			levelStr = acs[1:]
		}
		if level, err := strconv.Atoi(levelStr); err == nil {
			return userSecLevel == level
		}
	} else {
		// Direct level number
		if level, err := strconv.Atoi(acs); err == nil {
			return userSecLevel >= level
		}
	}

	// If we can't parse the ACS, deny access for security
	return false
}

// executeCommand executes a menu command
func (e *MenuExecutor) executeCommand(cmd database.MenuCommand, ctx *ExecutionContext) error {
	return e.registry.Execute(cmd.CmdKeys, ctx, cmd.Options)
}

// serveThemeFile serves the content of a theme file
func (e *MenuExecutor) serveThemeFile(themePath string) (int, error) {
	content, err := os.ReadFile(themePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read theme file %s: %w", themePath, err)
	}
	data := ui.StripSauce(string(content))
	e.io.Print(data)

	lineCount := strings.Count(data, "\n")
	if len(data) > 0 && !strings.HasSuffix(data, "\n") {
		lineCount++
	}
	if lineCount < 1 {
		lineCount = 1
	}

	if maxRow := detectANSIAbsoluteRow(data); maxRow > lineCount {
		lineCount = maxRow
	}
	e.currentRow += lineCount
	return lineCount, nil
}

func detectANSIAbsoluteRow(data string) int {
	maxRow := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\x1b' || i+1 >= len(data) || data[i+1] != '[' {
			continue
		}
		j := i + 2
		for j < len(data) && ((data[j] >= '0' && data[j] <= '9') || data[j] == ';') {
			j++
		}
		if j >= len(data) {
			break
		}
		cmd := data[j]
		params := data[i+2 : j]

		switch cmd {
		case 'H', 'f':
			row := parseANSIRow(params)
			if row > maxRow {
				maxRow = row
			}
		case 'd':
			if row, err := strconv.Atoi(params); err == nil && row > maxRow {
				maxRow = row
			}
		}
		i = j
	}
	return maxRow
}

func parseANSIRow(params string) int {
	if params == "" {
		return 1
	}
	parts := strings.Split(params, ";")
	if len(parts) == 0 || parts[0] == "" {
		return 1
	}
	if row, err := strconv.Atoi(parts[0]); err == nil && row > 0 {
		return row
	}
	return 1
}

// clearScreen clears the screen if ClearScreen is true
func (e *MenuExecutor) clearScreen(clear bool) {
	if clear {
		e.io.Print(ui.ClearScreenSequence()) // ANSI clear screen and move cursor to top-left
		e.currentRow = 1                     // Reset cursor row after clear screen
	}
}

// centerTitle centers the title text on screen
func (e *MenuExecutor) centerTitle(title string, ctx *ExecutionContext) string {
	width := 80
	if ctx != nil && ctx.Session != nil && ctx.Session.Width > 0 {
		width = ctx.Session.Width
	}
	// Strip both pipe codes and ANSI to get visible length
	visible := ui.StripANSI(ui.StripPipeCodes(title))
	if len(visible) >= width {
		return title
	}
	padding := (width - len(visible)) / 2
	return strings.Repeat(" ", padding) + title
}

// displayCommandsInColumns displays commands in columns with colors
func (e *MenuExecutor) displayCommandsInColumns(commands []database.MenuCommand, menu *database.Menu, ctx *ExecutionContext) {
	if len(commands) == 0 {
		return
	}

	// Filter commands that have short descriptions and are active
	var displayCommands []database.MenuCommand
	for _, cmd := range commands {
		if cmd.ShortDescription != "" && cmd.Active && !cmd.Hidden {
			displayCommands = append(displayCommands, cmd)
		}
	}

	if len(displayCommands) == 0 {
		return
	}

	// Use configured columns for the layout
	columns := menu.GenericColumns

	leftBracket := menu.LeftBracket
	if len([]rune(leftBracket)) == 0 {
		leftBracket = "["
	} else if runes := []rune(leftBracket); len(runes) > 2 {
		leftBracket = string(runes[:2])
	}
	rightBracket := menu.RightBracket
	if len([]rune(rightBracket)) == 0 {
		rightBracket = "]"
	} else if runes := []rune(rightBracket); len(runes) > 2 {
		rightBracket = string(runes[:2])
	}

	// ANSI color codes
	bracketColor := ui.ColorFromNumber(menu.GenericBracketColor)
	commandColor := ui.ColorFromNumber(menu.GenericCommandColor)
	descColor := ui.ColorFromNumber(menu.GenericDescColor)
	resetColor := ui.Ansi.Reset

	// Calculate items per column
	itemsPerColumn := (len(displayCommands) + columns - 1) / columns
	screenWidth := 80
	if ctx != nil && ctx.Session != nil && ctx.Session.Width > 0 {
		screenWidth = ctx.Session.Width
	}
	const margin = 2
	const interColumnPadding = 2

	// Calculate the maximum column width across all columns
	maxWidth := 0
	for _, cmd := range displayCommands {
		formatted := fmt.Sprintf("%s%s%s%s%s%s%s%s%s %s%s%s",
			bracketColor, leftBracket, resetColor,
			commandColor, cmd.Keys, resetColor,
			bracketColor, rightBracket, resetColor,
			descColor, cmd.ShortDescription, resetColor)
		visibleLen := len(ui.StripANSI(formatted))
		if visibleLen > maxWidth {
			maxWidth = visibleLen
		}
	}
	colWidths := make([]int, columns)
	for i := range colWidths {
		colWidths[i] = maxWidth
	}

	// Calculate total menu width and centering padding
	totalMenuWidth := 0
	for _, w := range colWidths {
		totalMenuWidth += w
	}
	totalMenuWidth += margin + margin + (columns-1)*interColumnPadding
	padding := (screenWidth - totalMenuWidth) / 2
	if padding < 0 {
		padding = 0
	}

	for row := 0; row < itemsPerColumn; row++ {
		line := ""
		for col := 0; col < columns; col++ {
			idx := col*itemsPerColumn + row
			if idx < len(displayCommands) {
				cmd := displayCommands[idx]
				// Format: [keys] Short Description
				formatted := fmt.Sprintf("%s%s%s%s%s%s%s%s%s %s%s%s",
					bracketColor, leftBracket, resetColor,
					commandColor, cmd.Keys, resetColor,
					bracketColor, rightBracket, resetColor,
					descColor, cmd.ShortDescription, resetColor)

				visibleLen := len(ui.StripANSI(formatted))
				if visibleLen < colWidths[col] {
					formatted += strings.Repeat(" ", colWidths[col]-visibleLen)
				}
				// Add left margin for first column, inter-column padding and right margin for others
				if col == 0 {
					line += strings.Repeat(" ", margin) + formatted
				} else {
					line += strings.Repeat(" ", interColumnPadding) + formatted + strings.Repeat(" ", margin)
				}
			}
		}
		// Apply centering padding to the beginning of each row
		line = strings.Repeat(" ", padding) + line
		e.io.Printf("%s\r\n", strings.TrimRight(line, " "))
		e.currentRow += 1 // Each command row adds 1 row
	}
}

func (e *MenuExecutor) getThemeBaseDir() string {
	if dir := ui.GetThemeDirectory(); strings.TrimSpace(dir) != "" {
		return strings.TrimSpace(dir)
	}
	if dir, err := e.db.GetConfig("Configuration.Paths", "", "Themes"); err == nil {
		dir = strings.TrimSpace(dir)
		if dir != "" {
			return dir
		}
	}
	return "theme"
}

func (e *MenuExecutor) lookupMenuByName(name string) (*database.Menu, error) {
	menu, err := e.db.GetMenuByName(name)
	if err == nil {
		return menu, nil
	}

	menus, listErr := e.db.GetAllMenus()
	if listErr != nil {
		return nil, fmt.Errorf("failed to load menu %s: %w", name, err)
	}

	for _, m := range menus {
		if strings.EqualFold(m.Name, name) {
			copy := m
			return &copy, nil
		}
	}

	return nil, fmt.Errorf("failed to load menu %s: %w", name, err)
}
