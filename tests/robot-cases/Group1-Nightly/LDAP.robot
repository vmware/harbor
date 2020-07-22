// Copyright Project Harbor Authors
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

*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - Ldap Verify Cert
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Test LDAP Server Success
    Close Browser

Test Case - Ldap Sign in and out
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Close Browser

Test Case - System Admin On-board New Member
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To User Tag
    Sleep  2
    Page Should Not Contain  mike02
    Navigate To Projects
    Create An New Project And Go Into Project  project${d}
    Switch To Member
    Add Guest Member To Project  mike02
    Page Should Contain  mike02
    Close Browser

Test Case - LDAP User On-borad New Member
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  mike03  zhu88jie
    Create An New Project And Go Into Project  project${d}
    Switch To Member
    Sleep  2
    Page Should Not Contain  mike04
    Add Guest Member To Project  mike04
    Sleep  2
    Page Should Contain  mike04
    Close Browser

Test Case - Home Page Differences With DB Mode
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Logout Harbor
    Sleep  2
    Page Should Not Contain  Sign up
    Page Should Not Contain  Forgot password
    Close Browser

Test Case - New User Button Is Unusable
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To User Tag
    Add User Button Should Be Disabled
    Close Browser

Test Case - Change Password Is Invisible
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  mike05  zhu88jie
    Ldap User Should Not See Change Password
    Close Browser

Test Case - Ldap User Create Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Create An New Project And Go Into Project  project${d}
    Logout Harbor
    Manage Project Member  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  mike02  Add  has_image=${false}
    Close Browser

Test Case - Ldap User Push An Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Create An New Project And Go Into Project  project${d}

    Push Image  ${ip}  mike  zhu88jie  project${d}  hello-world:latest
    Go Into Project  project${d}
    Wait Until Page Contains  project${d}/hello-world
    Close Browser

Test Case - Ldap User Can Not login
    Docker Login Fail  ${ip}  testerDeesExist  123456

Test Case - Run LDAP Group Related API Test
    Harbor API Test  ./tests/apitests/python/test_ldap_admin_role.py
    Harbor API Test  ./tests/apitests/python/test_user_group.py
    Harbor API Test  ./tests/apitests/python/test_assign_role_to_ldap_group.py
