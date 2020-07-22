# -*- coding: utf-8 -*-

import time
import base
import swagger_client
from docker_api import DockerAPI
from swagger_client.rest import ApiException

def pull_harbor_image(registry, username, password, image, tag, expected_login_error_message = None, expected_error_message = None):
    _docker_api = DockerAPI()
    _docker_api.docker_login(registry, username, password, expected_error_message = expected_login_error_message)
    if expected_login_error_message != None:
        return
    time.sleep(2)
    ret = _docker_api.docker_image_pull(r'{}/{}'.format(registry, image), tag = tag, expected_error_message = expected_error_message)
    print ret

def push_image_to_project(project_name, registry, username, password, image, tag, expected_login_error_message = None, expected_error_message = None, profix_for_image = None):
    _docker_api = DockerAPI()
    _docker_api.docker_login(registry, username, password, expected_error_message = expected_login_error_message)
    time.sleep(2)
    if expected_login_error_message != None:
        return
    _docker_api.docker_image_pull(image, tag = tag)
    time.sleep(2)

    if profix_for_image == None:
        new_harbor_registry, new_tag = _docker_api.docker_image_tag(r'{}:{}'.format(image, tag), r'{}/{}/{}'.format(registry, project_name, image))
    else:
        new_harbor_registry, new_tag = _docker_api.docker_image_tag(r'{}:{}'.format(image, tag), r'{}/{}/{}/{}'.format(registry, project_name, profix_for_image, image))
    time.sleep(2)

    _docker_api.docker_image_push(new_harbor_registry, new_tag, expected_error_message = expected_error_message)

    return r'{}/{}'.format(project_name, image), new_tag

def push_special_image_to_project(project_name, registry, username, password, image, tags=None, size=1, expected_login_error_message=None, expected_error_message = None):
    _docker_api = DockerAPI()
    _docker_api.docker_login(registry, username, password, expected_error_message = expected_login_error_message)
    time.sleep(2)
    if expected_login_error_message != None:
        return
    return _docker_api.docker_image_build(r'{}/{}/{}'.format(registry, project_name, image), tags = tags, size=size, expected_error_message=expected_error_message)

def is_repo_exist_in_project(repositories, repo_name):
    result = False
    for reop in repositories:
        if reop.name == repo_name:
            return True
    return result

class Repository(base.Base, object):
    def __init__(self):
        super(Repository,self).__init__(api_type = "repository")

    def list_tags(self, repository, **kwargs):
        client = self._get_client(**kwargs)
        return client.repositories_repo_name_tags_get(repository)

    def get_tag(self, repo_name, tag, **kwargs):
        client = self._get_client(**kwargs)
        return client.repositories_repo_name_tags_tag_get(repo_name, tag)

    def image_exists(self, repository, tag, **kwargs):
        tags = self.list_tags(repository, **kwargs)
        exist = False
        for t in tags:
            if t.name == tag:
                exist = True
                break
        return exist

    def image_should_exist(self, repository, tag, **kwargs):
        if not self.image_exists(repository, tag, **kwargs):
            raise Exception("image %s:%s not exist" % (repository, tag))

    def image_should_not_exist(self, repository, tag, **kwargs):
        if self.image_exists(repository, tag, **kwargs):
            raise Exception("image %s:%s exists" % (repository, tag))

    def delete_repoitory(self, project_name, repo_name, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)
        try:
            _, status_code, _ = client.delete_repository_with_http_info(project_name, repo_name)
        except Exception as e:
            base._assert_status_code(expect_status_code, e.status)
            return e.body
        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)


    def list_repositories(self, project_name, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.list_repositories_with_http_info(project_name)
        base._assert_status_code(200, status_code)
        return data

    def get_repository(self, project_name, repo_name, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.get_repository_with_http_info(project_name, repo_name)
        base._assert_status_code(200, status_code)
        return data

    def add_label_to_tag(self, repo_name, tag, label_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        label = swagger_client.Label(id=label_id)
        _, status_code, _ = client.repositories_repo_name_tags_tag_labels_post_with_http_info(repo_name, tag, label)
        base._assert_status_code(expect_status_code, status_code)

    def get_repo_signatures(self, repo_name, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        data, status_code, _ = client.repositories_repo_name_signatures_get_with_http_info(repo_name)
        base._assert_status_code(expect_status_code, status_code)
        return data

    def check_image_not_scanned(self, repo_name, tag, **kwargs):
        tag = self.get_tag(repo_name, tag, **kwargs)
        if tag.scan_overview != None:
            raise Exception("Image should be <Not Scanned> state!")

    def repository_should_exist(self, project_id, repo_name, **kwargs):
        repositories = self.list_repositories(project_id, **kwargs)
        if is_repo_exist_in_project(repositories, repo_name) == False:
            raise Exception("Repository {} is not exist.".format(repo_name))

    def signature_should_exist(self, repo_name, tag, **kwargs):
        signatures = self.get_repo_signatures(repo_name, **kwargs)
        for each_sign in signatures:
            if each_sign.tag == tag and len(each_sign.hashes["sha256"]) == 44:
                print "sha256:", len(each_sign.hashes["sha256"])
                return
        raise Exception(r"Signature of {}:{} is not exist!".format(repo_name, tag))

    def retag_image(self, repo_name, tag, src_image, override=True, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)
        request = swagger_client.RetagReq(tag=tag, src_image=src_image, override=override)

        try:
            data, status_code, _ = client.repositories_repo_name_tags_post_with_http_info(repo_name, request)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return

        base._assert_status_code(expect_status_code, status_code)
        base._assert_status_code(200, status_code)
        return data