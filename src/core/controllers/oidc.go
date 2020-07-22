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

package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/dao/group"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/oidc"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

const tokenKey = "oidc_token"
const stateKey = "oidc_state"
const userInfoKey = "oidc_user_info"
const oidcUserComment = "Onboarded via OIDC provider"

// OIDCController handles requests for OIDC login, callback and user onboard
type OIDCController struct {
	api.BaseController
}

type onboardReq struct {
	Username string `json:"username"`
}

// Prepare include public code path for call request handler of OIDCController
func (oc *OIDCController) Prepare() {
	if mode, _ := config.AuthMode(); mode != common.OIDCAuth {
		oc.SendPreconditionFailedError(fmt.Errorf("auth mode: %s is not OIDC based", mode))
		return
	}
}

// RedirectLogin redirect user's browser to OIDC provider's login page
func (oc *OIDCController) RedirectLogin() {
	state := utils.GenerateRandomString()
	url, err := oidc.AuthCodeURL(state)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	oc.SetSession(stateKey, state)
	log.Debugf("State dumped to session: %s", state)
	// Force to use the func 'Redirect' of beego.Controller
	oc.Controller.Redirect(url, http.StatusFound)
}

// Callback handles redirection from OIDC provider.  It will exchange the token and
// kick off onboard if needed.
func (oc *OIDCController) Callback() {
	if oc.Ctx.Request.URL.Query().Get("state") != oc.GetSession(stateKey) {
		log.Errorf("State mismatch, in session: %s, in url: %s", oc.GetSession(stateKey),
			oc.Ctx.Request.URL.Query().Get("state"))
		oc.SendBadRequestError(errors.New("State mismatch"))
		return
	}

	errorCode := oc.Ctx.Request.URL.Query().Get("error")
	if errorCode != "" {
		errorDescription := oc.Ctx.Request.URL.Query().Get("error_description")
		log.Errorf("OIDC callback returned error: %s - %s", errorCode, errorDescription)
		oc.SendBadRequestError(errors.Errorf("OIDC callback returned error: %s - %s", errorCode, errorDescription))
		return
	}

	code := oc.Ctx.Request.URL.Query().Get("code")
	ctx := oc.Ctx.Request.Context()
	token, err := oidc.ExchangeToken(ctx, code)
	if err != nil {
		log.Errorf("Failed to exchange token, error: %v", err)
		// Return a 4xx error so user can see the details in case it's due to misconfiguration.
		oc.SendBadRequestError(err)
		return
	}
	_, err = oidc.VerifyToken(ctx, token.RawIDToken)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	info, err := oidc.UserInfoFromToken(ctx, token)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	ouDataStr, err := json.Marshal(info)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	u, err := dao.GetUserBySubIss(info.Subject, info.Issuer)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	oc.SetSession(tokenKey, tokenBytes)

	oidcSettings, err := config.OIDCSetting()
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}

	if u == nil {
		// Recover the username from d.Username by default
		username := info.Username

		// Fix blanks in username
		username = strings.Replace(username, " ", "_", -1)

		// If automatic onboard is enabled, skip the onboard page
		if oidcSettings.AutoOnboard {
			log.Debug("Doing automatic onboarding\n")
			user, onboarded := userOnboard(oc, info, username, tokenBytes)
			if onboarded == false {
				log.Error("User not onboarded\n")
				return
			}
			log.Debug("User automatically onboarded\n")
			u = user
		} else {
			oc.SetSession(userInfoKey, string(ouDataStr))
			oc.Controller.Redirect(fmt.Sprintf("/oidc-onboard?username=%s", username), http.StatusFound)
			// Once redirected, no further actions are done
			return
		}
	}

	gids, err := group.PopulateGroup(models.UserGroupsFromName(info.Groups, common.OIDCGroupType))
	if err != nil {
		log.Warningf("Failed to populate groups, error: %v, user will have empty group list, username: %s", err, info.Username)
	}
	u.GroupIDs = gids
	oidcUser, err := dao.GetOIDCUserByUserID(u.UserID)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	_, t, err := secretAndToken(tokenBytes)
	oidcUser.Token = t
	if err := dao.UpdateOIDCUser(oidcUser); err != nil {
		oc.SendInternalServerError(err)
		return
	}
	oc.PopulateUserSession(*u)
	oc.Controller.Redirect("/", http.StatusFound)

}

func userOnboard(oc *OIDCController, info *oidc.UserInfo, username string, tokenBytes []byte) (*models.User, bool) {
	s, t, err := secretAndToken(tokenBytes)
	if err != nil {
		oc.SendInternalServerError(err)
		return nil, false
	}

	gids, err := group.PopulateGroup(models.UserGroupsFromName(info.Groups, common.OIDCGroupType))
	if err != nil {
		log.Warningf("Failed to populate group user will have empty group list. username: %s", username)
	}

	oidcUser := models.OIDCUser{
		SubIss: info.Subject + info.Issuer,
		Secret: s,
		Token:  t,
	}

	user := models.User{
		Username:     username,
		Realname:     username,
		Email:        info.Email,
		GroupIDs:     gids,
		OIDCUserMeta: &oidcUser,
		Comment:      oidcUserComment,
	}

	log.Debugf("User created: %+v\n", user)

	err = dao.OnBoardOIDCUser(&user)
	if err != nil {
		if strings.Contains(err.Error(), dao.ErrDupUser.Error()) {
			oc.RenderError(http.StatusConflict, "Conflict, the user with same username or email has been onboarded.")
			return nil, false
		}

		oc.SendInternalServerError(err)
		return nil, false
	}

	return &user, true
}

// Onboard handles the request to onboard a user authenticated via OIDC provider
func (oc *OIDCController) Onboard() {
	u := &onboardReq{}
	if err := oc.DecodeJSONReq(u); err != nil {
		oc.SendBadRequestError(err)
		return
	}
	username := u.Username
	if utils.IsIllegalLength(username, 1, 255) {
		oc.SendBadRequestError(errors.New("username with illegal length"))
		return
	}
	if utils.IsContainIllegalChar(username, []string{",", "~", "#", "$", "%"}) {
		oc.SendBadRequestError(errors.New("username contains illegal characters"))
		return
	}

	userInfoStr, ok := oc.GetSession(userInfoKey).(string)
	if !ok {
		oc.SendBadRequestError(errors.New("Failed to get OIDC user info from session"))
		return
	}
	log.Debugf("User info string: %s\n", userInfoStr)
	tb, ok := oc.GetSession(tokenKey).([]byte)
	if !ok {
		oc.SendBadRequestError(errors.New("Failed to get OIDC token from session"))
		return
	}

	d := &oidc.UserInfo{}
	err := json.Unmarshal([]byte(userInfoStr), &d)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}

	if user, onboarded := userOnboard(oc, d, username, tb); onboarded {
		user.OIDCUserMeta = nil
		oc.DelSession(userInfoKey)
		oc.PopulateUserSession(*user)
	}

}

func secretAndToken(tokenBytes []byte) (string, string, error) {
	key, err := config.SecretKey()
	if err != nil {
		return "", "", err
	}
	token, err := utils.ReversibleEncrypt((string)(tokenBytes), key)
	if err != nil {
		return "", "", err
	}
	str := utils.GenerateRandomString()
	secret, err := utils.ReversibleEncrypt(str, key)
	if err != nil {
		return "", "", err
	}
	return secret, token, nil
}
