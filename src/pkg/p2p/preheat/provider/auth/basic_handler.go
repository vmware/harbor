package auth

import (
	"errors"
	"net/http"
	"reflect"
)

// BasicAuthHandler handle the basic auth mode.
type BasicAuthHandler struct {
	*BaseHandler
}

// Mode implements @Handler.Mode
func (b *BasicAuthHandler) Mode() string {
	return AuthModeBasic
}

// Authorize implements @Handler.Authorize
func (b *BasicAuthHandler) Authorize(req *http.Request, cred *Credential) error {
	if err := b.BaseHandler.Authorize(req, cred); err != nil {
		return err
	}

	if len(cred.Data) == 0 {
		return errors.New("missing username and/or password")
	}

	key := reflect.ValueOf(cred.Data).MapKeys()[0].String()
	req.SetBasicAuth(key, cred.Data[key])

	return nil
}
