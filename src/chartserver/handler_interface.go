package chartserver

import (
	"net/http"

	"helm.sh/helm/v3/cmd/helm/search"
	helm_repo "helm.sh/helm/v3/pkg/repo"
)

// ServiceHandler defines the related methods to handle kinds of chart service requests.
type ServiceHandler interface {
	// ListCharts lists all the charts under the specified namespace.
	//
	//  namespace string: the chart namespace.
	//
	//  If succeed, a chart info list with nil error will be returned;
	//  otherwise, a non-nil error will be got.
	ListCharts(namespace string) ([]*ChartInfo, error)

	// Get all the chart versions of the specified chart under the namespace.
	//
	// namespace string: the chart namespace.
	// chartName string: the name of the chart, e.g: "harbor"
	//
	// If succeed, a chart version list with nil error will be returned;
	// otherwise, a non-nil error will be got.
	GetChart(namespace, chartName string) (helm_repo.ChartVersions, error)

	// Get the detailed info of the specified chart version under the namespace.
	// The detailed info includes chart summary, dependencies, values and signature status etc.
	//
	// namespace string: the chart namespace.
	// chartName string: the name of the chart, e.g: "harbor"
	// version string: the SemVer version of the chart, e.g: "0.2.0"
	//
	// If succeed, chart version details with nil error will be returned;
	// otherwise, a non-nil error will be got.
	GetChartVersionDetails(namespace, chartName, version string) (*ChartVersionDetails, error)

	// SearchChart search charts in the specified namespaces with the keyword q.
	// RegExp mode is enabled as default.
	// For each chart, only the latest version will shown in the result list if matched to avoid duplicated entries.
	// Keep consistent with `helm search` command.
	//
	// q string            : the searching keyword
	// namespaces []string : the search namespace scope
	//
	// If succeed, a search result list with nil error will be returned;
	// otherwise, a non-nil error will be got.
	SearchChart(q string, namespaces []string) ([]*search.Result, error)

	// GetIndexFile will read the index.yaml under all namespaces and merge them as a single one
	// Please be aware that, to support this function, the backend chart repository server should
	// enable multi-tenancies
	//
	// namespaces []string : all the namespaces with accessing permissions
	//
	// If succeed, a unified merged index file with nil error will be returned;
	// otherwise, a non-nil error will be got.
	GetIndexFile(namespaces []string) (*helm_repo.IndexFile, error)

	// Get the chart summary of the specified chart version.
	//
	// namespace string: the chart namespace.
	// chartName string: the name of the chart, e.g: "harbor"
	// version string: the SemVer version of the chart, e.g: "0.2.0"
	//
	// If succeed, chart version summary with nil error will be returned;
	// otherwise, a non-nil error will be got.
	GetChartVersion(namespace, name, version string) (*helm_repo.ChartVersion, error)

	// DeleteChart deletes all the chart versions of the specified chart under the namespace.
	//
	// namespace string: the chart namespace.
	// chartName string: the name of the chart, e.g: "harbor"
	//
	// If succeed, a nil error will be returned;
	// otherwise, a non-nil error will be got.
	DeleteChart(namespace, chartName string) error

	// GetCountOfCharts calculates and returns the total count of charts under the specified namespaces.
	//
	// namespaces []string : the namespaces to count charts
	//
	// If succeed, a unsigned integer with nil error will be returned;
	// otherwise, a non-nil error will be got.
	GetCountOfCharts(namespaces []string) (uint64, error)
}

// ProxyTrafficHandler defines the handler methods to handle the proxy traffic.
type ProxyTrafficHandler interface {
	// Proxy the traffic to the backended server
	//
	// Req *http.Request     : The incoming http request
	// w http.ResponseWriter : The response writer reference
	ProxyTraffic(w http.ResponseWriter, req *http.Request)
}
