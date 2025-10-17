package tui

import (
	"sort"

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
