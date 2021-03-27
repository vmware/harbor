package tencentcr

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/distribution/distribution/registry/client/auth/challenge"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/registry/auth/bearer"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

var (
	errInvalidTcrEndpoint    error = errors.New("[tencent-tcr.newAdapter] Invalid TCR instance endpoint")
	errPingTcrEndpointFailed error = errors.New("[tencent-tcr.newAdapter] Ping TCR instance endpoint failed")
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeTencentTcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeTencentTcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeTencentTcr)
}

type factory struct{}

/**
	* Implement Factory Interface
**/
var _ adp.Factory = &factory{}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterInfo()
}

func getAdapterInfo() *model.AdapterPattern {
	return nil
}

type adapter struct {
	*native.Adapter
	registryID *string
	regionName *string
	tcrClient  *tcr.Client
	pageSize   *int64
	client     *commonhttp.Client
	registry   *model.Registry
}

/**
	* Implement Adapter Interface
**/
var _ adp.Adapter = &adapter{}

func newAdapter(registry *model.Registry) (a *adapter, err error) {
	if !isSecretID(registry.Credential.AccessKey) {
		err = errors.New("[tencent-tcr.newAdapter] Please use SecretId/SecretKey, NOT docker login Username/Password")
		log.Debugf("[tencent-tcr.newAdapter] error=%v", err)
		return
	}

	// Query TCR instance info via endpoint.
	var registryURL *url.URL
	registryURL, _ = url.Parse(registry.URL)

	if strings.Index(registryURL.Host, ".tencentcloudcr.com") < 0 {
		log.Errorf("[tencent-tcr.newAdapter] errInvalidTcrEndpoint=%v", err)
		return nil, errInvalidTcrEndpoint
	}

	realm, service, err := ping(registry)
	log.Debugf("[tencent-tcr.newAdapter] realm=%s, service=%s error=%v", realm, service, err)
	if err != nil {
		log.Errorf("[tencent-tcr.newAdapter] ping failed. error=%v", err)
		return
	}

	// Create TCR API client
	var tcrCredential = common.NewCredential(registry.Credential.AccessKey, registry.Credential.AccessSecret)
	var cfp = profile.NewClientProfile()
	var client *tcr.Client
	// temp client used to get TCR instance info
	client, err = tcr.NewClient(tcrCredential, regions.Guangzhou, cfp)
	if err != nil {
		return
	}

	var req = tcr.NewDescribeInstancesRequest()
	req.AllRegion = common.BoolPtr(true)
	req.Filters = []*tcr.Filter{
		{
			Name:   common.StringPtr("RegistryName"),
			Values: []*string{common.StringPtr(strings.ReplaceAll(registryURL.Host, ".tencentcloudcr.com", ""))},
		},
	}
	var resp = tcr.NewDescribeInstancesResponse()
	resp, err = client.DescribeInstances(req)
	if err != nil {
		log.Errorf("DescribeInstances error=%s", err.Error())
		return
	}
	if *resp.Response.TotalCount == 0 {
		err = fmt.Errorf("[tencent-tcr.newAdapter] Can not get TCR instance info. RequestId=%s", *resp.Response.RequestId)
		return
	}
	var instanceInfo = resp.Response.Registries[0]
	log.Debugf("[tencent-tcr.InstanceInfo] registry.URL=%s, host=%s, PublicDomain=%s, RegionName=%s, RegistryId=%s",
		registry.URL, registryURL.Host, *instanceInfo.PublicDomain, *instanceInfo.RegionName, *instanceInfo.RegistryId)

	// rebuild TCR SDK client
	client, err = tcr.NewClient(tcrCredential, *instanceInfo.RegionName, cfp)
	if err != nil {
		return
	}

	var credential = NewAuth(instanceInfo.RegistryId, client)
	var transport = util.GetHTTPTransport(registry.Insecure)
	var authorizer = bearer.NewAuthorizer(realm, service, credential, transport)

	return &adapter{
		registry:   registry,
		registryID: instanceInfo.RegistryId,
		regionName: instanceInfo.RegionName,
		tcrClient:  client,
		pageSize:   common.Int64Ptr(20),
		client: commonhttp.NewClient(
			&http.Client{
				Transport: transport,
			},
			credential,
		),
		Adapter: native.NewAdapterWithAuthorizer(registry, authorizer),
	}, nil
}

func ping(registry *model.Registry) (string, string, error) {
	client := &http.Client{
		Transport: util.GetHTTPTransport(registry.Insecure),
	}

	resp, err := client.Get(registry.URL + "/v2/")
	log.Debugf("[tencent-tcr.ping] error=%v", err)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	challenges := challenge.ResponseChallenges(resp)
	for _, challenge := range challenges {
		if challenge.Scheme == "bearer" {
			return challenge.Parameters["realm"], challenge.Parameters["service"], nil
		}
	}
	return "", "", fmt.Errorf("[tencent-tcr.ping] bearer auth scheme isn't supported: %v", challenges)
}

func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
		Type: model.RegistryTypeTencentTcr,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeImage,
			model.ResourceTypeChart,
		},
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
	return
}

func (a *adapter) PrepareForPush(resources []*model.Resource) (err error) {
	log.Debugf("[tencent-tcr.PrepareForPush]")
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be null")
		}
		if resource.Metadata == nil {
			return errors.New("[tencent-tcr.PrepareForPush] the metadata of resource cannot be null")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("[tencent-tcr.PrepareForPush] the namespace of resource cannot be null")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("[tencent-tcr.PrepareForPush] the name of the namespace cannot be null")
		}
		var paths = strings.Split(resource.Metadata.Repository.Name, "/")
		var namespace = paths[0]
		var repository = path.Join(paths[1:]...)

		log.Debugf("[tencent-tcr.PrepareForPush.createPrivateNamespace] namespace=%s", namespace)
		err = a.createPrivateNamespace(namespace)
		if err != nil {
			return
		}
		log.Debugf("[tencent-tcr.PrepareForPush.createRepository] namespace=%s, repository=%s", namespace, repository)
		err = a.createRepository(namespace, repository)
		if err != nil {
			return
		}
		return
	}

	return
}
