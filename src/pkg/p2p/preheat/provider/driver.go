package provider

const (
	// DriverStatusHealthy represents the healthy status
	DriverStatusHealthy = "Healthy"

	// DriverStatusUnHealthy represents the unhealthy status
	DriverStatusUnHealthy = "Unhealthy"

	// ArtifactTypeImage represents the image
	ArtifactTypeImage = "image"
)

// Driver defines the capabilities one distribution provider should have.
// Includes:
//   Self descriptor
//   Health checking
//   Preheat related : Preheat means transfer the preheating artifact to the network of distribution provider in advance.
type Driver interface {
	// Self returns the metadata of the driver
	Self() *Metadata

	// Try to get the health status of the driver.
	// If succeed, a non nil status object will be returned;
	// otherwise, a non nil error will be set.
	GetHealth() (*DriverStatus, error)

	// Preheat the specified artifact
	// If succeed, a non nil result object with preheating task id will be returned;
	// otherwise, a non nil error will be set.
	Preheat(artifact *Artifact) (*PreheatingStatus, error)

	// Check the progress of the preheating process.
	// If succeed, a non nil status object with preheating status will be returned;
	// otherwise, a non nil error will be set.
	CheckProgress(taskID string) (*PreheatingStatus, error)
}

// Metadata contains the basic information of the provider.
type Metadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon,omitempty"`
	Maintainers []string `json:"maintainers"`
	Version     string   `json:"version"`
	Source      string   `json:"source,omitempty"`
	AuthMode    string   `json:"auth_mode"`
}

// DriverStatus keeps the health status of driver.
type DriverStatus struct {
	Status string `json:"status"`
}

// Artifact contains related information which can help providers to get/pull the artifacts.
type Artifact struct {
	// The artifact content type, only support 'image' now
	Type string `json:"type"`

	// The access URL of the preheating artifact
	URL string `json:"url"`

	// The headers which will be sent to the above URL of preheating artifact
	Headers map[string]interface{} `json:"headers"`

	// The artifact name
	Name string `json:"name,omitempty"`

	// The tag
	Tag string `json:"tag,omitempty"`
}

// PreheatingStatus contains the related results/status of the preheating operation
// from the provider.
type PreheatingStatus struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	StartTime  string `json:"start_time"`
	FinishTime string `json:"finish_time"`
}
