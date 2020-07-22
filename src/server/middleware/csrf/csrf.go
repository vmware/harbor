package csrf

import (
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/gorilla/csrf"
)

const (
	csrfKeyEnv  = "CSRF_KEY"
	tokenHeader = "X-Harbor-CSRF-Token"
	tokenCookie = "__csrf"
)

var (
	once       sync.Once
	secureFlag = true
	protect    func(handler http.Handler) http.Handler
)

// attachToken makes sure if csrf generate a new token it will be included in the response header
func attachToken(w http.ResponseWriter, r *http.Request) {
	if t := csrf.Token(r); len(t) > 0 {
		http.SetCookie(w, &http.Cookie{
			Name:     tokenCookie,
			Secure:   secureFlag,
			Value:    t,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		log.Warningf("token not found in context, skip attaching")
	}
}

func handleError(w http.ResponseWriter, r *http.Request) {
	attachToken(w, r)
	lib_http.SendError(w, errors.New(csrf.FailureReason(r)).WithCode(errors.ForbiddenCode))
	return
}

func attach(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		attachToken(rw, req)
		handler.ServeHTTP(rw, req)
	})
}

// Middleware initialize the middleware to apply csrf selectively
func Middleware() func(handler http.Handler) http.Handler {
	once.Do(func() {
		key := os.Getenv(csrfKeyEnv)
		if len(key) != 32 {
			log.Warningf("Invalid CSRF key from environment: %s, generating random key...", key)
			key = utils.GenerateRandomString()
		}
		secureFlag = secureCookie()
		protect = csrf.Protect([]byte(key), csrf.RequestHeader(tokenHeader),
			csrf.Secure(secureFlag),
			csrf.ErrorHandler(http.HandlerFunc(handleError)),
			csrf.SameSite(csrf.SameSiteStrictMode),
			csrf.Path("/"))
	})
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		protect(attach(next)).ServeHTTP(rw, req)
	}, csrfSkipper)
}

// csrfSkipper makes sure only some of the uris accessed by non-UI client can skip the csrf check
func csrfSkipper(req *http.Request) bool {
	path := req.URL.Path
	if (strings.HasPrefix(path, "/v2/") ||
		strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/chartrepo/") ||
		strings.HasPrefix(path, "/service/")) && !lib.GetCarrySession(req.Context()) {
		return true
	}
	return false
}

func secureCookie() bool {
	ep, err := config.ExtEndpoint()
	if err != nil {
		log.Warningf("Failed to get external endpoint: %v, set cookie secure flag to true", err)
		return true
	}
	return !strings.HasPrefix(strings.ToLower(ep), "http://")
}
