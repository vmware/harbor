package artifactrash

import (
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	"github.com/stretchr/testify/mock"
)

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

import (
	"context"
)

// FakeManager is a fake tag manager that implement the src/pkg/tag.Manager interface
type FakeManager struct {
	mock.Mock
}

// Create ...
func (f *FakeManager) Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Delete ...
func (f *FakeManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

// Filter ...
func (f *FakeManager) Filter(ctx context.Context, timeWindow int64) (arts []model.ArtifactTrash, err error) {
	args := f.Called()
	return args.Get(0).([]model.ArtifactTrash), args.Error(1)
}

// Flush ...
func (f *FakeManager) Flush(ctx context.Context, timeWindow int64) (err error) {
	args := f.Called()
	return args.Error(0)
}
