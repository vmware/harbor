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
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - UAA Sign in and out
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
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

Test Case - UAA User Push An Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Create An New Project And Go Into Project  project${d}

    Push Image  ${ip}  mike  zhu88jie  project${d}  hello-world:latest
    Go Into Project  project${d}
    Wait Until Page Contains  project${d}/hello-world
    Close Browser

Test Case - UAA User Can Not login
    Docker Login Fail  ${ip}  testerDeesExist  123456
