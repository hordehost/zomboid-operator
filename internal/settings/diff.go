package settings

import (
	"reflect"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
)

// SettingsDiff compares current and desired settings, returning a list of settings that need to be updated
// Each returned pair contains the setting name and its new value as strings
func SettingsDiff(current, desired zomboidv1.ZomboidSettings) [][2]string {
	var updates [][2]string

	// Helper to add a setting update if values differ
	addIfDifferent := func(name string, current, desired, defaultValue interface{}) {
		// If desired is nil or points to nil, we want the default value
		var desiredStr string
		if desired == nil || reflect.ValueOf(desired).IsNil() {
			desiredStr = ValueToString(defaultValue)
		} else {
			desiredStr = ValueToString(desired)
		}

		currentStr := ValueToString(current)
		if currentStr != desiredStr {
			updates = append(updates, [2]string{name, desiredStr})
		}
	}

	// Identity settings
	addIfDifferent("Public", current.Identity.Public, desired.Identity.Public, zomboidv1.DefaultIdentity.Public)
	addIfDifferent("PublicName", current.Identity.PublicName, desired.Identity.PublicName, zomboidv1.DefaultIdentity.PublicName)
	addIfDifferent("PublicDescription", current.Identity.PublicDescription, desired.Identity.PublicDescription, zomboidv1.DefaultIdentity.PublicDescription)
	addIfDifferent("ResetID", current.Identity.ResetID, desired.Identity.ResetID, zomboidv1.DefaultIdentity.ResetID)
	addIfDifferent("ServerPlayerID", current.Identity.ServerPlayerID, desired.Identity.ServerPlayerID, zomboidv1.DefaultIdentity.ServerPlayerID)

	// Player settings
	addIfDifferent("MaxPlayers", current.Player.MaxPlayers, desired.Player.MaxPlayers, zomboidv1.DefaultPlayer.MaxPlayers)
	addIfDifferent("Open", current.Player.Open, desired.Player.Open, zomboidv1.DefaultPlayer.Open)
	addIfDifferent("PingLimit", current.Player.PingLimit, desired.Player.PingLimit, zomboidv1.DefaultPlayer.PingLimit)
	addIfDifferent("AutoCreateUserInWhiteList", current.Player.AutoCreateUserInWhiteList, desired.Player.AutoCreateUserInWhiteList, zomboidv1.DefaultPlayer.AutoCreateUserInWhiteList)
	addIfDifferent("DropOffWhiteListAfterDeath", current.Player.DropOffWhiteListAfterDeath, desired.Player.DropOffWhiteListAfterDeath, zomboidv1.DefaultPlayer.DropOffWhiteListAfterDeath)
	addIfDifferent("MaxAccountsPerUser", current.Player.MaxAccountsPerUser, desired.Player.MaxAccountsPerUser, zomboidv1.DefaultPlayer.MaxAccountsPerUser)
	addIfDifferent("AllowCoop", current.Player.AllowCoop, desired.Player.AllowCoop, zomboidv1.DefaultPlayer.AllowCoop)
	addIfDifferent("AllowNonAsciiUsername", current.Player.AllowNonAsciiUsername, desired.Player.AllowNonAsciiUsername, zomboidv1.DefaultPlayer.AllowNonAsciiUsername)
	addIfDifferent("DenyLoginOnOverloadedServer", current.Player.DenyLoginOnOverloadedServer, desired.Player.DenyLoginOnOverloadedServer, zomboidv1.DefaultPlayer.DenyLoginOnOverloadedServer)
	addIfDifferent("LoginQueueEnabled", current.Player.LoginQueueEnabled, desired.Player.LoginQueueEnabled, zomboidv1.DefaultPlayer.LoginQueueEnabled)
	addIfDifferent("LoginQueueConnectTimeout", current.Player.LoginQueueConnectTimeout, desired.Player.LoginQueueConnectTimeout, zomboidv1.DefaultPlayer.LoginQueueConnectTimeout)

	// Map settings
	addIfDifferent("Map", current.Map.Map, desired.Map.Map, zomboidv1.DefaultMap.Map)

	// Mods settings
	addIfDifferent("Mods", current.Mods.Mods, desired.Mods.Mods, zomboidv1.DefaultMods.Mods)
	addIfDifferent("WorkshopItems", current.Mods.WorkshopItems, desired.Mods.WorkshopItems, zomboidv1.DefaultMods.WorkshopItems)

	// Backup settings
	addIfDifferent("SaveWorldEveryMinutes", current.Backup.SaveWorldEveryMinutes, desired.Backup.SaveWorldEveryMinutes, zomboidv1.DefaultBackup.SaveWorldEveryMinutes)
	addIfDifferent("BackupsCount", current.Backup.BackupsCount, desired.Backup.BackupsCount, zomboidv1.DefaultBackup.BackupsCount)
	addIfDifferent("BackupsOnStart", current.Backup.BackupsOnStart, desired.Backup.BackupsOnStart, zomboidv1.DefaultBackup.BackupsOnStart)
	addIfDifferent("BackupsOnVersionChange", current.Backup.BackupsOnVersionChange, desired.Backup.BackupsOnVersionChange, zomboidv1.DefaultBackup.BackupsOnVersionChange)
	addIfDifferent("BackupsPeriod", current.Backup.BackupsPeriod, desired.Backup.BackupsPeriod, zomboidv1.DefaultBackup.BackupsPeriod)

	// Logging settings
	addIfDifferent("PerkLogs", current.Logging.PerkLogs, desired.Logging.PerkLogs, zomboidv1.DefaultLogging.PerkLogs)
	addIfDifferent("ClientCommandFilter", current.Logging.ClientCommandFilter, desired.Logging.ClientCommandFilter, zomboidv1.DefaultLogging.ClientCommandFilter)
	addIfDifferent("ClientActionLogs", current.Logging.ClientActionLogs, desired.Logging.ClientActionLogs, zomboidv1.DefaultLogging.ClientActionLogs)

	// Moderation settings
	addIfDifferent("DisableRadioStaff", current.Moderation.DisableRadioStaff, desired.Moderation.DisableRadioStaff, zomboidv1.DefaultModeration.DisableRadioStaff)
	addIfDifferent("DisableRadioAdmin", current.Moderation.DisableRadioAdmin, desired.Moderation.DisableRadioAdmin, zomboidv1.DefaultModeration.DisableRadioAdmin)
	addIfDifferent("DisableRadioGM", current.Moderation.DisableRadioGM, desired.Moderation.DisableRadioGM, zomboidv1.DefaultModeration.DisableRadioGM)
	addIfDifferent("DisableRadioOverseer", current.Moderation.DisableRadioOverseer, desired.Moderation.DisableRadioOverseer, zomboidv1.DefaultModeration.DisableRadioOverseer)
	addIfDifferent("DisableRadioModerator", current.Moderation.DisableRadioModerator, desired.Moderation.DisableRadioModerator, zomboidv1.DefaultModeration.DisableRadioModerator)
	addIfDifferent("DisableRadioInvisible", current.Moderation.DisableRadioInvisible, desired.Moderation.DisableRadioInvisible, zomboidv1.DefaultModeration.DisableRadioInvisible)
	addIfDifferent("BanKickGlobalSound", current.Moderation.BanKickGlobalSound, desired.Moderation.BanKickGlobalSound, zomboidv1.DefaultModeration.BanKickGlobalSound)

	// Steam settings
	addIfDifferent("SteamScoreboard", current.Steam.SteamScoreboard, desired.Steam.SteamScoreboard, zomboidv1.DefaultSteam.SteamScoreboard)

	// Communication settings
	addIfDifferent("GlobalChat", current.Communication.GlobalChat, desired.Communication.GlobalChat, zomboidv1.DefaultCommunication.GlobalChat)
	addIfDifferent("ChatStreams", current.Communication.ChatStreams, desired.Communication.ChatStreams, zomboidv1.DefaultCommunication.ChatStreams)
	addIfDifferent("ServerWelcomeMessage", current.Communication.ServerWelcomeMessage, desired.Communication.ServerWelcomeMessage, zomboidv1.DefaultCommunication.ServerWelcomeMessage)
	addIfDifferent("VoiceEnable", current.Communication.VoiceEnable, desired.Communication.VoiceEnable, zomboidv1.DefaultCommunication.VoiceEnable)
	addIfDifferent("VoiceMinDistance", current.Communication.VoiceMinDistance, desired.Communication.VoiceMinDistance, zomboidv1.DefaultCommunication.VoiceMinDistance)
	addIfDifferent("VoiceMaxDistance", current.Communication.VoiceMaxDistance, desired.Communication.VoiceMaxDistance, zomboidv1.DefaultCommunication.VoiceMaxDistance)
	addIfDifferent("Voice3D", current.Communication.Voice3D, desired.Communication.Voice3D, zomboidv1.DefaultCommunication.Voice3D)

	// Gameplay settings
	addIfDifferent("PauseEmpty", current.Gameplay.PauseEmpty, desired.Gameplay.PauseEmpty, zomboidv1.DefaultGameplay.PauseEmpty)
	addIfDifferent("DisplayUserName", current.Gameplay.DisplayUserName, desired.Gameplay.DisplayUserName, zomboidv1.DefaultGameplay.DisplayUserName)
	addIfDifferent("ShowFirstAndLastName", current.Gameplay.ShowFirstAndLastName, desired.Gameplay.ShowFirstAndLastName, zomboidv1.DefaultGameplay.ShowFirstAndLastName)
	addIfDifferent("SpawnPoint", current.Gameplay.SpawnPoint, desired.Gameplay.SpawnPoint, zomboidv1.DefaultGameplay.SpawnPoint)
	addIfDifferent("SpawnItems", current.Gameplay.SpawnItems, desired.Gameplay.SpawnItems, zomboidv1.DefaultGameplay.SpawnItems)
	addIfDifferent("NoFire", current.Gameplay.NoFire, desired.Gameplay.NoFire, zomboidv1.DefaultGameplay.NoFire)
	addIfDifferent("AnnounceDeath", current.Gameplay.AnnounceDeath, desired.Gameplay.AnnounceDeath, zomboidv1.DefaultGameplay.AnnounceDeath)
	addIfDifferent("MinutesPerPage", current.Gameplay.MinutesPerPage, desired.Gameplay.MinutesPerPage, zomboidv1.DefaultGameplay.MinutesPerPage)
	addIfDifferent("AllowDestructionBySledgehammer", current.Gameplay.AllowDestructionBySledgehammer, desired.Gameplay.AllowDestructionBySledgehammer, zomboidv1.DefaultGameplay.AllowDestructionBySledgehammer)
	addIfDifferent("SledgehammerOnlyInSafehouse", current.Gameplay.SledgehammerOnlyInSafehouse, desired.Gameplay.SledgehammerOnlyInSafehouse, zomboidv1.DefaultGameplay.SledgehammerOnlyInSafehouse)
	addIfDifferent("SleepAllowed", current.Gameplay.SleepAllowed, desired.Gameplay.SleepAllowed, zomboidv1.DefaultGameplay.SleepAllowed)
	addIfDifferent("SleepNeeded", current.Gameplay.SleepNeeded, desired.Gameplay.SleepNeeded, zomboidv1.DefaultGameplay.SleepNeeded)
	addIfDifferent("KnockedDownAllowed", current.Gameplay.KnockedDownAllowed, desired.Gameplay.KnockedDownAllowed, zomboidv1.DefaultGameplay.KnockedDownAllowed)
	addIfDifferent("SneakModeHideFromOtherPlayers", current.Gameplay.SneakModeHideFromOtherPlayers, desired.Gameplay.SneakModeHideFromOtherPlayers, zomboidv1.DefaultGameplay.SneakModeHideFromOtherPlayers)
	addIfDifferent("SpeedLimit", current.Gameplay.SpeedLimit, desired.Gameplay.SpeedLimit, zomboidv1.DefaultGameplay.SpeedLimit)
	addIfDifferent("PlayerRespawnWithSelf", current.Gameplay.PlayerRespawnWithSelf, desired.Gameplay.PlayerRespawnWithSelf, zomboidv1.DefaultGameplay.PlayerRespawnWithSelf)
	addIfDifferent("PlayerRespawnWithOther", current.Gameplay.PlayerRespawnWithOther, desired.Gameplay.PlayerRespawnWithOther, zomboidv1.DefaultGameplay.PlayerRespawnWithOther)
	addIfDifferent("FastForwardMultiplier", current.Gameplay.FastForwardMultiplier, desired.Gameplay.FastForwardMultiplier, zomboidv1.DefaultGameplay.FastForwardMultiplier)

	// PVP settings
	addIfDifferent("PVP", current.PVP.PVP, desired.PVP.PVP, zomboidv1.DefaultPVP.PVP)
	addIfDifferent("SafetySystem", current.PVP.SafetySystem, desired.PVP.SafetySystem, zomboidv1.DefaultPVP.SafetySystem)
	addIfDifferent("ShowSafety", current.PVP.ShowSafety, desired.PVP.ShowSafety, zomboidv1.DefaultPVP.ShowSafety)
	addIfDifferent("SafetyToggleTimer", current.PVP.SafetyToggleTimer, desired.PVP.SafetyToggleTimer, zomboidv1.DefaultPVP.SafetyToggleTimer)
	addIfDifferent("SafetyCooldownTimer", current.PVP.SafetyCooldownTimer, desired.PVP.SafetyCooldownTimer, zomboidv1.DefaultPVP.SafetyCooldownTimer)
	addIfDifferent("PVPMeleeDamageModifier", current.PVP.PVPMeleeDamageModifier, desired.PVP.PVPMeleeDamageModifier, zomboidv1.DefaultPVP.PVPMeleeDamageModifier)
	addIfDifferent("PVPFirearmDamageModifier", current.PVP.PVPFirearmDamageModifier, desired.PVP.PVPFirearmDamageModifier, zomboidv1.DefaultPVP.PVPFirearmDamageModifier)
	addIfDifferent("PVPMeleeWhileHitReaction", current.PVP.PVPMeleeWhileHitReaction, desired.PVP.PVPMeleeWhileHitReaction, zomboidv1.DefaultPVP.PVPMeleeWhileHitReaction)

	// Safehouse settings
	addIfDifferent("PlayerSafehouse", current.Safehouse.PlayerSafehouse, desired.Safehouse.PlayerSafehouse, zomboidv1.DefaultSafehouse.PlayerSafehouse)
	addIfDifferent("AdminSafehouse", current.Safehouse.AdminSafehouse, desired.Safehouse.AdminSafehouse, zomboidv1.DefaultSafehouse.AdminSafehouse)
	addIfDifferent("SafehouseAllowTrepass", current.Safehouse.SafehouseAllowTrepass, desired.Safehouse.SafehouseAllowTrepass, zomboidv1.DefaultSafehouse.SafehouseAllowTrepass)
	addIfDifferent("SafehouseAllowFire", current.Safehouse.SafehouseAllowFire, desired.Safehouse.SafehouseAllowFire, zomboidv1.DefaultSafehouse.SafehouseAllowFire)
	addIfDifferent("SafehouseAllowLoot", current.Safehouse.SafehouseAllowLoot, desired.Safehouse.SafehouseAllowLoot, zomboidv1.DefaultSafehouse.SafehouseAllowLoot)
	addIfDifferent("SafehouseAllowRespawn", current.Safehouse.SafehouseAllowRespawn, desired.Safehouse.SafehouseAllowRespawn, zomboidv1.DefaultSafehouse.SafehouseAllowRespawn)
	addIfDifferent("SafehouseDaySurvivedToClaim", current.Safehouse.SafehouseDaySurvivedToClaim, desired.Safehouse.SafehouseDaySurvivedToClaim, zomboidv1.DefaultSafehouse.SafehouseDaySurvivedToClaim)
	addIfDifferent("SafeHouseRemovalTime", current.Safehouse.SafeHouseRemovalTime, desired.Safehouse.SafeHouseRemovalTime, zomboidv1.DefaultSafehouse.SafeHouseRemovalTime)
	addIfDifferent("SafehouseAllowNonResidential", current.Safehouse.SafehouseAllowNonResidential, desired.Safehouse.SafehouseAllowNonResidential, zomboidv1.DefaultSafehouse.SafehouseAllowNonResidential)
	addIfDifferent("DisableSafehouseWhenPlayerConnected", current.Safehouse.DisableSafehouseWhenPlayerConnected, desired.Safehouse.DisableSafehouseWhenPlayerConnected, zomboidv1.DefaultSafehouse.DisableSafehouseWhenPlayerConnected)

	// Faction settings
	addIfDifferent("Faction", current.Faction.Faction, desired.Faction.Faction, zomboidv1.DefaultFaction.Faction)
	addIfDifferent("FactionDaySurvivedToCreate", current.Faction.FactionDaySurvivedToCreate, desired.Faction.FactionDaySurvivedToCreate, zomboidv1.DefaultFaction.FactionDaySurvivedToCreate)
	addIfDifferent("FactionPlayersRequiredForTag", current.Faction.FactionPlayersRequiredForTag, desired.Faction.FactionPlayersRequiredForTag, zomboidv1.DefaultFaction.FactionPlayersRequiredForTag)

	// Loot settings
	addIfDifferent("HoursForLootRespawn", current.Loot.HoursForLootRespawn, desired.Loot.HoursForLootRespawn, zomboidv1.DefaultLoot.HoursForLootRespawn)
	addIfDifferent("MaxItemsForLootRespawn", current.Loot.MaxItemsForLootRespawn, desired.Loot.MaxItemsForLootRespawn, zomboidv1.DefaultLoot.MaxItemsForLootRespawn)
	addIfDifferent("ConstructionPreventsLootRespawn", current.Loot.ConstructionPreventsLootRespawn, desired.Loot.ConstructionPreventsLootRespawn, zomboidv1.DefaultLoot.ConstructionPreventsLootRespawn)
	addIfDifferent("ItemNumbersLimitPerContainer", current.Loot.ItemNumbersLimitPerContainer, desired.Loot.ItemNumbersLimitPerContainer, zomboidv1.DefaultLoot.ItemNumbersLimitPerContainer)
	addIfDifferent("TrashDeleteAll", current.Loot.TrashDeleteAll, desired.Loot.TrashDeleteAll, zomboidv1.DefaultLoot.TrashDeleteAll)

	// AntiCheat settings
	addIfDifferent("DoLuaChecksum", current.AntiCheat.DoLuaChecksum, desired.AntiCheat.DoLuaChecksum, zomboidv1.DefaultAntiCheat.DoLuaChecksum)
	addIfDifferent("KickFastPlayers", current.AntiCheat.KickFastPlayers, desired.AntiCheat.KickFastPlayers, zomboidv1.DefaultAntiCheat.KickFastPlayers)

	addIfDifferent("AntiCheatProtectionType1", current.AntiCheat.AntiCheatProtectionType1, desired.AntiCheat.AntiCheatProtectionType1, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType1)
	addIfDifferent("AntiCheatProtectionType2", current.AntiCheat.AntiCheatProtectionType2, desired.AntiCheat.AntiCheatProtectionType2, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType2)
	addIfDifferent("AntiCheatProtectionType3", current.AntiCheat.AntiCheatProtectionType3, desired.AntiCheat.AntiCheatProtectionType3, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType3)
	addIfDifferent("AntiCheatProtectionType4", current.AntiCheat.AntiCheatProtectionType4, desired.AntiCheat.AntiCheatProtectionType4, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType4)
	addIfDifferent("AntiCheatProtectionType5", current.AntiCheat.AntiCheatProtectionType5, desired.AntiCheat.AntiCheatProtectionType5, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType5)
	addIfDifferent("AntiCheatProtectionType6", current.AntiCheat.AntiCheatProtectionType6, desired.AntiCheat.AntiCheatProtectionType6, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType6)
	addIfDifferent("AntiCheatProtectionType7", current.AntiCheat.AntiCheatProtectionType7, desired.AntiCheat.AntiCheatProtectionType7, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType7)
	addIfDifferent("AntiCheatProtectionType8", current.AntiCheat.AntiCheatProtectionType8, desired.AntiCheat.AntiCheatProtectionType8, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType8)
	addIfDifferent("AntiCheatProtectionType9", current.AntiCheat.AntiCheatProtectionType9, desired.AntiCheat.AntiCheatProtectionType9, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType9)
	addIfDifferent("AntiCheatProtectionType10", current.AntiCheat.AntiCheatProtectionType10, desired.AntiCheat.AntiCheatProtectionType10, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType10)
	addIfDifferent("AntiCheatProtectionType11", current.AntiCheat.AntiCheatProtectionType11, desired.AntiCheat.AntiCheatProtectionType11, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType11)
	addIfDifferent("AntiCheatProtectionType12", current.AntiCheat.AntiCheatProtectionType12, desired.AntiCheat.AntiCheatProtectionType12, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType12)
	addIfDifferent("AntiCheatProtectionType13", current.AntiCheat.AntiCheatProtectionType13, desired.AntiCheat.AntiCheatProtectionType13, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType13)
	addIfDifferent("AntiCheatProtectionType14", current.AntiCheat.AntiCheatProtectionType14, desired.AntiCheat.AntiCheatProtectionType14, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType14)
	addIfDifferent("AntiCheatProtectionType15", current.AntiCheat.AntiCheatProtectionType15, desired.AntiCheat.AntiCheatProtectionType15, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType15)
	addIfDifferent("AntiCheatProtectionType16", current.AntiCheat.AntiCheatProtectionType16, desired.AntiCheat.AntiCheatProtectionType16, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType16)
	addIfDifferent("AntiCheatProtectionType17", current.AntiCheat.AntiCheatProtectionType17, desired.AntiCheat.AntiCheatProtectionType17, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType17)
	addIfDifferent("AntiCheatProtectionType18", current.AntiCheat.AntiCheatProtectionType18, desired.AntiCheat.AntiCheatProtectionType18, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType18)
	addIfDifferent("AntiCheatProtectionType19", current.AntiCheat.AntiCheatProtectionType19, desired.AntiCheat.AntiCheatProtectionType19, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType19)
	addIfDifferent("AntiCheatProtectionType20", current.AntiCheat.AntiCheatProtectionType20, desired.AntiCheat.AntiCheatProtectionType20, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType20)
	addIfDifferent("AntiCheatProtectionType21", current.AntiCheat.AntiCheatProtectionType21, desired.AntiCheat.AntiCheatProtectionType21, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType21)
	addIfDifferent("AntiCheatProtectionType22", current.AntiCheat.AntiCheatProtectionType22, desired.AntiCheat.AntiCheatProtectionType22, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType22)
	addIfDifferent("AntiCheatProtectionType23", current.AntiCheat.AntiCheatProtectionType23, desired.AntiCheat.AntiCheatProtectionType23, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType23)
	addIfDifferent("AntiCheatProtectionType24", current.AntiCheat.AntiCheatProtectionType24, desired.AntiCheat.AntiCheatProtectionType24, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType24)

	// AntiCheat threshold multipliers
	addIfDifferent("AntiCheatProtectionType2ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType2ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType3ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType3ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType3ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType3ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType4ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType4ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType4ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType4ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType9ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType9ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType9ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType9ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType15ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType15ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType15ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType15ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType20ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType20ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType20ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType20ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType22ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType22ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType22ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType22ThresholdMultiplier)
	addIfDifferent("AntiCheatProtectionType24ThresholdMultiplier", current.AntiCheat.AntiCheatProtectionType24ThresholdMultiplier, desired.AntiCheat.AntiCheatProtectionType24ThresholdMultiplier, zomboidv1.DefaultAntiCheat.AntiCheatProtectionType24ThresholdMultiplier)

	return updates
}
