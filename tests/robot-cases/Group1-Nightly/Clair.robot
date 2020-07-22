
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
Test Case - Clair Is Default Scanner And It Is Immutable
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Scanners Page
    Should Display The Default Clair Scanner
    Clair Is Immutable Scanner

Test Case - Disable Scan Schedule
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Disable Scan Schedule
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Retry Wait Until Page Contains  None
    Close Browser

Test Case - Scan A Tag In The Repo
    Body Of Scan A Tag In The Repo  hello-world  latest  is_no_vulerabilty=${true}

Test Case - Scan As An Unprivileged User
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world

    Sign In Harbor  ${HARBOR_URL}  user024  Test1@34
    Go Into Project  library
    Go Into Repo  hello-world
    Select Object  latest
    Scan Is Disabled
    Close Browser

# Chose a empty Vul repo
Test Case - Scan Image With Empty Vul
    Body Of Scan Image With Empty Vul  busybox  latest

Test Case - Manual Scan All
    Body Of Manual Scan All  Low  High  Medium  Negligible

Test Case - View Scan Error
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user026  Test1@34
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  user026  Test1@34  project${d}  vmware/photon:1.0
    Go Into Project  project${d}
    Go Into Repo  project${d}/vmware/photon
    Scan Repo  1.0  Fail
    View Scan Error Log
    Close Browser

Test Case - Scan Image On Push
    [Tags]  run-once
    Body Of Scan Image On Push  Low  High  Medium  Negligible

Test Case - View Scan Results
    [Tags]  run-once
    Body Of View Scan Results  Critical  High  Medium

Test Case - Project Level Image Serverity Policy
    [Tags]  run-once
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=  get current date  result_format=%m%s
    #For docker-hub registry
    #${sha256}=  Set Variable  9755880356c4ced4ff7745bafe620f0b63dd17747caedba72504ef7bac882089
    #For internal CPE harbor registry
    ${sha256}=  Set Variable  0e67625224c1da47cb3270e7a861a83e332f708d3d89dde0cbed432c94824d9a
    ${image}=  Set Variable  redis
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  sha256=${sha256}
    Go Into Project  project${d}
    Go Into Repo  ${image}
    Scan Repo  ${sha256}  Succeed
    Navigate To Projects
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  3
    Cannot Pull Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  tag=${sha256}  err_msg=To continue with pull, please contact your project administrator to exempt matched vulnerabilities through configuring the CVE allowlist
    Close Browser

#Important Note: All CVE IDs in CVE Allowlist cases must unique!
Test Case - Verfiy System Level CVE Allowlist
    Body Of Verfiy System Level CVE Allowlist  mariadb  b5e273ed46d2b5a1c96bf8f3ae37aa5e90c6c481e7f7ae66744610d7df79cbd1  CVE-2019-13050\nCVE-2018-19591\nCVE-2018-11236\nCVE-2018-11237\nCVE-2019-13627\nCVE-2018-20839\nCVE-2019-2923\nCVE-2019-2922\nCVE-2019-2911\nCVE-2019-2914\nCVE-2019-2924\nCVE-2019-2910\nCVE-2019-2938\nCVE-2019-2993\nCVE-2019-2974\nCVE-2019-2960\nCVE-2019-2948\nCVE-2019-2946  CVE-2019-2969

Test Case - Verfiy Project Level CVE Allowlist
    Body Of Verfiy Project Level CVE Allowlist  mariadb  b5e273ed46d2b5a1c96bf8f3ae37aa5e90c6c481e7f7ae66744610d7df79cbd1  CVE-2019-13050\nCVE-2018-19591\nCVE-2018-11236\nCVE-2018-11237\nCVE-2019-13627\nCVE-2018-20839\nCVE-2019-2923\nCVE-2019-2922\nCVE-2019-2911\nCVE-2019-2914\nCVE-2019-2924\nCVE-2019-2910\nCVE-2019-2938\nCVE-2019-2993\nCVE-2019-2974\nCVE-2019-2960\nCVE-2019-2948\nCVE-2019-2946  CVE-2019-2969

Test Case - Verfiy Project Level CVE Allowlist By Quick Way of Add System
    Body Of Verfiy Project Level CVE Allowlist By Quick Way of Add System  mariadb  b5e273ed46d2b5a1c96bf8f3ae37aa5e90c6c481e7f7ae66744610d7df79cbd1  CVE-2019-13050\nCVE-2018-19591\nCVE-2018-11236\nCVE-2018-11237\nCVE-2019-13627\nCVE-2018-20839\nCVE-2019-2923\nCVE-2019-2922\nCVE-2019-2911\nCVE-2019-2914\nCVE-2019-2924\nCVE-2019-2910\nCVE-2019-2938\nCVE-2019-2993\nCVE-2019-2974\nCVE-2019-2960\nCVE-2019-2948\nCVE-2019-2946\nCVE-2019-2969
