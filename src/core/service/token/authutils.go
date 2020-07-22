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
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// Issuer is the issuer of the internal token service in Harbor for registry
	Issuer = "harbor-token-issuer"
)

var privateKey string

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
		actions := []string{}

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
func filterAccess(access []*token.ResourceActions, ctx security.Context,
	pm promgr.ProjectManager, filters map[string]accessFilter) error {
	var err error
	for _, a := range access {
		f, ok := filters[a.Type]
		if !ok {
			a.Actions = []string{}
			log.Warningf("No filter found for access type: %s, skip filter, the access of resource '%s' will be set empty.", a.Type, a.Name)
			continue
		}
		err = f.filter(ctx, pm, a)
		log.Debugf("user: %s, access: %v", ctx.GetUsername(), a)
		if err != nil {
			return err
		}
	}
	return nil
}

// MakeToken makes a valid jwt token based on parms.
func MakeToken(username, service string, access []*token.ResourceActions) (*models.Token, error) {
	pk, err := libtrust.LoadKeyFile(privateKey)
	if err != nil {
		return nil, err
	}
	expiration, err := config.TokenExpiration()
	if err != nil {
		return nil, err
	}

	tk, expiresIn, issuedAt, err := makeTokenCore(Issuer, username, service, expiration, access, pk)
	if err != nil {
		return nil, err
	}
	rs := fmt.Sprintf("%s.%s", tk.Raw, base64UrlEncode(tk.Signature))
	return &models.Token{
		Token:     rs,
		ExpiresIn: expiresIn,
		IssuedAt:  issuedAt.Format(time.RFC3339),
	}, nil
}

func permToActions(p string) []string {
	res := []string{}
	if strings.Contains(p, "W") {
		res = append(res, "push")
	}
	if strings.Contains(p, "M") {
		res = append(res, "*")
	}
	if strings.Contains(p, "R") {
		res = append(res, "pull")
	}
	if strings.Contains(p, "S") {
		res = append(res, "scanner-pull")
	}
	return res
}

// make token core
func makeTokenCore(issuer, subject, audience string, expiration int,
	access []*token.ResourceActions, signingKey libtrust.PrivateKey) (t *token.Token, expiresIn int, issuedAt *time.Time, err error) {

	joseHeader := &token.Header{
		Type:       "JWT",
		SigningAlg: "RS256",
		KeyID:      signingKey.KeyID(),
	}

	jwtID := utils.GenerateRandomStringWithLen(16)

	now := time.Now().UTC()
	issuedAt = &now
	expiresIn = expiration * 60

	claimSet := &token.ClaimSet{
		Issuer:     issuer,
		Subject:    subject,
		Audience:   audience,
		Expiration: now.Add(time.Duration(expiration) * time.Minute).Unix(),
		NotBefore:  now.Unix(),
		IssuedAt:   now.Unix(),
		JWTID:      jwtID,
		Access:     access,
	}

	var joseHeaderBytes, claimSetBytes []byte

	if joseHeaderBytes, err = json.Marshal(joseHeader); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to marshal jose header: %s", err)
	}
	if claimSetBytes, err = json.Marshal(claimSet); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to marshal claim set: %s", err)
	}

	encodedJoseHeader := base64UrlEncode(joseHeaderBytes)
	encodedClaimSet := base64UrlEncode(claimSetBytes)
	payload := fmt.Sprintf("%s.%s", encodedJoseHeader, encodedClaimSet)

	var signatureBytes []byte
	if signatureBytes, _, err = signingKey.Sign(strings.NewReader(payload), crypto.SHA256); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to sign jwt payload: %s", err)
	}

	signature := base64UrlEncode(signatureBytes)
	tokenString := fmt.Sprintf("%s.%s", payload, signature)
	t, err = token.NewToken(tokenString)
	return
}

func base64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
