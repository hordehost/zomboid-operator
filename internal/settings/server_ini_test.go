package settings

import (
	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func init() {
	format.MaxLength = 0
}

const referenceServerINI = `
[Server]
Public=true
PublicName=Test Server
PublicDescription=First line\nSecond line
ResetID=485871306
ServerPlayerID=63827612
MaxPlayers=32
PingLimit=400
Open=true
AutoCreateUserInWhiteList=false
DropOffWhiteListAfterDeath=false
MaxAccountsPerUser=0
AllowCoop=true
AllowNonAsciiUsername=false
DenyLoginOnOverloadedServer=true
LoginQueueEnabled=false
LoginQueueConnectTimeout=60
Map=Muldraugh, KY
WorkshopItems=123456;789012
Mods=mod1;mod2
SaveWorldEveryMinutes=0
BackupsCount=5
BackupsOnStart=true
BackupsOnVersionChange=true
BackupsPeriod=0
PerkLogs=true
ClientCommandFilter=-vehicle.*;+vehicle.damageWindow;+vehicle.fixPart;+vehicle.installPart;+vehicle.uninstallPart
ClientActionLogs=ISEnterVehicle;ISExitVehicle;ISTakeEngineParts;
DisableRadioStaff=false
DisableRadioAdmin=true
DisableRadioGM=true
DisableRadioOverseer=false
DisableRadioModerator=false
DisableRadioInvisible=true
BanKickGlobalSound=true
SteamScoreboard=true
DiscordEnable=false
DiscordToken=
DiscordChannel=
DiscordChannelID=
GlobalChat=true
ChatStreams=s,r,a,w,y,sh,f,all
ServerWelcomeMessage=Welcome\nto the server!
VoiceEnable=true
VoiceMinDistance=10.0
VoiceMaxDistance=100.0
Voice3D=true
PVP=true
PauseEmpty=true
DisplayUserName=true
ShowFirstAndLastName=false
SpawnPoint=0,0,0
SpawnItems=Base.Axe,Base.Bag_BigHikingBag
NoFire=false
AnnounceDeath=false
MinutesPerPage=1.0
AllowDestructionBySledgehammer=true
SledgehammerOnlyInSafehouse=false
SleepAllowed=false
SleepNeeded=false
KnockedDownAllowed=true
SneakModeHideFromOtherPlayers=true
SpeedLimit=70.0
PlayerRespawnWithSelf=false
PlayerRespawnWithOther=false
FastForwardMultiplier=40.0
SafetySystem=true
ShowSafety=true
SafetyToggleTimer=2
SafetyCooldownTimer=3
PVPMeleeDamageModifier=30.0
PVPFirearmDamageModifier=50.0
PVPMeleeWhileHitReaction=false
PlayerSafehouse=false
AdminSafehouse=false
SafehouseAllowTrepass=true
SafehouseAllowFire=true
SafehouseAllowLoot=true
SafehouseAllowRespawn=false
SafehouseDaySurvivedToClaim=0
SafeHouseRemovalTime=144
SafehouseAllowNonResidential=false
DisableSafehouseWhenPlayerConnected=false
Faction=true
FactionDaySurvivedToCreate=0
FactionPlayersRequiredForTag=1
HoursForLootRespawn=0
MaxItemsForLootRespawn=4
ConstructionPreventsLootRespawn=true
ItemNumbersLimitPerContainer=0
TrashDeleteAll=false
DoLuaChecksum=true
KickFastPlayers=false
AntiCheatProtectionType1=true
AntiCheatProtectionType2=true
AntiCheatProtectionType3=true
AntiCheatProtectionType4=true
AntiCheatProtectionType5=true
AntiCheatProtectionType6=true
AntiCheatProtectionType7=true
AntiCheatProtectionType8=true
AntiCheatProtectionType9=true
AntiCheatProtectionType10=true
AntiCheatProtectionType11=true
AntiCheatProtectionType12=true
AntiCheatProtectionType13=true
AntiCheatProtectionType14=true
AntiCheatProtectionType15=true
AntiCheatProtectionType16=true
AntiCheatProtectionType17=true
AntiCheatProtectionType18=true
AntiCheatProtectionType19=true
AntiCheatProtectionType20=true
AntiCheatProtectionType21=true
AntiCheatProtectionType22=true
AntiCheatProtectionType23=true
AntiCheatProtectionType24=true
AntiCheatProtectionType2ThresholdMultiplier=3.0
AntiCheatProtectionType3ThresholdMultiplier=1.0
AntiCheatProtectionType4ThresholdMultiplier=1.0
AntiCheatProtectionType9ThresholdMultiplier=1.0
AntiCheatProtectionType15ThresholdMultiplier=1.0
AntiCheatProtectionType20ThresholdMultiplier=1.0
AntiCheatProtectionType22ThresholdMultiplier=1.0
AntiCheatProtectionType24ThresholdMultiplier=6.0
`

var _ = Describe("Project Zomboid server.ini handling", func() {
	var (
		settings             zomboidv1.ZomboidSettings
		generatedINI         string
		roundTrippedSettings zomboidv1.ZomboidSettings
	)

	BeforeEach(func() {
		settings = ParseServerINI(referenceServerINI)
		generatedINI = GenerateServerINI(settings)
		roundTrippedSettings = ParseServerINI(generatedINI)
	})

	It("should maintain settings after round trip", func() {
		Expect(roundTrippedSettings).To(Equal(settings))
	})

	Context("Identity settings", func() {
		It("should parse identity settings correctly", func() {
			Expect(*settings.Identity.Public).To(BeTrue())
			Expect(*settings.Identity.PublicName).To(Equal("Test Server"))
			Expect(*settings.Identity.PublicDescription).To(Equal("First line\nSecond line"))
			Expect(*settings.Identity.ResetID).To(Equal(int32(485871306)))
			Expect(*settings.Identity.ServerPlayerID).To(Equal(int32(63827612)))
		})
	})

	Context("Generated INI content", func() {
		expectedSettings := []string{
			// Identity
			"Public=true",
			"PublicName=Test Server",
			"PublicDescription=First line\\nSecond line",
			"ResetID=485871306",
			"ServerPlayerID=63827612",

			// Player
			"MaxPlayers=32",
			"PingLimit=400",
			"Open=true",
			"AutoCreateUserInWhiteList=false",
			"DropOffWhiteListAfterDeath=false",
			"MaxAccountsPerUser=0",
			"AllowCoop=true",
			"AllowNonAsciiUsername=false",
			"DenyLoginOnOverloadedServer=true",
			"LoginQueueEnabled=false",
			"LoginQueueConnectTimeout=60",

			// Map and Mods
			"Map=Muldraugh, KY",
			"WorkshopItems=123456;789012",
			"Mods=mod1;mod2",

			// Backup
			"SaveWorldEveryMinutes=0",
			"BackupsCount=5",
			"BackupsOnStart=true",
			"BackupsOnVersionChange=true",
			"BackupsPeriod=0",

			// Logging
			"PerkLogs=true",
			"ClientCommandFilter=-vehicle.*;+vehicle.damageWindow;+vehicle.fixPart;+vehicle.installPart;+vehicle.uninstallPart",
			"ClientActionLogs=ISEnterVehicle;ISExitVehicle;ISTakeEngineParts;",

			// Communication
			"GlobalChat=true",
			"ChatStreams=s,r,a,w,y,sh,f,all",
			"ServerWelcomeMessage=Welcome\\nto the server!",
			"VoiceEnable=true",
			"VoiceMinDistance=10.0",
			"VoiceMaxDistance=100.0",
			"Voice3D=true",

			// Gameplay
			"PVP=true",
			"PauseEmpty=true",
			"DisplayUserName=true",
			"ShowFirstAndLastName=false",
			"SpawnPoint=0,0,0",
			"SpawnItems=Base.Axe,Base.Bag_BigHikingBag",
			"NoFire=false",
			"AnnounceDeath=false",
			"MinutesPerPage=1.0",
			"SpeedLimit=70.0",
			"FastForwardMultiplier=40.0",

			// PVP
			"SafetySystem=true",
			"ShowSafety=true",
			"SafetyToggleTimer=2",
			"SafetyCooldownTimer=3",
			"PVPMeleeDamageModifier=30.0",
			"PVPFirearmDamageModifier=50.0",
			"PVPMeleeWhileHitReaction=false",

			// Safehouse
			"PlayerSafehouse=false",
			"AdminSafehouse=false",
			"SafehouseAllowTrepass=true",
			"SafehouseAllowFire=true",
			"SafehouseAllowLoot=true",
			"SafehouseAllowRespawn=false",
			"SafehouseDaySurvivedToClaim=0",
			"SafeHouseRemovalTime=144",

			// AntiCheat
			"DoLuaChecksum=true",
			"KickFastPlayers=false",
			"AntiCheatProtectionType1=true",
			"AntiCheatProtectionType24=true",
			"AntiCheatProtectionType2ThresholdMultiplier=3.0",
			"AntiCheatProtectionType24ThresholdMultiplier=6.0",
		}

		It("should contain all expected settings", func() {
			for _, expected := range expectedSettings {
				Expect(generatedINI).To(ContainSubstring(expected))
			}
		})

		It("should format floating point values correctly", func() {
			Expect(generatedINI).To(ContainSubstring("VoiceMinDistance=10.0"))
			Expect(generatedINI).To(ContainSubstring("VoiceMaxDistance=100.0"))
			Expect(generatedINI).To(ContainSubstring("MinutesPerPage=1.0"))
		})

		It("should format boolean values correctly", func() {
			Expect(generatedINI).To(ContainSubstring("Public=true"))
			Expect(generatedINI).To(ContainSubstring("AutoCreateUserInWhiteList=false"))
		})

		It("should handle newlines in strings correctly", func() {
			Expect(generatedINI).To(ContainSubstring("ServerWelcomeMessage=Welcome\\nto the server!"))
		})
	})

	Context("Specific value types", func() {
		It("should parse float32 values correctly", func() {
			Expect(*settings.Communication.VoiceMinDistance).To(Equal(float32(10.0)))
			Expect(*settings.Communication.VoiceMaxDistance).To(Equal(float32(100.0)))
			Expect(*settings.PVP.PVPMeleeDamageModifier).To(Equal(float32(30.0)))
			Expect(*settings.PVP.PVPFirearmDamageModifier).To(Equal(float32(50.0)))
		})

		It("should parse string values correctly", func() {
			Expect(*settings.Mods.WorkshopItems).To(Equal("123456;789012"))
			Expect(*settings.Mods.Mods).To(Equal("mod1;mod2"))
		})
	})
})
