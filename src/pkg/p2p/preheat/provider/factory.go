package provider

import (
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
)

// Factory is responsible to create a new driver based on the metadata.
type Factory func(instance *provider.Instance) (Driver, error)

// DragonflyFactory creates dragonfly driver
func DragonflyFactory(instance *provider.Instance) (Driver, error) {
	return &DragonflyDriver{instance}, nil
}

// KrakenFactory creates kraken driver
func KrakenFactory(instance *provider.Instance) (Driver, error) {
	return &KrakenDriver{instance}, nil
}
