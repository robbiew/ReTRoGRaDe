package config

import (
	"net"
	"time"
)

// Security Level Constants
const (
	SecurityLevelGuest   = 0   // Unauthenticated users
	SecurityLevelRegular = 10  // Standard registered users
	SecurityLevelSysOp   = 100 // System operators with full access
	SecurityLevelAdmin   = 255 // Special admin level (legacy compatibility)
)

// Config struct to hold settings from the ini file with hierarchical structure
type Config struct {
	Configuration ConfigurationSection
	Servers       ServersSection
	Networking    NetworkingSection
	Editors       EditorsSection
	Events        EventsSection
	Other         OtherSection
}

// ConfigurationSection holds all Configuration.* settings
type ConfigurationSection struct {
	Paths    PathsConfig
	General  GeneralConfig
	NewUsers NewUsersConfig
	Auth     AuthConfig
}

// PathsConfig holds system paths
type PathsConfig struct {
	Database    string
	FileBase    string
	Logs        string
	MessageBase string
	System      string
	Themes      string
	Security    string
}

// GeneralConfig holds general BBS settings
type GeneralConfig struct {
	BBSLocation        string
	BBSName            string
	DefaultTheme       string
	StartMenu          string
	SysOpName          string
	SysOpTimeoutExempt bool
	SystemPassword     string
	TimeoutMinutes     int
}

type RegistrationFieldConfig struct {
	Enabled  bool
	Required bool
}

type FormLayoutConfig struct {
	Row int
	Col int
}

// NewUsersConfig holds new user settings
type NewUsersConfig struct {
	AllowNew                      bool
	AskLocation                   bool
	AskRealName                   bool
	RegistrationFormEnabledFields []string
	RegistrationFields            map[string]RegistrationFieldConfig
	SysopQuestionEnabled          bool
	SysopFields                   map[string]RegistrationFieldConfig
	FormLayout                    map[string]FormLayoutConfig
}

// AuthConfig holds authentication persistence and policy settings
type AuthConfig struct {
	UseSQLite          bool
	JSONFallback       bool
	MaxFailedAttempts  int
	AccountLockMinutes int
	PasswordAlgorithm  string
}

// ServersSection holds all Servers.* settings
type ServersSection struct {
	GeneralSettings GeneralServerSettings
	Telnet          TelnetConfig
	Security        SecurityConfig
}

// GeneralServerSettings holds general server settings
type GeneralServerSettings struct {
	MaxConnectionsPerIP int
	MaxNodes            int
}

// TelnetConfig holds telnet server settings
type TelnetConfig struct {
	Active bool
	Port   int
}

// SecurityConfig holds all security-related settings
type SecurityConfig struct {
	RateLimits    RateLimitsConfig
	LocalLists    LocalListsConfig
	ExternalLists ExternalListsConfig
	GeoBlock      GeoBlockConfig
	Logs          SecurityLogsConfig
}

// RateLimitsConfig holds rate limiting settings
type RateLimitsConfig struct {
	Enabled       bool
	WindowMinutes int
}

// LocalListsConfig holds local blacklist/whitelist settings
type LocalListsConfig struct {
	BlacklistEnabled bool
	BlacklistFile    string
	WhitelistEnabled bool
	WhitelistFile    string
}

// ExternalListsConfig holds external blocklist settings
type ExternalListsConfig struct {
	Enabled               bool
	ExternalBlocklistURLs []string
}

// GeoBlockConfig holds geographic blocking settings
type GeoBlockConfig struct {
	AllowedCountries     []string
	BlockedCountries     []string
	BlocklistUpdateHours int
	GeoAPIKey            string
	GeoAPIProvider       string
	GeoBlockEnabled      bool
	ThreatIntelEnabled   bool
}

// SecurityLogsConfig holds security logging settings
type SecurityLogsConfig struct {
	LogBlockedAttempts bool
	LogSecurityEvents  bool
	SecurityLogFile    string
}

// NetworkingSection holds networking configuration
type NetworkingSection struct {
	// Future network settings
}

// EditorsSection holds editor configurations
type EditorsSection struct {
	UserEditor  EditorConfig
	SecLevels   EditorConfig
	MessageBase EditorConfig
	FileBase    EditorConfig
	Menus       EditorConfig
	Logon       EditorConfig
	Logoff      EditorConfig
	Timed       EditorConfig
}

// EditorConfig holds configuration for a specific editor
type EditorConfig struct {
	// Future editor settings
}

// EventsSection holds event configurations
type EventsSection struct {
	// Future event settings
}

// OtherSection holds miscellaneous settings
type OtherSection struct {
	Discord DiscordConfig
}

// DiscordConfig holds Discord integration settings
type DiscordConfig struct {
	Enabled    bool
	InviteURL  string
	Title      string
	Username   string
	WebhookURL string
}

// Struct for storing drop file data (adapted for telnet sessions)
type DropFileData struct {
	CommType      int
	CommHandle    int
	BaudRate      int
	BBSID         string
	UserRecordPos int
	RealName      string
	Alias         string
	SecurityLevel int
	TimeLeft      int
	Emulation     int
	NodeNum       int
}

// Struct to hold details about the program's initial state
type ProgramState struct {
	TerminalHeight int
	TerminalWidth  int
	StartTime      time.Time
}

// TelnetSession holds connection state for each telnet user
type TelnetSession struct {
	Alias         string
	SecurityLevel int
	TimeLeft      int
	StartTime     time.Time
	LastActivity  time.Time
	NodeNumber    int
	IPAddress     string
	Connected     bool
	Conn          net.Conn // Add connection reference for timeout handling
	Width         int      // Terminal width from NAWS negotiation
	Height        int      // Terminal height from NAWS negotiation
}

// NodeConnection tracks individual connection details
type NodeConnection struct {
	NodeNumber   int
	Username     string
	IPAddress    string
	ConnectTime  time.Time
	LastActivity time.Time
	Connected    bool
}

// NodeManager manages all active connections
type NodeManager struct {
	MaxNodes    int
	Connections map[int]*NodeConnection
	NextNode    int
}

// LogEntry represents a log entry for the system
type LogEntry struct {
	Timestamp time.Time
	NodeID    int
	Username  string
	IPAddress string
	Action    string
	Details   string
}

// Security Management Structures

// ConnectionAttempt tracks connection attempts per IP
type ConnectionAttempt struct {
	IPAddress    string
	AttemptTime  time.Time
	AttemptCount int
	LastAttempt  time.Time
	IsBlocked    bool
	BlockReason  string
	BlockExpires time.Time
}

// IPListEntry represents an entry in blacklist or whitelist
type IPListEntry struct {
	IPAddress string
	CIDR      string
	AddedTime time.Time
	Reason    string
	Source    string // "manual", "auto", "external"
	ExpiresAt *time.Time
}

// GeoLocation holds IP geolocation data
type GeoLocation struct {
	IPAddress   string
	Country     string
	CountryCode string
	City        string
	ISP         string
	CachedAt    time.Time
}

// SecurityManager manages all security operations
type SecurityManager struct {
	Config            *SecurityConfig
	ConnectionTracker map[string]*ConnectionAttempt
	Blacklist         map[string]*IPListEntry
	Whitelist         map[string]*IPListEntry
	GeoCache          map[string]*GeoLocation
	ThreatIntelCache  map[string]bool
	LastUpdate        time.Time
}

// SecurityEvent represents a security-related event for logging
type SecurityEvent struct {
	Timestamp time.Time
	IPAddress string
	EventType string // "BLOCKED", "RATE_LIMITED", "GEO_BLOCKED", "THREAT_BLOCKED"
	Reason    string
	Details   string
	Action    string // "REJECT", "ALLOW", "LOG"
}

// NodeManager methods

// GetAvailableNode returns the next available node number, or -1 if all nodes are full
func (nm *NodeManager) GetAvailableNode() int {
	for i := 1; i <= nm.MaxNodes; i++ {
		if conn, exists := nm.Connections[i]; !exists || !conn.Connected {
			return i
		}
	}
	return -1
}

// AddConnection adds a connection to the node manager
func (nm *NodeManager) AddConnection(nodeNum int, conn net.Conn, username string) {
	ipAddr := conn.RemoteAddr().String()
	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		ipAddr = tcpAddr.IP.String()
	}

	nm.Connections[nodeNum] = &NodeConnection{
		NodeNumber:   nodeNum,
		Username:     username,
		IPAddress:    ipAddr,
		ConnectTime:  time.Now(),
		LastActivity: time.Now(),
		Connected:    true,
	}
}

// RemoveConnection removes a connection from the node manager
func (nm *NodeManager) RemoveConnection(nodeNum int) {
	if conn, exists := nm.Connections[nodeNum]; exists {
		conn.Connected = false
	}
}

// GetConnection returns the connection for a specific node
func (nm *NodeManager) GetConnection(nodeNum int) *NodeConnection {
	return nm.Connections[nodeNum]
}

// GetActiveNodes returns a list of active node numbers
func (nm *NodeManager) GetActiveNodes() []int {
	var active []int
	for nodeNum, conn := range nm.Connections {
		if conn.Connected {
			active = append(active, nodeNum)
		}
	}
	return active
}

// IsNodeActive returns whether a node is currently active
func (nm *NodeManager) IsNodeActive(nodeNum int) bool {
	if conn, exists := nm.Connections[nodeNum]; exists {
		return conn.Connected
	}
	return false
}

// GetNodeCount returns the total number of configured nodes
func (nm *NodeManager) GetNodeCount() int {
	return nm.MaxNodes
}

// AssignNode assigns a node to a connection and returns the node number
func (nm *NodeManager) AssignNode(conn net.Conn, username string) int {
	nodeID := nm.GetAvailableNode()
	if nodeID == -1 {
		return -1 // No nodes available
	}

	ipAddr := conn.RemoteAddr().String()
	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		ipAddr = tcpAddr.IP.String()
	}

	nm.Connections[nodeID] = &NodeConnection{
		NodeNumber:   nodeID,
		Username:     username,
		IPAddress:    ipAddr,
		ConnectTime:  time.Now(),
		LastActivity: time.Now(),
		Connected:    true,
	}

	return nodeID
}

// ReleaseNode releases a node and marks it as disconnected
func (nm *NodeManager) ReleaseNode(nodeID int) {
	if conn, exists := nm.Connections[nodeID]; exists {
		conn.Connected = false
	}
}

// UpdateActivity updates the last activity time for a node
func (nm *NodeManager) UpdateActivity(nodeID int) {
	if conn, exists := nm.Connections[nodeID]; exists && conn.Connected {
		conn.LastActivity = time.Now()
	}
}

// GetNodeInfo returns information about a specific node
func (nm *NodeManager) GetNodeInfo(nodeID int) (*NodeConnection, bool) {
	conn, exists := nm.Connections[nodeID]
	return conn, exists && conn.Connected
}

// GetActiveNodeCount returns the number of currently active nodes
func (nm *NodeManager) GetActiveNodeCount() int {
	count := 0
	for _, conn := range nm.Connections {
		if conn.Connected {
			count++
		}
	}
	return count
}
