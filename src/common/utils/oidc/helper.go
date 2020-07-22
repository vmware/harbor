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

package oidc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gooidc "github.com/coreos/go-oidc"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"golang.org/x/oauth2"
)

const (
	googleEndpoint = "https://accounts.google.com"
)

type claimsProvider interface {
	Claims(v interface{}) error
}

type providerHelper struct {
	sync.Mutex
	instance     atomic.Value
	setting      atomic.Value
	creationTime time.Time
}

func (p *providerHelper) get() (*gooidc.Provider, error) {
	if p.instance.Load() != nil {
		if time.Now().Sub(p.creationTime) > 3*time.Second {
			if err := p.create(); err != nil {
				return nil, err
			}
		}
	} else {
		p.Lock()
		defer p.Unlock()
		if p.instance.Load() == nil {
			if err := p.reloadSetting(); err != nil {
				return nil, err
			}
			if err := p.create(); err != nil {
				return nil, err
			}
			go func() {
				for {
					if err := p.reloadSetting(); err != nil {
						log.Warningf("Failed to refresh configuration, error: %v", err)
					}
					time.Sleep(3 * time.Second)
				}
			}()
		}
	}

	return p.instance.Load().(*gooidc.Provider), nil
}

func (p *providerHelper) reloadSetting() error {
	conf, err := config.OIDCSetting()
	if err != nil {
		return fmt.Errorf("failed to load OIDC setting: %v", err)
	}
	p.setting.Store(*conf)
	return nil
}

func (p *providerHelper) create() error {
	if p.setting.Load() == nil {
		return errors.New("the configuration is not loaded")
	}
	s := p.setting.Load().(models.OIDCSetting)
	ctx := clientCtx(context.Background(), s.VerifyCert)
	provider, err := gooidc.NewProvider(ctx, s.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider, error: %v", err)
	}
	p.instance.Store(provider)
	p.creationTime = time.Now()
	return nil
}

var provider = &providerHelper{}

var insecureTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}

// Token wraps the attributes of a oauth2 token plus the attribute of ID token
type Token struct {
	oauth2.Token
	RawIDToken string `json:"id_token,omitempty"`
}

// UserInfo wraps the information that is extracted via token.  It will be transformed to data object that is persisted
// in the DB
type UserInfo struct {
	Issuer        string   `json:"iss"`
	Subject       string   `json:"sub"`
	Username      string   `json:"name"`
	Email         string   `json:"email"`
	Groups        []string `json:"groups"`
	hasGroupClaim bool
}

func getOauthConf() (*oauth2.Config, error) {
	p, err := provider.get()
	if err != nil {
		return nil, err
	}
	setting := provider.setting.Load().(models.OIDCSetting)
	scopes := make([]string, 0)
	for _, sc := range setting.Scope {
		if strings.HasPrefix(p.Endpoint().AuthURL, googleEndpoint) && sc == gooidc.ScopeOfflineAccess {
			log.Warningf("Dropped unsupported scope: %s ", sc)
			continue
		}
		scopes = append(scopes, sc)
	}
	return &oauth2.Config{
		ClientID:     setting.ClientID,
		ClientSecret: setting.ClientSecret,
		Scopes:       scopes,
		RedirectURL:  setting.RedirectURL,
		Endpoint:     p.Endpoint(),
	}, nil
}

// AuthCodeURL returns the URL for OIDC provider's consent page.  The state should be verified when user is redirected
// back to Harbor.
func AuthCodeURL(state string) (string, error) {
	conf, err := getOauthConf()
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return "", err
	}
	if strings.HasPrefix(conf.Endpoint.AuthURL, googleEndpoint) { // make sure the refresh token will be returned
		return conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), nil
	}
	return conf.AuthCodeURL(state), nil
}

// ExchangeToken get the token from token provider via the code
func ExchangeToken(ctx context.Context, code string) (*Token, error) {
	oauth, err := getOauthConf()
	if err != nil {
		log.Errorf("Failed to get OAuth configuration, error: %v", err)
		return nil, err
	}
	setting := provider.setting.Load().(models.OIDCSetting)
	ctx = clientCtx(ctx, setting.VerifyCert)
	oauthToken, err := oauth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return &Token{Token: *oauthToken, RawIDToken: oauthToken.Extra("id_token").(string)}, nil
}

func parseIDToken(ctx context.Context, rawIDToken string) (*gooidc.IDToken, error) {
	conf := &gooidc.Config{SkipClientIDCheck: true, SkipExpiryCheck: true}
	return verifyTokenWithConfig(ctx, rawIDToken, conf)
}

// VerifyToken verifies the ID token based on the OIDC settings
func VerifyToken(ctx context.Context, rawIDToken string) (*gooidc.IDToken, error) {
	return verifyTokenWithConfig(ctx, rawIDToken, nil)
}

func verifyTokenWithConfig(ctx context.Context, rawIDToken string, conf *gooidc.Config) (*gooidc.IDToken, error) {
	log.Debugf("Raw ID token for verification: %s", rawIDToken)
	p, err := provider.get()
	if err != nil {
		return nil, err
	}
	settings := provider.setting.Load().(models.OIDCSetting)
	if conf == nil {
		conf = &gooidc.Config{ClientID: settings.ClientID}
	}
	verifier := p.Verifier(conf)
	ctx = clientCtx(ctx, settings.VerifyCert)
	return verifier.Verify(ctx, rawIDToken)
}

func clientCtx(ctx context.Context, verifyCert bool) context.Context {
	var client *http.Client
	if !verifyCert {
		client = &http.Client{
			Transport: insecureTransport,
		}
	} else {
		client = &http.Client{}
	}
	return gooidc.ClientContext(ctx, client)
}

// refreshToken tries to refresh the token if it's expired, if it doesn't the
// original one will be returned.
func refreshToken(ctx context.Context, token *Token) (*Token, error) {
	oauthCfg, err := getOauthConf()
	if err != nil {
		return nil, err
	}
	setting := provider.setting.Load().(models.OIDCSetting)
	cctx := clientCtx(ctx, setting.VerifyCert)
	ts := oauthCfg.TokenSource(cctx, &token.Token)
	nt, err := ts.Token()
	if err != nil {
		return nil, err
	}
	it, ok := nt.Extra("id_token").(string)
	if !ok {
		log.Debug("id_token not exist in refresh response")
	}
	return &Token{Token: *nt, RawIDToken: it}, nil
}

// UserInfoFromToken tries to call the UserInfo endpoint of the OIDC provider, and consolidate with ID token
// to generate a UserInfo object, if the ID token is not in the input token struct, some attributes will be empty
func UserInfoFromToken(ctx context.Context, token *Token) (*UserInfo, error) {
	setting := provider.setting.Load().(models.OIDCSetting)
	local, err := userInfoFromIDToken(ctx, token, setting)
	if err != nil {
		return nil, err
	}
	remote, err := userInfoFromRemote(ctx, token, setting)
	if err != nil {
		log.Warningf("Failed to get userInfo by calling remote userinfo endpoint, error: %v ", err)
	}
	if remote != nil && local != nil {
		if remote.Subject != local.Subject {
			return nil, fmt.Errorf("the subject from userinfo: %s does not match the subject from ID token: %s, probably a security attack happened", remote.Subject, local.Subject)
		}
		return mergeUserInfo(remote, local), nil
	} else if remote != nil && local == nil {
		return remote, nil
	} else if local != nil && remote == nil {
		log.Debugf("Fall back to user data from ID token.")
		return local, nil
	}
	return nil, fmt.Errorf("failed to get userinfo from both remote and ID token")
}

func mergeUserInfo(remote, local *UserInfo) *UserInfo {
	res := &UserInfo{
		// data only contained in ID token
		Subject: local.Subject,
		Issuer:  local.Issuer,
		// Used data from userinfo
		Username: remote.Username,
		Email:    remote.Email,
	}
	if remote.hasGroupClaim {
		res.Groups = remote.Groups
		res.hasGroupClaim = true
	} else if local.hasGroupClaim {
		res.Groups = local.Groups
		res.hasGroupClaim = true
	} else {
		res.Groups = []string{}
	}
	return res
}

func userInfoFromRemote(ctx context.Context, token *Token, setting models.OIDCSetting) (*UserInfo, error) {
	p, err := provider.get()
	if err != nil {
		return nil, err
	}
	cctx := clientCtx(ctx, setting.VerifyCert)
	u, err := p.UserInfo(cctx, oauth2.StaticTokenSource(&token.Token))
	if err != nil {
		return nil, err
	}
	return userInfoFromClaims(u, setting.GroupsClaim, setting.UserClaim)
}

func userInfoFromIDToken(ctx context.Context, token *Token, setting models.OIDCSetting) (*UserInfo, error) {
	if token.RawIDToken == "" {
		return nil, nil
	}
	idt, err := parseIDToken(ctx, token.RawIDToken)
	if err != nil {
		return nil, err
	}

	return userInfoFromClaims(idt, setting.GroupsClaim, setting.UserClaim)
}

func userInfoFromClaims(c claimsProvider, g, u string) (*UserInfo, error) {
	res := &UserInfo{}
	if err := c.Claims(res); err != nil {
		return nil, err
	}
	if u != "" {
		allClaims := make(map[string]interface{})
		if err := c.Claims(&allClaims); err != nil {
			return nil, err
		}

		username, ok := allClaims[u].(string)
		if !ok {
			return nil, fmt.Errorf("OIDC. Failed to recover Username from claim. Claim '%s' is invalid or not a string", u)
		}
		res.Username = username

	}
	res.Groups, res.hasGroupClaim = GroupsFromClaims(c, g)
	return res, nil
}

// GroupsFromClaims fetches the group name list from claimprovider, such as decoded ID token.
// If the claims does not have the claim defined as k, the second return value will be false, otherwise true
func GroupsFromClaims(gp claimsProvider, k string) ([]string, bool) {
	res := make([]string, 0)
	claimMap := make(map[string]interface{})
	if err := gp.Claims(&claimMap); err != nil {
		log.Errorf("failed to fetch claims, error: %v", err)
		return res, false
	}
	g, ok := claimMap[k].([]interface{})
	if !ok {
		log.Warningf("Unable to get groups from claims, claims: %+v, groups claims key: %s", claimMap, k)
		return res, false
	}
	for _, e := range g {
		s, ok := e.(string)
		if !ok {
			log.Warningf("Element in group list is not string: %v, list: %v", e, g)
			continue
		}
		res = append(res, s)
	}
	return res, true
}

// Conn wraps connection info of an OIDC endpoint
type Conn struct {
	URL        string `json:"url"`
	VerifyCert bool   `json:"verify_cert"`
}

// TestEndpoint tests whether the endpoint is a valid OIDC endpoint.
// The nil return value indicates the success of the test
func TestEndpoint(conn Conn) error {

	// gooidc will try to call the discovery api when creating the provider and that's all we need to check
	ctx := clientCtx(context.Background(), conn.VerifyCert)
	_, err := gooidc.NewProvider(ctx, conn.URL)
	return err
}
