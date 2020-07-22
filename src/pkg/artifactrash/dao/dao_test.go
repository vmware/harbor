package dao

import (
	"context"
	beegoorm "github.com/astaxie/beego/orm"
	errors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	artdao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	htesting "github.com/goharbor/harbor/src/testing"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type daoTestSuite struct {
	dao    DAO
	afDao  artdao.DAO
	id     int64
	ctx    context.Context
	digest string
	htesting.Suite
}

func (d *daoTestSuite) SetupSuite() {
	d.Suite.SetupSuite()
	d.Suite.ClearTables = []string{"artifact", "artifact_trash"}
	d.dao = New()
	d.afDao = artdao.New()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())

	d.digest = d.Suite.DigestString()
	art1 := &artdao.Artifact{
		Type:              "image",
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         10,
		RepositoryID:      10,
		Digest:            d.digest,
	}
	id, err := d.afDao.Create(d.ctx, art1)
	d.Require().Nil(err)
	err = d.afDao.Delete(d.ctx, id)
	d.Require().Nil(err)
	art2 := &artdao.Artifact{
		Type:              "image",
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         10,
		RepositoryID:      10,
		Digest:            d.Suite.DigestString(),
	}
	_, err = d.afDao.Create(d.ctx, art2)
	d.Require().Nil(err)

	aft := &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "test/hello-world",
		Digest:            d.digest,
	}
	id, err = d.dao.Create(d.ctx, aft)
	d.Require().Nil(err)
	d.id = art2.ID
}

func (d *daoTestSuite) TearDownSuite() {
	d.afDao.Delete(d.ctx, d.id)
}

func (d *daoTestSuite) TestCreate() {
	// conflict
	aft := &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "test/hello-world",
		Digest:            d.digest,
	}

	_, err := d.dao.Create(d.ctx, aft)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))
}

func (d *daoTestSuite) TestDelete() {
	err := d.dao.Delete(d.ctx, 100021)
	d.Require().NotNil(err)
	var e *errors.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(errors.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestFilter() {
	afs, err := d.dao.Filter(d.ctx, time.Now().Add(time.Second*10))
	d.Require().Nil(err)
	d.Require().Equal(afs[0].Digest, d.digest)

	// clean it in GC
	err = d.dao.Flush(d.ctx, time.Now().Add(time.Second*10))
	d.Require().Nil(err)

	// push hello-world to projecta
	digest := d.Suite.DigestString()
	art1 := &artdao.Artifact{
		Type:              "image",
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         11,
		RepositoryID:      11,
		RepositoryName:    "projectA/hello-world",
		Digest:            digest,
	}
	_, err = d.afDao.Create(d.ctx, art1)
	d.Require().Nil(err)

	// push hello-world to projectb
	art2 := &artdao.Artifact{
		Type:              "image",
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         12,
		RepositoryID:      12,
		RepositoryName:    "projectB/hello-world",
		Digest:            digest,
	}
	_, err = d.afDao.Create(d.ctx, art2)
	d.Require().Nil(err)

	// remove hello-world to projectA
	err = d.afDao.Delete(d.ctx, art1.ID)
	d.Require().Nil(err)

	aft2 := &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "projectA/hello-world",
		Digest:            digest,
	}
	_, err = d.dao.Create(d.ctx, aft2)
	d.Require().Nil(err)

	// filter results should contain projectA hello-world
	afs1, err := d.dao.Filter(d.ctx, time.Now().Add(time.Second*10))
	d.Require().Nil(err)
	d.Require().Equal(afs1[0].Digest, digest)
	d.Require().Equal(afs1[0].RepositoryName, "projectA/hello-world")

	afs1, err = d.dao.Filter(d.ctx, time.Now().Add(-1*time.Hour))
	d.Require().Nil(err)
	d.Require().Equal(0, len(afs1))

	// push hello-world again to projecta
	art3 := &artdao.Artifact{
		Type:              "image",
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         11,
		RepositoryID:      13,
		RepositoryName:    "projectA/hello-world",
		Digest:            digest,
	}
	_, err = d.afDao.Create(d.ctx, art3)
	d.Require().Nil(err)

	// filter results should contain nothing
	afs2, err := d.dao.Filter(d.ctx, time.Now())
	d.Require().Nil(err)
	d.Require().Equal(0, len(afs2))

}

func (d *daoTestSuite) TestFlush() {
	_, err := d.dao.Create(d.ctx, &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "hello-world",
		Digest:            d.Suite.DigestString(),
	})
	d.Require().Nil(err)
	_, err = d.dao.Create(d.ctx, &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "hello-world2",
		Digest:            d.Suite.DigestString(),
	})
	d.Require().Nil(err)
	_, err = d.dao.Create(d.ctx, &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "hello-world3",
		Digest:            d.Suite.DigestString(),
	})
	d.Require().Nil(err)

	err = d.dao.Flush(d.ctx, time.Now())
	d.Require().Nil(err)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
