package auth

import (
	"errors"
	"net/http"
)

// NoneAuthHandler handles the case of no credentail required.
type NoneAuthHandler struct{}

// Mode implements @Handler.Mode
func (nah *NoneAuthHandler) Mode() string {
	return AuthModeNone
}

// Authorize implements @Handler.Authorize
func (nah *NoneAuthHandler) Authorize(req *http.Request, cred *Credential) error {
	if req == nil {
		return errors.New("nil request cannot be authorized")
	}

	// Do nothing
	return nil
}
