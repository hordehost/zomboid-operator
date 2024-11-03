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

	// Resources defines the compute resources required by the Zomboid server.
	Resources corev1.ResourceRequirements `json:"resources"`

	// Storage defines the persistent storage configuration for the Zomboid server.
	Storage Storage `json:"storage"`

	// Administrator defines the admin user credentials for the Zomboid server.
	Administrator Administrator `json:"administrator"`

	// Identity contains settings about how the server is identified and accessed
	// +optional
	Identity Identity `json:"identity"`

	// Player contains player management settings
	// +optional
	Player Player `json:"player"`

	// Map contains map configuration settings
	// +optional
	Map Map `json:"map"`

	// Mods contains mod configuration settings using the classic format of parallel mod/workshop lists.
	// This is the traditional way to specify mods in the server.ini but is less structured. Consider using WorkshopMods instead.
	// +optional
	Mods Mods `json:"mods"`

	// WorkshopMods contains Steam Workshop mods in a structured format.
	// This is the recommended way to specify mods for the zomboid-operator, as it provides better organization and validation.
	// +optional
	WorkshopMods []WorkshopMod `json:"workshopMods"`

	// Backup contains backup-related server settings
	// +optional
	Backup Backup `json:"backup"`

	// Logging contains logging configuration settings
	// +optional
	Logging Logging `json:"logging"`

	// Moderation contains admin and moderation settings
	// +optional
	Moderation Moderation `json:"moderation"`

	// Steam contains Steam-specific settings and anti-cheat
	// +optional
	Steam Steam `json:"steam"`

	// Discord contains Discord integration settings
	// +optional
	Discord Discord `json:"discord"`

	// Communication contains chat and VOIP settings
	// +optional
	Communication Communication `json:"communication"`

	// Gameplay contains general gameplay rules and settings
	// +optional
	Gameplay Gameplay `json:"gameplay"`

	// PVP contains PVP-specific settings
	// +optional
	PVP PVP `json:"pvp"`

	// Loot contains loot-related settings
	// +optional
	Loot Loot `json:"loot"`

	// Safehouse contains safehouse-related settings
	// +optional
	Safehouse Safehouse `json:"safehouse"`

	// Faction contains faction-related settings
	// +optional
	Faction Faction `json:"faction"`

	// AntiCheat configures the anti-cheat protection system
	// +optional
	AntiCheat AntiCheat `json:"antiCheat"`
}

// Storage defines the persistent storage configuration for the Zomboid server.
type Storage struct {
	// StorageClassName is the name of the storage class to use for the PVC, if not set, the default storage class for the cluster will be used.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// Request specifies the amount of storage requested
	// +kubebuilder:validation:Required
	Request resource.Quantity `json:"request"`
}

// Administrator defines the credentials for the admin user.
type Administrator struct {
	// Username is the admin username.
	Username string `json:"username"`

	// Password is a reference to a secret key containing the admin password.
	Password corev1.SecretKeySelector `json:"password"`
}

// ZomboidServerStatus defines the observed state of ZomboidServer.
type ZomboidServerStatus struct {
	// Ready indicates whether the server is ready to accept players
	Ready bool `json:"ready"`

	// Conditions represent the latest available observations of the ZomboidServer's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
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

	ReasonInfrastructureReady = "InfrastructureReady"
	ReasonMissingPVC          = "MissingPVC"
	ReasonMissingDeployment   = "MissingDeployment"
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
