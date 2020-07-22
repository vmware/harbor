---
title: Configure Harbor User Settings at the Command Line
weight: 65
---

From release 1.8.0 onwards, user settings are configured separately from the system settings. You do not configure user settings in the `harbor.yml` file, but rather in the Harbor interface or via HTTP requests.

## Example Configuration Commands:

**Add a new user in the local database:**

```sh
curl -X PUT -u "<username>:<password>" -H "Content-Type: application/json" -ki <Harbor Server URL>/api/configurations -d'{"<item_name>":"<item_value>"}'
```

**Get the current configuration:**

```sh
curl -u "<username>:<password>" -H "Content-Type: application/json" -ki <Harbor Server URL>/api/configurations
```

**Update Harbor to use LDAP authentication:**

Command

```shell
curl -X PUT -u "<username>:<password>" -H "Content-Type: application/json" -ki https://harbor.sample.domain/api/configurations -d'{"auth_mode":"ldap_auth"}'
```

Output

```
HTTP/1.1 200 OK
Server: nginx
Date: Wed, 08 May 2019 08:22:02 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 0
Connection: keep-alive
Set-Cookie: sid=a5803a1265e2b095cf65ce1d8bbd79b1; Path=/; HttpOnly
```

**Restrict project creation to Harbor administrators:**

Command

```shell
curl -X PUT -u "<username>:<password>" -H "Content-Type: application/json" -ki https://harbor.sample.domain/api/configurations -d'{"project_creation_restriction":"adminonly"}'
```

Output

```
HTTP/1.1 200 OK
Server: nginx
Date: Wed, 08 May 2019 08:24:32 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 0
Connection: keep-alive
Set-Cookie: sid=b7925eaf7af53bdefb13bdcae201a14a; Path=/; HttpOnly
```

**Update the token expiration time:**

Command

```shell
curl -X PUT -u "<username>:<password>" -H "Content-Type: application/json" -ki https://harbor.sample.domain/api/configurations -d'{"token_expiration":"300"}'
```

Output

```
HTTP/1.1 200 OK
Server: nginx
Date: Wed, 08 May 2019 08:23:38 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 0
Connection: keep-alive
Set-Cookie: sid=cc1bc93ffa2675253fc62b4bf3d9de0e; Path=/; HttpOnly
```

## Harbor user settings

| Configure item name | Description  | Type | Required | Default Value | 
| ------------ |------------ | ---- | ----- | ----- |
auth_mode | Authentication mode, it can be db_auth, ldap_auth, uaa_auth or oidc_auth  | string
email_from |   Email from  |  string | required (email feature)
email_host |   Email server  |  string | required (email feature)
email_identity |  Email identity  | string | optional (email feature)
email_password |  Email password | string | required (email feature)
email_insecure |  Email verify certificate, true or false |boolean  | optional (email feature) | false
email_port |    Email server port | number | required (email feature)
email_ssl |    Email SSL | boolean    | optional | false
email_username |  Email username | string | required (email feature)
ldap_url |  LDAP URL | string | required | 
ldap_base_dn |   LDAP base DN  | string | required(ldap_auth)
ldap_filter |    LDAP filter | string | optional
ldap_scope |  LDAP search scope, 0-Base Level, 1- One Level, 2-Sub Tree | number | optional | 2-Sub Tree
ldap_search_dn |  LDAP DN to search LDAP users| string | required(ldap_auth)
ldap_search_password |  LDAP DN's password  |string | required(ldap_auth)
ldap_timeout |  LDAP connection timeout | number | optional | 5
ldap_uid |  LDAP attribute to indicate the username in Harbor | string | optional | cn
ldap_verify_cert | Verify cert when create SSL connection with LDAP server, true or false | boolean | optional | true
ldap_group_admin_dn |  LDAP Group Admin DN | string | optional
ldap_group_attribute_name | LDAP Group Attribute, the LDAP attribute indicate the groupname in Harbor, it can be gid or cn | string | optional | cn
ldap_group_base_dn |  The Base DN which to search the LDAP groups | string | required(ldap_auth and LDAP group)
ldap_group_search_filter | The filter to search LDAP groups | string | optional
ldap_group_search_scope | LDAP group search scope, 0-Base Level, 1- One Level, 2-Sub Tree | number | optional | 2-Sub Tree|
ldap_group_membership_attribute |  LDAP group membership attribute, to indicate the group membership, it can be memberof, or ismemberof | string | optional | memberof
project_creation_restriction | The option to indicate user can be create object, it can be everyone, adminonly | string | optional | everyone
read_only | The option to set repository read only, it can be true or false | boolean | optional | false
self_registration | User can register account in Harbor, it can be true or false | boolean | optional| true
token_expiration | Security token expirtation time in minutes | number |optional| 30
uaa_client_id | UAA client ID | string | required(uaa_auth)
uaa_client_secret | UAA certificate | string | required(uaa_auth)
uaa_endpoint | UAA endpoint | string |  required(uaa_auth)
uaa_verify_cert | UAA verify cert, true or false | boolean | optional | true
oidc_name | Name for OIDC authentication | string | required(oidc_auth)
oidc_endpoint | Endpoint for OIDC auth | string | required(oidc_auth)
oidc_client_id | Client id for OIDC auth | string | required(oidc_auth)
oidc_client_secret | Client secret for OIDC auth |string | required(oidc_auth)
oidc_scope | Ccope for OIDC auth | string| required(oidc_auth)
oidc_verify_cert | Verify certificate for OIDC auth, true or false | boolean | optional| true
robot_token_duration | Robot token expiration time in minutes | number | optional | 43200 (30days)

{{< note >}}
Both booleans and numbers can be enclosed with double quote in the request json, for example: `123`, `"123"`, `"true"` or `true` is OK.
{{< /note >}}
