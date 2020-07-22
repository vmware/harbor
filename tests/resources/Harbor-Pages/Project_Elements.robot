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
${create_project_button_xpath}  //clr-main-container//button[contains(., 'New Project')]
${project_name_xpath}  //*[@id='create_project_name']
${project_public_xpath}  //input[@name='public']/..//label
${project_save_css}  html body.no-scrolling harbor-app harbor-shell clr-main-container.main-container div.content-container div.content-area.content-area-override project div.row div.col-lg-12.col-md-12.col-sm-12.col-xs-12 div.row.flex-items-xs-between div.option-left create-project clr-modal div.modal div.modal-dialog div.modal-content div.modal-footer button.btn.btn-primary
${log_xpath}  //clr-main-container//clr-vertical-nav//a[contains(.,'Logs')]
${projects_xpath}  //clr-main-container//clr-vertical-nav//a[contains(.,'Projects')]
${project_replication_xpath}  //project-detail//a[contains(.,'Replication')]
${project_log_xpath}  //project-detail//a[contains(.,'Logs')]
${project_member_xpath}  //project-detail//a[contains(.,'Members')]
${project_config_tabsheet}  xpath=//project-detail//a[contains(.,'Configuration')]
${project_tag_strategy_xpath}  //clr-tabs//a[contains(.,'Policy')]
${project_tab_overflow_btn}  //clr-tabs//li//button[contains(@class,"dropdown-toggle")]

${project_tag_immutability_switch}  //project-detail/app-tag-feature-integration//label/a[contains(.,'Tag Immutability')]

${create_project_CANCEL_button_xpath}  xpath=//button[contains(.,'CANCEL')]
${create_project_OK_button_xpath}  xpath=//button[contains(.,'OK')]
${delete_confirm_btn}  xpath=//confirmation-dialog//button[contains(.,'DELETE')]
${project_statistics_private_repository_icon}  xpath=//project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[2]/div[2]/statistics/div/span[1]
${repo_delete_confirm_btn}  xpath=//clr-modal//button[2]
${repo_retag_confirm_dlg}  css=${modal-dialog}
${repo_delete_on_card_view_btn}  //clr-modal//button[contains(.,'DELETE')]
${delete_btn}  //button[contains(.,'Delete')]
${repo_delete_btn}  xpath=//hbr-repository-gridview//button[contains(.,'Delete')]
${project_delete_btn}  xpath=//button[@id='delete-project']
${tag_delete_btn}  xpath=//tag-repository//clr-datagrid//button[contains(.,'Delete')]
${user_delete_btn}  xpath=/clr-dropdown-menu//button[contains(.,'Delete')]
${repo_search_icon}  xpath=//hbr-filter//clr-icon
${repo_search_input}  xpath=//hbr-filter//input
${repo_list_spinner}  xpath=//clr-datagrid//clr-spinner
#${repo_search_icon}  xpath=//hbr-repository-gridview//clr-datagrid//clr-dg-column[contains(.,'Name')]//clr-dg-string-filter//button//clr-icon
#${repo_search_input}  xpath=//div[@class[contains(.,'datagrid-filter')]]//input
${repo_tag_1st_checkbox}  xpath=//clr-datagrid//clr-dg-row//clr-checkbox-wrapper
${tag_table_column_pull_command}  xpath=//clr-dg-column//span[contains(.,'Pull Command')]
${tag_table_column_vulnerabilities}  xpath=//clr-dg-column//span[contains(.,'Vulnerabilities')]
${tag_table_column_os_arch}  xpath=//clr-dg-column//span[contains(.,'OS/ARCH')]
${tag_table_column_tag}  xpath=//clr-dg-column//span[contains(.,'Tag')]
${tag_table_column_size}  xpath=//clr-dg-column//span[contains(.,'Size')]
${tag_table_column_vulnerability}  xpath=//clr-dg-column//span[contains(.,'Vulnerability')]
${tag_images_btn}  xpath=//hbr-repository//button[contains(.,'Images')]
${project_member_action_xpath}  xpath=//*[@id='member-action']
${project_member_set_role_xpath}  xpath=//clr-dropdown-menu//label[contains(.,'Set Role')]
${project_config_public_checkbox}  xpath=//input[@name='public']
${project_config_content_trust_checkbox}  xpath=//input[@name='content-trust']
${project_config_scan_images_on_push_checkbox}  xpath=//input[@name='scan-image-on-push']
${project_config_prevent_vulnerable_images_from_running_checkbox}  xpath=//input[@name='prevent-vulenrability-image-input']
${project_config_severity_select}  xpath=//select[@id='severity']
${project_config_public_checkbox_label}  xpath=//*[@id="clr-wrapper-public"]/div/clr-checkbox-wrapper/label
${project_config_prevent_vulenrability_checkbox_label}    xpath=//*[@id='prevent-vulenrability-image']//clr-checkbox-wrapper//label
${project_config_system_wl_radio_input}    xpath=//clr-radio-wrapper//label[contains(.,'System allowlist')]
${project_config_project_wl_radio_input}    xpath=//clr-radio-wrapper//label[contains(.,'Project allowlist')]
${project_config_system_wl_radio_input_id}    systemAllowlistOrProjectAllowlist
${project_config_project_wl_radio_input_id}    systemAllowlistOrProjectAllowlist
${project_config_project_wl_add_btn}    xpath=//*[@id='show-add-modal']
${project_config_project_wl_add_confirm_btn}    xpath=//*[@id='add-to-allowlist']
${project_config_save_btn}    xpath=//hbr-project-policy-config//button[contains(.,'SAVE')]
${project_add_count_quota_input_text_id}    xpath=//*[@id='create_project_count_limit']
${project_add_storage_quota_input_text_id}    xpath=//*[@id='create_project_storage_limit']
${project_add_storage_quota_unit_id}    xpath=//*[@id='create_project_storage_limit_unit']
