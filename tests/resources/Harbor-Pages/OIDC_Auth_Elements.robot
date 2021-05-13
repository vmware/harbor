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
${log_oidc_provider_btn}       //*[@id='log_oidc']
${login_with_email_btn}    //span[contains(., 'Log in with Email')]
${login_with_ldap_btn}     //span[contains(., 'Log in with LDAP')]
${dex_login_btn}    //*[@id='login']
${dex_pwd_btn}    //*[@id='password']
${submit_login_btn}    //*[@id='submit-login']
${grant_btn}      xpath=/html/body/div[2]/div/div[2]/div[1]/form/button
${oidc_username_input}    //*[@id='oidcUsername']
${save_btn}       //*[@id='saveButton']
${OIDC_USERNAME}  test1
${generate_secret_btn}       //*[@id='generate-cli-btn']
${more_btn}       //*[@id='hidden-generate-cli']