// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package utils provides reusable and sharable utilities for other packages and components.
package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/lib/errors"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

// NodeIDContextKey is used to keep node ID in the system context
type NodeIDContextKey string

const (
	// NodeID is const of the ID context key
	NodeID NodeIDContextKey = "node_id"
)

// MakeIdentifier creates uuid for job.
func MakeIdentifier() string {
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

// IsEmptyStr check if the specified str is empty (len ==0) after triming prefix and suffix spaces.
func IsEmptyStr(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// ReadEnv return the value of env variable.
func ReadEnv(key string) string {
	return os.Getenv(key)
}

// FileExists check if the specified exists.
func FileExists(file string) bool {
	if !IsEmptyStr(file) {
		_, err := os.Stat(file)
		if err == nil {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}

		return true
	}

	return false
}

// DirExists check if the specified dir exists
func DirExists(path string) bool {
	if IsEmptyStr(path) {
		return false
	}

	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.IsDir()
}

// IsValidPort check if port is valid.
func IsValidPort(port uint) bool {
	return port != 0 && port < 65536
}

// IsValidURL validates if the url is well-formted
func IsValidURL(address string) bool {
	if IsEmptyStr(address) {
		return false
	}

	if _, err := url.Parse(address); err != nil {
		return false
	}

	return true
}

// SerializeJob encodes work.Job to json data.
func SerializeJob(job *work.Job) ([]byte, error) {
	return json.Marshal(job)
}

// DeSerializeJob decodes bytes to ptr of work.Job.
func DeSerializeJob(jobBytes []byte) (*work.Job, error) {
	var j work.Job
	err := json.Unmarshal(jobBytes, &j)

	return &j, err
}

// ResolveHostnameAndIP gets the local hostname and IP
func ResolveHostnameAndIP() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return fmt.Sprintf("%s:%s", host, ipnet.IP.String()), nil
			}
		}
	}

	return "", errors.New("failed to resolve local host&ip")
}

// GenerateNodeID returns ID of current node
func GenerateNodeID() string {
	hIP, err := ResolveHostnameAndIP()
	if err != nil {
		return MakeIdentifier()
	}

	return hIP
}
