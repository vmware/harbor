# -*- coding: utf-8 -*-

import base
import swagger_client
from swagger_client.rest import ApiException

class User(base.Base):

    def create_user(self, name=None,
        email = None, user_password=None, realname = None, role_id = None, expect_status_code=201, **kwargs):
        if name is None:
            name = base._random_name("user")
        if realname is None:
            realname = base._random_name("realname")
        if email is None:
            email = '%s@%s.com' % (realname,"harbortest")
        if user_password is None:
            user_password = "Harbor12345678"
        if role_id is None:
            role_id = 0

        client = self._get_client(**kwargs)
        user = swagger_client.User(username = name, email = email, password = user_password, realname = realname, role_id = role_id)

        try:
            _, status_code, header = client.users_post_with_http_info(user)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            return base._get_id_from_header(header), name

    def get_users(self, user_name=None, email=None, page=None, page_size=None, expect_status_code=200, **kwargs):
        client = self._get_client(**kwargs)
        params={}
        if user_name is not None:
            params["username"] = user_name
        if email is not None:
            params["email"] = email
        if page is not None:
            params["page"] = page
        if page_size is not None:
            params["page_size"] = page_size
        try:
            data, status_code, _ = client.users_get_with_http_info(**params)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            return data

    def get_user_by_id(self, user_id, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.users_user_id_get_with_http_info(user_id)
        base._assert_status_code(200, status_code)
        return data

    def get_user_by_name(self, name, expect_status_code=200, **kwargs):
        users = self.get_users(user_name=name, expect_status_code=expect_status_code , **kwargs)
        for user in users:
            if user.username == name:
                return user
        return None


    def get_user_current(self, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.users_current_get_with_http_info()
        base._assert_status_code(200, status_code)
        return data

    def delete_user(self, user_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.users_user_id_delete_with_http_info(user_id)
        base._assert_status_code(expect_status_code, status_code)
        return user_id

    def update_user_pwd(self, user_id, new_password=None, old_password=None, **kwargs):
        if old_password is None:
            old_password  = ""
        password = swagger_client.Password(old_password, new_password)
        client = self._get_client(**kwargs)
        _, status_code, _ = client.users_user_id_password_put_with_http_info(user_id, password)
        base._assert_status_code(200, status_code)
        return user_id

    def update_user_profile(self, user_id, email=None, realname=None, comment=None, **kwargs):
        client = self._get_client(**kwargs)
        user_rofile = swagger_client.UserProfile(email, realname, comment)
        _, status_code, _ = client.users_user_id_put_with_http_info(user_id, user_rofile)
        base._assert_status_code(200, status_code)
        return user_id

    def update_user_role_as_sysadmin(self, user_id, IsAdmin, **kwargs):
        client = self._get_client(**kwargs)
        sysadmin_flag = swagger_client.SysAdminFlag(IsAdmin)
        _, status_code, _ = client.users_user_id_sysadmin_put_with_http_info(user_id, sysadmin_flag)
        base._assert_status_code(200, status_code)
        return user_id
