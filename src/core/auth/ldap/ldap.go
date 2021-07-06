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

package ldap

import (
	"fmt"
	"regexp"
	"strings"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/utils"

	"github.com/goharbor/harbor/src/common/models"
	ldapUtils "github.com/goharbor/harbor/src/common/utils/ldap"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
)

// Auth implements AuthenticateHelper interface to authenticate against LDAP
type Auth struct {
	auth.DefaultAuthenticateHelper
}

// Authenticate checks user's credential against LDAP based on basedn template and LDAP URL,
// if the check is successful a dummy record will be inserted into DB, such that this user can
// be associated to other entities in the system.
func (l *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	p := m.Principal
	if len(strings.TrimSpace(p)) == 0 {
		log.Debugf("LDAP authentication failed for empty user id.")
		return nil, auth.NewErrAuth("Empty user id")
	}

	ldapSession, err := ldapUtils.LoadSystemLdapConfig()

	if err != nil {
		return nil, fmt.Errorf("can not load system ldap config: %v", err)
	}

	if err = ldapSession.Open(); err != nil {
		log.Warningf("ldap connection fail: %v", err)
		return nil, err
	}
	defer ldapSession.Close()

	ldapUsers, err := ldapSession.SearchUser(p)
	if err != nil {
		log.Warningf("ldap search fail: %v", err)
		return nil, err
	}
	if len(ldapUsers) == 0 {
		log.Warningf("Not found an entry.")
		return nil, auth.NewErrAuth("Not found an entry")
	} else if len(ldapUsers) != 1 {
		log.Warningf("Found more than one entry.")
		return nil, auth.NewErrAuth("Multiple entries found")
	}
	log.Debugf("Found ldap user %+v", ldapUsers[0])

	dn := ldapUsers[0].DN
	if err = ldapSession.Bind(dn, m.Password); err != nil {
		log.Warningf("Failed to bind user, username: %s, dn: %s, error: %v", p, dn, err)
		return nil, auth.NewErrAuth(err.Error())
	}

	u := models.User{}
	u.Username = ldapUsers[0].Username
	u.Realname = ldapUsers[0].Realname
	u.Email = strings.TrimSpace(ldapUsers[0].Email)

	l.syncUserInfoFromDB(&u)
	l.attachLDAPGroup(ldapUsers, &u, ldapSession)

	return &u, nil
}

func (l *Auth) attachLDAPGroup(ldapUsers []models.LdapUser, u *models.User, sess *ldapUtils.Session) {
	// Retrieve ldap related info in login to avoid too many traffic with LDAP server.
	// Get group admin dn
	groupCfg, err := config.LDAPGroupConf()
	if err != nil {
		log.Warningf("Failed to fetch ldap group configuration:%v", err)
		// most likely user doesn't configure user group info, it should not block user login
	}
	groupAdminDN := utils.TrimLower(groupCfg.LdapGroupAdminDN)
	// Attach user group
	for _, groupDN := range ldapUsers[0].GroupDNList {

		groupDN = utils.TrimLower(groupDN)
		// Attach LDAP group admin
		if len(groupAdminDN) > 0 && groupAdminDN == groupDN {
			u.AdminRoleInAuth = true
		}

	}
	userGroups := make([]models.UserGroup, 0)
	// fetch all groups from LDAP
	ldapGroups, err := sess.SearchAllGroups()
	if err != nil {
		log.Warningf("Failed to fetch ldap groups:%v", err)
	}
	// if group is in list, append it to user
	for _, dn := range ldapUsers[0].GroupDNList {
		for _, lGroup := range ldapGroups {
			if lGroup.GroupDN == dn {
				userGroups = append(userGroups, models.UserGroup{GroupName: lGroup.GroupName, LdapGroupDN: lGroup.GroupDN, GroupType: common.LDAPGroupType})
				break
			}
		}
	}
	u.GroupIDs, err = group.PopulateGroup(userGroups)
	if err != nil {
		log.Warningf("Failed to fetch ldap group configuration:%v", err)
	}
}

func (l *Auth) syncUserInfoFromDB(u *models.User) {
	// Retrieve SysAdminFlag from DB so that it transfer to session
	dbUser, err := dao.GetUser(models.User{Username: u.Username})
	if err != nil {
		log.Errorf("failed to sync user info from DB error %v", err)
		return
	}
	if dbUser == nil {
		return
	}
	u.SysAdminFlag = dbUser.SysAdminFlag
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func (l *Auth) OnBoardUser(u *models.User) error {
	if u.Email == "" {
		if strings.Contains(u.Username, "@") {
			u.Email = u.Username
		}
	}
	u.Password = "12345678AbC" // Password is not kept in local db
	u.Comment = "from LDAP."   // Source is from LDAP

	return dao.OnBoardUser(u)
}

// SearchUser -- Search user in ldap
func (l *Auth) SearchUser(username string) (*models.User, error) {
	var user models.User
	ldapSession, err := ldapUtils.LoadSystemLdapConfig()
	if err = ldapSession.Open(); err != nil {
		return nil, fmt.Errorf("Failed to load system ldap config, %v", err)
	}
	defer ldapSession.Close()
	ldapUsers, err := ldapSession.SearchUser(username)
	if err != nil {
		return nil, fmt.Errorf("Failed to search user in ldap")
	}

	if len(ldapUsers) > 1 {
		log.Warningf("There are more than one user found, return the first user")
	}
	if len(ldapUsers) > 0 {

		user.Username = strings.TrimSpace(ldapUsers[0].Username)
		user.Realname = strings.TrimSpace(ldapUsers[0].Realname)
		user.Email = strings.TrimSpace(ldapUsers[0].Email)

		log.Debugf("Found ldap user %v", user)
	} else {
		return nil, fmt.Errorf("No user found, %v", username)
	}

	return &user, nil
}

// SearchGroup -- Search group in ldap authenticator, groupKey is LDAP group DN.
func (l *Auth) SearchGroup(groupKey string) (*models.UserGroup, error) {
	if _, err := goldap.ParseDN(groupKey); err != nil {
		return nil, auth.ErrInvalidLDAPGroupDN
	}
	ldapSession, err := ldapUtils.LoadSystemLdapConfig()

	if err != nil {
		return nil, fmt.Errorf("can not load system ldap config: %v", err)
	}

	if err = ldapSession.Open(); err != nil {
		log.Warningf("ldap connection fail: %v", err)
		return nil, err
	}
	defer ldapSession.Close()
	userGroupList, err := ldapSession.SearchGroupByDN(groupKey)

	if err != nil {
		log.Warningf("ldap search group fail: %v", err)
		return nil, err
	}

	if len(userGroupList) == 0 {
		return nil, fmt.Errorf("Failed to searh ldap group with groupDN:%v", groupKey)
	}
	userGroup := models.UserGroup{
		GroupName:   userGroupList[0].GroupName,
		LdapGroupDN: userGroupList[0].GroupDN,
	}
	return &userGroup, nil
}

// OnBoardGroup -- Create Group in harbor DB, if altGroupName is not empty, take the altGroupName as groupName in harbor DB.
func (l *Auth) OnBoardGroup(u *models.UserGroup, altGroupName string) error {
	if _, err := goldap.ParseDN(u.LdapGroupDN); err != nil {
		return auth.ErrInvalidLDAPGroupDN
	}
	if len(altGroupName) > 0 {
		u.GroupName = altGroupName
	}
	u.GroupType = common.LDAPGroupType
	// Check duplicate LDAP DN in usergroup, if usergroup exist, return error
	userGroupList, err := group.QueryUserGroup(models.UserGroup{LdapGroupDN: u.LdapGroupDN})
	if err != nil {
		return err
	}
	if len(userGroupList) > 0 {
		return auth.ErrDuplicateLDAPGroup
	}
	return group.OnBoardUserGroup(u)
}

// PostAuthenticate -- If user exist in harbor DB, sync email address, if not exist, call OnBoardUser
func (l *Auth) PostAuthenticate(u *models.User) error {

	exist, err := dao.UserExists(*u, "username")
	if err != nil {
		return err
	}

	if exist {
		queryCondition := models.User{
			Username: u.Username,
		}
		dbUser, err := dao.GetUser(queryCondition)
		if err != nil {
			return err
		}
		if dbUser == nil {
			fmt.Printf("User not found in DB %+v", u)
			return nil
		}
		u.UserID = dbUser.UserID

		if dbUser.Email != u.Email {
			Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
			if !Re.MatchString(u.Email) {
				log.Debugf("Not a valid email address: %v, skip to sync", u.Email)
			} else {
				if err = dao.ChangeUserProfile(*u, "Email"); err != nil {
					u.Email = dbUser.Email
					log.Errorf("failed to sync user email: %v", err)
				}
			}
		}

		return nil
	}

	err = auth.OnBoardUser(u)
	if err != nil {
		return err
	}
	if u.UserID <= 0 {
		return fmt.Errorf("Can not OnBoardUser %v", u)
	}
	return nil
}

func init() {
	auth.Register(common.LDAPAuth, &Auth{})
}
