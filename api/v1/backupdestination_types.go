package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupDestinationSpec defines the desired state of BackupDestination.
type BackupDestinationSpec struct {
	S3 *S3 `json:"s3,omitempty"`

	Dropbox *Dropbox `json:"dropbox,omitempty"`
}

type S3 struct {
	// BucketName is the name of the remote bucket that should be used for storing backups
	// +kubebuilder:validation:Required
	BucketName string `json:"bucketName"`

	// Path is the path within the bucket where backups will be stored. Must not contain a leading slash.
	// +optional
	Path string `json:"path,omitempty"`

	// AccessKeyID is the AWS access key ID for authentication
	// +optional
	AccessKeyID *corev1.SecretKeySelector `json:"accessKeyId,omitempty"`

	// SecretAccessKey is the AWS secret access key for authentication
	// +optional
	SecretAccessKey *corev1.SecretKeySelector `json:"secretAccessKey,omitempty"`

	// IAMRoleEndpoint is the endpoint for IAM role credentials (e.g. http://169.254.169.254)
	// +optional
	IAMRoleEndpoint string `json:"iamRoleEndpoint,omitempty"`

	// Endpoint is the FQDN of the S3 storage server. Defaults to s3.amazonaws.com if not specified.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// EndpointProtocol is the protocol (http/https) to use. Defaults to https.
	// +optional
	EndpointProtocol string `json:"endpointProtocol,omitempty"`

	// EndpointInsecure disables SSL certificate verification when true
	// +optional
	EndpointInsecure bool `json:"endpointInsecure,omitempty"`

	// EndpointCACert is a PEM encoded CA certificate for validating self-signed certificates
	// +optional
	EndpointCACert string `json:"endpointCACert,omitempty"`

	// StorageClass changes the S3 storage class header. Defaults to STANDARD.
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

	// PartSize changes the S3 multipart upload part size in MB. Defaults to 16.
	// +optional
	// +kubebuilder:validation:Minimum=1
	PartSize *int32 `json:"partSize,omitempty"`
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
