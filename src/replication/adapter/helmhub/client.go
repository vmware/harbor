package helmhub

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"io/ioutil"
	"net/http"
)

// ErrHTTPNotFound defines the return error when receiving 404 response code
var ErrHTTPNotFound = errors.New("Not Found")

// Client is a client to interact with HelmHub
type Client struct {
	client *http.Client
}

// NewClient creates a new HelmHub client.
func NewClient(registry *model.Registry) *Client {
	return &Client{
		client: &http.Client{
			Transport: util.GetHTTPTransport(registry.Insecure),
		},
	}
}

// fetchCharts fetches the chart list from helm hub.
func (c *Client) fetchCharts() (*chartList, error) {
	request, err := http.NewRequest(http.MethodGet, baseURL+listCharts, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch chart list error %d: %s", resp.StatusCode, string(body))
	}

	list := &chartList{}
	err = json.Unmarshal(body, list)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chart list response error: %v", err)
	}

	return list, nil
}

// fetchChartDetail fetches the chart detail of a chart from helm hub.
func (c *Client) fetchChartDetail(chartName string) (*chartVersionList, error) {
	request, err := http.NewRequest(http.MethodGet, baseURL+listVersions(chartName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrHTTPNotFound
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch chart detail error %d: %s", resp.StatusCode, string(body))
	}

	list := &chartVersionList{}
	err = json.Unmarshal(body, list)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chart detail response error: %v", err)
	}

	return list, nil
}

func (c *Client) checkHealthy() error {
	request, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return errors.New("helm hub is unhealthy")
}

// do work as a proxy of Do function from net.http
func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
