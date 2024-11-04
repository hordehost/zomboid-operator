package settings

import (
	"fmt"
	"strings"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
)

// ParseServerINI reads a server.ini file content and returns a ZomboidSettings
func ParseServerINI(content string) zomboidv1.ZomboidSettings {
	settings := zomboidv1.ZomboidSettings{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if value != "" {
			ParseSettingValue(&settings, key, value)
		}
	}

	return settings
}

// GenerateServerINI creates a server.ini configuration file from the ZomboidServer settings
func GenerateServerINI(settings zomboidv1.ZomboidSettings) string {
	var sb strings.Builder

	// Identity settings
	writeString(&sb, "Public", settings.Identity.Public)
	writeString(&sb, "PublicName", settings.Identity.PublicName)
	writeString(&sb, "PublicDescription", settings.Identity.PublicDescription)
	writeString(&sb, "ResetID", settings.Identity.ResetID)
	writeString(&sb, "ServerPlayerID", settings.Identity.ServerPlayerID)

	// Player settings
	writeString(&sb, "MaxPlayers", settings.Player.MaxPlayers)
	writeString(&sb, "PingLimit", settings.Player.PingLimit)
	writeString(&sb, "Open", settings.Player.Open)
	writeString(&sb, "AutoCreateUserInWhiteList", settings.Player.AutoCreateUserInWhiteList)
	writeString(&sb, "DropOffWhiteListAfterDeath", settings.Player.DropOffWhiteListAfterDeath)
	writeString(&sb, "MaxAccountsPerUser", settings.Player.MaxAccountsPerUser)
	writeString(&sb, "AllowCoop", settings.Player.AllowCoop)
	writeString(&sb, "AllowNonAsciiUsername", settings.Player.AllowNonAsciiUsername)
	writeString(&sb, "DenyLoginOnOverloadedServer", settings.Player.DenyLoginOnOverloadedServer)
	writeString(&sb, "LoginQueueEnabled", settings.Player.LoginQueueEnabled)
	writeString(&sb, "LoginQueueConnectTimeout", settings.Player.LoginQueueConnectTimeout)

	// Map settings
	writeString(&sb, "Map", settings.Map.Map)

	// Mods settings - combine WorkshopMods and classic Mods format
	var workshopItems, mods []string

	// Add workshop mods
	for _, mod := range settings.WorkshopMods {
		if mod.WorkshopID != nil {
			workshopItems = append(workshopItems, *mod.WorkshopID)
		}
		if mod.ModID != nil {
			mods = append(mods, *mod.ModID)
		}
	}

	// Add classic mods
	if *settings.Mods.WorkshopItems != "" {
		workshopItems = append(workshopItems, strings.Split(*settings.Mods.WorkshopItems, ";")...)
	}
	if *settings.Mods.Mods != "" {
		mods = append(mods, strings.Split(*settings.Mods.Mods, ";")...)
	}

	writeString(&sb, "WorkshopItems", strings.Join(workshopItems, ";"))
	writeString(&sb, "Mods", strings.Join(mods, ";"))

	// Backup settings
	writeString(&sb, "SaveWorldEveryMinutes", settings.Backup.SaveWorldEveryMinutes)
	writeString(&sb, "BackupsCount", settings.Backup.BackupsCount)
	writeString(&sb, "BackupsOnStart", settings.Backup.BackupsOnStart)
	writeString(&sb, "BackupsOnVersionChange", settings.Backup.BackupsOnVersionChange)
	writeString(&sb, "BackupsPeriod", settings.Backup.BackupsPeriod)

	// Logging settings
	writeString(&sb, "PerkLogs", settings.Logging.PerkLogs)
	writeString(&sb, "ClientCommandFilter", settings.Logging.ClientCommandFilter)
	writeString(&sb, "ClientActionLogs", settings.Logging.ClientActionLogs)

	// Moderation settings
	writeString(&sb, "DisableRadioStaff", settings.Moderation.DisableRadioStaff)
	writeString(&sb, "DisableRadioAdmin", settings.Moderation.DisableRadioAdmin)
	writeString(&sb, "DisableRadioGM", settings.Moderation.DisableRadioGM)
	writeString(&sb, "DisableRadioOverseer", settings.Moderation.DisableRadioOverseer)
	writeString(&sb, "DisableRadioModerator", settings.Moderation.DisableRadioModerator)
	writeString(&sb, "DisableRadioInvisible", settings.Moderation.DisableRadioInvisible)
	writeString(&sb, "BanKickGlobalSound", settings.Moderation.BanKickGlobalSound)

	// Steam settings
	writeString(&sb, "SteamScoreboard", settings.Steam.SteamScoreboard)

	// Discord settings
	writeString(&sb, "DiscordEnable", settings.Discord.DiscordEnable)
	writeString(&sb, "DiscordToken", settings.Discord.DiscordToken)
	writeString(&sb, "DiscordChannel", settings.Discord.DiscordChannel)
	writeString(&sb, "DiscordChannelID", settings.Discord.DiscordChannelID)

	// Communication settings
	writeString(&sb, "GlobalChat", settings.Communication.GlobalChat)
	writeString(&sb, "ChatStreams", settings.Communication.ChatStreams)
	writeString(&sb, "ServerWelcomeMessage", settings.Communication.ServerWelcomeMessage)
	writeString(&sb, "VoiceEnable", settings.Communication.VoiceEnable)
	writeString(&sb, "VoiceMinDistance", settings.Communication.VoiceMinDistance)
	writeString(&sb, "VoiceMaxDistance", settings.Communication.VoiceMaxDistance)
	writeString(&sb, "Voice3D", settings.Communication.Voice3D)

	// Gameplay settings
	writeString(&sb, "PauseEmpty", settings.Gameplay.PauseEmpty)
	writeString(&sb, "DisplayUserName", settings.Gameplay.DisplayUserName)
	writeString(&sb, "ShowFirstAndLastName", settings.Gameplay.ShowFirstAndLastName)
	writeString(&sb, "SpawnPoint", settings.Gameplay.SpawnPoint)
	writeString(&sb, "SpawnItems", settings.Gameplay.SpawnItems)
	writeString(&sb, "NoFire", settings.Gameplay.NoFire)
	writeString(&sb, "AnnounceDeath", settings.Gameplay.AnnounceDeath)
	writeString(&sb, "MinutesPerPage", settings.Gameplay.MinutesPerPage)
	writeString(&sb, "AllowDestructionBySledgehammer", settings.Gameplay.AllowDestructionBySledgehammer)
	writeString(&sb, "SledgehammerOnlyInSafehouse", settings.Gameplay.SledgehammerOnlyInSafehouse)
	writeString(&sb, "SleepAllowed", settings.Gameplay.SleepAllowed)
	writeString(&sb, "SleepNeeded", settings.Gameplay.SleepNeeded)
	writeString(&sb, "KnockedDownAllowed", settings.Gameplay.KnockedDownAllowed)
	writeString(&sb, "SneakModeHideFromOtherPlayers", settings.Gameplay.SneakModeHideFromOtherPlayers)
	writeString(&sb, "SpeedLimit", settings.Gameplay.SpeedLimit)
	writeString(&sb, "PlayerRespawnWithSelf", settings.Gameplay.PlayerRespawnWithSelf)
	writeString(&sb, "PlayerRespawnWithOther", settings.Gameplay.PlayerRespawnWithOther)
	writeString(&sb, "FastForwardMultiplier", settings.Gameplay.FastForwardMultiplier)

	// PVP settings
	writeString(&sb, "PVP", settings.PVP.PVP)
	writeString(&sb, "SafetySystem", settings.PVP.SafetySystem)
	writeString(&sb, "ShowSafety", settings.PVP.ShowSafety)
	writeString(&sb, "SafetyToggleTimer", settings.PVP.SafetyToggleTimer)
	writeString(&sb, "SafetyCooldownTimer", settings.PVP.SafetyCooldownTimer)
	writeString(&sb, "PVPMeleeDamageModifier", settings.PVP.PVPMeleeDamageModifier)
	writeString(&sb, "PVPFirearmDamageModifier", settings.PVP.PVPFirearmDamageModifier)
	writeString(&sb, "PVPMeleeWhileHitReaction", settings.PVP.PVPMeleeWhileHitReaction)

	// Safehouse settings
	writeString(&sb, "PlayerSafehouse", settings.Safehouse.PlayerSafehouse)
	writeString(&sb, "AdminSafehouse", settings.Safehouse.AdminSafehouse)
	writeString(&sb, "SafehouseAllowTrepass", settings.Safehouse.SafehouseAllowTrepass)
	writeString(&sb, "SafehouseAllowFire", settings.Safehouse.SafehouseAllowFire)
	writeString(&sb, "SafehouseAllowLoot", settings.Safehouse.SafehouseAllowLoot)
	writeString(&sb, "SafehouseAllowRespawn", settings.Safehouse.SafehouseAllowRespawn)
	writeString(&sb, "SafehouseDaySurvivedToClaim", settings.Safehouse.SafehouseDaySurvivedToClaim)
	writeString(&sb, "SafeHouseRemovalTime", settings.Safehouse.SafeHouseRemovalTime)
	writeString(&sb, "SafehouseAllowNonResidential", settings.Safehouse.SafehouseAllowNonResidential)
	writeString(&sb, "DisableSafehouseWhenPlayerConnected", settings.Safehouse.DisableSafehouseWhenPlayerConnected)

	// Faction settings
	writeString(&sb, "Faction", settings.Faction.Faction)
	writeString(&sb, "FactionDaySurvivedToCreate", settings.Faction.FactionDaySurvivedToCreate)
	writeString(&sb, "FactionPlayersRequiredForTag", settings.Faction.FactionPlayersRequiredForTag)

	// Loot settings
	writeString(&sb, "HoursForLootRespawn", settings.Loot.HoursForLootRespawn)
	writeString(&sb, "MaxItemsForLootRespawn", settings.Loot.MaxItemsForLootRespawn)
	writeString(&sb, "ConstructionPreventsLootRespawn", settings.Loot.ConstructionPreventsLootRespawn)
	writeString(&sb, "ItemNumbersLimitPerContainer", settings.Loot.ItemNumbersLimitPerContainer)
	writeString(&sb, "TrashDeleteAll", settings.Loot.TrashDeleteAll)

	// AntiCheat settings
	writeString(&sb, "DoLuaChecksum", settings.AntiCheat.DoLuaChecksum)
	writeString(&sb, "KickFastPlayers", settings.AntiCheat.KickFastPlayers)

	// AntiCheat protection types
	writeString(&sb, "AntiCheatProtectionType1", settings.AntiCheat.AntiCheatProtectionType1)
	writeString(&sb, "AntiCheatProtectionType2", settings.AntiCheat.AntiCheatProtectionType2)
	writeString(&sb, "AntiCheatProtectionType3", settings.AntiCheat.AntiCheatProtectionType3)
	writeString(&sb, "AntiCheatProtectionType4", settings.AntiCheat.AntiCheatProtectionType4)
	writeString(&sb, "AntiCheatProtectionType5", settings.AntiCheat.AntiCheatProtectionType5)
	writeString(&sb, "AntiCheatProtectionType6", settings.AntiCheat.AntiCheatProtectionType6)
	writeString(&sb, "AntiCheatProtectionType7", settings.AntiCheat.AntiCheatProtectionType7)
	writeString(&sb, "AntiCheatProtectionType8", settings.AntiCheat.AntiCheatProtectionType8)
	writeString(&sb, "AntiCheatProtectionType9", settings.AntiCheat.AntiCheatProtectionType9)
	writeString(&sb, "AntiCheatProtectionType10", settings.AntiCheat.AntiCheatProtectionType10)
	writeString(&sb, "AntiCheatProtectionType11", settings.AntiCheat.AntiCheatProtectionType11)
	writeString(&sb, "AntiCheatProtectionType12", settings.AntiCheat.AntiCheatProtectionType12)
	writeString(&sb, "AntiCheatProtectionType13", settings.AntiCheat.AntiCheatProtectionType13)
	writeString(&sb, "AntiCheatProtectionType14", settings.AntiCheat.AntiCheatProtectionType14)
	writeString(&sb, "AntiCheatProtectionType15", settings.AntiCheat.AntiCheatProtectionType15)
	writeString(&sb, "AntiCheatProtectionType16", settings.AntiCheat.AntiCheatProtectionType16)
	writeString(&sb, "AntiCheatProtectionType17", settings.AntiCheat.AntiCheatProtectionType17)
	writeString(&sb, "AntiCheatProtectionType18", settings.AntiCheat.AntiCheatProtectionType18)
	writeString(&sb, "AntiCheatProtectionType19", settings.AntiCheat.AntiCheatProtectionType19)
	writeString(&sb, "AntiCheatProtectionType20", settings.AntiCheat.AntiCheatProtectionType20)
	writeString(&sb, "AntiCheatProtectionType21", settings.AntiCheat.AntiCheatProtectionType21)
	writeString(&sb, "AntiCheatProtectionType22", settings.AntiCheat.AntiCheatProtectionType22)
	writeString(&sb, "AntiCheatProtectionType23", settings.AntiCheat.AntiCheatProtectionType23)
	writeString(&sb, "AntiCheatProtectionType24", settings.AntiCheat.AntiCheatProtectionType24)

	// AntiCheat threshold multipliers
	writeString(&sb, "AntiCheatProtectionType2ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier)
	writeString(&sb, "AntiCheatProtectionType3ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType3ThresholdMultiplier)
	writeString(&sb, "AntiCheatProtectionType4ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType4ThresholdMultiplier)
	writeString(&sb, "AntiCheatProtectionType9ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType9ThresholdMultiplier)
	writeString(&sb, "AntiCheatProtectionType15ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType15ThresholdMultiplier)
	writeString(&sb, "AntiCheatProtectionType20ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType20ThresholdMultiplier)
	writeString(&sb, "AntiCheatProtectionType22ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType22ThresholdMultiplier)
	writeString(&sb, "AntiCheatProtectionType24ThresholdMultiplier", settings.AntiCheat.AntiCheatProtectionType24ThresholdMultiplier)

	return sb.String()
}

// writeString writes a key-value pair to the string builder
func writeString(sb *strings.Builder, key string, value interface{}) {
	if value == nil {
		return
	}

	switch v := value.(type) {
	case *bool:
		if v != nil {
			sb.WriteString(fmt.Sprintf("%s=%v\n", key, *v))
		}
	case *float32:
		if v != nil {
			sb.WriteString(fmt.Sprintf("%s=%.1f\n", key, *v))
		}
	case *string:
		if v != nil && *v != "" {
			sb.WriteString(fmt.Sprintf("%s=%s\n", key, strings.ReplaceAll(*v, "\n", "\\n")))
		}
	case *int32:
		if v != nil {
			sb.WriteString(fmt.Sprintf("%s=%d\n", key, *v))
		}
	case string:
		if v != "" {
			sb.WriteString(fmt.Sprintf("%s=%s\n", key, v))
		}
	default:
		if v != nil {
			sb.WriteString(fmt.Sprintf("%s=%v\n", key, v))
		}
	}
}
