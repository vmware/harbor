package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/distribution/distribution/registry/client/auth/challenge"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"io"
	"io/ioutil"
	"net/http"

	common_http "github.com/goharbor/harbor/src/common/http"
	"net/url"
	"reflect"
)

const (
	scheme = "bearer"
)

// Client is a client to interact with GitLab
type Client struct {
	client   *common_http.Client
	url      string
	username string
	token    string
}

// NewClient creates a new GitLab client.
func NewClient(registry *model.Registry) (*Client, error) {

	realm, _, err := ping(&http.Client{
		Transport: util.GetHTTPTransport(registry.Insecure),
	}, registry.URL)
	if err != nil {
		return nil, err
	}
	if realm == "" {
		return nil, fmt.Errorf("empty realm")
	}
	location, err := url.Parse(realm)
	if err != nil {
		return nil, err
	}
	client := &Client{
		url:      location.Scheme + "://" + location.Host,
		username: registry.Credential.AccessKey,
		token:    registry.Credential.AccessSecret,
		client: common_http.NewClient(
			&http.Client{
				Transport: util.GetHTTPTransport(registry.Insecure),
			}),
	}
	return client, nil
}

// ping returns the realm, service and error
func ping(client *http.Client, endpoint string) (string, string, error) {
	resp, err := client.Get(buildPingURL(endpoint))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	challenges := challenge.ResponseChallenges(resp)
	for _, challenge := range challenges {
		if scheme == challenge.Scheme {
			realm := challenge.Parameters["realm"]
			service := challenge.Parameters["service"]
			return realm, service, nil
		}
	}

	log.Warningf("Schemas %v are unsupported", challenges)
	return "", "", nil
}
func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}
func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)
	return req, nil
}

func (c *Client) getProjects() ([]*Project, error) {
	var projects []*Project
	urlAPI := fmt.Sprintf("%s/api/v4/projects?membership=1&per_page=50", c.url)
	if err := c.GetAndIteratePagination(urlAPI, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) getProjectsByName(name string) ([]*Project, error) {
	var projects []*Project
	urlAPI := fmt.Sprintf("%s/api/v4/projects?search=%s&membership=1&per_page=50", c.url, name)
	if err := c.GetAndIteratePagination(urlAPI, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}
func (c *Client) getRepositories(projectID int64) ([]*Repository, error) {
	var repositories []*Repository
	urlAPI := fmt.Sprintf("%s/api/v4/projects/%d/registry/repositories?per_page=50", c.url, projectID)
	if err := c.GetAndIteratePagination(urlAPI, &repositories); err != nil {
		return nil, err
	}
	return repositories, nil
}

func (c *Client) getTags(projectID int64, repositoryID int64) ([]*Tag, error) {
	var tags []*Tag
	urlAPI := fmt.Sprintf("%s/api/v4/projects/%d/registry/repositories/%d/tags?per_page=50", c.url, projectID, repositoryID)
	if err := c.GetAndIteratePagination(urlAPI, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetAndIteratePagination iterates the pagination header and returns all resources
// The parameter "v" must be a pointer to a slice
func (c *Client) GetAndIteratePagination(endpoint string, v interface{}) error {
	urlAPI, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v should be a pointer to a slice")
	}
	elemType := rv.Elem().Type()
	if elemType.Kind() != reflect.Slice {
		return errors.New("v should be a pointer to a slice")
	}

	resources := reflect.Indirect(reflect.New(elemType))
	for len(endpoint) > 0 {
		req, err := c.newRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return &common_http.Error{
				Code:    resp.StatusCode,
				Message: string(data),
			}
		}

		res := reflect.New(elemType)
		if err = json.Unmarshal(data, res.Interface()); err != nil {
			return err
		}
		resources = reflect.AppendSlice(resources, reflect.Indirect(res))
		endpoint = ""

		nextPage := resp.Header.Get("X-Next-Page")
		if len(nextPage) > 0 {
			query := urlAPI.Query()
			query.Set("page", nextPage)
			endpoint = urlAPI.Scheme + "://" + urlAPI.Host + urlAPI.Path + "?" + query.Encode()
		}
	}
	rv.Elem().Set(resources)
	return nil
}
