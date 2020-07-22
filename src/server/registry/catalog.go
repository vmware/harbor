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

package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/registry/util"
	"net/http"
	"sort"
	"strconv"
)

func newRepositoryHandler() http.Handler {
	return &repositoryHandler{
		repoCtl: repository.Ctl,
		artCtl:  artifact.Ctl,
	}
}

type repositoryHandler struct {
	repoCtl repository.Controller
	artCtl  artifact.Controller
}

func (r *repositoryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var maxEntries int
	var err error

	reqQ := req.URL.Query()
	lastEntry := reqQ.Get("last")
	withN := reqQ.Get("n") != ""
	if withN {
		maxEntries, err = strconv.Atoi(reqQ.Get("n"))
		if err != nil || maxEntries < 0 {
			err := errors.New(err).WithCode(errors.BadRequestCode).WithMessage("the N must be a positive int type")
			lib_http.SendError(w, err)
			return
		}
	}

	repoNames := make([]string, 0)
	// get all repositories
	repoRecords, err := r.repoCtl.List(req.Context(), nil)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	if len(repoRecords) <= 0 {
		r.sendResponse(w, req, repoNames)
		return
	}
	for _, repo := range repoRecords {
		valid, err := r.validateRepo(req.Context(), repo.RepositoryID)
		if err != nil {
			lib_http.SendError(w, err)
			return
		}
		if valid {
			repoNames = append(repoNames, repo.Name)
		}
	}
	sort.Strings(repoNames)
	if !withN {
		r.sendResponse(w, req, repoNames)
		return
	}

	// handle the pagination
	resRepos := repoNames
	repoNamesLen := len(repoNames)
	// with "last", get items form lastEntryIndex+1 to lastEntryIndex+maxEntries
	// without "last", get items from 0 to maxEntries'
	if lastEntry != "" {
		lastEntryIndex := util.IndexString(repoNames, lastEntry)
		if lastEntryIndex == -1 {
			err := errors.New(nil).WithCode(errors.BadRequestCode).WithMessage(fmt.Sprintf("the last: %s should be a valid repository name.", lastEntry))
			lib_http.SendError(w, err)
			return
		}
		if lastEntryIndex+1+maxEntries > repoNamesLen {
			resRepos = repoNames[lastEntryIndex+1 : repoNamesLen]
		} else {
			resRepos = repoNames[lastEntryIndex+1 : lastEntryIndex+1+maxEntries]
		}
	} else {
		if maxEntries > repoNamesLen {
			maxEntries = repoNamesLen
		}
		resRepos = repoNames[0:maxEntries]
	}

	if len(resRepos) == 0 {
		r.sendResponse(w, req, resRepos)
		return
	}

	// compare the last item to define whether return the link header.
	// if equals, means that there is no more items in DB. Do not need to give the link header.
	if repoNames[len(repoNames)-1] != resRepos[len(resRepos)-1] {
		urlStr, err := util.SetLinkHeader(req.URL.String(), maxEntries, resRepos[len(resRepos)-1])
		if err != nil {
			lib_http.SendError(w, err)
			return
		}
		w.Header().Set("Link", urlStr)
	}

	r.sendResponse(w, req, resRepos)
	return
}

// sendResponse ...
func (r *repositoryHandler) sendResponse(w http.ResponseWriter, req *http.Request, repositoryNames []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(catalogAPIResponse{
		Repositories: repositoryNames,
	}); err != nil {
		lib_http.SendError(w, err)
		return
	}
}

// empty repo and all of artifacts are untagged should be filtered out.
func (r *repositoryHandler) validateRepo(ctx context.Context, repositoryID int64) (bool, error) {
	arts, err := r.artCtl.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repositoryID,
		},
	}, &artifact.Option{
		WithTag: true,
	})
	if err != nil {
		return false, err
	}

	// empty repo
	if len(arts) == 0 {
		return false, nil
	}

	for _, art := range arts {
		if len(art.Tags) != 0 {
			return true, nil
		}
	}

	// if all of artifact are untagged, filter out
	return false, nil
}

type catalogAPIResponse struct {
	Repositories []string `json:"repositories"`
}
