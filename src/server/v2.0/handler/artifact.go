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

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/distribution/distribution/reference"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	"github.com/goharbor/harbor/src/server/v2.0/handler/assembler"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/artifact"
	"github.com/opencontainers/go-digest"
)

func newArtifactAPI() *artifactAPI {
	return &artifactAPI{
		artCtl:  artifact.Ctl,
		proCtl:  project.Ctl,
		repoCtl: repository.Ctl,
		scanCtl: scan.DefaultController,
		tagCtl:  tag.Ctl,
	}
}

type artifactAPI struct {
	BaseAPI
	artCtl  artifact.Controller
	proCtl  project.Controller
	repoCtl repository.Controller
	scanCtl scan.Controller
	tagCtl  tag.Controller
}

func (a *artifactAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	if err := unescapePathParams(params, "RepositoryName"); err != nil {
		a.SendError(ctx, err)
	}

	return nil
}

func (a *artifactAPI) ListArtifacts(ctx context.Context, params operation.ListArtifactsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}

	// set query
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["RepositoryName"] = fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)

	// set option
	option := option(params.WithTag, params.WithImmutableStatus,
		params.WithLabel, params.WithSignature)

	// get the total count of artifacts
	total, err := a.artCtl.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}
	// list artifacts according to the query and option
	arts, err := a.artCtl.List(ctx, query, option)
	if err != nil {
		return a.SendError(ctx, err)
	}

	assembler := assembler.NewVulAssembler(lib.BoolValue(params.WithScanOverview), parseScanReportMimeTypes(params.XAcceptVulnerabilities))
	var artifacts []*models.Artifact
	for _, art := range arts {
		artifact := &model.Artifact{}
		artifact.Artifact = *art
		assembler.WithArtifacts(artifact).Assemble(ctx)
		artifacts = append(artifacts, artifact.ToSwagger())
	}

	return operation.NewListArtifactsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(artifacts)
}

func (a *artifactAPI) GetArtifact(ctx context.Context, params operation.GetArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	// set option
	option := option(params.WithTag, params.WithImmutableStatus,
		params.WithLabel, params.WithSignature)

	// get the artifact
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, option)
	if err != nil {
		return a.SendError(ctx, err)
	}
	art := &model.Artifact{}
	art.Artifact = *artifact

	assembler.NewVulAssembler(lib.BoolValue(params.WithScanOverview), parseScanReportMimeTypes(params.XAcceptVulnerabilities)).WithArtifacts(art).Assemble(ctx)

	return operation.NewGetArtifactOK().WithPayload(art.ToSwagger())
}

func (a *artifactAPI) DeleteArtifact(ctx context.Context, params operation.DeleteArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.Delete(ctx, artifact.ID); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewDeleteArtifactOK()
}

func (a *artifactAPI) CopyArtifact(ctx context.Context, params operation.CopyArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}

	if err := a.requireNonProxyCacheProject(ctx, params.ProjectName); err != nil {
		return a.SendError(ctx, err)
	}

	srcRepo, ref, err := parse(params.From)
	if err != nil {
		return a.SendError(ctx, err)
	}

	srcPro, _ := utils.ParseRepository(srcRepo)
	if err = a.RequireProjectAccess(ctx, srcPro, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}

	dstRepo := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	_, _, err = a.repoCtl.Ensure(ctx, dstRepo)
	if err != nil {
		return a.SendError(ctx, err)
	}

	_, err = a.artCtl.Copy(ctx, srcRepo, ref, dstRepo)
	if err != nil {
		return a.SendError(ctx, err)
	}
	location := strings.TrimSuffix(params.HTTPRequest.URL.Path, "/") + "/" + ref
	return operation.NewCopyArtifactCreated().WithLocation(location)
}

// parse "repository:tag" or "repository@digest" into repository and reference parts
func parse(s string) (string, string, error) {
	matches := reference.ReferenceRegexp.FindStringSubmatch(s)
	if matches == nil {
		return "", "", errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("invalid input: %s", s)
	}
	repository := matches[1]
	reference := matches[2]
	if matches[3] != "" {
		_, err := digest.Parse(matches[3])
		if err != nil {
			return "", "", errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessage("invalid input: %s", s)
		}
		reference = matches[3]
	}
	return repository, reference, nil
}

func (a *artifactAPI) CreateTag(ctx context.Context, params operation.CreateTagParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceTag); err != nil {
		return a.SendError(ctx, err)
	}

	if err := a.requireNonProxyCacheProject(ctx, params.ProjectName); err != nil {
		return a.SendError(ctx, err)
	}

	art, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName),
		params.Reference, &artifact.Option{
			WithTag: true,
		})
	if err != nil {
		return a.SendError(ctx, err)
	}
	tag := &tag.Tag{}
	tag.RepositoryID = art.RepositoryID
	tag.ArtifactID = art.ID
	tag.Name = params.Tag.Name
	tag.PushTime = time.Now()
	if _, err = a.tagCtl.Create(ctx, tag); err != nil {
		return a.SendError(ctx, err)
	}

	// fire event
	notification.AddEvent(ctx, &metadata.CreateTagEventMetadata{
		Ctx:              ctx,
		Tag:              tag.Name,
		AttachedArtifact: &art.Artifact,
	})

	// as we provide no API for get the single tag, ignore setting the location header here
	return operation.NewCreateTagCreated()
}

func (a *artifactAPI) requireNonProxyCacheProject(ctx context.Context, name string) error {
	pro, err := a.proCtl.GetByName(ctx, name)
	if err != nil {
		return err
	}
	if pro.IsProxy() {
		return errors.New(nil).WithCode(errors.MethodNotAllowedCode).
			WithMessage("the operation isn't supported for a proxy cache project")
	}
	return nil
}

func (a *artifactAPI) DeleteTag(ctx context.Context, params operation.DeleteTagParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceTag); err != nil {
		return a.SendError(ctx, err)
	}
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName),
		params.Reference, &artifact.Option{
			WithTag: true,
		})
	if err != nil {
		return a.SendError(ctx, err)
	}
	var id int64
	for _, tag := range artifact.Tags {
		if tag.Name == params.TagName {
			id = tag.ID
			break
		}
	}
	// the tag not found
	if id == 0 {
		err = errors.New(nil).WithCode(errors.NotFoundCode).WithMessage(
			"tag %s attached to artifact %d not found", params.TagName, artifact.ID)
		return a.SendError(ctx, err)
	}
	if err = a.tagCtl.Delete(ctx, id); err != nil {
		return a.SendError(ctx, err)
	}

	// fire event
	notification.AddEvent(ctx, &metadata.DeleteTagEventMetadata{
		Ctx:              ctx,
		Tag:              params.TagName,
		AttachedArtifact: &artifact.Artifact,
	})

	return operation.NewDeleteTagOK()
}

func (a *artifactAPI) ListTags(ctx context.Context, params operation.ListTagsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceTag); err != nil {
		return a.SendError(ctx, err)
	}
	// set query
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}

	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["ArtifactID"] = artifact.ID

	// get the total count of tags
	total, err := a.tagCtl.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	// set option
	option := &tag.Option{}
	if params.WithSignature != nil {
		option.WithSignature = *params.WithSignature
	}
	if params.WithImmutableStatus != nil {
		option.WithImmutableStatus = *params.WithImmutableStatus
	}
	// list tags according to the query and option
	tags, err := a.tagCtl.List(ctx, query, option)
	if err != nil {
		return a.SendError(ctx, err)
	}

	var ts []*models.Tag
	for _, tag := range tags {
		ts = append(ts, model.NewTag(tag).ToSwagger())
	}
	return operation.NewListTagsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(ts)
}

func (a *artifactAPI) GetVulnerabilitiesAddition(ctx context.Context, params operation.GetVulnerabilitiesAdditionParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifactAddition); err != nil {
		return a.SendError(ctx, err)
	}

	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}

	vulnerabilities := make(map[string]interface{})

	for _, mimeType := range parseScanReportMimeTypes(params.XAcceptVulnerabilities) {
		reports, err := a.scanCtl.GetReport(ctx, artifact, []string{mimeType})
		if err != nil {
			return a.SendError(ctx, err)
		}

		for _, rp := range reports {
			// Resolve scan report data only when it is ready
			if len(rp.Report) == 0 {
				continue
			}

			vrp, err := report.ResolveData(rp.MimeType, []byte(rp.Report), report.WithArtifactDigest(rp.Digest))
			if err != nil {
				return a.SendError(ctx, err)
			}

			if v, ok := vulnerabilities[rp.MimeType]; ok {
				r, err := report.Merge(rp.MimeType, v, vrp)
				if err != nil {
					return a.SendError(ctx, err)
				}
				vulnerabilities[rp.MimeType] = r
			} else {
				vulnerabilities[rp.MimeType] = vrp
			}
		}

		if len(vulnerabilities) != 0 {
			break
		}
	}

	content, _ := json.Marshal(vulnerabilities)

	return middleware.ResponderFunc(func(w http.ResponseWriter, p runtime.Producer) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
}

func (a *artifactAPI) GetAddition(ctx context.Context, params operation.GetAdditionParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifactAddition); err != nil {
		return a.SendError(ctx, err)
	}

	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}

	addition, err := a.artCtl.GetAddition(ctx, artifact.ID, strings.ToUpper(params.Addition))
	if err != nil {
		return a.SendError(ctx, err)
	}

	return middleware.ResponderFunc(func(w http.ResponseWriter, p runtime.Producer) {
		w.Header().Set("Content-Type", addition.ContentType)
		w.Write(addition.Content)
	})
}

func (a *artifactAPI) AddLabel(ctx context.Context, params operation.AddLabelParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceArtifactLabel); err != nil {
		return a.SendError(ctx, err)
	}
	art, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.AddLabel(ctx, art.ID, params.Label.ID); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewAddLabelOK()
}

func (a *artifactAPI) RemoveLabel(ctx context.Context, params operation.RemoveLabelParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceArtifactLabel); err != nil {
		return a.SendError(ctx, err)
	}
	art, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.RemoveLabel(ctx, art.ID, params.LabelID); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewRemoveLabelOK()
}

func option(withTag, withImmutableStatus, withLabel, withSignature *bool) *artifact.Option {
	option := &artifact.Option{
		WithTag:   true, // return the tag by default
		WithLabel: lib.BoolValue(withLabel),
	}

	if withTag != nil {
		option.WithTag = *(withTag)
	}

	if option.WithTag {
		option.TagOption = &tag.Option{
			WithImmutableStatus: lib.BoolValue(withImmutableStatus),
			WithSignature:       lib.BoolValue(withSignature),
		}
	}

	return option
}
