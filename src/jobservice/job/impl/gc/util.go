package gc

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/gomodule/redigo/redis"
)

// delKeys ...
func delKeys(con redis.Conn, pattern string) error {
	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(con.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving '%s' keys", pattern)
		}
		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return fmt.Errorf("unexpected type for Int, got type %T", err)
		}
		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return fmt.Errorf("converts an array command reply to a []string %v", err)
		}
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}
	for _, key := range keys {
		_, err := con.Do("DEL", key)
		if err != nil {
			return fmt.Errorf("failed to clean registry cache %v", err)
		}
	}
	return nil
}

// v2DeleteManifest calls the registry API to remove manifest
func v2DeleteManifest(repository, digest string) error {
	exist, _, err := registry.Cli.ManifestExist(repository, digest)
	if err != nil {
		return err
	}
	// it could be happened at remove manifest success but fail to delete harbor DB.
	// when the GC job executes again, the manifest should not exist.
	if !exist {
		return nil
	}
	if err := registry.Cli.DeleteManifest(repository, digest); err != nil {
		return err
	}
	return nil
}

// ignoreNotFound ignores the NotFoundErr error
func ignoreNotFound(f func() error) error {
	if err := f(); err != nil && !errors.IsNotFoundErr(err) {
		return err
	}
	return nil
}
