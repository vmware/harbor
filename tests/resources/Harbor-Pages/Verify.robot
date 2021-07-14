*** settings ***
Library  ../../robot-cases/Group3-Upgrade/util.py
Resource  ../../resources/Util.robot

*** Keywords ***
#for jsonpath refer to http://goessner.net/articles/JsonPath/ or https://nottyo.github.io/robotframework-jsonlibrary/JSONLibrary.html

Verify User
    [Arguments]    ${json}
    Log To Console  "Verify User..."
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To User Tag
    @{user}=  Get Value From Json  ${json}  $.users..name
    FOR    ${user}    IN    @{user}
        Page Should Contain    ${user}
    END
    Logout Harbor
    #verify user can login
    @{user}=  Get Value From Json  ${json}  $.users..name
    FOR    ${user}    IN    @{user}
        Sign In Harbor    ${HARBOR_URL}  ${user}  ${HARBOR_PASSWORD}
        Logout Harbor
    END
    Close Browser

Verify Project
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Project..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop
        Retry Wait Until Page Contains    ${project}
    END
    Close Browser


Verify Project Metadata
    # check_content_trust has been removed from Harbor since v2.0
    # verify_registry_name is for proxy cache project, this feature developed since 2.1
    [Arguments]    ${json}  ${check_content_trust}=${true}  ${verify_registry_name}=${false}
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop
        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Project Configuration
        Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.public  ${project_config_public_checkbox}
        Run Keyword If  '${check_content_trust}' == '${true}'  Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.enable_content_trust  ${project_config_content_trust_checkbox}
        Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.auto_scan  ${project_config_scan_images_on_push_checkbox}
        Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.prevent_vul  ${project_config_prevent_vulnerable_images_from_running_checkbox}
        ${ret}    Get Selected List Value    ${project_config_severity_select}
        @{severity}=    Get Value From Json    ${json}    $.projects[?(@.name=${project})].configuration.severity
        Should Contain    ${ret}    @{severity}[0]
        Navigate To Projects
    END
    Close Browser

Verify Image Tag
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Image Tag..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        @{repo}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..repo..name
        Run Keyword If  ${has_image} == ${true}  Loop Image Repo  @{repo}
        Navigate To Projects
    END
    Close Browser

Verify Checkbox
    [Arguments]    ${json}    ${key}    ${checkbox}  ${is_opposite}=${false}
    @{out}=    Get Value From Json    ${json}    ${key}
    ${value}=  Set Variable If  '${is_opposite}'=='${true}'  'false'  'true'
    Run Keyword If    '@{out}[0]'==${value}    Checkbox Should Be Selected    ${checkbox}
    ...    ELSE    Checkbox Should Not Be Selected    ${checkbox}


Loop Image Repo
    [Arguments]    @{repo}
    FOR    ${repo}    IN    @{repo}
        Page Should Contain  ${repo}
    END

Verify Member Exist
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Member Exist..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Member
        @{members}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].member..name
        Loop Member  @{members}
        Navigate To Projects
    END
    Close Browser

Verify Webhook
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Webhook..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Project Webhooks
        ${enabled}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].webhook.enabled
        ${enable_count}  Get Element Count  xpath=//span[contains(.,'Enabled')]
        ${disable_count}  Get Element Count  xpath=//span[contains(.,'Disabled')]
        Log To Console  '${enabled}[0]'
        Log To Console  '${true}'
        Run Keyword If  '${enabled}[0]' == '${true}'  Page Should Contain  Enabled
        ...  ELSE  Page Should Contain  Disabled
        ${address}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].webhook.address
        Log To Console  '${address}[0]'
        Page Should Contain  ${address}[0]
        Page Should Contain  policy
        Page Should Contain  http
        Navigate To Projects
    END
    Close Browser

Verify Webhook For 2.0
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Webhook..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Project Webhooks
        ${enabled}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].webhook.enabled
        ${enable_count}  Get Element Count  xpath=//span[contains(.,'Enabled')]
        ${disable_count}  Get Element Count  xpath=//span[contains(.,'Disabled')]
        Log To Console  '${enabled}[0]'
        Log To Console  '${true}'
        Run Keyword If  '${enabled}[0]' == '${true}'  Page Should Contain  Enabled
        ...  ELSE  Page Should Contain  Disabled
        ${address}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].webhook.address
        ${name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].webhook.name
        ${notify_type}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].webhook.notify_type
        Log To Console  '${address}[0]'
        Log To Console  '${name}[0]'
        Log To Console  '${notify_type}[0]'
        Page Should Contain  ${address}[0]
        Page Should Contain  ${name}[0]
        Page Should Contain  ${notify_type}[0]
        Navigate To Projects
    END
    Close Browser

Verify Tag Retention Rule
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Tag Retention Rule..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        ${tag_retention_rule}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_retention_rule
        Run Keyword If  ${tag_retention_rule}[0] == ${null}  Continue For Loop
        ${out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Tag Retention
        ${actions_count}=  Set Variable  8
        ${repository_patten}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_retention_rule.repository_patten
        ${tag_decoration}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_retention_rule.tag_decoration
        ${latestPushedK}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_retention_rule.latestPushedK_verify
        ${cron}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_retention_rule.cron
        Log To Console  '${repository_patten}[0]'
        Page Should Contain  ${repository_patten}[0]
        Page Should Contain  ${tag_decoration}[0]
        Page Should Contain  ${latestPushedK}[0]
        Page Should Contain  ${cron}[0]
        Navigate To Projects
    END
    Close Browser

Verify Tag Immutability Rule
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Tag Immutability Rule..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Tag Immutability
        @{repo_decoration}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_immutability_rule.repo_decoration
        @{tag_decoration}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_immutability_rule.tag_decoration
        @{repo_pattern}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_immutability_rule.repo_pattern
        @{tag_pattern}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].tag_immutability_rule.tag_pattern
        Log To Console  '@{repo_decoration}[0]'
        #Page Should Contain  @{repo_decoration}[0]
        #Page Should Contain  @{tag_decoration}[0]
        Page Should Contain  @{repo_pattern}[0]
        Page Should Contain  @{tag_pattern}[0]
        Navigate To Projects
    END
    Close Browser

Loop Member
    [Arguments]    @{members}
    FOR    ${member}    IN    @{members}
        Page Should Contain    ${member}
    END

Verify Robot Account Exist
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Robot Account Exist..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Project Robot Account
        @{robot_accounts}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].robot_account..name
        Loop Verify Robot Account  @{robot_accounts}
        Navigate To Projects
    END
    Close Browser

Loop Verify Robot Account
    [Arguments]    @{robot_accounts}
    FOR   ${robot_account}    IN    @{robot_accounts}
        Page Should Contain    ${robot_account}
    END

Verify User System Admin Role
    [Arguments]    ${json}
    Log To Console  "Verify User System Admin Role..."
    @{user}=  Get Value From Json  ${json}  $.admin..name
    Init Chrome Driver
    FOR    ${user}    IN    @{user}
        Sign In Harbor  ${HARBOR_URL}  ${user}  ${HARBOR_PASSWORD}
        Page Should Contain  Administration
        Logout Harbor
    END
    Close Browser

Verify System Label
    [Arguments]    ${json}
    Log To Console  "Verify  System Label..."
    @{label}=   Get Value From Json  ${json}  $..syslabel..name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To System Labels
    FOR    ${label}    IN    @{label}
        Page Should Contain    ${label}
    END
    Close Browser

Verify Project Label
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Project Label..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Project Label
        @{projectlabel}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..labels..name
        FOR    ${label}    IN    @{label}
            Page Should Contain    ${projectlabel}
        END
        Navigate To Projects
    END
   Close Browser

Verify Endpoint
    [Arguments]    ${json}
    Log To Console  "Verify Endpoint..."
    @{endpoint}=  Get Value From Json  ${json}  $.endpoint..name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    FOR    ${endpoint}    IN    @{endpoint}
        Page Should Contain    ${endpoint}
    END
    Close Browser

Verify Replicationrule
    [Arguments]    ${json}
    Log To Console  "Verify Replicationrule..."
    @{replicationrules}=    Get Value From Json    ${json}    $.replicationrule.[*].rulename
    @{endpoints}=    Get Value From Json    ${json}    $.endpoint.[*].name
    FOR    ${replicationrule}    IN    @{replicationrules}
        Init Chrome Driver
        Log To Console    -----replicationrule-----"${replicationrule}"------------
        Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
        Edit Replication Rule    ${replicationrule}
        Capture Page Screenshot
        @{is_src_registry}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].is_src_registry
        @{trigger_type}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].trigger_type
        @{name_filters}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].name_filters
        @{tag_filters}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].tag_filters
        @{dest_namespace}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].dest_namespace
        @{cron}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].cron
        @{is_src_registry}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].is_src_registry
        Log To Console    -----is_src_registry-----@{is_src_registry}[0]------------
        @{endpoint}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].endpoint
        Log To Console    -----endpoint-----@{endpoint}------------
        ${endpoint0}=   Set Variable    @{endpoint}[0]
        Log To Console    -----endpoint0-----${endpoint0}------------
        @{endpoint_type}=    Get Value From Json    ${json}    $.endpoint[?(@.name=${endpoint0})].type
        @{endpoint_url}=    Get Value From Json    ${json}    $.endpoint[?(@.name=${endpoint0})].url
        Retry Textfield Value Should Be    ${filter_name_id}    @{name_filters}[0]
        Retry Textfield Value Should Be    ${filter_tag_id}    @{tag_filters}[0]
        Retry Textfield Value Should Be    ${rule_name_input}    ${replicationrule}
        Retry Textfield Value Should Be    ${dest_namespace_xpath}    @{dest_namespace}[0]
        Log To Console    -----endpoint_type-----@{endpoint_type}[0]------------
        ${registry}=    Set Variable If    "@{endpoint_type}[0]"=="harbor"    ${endpoint0}-@{endpoint_url}[0]    ${endpoint0}-https://hub.docker.com
        Log To Console    -------registry---${registry}------------
        Run Keyword If    '@{is_src_registry}[0]' == '${true}'    Retry List Selection Should Be    ${src_registry_dropdown_list}    ${registry}
        ...    ELSE    Retry List Selection Should Be    ${dest_registry_dropdown_list}    ${registry}
            #Retry List Selection Should Be    ${rule_resource_selector}    ${resource_type}
        Retry List Selection Should Be    ${rule_trigger_select}    @{trigger_type}[0]
        Run Keyword If    '@{trigger_type}[0]' == 'scheduled'    Log To Console    ----------@{trigger_type}[0]------------
        Run Keyword If    '@{trigger_type}[0]' == 'scheduled'    Retry Textfield Value Should Be    ${targetCron_id}    @{cron}[0]
    END

    Reload Page
    FOR    ${replicationrule}    IN    @{replicationrules}
        Delete Replication Rule  ${replicationrule}
    END
    Close Browser

Verify Interrogation Services
    [Arguments]    ${json}
    Log To Console  "Verify Interrogation Services..."
    @{cron}=  Get Value From Json  ${json}  $.interrogation_services..cron
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Page Should Contain  Custom
    Page Should Contain  @{cron}[0]
    Close Browser

Verify System Setting
    [Arguments]    ${json}
    Log To Console  "Verify System Setting..."
    @{authtype}=  Get Value From Json  ${json}  $.configuration.authmode
    @{creation}=  Get Value From Json  ${json}  $.configuration..projectcreation
    @{selfreg}=  Get Value From Json  ${json}  $.configuration..selfreg
    @{emailserver}=  Get Value From Json  ${json}  $.configuration..emailserver
    @{emailport}=  Get Value From Json  ${json}  $.configuration..emailport
    @{emailuser}=  Get Value From Json  ${json}  $.configuration..emailuser
    @{emailfrom}=  Get Value From Json  ${json}  $.configuration..emailfrom
    @{token}=  Get Value From Json  ${json}  $.configuration..token
    @{robot_token}=  Get Value From Json  ${json}  $.configuration..robot_token
    @{scanschedule}=  Get Value From Json  ${json}  $.configuration..scanall
    @{cve_ids}=  Get Value From Json  ${json}  $.configuration..cve
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Page Should Contain  @{authtype}[0]
    Run Keyword If  @{selfreg}[0] == 'True'  Checkbox Should Be Checked  //clr-checkbox-wrapper[@id='selfReg']//label
    Run Keyword If  @{selfreg}[0] == 'False'  Checkbox Should Not Be Checked  //clr-checkbox-wrapper[@id='selfReg']//label
    Switch To Email
    Textfield Value Should Be  xpath=//*[@id='mailServer']  @{emailserver}[0]
    Textfield Value Should Be  xpath=//*[@id='emailPort']  @{emailport}[0]
    Textfield Value Should Be  xpath=//*[@id='emailUsername']  @{emailuser}[0]
    Textfield Value Should Be  xpath=//*[@id='emailFrom']  @{emailfrom}[0]
    Switch To System Settings
    ${ret}  Get Selected List Value  xpath=//select[@id='proCreation']
    Should Be Equal As Strings  ${ret}  @{creation}[0]
    Token Must Be Match  @{token}[0]
    Robot Account Token Must Be Match  @{robot_token}[0]
    Close Browser

Verify Project-level Allowlist
    [Arguments]    ${json}  ${verify_registry_name}=${false}
    Log To Console  "Verify Project-level Allowlist..."
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{registry_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].registry_name
        Run Keyword If  '${registry_name}[0]' != '${null}' and '${verify_registry_name}' == '${false}'   Continue For Loop

        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Go Into Project  ${project}  has_image=${has_image}
        Switch To Project Configuration
        @{is_reuse_sys_cve_allowlist}=    Get Value From Json    ${json}    $.projects[?(@.name=${project})].configuration.reuse_sys_cve_allowlist
        Run Keyword If  "@{is_reuse_sys_cve_allowlist}[0]" == "true"  Retry Wait Element Should Be Disabled   ${project_config_project_wl_add_btn}
        ...  ELSE  Retry Wait Element  ${project_config_project_wl_add_btn}
        @{cve_ids}=    Get Value From Json    ${json}    $.projects[?(@.name=${project})].configuration.cve
        Loop Verifiy CVE_IDs  @{cve_ids}
        Navigate To Projects
    END
    Close Browser

Loop Verifiy CVE_IDs
    [Arguments]    @{cve_ids}
    FOR    ${cve_id}    IN    @{cve_ids}
        Page Should Contain    ${cve_id}
    END

Verify System Setting Allowlist
    [Arguments]    ${json}
    Log To Console  "Verify Verify System Setting Allowlist..."
    @{cve_ids}=  Get Value From Json  ${json}  $.configuration..cve..id
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To System Settings
    Log To Console  "@{cve_ids}"
    Loop Verifiy CVE_IDs  @{cve_ids}
    Close Browser

Verify Trivy Is Default Scanner
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Scanners Page
    Should Display The Default Trivy Scanner
    Close Browser

Verify Artifact Index
    [Arguments]    ${json}
    Log To Console  "Verify Artifact Index..."
    # Only the 1st project has manifest image, so use index 0 of projects for verification.
    @{project}=  Get Value From Json  ${json}  $.projects.[0].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        ${name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].artifact_index.name
        ${tag}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].artifact_index.tag
        Go Into Project  ${project}  has_image=${true}
        Go Into Repo  ${project}/${name}[0]
        Go Into Index And Contain Artifacts  ${tag}[0]  total_artifact_count=2
        Pull image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project}  ${name}[0]:${tag}[0]
        Navigate To Projects
    END
    Close Browser

Loop Repo
    [Arguments]  ${project}  @{repos}
    FOR    ${repo}    IN    @{repos}
        Navigate To Projects
        Go Into Project  ${project}  has_image=${true}
        Go Into Repo  ${project}/${repo}[0][cache_image_namespace]/${repo}[0][cache_image]
        Pull image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project}  ${repo}[0][cache_image_namespace]/${repo}[0][cache_image]:${repo}[0][tag]
    END

Verify Proxy Cache Image Existence
    [Arguments]    ${json}
    Log To Console  "Verify Proxy Cache Image Existence..."
    # Only the 3rd project has cached image, so use index 2 of projects for verification.
    @{project}=  Get Value From Json  ${json}  $.projects.[2].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        @{repo}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].repo
        Loop Repo  ${project}  @{repo}
    END
    Close Browser

Verify Distributions
    [Arguments]    ${json}
    Log To Console  "Verify Distributions..."
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    @{distribution_names}=  Get Value From Json  ${json}  $.distributions..name
    Switch To Distribution
    FOR    ${name}    IN    @{distribution_names}
        ${endpoint}=  Get Value From Json  ${json}  $.distributions[?(@.name=${name})].endpoint
        ${vendor}=  Get Value From Json  ${json}  $.distributions[?(@.name=${name})].vendor
        ${auth_mode}=  Get Value From Json  ${json}  $.distributions[?(@.name=${name})].auth_mode
        Retry Wait Until Page Contains Element  //clr-dg-row[contains(.,'${name}') and contains(.,'${endpoint}[0]') and contains(.,'${vendor}[0]') and contains(.,'${auth_mode}[0]')]
    END

Verify P2P Preheat Policy
    [Arguments]    ${json}
    Log To Console  "P2P Preheat Policy..."
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Navigate To Projects
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    FOR    ${project}    IN    @{project}
        @{p2p_preheat_policys}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].p2p_preheat_policy
        @{policy_names}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].p2p_preheat_policy..name
        @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
        ${has_image}  Set Variable If  ${out_has_image}[0] == ${true}  ${true}  ${false}
        Run Keyword If  ${p2p_preheat_policys}[0] == ${null}  Continue For Loop
        Go Into Project  ${project}  has_image=${has_image}
        Switch To P2P Preheat
        Loop P2P Preheat Policys  ${json}  ${project}  @{policy_names}
    END
    Close Browser

Loop P2P Preheat Policys
    [Arguments]  ${json}  ${project}  @{policy_names}
    FOR    ${policy}    IN    @{policy_names}
        ${provider_name}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].p2p_preheat_policy[?(@.name=${policy})].provider_name
        Retry Wait Until Page Contains Element   //clr-dg-row[contains(.,'${policy}') and contains(.,'${provider_name}[0]')]
    END


Verify Quotas Display
    [Arguments]    ${json}
    Log To Console  "Verify Quotas Display..."
    @{project}=  Get Value From Json  ${json}  $.quotas.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    FOR    ${project}    IN    @{project}
        ${storage_quota_ret}=  Get Project Storage Quota Text From Project Quotas List  ${project}
        ${storage_limit}=  Get Value From Json  ${json}  $.quotas[?(@.name=${project})].storage_limit
        ${size}=  Get Value From Json  ${json}  $.quotas[?(@.name=${project})].size
        ${size_in_mb}=  Evaluate  ${size}[0] * 1024 * 1024
        ${storage_usage}=  Convert Int To Readable File Size  ${size_in_mb}
        ${storage_usage_without_unit}=  Get Substring  ${storage_usage}  0  -2
        ${storage_usage_unit}=  Get Substring  ${storage_usage}  -2
        ${storage_total_size}=  Convert Int To Readable File Size  ${storage_limit}[0]
        Log All  storage_usage_without_unit:${storage_usage_without_unit}
        Log All  storage_usage_unit:${storage_usage_unit}
        Log All  storage_total_size:${storage_total_size}
        Log All  storage_quota_ret:${storage_quota_ret}
        ${str_expected}=  Replace String  ${storage_usage_without_unit}(\\\.\\d{1,2})*${storage_usage_unit} of ${storage_total_size}  B  iB
        Should Match Regexp  ${storage_quota_ret}  ${str_expected}
    END
    Close Browser
