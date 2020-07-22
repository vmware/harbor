package chartserver

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/lib/errors"
	"helm.sh/helm/v3/cmd/helm/search"

	"github.com/goharbor/harbor/src/core/config"
	hlog "github.com/goharbor/harbor/src/lib/log"
)

const (
	maxDeletionThreads = 10
)

// GetCountOfCharts calculates and returns the total count of charts under the specified namespaces.
// See @ServiceHandler.GetCountOfCharts
func (c *Controller) GetCountOfCharts(namespaces []string) (uint64, error) {
	if namespaces == nil || len(namespaces) == 0 {
		return 0, nil // Directly return 0 instead of non-nil error
	}

	indexFile, err := c.getIndexYaml(namespaces)
	if err != nil {
		return 0, err
	}

	return (uint64)(len(indexFile.Entries)), nil
}

// DeleteChart deletes all the chart versions of the specified chart under the namespace.
// See @ServiceHandler.DeleteChart
func (c *Controller) DeleteChart(namespace, chartName string) error {
	if len(strings.TrimSpace(namespace)) == 0 {
		return errors.New("empty namespace when deleting chart")
	}

	if len(strings.TrimSpace(chartName)) == 0 {
		return errors.New("empty chart name when deleting chart")
	}

	url := fmt.Sprintf("%s/%s", c.APIPrefix(namespace), chartName)

	content, err := c.apiClient.GetContent(url)
	if err != nil {
		return err
	}

	allVersions, err := c.chartOperator.GetChartVersions(content)
	if err != nil {
		return err
	}

	// Let's delete the versions in parallel
	// The number of goroutine is controlled by the const maxDeletionThreads
	qSize := len(allVersions)
	if qSize > maxDeletionThreads {
		qSize = maxDeletionThreads
	}
	tokenQueue := make(chan struct{}, qSize)
	errChan := make(chan error, 1)
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(allVersions))

	// Append initial tokens
	for i := 0; i < qSize; i++ {
		tokenQueue <- struct{}{}
	}

	// Collect errors
	errs := make([]error, 0)
	errWrapper := make(chan error, 1)
	go func() {
		defer func() {
			// pass to the out func
			if len(errs) > 0 {
				errWrapper <- fmt.Errorf("%v", errs)
			}
			close(errWrapper)
		}()

		for deletionErr := range errChan {
			errs = append(errs, deletionErr)
		}
	}()

	// Schedule deletion tasks
	for _, deletingVersion := range allVersions {
		// Apply for token first
		// If no available token, pending here
		<-tokenQueue

		// Got one token
		go func(deletingVersion *ChartVersion) {
			defer func() {
				// return the token back
				tokenQueue <- struct{}{}

				// done
				waitGroup.Done()
			}()

			if err := c.DeleteChartVersion(namespace, chartName, deletingVersion.Version); err != nil {
				errChan <- err
			}
		}(deletingVersion)
	}

	// Wait all goroutines are done
	waitGroup.Wait()
	// Safe to quit error collection goroutine
	close(errChan)

	err = <-errWrapper

	return err
}

// GetChartVersionDetails get the specified version for one chart
// This handler should return the details of the chart version,
// maybe including metadata,dependencies and values etc.
// See @ServiceHandler.GetChartVersionDetails
func (c *Controller) GetChartVersionDetails(namespace, chartName, version string) (*ChartVersionDetails, error) {
	chartV, err := c.GetChartVersion(namespace, chartName, version)
	if err != nil {
		return nil, err
	}

	// Query cache
	chartDetails := c.chartCache.GetChart(chartV.Digest)
	if chartDetails == nil {
		// NOT hit!!
		content, err := c.getChartVersionContent(namespace, chartV.URLs[0])
		if err != nil {
			return nil, err
		}

		// Process bytes and get more details of chart version
		chartDetails, err = c.chartOperator.GetChartDetails(content)
		if err != nil {
			return nil, err
		}
		chartDetails.Metadata = chartV

		// Put it into the cache for next access
		c.chartCache.PutChart(chartDetails)
	} else {
		// Just logged
		hlog.Debugf("Get detailed data from cache for chart: %s:%s (%s)",
			chartDetails.Metadata.Name,
			chartDetails.Metadata.Version,
			chartDetails.Metadata.Digest)
	}
	// The change of prov file will not cause any influence to the digest of chart,
	// and then the digital signature status should be not cached
	//
	// Generate the security report
	// prov file share same endpoint with the chart version
	// Just add .prov suffix to the chart version to form the path of prov file
	// Anyway, there will be a report about the digital signature status
	chartDetails.Security = &SecurityReport{
		Signature: &DigitalSignature{
			Signed: false,
		},
	}
	// Try to get the prov file to confirm if it is exitsing
	provFilePath := fmt.Sprintf("%s.prov", chartV.URLs[0])
	provBytes, err := c.getChartVersionContent(namespace, provFilePath)
	if err == nil && len(provBytes) > 0 {
		chartDetails.Security.Signature.Signed = true
		chartDetails.Security.Signature.Provenance = provFilePath
	} else {
		// Just log it
		hlog.Errorf("Failed to get prov file for chart %s with error: %s, got %d bytes", chartV.Name, err.Error(), len(provBytes))
	}

	return chartDetails, nil
}

// SearchChart search charts in the specified namespaces with the keyword q.
// RegExp mode is enabled as default.
// For each chart, only the latest version will shown in the result list if matched to avoid duplicated entries.
// Keep consistent with `helm search` command.
func (c *Controller) SearchChart(q string, namespaces []string) ([]*search.Result, error) {
	if len(q) == 0 || len(namespaces) == 0 {
		// Return empty list
		return []*search.Result{}, nil
	}

	// Get the merged index yaml file of the namespaces
	ind, err := c.getIndexYaml(namespaces)
	if err != nil {
		return nil, err
	}

	// Build the search index
	index := search.NewIndex()
	// As the repo name is already merged into the index yaml, we use empty repo name.
	// Set 'All' to false to return only one version for each chart.
	index.AddRepo("", ind, false)

	// Search
	// RegExp is enabled
	results, err := index.Search(q, searchMaxScore, true)
	if err != nil {
		return nil, err
	}

	// Sort results.
	search.SortScore(results)

	return results, nil
}

// Get the content bytes of the chart version
func (c *Controller) getChartVersionContent(namespace string, subPath string) ([]byte, error) {
	var url string
	if strings.HasPrefix(subPath, "http") {
		extEndpoint, err := config.ExtEndpoint()
		if err != nil {
			return nil, errors.Wrap(err, "can not get ext endpoint")
		}
		url = strings.TrimPrefix(subPath, fmt.Sprintf("%s/%s", extEndpoint, "chartrepo/"))
	} else {
		url = path.Join(namespace, subPath)
	}
	url = fmt.Sprintf("%s/%s", c.backendServerAddress.String(), url)
	return c.apiClient.GetContent(url)
}
