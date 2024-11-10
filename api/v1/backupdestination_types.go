package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupDestinationSpec defines the desired state of BackupDestination.
type BackupDestinationSpec struct {
	Dropbox *Dropbox `json:"dropbox,omitempty"`

	GoogleDrive *GoogleDrive `json:"googleDrive,omitempty"`

	S3 *S3 `json:"s3,omitempty"`
}

// Dropbox defines configuration for Dropbox storage.
type Dropbox struct {
	// Token for Dropbox OAuth.
	// +kubebuilder:validation:Required
	Token corev1.SecretKeySelector `json:"refreshToken"`

	// Path in Dropbox where files will be stored.
	// +optional
	Path string `json:"path,omitempty"`
}

// S3 defines configuration for S3-compatible storage providers.
type S3 struct {
	// Provider specifies which S3-compatible service to use.
	// +kubebuilder:validation:Enum=AWS;Alibaba;ArvanCloud;Ceph;ChinaMobile;Cloudflare;DigitalOcean;Dreamhost;HuaweiOBS;IBMCOS;IDrive;IONOS;Liara;Lyve;Minio;Netease;RackCorp;Scaleway;SeaweedFS;StackPath;Storj;TencentCOS;Wasabi;Other
	Provider string `json:"provider"`

	// Region to connect to.
	// +optional
	Region string `json:"region,omitempty"`

	// Endpoint for S3 API.
	// Leave blank if using AWS to use the default endpoint for the region.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// BucketName is the name of the bucket to use.
	// +kubebuilder:validation:Required
	BucketName string `json:"bucketName"`

	// Path within the bucket.
	// +optional
	Path string `json:"path,omitempty"`

	// AccessKeyID for authentication.
	// +optional
	AccessKeyID *corev1.SecretKeySelector `json:"accessKeyId,omitempty"`

	// SecretAccessKey for authentication.
	// +optional
	SecretAccessKey *corev1.SecretKeySelector `json:"secretAccessKey,omitempty"`

	// StorageClass to use when storing objects.
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

	// ServerSideEncryption algorithm used when storing objects.
	// +optional
	ServerSideEncryption string `json:"serverSideEncryption,omitempty"`
}

// GoogleDrive defines configuration for Google Drive storage.
type GoogleDrive struct {
	// Token for Google Drive OAuth.
	// +kubebuilder:validation:Required
	Token corev1.SecretKeySelector `json:"token"`

	// Path in Google Drive where files will be stored.
	// +optional
	Path string `json:"path,omitempty"`

	// RootFolderID is the ID of the root folder.
	// Leave blank normally.
	// +optional
	RootFolderID string `json:"rootFolderId,omitempty"`

	// TeamDriveID is the ID of the Shared Drive (Team Drive).
	// +optional
	TeamDriveID string `json:"teamDriveId,omitempty"`
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
