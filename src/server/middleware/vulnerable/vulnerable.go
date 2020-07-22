// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vulnerable

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/controller/artifact/processor/cnab"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/util"
)

var (
	scanChecker = func() scan.Checker {
		return scan.NewChecker()
	}
)

// Middleware middleware which does the vulnerability prevention checking for the artifact in GET /v2/<name>/manifests/<reference> API
func Middleware() func(http.Handler) http.Handler {
	return middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()

		logger := log.G(ctx).WithFields(log.Fields{"middleware": "vulnerable"})

		none := lib.ArtifactInfo{}
		info := lib.GetArtifactInfo(ctx)
		if info == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}

		art, err := artifactController.GetByReference(ctx, info.Repository, info.Reference, nil)
		if err != nil {
			if !errors.IsNotFoundErr(err) {
				logger.Errorf("get artifact failed, error %v", err)
			}
			return err
		}

		proj, err := projectController.Get(ctx, art.ProjectID, project.CVEAllowlist(true))
		if err != nil {
			logger.Errorf("get the project %d failed, error: %v", art.ProjectID, err)
			return err
		}

		if !proj.VulPrevented() {
			// vulnerability prevention disabled, skip the checking
			logger.Debugf("project %s vulnerability prevention disabled, skip the checking", proj.Name)
			return nil
		}

		if util.SkipPolicyChecking(ctx, proj.ProjectID) {
			// the artifact is pulling by the scanner, skip the checking
			logger.Debugf("artifact %s@%s is pulling by the scanner, skip the checking", art.RepositoryName, art.Digest)
			return nil
		}

		checker := scanChecker()

		scannable, err := checker.IsScannable(ctx, art)
		if err != nil {
			logger.Errorf("check the scannable status of the artifact %s@%s failed, error: %v", art.RepositoryName, art.Digest, err)
			return err
		}

		if !scannable {
			// the artifact is not scannable, skip the checking
			logger.Debugf("artifact %s@%s is not scannable, skip the checking", art.RepositoryName, art.Digest)
			return nil
		}

		allowlist := report.CVESet(proj.CVEAllowlist.CVESet())
		summaries, err := scanController.GetSummary(ctx, art, []string{v1.MimeTypeNativeReport}, report.WithCVEAllowlist(&allowlist))
		if err != nil {
			logger.Errorf("get vulnerability summary of the artifact %s@%s failed, error: %v", art.RepositoryName, art.Digest, err)
			return err
		}

		projectSeverity := vuln.ParseSeverityVersion3(proj.Severity())

		rawSummary, ok := summaries[v1.MimeTypeNativeReport]
		if !ok {
			// No report yet?
			msg := fmt.Sprintf(`current image without vulnerability scanning cannot be pulled due to configured policy in 'Prevent images with vulnerability severity of "%s" or higher from running.' `+
				`To continue with pull, please contact your project administrator for help.`, projectSeverity)
			return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage(msg)
		}

		summary, ok := rawSummary.(*vuln.NativeReportSummary)
		if !ok {
			return fmt.Errorf("report summary is invalid")
		}

		if art.IsImageIndex() {
			// artifact is image index, skip the checking when it is in the allowlist
			skippingAllowlist := []string{image.ArtifactTypeImage, cnab.ArtifactTypeCNAB}
			for _, t := range skippingAllowlist {
				if art.Type == t {
					logger.Debugf("artifact %s@%s is image index and its type is %s in skipping allowlist, "+
						"skip the vulnerability prevention checking", art.RepositoryName, art.Digest, art.Type)
					return nil
				}
			}
		}

		if !summary.IsSuccessStatus() {
			msg := fmt.Sprintf(`current image with "%s" status of vulnerability scanning cannot be pulled due to configured policy in 'Prevent images with vulnerability severity of "%s" or higher from running.' `+
				`To continue with pull, please contact your project administrator for help.`, summary.ScanStatus, projectSeverity)
			return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage(msg)
		}

		if summary.Summary == nil || summary.Summary.Total == 0 {
			// No vulnerabilities found in the artifact, skip the checking
			// See https://github.com/goharbor/harbor/issues/11210 to get more details
			logger.Debugf("no vulnerabilities found in artifact %s@%s, skip the vulnerability prevention checking", art.RepositoryName, art.Digest)
			return nil
		}

		// Do judgement
		if summary.Severity.Code() >= projectSeverity.Code() {
			thing := "vulnerability"
			if summary.Summary.Total > 1 {
				thing = "vulnerabilities"
			}
			msg := fmt.Sprintf(`current image with %d %s cannot be pulled due to configured policy in 'Prevent images with vulnerability severity of "%s" or higher from running.' `+
				`To continue with pull, please contact your project administrator to exempt matched vulnerabilities through configuring the CVE allowlist.`,
				summary.Summary.Total, thing, projectSeverity)
			return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage(msg)
		}

		// Print scannerPull CVE list
		if len(summary.CVEBypassed) > 0 {
			for _, cve := range summary.CVEBypassed {
				logger.Infof("Vulnerable policy check: bypassed CVE %s", cve)
			}
		}

		return nil
	})
}
