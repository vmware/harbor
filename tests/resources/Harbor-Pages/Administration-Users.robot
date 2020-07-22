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
Assign User Admin
    [Arguments]  ${user}
    Retry Element Click  xpath=//harbor-user//hbr-filter//clr-icon
    Input Text  xpath=//harbor-user//hbr-filter//input  ${user}
    Sleep  2
    #select checkbox
    Retry Element Click  //clr-dg-row[contains(.,'${user}')]//label
    #click assign admin
    Retry Element Click  //*[@id='set-admin']
    Sleep  1

Switch to User Tag
    Retry Element Click  xpath=${administration_user_tag_xpath}
    Sleep  1

Administration Tag Should Display
    Retry Wait Until Page Contains Element  xpath=${administration_tag_xpath}

User Email Should Exist
    [Arguments]  ${email}
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch to User Tag
    Retry Wait Until Page Contains Element  xpath=//clr-dg-cell[contains(., '${email}')]

Add User Button Should Be Disabled
    Sleep  1
    Retry Wait Until Page Contains Element  //button[contains(.,'New') and @disabled='']

Add A New User
    [Arguments]   ${username}  ${email}  ${realname}  ${newPassword}  ${comment}
    Retry Element Click  xpath=${add_new_user_button}
    Retry Text Input  xpath=${username_xpath}  ${username}
    Retry Text Input  xpath=${email_xpath}  ${email}
    Retry Text Input  xpath=${realname_xpath}  ${realname}
    Retry Text Input  xpath=${newPassword_xpath}  ${newPassword}
    Retry Text Input  xpath=${confirmPassword_xpath}  ${newPassword}
    Retry Text Input  xpath=${comment_xpath}  ${comment}
    Retry Double Keywords When Error  Retry Element Click  xpath=${save_new_user_button}  Retry Wait Until Page Not Contains Element  xpath=${save_new_user_button}
    Retry Wait Until Page Contains Element  xpath=//harbor-user//clr-dg-row//clr-dg-cell[contains(., '${username}')]
