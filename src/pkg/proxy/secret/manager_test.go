package secret

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManger(t *testing.T) {
	manager := GetManager()
	rn1 := "project1/golang"
	assert.False(t, manager.Verify("whatever", rn1))
	s1 := manager.Generate(rn1)
	s2 := manager.Generate(rn1)
	assert.False(t, s1 == s2)

	assert.False(t, manager.Verify(s1, "project1/donotexist"))
	assert.True(t, manager.Verify(s1, rn1))
	// A secret can be used only once.
	assert.False(t, manager.Verify(s1, rn1))
	manager2 := GetManager()
	assert.Equal(t, manager2, manager)
}

func TestExpiration(t *testing.T) {
	manager := createManager(1*time.Second, defaultCap, defaultGCInterval)
	rn1 := "project1/golang"
	s := manager.Generate(rn1)
	// Sleep till the secret expires
	time.Sleep(2 * time.Second)
	assert.False(t, manager.Verify(s, rn1))
}

func TestGC(t *testing.T) {
	manager := createManager(1*time.Second, 10, 1*time.Second).(*mgr)
	for i := 0; i < 10; i++ {
		rn := fmt.Sprintf("project%d/golang", i)
		manager.Generate(rn)
	}
	time.Sleep(2 * time.Second)
	assert.Equal(t, uint64(10), manager.size)
	for i := 0; i < 1000; i++ {
		rn := fmt.Sprintf("project%d/redis", i)
		manager.Generate(rn)
	}
	assert.Equal(t, uint64(1000), atomic.LoadUint64(&manager.size))
	time.Sleep(4 * time.Second)
	assert.Equal(t, uint64(0), atomic.LoadUint64(&manager.size))

}
