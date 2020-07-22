package secret

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	defaultExpiration        = 15 * time.Second
	defaultGCInterval        = 10 * time.Second
	defaultCap        uint64 = 1024 * 1024
)

type targetRepository struct {
	name      string
	expiresAt time.Time
}

// Manager generates and verifies the secret for repositories under proxy project
// The secret normally is used for authorizing a request trying to push artifact to a project
// A secret can be used only once and expires in a short period of time.
// As the request will be sent to 127.0.0.1 so the secret will live in one process.
type Manager interface {
	// Generate generates a secret for the given repository, sample value for repository: "library/ubuntu"
	Generate(repository string) string
	// Verify verifies the secret against repo name, after the verification the secret should be invalid
	Verify(secret, repository string) bool
}

type flag struct {
	v uint32
}

func (f *flag) grab() bool {
	return atomic.CompareAndSwapUint32(&f.v, 0, 1)
}

func (f *flag) release() {
	atomic.StoreUint32(&f.v, 0)
}

type mgr struct {
	gcFlag         *flag
	gcScheduleFlag *flag
	// the minimal interval for gc
	gcInterval time.Duration
	lastGC     time.Time
	// the size of the map, it must be read and write via atomic
	size uint64
	// the capacity of the map, if the size is above the cap the gc will be triggered
	cap  uint64
	m    *sync.Map
	lock sync.Mutex
	exp  time.Duration
}

func (man *mgr) Generate(rn string) string {
	if atomic.LoadUint64(&man.size) > man.cap {
		man.gc()
	}
	s := utils.GenerateRandomStringWithLen(8)
	man.m.Store(s, targetRepository{name: rn, expiresAt: time.Now().Add(man.exp)})
	atomic.AddUint64(&man.size, 1)
	return s
}

func (man *mgr) Verify(sec, rn string) bool {
	v, ok := man.m.Load(sec)
	if !ok {
		return false
	}
	p, ok := v.(targetRepository)
	if ok && p.name == rn {
		defer man.delete(sec)
		return p.expiresAt.After(time.Now())
	}
	return false
}

func (man *mgr) delete(sec string) {
	if _, ok := man.m.Load(sec); ok {
		man.lock.Lock()
		defer man.lock.Unlock()
		if _, ok := man.m.Load(sec); ok {
			man.m.Delete(sec)
			atomic.AddUint64(&man.size, ^uint64(0))
		}

	}
}

// gc removes the expired entries so it's possible that after running gc the size is still larger than cap
// If that happens it will try to start a go routine to run another gc
func (man *mgr) gc() {
	if !man.gcFlag.grab() {
		log.Debugf("There is GC in progress, skip")
		return
	}
	defer func() {
		if atomic.LoadUint64(&man.size) > man.cap && man.gcScheduleFlag.grab() {
			log.Debugf("Size is still larger than cap, schedule a gc in next cycle")
			go func() {
				time.Sleep(man.gcInterval)
				man.gcScheduleFlag.release()
				man.gc()
			}()
		}
		man.gcFlag.release()
	}()
	if time.Now().Before(man.lastGC.Add(man.gcInterval)) {
		log.Debugf("Skip too frequent GC, last one: %v, ", man.lastGC)
		return
	}
	log.Debugf("Running GC on secret map...")
	man.m.Range(func(k, v interface{}) bool {
		repoV, ok := v.(targetRepository)
		if ok && repoV.expiresAt.Before(time.Now()) {
			log.Debugf("Removed expire secret: %s, repo: %s", k, repoV.name)
			man.delete(k.(string))
		}
		return true
	})
	man.lastGC = time.Now()
	log.Debugf("GC on secret map finished.")
}

var (
	defaultManager Manager
	once           sync.Once
)

// GetManager returns the default manager which is a singleton in the package
func GetManager() Manager {
	once.Do(func() {
		defaultManager = createManager(defaultExpiration, defaultCap, defaultGCInterval)
	})
	return defaultManager
}

func createManager(d time.Duration, c uint64, interval time.Duration) Manager {
	return &mgr{
		m:              &sync.Map{},
		exp:            d,
		cap:            c,
		gcInterval:     interval,
		gcFlag:         &flag{},
		gcScheduleFlag: &flag{},
	}
}
