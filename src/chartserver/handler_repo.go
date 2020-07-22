package chartserver

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	helm_repo "helm.sh/helm/v3/pkg/repo"

	hlog "github.com/goharbor/harbor/src/lib/log"
)

const (
	maxWorkers = 10

	// Keep consistent with 'helm search' command
	searchMaxScore = 25
)

// Result returned by worker
type processedResult struct {
	namespace       string
	indexFileOfRepo *helm_repo.IndexFile
}

// GetIndexFile will read the index.yaml under all namespaces and merge them as a single one
// Please be aware that, to support this function, the backend chart repository server should
// enable multi-tenancies
//
// See @ServiceHandler.GetIndexFile
func (c *Controller) GetIndexFile(namespaces []string) (*helm_repo.IndexFile, error) {
	if namespaces == nil || len(namespaces) == 0 {
		return emptyIndexFile(), nil
	}

	return c.getIndexYaml(namespaces)
}

// getIndexYaml will get the index yaml files for all the namespaces and merge them
// as one unified index yaml file.
func (c *Controller) getIndexYaml(namespaces []string) (*helm_repo.IndexFile, error) {
	// The final merged index file
	mergedIndexFile := &helm_repo.IndexFile{
		APIVersion: "v1",
		Entries:    make(map[string]helm_repo.ChartVersions),
		Generated:  time.Now().Round(time.Second),
		PublicKeys: []string{},
	}

	// Sync the output results from the retriever
	resultChan := make(chan *processedResult, 1)
	// Receive error
	errorChan := make(chan error, 1)
	// Signal chan for merging work
	mergeDone := make(chan struct{}, 1)
	// Total projects/namespaces
	total := len(namespaces)
	// Initialize
	initialItemCount := maxWorkers
	if total < maxWorkers {
		initialItemCount = total
	}
	// Retrieve index.yaml for repositories
	workerPool := make(chan struct{}, initialItemCount)

	// Add initial tokens to the worker
	for i := 0; i < initialItemCount; i++ {
		workerPool <- struct{}{}
	}
	// Track all the background threads
	waitGroup := new(sync.WaitGroup)

	// Start the index files merging thread
	go func() {
		defer func() {
			mergeDone <- struct{}{}
		}()

		for res := range resultChan {
			c.mergeIndexFile(res.namespace, mergedIndexFile, res.indexFileOfRepo)
		}
	}()

	// Retrieve the index files for the repositories
	// and blocking here
	var err error
LOOP:
	for _, ns := range namespaces {
		// Check if error has occurred in some goroutines
		select {
		case err = <-errorChan:
			break LOOP
		default:
			// do nothing
		}

		// Apply one token before processing
		<-workerPool

		waitGroup.Add(1)
		go func(ns string) {
			defer func() {
				waitGroup.Done() // done
				// Return the worker back to the worker
				workerPool <- struct{}{}
			}()

			indexFile, err := c.getIndexYamlWithNS(ns)
			if err != nil {
				if len(errorChan) == 0 {
					// Only need one error as failure signal
					errorChan <- err
				}
				return
			}

			// Output
			resultChan <- &processedResult{
				namespace:       ns,
				indexFileOfRepo: indexFile,
			}
		}(ns)
	}

	// Hold util all the retrieving work are done
	waitGroup.Wait()

	// close merge channel
	close(resultChan)

	// Wait until merging thread quit
	<-mergeDone

	// All the threads are done
	// Make sure error in the chan is read
	if err == nil && len(errorChan) > 0 {
		err = <-errorChan
	}

	// Met an error
	if err != nil {
		return nil, err
	}

	// Remove duplicated keys in public key list
	hash := make(map[string]string)
	for _, key := range mergedIndexFile.PublicKeys {
		hash[key] = key
	}
	mergedIndexFile.PublicKeys = []string{}
	for k := range hash {
		mergedIndexFile.PublicKeys = append(mergedIndexFile.PublicKeys, k)
	}

	return mergedIndexFile, nil
}

// Get the index yaml file under the specified namespace from the backend server
func (c *Controller) getIndexYamlWithNS(namespace string) (*helm_repo.IndexFile, error) {
	// Join url path
	url := path.Join(namespace, "index.yaml")
	url = fmt.Sprintf("%s/%s", c.backendServerAddress.String(), url)
	hlog.Debugf("Getting index.yaml from '%s'", url)

	content, err := c.apiClient.GetContent(url)
	if err != nil {
		return nil, err
	}

	// Traverse to index file object for merging
	indexFile := helm_repo.NewIndexFile()
	if err := yaml.Unmarshal(content, indexFile); err != nil {
		return nil, err
	}

	return indexFile, nil
}

// Merge the content of mergingIndexFile to the baseIndex
// The chart url should be without --chart-url prefix
func (c *Controller) mergeIndexFile(namespace string,
	baseIndex *helm_repo.IndexFile,
	mergingIndexFile *helm_repo.IndexFile) {
	// Append entries
	for chartName, chartVersions := range mergingIndexFile.Entries {
		nameWithNS := fmt.Sprintf("%s/%s", namespace, chartName)
		for _, version := range chartVersions {
			version.Name = nameWithNS
			// Currently there is only one url
			for index, url := range version.URLs {
				if !strings.HasPrefix(url, "http") {
					version.URLs[index] = path.Join(namespace, url)
				}
			}
		}

		// Appended
		baseIndex.Entries[nameWithNS] = chartVersions
	}

	// Update generated time
	if mergingIndexFile.Generated.After(baseIndex.Generated) {
		baseIndex.Generated = mergingIndexFile.Generated
	}

	// Merge public keys
	baseIndex.PublicKeys = append(baseIndex.PublicKeys, mergingIndexFile.PublicKeys...)
}

// Generate empty index file
func emptyIndexFile() *helm_repo.IndexFile {
	emptyIndexFile := &helm_repo.IndexFile{}
	emptyIndexFile.Generated = time.Now().Round(time.Second)

	return emptyIndexFile
}
