package preheat

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/instance"
	policyModels "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	providerModels "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/policy"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
)

var (
	// Ctl is a global preheat controller instance
	Ctl = NewController()
)

// ErrorConflict for handling conflicts
var ErrorConflict = errors.New("resource conflict")

// ErrorUnhealthy for unhealthy
var ErrorUnhealthy = errors.New("instance unhealthy")

// Controller defines related top interfaces to handle the workflow of
// the image distribution.
type Controller interface {
	// Get all the supported distribution providers
	//
	// If succeed, an metadata of provider list will be returned.
	// Otherwise, a non nil error will be returned
	//
	GetAvailableProviders() ([]*provider.Metadata, error)

	// CountInstance all the setup instances of distribution providers
	//
	// params *q.Query : parameters for querying
	//
	// If succeed, matched provider instance count will be returned.
	// Otherwise, a non nil error will be returned
	//
	CountInstance(ctx context.Context, query *q.Query) (int64, error)

	// ListInstance all the setup instances of distribution providers
	//
	// params *q.Query : parameters for querying
	//
	// If succeed, matched provider instance list will be returned.
	// Otherwise, a non nil error will be returned
	//
	ListInstance(ctx context.Context, query *q.Query) ([]*providerModels.Instance, error)

	// GetInstance the metadata of the specified instance
	//
	// id string : ID of the instance being deleted
	//
	// If succeed, the metadata with nil error are returned
	// Otherwise, a non nil error is returned
	//
	GetInstance(ctx context.Context, id int64) (*providerModels.Instance, error)

	// GetInstance the metadata of the specified instance
	GetInstanceByName(ctx context.Context, name string) (*providerModels.Instance, error)

	// Create a new instance for the specified provider.
	//
	// If succeed, the ID of the instance will be returned.
	// Any problems met, a non nil error will be returned.
	//
	CreateInstance(ctx context.Context, instance *providerModels.Instance) (int64, error)

	// Delete the specified provider instance.
	//
	// id string : ID of the instance being deleted
	//
	// Any problems met, a non nil error will be returned.
	//
	DeleteInstance(ctx context.Context, id int64) error

	// Update the instance with incremental way;
	// Including update the enabled flag of the instance.
	//
	// id string                     : ID of the instance being updated
	// properties ...string 				 : The properties being updated
	//
	// Any problems met, a non nil error will be returned
	//
	UpdateInstance(ctx context.Context, instance *providerModels.Instance, properties ...string) error

	// do not provide another policy controller, mixed in preheat controller

	// CountPolicy returns the total count of the policy.
	CountPolicy(ctx context.Context, query *q.Query) (int64, error)
	// CreatePolicy creates the policy.
	CreatePolicy(ctx context.Context, schema *policyModels.Schema) (int64, error)
	// GetPolicy gets the policy by id.
	GetPolicy(ctx context.Context, id int64) (*policyModels.Schema, error)
	// GetPolicyByName gets the policy by name.
	GetPolicyByName(ctx context.Context, projectID int64, name string) (*policyModels.Schema, error)
	// UpdatePolicy updates the policy.
	UpdatePolicy(ctx context.Context, schema *policyModels.Schema, props ...string) error
	// DeletePolicy deletes the policy by id.
	DeletePolicy(ctx context.Context, id int64) error
	// ListPolicies lists policies by query.
	ListPolicies(ctx context.Context, query *q.Query) ([]*policyModels.Schema, error)
	// ListPoliciesByProject lists policies by project.
	ListPoliciesByProject(ctx context.Context, project int64, query *q.Query) ([]*policyModels.Schema, error)
	// CheckHealth checks the instance health, for test connection
	CheckHealth(ctx context.Context, instance *providerModels.Instance) error
}

var _ Controller = (*controller)(nil)

// controller is the default implementation of Controller interface.
//
type controller struct {
	// For instance
	iManager instance.Manager
	// For policy
	pManager policy.Manager
}

// NewController is constructor of controller
func NewController() Controller {
	return &controller{
		iManager: instance.Mgr,
		pManager: policy.Mgr,
	}
}

// GetAvailableProviders implements @Controller.GetAvailableProviders
func (c *controller) GetAvailableProviders() ([]*provider.Metadata, error) {
	return provider.ListProviders()
}

// CountInstance implements @Controller.CountInstance
func (c *controller) CountInstance(ctx context.Context, query *q.Query) (int64, error) {
	return c.iManager.Count(ctx, query)
}

// ListInstance implements @Controller.ListInstance
func (c *controller) ListInstance(ctx context.Context, query *q.Query) ([]*providerModels.Instance, error) {
	return c.iManager.List(ctx, query)
}

// CreateInstance implements @Controller.CreateInstance
func (c *controller) CreateInstance(ctx context.Context, instance *providerModels.Instance) (int64, error) {
	if instance == nil {
		return 0, errors.New("nil instance object provided")
	}

	// Avoid duplicated endpoint
	var query = &q.Query{
		Keywords: map[string]interface{}{
			"endpoint": instance.Endpoint,
		},
	}
	num, err := c.iManager.Count(ctx, query)
	if err != nil {
		return 0, err
	}
	if num > 0 {
		return 0, ErrorConflict
	}

	// !WARN: We don't check the health of the instance here.
	// That is ok because the health of instance will be checked before enforcing the policy each time.

	instance.SetupTimestamp = time.Now().Unix()

	return c.iManager.Save(ctx, instance)
}

// DeleteInstance implements @Controller.Delete
func (c *controller) DeleteInstance(ctx context.Context, id int64) error {
	return c.iManager.Delete(ctx, id)
}

// UpdateInstance implements @Controller.Update
func (c *controller) UpdateInstance(ctx context.Context, instance *providerModels.Instance, properties ...string) error {
	return c.iManager.Update(ctx, instance, properties...)
}

// GetInstance implements @Controller.Get
func (c *controller) GetInstance(ctx context.Context, id int64) (*providerModels.Instance, error) {
	return c.iManager.Get(ctx, id)
}

func (c *controller) GetInstanceByName(ctx context.Context, name string) (*providerModels.Instance, error) {
	return c.iManager.GetByName(ctx, name)
}

// CountPolicy returns the total count of the policy.
func (c *controller) CountPolicy(ctx context.Context, query *q.Query) (int64, error) {
	return c.pManager.Count(ctx, query)
}

// CreatePolicy creates the policy.
func (c *controller) CreatePolicy(ctx context.Context, schema *policyModels.Schema) (int64, error) {
	if schema != nil {
		now := time.Now()
		schema.CreatedAt = now
		schema.UpdatedTime = now
	}
	return c.pManager.Create(ctx, schema)
}

// GetPolicy gets the policy by id.
func (c *controller) GetPolicy(ctx context.Context, id int64) (*policyModels.Schema, error) {
	return c.pManager.Get(ctx, id)
}

// GetPolicyByName gets the policy by name.
func (c *controller) GetPolicyByName(ctx context.Context, projectID int64, name string) (*policyModels.Schema, error) {
	return c.pManager.GetByName(ctx, projectID, name)
}

// UpdatePolicy updates the policy.
func (c *controller) UpdatePolicy(ctx context.Context, schema *policyModels.Schema, props ...string) error {
	if schema != nil {
		schema.UpdatedTime = time.Now()
	}
	return c.pManager.Update(ctx, schema, props...)
}

// DeletePolicy deletes the policy by id.
func (c *controller) DeletePolicy(ctx context.Context, id int64) error {
	return c.pManager.Delete(ctx, id)
}

// ListPolicies lists policies by query.
func (c *controller) ListPolicies(ctx context.Context, query *q.Query) ([]*policyModels.Schema, error) {
	return c.pManager.ListPolicies(ctx, query)
}

// ListPoliciesByProject lists policies by project.
func (c *controller) ListPoliciesByProject(ctx context.Context, project int64, query *q.Query) ([]*policyModels.Schema, error) {
	return c.pManager.ListPoliciesByProject(ctx, project, query)
}

// CheckHealth checks the instance health, for test connection
func (c *controller) CheckHealth(ctx context.Context, instance *providerModels.Instance) error {
	if instance == nil {
		return errors.New("instance can not be nil")
	}

	fac, ok := provider.GetProvider(instance.Vendor)
	if !ok {
		return errors.Errorf("no driver registered for provider %s", instance.Vendor)
	}

	// Construct driver
	driver, err := fac(instance)
	if err != nil {
		return err
	}

	// Check health
	h, err := driver.GetHealth()
	if err != nil {
		return err
	}

	if h.Status != provider.DriverStatusHealthy {
		return errors.Errorf("preheat provider instance %s-%s:%s is not healthy", instance.Vendor, instance.Name, instance.Endpoint)
	}

	return nil
}
