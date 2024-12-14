package settings

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	zomboidv1 "github.com/zomboidhost/zomboid-operator/api/v1"
)

const sampleRCONOutput = `
List of Server Options:
* AdminSafehouse=false
* AllowCoop=true
* AllowDestructionBySledgehammer=true
* AllowNonAsciiUsername=false
* AnnounceDeath=false
* Public=true
* PublicName=Test Server
* PublicDescription=First line\nSecond line
* ResetID=485871306
* ServerPlayerID=63827612
* MaxPlayers=32
* PingLimit=400
* Open=true
* AutoCreateUserInWhiteList=false
* Map=Muldraugh, KY
* WorkshopItems=123456;789012
* Mods=mod1;mod2
* SaveWorldEveryMinutes=0
* BackupsCount=5
* BackupsOnStart=true
* BackupsOnVersionChange=true
* BackupsPeriod=0
* PerkLogs=true
* ClientCommandFilter=-vehicle.*;+vehicle.damageWindow;+vehicle.fixPart;+vehicle.installPart;+vehicle.uninstallPart
* ClientActionLogs=ISEnterVehicle;ISExitVehicle;ISTakeEngineParts;
* GlobalChat=true
* ChatStreams=s,r,a,w,y,sh,f,all
* ServerWelcomeMessage=Welcome\nto the server!
* VoiceEnable=true
* VoiceMinDistance=10.0
* VoiceMaxDistance=100.0
* Voice3D=true
* PVP=true
* SafetySystem=true
* ShowSafety=true
* SafetyToggleTimer=2
* SafetyCooldownTimer=3
* PVPMeleeDamageModifier=30.0
* PVPFirearmDamageModifier=50.0
* PVPMeleeWhileHitReaction=false
* DoLuaChecksum=true
* KickFastPlayers=false
* AntiCheatProtectionType1=true
* AntiCheatProtectionType2=true
* AntiCheatProtectionType2ThresholdMultiplier=3.0
`

var _ = Describe("RCON Settings Parser", func() {
	var settings zomboidv1.ZomboidSettings

	BeforeEach(func() {
		ParseRCONShowOptions(sampleRCONOutput, &settings)
	})

	Context("when parsing RCON output", func() {
		It("should parse identity settings correctly", func() {
			Expect(*settings.Identity.Public).To(BeTrue())
			Expect(*settings.Identity.PublicName).To(Equal("Test Server"))
			Expect(*settings.Identity.PublicDescription).To(Equal("First line\nSecond line"))
			Expect(*settings.Identity.ResetID).To(Equal(int32(485871306)))
			Expect(*settings.Identity.ServerPlayerID).To(Equal(int32(63827612)))
		})

		It("should parse player settings correctly", func() {
			Expect(*settings.Player.MaxPlayers).To(Equal(int32(32)))
			Expect(*settings.Player.PingLimit).To(Equal(int32(400)))
			Expect(*settings.Player.Open).To(BeTrue())
			Expect(*settings.Player.AutoCreateUserInWhiteList).To(BeFalse())
			Expect(*settings.Player.AllowCoop).To(BeTrue())
			Expect(*settings.Player.AllowNonAsciiUsername).To(BeFalse())
		})

		It("should parse map and mod settings correctly", func() {
			Expect(*settings.Map.Map).To(Equal("Muldraugh, KY"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("123456;789012"))
			Expect(*settings.Mods.Mods).To(Equal("mod1;mod2"))
		})

		It("should parse backup settings correctly", func() {
			Expect(*settings.Backup.SaveWorldEveryMinutes).To(Equal(int32(0)))
			Expect(*settings.Backup.BackupsCount).To(Equal(int32(5)))
			Expect(*settings.Backup.BackupsOnStart).To(BeTrue())
			Expect(*settings.Backup.BackupsOnVersionChange).To(BeTrue())
			Expect(*settings.Backup.BackupsPeriod).To(Equal(int32(0)))
		})

		It("should parse communication settings correctly", func() {
			Expect(*settings.Communication.GlobalChat).To(BeTrue())
			Expect(*settings.Communication.ChatStreams).To(Equal("s,r,a,w,y,sh,f,all"))
			Expect(*settings.Communication.ServerWelcomeMessage).To(Equal("Welcome\nto the server!"))
			Expect(*settings.Communication.VoiceEnable).To(BeTrue())
			Expect(*settings.Communication.VoiceMinDistance).To(Equal(float32(10.0)))
			Expect(*settings.Communication.VoiceMaxDistance).To(Equal(float32(100.0)))
			Expect(*settings.Communication.Voice3D).To(BeTrue())
		})

		It("should parse PVP settings correctly", func() {
			Expect(*settings.PVP.PVP).To(BeTrue())
			Expect(*settings.PVP.SafetySystem).To(BeTrue())
			Expect(*settings.PVP.ShowSafety).To(BeTrue())
			Expect(*settings.PVP.SafetyToggleTimer).To(Equal(int32(2)))
			Expect(*settings.PVP.SafetyCooldownTimer).To(Equal(int32(3)))
			Expect(*settings.PVP.PVPMeleeDamageModifier).To(Equal(float32(30.0)))
			Expect(*settings.PVP.PVPFirearmDamageModifier).To(Equal(float32(50.0)))
			Expect(*settings.PVP.PVPMeleeWhileHitReaction).To(BeFalse())
		})

		It("should parse anti-cheat settings correctly", func() {
			Expect(*settings.AntiCheat.DoLuaChecksum).To(BeTrue())
			Expect(*settings.AntiCheat.KickFastPlayers).To(BeFalse())
			Expect(*settings.AntiCheat.AntiCheatProtectionType1).To(BeTrue())
			Expect(*settings.AntiCheat.AntiCheatProtectionType2).To(BeTrue())
			Expect(*settings.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier).To(Equal(float32(3.0)))
		})
	})
})
