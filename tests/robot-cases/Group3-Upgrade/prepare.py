import os
import sys
import json
import argparse
import requests
from functools import wraps
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

parser = argparse.ArgumentParser(description='The script to generate data for harbor v1.4.0')
parser.add_argument('--endpoint', '-e', dest='endpoint', required=True, help='The endpoint to harbor')
parser.add_argument('--version', '-v', dest='version', required=False, help='The version to harbor')
args = parser.parse_args()

url = "https://"+args.endpoint+"/api/"
endpoint_url = "https://"+args.endpoint
print url

with open("feature_map.json") as f:
    feature_map = json.load(f)

def get_branch(func_name, version):
    has_feature = False
    for node in feature_map[func_name]:
        has_feature = True
        if node["version"] == version:
            return node["branch"]
    if has_feature is False:
        return "No Restriction"
    else:
        return "Not Supported"

def get_feature_branch(func):
    @wraps(func)
    def inner_func(*args,**kwargs):
        branch=get_branch(inner_func.__name__, kwargs.get("version"))
        if branch == "No Restriction":
            func(*args,**kwargs)
        elif branch == "Not Supported":
            print("Feature {} is not supported in version {}".format(inner_func.__name__, kwargs.get("version")))
        else:
            kwargs["branch"] = branch
            func(*args,**kwargs)
        return
    return inner_func

class HarborAPI:
    @get_feature_branch
    def create_project(self, project, **kwargs):
        if kwargs["branch"] == 1:
            body=dict(body={"project_name": ""+project["name"]+"", "metadata": {"public": "true"}})
        elif kwargs["branch"] == 2:
            body=dict(body={"project_name": ""+project["name"]+"", "metadata": {"public": "true"},"count_limit":project["count_limit"],"storage_limit":project["storage_limit"]})
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))
        request(url+"projects", 'post', **body)

    def create_user(self, username):
        payload = {"username":username, "email":username+"@vmware.com", "password":"Harbor12345", "realname":username, "comment":"string"}
        body=dict(body=payload)
        request(url+"users", 'post', **body)

    @get_feature_branch
    def set_user_admin(self, user, **kwargs):
        r = request(url+"users?username="+user+"", 'get')
        userid = str(r.json()[0]['user_id'])

        if kwargs["branch"] == 1:
            body=dict(body={"has_admin_role": 1})
        elif kwargs["branch"] == 2:
            body=dict(body={"has_admin_role": True})
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))
        request(url+"users/"+userid+"/sysadmin", 'put', **body)

    @get_feature_branch
    def add_member(self, project, user, role, **kwargs):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])

        if kwargs["branch"] == 1:
            payload = {"roles": [role], "username":""+user+""}
        elif kwargs["branch"] == 2:
            payload = {"member_user":{ "username": ""+user+""},"role_id": role}
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))
        body=dict(body=payload)
        request(url+"projects/"+projectid+"/members", 'post', **body)

    @get_feature_branch
    def add_endpoint(self, endpointurl, endpointname, username, password, insecure, registry_type, **kwargs):
        if kwargs["branch"] == 1:
            payload = {"endpoint": ""+endpointurl+"", "name": ""+endpointname+"", "username": ""+username+"", "password": ""+password+"", "insecure": insecure}
            body=dict(body=payload)
            request(url+"targets", 'post', **body)
        elif kwargs["branch"] == 2:
            if registry_type == "harbor":
                endpointurl = endpoint_url
            payload = {
                "credential":{
                    "access_key":""+username+"",
                    "access_secret":""+password+"",
                    "type":"basic"
                },
                "insecure":insecure,
                "name":""+endpointname+"",
                "type":""+registry_type+"",
                "url":""+endpointurl+""
            }
            body=dict(body=payload)
            print  body
            request(url+"/registries", 'post', **body)
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))

    @get_feature_branch
    def add_replication_rule(self, replicationrule, **kwargs):
        if kwargs["branch"] == 1:
            r = request(url+"projects?name="+replicationrule["project"]+"", 'get')
            projectid = r.json()[0]['project_id']
            r = request(url+"targets?name="+replicationrule["endpoint"]+"", 'get')
            targetid = r.json()[0]['id']
            payload = {"name": ""+replicationrule["rulename"]+"", "description": "string", "projects": [{"project_id": projectid,}], "targets": [{"id": targetid,}], "trigger": {"kind": ""+replicationrule["trigger"]+"", "schedule_param": {"type": "weekly", "weekday": 1, "offtime": 0}}}
            body=dict(body=payload)
            request(url+"policies/replication", 'post', **body)
        elif kwargs["branch"] == 2:
            r = request(url+"registries?name="+replicationrule["endpoint"]+"", 'get')
            targetid = r.json()[0]['id']
            if replicationrule["is_src_registry"] is True:
                registry = r'"src_registry": { "id": '+str(targetid)+r'},'
            else:
                registry = r'"dest_registry": { "id": '+str(targetid)+r'},'

            body=dict(body=json.loads(r'{"name":"'+replicationrule["rulename"].encode('utf-8')+r'","dest_namespace":"'+replicationrule["dest_namespace"].encode('utf-8')+r'","deletion": '+str(replicationrule["deletion"]).lower()+r',"enabled": '+str(replicationrule["enabled"]).lower()+r',"override": '+str(replicationrule["override"]).lower()+r',"description": "string",'+ registry + r'"trigger":{"type": "'+replicationrule["trigger_type"]+r'", "trigger_settings":{"cron":"'+replicationrule["cron"]+r'"}},"filters":[ {"type":"name","value":"'+replicationrule["name_filters"]+r'"},{"type":"tag","value":"'+replicationrule["tag_filters"]+r'"}]}'))
            print body
            request(url+"replication/policies", 'post', **body)
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))

    #@get_feature_branch
    def update_project_setting_metadata(self, project, public, contenttrust, preventrunning, preventseverity, scanonpush):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])
        payload = {
            "metadata": {
                "public": public,
                "enable_content_trust": contenttrust,
                "prevent_vul": preventrunning,
                "severity": preventseverity,
                "auto_scan": scanonpush
            }
        }
        body=dict(body=payload)
        print body
        request(url+"projects/"+projectid+"", 'put', **body)

    @get_feature_branch
    def add_sys_allowlist(self, cve_id_list, **kwargs):
        cve_id_str = ""
        if kwargs["branch"] == 1:
            for index, cve_id in enumerate(cve_id_list["cve"]):
                cve_id_str = cve_id_str + '{"cve_id":"' +cve_id["id"] + '"}'
                if index != len(cve_id_list["cve"]) - 1:
                    cve_id_str = cve_id_str + ","
            body=dict(body=json.loads(r'{"items":['+cve_id_str.encode('utf-8')+r'],"expires_at":'+cve_id_list["expires_at"]+'}'))
            request(url+"system/CVEWhitelist", 'put', **body)
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))

    @get_feature_branch
    def update_project_setting_allowlist(self, project, reuse_sys_cve_whitelist, cve_id_list, **kwargs):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])
        cve_id_str = ""
        if kwargs["branch"] == 1:
            for index, cve_id in enumerate(cve_id_list["cve"]):
                cve_id_str = cve_id_str + '{"cve_id":"' +cve_id["id"] + '"}'
                if index != len(cve_id_list["cve"]) - 1:
                    cve_id_str = cve_id_str + ","
            print cve_id_str
            if reuse_sys_cve_whitelist == "true":
                payload = r'{"metadata":{"reuse_sys_cve_whitelist":"true"}}'
            else:
                payload = r'{"metadata":{"reuse_sys_cve_whitelist":"false"},"cve_whitelist":{"project_id":'+projectid+',"items":['+cve_id_str.encode('utf-8')+r'],"expires_at":'+cve_id_list["expires_at"]+'}}'
            print payload
            body=dict(body=json.loads(payload))
            request(url+"projects/"+projectid+"", 'put', **body)
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))

    @get_feature_branch
    def update_interrogation_services(self, cron, **kwargs):
        payload = {"schedule":{"type":"Custom","cron": cron}}
        print payload
        body=dict(body=payload)
        request(url+"system/scanAll/schedule", 'post', **body)

    def update_systemsetting(self, emailfrom, emailhost, emailport, emailuser, creation, selfreg, token, robot_token):
        payload = {
            "auth_mode": "db_auth",
            "email_from": emailfrom,
            "email_host": emailhost,
            "email_port": emailport,
            "email_identity": "string",
            "email_username": emailuser,
            "email_ssl": True,
            "email_insecure": True,
            "project_creation_restriction": creation,
            "read_only": False,
            "self_registration": selfreg,
            "token_expiration": token,
            "robot_token_duration":robot_token,
            "scan_all_policy": {
                "type": "none",
                "parameter": {
                    "daily_time": 0
                }
            }
        }
        print payload
        body=dict(body=payload)
        request(url+"configurations", 'put', **body)

    @get_feature_branch
    def add_project_robot_account(self, project, robot_account, **kwargs):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])

        if kwargs["branch"] == 1:
            if len(robot_account["access"]) == 1:
                robot_account_ac = robot_account["access"][0]
                payload = {
                    "name": robot_account["name"],
                    "access": [
                        {
                            "resource": "/project/"+projectid+"/repository",
                            "action": robot_account_ac["action"]
                        }
                    ]
                }
            elif len(robot_account["access"]) == 2:
                payload = {
                    "name": robot_account["name"],
                    "access": [
                        {
                            "resource": "/project/"+projectid+"/repository",
                            "action": "pull"
                        },
                        {
                            "resource": "/project/"+projectid+"/repository",
                            "action": "push"
                        }
                    ]
                }
            else:
                raise Exception(r"Error: Robot account count {} is not legal!".format(len(robot_account["access"])))
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, branch))
        print payload
        body=dict(body=payload)
        request(url+"projects/"+projectid+"/robots", 'post', **body)

    @get_feature_branch
    def add_tag_retention_rule(self, project, tag_retention_rule, **kwargs):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])
        if kwargs["branch"] == 1:
            payload = {
                "algorithm":"or",
                "rules":
                [
                    {
                        "disabled":False,
                        "action":"retain",
                        "scope_selectors":
                        {
                            "repository":
                            [
                                {
                                    "kind":"doublestar",
                                    "decoration":"repoMatches",
                                    "pattern":tag_retention_rule["repository_patten"]
                                }
                            ]
                        },
                        "tag_selectors":
                        [
                            {
                                "kind":"doublestar",
                                "decoration":"matches","pattern":tag_retention_rule["tag_decoration"]
                            }
                        ],
                        "params":{"latestPushedK":tag_retention_rule["latestPushedK"]},
                        "template":"latestPushedK"
                    }
                ],
                "trigger":
                {
                    "kind":"Schedule",
                    "references":{},
                    "settings":{"cron":tag_retention_rule["cron"]}
                },
                "scope":
                {
                "level":"project",
                    "ref":int(projectid)
                }
            }
            print payload
            body=dict(body=payload)
            request(url+"retentions", 'post', **body)
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, kwargs["branch"]))

    @get_feature_branch
    def add_tag_immutability_rule(self, project, tag_immutability_rule, **kwargs):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])
        if kwargs["branch"] == 1:
            payload = {
                "disabled":False,
                "action":"immutable",
                "scope_selectors":
                {
                    "repository":
                    [
                        {
                            "kind":"doublestar",
                            "decoration":tag_immutability_rule["repo_decoration"],
                            "pattern":tag_immutability_rule["repo_pattern"]
                        }
                    ]
                },
                "tag_selectors":
                [
                    {
                        "kind":"doublestar",
                        "decoration":tag_immutability_rule["tag_decoration"],
                        "pattern":tag_immutability_rule["tag_pattern"]
                    }
                ],
                "project_id":int(projectid),
                "priority":0,
                "template":"immutable_template"
            }
            print payload
            body=dict(body=payload)
            request(url+"projects/"+projectid+"/immutabletagrules", 'post', **body)
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, kwargs["branch"]))

    @get_feature_branch
    def add_webhook(self, project, webhook, **kwargs):
        r = request(url+"projects?name="+project+"", 'get')
        projectid = str(r.json()[0]['project_id'])
        if kwargs["branch"] == 1:
            payload = {
                "targets":[
                    {
                        "type":"http",
                        "address":webhook["address"],
                        "skip_cert_verify":webhook["skip_cert_verify"],
                        "auth_header":webhook["auth_header"]
                    }
                ],
                "event_types":[
                    "downloadChart",
                    "deleteChart",
                    "uploadChart",
                    "deleteImage",
                    "pullImage",
                    "pushImage",
                    "scanningFailed",
                    "scanningCompleted"
                ],
                "enabled":webhook["enabled"]
            }
            print payload
            body=dict(body=payload)
            request(url+"projects/"+projectid+"/webhook/policies", 'post', **body)
        else:
            raise Exception(r"Error: Feature {} has no branch {}.".format(sys._getframe().f_code.co_name, kwargs["branch"]))

    def update_repoinfo(self, reponame):
        payload = {"description": "testdescription"}
        print payload
        body=dict(body=payload)
        request(url+"repositories/"+reponame+"", 'put', **body)

    def get_ca(self, target='/harbor/ca/ca.crt'):
        url = "https://" + args.endpoint + "/api/systeminfo/getcert"
        resp = request(url, 'get')
        try:
            ca_content = json.loads(resp.text)
        except ValueError:
            ca_content = resp.text
        ca_path = '/harbor/ca'
        if not os.path.exists(ca_path):
            try:
                os.makedirs(ca_path)
            except Exception, e:
                print str(e)
                pass
        open(target, 'wb').write(ca_content)


def request(url, method, user = None, userp = None, **kwargs):
    if user is None:
        user = "admin"
    if userp is None:
        userp = "Harbor12345"
    kwargs.setdefault('headers', kwargs.get('headers', {}))
    kwargs['headers']['Accept'] = 'application/json'
    if 'body' in kwargs:
        kwargs['headers']['Content-Type'] = 'application/json'
        kwargs['data'] = json.dumps(kwargs['body'])
        del kwargs['body']
    print "url: ", url
    resp = requests.request(method, url, verify=False, auth=(user, userp), **kwargs)
    if resp.status_code >= 400:
        raise Exception("[Exception Message] - {}".format(resp.text))
    return resp

with open("data.json") as f:
    data = json.load(f)

def pull_image(*image):
    for i in image:
        os.system("docker pull "+i)

def push_image(image, project):
    os.system("docker tag "+image+" "+args.endpoint+"/"+project+"/"+image)
    os.system("docker login "+args.endpoint+" -u admin"+" -p Harbor12345")
    os.system("docker push "+args.endpoint+"/"+project+"/"+image)

def push_signed_image(image, project, tag):
    os.system("./sign_image.sh" + " " + args.endpoint + " " + project + " " + image + " " + tag)

def do_data_creation():
    harborAPI = HarborAPI()
    harborAPI.get_ca()

    for user in data["users"]:
        harborAPI.create_user(user["name"])

    for user in data["admin"]:
        harborAPI.set_user_admin(user["name"], version=args.version)

    for project in data["projects"]:
        harborAPI.create_project(project, version=args.version)
        for member in project["member"]:
            harborAPI.add_member(project["name"], member["name"], member["role"], version=args.version)
        for robot_account in project["robot_account"]:
            harborAPI.add_project_robot_account(project["name"], robot_account, version=args.version)
        harborAPI.add_webhook(project["name"], project["webhook"], version=args.version)
        harborAPI.add_tag_retention_rule(project["name"], project["tag_retention_rule"], version=args.version)
        harborAPI.add_tag_immutability_rule(project["name"], project["tag_immutability_rule"], version=args.version)

    pull_image("busybox", "redis", "haproxy", "alpine", "httpd:2")
    push_image("busybox", data["projects"][0]["name"])
    push_signed_image("alpine", data["projects"][0]["name"], "latest")

    for endpoint in data["endpoint"]:
        harborAPI.add_endpoint(endpoint["url"], endpoint["name"], endpoint["user"], endpoint["pass"], endpoint["insecure"], endpoint["type"], version=args.version)

    for replicationrule in data["replicationrule"]:
        harborAPI.add_replication_rule(replicationrule, version=args.version)

    for project in data["projects"]:
        harborAPI.update_project_setting_metadata(project["name"],
                                        project["configuration"]["public"],
                                        project["configuration"]["enable_content_trust"],
                                        project["configuration"]["prevent_vul"],
                                        project["configuration"]["severity"],
                                        project["configuration"]["auto_scan"])

    for project in data["projects"]:
        harborAPI.update_project_setting_allowlist(project["name"],
                                    project["configuration"]["reuse_sys_cve_allowlist"],
                                    project["configuration"]["deployment_security"], version=args.version)

    harborAPI.update_interrogation_services(data["interrogation_services"]["cron"], version=args.version)

    harborAPI.update_systemsetting(data["configuration"]["emailsetting"]["emailfrom"],
                                   data["configuration"]["emailsetting"]["emailserver"],
                                   float(data["configuration"]["emailsetting"]["emailport"]),
                                   data["configuration"]["emailsetting"]["emailuser"],
                                   data["configuration"]["projectcreation"],
                                   data["configuration"]["selfreg"],
                                   float(data["configuration"]["token"]),
                                   float(data["configuration"]["robot_token"])*60*24)

    harborAPI.add_sys_allowlist(data["configuration"]["deployment_security"], version=args.version)

do_data_creation()
