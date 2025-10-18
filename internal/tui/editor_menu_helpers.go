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

// getSecurityLevelSelectOptionsWithPublic returns security level options including Public fallback
func (m *Model) getSecurityLevelSelectOptionsWithPublic() []SelectOption {
	options := []SelectOption{
		{
			Value:       "public",
			Label:       "Public",
			Description: "Accessible to all users",
			Implemented: true,
		},
	}
	return append(options, m.getSecurityLevelSelectOptions()...)
}

func (m *Model) getConferenceSelectOptions() []SelectOption {
	conferences := m.conferenceList
	if len(conferences) == 0 && m.db != nil {
		if list, err := m.db.GetAllConferences(); err == nil {
			conferences = list
			m.conferenceList = list
		}
	}

	options := make([]SelectOption, 0, len(conferences))
	for _, conf := range conferences {
		options = append(options, SelectOption{
			Value:       fmt.Sprintf("%d", conf.ID),
			Label:       conf.Name,
			Description: strings.TrimSpace(conf.Description),
			Implemented: true,
		})
	}

	return options
}

func getAreaTypeOptions() []SelectOption {
	return []SelectOption{
		{Value: "local", Label: "Local", Description: "Local-only message area", Implemented: true},
		{Value: "echomail", Label: "Echomail", Description: "Network echoed message area", Implemented: true},
		{Value: "netmail", Label: "Netmail", Description: "Direct network mail", Implemented: true},
	}
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
