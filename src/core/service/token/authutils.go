// Copyright 2018 Project Harbor Authors
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

package token

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/distribution/distribution/registry/auth/token"
	"github.com/docker/libtrust"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	tokenpkg "github.com/goharbor/harbor/src/pkg/token"
	v2 "github.com/goharbor/harbor/src/pkg/token/claims/v2"
)

const (
	signingMethod = "RS256"
)

var (
	privateKey string
)

func init() {
	privateKey = config.TokenPrivateKeyPath()
}

// GetResourceActions ...
func GetResourceActions(scopes []string) []*token.ResourceActions {
	log.Debugf("scopes: %+v", scopes)
	var res []*token.ResourceActions
	for _, s := range scopes {
		if s == "" {
			continue
		}
		items := strings.Split(s, ":")
		length := len(items)

		typee := ""
		name := ""
		actions := make([]string, 0)

		if length == 1 {
			typee = items[0]
		} else if length == 2 {
			typee = items[0]
			name = items[1]
		} else {
			typee = items[0]
			name = strings.Join(items[1:length-1], ":")
			if len(items[length-1]) > 0 {
				actions = strings.Split(items[length-1], ",")
			}
		}

		res = append(res, &token.ResourceActions{
			Type:    typee,
			Name:    name,
			Actions: actions,
		})
	}
	return res
}

// filterAccess iterate a list of resource actions and try to use the filter that matches the resource type to filter the actions.
func filterAccess(ctx context.Context, access []*token.ResourceActions,
	ctl project.Controller, filters map[string]accessFilter) error {
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to  get security context from request")
	}
	var err error
	for _, a := range access {
		f, ok := filters[a.Type]
		if !ok {
			a.Actions = []string{}
			log.Warningf("No filter found for access type: %s, skip filter, the access of resource '%s' will be set empty.", a.Type, a.Name)
			continue
		}
		err = f.filter(ctx, ctl, a)
		log.Debugf("user: %s, access: %v", secCtx.GetUsername(), a)
		if err != nil {
			log.Errorf("Failed to handle the resource %s:%s, due to error %v, returning empty access for it.",
				a.Type, a.Name, err)
			a.Actions = []string{}
		}
	}
	return nil
}

// MakeToken makes a valid jwt token based on parms.
func MakeToken(username, service string, access []*token.ResourceActions) (*models.Token, error) {
	options, err := tokenpkg.NewOptions(signingMethod, v2.Issuer, privateKey)
	if err != nil {
		return nil, err
	}
	expiration, err := config.TokenExpiration()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	claims := &v2.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    options.Issuer,
			Subject:   username,
			Audience:  service,
			ExpiresAt: now.Add(time.Duration(expiration) * time.Minute).Unix(),
			NotBefore: now.Unix(),
			IssuedAt:  now.Unix(),
			Id:        utils.GenerateRandomStringWithLen(16),
		},
		Access: access,
	}
	tok, err := tokenpkg.New(options, claims)
	if err != nil {
		return nil, err
	}
	// Add kid to token header for compatibility with docker distribution's code
	// see https://github.com/distribution/distribution/blob/release/2.7/registry/auth/token/token.go#L197
	k, err := libtrust.UnmarshalPrivateKeyPEM(options.PrivateKey)
	if err != nil {
		return nil, err
	}
	tok.Header["kid"] = k.KeyID()

	rawToken, err := tok.Raw()
	if err != nil {
		return nil, err
	}
	return &models.Token{
		Token:     rawToken,
		ExpiresIn: expiration * 60,
		IssuedAt:  now.Format(time.RFC3339),
	}, nil
}
