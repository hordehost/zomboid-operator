package settings

import (
	"fmt"
	"strconv"
	"strings"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	"k8s.io/utils/ptr"
)

func ParseSettingValue(settings *zomboidv1.ZomboidSettings, key, value string) {
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
	case "MapRemotePlayerVisibility":
		settings.Gameplay.MapRemotePlayerVisibility = parseInt32(value)
	case "MouseOverToSeeDisplayName":
		settings.Gameplay.MouseOverToSeeDisplayName = parseBool(value)
	case "HidePlayersBehindYou":
		settings.Gameplay.HidePlayersBehindYou = parseBool(value)
	case "CarEngineAttractionModifier":
		settings.Gameplay.CarEngineAttractionModifier = parseFloat32(value)
	case "PlayerBumpPlayer":
		settings.Gameplay.PlayerBumpPlayer = parseBool(value)
	case "BloodSplatLifespanDays":
		settings.Gameplay.BloodSplatLifespanDays = parseInt32(value)
	case "RemovePlayerCorpsesOnCorpseRemoval":
		settings.Gameplay.RemovePlayerCorpsesOnCorpseRemoval = parseBool(value)

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

func ValueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case *bool:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
	case *float32:
		if v != nil {
			return fmt.Sprintf("%.1f", *v)
		}
	case *string:
		if v != nil && *v != "" {
			return strings.ReplaceAll(*v, "\n", "\\n")
		}
	case *int32:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
	case string:
		if v != "" {
			return v
		}
	default:
		if v != nil {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}
