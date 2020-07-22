package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncode(t *testing.T) {
	repo := "library/ns1/busybox"
	assert.Equal(t, "library%252Fns1%252Fbusybox", Encode(repo))
}
