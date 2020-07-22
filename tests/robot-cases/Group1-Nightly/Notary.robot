
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
Test Case - Project Level Policy Content Trust
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world:latest
    Go Into Project  project${d}
    Goto Project Config
    Click Content Trust
    Save Project Config
    # Verify
    # Unsigned image can not be pulled
    Content Trust Should Be Selected
    Cannot Pull Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world:latest  err_msg=The image is not signed in Notary
    # Signed image can be pulled
    Body Of Admin Push Signed Image  image=redis  project=project${d}
    Pull image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  redis  tag=latest
    Close Browser

Test Case - Admin Push Signed Image
    Body Of Admin Push Signed Image

Test Case - Admin Push Signed Image And Remove Signature
    Body Of Admin Push Signed Image  image=alpine  with_remove=${true}
