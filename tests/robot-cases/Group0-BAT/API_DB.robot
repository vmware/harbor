*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/APITest-Util.robot
Resource  ../../resources/Docker-Util.robot
Library  ../../apitests/python/library/Harbor.py  ${SERVER_CONFIG}
Library  OperatingSystem
Library  String
Library  Collections
Library  requests
Library  Process
Default Tags  APIDB

*** Variables ***
${SERVER}  ${ip}
${SERVER_URL}  https://${SERVER}
${SERVER_API_ENDPOINT}  ${SERVER_URL}/api
&{SERVER_CONFIG}  endpoint=${SERVER_API_ENDPOINT}  verify_ssl=False

*** Test Cases ***

Test Case - Garbage Collection
    [Tags]  gc
    Harbor API Test  ./tests/apitests/python/test_garbage_collection.py

Test Case - Add Private Project Member and Check User Can See It
    [Tags]  private_member
    Harbor API Test  ./tests/apitests/python/test_add_member_to_private_project.py

Test Case - Delete a Repository of a Certain Project Created by Normal User
    [Tags]  del_repo
    Harbor API Test  ./tests/apitests/python/test_del_repo.py

Test Case - Add a System Global Label to a Certain Tag
    [Tags]  global_lbl
    Harbor API Test  ./tests/apitests/python/test_add_sys_label_to_tag.py

Test Case - Add Replication Rule
    [Tags]  replic_rule
    Harbor API Test  ./tests/apitests/python/test_add_replication_rule.py

Test Case - Edit Project Creation
    [Tags]  pro_creation
    Harbor API Test  ./tests/apitests/python/test_edit_project_creation.py

Test Case - Manage Project Member
    [Tags]  member
    Harbor API Test  ./tests/apitests/python/test_manage_project_member.py

Test Case - Project Level Policy Content Trust
    [Tags]  content_trust
    Harbor API Test  ./tests/apitests/python/test_project_level_policy_content_trust.py

Test Case - User View Logs
    [Tags]  view_logs
    Harbor API Test  ./tests/apitests/python/test_user_view_logs.py

Test Case - List Helm Charts
    [Tags]  list_helm_charts
    Harbor API Test  ./tests/apitests/python/test_list_helm_charts.py

Test Case - Assign Sys Admin
    [Tags]  assign_adin
    Harbor API Test  ./tests/apitests/python/test_assign_sys_admin.py

Test Case - Copy Artifact Outside Project
    [Tags]  copy_artifact
    Harbor API Test  ./tests/apitests/python/test_copy_artifact_outside_project.py

Test Case - Robot Account
    [Tags]  robot_account
    Harbor API Test  ./tests/apitests/python/test_robot_account.py

Test Case - Sign A Image
    [Tags]  sign_image
    Harbor API Test  ./tests/apitests/python/test_sign_image.py

Test Case - Project Quota
    [Tags]  quota
    Harbor API Test  ./tests/apitests/python/test_project_quota.py

Test Case - System Level CVE Allowlist
    [Tags]  sys_cve
    Harbor API Test  ./tests/apitests/python/test_sys_cve_allowlists.py

Test Case - Project Level CVE Allowlist
    [Tags]  pro_cve
    Harbor API Test  ./tests/apitests/python/test_project_level_cve_allowlist.py

Test Case - Tag Retention
    [Tags]  tag_retention
    Harbor API Test  ./tests/apitests/python/test_retention.py

Test Case - Health Check
    [Tags]  health
    Harbor API Test  ./tests/apitests/python/test_health_check.py

Test Case - Push Index By Docker Manifest
    [Tags]  push_index
    Harbor API Test  ./tests/apitests/python/test_push_index_by_docker_manifest.py

Test Case - Push Chart By Helm3 Chart CLI
    [Tags]  push_chart
    Harbor API Test  ./tests/apitests/python/test_push_chart_by_helm3_chart_cli.py

Test Case - Push Cnab Bundle
    [Tags]  push_cnab
    Harbor API Test  ./tests/apitests/python/test_push_cnab_bundle.py

Test Case - Create/Delete tag
    [Tags]  tag_cuid
    Harbor API Test  ./tests/apitests/python/test_create_delete_tag.py

Test Case - Scan Image
    [Tags]  scan
    Harbor API Test  ./tests/apitests/python/test_scan_image_artifact.py

Test Case - Scan All Images
    [Tags]  scan_all
    Harbor API Test  ./tests/apitests/python/test_system_level_scan_all.py

Test Case - Registry API
    [Tags]  reg_api
    Harbor API Test  ./tests/apitests/python/test_registry_api.py

Test Case - Push Image With Special Name
    [Tags]  special_repo_name
    Harbor API Test  ./tests/apitests/python/test_push_image_with_special_name.py

Test Case - Push Artifact With ORAS CLI
    [Tags]  oras
    Harbor API Test  ./tests/apitests/python/test_push_files_by_oras.py

Test Case - Push Singularity file With Singularity CLI
    [Tags]  singularity
    Harbor API Test  ./tests/apitests/python/test_push_sif_by_singularity.py

Test Case - Push Chart File To Chart Repository By Helm V2 With Robot Account
    [Tags]  helm2
    Harbor API Test  ./tests/apitests/python/test_push_chart_by_helm2_helm3_with_robot_Account.py

Test Case - Replication From Dockerhub
    [Tags]  replic_dockerhub
    Harbor API Test  ./tests/apitests/python/test_replication_from_dockerhub.py
