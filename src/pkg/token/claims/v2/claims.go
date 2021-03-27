package v2

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/distribution/distribution/registry/auth/token"
)

const (
	// Issuer is the only valid issuer for jwt token sent to /v2/xxxx
	Issuer = "harbor-token-issuer"
)

// Claims represents the token claims that encapsulated in a JWT token for registry/notary resources
type Claims struct {
	jwt.StandardClaims
	Access []*token.ResourceActions `json:"access"`
}

// Valid checks if the issuer is harbor
func (c *Claims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}
	if !c.VerifyIssuer(Issuer, true) {
		return fmt.Errorf("invalid token issuer: %s", c.Issuer)
	}
	return nil
}
