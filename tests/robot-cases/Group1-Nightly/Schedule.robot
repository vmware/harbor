
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
Test Case - Scan Schedule Job
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%M
    Log To Console  ${d}
    ${project_name}=  Set Variable  scan_schedule_proj${d}
    ${image}=  Set Variable  redis
    ${tag}=  Set Variable  latest
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${project_name}
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}:${tag}
    Sleep  50
    Go Into Repo  ${project_name}/${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}
    Switch To Vulnerability Page
    ${flag}=  Set Variable  ${false}
    :FOR    ${i}    IN RANGE    999999
    \    ${minite}=  Get Current Date  result_format=%M
    \    ${minite_int} =  Convert To Integer  ${minite}
    \    ${left} =  Evaluate 	${minite_int}%10
    \    Log To Console    ${i}/${left}
    \    Sleep  55
    \    Run Keyword If  ${left} <= 3 and ${left} != 0   Run Keywords  Set Scan Schedule  custom  value=* */10 * * * *  AND  Set Suite Variable  ${flag}  ${true}
    \    Exit For Loop If    '${flag}' == '${true}'
    # After scan custom schedule is set, image should stay in unscanned status.
    Log To Console  Sleep for 300 seconds......
    Sleep  300
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}

    Log To Console  Sleep for 500 seconds......
    Sleep  500
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image}
    Scan Result Should Display In List Row  ${tag}
    View Repo Scan Details  High  Medium

Test Case - Replication Schedule Job
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%M
    Log To Console  ${d}
    ${project_name}=  Set Variable  replication_schedule_proj${d}
    ${image_a}=  Set Variable  mariadb
    ${tag_a}=  Set Variable  111
    ${image_b}=  Set Variable  centos
    ${tag_b}=  Set Variable  222
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${project_name}
    Switch To Registries
    Create A New Endpoint    docker-hub    e${d}    https://hub.docker.com/    danfengliu    Aa123456    Y
    Switch To Replication Manage
    ${flag}=  Set Variable  ${false}
    :FOR    ${i}    IN RANGE    999999
    \    ${minite}=  Get Current Date  result_format=%M
    \    ${minite_int} =  Convert To Integer  ${minite}
    \    ${left} =  Evaluate 	${minite_int}%10
    \    Log To Console    ${i}/${left}
    \    Run Keyword If  ${left} <= 3 and ${left} != 0   Run Keywords  Create A Rule With Existing Endpoint    rule${d}    pull    danfengliu/*    image    e${d}    ${project_name}  mode=Scheduled  cron=* */10 * * * *  AND  Set Suite Variable  ${flag}  ${true}
    \    Sleep  40
    \    Exit For Loop If    '${flag}' == '${true}'

    # After replication schedule is set, project should contain 2 images.
    Log To Console  Sleep for 720 seconds......
    Sleep  720
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_a}
    Artifact Exist  ${tag_a}
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_b}
    Artifact Exist  ${tag_b}

    # Delete repository
    Go Into Project  ${project_name}
    Delete Repo  ${project_name}/${image_a}
    Delete Repo  ${project_name}/${image_b}

    # After replication schedule is set, project should contain 2 images.
    Log To Console  Sleep for 600 seconds......
    Sleep  600
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_a}
    Artifact Exist  ${tag_a}
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_b}
    Artifact Exist  ${tag_b}
