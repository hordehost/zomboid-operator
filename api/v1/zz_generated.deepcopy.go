//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Administrator) DeepCopyInto(out *Administrator) {
	*out = *in
	in.Password.DeepCopyInto(&out.Password)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Administrator.
func (in *Administrator) DeepCopy() *Administrator {
	if in == nil {
		return nil
	}
	out := new(Administrator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AntiCheat) DeepCopyInto(out *AntiCheat) {
	*out = *in
	if in.DoLuaChecksum != nil {
		in, out := &in.DoLuaChecksum, &out.DoLuaChecksum
		*out = new(bool)
		**out = **in
	}
	if in.KickFastPlayers != nil {
		in, out := &in.KickFastPlayers, &out.KickFastPlayers
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType1 != nil {
		in, out := &in.AntiCheatProtectionType1, &out.AntiCheatProtectionType1
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType2 != nil {
		in, out := &in.AntiCheatProtectionType2, &out.AntiCheatProtectionType2
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType3 != nil {
		in, out := &in.AntiCheatProtectionType3, &out.AntiCheatProtectionType3
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType4 != nil {
		in, out := &in.AntiCheatProtectionType4, &out.AntiCheatProtectionType4
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType5 != nil {
		in, out := &in.AntiCheatProtectionType5, &out.AntiCheatProtectionType5
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType6 != nil {
		in, out := &in.AntiCheatProtectionType6, &out.AntiCheatProtectionType6
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType7 != nil {
		in, out := &in.AntiCheatProtectionType7, &out.AntiCheatProtectionType7
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType8 != nil {
		in, out := &in.AntiCheatProtectionType8, &out.AntiCheatProtectionType8
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType9 != nil {
		in, out := &in.AntiCheatProtectionType9, &out.AntiCheatProtectionType9
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType10 != nil {
		in, out := &in.AntiCheatProtectionType10, &out.AntiCheatProtectionType10
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType11 != nil {
		in, out := &in.AntiCheatProtectionType11, &out.AntiCheatProtectionType11
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType12 != nil {
		in, out := &in.AntiCheatProtectionType12, &out.AntiCheatProtectionType12
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType13 != nil {
		in, out := &in.AntiCheatProtectionType13, &out.AntiCheatProtectionType13
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType14 != nil {
		in, out := &in.AntiCheatProtectionType14, &out.AntiCheatProtectionType14
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType15 != nil {
		in, out := &in.AntiCheatProtectionType15, &out.AntiCheatProtectionType15
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType16 != nil {
		in, out := &in.AntiCheatProtectionType16, &out.AntiCheatProtectionType16
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType17 != nil {
		in, out := &in.AntiCheatProtectionType17, &out.AntiCheatProtectionType17
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType18 != nil {
		in, out := &in.AntiCheatProtectionType18, &out.AntiCheatProtectionType18
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType19 != nil {
		in, out := &in.AntiCheatProtectionType19, &out.AntiCheatProtectionType19
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType20 != nil {
		in, out := &in.AntiCheatProtectionType20, &out.AntiCheatProtectionType20
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType21 != nil {
		in, out := &in.AntiCheatProtectionType21, &out.AntiCheatProtectionType21
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType22 != nil {
		in, out := &in.AntiCheatProtectionType22, &out.AntiCheatProtectionType22
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType23 != nil {
		in, out := &in.AntiCheatProtectionType23, &out.AntiCheatProtectionType23
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType24 != nil {
		in, out := &in.AntiCheatProtectionType24, &out.AntiCheatProtectionType24
		*out = new(bool)
		**out = **in
	}
	if in.AntiCheatProtectionType2ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType2ThresholdMultiplier, &out.AntiCheatProtectionType2ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.AntiCheatProtectionType3ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType3ThresholdMultiplier, &out.AntiCheatProtectionType3ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.AntiCheatProtectionType4ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType4ThresholdMultiplier, &out.AntiCheatProtectionType4ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.AntiCheatProtectionType9ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType9ThresholdMultiplier, &out.AntiCheatProtectionType9ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.AntiCheatProtectionType15ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType15ThresholdMultiplier, &out.AntiCheatProtectionType15ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.AntiCheatProtectionType20ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType20ThresholdMultiplier, &out.AntiCheatProtectionType20ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.AntiCheatProtectionType22ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType22ThresholdMultiplier, &out.AntiCheatProtectionType22ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.AntiCheatProtectionType24ThresholdMultiplier != nil {
		in, out := &in.AntiCheatProtectionType24ThresholdMultiplier, &out.AntiCheatProtectionType24ThresholdMultiplier
		*out = new(float32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AntiCheat.
func (in *AntiCheat) DeepCopy() *AntiCheat {
	if in == nil {
		return nil
	}
	out := new(AntiCheat)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backup) DeepCopyInto(out *Backup) {
	*out = *in
	if in.SaveWorldEveryMinutes != nil {
		in, out := &in.SaveWorldEveryMinutes, &out.SaveWorldEveryMinutes
		*out = new(int32)
		**out = **in
	}
	if in.BackupsCount != nil {
		in, out := &in.BackupsCount, &out.BackupsCount
		*out = new(int32)
		**out = **in
	}
	if in.BackupsOnStart != nil {
		in, out := &in.BackupsOnStart, &out.BackupsOnStart
		*out = new(bool)
		**out = **in
	}
	if in.BackupsOnVersionChange != nil {
		in, out := &in.BackupsOnVersionChange, &out.BackupsOnVersionChange
		*out = new(bool)
		**out = **in
	}
	if in.BackupsPeriod != nil {
		in, out := &in.BackupsPeriod, &out.BackupsPeriod
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backup.
func (in *Backup) DeepCopy() *Backup {
	if in == nil {
		return nil
	}
	out := new(Backup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backups) DeepCopyInto(out *Backups) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	if in.Request != nil {
		in, out := &in.Request, &out.Request
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backups.
func (in *Backups) DeepCopy() *Backups {
	if in == nil {
		return nil
	}
	out := new(Backups)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Communication) DeepCopyInto(out *Communication) {
	*out = *in
	if in.GlobalChat != nil {
		in, out := &in.GlobalChat, &out.GlobalChat
		*out = new(bool)
		**out = **in
	}
	if in.ChatStreams != nil {
		in, out := &in.ChatStreams, &out.ChatStreams
		*out = new(string)
		**out = **in
	}
	if in.ServerWelcomeMessage != nil {
		in, out := &in.ServerWelcomeMessage, &out.ServerWelcomeMessage
		*out = new(string)
		**out = **in
	}
	if in.VoiceEnable != nil {
		in, out := &in.VoiceEnable, &out.VoiceEnable
		*out = new(bool)
		**out = **in
	}
	if in.VoiceMinDistance != nil {
		in, out := &in.VoiceMinDistance, &out.VoiceMinDistance
		*out = new(float32)
		**out = **in
	}
	if in.VoiceMaxDistance != nil {
		in, out := &in.VoiceMaxDistance, &out.VoiceMaxDistance
		*out = new(float32)
		**out = **in
	}
	if in.Voice3D != nil {
		in, out := &in.Voice3D, &out.Voice3D
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Communication.
func (in *Communication) DeepCopy() *Communication {
	if in == nil {
		return nil
	}
	out := new(Communication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Discord) DeepCopyInto(out *Discord) {
	*out = *in
	if in.DiscordToken != nil {
		in, out := &in.DiscordToken, &out.DiscordToken
		*out = new(corev1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.DiscordChannel != nil {
		in, out := &in.DiscordChannel, &out.DiscordChannel
		*out = new(corev1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.DiscordChannelID != nil {
		in, out := &in.DiscordChannelID, &out.DiscordChannelID
		*out = new(corev1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Discord.
func (in *Discord) DeepCopy() *Discord {
	if in == nil {
		return nil
	}
	out := new(Discord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Faction) DeepCopyInto(out *Faction) {
	*out = *in
	if in.Faction != nil {
		in, out := &in.Faction, &out.Faction
		*out = new(bool)
		**out = **in
	}
	if in.FactionDaySurvivedToCreate != nil {
		in, out := &in.FactionDaySurvivedToCreate, &out.FactionDaySurvivedToCreate
		*out = new(int32)
		**out = **in
	}
	if in.FactionPlayersRequiredForTag != nil {
		in, out := &in.FactionPlayersRequiredForTag, &out.FactionPlayersRequiredForTag
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Faction.
func (in *Faction) DeepCopy() *Faction {
	if in == nil {
		return nil
	}
	out := new(Faction)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Gameplay) DeepCopyInto(out *Gameplay) {
	*out = *in
	if in.PauseEmpty != nil {
		in, out := &in.PauseEmpty, &out.PauseEmpty
		*out = new(bool)
		**out = **in
	}
	if in.DisplayUserName != nil {
		in, out := &in.DisplayUserName, &out.DisplayUserName
		*out = new(bool)
		**out = **in
	}
	if in.ShowFirstAndLastName != nil {
		in, out := &in.ShowFirstAndLastName, &out.ShowFirstAndLastName
		*out = new(bool)
		**out = **in
	}
	if in.SpawnPoint != nil {
		in, out := &in.SpawnPoint, &out.SpawnPoint
		*out = new(string)
		**out = **in
	}
	if in.SpawnItems != nil {
		in, out := &in.SpawnItems, &out.SpawnItems
		*out = new(string)
		**out = **in
	}
	if in.NoFire != nil {
		in, out := &in.NoFire, &out.NoFire
		*out = new(bool)
		**out = **in
	}
	if in.AnnounceDeath != nil {
		in, out := &in.AnnounceDeath, &out.AnnounceDeath
		*out = new(bool)
		**out = **in
	}
	if in.MinutesPerPage != nil {
		in, out := &in.MinutesPerPage, &out.MinutesPerPage
		*out = new(float32)
		**out = **in
	}
	if in.AllowDestructionBySledgehammer != nil {
		in, out := &in.AllowDestructionBySledgehammer, &out.AllowDestructionBySledgehammer
		*out = new(bool)
		**out = **in
	}
	if in.SledgehammerOnlyInSafehouse != nil {
		in, out := &in.SledgehammerOnlyInSafehouse, &out.SledgehammerOnlyInSafehouse
		*out = new(bool)
		**out = **in
	}
	if in.SleepAllowed != nil {
		in, out := &in.SleepAllowed, &out.SleepAllowed
		*out = new(bool)
		**out = **in
	}
	if in.SleepNeeded != nil {
		in, out := &in.SleepNeeded, &out.SleepNeeded
		*out = new(bool)
		**out = **in
	}
	if in.KnockedDownAllowed != nil {
		in, out := &in.KnockedDownAllowed, &out.KnockedDownAllowed
		*out = new(bool)
		**out = **in
	}
	if in.SneakModeHideFromOtherPlayers != nil {
		in, out := &in.SneakModeHideFromOtherPlayers, &out.SneakModeHideFromOtherPlayers
		*out = new(bool)
		**out = **in
	}
	if in.SpeedLimit != nil {
		in, out := &in.SpeedLimit, &out.SpeedLimit
		*out = new(float32)
		**out = **in
	}
	if in.PlayerRespawnWithSelf != nil {
		in, out := &in.PlayerRespawnWithSelf, &out.PlayerRespawnWithSelf
		*out = new(bool)
		**out = **in
	}
	if in.PlayerRespawnWithOther != nil {
		in, out := &in.PlayerRespawnWithOther, &out.PlayerRespawnWithOther
		*out = new(bool)
		**out = **in
	}
	if in.FastForwardMultiplier != nil {
		in, out := &in.FastForwardMultiplier, &out.FastForwardMultiplier
		*out = new(float32)
		**out = **in
	}
	if in.MapRemotePlayerVisibility != nil {
		in, out := &in.MapRemotePlayerVisibility, &out.MapRemotePlayerVisibility
		*out = new(int32)
		**out = **in
	}
	if in.MouseOverToSeeDisplayName != nil {
		in, out := &in.MouseOverToSeeDisplayName, &out.MouseOverToSeeDisplayName
		*out = new(bool)
		**out = **in
	}
	if in.HidePlayersBehindYou != nil {
		in, out := &in.HidePlayersBehindYou, &out.HidePlayersBehindYou
		*out = new(bool)
		**out = **in
	}
	if in.CarEngineAttractionModifier != nil {
		in, out := &in.CarEngineAttractionModifier, &out.CarEngineAttractionModifier
		*out = new(float32)
		**out = **in
	}
	if in.PlayerBumpPlayer != nil {
		in, out := &in.PlayerBumpPlayer, &out.PlayerBumpPlayer
		*out = new(bool)
		**out = **in
	}
	if in.BloodSplatLifespanDays != nil {
		in, out := &in.BloodSplatLifespanDays, &out.BloodSplatLifespanDays
		*out = new(int32)
		**out = **in
	}
	if in.RemovePlayerCorpsesOnCorpseRemoval != nil {
		in, out := &in.RemovePlayerCorpsesOnCorpseRemoval, &out.RemovePlayerCorpsesOnCorpseRemoval
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Gameplay.
func (in *Gameplay) DeepCopy() *Gameplay {
	if in == nil {
		return nil
	}
	out := new(Gameplay)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Identity) DeepCopyInto(out *Identity) {
	*out = *in
	if in.Public != nil {
		in, out := &in.Public, &out.Public
		*out = new(bool)
		**out = **in
	}
	if in.PublicName != nil {
		in, out := &in.PublicName, &out.PublicName
		*out = new(string)
		**out = **in
	}
	if in.PublicDescription != nil {
		in, out := &in.PublicDescription, &out.PublicDescription
		*out = new(string)
		**out = **in
	}
	if in.ResetID != nil {
		in, out := &in.ResetID, &out.ResetID
		*out = new(int32)
		**out = **in
	}
	if in.ServerPlayerID != nil {
		in, out := &in.ServerPlayerID, &out.ServerPlayerID
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Identity.
func (in *Identity) DeepCopy() *Identity {
	if in == nil {
		return nil
	}
	out := new(Identity)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Logging) DeepCopyInto(out *Logging) {
	*out = *in
	if in.PerkLogs != nil {
		in, out := &in.PerkLogs, &out.PerkLogs
		*out = new(bool)
		**out = **in
	}
	if in.ClientCommandFilter != nil {
		in, out := &in.ClientCommandFilter, &out.ClientCommandFilter
		*out = new(string)
		**out = **in
	}
	if in.ClientActionLogs != nil {
		in, out := &in.ClientActionLogs, &out.ClientActionLogs
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Logging.
func (in *Logging) DeepCopy() *Logging {
	if in == nil {
		return nil
	}
	out := new(Logging)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Loot) DeepCopyInto(out *Loot) {
	*out = *in
	if in.HoursForLootRespawn != nil {
		in, out := &in.HoursForLootRespawn, &out.HoursForLootRespawn
		*out = new(int32)
		**out = **in
	}
	if in.MaxItemsForLootRespawn != nil {
		in, out := &in.MaxItemsForLootRespawn, &out.MaxItemsForLootRespawn
		*out = new(int32)
		**out = **in
	}
	if in.ConstructionPreventsLootRespawn != nil {
		in, out := &in.ConstructionPreventsLootRespawn, &out.ConstructionPreventsLootRespawn
		*out = new(bool)
		**out = **in
	}
	if in.ItemNumbersLimitPerContainer != nil {
		in, out := &in.ItemNumbersLimitPerContainer, &out.ItemNumbersLimitPerContainer
		*out = new(int32)
		**out = **in
	}
	if in.TrashDeleteAll != nil {
		in, out := &in.TrashDeleteAll, &out.TrashDeleteAll
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Loot.
func (in *Loot) DeepCopy() *Loot {
	if in == nil {
		return nil
	}
	out := new(Loot)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Map) DeepCopyInto(out *Map) {
	*out = *in
	if in.Map != nil {
		in, out := &in.Map, &out.Map
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Map.
func (in *Map) DeepCopy() *Map {
	if in == nil {
		return nil
	}
	out := new(Map)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Moderation) DeepCopyInto(out *Moderation) {
	*out = *in
	if in.DisableRadioStaff != nil {
		in, out := &in.DisableRadioStaff, &out.DisableRadioStaff
		*out = new(bool)
		**out = **in
	}
	if in.DisableRadioAdmin != nil {
		in, out := &in.DisableRadioAdmin, &out.DisableRadioAdmin
		*out = new(bool)
		**out = **in
	}
	if in.DisableRadioGM != nil {
		in, out := &in.DisableRadioGM, &out.DisableRadioGM
		*out = new(bool)
		**out = **in
	}
	if in.DisableRadioOverseer != nil {
		in, out := &in.DisableRadioOverseer, &out.DisableRadioOverseer
		*out = new(bool)
		**out = **in
	}
	if in.DisableRadioModerator != nil {
		in, out := &in.DisableRadioModerator, &out.DisableRadioModerator
		*out = new(bool)
		**out = **in
	}
	if in.DisableRadioInvisible != nil {
		in, out := &in.DisableRadioInvisible, &out.DisableRadioInvisible
		*out = new(bool)
		**out = **in
	}
	if in.BanKickGlobalSound != nil {
		in, out := &in.BanKickGlobalSound, &out.BanKickGlobalSound
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Moderation.
func (in *Moderation) DeepCopy() *Moderation {
	if in == nil {
		return nil
	}
	out := new(Moderation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Mods) DeepCopyInto(out *Mods) {
	*out = *in
	if in.WorkshopItems != nil {
		in, out := &in.WorkshopItems, &out.WorkshopItems
		*out = new(string)
		**out = **in
	}
	if in.Mods != nil {
		in, out := &in.Mods, &out.Mods
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Mods.
func (in *Mods) DeepCopy() *Mods {
	if in == nil {
		return nil
	}
	out := new(Mods)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PVP) DeepCopyInto(out *PVP) {
	*out = *in
	if in.PVP != nil {
		in, out := &in.PVP, &out.PVP
		*out = new(bool)
		**out = **in
	}
	if in.SafetySystem != nil {
		in, out := &in.SafetySystem, &out.SafetySystem
		*out = new(bool)
		**out = **in
	}
	if in.ShowSafety != nil {
		in, out := &in.ShowSafety, &out.ShowSafety
		*out = new(bool)
		**out = **in
	}
	if in.SafetyToggleTimer != nil {
		in, out := &in.SafetyToggleTimer, &out.SafetyToggleTimer
		*out = new(int32)
		**out = **in
	}
	if in.SafetyCooldownTimer != nil {
		in, out := &in.SafetyCooldownTimer, &out.SafetyCooldownTimer
		*out = new(int32)
		**out = **in
	}
	if in.PVPMeleeDamageModifier != nil {
		in, out := &in.PVPMeleeDamageModifier, &out.PVPMeleeDamageModifier
		*out = new(float32)
		**out = **in
	}
	if in.PVPFirearmDamageModifier != nil {
		in, out := &in.PVPFirearmDamageModifier, &out.PVPFirearmDamageModifier
		*out = new(float32)
		**out = **in
	}
	if in.PVPMeleeWhileHitReaction != nil {
		in, out := &in.PVPMeleeWhileHitReaction, &out.PVPMeleeWhileHitReaction
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PVP.
func (in *PVP) DeepCopy() *PVP {
	if in == nil {
		return nil
	}
	out := new(PVP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Player) DeepCopyInto(out *Player) {
	*out = *in
	if in.MaxPlayers != nil {
		in, out := &in.MaxPlayers, &out.MaxPlayers
		*out = new(int32)
		**out = **in
	}
	if in.PingLimit != nil {
		in, out := &in.PingLimit, &out.PingLimit
		*out = new(int32)
		**out = **in
	}
	if in.Open != nil {
		in, out := &in.Open, &out.Open
		*out = new(bool)
		**out = **in
	}
	if in.AutoCreateUserInWhiteList != nil {
		in, out := &in.AutoCreateUserInWhiteList, &out.AutoCreateUserInWhiteList
		*out = new(bool)
		**out = **in
	}
	if in.DropOffWhiteListAfterDeath != nil {
		in, out := &in.DropOffWhiteListAfterDeath, &out.DropOffWhiteListAfterDeath
		*out = new(bool)
		**out = **in
	}
	if in.MaxAccountsPerUser != nil {
		in, out := &in.MaxAccountsPerUser, &out.MaxAccountsPerUser
		*out = new(int32)
		**out = **in
	}
	if in.AllowCoop != nil {
		in, out := &in.AllowCoop, &out.AllowCoop
		*out = new(bool)
		**out = **in
	}
	if in.AllowNonAsciiUsername != nil {
		in, out := &in.AllowNonAsciiUsername, &out.AllowNonAsciiUsername
		*out = new(bool)
		**out = **in
	}
	if in.DenyLoginOnOverloadedServer != nil {
		in, out := &in.DenyLoginOnOverloadedServer, &out.DenyLoginOnOverloadedServer
		*out = new(bool)
		**out = **in
	}
	if in.LoginQueueEnabled != nil {
		in, out := &in.LoginQueueEnabled, &out.LoginQueueEnabled
		*out = new(bool)
		**out = **in
	}
	if in.LoginQueueConnectTimeout != nil {
		in, out := &in.LoginQueueConnectTimeout, &out.LoginQueueConnectTimeout
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Player.
func (in *Player) DeepCopy() *Player {
	if in == nil {
		return nil
	}
	out := new(Player)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Safehouse) DeepCopyInto(out *Safehouse) {
	*out = *in
	if in.PlayerSafehouse != nil {
		in, out := &in.PlayerSafehouse, &out.PlayerSafehouse
		*out = new(bool)
		**out = **in
	}
	if in.AdminSafehouse != nil {
		in, out := &in.AdminSafehouse, &out.AdminSafehouse
		*out = new(bool)
		**out = **in
	}
	if in.SafehouseAllowTrepass != nil {
		in, out := &in.SafehouseAllowTrepass, &out.SafehouseAllowTrepass
		*out = new(bool)
		**out = **in
	}
	if in.SafehouseAllowFire != nil {
		in, out := &in.SafehouseAllowFire, &out.SafehouseAllowFire
		*out = new(bool)
		**out = **in
	}
	if in.SafehouseAllowLoot != nil {
		in, out := &in.SafehouseAllowLoot, &out.SafehouseAllowLoot
		*out = new(bool)
		**out = **in
	}
	if in.SafehouseAllowRespawn != nil {
		in, out := &in.SafehouseAllowRespawn, &out.SafehouseAllowRespawn
		*out = new(bool)
		**out = **in
	}
	if in.SafehouseDaySurvivedToClaim != nil {
		in, out := &in.SafehouseDaySurvivedToClaim, &out.SafehouseDaySurvivedToClaim
		*out = new(int32)
		**out = **in
	}
	if in.SafeHouseRemovalTime != nil {
		in, out := &in.SafeHouseRemovalTime, &out.SafeHouseRemovalTime
		*out = new(int32)
		**out = **in
	}
	if in.SafehouseAllowNonResidential != nil {
		in, out := &in.SafehouseAllowNonResidential, &out.SafehouseAllowNonResidential
		*out = new(bool)
		**out = **in
	}
	if in.DisableSafehouseWhenPlayerConnected != nil {
		in, out := &in.DisableSafehouseWhenPlayerConnected, &out.DisableSafehouseWhenPlayerConnected
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Safehouse.
func (in *Safehouse) DeepCopy() *Safehouse {
	if in == nil {
		return nil
	}
	out := new(Safehouse)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Steam) DeepCopyInto(out *Steam) {
	*out = *in
	if in.SteamScoreboard != nil {
		in, out := &in.SteamScoreboard, &out.SteamScoreboard
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Steam.
func (in *Steam) DeepCopy() *Steam {
	if in == nil {
		return nil
	}
	out := new(Steam)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Storage) DeepCopyInto(out *Storage) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	out.Request = in.Request.DeepCopy()
	if in.WorkshopRequest != nil {
		in, out := &in.WorkshopRequest, &out.WorkshopRequest
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Storage.
func (in *Storage) DeepCopy() *Storage {
	if in == nil {
		return nil
	}
	out := new(Storage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkshopMod) DeepCopyInto(out *WorkshopMod) {
	*out = *in
	if in.ModID != nil {
		in, out := &in.ModID, &out.ModID
		*out = new(string)
		**out = **in
	}
	if in.WorkshopID != nil {
		in, out := &in.WorkshopID, &out.WorkshopID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkshopMod.
func (in *WorkshopMod) DeepCopy() *WorkshopMod {
	if in == nil {
		return nil
	}
	out := new(WorkshopMod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZomboidServer) DeepCopyInto(out *ZomboidServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZomboidServer.
func (in *ZomboidServer) DeepCopy() *ZomboidServer {
	if in == nil {
		return nil
	}
	out := new(ZomboidServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZomboidServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZomboidServerList) DeepCopyInto(out *ZomboidServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ZomboidServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZomboidServerList.
func (in *ZomboidServerList) DeepCopy() *ZomboidServerList {
	if in == nil {
		return nil
	}
	out := new(ZomboidServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZomboidServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZomboidServerSpec) DeepCopyInto(out *ZomboidServerSpec) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
	in.Storage.DeepCopyInto(&out.Storage)
	in.Backups.DeepCopyInto(&out.Backups)
	in.Administrator.DeepCopyInto(&out.Administrator)
	if in.Password != nil {
		in, out := &in.Password, &out.Password
		*out = new(corev1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.Suspended != nil {
		in, out := &in.Suspended, &out.Suspended
		*out = new(bool)
		**out = **in
	}
	in.Settings.DeepCopyInto(&out.Settings)
	if in.Discord != nil {
		in, out := &in.Discord, &out.Discord
		*out = new(Discord)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZomboidServerSpec.
func (in *ZomboidServerSpec) DeepCopy() *ZomboidServerSpec {
	if in == nil {
		return nil
	}
	out := new(ZomboidServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZomboidServerStatus) DeepCopyInto(out *ZomboidServerStatus) {
	*out = *in
	if in.SettingsLastObserved != nil {
		in, out := &in.SettingsLastObserved, &out.SettingsLastObserved
		*out = (*in).DeepCopy()
	}
	if in.Settings != nil {
		in, out := &in.Settings, &out.Settings
		*out = new(ZomboidSettings)
		(*in).DeepCopyInto(*out)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZomboidServerStatus.
func (in *ZomboidServerStatus) DeepCopy() *ZomboidServerStatus {
	if in == nil {
		return nil
	}
	out := new(ZomboidServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZomboidSettings) DeepCopyInto(out *ZomboidSettings) {
	*out = *in
	in.Identity.DeepCopyInto(&out.Identity)
	in.Player.DeepCopyInto(&out.Player)
	in.Map.DeepCopyInto(&out.Map)
	in.Mods.DeepCopyInto(&out.Mods)
	if in.WorkshopMods != nil {
		in, out := &in.WorkshopMods, &out.WorkshopMods
		*out = make([]WorkshopMod, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Backup.DeepCopyInto(&out.Backup)
	in.Logging.DeepCopyInto(&out.Logging)
	in.Moderation.DeepCopyInto(&out.Moderation)
	in.Steam.DeepCopyInto(&out.Steam)
	in.Communication.DeepCopyInto(&out.Communication)
	in.Gameplay.DeepCopyInto(&out.Gameplay)
	in.PVP.DeepCopyInto(&out.PVP)
	in.Loot.DeepCopyInto(&out.Loot)
	in.Safehouse.DeepCopyInto(&out.Safehouse)
	in.Faction.DeepCopyInto(&out.Faction)
	in.AntiCheat.DeepCopyInto(&out.AntiCheat)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZomboidSettings.
func (in *ZomboidSettings) DeepCopy() *ZomboidSettings {
	if in == nil {
		return nil
	}
	out := new(ZomboidSettings)
	in.DeepCopyInto(out)
	return out
}
