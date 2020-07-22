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

package project

import (
	"context"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/scan/allowlist"
)

var (
	// Ctl is a global project controller instance
	Ctl = NewController()
)

// Controller defines the operations related with blobs
type Controller interface {
	// Get get the project by project id
	Get(ctx context.Context, projectID int64, options ...Option) (*models.Project, error)
	// GetByName get the project by project name
	GetByName(ctx context.Context, projectName string, options ...Option) (*models.Project, error)
	// List list projects
	List(ctx context.Context, query *models.ProjectQueryParam, options ...Option) ([]*models.Project, error)
}

// NewController creates an instance of the default project controller
func NewController() Controller {
	return &controller{
		projectMgr:   project.Mgr,
		metaMgr:      metamgr.NewDefaultProjectMetadataManager(),
		allowlistMgr: allowlist.NewDefaultManager(),
	}
}

type controller struct {
	projectMgr   project.Manager
	metaMgr      metamgr.ProjectMetadataManager
	allowlistMgr allowlist.Manager
}

func (c *controller) Get(ctx context.Context, projectID int64, options ...Option) (*models.Project, error) {
	p, err := c.projectMgr.Get(projectID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.NotFoundError(nil).WithMessage("project %d not found", projectID)
	}

	return c.assembleProject(ctx, p, newOptions(options...))
}

func (c *controller) GetByName(ctx context.Context, projectName string, options ...Option) (*models.Project, error) {
	if projectName == "" {
		return nil, errors.BadRequestError(nil).WithMessage("project name required")
	}

	p, err := c.projectMgr.Get(projectName)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.NotFoundError(nil).WithMessage("project %s not found", projectName)
	}

	return c.assembleProject(ctx, p, newOptions(options...))
}

func (c *controller) List(ctx context.Context, query *models.ProjectQueryParam, options ...Option) ([]*models.Project, error) {
	projects, err := c.projectMgr.List(query)
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)
	for _, p := range projects {
		if _, err := c.assembleProject(ctx, p, opts); err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (c *controller) assembleProject(ctx context.Context, p *models.Project, opts *Options) (*models.Project, error) {
	if opts.Metadata {
		meta, err := c.metaMgr.Get(p.ProjectID)
		if err != nil {
			return nil, err
		}
		if len(p.Metadata) == 0 {
			p.Metadata = make(map[string]string)
		}

		for k, v := range meta {
			p.Metadata[k] = v
		}
	}

	if opts.CVEAllowlist {
		if p.ReuseSysCVEAllowlist() {
			wl, err := c.allowlistMgr.GetSys()
			if err != nil {
				log.Errorf("get system CVE allowlist failed, error: %v", err)
				return nil, err
			}

			wl.ProjectID = p.ProjectID
			p.CVEAllowlist = *wl
		} else {
			wl, err := c.allowlistMgr.Get(p.ProjectID)
			if err != nil {
				return nil, err
			}

			p.CVEAllowlist = *wl
		}

	}

	return p, nil
}
