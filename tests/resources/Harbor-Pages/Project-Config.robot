*** Settings ***

Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***

Goto Project Config
    Sleep  3
    Retry Element Click  //project-detail//ul/li[contains(.,'Configuration')]
    Sleep  2

Click Project Public
    Mouse Down  //hbr-project-policy-config//input[@name='public']
    Mouse Up  //hbr-project-policy-config//input[@name='public']

Click Content Trust
    Mouse Down  //hbr-project-policy-config//input[@name='content-trust']
    Mouse Up  //hbr-project-policy-config//input[@name='content-trust']

Click Prevent Running
    Mouse Down  //hbr-project-policy-config//input[@name='prevent-vulnerability-image']
    Mouse Up  //hbr-project-policy-config//input[@name='prevent-vulnerability-image']

Select Prevent Level
#value NEGLIGIBLE LOW MEDIUM HIGH
    [Arguments]  ${level}
    Retry Element Click  //hbr-project-policy-config//select
    Retry Element Click  //hbr-project-policy-config//select/option[contains(.,'${level}')]

Click Auto Scan
    Mouse Down  //hbr-project-policy-config//input[@name='scan-image-on-push']
    Mouse Up  //hbr-project-policy-config//input[@name='scan-image-on-push']

Save Project Config
    Sleep  1
    Retry Element Click  //hbr-project-policy-config//button[contains(.,'SAVE')]

Public Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='public']

Project Should Be Public
    [Arguments]  ${projectName}
    Retry Wait Until Page Contains Element  //clr-dg-row[contains(.,'${projectName}')]//clr-dg-cell[contains(.,'Public')]

Content Trust Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='content-trust']

Prevent Running Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='prevent-vulnerability-image']

Auto Scan Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='scan-image-on-push']

Select System CVE Allowlist
    Retry Element Click    ${project_config_system_wl_radio_input}

Select Prject CVE Allowlist
    Retry Element Click    ${project_config_project_wl_radio_input}

Add System CVE Allowlist to Project CVE Allowlist By Add System Button Click
    Goto Project Config
    Select Prject CVE Allowlist
    Retry Element Click    ${project_configuration_wl_project_add_system_btn}
    Retry Element Click    ${project_config_save_btn}

Set Project To Project Level CVE Allowlist
    Goto Project Config
    Select Prject CVE Allowlist
    Retry Element Click    ${project_config_save_btn}

Add Items to Project CVE Allowlist
    [Arguments]    ${cve_id}
    Goto Project Config
    Select Prject CVE Allowlist
    Retry Element Click    ${project_config_project_wl_add_btn}
    Retry Text Input    ${configuration_system_wl_textarea}    ${cve_id}
    Retry Element Click    ${project_config_project_wl_add_confirm_btn}
    Retry Element Click    ${project_config_save_btn}

Delete Top Item In Project CVE Allowlist
    [Arguments]
    Goto Project Config
    Retry Element Click    ${project_configuration_wl_delete_a_cve_id_icon}
    Retry Element Click    ${project_config_save_btn}
