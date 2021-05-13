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
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - List Helm Charts And Delete Chart Files
    Body Of List Helm Charts

Test Case - Helm CLI Push
    Init Chrome Driver
    ${user}=    Set Variable    user004
    ${pwd}=    Set Variable    Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Helm CLI Push Without Sign In Harbor  ${user}  ${pwd}

Test Case - Helm3 CLI Push
    Init Chrome Driver
    ${user}=    Set Variable    user004
    ${pwd}=    Set Variable    Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Helm3 CLI Push Without Sign In Harbor  ${user}  ${pwd}