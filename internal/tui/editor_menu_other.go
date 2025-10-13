package tui

import (
	"github.com/robbiew/retrograde/internal/config"
)

func otherMenu(cfg *config.Config) MenuCategory {
	return MenuCategory{
		ID:     "other",
		Label:  "Other",
		HotKey: 'O',
		SubItems: []SubmenuItem{
			{
				ID:       "discord-integration",
				Label:    "Discord Integration",
				ItemType: SectionHeader,
				SubItems: []SubmenuItem{
					{
						ID:       "discord-enabled",
						Label:    "Discord Enabled",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "other.discord.enabled",
							Label:     "Discord Enabled",
							ValueType: BoolValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Other.Discord.Enabled },
								SetValue: func(v interface{}) error {
									cfg.Other.Discord.Enabled = v.(bool)
									return nil
								},
							},
							HelpText: "Enable Discord webhook integration",
						},
					},
					{
						ID:       "discord-webhook-url",
						Label:    "Discord Webhook URL",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "other.discord.webhook_url",
							Label:     "Discord Webhook URL",
							ValueType: StringValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Other.Discord.WebhookURL },
								SetValue: func(v interface{}) error {
									cfg.Other.Discord.WebhookURL = v.(string)
									return nil
								},
							},
							HelpText: "Discord webhook URL for notifications",
						},
					},
					{
						ID:       "discord-username",
						Label:    "Discord Username",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "other.discord.username",
							Label:     "Discord Username",
							ValueType: StringValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Other.Discord.Username },
								SetValue: func(v interface{}) error {
									cfg.Other.Discord.Username = v.(string)
									return nil
								},
							},
							HelpText: "Bot username for Discord notifications",
						},
					},
					{
						ID:       "discord-title",
						Label:    "Discord Title",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "other.discord.title",
							Label:     "Discord Title",
							ValueType: StringValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Other.Discord.Title },
								SetValue: func(v interface{}) error {
									cfg.Other.Discord.Title = v.(string)
									return nil
								},
							},
							HelpText: "Title for Discord notifications",
						},
					},
					{
						ID:       "discord-invite-url",
						Label:    "Discord Invite URL",
						ItemType: EditableField,
						EditableItem: &MenuItem{
							ID:        "other.discord.invite_url",
							Label:     "Discord Invite URL",
							ValueType: StringValue,
							Field: ConfigField{
								GetValue: func() interface{} { return cfg.Other.Discord.InviteURL },
								SetValue: func(v interface{}) error {
									cfg.Other.Discord.InviteURL = v.(string)
									return nil
								},
							},
							HelpText: "Discord server invite URL",
						},
					},
				},
			},
		},
	}
}
