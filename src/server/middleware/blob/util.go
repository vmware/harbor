package blob

import (
	"fmt"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/server/middleware/requestid"
	"net/http"
	"time"
)

// probeBlob handles config/layer and manifest status in the PUT Blob & Manifest middleware, and update the status before it passed into proxy(distribution).
func probeBlob(r *http.Request, digest string) error {
	logger := log.G(r.Context())

	// digest empty is handled by the blob controller GET method
	bb, err := blobController.Get(r.Context(), digest)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return nil
		}
		return err
	}

	switch bb.Status {
	case models.StatusNone, models.StatusDelete, models.StatusDeleteFailed:
		if err := blobController.Touch(r.Context(), bb); err != nil {
			logger.Errorf("failed to update blob: %s status to StatusNone, error:%v", bb.Digest, err)
			return errors.Wrapf(err, fmt.Sprintf("the request id is: %s", r.Header.Get(requestid.HeaderXRequestID)))
		}
	case models.StatusDeleting:
		now := time.Now().UTC()
		// if the deleting exceed 2 hours, marks the blob as StatusDeleteFailed
		if now.Sub(bb.UpdateTime) > time.Duration(config.GetGCTimeWindow())*time.Hour {
			if err := blob.Ctl.Fail(r.Context(), bb); err != nil {
				log.Errorf("failed to update blob: %s status to StatusDeleteFailed, error:%v", bb.Digest, err)
				return errors.Wrapf(err, fmt.Sprintf("the request id is: %s", r.Header.Get(requestid.HeaderXRequestID)))
			}
			// StatusDeleteFailed => StatusNone, and then let the proxy to handle manifest upload
			return probeBlob(r, digest)
		}
		return errors.New(nil).WithMessage(fmt.Sprintf("the asking blob is delete failed, mark it as non existing, request id: %s", r.Header.Get(requestid.HeaderXRequestID))).WithCode(errors.NotFoundCode)
	default:
		return nil
	}
	return nil
}
