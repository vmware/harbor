package provider

const (
	// DriverStatusHealthy represents the healthy status
	DriverStatusHealthy = "Healthy"

	// DriverStatusUnHealthy represents the unhealthy status
	DriverStatusUnHealthy = "Unhealthy"
)

// Driver defines the capabilities one distribution provider should have.
// Includes:
//   Self descriptor
//   Health checking
//   Preheat related : Preheat means transfer the preheating image to the network of distribution provider in advance.
type Driver interface {
	// Self returns the metadata of the driver.
	// The metadata includes: name, icon(optional), maintainers(optional), version and source repo.
	Self() *Metadata

	// Try to get the health status of the driver.
	// If succeed, a non nil status object will be returned;
	// otherwise, a non nil error will be set.
	GetHealth() (*DriverStatus, error)

	// Preheat the specified image
	// If succeed, a non nil result object with preheating task id will be returned;
	// otherwise, a non nil error will be set.
	Preheat(preheatingImage *PreheatImage) (*PreheatingStatus, error)

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
	Maintainers []string `json:"maintainers,omitempty"`
	Version     string   `json:"version"`
	Source      string   `json:"source,omitempty"`
}

// DriverStatus keeps the health status of driver.
type DriverStatus struct {
	Status string `json:"status"`
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
