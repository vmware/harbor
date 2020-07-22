package dao

import (
	"context"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	liborm "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/replication/dao/models"
)

// AddRegistry add a new registry
func AddRegistry(registry *models.Registry) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(registry)
}

// GetRegistry gets one registry from database by id.
func GetRegistry(id int64) (*models.Registry, error) {
	o := dao.GetOrmer()
	r := models.Registry{ID: id}
	err := o.Read(&r, "ID")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// GetRegistryByName gets one registry from database by its name.
func GetRegistryByName(name string) (*models.Registry, error) {
	o := dao.GetOrmer()
	r := models.Registry{Name: name}
	err := o.Read(&r, "Name")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// GetRegistryByURL gets one registry from database by its URL.
func GetRegistryByURL(url string) (*models.Registry, error) {
	o := dao.GetOrmer()
	r := models.Registry{URL: url}
	err := o.Read(&r, "URL")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// ListRegistries lists registries
func ListRegistries(ctx context.Context, query *q.Query) (int64, []*models.Registry, error) {
	var countQuery *q.Query
	if query != nil {
		// ignore the page number and size
		countQuery = &q.Query{
			Keywords: query.Keywords,
		}
	}
	countQs, err := liborm.QuerySetter(ctx, &models.Registry{}, countQuery)
	if err != nil {
		return 0, nil, err
	}
	count, err := countQs.Count()
	if err != nil {
		return 0, nil, err
	}

	qs, err := liborm.QuerySetter(ctx, &models.Registry{}, query)
	if err != nil {
		return 0, nil, err
	}
	var registries []*models.Registry
	_, err = qs.All(&registries)
	if err != nil {
		return 0, nil, err
	}
	if registries == nil {
		registries = []*models.Registry{}
	}
	return count, registries, nil
}

// UpdateRegistry updates one registry
func UpdateRegistry(registry *models.Registry, props ...string) error {
	o := dao.GetOrmer()

	_, err := o.Update(registry, props...)
	return err
}

// DeleteRegistry deletes a registry
func DeleteRegistry(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Registry{ID: id})
	return err
}
