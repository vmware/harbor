from __future__ import absolute_import


import unittest

import library.repository
import library.cnab

from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact
from library.docker_api import DockerAPI

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo= Repository()
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_push_cnab_password = "Aa123456"
        self.cnab_repo_name = "test_cnab"
        self.cnab_tag = "test_cnab_tag"

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_push_bundle_name, self.cnab_repo_name, **TestProjects.USER_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_push_bundle_id, **TestProjects.USER_CLIENT)

        #3. Delete user(UA).
        self.user.delete_user(TestProjects.user_id, **ADMIN_CLIENT)

    def testPushBundleByCnab(self):
        """
        Test case:
            Push Bundle By Cnab
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Pull images for bundle;
            4. Push bundle to harbor as repository(RA);
            5. Get repository from Harbor successfully;
            6. Verfiy bundle name;
            7. Get artifact by sha256;
            8. Verify artifact information.
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        #1. Create a new user(UA);
        TestProjects.user_id, user_name = self.user.create_user(user_password = self.user_push_cnab_password, **ADMIN_CLIENT)
        TestProjects.USER_CLIENT=dict(endpoint = self.url, username = user_name, password = self.user_push_cnab_password)


        #2. Create a new project(PA) by user(UA);
        TestProjects.project_push_bundle_id, TestProjects.project_push_bundle_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CLIENT)

        #3. Pull images for bundle;
        _docker_api = DockerAPI()
        _docker_api.docker_image_pull("alpine", tag = "latest")
        _docker_api.docker_image_pull("haproxy", tag = "latest")

        #4. Push bundle to harbor as repository(RA);
        target = harbor_server + "/" + TestProjects.project_push_bundle_name  + "/" + self.cnab_repo_name  + ":" + self.cnab_tag
        reference_sha256 = library.cnab.push_cnab_bundle(harbor_server, user_name, self.user_push_cnab_password, "alpine:latest", "haproxy:latest", target)

        #5. Get repository from Harbor successfully;
        index_data = self.repo.get_repository(TestProjects.project_push_bundle_name, self.cnab_repo_name, **TestProjects.USER_CLIENT)
        print "index_data:", index_data

        #5.2 Cnab bundle can be pulled by ctr successfully;
        # This step might not successful since ctr does't support cnab fully, it might be uncomment sometime in future.
        # Please keep them in comment!
        #library.containerd.ctr_images_pull(user_name, self.user_push_cnab_password, target)
        #library.containerd.ctr_images_list(oci_ref = target)

        #6. Verfiy bundle name;
        self.assertEqual(index_data.name, TestProjects.project_push_bundle_name + "/" + self.cnab_repo_name)

        #7. Get artifact by sha256;
        artifact = self.artifact.get_reference_info(TestProjects.project_push_bundle_name, self.cnab_repo_name, reference_sha256, **TestProjects.USER_CLIENT)

        #8. Verify artifact information;
        self.assertEqual(artifact[0].type, 'CNAB')
        self.assertEqual(artifact[0].digest, reference_sha256)

if __name__ == '__main__':
    unittest.main()

