package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/robbiew/retrograde/internal/auth"
	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/database"
	"github.com/robbiew/retrograde/internal/logging"
	"github.com/robbiew/retrograde/internal/menu"
	"github.com/robbiew/retrograde/internal/security"
	"github.com/robbiew/retrograde/internal/telnet"
	"github.com/robbiew/retrograde/internal/tui"
	"github.com/robbiew/retrograde/internal/ui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config", "edit", "-config", "--config", "/config":
			runConfigEditor()
			return
		case "setup", "install", "-setup", "--setup", "-install", "--install":
			err := runGuidedSetup()
			if err != nil {
				fmt.Printf("Guided setup failed: %v\n", err)
				os.Exit(1)
			}
			// runGuidedSetup returns nil on cancellation, so we only show success message on actual success
			// (success message is handled inside runGuidedSetup)
			return
		}
	}

	runServer()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func seedDefaultMenu(db database.Database) error {
	// Check if MAIN menu already exists
	menu, err := db.GetMenuByName("MAIN")
	if err == nil {
		// Menu exists, check if it has commands
		commands, err := db.GetMenuCommands(menu.ID)
		if err != nil {
			return fmt.Errorf("failed to get menu commands: %w", err)
		}
		if len(commands) > 0 {
			// Already has commands, skip
			return nil
		}
		// Add default commands
		defaultCommands := []database.MenuCommand{
			{
				MenuID:           menu.ID,
				CommandNumber:    1,
				Keys:             "R",
				ShortDescription: "Read Mail",
				ACSRequired:      "",
				CmdKeys:          "MM",
				Options:          "",
			},
			{
				MenuID:           menu.ID,
				CommandNumber:    2,
				Keys:             "P",
				ShortDescription: "Post Message",
				ACSRequired:      "",
				CmdKeys:          "MP",
				Options:          "",
			},
			{
				MenuID:           menu.ID,
				CommandNumber:    3,
				Keys:             "G",
				ShortDescription: "Goodbye",
				ACSRequired:      "",
				CmdKeys:          "G",
				Options:          "",
			},
		}
		for _, cmd := range defaultCommands {
			_, err := db.CreateMenuCommand(&cmd)
			if err != nil {
				return fmt.Errorf("failed to create menu command %s: %w", cmd.Keys, err)
			}
		}
		return nil
	}

	// Menu doesn't exist, create it with commands
	menu = &database.Menu{
		Name:                "MAIN",
		Titles:              []string{"-= Retrograde BBS =-", "-- Main Menu --"},
		Prompt:              "[@1 - @2]@MTime Left: [@V] (?=Help)@MMain Menu :",
		ACSRequired:         "",
		Password:            "",
		GenericColumns:      4,
		GenericBracketColor: 1,
		GenericCommandColor: 9,
		GenericDescColor:    1,
		ClearScreen:         true,
	}

	menuID, err := db.CreateMenu(menu)
	if err != nil {
		return fmt.Errorf("failed to create MAIN menu: %w", err)
	}

	// Create menu commands
	commands := []database.MenuCommand{
		{
			MenuID:           int(menuID),
			CommandNumber:    1,
			Keys:             "R",
			ShortDescription: "Read Mail",
			ACSRequired:      "",
			CmdKeys:          "MM",
			Options:          "",
		},
		{
			MenuID:           int(menuID),
			CommandNumber:    2,
			Keys:             "P",
			ShortDescription: "Post Message",
			ACSRequired:      "",
			CmdKeys:          "MP",
			Options:          "",
		},
		{
			MenuID:           int(menuID),
			CommandNumber:    3,
			Keys:             "G",
			ShortDescription: "Goodbye",
			ACSRequired:      "",
			CmdKeys:          "G",
			Options:          "",
		},
	}

	for _, cmd := range commands {
		_, err := db.CreateMenuCommand(&cmd)
		if err != nil {
			return fmt.Errorf("failed to create menu command %s: %w", cmd.Keys, err)
		}
	}

	return nil
}

func runGuidedSetup() error {
	// Get default root directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	rootDir := filepath.Join(cwd)

	// Run the TUI form to collect directory paths
	setupConfig, err := tui.RunGuidedSetupTUI(rootDir)
	if err != nil {
		return fmt.Errorf("guided setup failed: %v", err)
	}

	// Check if user cancelled (nil return indicates cancellation)
	if setupConfig == nil {
		fmt.Println("\nCancelled! Run install again to complete setup")
		return nil
	}

	// Create config from collected data
	cfg := config.GetDefaultConfig()
	cfg.Configuration.Paths.Database = filepath.Join(setupConfig.Data, "retrograde.db")
	cfg.Configuration.Paths.FileBase = setupConfig.Files
	cfg.Configuration.Paths.Logs = setupConfig.Logs
	cfg.Configuration.Paths.MessageBase = setupConfig.Msgs
	cfg.Configuration.Paths.System = setupConfig.Root
	cfg.Configuration.Paths.Themes = setupConfig.Theme
	cfg.Configuration.Paths.Security = setupConfig.Security

	// Create directories
	paths := []struct {
		name string
		path string
	}{
		{"Database", filepath.Dir(cfg.Configuration.Paths.Database)},
		{"Files", cfg.Configuration.Paths.FileBase},
		{"Logs", cfg.Configuration.Paths.Logs},
		{"Messages", cfg.Configuration.Paths.MessageBase},
		{"Security", cfg.Configuration.Paths.Security},
		{"System", cfg.Configuration.Paths.System},
		{"Text", cfg.Configuration.Paths.Themes},
	}

	for _, p := range paths {
		if _, err := os.Stat(p.path); err == nil {
			// Directory already exists
			fmt.Printf("✓ %s directory exists.\n", p.name)
		} else if err := os.MkdirAll(p.path, 0755); err != nil {
			fmt.Printf(ui.Ansi.RedHi+"Failed to create %s: %v\n"+ui.Ansi.Reset, p.name, err)
		} else {
			fmt.Printf("✓ Created %s directory.\n", p.name)
		}

	}

	// Show success message only on successful completion
	fmt.Println("\nRetrograde BBS successfully installed... Next steps:")
	fmt.Printf("- \"retrograde\" to start server on port %d\n", 2323) // default port
	fmt.Println("- \"retrograde config\" to customize")
	fmt.Println("- Copy files from git themes directory to your themes folder")

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

	// Seed default menu
	if err := seedDefaultMenu(db); err != nil {
		return fmt.Errorf("failed to seed default menu: %w", err)
	}

	// Save config
	if err := config.SaveConfigToDB(db, cfg, "system"); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func runConfigEditor() {
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
		cfg.Other.Discord.Username = "Retrograde BBS Bot"

		// Seed default menu for new installations
		db := config.GetDatabase()
		if db != nil {
			if seedErr := seedDefaultMenu(db); seedErr != nil {
				fmt.Printf("Warning: Failed to seed default menu: %v\n", seedErr)
			}
		}
	}

	if err := tui.RunConfigEditorTUI(cfg); err != nil {
		fmt.Printf("Error running configuration editor: %v\n", err)
		os.Exit(1)
	}
}

func runConfigEditorFromServer(cfg *config.Config) error {
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

		os.Exit(0)
	}

	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	ui.SetThemeDirectory(cfg.Configuration.Paths.Themes)
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
		ui.SetThemeDirectory(cfg.Configuration.Paths.Themes)

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

	// Create a session for this connection
	session := &config.TelnetSession{
		Alias:         "Guest",                   // Default for unauthenticated users
		SecurityLevel: config.SecurityLevelGuest, // Guest security level
		TimeLeft:      10,                        // 60 minutes
		StartTime:     time.Now(),
		LastActivity:  time.Now(),
		NodeNumber:    nodeID,
		IPAddress:     ipAddr,
		Connected:     true,
		Conn:          conn, // Store connection reference for timeout handling
		Width:         80,   // Default width
		Height:        25,   // Default height
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

	// Bypass main menu and proceed directly to login for telnet connections
	userRecord, err := auth.LoginPrompt(io, session, cfg)
	if err != nil {
		// Login failed or cancelled - disconnect
		if err.Error() != "login cancelled" {
			io.Printf(ui.Ansi.RedHi+"\r\n Login failed: %v\r\n"+ui.Ansi.Reset, err)
			io.Pause()
		}
		return
	}

	// Login successful - update session with user info
	session.Alias = userRecord.Username
	session.SecurityLevel = userRecord.SecurityLevel
	// Update node manager with new username
	if nm := logging.GetNodeManager(); nm != nil && session.NodeNumber > 0 {
		if conn, exists := nm.Connections[session.NodeNumber]; exists {
			conn.Username = userRecord.Username
		}
	}

	// Load and execute start menu
	db := config.GetDatabase()
	if db != nil {
		executor := menu.NewMenuExecutor(db, io)
		ctx := &menu.ExecutionContext{
			UserID:   userRecord.ID,
			Username: userRecord.Username,
			IO:       io,
			Session:  session,
		}
		startMenu := cfg.Configuration.General.StartMenu
		if startMenu == "" {
			startMenu = "MAIN"
		}
		if err := executor.ExecuteMenu(startMenu, ctx); err != nil {
			io.Printf("Menu error: %v\r\n", err)
		}
	}
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
		NAWS              = 31  // Negotiate About Window Size option
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

	// Ask client to DO NAWS (Negotiate About Window Size)
	writer.WriteByte(IAC)
	writer.WriteByte(DO)
	writer.WriteByte(NAWS)

	writer.Flush()

	// Give client time to process negotiations
	time.Sleep(300 * time.Millisecond)

	fmt.Println("Sent telnet option negotiations")
}

// monitorSessionTimeout monitors a session for inactivity and handles disconnection
func monitorSessionTimeout(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config) {

	warningShown := false
	timeoutDuration := time.Duration(cfg.Configuration.General.TimeoutMinutes) * time.Minute
	warningTime := timeoutDuration - (30 * time.Second) // Show warning 30 seconds before timeout

	for session.Connected {
		time.Sleep(10 * time.Second) // Check every 10 seconds

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
	io.Print("Thank you for using Retrograde BBS. Goodbye!\r\n\r\n")
	time.Sleep(2 * time.Second) // Give time to read message
}
