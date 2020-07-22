package auth

import (
	"errors"
	"fmt"
	"net/http"
)

// TokenAuthHandler handles the OAuth auth mode.
type TokenAuthHandler struct {
	*BaseHandler
}

// Mode implements @Handler.Mode
func (t *TokenAuthHandler) Mode() string {
	return AuthModeOAuth
}

// Authorize implements @Handler.Authorize
func (t *TokenAuthHandler) Authorize(req *http.Request, cred *Credential) error {
	if err := t.BaseHandler.Authorize(req, cred); err != nil {
		return err
	}

	if _, ok := cred.Data["token"]; !ok {
		return errors.New("missing OAuth token")
	}

	authData := fmt.Sprintf("%s %s", "Bearer", cred.Data["token"])
	req.Header.Set("Authorization", authData)

	return nil
}
