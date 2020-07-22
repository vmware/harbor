package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/controller/event/metadata"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/label"
	hlog "github.com/goharbor/harbor/src/lib/log"
	n_event "github.com/goharbor/harbor/src/pkg/notifier/event"
	rep_event "github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/server/middleware/orm"
)

const (
	namespaceParam          = ":repo"
	nameParam               = ":name"
	filenameParam           = ":filename"
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
	// chartPackageFileExtension is the file extension used for chart packages
	chartPackageFileExtension = "tgz"
)

// chartController is a singleton instance
var chartController *chartserver.Controller

// GetChartController returns the chart controller
func GetChartController() *chartserver.Controller {
	return chartController
}

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

	// Init label manager
	cra.labelManager = &label.BaseManager{}
}

func (cra *ChartRepositoryAPI) requireAccess(action rbac.Action, subresource ...rbac.Resource) bool {
	if len(subresource) == 0 {
		subresource = append(subresource, rbac.ResourceHelmChart)
	}

	return cra.RequireProjectAccess(cra.namespace, action, subresource...)
}

// GetHealthStatus handles GET /api/chartrepo/health
func (cra *ChartRepositoryAPI) GetHealthStatus() {
	// Check access
	if !cra.SecurityCtx.IsAuthenticated() {
		cra.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	if !cra.SecurityCtx.IsSysAdmin() {
		cra.SendForbiddenError(errors.New(cra.SecurityCtx.GetUsername()))
		return
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// GetIndexByRepo handles GET /:repo/index.yaml
func (cra *ChartRepositoryAPI) GetIndexByRepo() {
	// Check access
	if !cra.requireAccess(rbac.ActionRead) {
		return
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// GetIndex handles GET /index.yaml
func (cra *ChartRepositoryAPI) GetIndex() {
	// Check access
	if !cra.SecurityCtx.IsAuthenticated() {
		cra.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	if !cra.SecurityCtx.IsSysAdmin() {
		cra.SendForbiddenError(errors.New(cra.SecurityCtx.GetUsername()))
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
	if !cra.requireAccess(rbac.ActionRead) {
		return
	}

	namespace := cra.GetStringFromPath(namespaceParam)
	fileName := cra.GetStringFromPath(filenameParam)
	// Add hook event to request context
	cra.addDownloadChartEventContext(fileName, namespace, cra.Ctx.Request)

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// ListCharts handles GET /api/:repo/charts
func (cra *ChartRepositoryAPI) ListCharts() {
	// Check access
	if !cra.requireAccess(rbac.ActionList) {
		return
	}

	charts, err := chartController.ListCharts(cra.namespace)
	if err != nil {
		cra.ParseAndHandleError("fail to list charts", err)
		return
	}

	cra.WriteJSONData(charts)
}

// ListChartVersions GET /api/:repo/charts/:name
func (cra *ChartRepositoryAPI) ListChartVersions() {
	// Check access
	if !cra.requireAccess(rbac.ActionList, rbac.ResourceHelmChartVersion) {
		return
	}

	chartName := cra.GetStringFromPath(nameParam)

	versions, err := chartController.GetChart(cra.namespace, chartName)
	if err != nil {
		cra.ParseAndHandleError("fail to get chart", err)
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
	if !cra.requireAccess(rbac.ActionRead, rbac.ResourceHelmChartVersion) {
		return
	}

	// Get other parameters
	chartName := cra.GetStringFromPath(nameParam)
	version := cra.GetStringFromPath(versionParam)

	chartVersion, err := chartController.GetChartVersionDetails(cra.namespace, chartName, version)
	if err != nil {
		cra.ParseAndHandleError("fail to get chart version", err)
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
	if !cra.requireAccess(rbac.ActionDelete, rbac.ResourceHelmChartVersion) {
		return
	}

	// Get other parameters
	chartName := cra.GetStringFromPath(nameParam)
	version := cra.GetStringFromPath(versionParam)

	// Try to remove labels from deleting chart if existing
	if err := cra.removeLabelsFromChart(chartName, version); err != nil {
		cra.SendInternalServerError(err)
		return
	}

	if err := chartController.DeleteChartVersion(cra.namespace, chartName, version); err != nil {
		cra.ParseAndHandleError("fail to delete chart version", err)
		return
	}

	event := &n_event.Event{}
	metaData := &metadata.ChartDeleteMetaData{
		ChartMetaData: metadata.ChartMetaData{
			ProjectName: cra.namespace,
			ChartName:   chartName,
			Versions:    []string{version},
			OccurAt:     time.Now(),
			Operator:    cra.SecurityCtx.GetUsername(),
		},
	}
	if err := event.Build(metaData); err == nil {
		if err := event.Publish(); err != nil {
			hlog.Errorf("failed to publish chart delete event: %v", err)
		}
	} else {
		hlog.Errorf("failed to build chart delete event metadata: %v", err)
	}
}

// UploadChartVersion handles POST /api/:repo/charts
func (cra *ChartRepositoryAPI) UploadChartVersion() {
	hlog.Debugf("Header of request of uploading chart: %#v, content-len=%d", cra.Ctx.Request.Header, cra.Ctx.Request.ContentLength)

	// Check access
	if !cra.requireAccess(rbac.ActionCreate, rbac.ResourceHelmChartVersion) {
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
		if err := cra.addEventContext(formFiles, cra.Ctx.Request); err != nil {
			hlog.Errorf("Failed to add chart upload context, %v", err)
		}
	}

	// Directly proxy to the backend
	chartController.ProxyTraffic(cra.Ctx.ResponseWriter, cra.Ctx.Request)
}

// UploadChartProvFile handles POST /api/:repo/prov
func (cra *ChartRepositoryAPI) UploadChartProvFile() {
	// Check access
	if !cra.requireAccess(rbac.ActionCreate) {
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
	if !cra.requireAccess(rbac.ActionDelete) {
		return
	}

	// Get other parameters from the request
	chartName := cra.GetStringFromPath(nameParam)

	// Remove labels from all the deleting chart versions under the chart
	chartVersions, err := chartController.GetChart(cra.namespace, chartName)
	if err != nil {
		cra.ParseAndHandleError("fail to get chart", err)
		return
	}

	versions := []string{}
	for _, chartVersion := range chartVersions {
		versions = append(versions, chartVersion.Version)
		if err := cra.removeLabelsFromChart(chartName, chartVersion.Version); err != nil {
			cra.SendInternalServerError(err)
			return
		}
	}

	if err := chartController.DeleteChart(cra.namespace, chartName); err != nil {
		cra.SendInternalServerError(err)
		return
	}

	event := &n_event.Event{}
	metaData := &metadata.ChartDeleteMetaData{
		ChartMetaData: metadata.ChartMetaData{
			ProjectName: cra.namespace,
			ChartName:   chartName,
			Versions:    versions,
			OccurAt:     time.Now(),
			Operator:    cra.SecurityCtx.GetUsername(),
		},
	}
	if err := event.Build(metaData); err == nil {
		if err := event.Publish(); err != nil {
			hlog.Errorf("failed to publish chart delete event: %v", err)
		}
	} else {
		hlog.Errorf("failed to build chart delete event metadata: %v", err)
	}
}

func (cra *ChartRepositoryAPI) removeLabelsFromChart(chartName, version string) error {
	// Try to remove labels from deleting chart if existing
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

	existing, err := cra.ProjectMgr.Exists(namespace)
	if err != nil {
		// Check failed with error
		cra.SendInternalServerError(fmt.Errorf("failed to check existence of namespace %s with error: %s", namespace, err.Error()))
		return false
	}

	// Not existing
	if !existing {
		cra.handleProjectNotFound(namespace)
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

// The func is for event based chart replication policy.
// It will add a context for uploading request with key chart_upload, and consumed by upload response.
func (cra *ChartRepositoryAPI) addEventContext(files []formFile, request *http.Request) error {
	if len(files) == 0 {
		return nil
	}

	for _, f := range files {
		if f.formField == formFieldNameForChart {
			mFile, _, err := cra.GetFile(f.formField)
			if err != nil {
				hlog.Errorf("failed to read file content for upload event, %v", err)
				return err
			}
			var Buf bytes.Buffer
			_, err = io.Copy(&Buf, mFile)
			if err != nil {
				hlog.Errorf("failed to copy file content for upload event, %v", err)
				return err
			}
			chartOpr := chartserver.ChartOperator{}
			chartDetails, err := chartOpr.GetChartData(Buf.Bytes())
			if err != nil {
				hlog.Errorf("failed to get chart content for upload event, %v", err)
				return err
			}

			extInfo := make(map[string]interface{})
			extInfo["operator"] = cra.SecurityCtx.GetUsername()
			extInfo["projectName"] = cra.namespace
			extInfo["chartName"] = chartDetails.Metadata.Name

			public, err := cra.ProjectMgr.IsPublic(cra.namespace)
			if err != nil {
				hlog.Errorf("failed to check the public of project %s: %v", cra.namespace, err)
				public = false
			}
			e := &rep_event.Event{
				Type: rep_event.EventTypeChartUpload,
				Resource: &model.Resource{
					Type: model.ResourceTypeChart,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: fmt.Sprintf("%s/%s", cra.namespace, chartDetails.Metadata.Name),
							Metadata: map[string]interface{}{
								"public": strconv.FormatBool(public),
							},
						},
						Artifacts: []*model.Artifact{
							{
								Tags: []string{chartDetails.Metadata.Version},
							},
						},
					},
					ExtendedInfo: extInfo,
				},
			}
			*request = *(request.WithContext(context.WithValue(request.Context(), common.ChartUploadCtxKey, e)))
			break
		}
	}

	return nil
}

func (cra *ChartRepositoryAPI) addDownloadChartEventContext(fileName, namespace string, request *http.Request) {
	chartName, version := parseChartVersionFromFilename(fileName)
	event := &metadata.ChartDownloadMetaData{
		ChartMetaData: metadata.ChartMetaData{
			ProjectName: namespace,
			ChartName:   chartName,
			Versions:    []string{version},
			OccurAt:     time.Now(),
			Operator:    cra.SecurityCtx.GetUsername(),
		},
	}
	*request = *(request.WithContext(context.WithValue(request.Context(), common.ChartDownloadCtxKey, event)))
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

	controller, err := chartserver.NewController(url, orm.Middleware())
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
	if strings.HasPrefix(chartName, "http") {
		return fmt.Sprintf("%s:%s", chartName, version)
	}
	return fmt.Sprintf("%s/%s:%s", namespace, chartName, version)
}

// parseChartVersionFromFilename parse chart and version from file name
func parseChartVersionFromFilename(filename string) (string, string) {
	noExt := strings.TrimSuffix(path.Base(filename), fmt.Sprintf(".%s", chartPackageFileExtension))
	parts := strings.Split(noExt, "-")
	name := parts[0]
	version := ""
	for idx, part := range parts[1:] {
		if _, err := strconv.Atoi(string(part[0])); err == nil { // see if this part looks like a version (starts w int)
			version = strings.Join(parts[idx+1:], "-")
			break
		}
		name = fmt.Sprintf("%s-%s", name, part)
	}
	if version == "" { // no parts looked like a real version, just take everything after last hyphen
		lastIndex := len(parts) - 1
		name = strings.Join(parts[:lastIndex], "-")
		version = parts[lastIndex]
	}
	return name, version
}
