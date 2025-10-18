package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/robbiew/retrograde/internal/database"
	"github.com/robbiew/retrograde/internal/menu"
)

// getCmdKeySelectOptions returns SelectOptions for all implemented command keys
func getCmdKeySelectOptions() []SelectOption {
	registry := menu.NewCmdKeyRegistry()
	definitions := registry.GetAllDefinitions()

	// Filter to only implemented commands
	var implementedDefs []*menu.CmdKeyDefinition
	for _, def := range definitions {
		if def.Implemented {
			implementedDefs = append(implementedDefs, def)
		}
	}

	// Sort by category, then by name
	sort.Slice(implementedDefs, func(i, j int) bool {
		if implementedDefs[i].Category == implementedDefs[j].Category {
			return implementedDefs[i].Name < implementedDefs[j].Name
		}
		return implementedDefs[i].Category < implementedDefs[j].Category
	})

	options := make([]SelectOption, 0, len(implementedDefs))
	for _, def := range implementedDefs {
		options = append(options, SelectOption{
			Value:       def.CmdKey,
			Label:       def.Name,
			Description: def.Description,
			Category:    def.Category,
			Implemented: def.Implemented,
		})
	}

	return options
}

// getSecurityLevelSelectOptions returns SelectOptions for all security levels
func (m *Model) getSecurityLevelSelectOptions() []SelectOption {
	// Get all security levels from database
	levels, err := m.db.GetAllSecurityLevels()
	if err != nil {
		// Return empty options if we can't load levels
		return []SelectOption{}
	}

	// Sort by SecLevel (numeric value)
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].SecLevel < levels[j].SecLevel
	})

	options := make([]SelectOption, 0, len(levels))
	for _, level := range levels {
		options = append(options, SelectOption{
			Value:       fmt.Sprintf("%d", level.SecLevel),
			Label:       level.Name,
			Description: fmt.Sprintf("Level %d", level.SecLevel),
			Category:    "", // No category needed for security levels
			Implemented: true,
		})
	}

	return options
}

func (m *Model) getMenuSelectOptions() []SelectOption {
	menus, err := m.db.GetAllMenus()
	if err != nil {
		return []SelectOption{}
	}

	sort.Slice(menus, func(i, j int) bool {
		return strings.ToLower(menus[i].Name) < strings.ToLower(menus[j].Name)
	})

	options := make([]SelectOption, 0, len(menus))
	for _, menu := range menus {
		options = append(options, SelectOption{
			Value:       menu.Name,
			Label:       menu.Name,
			Description: fmt.Sprintf("Menu ID %d", menu.ID),
			Implemented: true,
		})
	}

	return options
}

func getMenuDisplayModeOptions() []SelectOption {
	return []SelectOption{
		{
			Value:       database.DisplayModeTitlesGenerated,
			Label:       "Titles + Generic Menu",
			Description: "Generic titles + menu",
			Implemented: true,
		},
		{
			Value:       database.DisplayModeHeaderGenerated,
			Label:       "[MenuName].hdr.ans/asc + Generic Menu",
			Description: "Render [Menu].hdr before menu",
			Implemented: true,
		},
		{
			Value:       database.DisplayModeThemeOnly,
			Label:       "[MenuName].ans/asc",
			Description: "Render [Menu].ans only",
			Implemented: true,
		},
	}
}
