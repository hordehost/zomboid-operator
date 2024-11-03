package v1

// Identity controls how your server appears in server browsers and handles player authentication,
// including server naming, visibility, password protection, and character persistence across server resets
type Identity struct {
	// Public determines if server is visible in in-game browser. Note: Steam-enabled servers are always visible in Steam browser.
	// +kubebuilder:default=false
	Public bool `json:"Public"`

	// PublicName is the server name shown in browsers
	// +kubebuilder:default="My PZ Server"
	PublicName string `json:"PublicName"`

	// PublicDescription is the server description shown in browsers. Use \n for newlines.
	PublicDescription string `json:"PublicDescription"`

	// ResetID determines if server has undergone soft-reset. If this number doesn't match client, client must create new character.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=485871306
	ResetID int32 `json:"ResetID"`

	// ServerPlayerID identifies characters from different servers. Used with ResetID.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=63827612
	ServerPlayerID int32 `json:"ServerPlayerID"`
}

// Player manages the core multiplayer experience including server capacity, connection requirements,
// whitelist behavior, and basic multiplayer features like co-op and username restrictions
type Player struct {
	// MaxPlayers is maximum concurrent players excluding admins. WARNING: Values above 32 may cause poor map streaming and desync.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=32
	MaxPlayers int32 `json:"MaxPlayers"`

	// PingLimit is max ping in ms before kick. Set to 100 to disable.
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=400
	PingLimit int32 `json:"PingLimit"`

	// Open allows joining without whitelist account. If false, admins must manually create accounts.
	// +kubebuilder:default=true
	Open bool `json:"Open"`

	// AutoCreateUserInWhiteList adds unknown users to whitelist. Only for Open=true servers.
	// +kubebuilder:default=false
	AutoCreateUserInWhiteList bool `json:"AutoCreateUserInWhiteList"`

	// DropOffWhiteListAfterDeath removes accounts after death. Prevents new characters after death on Open=false servers.
	// +kubebuilder:default=false
	DropOffWhiteListAfterDeath bool `json:"DropOffWhiteListAfterDeath"`

	// MaxAccountsPerUser limits accounts per Steam user. Ignored when using Host button.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	MaxAccountsPerUser int32 `json:"MaxAccountsPerUser"`

	// AllowCoop enables splitscreen/co-op play
	// +kubebuilder:default=true
	AllowCoop bool `json:"AllowCoop"`

	// AllowNonAsciiUsername enables non-ASCII characters in usernames
	// +kubebuilder:default=false
	AllowNonAsciiUsername bool `json:"AllowNonAsciiUsername"`

	// DenyLoginOnOverloadedServer prevents logins when server is overloaded
	// +kubebuilder:default=true
	DenyLoginOnOverloadedServer bool `json:"DenyLoginOnOverloadedServer"`

	// LoginQueueEnabled enables login queue
	// +kubebuilder:default=false
	LoginQueueEnabled bool `json:"LoginQueueEnabled"`

	// LoginQueueConnectTimeout is timeout for login queue in seconds
	// +kubebuilder:validation:Minimum=20
	// +kubebuilder:validation:Maximum=1200
	// +kubebuilder:default=60
	LoginQueueConnectTimeout int32 `json:"LoginQueueConnectTimeout"`
}

// Map specifies which game world players will spawn into and explore,
// supporting both vanilla maps and custom map mods from the Steam Workshop
type Map struct {
	// Map is the folder name of the map mod. Found in Steam/steamapps/workshop/modID/mods/modName/media/maps/
	// +kubebuilder:default="Muldraugh, KY"
	Map string `json:"Map"`
}

// Mods provides two parallel lists for managing Steam Workshop content:
// one for Workshop IDs to download mods, and another for mod IDs to load them
type Mods struct {
	// WorkshopItems lists Workshop Mod IDs to download. Separate with semicolons.
	WorkshopItems string `json:"WorkshopItems"`

	// Mods lists mod loading IDs. Found in Steam/steamapps/workshop/modID/mods/modName/info.txt
	Mods string `json:"Mods"`
}

// WorkshopMod pairs a mod's loading ID with its Steam Workshop ID,
// providing a more structured way to specify mods compared to the classic parallel lists
type WorkshopMod struct {
	// ModID is the mod loading ID found in Steam/steamapps/workshop/modID/mods/modName/info.txt
	// +kubebuilder:validation:Required
	ModID string `json:"modID"`

	// WorkshopID is the Steam Workshop ID used to download the mod
	// +kubebuilder:validation:Required
	WorkshopID string `json:"workshopID"`
}

// Backup handles automatic world saving and backup creation,
// protecting against data loss from crashes, updates, or corruption
type Backup struct {
	// SaveWorldEveryMinutes is how often loaded map parts are saved. Map usually only saves when clients leave area.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	SaveWorldEveryMinutes int32 `json:"SaveWorldEveryMinutes"`

	// BackupsCount is the number of backups to keep
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=300
	// +kubebuilder:default=5
	BackupsCount int32 `json:"BackupsCount"`

	// BackupsOnStart enables backups when server starts
	// +kubebuilder:default=true
	BackupsOnStart bool `json:"BackupsOnStart"`

	// BackupsOnVersionChange enables backups on version changes
	// +kubebuilder:default=true
	BackupsOnVersionChange bool `json:"BackupsOnVersionChange"`

	// BackupsPeriod is the backup interval in minutes
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1500
	// +kubebuilder:default=0
	BackupsPeriod int32 `json:"BackupsPeriod"`
}

// Logging configures what client actions and commands are logged to files
type Logging struct {
	// PerkLogs enables tracking player perk changes in PerkLog.txt
	// +kubebuilder:default=true
	PerkLogs bool `json:"PerkLogs"`

	// ClientCommandFilter lists commands not written to cmd.txt log
	// +kubebuilder:default="-vehicle.*;+vehicle.damageWindow;+vehicle.fixPart;+vehicle.installPart;+vehicle.uninstallPart"
	ClientCommandFilter string `json:"ClientCommandFilter"`

	// ClientActionLogs lists actions written to ClientActionLogs.txt
	// +kubebuilder:default="ISEnterVehicle;ISExitVehicle;ISTakeEngineParts;"
	ClientActionLogs string `json:"ClientActionLogs"`
}

// Moderation customizes staff member capabilities and restrictions,
// particularly focusing on radio usage permissions for different staff roles
type Moderation struct {
	// DisableRadioStaff disables radio for staff
	// +kubebuilder:default=false
	DisableRadioStaff bool `json:"DisableRadioStaff"`

	// DisableRadioAdmin disables radio for admins
	// +kubebuilder:default=true
	DisableRadioAdmin bool `json:"DisableRadioAdmin"`

	// DisableRadioGM disables radio for GMs
	// +kubebuilder:default=true
	DisableRadioGM bool `json:"DisableRadioGM"`

	// DisableRadioOverseer disables radio for overseers
	// +kubebuilder:default=false
	DisableRadioOverseer bool `json:"DisableRadioOverseer"`

	// DisableRadioModerator disables radio for moderators
	// +kubebuilder:default=false
	DisableRadioModerator bool `json:"DisableRadioModerator"`

	// DisableRadioInvisible disables radio for invisible players
	// +kubebuilder:default=true
	DisableRadioInvisible bool `json:"DisableRadioInvisible"`

	// BanKickGlobalSound enables global sound on ban/kick
	// +kubebuilder:default=true
	BanKickGlobalSound bool `json:"BanKickGlobalSound"`
}

// Steam manages Steam platform integration and anti-cheat measures,
// including VAC and player visibility settings
type Steam struct {
	// SteamScoreboard controls visibility of Steam names/avatars. Can be "true" (visible to everyone), "false" (visible to no one), or "admin" (visible to only admins)
	// +kubebuilder:default="true"
	SteamScoreboard string `json:"SteamScoreboard"`
}

// Discord enables and configures integration with Discord,
// allowing server chat to be bridged with a Discord channel
type Discord struct {
	// DiscordEnable enables Discord chat integration
	// +kubebuilder:default=false
	DiscordEnable bool `json:"DiscordEnable"`

	// DiscordToken is the Discord bot token
	DiscordToken string `json:"DiscordToken"`

	// DiscordChannel is the Discord channel name. Try channel ID if having difficulties.
	DiscordChannel string `json:"DiscordChannel"`

	// DiscordChannelID is the Discord channel ID. Use if having difficulties with channel name.
	DiscordChannelID string `json:"DiscordChannelID"`
}

// Communication manages all player interaction features including global chat,
// chat streams, welcome messages, and VOIP with distance-based audio
type Communication struct {
	// GlobalChat enables global chat
	// +kubebuilder:default=true
	GlobalChat bool `json:"GlobalChat"`

	// ChatStreams lists available chat streams
	// +kubebuilder:default="s,r,a,w,y,sh,f,all"
	ChatStreams string `json:"ChatStreams"`

	// ServerWelcomeMessage is shown to players on login. Use <LINE> for newlines and <RGB:r,g,b> for colors.
	ServerWelcomeMessage string `json:"ServerWelcomeMessage"`

	// VoiceEnable enables VOIP
	// +kubebuilder:default=true
	VoiceEnable bool `json:"VoiceEnable"`

	// VoiceMinDistance is minimum VOIP audible distance
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=100000.00
	// +kubebuilder:default=10.00
	VoiceMinDistance float32 `json:"VoiceMinDistance"`

	// VoiceMaxDistance is maximum VOIP audible distance
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=100000.00
	// +kubebuilder:default=100.00
	VoiceMaxDistance float32 `json:"VoiceMaxDistance"`

	// Voice3D enables directional VOIP audio
	// +kubebuilder:default=true
	Voice3D bool `json:"Voice3D"`
}

// Gameplay controls fundamental aspects of the player experience including
// PvP, time progression, player visibility, spawning, movement, and sleep mechanics
type Gameplay struct {
	// PVP enables player vs player combat
	// +kubebuilder:default=true
	PVP bool `json:"PVP"`

	// PauseEmpty pauses time when no players online
	// +kubebuilder:default=true
	PauseEmpty bool `json:"PauseEmpty"`

	// DisplayUserName shows player names
	// +kubebuilder:default=true
	DisplayUserName bool `json:"DisplayUserName"`

	// ShowFirstAndLastName shows full player names
	// +kubebuilder:default=false
	ShowFirstAndLastName bool `json:"ShowFirstAndLastName"`

	// SpawnPoint forces spawn location (x,y,z). Find coordinates at map.projectzomboid.com. Ignored when 0,0,0.
	// +kubebuilder:default="0,0,0"
	SpawnPoint string `json:"SpawnPoint"`

	// SpawnItems lists items given to new players. Example: Base.Axe,Base.Bag_BigHikingBag
	SpawnItems string `json:"SpawnItems"`

	// NoFire disables all forms of fire except campfires
	// +kubebuilder:default=false
	NoFire bool `json:"NoFire"`

	// AnnounceDeath broadcasts player deaths
	// +kubebuilder:default=false
	AnnounceDeath bool `json:"AnnounceDeath"`

	// MinutesPerPage is reading time per book page
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=60.00
	// +kubebuilder:default=1.00
	MinutesPerPage float32 `json:"MinutesPerPage"`

	// AllowDestructionBySledgehammer enables sledgehammer destruction
	// +kubebuilder:default=true
	AllowDestructionBySledgehammer bool `json:"AllowDestructionBySledgehammer"`

	// SledgehammerOnlyInSafehouse restricts sledgehammer use to safehouses
	// +kubebuilder:default=false
	SledgehammerOnlyInSafehouse bool `json:"SledgehammerOnlyInSafehouse"`

	// SleepAllowed enables sleeping
	// +kubebuilder:default=false
	SleepAllowed bool `json:"SleepAllowed"`

	// SleepNeeded requires sleeping. Ignored if SleepAllowed=false
	// +kubebuilder:default=false
	SleepNeeded bool `json:"SleepNeeded"`

	// KnockedDownAllowed enables knock downs
	// +kubebuilder:default=true
	KnockedDownAllowed bool `json:"KnockedDownAllowed"`

	// SneakModeHideFromOtherPlayers enables sneaking from players
	// +kubebuilder:default=true
	SneakModeHideFromOtherPlayers bool `json:"SneakModeHideFromOtherPlayers"`

	// SpeedLimit caps movement speed
	// +kubebuilder:validation:Minimum=10.00
	// +kubebuilder:validation:Maximum=150.00
	// +kubebuilder:default=70.00
	SpeedLimit float32 `json:"SpeedLimit"`

	// PlayerRespawnWithSelf enables respawning at death location
	// +kubebuilder:default=false
	PlayerRespawnWithSelf bool `json:"PlayerRespawnWithSelf"`

	// PlayerRespawnWithOther enables respawning at other players
	// +kubebuilder:default=false
	PlayerRespawnWithOther bool `json:"PlayerRespawnWithOther"`

	// FastForwardMultiplier affects sleep time passage
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=100.00
	// +kubebuilder:default=40.00
	FastForwardMultiplier float32 `json:"FastForwardMultiplier"`
}

// PVP fine-tunes player versus player combat with safety systems,
// damage modifiers, and combat mechanics
type PVP struct {
	// SafetySystem enables PVP safety system. When false, players can hurt each other anytime if PVP enabled.
	// +kubebuilder:default=true
	SafetySystem bool `json:"SafetySystem"`

	// ShowSafety shows safety status with skull icon
	// +kubebuilder:default=true
	ShowSafety bool `json:"ShowSafety"`

	// SafetyToggleTimer is delay for toggling safety
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=2
	SafetyToggleTimer int32 `json:"SafetyToggleTimer"`

	// SafetyCooldownTimer is cooldown between safety toggles
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=3
	SafetyCooldownTimer int32 `json:"SafetyCooldownTimer"`

	// PVPMeleeDamageModifier affects melee damage
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=500.00
	// +kubebuilder:default=30.00
	PVPMeleeDamageModifier float32 `json:"PVPMeleeDamageModifier"`

	// PVPFirearmDamageModifier affects firearm damage
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=500.00
	// +kubebuilder:default=50.00
	PVPFirearmDamageModifier float32 `json:"PVPFirearmDamageModifier"`

	// PVPMeleeWhileHitReaction enables hit reactions
	// +kubebuilder:default=false
	PVPMeleeWhileHitReaction bool `json:"PVPMeleeWhileHitReaction"`
}

// Safehouse manages player-claimed safe areas including access permissions,
// allowed activities within safehouses, and claim requirements
type Safehouse struct {
	// PlayerSafehouse enables player safehouses
	// +kubebuilder:default=false
	PlayerSafehouse bool `json:"PlayerSafehouse"`

	// AdminSafehouse enables admin safehouses
	// +kubebuilder:default=false
	AdminSafehouse bool `json:"AdminSafehouse"`

	// SafehouseAllowTrepass allows entering others' safehouses
	// +kubebuilder:default=true
	SafehouseAllowTrepass bool `json:"SafehouseAllowTrepass"`

	// SafehouseAllowFire allows fire in safehouses
	// +kubebuilder:default=true
	SafehouseAllowFire bool `json:"SafehouseAllowFire"`

	// SafehouseAllowLoot allows looting in safehouses
	// +kubebuilder:default=true
	SafehouseAllowLoot bool `json:"SafehouseAllowLoot"`

	// SafehouseAllowRespawn allows respawning in safehouses
	// +kubebuilder:default=false
	SafehouseAllowRespawn bool `json:"SafehouseAllowRespawn"`

	// SafehouseDaySurvivedToClaim is days before claiming
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	SafehouseDaySurvivedToClaim int32 `json:"SafehouseDaySurvivedToClaim"`

	// SafeHouseRemovalTime is hours before removal when not visited
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=144
	SafeHouseRemovalTime int32 `json:"SafeHouseRemovalTime"`

	// SafehouseAllowNonResidential allows non-residential safehouses
	// +kubebuilder:default=false
	SafehouseAllowNonResidential bool `json:"SafehouseAllowNonResidential"`

	// DisableSafehouseWhenPlayerConnected disables when owner online
	// +kubebuilder:default=false
	DisableSafehouseWhenPlayerConnected bool `json:"DisableSafehouseWhenPlayerConnected"`
}

// Faction configures the player group system including creation requirements
// and the number of members needed for faction tags to appear
type Faction struct {
	// Faction enables faction system
	// +kubebuilder:default=true
	Faction bool `json:"Faction"`

	// FactionDaySurvivedToCreate is days before creation
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	FactionDaySurvivedToCreate int32 `json:"FactionDaySurvivedToCreate"`

	// FactionPlayersRequiredForTag is players needed for tag
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=1
	FactionPlayersRequiredForTag int32 `json:"FactionPlayersRequiredForTag"`
}

// Loot manages the respawning and limitations of items in containers,
// including timing, quantity restrictions, and cleanup behavior
type Loot struct {
	// HoursForLootRespawn is hours before loot respawns. Container must be looted once.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	HoursForLootRespawn int32 `json:"HoursForLootRespawn"`

	// MaxItemsForLootRespawn is max items per respawn
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=4
	MaxItemsForLootRespawn int32 `json:"MaxItemsForLootRespawn"`

	// ConstructionPreventsLootRespawn prevents respawn near construction
	// +kubebuilder:default=true
	ConstructionPreventsLootRespawn bool `json:"ConstructionPreventsLootRespawn"`

	// ItemNumbersLimitPerContainer caps items per container. Includes small items like nails.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=9000
	// +kubebuilder:default=0
	ItemNumbersLimitPerContainer int32 `json:"ItemNumbersLimitPerContainer"`

	// TrashDeleteAll enables complete trash deletion
	// +kubebuilder:default=false
	TrashDeleteAll bool `json:"TrashDeleteAll"`
}

// AntiCheat configures the 24 different protection types and their thresholds
type AntiCheat struct {
	// DoLuaChecksum enables kicking clients with mismatched game files
	// +kubebuilder:default=true
	DoLuaChecksum bool `json:"DoLuaChecksum"`

	// KickFastPlayers enables kicking speed hackers. May be buggy - use with caution.
	// +kubebuilder:default=false
	KickFastPlayers bool `json:"KickFastPlayers"`

	// AntiCheatProtectionType1-24 enable different protections
	// +kubebuilder:default=true
	AntiCheatProtectionType1 bool `json:"AntiCheatProtectionType1"`
	// +kubebuilder:default=true
	AntiCheatProtectionType2 bool `json:"AntiCheatProtectionType2"`
	// +kubebuilder:default=true
	AntiCheatProtectionType3 bool `json:"AntiCheatProtectionType3"`
	// +kubebuilder:default=true
	AntiCheatProtectionType4 bool `json:"AntiCheatProtectionType4"`
	// +kubebuilder:default=true
	AntiCheatProtectionType5 bool `json:"AntiCheatProtectionType5"`
	// +kubebuilder:default=true
	AntiCheatProtectionType6 bool `json:"AntiCheatProtectionType6"`
	// +kubebuilder:default=true
	AntiCheatProtectionType7 bool `json:"AntiCheatProtectionType7"`
	// +kubebuilder:default=true
	AntiCheatProtectionType8 bool `json:"AntiCheatProtectionType8"`
	// +kubebuilder:default=true
	AntiCheatProtectionType9 bool `json:"AntiCheatProtectionType9"`
	// +kubebuilder:default=true
	AntiCheatProtectionType10 bool `json:"AntiCheatProtectionType10"`
	// +kubebuilder:default=true
	AntiCheatProtectionType11 bool `json:"AntiCheatProtectionType11"`
	// +kubebuilder:default=true
	AntiCheatProtectionType12 bool `json:"AntiCheatProtectionType12"`
	// +kubebuilder:default=true
	AntiCheatProtectionType13 bool `json:"AntiCheatProtectionType13"`
	// +kubebuilder:default=true
	AntiCheatProtectionType14 bool `json:"AntiCheatProtectionType14"`
	// +kubebuilder:default=true
	AntiCheatProtectionType15 bool `json:"AntiCheatProtectionType15"`
	// +kubebuilder:default=true
	AntiCheatProtectionType16 bool `json:"AntiCheatProtectionType16"`
	// +kubebuilder:default=true
	AntiCheatProtectionType17 bool `json:"AntiCheatProtectionType17"`
	// +kubebuilder:default=true
	AntiCheatProtectionType18 bool `json:"AntiCheatProtectionType18"`
	// +kubebuilder:default=true
	AntiCheatProtectionType19 bool `json:"AntiCheatProtectionType19"`
	// +kubebuilder:default=true
	AntiCheatProtectionType20 bool `json:"AntiCheatProtectionType20"`
	// +kubebuilder:default=true
	AntiCheatProtectionType21 bool `json:"AntiCheatProtectionType21"`
	// +kubebuilder:default=true
	AntiCheatProtectionType22 bool `json:"AntiCheatProtectionType22"`
	// +kubebuilder:default=true
	AntiCheatProtectionType23 bool `json:"AntiCheatProtectionType23"`
	// +kubebuilder:default=true
	AntiCheatProtectionType24 bool `json:"AntiCheatProtectionType24"`

	// Protection type threshold multipliers
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=3.00
	AntiCheatProtectionType2ThresholdMultiplier float32 `json:"AntiCheatProtectionType2ThresholdMultiplier"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	AntiCheatProtectionType3ThresholdMultiplier float32 `json:"AntiCheatProtectionType3ThresholdMultiplier"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	AntiCheatProtectionType4ThresholdMultiplier float32 `json:"AntiCheatProtectionType4ThresholdMultiplier"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	AntiCheatProtectionType9ThresholdMultiplier float32 `json:"AntiCheatProtectionType9ThresholdMultiplier"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	AntiCheatProtectionType15ThresholdMultiplier float32 `json:"AntiCheatProtectionType15ThresholdMultiplier"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	AntiCheatProtectionType20ThresholdMultiplier float32 `json:"AntiCheatProtectionType20ThresholdMultiplier"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	AntiCheatProtectionType22ThresholdMultiplier float32 `json:"AntiCheatProtectionType22ThresholdMultiplier"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=6.00
	AntiCheatProtectionType24ThresholdMultiplier float32 `json:"AntiCheatProtectionType24ThresholdMultiplier"`
}
