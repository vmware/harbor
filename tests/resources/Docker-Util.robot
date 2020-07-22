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
Documentation  This resource provides helper functions for docker operations
Library  OperatingSystem
Library  Process

*** Keywords ***
Run Docker Info
    [Arguments]  ${docker-params}
    Wait Unitl Command Success  docker ${docker-params} info

Pull image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}=${null}  ${is_robot}=${false}
    Log To Console  \nRunning docker pull ${image}...
    ${image_with_tag}=  Set Variable If  '${tag}'=='${null}'  ${image}  ${image}:${tag}
    Run Keyword If  ${is_robot}==${false}  Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    ...  ELSE  Wait Unitl Command Success  docker login -u robot\\\$${user} -p ${pwd} ${ip}
    ${output}=  Docker Pull  ${ip}/${project}/${image_with_tag}
    Log  ${output}
    Log To Console  ${output}
    Should Contain  ${output}  Digest:
    Should Contain  ${output}  Status:
    Should Not Contain  ${output}  No such image:

Push image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${sha256}=${null}  ${is_robot}=${false}  ${tag_suffix}=${false}
    ${image_with_sha256}=  Set Variable If  '${sha256}'=='${null}'  ${image}  ${image}@sha256:${sha256}
    ${image_with_tag}=  Set Variable If  '${sha256}'=='${null}'  ${image}  ${image}:${sha256}
    Sleep  3
    Log To Console  \nRunning docker push ${image}...
    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image_with_sha256}
    ${d}=    Get Current Date    result_format=%m%s
    ${image_with_tag}=  Set Variable If  '${tag_suffix}'=='${true}'  ${image_with_tag}_${d}  ${image_with_tag}
    Run Keyword If  ${is_robot}==${false}  Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    ...  ELSE  Wait Unitl Command Success  docker login -u robot\\\$${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image_with_sha256} ${ip}/${project}/${image_with_tag}
    Wait Unitl Command Success  docker push ${ip}/${project}/${image_with_tag}
    Wait Unitl Command Success  docker logout ${ip}
    Sleep  1
    [Return]  ${image_with_tag}

Push Image With Tag
#tag1 is tag of image on docker hub,default latest,use a version existing if you do not want to use latest
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}  ${tag1}=latest
    Log To Console  \nRunning docker push ${image}...
    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${tag1}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${tag1} ${ip}/${project}/${image}:${tag}
    Wait Unitl Command Success  docker push ${ip}/${project}/${image}:${tag}
    Wait Unitl Command Success  docker logout ${ip}

Cannot Docker Login Harbor
    [Arguments]  ${ip}  ${user}  ${pwd}
    Command Should be Failed  docker login -u ${user} -p ${pwd} ${ip}

Cannot Pull Image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}=${null}  ${err_msg}=${null}
    ${image_with_tag}=  Set Variable If  '${tag}'=='${null}'  ${image}  ${image}:${tag}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    :FOR  ${idx}  IN RANGE  0  30
    \   ${out}  Run Keyword And Ignore Error  Command Should be Failed  docker pull ${ip}/${project}/${image_with_tag}
    \   Exit For Loop If  '${out[0]}'=='PASS'
    \   Sleep  3
    Log To Console  Cannot Pull Image - Pull Log: ${out[1]}
    Should Be Equal As Strings  '${out[0]}'  'PASS'
    Run Keyword If  '${err_msg}' != '${null}'  Should Contain  ${out[1]}  ${err_msg}

Cannot Pull Unsigned Image
    [Arguments]  ${ip}  ${user}  ${pass}  ${proj}  ${imagewithtag}
    Wait Unitl Command Success  docker login -u ${user} -p ${pass} ${ip}
    ${output}=  Command Should be Failed  docker pull ${ip}/${proj}/${imagewithtag}
    Log To Console  ${output}
    Should Contain  ${output}  The image is not signed in Notary

Cannot Push image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${err_msg}=${null}  ${err_msg_2}=${null}
    Log To Console  \nRunning docker push ${image}...
    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image} ${ip}/${project}/${image}
    ${output}=  Command Should be Failed  docker push ${ip}/${project}/${image}
    Run Keyword If  '${err_msg}' != '${null}'  Should Contain  ${output}  ${err_msg}
    Run Keyword If  '${err_msg_2}' != '${null}'  Should Contain  ${output}  ${err_msg_2}
    Wait Unitl Command Success  docker logout ${ip}

Wait Until Container Stops
    [Arguments]  ${container}
    :FOR  ${idx}  IN RANGE  0  60
    \   ${out}=  Run  docker %{VCH-PARAMS} inspect ${container} | grep Status
    \   ${status}=  Run Keyword And Return Status  Should Contain  ${out}  exited
    \   Return From Keyword If  ${status}
    \   Sleep  1
    Fail  Container did not stop within 60 seconds

Hit Nginx Endpoint
    [Arguments]  ${vch-ip}  ${port}
    Wait Unitl Command Success  wget ${vch-ip}:${port}

Get Container IP
    [Arguments]  ${docker-params}  ${id}  ${network}=default  ${dockercmd}=docker
    ${ip}=  Wait Unitl Command Success  ${dockercmd} ${docker-params} network inspect ${network} | jq '.[0].Containers."${id}".IPv4Address' | cut -d \\" -f 2 | cut -d \\/ -f 1
    [Return]  ${ip}

# The local dind version is embedded in Dockerfile
# docker:1.13-dind
# If you are running this keyword in a container, make sure it is run with --privileged turned on
Start Docker Daemon Locally
    ${pid}=  Run  pidof dockerd
    #${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/docker_config.sh
    #Log  ${output}
    #Should Be Equal As Integers  ${rc}  0
    Return From Keyword If  '${pid}' != '${EMPTY}'
    OperatingSystem.File Should Exist  /usr/local/bin/dockerd-entrypoint.sh
    ${handle}=  Start Process  /usr/local/bin/dockerd-entrypoint.sh dockerd>./daemon-local.log 2>&1  shell=True
    Process Should Be Running  ${handle}
    :FOR  ${IDX}  IN RANGE  5
    \   ${pid}=  Run  pidof dockerd
    \   Exit For Loop If  '${pid}' != '${EMPTY}'
    \   Sleep  2s
    Sleep  2s
    [Return]  ${handle}

Prepare Docker Cert
    [Arguments]  ${ip}
    Wait Unitl Command Success  mkdir -p /etc/docker/certs.d/${ip}
    Wait Unitl Command Success  cp harbor_ca.crt /etc/docker/certs.d/${ip}
    Wait Unitl Command Success  cp harbor_ca.crt /usr/local/share/ca-certificates/
    Wait Unitl Command Success  update-ca-certificates

Kill Local Docker Daemon
    [Arguments]  ${handle}  ${dockerd-pid}
    Terminate Process  ${handle}
    Process Should Be Stopped  ${handle}
    Wait Unitl Command Success  kill -9 ${dockerd-pid}

Docker Login Fail
    [Arguments]  ${ip}  ${user}  ${pwd}
    Log To Console  \nRunning docker login ${ip} ...
    ${output}=  Command Should be Failed  docker login -u ${user} -p ${pwd} ${ip}
    Should Contain  ${output}  unauthorized
    Should Not Contain  ${output}  500 Internal Server Error

Docker Login
    [Arguments]  ${server}  ${username}  ${password}
    Wait Unitl Command Success  docker login -u ${username} -p ${password} ${server}

Docker Pull
    [Arguments]  ${image}
    ${output}=  Retry Keyword N Times When Error  10  Wait Unitl Command Success  docker pull ${image}
    Log  ${output}
    Log To Console  Docker Pull: \n ${output}
    [Return]  ${output}

Docker Tag
    [Arguments]  ${src_image}   ${dst_image}
    Wait Unitl Command Success  docker tag ${src_image} ${dst_image}

Docker Push
    [Arguments]  ${image}
    Wait Unitl Command Success  docker push ${image}

Docker Push Index
    [Arguments]  ${ip}  ${user}  ${pwd}  ${index}  ${image1}  ${image2}
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/docker_push_manifest_list.sh ${ip} ${user} ${pwd} ${index} ${image1} ${image2}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Docker Image Can Not Be Pulled
    [Arguments]  ${image}
    :FOR  ${idx}  IN RANGE  0  30
    \   ${out}=  Run Keyword And Ignore Error  Command Should be Failed  docker pull ${image}
    \   Exit For Loop If  '${out[0]}'=='PASS'
    \   Sleep  3
    Log To Console  Cannot Pull Image From Docker - Pull Log: ${out[1]}
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Docker Image Can Be Pulled
    [Arguments]  ${image}  ${period}=60  ${times}=10
    :For  ${n}  IN RANGE  1  ${times}
    \    Sleep  ${period}
    \    ${out}=  Run Keyword And Ignore Error  Docker Pull  ${image}
    \    Log To Console  Return value is ${out[0]}
    \    Exit For Loop If  '${out[0]}'=='PASS'
    \    Sleep  5
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'