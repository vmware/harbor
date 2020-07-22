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
Go Into Project
    [Arguments]  ${project}  ${has_image}=${true}
    Retry Keyword N Times When Error  5  Clear Search Input And Go Into Project  ${project}  ${has_image}

Clear Search Input And Go Into Project
    [Arguments]  ${project}  ${has_image}
    # Close prompt in header
    Sleep  2
    #go To  ${url}
    Reload Page
    Sleep  2
    # Clear Search Input
    Retry Element Click  xpath=//harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[1]/a/span
    Sleep  2
    Retry Text Input  ${search_input}  ${project}
    Sleep  2
    # Go Into Project
    ${out}  Run Keyword If  ${has_image}==${false}  Run Keywords  Retry Element Click  xpath=//*[@id='project-results']//clr-dg-cell[contains(.,'${project}')]/a  AND  Wait Until Element Is Visible And Enabled  xpath=//clr-dg-placeholder[contains(.,\"We couldn\'t find any repositories!\")]
    ...  ELSE  Run Keywords  Retry Element Click  xpath=//*[@id='project-results']//clr-dg-cell[contains(.,'${project}')]/a  AND  Wait Until Element Is Visible And Enabled  xpath=//project-detail//hbr-repository-gridview//clr-dg-cell[contains(.,'${project}/')]
    Sleep  1

Add User To Project Admin
    [Arguments]  ${project}  ${user}
    # *** this keyword has not been used ***
    Go Into Project
    Retry Element Click  xpath=${project_member_tag_xpath}
    Retry Element Click  xpath=${project_member_add_button_xpath}
    Retry Text Input  xpath=${project_member_add_username_xpath}  ${user}
    Retry Element Click  xpath=${project_member_add_admin_xpath}
    Retry Element Click  xpath=${project_member_add_save_button_xpath}
    Sleep  4

Search Project Member
    [Arguments]  ${project}  ${user}
    # *** this keyword has not been used ***
    Go Into Project  ${project}
    Retry Element Click  xpath=//clr-dg-cell//a[contains(.,'${project}')]
    Retry Element Click  xpath=${project_member_search_button_xpath}
    Retry Element Click  xpath=${project_member_search_text_xpath}
    Retry Wait Until Page Contains  ${user}

Change Project Member Role
    [Arguments]  ${project}  ${user}  ${role}
    Retry Element Click  xpath=//clr-dg-cell//a[contains(.,'${project}')]
    Retry Element Click  xpath=${project_member_tag_xpath}
    Retry Element Click  xpath=//project-detail//clr-dg-row[contains(.,'${user}')]//clr-checkbox-wrapper
    #change role
    Retry Element Click  ${project_member_action_xpath}
    Retry Element Click  //button[contains(.,'${role}')]
    Retry Wait Until Page Not Contains Element  ${project_member_set_role_xpath}
    #Precondition is that only 1 member is in the list.
    Retry Wait Until Page Contains  ${role}

User Can Change Role
     [arguments]  ${username}
     Retry Element Click  xpath=//clr-dg-row[contains(.,'${username}')]//input/../label
     Retry Element Click  xpath=//*[@id='member-action']
     Retry Wait Until Page Not Contains Element  xpath=//button[@disabled='' and contains(.,'Admin')]

User Can Not Change Role
     [arguments]  ${username}
     Retry Element Click  xpath=//clr-dg-row[contains(.,'${username}')]//input/../label
     Retry Element Click  xpath=//*[@id='member-action']
     Retry Wait Until Page Contains Element  xpath=//button[@disabled='' and contains(.,'Admin')]

#this keyworkd seems will not use any more, will delete in the future
Non-admin View Member Account
    [arguments]  ${times}
    Xpath Should Match X Times  //clr-dg-row-maintainer  ${times}

User Can Not Add Member
    Retry Wait Until Page Contains Element  xpath=//button[@disabled='' and contains(.,'User')]

Add Guest Member To Project
    [arguments]  ${member}
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_member_add_button_xpath}  Retry Wait Until Page Contains Element  xpath=${project_member_add_button_xpath}
    Retry Text Input  xpath=${project_member_add_username_xpath}  ${member}
    #select guest
    Mouse Down  xpath=${project_member_guest_radio_checkbox}
    Mouse Up  xpath=${project_member_guest_radio_checkbox}
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_member_add_confirmation_ok_xpath}  Retry Wait Until Page Not Contains Element  xpath=${project_member_add_confirmation_ok_xpath}

Delete Project Member
    [arguments]  ${member}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${member}')]//input/../label
    Retry Double Keywords When Error  Retry Element Click  ${member_action_xpath}  Retry Wait Until Page Contains Element  ${delete_action_xpath}
    Retry Double Keywords When Error  Retry Element Click  ${delete_action_xpath}  Retry Wait Until Page Contains Element  ${repo_delete_on_card_view_btn}
    Retry Double Keywords When Error  Retry Element Click  ${repo_delete_on_card_view_btn}  Retry Wait Element  xpath=${project_member_xpath}
    Sleep  1

User Should Be Owner Of Project
    [Arguments]  ${user}  ${pwd}  ${project}
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Go Into Project  ${project}
    Switch To Member
    User Can Not Change Role  ${user}
    Push image  ${ip}  ${user}  ${pwd}  ${project}  hello-world
    Logout Harbor

User Should Not Be A Member Of Project
    [Arguments]  ${user}  ${pwd}  ${project}  ${is_oidc_mode}=${false}
    Run Keyword If  ${is_oidc_mode} == ${false}  Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    ...    ELSE  Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${user}
    ${pwd_oidc}=  Run Keyword And Return If  ${is_oidc_mode} == ${true}  Get Secrete By API  ${HARBOR_URL}
    ${password}=  Set Variable If  ${is_oidc_mode} == ${true}  ${pwd_oidc}  ${pwd}
    Project Should Not Display  ${project}
    Logout Harbor
    Cannot Pull Image  ${ip}  ${user}  ${password}  ${project}  ${ip}/${project}/hello-world
    Cannot Pull Image  ${ip}  ${user}  ${password}  ${project}  hello-world

Manage Project Member
    [Arguments]  ${admin}  ${pwd}  ${project}  ${user}  ${op}  ${has_image}=${true}  ${is_oidc_mode}=${false}
    Run Keyword If  ${is_oidc_mode} == ${false}  Sign In Harbor  ${HARBOR_URL}  ${admin}  ${pwd}
    ...    ELSE  Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${admin}
    Go Into Project  ${project}  ${has_image}
    Switch To Member
    Run Keyword If  '${op}' == 'Add'  Add Guest Member To Project  ${user}
    ...    ELSE IF  '${op}' == 'Remove'  Delete Project Member  ${user}
    ...    ELSE  Change Project Member Role  ${project}  ${user}  ${role}
    Logout Harbor

Change User Role In Project
    [Arguments]  ${admin}  ${pwd}  ${project}  ${user}  ${role}  ${is_oidc_mode}=${false}
    Run Keyword If  ${is_oidc_mode} == ${false}  Sign In Harbor   ${HARBOR_URL}  ${admin}  ${pwd}
    ...    ELSE  Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${admin}
    Retry Wait Element Visible  //clr-dg-cell//a[contains(.,'${project}')]
    Change Project Member Role  ${project}  ${user}  ${role}
    Logout Harbor

User Should Be Guest
    [Arguments]  ${user}  ${pwd}  ${project}  ${is_oidc_mode}=${false}
    Run Keyword If  ${is_oidc_mode} == ${false}  Sign In Harbor   ${HARBOR_URL}  ${user}  ${pwd}
    ...    ELSE  Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${user}
    ${pwd_oidc}=  Run Keyword And Return If  ${is_oidc_mode} == ${true}  Get Secrete By API  ${HARBOR_URL}
    ${password}=  Set Variable If  ${is_oidc_mode} == ${true}  ${pwd_oidc}  ${pwd}
    Project Should Display  ${project}
    Go Into Project  ${project}
    Switch To Member
    User Can Not Add Member
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'${user}')]//clr-dg-cell[contains(.,'Guest')]
    Logout Harbor
    Pull image  ${ip}  ${user}  ${password}  ${project}  hello-world
    Cannot Push image  ${ip}  ${user}  ${password}  ${project}  hello-world

User Should Be Developer
    [Arguments]  ${user}  ${pwd}  ${project}  ${is_oidc_mode}=${false}
    Run Keyword If  ${is_oidc_mode} == ${false}  Sign In Harbor   ${HARBOR_URL}  ${user}  ${pwd}
    ...    ELSE  Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${user}
    ${pwd_oidc}=  Run Keyword And Return If  ${is_oidc_mode} == ${true}  Get Secrete By API  ${HARBOR_URL}
    ${password}=  Set Variable If  ${is_oidc_mode} == ${true}  ${pwd_oidc}  ${pwd}
    Project Should Display  ${project}
    Go Into Project  ${project}
    Switch To Member
    User Can Not Add Member
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'${user}')]//clr-dg-cell[contains(.,'Developer')]
    Logout Harbor
    Push Image With Tag  ${ip}  ${user}  ${password}  ${project}  hello-world  v1

User Should Be Admin
    [Arguments]  ${user}  ${pwd}  ${project}  ${guest}  ${is_oidc_mode}=${false}
    Run Keyword If  ${is_oidc_mode} == ${false}  Sign In Harbor   ${HARBOR_URL}  ${user}  ${pwd}
    ...    ELSE  Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${user}
    ${pwd_oidc}=  Run Keyword And Return If  ${is_oidc_mode} == ${true}  Get Secrete By API  ${HARBOR_URL}
    ${password}=  Set Variable If  ${is_oidc_mode} == ${true}  ${pwd_oidc}  ${pwd}
    Project Should Display  ${project}
    Go Into Project  ${project}
    Switch To Member
    Add Guest Member To Project  ${guest}
    User Can Change Role  ${guest}
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'${user}')]//clr-dg-cell[contains(.,'Admin')]
    Logout Harbor
    Push Image With Tag  ${ip}  ${user}  ${password}  ${project}  hello-world  v2

User Should Be Maintainer
    [Arguments]  ${user}  ${pwd}  ${project}  ${is_oidc_mode}=${false}
    Run Keyword If  ${is_oidc_mode} == ${false}  Sign In Harbor   ${HARBOR_URL}  ${user}  ${pwd}
    ...    ELSE  Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${user}
    ${pwd_oidc}=  Run Keyword And Return If  ${is_oidc_mode} == ${true}  Get Secrete By API  ${HARBOR_URL}
    ${password}=  Set Variable If  ${is_oidc_mode} == ${true}  ${pwd_oidc}  ${pwd}
    Project Should Display  ${project}
    Go Into Project  ${project}
    Delete Repo  ${project}
    Switch To Member
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'${user}')]//clr-dg-cell[contains(.,'Maintainer')]
    Logout Harbor
    Push Image With Tag  ${ip}  ${user}  ${password}  ${project}  hello-world  v3

Project Should Have Member
    [Arguments]  ${project}  ${user}
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Go Into Project  ${project}
    Switch To Member
    Retry Wait Until Page Contains Element  xpath=//clr-dg-cell[contains(., '${user}')]
    Logout Harbor
