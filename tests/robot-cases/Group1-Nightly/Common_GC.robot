
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
Test Case - Garbage Collection
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world
    Sleep  2
    Go Into Project  project${d}
    Delete Repo  project${d}
    Sleep  2
    GC Now  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Retry GC Should Be Successful  1  0 blobs marked, 3 blobs and 0 manifests eligible for deletion
    Close Browser

Test Case - GC Untagged Images
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=    Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world  latest
    # make hello-world untagged
    Go Into Project   project${d}
    Go Into Repo   hello-world
    Go Into Artifact   latest
    Should Contain Tag   latest
    Delete A Tag   latest
    Should Not Contain Tag   latest
    # run gc without param delete untagged artifacts checked,  should not delete hello-world:latest
    GC Now  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Retry GC Should Be Successful  2  ${null}
    Go Into Project   project${d}
    Switch To Project Repo
    Go Into Repo   hello-world
    Should Contain Artifact
    # run gc with param delete untagged artifacts checked,  should delete hello-world
    Switch To Garbage Collection
    GC Now    ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  untag=${true}
    Retry GC Should Be Successful  3  ${null}
    Go Into Project   project${d}
    Switch To Project Repo
    Go Into Repo   hello-world
    Should Not Contain Any Artifact
    Close Browser

# Make sure image logstash was pushed to harbor for the 1st time, so GC will delete it.
Test Case - Project Quotas Control Under GC
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${storage_quota}=  Set Variable  200
    ${storage_quota_unit}=  Set Variable  MB
    ${image_a}=  Set Variable  logstash
    ${image_a_size}=    Set Variable    321.03MB
    ${image_a_ver}=  Set Variable  6.8.3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Capture Page Screenshot
    Create An New Project And Go Into Project  project${d}  storage_quota=${storage_quota}  storage_quota_unit=${storage_quota_unit}
    Capture Page Screenshot
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_a}:${image_a_ver}  err_msg=will exceed the configured upper limit of 200.0 MiB
    Capture Page Screenshot
    GC Now  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Retry GC Should Be Successful  4  ${null}
    @{param}  Create List  project${d}
    Retry Keyword When Return Value Mismatch  Get Project Storage Quota Text From Project Quotas List  0Byte of ${storage_quota}${storage_quota_unit}  60  @{param}
    Close Browser