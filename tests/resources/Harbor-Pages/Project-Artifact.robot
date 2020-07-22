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

*** Keywords ***
Go Into Artifact
    [Arguments]  ${tag}
    Retry Wait Until Page Not Contains Element  ${artifact_list_spinner}
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${tag}')]//a[contains(.,'sha256')]
    Retry Wait Until Page Contains Element   ${artifact_tag_component}
Should Contain Tag
    [Arguments]  ${tag}
    Retry Wait Until Page Contains Element   xpath=//artifact-tag//clr-dg-row//clr-dg-cell[contains(.,'${tag}')]

Should Not Contain Tag
    [Arguments]  ${tag}
    Retry Wait Until Page Not Contains Element   xpath=//artifact-tag//clr-dg-row//clr-dg-cell[contains(.,'${tag}')]

Add A New Tag
    [Arguments]  ${tag}
    Retry Double Keywords When Error  Retry Element Click   ${add_tag_button}  Retry Wait Element  ${tag_name_xpath}
    Retry Text Input   ${tag_name_xpath}   ${tag}
    Retry Double Keywords When Error  Retry Element Click   ${add_ok_button}  Should Contain Tag  ${tag}

Delete A Tag
    [Arguments]  ${tag}
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${tag}')]//clr-checkbox-wrapper//label[contains(@class,'clr-control-label')]
    Retry Double Keywords When Error  Retry Element Click    ${delete_tag_button}  Retry Wait Until Page Contains Element  ${dialog_delete_button}
    Retry Double Keywords When Error  Retry Element Click  ${dialog_delete_button}  Should Not Contain Tag  ${tag}

Should Contain Artifact
    Retry Wait Until Page Contains Element   xpath=//artifact-list-tab//clr-dg-row//a[contains(.,'sha256')]

Should Not Contain Any Artifact
    Retry Wait Until Page Not Contains Element   xpath=//artifact-list-tab//clr-dg-row