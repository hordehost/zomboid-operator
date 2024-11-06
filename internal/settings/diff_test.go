package settings

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
)

var _ = Describe("Settings Diff", func() {
	var (
		current zomboidv1.ZomboidSettings
		desired zomboidv1.ZomboidSettings
	)

	BeforeEach(func() {
		// Reset settings before each test
		current = *zomboidv1.DefaultZomboidSettings.DeepCopy()
		desired = zomboidv1.ZomboidSettings{}
	})

	Context("when comparing Identity settings", func() {
		It("should detect changes in Public setting", func() {
			current.Identity.Public = ptr.To(true)
			desired.Identity.Public = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"Public", "false"}))
		})

		It("should detect changes in PublicName", func() {
			current.Identity.PublicName = ptr.To("Server 1")
			desired.Identity.PublicName = ptr.To("Server 2")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"PublicName", "Server 2"}))
		})

		It("should use default value when both values are nil", func() {
			current.Identity.Public = nil
			desired.Identity.Public = nil

			Expect(*zomboidv1.DefaultIdentity.Public).To(BeFalse())

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"Public", "false"}))
		})
	})

	Context("when comparing Player settings", func() {
		It("should detect changes in MaxPlayers", func() {
			current.Player.MaxPlayers = ptr.To(int32(32))
			desired.Player.MaxPlayers = ptr.To(int32(64))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"MaxPlayers", "64"}))
		})

		It("should detect changes in multiple Player settings", func() {
			current.Player.Open = ptr.To(true)
			current.Player.PingLimit = ptr.To(int32(400))
			desired.Player.Open = ptr.To(false)
			desired.Player.PingLimit = ptr.To(int32(500))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"Open", "false"},
				[2]string{"PingLimit", "500"},
			))
		})
	})

	Context("when comparing Map settings", func() {
		It("should detect changes in Map value", func() {
			current.Map.Map = ptr.To("Muldraugh, KY")
			desired.Map.Map = ptr.To("Riverside, KY")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"Map", "Riverside, KY"}))
		})
	})

	Context("when comparing Backup settings", func() {
		It("should detect changes in numeric values", func() {
			current.Backup.BackupsPeriod = ptr.To(int32(0))
			current.Backup.BackupsCount = ptr.To(int32(5))
			desired.Backup.BackupsPeriod = ptr.To(int32(30))
			desired.Backup.BackupsCount = ptr.To(int32(10))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"BackupsPeriod", "30"},
				[2]string{"BackupsCount", "10"},
			))
		})

		It("should detect changes in boolean values", func() {
			current.Backup.BackupsOnStart = ptr.To(true)
			desired.Backup.BackupsOnStart = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"BackupsOnStart", "false"}))
		})
	})

	Context("when comparing Communication settings", func() {
		It("should detect changes in float values", func() {
			current.Communication.VoiceMinDistance = ptr.To(float32(10.1))
			current.Communication.VoiceMaxDistance = ptr.To(float32(100.0))
			desired.Communication.VoiceMinDistance = ptr.To(float32(15.0))
			desired.Communication.VoiceMaxDistance = ptr.To(float32(150.2))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"VoiceMinDistance", "15.0"},
				[2]string{"VoiceMaxDistance", "150.2"},
			))
		})

		It("should handle chat streams changes", func() {
			current.Communication.ChatStreams = ptr.To("s,r,a,w")
			desired.Communication.ChatStreams = ptr.To("s,r,a,w,y")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"ChatStreams", "s,r,a,w,y"}))
		})
	})

	Context("when comparing entire unset structs", func() {
		It("should handle nil struct fields", func() {
			// Only set Communication settings in desired
			desired.Communication.VoiceEnable = ptr.To(false)
			Expect(desired.Communication.VoiceEnable).NotTo(Equal(current.Communication.VoiceEnable))
			Expect(desired.Communication.VoiceEnable).NotTo(Equal(zomboidv1.DefaultCommunication.VoiceEnable))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"VoiceEnable", "false"}))
			Expect(diff).To(HaveLen(1))
		})

		It("should handle completely empty settings", func() {
			diff := SettingsDiff(current, desired)
			Expect(diff).To(BeEmpty())
		})
	})

	Context("when comparing PVP settings", func() {
		It("should detect changes in damage modifiers", func() {
			current.PVP.PVPMeleeDamageModifier = ptr.To(float32(30.0))
			current.PVP.PVPFirearmDamageModifier = ptr.To(float32(50.0))
			desired.PVP.PVPMeleeDamageModifier = ptr.To(float32(40.0))
			desired.PVP.PVPFirearmDamageModifier = ptr.To(float32(60.0))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"PVPMeleeDamageModifier", "40.0"},
				[2]string{"PVPFirearmDamageModifier", "60.0"},
			))
		})
	})

	Context("when comparing Safehouse settings", func() {
		It("should detect changes in time-based settings", func() {
			current.Safehouse.SafehouseDaySurvivedToClaim = ptr.To(int32(0))
			current.Safehouse.SafeHouseRemovalTime = ptr.To(int32(144))
			desired.Safehouse.SafehouseDaySurvivedToClaim = ptr.To(int32(7))
			desired.Safehouse.SafeHouseRemovalTime = ptr.To(int32(168))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"SafehouseDaySurvivedToClaim", "7"},
				[2]string{"SafeHouseRemovalTime", "168"},
			))
		})
	})

	Context("when comparing AntiCheat settings", func() {
		It("should detect changes in protection types", func() {
			current.AntiCheat.AntiCheatProtectionType1 = ptr.To(true)
			desired.AntiCheat.AntiCheatProtectionType1 = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"AntiCheatProtectionType1", "false"}))
		})

		It("should detect changes in threshold multipliers", func() {
			current.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier = ptr.To(float32(3.0))
			desired.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier = ptr.To(float32(5.0))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"AntiCheatProtectionType2ThresholdMultiplier", "5.0"}))
		})
	})

	Context("when comparing Logging settings", func() {
		It("should detect changes in PerkLogs", func() {
			current.Logging.PerkLogs = ptr.To(true)
			desired.Logging.PerkLogs = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"PerkLogs", "false"}))
		})

		It("should detect changes in command filters", func() {
			current.Logging.ClientCommandFilter = ptr.To("-vehicle.*")
			desired.Logging.ClientCommandFilter = ptr.To("-vehicle.*;+vehicle.damageWindow")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"ClientCommandFilter", "-vehicle.*;+vehicle.damageWindow"}))
		})

		It("should detect changes in action logs", func() {
			current.Logging.ClientActionLogs = ptr.To("ISEnterVehicle")
			desired.Logging.ClientActionLogs = ptr.To("ISEnterVehicle;ISExitVehicle")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"ClientActionLogs", "ISEnterVehicle;ISExitVehicle"}))
		})
	})

	Context("when comparing Moderation settings", func() {
		It("should detect changes in radio permissions", func() {
			current.Moderation.DisableRadioStaff = ptr.To(true)
			current.Moderation.DisableRadioAdmin = ptr.To(true)
			desired.Moderation.DisableRadioStaff = ptr.To(false)
			desired.Moderation.DisableRadioAdmin = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"DisableRadioStaff", "false"},
				[2]string{"DisableRadioAdmin", "false"},
			))
		})

		It("should detect changes in ban/kick sounds", func() {
			current.Moderation.BanKickGlobalSound = ptr.To(true)
			desired.Moderation.BanKickGlobalSound = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"BanKickGlobalSound", "false"}))
		})
	})

	Context("when comparing Steam settings", func() {
		It("should detect changes in scoreboard visibility", func() {
			current.Steam.SteamScoreboard = ptr.To("true")
			desired.Steam.SteamScoreboard = ptr.To("admin")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElement([2]string{"SteamScoreboard", "admin"}))
		})
	})

	Context("when comparing Gameplay settings", func() {
		It("should detect changes in display settings", func() {
			current.Gameplay.DisplayUserName = ptr.To(true)
			current.Gameplay.ShowFirstAndLastName = ptr.To(true)
			desired.Gameplay.DisplayUserName = ptr.To(false)
			desired.Gameplay.ShowFirstAndLastName = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"DisplayUserName", "false"},
				[2]string{"ShowFirstAndLastName", "false"},
			))
		})

		It("should detect changes in spawn settings", func() {
			current.Gameplay.SpawnPoint = ptr.To("0,0,0")
			current.Gameplay.SpawnItems = ptr.To("Base.Axe")
			desired.Gameplay.SpawnPoint = ptr.To("100,200,0")
			desired.Gameplay.SpawnItems = ptr.To("Base.Axe,Base.Bag_BigHikingBag")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"SpawnPoint", "100,200,0"},
				[2]string{"SpawnItems", "Base.Axe,Base.Bag_BigHikingBag"},
			))
		})

		It("should detect changes in gameplay mechanics", func() {
			current.Gameplay.KnockedDownAllowed = ptr.To(true)
			current.Gameplay.SneakModeHideFromOtherPlayers = ptr.To(true)
			current.Gameplay.FastForwardMultiplier = ptr.To(float32(40.0))
			desired.Gameplay.KnockedDownAllowed = ptr.To(false)
			desired.Gameplay.SneakModeHideFromOtherPlayers = ptr.To(false)
			desired.Gameplay.FastForwardMultiplier = ptr.To(float32(60.0))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"KnockedDownAllowed", "false"},
				[2]string{"SneakModeHideFromOtherPlayers", "false"},
				[2]string{"FastForwardMultiplier", "60.0"},
			))
		})
	})

	Context("when comparing Loot settings", func() {
		It("should detect changes in loot respawn settings", func() {
			current.Loot.HoursForLootRespawn = ptr.To(int32(0))
			current.Loot.MaxItemsForLootRespawn = ptr.To(int32(4))
			desired.Loot.HoursForLootRespawn = ptr.To(int32(24))
			desired.Loot.MaxItemsForLootRespawn = ptr.To(int32(8))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"HoursForLootRespawn", "24"},
				[2]string{"MaxItemsForLootRespawn", "8"},
			))
		})

		It("should detect changes in container limits", func() {
			current.Loot.ItemNumbersLimitPerContainer = ptr.To(int32(0))
			current.Loot.TrashDeleteAll = ptr.To(true)
			desired.Loot.ItemNumbersLimitPerContainer = ptr.To(int32(1000))
			desired.Loot.TrashDeleteAll = ptr.To(false)

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"ItemNumbersLimitPerContainer", "1000"},
				[2]string{"TrashDeleteAll", "false"},
			))
		})
	})

	Context("when comparing Workshop Mods", func() {
		It("should detect changes in mod lists", func() {
			current.Mods.Mods = ptr.To("mod1;mod2")
			current.Mods.WorkshopItems = ptr.To("2735567460;2392709985")
			desired.Mods.Mods = ptr.To("mod1;mod2;mod3")
			desired.Mods.WorkshopItems = ptr.To("2735567460;2392709985;2392709986")

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"Mods", "mod1;mod2;mod3"},
				[2]string{"WorkshopItems", "2735567460;2392709985;2392709986"},
			))
		})
	})

	Context("when comparing Identity settings", func() {
		It("should detect changes in server identification", func() {
			current.Identity.ResetID = ptr.To(int32(485871306))
			current.Identity.ServerPlayerID = ptr.To(int32(63827612))
			desired.Identity.ResetID = ptr.To(int32(485871307))
			desired.Identity.ServerPlayerID = ptr.To(int32(63827613))

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"ResetID", "485871307"},
				[2]string{"ServerPlayerID", "63827613"},
			))
		})
	})

	Context("when handling nil values", func() {
		It("should use default values when desired setting is nil", func() {
			// Set current values that differ from defaults
			current.Identity.Public = ptr.To(false)
			current.Player.MaxPlayers = ptr.To(int32(50))
			current.Communication.VoiceMinDistance = ptr.To(float32(20.0))

			// Leave desired values nil to request defaults
			desired.Identity.Public = nil                // Default is false
			desired.Player.MaxPlayers = nil              // Default is 32
			desired.Communication.VoiceMinDistance = nil // Default is 10.0

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"MaxPlayers", "32"},
				[2]string{"VoiceMinDistance", "10.0"},
			))
			// Public shouldn't be in diff since current matches default
			Expect(diff).NotTo(ContainElement(ContainElement("Public")))
		})

		It("should handle nil values in complex settings", func() {
			// Set current values
			current.PVP.PVPMeleeDamageModifier = ptr.To(float32(40.0))
			current.PVP.PVPFirearmDamageModifier = ptr.To(float32(60.0))
			current.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier = ptr.To(float32(5.0))

			// Leave desired values nil to request defaults
			desired.PVP.PVPMeleeDamageModifier = nil                            // Default is 30.0
			desired.PVP.PVPFirearmDamageModifier = nil                          // Default is 50.0
			desired.AntiCheat.AntiCheatProtectionType2ThresholdMultiplier = nil // Default is 3.0

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"PVPMeleeDamageModifier", "30.0"},
				[2]string{"PVPFirearmDamageModifier", "50.0"},
				[2]string{"AntiCheatProtectionType2ThresholdMultiplier", "3.0"},
			))
		})

		It("should handle nil boolean values", func() {
			current.Gameplay.PauseEmpty = ptr.To(false)
			current.Gameplay.DisplayUserName = ptr.To(false)
			current.Gameplay.NoFire = ptr.To(true)

			// Leave desired values nil to request defaults
			desired.Gameplay.PauseEmpty = nil      // Default is true
			desired.Gameplay.DisplayUserName = nil // Default is true
			desired.Gameplay.NoFire = nil          // Default is false

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"PauseEmpty", "true"},
				[2]string{"DisplayUserName", "true"},
				[2]string{"NoFire", "false"},
			))
		})

		It("should handle nil string values", func() {
			current.Map.Map = ptr.To("Riverside, KY")
			current.Communication.ChatStreams = ptr.To("s,r,a")
			current.Steam.SteamScoreboard = ptr.To("false")

			// Leave desired values nil to request defaults
			desired.Map.Map = nil                   // Default is "Muldraugh, KY"
			desired.Communication.ChatStreams = nil // Default is "s,r,a,w,y,sh,f,all"
			desired.Steam.SteamScoreboard = nil     // Default is "true"

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"Map", "Muldraugh, KY"},
				[2]string{"ChatStreams", "s,r,a,w,y,sh,f,all"},
				[2]string{"SteamScoreboard", "true"},
			))
		})

		It("should not produce diffs when current matches default and desired is nil", func() {
			// Set current values to match defaults
			current.Identity.Public = ptr.To(false)                        // Matches default
			current.Player.MaxPlayers = ptr.To(int32(32))                  // Matches default
			current.Communication.VoiceMinDistance = ptr.To(float32(10.0)) // Matches default

			// Leave desired values nil
			desired.Identity.Public = nil
			desired.Player.MaxPlayers = nil
			desired.Communication.VoiceMinDistance = nil

			diff := SettingsDiff(current, desired)
			Expect(diff).To(BeEmpty())
		})

		It("should handle partially nil structs", func() {
			// Set some current values different from defaults
			current.Communication.VoiceEnable = ptr.To(false)               // Default is true
			current.Communication.VoiceMinDistance = ptr.To(float32(15.0))  // Default is 10.0
			current.Communication.VoiceMaxDistance = ptr.To(float32(100.0)) // Matches default

			// Mix of nil and non-nil desired values
			desired.Communication.VoiceEnable = nil                        // Request default (true)
			desired.Communication.VoiceMinDistance = ptr.To(float32(20.0)) // Explicit new value
			desired.Communication.VoiceMaxDistance = nil                   // Default matches current, should not appear in diff

			diff := SettingsDiff(current, desired)
			Expect(diff).To(ContainElements(
				[2]string{"VoiceEnable", "true"},
				[2]string{"VoiceMinDistance", "20.0"},
			))
			Expect(diff).NotTo(ContainElement(ContainElement("VoiceMaxDistance")))
		})
	})
})
