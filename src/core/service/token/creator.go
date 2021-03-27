// Copyright 2018 Project Harbor Authors
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

package token

import (
	"context"
	"fmt"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"net/http"
	"net/url"
	"strings"

	"github.com/distribution/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

var creatorMap map[string]Creator
var registryFilterMap map[string]accessFilter
var notaryFilterMap map[string]accessFilter
var actionScopeMap = map[rbac.Action]string{
	// Scopes checked by distribution, see: https://github.com/distribution/distribution/blob/master/registry/handlers/app.go
	rbac.ActionPull:   "pull",
	rbac.ActionPush:   "push",
	rbac.ActionDelete: "delete",
	// For skipping policy check when scanner pulls artifacts
	rbac.ActionScannerPull: "scanner-pull",
}

const (
	// Notary service
	Notary = "harbor-notary"
	// Registry service
	Registry = "harbor-registry"
)

// InitCreators initialize the token creators for different services
func InitCreators() {
	creatorMap = make(map[string]Creator)
	registryFilterMap = map[string]accessFilter{
		"repository": &repositoryFilter{
			parser: &basicParser{},
		},
		"registry": &registryFilter{},
	}
	ext, err := config.ExtURL()
	if err != nil {
		log.Warningf("Failed to get ext url, err: %v, the token service will not be functional with notary requests", err)
	} else {
		notaryFilterMap = map[string]accessFilter{
			"repository": &repositoryFilter{
				parser: &endpointParser{
					endpoint: ext,
				},
			},
		}
		creatorMap[Notary] = &generalCreator{
			service:   Notary,
			filterMap: notaryFilterMap,
		}
	}

	creatorMap[Registry] = &generalCreator{
		service:   Registry,
		filterMap: registryFilterMap,
	}
}

// Creator creates a token ready to be served based on the http request.
type Creator interface {
	Create(r *http.Request) (*models.Token, error)
}

type imageParser interface {
	parse(s string) (*image, error)
}

type image struct {
	namespace string
	repo      string
	tag       string
}

type basicParser struct{}

func (b basicParser) parse(s string) (*image, error) {
	return parseImg(s)
}

type endpointParser struct {
	endpoint string
}

func (e endpointParser) parse(s string) (*image, error) {
	repo := strings.SplitN(s, "/", 2)
	if len(repo) < 2 {
		return nil, fmt.Errorf("Unable to parse image from string: %s", s)
	}
	if repo[0] != e.endpoint {
		return nil, fmt.Errorf("Mismatch endpoint from string: %s, expected endpoint: %s", s, e.endpoint)
	}
	return parseImg(repo[1])
}

// build Image accepts a string like library/ubuntu:14.04 and build a image struct
func parseImg(s string) (*image, error) {
	repo := strings.SplitN(s, "/", 2)
	if len(repo) < 2 {
		return nil, fmt.Errorf("Unable to parse image from string: %s", s)
	}
	i := strings.SplitN(repo[1], ":", 2)
	res := &image{
		namespace: repo[0],
		repo:      i[0],
	}
	if len(i) == 2 {
		res.tag = i[1]
	}
	return res, nil
}

// An accessFilter will filter access based on userinfo
type accessFilter interface {
	filter(ctx context.Context, ctl project.Controller, a *token.ResourceActions) error
}

type registryFilter struct {
}

func (reg registryFilter) filter(ctx context.Context, ctl project.Controller,
	a *token.ResourceActions) error {
	// Do not filter if the request is to access registry catalog
	if a.Name != "catalog" {
		return fmt.Errorf("Unable to handle, type: %s, name: %s", a.Type, a.Name)
	}

	secCtx, ok := security.FromContext(ctx)
	if !ok || !secCtx.IsSysAdmin() {
		// Set the actions to empty is the user is not admin
		a.Actions = []string{}
	}
	return nil
}

// repositoryFilter filters the access based on Harbor's permission model
type repositoryFilter struct {
	parser imageParser
}

func (rep repositoryFilter) filter(ctx context.Context, ctl project.Controller,
	a *token.ResourceActions) error {
	// clear action list to assign to new acess element after perm check.
	img, err := rep.parser.parse(a.Name)
	if err != nil {
		return err
	}
	projectName := img.namespace

	project, err := ctl.GetByName(ctx, projectName)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			log.Debugf("project %s does not exist, set empty permission", projectName)
			a.Actions = []string{}
			return nil
		}
		return err
	}

	resource := rbac_project.NewNamespace(project.ProjectID).Resource(rbac.ResourceRepository)
	scopeList := make([]string, 0)
	for s := range resourceScopes(ctx, resource) {
		scopeList = append(scopeList, s)
	}
	a.Actions = scopeList
	return nil
}

func resourceScopes(ctx context.Context, rc rbac.Resource) map[string]struct{} {
	sCtx, _ := security.FromContext(ctx)
	res := map[string]struct{}{}
	for a, s := range actionScopeMap {
		if sCtx.Can(ctx, a, rc) {
			res[s] = struct{}{}
		}
	}

	// "*" is needed in the token for some API in notary server
	// see https://github.com/goharbor/harbor/issues/14303#issuecomment-788010900
	// and https://github.com/theupdateframework/notary/blob/84287fd8df4f172c9a8289641cdfa355fc86989d/server/server.go#L200
	_, ok1 := res["push"]
	_, ok2 := res["pull"]
	_, ok3 := res["delete"]
	if ok1 && ok2 && ok3 {
		res["*"] = struct{}{}
	}
	return res
}

type generalCreator struct {
	service   string
	filterMap map[string]accessFilter
}

type unauthorizedError struct{}

func (e *unauthorizedError) Error() string {
	return "Unauthorized"
}

func (g generalCreator) Create(r *http.Request) (*models.Token, error) {
	var err error
	scopes := parseScopes(r.URL)
	log.Debugf("scopes: %v", scopes)

	ctx, ok := security.FromContext(r.Context())
	if !ok {
		return nil, fmt.Errorf("failed to  get security context from request")
	}

	// for docker login
	if !ctx.IsAuthenticated() {
		if len(scopes) == 0 {
			return nil, &unauthorizedError{}
		}
	}
	access := GetResourceActions(scopes)
	err = filterAccess(r.Context(), access, project.Ctl, g.filterMap)
	if err != nil {
		return nil, err
	}
	return MakeToken(ctx.GetUsername(), g.service, access)
}

func parseScopes(u *url.URL) []string {
	var sector string
	var result []string
	for _, sector = range u.Query()["scope"] {
		result = append(result, strings.Split(sector, " ")...)
	}
	return result
}
