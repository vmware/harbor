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
Library  ../../apitests/python/testutils.py
Library  ../../apitests/python/library/repository.py
Resource  ../../resources/Util.robot
Default Tags  Replication

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin
${SERVER}  ${ip}
${SERVER_URL}  https://${SERVER}
${SERVER_API_ENDPOINT}  ${SERVER_URL}/api
&{SERVER_CONFIG}  endpoint=${SERVER_API_ENDPOINT}  verify_ssl=False
${REMOTE_SERVER}  ${ip1}
${REMOTE_SERVER_URL}  https://${REMOTE_SERVER}
${REMOTE_SERVER_API_ENDPOINT}  ${REMOTE_SERVER_URL}/api

*** Test Cases ***
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - Pro Replication Rules Add
    Init Chrome Driver
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Switch To Replication Manage
    Check New Rule UI Without Endpoint
    Close Browser

Test Case - Harbor Endpoint Verification
    #This case need vailid info and selfsign cert
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint    harbor    edp1${d}    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    N
    Endpoint Is Pingable
    Enable Certificate Verification
    Endpoint Is Unpingable
    Close Browser

##Test Case - DockerHub Endpoint Add
    #This case need vailid info and selfsign cert
    ##Init Chrome Driver
    ##${d}=    Get Current Date    result_format=%m%s
    ##Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    ##Switch To Registries
    ##Create A New Endpoint    docker-hub    edp1${d}    https://hub.docker.com/    ${DOCKER_USER}    ${DOCKER_PWD}    Y
    ##Close Browser

Test Case - Harbor Endpoint Add
    #This case need vailid info and selfsign cert
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint    harbor    testabc    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    Y
    Close Browser

Test Case - Harbor Endpoint Edit
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Rename Endpoint  testabc  deletea
    Retry Wait Until Page Contains  deletea
    Close Browser

Test Case - Harbor Endpoint Delete
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Delete Endpoint  deletea
    Delete Success  deletea
    Close Browser

Test Case - Replication Rule Edit
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${endpoint1}=    Set Variable    e1${d}
    ${endpoint2}=    Set Variable    e2${d}
    ${rule_name_old}=    Set Variable    rule_testabc${d}
    ${rule_name_new}=    Set Variable    rule_abctest${d}
    ${resource_type}=    Set Variable    chart
    ${dest_namespace}=    Set Variable    dest_namespace${d}
    ${mode}=    Set Variable    Scheduled
    ${cron_str}=    Set Variable    10 10 10 * * *
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    #Due to docker-hub access limitation, remove docker-hub endpoint
    Create A New Endpoint    harbor    ${endpoint1}    https://cicd.harbor.vmwarecna.net    ${null}    ${null}    Y
    Create A New Endpoint    harbor    ${endpoint2}    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    ${rule_name_old}    pull    nightly/a*    image    ${endpoint1}    project${d}
    Edit Replication Rule  ${rule_name_old}
    #  Change rule-name, source-registry, filter, trigger-mode for edition verification
    Clear Field Of Characters    ${rule_name_input}    30
    Retry Text Input    ${rule_name_input}    ${rule_name_new}
    Select Source Registry  ${endpoint2}
    #Source Resource Filter
    Retry Text Input  ${filter_name_id}  project${d}
    Select From List By Value  ${rule_resource_selector}  ${resource_type}
    Retry Text Input  ${dest_namespace_xpath}  ${dest_namespace}
    Select Trigger  ${mode}
    Retry Text Input  ${targetCron_id}  ${cron_str}
    Retry Double Keywords When Error    Retry Element Click    ${rule_save_button}    Retry Wait Until Page Not Contains Element    ${rule_save_button}
    #  verify all items were changed as expected
    Edit Replication Rule    ${rule_name_new}
    Retry Textfield Value Should Be    ${rule_name_input}               ${rule_name_new}
    Retry List Selection Should Be     ${src_registry_dropdown_list}    ${endpoint2}-https://${ip}
    Retry Textfield Value Should Be    ${filter_name_id}                project${d}
    Retry Textfield Value Should Be    ${dest_namespace_xpath}          ${dest_namespace}
    Retry List Selection Should Be     ${rule_resource_selector}        ${resource_type}
    Retry List Selection Should Be     ${rule_trigger_select}           ${mode}
    Retry Textfield Value Should Be    ${targetCron_id}                 ${cron_str}
    Retry Element Click  ${rule_cancel_btn}
    Delete Replication Rule  ${rule_name_new}
    Close Browser

Test Case - Replication Rule Delete
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${endpoint1}=    Set Variable    e1${d}
    ${rule_name}=    Set Variable    rule_testabc${d}
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint    harbor    ${endpoint1}    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    ${rule_name}    pull    ${DOCKER_USER}/*    image    ${endpoint1}    project${d}
    Delete Replication Rule  ${rule_name}
    Close Browser

Test Case - Replication Of Pull Images from DockerHub To Self
    @{target_images}=  Create List  mariadb  centos
    &{image1_with_tag}=	 Create Dictionary  image=centos  tag=1.0
    &{image2_with_tag}=	 Create Dictionary  image=mariadb  tag=latest
    ${image1}=  Get From Dictionary  ${image1_with_tag}  image
    ${image1}=  Get Substring  ${image1}  0  -2
    Log All  image1:${image1}
    ${image2}=  Get From Dictionary  ${image2_with_tag}  image
    @{target_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'
    Body Of Replication Of Pull Images from Registry To Self   docker-hub  https://hub.docker.com/  ${DOCKER_USER}    ${DOCKER_PWD}  ${DOCKER_USER}/{${image1}*,${image2}}  ${null}  N  Flatten 1 Level  @{target_images}

Test Case - Replication Of Push Images from Self To Harbor
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    project${d}    hello-world
    Push Image    ${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    project${d}    busybox:latest
    Push Image With Tag    ${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    project${d}    hello-world    v1
    Switch To Registries
    Create A New Endpoint    harbor    e${d}    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    image    e${d}    project_dest${d}
    #logout and login target
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project_dest${d}
    #logout and login source
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule${d}
    Sleep  20
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Image Should Be Replicated To Project  project_dest${d}  hello-world
    Image Should Be Replicated To Project  project_dest${d}  busybox
    Close Browser

Test Case - Replication Of Push Chart from Self To Harbor
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Project Charts
    Upload Chart files
    Switch To Registries
    Create A New Endpoint    harbor    e${d}    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    chart    e${d}    project_dest${d}
    #logout and login target
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project_dest${d}
    #logout and login source
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate    rule${d}
    Sleep    20
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Go Into Project    project_dest${d}    has_image=${false}
    Switch To Project Charts
    Go Into Chart Version    ${harbor_chart_name}
    Retry Wait Until Page Contains    ${harbor_chart_version}
    Go Into Chart Detail    ${harbor_chart_version}
    Close Browser

Test Case - Replication Of Push Images from Self To Harbor By Push Event
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=   Set Variable    test_large_image
    ${image_size}=  Set Variable  4096
    ${tag1}=  Set Variable  large_f
    @{tags}   Create List  ${tag1}

    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Registries
    Create A New Endpoint    harbor    e${d}    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    image    e${d}    project_dest${d}
    ...    Event Based
    #logout and login target
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project_dest${d}
    Push Special Image To Project  project${d}  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${image}  tags=@{tags}  size=${image_size}
    # Use tag as identifier for this artifact
    Image Should Be Replicated To Project  project_dest${d}  ${image}  tag=${tag1}  expected_image_size_in_regexp=4(\\\.\\d{1,2})*GiB
    Close Browser

Test Case - Replication Of Pull Images from AWS-ECR To Self
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Registries
    Create A New Endpoint    aws-ecr    e${d}    us-east-2    ${ecr_ac_id}    ${ecr_ac_key}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    pull    a/*    image    e${d}    project${d}
    Select Rule And Replicate  rule${d}
    Image Should Be Replicated To Project  project${d}  httpd
    Image Should Be Replicated To Project  project${d}  alpine
    Image Should Be Replicated To Project  project${d}  hello-world
    Close Browser

Test Case - Replication Of Pull Images from Google-GCR To Self
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Registries
    Create A New Endpoint    google-gcr    e${d}    asia.gcr.io    ${null}    ${gcr_ac_key}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    pull    eminent-nation-87317/*    image    e${d}    project${d}
    Filter Replication Rule  rule${d}
    Select Rule And Replicate  rule${d}
    Image Should Be Replicated To Project  project${d}  httpd
    Image Should Be Replicated To Project  project${d}  tomcat
    Close Browser

Test Case - Replication Of Push Images to DockerHub Triggered By Event
    Body Of Replication Of Push Images to Registry Triggered By Event  docker-hub  https://hub.docker.com/  ${DOCKER_USER}  ${DOCKER_PWD}  ${DOCKER_USER}

#Due to issue of delete event replication
#Test Case - Replication Of Push Images to Google-GCR Triggered By Event
    #Body Of Replication Of Push Images to Registry Triggered By Event  google-gcr  gcr.io  ${null}  ${gcr_ac_key}  eminent-nation-87317/harbor-nightly-replication

Test Case - Replication Of Push Images to AWS-ECR Triggered By Event
    Body Of Replication Of Push Images to Registry Triggered By Event  aws-ecr  us-east-2  ${ecr_ac_id}  ${ecr_ac_key}  harbor-nightly-replication

Test Case - Replication Of Pull Images from Gitlab To Self
    &{image1_with_tag}=	 Create Dictionary  image=photon  tag=1.0
    &{image2_with_tag}=	 Create Dictionary  image=alpine  tag=latest
    ${image1}=  Get From Dictionary  ${image1_with_tag}  image
    ${image2}=  Get From Dictionary  ${image2_with_tag}  image
    @{target_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'
    Body Of Replication Of Pull Images from Registry To Self   gitlab   https://registry.gitlab.com    ${gitlab_id}    ${gitlab_key}    dannylunsa/test_replication/{${image1},${image2}}  ${null}  N  Flatten All Levels  @{target_images}

Test Case - Replication Of Push Images to Gitlab Triggered By Event
    Body Of Replication Of Push Images to Registry Triggered By Event    gitlab   https://registry.gitlab.com    ${gitlab_id}    ${gitlab_key}    dannylunsa/test_replication

Test Case - Replication Of Pull Manifest List and CNAB from Harbor To Self
    &{image1_with_tag}=	 Create Dictionary  image=busybox  tag=1.32.0  total_artifact_count=9  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=index101603308079  tag=index_tag101603308079  total_artifact_count=2  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=cnab011609785126  tag=cnab_tag011609785126  total_artifact_count=3  archive_count=2
    ${image1}=  Get From Dictionary  ${image1_with_tag}  image
    ${image2}=  Get From Dictionary  ${image2_with_tag}  image
    ${image3}=  Get From Dictionary  ${image3_with_tag}  image
    @{target_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Body Of Replication Of Pull Images from Registry To Self   harbor  https://cicd.harbor.vmwarecna.net  admin  qA5ZgV  nightly/{${image1},${image2},${image3}}  ${null}  Y  Flatten 1 Level  @{target_images}

Test Case - Image Namespace Level Flattening
    [tags]  flattening
    ${src_endpoint}=  Set Variable  ${ip1}

    #Test only for <Flatten All Levels>
    &{image1_with_tag}=	 Create Dictionary  image=test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  10  Flatten All Levels  /  @{src_images}

    #Test only for <Flatten 1 Level>
    &{image1_with_tag}=	 Create Dictionary  image=test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  10  Flatten 1 Level  ${null}  @{src_images}

    #Test only for <Flatten 2 Levels>
    &{image1_with_tag}=	 Create Dictionary  image=level_1/test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/level_3/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  10  Flatten 2 Levels  level_1/  @{src_images}

    #Test only for <Flatten 3 Levels>
    &{image1_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/level_2/level_3/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/level_3/level_4/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  20  Flatten 3 Levels  level_1/level_2/  @{src_images}
