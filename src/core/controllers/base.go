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
	"bytes"
	"context"
	"html/template"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils"
	email_util "github.com/goharbor/harbor/src/common/utils/email"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
)

// CommonController handles request from UI that doesn't expect a page, such as /SwitchLanguage /logout ...
type CommonController struct {
	api.BaseController
	i18n.Locale
}

// Render returns nil.
func (cc *CommonController) Render() error {
	return nil
}

// Prepare overwrites the Prepare func in api.BaseController to ignore unnecessary steps
func (cc *CommonController) Prepare() {}

type messageDetail struct {
	Hint string
	URL  string
	UUID string
}

func redirectForOIDC(ctx context.Context, username string) bool {
	if lib.GetAuthMode(ctx) != common.OIDCAuth {
		return false
	}
	u, err := dao.GetUser(models.User{Username: username})
	if err != nil {
		log.Warningf("Failed to get user by name: %s, error: %v", username, err)
	}
	if u == nil {
		return true
	}
	ou, err := dao.GetOIDCUserByUserID(u.UserID)
	if err != nil {
		log.Warningf("Failed to get OIDC user info for user, id: %d, error: %v", u.UserID, err)
	}
	if ou != nil {
		return true
	}
	return false
}

// Login handles login request from UI.
func (cc *CommonController) Login() {
	principal := cc.GetString("principal")
	password := cc.GetString("password")
	if redirectForOIDC(cc.Ctx.Request.Context(), principal) {
		ep, err := config.ExtEndpoint()
		if err != nil {
			log.Errorf("Failed to get the external endpoint, error: %v", err)
			cc.CustomAbort(http.StatusUnauthorized, "")
		}
		url := strings.TrimSuffix(ep, "/") + common.OIDCLoginPath
		log.Debugf("Redirect user %s to login page of OIDC provider", principal)
		// Return a json to UI with status code 403, as it cannot handle status 302
		cc.Ctx.Output.Status = http.StatusForbidden
		cc.Ctx.Output.JSON(struct {
			Location string `json:"redirect_location"`
		}{url}, false, false)
		return
	}

	user, err := auth.Login(models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		cc.CustomAbort(http.StatusUnauthorized, "")
	}
	cc.PopulateUserSession(*user)
}

// LogOut Habor UI
func (cc *CommonController) LogOut() {
	cc.DestroySession()
}

// UserExists checks if user exists when user input value in sign in form.
func (cc *CommonController) UserExists() {
	flag, err := config.SelfRegistration()
	if err != nil {
		log.Errorf("Failed to get the status of self registration flag, error: %v, disabling user existence check", err)
	}
	securityCtx, ok := security.FromContext(cc.Ctx.Request.Context())
	isAdmin := ok && securityCtx.IsSysAdmin()
	if !flag && !isAdmin {
		cc.CustomAbort(http.StatusPreconditionFailed, "self registration disabled, only sysadmin can check user existence")
	}

	target := cc.GetString("target")
	value := cc.GetString("value")

	user := models.User{}
	switch target {
	case "username":
		user.Username = value
	case "email":
		user.Email = value
	}

	exist, err := dao.UserExists(user, target)
	if err != nil {
		log.Errorf("Error occurred in UserExists: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	cc.Data["json"] = exist
	cc.ServeJSON()
}

// SendResetEmail verifies the Email address and contact SMTP server to send reset password Email.
func (cc *CommonController) SendResetEmail() {

	email := cc.GetString("email")

	valid, err := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, email)
	if err != nil {
		log.Errorf("failed to match regexp: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if !valid {
		cc.CustomAbort(http.StatusBadRequest, "invalid email")
	}

	queryUser := models.User{Email: email}
	u, err := dao.GetUser(queryUser)
	if err != nil {
		log.Errorf("Error occurred in GetUser: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if u == nil {
		log.Debugf("email %s not found", email)
		cc.CustomAbort(http.StatusNotFound, "email_does_not_exist")
	}

	if !isUserResetable(u) {
		log.Errorf("Resetting password for user with ID: %d is not allowed", u.UserID)
		cc.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	uuid := utils.GenerateRandomString()
	user := models.User{ResetUUID: uuid, Email: email}
	if err = dao.UpdateUserResetUUID(user); err != nil {
		log.Errorf("failed to update user reset UUID: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	messageTemplate, err := template.ParseFiles("views/reset-password-mail.tpl")
	if err != nil {
		log.Errorf("Parse email template file failed: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, err.Error())
	}

	message := new(bytes.Buffer)

	harborURL, err := config.ExtEndpoint()
	if err != nil {
		log.Errorf("failed to get domain name: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	err = messageTemplate.Execute(message, messageDetail{
		Hint: cc.Tr("reset_email_hint"),
		URL:  harborURL,
		UUID: uuid,
	})

	if err != nil {
		log.Errorf("Message template error: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "internal_error")
	}

	settings, err := config.Email()
	if err != nil {
		log.Errorf("failed to get email configurations: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "internal_error")
	}

	addr := net.JoinHostPort(settings.Host, strconv.Itoa(settings.Port))
	err = email_util.Send(addr,
		settings.Identity,
		settings.Username,
		settings.Password,
		60, settings.SSL,
		settings.Insecure,
		settings.From,
		[]string{email},
		"Reset Harbor user password",
		message.String())
	if err != nil {
		log.Errorf("Send email failed: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "send_email_failed")
	}
}

// ResetPassword handles request from the reset page and reset password
func (cc *CommonController) ResetPassword() {

	resetUUID := cc.GetString("reset_uuid")
	if resetUUID == "" {
		cc.CustomAbort(http.StatusBadRequest, "Reset uuid is blank.")
	}

	queryUser := models.User{ResetUUID: resetUUID}
	user, err := dao.GetUser(queryUser)

	if err != nil {
		log.Errorf("Error occurred in GetUser: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if user == nil {
		log.Error("User does not exist")
		cc.CustomAbort(http.StatusBadRequest, "User does not exist")
	}

	if !isUserResetable(user) {
		log.Errorf("Resetting password for user with ID: %d is not allowed", user.UserID)
		cc.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}

	rawPassword := cc.GetString("password")

	if rawPassword != "" {
		err = dao.ResetUserPassword(*user, rawPassword)
		if err != nil {
			log.Errorf("Error occurred in ResetUserPassword: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	} else {
		cc.CustomAbort(http.StatusBadRequest, "password_is_required")
	}
}

func isUserResetable(u *models.User) bool {
	if u == nil {
		return false
	}
	mode, err := config.AuthMode()
	if err != nil {
		log.Errorf("Failed to get the auth mode, error: %v", err)
		return false
	}
	if mode == common.DBAuth {
		return true
	}
	return u.UserID == 1
}

func init() {
	// conf/app.conf -> os.Getenv("config_path")
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		if err := beego.LoadAppConfig("ini", configPath); err != nil {
			log.Errorf("failed to load app config: %v", err)
		}
	}

}
