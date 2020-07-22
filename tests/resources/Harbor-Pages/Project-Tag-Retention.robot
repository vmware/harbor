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

Add A Tag Retention Rule
    Retry Element Click  xpath=${project_tag_retention_add_rule_xpath}
    Retry Element Click  xpath=${project_tag_retention_template_xpath}
    Retry Element Click  xpath=${project_tag_retention_option_always_xpath}
    Retry Element Click  xpath=${project_tag_retention_save_add_button_xpath}
    Retry Wait Until Page Contains Element   xpath=${project_tag_retention_rule_name_xpath}

Add A Tag Immutability Rule
    [Arguments]  ${scope}  ${tag}
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_tag_retention_add_rule_xpath}  Retry Wait Until Page Contains Element  xpath=${project_tag_immutability_save_add_button_xpath}
    Retry Clear Element Text  ${project_tag_immutability_scope_input_xpath}
    Retry Text Input  ${project_tag_immutability_scope_input_xpath}  ${scope}
    Retry Clear Element Text  ${project_tag_immutability_tag_input_xpath}
    Retry Text Input  ${project_tag_immutability_tag_input_xpath}  ${tag}
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_tag_immutability_save_add_button_xpath}  Retry Wait Until Page Contains Element  xpath=${project_tag_retention_rule_name_xpath}
    Retry Wait Until Page Contains  ${scope}
    Retry Wait Until Page Contains  ${tag}

Delete A Tag Retention Rule
    Retry Element Click  xpath=${project_tag_retention_action_button_xpath}
    Retry Element Click  xpath=${project_tag_retention_delete_button_xpath}
    Retry Wait Until Page Not Contains Element   xpath=${project_tag_retention_rule_name_xpath}

Delete A Tag Immutability Rule
    Retry Element Click  xpath=${project_tag_retention_action_button_xpath}
    Retry Element Click  xpath=${project_tag_retention_delete_button_xpath}
    Retry Wait Until Page Not Contains Element   xpath=${project_tag_retention_rule_name_xpath}

Edit A Tag Retention Rule
    [Arguments]  ${repos}   ${tags}
    Retry Element Click  xpath=${project_tag_retention_action_button_xpath}
    Retry Element Click  xpath=${project_tag_retention_edit_button_xpath}
    Retry Wait Until Page Contains Element   xpath=${project_tag_retention_modal_title_edit_xpath}
    Input Text  ${project_tag_retention_repo_input_xpath}  ${repos}
    Input Text  ${project_tag_retention_tags_input_xpath}  ${tags}
    Retry Element Click  xpath=${project_tag_retention_save_add_button_xpath}
    Retry Wait Until Page Contains Element   xpath=//span[contains(@class, 'rule-name')]//span[contains(.,'${tags}')]

Edit A Tag Immutability Rule
    [Arguments]  ${scope}  ${tag}
    Retry Element Click  xpath=${project_tag_retention_action_button_xpath}
    Retry Element Click  xpath=${project_tag_retention_edit_button_xpath}
    Retry Clear Element Text  ${project_tag_immutability_scope_input_xpath}
    Retry Text Input  ${project_tag_immutability_scope_input_xpath}  ${scope}
    Retry Clear Element Text  ${project_tag_immutability_tag_input_xpath}
    Retry Text Input  ${project_tag_immutability_tag_input_xpath}  ${tag}
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_tag_immutability_save_add_button_xpath}  Retry Wait Until Page Contains Element  xpath=${project_tag_retention_rule_name_xpath}
    Retry Wait Until Page Contains  ${scope}
    Retry Wait Until Page Contains  ${tag}

Set Daily Schedule
    Retry Element Click  xpath=${project_tag_retention_edit_schedule_xpath}
    Retry Element Click  xpath=${project_tag_retention_select_policy_xpath}
    Retry Element Click  xpath=${project_tag_retention_option_daily_xpath}
    Retry Element Click  xpath=${project_tag_retention_config_save_xpath}
    Retry Wait Until Page Contains Element  xpath=${project_tag_retention_schedule_ok_xpath}
    Retry Element Click   xpath=${project_tag_retention_schedule_ok_xpath}
    Retry Wait Until Page Contains Element  xpath=${project_tag_retention_span_daily_xpath}

Execute Result Should Be
    [Arguments]  ${result}
    :FOR  ${idx}  IN RANGE  0  20
    \   ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Contains Element  xpath=//clr-dg-cell[contains(., '${result}')]
    \   Exit For Loop If  '${out[0]}'=='PASS'
    \   Sleep  6
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Execute Dry Run
    Retry Element Click  xpath=${project_tag_retention_dry_run_xpath}
    Retry Wait Until Page Contains Element  xpath=${project_tag_retention_record_yes_xpath}
    Sleep    5
    Retry Element Click  xpath=${project_tag_retention_record_yes_xpath}
    # memcached:123 should be deleted and hello-world:latest should be retained
    Execute Result Should Be  0/1


Execute Run
    Retry Element Click  xpath=${project_tag_retention_run_now_xpath}
    Retry Element Click  xpath=${project_tag_retention_execute_run_xpath}
    Retry Wait Until Page Contains Element  xpath=${project_tag_retention_record_no_xpath}
    Sleep    5
    Retry Element Click  xpath=${project_tag_retention_record_no_xpath}
    # memcached:123 should be deleted and hello-world:latest should be retained
    Execute Result Should Be  0/1

