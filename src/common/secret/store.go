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

package secret

const (
	// JobserviceUser is the name of jobservice user
	JobserviceUser = "harbor-jobservice"
	// ProxyserviceUser is the name of proxyservice user
	ProxyserviceUser = "harbor-proxyservice"
	// CoreUser is the name of ui user
	CoreUser = "harbor-core"
)

// Store the secrets and provides methods to validate secrets
type Store struct {
	// the key is secret
	// the value is username
	secrets map[string]string
}

// NewStore ...
func NewStore(secrets map[string]string) *Store {
	return &Store{
		secrets: secrets,
	}
}

// IsValid returns whether the secret is valid
func (s *Store) IsValid(secret string) bool {
	return len(s.GetUsername(secret)) != 0
}

// GetUsername returns the corresponding username of the secret
func (s *Store) GetUsername(secret string) string {
	return s.secrets[secret]
}
