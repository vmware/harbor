# -*- coding: utf-8 -*-

import base
import subprocess
import json

try:
    import docker
except ImportError:
    import pip
    pip.main(['install', 'docker'])
    import docker

def docker_info_display():
    command = ["docker", "info", "-f", "'{{.OSType}}/{{.Architecture}}'"]
    print "Docker Info: ", command
    ret = base.run_command(command)
    print "Command return: ", ret

def docker_login_cmd(harbor_host, user, password, enable_manifest = True):
    command = ["sudo", "docker", "login", harbor_host, "-u", user, "-p", password]
    print "Docker Login Command: ", command
    base.run_command(command)
    if enable_manifest == True:
        try:
            ret = subprocess.check_output(["./tests/apitests/python/update_docker_cfg.sh"], shell=False)
        except subprocess.CalledProcessError, exc:
            raise Exception("Failed to update docker config, error is {} {}.".format(exc.returncode, exc.output))

def docker_manifest_create(index, manifests):
    command = ["sudo", "docker","manifest","create",index]
    command.extend(manifests)
    print "Docker Manifest Command: ", command
    base.run_command(command)

def docker_manifest_push(index):
    command = ["sudo", "docker","manifest","push",index]
    print "Docker Manifest Command: ", command
    ret = base.run_command(command)
    index_sha256=""
    manifest_list=[]
    for line in ret.split("\n"):
        if line[:7] == "sha256:":
            index_sha256 = line
        if line.find('Pushed ref') == 0:
            manifest_list.append(line[-71:])
    return index_sha256, manifest_list

def docker_manifest_push_to_harbor(index, manifests, harbor_server, user, password):
    docker_login_cmd(harbor_server, user, password)
    docker_manifest_create(index, manifests)
    return docker_manifest_push(index)

def list_repositories(harbor_host, user, password, n = None, last = None):
    if n is not None and last is not None:
        command = ["curl", "-s", "-u", user+":"+password, "https://"+harbor_host+"/v2/_catalog"+"?n=%d"%n+"&last="+last, "--insecure"]
    elif n is not None:
            command = ["curl", "-s", "-u", user+":"+password, "https://"+harbor_host+"/v2/_catalog"+"?n=%d"%n, "--insecure"]
    else:
        command = ["curl", "-s", "-u", user+":"+password, "https://"+harbor_host+"/v2/_catalog", "--insecure"]
    print "List Repositories Command: ", command
    ret = base.run_command(command)
    repos = json.loads(ret).get("repositories","")
    return repos

def list_image_tags(harbor_host, repository, user, password, n = None, last = None):
    if n is not None and last is not None:
        command = ["curl", "-s", "-u", user+":"+password, "https://"+harbor_host+"/v2/"+repository+"/tags/list"+"?n=%d"%n+"&last="+last, "--insecure"]
    elif n is not None:
        command = ["curl", "-s", "-u", user+":"+password, "https://"+harbor_host+"/v2/"+repository+"/tags/list"+"?n=%d"%n, "--insecure"]
    else:
        command = ["curl", "-s", "-u", user+":"+password, "https://"+harbor_host+"/v2/"+repository+"/tags/list", "--insecure"]
    print "List Image Tags Command: ", command
    ret = base.run_command(command)
    tags = json.loads(ret).get("tags","")
    return tags

class DockerAPI(object):
    def __init__(self):
        self.DCLIENT = docker.APIClient(base_url='unix://var/run/docker.sock',version='auto',timeout=30)
        self.DCLIENT2 = docker.from_env()

    def docker_login(self, registry, username, password, expected_error_message = None):
        if expected_error_message is "":
            expected_error_message = None
        try:
            self.DCLIENT.login(registry = registry, username=username, password=password)
        except docker.errors.APIError, err:
            if expected_error_message is not None:
                print "docker login error:", str(err)
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Docker login: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker login failed, error is [{}]".format (err.message))

    def docker_image_pull(self, image, tag = None, expected_error_message = None):
        if tag is not None:
            _tag = tag
        else:
            _tag = "latest"
        if expected_error_message is "":
            expected_error_message = None
        caught_err = False
        ret = ""
        try:
            ret = base._get_string_from_unicode(self.DCLIENT.pull(r'{}:{}'.format(image, _tag)))
            return ret
        except Exception, err:
            caught_err = True
            if expected_error_message is not None:
                print "docker image pull error:", str(err)
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Pull image: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker pull image {} failed, error is [{}]".format (image, err.message))
        if caught_err == False:
            if expected_error_message is not None:
                if str(ret).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when pull image {}, return message: {}".format (expected_error_message, image, str(ret)))
            else:
                if str(ret).lower().find("error".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when pull image {}, return message is [{}]".format (image, ret))

    def docker_image_tag(self, image, harbor_registry, tag = None):
        _tag = base._random_name("tag")
        if tag is not None:
            _tag = tag
        try:
            self.DCLIENT.tag(image, harbor_registry, _tag, force=True)
            return harbor_registry, _tag
        except docker.errors.APIError, e:
            raise Exception(r" Docker tag image {} failed, error is [{}]".format (image, e.message))

    def docker_image_push(self, harbor_registry, tag, expected_error_message = None):
        caught_err = False
        ret = ""
        if expected_error_message is "":
            expected_error_message = None
        try:
            ret = base._get_string_from_unicode(self.DCLIENT.push(harbor_registry, tag, stream=True))
            return ret
        except Exception, err:
            caught_err = True
            if expected_error_message is not None:
                print "docker image push error:", str(err)
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Push image: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker push image {} failed, error is [{}]".format (harbor_registry, err.message))
        if caught_err == False:
            if expected_error_message is not None:
                if str(ret).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when push image {}, return message: {}".
                                    format (expected_error_message, harbor_registry, str(ret)))
            else:
                if str(ret).lower().find("errorDetail".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when push image {}, return message is [{}]".
                                    format (harbor_registry, ret))

    def docker_image_build(self, harbor_registry, tags=None, size=1, expected_error_message = None):
        caught_err = False
        ret = ""
        try:
            baseimage='busybox:latest'
            if not self.DCLIENT.images(name=baseimage):
                self.DCLIENT.pull(baseimage)
            c=self.DCLIENT.create_container(image='busybox:latest',command='dd if=/dev/urandom of=test bs=1M count=%d' % size )
            self.DCLIENT.start(c)
            self.DCLIENT.wait(c)
            if not tags:
                tags=['latest']
            firstrepo="%s:%s" % (harbor_registry, tags[0])
            #self.DCLIENT.commit(c, firstrepo)
            self.DCLIENT2.containers.get(c).commit(harbor_registry, tags[0])
            for tag in tags[1:]:
                repo="%s:%s" % (harbor_registry, tag)
                self.DCLIENT.tag(firstrepo, repo)
            for tag in tags:
                repo="%s:%s" % (harbor_registry, tag)
                self.DCLIENT.push(repo)
                print("build image %s with size %d" % (repo, size))
                self.DCLIENT.remove_image(repo)
            self.DCLIENT.remove_container(c)
            self.DCLIENT.pull(repo)
            image = self.DCLIENT2.images.get(repo)
            return repo, image.id
        except Exception, err:
            caught_err = True
            if expected_error_message is not None:
                print "docker image build error:", str(err)
                if str(err).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r"Push image: Return message {} is not as expected {}".format(str(err), expected_error_message))
            else:
                raise Exception(r" Docker build image {} failed, error is [{}]".format (harbor_registry, err.message))
        if caught_err == False:
            if expected_error_message is not None:
                if str(ret).lower().find(expected_error_message.lower()) < 0:
                    raise Exception(r" Failed to catch error [{}] when build image {}, return message: {}".
                                    format (expected_error_message, harbor_registry, str(ret)))
            else:
                if str(ret).lower().find("errorDetail".lower()) >= 0:
                    raise Exception(r" It's was not suppose to catch error when push image {}, return message is [{}]".format (harbor_registry, ret))
