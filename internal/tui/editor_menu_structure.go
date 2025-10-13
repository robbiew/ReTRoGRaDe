package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/robbiew/retrograde/internal/config"
)

func buildMenuStructure(cfg *config.Config) MenuBar {
	return MenuBar{
		Items: []MenuCategory{
			configurationMenu(cfg),
			serversMenu(cfg),
			networkingMenu(),
			editorsMenu(),
			otherMenu(cfg),
		},
	}
}

func networkingMenu() MenuCategory {
	return MenuCategory{
		ID:       "networking",
		Label:    "Networking",
		HotKey:   'N',
		SubItems: []SubmenuItem{
			// TODO: Add network configuration when needed
		},
	}
}

// ============================================================================
// Initialization
// ============================================================================

// buildListItems creates list items from a menu category's submenu items
// Sub-menu should contain section headers and action items, not individual fields
func buildListItems(category MenuCategory) []list.Item {
	var items []list.Item

	for _, section := range category.SubItems {
		// Add section headers and action items to the submenu
		// Individual fields will be shown in a modal after selecting a section
		if section.ItemType == SectionHeader || section.ItemType == ActionItem {
			items = append(items, submenuListItem{submenuItem: section})
		}
	}

	return items
}
