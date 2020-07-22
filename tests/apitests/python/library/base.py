# -*- coding: utf-8 -*-

import sys
import time
import subprocess
import client
import swagger_client
import v2_swagger_client
try:
    from urllib import getproxies
except ImportError:
    from urllib.request import getproxies

class Server:
    def __init__(self, endpoint, verify_ssl):
        self.endpoint = endpoint
        self.verify_ssl = verify_ssl

class Credential:
    def __init__(self, type, username, password):
        self.type = type
        self.username = username
        self.password = password

def _create_client(server, credential, debug, api_type="products"):
    cfg = None
    if api_type in ('projectv2', 'artifact', 'repository', 'scan'):
        cfg = v2_swagger_client.Configuration()
    else:
        cfg = swagger_client.Configuration()

    cfg.host = server.endpoint
    cfg.verify_ssl = server.verify_ssl
    # support basic auth only for now
    cfg.username = credential.username
    cfg.password = credential.password
    cfg.debug = debug

    proxies = getproxies()
    proxy = proxies.get('http', proxies.get('all', None))
    if proxy:
        cfg.proxy = proxy
    return {
        "chart":   client.ChartRepositoryApi(client.ApiClient(cfg)),
        "products":   swagger_client.ProductsApi(swagger_client.ApiClient(cfg)),
        "projectv2":  v2_swagger_client.ProjectApi(v2_swagger_client.ApiClient(cfg)),
        "artifact":   v2_swagger_client.ArtifactApi(v2_swagger_client.ApiClient(cfg)),
        "repository": v2_swagger_client.RepositoryApi(v2_swagger_client.ApiClient(cfg)),
        "scan": v2_swagger_client.ScanApi(v2_swagger_client.ApiClient(cfg)),
        "scanner": swagger_client.ScannersApi(swagger_client.ApiClient(cfg)),
    }.get(api_type,'Error: Wrong API type')

def _assert_status_code(expect_code, return_code):
    if str(return_code) != str(expect_code):
        raise Exception(r"HTTPS status code s not as we expected. Expected {}, while actual HTTPS status code is {}.".format(expect_code, return_code))

def _assert_status_body(expect_status_body, returned_status_body):
    if expect_status_body.strip() != returned_status_body.strip():
        raise Exception(r"HTTPS status body s not as we expected. Expected {}, while actual HTTPS status body is {}.".format(expect_status_body, returned_status_body))

def _random_name(prefix):
    return "%s-%d" % (prefix, int(round(time.time() * 1000)))

def _get_id_from_header(header):
    try:
        location = header["Location"]
        return int(location.split("/")[-1])
    except Exception:
        return None

def _get_string_from_unicode(udata):
    result=''
    for u in udata:
        tmp = u.encode('utf8')
        result = result + tmp.strip('\n\r\t')
    return result

def run_command(command):
    print "Command: ", subprocess.list2cmdline(command)
    try:
        output = subprocess.check_output(command,
                                         stderr=subprocess.STDOUT,
                                         universal_newlines=True)
    except subprocess.CalledProcessError as e:
        raise Exception('Error: Exited with error code: %s. Output:%s'% (e.returncode, e.output))
    return output

class Base:
    def __init__(self,
        server = Server(endpoint="http://localhost:8080/api", verify_ssl=False),
        credential = Credential(type="basic_auth", username="admin", password="Harbor12345"),
        debug = True, api_type = "products"):
        if not isinstance(server.verify_ssl, bool):
            server.verify_ssl = server.verify_ssl == "True"
        self.server = server
        self.credential = credential
        self.debug = debug
        self.api_type = api_type
        self.client = _create_client(server, credential, debug, api_type=api_type)

    def _get_client(self, **kwargs):
        if len(kwargs) == 0:
            return self.client
        server = self.server
        if "api_type" in kwargs:
            server.api_type = kwargs.get("api_type")
        if "endpoint" in kwargs:
            server.endpoint = kwargs.get("endpoint")
        if "verify_ssl" in kwargs:
            if not isinstance(kwargs.get("verify_ssl"), bool):
                server.verify_ssl = kwargs.get("verify_ssl") == "True"
            else:
                server.verify_ssl = kwargs.get("verify_ssl")
        credential = self.credential
        if "type" in kwargs:
            credential.type = kwargs.get("type")
        if "username" in kwargs:
            credential.username = kwargs.get("username")
        if "password" in kwargs:
            credential.password = kwargs.get("password")
        return _create_client(server, credential, self.debug, self.api_type)