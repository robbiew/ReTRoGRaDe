package config

// GetDefaultConfig returns a Config struct populated with default values
// Used when creating a new database for the first time
func GetDefaultConfig() *Config {
	cfg := &Config{}

	// Configuration.Paths
	cfg.Configuration.Paths.Database = "C:\\retrograde\\data"
	cfg.Configuration.Paths.FileBase = "C:\\retrograde\\files"
	cfg.Configuration.Paths.Logs = "C:\\retrograde\\logs"
	cfg.Configuration.Paths.MessageBase = "C:\\retrograde\\msgs"
	cfg.Configuration.Paths.System = "C:\\retrograde"
	cfg.Configuration.Paths.Themes = "C:\\retrograde\\themes"

	// Configuration.General
	cfg.Configuration.General.BBSLocation = "Your City, State"
	cfg.Configuration.General.BBSName = "Retrograde BBS"
	cfg.Configuration.General.DefaultTheme = "default"
	cfg.Configuration.General.StartMenu = "prelogin"
	cfg.Configuration.General.SysOpName = "SysOp"
	cfg.Configuration.General.SysOpTimeoutExempt = true
	cfg.Configuration.General.SystemPassword = "CHANGEME"
	cfg.Configuration.General.TimeoutMinutes = 3

	// Configuration.NewUsers
	cfg.Configuration.NewUsers.AllowNew = true
	cfg.Configuration.NewUsers.AskLocation = true
	cfg.Configuration.NewUsers.AskRealName = true
	cfg.Configuration.NewUsers.RegistrationFormEnabledFields = []string{
		"Username",
		"Password",
		"Email",
		"RealName",
		"Location",
	}
	cfg.Configuration.NewUsers.RegistrationFields = map[string]RegistrationFieldConfig{
		"Email": {
			Enabled:  true,
			Required: true,
		},
		"RealName": {
			Enabled:  true,
			Required: true,
		},
		"Location": {
			Enabled:  true,
			Required: true,
		},
	}
	cfg.Configuration.NewUsers.SysopQuestionEnabled = false
	cfg.Configuration.NewUsers.SysopFields = map[string]RegistrationFieldConfig{
		"BBSName": {
			Enabled:  true,
			Required: true,
		},
		"BBSURL": {
			Enabled:  true,
			Required: false,
		},
		"BBSPort": {
			Enabled:  true,
			Required: false,
		},
		"BBSSoftware": {
			Enabled:  true,
			Required: false,
		},
		"BBSLocation": {
			Enabled:  true,
			Required: false,
		},
	}
	cfg.Configuration.NewUsers.FormLayout = map[string]FormLayoutConfig{}

	// Configuration.Auth
	cfg.Configuration.Auth.UseSQLite = true
	cfg.Configuration.Auth.JSONFallback = false
	cfg.Configuration.Auth.MaxFailedAttempts = 5
	cfg.Configuration.Auth.AccountLockMinutes = 15
	cfg.Configuration.Auth.PasswordAlgorithm = "sha256"

	// Servers.GeneralSettings
	cfg.Servers.GeneralSettings.MaxConnectionsPerIP = 5
	cfg.Servers.GeneralSettings.MaxNodes = 10

	// Servers.Telnet
	cfg.Servers.Telnet.Active = true
	cfg.Servers.Telnet.Port = 2323

	// Servers.Security.RateLimits
	cfg.Servers.Security.RateLimits.Enabled = true
	cfg.Servers.Security.RateLimits.WindowMinutes = 15

	// Servers.Security.LocalLists
	cfg.Servers.Security.LocalLists.BlacklistEnabled = true
	cfg.Servers.Security.LocalLists.BlacklistFile = "security/blacklist.txt"
	cfg.Servers.Security.LocalLists.WhitelistEnabled = false
	cfg.Servers.Security.LocalLists.WhitelistFile = "security/whitelist.txt"

	// Servers.Security.ExternalLists
	cfg.Servers.Security.ExternalLists.Enabled = true
	cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs = []string{
		"https://raw.githubusercontent.com/stamparm/ipsum/master/ipsum.txt",
	}

	// Servers.Security.GeoBlock
	cfg.Servers.Security.GeoBlock.AllowedCountries = []string{}
	cfg.Servers.Security.GeoBlock.BlockedCountries = []string{"CN", "RU", "KP", "IR"}
	cfg.Servers.Security.GeoBlock.BlocklistUpdateHours = 6
	cfg.Servers.Security.GeoBlock.GeoAPIKey = "your_api_key_here"
	cfg.Servers.Security.GeoBlock.GeoAPIProvider = "ipapi"
	cfg.Servers.Security.GeoBlock.GeoBlockEnabled = false
	cfg.Servers.Security.GeoBlock.ThreatIntelEnabled = true

	// Servers.Security.Logs
	cfg.Servers.Security.Logs.LogBlockedAttempts = true
	cfg.Servers.Security.Logs.LogSecurityEvents = true
	cfg.Servers.Security.Logs.SecurityLogFile = "logs/security.log"

	// Other.Discord
	cfg.Other.Discord.Enabled = false
	cfg.Other.Discord.InviteURL = "https://discord.gg/your-invite"
	cfg.Other.Discord.Title = "New User Application:"
	cfg.Other.Discord.Username = "Retrograde Bot"
	cfg.Other.Discord.WebhookURL = "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"

	return cfg
}
