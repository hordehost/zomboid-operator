package v1

import "k8s.io/utils/ptr"

type ZomboidSettings struct {
	// Identity contains settings about how the server is identified and accessed
	// +optional
	Identity Identity `json:"identity,omitempty"`

	// Player contains player management settings
	// +optional
	Player Player `json:"player,omitempty"`

	// Map contains map configuration settings
	// +optional
	Map Map `json:"map,omitempty"`

	// Mods contains mod configuration settings using the classic format of parallel mod/workshop lists.
	// This is the traditional way to specify mods in the server.ini but is less structured. Consider using WorkshopMods instead.
	// +optional
	Mods Mods `json:"mods,omitempty"`

	// WorkshopMods contains Steam Workshop mods in a structured format.
	// This is the recommended way to specify mods for the zomboid-operator, as it provides better organization and validation.
	// +optional
	WorkshopMods []WorkshopMod `json:"workshopMods,omitempty"`

	// Backup contains backup-related server settings
	// +optional
	Backup Backup `json:"backup,omitempty"`

	// Logging contains logging configuration settings
	// +optional
	Logging Logging `json:"logging,omitempty"`

	// Moderation contains admin and moderation settings
	// +optional
	Moderation Moderation `json:"moderation,omitempty"`

	// Steam contains Steam-specific settings and anti-cheat
	// +optional
	Steam Steam `json:"steam,omitempty"`

	// Discord contains Discord integration settings
	// +optional
	Discord Discord `json:"discord,omitempty"`

	// Communication contains chat and VOIP settings
	// +optional
	Communication Communication `json:"communication,omitempty"`

	// Gameplay contains general gameplay rules and settings
	// +optional
	Gameplay Gameplay `json:"gameplay,omitempty"`

	// PVP contains PVP-specific settings
	// +optional
	PVP PVP `json:"pvp,omitempty"`

	// Loot contains loot-related settings
	// +optional
	Loot Loot `json:"loot,omitempty"`

	// Safehouse contains safehouse-related settings
	// +optional
	Safehouse Safehouse `json:"safehouse,omitempty"`

	// Faction contains faction-related settings
	// +optional
	Faction Faction `json:"faction,omitempty"`

	// AntiCheat configures the anti-cheat protection system
	// +optional
	AntiCheat AntiCheat `json:"antiCheat,omitempty"`
}

var DefaultZomboidSettings = ZomboidSettings{
	Identity:      DefaultIdentity,
	Player:        DefaultPlayer,
	Map:           DefaultMap,
	Mods:          DefaultMods,
	WorkshopMods:  []WorkshopMod{},
	Backup:        DefaultBackup,
	Logging:       DefaultLogging,
	Moderation:    DefaultModeration,
	Steam:         DefaultSteam,
	Discord:       DefaultDiscord,
	Communication: DefaultCommunication,
	Gameplay:      DefaultGameplay,
	PVP:           DefaultPVP,
	Safehouse:     DefaultSafehouse,
	Faction:       DefaultFaction,
	Loot:          DefaultLoot,
	AntiCheat:     DefaultAntiCheat,
}

// Identity controls how your server appears in server browsers and handles player authentication,
// including server naming, visibility, password protection, and character persistence across server resets
type Identity struct {
	// Public determines if server is visible in in-game browser. Note: Steam-enabled servers are always visible in Steam browser.
	// +kubebuilder:default=false
	// +optional
	Public *bool `json:"Public,omitempty"`

	// PublicName is the server name shown in browsers
	// +kubebuilder:default="My PZ Server"
	// +optional
	PublicName *string `json:"PublicName,omitempty"`

	// PublicDescription is the server description shown in browsers. Use \n for newlines.
	// +optional
	PublicDescription *string `json:"PublicDescription,omitempty"`

	// ResetID determines if server has undergone soft-reset. If this number doesn't match client, client must create new character.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +optional
	ResetID *int32 `json:"ResetID,omitempty"`

	// ServerPlayerID identifies characters from different servers. Used with ResetID.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +optional
	ServerPlayerID *int32 `json:"ServerPlayerID,omitempty"`
}

var DefaultIdentity = Identity{
	Public:     ptr.To(false),
	PublicName: ptr.To("My PZ Server"),
}

// Player manages the core multiplayer experience including server capacity, connection requirements,
// whitelist behavior, and basic multiplayer features like co-op and username restrictions
type Player struct {
	// MaxPlayers is maximum concurrent players excluding admins. WARNING: Values above 32 may cause poor map streaming and desync.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=32
	// +optional
	MaxPlayers *int32 `json:"MaxPlayers,omitempty"`

	// PingLimit is max ping in ms before kick. Set to 100 to disable.
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=400
	// +optional
	PingLimit *int32 `json:"PingLimit,omitempty"`

	// Open allows joining without whitelist account. If false, admins must manually create accounts.
	// +kubebuilder:default=true
	// +optional
	Open *bool `json:"Open,omitempty"`

	// AutoCreateUserInWhiteList adds unknown users to whitelist. Only for Open=true servers.
	// +kubebuilder:default=false
	// +optional
	AutoCreateUserInWhiteList *bool `json:"AutoCreateUserInWhiteList,omitempty"`

	// DropOffWhiteListAfterDeath removes accounts after death. Prevents new characters after death on Open=false servers.
	// +kubebuilder:default=false
	// +optional
	DropOffWhiteListAfterDeath *bool `json:"DropOffWhiteListAfterDeath,omitempty"`

	// MaxAccountsPerUser limits accounts per Steam user. Ignored when using Host button.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	// +optional
	MaxAccountsPerUser *int32 `json:"MaxAccountsPerUser,omitempty"`

	// AllowCoop enables splitscreen/co-op play
	// +kubebuilder:default=true
	// +optional
	AllowCoop *bool `json:"AllowCoop,omitempty"`

	// AllowNonAsciiUsername enables non-ASCII characters in usernames
	// +kubebuilder:default=false
	// +optional
	AllowNonAsciiUsername *bool `json:"AllowNonAsciiUsername,omitempty"`

	// DenyLoginOnOverloadedServer prevents logins when server is overloaded
	// +kubebuilder:default=true
	// +optional
	DenyLoginOnOverloadedServer *bool `json:"DenyLoginOnOverloadedServer,omitempty"`

	// LoginQueueEnabled enables login queue
	// +kubebuilder:default=false
	// +optional
	LoginQueueEnabled *bool `json:"LoginQueueEnabled,omitempty"`

	// LoginQueueConnectTimeout is timeout for login queue in seconds
	// +kubebuilder:validation:Minimum=20
	// +kubebuilder:validation:Maximum=1200
	// +kubebuilder:default=60
	// +optional
	LoginQueueConnectTimeout *int32 `json:"LoginQueueConnectTimeout,omitempty"`
}

var DefaultPlayer = Player{
	MaxPlayers:                  ptr.To(int32(32)),
	PingLimit:                   ptr.To(int32(400)),
	Open:                        ptr.To(true),
	AutoCreateUserInWhiteList:   ptr.To(false),
	DropOffWhiteListAfterDeath:  ptr.To(false),
	MaxAccountsPerUser:          ptr.To(int32(0)),
	AllowCoop:                   ptr.To(true),
	AllowNonAsciiUsername:       ptr.To(false),
	DenyLoginOnOverloadedServer: ptr.To(true),
	LoginQueueEnabled:           ptr.To(false),
	LoginQueueConnectTimeout:    ptr.To(int32(60)),
}

// Map specifies which game world players will spawn into and explore,
// supporting both vanilla maps and custom map mods from the Steam Workshop
type Map struct {
	// Map is the folder name of the map mod. Found in Steam/steamapps/workshop/modID/mods/modName/media/maps/
	// +kubebuilder:default="Muldraugh, KY"
	// +optional
	Map *string `json:"Map,omitempty"`
}

var DefaultMap = Map{
	Map: ptr.To("Muldraugh, KY"),
}

// Mods provides two parallel lists for managing Steam Workshop content:
// one for Workshop IDs to download mods, and another for mod IDs to load them
type Mods struct {
	// WorkshopItems lists Workshop Mod IDs to download. Separate with semicolons.
	// +optional
	WorkshopItems *string `json:"WorkshopItems,omitempty"`

	// Mods lists mod loading IDs. Found in Steam/steamapps/workshop/modID/mods/modName/info.txt
	// +optional
	Mods *string `json:"Mods,omitempty"`
}

var DefaultMods = Mods{
	Mods:          ptr.To(""),
	WorkshopItems: ptr.To(""),
}

// WorkshopMod pairs a mod's loading ID with its Steam Workshop ID,
// providing a more structured way to specify mods compared to the classic parallel lists
type WorkshopMod struct {
	// ModID is the mod loading ID found in Steam/steamapps/workshop/modID/mods/modName/info.txt
	// +kubebuilder:validation:Required
	// +optional
	ModID *string `json:"modID,omitempty"`

	// WorkshopID is the Steam Workshop ID used to download the mod
	// +kubebuilder:validation:Required
	// +optional
	WorkshopID *string `json:"workshopID,omitempty"`
}

// Backup handles automatic world saving and backup creation,
// protecting against data loss from crashes, updates, or corruption
type Backup struct {
	// SaveWorldEveryMinutes is how often loaded map parts are saved. Map usually only saves when clients leave area.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	// +optional
	SaveWorldEveryMinutes *int32 `json:"SaveWorldEveryMinutes,omitempty"`

	// BackupsCount is the number of backups to keep
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=300
	// +kubebuilder:default=5
	// +optional
	BackupsCount *int32 `json:"BackupsCount,omitempty"`

	// BackupsOnStart enables backups when server starts
	// +kubebuilder:default=true
	// +optional
	BackupsOnStart *bool `json:"BackupsOnStart,omitempty"`

	// BackupsOnVersionChange enables backups on version changes
	// +kubebuilder:default=true
	// +optional
	BackupsOnVersionChange *bool `json:"BackupsOnVersionChange,omitempty"`

	// BackupsPeriod is the backup interval in minutes
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1500
	// +kubebuilder:default=0
	// +optional
	BackupsPeriod *int32 `json:"BackupsPeriod,omitempty"`
}

var DefaultBackup = Backup{
	SaveWorldEveryMinutes:  ptr.To(int32(0)),
	BackupsCount:           ptr.To(int32(5)),
	BackupsOnStart:         ptr.To(true),
	BackupsOnVersionChange: ptr.To(true),
	BackupsPeriod:          ptr.To(int32(0)),
}

// Logging configures what client actions and commands are logged to files
type Logging struct {
	// PerkLogs enables tracking player perk changes in PerkLog.txt
	// +kubebuilder:default=true
	// +optional
	PerkLogs *bool `json:"PerkLogs,omitempty"`

	// ClientCommandFilter lists commands not written to cmd.txt log
	// +kubebuilder:default="-vehicle.*;+vehicle.damageWindow;+vehicle.fixPart;+vehicle.installPart;+vehicle.uninstallPart"
	// +optional
	ClientCommandFilter *string `json:"ClientCommandFilter,omitempty"`

	// ClientActionLogs lists actions written to ClientActionLogs.txt
	// +kubebuilder:default="ISEnterVehicle;ISExitVehicle;ISTakeEngineParts;"
	// +optional
	ClientActionLogs *string `json:"ClientActionLogs,omitempty"`
}

var DefaultLogging = Logging{
	PerkLogs:            ptr.To(true),
	ClientCommandFilter: ptr.To("-vehicle.*;+vehicle.damageWindow;+vehicle.fixPart;+vehicle.installPart;+vehicle.uninstallPart"),
	ClientActionLogs:    ptr.To("ISEnterVehicle;ISExitVehicle;ISTakeEngineParts;"),
}

// Moderation customizes staff member capabilities and restrictions,
// particularly focusing on radio usage permissions for different staff roles
type Moderation struct {
	// DisableRadioStaff disables radio for staff
	// +kubebuilder:default=false
	// +optional
	DisableRadioStaff *bool `json:"DisableRadioStaff,omitempty"`

	// DisableRadioAdmin disables radio for admins
	// +kubebuilder:default=true
	// +optional
	DisableRadioAdmin *bool `json:"DisableRadioAdmin,omitempty"`

	// DisableRadioGM disables radio for GMs
	// +kubebuilder:default=true
	// +optional
	DisableRadioGM *bool `json:"DisableRadioGM,omitempty"`

	// DisableRadioOverseer disables radio for overseers
	// +kubebuilder:default=false
	// +optional
	DisableRadioOverseer *bool `json:"DisableRadioOverseer,omitempty"`

	// DisableRadioModerator disables radio for moderators
	// +kubebuilder:default=false
	// +optional
	DisableRadioModerator *bool `json:"DisableRadioModerator,omitempty"`

	// DisableRadioInvisible disables radio for invisible players
	// +kubebuilder:default=true
	// +optional
	DisableRadioInvisible *bool `json:"DisableRadioInvisible,omitempty"`

	// BanKickGlobalSound enables global sound on ban/kick
	// +kubebuilder:default=true
	// +optional
	BanKickGlobalSound *bool `json:"BanKickGlobalSound,omitempty"`
}

var DefaultModeration = Moderation{
	DisableRadioStaff:     ptr.To(false),
	DisableRadioAdmin:     ptr.To(true),
	DisableRadioGM:        ptr.To(true),
	DisableRadioOverseer:  ptr.To(false),
	DisableRadioModerator: ptr.To(false),
	DisableRadioInvisible: ptr.To(true),
	BanKickGlobalSound:    ptr.To(true),
}

// Steam manages Steam platform integration and anti-cheat measures,
// including VAC and player visibility settings
type Steam struct {
	// SteamScoreboard controls visibility of Steam names/avatars. Can be "true" (visible to everyone), "false" (visible to no one), or "admin" (visible to only admins)
	// +kubebuilder:default="true"
	// +optional
	SteamScoreboard *string `json:"SteamScoreboard,omitempty"`
}

var DefaultSteam = Steam{
	SteamScoreboard: ptr.To("true"),
}

// Discord enables and configures integration with Discord,
// allowing server chat to be bridged with a Discord channel
type Discord struct {
	// DiscordEnable enables Discord chat integration
	// +kubebuilder:default=false
	// +optional
	DiscordEnable *bool `json:"DiscordEnable,omitempty"`

	// DiscordToken is the Discord bot token
	// +optional
	DiscordToken *string `json:"DiscordToken,omitempty"`

	// DiscordChannel is the Discord channel name. Try channel ID if having difficulties.
	// +optional
	DiscordChannel *string `json:"DiscordChannel,omitempty"`

	// DiscordChannelID is the Discord channel ID. Use if having difficulties with channel name.
	// +optional
	DiscordChannelID *string `json:"DiscordChannelID,omitempty"`
}

var DefaultDiscord = Discord{
	DiscordEnable: ptr.To(false),
}

// Communication manages all player interaction features including global chat,
// chat streams, welcome messages, and VOIP with distance-based audio
type Communication struct {
	// GlobalChat enables global chat
	// +kubebuilder:default=true
	// +optional
	GlobalChat *bool `json:"GlobalChat,omitempty"`

	// ChatStreams lists available chat streams
	// +kubebuilder:default="s,r,a,w,y,sh,f,all"
	// +optional
	ChatStreams *string `json:"ChatStreams,omitempty"`

	// ServerWelcomeMessage is shown to players on login. Use <LINE> for newlines and <RGB:r,g,b> for colors.
	// +kubebuilder:default="Welcome to Project Zomboid Multiplayer! <LINE> <LINE> To interact with the Chat panel: press Tab, T, or Enter. <LINE> <LINE> The Tab key will change the target stream of the message. <LINE> <LINE> Global Streams: /all <LINE> Local Streams: /say, /yell <LINE> Special Steams: /whisper, /safehouse, /faction. <LINE> <LINE> Press the Up arrow to cycle through your message history. Click the Gear icon to customize chat. <LINE> <LINE> Happy surviving!"
	// +optional
	ServerWelcomeMessage *string `json:"ServerWelcomeMessage,omitempty"`

	// VoiceEnable enables VOIP
	// +kubebuilder:default=true
	// +optional
	VoiceEnable *bool `json:"VoiceEnable,omitempty"`

	// VoiceMinDistance is minimum VOIP audible distance
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=100000.00
	// +kubebuilder:default=10.00
	// +optional
	VoiceMinDistance *float32 `json:"VoiceMinDistance,omitempty"`

	// VoiceMaxDistance is maximum VOIP audible distance
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=100000.00
	// +kubebuilder:default=100.00
	// +optional
	VoiceMaxDistance *float32 `json:"VoiceMaxDistance,omitempty"`

	// Voice3D enables directional VOIP audio
	// +kubebuilder:default=true
	// +optional
	Voice3D *bool `json:"Voice3D,omitempty"`
}

var DefaultCommunication = Communication{
	GlobalChat:           ptr.To(true),
	ChatStreams:          ptr.To("s,r,a,w,y,sh,f,all"),
	ServerWelcomeMessage: ptr.To("Welcome to Project Zomboid Multiplayer! <LINE> <LINE> To interact with the Chat panel: press Tab, T, or Enter. <LINE> <LINE> The Tab key will change the target stream of the message. <LINE> <LINE> Global Streams: /all <LINE> Local Streams: /say, /yell <LINE> Special Steams: /whisper, /safehouse, /faction. <LINE> <LINE> Press the Up arrow to cycle through your message history. Click the Gear icon to customize chat. <LINE> <LINE> Happy surviving!"),
	VoiceEnable:          ptr.To(true),
	VoiceMinDistance:     ptr.To(float32(10.00)),
	VoiceMaxDistance:     ptr.To(float32(100.00)),
	Voice3D:              ptr.To(true),
}

// Gameplay controls fundamental aspects of the player experience including
// PvP, time progression, player visibility, spawning, movement, and sleep mechanics
type Gameplay struct {
	// PauseEmpty pauses time when no players online
	// +kubebuilder:default=true
	// +optional
	PauseEmpty *bool `json:"PauseEmpty,omitempty"`

	// DisplayUserName shows player names
	// +kubebuilder:default=true
	// +optional
	DisplayUserName *bool `json:"DisplayUserName,omitempty"`

	// ShowFirstAndLastName shows full player names
	// +kubebuilder:default=false
	// +optional
	ShowFirstAndLastName *bool `json:"ShowFirstAndLastName,omitempty"`

	// SpawnPoint forces spawn location (x,y,z). Find coordinates at map.projectzomboid.com. Ignored when 0,0,0.
	// +kubebuilder:default="0,0,0"
	// +optional
	SpawnPoint *string `json:"SpawnPoint,omitempty"`

	// SpawnItems lists items given to new players. Example: Base.Axe,Base.Bag_BigHikingBag
	// +optional
	SpawnItems *string `json:"SpawnItems,omitempty"`

	// NoFire disables all forms of fire except campfires
	// +kubebuilder:default=false
	// +optional
	NoFire *bool `json:"NoFire,omitempty"`

	// AnnounceDeath broadcasts player deaths
	// +kubebuilder:default=false
	// +optional
	AnnounceDeath *bool `json:"AnnounceDeath,omitempty"`

	// MinutesPerPage is reading time per book page
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=60.00
	// +kubebuilder:default=1.00
	// +optional
	MinutesPerPage *float32 `json:"MinutesPerPage,omitempty"`

	// AllowDestructionBySledgehammer enables sledgehammer destruction
	// +kubebuilder:default=true
	// +optional
	AllowDestructionBySledgehammer *bool `json:"AllowDestructionBySledgehammer,omitempty"`

	// SledgehammerOnlyInSafehouse restricts sledgehammer use to safehouses
	// +kubebuilder:default=false
	// +optional
	SledgehammerOnlyInSafehouse *bool `json:"SledgehammerOnlyInSafehouse,omitempty"`

	// SleepAllowed enables sleeping
	// +kubebuilder:default=false
	// +optional
	SleepAllowed *bool `json:"SleepAllowed,omitempty"`

	// SleepNeeded requires sleeping. Ignored if SleepAllowed=false
	// +kubebuilder:default=false
	// +optional
	SleepNeeded *bool `json:"SleepNeeded,omitempty"`

	// KnockedDownAllowed enables knock downs
	// +kubebuilder:default=true
	// +optional
	KnockedDownAllowed *bool `json:"KnockedDownAllowed,omitempty"`

	// SneakModeHideFromOtherPlayers enables sneaking from players
	// +kubebuilder:default=true
	// +optional
	SneakModeHideFromOtherPlayers *bool `json:"SneakModeHideFromOtherPlayers,omitempty"`

	// SpeedLimit caps movement speed
	// +kubebuilder:validation:Minimum=10.00
	// +kubebuilder:validation:Maximum=150.00
	// +kubebuilder:default=70.00
	// +optional
	SpeedLimit *float32 `json:"SpeedLimit,omitempty"`

	// PlayerRespawnWithSelf enables respawning at death location
	// +kubebuilder:default=false
	// +optional
	PlayerRespawnWithSelf *bool `json:"PlayerRespawnWithSelf,omitempty"`

	// PlayerRespawnWithOther enables respawning at other players
	// +kubebuilder:default=false
	// +optional
	PlayerRespawnWithOther *bool `json:"PlayerRespawnWithOther,omitempty"`

	// FastForwardMultiplier affects sleep time passage
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=100.00
	// +kubebuilder:default=40.00
	// +optional
	FastForwardMultiplier *float32 `json:"FastForwardMultiplier,omitempty"`
}

var DefaultGameplay = Gameplay{
	PauseEmpty:                     ptr.To(true),
	DisplayUserName:                ptr.To(true),
	ShowFirstAndLastName:           ptr.To(false),
	SpawnPoint:                     ptr.To("0,0,0"),
	NoFire:                         ptr.To(false),
	AnnounceDeath:                  ptr.To(false),
	MinutesPerPage:                 ptr.To(float32(1.00)),
	AllowDestructionBySledgehammer: ptr.To(true),
	SledgehammerOnlyInSafehouse:    ptr.To(false),
	SleepAllowed:                   ptr.To(false),
	SleepNeeded:                    ptr.To(false),
	KnockedDownAllowed:             ptr.To(true),
	SneakModeHideFromOtherPlayers:  ptr.To(true),
	SpeedLimit:                     ptr.To(float32(70.00)),
	PlayerRespawnWithSelf:          ptr.To(false),
	PlayerRespawnWithOther:         ptr.To(false),
	FastForwardMultiplier:          ptr.To(float32(40.00)),
}

// PVP fine-tunes player versus player combat with safety systems,
// damage modifiers, and combat mechanics
type PVP struct {
	// PVP enables player vs player combat
	// +kubebuilder:default=true
	// +optional
	PVP *bool `json:"PVP,omitempty"`

	// SafetySystem enables PVP safety system. When false, players can hurt each other anytime if PVP enabled.
	// +kubebuilder:default=true
	// +optional
	SafetySystem *bool `json:"SafetySystem,omitempty"`

	// ShowSafety shows safety status with skull icon
	// +kubebuilder:default=true
	// +optional
	ShowSafety *bool `json:"ShowSafety,omitempty"`

	// SafetyToggleTimer is delay for toggling safety
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=2
	// +optional
	SafetyToggleTimer *int32 `json:"SafetyToggleTimer,omitempty"`

	// SafetyCooldownTimer is cooldown between safety toggles
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=3
	// +optional
	SafetyCooldownTimer *int32 `json:"SafetyCooldownTimer,omitempty"`

	// PVPMeleeDamageModifier affects melee damage
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=500.00
	// +kubebuilder:default=30.00
	// +optional
	PVPMeleeDamageModifier *float32 `json:"PVPMeleeDamageModifier,omitempty"`

	// PVPFirearmDamageModifier affects firearm damage
	// +kubebuilder:validation:Minimum=0.00
	// +kubebuilder:validation:Maximum=500.00
	// +kubebuilder:default=50.00
	// +optional
	PVPFirearmDamageModifier *float32 `json:"PVPFirearmDamageModifier,omitempty"`

	// PVPMeleeWhileHitReaction enables hit reactions
	// +kubebuilder:default=false
	// +optional
	PVPMeleeWhileHitReaction *bool `json:"PVPMeleeWhileHitReaction,omitempty"`
}

var DefaultPVP = PVP{
	PVP:                      ptr.To(true),
	SafetySystem:             ptr.To(true),
	ShowSafety:               ptr.To(true),
	SafetyToggleTimer:        ptr.To(int32(2)),
	SafetyCooldownTimer:      ptr.To(int32(3)),
	PVPMeleeDamageModifier:   ptr.To(float32(30.00)),
	PVPFirearmDamageModifier: ptr.To(float32(50.00)),
	PVPMeleeWhileHitReaction: ptr.To(false),
}

// Safehouse manages player-claimed safe areas including access permissions,
// allowed activities within safehouses, and claim requirements
type Safehouse struct {
	// PlayerSafehouse enables player safehouses
	// +kubebuilder:default=false
	// +optional
	PlayerSafehouse *bool `json:"PlayerSafehouse,omitempty"`

	// AdminSafehouse enables admin safehouses
	// +kubebuilder:default=false
	// +optional
	AdminSafehouse *bool `json:"AdminSafehouse,omitempty"`

	// SafehouseAllowTrepass allows entering others' safehouses
	// +kubebuilder:default=true
	// +optional
	SafehouseAllowTrepass *bool `json:"SafehouseAllowTrepass,omitempty"`

	// SafehouseAllowFire allows fire in safehouses
	// +kubebuilder:default=true
	// +optional
	SafehouseAllowFire *bool `json:"SafehouseAllowFire,omitempty"`

	// SafehouseAllowLoot allows looting in safehouses
	// +kubebuilder:default=true
	// +optional
	SafehouseAllowLoot *bool `json:"SafehouseAllowLoot,omitempty"`

	// SafehouseAllowRespawn allows respawning in safehouses
	// +kubebuilder:default=false
	// +optional
	SafehouseAllowRespawn *bool `json:"SafehouseAllowRespawn,omitempty"`

	// SafehouseDaySurvivedToClaim is days before claiming
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	// +optional
	SafehouseDaySurvivedToClaim *int32 `json:"SafehouseDaySurvivedToClaim,omitempty"`

	// SafeHouseRemovalTime is hours before removal when not visited
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=144
	// +optional
	SafeHouseRemovalTime *int32 `json:"SafeHouseRemovalTime,omitempty"`

	// SafehouseAllowNonResidential allows non-residential safehouses
	// +kubebuilder:default=false
	// +optional
	SafehouseAllowNonResidential *bool `json:"SafehouseAllowNonResidential,omitempty"`

	// DisableSafehouseWhenPlayerConnected disables when owner online
	// +kubebuilder:default=false
	// +optional
	DisableSafehouseWhenPlayerConnected *bool `json:"DisableSafehouseWhenPlayerConnected,omitempty"`
}

var DefaultSafehouse = Safehouse{
	PlayerSafehouse:                     ptr.To(false),
	AdminSafehouse:                      ptr.To(false),
	SafehouseAllowTrepass:               ptr.To(true),
	SafehouseAllowFire:                  ptr.To(true),
	SafehouseAllowLoot:                  ptr.To(true),
	SafehouseAllowRespawn:               ptr.To(false),
	SafehouseDaySurvivedToClaim:         ptr.To(int32(0)),
	SafeHouseRemovalTime:                ptr.To(int32(144)),
	SafehouseAllowNonResidential:        ptr.To(false),
	DisableSafehouseWhenPlayerConnected: ptr.To(false),
}

// Faction configures the player group system including creation requirements
// and the number of members needed for faction tags to appear
type Faction struct {
	// Faction enables faction system
	// +kubebuilder:default=true
	// +optional
	Faction *bool `json:"Faction,omitempty"`

	// FactionDaySurvivedToCreate is days before creation
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	// +optional
	FactionDaySurvivedToCreate *int32 `json:"FactionDaySurvivedToCreate,omitempty"`

	// FactionPlayersRequiredForTag is players needed for tag
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=1
	// +optional
	FactionPlayersRequiredForTag *int32 `json:"FactionPlayersRequiredForTag,omitempty"`
}

var DefaultFaction = Faction{
	Faction:                      ptr.To(true),
	FactionDaySurvivedToCreate:   ptr.To(int32(0)),
	FactionPlayersRequiredForTag: ptr.To(int32(1)),
}

// Loot manages the respawning and limitations of items in containers,
// including timing, quantity restrictions, and cleanup behavior
type Loot struct {
	// HoursForLootRespawn is hours before loot respawns. Container must be looted once.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=0
	// +optional
	HoursForLootRespawn *int32 `json:"HoursForLootRespawn,omitempty"`

	// MaxItemsForLootRespawn is max items per respawn
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=2147483647
	// +kubebuilder:default=4
	// +optional
	MaxItemsForLootRespawn *int32 `json:"MaxItemsForLootRespawn,omitempty"`

	// ConstructionPreventsLootRespawn prevents respawn near construction
	// +kubebuilder:default=true
	// +optional
	ConstructionPreventsLootRespawn *bool `json:"ConstructionPreventsLootRespawn,omitempty"`

	// ItemNumbersLimitPerContainer caps items per container. Includes small items like nails.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=9000
	// +kubebuilder:default=0
	// +optional
	ItemNumbersLimitPerContainer *int32 `json:"ItemNumbersLimitPerContainer,omitempty"`

	// TrashDeleteAll enables complete trash deletion
	// +kubebuilder:default=false
	// +optional
	TrashDeleteAll *bool `json:"TrashDeleteAll,omitempty"`
}

var DefaultLoot = Loot{
	HoursForLootRespawn:             ptr.To(int32(0)),
	MaxItemsForLootRespawn:          ptr.To(int32(4)),
	ConstructionPreventsLootRespawn: ptr.To(true),
	ItemNumbersLimitPerContainer:    ptr.To(int32(0)),
	TrashDeleteAll:                  ptr.To(false),
}

// AntiCheat configures the 24 different protection types and their thresholds
type AntiCheat struct {
	// DoLuaChecksum enables kicking clients with mismatched game files
	// +kubebuilder:default=true
	// +optional
	DoLuaChecksum *bool `json:"DoLuaChecksum,omitempty"`

	// KickFastPlayers enables kicking speed hackers. May be buggy - use with caution.
	// +kubebuilder:default=false
	// +optional
	KickFastPlayers *bool `json:"KickFastPlayers,omitempty"`

	// AntiCheatProtectionType1-24 enable different protections
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType1 *bool `json:"AntiCheatProtectionType1,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType2 *bool `json:"AntiCheatProtectionType2,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType3 *bool `json:"AntiCheatProtectionType3,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType4 *bool `json:"AntiCheatProtectionType4,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType5 *bool `json:"AntiCheatProtectionType5,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType6 *bool `json:"AntiCheatProtectionType6,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType7 *bool `json:"AntiCheatProtectionType7,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType8 *bool `json:"AntiCheatProtectionType8,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType9 *bool `json:"AntiCheatProtectionType9,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType10 *bool `json:"AntiCheatProtectionType10,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType11 *bool `json:"AntiCheatProtectionType11,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType12 *bool `json:"AntiCheatProtectionType12,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType13 *bool `json:"AntiCheatProtectionType13,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType14 *bool `json:"AntiCheatProtectionType14,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType15 *bool `json:"AntiCheatProtectionType15,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType16 *bool `json:"AntiCheatProtectionType16,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType17 *bool `json:"AntiCheatProtectionType17,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType18 *bool `json:"AntiCheatProtectionType18,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType19 *bool `json:"AntiCheatProtectionType19,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType20 *bool `json:"AntiCheatProtectionType20,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType21 *bool `json:"AntiCheatProtectionType21,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType22 *bool `json:"AntiCheatProtectionType22,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType23 *bool `json:"AntiCheatProtectionType23,omitempty"`
	// +kubebuilder:default=true
	// +optional
	AntiCheatProtectionType24 *bool `json:"AntiCheatProtectionType24,omitempty"`

	// Protection type threshold multipliers
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=3.00
	// +optional
	AntiCheatProtectionType2ThresholdMultiplier *float32 `json:"AntiCheatProtectionType2ThresholdMultiplier,omitempty"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	// +optional
	AntiCheatProtectionType3ThresholdMultiplier *float32 `json:"AntiCheatProtectionType3ThresholdMultiplier,omitempty"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	// +optional
	AntiCheatProtectionType4ThresholdMultiplier *float32 `json:"AntiCheatProtectionType4ThresholdMultiplier,omitempty"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	// +optional
	AntiCheatProtectionType9ThresholdMultiplier *float32 `json:"AntiCheatProtectionType9ThresholdMultiplier,omitempty"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	// +optional
	AntiCheatProtectionType15ThresholdMultiplier *float32 `json:"AntiCheatProtectionType15ThresholdMultiplier,omitempty"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	// +optional
	AntiCheatProtectionType20ThresholdMultiplier *float32 `json:"AntiCheatProtectionType20ThresholdMultiplier,omitempty"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=1.00
	// +optional
	AntiCheatProtectionType22ThresholdMultiplier *float32 `json:"AntiCheatProtectionType22ThresholdMultiplier,omitempty"`
	// +kubebuilder:validation:Minimum=1.00
	// +kubebuilder:validation:Maximum=10.00
	// +kubebuilder:default=6.00
	// +optional
	AntiCheatProtectionType24ThresholdMultiplier *float32 `json:"AntiCheatProtectionType24ThresholdMultiplier,omitempty"`
}

var DefaultAntiCheat = AntiCheat{
	DoLuaChecksum:                                ptr.To(true),
	KickFastPlayers:                              ptr.To(false),
	AntiCheatProtectionType1:                     ptr.To(true),
	AntiCheatProtectionType2:                     ptr.To(true),
	AntiCheatProtectionType3:                     ptr.To(true),
	AntiCheatProtectionType4:                     ptr.To(true),
	AntiCheatProtectionType5:                     ptr.To(true),
	AntiCheatProtectionType6:                     ptr.To(true),
	AntiCheatProtectionType7:                     ptr.To(true),
	AntiCheatProtectionType8:                     ptr.To(true),
	AntiCheatProtectionType9:                     ptr.To(true),
	AntiCheatProtectionType10:                    ptr.To(true),
	AntiCheatProtectionType11:                    ptr.To(true),
	AntiCheatProtectionType12:                    ptr.To(true),
	AntiCheatProtectionType13:                    ptr.To(true),
	AntiCheatProtectionType14:                    ptr.To(true),
	AntiCheatProtectionType15:                    ptr.To(true),
	AntiCheatProtectionType16:                    ptr.To(true),
	AntiCheatProtectionType17:                    ptr.To(true),
	AntiCheatProtectionType18:                    ptr.To(true),
	AntiCheatProtectionType19:                    ptr.To(true),
	AntiCheatProtectionType20:                    ptr.To(true),
	AntiCheatProtectionType21:                    ptr.To(true),
	AntiCheatProtectionType22:                    ptr.To(true),
	AntiCheatProtectionType23:                    ptr.To(true),
	AntiCheatProtectionType24:                    ptr.To(true),
	AntiCheatProtectionType2ThresholdMultiplier:  ptr.To(float32(3.00)),
	AntiCheatProtectionType3ThresholdMultiplier:  ptr.To(float32(1.00)),
	AntiCheatProtectionType4ThresholdMultiplier:  ptr.To(float32(1.00)),
	AntiCheatProtectionType9ThresholdMultiplier:  ptr.To(float32(1.00)),
	AntiCheatProtectionType15ThresholdMultiplier: ptr.To(float32(1.00)),
	AntiCheatProtectionType20ThresholdMultiplier: ptr.To(float32(1.00)),
	AntiCheatProtectionType22ThresholdMultiplier: ptr.To(float32(1.00)),
	AntiCheatProtectionType24ThresholdMultiplier: ptr.To(float32(6.00)),
}
