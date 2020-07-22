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

package models

/*
// Authentication ...
type Authentication struct {
	Mode             string `json:"mode"`
	SelfRegistration bool   `json:"self_registration"`
	LDAP             *LDAP  `json:"ldap,omitempty"`
}
*/

// Database ...
type Database struct {
	Type       string      `json:"type"`
	PostGreSQL *PostGreSQL `json:"postgresql,omitempty"`
}

// MySQL ...
type MySQL struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Database string `json:"database"`
}

// SQLite ...
type SQLite struct {
	File string `json:"file"`
}

// PostGreSQL ...
type PostGreSQL struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	Database     string `json:"database"`
	SSLMode      string `json:"sslmode"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
}

// Email ...
type Email struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSL      bool   `json:"ssl"`
	Identity string `json:"identity"`
	From     string `json:"from"`
	Insecure bool   `json:"insecure"`
}

// HTTPAuthProxy wraps the settings for HTTP auth proxy
type HTTPAuthProxy struct {
	Endpoint            string `json:"endpoint"`
	TokenReviewEndpoint string `json:"tokenreivew_endpoint"`
	VerifyCert          bool   `json:"verify_cert"`
	SkipSearch          bool   `json:"skip_search"`
	ServerCertificate   string `json:"server_certificate"`
}

// OIDCSetting wraps the settings for OIDC auth endpoint
type OIDCSetting struct {
	Name         string   `json:"name"`
	Endpoint     string   `json:"endpoint"`
	VerifyCert   bool     `json:"verify_cert"`
	AutoOnboard  bool     `json:"auto_onboard"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	GroupsClaim  string   `json:"groups_claim"`
	RedirectURL  string   `json:"redirect_url"`
	Scope        []string `json:"scope"`
	UserClaim    string   `json:"user_claim"`
}

// QuotaSetting wraps the settings for Quota
type QuotaSetting struct {
	StoragePerProject int64 `json:"storage_per_project"`
}

// ConfigEntry ...
type ConfigEntry struct {
	ID    int64  `orm:"pk;auto;column(id)" json:"-"`
	Key   string `orm:"column(k)" json:"k"`
	Value string `orm:"column(v)" json:"v"`
}

// TableName ...
func (ce *ConfigEntry) TableName() string {
	return "properties"
}
