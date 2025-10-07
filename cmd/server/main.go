package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/robbiew/retrograde/internal/auth"
	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/database"
	"github.com/robbiew/retrograde/internal/logging"
	"github.com/robbiew/retrograde/internal/security"
	"github.com/robbiew/retrograde/internal/telnet"
	"github.com/robbiew/retrograde/internal/tui"
	"github.com/robbiew/retrograde/internal/ui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config", "edit":
			runConfigEditor()
			return
		}
	}

	runServer()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runGuidedSetup() error {
	// Clear screen for clean setup experience
	ui.ClearScreen()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(ui.Ansi.CyanHi + "╔══════════════════════════════════════╗" + ui.Ansi.Reset)
	fmt.Println(ui.Ansi.CyanHi + "║" + ui.Ansi.WhiteHi + "    Welcome to Retrograde BBS Setup   " + ui.Ansi.CyanHi + "║" + ui.Ansi.Reset)
	fmt.Println(ui.Ansi.CyanHi + "╚══════════════════════════════════════╝" + ui.Ansi.Reset)
	fmt.Println()
	fmt.Print(ui.Ansi.GreenHi + "Use Guided Setup? " + ui.Ansi.YellowHi + "(Y/n): " + ui.Ansi.Reset)
	scanner.Scan()
	response := strings.ToUpper(strings.TrimSpace(scanner.Text()))
	if response == "" {
		response = "Y" // Default to Yes
	}
	if response != "Y" {
		fmt.Println("Setup cancelled. Exiting.")
		os.Exit(0)
	}

	// Set Paths
	fmt.Println("\n" + ui.Ansi.BlueHi + "Setting up core paths..." + ui.Ansi.Reset)
	cfg := config.GetDefaultConfig()

	paths := []struct {
		name string
		path *string
	}{
		{"Database", &cfg.Configuration.Paths.Database},
		{"File Base", &cfg.Configuration.Paths.FileBase},
		{"Logs", &cfg.Configuration.Paths.Logs},
		{"Message Base", &cfg.Configuration.Paths.MessageBase},
		{"System", &cfg.Configuration.Paths.System},
		{"Themes", &cfg.Configuration.Paths.Themes},
	}

	for _, p := range paths {
		for {
			fmt.Printf(ui.Ansi.CyanHi + "Path for %s " + ui.Ansi.YellowHi + "[default: '%s']: " + ui.Ansi.Reset, p.name, *p.path)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if input != "" {
				*p.path = input
			}
			// Validate path is not empty
			if *p.path == "" {
				fmt.Println(ui.Ansi.RedHi + "Path cannot be empty. Please enter a valid path." + ui.Ansi.Reset)
				continue
			}
			break
		}

		fmt.Printf(ui.Ansi.CyanHi + "Create directory for %s at '%s'? " + ui.Ansi.YellowHi + "(Y/n): " + ui.Ansi.Reset, p.name, *p.path)
		scanner.Scan()
		response := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		if response == "" {
			response = "Y" // Default to Yes
		}

		if response == "Y" {
			if err := os.MkdirAll(*p.path, 0755); err != nil {
				fmt.Printf(ui.Ansi.RedHi + "Failed to create %s: %v\n" + ui.Ansi.Reset, p.name, err)
			} else {
				fmt.Printf(ui.Ansi.GreenHi + "✓ Created %s directory.\n" + ui.Ansi.Reset, p.name)
			}
		} else {
			// User said No - check if directory exists
			if !fileExists(*p.path) {
				fmt.Printf(ui.Ansi.YellowHi + "Warning: Directory '%s' does not exist. You will need to create it manually or the %s feature may not work properly.\n" + ui.Ansi.Reset, *p.path, strings.ToLower(p.name))
				fmt.Print(ui.Ansi.CyanHi + "Are you sure you want to continue without creating this directory? " + ui.Ansi.YellowHi + "(y/N): " + ui.Ansi.Reset)
				scanner.Scan()
				sureResponse := strings.ToUpper(strings.TrimSpace(scanner.Text()))
				if sureResponse == "" {
					sureResponse = "N" // Default to No (don't continue)
				}
				if sureResponse != "Y" {
					fmt.Println(ui.Ansi.MagentaHi + "Please reconsider creating the directory." + ui.Ansi.Reset)
					// Re-prompt for directory creation
					continue
				}
			}
		}
	}

	// Open db - ensure data directory exists
	dbPath := filepath.Join("data", "retrograde.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	dbConfig := database.ConnectionConfig{
		Path:    dbPath,
		Timeout: 5,
	}
	db, err := database.OpenSQLite(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.InitializeSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Save config
	if err := config.SaveConfigToDB(db, cfg, "system"); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Create Sysop Account
	fmt.Print("\n" + ui.Ansi.BlueHi + "Create Sysop Account? " + ui.Ansi.YellowHi + "(Y/n): " + ui.Ansi.Reset)
	scanner.Scan()
	response = strings.ToUpper(strings.TrimSpace(scanner.Text()))
	if response == "" {
		response = "Y" // Default to Yes
	}
	if response == "Y" {
		// Username input with validation
		var username string
		for {
			fmt.Print(ui.Ansi.CyanHi + "Username: " + ui.Ansi.Reset)
			scanner.Scan()
			username = strings.TrimSpace(scanner.Text())
			if username == "" {
				fmt.Println(ui.Ansi.RedHi + "Username cannot be empty." + ui.Ansi.Reset)
				continue
			}
			if len(username) < 3 {
				fmt.Println(ui.Ansi.RedHi + "Username must be at least 3 characters." + ui.Ansi.Reset)
				continue
			}
			break
		}

		// Password input with confirmation
		var password string
		for {
			fmt.Print(ui.Ansi.CyanHi + "Password: " + ui.Ansi.Reset)
			passwordBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
			if err != nil {
				fmt.Printf("\n" + ui.Ansi.RedHi + "Error reading password: %v\n" + ui.Ansi.Reset, err)
				os.Exit(1)
			}
			fmt.Println() // Print newline after password input
			password = strings.TrimSpace(string(passwordBytes))
			if password == "" {
				fmt.Println(ui.Ansi.RedHi + "Password cannot be empty." + ui.Ansi.Reset)
				continue
			}
			if len(password) < 6 {
				fmt.Println(ui.Ansi.RedHi + "Password must be at least 6 characters." + ui.Ansi.Reset)
				continue
			}

			fmt.Print(ui.Ansi.CyanHi + "Confirm Password: " + ui.Ansi.Reset)
			confirmBytes, err := term.ReadPassword(uintptr(syscall.Stdin))
			if err != nil {
				fmt.Printf("\n" + ui.Ansi.RedHi + "Error reading password: %v\n" + ui.Ansi.Reset, err)
				os.Exit(1)
			}
			fmt.Println() // Print newline after password input
			confirmPassword := strings.TrimSpace(string(confirmBytes))
			if password != confirmPassword {
				fmt.Println(ui.Ansi.RedHi + "Passwords do not match. Please try again." + ui.Ansi.Reset)
				continue
			}
			break
		}

		// Real Name input
		var realName string
		for {
			fmt.Print(ui.Ansi.CyanHi + "Real Name: " + ui.Ansi.Reset)
			scanner.Scan()
			realName = strings.TrimSpace(scanner.Text())
			if realName == "" {
				fmt.Println(ui.Ansi.RedHi + "Real Name cannot be empty." + ui.Ansi.Reset)
				continue
			}
			break
		}

		// Email input with basic validation
		var email string
		for {
			fmt.Print(ui.Ansi.CyanHi + "Email: " + ui.Ansi.Reset)
			scanner.Scan()
			email = strings.TrimSpace(scanner.Text())
			if email == "" {
				fmt.Println(ui.Ansi.RedHi + "Email cannot be empty." + ui.Ansi.Reset)
				continue
			}
			if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
				fmt.Println(ui.Ansi.RedHi + "Please enter a valid email address." + ui.Ansi.Reset)
				continue
			}
			break
		}

		// Location input
		var location string
		for {
			fmt.Print(ui.Ansi.CyanHi + "Location: " + ui.Ansi.Reset)
			scanner.Scan()
			location = strings.TrimSpace(scanner.Text())
			if location == "" {
				fmt.Println(ui.Ansi.RedHi + "Location cannot be empty." + ui.Ansi.Reset)
				continue
			}
			break
		}

		// Set global db
		config.SetDatabase(db)

		// Initialize auth
		if err := auth.Init(&cfg.Configuration.Auth, db); err != nil {
			return fmt.Errorf("failed to initialize auth: %w", err)
		}

		// Create user details
		userDetails := map[string]string{
			"realname": realName,
			"location": location,
		}

		// Create account with level 100
		if err := auth.CreateUser(username, password, email, 100, userDetails); err != nil {
			return fmt.Errorf("failed to create sysop account: %w", err)
		}
		fmt.Println(ui.Ansi.GreenHi + "✓ Sysop Account #1 created!" + ui.Ansi.Reset)
	}

	// Theme files setup instructions
	fmt.Printf("\n" + ui.Ansi.BlueHi + "Theme files setup:" + ui.Ansi.Reset + "\n")
	fmt.Printf(ui.Ansi.CyanHi + "Please copy theme files from the 'theme' directory to '%s'\n", cfg.Configuration.Paths.Themes)
	fmt.Printf("Example: " + ui.Ansi.YellowHi + "cp theme/* %s/\n" + ui.Ansi.Reset, cfg.Configuration.Paths.Themes)
	fmt.Printf(ui.Ansi.MagentaHi + "(This step is optional - you can do it later)\n" + ui.Ansi.Reset)

	return nil
}

func copyDir(src, dst string) error {
	// Simple copy, assume files
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func runConfigEditor() {
	fmt.Println("Starting configuration editor...")

	cfg, err := config.LoadConfig("")
	defer config.CloseDatabase()

	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		fmt.Printf("Creating new configuration from template...\n")

		cfg = &config.Config{}
		cfg.Configuration.General.BBSName = "Your BBS Name Here"
		cfg.Configuration.General.TimeoutMinutes = 3
		cfg.Servers.GeneralSettings.MaxNodes = 10
		cfg.Servers.GeneralSettings.MaxConnectionsPerIP = 5
		cfg.Servers.Telnet.Active = true
		cfg.Servers.Telnet.Port = 2323
		cfg.Servers.Security.RateLimits.Enabled = true
		cfg.Servers.Security.RateLimits.WindowMinutes = 15
		cfg.Servers.Security.LocalLists.BlacklistEnabled = true
		cfg.Servers.Security.LocalLists.BlacklistFile = "security/blacklist.txt"
		cfg.Servers.Security.LocalLists.WhitelistEnabled = false
		cfg.Servers.Security.LocalLists.WhitelistFile = "security/whitelist.txt"
		cfg.Servers.Security.GeoBlock.GeoBlockEnabled = false
		cfg.Servers.Security.GeoBlock.GeoAPIProvider = "ipapi"
		cfg.Servers.Security.GeoBlock.ThreatIntelEnabled = false
		cfg.Servers.Security.GeoBlock.BlocklistUpdateHours = 6
		cfg.Servers.Security.Logs.SecurityLogFile = "logs/security.log"
		cfg.Servers.Security.Logs.LogSecurityEvents = true
		cfg.Servers.Security.Logs.LogBlockedAttempts = true
		cfg.Other.Discord.Enabled = false
		cfg.Other.Discord.Username = "GHOSTnet Bot"
	}

	if err := tui.RunConfigEditorTUI(cfg); err != nil {
		fmt.Printf("Error running configuration editor: %v\n", err)
		os.Exit(1)
	}
}

func runConfigEditorFromServer(cfg *config.Config) error {
	fmt.Println("Starting configuration editor...")

	if err := tui.RunConfigEditorTUI(cfg); err != nil {
		return fmt.Errorf("error running configuration editor: %w", err)
	}
	return nil
}

func runServer() {
	// Check if this is first time setup
	dbPath := filepath.Join("data", "retrograde.db")
	if !fileExists(dbPath) {
		fmt.Println()
		fmt.Println("Retrograde Database not found.")
		if err := runGuidedSetup(); err != nil {
			fmt.Printf("Guided setup failed: %v\n", err)
			os.Exit(1)
		}

		// Exit after setup with instructions
		fmt.Println("\nRetrograde BBS successfully installed... Next steps:")
		fmt.Println("- \"retrograde config\" to customize, or")
		fmt.Printf("- \"retrograde\" to start server on port %d\n", 2323) // default port
		os.Exit(0)
	}

	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	defer config.CloseDatabase()

	// Check if required paths exist, if not, launch config editor
	if !config.CheckRequiredPathsExist(cfg) {
		fmt.Println("Some required directories do not exist.")
		fmt.Println("Launching configuration editor to set up paths...")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln() // Wait for user to press Enter

		if err := runConfigEditorFromServer(cfg); err != nil {
			fmt.Printf("Error running configuration editor: %v\n", err)
			os.Exit(1)
		}

		// Reload configuration after editing
		cfg, err = config.LoadConfig("")
		if err != nil {
			fmt.Printf("Error reloading configuration: %v\n", err)
			os.Exit(1)
		}

		// Try to create missing directories
		if err := config.EnsureRequiredPaths(cfg); err != nil {
			fmt.Printf("Warning: Could not create some directories: %v\n", err)
			fmt.Println("You may need to create directories manually or adjust permissions.")
		}
	}

	logging.InitializeNodeManager(cfg.Servers.GeneralSettings.MaxNodes)
	security.InitializeSecurity(cfg)
	go security.CleanupSecurityData()

	if err := auth.Init(&cfg.Configuration.Auth, config.GetDatabase()); err != nil {
		fmt.Printf("Error initializing auth: %v\n", err)
		os.Exit(1)
	}

	listenAddr := fmt.Sprintf(":%d", cfg.Servers.Telnet.Port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Retrograde Application Server starting...")
	fmt.Printf("Server listening on port %d\n", cfg.Servers.Telnet.Port)
	fmt.Printf("Connect with: telnet localhost %d\n", cfg.Servers.Telnet.Port)
	fmt.Printf("Maximum nodes: %d\n", cfg.Servers.GeneralSettings.MaxNodes)
	fmt.Println("Press Ctrl+C to stop the server")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Fprintf(conn, "Connecting...")
		allowed, reason := security.CheckConnectionSecurity(conn, cfg)

		if !allowed {
			fmt.Printf("Connection blocked from %s: %s\n", security.GetIPFromConn(conn), reason)
			fmt.Fprintf(conn, "\r\nConnection temporarily unavailable.\r\nPlease try again later.\r\n")
			conn.Close()
			continue
		}

		fmt.Fprintf(conn, "\r                    \r")

		nodeID := logging.GetNodeManager().GetAvailableNode()
		if nodeID == -1 {
			fmt.Fprintf(conn, "Sorry, all %d nodes are currently in use.\r\nPlease try again later.\r\n", cfg.Servers.GeneralSettings.MaxNodes)
			conn.Close()
			continue
		}

		go handleConnection(conn, cfg, nodeID)
	}
}

func handleConnection(conn net.Conn, cfg *config.Config, nodeID int) {
	defer func() {
		conn.Close()
		logging.ReleaseNodeWithLogging(nodeID)
	}()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Get client IP address
	ipAddr := conn.RemoteAddr().String()
	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		ipAddr = tcpAddr.IP.String()
	}

	// Assign node and log connection
	actualNodeID := logging.AssignNodeWithLogging(conn, "Guest")
	if actualNodeID != nodeID {
		fmt.Printf("Warning: Node ID mismatch - expected %d, got %d\n", nodeID, actualNodeID)
	}

	// Create a session for this connection - start as guest
	session := &config.TelnetSession{
		Alias:         "Guest",                    // Default for unauthenticated users
		SecurityLevel: config.SecurityLevelGuest, // Guest security level
		TimeLeft:      60,                        // 60 minutes
		StartTime:     time.Now(),
		LastActivity:  time.Now(),
		NodeNumber:    nodeID,
		IPAddress:     ipAddr,
		Connected:     true,
		Conn:          conn, // Store connection reference for timeout handling
	}

	// Create TelnetIO wrapper with session reference
	io := &telnet.TelnetIO{
		Reader:  reader,
		Writer:  writer,
		Session: session,
	}

	// Start timeout monitoring goroutine
	go monitorSessionTimeout(io, session, cfg)

	// NOW that security is cleared, send telnet options to enable character mode
	negotiateTelnetOptions(writer)

	// Main menu loop - only for security-cleared connections
	mainMenu(io, session, cfg)
}

func mainMenu(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config) {
	for session.Connected {
		// Display main menu
		showMainMenu(io, session, cfg)

		// Get user input
		key, err := io.GetKeyPressUpper()
		if err != nil {
			fmt.Printf("Connection error: %v\n", err)
			return
		}

		// Log what we received on server side
		fmt.Printf("Received key press: %d (char: '%c')\n", key, key)

		// Handle menu options
		switch key {
		case 'L':
			if session.Alias != "Guest" {
				// Logout functionality
				logging.LogLogout(session.NodeNumber, session.Alias, session.IPAddress)
				io.Printf("\r\n Logging out %s...\r\n", session.Alias)
				session.Alias = "Guest"
				session.SecurityLevel = config.SecurityLevelGuest
				// Update node manager
				if nm := logging.GetNodeManager(); nm != nil && session.NodeNumber > 0 {
					if conn, exists := nm.Connections[session.NodeNumber]; exists {
						conn.Username = "Guest"
					}
				}
				io.Pause()
			} else {
				// Direct login for guest users
				userRecord, err := auth.LoginPrompt(io, session)
				if err != nil {
					// Don't show error message or pause for cancelled logins
					if err.Error() != "login cancelled" {
						io.Printf(ui.Ansi.RedHi+"\r\n Login failed: %v\r\n"+ui.Ansi.Reset, err)
						io.Pause()
					}
					// For cancelled logins, silently return to main menu
				} else {
					// Update session with logged-in user info
					session.Alias = userRecord.Username
					session.SecurityLevel = userRecord.SecurityLevel
					// Update node manager with new username
					if nm := logging.GetNodeManager(); nm != nil && session.NodeNumber > 0 {
						if conn, exists := nm.Connections[session.NodeNumber]; exists {
							conn.Username = userRecord.Username
						}
					}
				}
			}

		case 'R':
			// Only allow registration for guest users
			if session.Alias == "Guest" {
				userRecord, err := auth.RegisterPrompt(io, session, cfg)
				if err != nil {
					// Don't show error message or pause for cancelled registration
					if err.Error() != "registration cancelled" {
						io.Printf(ui.Ansi.RedHi+"\r\n Registration failed: %v\r\n"+ui.Ansi.Reset, err)
						io.Pause()
					}
					// For cancelled registration, silently return to main menu
				} else {
					// Update session with registered user info
					session.Alias = userRecord.Username
					session.SecurityLevel = userRecord.SecurityLevel
					// Update node manager with new username
					if nm := logging.GetNodeManager(); nm != nil && session.NodeNumber > 0 {
						if conn, exists := nm.Connections[session.NodeNumber]; exists {
							conn.Username = userRecord.Username
						}
					}
				}
			}
			// Silently ignore 'R' key if user is logged in

		case 'Q':
			io.Print("\r\n\r\nGoodbye! Thanks for visiting.\r\n")
			return

		default:
			// Silently ignore invalid menu options
		}
	}
}

func showMainMenu(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config) {
	// Clear screen and show header
	io.ClearScreen()
	io.PrintAnsi("connect", 0, 6)
	io.MoveCursor(0, 6)
	io.Print(" " + ui.Ansi.BgCyanHi + ui.Ansi.WhiteHi + " " + cfg.Configuration.General.BBSName + " - Main Menu " + ui.Ansi.Reset + "\r\n")

	// Show user status
	if session.Alias == "Guest" {
		io.Print("\r\n " + ui.Ansi.CyanHi + "Welcome, Guest!" + ui.Ansi.Reset + "\r\n")
	} else {
		io.Print("\r\n " + ui.Ansi.CyanHi + "Logged in as: " + session.Alias + ui.Ansi.Reset + "\r\n")
	}

	io.Print("\r\n")

	// Display Login/Register/Logout as menu items
	if session.Alias == "Guest" {
		io.Print(" " + ui.Ansi.Cyan + "[" + ui.Ansi.CyanHi + "L" + ui.Ansi.Reset + ui.Ansi.Cyan + "] " + ui.Ansi.Cyan + "Login\r\n" + ui.Ansi.Reset)
		io.Print(" " + ui.Ansi.Cyan + "[" + ui.Ansi.CyanHi + "R" + ui.Ansi.Reset + ui.Ansi.Cyan + "] " + ui.Ansi.Cyan + "Register\r\n" + ui.Ansi.Reset)
	} else {
		io.Print(" " + ui.Ansi.Cyan + "[" + ui.Ansi.CyanHi + "L" + ui.Ansi.Reset + ui.Ansi.Cyan + "] " + ui.Ansi.Cyan + "Logout\r\n" + ui.Ansi.Reset)
	}

	io.Print(" " + ui.Ansi.Cyan + "[" + ui.Ansi.CyanHi + "Q" + ui.Ansi.Reset + ui.Ansi.Cyan + "] " + ui.Ansi.Cyan + "Quit\r\n" + ui.Ansi.Reset)

	io.Print("\r\n " + ui.Ansi.CyanHi + "Select an option" + ui.Ansi.Reset + ": ")
}

func negotiateTelnetOptions(writer *bufio.Writer) {
	// Telnet protocol constants
	const (
		IAC               = 255 // Interpret As Command
		WILL              = 251 // Server will enable option
		WONT              = 252 // Server won't enable option
		DO                = 253 // Request client to enable option
		DONT              = 254 // Request client to disable option
		ECHO              = 1   // Echo option
		SUPPRESS_GO_AHEAD = 3   // Suppress Go Ahead option
		LINEMODE          = 34  // Line mode option
	)

	// Send telnet negotiations to enable character mode

	// Server WILL ECHO (server echoes back what user types)
	writer.WriteByte(IAC)
	writer.WriteByte(WILL)
	writer.WriteByte(ECHO)

	// Server WILL SUPPRESS GO AHEAD (disable line buffering)
	writer.WriteByte(IAC)
	writer.WriteByte(WILL)
	writer.WriteByte(SUPPRESS_GO_AHEAD)

	// Ask client to DO SUPPRESS GO AHEAD (client should not buffer)
	writer.WriteByte(IAC)
	writer.WriteByte(DO)
	writer.WriteByte(SUPPRESS_GO_AHEAD)

	// Ask client to DON'T use LINEMODE (disable line mode)
	writer.WriteByte(IAC)
	writer.WriteByte(DONT)
	writer.WriteByte(LINEMODE)

	writer.Flush()

	// Give client time to process negotiations
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Sent telnet option negotiations")
}

// monitorSessionTimeout monitors a session for inactivity and handles disconnection
func monitorSessionTimeout(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config) {
	// Check if user is sysop and exempt from timeout
	isExempt := func() bool {
		if !cfg.Configuration.General.SysOpTimeoutExempt {
			return false
		}
		return strings.EqualFold(session.Alias, cfg.Configuration.General.SysOpName)
	}

	warningShown := false
	timeoutDuration := time.Duration(cfg.Configuration.General.TimeoutMinutes) * time.Minute
	warningTime := timeoutDuration - (30 * time.Second) // Show warning 30 seconds before timeout

	for session.Connected {
		time.Sleep(10 * time.Second) // Check every 10 seconds

		// Skip timeout for sysop if configured
		if isExempt() {
			continue
		}

		timeSinceActivity := time.Since(session.LastActivity)

		// Show warning at 30 seconds remaining
		if !warningShown && timeSinceActivity >= warningTime {
			showTimeoutWarning(io, 30)
			warningShown = true
		}

		// Disconnect if timeout exceeded
		if timeSinceActivity >= timeoutDuration {
			showTimeoutDisconnection(io, cfg.Configuration.General.TimeoutMinutes)
			session.Connected = false
			if session.Conn != nil {
				session.Conn.Close() // Actually close the TCP connection
			}
			return
		}

		// Reset warning if user becomes active again
		if warningShown && timeSinceActivity < warningTime {
			warningShown = false
		}
	}
}

// showTimeoutWarning displays a warning message about impending timeout
func showTimeoutWarning(io *telnet.TelnetIO, secondsRemaining int) {
	// Save current cursor position and display warning
	io.Print(fmt.Sprintf("\r\n%s WARNING: You will be disconnected in %d seconds due to inactivity!%s\r\n",
		ui.Ansi.YellowHi, secondsRemaining, ui.Ansi.Reset))
	io.Print("Press any key to remain connected...\r\n")
}

// showTimeoutDisconnection displays final disconnection message
func showTimeoutDisconnection(io *telnet.TelnetIO, timeoutMinutes int) {
	io.Print(fmt.Sprintf("\r\n%s Session timeout: Disconnected due to %d minutes of inactivity.%s\r\n",
		ui.Ansi.RedHi, timeoutMinutes, ui.Ansi.Reset))
	io.Print("Thank you for using GHOSTnet. Goodbye!\r\n\r\n")
	time.Sleep(2 * time.Second) // Give time to read message
}
