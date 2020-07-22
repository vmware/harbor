package notification

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHttpHelper(t *testing.T) {
	c1 := GetHTTPInstance(true)
	assert.NotNil(t, c1)

	c2 := GetHTTPInstance(false)
	assert.NotNil(t, c2)

	_, ok := httpHelper.clients["notExists"]
	assert.False(t, ok)
}
