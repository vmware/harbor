package registryctl

import (
	"github.com/goharbor/harbor/src/registryctl/api/registry/gc"
	"github.com/stretchr/testify/mock"
)

type Mockclient struct {
	mock.Mock
}

// Health ...
func (c *Mockclient) Health() error {
	return nil
}

// StartGC ...
func (c *Mockclient) StartGC() (*gc.Result, error) {
	result := &gc.Result{
		Status: true,
		Msg:    "this is a mock client",
	}
	return result, nil
}

// DeleteBlob ...
func (c *Mockclient) DeleteBlob(reference string) (err error) {
	return nil
}

// DeleteManifest ...
func (c *Mockclient) DeleteManifest(repository, reference string) (err error) {
	return nil
}
