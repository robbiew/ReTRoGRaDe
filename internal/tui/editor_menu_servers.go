package tui

import (
	"fmt"
	"strings"

	"github.com/robbiew/retrograde/internal/config"
)

func serversMenu(cfg *config.Config) MenuCategory {
	return MenuCategory{
		ID:     "servers",
		Label:  "Servers",
		HotKey: 'S',
		SubItems: []SubmenuItem{
			{
				ID:       "general-settings",
				Label:    "General Settings",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "max-nodes",
						Label:    "Max Nodes",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.general.max_nodes",
							Label:     "Max Nodes",
							ValueType: IntValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.GeneralSettings.MaxNodes },
								SetValue: func(v interface{}) error {
									cfg.Servers.GeneralSettings.MaxNodes = v.(int)
									return nil
								},
							},
							HelpText: "Maximum number of concurrent connections",
							Validation: func(v interface{}) error {
								nodes := v.(int)
								if nodes <= 0 {
									return fmt.Errorf("max nodes must be positive")
								}
								return nil
							},
						},
					},
					{
						ID:       "max-connections-per-ip",
						Label:    "Max Connections Per IP",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.general.max_connections_per_ip",
							Label:     "Max Connect",
							ValueType: IntValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.GeneralSettings.MaxConnectionsPerIP },
								SetValue: func(v interface{}) error {
									cfg.Servers.GeneralSettings.MaxConnectionsPerIP = v.(int)
									return nil
								},
							},
							HelpText: "Maximum connections allowed per IP address",
						},
					},
				},
			},
			{
				ID:       "telnet-server",
				Label:    "Telnet Server",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "telnet-active",
						Label:    "Active",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.telnet.active",
							Label:     "Active",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Telnet.Active },
								SetValue: func(v interface{}) error {
									cfg.Servers.Telnet.Active = v.(bool)
									return nil
								},
							},
							HelpText: "Enable/disable telnet server",
						},
					},
					{
						ID:       "telnet-port",
						Label:    "Port",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.telnet.port",
							Label:     "Telnet Port",
							ValueType: PortValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Telnet.Port },
								SetValue: func(v interface{}) error {
									cfg.Servers.Telnet.Port = v.(int)
									return nil
								},
							},
							HelpText: "TCP port for telnet connections (1-65535)",
							Validation: func(v interface{}) error {
								port := v.(int)
								if port < 1 || port > 65535 {
									return fmt.Errorf("port must be between 1 and 65535")
								}
								return nil
							},
						},
					},
				},
			},
			{
				ID:       "security-rate-limits",
				Label:    "Rate Limits",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "rate-limit-enabled",
						Label:    "Rate Limiting",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.rate_limits.enabled",
							Label:     "Rate Limiting",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.RateLimits.Enabled },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.RateLimits.Enabled = v.(bool)
									return nil
								},
							},
							HelpText: "Enable connection rate limiting",
						},
					},
					{
						ID:       "rate-limit-window",
						Label:    "Window Minutes",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.rate_limits.window_minutes",
							Label:     "Window Mins",
							ValueType: IntValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.RateLimits.WindowMinutes },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.RateLimits.WindowMinutes = v.(int)
									return nil
								},
							},
							HelpText: "Rate limit time window in minutes",
						},
					},
				},
			},
			{
				ID:       "security-local-lists",
				Label:    "Local Lists",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "blacklist-enabled",
						Label:    "Blacklist",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.local_lists.blacklist_enabled",
							Label:     "Blacklist",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.BlacklistEnabled },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.LocalLists.BlacklistEnabled = v.(bool)
									return nil
								},
							},
							HelpText: "Enable IP blacklisting",
						},
					},
					{
						ID:       "blacklist-file",
						Label:    "Blacklist File",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.local_lists.blacklist_file",
							Label:     "Blacklist File",
							ValueType: PathValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.BlacklistFile },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.LocalLists.BlacklistFile = v.(string)
									return nil
								},
							},
							HelpText: "Path to blacklist file",
						},
					},
					{
						ID:       "whitelist-enabled",
						Label:    "Whitelist",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.local_lists.whitelist_enabled",
							Label:     "Whitelist",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.WhitelistEnabled },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.LocalLists.WhitelistEnabled = v.(bool)
									return nil
								},
							},
							HelpText: "Enable IP whitelisting",
						},
					},
					{
						ID:       "whitelist-file",
						Label:    "Whitelist File",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.local_lists.whitelist_file",
							Label:     "Whitelist File",
							ValueType: PathValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.LocalLists.WhitelistFile },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.LocalLists.WhitelistFile = v.(string)
									return nil
								},
							},
							HelpText: "Path to whitelist file",
						},
					},
				},
			},
			{
				ID:       "security-external-lists",
				Label:    "External Lists",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "external-block-enabled",
						Label:    "External Block",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.external_lists.enabled",
							Label:     "External Block",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.ExternalLists.Enabled },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.ExternalLists.Enabled = v.(bool)
									return nil
								},
							},
							HelpText: "Enable external blocklist checking",
						},
					},
					{
						ID:       "external-blocklist-urls",
						Label:    "Blocklist URLs",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.external_lists.urls",
							Label:     "Blocklist URLs",
							ValueType: ListValue,
							Field: ConfigField{
								GetValue: func() interface{} {
									return strings.Join(cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs, ", ")
								},
								SetValue: func(v interface{}) error {
									s := v.(string)
									cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs = nil
									for _, u := range strings.Split(s, ",") {
										url := strings.TrimSpace(u)
										if url != "" {
											cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs = append(cfg.Servers.Security.ExternalLists.ExternalBlocklistURLs, url)
										}
									}
									return nil
								},
							},
							HelpText: "Comma-separated list of blocklist URLs",
						},
					},
				},
			},
			{
				ID:       "security-geo-block",
				Label:    "Geo Blocking",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "geo-block-enabled",
						Label:    "Geo Block",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.geo_block.enabled",
							Label:     "Geo Block",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.GeoBlockEnabled },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.GeoBlock.GeoBlockEnabled = v.(bool)
									return nil
								},
							},
							HelpText: "Enable geographic IP blocking",
						},
					},
					{
						ID:       "blocked-countries",
						Label:    "Blocked Countries",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.geo_block.blocked_countries",
							Label:     "Blocked",
							ValueType: ListValue,
							Field: ConfigField{
								GetValue: func() interface{} { return strings.Join(cfg.Servers.Security.GeoBlock.BlockedCountries, ", ") },
								SetValue: func(v interface{}) error {
									s := v.(string)
									cfg.Servers.Security.GeoBlock.BlockedCountries = nil
									for _, country := range strings.Split(s, ",") {
										country = strings.TrimSpace(country)
										if country != "" {
											cfg.Servers.Security.GeoBlock.BlockedCountries = append(cfg.Servers.Security.GeoBlock.BlockedCountries, country)
										}
									}
									return nil
								},
							},
							HelpText: "Comma-separated country codes to block",
						},
					},
					{
						ID:       "allowed-countries",
						Label:    "Allowed Countries",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.geo_block.allowed_countries",
							Label:     "Allowed",
							ValueType: ListValue,
							Field: ConfigField{
								GetValue: func() interface{} { return strings.Join(cfg.Servers.Security.GeoBlock.AllowedCountries, ", ") },
								SetValue: func(v interface{}) error {
									s := v.(string)
									cfg.Servers.Security.GeoBlock.AllowedCountries = nil
									for _, country := range strings.Split(s, ",") {
										country = strings.TrimSpace(country)
										if country != "" {
											cfg.Servers.Security.GeoBlock.AllowedCountries = append(cfg.Servers.Security.GeoBlock.AllowedCountries, country)
										}
									}
									return nil
								},
							},
							HelpText: "Comma-separated country codes to allow (leave empty for all)",
						},
					},
					{
						ID:       "geo-api-provider",
						Label:    "Geo API Provider",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.geo_block.api_provider",
							Label:     "API Provider",
							ValueType: StringValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.GeoAPIProvider },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.GeoBlock.GeoAPIProvider = v.(string)
									return nil
								},
							},
							HelpText: "Geolocation API provider (e.g., ipapi)",
						},
					},
					{
						ID:       "geo-api-key",
						Label:    "Geo API Key",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.geo_block.api_key",
							Label:     "API Key",
							ValueType: StringValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.GeoAPIKey },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.GeoBlock.GeoAPIKey = v.(string)
									return nil
								},
							},
							HelpText: "API key for geolocation service",
						},
					},
					{
						ID:       "threat-intel-enabled",
						Label:    "Threat Intel",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.geo_block.threat_intel",
							Label:     "Threat Intel",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.ThreatIntelEnabled },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.GeoBlock.ThreatIntelEnabled = v.(bool)
									return nil
								},
							},
							HelpText: "Enable threat intelligence checking",
						},
					},
					{
						ID:       "blocklist-update-hours",
						Label:    "Blocklist Hours",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.geo_block.update_hours",
							Label:     "Block Hours",
							ValueType: IntValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.GeoBlock.BlocklistUpdateHours },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.GeoBlock.BlocklistUpdateHours = v.(int)
									return nil
								},
							},
							HelpText: "Hours between blocklist updates",
						},
					},
				},
			},
			{
				ID:       "security-logs",
				Label:    "Security Logs",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "log-security-events",
						Label:    "Log Events",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.logs.log_events",
							Label:     "Log Events",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.Logs.LogSecurityEvents },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.Logs.LogSecurityEvents = v.(bool)
									return nil
								},
							},
							HelpText: "Log security-related events",
						},
					},
					{
						ID:       "log-blocked-attempts",
						Label:    "Log Blocked",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.logs.log_blocked",
							Label:     "Log Blocked",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.Logs.LogBlockedAttempts },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.Logs.LogBlockedAttempts = v.(bool)
									return nil
								},
							},
							HelpText: "Log blocked connection attempts",
						},
					},
					{
						ID:       "security-log-file",
						Label:    "Log File",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "servers.security.logs.log_file",
							Label:     "Log File",
							ValueType: PathValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Servers.Security.Logs.SecurityLogFile },
								SetValue: func(v interface{}) error {
									cfg.Servers.Security.Logs.SecurityLogFile = v.(string)
									return nil
								},
							},
							HelpText: "Path to security log file",
						},
					},
				},
			},
		},
	}
}
