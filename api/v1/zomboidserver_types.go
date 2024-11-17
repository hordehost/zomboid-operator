package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZomboidServerSpec defines the desired state of ZomboidServer.
type ZomboidServerSpec struct {
	// Version is the version of the Zomboid server to run.
	Version string `json:"version"`

	// ServerPort is the port used for establishing connections to the server (UDP)
	// +optional
	// +kubebuilder:default=16261
	ServerPort *int32 `json:"serverPort,omitempty"`

	// UDPPort is the port used for game traffic (UDP)
	// +optional
	// +kubebuilder:default=16262
	UDPPort *int32 `json:"udpPort,omitempty"`

	// Resources defines the compute resources required by the Zomboid server.
	Resources corev1.ResourceRequirements `json:"resources"`

	// Storage defines the persistent storage configuration for the Zomboid server.
	Storage Storage `json:"storage"`

	// Backups defines the persistent storage configuration for the Zomboid server backups.
	Backups Backups `json:"backups,omitempty"`

	// Administrator defines the admin user credentials for the Zomboid server.
	Administrator Administrator `json:"administrator"`

	// Users is a list of users to add to the server
	// +optional
	Users []User `json:"users,omitempty"`

	// Password is required for clients to join.
	// +optional
	Password *corev1.SecretKeySelector `json:"password,omitempty"`

	// Suspended indicates whether the server should be running. If true, the server will be stopped.
	// +optional
	Suspended *bool `json:"suspended,omitempty"`

	// Settings contains the server's current settings
	// +optional
	Settings ZomboidSettings `json:"settings,omitempty"`

	// Discord contains the Discord configuration
	// +optional
	Discord *Discord `json:"discord,omitempty"`
}

// Storage defines the persistent storage configuration for the Zomboid server.
type Storage struct {
	// StorageClassName is the name of the storage class to use for the PVC, if
	// not set, the default storage class for the cluster will be used.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// Request specifies the amount of storage requested
	// +kubebuilder:validation:Required
	Request resource.Quantity `json:"request"`

	// WorkshopRequest specifies the amount of storage requested for mods
	// +optional
	WorkshopRequest *resource.Quantity `json:"workshopRequest,omitempty"`
}

type Backups struct {
	// StorageClassName is the name of the storage class to use for the backup
	// PVC. This storage class should support ReadWriteMany access.  If backups
	// are requested and this is not set, the cluster's default storage class
	// will be used.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// Request specifies the amount of storage requested for backups in-cluster
	// before they are uploaded to external storage providers
	// +optional
	Request *resource.Quantity `json:"request,omitempty"`
}

// Administrator defines the credentials for the admin user.
type Administrator struct {
	// Username is the admin username.
	Username string `json:"username"`

	// Password is a reference to a secret key containing the admin password.
	Password corev1.SecretKeySelector `json:"password"`
}

// Discord enables and configures integration with Discord,
// allowing server chat to be bridged with a Discord channel
type Discord struct {
	// DiscordToken is a reference to a secret key containing the Discord bot token
	// +optional
	DiscordToken *corev1.SecretKeySelector `json:"DiscordToken,omitempty"`

	// DiscordChannel is a reference to a secret key containing the Discord channel name
	// +optional
	DiscordChannel *corev1.SecretKeySelector `json:"DiscordChannel,omitempty"`

	// DiscordChannelID is a reference to a secret key containing the Discord channel ID
	// +optional
	DiscordChannelID *corev1.SecretKeySelector `json:"DiscordChannelID,omitempty"`
}

// ZomboidServerStatus defines the observed state of ZomboidServer.
type ZomboidServerStatus struct {
	// Ready indicates whether the server is ready to accept players
	Ready bool `json:"ready"`

	// SettingsLastObserved is the timestamp of when we last successfully read the server's settings
	// +optional
	SettingsLastObserved *metav1.Time `json:"settingsLastObserved,omitempty"`

	// Settings contains the server's current settings, if they have ever been observed
	// +optional
	Settings *ZomboidSettings `json:"settings,omitempty"`

	// Allowlist contains the server's current allowlist
	// +optional
	Allowlist []AllowlistUser `json:"allowlist,omitempty"`

	// ConnectedPlayers contains the server's current connected players
	// +optional
	ConnectedPlayers []ConnectedPlayer `json:"connectedPlayers,omitempty"`

	// Conditions represent the latest available observations of the ZomboidServer's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions," patchStrategy:"merge" patchMergeKey:"type"`
}

// User represents a user to add to the server
type User struct {
	// Username is the username of the user
	Username string `json:"username"`

	// Password is a reference to a secret key containing the user's password
	Password *corev1.SecretKeySelector `json:"password"`

	// AccessLevel is the access level of the user
	// +kubebuilder:validation:Enum=Admin;Moderator;Overseer;GM;Observer
	// +optional
	AccessLevel string `json:"accesslevel,omitempty"`

	// Banned indicates whether the user is banned from the server
	// +optional
	Banned bool `json:"banned,omitempty"`
}

// AllowlistUser represents the current state of a user on the server's allowlist
type AllowlistUser struct {
	Username       string  `json:"username"`
	ID             int     `json:"id"`
	SteamID        *string `json:"steamid,omitempty"`
	OwnerID        *string `json:"ownerid,omitempty"`
	AccessLevel    string  `json:"accesslevel"`
	DisplayName    *string `json:"displayName,omitempty"`
	Banned         bool    `json:"banned"`
	HashedPassword string  `json:"hashedPassword,omitempty"`
	LastConnection *string `json:"lastConnection,omitempty"`
}

type ConnectedPlayer struct {
	Username string `json:"username"`
}

// Condition Types
const (
	// TypeReadyForPlayers indicates whether the ZomboidServer is ready to accept players
	TypeReadyForPlayers = "ReadyForPlayers"
	// TypeInfrastructureReady indicates whether all required infrastructure components exist
	TypeInfrastructureReady = "InfrastructureReady"
)

// Condition Reasons
const (
	ReasonServerStarting = "ServerStarting"
	ReasonServerReady    = "ServerReady"

	ReasonInfrastructureReady  = "InfrastructureReady"
	ReasonMissingPVC           = "MissingPVC"
	ReasonMissingDeployment    = "MissingDeployment"
	ReasonMissingRCONService   = "MissingRCONService"
	ReasonMissingGameService   = "MissingGameService"
	ReasonMissingSQLiteService = "MissingSQLiteService"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ZomboidServer is the Schema for the zomboidservers API.
type ZomboidServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZomboidServerSpec   `json:"spec,omitempty"`
	Status ZomboidServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ZomboidServerList contains a list of ZomboidServer.
type ZomboidServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZomboidServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZomboidServer{}, &ZomboidServerList{})
}
