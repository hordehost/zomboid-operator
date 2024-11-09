package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupDestinationSpec defines the desired state of BackupDestination.
type BackupDestinationSpec struct {
	Dropbox *Dropbox `json:"dropbox,omitempty"`
}

// Dropbox contains configuration for backing up to Dropbox
type Dropbox struct {
	// RefreshToken is the OAuth refresh token for Dropbox access
	RefreshToken corev1.SecretKeySelector `json:"refreshToken"`

	// RemotePath is the path in Dropbox where backups will be stored
	// If not specified, defaults to /<namespace>/zomboid/<server-name>
	// +optional
	RemotePath string `json:"remotePath,omitempty"`
}

// BackupDestinationStatus defines the observed state of BackupDestination.
type BackupDestinationStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// BackupDestination is the Schema for the backupdestinations API.
type BackupDestination struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupDestinationSpec   `json:"spec,omitempty"`
	Status BackupDestinationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupDestinationList contains a list of BackupDestination.
type BackupDestinationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupDestination `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupDestination{}, &BackupDestinationList{})
}
