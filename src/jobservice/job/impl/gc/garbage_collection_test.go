package gc

import (
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/models"
	commom_regctl "github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	pkg_blob "github.com/goharbor/harbor/src/pkg/blob/models"
	htesting "github.com/goharbor/harbor/src/testing"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	trashtesting "github.com/goharbor/harbor/src/testing/pkg/artifactrash"
	"github.com/goharbor/harbor/src/testing/pkg/blob"
	"github.com/goharbor/harbor/src/testing/registryctl"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type gcTestSuite struct {
	htesting.Suite
	artifactCtl       *artifacttesting.Controller
	artrashMgr        *trashtesting.FakeManager
	registryCtlClient *registryctl.Mockclient
	projectCtl        *projecttesting.Controller
	blobMgr           *blob.Manager

	regCtlInit func()
}

func (suite *gcTestSuite) SetupTest() {
	suite.artifactCtl = &artifacttesting.Controller{}
	suite.artrashMgr = &trashtesting.FakeManager{}
	suite.registryCtlClient = &registryctl.Mockclient{}
	suite.blobMgr = &blob.Manager{}
	suite.projectCtl = &projecttesting.Controller{}

	regCtlInit = func() { commom_regctl.RegistryCtlClient = suite.registryCtlClient }
}

func (suite *gcTestSuite) TestMaxFails() {
	gc := &GarbageCollector{}
	suite.Equal(uint(1), gc.MaxFails())
}

func (suite *gcTestSuite) TestShouldRetry() {
	gc := &GarbageCollector{}
	suite.False(gc.ShouldRetry())
}

func (suite *gcTestSuite) TestValidate() {
	gc := &GarbageCollector{}
	suite.Nil(gc.Validate(nil))
}

func (suite *gcTestSuite) TestDeletedArt() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	suite.artrashMgr.On("Flush").Return(nil)
	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	suite.artrashMgr.On("Filter").Return([]model.ArtifactTrash{
		{
			ID:                1,
			Digest:            suite.DigestString(),
			ManifestMediaType: schema2.MediaTypeManifest,
		},
	}, nil)

	gc := &GarbageCollector{
		artCtl:     suite.artifactCtl,
		artrashMgr: suite.artrashMgr,
	}

	arts, err := gc.deletedArt(ctx)
	suite.Nil(err)
	suite.Equal(1, len(arts))
}

func (suite *gcTestSuite) TestRemoveUntaggedBlobs() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	mock.OnAnything(suite.projectCtl, "List").Return([]*models.Project{
		{
			ProjectID: 1234,
			Name:      "test GC",
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "List").Return([]*pkg_blob.Blob{
		{
			ID:     1234,
			Digest: "sha256:1234",
			Size:   1234,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "CleanupAssociationsForProject").Return(nil)

	gc := &GarbageCollector{
		projectCtl: suite.projectCtl,
		blobMgr:    suite.blobMgr,
	}

	suite.NotPanics(func() {
		gc.removeUntaggedBlobs(ctx)
	})
}

func (suite *gcTestSuite) TestInit() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	mock.OnAnything(ctx, "Get").Return("core url", true)
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.NilCommand, true)

	gc := &GarbageCollector{
		registryCtlClient: suite.registryCtlClient,
	}
	params := map[string]interface{}{
		"delete_untagged": true,
		"redis_url_reg":   "redis url",
		"time_window":     1,
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)

	params = map[string]interface{}{
		"delete_untagged": "unsupported",
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)

	params = map[string]interface{}{
		"delete_untagged": false,
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.False(gc.deleteUntagged)

	params = map[string]interface{}{
		"redis_url_reg": "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)
}

func (suite *gcTestSuite) TestStop() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	mock.OnAnything(ctx, "Get").Return("core url", true)
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.StopCommand, true)

	gc := &GarbageCollector{
		registryCtlClient: suite.registryCtlClient,
	}
	params := map[string]interface{}{
		"delete_untagged": true,
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))

	ctx = &mockjobservice.MockJobContext{}
	mock.OnAnything(ctx, "Get").Return("core url", true)
	ctx.On("OPCommand").Return(job.StopCommand, false)
	suite.Nil(gc.init(ctx, params))

	ctx = &mockjobservice.MockJobContext{}
	mock.OnAnything(ctx, "Get").Return("core url", true)
	ctx.On("OPCommand").Return(job.NilCommand, true)
	suite.Nil(gc.init(ctx, params))

}

func (suite *gcTestSuite) TestRun() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.NilCommand, true)
	mock.OnAnything(ctx, "Get").Return("core url", true)

	suite.artrashMgr.On("Flush").Return(nil)
	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	suite.artrashMgr.On("Filter").Return([]model.ArtifactTrash{}, nil)

	mock.OnAnything(suite.projectCtl, "List").Return([]*models.Project{
		{
			ProjectID: 12345,
			Name:      "test GC",
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "List").Return([]*pkg_blob.Blob{
		{
			ID:     12345,
			Digest: "sha256:12345",
			Size:   12345,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "CleanupAssociationsForProject").Return(nil)

	gc := &GarbageCollector{
		artCtl:            suite.artifactCtl,
		artrashMgr:        suite.artrashMgr,
		projectCtl:        suite.projectCtl,
		blobMgr:           suite.blobMgr,
		registryCtlClient: suite.registryCtlClient,
	}
	params := map[string]interface{}{
		"delete_untagged": false,
		// ToDo add a redis testing pkg, we do have a 'localhost' redis server in UT
		"redis_url_reg": "redis://localhost:6379",
		"time_window":   1,
	}

	suite.Nil(gc.Run(ctx, params))
}

func (suite *gcTestSuite) TestMark() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	suite.artrashMgr.On("Flush").Return(nil)
	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	suite.artrashMgr.On("Filter").Return([]model.ArtifactTrash{
		{
			ID:                1,
			Digest:            suite.DigestString(),
			ManifestMediaType: schema2.MediaTypeManifest,
		},
	}, nil)

	mock.OnAnything(suite.projectCtl, "List").Return([]*models.Project{
		{
			ProjectID: 1234,
			Name:      "test GC",
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "List").Return([]*pkg_blob.Blob{
		{
			ID:     1234,
			Digest: suite.DigestString(),
			Size:   1234,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "CleanupAssociationsForProject").Return(nil)

	mock.OnAnything(suite.blobMgr, "UselessBlobs").Return([]*pkg_blob.Blob{
		{
			ID:          1,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeManifest,
		},
		{
			ID:          2,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeLayer,
		},
		{
			ID:          3,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeManifest,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "UpdateBlobStatus").Return(int64(1), nil)

	gc := &GarbageCollector{
		artCtl:     suite.artifactCtl,
		artrashMgr: suite.artrashMgr,
		projectCtl: suite.projectCtl,
		blobMgr:    suite.blobMgr,
	}

	suite.Nil(gc.mark(ctx))
}

func (suite *gcTestSuite) TestSweep() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	mock.OnAnything(suite.blobMgr, "UpdateBlobStatus").Return(int64(1), nil)
	mock.OnAnything(suite.blobMgr, "Delete").Return(nil)

	gc := &GarbageCollector{
		artCtl:            suite.artifactCtl,
		artrashMgr:        suite.artrashMgr,
		projectCtl:        suite.projectCtl,
		blobMgr:           suite.blobMgr,
		registryCtlClient: suite.registryCtlClient,
		deleteSet: []*pkg_blob.Blob{
			{
				ID:          1,
				Digest:      suite.DigestString(),
				ContentType: schema2.MediaTypeLayer,
			},
		},
	}

	suite.Nil(gc.sweep(ctx))
}

func TestGCTestSuite(t *testing.T) {
	os.Setenv("UTTEST", "true")
	suite.Run(t, &gcTestSuite{})
}
