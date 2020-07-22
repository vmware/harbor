---
title: E2E (API) Python Based Test Scripting Guide
draft: true
---

#### Preparation ####

After getting Harbor source code (git clone https://github.com/goharbor/harbor.git), Harbor E2E API  test scripts  can be found in tests/apitests/python directory.

Before scripting, please make sure swagger client has been made by "make swagger_client", then archive "harborclient" will be made

1. `git clone https://github.com/goharbor/harbor.git`

2. `cd harbor`

3. `make swagger_client`

4. `cd harborclient/`


We use $HARBORCLIENT_PATH represent the path of "harborclient" you've made by "make swagger_client".

Deploy Harbor instance for testing, and we will use $HARBOR_IP_ADDR to represent the deployed Harbor in this document.

Harbor E2E API test scripts will import python library under archive "harborclient", please set OS environment variant for Harbor E2E API test scripts:


1. `export HARBOR_HOST=$HARBOR_IP_ADDR`
2. `export SWAGGER_CLIENT_PATH=$HARBORCLIENT_PATH`

Until now, we have all preparation work done.

#### Scripting ####

As you can see, we will use python library made by "make swagger_client", in this library, we have all API functions and models, but for more convenience we encapsulate one more level in archive "library", so the script structure is as bellow:

	-library/

	-test_project_level_policy_content_trust.py

	-test_project_quota.py

	-test_retention.py

	-...

You can add both library code and script code, since not all APIs have been encapsulated.


#### Manual Execution Example ####

root@harbor:/harbor/code/harbor# `python ./tests/apitests/python/test_add_sys_label_to_tag.py`

2020-03-11 13:40:07,269 DEBUG Starting new HTTPS connection (1): 1.1.1.1:443
send: 'POST /api/v2.0/users HTTP/1.1\r\nHost: 1.1.1.1\r\nAccept-Encoding: identity\r\nContent-Length: 156\r\nContent-Type: application/json\r\nAccept: application/json\r\nAuthorization: Basic YWRtaW46SGFyYm9yMTIzNDU=\r\nUser-Agent: Swagger-Codegen/1.0.0/python\r\n\r\n{"username": "user-1583934007059", "role_id": 0, "password": "xxxxxxxx", "email": "realname-1583934007059@vmware.com", "realname": "realname-1583934007059"}'
reply: 'HTTP/1.1 201 Created\r\n'
header: Server: nginx
header: Date: Wed, 11 Mar 2020 13:40:07 GMT
header: Content-Length: 0
header: Connection: keep-alive
header: Location: /api/v2.0/users/9
header: Set-Cookie: sid=2e05f902f345b855ec33221ade1c6d09; Path=/; Secure; HttpOnly
header: X-Request-Id: 9ba11b26-2cdb-432f-878e-3fed04fa61b1
header: Strict-Transport-Security: max-age=31536000; includeSubdomains; preload
header: X-Frame-Options: DENY
header: Content-Security-Policy: frame-ancestors 'none'
2020-03-11 13:40:07,482 DEBUG https://1.1.1.1:443 "POST /api/v2.0/users HTTP/1.1" 201 0
2020-03-11 13:40:07,483 DEBUG response body:

......

......

......

#### How To Trigger Script In CI ####

If you like to have your scripts running in CI which is for verification of pull requests, please add your scripts into *https://github.com/goharbor/harbor/blob/master/tests/robot-cases/Group0-BAT/API_DB.robot* file, then scripts can be triggered once there is a pull request.



