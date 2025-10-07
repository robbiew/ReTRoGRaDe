package security

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/robbiew/retrograde/internal/config"
)

var (
	securityManager *config.SecurityManager
	securityMutex   sync.RWMutex
)

// InitializeSecurity initializes the security system
func InitializeSecurity(cfg *config.Config) {
	securityMutex.Lock()
	defer securityMutex.Unlock()

	securityManager = &config.SecurityManager{
		Config:            &cfg.Servers.Security,
		ConnectionTracker: make(map[string]*config.ConnectionAttempt),
		Blacklist:         make(map[string]*config.IPListEntry),
		Whitelist:         make(map[string]*config.IPListEntry),
		GeoCache:          make(map[string]*config.GeoLocation),
		ThreatIntelCache:  make(map[string]bool),
		LastUpdate:        time.Now(),
	}

	// Load IP lists from files
	if cfg.Servers.Security.LocalLists.BlacklistEnabled {
		loadIPList(cfg.Servers.Security.LocalLists.BlacklistFile, securityManager.Blacklist)
	}
	if cfg.Servers.Security.LocalLists.WhitelistEnabled {
		loadIPList(cfg.Servers.Security.LocalLists.WhitelistFile, securityManager.Whitelist)
	}

	// Load external threat intelligence
	if cfg.Servers.Security.GeoBlock.ThreatIntelEnabled {
		go updateThreatIntelligence()

		// Test blocklist connectivity asynchronously (don't block startup)
		go func() {
			fmt.Println("Testing external blocklist connectivity...")
			for _, url := range cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs {
				if err := testBlocklistURL(url); err != nil {
					fmt.Printf("⚠ WARNING: Cannot access blocklist %s: %v\n", url, err)
					logSecurityEvent("BLOCKLIST_ERROR", "startup", fmt.Sprintf("Failed to access %s: %v", url, err), "LOG")
				} else {
					fmt.Printf("✓ Blocklist accessible: %s\n", url)
				}
			}
		}()
	}

	fmt.Printf("Security system initialized - Rate limiting: %v, Geo-blocking: %v, Threat intel: %v\n",
		cfg.Servers.Security.RateLimits.Enabled, cfg.Servers.Security.GeoBlock.GeoBlockEnabled, cfg.Servers.Security.GeoBlock.ThreatIntelEnabled)
}

// CheckConnectionSecurity validates an incoming connection
func CheckConnectionSecurity(conn net.Conn, cfg *config.Config) (bool, string) {
	if securityManager == nil {
		return true, "" // Allow if security not initialized
	}

	ipAddr := getIPFromConn(conn)

	securityMutex.RLock()
	defer securityMutex.RUnlock()

	// Check whitelist first (if enabled)
	if cfg.Servers.Security.LocalLists.WhitelistEnabled {
		if isIPInList(ipAddr, securityManager.Whitelist) {
			logSecurityEvent("WHITELIST_ALLOW", ipAddr, "IP in whitelist", "ALLOW")
			return true, ""
		}
	}

	// Check blacklist
	if cfg.Servers.Security.LocalLists.BlacklistEnabled {
		if entry := getIPListEntry(ipAddr, securityManager.Blacklist); entry != nil {
			reason := fmt.Sprintf("IP blacklisted: %s", entry.Reason)
			logSecurityEvent("BLACKLIST_BLOCK", ipAddr, reason, "REJECT")
			return false, reason
		}
	}

	// Check rate limiting
	if cfg.Servers.Security.RateLimits.Enabled {
		if blocked, reason := checkRateLimit(ipAddr, cfg); blocked {
			logSecurityEvent("RATE_LIMITED", ipAddr, reason, "REJECT")
			return false, reason
		}
	}

	// Check geographic blocking
	if cfg.Servers.Security.GeoBlock.GeoBlockEnabled {
		if blocked, reason := checkGeoBlocking(ipAddr, cfg); blocked {
			logSecurityEvent("GEO_BLOCKED", ipAddr, reason, "REJECT")
			return false, reason
		}
	}

	// Check threat intelligence
	if cfg.Servers.Security.GeoBlock.ThreatIntelEnabled {
		if isThreatIP(ipAddr) {
			reason := "IP found in threat intelligence feeds"
			logSecurityEvent("THREAT_BLOCKED", ipAddr, reason, "REJECT")
			return false, reason
		}
	}

	// Track successful connection
	trackConnection(ipAddr)
	logSecurityEvent("CONNECTION_ALLOWED", ipAddr, "Passed all security checks", "ALLOW")

	return true, ""
}

// Rate limiting functions
func checkRateLimit(ipAddr string, cfg *config.Config) (bool, string) {
	now := time.Now()
	windowStart := now.Add(-time.Duration(cfg.Servers.Security.RateLimits.WindowMinutes) * time.Minute)

	// Get or create connection attempt record
	attempt, exists := securityManager.ConnectionTracker[ipAddr]
	if !exists {
		attempt = &config.ConnectionAttempt{
			IPAddress:    ipAddr,
			AttemptTime:  now,
			AttemptCount: 1,
			LastAttempt:  now,
		}
		securityManager.ConnectionTracker[ipAddr] = attempt
		return false, ""
	}

	// Check if IP is currently blocked
	if attempt.IsBlocked && now.Before(attempt.BlockExpires) {
		return true, fmt.Sprintf("IP temporarily blocked until %s", attempt.BlockExpires.Format("15:04:05"))
	}

	// Reset block status if expired
	if attempt.IsBlocked && now.After(attempt.BlockExpires) {
		attempt.IsBlocked = false
		attempt.AttemptCount = 0
	}

	// Count attempts in current window
	if attempt.LastAttempt.After(windowStart) {
		attempt.AttemptCount++
	} else {
		attempt.AttemptCount = 1
		attempt.AttemptTime = now
	}

	attempt.LastAttempt = now

	// Check if limit exceeded
	if attempt.AttemptCount > cfg.Servers.GeneralSettings.MaxConnectionsPerIP {
		blockDuration := time.Duration(cfg.Servers.Security.RateLimits.WindowMinutes) * time.Minute
		attempt.IsBlocked = true
		attempt.BlockExpires = now.Add(blockDuration)
		attempt.BlockReason = "Rate limit exceeded"

		// Auto-add to temporary blacklist
		AddToBlacklist(ipAddr, "Auto-blocked: Rate limit exceeded", "auto", &attempt.BlockExpires)

		return true, fmt.Sprintf("Rate limit exceeded: %d connections in %d minutes",
			attempt.AttemptCount, cfg.Servers.Security.RateLimits.WindowMinutes)
	}

	return false, ""
}

// Geographic blocking functions
func checkGeoBlocking(ipAddr string, cfg *config.Config) (bool, string) {
	geo := getGeoLocation(ipAddr, cfg)
	if geo == nil {
		return false, "" // Allow if geo lookup fails
	}

	// Check blocked countries
	for _, blockedCountry := range cfg.Servers.Security.GeoBlock.BlockedCountries {
		if strings.EqualFold(geo.CountryCode, blockedCountry) || strings.EqualFold(geo.Country, blockedCountry) {
			return true, fmt.Sprintf("Connection from blocked country: %s (%s)", geo.Country, geo.CountryCode)
		}
	}

	// Check allowed countries (if specified, only these are allowed)
	if len(cfg.Servers.Security.GeoBlock.AllowedCountries) > 0 {
		allowed := false
		for _, allowedCountry := range cfg.Servers.Security.GeoBlock.AllowedCountries {
			if strings.EqualFold(geo.CountryCode, allowedCountry) || strings.EqualFold(geo.Country, allowedCountry) {
				allowed = true
				break
			}
		}
		if !allowed {
			return true, fmt.Sprintf("Connection from non-allowed country: %s (%s)", geo.Country, geo.CountryCode)
		}
	}

	return false, ""
}

// Geolocation API integration
func getGeoLocation(ipAddr string, cfg *config.Config) *config.GeoLocation {
	// Check cache first
	if cached, exists := securityManager.GeoCache[ipAddr]; exists {
		if time.Since(cached.CachedAt) < 24*time.Hour {
			return cached
		}
	}

	// Skip private IPs
	if isPrivateIP(ipAddr) {
		geo := &config.GeoLocation{
			IPAddress:   ipAddr,
			Country:     "Private",
			CountryCode: "XX",
			CachedAt:    time.Now(),
		}
		securityManager.GeoCache[ipAddr] = geo
		return geo
	}

	var geo *config.GeoLocation

	switch cfg.Servers.Security.GeoBlock.GeoAPIProvider {
	case "ipstack":
		geo = getGeoFromIPStack(ipAddr, cfg.Servers.Security.GeoBlock.GeoAPIKey)
	case "nerds":
		geo = getGeoFromNerdsDK(ipAddr)
	default:
		geo = getGeoFromIPAPI(ipAddr) // Free fallback
	}

	if geo != nil {
		geo.CachedAt = time.Now()
		securityManager.GeoCache[ipAddr] = geo
	}

	return geo
}

// IPStack API integration
func getGeoFromIPStack(ipAddr, apiKey string) *config.GeoLocation {
	url := fmt.Sprintf("http://api.ipstack.com/%s?access_key=%s&format=1", ipAddr, apiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		CountryCode string `json:"country_code"`
		CountryName string `json:"country_name"`
		City        string `json:"city"`
		ISP         string `json:"isp"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	return &config.GeoLocation{
		IPAddress:   ipAddr,
		Country:     result.CountryName,
		CountryCode: result.CountryCode,
		City:        result.City,
		ISP:         result.ISP,
	}
}

// Free IP-API integration
func getGeoFromIPAPI(ipAddr string) *config.GeoLocation {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipAddr)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		CountryCode string `json:"countryCode"`
		Country     string `json:"country"`
		City        string `json:"city"`
		ISP         string `json:"isp"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	if result.Status != "success" {
		return nil
	}

	return &config.GeoLocation{
		IPAddress:   ipAddr,
		Country:     result.Country,
		CountryCode: result.CountryCode,
		City:        result.City,
		ISP:         result.ISP,
	}
}

// Nerds.dk API integration
func getGeoFromNerdsDK(ipAddr string) *config.GeoLocation {
	url := fmt.Sprintf("https://nerds.dk/geoip/%s", ipAddr)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	// Parse simple text response: "Country: CountryCode"
	parts := strings.Split(strings.TrimSpace(string(body)), ":")
	if len(parts) != 2 {
		return nil
	}

	return &config.GeoLocation{
		IPAddress:   ipAddr,
		Country:     strings.TrimSpace(parts[0]),
		CountryCode: strings.TrimSpace(parts[1]),
	}
}

// Threat intelligence functions
func updateThreatIntelligence() {
	cfg := securityManager.Config

	for {
		fmt.Println("Updating threat intelligence feeds...")

		for _, url := range cfg.ExternalLists.ExternalBlocklistURLs {
			updateBlocklistFromURL(url)
		}

		securityManager.LastUpdate = time.Now()
		fmt.Printf("Threat intelligence updated at %s\n", securityManager.LastUpdate.Format("2006-01-02 15:04:05"))

		// Wait for next update
		time.Sleep(time.Duration(cfg.GeoBlock.BlocklistUpdateHours) * time.Hour)
	}
}

func updateBlocklistFromURL(url string) {
	// Use longer timeout for slow networks and TLS handshakes
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching blocklist from %s: %v\n", url, err)
		logSecurityEvent("BLOCKLIST_ERROR", "update", fmt.Sprintf("Failed to fetch %s: %v", url, err), "LOG")
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	count := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse different formats
		ip := extractIPFromLine(line)
		if ip != "" && isValidIP(ip) {
			securityManager.ThreatIntelCache[ip] = true
			count++
		}
	}

	fmt.Printf("Loaded %d threat IPs from %s\n", count, url)
}

// IP list management
func loadIPList(filename string, list map[string]*config.IPListEntry) {
	if filename == "" {
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Error opening IP list file %s: %v\n", filename, err)
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, " ")
		ip := parts[0]
		reason := "Manual entry"
		if len(parts) > 1 {
			reason = strings.Join(parts[1:], " ")
		}

		if isValidIP(ip) {
			list[ip] = &config.IPListEntry{
				IPAddress: ip,
				AddedTime: time.Now(),
				Reason:    reason,
				Source:    "file",
			}
			count++
		}
	}

	fmt.Printf("Loaded %d IPs from %s\n", count, filename)
}

// Utility functions
func getIPFromConn(conn net.Conn) string {
	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		return tcpAddr.IP.String()
	}
	return strings.Split(conn.RemoteAddr().String(), ":")[0]
}

func isIPInList(ipAddr string, list map[string]*config.IPListEntry) bool {
	return getIPListEntry(ipAddr, list) != nil
}

func getIPListEntry(ipAddr string, list map[string]*config.IPListEntry) *config.IPListEntry {
	// Direct IP match
	if entry, exists := list[ipAddr]; exists {
		return entry
	}

	// CIDR range match
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return nil
	}

	for _, entry := range list {
		if entry.CIDR != "" {
			_, cidr, err := net.ParseCIDR(entry.CIDR)
			if err == nil && cidr.Contains(ip) {
				return entry
			}
		}
	}

	return nil
}

// AddToBlacklist adds an IP to the blacklist (exported for use by other packages)
func AddToBlacklist(ipAddr, reason, source string, expiresAt *time.Time) {
	securityMutex.Lock()
	defer securityMutex.Unlock()

	securityManager.Blacklist[ipAddr] = &config.IPListEntry{
		IPAddress: ipAddr,
		AddedTime: time.Now(),
		Reason:    reason,
		Source:    source,
		ExpiresAt: expiresAt,
	}

	// Also save to blacklist file if it's a permanent ban (no expiration)
	if expiresAt == nil {
		saveIPToBlacklistFile(ipAddr, reason, source)
	}
}

// saveIPToBlacklistFile appends an IP to the blacklist file for permanent storage
func saveIPToBlacklistFile(ipAddr, reason, source string) {
	if securityManager.Config.LocalLists.BlacklistFile == "" {
		return
	}

	// Open file for appending
	file, err := os.OpenFile(securityManager.Config.LocalLists.BlacklistFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error writing to blacklist file: %v\n", err)
		return
	}
	defer file.Close()

	// Write entry in format: IP reason (source - timestamp)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("%s %s (%s - %s)\n", ipAddr, reason, source, timestamp)

	if _, err := file.WriteString(entry); err != nil {
		fmt.Printf("Error writing to blacklist file: %v\n", err)
		return
	}

	fmt.Printf("Added IP %s to permanent blacklist: %s\n", ipAddr, reason)
	logSecurityEvent("BLACKLIST_ADDED", ipAddr, fmt.Sprintf("Permanently blacklisted: %s", reason), "BLOCKED")
}

func trackConnection(ipAddr string) {
	if _, exists := securityManager.ConnectionTracker[ipAddr]; !exists {
		securityManager.ConnectionTracker[ipAddr] = &config.ConnectionAttempt{
			IPAddress:    ipAddr,
			AttemptTime:  time.Now(),
			LastAttempt:  time.Now(),
			AttemptCount: 1,
		}
	}
}

func isThreatIP(ipAddr string) bool {
	_, exists := securityManager.ThreatIntelCache[ipAddr]
	return exists
}

func isPrivateIP(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}

	for _, cidrRange := range privateRanges {
		_, cidr, err := net.ParseCIDR(cidrRange)
		if err == nil && cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// testBlocklistURL tests if a blocklist URL is accessible at startup
func testBlocklistURL(url string) error {
	// Use longer timeout for TLS handshake and slow networks
	client := &http.Client{Timeout: 30 * time.Second}

	// Use GET instead of HEAD - some servers don't handle HEAD well
	// Read only first few bytes to verify connectivity without downloading full file
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read first 100 bytes to verify the connection works
	buffer := make([]byte, 100)
	_, err = resp.Body.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read response: %v", err)
	}

	return nil
}

func isValidIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}

func extractIPFromLine(line string) string {
	// Try to extract IP from various formats
	ipRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	matches := ipRegex.FindStringSubmatch(line)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// Security logging
func logSecurityEvent(eventType, ipAddr, reason, action string) {
	cfg := securityManager.Config

	if !cfg.Logs.LogSecurityEvents {
		return
	}

	event := config.SecurityEvent{
		Timestamp: time.Now(),
		IPAddress: ipAddr,
		EventType: eventType,
		Reason:    reason,
		Action:    action,
	}

	// Console logging
	fmt.Printf("[SECURITY] %s - %s: %s (%s) -> %s\n",
		event.Timestamp.Format("2006-01-02 15:04:05"),
		event.EventType, event.IPAddress, event.Reason, event.Action)

	// File logging
	if cfg.Logs.SecurityLogFile != "" {
		logSecurityEventToFile(event, cfg.Logs.SecurityLogFile)
	}
}

func logSecurityEventToFile(event config.SecurityEvent, filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	logLine := fmt.Sprintf("%s [%s] %s: %s (%s) -> %s\n",
		event.Timestamp.Format("2006-01-02 15:04:05"),
		event.EventType, event.IPAddress, event.Reason, event.Action, event.Details)

	file.WriteString(logLine)
}

// Cleanup expired entries
func cleanupSecurityData() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cleanupExpiredEntries()
	}
}

func cleanupExpiredEntries() {
	securityMutex.Lock()
	defer securityMutex.Unlock()

	now := time.Now()

	// Cleanup expired blacklist entries
	for ip, entry := range securityManager.Blacklist {
		if entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
			delete(securityManager.Blacklist, ip)
		}
	}

	// Cleanup old connection attempts (older than 24 hours)
	cutoff := now.Add(-24 * time.Hour)
	for ip, attempt := range securityManager.ConnectionTracker {
		if attempt.LastAttempt.Before(cutoff) && !attempt.IsBlocked {
			delete(securityManager.ConnectionTracker, ip)
		}
	}

	// Cleanup old geo cache (older than 7 days)
	geoCutoff := now.Add(-7 * 24 * time.Hour)
	for ip, geo := range securityManager.GeoCache {
		if geo.CachedAt.Before(geoCutoff) {
			delete(securityManager.GeoCache, ip)
		}
	}
}

// CleanupSecurityData exports cleanupSecurityData for external use
func CleanupSecurityData() {
	cleanupSecurityData()
}

// GetIPFromConn exports getIPFromConn for external use
func GetIPFromConn(conn net.Conn) string {
	return getIPFromConn(conn)
}
