package notification

import (
	"net/http"

	commonhttp "github.com/goharbor/harbor/src/common/http"
)

const (
	secure   = "secure"
	insecure = "insecure"
)

var (
	httpHelper *HTTPHelper
)

// HTTPHelper in charge of sending notification messages to remote endpoint
type HTTPHelper struct {
	clients map[string]*http.Client
}

func init() {
	httpHelper = &HTTPHelper{
		clients: map[string]*http.Client{},
	}
	httpHelper.clients[secure] = &http.Client{
		Transport: commonhttp.GetHTTPTransport(commonhttp.SecureTransport),
	}
	httpHelper.clients[insecure] = &http.Client{
		Transport: commonhttp.GetHTTPTransport(commonhttp.InsecureTransport),
	}
}

// GetHTTPInstance ...
func GetHTTPInstance(insec bool) *http.Client {
	if insec {
		return httpHelper.clients[insecure]
	}
	return httpHelper.clients[secure]
}
