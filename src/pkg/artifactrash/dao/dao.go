package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
)

// DAO is the data access object interface for artifact trash
type DAO interface {
	// Create the artifact trash
	Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error)
	// Delete the artifact trash specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Filter lists the artifact that needs to be cleaned, which creation_time must be less than or equal to the cut-off.
	Filter(ctx context.Context, cutOff time.Time) (arts []model.ArtifactTrash, err error)
	// Flush cleans the trash table record, which creation_time must be less than or equal to the cut-off.
	Flush(ctx context.Context, cutOff time.Time) (err error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create ...
func (d *dao) Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	artifactrsh.CreationTime = time.Now()
	id, err = ormer.Insert(artifactrsh)
	if err != nil {
		if e := orm.AsConflictError(err, "artifact trash %s already exists under the repository %s",
			artifactrsh.Digest, artifactrsh.RepositoryName); e != nil {
			err = e
		}
	}
	return id, err
}

// Delete ...
func (d *dao) Delete(ctx context.Context, id int64) (err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.ArtifactTrash{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("artifact trash %d not found", id)
	}
	return nil
}

// Filter the results are: all of records in artifact_trash excludes the records in artifact with same repo and digest.
func (d *dao) Filter(ctx context.Context, cutOff time.Time) (arts []model.ArtifactTrash, err error) {
	var deletedAfs []model.ArtifactTrash
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return deletedAfs, err
	}

	sql := fmt.Sprintf(`SELECT aft.* FROM artifact_trash AS aft LEFT JOIN artifact af ON (aft.repository_name=af.repository_name AND aft.digest=af.digest) WHERE (af.digest IS NULL AND af.repository_name IS NULL) AND aft.creation_time <= TO_TIMESTAMP('%f')`, float64(cutOff.UnixNano())/float64((time.Second)))

	_, err = ormer.Raw(sql).QueryRows(&deletedAfs)
	if err != nil {
		return deletedAfs, err
	}

	return deletedAfs, nil
}

// Flush delete all of items beside the one in the time window.
func (d *dao) Flush(ctx context.Context, cutOff time.Time) (err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(`DELETE FROM artifact_trash where creation_time <= TO_TIMESTAMP('%f')`, float64(cutOff.UnixNano())/float64((time.Second)))
	if err != nil {
		return err
	}
	_, err = ormer.Raw(sql).Exec()
	if err != nil {
		return err
	}
	return nil
}
