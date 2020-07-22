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

package gc

import (
	"os"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	redislib "github.com/goharbor/harbor/src/lib/redis"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifactrash"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/registryctl/client"
)

var (
	regCtlInit = registryctl.Init
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	blobPrefix            = "blobs::*"
	repoPrefix            = "repository::*"
	uploadSizePattern     = "upload:*:size"
)

// GarbageCollector is the struct to run registry's garbage collection
type GarbageCollector struct {
	artCtl            artifact.Controller
	artrashMgr        artifactrash.Manager
	blobMgr           blob.Manager
	projectCtl        project.Controller
	registryCtlClient client.Client
	logger            logger.Interface
	redisURL          string
	deleteUntagged    bool
	dryRun            bool
	// holds all of trashed artifacts' digest and repositories.
	// The source data of trashedArts is the table ArtifactTrash and it's only used as a dictionary by sweep when to delete a manifest.
	// As table blob has no repositories data, and the repositories are required when to delete a manifest, so use the table ArtifactTrash to capture them.
	trashedArts map[string][]string
	// hold all of GC candidates(non-referenced blobs), it's captured by mark and consumed by sweep.
	deleteSet       []*blob_models.Blob
	timeWindowHours int64
}

// MaxFails implements the interface in job/Interface
func (gc *GarbageCollector) MaxFails() uint {
	return 1
}

// MaxCurrency is implementation of same method in Interface.
func (gc *GarbageCollector) MaxCurrency() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (gc *GarbageCollector) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (gc *GarbageCollector) Validate(params job.Parameters) error {
	return nil
}

func (gc *GarbageCollector) init(ctx job.Context, params job.Parameters) error {
	regCtlInit()
	gc.logger = ctx.GetLogger()
	gc.deleteSet = make([]*blob_models.Blob, 0)
	gc.trashedArts = make(map[string][]string, 0)
	opCmd, flag := ctx.OPCommand()
	if flag && opCmd.IsStop() {
		gc.logger.Info("received the stop signal, quit GC job.")
		return nil
	}
	// UT will use the mock client, ctl and mgr
	if os.Getenv("UTTEST") != "true" {
		gc.registryCtlClient = registryctl.RegistryCtlClient
		gc.artCtl = artifact.Ctl
		gc.artrashMgr = artifactrash.NewManager()
		gc.blobMgr = blob.NewManager()
		gc.projectCtl = project.Ctl
	}
	if err := gc.registryCtlClient.Health(); err != nil {
		gc.logger.Errorf("failed to start gc as registry controller is unreachable: %v", err)
		return err
	}
	gc.parseParams(params)
	return nil
}

// parseParams set the parameters according to the GC API call.
func (gc *GarbageCollector) parseParams(params job.Parameters) {
	// redis url
	gc.logger.Info(params)
	gc.redisURL = params["redis_url_reg"].(string)

	// delete untagged: default is to delete the untagged artifact
	gc.deleteUntagged = true
	deleteUntagged, exist := params["delete_untagged"]
	if exist {
		if untagged, ok := deleteUntagged.(bool); ok && !untagged {
			gc.deleteUntagged = untagged
		}
	}

	// time window: default is 2 hours, and for testing/debugging, it can be set to 0.
	gc.timeWindowHours = 2
	timeWindow, exist := params["time_window"]
	if exist {
		if timeWindow, ok := timeWindow.(float64); ok {
			gc.timeWindowHours = int64(timeWindow)
		}
	}

	// dry run: default is false. And for dry run we can have button in the UI.
	gc.dryRun = false
	dryRun, exist := params["dry_run"]
	if exist {
		if dryRun, ok := dryRun.(bool); ok && dryRun {
			gc.dryRun = dryRun
		}
	}
}

// Run implements the interface in job/Interface
func (gc *GarbageCollector) Run(ctx job.Context, params job.Parameters) error {
	if err := gc.init(ctx, params); err != nil {
		return err
	}

	gc.logger.Infof("start to run gc in job.")

	// mark
	if err := gc.mark(ctx); err != nil {
		gc.logger.Errorf("failed to execute GC job at mark phase, error: %v", err)
		return err
	}

	// sweep
	if !gc.dryRun {
		if err := gc.sweep(ctx); err != nil {
			gc.logger.Errorf("failed to execute GC job at sweep phase, error: %v", err)
			return err
		}

		if err := gc.cleanCache(); err != nil {
			return err
		}
	}
	gc.logger.Infof("success to run gc in job.")
	return nil
}

// mark
func (gc *GarbageCollector) mark(ctx job.Context) error {
	arts, err := gc.deletedArt(ctx)
	if err != nil {
		gc.logger.Errorf("failed to get deleted Artifacts in gc job, with error: %v", err)
		return err
	}
	// no need to execute GC as there is no removed artifacts.
	// Do this is to handle if user trigger GC job several times, only one job should do the following logic as artifact trash table is flushed.
	if len(arts) == 0 {
		gc.logger.Info("no need to execute GC as there is no removed artifacts.")
		return nil
	}
	gc.trashedArts = arts

	// get gc candidates, and set the repositories.
	// AS the reference count is calculated by joining table project_blob and blob, here needs to call removeUntaggedBlobs to remove these non-used blobs from table project_blob firstly.
	if !gc.dryRun {
		gc.removeUntaggedBlobs(ctx)
	}
	blobs, err := gc.blobMgr.UselessBlobs(ctx.SystemContext(), gc.timeWindowHours)
	if err != nil {
		gc.logger.Errorf("failed to get gc candidate: %v", err)
		return err
	}
	if len(blobs) == 0 {
		gc.logger.Info("no need to execute GC as there is no non referenced artifacts.")
		return nil
	}

	// update delete status for the candidates.
	blobCt := 0
	mfCt := 0
	makeSize := int64(0)
	for _, blob := range blobs {
		if !gc.dryRun {
			blob.Status = blob_models.StatusDelete
			count, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
			if err != nil {
				gc.logger.Warningf("failed to mark gc candidate, skip it.: %s, error: %v", blob.Digest, err)
				continue
			}
			if count == 0 {
				gc.logger.Warningf("no blob found to mark gc candidate, skip it. ID:%d, digest:%s", blob.ID, blob.Digest)
				continue
			}
		}
		gc.logger.Infof("blob eligible for deletion: %s", blob.Digest)
		gc.deleteSet = append(gc.deleteSet, blob)
		// as table blob has no repository name, here needs to use the ArtifactTrash to fill it in.
		if blob.IsManifest() {
			mfCt++
		} else {
			blobCt++
		}
		makeSize = makeSize + blob.Size
	}
	gc.logger.Infof("%d blobs and %d manifests eligible for deletion", blobCt, mfCt)
	gc.logger.Infof("The GC could free up %d MB space, the size is a rough estimate.", makeSize/1024/1024)
	return nil
}

func (gc *GarbageCollector) sweep(ctx job.Context) error {
	gc.logger = ctx.GetLogger()
	sweepSize := int64(0)
	for _, blob := range gc.deleteSet {
		// set the status firstly, if the blob is updated by any HEAD/PUT request, it should be fail and skip.
		blob.Status = blob_models.StatusDeleting
		count, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
		if err != nil {
			gc.logger.Errorf("failed to mark gc candidate deleting, skip: %s, %s", blob.Digest, blob.Status)
			continue
		}
		if count == 0 {
			gc.logger.Warningf("no blob found to mark gc candidate deleting, ID:%d, digest:%s", blob.ID, blob.Digest)
			continue
		}

		// remove tags and revisions of a manifest
		if _, exist := gc.trashedArts[blob.Digest]; exist && blob.IsManifest() {
			for _, repo := range gc.trashedArts[blob.Digest] {
				// Harbor cannot know the existing tags in the backend from its database, so let the v2 DELETE manifest to remove all of them.
				gc.logger.Infof("delete the manifest with registry v2 API: %s, %s, %s",
					repo, blob.ContentType, blob.Digest)
				if err := v2DeleteManifest(repo, blob.Digest); err != nil {
					gc.logger.Errorf("failed to delete manifest with v2 API, %s, %s, %v", repo, blob.Digest, err)
					if err := ignoreNotFound(func() error {
						return gc.markDeleteFailed(ctx, blob)
					}); err != nil {
						return err
					}
					return errors.Wrapf(err, "failed to delete manifest with v2 API: %s, %s", repo, blob.Digest)
				}
				// for manifest, it has to delete the revisions folder of each repository
				gc.logger.Infof("delete manifest from storage: %s", blob.Digest)
				if err := ignoreNotFound(func() error {
					return gc.registryCtlClient.DeleteManifest(repo, blob.Digest)
				}); err != nil {
					if err := ignoreNotFound(func() error {
						return gc.markDeleteFailed(ctx, blob)
					}); err != nil {
						return err
					}
					return errors.Wrapf(err, "failed to remove manifest from storage: %s, %s", repo, blob.Digest)
				}
			}
		}

		// delete all of blobs, which include config, layer and manifest
		gc.logger.Infof("delete blob from storage: %s", blob.Digest)
		if err := ignoreNotFound(func() error {
			return gc.registryCtlClient.DeleteBlob(blob.Digest)
		}); err != nil {
			if err := ignoreNotFound(func() error {
				return gc.markDeleteFailed(ctx, blob)
			}); err != nil {
				return err
			}
			return errors.Wrapf(err, "failed to delete blob from storage: %s, %s", blob.Digest, blob.Status)

		}

		// remove the blob record
		if err := ignoreNotFound(func() error {
			return gc.blobMgr.Delete(ctx.SystemContext(), blob.ID)
		}); err != nil {
			if err := ignoreNotFound(func() error {
				return gc.markDeleteFailed(ctx, blob)
			}); err != nil {
				return err
			}
			return errors.Wrapf(err, "failed to delete blob from database: %s, %s", blob.Digest, blob.Status)
		}
		sweepSize = sweepSize + blob.Size
	}
	gc.logger.Infof("The GC job actual frees up %d MB space.", sweepSize/1024/1024)
	return nil
}

// cleanCache is to clean the registry cache for GC.
// To do this is because the issue https://github.com/docker/distribution/issues/2094
func (gc *GarbageCollector) cleanCache() error {
	pool, err := redislib.GetRedisPool("GarbageCollector", gc.redisURL, &redislib.PoolParam{
		PoolMaxIdle:           0,
		PoolMaxActive:         1,
		PoolIdleTimeout:       60 * time.Second,
		DialConnectionTimeout: dialConnectionTimeout,
		DialReadTimeout:       dialReadTimeout,
		DialWriteTimeout:      dialWriteTimeout,
	})
	if err != nil {
		gc.logger.Errorf("failed to connect to redis %v", err)
		return err
	}
	con := pool.Get()
	defer con.Close()

	// clean all keys in registry redis DB.

	// sample of keys in registry redis:
	// 1) "blobs::sha256:1a6fd470b9ce10849be79e99529a88371dff60c60aab424c077007f6979b4812"
	// 2) "repository::library/hello-world::blobs::sha256:4ab4c602aa5eed5528a6620ff18a1dc4faef0e1ab3a5eddeddb410714478c67f"
	// 3) "upload:fbd2e0a3-262d-40bb-abe4-2f43aa6f9cda:size"
	patterns := []string{blobPrefix, repoPrefix, uploadSizePattern}
	for _, pattern := range patterns {
		if err := delKeys(con, pattern); err != nil {
			gc.logger.Errorf("failed to clean registry cache %v, pattern %s", err, pattern)
			return err
		}
	}

	return nil
}

// deletedArt contains the two parts of artifact
// 1, required part, the artifacts were removed from Harbor.
// 2, optional part, the untagged artifacts.
func (gc *GarbageCollector) deletedArt(ctx job.Context) (map[string][]string, error) {
	if os.Getenv("UTTEST") == "true" {
		gc.logger = ctx.GetLogger()
	}
	// default is not to clean trash
	flushTrash := false
	defer func() {
		if flushTrash {
			gc.logger.Info("flush artifact trash")
			if err := gc.artrashMgr.Flush(ctx.SystemContext(), gc.timeWindowHours); err != nil {
				gc.logger.Errorf("failed to flush artifact trash: %v", err)
			}
		}
	}()
	arts := make([]model.ArtifactTrash, 0)

	// artMap : map[digest : []repo list]
	artMap := make(map[string][]string)
	// handle the optional ones, and the artifact controller will move them into trash.
	if gc.deleteUntagged {
		untagged, err := gc.artCtl.List(ctx.SystemContext(), &q.Query{
			Keywords: map[string]interface{}{
				"Tags": "nil",
			},
		}, nil)
		if err != nil {
			return artMap, err
		}
		gc.logger.Info("start to delete untagged artifact.")
		for _, art := range untagged {
			if err := gc.artCtl.Delete(ctx.SystemContext(), art.ID); err != nil {
				// the failure ones can be GCed by the next execution
				gc.logger.Errorf("failed to delete untagged:%d artifact in DB, error, %v", art.ID, err)
				continue
			}
			gc.logger.Infof("delete the untagged artifact: ProjectID:(%d)-RepositoryName(%s)-MediaType:(%s)-Digest:(%s)",
				art.ProjectID, art.RepositoryName, art.ManifestMediaType, art.Digest)
		}
		gc.logger.Info("end to delete untagged artifact.")
	}

	// filter gets all of deleted artifact, here do not need time window as the manifest candidate has to remove all of its reference.
	arts, err := gc.artrashMgr.Filter(ctx.SystemContext(), 0)
	if err != nil {
		return artMap, err
	}

	// the repositories of blob is needed when to delete as a manifest.
	for _, art := range arts {
		_, exist := artMap[art.Digest]
		if !exist {
			artMap[art.Digest] = []string{art.RepositoryName}
		} else {
			repos := artMap[art.Digest]
			repos = append(repos, art.RepositoryName)
			artMap[art.Digest] = repos
		}
	}

	gc.logger.Info("required candidate: %+v", arts)
	if !gc.dryRun {
		flushTrash = true
	}
	return artMap, nil
}

// clean the untagged blobs in each project, these blobs are not referenced by any manifest and will be cleaned by GC
func (gc *GarbageCollector) removeUntaggedBlobs(ctx job.Context) {
	// get all projects
	projects := func(chunkSize int) <-chan *models.Project {
		ch := make(chan *models.Project, chunkSize)

		go func() {
			defer close(ch)

			params := &models.ProjectQueryParam{
				Pagination: &models.Pagination{Page: 1, Size: int64(chunkSize)},
			}

			for {
				results, err := gc.projectCtl.List(ctx.SystemContext(), params, project.Metadata(false))
				if err != nil {
					gc.logger.Errorf("list projects failed, error: %v", err)
					return
				}

				for _, p := range results {
					ch <- p
				}

				if len(results) < chunkSize {
					break
				}

				params.Pagination.Page++
			}

		}()

		return ch
	}(50)

	for project := range projects {
		all, err := gc.blobMgr.List(ctx.SystemContext(), blob.ListParams{
			ProjectID:  project.ProjectID,
			UpdateTime: time.Now().Add(-time.Duration(gc.timeWindowHours) * time.Hour),
		})
		if err != nil {
			gc.logger.Errorf("failed to get blobs of project, %v", err)
			continue
		}
		if err := gc.blobMgr.CleanupAssociationsForProject(ctx.SystemContext(), project.ProjectID, all); err != nil {
			gc.logger.Errorf("failed to clean untagged blobs of project, %v", err)
			continue
		}
	}
}

// markDeleteFailed set the blob status to StatusDeleteFailed
func (gc *GarbageCollector) markDeleteFailed(ctx job.Context, blob *blob_models.Blob) error {
	blob.Status = blob_models.StatusDeleteFailed
	count, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
	if err != nil {
		gc.logger.Errorf("failed to mark gc candidate delete failed: %s, %s", blob.Digest, blob.Status)
		return errors.Wrapf(err, "failed to mark gc candidate delete failed: %s, %s", blob.Digest, blob.Status)
	}
	if count == 0 {
		return errors.New(nil).WithMessage("no blob found to mark gc candidate, ID:%d, digest:%s", blob.ID, blob.Digest).WithCode(errors.NotFoundCode)
	}
	return nil
}
