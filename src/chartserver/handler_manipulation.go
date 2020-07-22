package chartserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication"
	rep_event "github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
	helm_repo "helm.sh/helm/v3/pkg/repo"
)

// ListCharts gets the chart list under the namespace
// See @ServiceHandler.ListCharts
func (c *Controller) ListCharts(namespace string) ([]*ChartInfo, error) {
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, errors.New("empty namespace when getting chart list")
	}

	content, err := c.apiClient.GetContent(c.APIPrefix(namespace))
	if err != nil {
		return nil, err
	}

	return c.chartOperator.GetChartList(content)
}

// GetChart returns all the chart versions under the specified chart
// See @ServiceHandler.GetChart
func (c *Controller) GetChart(namespace, chartName string) (ChartVersions, error) {
	if len(namespace) == 0 {
		return nil, errors.New("empty name when getting chart versions")
	}

	if len(chartName) == 0 {
		return nil, errors.New("no chart name specified")
	}

	url := fmt.Sprintf("%s/%s", c.APIPrefix(namespace), chartName)
	data, err := c.apiClient.GetContent(url)
	if err != nil {
		return nil, err
	}

	versions := make(ChartVersions, 0)
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// DeleteChartVersion will delete the specified version of the chart
// See @ServiceHandler.DeleteChartVersion
func (c *Controller) DeleteChartVersion(namespace, chartName, version string) error {
	if len(namespace) == 0 {
		return errors.New("empty namespace when deleting chart version")
	}

	if len(chartName) == 0 || len(version) == 0 {
		return errors.New("invalid chart for deleting")
	}

	url := fmt.Sprintf("/api/chartrepo/%s/charts/%s/%s", namespace, chartName, version)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	c.trafficProxy.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		text, err := extractError(w.Body.Bytes())
		if err != nil {
			return err
		}
		return &commonhttp.Error{
			Code:    w.Code,
			Message: text,
		}
	}

	// send notification to replication handler
	// Todo: it used as the replacement of webhook, will be removed when webhook to be introduced.
	if os.Getenv("UTTEST") != "true" {
		go func() {
			e := &rep_event.Event{
				Type: rep_event.EventTypeChartDelete,
				Resource: &model.Resource{
					Type:    model.ResourceTypeChart,
					Deleted: true,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: fmt.Sprintf("%s/%s", namespace, chartName),
						},
						Artifacts: []*model.Artifact{
							{
								Tags: []string{version},
							},
						},
					},
				},
			}
			if err := replication.EventHandler.Handle(e); err != nil {
				log.Errorf("failed to handle event: %v", err)
			}
		}()
	}

	return nil
}

// GetChartVersion returns the summary of the specified chart version.
// See @ServiceHandler.GetChartVersion
func (c *Controller) GetChartVersion(namespace, name, version string) (*helm_repo.ChartVersion, error) {
	if len(namespace) == 0 {
		return nil, errors.New("empty namespace when getting summary of chart version")
	}

	if len(name) == 0 || len(version) == 0 {
		return nil, errors.New("invalid chart when getting summary")
	}

	url := fmt.Sprintf("%s/%s/%s", c.APIPrefix(namespace), name, version)

	content, err := c.apiClient.GetContent(url)
	if err != nil {
		return nil, err
	}

	chartVersion := &helm_repo.ChartVersion{}
	if err := yaml.Unmarshal(content, chartVersion); err != nil {
		return nil, err
	}

	return chartVersion, nil
}
