package settings

import (
	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
)

// SettingsDiff compares current and desired settings, returning a list of settings that need to be updated
// Each returned pair contains the setting name and its new value as strings
func SettingsDiff(current, desired zomboidv1.ZomboidSettings) [][2]string {
	var updates [][2]string

	// Helper to add a setting update if values differ
	addIfDifferent := func(name string, current, desired interface{}) {
		if current == nil && desired == nil {
			return
		}

		// Convert both values to strings for comparison
		currentStr := ValueToString(current)
		desiredStr := ValueToString(desired)
		if currentStr != desiredStr {
			updates = append(updates, [2]string{name, desiredStr})
		}
	}

	// Identity settings
	addIfDifferent("Public", current.Identity.Public, desired.Identity.Public)
	addIfDifferent("PublicName", current.Identity.PublicName, desired.Identity.PublicName)
	addIfDifferent("PublicDescription", current.Identity.PublicDescription, desired.Identity.PublicDescription)
	addIfDifferent("ResetID", current.Identity.ResetID, desired.Identity.ResetID)
	addIfDifferent("ServerPlayerID", current.Identity.ServerPlayerID, desired.Identity.ServerPlayerID)

	// Player settings
	addIfDifferent("MaxPlayers", current.Player.MaxPlayers, desired.Player.MaxPlayers)
	addIfDifferent("Open", current.Player.Open, desired.Player.Open)
	addIfDifferent("PingLimit", current.Player.PingLimit, desired.Player.PingLimit)
	addIfDifferent("AutoCreateUserInWhiteList", current.Player.AutoCreateUserInWhiteList, desired.Player.AutoCreateUserInWhiteList)
	addIfDifferent("DropOffWhiteListAfterDeath", current.Player.DropOffWhiteListAfterDeath, desired.Player.DropOffWhiteListAfterDeath)
	addIfDifferent("MaxAccountsPerUser", current.Player.MaxAccountsPerUser, desired.Player.MaxAccountsPerUser)
	addIfDifferent("AllowCoop", current.Player.AllowCoop, desired.Player.AllowCoop)
	addIfDifferent("AllowNonAsciiUsername", current.Player.AllowNonAsciiUsername, desired.Player.AllowNonAsciiUsername)
	addIfDifferent("DenyLoginOnOverloadedServer", current.Player.DenyLoginOnOverloadedServer, desired.Player.DenyLoginOnOverloadedServer)
	addIfDifferent("LoginQueueEnabled", current.Player.LoginQueueEnabled, desired.Player.LoginQueueEnabled)
	addIfDifferent("LoginQueueConnectTimeout", current.Player.LoginQueueConnectTimeout, desired.Player.LoginQueueConnectTimeout)

	// Map settings
	addIfDifferent("Map", current.Map.Map, desired.Map.Map)

	// Mods settings
	addIfDifferent("Mods", current.Mods.Mods, desired.Mods.Mods)
	addIfDifferent("WorkshopItems", current.Mods.WorkshopItems, desired.Mods.WorkshopItems)

	// Backup settings
	addIfDifferent("SaveWorldEveryMinutes", current.Backup.SaveWorldEveryMinutes, desired.Backup.SaveWorldEveryMinutes)
	addIfDifferent("BackupsCount", current.Backup.BackupsCount, desired.Backup.BackupsCount)
	addIfDifferent("BackupsOnStart", current.Backup.BackupsOnStart, desired.Backup.BackupsOnStart)
	addIfDifferent("BackupsOnVersionChange", current.Backup.BackupsOnVersionChange, desired.Backup.BackupsOnVersionChange)
	addIfDifferent("BackupsPeriod", current.Backup.BackupsPeriod, desired.Backup.BackupsPeriod)

	// Logging settings
	addIfDifferent("PerkLogs", current.Logging.PerkLogs, desired.Logging.PerkLogs)
	addIfDifferent("ClientCommandFilter", current.Logging.ClientCommandFilter, desired.Logging.ClientCommandFilter)
	addIfDifferent("ClientActionLogs", current.Logging.ClientActionLogs, desired.Logging.ClientActionLogs)

	// Moderation settings
	addIfDifferent("DisableRadioStaff", current.Moderation.DisableRadioStaff, desired.Moderation.DisableRadioStaff)
	addIfDifferent("DisableRadioAdmin", current.Moderation.DisableRadioAdmin, desired.Moderation.DisableRadioAdmin)
	addIfDifferent("DisableRadioGM", current.Moderation.DisableRadioGM, desired.Moderation.DisableRadioGM)
	addIfDifferent("DisableRadioOverseer", current.Moderation.DisableRadioOverseer, desired.Moderation.DisableRadioOverseer)
	addIfDifferent("DisableRadioModerator", current.Moderation.DisableRadioModerator, desired.Moderation.DisableRadioModerator)
	addIfDifferent("DisableRadioInvisible", current.Moderation.DisableRadioInvisible, desired.Moderation.DisableRadioInvisible)
	addIfDifferent("BanKickGlobalSound", current.Moderation.BanKickGlobalSound, desired.Moderation.BanKickGlobalSound)

	// Steam settings
	addIfDifferent("SteamScoreboard", current.Steam.SteamScoreboard, desired.Steam.SteamScoreboard)

	// Discord settings
	addIfDifferent("DiscordEnable", current.Discord.DiscordEnable, desired.Discord.DiscordEnable)
	addIfDifferent("DiscordToken", current.Discord.DiscordToken, desired.Discord.DiscordToken)
	addIfDifferent("DiscordChannel", current.Discord.DiscordChannel, desired.Discord.DiscordChannel)
	addIfDifferent("DiscordChannelID", current.Discord.DiscordChannelID, desired.Discord.DiscordChannelID)

	// Communication settings
	addIfDifferent("GlobalChat", current.Communication.GlobalChat, desired.Communication.GlobalChat)
	addIfDifferent("ChatStreams", current.Communication.ChatStreams, desired.Communication.ChatStreams)
	addIfDifferent("ServerWelcomeMessage", current.Communication.ServerWelcomeMessage, desired.Communication.ServerWelcomeMessage)
	addIfDifferent("VoiceEnable", current.Communication.VoiceEnable, desired.Communication.VoiceEnable)
	addIfDifferent("VoiceMinDistance", current.Communication.VoiceMinDistance, desired.Communication.VoiceMinDistance)
	addIfDifferent("VoiceMaxDistance", current.Communication.VoiceMaxDistance, desired.Communication.VoiceMaxDistance)
	addIfDifferent("Voice3D", current.Communication.Voice3D, desired.Communication.Voice3D)

	// Gameplay settings
	addIfDifferent("PauseEmpty", current.Gameplay.PauseEmpty, desired.Gameplay.PauseEmpty)
	addIfDifferent("DisplayUserName", current.Gameplay.DisplayUserName, desired.Gameplay.DisplayUserName)
	addIfDifferent("ShowFirstAndLastName", current.Gameplay.ShowFirstAndLastName, desired.Gameplay.ShowFirstAndLastName)
	addIfDifferent("SpawnPoint", current.Gameplay.SpawnPoint, desired.Gameplay.SpawnPoint)
	addIfDifferent("SpawnItems", current.Gameplay.SpawnItems, desired.Gameplay.SpawnItems)
	addIfDifferent("NoFire", current.Gameplay.NoFire, desired.Gameplay.NoFire)
	addIfDifferent("AnnounceDeath", current.Gameplay.AnnounceDeath, desired.Gameplay.AnnounceDeath)
	addIfDifferent("MinutesPerPage", current.Gameplay.MinutesPerPage, desired.Gameplay.MinutesPerPage)
	addIfDifferent("AllowDestructionBySledgehammer", current.Gameplay.AllowDestructionBySledgehammer, desired.Gameplay.AllowDestructionBySledgehammer)
	addIfDifferent("SledgehammerOnlyInSafehouse", current.Gameplay.SledgehammerOnlyInSafehouse, desired.Gameplay.SledgehammerOnlyInSafehouse)
	addIfDifferent("SleepAllowed", current.Gameplay.SleepAllowed, desired.Gameplay.SleepAllowed)
	addIfDifferent("SleepNeeded", current.Gameplay.SleepNeeded, desired.Gameplay.SleepNeeded)
	addIfDifferent("KnockedDownAllowed", current.Gameplay.KnockedDownAllowed, desired.Gameplay.KnockedDownAllowed)
	addIfDifferent("SneakModeHideFromOtherPlayers", current.Gameplay.SneakModeHideFromOtherPlayers, desired.Gameplay.SneakModeHideFromOtherPlayers)
	addIfDifferent("SpeedLimit", current.Gameplay.SpeedLimit, desired.Gameplay.SpeedLimit)
	addIfDifferent("PlayerRespawnWithSelf", current.Gameplay.PlayerRespawnWithSelf, desired.Gameplay.PlayerRespawnWithSelf)
	addIfDifferent("PlayerRespawnWithOther", current.Gameplay.PlayerRespawnWithOther, desired.Gameplay.PlayerRespawnWithOther)
	addIfDifferent("FastForwardMultiplier", current.Gameplay.FastForwardMultiplier, desired.Gameplay.FastForwardMultiplier)

	// PVP settings
	addIfDifferent("PVP", current.PVP.PVP, desired.PVP.PVP)
	addIfDifferent("SafetySystem", current.PVP.SafetySystem, desired.PVP.SafetySystem)
	addIfDifferent("ShowSafety", current.PVP.ShowSafety, desired.PVP.ShowSafety)
	addIfDifferent("SafetyToggleTimer", current.PVP.SafetyToggleTimer, desired.PVP.SafetyToggleTimer)
	addIfDifferent("SafetyCooldownTimer", current.PVP.SafetyCooldownTimer, desired.PVP.SafetyCooldownTimer)
	addIfDifferent("PVPMeleeDamageModifier", current.PVP.PVPMeleeDamageModifier, desired.PVP.PVPMeleeDamageModifier)
	addIfDifferent("PVPFirearmDamageModifier", current.PVP.PVPFirearmDamageModifier, desired.PVP.PVPFirearmDamageModifier)
	addIfDifferent("PVPMeleeWhileHitReaction", current.PVP.PVPMeleeWhileHitReaction, desired.PVP.PVPMeleeWhileHitReaction)

	// Safehouse settings
	addIfDifferent("PlayerSafehouse", current.Safehouse.PlayerSafehouse, desired.Safehouse.PlayerSafehouse)
	addIfDifferent("AdminSafehouse", current.Safehouse.AdminSafehouse, desired.Safehouse.AdminSafehouse)
	addIfDifferent("SafehouseAllowTrepass", current.Safehouse.SafehouseAllowTrepass, desired.Safehouse.SafehouseAllowTrepass)
	addIfDifferent("SafehouseAllowFire", current.Safehouse.SafehouseAllowFire, desired.Safehouse.SafehouseAllowFire)
	addIfDifferent("SafehouseAllowLoot", current.Safehouse.SafehouseAllowLoot, desired.Safehouse.SafehouseAllowLoot)
	addIfDifferent("SafehouseAllowRespawn", current.Safehouse.SafehouseAllowRespawn, desired.Safehouse.SafehouseAllowRespawn)
	addIfDifferent("SafehouseDaySurvivedToClaim", current.Safehouse.SafehouseDaySurvivedToClaim, desired.Safehouse.SafehouseDaySurvivedToClaim)
	addIfDifferent("SafeHouseRemovalTime", current.Safehouse.SafeHouseRemovalTime, desired.Safehouse.SafeHouseRemovalTime)
	addIfDifferent("SafehouseAllowNonResidential", current.Safehouse.SafehouseAllowNonResidential, desired.Safehouse.SafehouseAllowNonResidential)
	addIfDifferent("DisableSafehouseWhenPlayerConnected", current.Safehouse.DisableSafehouseWhenPlayerConnected, desired.Safehouse.DisableSafehouseWhenPlayerConnected)

	// Faction settings
	addIfDifferent("Faction", current.Faction.Faction, desired.Faction.Faction)
	addIfDifferent("FactionDaySurvivedToCreate", current.Faction.FactionDaySurvivedToCreate, desired.Faction.FactionDaySurvivedToCreate)
	addIfDifferent("FactionPlayersRequiredForTag", current.Faction.FactionPlayersRequiredForTag, desired.Faction.FactionPlayersRequiredForTag)

	// Loot settings
	addIfDifferent("HoursForLootRespawn", current.Loot.HoursForLootRespawn, desired.Loot.HoursForLootRespawn)
	addIfDifferent("MaxItemsForLootRespawn", current.Loot.MaxItemsForLootRespawn, desired.Loot.MaxItemsForLootRespawn)
	addIfDifferent("ConstructionPreventsLootRespawn", current.Loot.ConstructionPreventsLootRespawn, desired.Loot.ConstructionPreventsLootRespawn)
	addIfDifferent("ItemNumbersLimitPerContainer", current.Loot.ItemNumbersLimitPerContainer, desired.Loot.ItemNumbersLimitPerContainer)
	addIfDifferent("TrashDeleteAll", current.Loot.TrashDeleteAll, desired.Loot.TrashDeleteAll)

	// AntiCheat settings
	addIfDifferent("DoLuaChecksum", current.AntiCheat.DoLuaChecksum, desired.AntiCheat.DoLuaChecksum)
	addIfDifferent("KickFastPlayers", current.AntiCheat.KickFastPlayers, desired.AntiCheat.KickFastPlayers)

	addIfDifferent("AntiCheatProtectionType1", current.AntiCheat.AntiCheatProtectionType1, desired.AntiCheat.AntiCheatProtectionType1)
	addIfDifferent("AntiCheatProtectionType2", current.AntiCheat.AntiCheatProtectionType2, desired.AntiCheat.AntiCheatProtectionType2)
	addIfDifferent("AntiCheatProtectionType3", current.AntiCheat.AntiCheatProtectionType3, desired.AntiCheat.AntiCheatProtectionType3)
	addIfDifferent("AntiCheatProtectionType4", current.AntiCheat.AntiCheatProtectionType4, desired.AntiCheat.AntiCheatProtectionType4)
	addIfDifferent("AntiCheatProtectionType5", current.AntiCheat.AntiCheatProtectionType5, desired.AntiCheat.AntiCheatProtectionType5)
	addIfDifferent("AntiCheatProtectionType6", current.AntiCheat.AntiCheatProtectionType6, desired.AntiCheat.AntiCheatProtectionType6)
	addIfDifferent("AntiCheatProtectionType7", current.AntiCheat.AntiCheatProtectionType7, desired.AntiCheat.AntiCheatProtectionType7)
	addIfDifferent("AntiCheatProtectionType8", current.AntiCheat.AntiCheatProtectionType8, desired.AntiCheat.AntiCheatProtectionType8)
	addIfDifferent("AntiCheatProtectionType9", current.AntiCheat.AntiCheatProtectionType9, desired.AntiCheat.AntiCheatProtectionType9)
	addIfDifferent("AntiCheatProtectionType10", current.AntiCheat.AntiCheatProtectionType10, desired.AntiCheat.AntiCheatProtectionType10)
	addIfDifferent("AntiCheatProtectionType11", current.AntiCheat.AntiCheatProtectionType11, desired.AntiCheat.AntiCheatProtectionType11)
	addIfDifferent("AntiCheatProtectionType12", current.AntiCheat.AntiCheatProtectionType12, desired.AntiCheat.AntiCheatProtectionType12)
	addIfDifferent("AntiCheatProtectionType13", current.AntiCheat.AntiCheatProtectionType13, desired.AntiCheat.AntiCheatProtectionType13)
	addIfDifferent("AntiCheatProtectionType14", current.AntiCheat.AntiCheatProtectionType14, desired.AntiCheat.AntiCheatProtectionType14)
	addIfDifferent("AntiCheatProtectionType15", current.AntiCheat.AntiCheatProtectionType15, desired.AntiCheat.AntiCheatProtectionType15)
	addIfDifferent("AntiCheatProtectionType16", current.AntiCheat.AntiCheatProtectionType16, desired.AntiCheat.AntiCheatProtectionType16)
	addIfDifferent("AntiCheatProtectionType17", current.AntiCheat.AntiCheatProtectionType17, desired.AntiCheat.AntiCheatProtectionType17)
	addIfDifferent("AntiCheatProtectionType18", current.AntiCheat.AntiCheatProtectionType18, desired.AntiCheat.AntiCheatProtectionType18)
	addIfDifferent("AntiCheatProtectionType19", current.AntiCheat.AntiCheatProtectionType19, desired.AntiCheat.AntiCheatProtectionType19)
	addIfDifferent("AntiCheatProtectionType20", current.AntiCheat.AntiCheatProtectionType20, desired.AntiCheat.AntiCheatProtectionType20)
	addIfDifferent("AntiCheatProtectionType21", current.AntiCheat.AntiCheatProtectionType21, desired.AntiCheat.AntiCheatProtectionType21)
	addIfDifferent("AntiCheatProtectionType22", current.AntiCheat.AntiCheatProtectionType22, desired.AntiCheat.AntiCheatProtectionType22)
	addIfDifferent("AntiCheatProtectionType23", current.AntiCheat.AntiCheatProtectionType23, desired.AntiCheat.AntiCheatProtectionType23)
	addIfDifferent("AntiCheatProtectionType24", current.AntiCheat.AntiCheatProtectionType24, desired.AntiCheat.AntiCheatProtectionType24)

	// AntiCheat threshold multipliers
	addIfDifferent("AntiCheatProtectionType2ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType3ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType3ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType3ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType4ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType4ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType4ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType9ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType9ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType9ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType15ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType15ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType15ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType20ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType20ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType20ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType22ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType22ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType22ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType24ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType24ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType24ThresholdMultiplier)

	return updates
}
