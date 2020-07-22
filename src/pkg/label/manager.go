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

package label

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"time"
)

// Mgr is a global instance of label manager
var Mgr = New()

// Manager manages the labels and references between label and resource
type Manager interface {
	// Get the label specified by ID
	Get(ctx context.Context, id int64) (label *models.Label, err error)
	// List labels that added to the artifact specified by the ID
	ListByArtifact(ctx context.Context, artifactID int64) (labels []*models.Label, err error)
	// Add label to the artifact specified the ID
	AddTo(ctx context.Context, labelID int64, artifactID int64) (err error)
	// Remove the label added to the artifact specified by the ID
	RemoveFrom(ctx context.Context, labelID int64, artifactID int64) (err error)
	// Remove all labels added to the artifact specified by the ID
	RemoveAllFrom(ctx context.Context, artifactID int64) (err error)
	// RemoveFromAllArtifacts removes the label specified by the ID from all artifacts
	RemoveFromAllArtifacts(ctx context.Context, labelID int64) (err error)
}

// New creates an instance of the default label manager
func New() Manager {
	return &manager{
		dao: &defaultDAO{},
	}
}

type manager struct {
	dao DAO
}

func (m *manager) Get(ctx context.Context, id int64) (*models.Label, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) ListByArtifact(ctx context.Context, artifactID int64) ([]*models.Label, error) {
	return m.dao.ListByArtifact(ctx, artifactID)
}

func (m *manager) AddTo(ctx context.Context, labelID int64, artifactID int64) error {
	now := time.Now()
	_, err := m.dao.CreateReference(ctx, &Reference{
		LabelID:      labelID,
		ArtifactID:   artifactID,
		CreationTime: now,
		UpdateTime:   now,
	})
	return err
}
func (m *manager) RemoveFrom(ctx context.Context, labelID int64, artifactID int64) error {
	n, err := m.dao.DeleteReferences(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"LabelID":    labelID,
			"ArtifactID": artifactID,
		},
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("reference with label %d and artifact %d not found", labelID, artifactID)
	}
	return nil
}

func (m *manager) RemoveAllFrom(ctx context.Context, artifactID int64) error {
	_, err := m.dao.DeleteReferences(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ArtifactID": artifactID,
		},
	})
	return err
}

func (m *manager) RemoveFromAllArtifacts(ctx context.Context, labelID int64) error {
	_, err := m.dao.DeleteReferences(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"LabelID": labelID,
		},
	})
	return err
}
