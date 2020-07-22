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
Test Case - Create An New User
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Self Reg Should Be Disabled
    Sleep  1
    Enable Self Reg
    Logout Harbor

    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Close Browser

Test Case - Update User Comment
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Update User Comment  Test12#4
    Logout Harbor

Test Case - Update Password
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Change Password  Test1@34  Test12#4
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  tester${d}  Test12#4
    Close Browser

Test Case - Delete Multi User
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Create An New User  ${HARBOR_URL}  deletea${d}  testa${d}@vmware.com  test${d}  Test1@34  harbor
    Logout Harbor
    Create An New User  ${HARBOR_URL}  deleteb${d}  testb${d}@vmware.com  test${d}  Test1@34  harbor
    Logout Harbor
    Create An New User  ${HARBOR_URL}  deletec${d}  testc${d}@vmware.com  test${d}  Test1@34  harbor
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  admin  Harbor12345
    Switch To User Tag
    Filter Object  delete
    Multi-delete User  deletea  deleteb  deletec
    # Assert delete
    Delete Success  deletea  deleteb  deletec
    Sleep  1
    Close Browser

Test Case - Edit Self-Registration
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Disable Self Reg
    Logout Harbor

    Sign Up Should Not Display

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Self Reg Should Be Disabled
    Sleep  1

    # Restore setting
    Enable Self Reg
    Close Browser

Test Case - Admin Add New Users
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    # can add a new user when self-reg is enbled
    Switch To Configure
    Self Reg Should Be Enabled
    Switch to User Tag
    Add A New User   username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    # can add a new user when self-reg is disabled
    Switch To Configure
    Self Reg Should Be Enabled
    Sleep  1
    Disable Self Reg
    Self Reg Should Be Disabled
    Switch to User Tag
    Add A New User   username=tester2${d}  email=tester2${d}@vmware.com  realname=harbortest2  newPassword=Test1@34  comment=harbortest2
    Close Browser