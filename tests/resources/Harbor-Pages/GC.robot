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
GC Now
    [Arguments]  ${harbor_url}  ${login_user}  ${login_pwd}  ${untag}=${false}
    Switch To Garbage Collection
    Capture Page Screenshot
    Run Keyword If  '${untag}' == '${true}'  Retry Element Click  xpath=${checkbox_delete_untagged_artifacts}
    Capture Page Screenshot
    Click GC Now
    Logout Harbor
    Sleep  2
    Sign In Harbor  ${harbor_url}  ${login_user}  ${login_pwd}
    Switch To Garbage Collection
    Sleep  1
    #Switch To GC History
    #Retry Keyword N Times When Error  60  Retry Wait Until Page Contains  Finished

Retry GC Should Be Successful
    [Arguments]  ${history_id}  ${expected_msg}
    Retry Keyword N Times When Error  15  GC Should Be Successful  ${history_id}  ${expected_msg}

GC Should Be Successful
    [Arguments]  ${history_id}  ${expected_msg}
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -i --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/system/gc/${history_id}/log"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Run Keyword If  '${expected_msg}' != '${null}'  Should Contain  ${output}  ${expected_msg}
    Should Contain  ${output}  success to run gc in job.