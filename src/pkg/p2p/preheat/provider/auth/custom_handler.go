package auth

import (
	"errors"
	"net/http"
	"reflect"
)

// CustomAuthHandler handle the custom auth mode.
type CustomAuthHandler struct {
	*BaseHandler
}

// Mode implements @Handler.Mode
func (c *CustomAuthHandler) Mode() string {
	return AuthModeCustom
}

// Authorize implements @Handler.Authorize
func (c *CustomAuthHandler) Authorize(req *http.Request, cred *Credential) error {
	if err := c.BaseHandler.Authorize(req, cred); err != nil {
		return err
	}

	if len(cred.Data) == 0 {
		return errors.New("missing custom token/key data")
	}

	key := reflect.ValueOf(cred.Data).MapKeys()[0].String()
	req.Header.Set(key, cred.Data[key])

	return nil
}
