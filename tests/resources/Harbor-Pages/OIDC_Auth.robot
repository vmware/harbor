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
Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***

Sign In Harbor With OIDC User
    [Arguments]  ${url}  ${username}=${OIDC_USERNAME}  ${password}=password  ${is_onboard}=${false}  ${username_claim}=${null}  ${login_with_provider}=email
    ${full_name}=   Set Variable If  '${login_with_provider}' == 'email'  ${username}@example.com  ${username}
    ${head_username}=   Set Variable If  '${username_claim}' == 'email'  xpath=//harbor-app/harbor-shell/clr-main-container/navigator/clr-header//clr-dropdown//button[contains(.,'${full_name}')]  xpath=//harbor-app/harbor-shell/clr-main-container/navigator/clr-header//clr-dropdown//button[contains(.,'${username}')]
    Init Chrome Driver
    Go To    ${url}
    Retry Element Click    ${log_oidc_provider_btn}
    Run Keyword If  '${login_with_provider}' == 'email'  Retry Element Click  ${login_with_email_btn}
    Run Keyword If  '${login_with_provider}' == 'ldap'   Retry Element Click  ${login_with_ldap_btn}
    Retry Text Input    ${dex_login_btn}    ${full_name}
    Retry Text Input    ${dex_pwd_btn}    ${password}
    Retry Element Click    ${submit_login_btn}
    Retry Element Click    ${grant_btn}

    #If input box for harbor user name is visible, it means it's the 1st time login of this user,
    #  but if this user has been logged into harbor successfully, this input box will not show up,
    #  so there is condition branch for this stituation.
    ${isVisible}=  Run Keyword And Return Status  Element Should Be Visible  ${oidc_username_input}
    Run Keyword If  ${is_onboard} == ${true}  Should Not Be True  ${isVisible}
    Run Keyword If  '${isVisible}' == 'True'  Run Keywords  Retry Text Input    ${oidc_username_input}    ${username}  AND  Retry Element Click    ${save_btn}
    Retry Wait Element  ${head_username}
    Capture Page Screenshot
    ${name_display}=  Get Text  ${header_user}
    Run Keyword If  '${username_claim}' == 'email'  Should Be Equal As Strings  ${name_display}  ${full_name}
    ...  ELSE    Should Be Equal As Strings  ${name_display}  ${username}

Get Secrete By API
    [Arguments]  ${url}  ${username}=${OIDC_USERNAME}
    ${json}=  Run Curl And Return Json  curl -s -k -X GET --header 'Accept: application/json' -u '${HARBOR_ADMIN}:${HARBOR_PASSWORD}' '${url}/api/v2.0/users/search?username=${username}'
    ${user_info}=    Set Variable    ${json[0]}
    ${user_id}=    Set Variable    ${user_info["user_id"]}
    ${json}=  Run Curl And Return Json   curl -s -k -X GET --header 'Accept: application/json' -u '${HARBOR_ADMIN}:${HARBOR_PASSWORD}' '${url}/api/v2.0/users/${user_id}'
    ${secret}=    Set Variable    ${json["oidc_user_meta"]["secret"]}
    [Return]  ${secret}

Generate And Return Secret
    [Arguments]  ${url}
    Retry Element Click  ${head_admin_xpath}
    Retry Element Click  ${user_profile_xpath}
    Retry Element Click  ${more_btn}
    Retry Element Click  ${generate_secret_btn}
    Retry Double Keywords When Error  Retry Element Click  ${confirm_btn}  Retry Wait Until Page Not Contains Element  ${confirm_btn}
    Retry Wait Until Page Contains  Cli secret setting is successful
    ${secret}=  Get Secrete By API  ${url}
    [Return]  ${secret}

Able To Delete An OIDC User
    Switch to User Tag
    Retry Element Click   ${user_test7_checkbox}
    Retry Element Click   ${member_action_btn}
    Retry Element Click   ${delete_btn}
    Retry Element Click   ${confirm_delete_btn}
    Retry Wait Until Page Not Contains Element   ${user_test7_checkbox}

Should Contain Target User
    Switch to User Tag
    Retry Wait Until Page Contains Element  ${user_test7_row}
