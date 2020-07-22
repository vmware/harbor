// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scanner

import (
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	rscanner "github.com/goharbor/harbor/src/pkg/scan/scanner"
)

const (
	proScannerMetaKey = "projectScanner"
	statusUnhealthy   = "unhealthy"
	statusHealthy     = "healthy"
)

// DefaultController is a singleton api controller for plug scanners
var DefaultController = New()

// New a basic controller
func New() Controller {
	return &basicController{
		manager:    rscanner.New(),
		proMetaMgr: metamgr.NewDefaultProjectMetadataManager(),
		clientPool: v1.DefaultClientPool,
	}
}

// basicController is default implementation of api.Controller interface
type basicController struct {
	// Managers for managing the scanner registrations
	manager rscanner.Manager
	// For operating the project level configured scanner
	proMetaMgr metamgr.ProjectMetadataManager
	// Client pool for talking to adapters
	clientPool v1.ClientPool
}

// ListRegistrations ...
func (bc *basicController) ListRegistrations(query *q.Query) ([]*scanner.Registration, error) {
	l, err := bc.manager.List(query)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: list registrations")
	}

	return l, nil
}

// CreateRegistration ...
func (bc *basicController) CreateRegistration(registration *scanner.Registration) (string, error) {
	if isReservedName(registration.Name) {
		return "", errors.BadRequestError(nil).WithMessage(`name "%s" is reserved, please try a different name`, registration.Name)
	}

	// Check if the registration is available
	if _, err := bc.Ping(registration); err != nil {
		return "", errors.Wrap(err, "api controller: create registration")
	}

	// Check if there are any registrations already existing.
	l, err := bc.manager.List(&q.Query{
		PageSize:   1,
		PageNumber: 1,
	})
	if err != nil {
		return "", errors.Wrap(err, "api controller: create registration")
	}

	if len(l) == 0 && !registration.IsDefault {
		// Mark the 1st as default automatically
		registration.IsDefault = true
	}

	return bc.manager.Create(registration)
}

// GetRegistration ...
func (bc *basicController) GetRegistration(registrationUUID string) (*scanner.Registration, error) {
	r, err := bc.manager.Get(registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: get registration")
	}

	return r, nil
}

// RegistrationExists ...
func (bc *basicController) RegistrationExists(registrationUUID string) bool {
	registration, err := bc.manager.Get(registrationUUID)

	// Just logged when an error occurred
	if err != nil {
		logger.Errorf("Check existence of registration error: %s", err)
	}

	return !(err == nil && registration == nil)
}

// UpdateRegistration ...
func (bc *basicController) UpdateRegistration(registration *scanner.Registration) error {
	if registration.IsDefault && registration.Disabled {
		return errors.Errorf("default registration %s can not be marked to disabled", registration.UUID)
	}

	if isReservedName(registration.Name) {
		return errors.BadRequestError(nil).WithMessage(`name "%s" is reserved, please try a different name`, registration.Name)
	}

	return bc.manager.Update(registration)
}

// SetDefaultRegistration ...
func (bc *basicController) DeleteRegistration(registrationUUID string) (*scanner.Registration, error) {
	registration, err := bc.manager.Get(registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: delete registration")
	}

	if registration == nil {
		// Not found
		return nil, nil
	}

	if err := bc.manager.Delete(registrationUUID); err != nil {
		return nil, errors.Wrap(err, "api controller: delete registration")
	}

	return registration, nil
}

// SetDefaultRegistration ...
func (bc *basicController) SetDefaultRegistration(registrationUUID string) error {
	return bc.manager.SetAsDefault(registrationUUID)
}

// SetRegistrationByProject ...
func (bc *basicController) SetRegistrationByProject(projectID int64, registrationID string) error {
	if projectID == 0 {
		return errors.New("invalid project ID")
	}

	if len(registrationID) == 0 {
		return errors.New("missing scanner UUID")
	}

	// Only keep the UUID in the metadata of the given project
	// Scanner metadata existing?
	m, err := bc.proMetaMgr.Get(projectID, proScannerMetaKey)
	if err != nil {
		return errors.Wrap(err, "api controller: set project scanner")
	}

	// Update if exists
	if len(m) > 0 {
		// Compare and set new
		if registrationID != m[proScannerMetaKey] {
			m[proScannerMetaKey] = registrationID
			if err := bc.proMetaMgr.Update(projectID, m); err != nil {
				return errors.Wrap(err, "api controller: set project scanner")
			}
		}
	} else {
		meta := make(map[string]string, 1)
		meta[proScannerMetaKey] = registrationID
		if err := bc.proMetaMgr.Add(projectID, meta); err != nil {
			return errors.Wrap(err, "api controller: set project scanner")
		}
	}

	return nil
}

// GetRegistrationByProject ...
func (bc *basicController) GetRegistrationByProject(projectID int64, options ...Option) (*scanner.Registration, error) {
	if projectID == 0 {
		return nil, errors.New("invalid project ID")
	}

	// First, get it from the project metadata
	m, err := bc.proMetaMgr.Get(projectID, proScannerMetaKey)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: get project scanner")
	}

	var registration *scanner.Registration
	if len(m) > 0 {
		if registrationID, ok := m[proScannerMetaKey]; ok && len(registrationID) > 0 {
			registration, err = bc.manager.Get(registrationID)
			if err != nil {
				return nil, errors.Wrap(err, "api controller: get project scanner")
			}

			if registration == nil {
				// Not found
				// Might be deleted by the admin, the project scanner ID reference should be cleared
				if err := bc.proMetaMgr.Delete(projectID, proScannerMetaKey); err != nil {
					return nil, errors.Wrap(err, "api controller: get project scanner")
				}
			}
		}
	}

	if registration == nil {
		// Second, get the default one
		registration, err = bc.manager.GetDefault()
		if err != nil {
			return nil, errors.Wrap(err, "api controller: get project scanner")
		}
	}

	// No scanner configured
	if registration == nil {
		return nil, nil
	}

	opts := newOptions(options...)

	if opts.Ping {
		// Get metadata of the configured registration
		meta, err := bc.Ping(registration)
		if err != nil {
			// Not blocked, just logged it
			log.Error(errors.Wrap(err, "api controller: get project scanner"))
			registration.Health = statusUnhealthy
		} else {
			registration.Health = statusHealthy
			// Fill in some metadata
			registration.Adapter = meta.Scanner.Name
			registration.Vendor = meta.Scanner.Vendor
			registration.Version = meta.Scanner.Version

			registration.Metadata = meta
		}
	}

	return registration, nil
}

// Ping ...
func (bc *basicController) Ping(registration *scanner.Registration) (*v1.ScannerAdapterMetadata, error) {
	if registration == nil {
		return nil, errors.New("nil registration to ping")
	}

	client, err := registration.Client(bc.clientPool)
	if err != nil {
		return nil, errors.Wrap(err, "scanner controller: ping")
	}

	meta, err := client.GetMetadata()
	if err != nil {
		return nil, errors.Wrap(err, "scanner controller: ping")
	}

	// Validate the required properties
	if meta.Scanner == nil ||
		len(meta.Scanner.Name) == 0 ||
		len(meta.Scanner.Version) == 0 ||
		len(meta.Scanner.Vendor) == 0 {
		return nil, errors.New("invalid scanner in metadata")
	}

	if len(meta.Capabilities) == 0 {
		return nil, errors.New("invalid capabilities in metadata")
	}

	for _, ca := range meta.Capabilities {
		// v1.MimeTypeDockerArtifact is required now
		found := false
		for _, cm := range ca.ConsumesMimeTypes {
			if cm == v1.MimeTypeDockerArtifact {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.Errorf("missing %s in consumes_mime_types", v1.MimeTypeDockerArtifact)
		}

		// v1.MimeTypeNativeReport is required
		found = false
		for _, pm := range ca.ProducesMimeTypes {
			if pm == v1.MimeTypeNativeReport {
				found = true
				break
			}
		}

		if !found {
			return nil, errors.Errorf("missing %s in produces_mime_types", v1.MimeTypeNativeReport)
		}
	}

	return meta, err
}

// GetMetadata ...
func (bc *basicController) GetMetadata(registrationUUID string) (*v1.ScannerAdapterMetadata, error) {
	if len(registrationUUID) == 0 {
		return nil, errors.New("empty registration uuid")
	}

	r, err := bc.manager.Get(registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "scanner controller: get metadata")
	}

	return bc.Ping(r)
}

var (
	reservedNames = []string{"Clair", "Trivy"}
)

func isReservedName(name string) bool {
	for _, reservedName := range reservedNames {
		if name == reservedName {
			return true
		}
	}

	return false
}
