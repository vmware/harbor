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
	"github.com/stretchr/testify/mock"
)

// FakeManager is a fake label manager that implement the src/pkg/label.Manager interface
type FakeManager struct {
	mock.Mock
}

// Get ...
func (f *FakeManager) Get(ctx context.Context, id int64) (*models.Label, error) {
	args := f.Called()
	var label *models.Label
	if args.Get(0) != nil {
		label = args.Get(0).(*models.Label)
	}
	return label, args.Error(1)
}

// ListByArtifact ...
func (f *FakeManager) ListByArtifact(ctx context.Context, artifactID int64) ([]*models.Label, error) {
	args := f.Called()
	var labels []*models.Label
	if args.Get(0) != nil {
		labels = args.Get(0).([]*models.Label)
	}
	return labels, args.Error(1)
}

// AddTo ...
func (f *FakeManager) AddTo(ctx context.Context, labelID int64, artifactID int64) error {
	args := f.Called()
	return args.Error(0)
}

// RemoveFrom ...
func (f *FakeManager) RemoveFrom(ctx context.Context, labelID int64, artifactID int64) error {
	args := f.Called()
	return args.Error(0)
}

// RemoveAllFrom ...
func (f *FakeManager) RemoveAllFrom(ctx context.Context, artifactID int64) error {
	args := f.Called()
	return args.Error(0)
}

// RemoveFromAllArtifacts ...
func (f *FakeManager) RemoveFromAllArtifacts(ctx context.Context, labelID int64) error {
	args := f.Called()
	return args.Error(0)
}
