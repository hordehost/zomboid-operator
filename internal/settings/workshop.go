package settings

import (
	"strings"

	zomboidv1 "github.com/zomboidhost/zomboid-operator/api/v1"
)

func MergeWorkshopMods(settings *zomboidv1.ZomboidSettings) {
	if len(settings.WorkshopMods) == 0 {
		return
	}

	var modIDs []string
	var workshopIDs []string

	// First collect any existing mods from the semicolon-separated lists
	if settings.Mods.Mods != nil && *settings.Mods.Mods != "" {
		modIDs = append(modIDs, strings.Split(*settings.Mods.Mods, ";")...)
	}
	if settings.Mods.WorkshopItems != nil && *settings.Mods.WorkshopItems != "" {
		workshopIDs = append(workshopIDs, strings.Split(*settings.Mods.WorkshopItems, ";")...)
	}

	// Add the structured workshop mods
	for _, mod := range settings.WorkshopMods {
		if mod.ModID != nil {
			modIDs = append(modIDs, *mod.ModID)
		}
		if mod.WorkshopID != nil {
			workshopIDs = append(workshopIDs, *mod.WorkshopID)
		}
	}

	// Convert back to semicolon-separated strings if we have any items
	if len(modIDs) > 0 {
		modString := strings.Join(modIDs, ";")
		settings.Mods.Mods = &modString
	}
	if len(workshopIDs) > 0 {
		workshopString := strings.Join(workshopIDs, ";")
		settings.Mods.WorkshopItems = &workshopString
	}
}
