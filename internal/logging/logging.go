package logging

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/robbiew/retrograde/internal/config"
)

// Type aliases for node management
type NodeManager = config.NodeManager
type NodeConnection = config.NodeConnection

var (
	nodeManager *NodeManager
	logMutex    sync.Mutex
)

// GetNodeManager returns the global node manager instance
func GetNodeManager() *NodeManager {
	return nodeManager
}

// InitializeNodeManager creates and initializes the global node manager
func InitializeNodeManager(maxNodes int) {
	nodeManager = &NodeManager{
		MaxNodes:    maxNodes,
		Connections: make(map[int]*NodeConnection),
		NextNode:    1,
	}
}

// AssignNodeWithLogging assigns a node to a connection, logs it, and returns the node number
func AssignNodeWithLogging(conn net.Conn, username string) int {
	nodeID := GetNodeManager().AssignNode(conn, username)
	if nodeID != -1 {
		ipAddr := conn.RemoteAddr().String()
		if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
			ipAddr = tcpAddr.IP.String()
		}
		LogEvent(nodeID, username, ipAddr, "CONNECT", fmt.Sprintf("Connected to node %d", nodeID))
	}
	return nodeID
}

// ReleaseNodeWithLogging releases a node, logs the disconnection, and marks it as disconnected
func ReleaseNodeWithLogging(nodeID int) {
	nm := GetNodeManager()
	if conn, exists := nm.Connections[nodeID]; exists && conn.Connected {
		LogEvent(nodeID, conn.Username, conn.IPAddress, "DISCONNECT",
			fmt.Sprintf("Disconnected from node %d after %v",
				nodeID, time.Since(conn.ConnectTime).Round(time.Second)))
		nm.ReleaseNode(nodeID)
	}
}

// LogEvent writes a log entry to the appropriate log file
func LogEvent(nodeID int, username, ipAddress, action, details string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Error creating logs directory: %v\n", err)
		return
	}

	// Generate log filename based on current date
	now := time.Now()
	logFileName := fmt.Sprintf("%s.log", now.Format("2006-01-02"))
	logFilePath := filepath.Join(logDir, logFileName)

	// Open log file (create if doesn't exist, append if it does)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	// Format the log entry
	timestamp := now.Format("2006-01-02 15:04:05")
	if username == "" {
		username = "Unknown"
	}

	logEntry := fmt.Sprintf("[%s] Node:%02d User:%s IP:%s Action:%s Details:%s\n",
		timestamp, nodeID, username, ipAddress, action, details)

	// Write to file
	if _, err := file.WriteString(logEntry); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
		return
	}

	// Also output to console for debugging
	fmt.Printf("LOG: %s", logEntry)
}

// LogConnection logs a connection event
func LogConnection(nodeID int, username, ipAddress string) {
	LogEvent(nodeID, username, ipAddress, "CONNECT", fmt.Sprintf("Connected to node %d", nodeID))
}

// LogDisconnection logs a disconnection event
func LogDisconnection(nodeID int, username, ipAddress string, duration time.Duration) {
	LogEvent(nodeID, username, ipAddress, "DISCONNECT",
		fmt.Sprintf("Disconnected from node %d after %v", nodeID, duration.Round(time.Second)))
}

// LogLogin logs a successful login event
func LogLogin(nodeID int, username, ipAddress string) {
	LogEvent(nodeID, username, ipAddress, "LOGIN", fmt.Sprintf("User %s logged in successfully", username))
}

// LogLogout logs a logout event
func LogLogout(nodeID int, username, ipAddress string) {
	LogEvent(nodeID, username, ipAddress, "LOGOUT", fmt.Sprintf("User %s logged out", username))
}

// LogLoginFailed logs a failed login attempt
func LogLoginFailed(nodeID int, username, ipAddress, reason string) {
	LogEvent(nodeID, username, ipAddress, "LOGIN_FAILED", fmt.Sprintf("Login failed for %s: %s", username, reason))
}

// LogApplicationSubmit logs when a user submits an application
func LogApplicationSubmit(nodeID int, username, ipAddress, networkType string) {
	LogEvent(nodeID, username, ipAddress, "APPLICATION_SUBMIT",
		fmt.Sprintf("User %s submitted %s application", username, networkType))
}

// LogApplicationEdit logs when a user or admin edits an application
func LogApplicationEdit(nodeID int, username, ipAddress, networkType, editor string) {
	if editor != username {
		LogEvent(nodeID, username, ipAddress, "APPLICATION_EDIT",
			fmt.Sprintf("Admin %s edited %s application for user %s", editor, networkType, username))
	} else {
		LogEvent(nodeID, username, ipAddress, "APPLICATION_EDIT",
			fmt.Sprintf("User %s edited their %s application", username, networkType))
	}
}

// LogApplicationApproval logs when an admin approves an application
func LogApplicationApproval(nodeID int, username, ipAddress, networkType, adminUser string) {
	LogEvent(nodeID, adminUser, ipAddress, "APPLICATION_APPROVED",
		fmt.Sprintf("Admin %s approved %s application for user %s", adminUser, networkType, username))
}

// LogAdminAction logs general admin actions
func LogAdminAction(nodeID int, adminUser, ipAddress, action, details string) {
	LogEvent(nodeID, adminUser, ipAddress, "ADMIN_ACTION", fmt.Sprintf("%s: %s", action, details))
}

// GetLogFilePath returns the path to today's log file
func GetLogFilePath() string {
	logDir := "logs"
	now := time.Now()
	logFileName := fmt.Sprintf("%s.log", now.Format("2006-01-02"))
	return filepath.Join(logDir, logFileName)
}

// ListLogFiles returns a list of all available log files
func ListLogFiles() ([]string, error) {
	logDir := "logs"
	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			logFiles = append(logFiles, file.Name())
		}
	}
	return logFiles, nil
}
