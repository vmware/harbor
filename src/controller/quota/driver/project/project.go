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
	"fmt"
	"strconv"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib/log"
	dr "github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/graph-gophers/dataloader"
)

func init() {
	dr.Register("project", newDriver())
}

type driver struct {
	cfg    *config.CfgManager
	loader *dataloader.Loader

	blobCtl blob.Controller
}

func (d *driver) Enabled(ctx context.Context, key string) (bool, error) {
	// NOTE: every time load the new configurations from the db to get the latest configurations may have performance problem.
	if err := d.cfg.Load(); err != nil {
		return false, err
	}
	return d.cfg.Get(common.QuotaPerProjectEnable).GetBool(), nil
}

func (d *driver) HardLimits(ctx context.Context) types.ResourceList {
	// NOTE: every time load the new configurations from the db to get the latest configurations may have performance problem.
	if err := d.cfg.Load(); err != nil {
		log.Warningf("load configurations failed, error: %v", err)
	}

	return types.ResourceList{
		types.ResourceStorage: d.cfg.Get(common.StoragePerProject).GetInt64(),
	}
}

func (d *driver) Load(ctx context.Context, key string) (dr.RefObject, error) {
	thunk := d.loader.Load(ctx, dataloader.StringKey(key))

	result, err := thunk()
	if err != nil {
		return nil, err
	}

	project, ok := result.(*models.Project)
	if !ok {
		return nil, fmt.Errorf("bad result for project: %s", key)
	}

	return dr.RefObject{
		"id":         project.ProjectID,
		"name":       project.Name,
		"owner_name": project.OwnerName,
	}, nil
}

func (d *driver) Validate(hardLimits types.ResourceList) error {
	resources := map[types.ResourceName]bool{
		types.ResourceStorage: true,
	}

	for resource, value := range hardLimits {
		if !resources[resource] {
			return fmt.Errorf("resource %s not support", resource)
		}

		if value <= 0 && value != types.UNLIMITED {
			return fmt.Errorf("invalid value for resource %s", resource)
		}
	}

	for resource := range resources {
		if _, found := hardLimits[resource]; !found {
			return fmt.Errorf("resource %s not found", resource)
		}
	}

	return nil
}

func (d *driver) CalculateUsage(ctx context.Context, key string) (types.ResourceList, error) {
	projectID, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return nil, err
	}

	size, err := d.blobCtl.CalculateTotalSizeByProject(ctx, projectID, true)
	if err != nil {
		return nil, err
	}

	return types.ResourceList{types.ResourceStorage: size}, nil
}

func newDriver() dr.Driver {
	cfg := config.NewDBCfgManager()

	loader := dataloader.NewBatchedLoader(getProjectsBatchFn, dataloader.WithClearCacheOnBatch())

	return &driver{
		cfg:     cfg,
		loader:  loader,
		blobCtl: blob.Ctl,
	}
}
