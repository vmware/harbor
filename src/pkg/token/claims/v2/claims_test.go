package v2

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/distribution/distribution/registry/auth/token"
	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	cases := []struct {
		claims Claims
		valid  bool
	}{
		{
			claims: Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer: "anonymous",
				},
				Access: []*token.ResourceActions{},
			},
			valid: false,
		},
		{
			claims: Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer: Issuer,
				},
				Access: []*token.ResourceActions{},
			},
			valid: true,
		},
	}

	for _, tc := range cases {
		if tc.valid {
			assert.Nil(t, tc.claims.Valid())
		} else {
			assert.NotNil(t, tc.claims.Valid())
		}
	}
}
