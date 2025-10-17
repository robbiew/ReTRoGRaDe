package tui

import (
	"sort"

	"github.com/robbiew/retrograde/internal/menu"
)

// getCmdKeySelectOptions returns SelectOptions for all registered command keys
func getCmdKeySelectOptions() []SelectOption {
	registry := menu.NewCmdKeyRegistry()
	definitions := registry.GetAllDefinitions()

	// Sort by category, then by name
	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].Category == definitions[j].Category {
			return definitions[i].Name < definitions[j].Name
		}
		return definitions[i].Category < definitions[j].Category
	})

	options := make([]SelectOption, 0, len(definitions))
	for _, def := range definitions {
		options = append(options, SelectOption{
			Value:       def.CmdKey,
			Label:       def.Name,
			Description: def.Description,
			Category:    def.Category,
		})
	}

	return options
}
