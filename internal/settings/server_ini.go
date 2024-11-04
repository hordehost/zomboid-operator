package settings

import (
	"fmt"
	"strconv"
	"strings"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	"k8s.io/utils/ptr"
)

// GenerateServerINI creates a server.ini configuration file from the ZomboidServer settings
func GenerateServerINI(settings zomboidv1.ZomboidServerSpec) string {
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

// ParseServerINI reads a server.ini file content and returns a ZomboidServerSpec
func ParseServerINI(content string) zomboidv1.ZomboidServerSpec {
	settings := zomboidv1.ZomboidServerSpec{}
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
			parseSettingValue(&settings, key, value)
		}
	}

	return settings
}

func parseSettingValue(settings *zomboidv1.ZomboidServerSpec, key, value string) {
	switch key {
	// Identity settings
	case "Public":
		settings.Identity.Public = parseBool(value)
	case "PublicName":
		settings.Identity.PublicName = ptr.To(value)
	case "PublicDescription":
		settings.Identity.PublicDescription = ptr.To(strings.ReplaceAll(value, "\\n", "\n"))
	case "ResetID":
		settings.Identity.ResetID = parseInt32(value)
	case "ServerPlayerID":
		settings.Identity.ServerPlayerID = parseInt32(value)

	// Player settings
	case "MaxPlayers":
		settings.Player.MaxPlayers = parseInt32(value)
	case "PingLimit":
		settings.Player.PingLimit = parseInt32(value)
	case "Open":
		settings.Player.Open = parseBool(value)
	case "AutoCreateUserInWhiteList":
		settings.Player.AutoCreateUserInWhiteList = parseBool(value)
	case "DropOffWhiteListAfterDeath":
		settings.Player.DropOffWhiteListAfterDeath = parseBool(value)
	case "MaxAccountsPerUser":
		settings.Player.MaxAccountsPerUser = parseInt32(value)
	case "AllowCoop":
		settings.Player.AllowCoop = parseBool(value)
	case "AllowNonAsciiUsername":
		settings.Player.AllowNonAsciiUsername = parseBool(value)
	case "DenyLoginOnOverloadedServer":
		settings.Player.DenyLoginOnOverloadedServer = parseBool(value)
	case "LoginQueueEnabled":
		settings.Player.LoginQueueEnabled = parseBool(value)
	case "LoginQueueConnectTimeout":
		settings.Player.LoginQueueConnectTimeout = parseInt32(value)

	// Map settings
	case "Map":
		settings.Map.Map = ptr.To(value)

	// Mods settings
	case "WorkshopItems":
		settings.Mods.WorkshopItems = ptr.To(value)
	case "Mods":
		settings.Mods.Mods = ptr.To(value)

	// Backup settings
	case "SaveWorldEveryMinutes":
		settings.Backup.SaveWorldEveryMinutes = parseInt32(value)
	case "BackupsCount":
		settings.Backup.BackupsCount = parseInt32(value)
	case "BackupsOnStart":
		settings.Backup.BackupsOnStart = parseBool(value)
	case "BackupsOnVersionChange":
		settings.Backup.BackupsOnVersionChange = parseBool(value)
	case "BackupsPeriod":
		settings.Backup.BackupsPeriod = parseInt32(value)

	// Logging settings
	case "PerkLogs":
		settings.Logging.PerkLogs = parseBool(value)
	case "ClientCommandFilter":
		settings.Logging.ClientCommandFilter = ptr.To(value)
	case "ClientActionLogs":
		settings.Logging.ClientActionLogs = ptr.To(value)

	// Moderation settings
	case "DisableRadioStaff":
		settings.Moderation.DisableRadioStaff = parseBool(value)
	case "DisableRadioAdmin":
		settings.Moderation.DisableRadioAdmin = parseBool(value)
	case "DisableRadioGM":
		settings.Moderation.DisableRadioGM = parseBool(value)
	case "DisableRadioOverseer":
		settings.Moderation.DisableRadioOverseer = parseBool(value)
	case "DisableRadioModerator":
		settings.Moderation.DisableRadioModerator = parseBool(value)
	case "DisableRadioInvisible":
		settings.Moderation.DisableRadioInvisible = parseBool(value)
	case "BanKickGlobalSound":
		settings.Moderation.BanKickGlobalSound = parseBool(value)

	// Steam settings
	case "SteamScoreboard":
		settings.Steam.SteamScoreboard = ptr.To(value)

	// Discord settings
	case "DiscordEnable":
		settings.Discord.DiscordEnable = parseBool(value)
	case "DiscordToken":
		settings.Discord.DiscordToken = ptr.To(value)
	case "DiscordChannel":
		settings.Discord.DiscordChannel = ptr.To(value)
	case "DiscordChannelID":
		settings.Discord.DiscordChannelID = ptr.To(value)

	// Communication settings
	case "GlobalChat":
		settings.Communication.GlobalChat = parseBool(value)
	case "ChatStreams":
		settings.Communication.ChatStreams = ptr.To(value)
	case "ServerWelcomeMessage":
		settings.Communication.ServerWelcomeMessage = ptr.To(strings.ReplaceAll(value, "\\n", "\n"))
	case "VoiceEnable":
		settings.Communication.VoiceEnable = parseBool(value)
	case "VoiceMinDistance":
		settings.Communication.VoiceMinDistance = parseFloat32(value)
	case "VoiceMaxDistance":
		settings.Communication.VoiceMaxDistance = parseFloat32(value)
	case "Voice3D":
		settings.Communication.Voice3D = parseBool(value)

	// Gameplay settings
	case "PauseEmpty":
		settings.Gameplay.PauseEmpty = parseBool(value)
	case "DisplayUserName":
		settings.Gameplay.DisplayUserName = parseBool(value)
	case "ShowFirstAndLastName":
		settings.Gameplay.ShowFirstAndLastName = parseBool(value)
	case "SpawnPoint":
		settings.Gameplay.SpawnPoint = ptr.To(value)
	case "SpawnItems":
		settings.Gameplay.SpawnItems = ptr.To(value)
	case "NoFire":
		settings.Gameplay.NoFire = parseBool(value)
	case "AnnounceDeath":
		settings.Gameplay.AnnounceDeath = parseBool(value)
	case "MinutesPerPage":
		settings.Gameplay.MinutesPerPage = parseFloat32(value)
	case "AllowDestructionBySledgehammer":
		settings.Gameplay.AllowDestructionBySledgehammer = parseBool(value)
	case "SledgehammerOnlyInSafehouse":
		settings.Gameplay.SledgehammerOnlyInSafehouse = parseBool(value)
	case "SleepAllowed":
		settings.Gameplay.SleepAllowed = parseBool(value)
	case "SleepNeeded":
		settings.Gameplay.SleepNeeded = parseBool(value)
	case "KnockedDownAllowed":
		settings.Gameplay.KnockedDownAllowed = parseBool(value)
	case "SneakModeHideFromOtherPlayers":
		settings.Gameplay.SneakModeHideFromOtherPlayers = parseBool(value)
	case "SpeedLimit":
		settings.Gameplay.SpeedLimit = parseFloat32(value)
	case "PlayerRespawnWithSelf":
		settings.Gameplay.PlayerRespawnWithSelf = parseBool(value)
	case "PlayerRespawnWithOther":
		settings.Gameplay.PlayerRespawnWithOther = parseBool(value)
	case "FastForwardMultiplier":
		settings.Gameplay.FastForwardMultiplier = parseFloat32(value)

	// PVP settings
	case "PVP":
		settings.PVP.PVP = parseBool(value)
	case "SafetySystem":
		settings.PVP.SafetySystem = parseBool(value)
	case "ShowSafety":
		settings.PVP.ShowSafety = parseBool(value)
	case "SafetyToggleTimer":
		settings.PVP.SafetyToggleTimer = parseInt32(value)
	case "SafetyCooldownTimer":
		settings.PVP.SafetyCooldownTimer = parseInt32(value)
	case "PVPMeleeDamageModifier":
		settings.PVP.PVPMeleeDamageModifier = parseFloat32(value)
	case "PVPFirearmDamageModifier":
		settings.PVP.PVPFirearmDamageModifier = parseFloat32(value)
	case "PVPMeleeWhileHitReaction":
		settings.PVP.PVPMeleeWhileHitReaction = parseBool(value)

	// Safehouse settings
	case "PlayerSafehouse":
		settings.Safehouse.PlayerSafehouse = parseBool(value)
	case "AdminSafehouse":
		settings.Safehouse.AdminSafehouse = parseBool(value)
	case "SafehouseAllowTrepass":
		settings.Safehouse.SafehouseAllowTrepass = parseBool(value)
	case "SafehouseAllowFire":
		settings.Safehouse.SafehouseAllowFire = parseBool(value)
	case "SafehouseAllowLoot":
		settings.Safehouse.SafehouseAllowLoot = parseBool(value)
	case "SafehouseAllowRespawn":
		settings.Safehouse.SafehouseAllowRespawn = parseBool(value)
	case "SafehouseDaySurvivedToClaim":
		settings.Safehouse.SafehouseDaySurvivedToClaim = parseInt32(value)
	case "SafeHouseRemovalTime":
		settings.Safehouse.SafeHouseRemovalTime = parseInt32(value)
	case "SafehouseAllowNonResidential":
		settings.Safehouse.SafehouseAllowNonResidential = parseBool(value)
	case "DisableSafehouseWhenPlayerConnected":
		settings.Safehouse.DisableSafehouseWhenPlayerConnected = parseBool(value)

	// Faction settings
	case "Faction":
		settings.Faction.Faction = parseBool(value)
	case "FactionDaySurvivedToCreate":
		settings.Faction.FactionDaySurvivedToCreate = parseInt32(value)
	case "FactionPlayersRequiredForTag":
		settings.Faction.FactionPlayersRequiredForTag = parseInt32(value)

	// Loot settings
	case "HoursForLootRespawn":
		settings.Loot.HoursForLootRespawn = parseInt32(value)
	case "MaxItemsForLootRespawn":
		settings.Loot.MaxItemsForLootRespawn = parseInt32(value)
	case "ConstructionPreventsLootRespawn":
		settings.Loot.ConstructionPreventsLootRespawn = parseBool(value)
	case "ItemNumbersLimitPerContainer":
		settings.Loot.ItemNumbersLimitPerContainer = parseInt32(value)
	case "TrashDeleteAll":
		settings.Loot.TrashDeleteAll = parseBool(value)

	// AntiCheat settings
	case "DoLuaChecksum":
		settings.AntiCheat.DoLuaChecksum = parseBool(value)
	case "KickFastPlayers":
		settings.AntiCheat.KickFastPlayers = parseBool(value)
	case "AntiCheatProtectionType1":
		settings.AntiCheat.AntiCheatProtectionType1 = parseBool(value)
	case "AntiCheatProtectionType2":
		settings.AntiCheat.AntiCheatProtectionType2 = parseBool(value)
	case "AntiCheatProtectionType3":
		settings.AntiCheat.AntiCheatProtectionType3 = parseBool(value)
	case "AntiCheatProtectionType4":
		settings.AntiCheat.AntiCheatProtectionType4 = parseBool(value)
	case "AntiCheatProtectionType5":
		settings.AntiCheat.AntiCheatProtectionType5 = parseBool(value)
	case "AntiCheatProtectionType6":
		settings.AntiCheat.AntiCheatProtectionType6 = parseBool(value)
	case "AntiCheatProtectionType7":
		settings.AntiCheat.AntiCheatProtectionType7 = parseBool(value)
	case "AntiCheatProtectionType8":
		settings.AntiCheat.AntiCheatProtectionType8 = parseBool(value)
	case "AntiCheatProtectionType9":
		settings.AntiCheat.AntiCheatProtectionType9 = parseBool(value)
	case "AntiCheatProtectionType10":
		settings.AntiCheat.AntiCheatProtectionType10 = parseBool(value)
	case "AntiCheatProtectionType11":
		settings.AntiCheat.AntiCheatProtectionType11 = parseBool(value)
	case "AntiCheatProtectionType12":
		settings.AntiCheat.AntiCheatProtectionType12 = parseBool(value)
	case "AntiCheatProtectionType13":
		settings.AntiCheat.AntiCheatProtectionType13 = parseBool(value)
	case "AntiCheatProtectionType14":
		settings.AntiCheat.AntiCheatProtectionType14 = parseBool(value)
	case "AntiCheatProtectionType15":
		settings.AntiCheat.AntiCheatProtectionType15 = parseBool(value)
	case "AntiCheatProtectionType16":
		settings.AntiCheat.AntiCheatProtectionType16 = parseBool(value)
	case "AntiCheatProtectionType17":
		settings.AntiCheat.AntiCheatProtectionType17 = parseBool(value)
	case "AntiCheatProtectionType18":
		settings.AntiCheat.AntiCheatProtectionType18 = parseBool(value)
	case "AntiCheatProtectionType19":
		settings.AntiCheat.AntiCheatProtectionType19 = parseBool(value)
	case "AntiCheatProtectionType20":
		settings.AntiCheat.AntiCheatProtectionType20 = parseBool(value)
	case "AntiCheatProtectionType21":
		settings.AntiCheat.AntiCheatProtectionType21 = parseBool(value)
	case "AntiCheatProtectionType22":
		settings.AntiCheat.AntiCheatProtectionType22 = parseBool(value)
	case "AntiCheatProtectionType23":
		settings.AntiCheat.AntiCheatProtectionType23 = parseBool(value)
	case "AntiCheatProtectionType24":
		settings.AntiCheat.AntiCheatProtectionType24 = parseBool(value)
	case "AntiCheatProtectionType2ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier = parseFloat32(value)
	case "AntiCheatProtectionType3ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType3ThresholdMultiplier = parseFloat32(value)
	case "AntiCheatProtectionType4ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType4ThresholdMultiplier = parseFloat32(value)
	case "AntiCheatProtectionType9ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType9ThresholdMultiplier = parseFloat32(value)
	case "AntiCheatProtectionType15ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType15ThresholdMultiplier = parseFloat32(value)
	case "AntiCheatProtectionType20ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType20ThresholdMultiplier = parseFloat32(value)
	case "AntiCheatProtectionType22ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType22ThresholdMultiplier = parseFloat32(value)
	case "AntiCheatProtectionType24ThresholdMultiplier":
		settings.AntiCheat.AntiCheatProtectionType24ThresholdMultiplier = parseFloat32(value)
	}
}

func parseBool(value string) *bool {
	return ptr.To(strings.ToLower(value) == "true")
}

func parseInt32(value string) *int32 {
	i, _ := strconv.ParseInt(value, 10, 32)
	return ptr.To(int32(i))
}

func parseFloat32(value string) *float32 {
	f, _ := strconv.ParseFloat(value, 32)
	return ptr.To(float32(f))
}
