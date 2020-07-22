package huawei

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	err := adp.RegisterFactory(model.RegistryTypeHuawei, new(factory))
	if err != nil {
		log.Errorf("failed to register factory for Huawei: %v", err)
		return
	}
	log.Infof("the factory of Huawei adapter was registered")
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

// Adapter is for images replications between harbor and Huawei image repository(SWR)
type adapter struct {
	*native.Adapter
	registry *model.Registry
	client   *common_http.Client
	// original http client with no modifer,
	// huawei's some api interface with basic authorization,
	// some with bearer token authorization.
	oriClient *http.Client
}

// Info gets info about Huawei SWR
func (a *adapter) Info() (*model.RegistryInfo, error) {
	registryInfo := model.RegistryInfo{
		Type:                   model.RegistryTypeHuawei,
		Description:            "Adapter for SWR -- The image registry of Huawei Cloud",
		SupportedResourceTypes: []model.ResourceType{model.ResourceTypeImage},
		SupportedResourceFilters: []*model.FilterStyle{
			{
				Type:  model.FilterTypeName,
				Style: model.FilterStyleTypeText,
			},
			{
				Type:  model.FilterTypeTag,
				Style: model.FilterStyleTypeText,
			},
		},
		SupportedTriggers: []model.TriggerType{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}
	return &registryInfo, nil
}

// ListNamespaces lists namespaces from Huawei SWR with the provided query conditions.
func (a *adapter) ListNamespaces(query *model.NamespaceQuery) ([]*model.Namespace, error) {
	var namespaces []*model.Namespace

	urls := fmt.Sprintf("%s/dockyard/v2/visible/namespaces", a.registry.URL)

	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return namespaces, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")

	resp, err := a.client.Do(r)
	if err != nil {
		return namespaces, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return namespaces, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return namespaces, err
	}

	var namespacesData hwNamespaceList
	err = json.Unmarshal(body, &namespacesData)
	if err != nil {
		return namespaces, err
	}
	reg := fmt.Sprintf(".*%s.*", strings.Replace(query.Name, " ", "", -1))

	for _, namespaceData := range namespacesData.Namespace {
		namespace := model.Namespace{
			Name:     namespaceData.Name,
			Metadata: namespaceData.metadata(),
		}
		b, err := regexp.MatchString(reg, namespace.Name)
		if err != nil {
			return namespaces, nil
		}
		if b {
			namespaces = append(namespaces, &namespace)
		}
	}
	return namespaces, nil
}

// ConvertResourceMetadata convert resource metadata for Huawei SWR
func (a *adapter) ConvertResourceMetadata(resourceMetadata *model.ResourceMetadata, namespace *model.Namespace) (*model.ResourceMetadata, error) {
	metadata := &model.ResourceMetadata{
		Repository: resourceMetadata.Repository,
		Vtags:      resourceMetadata.Vtags,
	}
	return metadata, nil
}

// PrepareForPush prepare for push to Huawei SWR
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	namespaces := map[string]struct{}{}
	for _, resource := range resources {
		var namespace string
		paths := strings.Split(resource.Metadata.Repository.Name, "/")
		if len(paths) > 0 {
			namespace = paths[0]
		}
		ns, err := a.GetNamespace(namespace)
		if err != nil {
			return err
		}
		if ns != nil && ns.Name == namespace {
			continue
		}
		namespaces[namespace] = struct{}{}
	}

	url := fmt.Sprintf("%s/dockyard/v2/namespaces", a.registry.URL)

	for namespace := range namespaces {
		namespacebyte, err := json.Marshal(struct {
			Namespace string `json:"namespace"`
		}{
			Namespace: namespace,
		})
		if err != nil {
			return err
		}

		r, err := http.NewRequest("POST", url, strings.NewReader(string(namespacebyte)))
		if err != nil {
			return err
		}

		r.Header.Add("content-type", "application/json; charset=utf-8")

		resp, err := a.client.Do(r)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		code := resp.StatusCode
		if code >= 300 || code < 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("[%d][%s]", code, string(body))
		}

		log.Debugf("namespace %s created", namespace)
	}
	return nil
}

// GetNamespace gets a namespace from Huawei SWR
func (a *adapter) GetNamespace(namespaceStr string) (*model.Namespace, error) {
	var namespace = &model.Namespace{
		Name:     "",
		Metadata: make(map[string]interface{}),
	}

	urls := fmt.Sprintf("%s/dockyard/v2/namespaces/%s", a.registry.URL, namespaceStr)
	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return namespace, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")

	resp, err := a.client.Do(r)
	if err != nil {
		return namespace, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return namespace, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return namespace, err
	}

	var namespaceData hwNamespace
	err = json.Unmarshal(body, &namespaceData)
	if err != nil {
		return namespace, err
	}

	namespace.Name = namespaceData.Name
	namespace.Metadata = namespaceData.metadata()

	return namespace, nil
}

// HealthCheck check health for huawei SWR
func (a *adapter) HealthCheck() (model.HealthStatus, error) {
	return model.Healthy, nil
}

func newAdapter(registry *model.Registry) (adp.Adapter, error) {
	var (
		modifiers  = []modifier.Modifier{}
		authorizer modifier.Modifier
	)
	if registry.Credential != nil {
		authorizer = basic.NewAuthorizer(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret)
		modifiers = append(modifiers, authorizer)
	}

	transport := util.GetHTTPTransport(registry.Insecure)
	return &adapter{
		Adapter:  native.NewAdapter(registry),
		registry: registry,
		client: common_http.NewClient(
			&http.Client{
				Transport: transport,
			},
			modifiers...,
		),
		oriClient: &http.Client{
			Transport: transport,
		},
	}, nil

}

type hwNamespaceList struct {
	Namespace []hwNamespace `json:"namespaces"`
}

type hwNamespace struct {
	ID           int64  `json:"id" orm:"column(id)"`
	Name         string `json:"name"`
	CreatorName  string `json:"creator_name,omitempty"`
	DomainPublic int    `json:"-"`
	Auth         int    `json:"auth"`
	DomainName   string `json:"-"`
	UserCount    int64  `json:"user_count"`
	ImageCount   int64  `json:"image_count"`
}

func (ns hwNamespace) metadata() map[string]interface{} {
	var metadata = make(map[string]interface{})
	metadata["id"] = ns.ID
	metadata["creator_name"] = ns.CreatorName
	metadata["domain_public"] = ns.DomainPublic
	metadata["auth"] = ns.Auth
	metadata["domain_name"] = ns.DomainName
	metadata["user_count"] = ns.UserCount
	metadata["image_count"] = ns.ImageCount

	return metadata
}
