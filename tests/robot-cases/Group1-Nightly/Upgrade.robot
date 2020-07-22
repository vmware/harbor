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
Test Case - Manage project publicity
    Body Of Manage project publicity

Test Case - Scan A Tag In The Repo
    [Tags]  clair
    Body Of Scan A Tag In The Repo  hello-world  latest  is_no_vulerabilty=${true}

Test Case - Scan A Tag In The Repo
    [Tags]  trivy
    Body Of Scan A Tag In The Repo  vmware/photon  1.0

Test Case - List Helm Charts
    Body Of List Helm Charts

Test Case - Admin Push Signed Image
    Body Of Admin Push Signed Image
