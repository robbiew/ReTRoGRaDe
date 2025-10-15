package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robbiew/retrograde/internal/config"
	"github.com/robbiew/retrograde/internal/logging"
	"github.com/robbiew/retrograde/internal/security"
	"github.com/robbiew/retrograde/internal/telnet"
	"github.com/robbiew/retrograde/internal/ui"
	"github.com/robbiew/retrograde/internal/util"
)

const userTimestampLayout = "2006-01-02 15:04:05"

// SanitizeFilename replaces unsafe characters in the username to make it file-system safe
func SanitizeFilename(name string) string {
	return util.SanitizeFilename(name)
}

// Type aliases for convenience
type Config = config.Config
type TelnetIO = telnet.TelnetIO
type TelnetSession = config.TelnetSession

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

	params := CreateUserParams{
		Username:      username,
		Password:      digest,
		SecurityLevel: securityLevel,
		Email:         email,
		CreatedAt:     time.Now().UTC(),
	}

	userRecord, err := getStorage().CreateUser(params)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return fmt.Errorf("user %s already exists", username)
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

// LoginPrompt handles the login process for telnet clients
func LoginPrompt(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config) (*UserRecord, error) {
	io.ClearScreen()

	// Simple one-liner to display login art
	if err := ui.PrintAnsiArt(io, "login.ans"); err != nil {
		// Log the actual error so we can see what's wrong
		fmt.Printf("DEBUG: Failed to load login art: %v\n", err)
		io.Print("Welcome to the BBS\r\n")
	}
	io.Print("\r\n\r\n")

	io.Print(ui.Ansi.Cyan + "Enter your username or 'NEW' to apply and hit RETURN.\r\n\r\n" + ui.Ansi.Reset)

	io.Print("\r\n")

	// Get username with validation
	var username string
	for {
		var err error
		username, err = io.Prompt("Username: ", 2, 10, 20)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogLoginFailed(session.NodeNumber, "Unknown", session.IPAddress, "login cancelled by user")
					return nil, fmt.Errorf("login cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Username: ", 2, 10, 20)
				continue
			}
			return nil, err
		}

		if username == "" {
			io.ShowTimedError("Username cannot be empty.", 2, 13)
			io.ClearField("Username: ", 2, 10, 20)
			continue
		} else {
			break
		}
	}

	// Get password with validation
	var password string
	for {
		var err error
		password, err = io.PromptPassword("Password: ", 2, 11, 20)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogLoginFailed(session.NodeNumber, username, session.IPAddress, "login cancelled by user")
					return nil, fmt.Errorf("login cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Password: ", 2, 11, 20)
				continue
			}
			return nil, err
		}

		if password == "" {
			io.ShowTimedError("Password cannot be empty.", 2, 13)
			io.ClearField("Password: ", 2, 11, 20)
			continue
		} else {
			break
		}
	}

	// Authenticate user - use a loop to handle "user not found" errors properly
	var user *UserRecord
	failedAttempts := 0
	maxAttempts := 3

	for {
		var err error
		user, err = AuthenticateUser(username, password)
		if err != nil {
			failedAttempts++
			logging.LogLoginFailed(session.NodeNumber, username, session.IPAddress, fmt.Sprintf("Attempt %d: %s", failedAttempts, err.Error()))

			// Check if max attempts exceeded - force disconnection and permanent blacklist
			if failedAttempts >= maxAttempts {
				io.Print(ui.Ansi.RedHi + "\r\n\r\n Too many login tries, hacker -- see ya!\r\n\r\n" + ui.Ansi.Reset)
				logging.LogLoginFailed(session.NodeNumber, username, session.IPAddress, fmt.Sprintf("Disconnected after %d failed attempts", maxAttempts))

				// Add IP to permanent blacklist
				security.AddToBlacklist(session.IPAddress, "Failed login attempts exceeded", "login_security", nil)

				time.Sleep(2 * time.Second) // Let them read the message

				// Force disconnection by setting session state and closing connection
				session.Connected = false
				if session.Conn != nil {
					session.Conn.Close()
				}
				return nil, fmt.Errorf("too many login attempts - disconnected")
			}

			// Use generic error message for security (don't reveal if user exists or not)
			// Show error message synchronously to avoid goroutine interference
			io.PrintAt(ui.Ansi.RedHi+"Invalid login, try again."+ui.Ansi.Reset, 2, 13)

			// Wait for user to read the error message, form remains visible during display
			time.Sleep(2 * time.Second)

			// Clear error message first
			io.PrintAt(strings.Repeat(" ", 78), 2, 13) // Clear error message line

			// Now clear form fields and restart
			io.PrintAt(strings.Repeat(" ", 78), 2, 10) // Clear username line
			io.PrintAt(strings.Repeat(" ", 78), 2, 11) // Clear password line

			// Reset completely and restart from the very beginning
			username = ""
			password = ""

			// Get username with validation (exactly like original pattern at top of function)
			for {
				var err error
				username, err = io.Prompt("Username: ", 2, 10, 20)
				if err != nil {
					if err.Error() == "ESC_PRESSED" {
						if io.HandleEscQuit() {
							logging.LogLoginFailed(session.NodeNumber, "Unknown", session.IPAddress, "login cancelled by user")
							return nil, fmt.Errorf("login cancelled")
						}
						// User chose to continue, clear field and retry
						io.ClearField("Username: ", 2, 10, 20)
						continue
					}
					return nil, err
				}

				if username == "" {
					io.ShowTimedError("Username cannot be empty.", 2, 13)
					io.ClearField("Username: ", 2, 10, 20)
					continue
				} else if username == "NEW" {
					// Handle new user registration
					userRecord, err := RegisterPrompt(io, session, cfg)
					if err != nil {
						// Registration failed or cancelled, restart login prompt
						io.ClearScreen()
						// Display theme file first
						if content, err := ui.LoadAnsiArt("login"); err == nil {
							io.Print(ui.TrimStringFromSauce(content))
						} else {
							// Fallback to original behavior if theme file doesn't exist
							io.PrintAnsi("login", 0, 6) // Use login.ans as header
						}
						io.Print("\r\n\r\n")
						io.Print(ui.Ansi.Cyan + "Enter your username or 'NEW' to apply and hit RETURN.\r\n\r\n" + ui.Ansi.Reset)
						continue
					}
					// Registration successful, return the user
					return userRecord, nil
				} else {
					break
				}
			}

			// Get password with validation (exactly like original pattern)
			for {
				var err error
				password, err = io.PromptPassword("Password: ", 2, 11, 20)
				if err != nil {
					if err.Error() == "ESC_PRESSED" {
						if io.HandleEscQuit() {
							logging.LogLoginFailed(session.NodeNumber, username, session.IPAddress, "login cancelled by user")
							return nil, fmt.Errorf("login cancelled")
						}
						// User chose to continue, clear field and retry
						io.ClearField("Password: ", 2, 11, 20)
						continue
					}
					return nil, err
				}

				if password == "" {
					io.ShowTimedError("Password cannot be empty.", 2, 13)
					io.ClearField("Password: ", 2, 11, 20)
					continue
				} else {
					break
				}
			}

			// Continue the loop to try authentication again with new credentials
			continue
		} else {
			// Authentication successful, break out of the loop
			break
		}
	}

	// Log successful login
	logging.LogLogin(session.NodeNumber, user.Username, session.IPAddress)

	io.Printf(ui.Ansi.GreenHi+"\r\n\r\n Welcome back, %s!\r\n"+ui.Ansi.Reset, user.Username)
	io.Pause()

	return user, nil
}

// RegisterPrompt handles new user registration
func RegisterPrompt(io *telnet.TelnetIO, session *config.TelnetSession, cfg *config.Config) (*UserRecord, error) {
	io.ClearScreen()
	io.PrintAnsi("newuser", 0, 6) // Use newuser.ans as header

	// Create full-width header bar with ESC indicator at the end (columns 2-79, width 78)
	headerText := " New User Registration "
	escText := "[ESC] Quit/Cancel"
	middlePadding := 78 - len(headerText) - len(escText) - 1 // -1 for space after escText
	if middlePadding < 0 {
		middlePadding = 0
	}
	fullHeader := headerText + strings.Repeat(" ", middlePadding) + escText + " "
	io.PrintAt(ui.Ansi.BgCyan+ui.Ansi.WhiteHi+fullHeader+ui.Ansi.Reset, 2, 6)
	io.Print("\r\n\r\n")

	io.Print(ui.Ansi.Cyan + " This will create your account for accessing Retrograde BBS.\r\n\r\n" + ui.Ansi.Reset)

	// Compile regex once for username validation
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)

	// Get username with validation
	var username string
	for {
		var err error
		username, err = io.Prompt("Username: ", 2, 10, 20)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogEvent(session.NodeNumber, "Unknown", session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Username: ", 2, 10, 20)
				continue
			}
			return nil, err
		}

		if username == "" {
			io.ShowTimedError("Username cannot be empty.", 2, 14)
			io.ClearField("Username: ", 2, 10, 20)
			continue
		} else if len(username) < 3 {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "username too short")
			io.ShowTimedError("Username must be at least 3 characters.", 2, 14)
			io.ClearField("Username: ", 2, 10, 20)
			continue
		} else if UserExists(username) {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "username already exists")
			io.ShowTimedError("Username "+username+" already exists.", 2, 14)
			io.ClearField("Username: ", 2, 10, 20)
			continue
		} else if !usernameRegex.MatchString(username) {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "username contains illegal characters")
			io.ShowTimedError("Username can only contain letters, numbers, and spaces.", 2, 14)
			io.ClearField("Username: ", 2, 10, 20)
			continue
		} else {
			break
		}
	}

	// Get password with validation
	var password string
	for {
		var err error
		password, err = io.PromptPassword("Password: ", 2, 11, 20)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Password: ", 2, 11, 20)
				continue
			}
			return nil, err
		}

		if password == "" {
			io.ShowTimedError("Password cannot be empty.", 2, 14)
			io.ClearField("Password: ", 2, 11, 20)
			continue
		} else if len(password) < 4 {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "password too short")
			io.ShowTimedError("Password must be at least 4 characters.", 2, 14)
			io.ClearField("Password: ", 2, 11, 20)
			continue
		} else {
			break
		}
	}

	// Confirm password
	var confirmPassword string
	for {
		var err error
		confirmPassword, err = io.PromptPassword("Confirm Password: ", 2, 12, 20)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Confirm Password: ", 2, 12, 20)
				continue
			}
			return nil, err
		}

		if confirmPassword != password {
			io.ShowTimedError("Passwords do not match.", 2, 14)
			io.ClearField("Confirm Password: ", 2, 12, 20)
			continue
		} else {
			break
		}
	}

	// Get email with validation (required)
	var email string
	for {
		var err error
		email, err = io.Prompt("   Email: ", 2, 13, 30)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("   Email: ", 2, 13, 30)
				continue
			}
			return nil, err
		}

		if email == "" {
			io.ShowTimedError("Email is required.", 2, 14)
			io.ClearField("   Email: ", 2, 13, 30)
			continue
		} else {
			break
		}
	}

	// Collect additional registration fields based on configuration
	userDetails := make(map[string]string)

	// Get terminal width preference
	var terminalWidth int
	for {
		var err error
		widthStr, err := io.Prompt("Terminal Width (80): ", 2, 14, 3)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Terminal Width (80): ", 2, 14, 3)
				continue
			}
			return nil, err
		}

		if widthStr == "" {
			terminalWidth = 80
			break
		}

		width, err := strconv.Atoi(widthStr)
		if err != nil || width < 1 || width > 80 {
			io.ShowTimedError("Width must be between 1 and 80.", 2, 15)
			io.ClearField("Terminal Width (80): ", 2, 14, 3)
			continue
		}

		terminalWidth = width
		break
	}

	// Get terminal height preference
	var terminalHeight int
	for {
		var err error
		heightStr, err := io.Prompt("Terminal Height (25): ", 2, 15, 3)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Terminal Height (25): ", 2, 15, 3)
				continue
			}
			return nil, err
		}

		if heightStr == "" {
			terminalHeight = 25
			break
		}

		height, err := strconv.Atoi(heightStr)
		if err != nil || height < 1 || height > 25 {
			io.ShowTimedError("Height must be between 1 and 25.", 2, 16)
			io.ClearField("Terminal Height (25): ", 2, 15, 3)
			continue
		}

		terminalHeight = height
		break
	}

	// Store terminal preferences in userDetails
	userDetails["terminal_width"] = strconv.Itoa(terminalWidth)
	userDetails["terminal_height"] = strconv.Itoa(terminalHeight)

	row := 14 // Start after email field

	// Check for additional fields from RegistrationFields config
	if cfg.Configuration.NewUsers.RegistrationFields != nil {
		for fieldName, fieldConfig := range cfg.Configuration.NewUsers.RegistrationFields {
			if fieldConfig.Enabled {
				// Skip email since it's already prompted separately
				if strings.ToLower(fieldName) == "email" {
					continue
				}
				var promptText string
				var maxLength int

				// Customize prompt based on field name
				switch strings.ToLower(fieldName) {
				case "realname":
					promptText = "Real Name: "
					maxLength = 50
				case "location":
					promptText = "Location: "
					maxLength = 50
				case "phone":
					promptText = "Phone: "
					maxLength = 20
				case "website":
					promptText = "Website: "
					maxLength = 100
				default:
					promptText = fieldName + ": "
					maxLength = 50
				}

				var value string
				if fieldConfig.Required {
					// Required field - keep prompting until we get a value
					for {
						var err error
						value, err = io.Prompt(promptText, 2, row, maxLength)
						if err != nil {
							if err.Error() == "ESC_PRESSED" {
								if io.HandleEscQuit() {
									logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
									return nil, fmt.Errorf("registration cancelled")
								}
								// User chose to continue, clear field and retry
								io.ClearField(promptText, 2, row, maxLength)
								continue
							}
							return nil, err
						}

						if value == "" {
							io.ShowTimedError(fieldName+" is required.", 2, row+1)
							io.ClearField(promptText, 2, row, maxLength)
							continue
						} else {
							break
						}
					}
				} else {
					// Optional field
					var err error
					value, err = io.Prompt(promptText, 2, row, maxLength)
					if err != nil {
						if err.Error() == "ESC_PRESSED" {
							if io.HandleEscQuit() {
								logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
								return nil, fmt.Errorf("registration cancelled")
							}
							// User chose to continue, clear field and retry
							io.ClearField(promptText, 2, row, maxLength)
							continue
						}
						return nil, err
					}
				}

				if value != "" {
					userDetails[fieldName] = value
				}
				row++
			}
		}
	}

	// Display summary and confirm
	io.Print("\r\n\r\n" + ui.Ansi.Cyan + "Account Summary:" + ui.Ansi.Reset + "\r\n")
	io.Printf(ui.Ansi.WhiteHi+"Username: "+ui.Ansi.Reset+"%s\r\n", username)
	io.Printf(ui.Ansi.WhiteHi+"Password: "+ui.Ansi.Reset+"%s\r\n", strings.Repeat("*", len(password)))
	io.Printf(ui.Ansi.WhiteHi+"Email: "+ui.Ansi.Reset+"%s\r\n", email)
	io.Printf(ui.Ansi.WhiteHi+"Terminal Width: "+ui.Ansi.Reset+"%s\r\n", userDetails["terminal_width"])
	io.Printf(ui.Ansi.WhiteHi+"Terminal Height: "+ui.Ansi.Reset+"%s\r\n", userDetails["terminal_height"])
	for fieldName, value := range userDetails {
		if fieldName != "terminal_width" && fieldName != "terminal_height" {
			io.Printf(ui.Ansi.WhiteHi+"%s: "+ui.Ansi.Reset+"%s\r\n", fieldName, value)
		}
	}
	io.Print("\r\n")

	// Confirm creation
	var confirm string
	for {
		var err error
		confirm, err = io.Prompt("Create an account with this info? Y/N: ", 2, row, 1)
		if err != nil {
			if err.Error() == "ESC_PRESSED" {
				if io.HandleEscQuit() {
					logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
					return nil, fmt.Errorf("registration cancelled")
				}
				// User chose to continue, clear field and retry
				io.ClearField("Create an account with this info? Y/N: ", 2, row, 1)
				continue
			}
			return nil, err
		}

		confirm = strings.ToUpper(confirm)
		if confirm == "Y" {
			break
		} else if confirm == "N" {
			logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", "registration cancelled by user")
			return nil, fmt.Errorf("registration cancelled")
		} else {
			io.ShowTimedError("Please enter Y or N.", 2, row+1)
			io.ClearField("Create an account with this info? Y/N: ", 2, row, 1)
			continue
		}
	}

	// Create user account
	err := CreateUser(username, password, email, config.SecurityLevelRegular, userDetails) // Default security level for regular users
	if err != nil {
		logging.LogEvent(session.NodeNumber, username, session.IPAddress, "REGISTER_FAILED", err.Error())
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

	io.Printf(ui.Ansi.GreenHi+"\r\n\r\n Account created successfully. Welcome, %s!"+ui.Ansi.Reset, username)
	io.Pause()

	return user, nil
}
