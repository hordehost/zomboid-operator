package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZomboidBackupPlanSpec defines the desired state of ZomboidBackupPlan.
type ZomboidBackupPlanSpec struct {
	// Server references the ZomboidServer whose backups should be copied
	// +kubebuilder:validation:Required
	Server corev1.LocalObjectReference `json:"server"`

	// Destination references the BackupDestination to copy backups to
	// +kubebuilder:validation:Required
	Destination corev1.LocalObjectReference `json:"destination"`

	// Schedule specifies when backups should occur in cron format
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^(\d+|\*)(/\d+)?(\s+(\d+|\*)(/\d+)?){4}$`
	Schedule string `json:"schedule"`
}

// ZomboidBackupPlanStatus defines the observed state of ZomboidBackupPlan.
type ZomboidBackupPlanStatus struct {
	// LastBackupTime is the timestamp of when we last successfully backed up to this destination
	// +optional
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`

	// Conditions represent the latest available observations of the ZomboidBackupPlan's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ZomboidBackupPlan is the Schema for the zomboidbackupplans API.
type ZomboidBackupPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZomboidBackupPlanSpec   `json:"spec,omitempty"`
	Status ZomboidBackupPlanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ZomboidBackupPlanList contains a list of ZomboidBackupPlan.
type ZomboidBackupPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZomboidBackupPlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZomboidBackupPlan{}, &ZomboidBackupPlanList{})
}
