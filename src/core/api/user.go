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

package api

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// UserAPI handles request to /api/users/{}
type UserAPI struct {
	BaseController
	currentUserID    int
	userID           int
	SelfRegistration bool
	IsAdmin          bool
	AuthMode         string
	secretKey        string
}

type passwordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type userSearch struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

type secretReq struct {
	Secret string `json:"secret"`
}

// Prepare validates the URL and parms
func (ua *UserAPI) Prepare() {
	ua.BaseController.Prepare()
	mode, err := config.AuthMode()
	if err != nil {
		log.Errorf("failed to get auth mode: %v", err)
		ua.SendInternalServerError(errors.New(""))
		return
	}

	ua.AuthMode = mode
	if mode == common.OIDCAuth {
		key, err := config.SecretKey()
		if err != nil {
			log.Errorf("failed to get secret key: %v", err)
			ua.SendInternalServerError(fmt.Errorf("failed to get secret key: %v", err))
			return
		}
		ua.secretKey = key
	}

	self, err := config.SelfRegistration()
	if err != nil {
		log.Errorf("failed to get self registration: %v", err)
		ua.SendInternalServerError(errors.New(""))
		return
	}

	ua.SelfRegistration = self

	if !ua.SecurityCtx.IsAuthenticated() {
		if ua.Ctx.Input.IsPost() && ua.SelfRegistration {
			return
		}
		ua.SendUnAuthorizedError(errors.New("UnAuthorize"))
		return
	}

	user, err := dao.GetUser(models.User{
		Username: ua.SecurityCtx.GetUsername(),
	})
	if err != nil {
		ua.SendInternalServerError(fmt.Errorf("failed to get user %s: %v",
			ua.SecurityCtx.GetUsername(), err))
		return
	}

	if user == nil {
		log.Errorf("User with username %s does not exist in DB.", ua.SecurityCtx.GetUsername())
		ua.SendInternalServerError(fmt.Errorf("user %s does not exist in DB", ua.SecurityCtx.GetUsername()))
		return
	}

	ua.currentUserID = user.UserID
	id := ua.Ctx.Input.Param(":id")
	if id == "current" {
		ua.userID = ua.currentUserID
	} else if len(id) > 0 {
		var err error
		ua.userID, err = strconv.Atoi(id)
		if err != nil {
			log.Errorf("Invalid user id, error: %v", err)
			ua.SendBadRequestError(errors.New("invalid user Id"))
			return
		}
		userQuery := models.User{UserID: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ua.SendInternalServerError(errors.New("internal error"))
			return
		}
		if u == nil {
			log.Errorf("User with Id: %d does not exist", ua.userID)
			ua.SendNotFoundError(errors.New(""))
			return
		}
	}

	ua.IsAdmin = ua.SecurityCtx.IsSysAdmin()
}

// Get ...
func (ua *UserAPI) Get() {
	if ua.userID == ua.currentUserID || ua.IsAdmin {
		userQuery := models.User{UserID: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ua.SendInternalServerError(err)
			return
		}
		u.Password = ""
		if ua.userID == ua.currentUserID {
			sc := ua.SecurityCtx
			switch lsc := sc.(type) {
			case *local.SecurityContext:
				u.AdminRoleInAuth = lsc.User().AdminRoleInAuth
			}
		}
		if ua.AuthMode == common.OIDCAuth {
			o, err := ua.getOIDCUserInfo()
			if err != nil {
				ua.SendInternalServerError(err)
				return
			}
			u.OIDCUserMeta = o
		}
		ua.Data["json"] = u
		ua.ServeJSON()
		return
	}

	log.Errorf("Current user, id: %d does not have admin role, can not view other user's detail", ua.currentUserID)
	ua.SendForbiddenError(errors.New("user does not have admin role"))
	return
}

// List ...
func (ua *UserAPI) List() {
	if !ua.IsAdmin {
		log.Errorf("Current user, id: %d does not have admin role, can not list users", ua.currentUserID)
		ua.SendForbiddenError(errors.New("user does not have admin role"))
		return
	}

	page, size, err := ua.GetPaginationParams()
	if err != nil {
		ua.SendBadRequestError(err)
		return
	}

	query := &models.UserQuery{
		Username: ua.GetString("username"),
		Email:    ua.GetString("email"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	total, err := dao.GetTotalOfUsers(query)
	if err != nil {
		ua.SendInternalServerError(fmt.Errorf("failed to get total of users: %v", err))
		return
	}

	users, err := dao.ListUsers(query)
	if err != nil {
		ua.SendInternalServerError(fmt.Errorf("failed to get users: %v", err))
		return
	}
	for i := range users {
		user := &users[i]
		user.Password = ""
	}
	ua.SetPaginationHeader(total, page, size)
	ua.Data["json"] = users
	ua.ServeJSON()
}

// Search ...
func (ua *UserAPI) Search() {
	page, size, err := ua.GetPaginationParams()
	if err != nil {
		ua.SendBadRequestError(err)
		return
	}
	query := &models.UserQuery{
		Username: ua.GetString("username"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}
	if len(query.Username) == 0 {
		ua.SendBadRequestError(errors.New("username is required"))
		return
	}

	total, err := dao.GetTotalOfUsers(query)
	if err != nil {
		ua.SendInternalServerError(fmt.Errorf("failed to get total of users: %v", err))
		return
	}

	users, err := dao.ListUsers(query)
	if err != nil {
		ua.SendInternalServerError(fmt.Errorf("failed to get users: %v", err))
		return
	}

	var userSearches []userSearch
	for _, user := range users {
		userSearches = append(userSearches, userSearch{UserID: user.UserID, Username: user.Username})
	}

	ua.SetPaginationHeader(total, page, size)
	ua.Data["json"] = userSearches
	ua.ServeJSON()
}

// Put ...
func (ua *UserAPI) Put() {
	if !ua.modifiable() {
		ua.SendForbiddenError(fmt.Errorf("User with ID %d cannot be modified", ua.userID))
		return
	}
	user := models.User{}
	if err := ua.DecodeJSONReq(&user); err != nil {
		ua.SendBadRequestError(err)
		return
	}
	user.UserID = ua.userID
	err := commonValidate(user)
	if err != nil {
		log.Warningf("Bad request in change user profile: %v", err)
		ua.SendBadRequestError(fmt.Errorf("change user profile error:" + err.Error()))
		return
	}
	userQuery := models.User{UserID: ua.userID}
	u, err := dao.GetUser(userQuery)
	if err != nil {
		log.Errorf("Error occurred in GetUser, error: %v", err)
		ua.SendInternalServerError(errors.New("internal error"))
		return
	}
	if u == nil {
		log.Errorf("User with Id: %d does not exist", ua.userID)
		ua.SendNotFoundError(errors.New(""))
		return
	}
	if u.Email != user.Email {
		emailExist, err := dao.UserExists(user, "email")
		if err != nil {
			log.Errorf("Error occurred in change user profile: %v", err)
			ua.SendInternalServerError(errors.New("internal error"))
			return
		}
		if emailExist {
			log.Warning("email has already been used!")
			ua.SendConflictError(errors.New("email has already been used"))
			return
		}
	}
	if err := dao.ChangeUserProfile(user); err != nil {
		log.Errorf("Failed to update user profile, error: %v", err)
		ua.SendInternalServerError(err)
		return
	}
}

// Post ...
func (ua *UserAPI) Post() {

	if !(ua.AuthMode == common.DBAuth) {
		ua.SendForbiddenError(errors.New(""))
		return
	}

	if !(ua.SelfRegistration || ua.IsAdmin) {
		log.Warning("Registration can only be used by admin role user when self-registration is off.")
		ua.SendForbiddenError(errors.New(""))
		return
	}

	if !ua.IsAdmin && !lib.GetCarrySession(ua.Ctx.Request.Context()) {
		ua.SendForbiddenError(errors.New("self-registration cannot be triggered via API"))
		return
	}

	user := models.User{}
	if err := ua.DecodeJSONReq(&user); err != nil {
		ua.SendBadRequestError(err)
		return
	}
	err := validate(user)
	if err != nil {
		log.Warningf("Bad request in Register: %v", err)
		ua.RenderError(http.StatusBadRequest, "register error:"+err.Error())
		return
	}

	if !ua.IsAdmin && user.SysAdminFlag {
		msg := "Non-admin cannot create an admin user."
		log.Errorf(msg)
		ua.SendForbiddenError(errors.New(msg))
		return
	}

	userExist, err := dao.UserExists(user, "username")
	if err != nil {
		log.Errorf("Error occurred in Register: %v", err)
		ua.SendInternalServerError(errors.New("internal error"))
		return
	}
	if userExist {
		log.Warning("username has already been used!")
		ua.SendConflictError(errors.New("username has already been used"))
		return
	}
	emailExist, err := dao.UserExists(user, "email")
	if err != nil {
		log.Errorf("Error occurred in change user profile: %v", err)
		ua.SendInternalServerError(errors.New("internal error"))
		return
	}
	if emailExist {
		log.Warning("email has already been used!")
		ua.SendConflictError(errors.New("email has already been used"))
		return
	}

	userID, err := dao.Register(user)
	if err != nil {
		log.Errorf("Error occurred in Register: %v", err)
		ua.SendInternalServerError(errors.New("internal error"))
		return
	}

	ua.Redirect(http.StatusCreated, strconv.FormatInt(userID, 10))
}

// Delete ...
func (ua *UserAPI) Delete() {
	if !ua.IsAdmin || ua.AuthMode != common.DBAuth || ua.userID == 1 || ua.currentUserID == ua.userID {
		ua.SendForbiddenError(fmt.Errorf("User with ID: %d cannot be removed, auth mode: %s, current user ID: %d", ua.userID, ua.AuthMode, ua.currentUserID))
		return
	}

	var err error
	err = dao.DeleteUser(ua.userID)
	if err != nil {
		log.Errorf("Failed to delete data from database, error: %v", err)
		ua.SendInternalServerError(errors.New("failed to delete User"))
		return
	}
}

// ChangePassword handles PUT to /api/users/{}/password
func (ua *UserAPI) ChangePassword() {
	if !ua.modifiable() {
		ua.SendForbiddenError(fmt.Errorf("User with ID: %d is not modifiable", ua.userID))
		return
	}

	changePwdOfOwn := ua.userID == ua.currentUserID

	var req passwordReq
	if err := ua.DecodeJSONReq(&req); err != nil {
		ua.SendBadRequestError(err)
		return
	}

	if changePwdOfOwn && len(req.OldPassword) == 0 {
		ua.SendBadRequestError(errors.New("empty old_password"))
		return
	}

	if err := validateSecret(req.NewPassword); err != nil {
		ua.SendBadRequestError(err)
		return
	}

	user, err := dao.GetUser(models.User{UserID: ua.userID})
	if err != nil {
		ua.SendInternalServerError(fmt.Errorf("failed to get user %d: %v", ua.userID, err))
		return
	}
	if user == nil {
		ua.SendNotFoundError(fmt.Errorf("user %d not found", ua.userID))
		return
	}
	if changePwdOfOwn {
		if user.Password != utils.Encrypt(req.OldPassword, user.Salt, user.PasswordVersion) {
			log.Info("incorrect old_password")
			ua.SendForbiddenError(errors.New("incorrect old_password"))
			return
		}
	}
	if user.Password == utils.Encrypt(req.NewPassword, user.Salt, user.PasswordVersion) {
		ua.SendBadRequestError(errors.New("the new password can not be same with the old one"))
		return
	}

	updatedUser := models.User{
		UserID:          ua.userID,
		Password:        req.NewPassword,
		PasswordVersion: user.PasswordVersion,
	}
	if err = dao.ChangeUserPassword(updatedUser); err != nil {
		ua.SendInternalServerError(fmt.Errorf("failed to change password of user %d: %v", ua.userID, err))
		return
	}
}

// ToggleUserAdminRole handles PUT api/users/{}/sysadmin
func (ua *UserAPI) ToggleUserAdminRole() {
	if !ua.IsAdmin {
		log.Warningf("current user, id: %d does not have admin role, can not update other user's role", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	userQuery := models.User{UserID: ua.userID}
	if err := ua.DecodeJSONReq(&userQuery); err != nil {
		ua.SendBadRequestError(err)
		return
	}
	if err := dao.ToggleUserAdminRole(userQuery.UserID, userQuery.SysAdminFlag); err != nil {
		log.Errorf("Error occurred in ToggleUserAdminRole: %v", err)
		ua.SendInternalServerError(errors.New("internal error"))
		return
	}
}

// ListUserPermissions handles GET to /api/users/{}/permissions
func (ua *UserAPI) ListUserPermissions() {
	if ua.userID != ua.currentUserID {
		log.Warningf("Current user, id: %d can not view other user's permissions", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have permission")
		return
	}

	relative := ua.Ctx.Input.Query("relative") == "true"

	scope := rbac.Resource(ua.Ctx.Input.Query("scope"))
	policies := []*types.Policy{}

	if ns, ok := types.NamespaceFromResource(scope); ok {
		for _, policy := range ns.GetPolicies() {
			if ua.SecurityCtx.Can(policy.Action, policy.Resource) {
				policies = append(policies, policy)
			}
		}
	}

	results := []map[string]string{}
	for _, policy := range policies {
		var resource rbac.Resource

		// for resource `/project/1/repository` if `relative` is `true` then the resource in response will be `repository`
		if relative {
			relativeResource, err := policy.Resource.RelativeTo(scope)
			if err != nil {
				continue
			}
			resource = relativeResource
		} else {
			resource = policy.Resource
		}

		results = append(results, map[string]string{
			"resource": resource.String(),
			"action":   policy.Action.String(),
		})
	}

	ua.Data["json"] = results
	ua.ServeJSON()
	return
}

// SetCLISecret handles request PUT /api/users/:id/cli_secret to update the CLI secret of the user
func (ua *UserAPI) SetCLISecret() {
	if ua.AuthMode != common.OIDCAuth {
		ua.SendPreconditionFailedError(errors.New("the auth mode has to be oidc auth"))
		return
	}
	if ua.userID != ua.currentUserID && !ua.IsAdmin {
		ua.SendForbiddenError(errors.New(""))
		return
	}
	oidcData, err := dao.GetOIDCUserByUserID(ua.userID)
	if err != nil {
		log.Errorf("Failed to get OIDC User meta for user, id: %d, error: %v", ua.userID, err)
		ua.SendInternalServerError(errors.New("failed to get OIDC meta data for user"))
		return
	}
	if oidcData == nil {
		log.Errorf("User is not onboarded via OIDC AuthN, user id: %d", ua.userID)
		ua.SendPreconditionFailedError(errors.New("user is not onboarded via OIDC AuthN"))
		return
	}

	s := &secretReq{}
	if err := ua.DecodeJSONReq(s); err != nil {
		ua.SendBadRequestError(err)
		return
	}
	if err := validateSecret(s.Secret); err != nil {
		ua.SendBadRequestError(err)
		return
	}

	encSec, err := utils.ReversibleEncrypt(s.Secret, ua.secretKey)
	if err != nil {
		log.Errorf("Failed to encrypt secret, error: %v", err)
		ua.SendInternalServerError(errors.New("failed to encrypt secret"))
		return
	}
	oidcData.Secret = encSec
	err = dao.UpdateOIDCUserSecret(oidcData)
	if err != nil {
		log.Errorf("Failed to update secret in DB, error: %v", err)
		ua.SendInternalServerError(errors.New("failed to update secret in DB"))
		return
	}
}

func (ua *UserAPI) getOIDCUserInfo() (*models.OIDCUser, error) {
	o, err := dao.GetOIDCUserByUserID(ua.userID)
	if err != nil || o == nil {
		return nil, err
	}
	if len(o.Secret) > 0 {
		p, err := utils.ReversibleDecrypt(o.Secret, ua.secretKey)
		if err != nil {
			return nil, err
		}
		o.PlainSecret = p
	}
	return o, nil
}

// modifiable returns whether the modify is allowed based on current auth mode and context
func (ua *UserAPI) modifiable() bool {
	if ua.AuthMode == common.DBAuth {
		// When the auth mode is local DB, admin can modify anyone, non-admin can modify himself.
		return ua.IsAdmin || ua.userID == ua.currentUserID
	}
	// When the auth mode is external IDM backend, only the super user can modify himself,
	// because he's the only one whose information is stored in local DB.
	return ua.userID == 1 && ua.userID == ua.currentUserID

}

// validate only validate when user register
func validate(user models.User) error {

	if utils.IsIllegalLength(user.Username, 1, 255) {
		return fmt.Errorf("username with illegal length")
	}
	if utils.IsContainIllegalChar(user.Username, []string{",", "~", "#", "$", "%"}) {
		return fmt.Errorf("username contains illegal characters")
	}

	if err := validateSecret(user.Password); err != nil {
		return err
	}

	return commonValidate(user)
}

func validateSecret(in string) error {
	hasLower := regexp.MustCompile(`[a-z]`)
	hasUpper := regexp.MustCompile(`[A-Z]`)
	hasNumber := regexp.MustCompile(`[0-9]`)
	if len(in) >= 8 && hasLower.MatchString(in) && hasUpper.MatchString(in) && hasNumber.MatchString(in) {
		return nil
	}
	return errors.New("the password or secret must longer than 8 chars with at least 1 uppercase letter, 1 lowercase letter and 1 number")
}

// commonValidate validates email, realname, comment information when user register or change their profile
func commonValidate(user models.User) error {

	if len(user.Email) > 0 {
		if m, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, user.Email); !m {
			return fmt.Errorf("email with illegal format")
		}
	} else {
		return fmt.Errorf("Email can't be empty")
	}

	if utils.IsIllegalLength(user.Realname, 1, 255) {
		return fmt.Errorf("realname with illegal length")
	}

	if utils.IsContainIllegalChar(user.Realname, []string{",", "~", "#", "$", "%"}) {
		return fmt.Errorf("realname contains illegal characters")
	}
	if utils.IsIllegalLength(user.Comment, -1, 30) {
		return fmt.Errorf("comment with illegal length")
	}
	return nil

}
