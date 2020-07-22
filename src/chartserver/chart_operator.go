package chartserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"

	"github.com/goharbor/harbor/src/common/models"
	hlog "github.com/goharbor/harbor/src/lib/log"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	helm_repo "helm.sh/helm/v3/pkg/repo"
)

const (
	readmeFileName = "README.md"
	valuesFileName = "values.yaml"
)

// ChartVersion extends the helm ChartVersion with additional labels
type ChartVersion struct {
	helm_repo.ChartVersion
	Labels []*models.Label `json:"labels"`
}

// ChartVersions is an array of extended ChartVersion
type ChartVersions []*ChartVersion

// ChartVersionDetails keeps the detailed data info of the chart version
type ChartVersionDetails struct {
	Metadata     *helm_repo.ChartVersion `json:"metadata"`
	Dependencies []*chart.Dependency     `json:"dependencies"`
	Values       map[string]interface{}  `json:"values"`
	Files        map[string]string       `json:"files"`
	Security     *SecurityReport         `json:"security"`
	Labels       []*models.Label         `json:"labels"`
}

// SecurityReport keeps the info related with security
// e.g.: digital signature, vulnerability scanning etc.
type SecurityReport struct {
	Signature *DigitalSignature `json:"signature"`
}

// DigitalSignature used to indicate if the chart has been signed
type DigitalSignature struct {
	Signed     bool   `json:"signed"`
	Provenance string `json:"prov_file"`
}

// ChartInfo keeps the information of the chart
type ChartInfo struct {
	Name          string    `json:"name"`
	TotalVersions uint32    `json:"total_versions"`
	LatestVersion string    `json:"latest_version"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
	Icon          string    `json:"icon"`
	Home          string    `json:"home"`
	Deprecated    bool      `json:"deprecated"`
}

// ChartOperator is designed to process the contents of
// the specified chart version to get more details
type ChartOperator struct{}

// GetChartDetails parse the details from the provided content bytes
func (cho *ChartOperator) GetChartDetails(content []byte) (*ChartVersionDetails, error) {
	chartData, err := cho.GetChartData(content)
	if err != nil {
		return nil, err
	}
	dependencies := chartData.Metadata.Dependencies
	var values map[string]interface{}
	var buf bytes.Buffer
	files := make(map[string]string)
	// Parse values
	if chartData.Values != nil {
		// values = parseRawValues([]byte(chartData.Values.GetRaw()))
		if len(chartData.Values) > 0 {
			c := chartutil.Values(chartData.Values)
			ValYaml, err := c.YAML()

			if err != nil {
				return nil, err
			}
			c.Encode(&buf)
			values = parseRawValues(buf.Bytes())
			// Append values.yaml file
			files[valuesFileName] = ValYaml
		}
	}

	// Append other files like 'README.md'
	for _, v := range chartData.Files {
		if v.Name == readmeFileName {
			files[readmeFileName] = string(v.Data)
			break
		}
	}

	theChart := &ChartVersionDetails{
		Dependencies: dependencies,
		Values:       values,
		Files:        files,
	}

	return theChart, nil
}

// GetChartList returns a reorganized chart list
func (cho *ChartOperator) GetChartList(content []byte) ([]*ChartInfo, error) {
	if content == nil || len(content) == 0 {
		return nil, errors.New("zero content")
	}

	allCharts := make(map[string]helm_repo.ChartVersions)
	if err := json.Unmarshal(content, &allCharts); err != nil {
		return nil, err
	}

	chartList := make([]*ChartInfo, 0)
	for key, chartVersions := range allCharts {
		lVersion, oVersion := getTheTwoCharts(chartVersions)
		if lVersion != nil && oVersion != nil {
			chartInfo := &ChartInfo{
				Name:          key,
				TotalVersions: uint32(len(chartVersions)),
			}
			chartInfo.Created = oVersion.Created
			chartInfo.Home = lVersion.Home
			chartInfo.Icon = lVersion.Icon
			chartInfo.Deprecated = lVersion.Deprecated
			chartInfo.LatestVersion = lVersion.Version
			chartList = append(chartList, chartInfo)
		}
	}

	// Sort the chart list by the updated time which is the create time
	// of the latest version of the chart.
	sort.Slice(chartList, func(i, j int) bool {
		if chartList[i].Updated.Equal(chartList[j].Updated) {
			return strings.Compare(chartList[i].Name, chartList[j].Name) < 0
		}

		return chartList[i].Updated.After(chartList[j].Updated)
	})

	return chartList, nil
}

// GetChartData returns raw data of chart
func (cho *ChartOperator) GetChartData(content []byte) (*chart.Chart, error) {
	if content == nil || len(content) == 0 {
		return nil, errors.New("zero content")
	}

	reader := bytes.NewReader(content)
	chartData, err := loader.LoadArchive(reader)
	if err != nil {
		return nil, err
	}

	return chartData, nil
}

// GetChartVersions returns the chart versions
func (cho *ChartOperator) GetChartVersions(content []byte) (ChartVersions, error) {
	if content == nil || len(content) == 0 {
		return nil, errors.New("zero content")
	}

	chartVersions := make(ChartVersions, 0)
	if err := json.Unmarshal(content, &chartVersions); err != nil {
		return nil, err
	}

	return chartVersions, nil
}

// Get the latest and oldest chart versions
func getTheTwoCharts(chartVersions helm_repo.ChartVersions) (latestChart *helm_repo.ChartVersion, oldestChart *helm_repo.ChartVersion) {
	if len(chartVersions) == 1 {
		return chartVersions[0], chartVersions[0]
	}

	for _, chartVersion := range chartVersions {
		currentV, err := semver.NewVersion(chartVersion.Version)
		if err != nil {
			// ignore it, just logged
			hlog.Warningf("Malformed semversion %s for the chart %s", chartVersion.Version, chartVersion.Name)
			continue
		}

		// Find latest chart
		if latestChart == nil {
			latestChart = chartVersion
		} else {
			lVersion, err := semver.NewVersion(latestChart.Version)
			if err != nil {
				// ignore it, just logged
				hlog.Warningf("Malformed semversion %s for the chart %s", latestChart.Version, chartVersion.Name)
				continue
			}
			if lVersion.LessThan(currentV) {
				latestChart = chartVersion
			}
		}

		if oldestChart == nil {
			oldestChart = chartVersion
		} else {
			if oldestChart.Created.After(chartVersion.Created) {
				oldestChart = chartVersion
			}
		}
	}

	return latestChart, oldestChart
}

// Parse the raw values to value map
func parseRawValues(rawValue []byte) map[string]interface{} {
	valueMap := make(map[string]interface{})

	if len(rawValue) == 0 {
		return valueMap
	}
	values, err := chartutil.ReadValues(rawValue)
	if err != nil || len(values) == 0 {
		return valueMap
	}

	readValue(values, "", valueMap)

	return valueMap
}

// Recursively read value
func readValue(values map[string]interface{}, keyPrefix string, valueMap map[string]interface{}) {
	for key, value := range values {
		longKey := key
		if keyPrefix != "" {
			longKey = fmt.Sprintf("%s.%s", keyPrefix, key)
		}

		if subValues, ok := value.(map[string]interface{}); ok {
			readValue(subValues, longKey, valueMap)
		} else {
			valueMap[longKey] = value
		}
	}
}
