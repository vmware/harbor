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

package v2

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/encode/repository"
	"github.com/goharbor/harbor/src/replication/adapter/harbor/base"
	"github.com/goharbor/harbor/src/replication/model"
)

type client struct {
	*base.Client
}

func (c *client) listRepositories(project *base.Project) ([]*model.Repository, error) {
	repositories := []*models.RepoRecord{}
	url := fmt.Sprintf("%s/projects/%s/repositories", c.BasePath(), project.Name)
	if err := c.C.GetAndIteratePagination(url, &repositories); err != nil {
		return nil, err
	}
	var repos []*model.Repository
	for _, repository := range repositories {
		repos = append(repos, &model.Repository{
			Name:     repository.Name,
			Metadata: project.Metadata,
		})
	}
	return repos, nil
}

func (c *client) listArtifacts(repo string) ([]*model.Artifact, error) {
	project, repo := utils.ParseRepository(repo)
	repo = repository.Encode(repo)
	url := fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts?with_label=true",
		c.BasePath(), project, repo)
	artifacts := []*artifact.Artifact{}
	if err := c.C.GetAndIteratePagination(url, &artifacts); err != nil {
		return nil, err
	}
	var arts []*model.Artifact
	for _, artifact := range artifacts {
		art := &model.Artifact{
			Type:   artifact.Type,
			Digest: artifact.Digest,
		}
		for _, label := range artifact.Labels {
			art.Labels = append(art.Labels, label.Name)
		}
		for _, tag := range artifact.Tags {
			art.Tags = append(art.Tags, tag.Name)
		}
		arts = append(arts, art)
	}
	return arts, nil
}

func (c *client) deleteTag(repo, tag string) error {
	project, repo := utils.ParseRepository(repo)
	repo = repository.Encode(repo)
	url := fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts/%s/tags/%s",
		c.BasePath(), project, repo, tag, tag)
	return c.C.Delete(url)
}
