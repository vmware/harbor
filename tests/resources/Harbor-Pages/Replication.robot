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
Filter Replicatin Rule
    [Arguments]  ${ruleName}
    ${rule_name_element}=  Set Variable  xpath=//clr-dg-cell[contains(.,'${ruleName}')]
    Retry Element Click  ${filter_rules_btn}
    Retry Text Input  ${filter_rules_input}  ${ruleName}
    Retry Wait Until Page Contains Element   ${rule_name_element}
    Capture Page Screenshot  filter_replic_${ruleName}.png

Select Dest Registry
    [Arguments]    ${endpoint}
    Retry Element Click    ${dest_registry_dropdown_list}
    Retry Element Click    ${dest_registry_dropdown_list}//option[contains(.,'${endpoint}')]

Select Source Registry
    [Arguments]    ${endpoint}
    Retry Element Click    ${src_registry_dropdown_list}
    Retry Element Click    ${src_registry_dropdown_list}//option[contains(.,'${endpoint}')]

Select Trigger
    [Arguments]    ${mode}
    Retry Element Click    ${rule_trigger_select}
    Retry Element Click    ${rule_trigger_select}//option[contains(.,'${mode}')]

Select Destination URL
    [Arguments]    ${type}
    Retry Element Click  ${destination_url_xpath}
    Retry Element Click  //div[contains(@class, 'selectBox')]//li[contains(.,'${type}')]

Check New Rule UI Without Endpoint
    Retry Element Click    ${new_replication-rule_button}
    Page Should Contain    Please add an endpoint first
    Retry Element Click    ${link_to_registries}
    Retry Wait Until Page Contains    Endpoint URL
    Retry Wait Element  ${new_endpoint_button}

Create A New Endpoint
    [Arguments]    ${provider}    ${name}    ${url}    ${username}    ${pwd}    ${save}=Y
    #click new button
    Retry Element Click  xpath=${new_endpoint_button}
    #input necessary info
    Select From List By Value  ${provider_selector}  ${provider}
    Retry Text Input  xpath=${destination_name_xpath}    ${name}
    Run Keyword If  '${provider}' == 'harbor'  Run keyword  Retry Text Input  xpath=${destination_url_xpath}  ${url}
    Run Keyword If  '${provider}' == 'aws-ecr' or '${provider}' == 'google-gcr'   Run keyword  Select Destination URL  ${url}
    Run Keyword If  '${provider}' != 'google-gcr'   Retry Text Input  xpath=${destination_username_xpath}  ${username}
    Retry Text Input  xpath=${destination_password_xpath}  ${pwd}
    #cancel verify cert since we use a selfsigned cert
    Retry Element Click  ${destination_insecure_xpath}
    Run Keyword If  '${save}' == 'Y'  Run keyword  Retry Double Keywords When Error  Retry Element Click  ${replication_save_xpath}  Retry Wait Until Page Not Contains Element  ${replication_save_xpath}
    Run Keyword If  '${save}' == 'Y'  Run keyword  Retry Wait Until Page Contains  ${name}
    Run Keyword If  '${save}' == 'N'  No Operation

Create A Rule With Existing Endpoint
    [Arguments]    ${name}    ${replication_mode}    ${project_name}    ${resource_type}    ${endpoint}    ${dest_namespace}
    ...    ${mode}=Manual  ${cron}="* */59 * * * *"  ${del_remote}=${false}
    #click new
    Retry Element Click    ${new_name_xpath}
    #input name
    Retry Text Input    ${rule_name}    ${name}
    Run Keyword If    '${replication_mode}' == 'push'  Run Keywords  Retry Element Click  ${replication_mode_radio_push}  AND  Select Dest Registry  ${endpoint}
    ...    ELSE  Run Keywords  Retry Element Click  ${replication_mode_radio_pull}  AND  Select Source Registry  ${endpoint}

    #set filter
    Retry Text Input    ${source_project}    ${project_name}
    Run Keyword And Ignore Error    Select From List By Value    ${rule_resource_selector}    ${resource_type}
    Retry Text Input    ${dest_namespace_xpath}    ${dest_namespace}
    #set trigger
    Select Trigger  ${mode}
    Run Keyword If  '${mode}' == 'Scheduled'  Retry Text Input  ${targetCron_id}  ${cron}
    Run Keyword If  '${mode}' == 'Event Based' and '${del_remote}' == '${true}'  Retry Element Click  ${del_remote_checkbox}
    #click save
    Retry Double Keywords When Error  Retry Element Click  ${rule_save_button}  Retry Wait Until Page Not Contains Element  ${rule_save_button}
    Sleep  2

Endpoint Is Unpingable
    Retry Element Click  ${ping_test_button}
    Wait Until Page Contains  Failed

Endpoint Is Pingable
    Retry Element Click  ${ping_test_button}
    Wait Until Page Contains  successfully

Disable Certificate Verification
    Checkbox Should Be Selected  ${destination_insecure_checkbox}
    Retry Element Click  ${destination_insecure_xpath}
    Sleep  1

Enable Certificate Verification
    Checkbox Should Not Be Selected  ${destination_insecure_checkbox}
    Retry Element Click  ${destination_insecure_xpath}
    Sleep  1

Switch To Registries
    Retry Element Click  ${nav_to_registries}
    Sleep  1

Switch To Replication Manage
    Retry Element Click  ${nav_to_replications}
    Sleep  1

Trigger Replication Manual
    [Arguments]  ${rule}
    Retry Element Click  ${rule_filter_search}
    Retry Text Input   ${rule_filter_input}  ${rule}
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label
    Retry Element Click  ${action_bar_replicate}
    Retry Wait Until Page Contains Element  ${dialog_replicate}
    #change from click to mouse down and up
    Mouse Down  ${dialog_replicate}
    Mouse Up  ${dialog_replicate}
    Sleep  2
    Retry Wait Until Page Contains Element  //*[@id='contentAll']//div[contains(.,'${rule}')]/../div/clr-icon[@shape='success-standard']
    Sleep  1

Rename Rule
    [Arguments]  ${rule}  ${newname}
    Retry Element Click  ${rule_filter_search}
    Retry Text Input  ${rule_filter_input}  ${rule}
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label
    Retry Element Click  ${action_bar_edit}
    Retry Text Input  ${rule_name}  ${newname}
    Retry Element Click  ${rule_save_button}

Delete Rule
    [Arguments]  ${rule}
    Retry Element Click  ${rule_filter_search}
    Retry Text Input   ${rule_filter_input}  ${rule}
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label
    Retry Element Click  ${action_bar_delete}
    Retry Wait Until Page Contains Element  ${dialog_delete}
    #change from click to mouse down and up
    Mouse Down  ${dialog_delete}
    Mouse Up  ${dialog_delete}
    Sleep  2

Filter Rule
    [Arguments]  ${rule}
    Retry Element Click  ${rule_filter_search}
    Retry Text Input   ${rule_filter_input}  ${rule}
    Sleep  1

Select Rule
    [Arguments]  ${rule}
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label

Stop Jobs
    Retry Element Click  ${stop_jobs_button}

View Job Log
    [arguments]  ${job}
    Retry Element Click  ${job_filter_search}
    Retry Text Input  ${job_filter_input}  ${job}
    Retry Link Click  //clr-dg-row[contains(.,'${job}')]//a

Find Item And Click Edit Button
    [Arguments]    ${name}
    Filter Object    ${name}
    Retry Select Object    ${name}
    Retry Element Click    ${action_bar_edit}

Find Item And Click Delete Button
    [Arguments]    ${name}
    Filter Object    ${name}
    Retry Select Object    ${name}
    Retry Element Click    ${action_bar_delete}

Switch To Replication Manage Page
    [Arguments]    ${name}
    Switch To Registries
    Switch To Replication Manage

Edit Replication Rule By Name
    [Arguments]    ${name}
    Retry Double Keywords When Error  Switch To Replication Manage Page  "NULL"  Find Item And Click Edit Button  ${name}

Delete Replication Rule By Name
    [Arguments]    ${name}
    Switch To Registries
    Switch To Replication Manage
    Find Item And Click Delete Button  ${name}

Ensure Delete Replication Rule By Name
    [Arguments]    ${name}
    Delete Replication Rule By Name    ${name}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}
    Retry Wait Element  xpath=//clr-dg-placeholder[contains(.,\"We couldn\'t find any replication rules!\")]

Rename Endpoint
    [arguments]  ${name}  ${newname}
    Find Item And Click Edit Button  ${name}
    Retry Wait Until Page Contains Element  ${destination_name_xpath}
    Retry Text Input  ${destination_name_xpath}  ${newname}
    Retry Element Click  ${replication_save_xpath}

Delete Endpoint
    [Arguments]  ${name}
    Retry Element Click  ${endpoint_filter_search}
    Retry Text Input   ${endpoint_filter_input}  ${name}
    #click checkbox before target endpoint
    Retry Double Keywords When Error  Retry Element Click  //clr-dg-row[contains(.,'${name}')]//clr-checkbox-wrapper  Retry Wait Element  ${action_bar_delete}
    Retry Element Click  ${action_bar_delete}
    Wait Until Page Contains Element  ${dialog_delete}
    Retry Element Click  ${dialog_delete}

Select Rule And Replicate
    [Arguments]  ${rule_name}
    Select Rule  ${rule_name}
    Retry Element Click    ${replication_exec_id}
    Retry Double Keywords When Error    Retry Element Click    xpath=${dialog_replicate}    Retry Wait Until Page Not Contains Element    xpath=${dialog_replicate}

Select Rule And Click Edit Button
    [Arguments]  ${rule_name}
    Retry Element Click  //clr-dg-row[contains(.,'${rule_name}')]//clr-radio-wrapper/label
    Retry Element Click  ${edit_replication_rule_id}

Delete Replication Rule
    [Arguments]  ${name}
    Retry Element Click  ${endpoint_filter_search}
    Retry Text Input   ${endpoint_filter_input}  ${name}
    #click checkbox before target endpoint
    Retry Element Click  //clr-dg-row[contains(.,'${name}')]//label
    Retry Element Click  ${action_bar_delete}
    Wait Until Page Contains Element  ${dialog_delete}
    Retry Element Click  ${dialog_delete}

Image Should Be Replicated To Project
    [Arguments]  ${project}  ${image}  ${period}=60  ${times}=10
    :For  ${n}  IN RANGE  1  ${times}
    \    Sleep  ${period}
    \    Go Into Project    ${project}
    \    Switch To Project Repo
    \    #In AWS-ECR, under repository a, there're only several images: httpd,alpine,hello-world.
    \    ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Contains  ${project}/${image}
    \    Log To Console  Return value is ${out[0]}
    \    Exit For Loop If  '${out[0]}'=='PASS'
    \    Sleep  5
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Executions Result Count Should Be
    [Arguments]  ${expected_status}  ${expected_trigger_type}  ${expected_result_count}
    Sleep  10
    ${count}=  Get Element Count  xpath=//clr-dg-row[contains(.,'${expected_status}') and contains(.,'${expected_trigger_type}')]
    Capture Page Screenshot
    Should Be Equal As Integers  ${count}  ${expected_result_count}
    Capture Page Screenshot

