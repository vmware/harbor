//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	securitySecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/core/config"
	"testing"
)

func TestIsProxyProject(t *testing.T) {
	cases := []struct {
		name string
		in   *models.Project
		want bool
	}{
		{
			name: `no proxy`,
			in:   &models.Project{RegistryID: 0},
			want: false,
		},
		{
			name: `normal proxy`,
			in:   &models.Project{RegistryID: 1},
			want: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			got := isProxyProject(tt.in)

			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}
}

func TestIsProxySession(t *testing.T) {
	config.Init()
	sc1 := securitySecret.NewSecurityContext("123456789", config.SecretStore)
	otherCtx := security.NewContext(context.Background(), sc1)

	sc2 := securitySecret.NewSecurityContext(config.ProxyServiceSecret, config.SecretStore)
	proxyCtx := security.NewContext(context.Background(), sc2)
	cases := []struct {
		name string
		in   context.Context
		want bool
	}{
		{
			name: `normal`,
			in:   otherCtx,
			want: false,
		},
		{
			name: `proxy user`,
			in:   proxyCtx,
			want: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := isProxySession(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}
}
