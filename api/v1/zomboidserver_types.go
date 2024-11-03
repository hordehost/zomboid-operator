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
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

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
