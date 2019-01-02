package api

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/label"

	"github.com/goharbor/harbor/src/chartserver"
	hlog "github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
)

const (
	namespaceParam          = ":repo"
	nameParam               = ":name"
	defaultRepo             = "library"
	rootUploadingEndpoint   = "/api/chartrepo/charts"
	rootIndexEndpoint       = "/chartrepo/index.yaml"
	chartRepoHealthEndpoint = "/api/chartrepo/health"

	accessLevelPublic = iota
	accessLevelRead
	accessLevelWrite
	accessLevelAll
	accessLevelSystem

	formFieldNameForChart = "chart"
	formFiledNameForProv  = "prov"
	headerContentType     = "Content-Type"
	contentTypeMultipart  = "multipart/form-data"
)

// chartController is a singleton instance
var chartController *chartserver.Controller

// ChartRepositoryAPI provides related API handlers for the chart repository APIs
type ChartRepositoryAPI struct {
	// The base controller to provide common utilities
	BaseController

	// For label management
	labelManager *label.BaseManager

	// Keep the namespace if existing
	namespace string
}

// Prepare something for the following actions
func (cra *ChartRepositoryAPI) Prepare() {
	// Call super prepare method
	cra.BaseController.Prepare()

	// Try to extract namespace for parameter of path
	// It may not exist
	cra.namespace = strings.TrimSpace(cra.GetStringFromPath(namespaceParam))

	// Check the existence of namespace
	// Exclude the following URI
	// -/index.yaml
	// -/api/chartserver/health
	incomingURI := cra.Ctx.Request.URL.Path
	if incomingURI == rootUploadingEndpoint {
		// Forward to the default repository
		cra.namespace = defaultRepo
	}

	if incomingURI != rootIndexEndpoint &&
		incomingURI != chartRepoHealthEndpoint {
		if !cra.requireNamespace(cra.namespace) {
			return
		}
	}

	// IFIX: As the `api` is a segment of upstream chartmuseum URL,
	// it cannot be directly passed to the chartmuseum or an 404 error will occur.
	// We'll encode it here with appending "=" prefix and suffix to avoid the potential
	// namespace conflicts. (The user inputting namespace cannot contain "=")
	cra.namespace = encodeNamespace(cra.namespace)

	// Init label manager
	cra.labelManager = &label.BaseManager{}
}

// GetHealthStatus handles GET /api/chartrepo/health
func (cra *ChartRepositoryAPI) GetHealthStatus() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelSystem) {
		return
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// GetIndexByRepo handles GET /:repo/index.yaml
func (cra *ChartRepositoryAPI) GetIndexByRepo() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// GetIndex handles GET /index.yaml
func (cra *ChartRepositoryAPI) GetIndex() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelSystem) {
		return
	}

	results, err := cra.ProjectMgr.List(nil)
	if err != nil {
		cra.SendInternalServerError(err)
		return
	}

	namespaces := []string{}
	for _, r := range results.Projects {
		namespaces = append(namespaces, r.Name)
	}

	indexFile, err := chartController.GetIndexFile(namespaces)
	if err != nil {
		cra.SendInternalServerError(err)
		return
	}

	cra.WriteYamlData(indexFile)
}

// DownloadChart handles GET /:repo/charts/:filename
func (cra *ChartRepositoryAPI) DownloadChart() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// ListCharts handles GET /api/:repo/charts
func (cra *ChartRepositoryAPI) ListCharts() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	charts, err := chartController.ListCharts(cra.namespace)
	if err != nil {
		cra.SendInternalServerError(err)
		return
	}

	cra.WriteJSONData(charts)
}

// ListChartVersions GET /api/:repo/charts/:name
func (cra *ChartRepositoryAPI) ListChartVersions() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	chartName := cra.GetStringFromPath(nameParam)

	versions, err := chartController.GetChart(cra.namespace, chartName)
	if err != nil {
		cra.SendInternalServerError(err)
		return
	}

	// Append labels
	for _, chartVersion := range versions {
		labels, err := cra.labelManager.GetLabelsOfResource(common.ResourceTypeChart, chartFullName(cra.namespace, chartVersion.Name, chartVersion.Version))
		if err != nil {
			cra.SendInternalServerError(err)
			return
		}
		chartVersion.Labels = labels
	}

	cra.WriteJSONData(versions)
}

// GetChartVersion handles GET /api/:repo/charts/:name/:version
func (cra *ChartRepositoryAPI) GetChartVersion() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelRead) {
		return
	}

	// Get other parameters
	chartName := cra.GetStringFromPath(nameParam)
	version := cra.GetStringFromPath(versionParam)

	chartVersion, err := chartController.GetChartVersionDetails(cra.namespace, chartName, version)
	if err != nil {
		cra.SendInternalServerError(err)
		return
	}

	// Append labels
	labels, err := cra.labelManager.GetLabelsOfResource(common.ResourceTypeChart, chartFullName(cra.namespace, chartName, version))
	if err != nil {
		cra.SendInternalServerError(err)
		return
	}
	chartVersion.Labels = labels

	cra.WriteJSONData(chartVersion)
}

// DeleteChartVersion handles DELETE /api/:repo/charts/:name/:version
func (cra *ChartRepositoryAPI) DeleteChartVersion() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelAll) {
		return
	}

	// Get other parameters
	chartName := cra.GetStringFromPath(nameParam)
	version := cra.GetStringFromPath(versionParam)

	// Try to remove labels from deleting chart if exitsing
	if err := cra.removeLabelsFromChart(chartName, version); err != nil {
		cra.SendInternalServerError(err)
		return
	}

	if err := chartController.DeleteChartVersion(cra.namespace, chartName, version); err != nil {
		cra.SendInternalServerError(err)
		return
	}
}

// UploadChartVersion handles POST /api/:repo/charts
func (cra *ChartRepositoryAPI) UploadChartVersion() {
	hlog.Debugf("Header of request of uploading chart: %#v, content-len=%d", cra.Ctx.Request.Header, cra.Ctx.Request.ContentLength)

	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelWrite) {
		return
	}

	// Rewrite file content if the content type is "multipart/form-data"
	if isMultipartFormData(cra.Ctx.Request) {
		formFiles := make([]formFile, 0)
		formFiles = append(formFiles,
			formFile{
				formField: formFieldNameForChart,
				mustHave:  true,
			},
			formFile{
				formField: formFiledNameForProv,
			})
		if err := cra.rewriteFileContent(formFiles, cra.Ctx.Request); err != nil {
			cra.SendInternalServerError(err)
			return
		}
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// UploadChartProvFile handles POST /api/:repo/prov
func (cra *ChartRepositoryAPI) UploadChartProvFile() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelWrite) {
		return
	}

	// Rewrite file content if the content type is "multipart/form-data"
	if isMultipartFormData(cra.Ctx.Request) {
		formFiles := make([]formFile, 0)
		formFiles = append(formFiles,
			formFile{
				formField: formFiledNameForProv,
				mustHave:  true,
			})
		if err := cra.rewriteFileContent(formFiles, cra.Ctx.Request); err != nil {
			cra.SendInternalServerError(err)
			return
		}
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// DeleteChart deletes all the chart versions of the specified chart.
func (cra *ChartRepositoryAPI) DeleteChart() {
	// Check access
	if !cra.requireAccess(cra.namespace, accessLevelWrite) {
		return
	}

	// Get other parameters from the request
	chartName := cra.GetStringFromPath(nameParam)

	// Remove labels from all the deleting chart versions under the chart
	chartVersions, err := chartController.GetChart(cra.namespace, chartName)
	if err != nil {
		cra.SendInternalServerError(err)
		return
	}

	for _, chartVersion := range chartVersions {
		if err := cra.removeLabelsFromChart(chartName, chartVersion.GetVersion()); err != nil {
			cra.SendInternalServerError(err)
			return
		}
	}

	if err := chartController.DeleteChart(cra.namespace, chartName); err != nil {
		cra.SendInternalServerError(err)
		return
	}
}

func (cra *ChartRepositoryAPI) removeLabelsFromChart(chartName, version string) error {
	// Try to remove labels from deleting chart if exitsing
	resourceID := chartFullName(cra.namespace, chartName, version)
	labels, err := cra.labelManager.GetLabelsOfResource(common.ResourceTypeChart, resourceID)
	if err == nil && len(labels) > 0 {
		for _, l := range labels {
			if err := cra.labelManager.RemoveLabelFromResource(common.ResourceTypeChart, resourceID, l.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// Check if there exists a valid namespace
// Return true if it does
// Return false if it does not
func (cra *ChartRepositoryAPI) requireNamespace(namespace string) bool {
	// Actually, never should be like this
	if len(namespace) == 0 {
		cra.SendBadRequestError(errors.New(":repo should be in the request URL"))
		return false
	}

	existsing, err := cra.ProjectMgr.Exists(namespace)
	if err != nil {
		// Check failed with error
		cra.SendInternalServerError(fmt.Errorf("failed to check existence of namespace %s with error: %s", namespace, err.Error()))
		return false
	}

	// Not existing
	if !existsing {
		cra.SendBadRequestError(fmt.Errorf("namespace %s is not existing", namespace))
		return false
	}

	return true
}

// Check if the related access match the expected requirement
// If with right access, return true
// If without right access, return false
func (cra *ChartRepositoryAPI) requireAccess(namespace string, accessLevel uint) bool {
	if accessLevel == accessLevelPublic {
		return true // do nothing
	}

	theLevel := accessLevel
	// If repo is empty, system admin role must be required
	if len(namespace) == 0 {
		theLevel = accessLevelSystem
	}

	var err error

	switch theLevel {
	// Should be system admin role
	case accessLevelSystem:
		if !cra.SecurityCtx.IsSysAdmin() {
			err = errors.New("permission denied: system admin role is required")
		}
	case accessLevelAll:
		if !cra.SecurityCtx.HasAllPerm(namespace) {
			err = errors.New("permission denied: project admin or higher role is required")
		}
	case accessLevelWrite:
		if !cra.SecurityCtx.HasWritePerm(namespace) {
			err = errors.New("permission denied: developer or higher role is required")
		}
	case accessLevelRead:
		if !cra.SecurityCtx.HasReadPerm(namespace) {
			err = errors.New("permission denied: guest or higher role is required")
		}
	default:
		// access rejected for invalid scope
		cra.SendForbiddenError(errors.New("unrecognized access scope"))
		return false
	}

	// Access is not granted, check if user has authenticated
	if err != nil {
		// Unauthenticated, return 401
		if !cra.SecurityCtx.IsAuthenticated() {
			cra.SendUnAuthorizedError(errors.New("Unauthorized"))
			return false
		}

		// Authenticated, return 403
		cra.SendForbiddenError(err)
		return false
	}

	return true
}

// formFile is used to represent the uploaded files in the form
type formFile struct {
	// form field key contains the form file
	formField string

	// flag to indicate if the file identified by the 'formField'
	// must exist
	mustHave bool
}

// If the files are uploaded with multipart/form-data mimetype, beego will extract the data
// from the request automatically. Then the request passed to the backend server with proxying
// way will have empty content.
// This method will refill the requests with file content.
func (cra *ChartRepositoryAPI) rewriteFileContent(files []formFile, request *http.Request) error {
	if len(files) == 0 {
		return nil // no files, early return
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		if err := w.Close(); err != nil {
			// Just log it
			hlog.Errorf("Failed to defer close multipart writer with error: %s", err.Error())
		}
	}()

	// Process files by key one by one
	for _, f := range files {
		mFile, mHeader, err := cra.GetFile(f.formField)
		// Handle error case by case
		if err != nil {
			formatedErr := fmt.Errorf("Get file content with multipart header from key '%s' failed with error: %s", f.formField, err.Error())
			if f.mustHave || err != http.ErrMissingFile {
				return formatedErr
			}

			// Error can be ignored, just log it
			hlog.Warning(formatedErr.Error())
			continue
		}

		fw, err := w.CreateFormFile(f.formField, mHeader.Filename)
		if err != nil {
			return fmt.Errorf("Create form file with multipart header failed with error: %s", err.Error())
		}

		_, err = io.Copy(fw, mFile)
		if err != nil {
			return fmt.Errorf("Copy file stream in multipart form data failed with error: %s", err.Error())
		}
	}

	request.Header.Set(headerContentType, w.FormDataContentType())
	request.ContentLength = -1
	request.Body = ioutil.NopCloser(&body)

	return nil
}

// Initialize the chart service controller
func initializeChartController() (*chartserver.Controller, error) {
	addr, err := config.GetChartMuseumEndpoint()
	if err != nil {
		return nil, fmt.Errorf("Failed to get the endpoint URL of chart storage server: %s", err.Error())
	}

	addr = strings.TrimSuffix(addr, "/")
	url, err := url.Parse(addr)
	if err != nil {
		return nil, errors.New("Endpoint URL of chart storage server is malformed")
	}

	controller, err := chartserver.NewController(url)
	if err != nil {
		return nil, errors.New("Failed to initialize chart API controller")
	}

	hlog.Debugf("Chart storage server is set to %s", url.String())
	hlog.Info("API controller for chart repository server is successfully initialized")

	return controller, nil
}

// Check if the request content type is "multipart/form-data"
func isMultipartFormData(req *http.Request) bool {
	return strings.Contains(req.Header.Get(headerContentType), contentTypeMultipart)
}

// Return the chart full name
func chartFullName(namespace, chartName, version string) string {
	return fmt.Sprintf("%s/%s:%s", namespace, chartName, version)
}

// Encode the namespace to avoid some name issues, e.g: namespace is 'api'
func encodeNamespace(namespace string) string {
	if strings.ToLower(namespace) == "api" {
		return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("=%s=", namespace)))
	}

	return namespace
}
