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
${project_create_xpath}  //clr-dg-action-bar//button[contains(.,'New')]
${self_reg_xpath}  //input[@id='selfReg']
${test_ldap_xpath}  //*[@id='authentication']/config-auth/div/button[3]
${config_save_button_xpath}  //config//div/button[contains(.,'SAVE')]
${config_email_save_button_xpath}  //*[@id='config_email_save']
${config_auth_save_button_xpath}  //*[@id='config_auth_save']
${config_system_save_button_xpath}  //*[@id='config_system_save']
${vulnerbility_save_button_xpath}  //*[@id='config-save']
${configuration_xpath}  //clr-main-container//clr-vertical-nav//a[contains(.,' Configuration ')]
${garbage_collection_xpath}  //*[@id='config-gc']
${gc_log_xpath}  //*[@id='gc-log']
${gc_config_page}  //clr-vertical-nav-group-children/a[contains(.,'Garbage')]
${gc_now_xpath}  //*[@id='gc']/gc-config//button[contains(.,'GC')]
${gc_log_details_xpath}  //*[@id='clr-dg-row26']/clr-dg-cell[6]/a
${configuration_system_tabsheet_id}  //*[@id='config-system']
${configuration_project_quotas_tabsheet_id}  //*[@id='config-quotas']
${configuration_system_wl_add_btn}    //*[@id='show-add-modal-button']
${configuration_system_wl_textarea}    //*[@id='allowlist-textarea']
${configuration_system_wl_add_confirm_btn}    //*[@id='add-to-system']
${configuration_system_wl_delete_a_cve_id_icon}    //system-settings/form/section//ul/li[1]/a[2]/clr-icon
${configuration_sys_repo_readonly_chb_id}  //*[@id='repo_read_only_lbl']
${checkbox_delete_untagged_artifacts}  //gc-config//clr-toggle-wrapper/label[contains(@for,'delete_untagged')]
