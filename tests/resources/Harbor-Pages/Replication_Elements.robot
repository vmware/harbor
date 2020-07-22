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

*** Variables ***
${new_name_xpath}  	//hbr-list-replication-rule//button[contains(.,'New')]
${policy_name_xpath}  //*[@id='policy_name']
${policy_description_xpath}  //*[@id='policy_description']
${policy_enable_checkbox}  //input[@id='policy_enable']/../label
${policy_endpoint_checkbox}  //input[@id='check_new']/../label
${destination_name_xpath}  //*[@id='destination_name']
${destination_url_xpath}  //*[@id='destination_url']
${destination_username_xpath}    //*[@id='destination_access_key']
${destination_password_xpath}  //*[@id='destination_password']
${replication_save_xpath}  //button[contains(.,'OK')]
${replication_xpath}  //clr-vertical-nav-group-children/a[contains(.,'Replication')]
${destination_insecure_xpath}  //label[@id='destination_insecure_checkbox']

${new_replication-rule_button}  //button[contains(.,'New Replication Rule')]
${link_to_registries}  //clr-modal//span[contains(.,'Endpoint')]
${new_endpoint_button}  //hbr-endpoint//button[contains(.,'New Endpoint')]
${rule_name}  //input[@id='ruleName']
${source_image_filter_add}  //hbr-create-edit-rule/clr-modal//clr-icon[@id='add-label-list']
${source_iamge_repo_filter}  //hbr-create-edit-rule//section/div[4]/div/div[1]/div/label/input
${source_image_tag_filter}  //hbr-create-edit-rule//section/div[4]/div/div[2]/div/label/input
${rule_target_select}  //select[@id='ruleTarget']
${rule_trigger_select}  //select[@id='ruleTrigger']
${schedule_type_select}  //select[@name='scheduleType']
${schedule_day_select}  //select[@name='scheduleDay']
${shcedule_time}  //input[@type='time']
${destination_insecure_checkbox}    //hbr-create-edit-endpoint/clr-modal//input[@id='destination_insecure']
${ping_test_button}  //button[contains(.,'Test')]
${nav_to_registries}  //clr-vertical-nav//span[contains(.,'Registries')]
${nav_to_replications}  //clr-vertical-nav//span[contains(.,'Replications')]
${rule_filter_search}  //hbr-replication/div/div[1]//hbr-filter/span/clr-icon
${rule_filter_input}  //hbr-replication/div/div[1]//hbr-filter/span//input
${job_filter_search}  //hbr-replication/div/div[3]//hbr-filter/span/clr-icon
${job_filter_input}  //hbr-replication/div/div[3]//hbr-filter/span//input
${endpoint_filter_search}  //hbr-filter/span/clr-icon
${endpoint_filter_input}  //hbr-filter/span//input
${action_bar_edit}  //button[contains(.,'Edit')]
${action_bar_delete}  //button[contains(.,'Delete')]
${stop_jobs_button}  //button[contains(.,'Stop Jobs')]
${dialog_close}  //clr-modal//button[contains(.,'CLOSE')]
${dialog_delete}  //clr-modal//button[contains(.,'DELETE')]
${dialog_replicate}  //clr-modal//button[contains(.,'REPLICATE')]
${action_bar_replicate}  //button[contains(.,'Replicate')]
${rule_save_button}  //button[contains(.,'SAVE')]
${provider_selector}    //*[@id='adapter']
${replication_mode_radio_push}    //clr-main-container//hbr-create-edit-rule//label[contains(.,'Push-based')]
${replication_mode_radio_pull}    //clr-main-container//hbr-create-edit-rule//label[contains(.,'Pull-based')]
${source_project}    //input[@id='filter_name']
${rule_resource_selector}    //*[@id='select_resource']
${trigger_mode_selector}    //*[@id='ruleTrigger']
${dest_namespace_xpath}    //*[@id='dest_namespace']
${new_replication_rule_id}    //*[@id='new_replication_rule_id']
${edit_replication_rule_id}    //*[@id='edit_replication_rule_id']
${delete_replication_rule_id}    //*[@id='delete_replication_rule_id']
${replication_exec_id}    //*[@id='replication_exe_id']
${replication_task_line_1}    //clr-datagrid//clr-dg-row/div/div[2]//clr-checkbox-wrapper/label[1]
${filter_tag}    //*[@id='filter_tag']
${is_overide_xpath}    //label[contains(.,'Replace the destination resources if name exists')]
${enable_rule_xpath}    //label[contains(.,'Enable rule')]
${targetCron_id}    //*[@id='targetCron']
${rule_name_input}    //*[@id='ruleName']
${src_registry_dropdown_list}    //select[@id='src_registry_id']
${dest_registry_dropdown_list}    //select[@id='dest_registry']
${rule_confirm_btn}    //*[@id='ruleBtnOk']
${rule_cancel_btn}    //*[@id='ruleBtnCancel']
${filter_rules_btn}    //*[@id='filter-rules']
${filter_rules_input}  //*[@id='filter-rules']//input
${del_remote_checkbox}  //label[@for='ruleDeletion']






