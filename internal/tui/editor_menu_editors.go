package tui

func editorsMenu() MenuCategory {
	return MenuCategory{
		ID:     "editors",
		Label:  "Editors",
		HotKey: 'E',
		SubItems: []SubmenuItem{
			{
				ID:       "user-editor",
				Label:    "Users",
				ItemType: ActionItem,
			},
			{
				ID:       "security-levels-editor",
				Label:    "Security Levels",
				ItemType: ActionItem,
			},
			{
				ID:       "menu-editor",
				Label:    "Menus",
				ItemType: ActionItem,
			},
		},
	}
}
