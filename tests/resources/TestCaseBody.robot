# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource wrap test case body

*** Variables ***

*** Keywords ***
Body Of Manage project publicity
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Create An New Project And Go Into Project  project${d}  public=true

    Push image  ${ip}  user007  Test1@34  project${d}  hello-world:latest
    Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Search Private Projects
    Project Should Not Display  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Private  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Not Display  project${d}
    Cannot Pull Image  ${ip}  user008  Test1@34  project${d}  hello-world:latest  err_msg=unauthorized to access repository

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Close Browser

Body Of Scan A Tag In The Repo
    [Arguments]  ${image_argument}  ${tag_argument}  ${is_no_vulerabilty}=${false}
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user023  Test1@34
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  user023  Test1@34  project${d}  ${image_argument}:${tag_argument}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image_argument}
    Scan Repo  ${tag_argument}  Succeed
    Scan Result Should Display In List Row  ${tag_argument}  is_no_vulerabilty=${is_no_vulerabilty}
    Pull Image  ${ip}  user023  Test1@34  project${d}  ${image_argument}  ${tag_argument}
    # Edit Repo Info
    Close Browser

Body Of Scan Image With Empty Vul
    [Arguments]  ${image_argument}  ${tag_argument}
    Init Chrome Driver
    ${tag}=  Set Variable  ${tag_argument}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  ${image_argument}:${tag_argument}
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Go Into Repo  ${image_argument}
    Scan Repo  ${tag}  Succeed
    Move To Summary Chart
    Scan Result Should Display In List Row  ${tag}  is_no_vulerabilty=${true}
    Close Browser

Body Of Manual Scan All
    [Arguments]  @{vulnerability_levels}
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  redis
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Trigger Scan Now And Wait Until The Result Appears
    Navigate To Projects
    Go Into Project  library
    Go Into Repo  redis
    Scan Result Should Display In List Row  latest
    View Repo Scan Details  @{vulnerability_levels}
    Close Browser

Body Of View Scan Results
    [Arguments]  @{vulnerability_levels}
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user025  Test1@34
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  user025  Test1@34  project${d}  tomcat
    Go Into Project  project${d}
    Go Into Repo  project${d}/tomcat
    Scan Repo  latest  Succeed
    Scan Result Should Display In List Row  latest
    View Repo Scan Details  @{vulnerability_levels}
    Close Browser

Body Of Scan Image On Push
    [Arguments]  @{vulnerability_levels}
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Goto Project Config
    Enable Scan On Push
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  memcached
    Navigate To Projects
    Go Into Project  library
    Go Into Repo  memcached
    Scan Result Should Display In List Row  latest
    View Repo Scan Details  @{vulnerability_levels}
    Close Browser

Body Of List Helm Charts
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user027  Test1@34
    Create An New Project And Go Into Project  project${d}

    Switch To Project Charts
    Upload Chart files
    Go Into Chart Version  ${prometheus_chart_name}
    Retry Wait Until Page Contains  ${prometheus_chart_version}
    Go Into Chart Detail  ${prometheus_chart_version}

    # Summary tab
    Retry Wait Until Page Contains Element  ${summary_markdown}
    Retry Wait Until Page Contains Element  ${summary_container}

    # Dependency tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_dependency}  Retry Wait Until Page Contains Element  ${dependency_content}

    # Values tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_value}  Retry Wait Until Page Contains Element  ${value_content}

    Go Into Project  project${d}  has_image=${false}
    Switch To Project Charts
    Multi-delete Chart Files  ${prometheus_chart_name}  ${harbor_chart_name}
    Close Browser

Body Of Admin Push Signed Image
    [Arguments]  ${image}=tomcat  ${project}=library  ${with_remove}=${false}
    Enable Notary Client

    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/notary-push-image.sh ${ip} ${project} ${image} latest ${notaryServerEndpoint} ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:latest
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/projects/${project}/repositories/${image}/artifacts/latest?with_signature=true"

    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  "signed":true

    Run Keyword If  ${with_remove} == ${true}  Remove Notary Signature  ${ip}  ${image}

Delete A Project Without Sign In Harbor
    [Arguments]  ${harbor_ip}=${ip}  ${username}=${HARBOR_ADMIN}  ${password}=${HARBOR_PASSWORD}
    ${d}=    Get Current Date    result_format=%m%s
    ${project_name}=  Set Variable  000${d}
    Create An New Project And Go Into Project  ${project_name}
    Push Image  ${harbor_ip}  ${username}  ${password}  ${project_name}  hello-world
    Project Should Not Be Deleted  ${project_name}
    Go Into Project  ${project_name}
    Delete Repo  ${project_name}
    Navigate To Projects
    Project Should Be Deleted  ${project_name}

Manage Project Member Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}  ${test_user1}=user005  ${test_user2}=user006  ${is_oidc_mode}=${false}
    ${d}=    Get current Date  result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Push image  ip=${ip}  user=${sign_in_user}  pwd=${sign_in_pwd}  project=project${d}  image=hello-world
    Logout Harbor

    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Add  is_oidc_mode=${is_oidc_mode}
    User Should Be Guest  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Developer  is_oidc_mode=${is_oidc_mode}
    User Should Be Developer  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Admin  is_oidc_mode=${is_oidc_mode}
    User Should Be Admin  ${test_user1}  ${sign_in_pwd}  project${d}  ${test_user2}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Maintainer  is_oidc_mode=${is_oidc_mode}
    User Should Be Maintainer  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Remove  is_oidc_mode=${is_oidc_mode}
    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}    is_oidc_mode=${is_oidc_mode}
    Push image  ip=${ip}  user=${sign_in_user}  pwd=${sign_in_pwd}  project=project${d}  image=hello-world
    User Should Be Guest  ${test_user2}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}

Helm CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Helm Repo Add  ${HARBOR_URL}  ${sign_in_user}  ${sign_in_pwd}  project_name=project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}
    Switch To Project Charts
    Go Into Chart Version  ${harbor_chart_name}
    Retry Wait Until Page Contains  ${harbor_chart_version}
    Capture Page Screenshot

Helm3 CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}  helm_repo_name=${HARBOR_URL}/chartrepo/project${d}  helm_cmd=helm3
    Switch To Project Charts
    Retry Double Keywords When Error  Go Into Chart Version  ${harbor_chart_name}  Retry Wait Until Page Contains  ${harbor_chart_version}
    Capture Page Screenshot

#Important Note: All CVE IDs in CVE Allowlist cases must unique!
Body Of Verfiy System Level CVE Allowlist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    # ${image}=    Set Variable    goharbor/harbor-portal
    ${sha256}=  Set Variable  ${sha256_argument}
    # ${sha256}=  Set Variable  2cb6a1c24dd6b88f11fd44ccc6560cb7be969f8ac5f752802c99cae6bcd592bb
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}  err_msg=current image without vulnerability scanning cannot be pulled due to configured policy
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To Configuration System Setting
    # Add Items To System CVE Allowlist    CVE-2019-19317\nCVE-2019-19646 \nCVE-2019-5188 \nCVE-2019-20387 \nCVE-2019-17498 \nCVE-2019-20372 \nCVE-2019-19244 \nCVE-2019-19603 \nCVE-2019-19880 \nCVE-2019-19923 \nCVE-2019-19925 \nCVE-2019-19926 \nCVE-2019-19959 \nCVE-2019-20218 \nCVE-2019-19232 \nCVE-2019-19234 \nCVE-2019-19645
    Add Items To System CVE Allowlist    ${most_cve_list}
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}  err_msg=cannot be pulled due to configured policy
    # Add Items To System CVE Allowlist    CVE-2019-18276
    Add Items To System CVE Allowlist    ${single_cve}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In System CVE Allowlist  count=6
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}  err_msg=cannot be pulled due to configured policy
    Close Browser

Body Of Verfiy Project Level CVE Allowlist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Go Into Project  project${d}
    Add Items to Project CVE Allowlist    ${most_cve_list}
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add Items to Project CVE Allowlist    ${single_cve}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In Project CVE Allowlist
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser

Body Of Verfiy Project Level CVE Allowlist By Quick Way of Add System
    [Arguments]  ${image_argument}  ${sha256_argument}  ${cve_list}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To Configuration System Setting
    Add Items To System CVE Allowlist    ${cve_list}
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Project To Project Level CVE Allowlist
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add System CVE Allowlist to Project CVE Allowlist By Add System Button Click
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser
