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
Create An New Project And Go Into Project
    [Arguments]  ${projectname}  ${public}=false  ${count_quota}=${null}  ${storage_quota}=${null}  ${storage_quota_unit}=${null}  ${proxy_cache}=${false}  ${registry}=${null}
    Navigate To Projects
    FOR  ${n}  IN RANGE  1  8
        ${out}  Run Keyword And Ignore Error  Retry Button Click  xpath=${create_project_button_xpath}
        Log All  Return value is ${out[0]}
        Exit For Loop If  '${out[0]}'=='PASS'
        Sleep  1
    END
    Log To Console  Project Name: ${projectname}
    Retry Text Input  xpath=${project_name_xpath}  ${projectname}
    ${element_project_public}=  Set Variable  xpath=${project_public_xpath}
    Run Keyword If  '${public}' == 'true'  Run Keywords  Wait Until Element Is Visible And Enabled  ${element_project_public}  AND  Retry Element Click  ${element_project_public}
    Run Keyword If  '${count_quota}'!='${null}'  Input Count Quota  ${count_quota}
    Run Keyword If  '${storage_quota}'!='${null}'  Input Storage Quota  ${storage_quota}  ${storage_quota_unit}
    Run Keyword If  '${proxy_cache}' == '${true}'  Run Keywords  Mouse Down  ${project_proxy_cache_switcher_id}  AND  Mouse Up  ${project_proxy_cache_switcher_id}  AND  Retry Element Click  ${project_registry_select_id}  AND  Retry Element Click  xpath=//select[@id='registry']//option[contains(.,'${registry}')]
    Retry Double Keywords When Error  Retry Element Click  ${create_project_OK_button_xpath}  Retry Wait Until Page Not Contains Element  ${create_project_OK_button_xpath}
    Sleep  2
    Go Into Project  ${projectname}  has_image=${false}

Create An New Project With New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}  ${projectname}  ${public}
    Create An New User  url=${url}  username=${username}  email=${email}  realname=${realname}  newPassword=${newPassword}  comment=${comment}
    Logout Harbor
    Sign In Harbor  ${url}  ${username}  ${newPassword}
    Create An New Project And Go Into Project  ${projectname}  ${public}
    Sleep  1

Artifact Exist
    [Arguments]  ${tag_name}
    Retry Wait Until Page Contains Element  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256') and contains(.,'${tag_name}')]
#It's the log of project.
Go To Project Log
    #Switch To Project Tab Overflow
    Retry Element Click  xpath=${project_log_xpath}
    Sleep  2

Switch To Member
    Sleep  3
    Retry Element Click  xpath=${project_member_xpath}
    Sleep  1

Switch To Log
    Retry Element Click  xpath=${log_xpath}
    Sleep  1

Switch To Replication
    Retry Element Click  xpath=${project_replication_xpath}
    Sleep  1

Switch To Project Configuration
    Retry Element Click  ${project_config_tabsheet}
    Sleep  1

Switch To Tag Retention
    #Switch To Project Tab Overflow
    Retry Element Click  xpath=${project_tag_strategy_xpath}
    Sleep  1

Switch To Tag Immutability
    #Switch To Project Tab Overflow
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_tag_strategy_xpath}  Retry Wait Until Page Contains Element  ${project_tag_immutability_switch}
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_tag_immutability_switch}  Retry Wait Until Page Contains  Immutability rules
    Sleep  1

Switch To Project Tab Overflow
    Retry Element Click  xpath=${project_tab_overflow_btn}
    Sleep  1

Navigate To Projects
    Reload Page
    Sleep  3
    Retry Element Click  xpath=${projects_xpath}
    Sleep  1

Project Should Display
    [Arguments]  ${projectname}
    Retry Wait Element  xpath=//projects//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

Project Should Not Display
    [Arguments]  ${projectname}
    Retry Wait Until Page Not Contains Element  xpath=//projects//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

Search Private Projects
    Retry Element Click  xpath=//select
    Retry Element Click  xpath=//select/option[@value=1]
    Sleep  1

Make Project Private
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Be Selected  ${project_config_public_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${project_config_public_checkbox_label}  Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}
    Retry Element Click  //button[contains(.,'SAVE')]
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}

Make Project Public
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${project_config_public_checkbox_label}  Retry Checkbox Should Be Selected  ${project_config_public_checkbox}
    Retry Element Click  //button[contains(.,'SAVE')]
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Be Selected  ${project_config_public_checkbox}

Repo Exist
    [Arguments]  ${pro_name}  ${repo_name}
    Retry Wait Until Page Contains Element  //clr-dg-row[contains(.,'${pro_name}/${repo_name}')]

Repo Not Exist
    [Arguments]  ${pro_name}  ${repo_name}
    Retry Wait Until Page Not Contains Element  //clr-dg-row[contains(.,'${pro_name}/${repo_name}')]

Filter Repo
    [Arguments]  ${pro_name}  ${repo_name}  ${exsit}=${true}
    Retry Double Keywords When Error  Retry Element Click  ${filter_dist_btn}  Wait Until Element Is Visible And Enabled  ${filter_dist_input}
    Retry Clear Element Text  ${filter_dist_input}
    Retry Text Input  ${filter_dist_input}  ${pro_name}/${repo_name}
    Run Keyword If  ${exsit}==${true}    Repo Exist  ${pro_name}  ${repo_name}
    ...  ELSE  Repo Not Exist  ${pro_name}  ${repo_name}

Delete Repo
    [Arguments]  ${pro_name}  ${repo_name}
    ${element_repo_checkbox}=  Set Variable  xpath=//clr-dg-row[contains(.,'${pro_name}/${repo_name}')]//div[contains(@class,'clr-checkbox-wrapper')]//label
    Filter Repo  ${pro_name}  ${repo_name}
    Retry Double Keywords When Error  Retry Element Click  ${element_repo_checkbox}  Wait Until Element Is Visible And Enabled  ${repo_delete_btn}
    Retry Double Keywords When Error  Retry Element Click  ${repo_delete_btn}  Wait Until Element Is Visible And Enabled  ${delete_confirm_btn}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}
    Retry Wait Until Page Not Contains Element  ${element_repo_checkbox}
    Filter Repo  ${pro_name}  ${repo_name}  exsit=${false}

Delete Repo on CardView
    [Arguments]  ${reponame}
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/button
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/clr-dropdown-menu/button[contains(.,'Delete')]
    Retry Element Click  ${repo_delete_on_card_view_btn}
    Sleep  2

Delete Project
    [Arguments]  ${projectname}
    Navigate To Projects
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${projectname}')]//div[contains(@class,'clr-checkbox-wrapper')]//label
    Retry Element Click  xpath=//*[@id='delete-project']
    Retry Element Click  //clr-modal//button[contains(.,'DELETE')]
    Sleep  1

Project Should Not Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Retry Wait Until Page Contains Element  //*[@id='contentAll']//div[contains(.,'${projname}')]/../div/clr-icon[@shape='error-standard']

Project Should Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Retry Wait Until Page Contains Element  //*[@id='contentAll']//div[contains(.,'${projname}')]/../div/clr-icon[@shape='success-standard']

Advanced Search Should Display
    Retry Wait Until Page Contains Element  xpath=//audit-log//div[@class='flex-xs-middle']/button

# it's not a common keywords, only used into log case.
Do Log Advanced Search
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'artifact') and contains(.,'pull')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'artifact') and contains(.,'create')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'artifact') and contains(.,'delete')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'project') and contains(.,'create')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'repository') and contains(.,'delete')]
    Retry Element Click  xpath=//audit-log//div[@class='flex-xs-middle']/button
    Retry Element Click  xpath=//project-detail//audit-log//clr-dropdown/button
    #pull log
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Pull')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'pull')]
    #push log
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Push')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'push')]
    #create log
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Create')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'create')]
    #delete log
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Delete')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'delete')]
    #others
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Others')]
    Retry Element Click  xpath=//audit-log//hbr-filter//clr-icon
    Retry Text Input  xpath=//audit-log//hbr-filter//input  harbor-jobservice
    Sleep  1
    ${rc} =  Get Element Count  //audit-log//clr-dg-row
    Should Be Equal As Integers  ${rc}  1

Retry Click Repo Name
    [Arguments]  ${repo_name_element}
    FOR  ${n}  IN RANGE  1  2
        ${out}  Run Keyword And Ignore Error  Retry Double Keywords When Error  Retry Element Click  ${repo_name_element}   Retry Wait Element  ${tag_table_column_vulnerabilities}
        Exit For Loop If  '${out[0]}'=='PASS'
    END
    Should Be Equal As Strings  '${out[0]}'  'PASS'

    FOR  ${n}  IN RANGE  1  2
        ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Not Contains Element  ${repo_list_spinner}
        Exit For Loop If  '${out[0]}'=='PASS'
    END
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Go Into Repo
    [Arguments]  ${repoName}
    Sleep  2
    Retry Wait Until Page Not Contains Element  ${repo_list_spinner}
    ${repo_name_element}=  Set Variable  xpath=//clr-dg-cell[contains(.,'${repoName}')]/a
    FOR  ${n}  IN RANGE  1  3
        Retry Element Click  ${repo_search_icon}
        Retry Clear Element Text  ${repo_search_input}
        Retry Text Input  ${repo_search_input}  ${repoName}
        ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Contains Element  ${repo_name_element}
        Sleep  2
        Run Keyword If  '${out[0]}'=='FAIL'  Reload Page
        Continue For Loop If  '${out[0]}'=='FAIL'
        Retry Click Repo Name  ${repo_name_element}
        Sleep  2
        Exit For Loop
    END
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Click Index Achieve
    [Arguments]  ${tag_name}
    Retry Element Click  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256') and contains(.,'${tag_name}')]//clr-dg-cell[1]//clr-tooltip//a

Go Into Index And Contain Artifacts
    [Arguments]  ${tag_name}  ${total_artifact_count}=3  ${archive_count}=0  ${return_immediately}=${false}
    Retry Double Keywords When Error  Click Index Achieve  ${tag_name}  Page Should Contain Element  ${tag_table_column_os_arch}
    FOR  ${n}  IN RANGE  1  10
        ${out1}  Run Keyword And Ignore Error  Page Should Contain Element  ${artifact_rows}  limit=${total_artifact_count}
        ${out2}  Run Keyword And Ignore Error  Page Should Contain Element  ${archive_rows}  limit=${archive_count}
        Exit For Loop If  '${out1[0]}'=='PASS' and '${out2[0]}'=='PASS'
        Sleep  3
    END
    ${result}=  Set Variable If  '${out1[0]}'=='FAIL' or '${out2[0]}'=='FAIL'  FAIL  PASS
    Return From Keyword If  '${return_immediately}' == '${true}'  ${result}
    Should Be Equal As Strings  '${out1[0]}'  'PASS'
    Should Be Equal As Strings  '${out2[0]}'  'PASS'

Switch To CardView
    Retry Element Click  xpath=//hbr-repository-gridview//span[@class='card-btn']/clr-icon
    Sleep  5

Expand Repo
    [Arguments]  ${projectname}
    Retry Element Click  //repository//clr-dg-row[contains(.,'${projectname}')]//button/clr-icon
    Sleep  1

Edit Repo Info
    Retry Element Click  //*[@id='repo-info']
    Retry Wait Until Page Contains Element  //*[@id='info']/form/div[2]
    # Cancel input
    Retry Element Click  xpath=//*[@id='info-edit-button']/button
    Input Text  xpath=//*[@id='info-edit-textarea']  test_description_info
    Retry Element Click  xpath=//*[@id='edit-cancel']
    Retry Element Click  xpath=//clr-modal//button[contains(.,'CONFIRM')]
    Retry Wait Until Page Contains Element  //*[@id='no-editing']
    # Confirm input
    Retry Element Click  xpath=//*[@id='info-edit-button']/button
    Input Text  xpath=//*[@id='info-edit-textarea']  test_description_info
    Retry Element Click  xpath=//*[@id='edit-save']
    Retry Wait Until Page Contains  test_description_info

Switch To Project Label
    Retry Element Click  xpath=//project-detail//a[contains(.,'Labels')]
    Sleep  1

Switch To Project Repo
    Retry Element Click  xpath=//project-detail//a[contains(.,'Repositories')]
    Sleep  1

Add Labels To Tag
    [Arguments]  ${tagName}  ${labelName}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${tagName}')]//label
    Retry Element Click  xpath=//clr-dg-action-bar//clr-dropdown//span
    Retry Element Click  xpath=//clr-dropdown-menu//clr-dropdown//button[contains(.,'Add Labels')]
    Retry Element Click  xpath=//clr-dropdown//div//label[contains(.,'${labelName}')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row//label[contains(.,'${labelName}')]

Filter Labels In Tags
    [Arguments]  ${labelName1}  ${labelName2}
    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Retry Element Click  xpath=//clr-main-container//artifact-list-tab//clr-select-container//select
    Retry Element Click  xpath=//clr-main-container//artifact-list-tab//clr-select-container//select/option[@value='labels']
    Retry Wait Until Page Contains Element  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Retry Element Click  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Retry Wait Until Page Contains Element  xpath=//clr-datagrid//label[contains(.,'${labelName1}')]

    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Retry Element Click  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName2}')]
    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Sleep  2
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'${labelName2}')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'${labelName1}')]

Get Statics
    [Arguments]  ${locator}
    Reload Page
    Sleep  5
    ${privaterepo}=  Get Text  ${locator}
    [Return]  ${privaterepo}

Retry Get Statics
    [Arguments]  ${locator}
    @{param}  Create List  ${locator}
    ${ret}=  Retry Keyword N Times When Error  5  Get Statics  @{param}
    [Return]  ${ret}

Get Statics Private Repo
    ${privaterepo}=  Retry Get Statics  ${project_statistics_private_repository_icon}
    [Return]  ${privaterepo}

Get Statics Private Project
    ${privateproj}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[1]/div[2]
    [Return]  ${privateproj}

Get Statics Public Repo
    ${publicrepo}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[2]/div[2]
    [Return]  ${publicrepo}

Get Statics Public Project
    ${publicproj}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[2]/div[2]
    [Return]  ${publicproj}

Get Statics Total Repo
    ${totalrepo}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[2]/div/div[3]/div[2]
    [Return]  ${totalrepo}

Get Statics Total Project
    ${totalproj}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[3]/div[2]
    [Return]  ${totalproj}

Input Count Quota
    [Arguments]  ${text}
    ${element_xpath}=  Set Variable  ${project_add_count_quota_input_text_id}
    Retry Clear Element Text  ${element_xpath}
    Retry Text Input  ${element_xpath}  ${text}

Input Storage Quota
    [Arguments]  ${text}  ${unit}=${null}
    ${element_xpath}=  Set Variable  ${project_add_storage_quota_input_text_id}
    Retry Clear Element Text  ${element_xpath}
    Retry Text Input  ${element_xpath}  ${text}
    Run Keyword If  '${unit}'!='${null}'  Select Storage Quota unit  ${unit}

Select Storage Quota unit
    [Arguments]  ${unit}
    Select From List By Value  ${project_add_storage_quota_unit_id}  ${unit}

