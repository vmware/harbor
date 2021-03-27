package security

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	registry_token "github.com/distribution/distribution/registry/auth/token"
)

func TestGenerate(t *testing.T) {
	config.Init()
	vt := &v2Token{}
	req1, _ := http.NewRequest(http.MethodHead, "/api/2.0/", nil)
	assert.Nil(t, vt.Generate(req1))
	req2, _ := http.NewRequest(http.MethodGet, "/v2/library/ubuntu/manifests/v1.0", nil)
	req2.Header.Set("Authorization", "Bearer 123")
	assert.Nil(t, vt.Generate(req2))
	mt, err := token.MakeToken("admin", "none", []*registry_token.ResourceActions{})
	require.Nil(t, err)
	req3 := req2.Clone(req2.Context())
	req3.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mt.Token))
	assert.Nil(t, vt.Generate(req3))
	req4 := req3.Clone(req3.Context())
	mt2, err2 := token.MakeToken("admin", token.Registry, []*registry_token.ResourceActions{})
	require.Nil(t, err2)
	req4.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mt2.Token))
	assert.NotNil(t, vt.Generate(req4))
}
