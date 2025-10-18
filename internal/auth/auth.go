package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/logging"
	"github.com/robbiew/retrograde/internal/security"
	"github.com/robbiew/retrograde/internal/telnet"
	"github.com/robbiew/retrograde/internal/ui"
)

// FieldEdit represents an editable field in the confirmation screen
type FieldEdit struct {
	Key        string
	Label      string
	Value      string
	DataType   string // "username", "email", "password", "text", "number"
	MapKey     string // Key in userDetails map (empty for username/email/password)
	Required   bool
	MaxLength  int
	Validation func(string) error
}

const userTimestampLayout = "2006-01-02 15:04:05"

// Type aliases for convenience
type Config = config.Config
type TelnetIO = telnet.TelnetIO
type TelnetSession = config.TelnetSession

// reservedUsernames contains usernames that cannot be registered
var reservedUsernames = map[string]bool{
	"new":           true,
	"sysop":         true,
	"admin":         true,
	"administrator": true,
	"root":          true,
	"system":        true,
	"guest":         true,
	"user":          true,
	"moderator":     true,
	"mod":           true,
	"operator":      true,
	"op":            true,
	"bot":           true,
	"staff":         true,
	"support":       true,
	"help":          true,
	"info":          true,
	"test":          true,
	"demo":          true,
	"sample":        true,
}

// IsReservedUsername checks if a username is reserved/banned (case-insensitive)
func IsReservedUsername(username string) bool {
	return reservedUsernames[strings.ToLower(username)]
}

// CreateUser creates a new user account
func CreateUser(username, password, email string, securityLevel int, userDetails map[string]string) error {
	// Hash the password
	passwordHash := HashPassword(password)
	digest := PasswordDigest{
		Hash:      passwordHash,
		Algorithm: "sha256",
		Salt:      "retrograde_salt_2025",
		UpdatedAt: time.Now().UTC(),
	}

	// Extract user details for direct storage in users table
	firstName := userDetails["first_name"]
	lastName := userDetails["last_name"]
	location := userDetails["locations"]

	params := CreateUserParams{
		Username:      username,
		Password:      digest,
		SecurityLevel: securityLevel,
		Email:         email,
		CreatedAt:     time.Now().UTC(),
		FirstName:     firstName,
		LastName:      lastName,
		Location:      location,
	}

	userRecord, err := getStorage().CreateUser(params)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return fmt.Errorf("user %s already exists", username)
		}
		if errors.Is(err, ErrEmailExists) {
			return fmt.Errorf("email %s already exists", email)
		}
		return fmt.Errorf("could not create user: %w", err)
	}

	// Store additional user details
	if len(userDetails) > 0 {
		for key, value := range userDetails {
			if err := getStorage().UpsertUserDetail(userRecord.ID, key, value); err != nil {
				// Log error but don't fail user creation
				fmt.Printf("Warning: could not store user detail %s: %v\n", key, err)
			}
		}
	}

	return nil
}

// AuthenticateUser verifies user credentials and returns user data
func AuthenticateUser(username, password string) (*UserRecord, error) {
	userRecord, err := getStorage().GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, fmt.Errorf("user %s not found", username)
		}
		return nil, fmt.Errorf("could not retrieve user: %w", err)
	}

	// Verify password
	if !VerifyPassword(password, userRecord.Password.Hash) {
		return nil, fmt.Errorf("invalid password")
	}

	// Update last login time
	now := time.Now().UTC()
	if err := getStorage().UpdateLastLogin(userRecord.ID, now); err != nil {
		// Log error but don't fail authentication
		fmt.Printf("Warning: could not update last login: %v\n", err)
	}

	return userRecord, nil
}

// HashPassword creates a SHA-256 hash of the password with salt
func HashPassword(password string) string {
	// Simple hash for demo - in production, use bcrypt or similar
	salt := "retrograde_salt_2025"
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

// GetUser loads a user by username
func GetUser(username string) (*UserRecord, error) {
	userRecord, err := getStorage().GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, fmt.Errorf("user %s not found", username)
		}
		return nil, fmt.Errorf("could not retrieve user: %w", err)
	}

	return userRecord, nil
}

// UserExists checks if a username exists
func UserExists(username string) bool {
	_, err := getStorage().GetUserByUsername(username)
	return err == nil
}

// EmailExists checks if an email address is already registered
func EmailExists(email string) bool {
	_, err := getStorage().GetUserByEmail(email)
	return err == nil
}

// LoginPrompt handles the login process for telnet clients
func LoginPrompt(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config) (*UserRecord, error) {
	io.ClearScreen()

	// Display login art
	if err := ui.PrintAnsiArt(io, "login"); err != nil {
		io.Print(" Welcome to the BBS\r\n")
	}

	io.Print("\r\n")
	io.Print(ui.Ansi.Cyan + " Enter your username or 'NEW'.\r\n\r\n" + ui.Ansi.Reset)

	// Flush any stray input
	io.FlushInput()

	// Track failed login attempts across all retries
	failedAttempts := 0
	maxAttempts := 3

	// Main login loop
	for {
		// Get username with validation
		var username string
		for {
			var err error
			username, err = ui.PromptSimple(io, " Username: ", 20, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, "")
			if err != nil {
				if err.Error() == "ESC_PRESSED" {
					if ui.HandleEscQuit(io) {
						logging.LogLoginFailed(session.NodeNumber, "Unknown", session.IPAddress, "login cancelled by user")
						return nil, fmt.Errorf("login cancelled")
					}
					io.Print("\r\n")
					continue
				}
				return nil, err
			}

			if username == "" {
				ui.ShowErrorAndClearPrompt(io, " Username cannot be empty.")
				continue
			} else if strings.EqualFold(username, "NEW") {
				// Handle new user registration
				userRecord, err := RegisterPrompt(io, session, cfg, "")
				if err != nil {
					// Registration failed or cancelled, restart login prompt
					io.ClearScreen()
					if err := ui.PrintAnsiArt(io, "login"); err != nil {
						io.Print(" Welcome to the BBS\r\n")
					}
					io.Print("\r\n\r\n")
					io.Print(ui.Ansi.Cyan + " Enter your username or 'NEW'.\r\n\r\n" + ui.Ansi.Reset)
					continue
				}
				// Registration successful, return the user
				return userRecord, nil
			} else if !UserExists(username) {
				// USERNAME NOT FOUND - OFFER TO CREATE ACCOUNT
				io.Print(ui.Ansi.YellowHi + " User not found. Create an account? (Y/N): " + ui.Ansi.Reset)

				createAccount := false
				for {
					key, err := io.GetKeyPress()
					if err != nil {
						return nil, err
					}

					if key == 'Y' || key == 'y' {
						io.Print("Y\r\n\r\n")
						createAccount = true
						break
					} else if key == 'N' || key == 'n' {
						io.Print("N\r\n")
						break
					}
					// Ignore other keys, keep waiting for Y/N
				}

				if createAccount {
					// Start registration process
					userRecord, err := RegisterPrompt(io, session, cfg, username)
					if err != nil {
						// Registration failed or cancelled, restart login prompt
						io.ClearScreen()
						if err := ui.PrintAnsiArt(io, "login"); err != nil {
							io.Print(" Welcome to the BBS\r\n")
						}
						io.Print("\r\n\r\n")
						io.Print(ui.Ansi.Cyan + " Enter your username or 'NEW'.\r\n\r\n" + ui.Ansi.Reset)
						continue
					}
					// Registration successful, return the user
					return userRecord, nil
				} else {
					// User chose not to create account, clear and restart
					ui.ShowErrorAndClearMultiplePrompts(io, "\r\n OK, try your Username again!", 5)
					continue
				}
			} else {
				break // User exists, continue to password prompt
			}
		}

		// Get password with validation (only reached if user exists)
		var password string
		for {
			var err error
			password, err = ui.PromptPasswordSimple(io, " Password: ", 20, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan)
			if err != nil {
				if err.Error() == "ESC_PRESSED" {
					if ui.HandleEscQuit(io) {
						logging.LogLoginFailed(session.NodeNumber, username, session.IPAddress, "login cancelled by user")
						return nil, fmt.Errorf("login cancelled")
					}
					io.Print("\r\n")
					continue
				}
				return nil, err
			}

			if password == "" {
				ui.ShowErrorAndClearMultiplePrompts(io, "\r\n Password cannot be empty.", 3)
				continue
			} else {
				break
			}
		}

		// Authenticate user
		user, err := AuthenticateUser(username, password)
		if err != nil {
			failedAttempts++
			logging.LogLoginFailed(session.NodeNumber, username, session.IPAddress, fmt.Sprintf("Attempt %d: %s", failedAttempts, err.Error()))

			// Check if max attempts exceeded - force disconnection and permanent blocklist
			if failedAttempts >= maxAttempts {
				io.Print(ui.Ansi.RedHi + "\r\n\r\n Too many login tries, hacker -- see ya!\r\n\r\n" + ui.Ansi.Reset)
				logging.LogLoginFailed(session.NodeNumber, username, session.IPAddress, fmt.Sprintf("Disconnected after %d failed attempts", maxAttempts))

				// Add IP to permanent blocklist
				security.AddToBlocklist(session.IPAddress, "Failed login attempts exceeded", "login_security", nil)

				time.Sleep(2 * time.Second) // Let them read the message

				// Force disconnection by setting session state and closing connection
				session.Connected = false
				if session.Conn != nil {
					session.Conn.Close()
				}
				return nil, fmt.Errorf("too many login attempts - disconnected")
			}

			// Incorrect password - clear both prompts and restart from username
			ui.ShowErrorAndClearMultiplePrompts(io, "\r\n Incorrect password, please try again.", 4)

			// Continue outer loop to restart from username prompt
			continue
		}

		// Authentication successful!
		logging.LogLogin(session.NodeNumber, user.Username, session.IPAddress)
		io.Printf(ui.Ansi.GreenHi+"\r\n\r\n Welcome back, %s!\r\n"+ui.Ansi.Reset, user.Username)
		ui.Pause(io)

		return user, nil
	}
}

// RegisterPrompt handles new user registration
func RegisterPrompt(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config, initialUsername string) (*UserRecord, error) {
	io.ClearScreen()

	// Display new user art
	if err := ui.PrintAnsiArt(io, "newuser"); err != nil {
		// Show error to user
		fmt.Printf("Failed to load art: %v\n", err)
	}

	io.Print("\r\n")

	// Compile regex once for username validation
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)

	// Get username with validation
	var username string
	for {
		var err error
		username, err = ui.PromptSimple(io, " Desired Username: ", 20, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, initialUsername)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if ui.HandleEscQuit(io) {
					logging.LogEvent(session.NodeNumber, "Unknown", session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				ui.ShowErrorAndClearMultiplePrompts(io, " Registration cancelled.", 2)
				continue
			}
			return nil, err
		}

		if username == "" {
			ui.ShowErrorAndClearMultiplePrompts(io, " Username cannot be empty.", 2)
			continue
		} else if len(username) < 3 {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "username too short")
			ui.ShowErrorAndClearMultiplePrompts(io, " Username must be at least 3 characters.", 2)
			continue
		} else if IsReservedUsername(username) {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "username is reserved")
			ui.ShowErrorAndClearMultiplePrompts(io, " Username '"+username+"' is reserved and cannot be used.", 2)
			continue
		} else if UserExists(username) {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "username already exists")
			ui.ShowErrorAndClearMultiplePrompts(io, " Username '"+username+"' already exists.", 2)
			continue
		} else if !usernameRegex.MatchString(username) {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "username contains illegal characters")
			ui.ShowErrorAndClearMultiplePrompts(io, " Username can only contain letters, numbers, and spaces.", 2)
			continue
		} else {
			break
		}
	}

	// Outer loop for password + confirmation
	var password string
	var confirmPassword string
	for {
		// Get password with validation
		password = "" // Reset at start of outer loop
		for {
			var err error
			password, err = ui.PromptPasswordSimple(io, " Desired Password: ", 20, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan)
			if err != nil {
				if err.Error() == "ESC_PRESSED" {
					if ui.HandleEscQuit(io) {
						logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
						return nil, fmt.Errorf("registration cancelled")
					}
					ui.ShowErrorAndClearPrompt(io, " Registration cancelled.")
					continue
				}
				return nil, err
			}

			if password == "" {
				ui.ShowErrorAndClearPrompt(io, " Password cannot be empty.")
				continue
			} else if len(password) < 4 {
				logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "password too short")
				ui.ShowErrorAndClearPrompt(io, " Password must be at least 4 characters.")
				continue
			} else {
				break // Password is valid, move to confirmation
			}
		}

		// Confirm password
		confirmPassword = "" // Reset
		for {
			var err error
			confirmPassword, err = ui.PromptPasswordSimple(io, " Confirm Password: ", 20, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan)
			if err != nil {
				if err.Error() == "ESC_PRESSED" {
					if ui.HandleEscQuit(io) {
						logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
						return nil, fmt.Errorf("registration cancelled")
					}
					ui.ShowErrorAndClearPrompt(io, " Registration cancelled.")
					continue
				}
				return nil, err
			}

			if confirmPassword != password {
				// Passwords don't match - clear both prompts and restart from password
				ui.ShowErrorAndClearMultiplePrompts(io, " Passwords do not match.", 3)
				break // Break inner loop to restart outer loop
			} else {
				// Passwords match!
				break
			}
		}

		// Check if passwords matched
		if confirmPassword == password {
			break // Exit outer loop - we're done with passwords
		}
		// Otherwise, outer loop continues and re-prompts for password
	}

	// Get email with validation (required)
	var email string
	for {
		var err error
		email, err = ui.PromptSimple(io, "    Email Address: ", 30, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, "")
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if ui.HandleEscQuit(io) {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				ui.ShowErrorAndClearMultiplePrompts(io, " Registration cancelled.", 2)
				continue
			}
			return nil, err
		}

		if email == "" {
			ui.ShowErrorAndClearMultiplePrompts(io, " Email is required.", 2)
			continue
		} else if EmailExists(email) {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "email already exists")
			ui.ShowErrorAndClearMultiplePrompts(io, " Email address already exists.", 2)
			continue
		} else {
			break
		}
	}

	// Collect additional registration fields based on configuration
	userDetails := make(map[string]string)

	// Get terminal width preference
	var terminalWidth int
	widthDefault := 80
	if session.Width > 0 && session.Width <= 255 {
		widthDefault = session.Width
	}

	// Show helpful message
	if session.Width > 0 {
		io.Print(ui.Ansi.BlackHi + fmt.Sprintf("\r\n Detected terminal size: %dx%d (press Enter to accept or edit)\r\n", session.Width, session.Height) + ui.Ansi.Reset)
	} else {
		io.Print(ui.Ansi.YellowHi + " Unable to detect terminal size. Using defaults (press Enter to accept)\r\n" + ui.Ansi.Reset)
	}

	for {
		var err error
		widthStr, err := ui.PromptSimple(io, "   Terminal Width: ", 3, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, strconv.Itoa(widthDefault))
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if ui.HandleEscQuit(io) {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				ui.ShowErrorAndClearMultiplePrompts(io, " Registration cancelled.", 2)
				continue
			}
			return nil, err
		}

		if widthStr == "" {
			terminalWidth = widthDefault
			break
		}

		width, err := strconv.Atoi(widthStr)
		if err != nil || width < 1 || width > 80 {
			ui.ShowErrorAndClearMultiplePrompts(io, " Width must be between 1 and 80.", 2)
			continue
		}

		terminalWidth = width
		break
	}

	// Get terminal height preference - always prompt, show detected value in label
	var terminalHeight int
	heightDefault := 24 // In case Syncterm Staus bar is on
	if session.Height > 0 {
		heightDefault = session.Height
		if heightDefault > 24 {
			heightDefault = 24
		}
	}
	for {
		var err error

		heightStr, err := ui.PromptSimple(io, "  Terminal Height: ", 3, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, strconv.Itoa(heightDefault))
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if ui.HandleEscQuit(io) {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				ui.ShowErrorAndClearMultiplePrompts(io, " Registration cancelled.", 2)
				continue
			}
			return nil, err
		}

		if heightStr == "" {
			terminalHeight = heightDefault
			break
		}

		height, err := strconv.Atoi(heightStr)
		if err != nil || height < 1 || height > 25 {
			ui.ShowErrorAndClearMultiplePrompts(io, " Height must be between 1 and 25.", 2)
			continue
		}

		terminalHeight = height
		break
	}

	// Store terminal preferences in userDetails
	userDetails["terminal_width"] = strconv.Itoa(terminalWidth)
	userDetails["terminal_height"] = strconv.Itoa(terminalHeight)

	// Check for additional fields from RegistrationFields config
	if cfg.Configuration.NewUsers.RegistrationFields != nil {
		// Get field names and sort them for consistent ordering
		fieldNames := make([]string, 0, len(cfg.Configuration.NewUsers.RegistrationFields))
		for fieldName := range cfg.Configuration.NewUsers.RegistrationFields {
			fieldNames = append(fieldNames, fieldName)
		}

		// Sort alphabetically to ensure consistent order (First Name before Last Name)
		sort.Strings(fieldNames)

		for _, fieldName := range fieldNames {
			fieldConfig := cfg.Configuration.NewUsers.RegistrationFields[fieldName]
			if fieldConfig.Enabled {
				// Skip email since it's already prompted separately
				if strings.ToLower(fieldName) == "email" {
					continue
				}
				var promptText string
				var maxLength int

				// Customize prompt based on field name
				switch strings.ToLower(fieldName) {
				case "first name":
					io.Print(ui.Ansi.BlackHi + "\r\n Real Names are required for some message networks.\r\n" + ui.Ansi.Reset)
					promptText = "       First Name: "
					maxLength = 30
				case "last name":
					promptText = "        Last Name: "
					maxLength = 30
				case "location":
					io.Print(ui.Ansi.BlackHi + "\r\n Location can be used to show your city/state or other info.\r\n" + ui.Ansi.Reset)
					promptText = "         Location: "
					maxLength = 30
				default:
					promptText = fieldName + ": "
					maxLength = 30
				}

				var value string
				if fieldConfig.Required {
					// Required field - keep prompting until we get a value
					for {
						var err error
						value, err = ui.PromptSimple(io, promptText, maxLength, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, "")
						if err != nil {
							if err.Error() == "ESC_PRESSED" {
								if ui.HandleEscQuit(io) {
									logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
									return nil, fmt.Errorf("registration cancelled")
								}
								ui.ShowErrorAndClearPrompt(io, "Registration cancelled.")
								continue
							}
							return nil, err
						}

						if strings.TrimSpace(value) == "" {
							ui.ShowErrorAndClearPrompt(io, fieldName+" is required.")
							continue
						} else {
							break
						}
					}
				} else {
					// Optional field
					var err error
					value, err = ui.PromptSimple(io, promptText, maxLength, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, "")
					if err != nil {
						if err.Error() == "ESC_PRESSED" {
							if ui.HandleEscQuit(io) {
								logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
								return nil, fmt.Errorf("registration cancelled")
							}
							ui.ShowErrorAndClearPrompt(io, "Registration cancelled.")
							continue
						}
						return nil, err
					}
				}

				if value != "" {
					// Map field names to match database column names for consistency
					var key string
					switch strings.ToLower(fieldName) {
					case "firstname", "first name":
						key = "first_name"
					case "lastname", "last name":
						key = "last_name"
					case "location":
						key = "locations"
					default:
						key = fieldName
					}
					userDetails[key] = value
				}
			}
		}
	}

	// Debug logging: show userDetails map contents
	fmt.Printf("DEBUG: userDetails map contents:\n")
	for key, value := range userDetails {
		fmt.Printf("  %s = '%s'\n", key, value)
	}

	// Display summary and confirm
	// Show interactive confirmation screen
	confirmed, err := ShowAccountConfirmation(io, &username, &email, &password, userDetails, session, cfg, initialUsername)
	if err != nil {
		logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", err.Error())
		return nil, err
	}
	if !confirmed {
		logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "user cancelled at confirmation")
		return nil, fmt.Errorf("registration cancelled")
	}

	// Create user account
	err = CreateUser(username, password, email, config.SecurityLevelRegular, userDetails)
	if err != nil {
		logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "could not create user")
		io.Print(ui.Ansi.Red + "Something went wrong creating your account. Please contact the sysop." + ui.Ansi.Reset + "\r\n")
		ui.Pause(io)
		return nil, err
	}

	// Load the created user
	user, err := GetUser(username)
	if err != nil {
		logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "could not load created user")
		return nil, err
	}

	// Log successful registration and auto-login
	logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_SUCCESS", "new account created")
	logging.LogLogin(session.NodeNumber, user.Username, session.IPAddress)

	io.Printf(ui.Ansi.GreenHi+"\r\n\r\n Account created successfully. Welcome, %s!\r\n"+ui.Ansi.Reset, username)
	ui.Pause(io)

	// Show new user welcome screen
	io.ClearScreen()
	if err := ui.PrintAnsiArt(io, "welcome"); err != nil {
		io.Print("\r\n Welcome to the BBS!\r\n\r\n")
	}
	ui.Pause(io)

	return user, nil
}

// ShowAccountConfirmation displays an interactive account summary where users can edit fields
func ShowAccountConfirmation(io *telnet.TelnetIO, username *string, email *string, password *string,
	userDetails map[string]string, session *config.TelnetSession, cfg *config.Config,
	initialUsername string) (bool, error) {

	for {
		io.ClearScreen()

		// Display confirmation art
		if err := ui.PrintAnsiArt(io, "confirm"); err != nil {
			io.Print("\r\n Account Confirmation\r\n")
		}

		io.Print("\r\n")
		io.Print(ui.Ansi.Cyan + " Review and edit your account information:\r\n\r\n" + ui.Ansi.Reset)

		// Dynamically build field list based on what was collected
		fields := []FieldEdit{}
		keyLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P"}
		keyIndex := 0

		// Always have username and email
		fields = append(fields, FieldEdit{
			Key:       keyLetters[keyIndex],
			Label:     "Username",
			Value:     *username,
			DataType:  "username",
			Required:  true,
			MaxLength: 20,
		})
		keyIndex++

		fields = append(fields, FieldEdit{
			Key:       keyLetters[keyIndex],
			Label:     "Email",
			Value:     *email,
			DataType:  "email",
			Required:  true,
			MaxLength: 30,
		})
		keyIndex++

		// Dynamically add fields from userDetails (in a consistent order)
		// First, add the standard fields in preferred order
		standardFields := []struct {
			mapKey string
			label  string
			maxLen int
		}{
			{"first_name", "First Name", 30},
			{"last_name", "Last Name", 30},
			{"locations", "Location", 30},
			{"terminal_width", "Term Width", 3},
			{"terminal_height", "Term Height", 3},
		}

		for _, sf := range standardFields {
			if val, exists := userDetails[sf.mapKey]; exists {
				dataType := "text"
				if sf.mapKey == "terminal_width" || sf.mapKey == "terminal_height" {
					dataType = "number"
				}

				fields = append(fields, FieldEdit{
					Key:       keyLetters[keyIndex],
					Label:     sf.label,
					Value:     val,
					DataType:  dataType,
					MapKey:    sf.mapKey,
					MaxLength: sf.maxLen,
				})
				keyIndex++
			}
		}

		// Add any other custom fields that aren't in the standard list
		for key, value := range userDetails {
			// Skip if already added
			isStandard := false
			for _, sf := range standardFields {
				if sf.mapKey == key {
					isStandard = true
					break
				}
			}
			if !isStandard && value != "" {
				fields = append(fields, FieldEdit{
					Key:       keyLetters[keyIndex],
					Label:     key,
					Value:     value,
					DataType:  "text",
					MapKey:    key,
					MaxLength: 50,
				})
				keyIndex++
			}
		}

		// Display all fields
		for _, field := range fields {
			io.Printf(" "+ui.Ansi.YellowHi+"%s"+ui.Ansi.Reset+") "+ui.Ansi.WhiteHi+"%-13s: "+ui.Ansi.Reset+"%s\r\n",
				field.Key, field.Label, field.Value)
		}

		io.Print("\r\n")
		io.Print(" " + ui.Ansi.GreenHi + "S" + ui.Ansi.Reset + ") Save and Create Account\r\n")
		io.Print(" " + ui.Ansi.RedHi + "Q" + ui.Ansi.Reset + ") Cancel Registration\r\n")
		io.Print("\r\n")
		io.Print(ui.Ansi.Cyan + " Select an option to edit, or S to save: " + ui.Ansi.Reset)

		// Get user choice
		key, err := io.GetKeyPressUpper()
		if err != nil {
			return false, err
		}

		io.Printf("%c\r\n\r\n", key)

		switch key {
		case 'S':
			// Confirm save
			return true, nil

		case 'Q':
			// Cancel
			io.Print(ui.Ansi.YellowHi + " Are you sure you want to cancel? (Y/N): " + ui.Ansi.Reset)
			confirmKey, err := io.GetKeyPressUpper()
			if err != nil {
				return false, err
			}
			if confirmKey == 'Y' {
				return false, fmt.Errorf("registration cancelled")
			}
			continue

		default:
			// Find the selected field
			var selectedField *FieldEdit
			for i := range fields {
				if fields[i].Key == string(key) {
					selectedField = &fields[i]
					break
				}
			}

			if selectedField == nil {
				// Invalid key, just redisplay
				continue
			}

			// Edit the selected field
			io.Printf(ui.Ansi.Cyan+" Edit %s:\r\n\r\n"+ui.Ansi.Reset, selectedField.Label)

			var newValue string
			var promptErr error

			// Use appropriate prompt based on field type
			if selectedField.DataType == "password" {
				newValue, promptErr = ui.PromptPasswordSimple(io, " "+selectedField.Label+": ",
					selectedField.MaxLength, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan)
			} else {
				newValue, promptErr = ui.PromptSimple(io, " "+selectedField.Label+": ",
					selectedField.MaxLength, ui.Ansi.Cyan, ui.Ansi.WhiteHi, ui.Ansi.BgCyan, selectedField.Value)
			}

			if promptErr != nil {
				if promptErr.Error() == "ESC_PRESSED" {
					continue
				}
				return false, promptErr
			}

			// Validate the new value based on field type
			valid := true
			errorMsg := ""

			switch selectedField.DataType {
			case "username":
				usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)
				if newValue == "" {
					valid = false
					errorMsg = "Username cannot be empty."
				} else if len(newValue) < 3 {
					valid = false
					errorMsg = "Username must be at least 3 characters."
				} else if IsReservedUsername(newValue) {
					valid = false
					errorMsg = "Username is reserved."
				} else if UserExists(newValue) && newValue != *username {
					valid = false
					errorMsg = "Username already exists."
				} else if !usernameRegex.MatchString(newValue) {
					valid = false
					errorMsg = "Username can only contain letters, numbers, and spaces."
				}
				if valid {
					*username = newValue
				}

			case "email":
				if newValue == "" {
					valid = false
					errorMsg = "Email cannot be empty."
				} else if EmailExists(newValue) && newValue != *email {
					valid = false
					errorMsg = "Email already exists."
				}
				if valid {
					*email = newValue
				}

			case "password":
				if newValue == "" {
					valid = false
					errorMsg = "Password cannot be empty."
				} else if len(newValue) < 4 {
					valid = false
					errorMsg = "Password must be at least 4 characters."
				}
				if valid {
					*password = newValue
				}

			case "number":
				num, err := strconv.Atoi(newValue)
				if err != nil {
					valid = false
					errorMsg = "Must be a valid number."
				} else {
					// Check specific number ranges
					if selectedField.MapKey == "terminal_width" && (num < 1 || num > 80) {
						valid = false
						errorMsg = "Width must be between 1 and 80."
					} else if selectedField.MapKey == "terminal_height" && (num < 1 || num > 25) {
						valid = false
						errorMsg = "Height must be between 1 and 25."
					}
				}
				if valid {
					userDetails[selectedField.MapKey] = newValue
				}

			case "text":
				// Basic text validation
				if selectedField.Required && strings.TrimSpace(newValue) == "" {
					valid = false
					errorMsg = selectedField.Label + " is required."
				}
				if valid {
					userDetails[selectedField.MapKey] = newValue
				}
			}

			// Show error if validation failed
			if !valid {
				io.Print(ui.Ansi.RedHi + " " + errorMsg + "\r\n" + ui.Ansi.Reset)
				time.Sleep(2 * time.Second)
			}
		}
	}
}
