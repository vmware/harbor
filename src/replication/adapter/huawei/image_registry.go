package huawei

import (
	"encoding/json"
	"fmt"
	"github.com/distribution/distribution"
	"github.com/goharbor/harbor/src/replication/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// FetchArtifacts gets resources from Huawei SWR
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {

	resources := []*model.Resource{}

	urls := fmt.Sprintf("%s/dockyard/v2/repositories?filter=center::self", a.registry.URL)

	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return resources, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")

	resp, err := a.client.Do(r)
	if err != nil {
		return resources, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resources, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resources, err
	}
	repos := []hwRepoQueryResult{}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return resources, err
	}
	for _, repo := range repos {
		resource := parseRepoQueryResultToResource(repo)
		resource.Registry = a.registry
		resources = append(resources, resource)
	}
	return resources, nil

}

// ManifestExist check the manifest of Huawei SWR
func (a *adapter) ManifestExist(repository, reference string) (exist bool, desc *distribution.Descriptor, err error) {
	token, err := getJwtToken(a, repository)
	if err != nil {
		return exist, nil, err
	}

	urls := fmt.Sprintf("%s/v2/%s/manifests/%s", a.registry.URL, repository, reference)

	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return exist, nil, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")
	r.Header.Add("Authorization", "Bearer "+token.Token)

	resp, err := a.oriClient.Do(r)
	if err != nil {
		return exist, nil, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		if code == 404 {
			return false, nil, nil
		}
		body, _ := ioutil.ReadAll(resp.Body)
		return exist, nil, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return exist, nil, err
	}
	exist = true
	manifest := hwManifest{}
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		return exist, nil, err
	}
	contentType := resp.Header.Get(http.CanonicalHeaderKey("Content-Type"))
	contentLen := resp.Header.Get(http.CanonicalHeaderKey("Content-Length"))
	len, _ := strconv.Atoi(contentLen)

	return exist, &distribution.Descriptor{MediaType: contentType, Size: int64(len)}, nil
}

// DeleteManifest delete the manifest of Huawei SWR
func (a *adapter) DeleteManifest(repository, reference string) error {
	token, err := getJwtToken(a, repository)
	if err != nil {
		return err
	}

	urls := fmt.Sprintf("%s/v2/%s/manifests/%s", a.registry.URL, repository, reference)

	r, err := http.NewRequest("DELETE", urls, nil)
	if err != nil {
		return err
	}
	r.Header.Add("content-type", "application/json; charset=utf-8")
	r.Header.Add("Authorization", "Bearer "+token.Token)

	resp, err := a.oriClient.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("[%d][%s]", code, string(body))
	}

	return nil
}

func parseRepoQueryResultToResource(repo hwRepoQueryResult) *model.Resource {
	var resource model.Resource
	info := make(map[string]interface{})
	info["category"] = repo.Category
	info["description"] = repo.Description
	info["size"] = repo.Size
	info["is_public"] = repo.IsPublic
	info["num_images"] = repo.NumImages
	info["num_download"] = repo.NumDownload
	info["created_at"] = repo.CreatedAt
	info["updated_at"] = repo.UpdatedAt
	info["domain_name"] = repo.DomainName
	info["status"] = repo.Status
	info["total_range"] = repo.TotalRange

	repository := &model.Repository{
		Name:     fmt.Sprintf("%s/%s", repo.NamespaceName, repo.Name),
		Metadata: info,
	}
	resource.ExtendedInfo = info
	resource.Metadata = &model.ResourceMetadata{
		Repository: repository,
		Vtags:      repo.Tags,
	}
	resource.Deleted = false
	resource.Override = false
	resource.Type = model.ResourceTypeImage

	return &resource
}

type hwRepoQueryResult struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`

	Size         int64     `json:"size" `
	IsPublic     bool      `json:"is_public"`
	NumImages    int64     `json:"num_images"`
	NumDownload  int64     `json:"num_download"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Logo         string    `json:"logo"`
	LogoURL      string    `json:"url"`
	Path         string    `json:"path"`
	InternalPath string    `json:"internal_path"`

	DomainName    string   `json:"domain_name"`
	NamespaceName string   `json:"namespace"`
	Tags          []string `json:"tags"`
	Status        bool     `json:"status"`
	TotalRange    int64    `json:"total_range"`
}

func getJwtToken(a *adapter, repository string) (token jwtToken, err error) {
	urls := fmt.Sprintf("%s/swr/auth/v2/registry/auth?scope=repository:%s:push,pull", a.registry.URL, repository)

	r, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return token, err
	}

	r.Header.Add("content-type", "application/json; charset=utf-8")

	resp, err := a.client.Do(r)
	if err != nil {
		return token, err
	}

	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 300 || code < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return token, fmt.Errorf("[%d][%s]", code, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return token, err
	}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return token, err
	}
	return token, nil
}

type jwtToken struct {
	Token     string    `json:"token" description:"token return to user"`
	ExpiresIn int       `json:"expires_in" description:"describes token  will expires in how many seconds later"`
	IssuedAt  time.Time `json:"issued_at" description:"token issued time"`
}

type hwManifest struct {
	// SchemaVersion is the image manifest schema that this image follows
	SchemaVersion int `json:"schemaVersion"`

	// MediaType is the media type of this schema.
	MediaType string `json:"mediaType,omitempty"`

	// Config references the image configuration as a blob.
	Config hwDescriptor `json:"config"`

	// Layers lists descriptors for the layers referenced by the
	// configuration.
	Layers []hwDescriptor `json:"layers"`

	// summary keeps the summary infos
	Summary hwManifestSummary `json:"-"`
}

type hwDescriptor struct {
	// MediaType describe the type of the content. All text based formats are
	// encoded as utf-8.
	MediaType string `json:"mediaType,omitempty"`

	// Size in bytes of content.
	Size int64 `json:"size,omitempty"`

	// Digest uniquely identifies the content. A byte stream can be verified
	// against this digest.
	Digest string `json:"digest,omitempty"`

	// URLs contains the source URLs of this content.
	URLs []string `json:"urls,omitempty"`

	// depandence
	Dependence string `json:"dependence,omitempty"`
}

type hwManifestSummary struct {
	Config   string
	RepoTags []string
	Layers   []string
}
