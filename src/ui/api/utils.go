// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/notifier"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/clair"
	registry_error "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/projectmanager"
	uiutils "github.com/vmware/harbor/src/ui/utils"
)

//sysadmin has all privileges to all projects
func listRoles(userID int, projectID int64) ([]models.Role, error) {
	roles := make([]models.Role, 0, 1)
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("failed to determine whether the user %d is system admin: %v", userID, err)
		return roles, err
	}
	if isSysAdmin {
		role, err := dao.GetRoleByID(models.PROJECTADMIN)
		if err != nil {
			log.Errorf("failed to get role %d: %v", models.PROJECTADMIN, err)
			return roles, err
		}
		roles = append(roles, *role)
		return roles, nil
	}

	rs, err := dao.GetUserProjectRoles(userID, projectID)
	if err != nil {
		log.Errorf("failed to get user %d 's roles for project %d: %v", userID, projectID, err)
		return roles, err
	}
	roles = append(roles, rs...)
	return roles, nil
}

func checkUserExists(name string) int {
	u, err := dao.GetUser(models.User{Username: name})
	if err != nil {
		log.Errorf("Error occurred in GetUser, error: %v", err)
		return 0
	}
	if u != nil {
		return u.UserID
	}
	return 0
}

// TriggerReplication triggers the replication according to the policy
func TriggerReplication(policyID int64, repository string,
	tags []string, operation string) error {
	data := struct {
		PolicyID  int64    `json:"policy_id"`
		Repo      string   `json:"repository"`
		Operation string   `json:"operation"`
		TagList   []string `json:"tags"`
	}{
		PolicyID:  policyID,
		Repo:      repository,
		TagList:   tags,
		Operation: operation,
	}

	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	url := buildReplicationURL()

	return uiutils.RequestAsUI("POST", url, bytes.NewBuffer(b), uiutils.NewStatusRespHandler(http.StatusOK))
}

// TriggerReplicationByRepository triggers the replication according to the repository
func TriggerReplicationByRepository(projectID int64, repository string, tags []string, operation string) {
	policies, err := dao.GetRepPolicyByProject(projectID)
	if err != nil {
		log.Errorf("failed to get policies for repository %s: %v", repository, err)
		return
	}

	for _, policy := range policies {
		if policy.Enabled == 0 {
			continue
		}
		if err := TriggerReplication(policy.ID, repository, tags, operation); err != nil {
			log.Errorf("failed to trigger replication of policy %d for %s: %v", policy.ID, repository, err)
		} else {
			log.Infof("replication of policy %d for %s triggered", policy.ID, repository)
		}
	}
}

func postReplicationAction(policyID int64, acton string) error {
	data := struct {
		PolicyID int64  `json:"policy_id"`
		Action   string `json:"action"`
	}{
		PolicyID: policyID,
		Action:   acton,
	}

	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	url := buildReplicationActionURL()

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	uiutils.AddUISecret(req)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("%d %s", resp.StatusCode, string(b))
}

// SyncRegistry syncs the repositories of registry with database.
func SyncRegistry(pm projectmanager.ProjectManager) error {

	log.Infof("Start syncing repositories from registry to DB... ")

	reposInRegistry, err := catalog()
	if err != nil {
		log.Error(err)
		return err
	}

	var repoRecordsInDB []models.RepoRecord
	repoRecordsInDB, err = dao.GetAllRepositories()
	if err != nil {
		log.Errorf("error occurred while getting all registories. %v", err)
		return err
	}

	var reposInDB []string
	for _, repoRecordInDB := range repoRecordsInDB {
		reposInDB = append(reposInDB, repoRecordInDB.Name)
	}

	var reposToAdd []string
	var reposToDel []string
	reposToAdd, reposToDel, err = diffRepos(reposInRegistry, reposInDB, pm)
	if err != nil {
		return err
	}

	if len(reposToAdd) > 0 {
		log.Debugf("Start adding repositories into DB... ")
		for _, repoToAdd := range reposToAdd {
			project, _ := utils.ParseRepository(repoToAdd)
			pullCount, err := dao.CountPull(repoToAdd)
			if err != nil {
				log.Errorf("Error happens when counting pull count from access log: %v", err)
			}
			pro, err := pm.Get(project)
			if err != nil {
				log.Errorf("failed to get project %s: %v", project, err)
				continue
			}
			repoRecord := models.RepoRecord{
				Name:      repoToAdd,
				ProjectID: pro.ProjectID,
				PullCount: pullCount,
			}

			if err := dao.AddRepository(repoRecord); err != nil {
				log.Errorf("Error happens when adding the missing repository: %v", err)
			} else {
				log.Debugf("Add repository: %s success.", repoToAdd)
			}
		}
	}

	if len(reposToDel) > 0 {
		log.Debugf("Start deleting repositories from DB... ")
		for _, repoToDel := range reposToDel {
			if err := dao.DeleteRepository(repoToDel); err != nil {
				log.Errorf("Error happens when deleting the repository: %v", err)
			} else {
				log.Debugf("Delete repository: %s success.", repoToDel)
			}
		}
	}

	log.Infof("Sync repositories from registry to DB is done.")
	return nil
}

func catalog() ([]string, error) {
	repositories := []string{}

	rc, err := initRegistryClient()
	if err != nil {
		return repositories, err
	}

	repositories, err = rc.Catalog()
	if err != nil {
		return repositories, err
	}

	return repositories, nil
}

func diffRepos(reposInRegistry []string, reposInDB []string,
	pm projectmanager.ProjectManager) ([]string, []string, error) {
	var needsAdd []string
	var needsDel []string

	sort.Strings(reposInRegistry)
	sort.Strings(reposInDB)

	i, j := 0, 0
	repoInR, repoInD := "", ""
	for i < len(reposInRegistry) && j < len(reposInDB) {
		repoInR = reposInRegistry[i]
		repoInD = reposInDB[j]
		d := strings.Compare(repoInR, repoInD)
		if d < 0 {
			i++
			exist, err := projectExists(pm, repoInR)
			if err != nil {
				log.Errorf("failed to check the existence of project %s: %v", repoInR, err)
				continue
			}

			if !exist {
				continue
			}

			// TODO remove the workaround when the bug of registry is fixed
			endpoint, err := config.RegistryURL()
			if err != nil {
				return needsAdd, needsDel, err
			}
			client, err := uiutils.NewRepositoryClientForUI(endpoint, true,
				"admin", repoInR, "pull")
			if err != nil {
				return needsAdd, needsDel, err
			}

			exist, err = repositoryExist(repoInR, client)
			if err != nil {
				return needsAdd, needsDel, err
			}

			if !exist {
				continue
			}

			needsAdd = append(needsAdd, repoInR)
		} else if d > 0 {
			needsDel = append(needsDel, repoInD)
			j++
		} else {
			// TODO remove the workaround when the bug of registry is fixed
			endpoint, err := config.RegistryURL()
			if err != nil {
				return needsAdd, needsDel, err
			}
			client, err := uiutils.NewRepositoryClientForUI(endpoint, true, "admin", repoInR, "pull")
			if err != nil {
				return needsAdd, needsDel, err
			}

			exist, err := repositoryExist(repoInR, client)
			if err != nil {
				return needsAdd, needsDel, err
			}

			if !exist {
				needsDel = append(needsDel, repoInD)
			}

			i++
			j++
		}
	}

	for i < len(reposInRegistry) {
		repoInR = reposInRegistry[i]
		i++
		exist, err := projectExists(pm, repoInR)
		if err != nil {
			log.Errorf("failed to check whether project of %s exists: %v", repoInR, err)
			continue
		}

		if !exist {
			continue
		}

		endpoint, err := config.RegistryURL()
		if err != nil {
			log.Errorf("failed to get registry URL: %v", err)
			continue
		}
		client, err := uiutils.NewRepositoryClientForUI(endpoint, true, "admin", repoInR, "pull")
		if err != nil {
			log.Errorf("failed to create repository client: %v", err)
			continue
		}

		exist, err = repositoryExist(repoInR, client)
		if err != nil {
			log.Errorf("failed to check the existence of repository %s: %v", repoInR, err)
			continue
		}

		if !exist {
			continue
		}

		needsAdd = append(needsAdd, repoInR)
	}

	for j < len(reposInDB) {
		needsDel = append(needsDel, reposInDB[j])
		j++
	}

	return needsAdd, needsDel, nil
}

func projectExists(pm projectmanager.ProjectManager, repository string) (bool, error) {
	project, _ := utils.ParseRepository(repository)
	return pm.Exist(project)
}

// TODO need a registry client which accept a raw token as param
func initRegistryClient() (r *registry.Registry, err error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	addr := endpoint
	if strings.Contains(endpoint, "://") {
		addr = strings.Split(endpoint, "://")[1]
	}

	if err := utils.TestTCPConn(addr, 60, 2); err != nil {
		return nil, err
	}

	registryClient, err := NewRegistryClient(endpoint, true, "admin",
		"registry", "catalog", "*")
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func buildReplicationURL() string {
	url := config.InternalJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/replication", url)
}

func buildJobLogURL(jobID string, jobType string) string {
	url := config.InternalJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/%s/%s/log", url, jobType, jobID)
}

func buildReplicationActionURL() string {
	url := config.InternalJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/replication/actions", url)
}

func getReposByProject(name string, keyword ...string) ([]string, error) {
	repositories := []string{}

	repos, err := dao.GetRepositoryByProjectName(name)
	if err != nil {
		return repositories, err
	}

	needMatchKeyword := len(keyword) > 0 && len(keyword[0]) != 0

	for _, repo := range repos {
		if needMatchKeyword &&
			!strings.Contains(repo.Name, keyword[0]) {
			continue
		}

		repositories = append(repositories, repo.Name)
	}

	return repositories, nil
}

func repositoryExist(name string, client *registry.Repository) (bool, error) {
	tags, err := client.ListTag()
	if err != nil {
		if regErr, ok := err.(*registry_error.HTTPError); ok && regErr.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return len(tags) != 0, nil
}

// NewRegistryClient ...
// TODO need a registry client which accept a raw token as param
func NewRegistryClient(endpoint string, insecure bool, username, scopeType, scopeName string,
	scopeActions ...string) (*registry.Registry, error) {
	authorizer := auth.NewRegistryUsernameTokenAuthorizer(username, scopeType, scopeName, scopeActions...)

	store, err := auth.NewAuthorizerStore(endpoint, insecure, authorizer)
	if err != nil {
		return nil, err
	}

	client, err := registry.NewRegistryWithModifiers(endpoint, insecure, store)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// transformVulnerabilities transforms the returned value of Clair API to a list of VulnerabilityItem
func transformVulnerabilities(layerWithVuln *models.ClairLayerEnvelope) []*models.VulnerabilityItem {
	res := []*models.VulnerabilityItem{}
	l := layerWithVuln.Layer
	if l == nil {
		return res
	}
	features := l.Features
	if features == nil {
		return res
	}
	for _, f := range features {
		vulnerabilities := f.Vulnerabilities
		if vulnerabilities == nil {
			continue
		}
		for _, v := range vulnerabilities {
			vItem := &models.VulnerabilityItem{
				ID:          v.Name,
				Pkg:         f.Name,
				Version:     f.Version,
				Severity:    clair.ParseClairSev(v.Severity),
				Fixed:       v.FixedBy,
				Description: v.Description,
			}
			res = append(res, vItem)
		}
	}
	return res
}

//Watch the configuration changes.
//Wrap the same method in common utils.
func watchConfigChanges(cfg map[string]interface{}) error {
	return notifier.WatchConfigChanges(cfg)
}
