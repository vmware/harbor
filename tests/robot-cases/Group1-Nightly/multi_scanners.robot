
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
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - Switch Scanner
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=  get current date  result_format=%m%s

    Switch To Scanners Page

    Should Display The Default Trivy Scanner

    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  admin  Harbor12345  project${d}  hello-world:latest
    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Scan Repo  latest  Succeed
    Move To Summary Chart
    Wait Until Page Contains  No vulnerability

    Switch To Scanners Page

    Set Default Scanner  Clair
    Should Display The Default Clair Scanner

    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Retry Wait Until Page Contains Element  ${not_scanned_icon}
    Scan Repo  latest  Succeed
    Move To Summary Chart
    Wait Until Page Contains  No vulnerability

    Switch To Scanners Page
    Set Default Scanner  Trivy
    Close Browser