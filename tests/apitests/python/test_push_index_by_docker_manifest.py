from __future__ import absolute_import


import unittest

import library.repository
import library.docker_api
import library.containerd
from library.base import _assert_status_code
from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact
from library.repository import push_image_to_project
from library.repository import pull_harbor_image

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo= Repository()
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_push_index_password = "Aa123456"
        self.index_name = "ci_test_index"
        self.index_tag = "test_tag"
        self.image_a = "alpine"
        self.image_b = "busybox"

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA,RB,IA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_push_index_name, self.index_name, **TestProjects.USER_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_push_index_name, self.image_a, **TestProjects.USER_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_push_index_name, self.image_b, **TestProjects.USER_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_push_index_id, **TestProjects.USER_CLIENT)

        #3. Delete user(UA).
        self.user.delete_user(TestProjects.user_id, **ADMIN_CLIENT)

    def testAddIndexByDockerManifest(self):
        """
        Test case:
            Push Index By Docker Manifest
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Create 2 new repositorys(RA,RB) in project(PA) by user(UA);
            4. Push an index(IA) to Harbor by docker manifest CLI successfully;
            5. Get Artifacts successfully;
            6. Get index(IA) by reference successfully;
            7. Verify harbor index is index(IA) pushed by docker manifest CLI;
            8.1 Verify harbor index(IA) can be pulled by docker CLI successfully;
            8.2 Verify harbor index(IA) can be pulled by docker CLI successfully;
            9. Get addition successfully;
            10. Unable to Delete artifact in manifest list;
            11. Delete index successfully.
        Tear down:
            1. Delete repository(RA,RB,IA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        #1. Create a new user(UA);
        TestProjects.user_id, user_name = self.user.create_user(user_password = self.user_push_index_password, **ADMIN_CLIENT)

        TestProjects.USER_CLIENT=dict(endpoint = self.url, username = user_name, password = self.user_push_index_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_push_index_id, TestProjects.project_push_index_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CLIENT)

        #3. Create 2 new repositorys(RA,RB) in project(PA) by user(UA);
        repo_name_a, tag_a = push_image_to_project(TestProjects.project_push_index_name, harbor_server, 'admin', 'Harbor12345', self.image_a, "latest")
        repo_name_b, tag_b = push_image_to_project(TestProjects.project_push_index_name, harbor_server, 'admin', 'Harbor12345', self.image_b, "latest")

        #4. Push an index(IA) to Harbor by docker manifest CLI successfully;
        manifests = [harbor_server+"/"+repo_name_a+":"+tag_a, harbor_server+"/"+repo_name_b+":"+tag_b]
        index = harbor_server+"/"+TestProjects.project_push_index_name+"/"+self.index_name+":"+self.index_tag
        index_sha256_cli_ret, manifests_sha256_cli_ret = library.docker_api.docker_manifest_push_to_harbor(index, manifests, harbor_server, user_name, self.user_push_index_password)

        #5. Get Artifacts successfully;
        artifacts = self.artifact.list_artifacts(TestProjects.project_push_index_name, self.index_name, **TestProjects.USER_CLIENT)
        artifacts_ref_child_list = [artifacts[0].references[1].child_digest, artifacts[0].references[0].child_digest]
        self.assertEqual(artifacts_ref_child_list.count(manifests_sha256_cli_ret[0]), 1)
        self.assertEqual(artifacts_ref_child_list.count(manifests_sha256_cli_ret[1]), 1)

        #6. Get index(IA) by reference successfully;
        index_data = self.artifact.get_reference_info(TestProjects.project_push_index_name, self.index_name, self.index_tag, **TestProjects.USER_CLIENT)
        manifests_sha256_harbor_ret = [index_data[0].references[1].child_digest, index_data[0].references[0].child_digest]

        #7. Verify harbor index is index(IA) pushed by docker manifest CLI;
        self.assertEqual(index_data[0].digest, index_sha256_cli_ret)
        self.assertEqual(manifests_sha256_harbor_ret.count(manifests_sha256_cli_ret[0]), 1)
        self.assertEqual(manifests_sha256_harbor_ret.count(manifests_sha256_cli_ret[1]), 1)

        #8.1 Verify harbor index(IA) can be pulled by docker CLI successfully;
        pull_harbor_image(harbor_server, user_name, self.user_push_index_password, TestProjects.project_push_index_name+"/"+self.index_name, self.index_tag)

        #8.2 Verify harbor index(IA) can be pulled by ctr successfully;
        oci_ref = harbor_server+"/"+TestProjects.project_push_index_name+"/"+self.index_name+":"+self.index_tag
        library.containerd.ctr_images_pull(user_name, self.user_push_index_password, oci_ref)
        library.containerd.ctr_images_list(oci_ref = oci_ref)

        #9. Get addition successfully;
        addition_v = self.artifact.get_addition(TestProjects.project_push_index_name, self.index_name, self.index_tag, "vulnerabilities", **TestProjects.USER_CLIENT)
        self.assertEqual(addition_v[0], '{}')
        #This artifact has no build history

        addition_v = self.artifact.get_addition(TestProjects.project_push_index_name, self.index_name, manifests_sha256_cli_ret[0], "vulnerabilities", **TestProjects.USER_CLIENT)
        self.assertEqual(addition_v[0], '{}')
        addition_b = self.artifact.get_addition(TestProjects.project_push_index_name, self.index_name, manifests_sha256_cli_ret[0], "build_history", **TestProjects.USER_CLIENT)
        self.assertIn("ADD file:", addition_b[0])
        image_data = self.artifact.get_reference_info(TestProjects.project_push_index_name, self.index_name, manifests_sha256_cli_ret[0], **TestProjects.USER_CLIENT)

        addition_v = self.artifact.get_addition(TestProjects.project_push_index_name, self.index_name, manifests_sha256_cli_ret[1], "vulnerabilities", **TestProjects.USER_CLIENT)
        self.assertEqual(addition_v[0], '{}')
        addition_b = self.artifact.get_addition(TestProjects.project_push_index_name, self.index_name, manifests_sha256_cli_ret[1], "build_history", **TestProjects.USER_CLIENT)
        self.assertIn("ADD file:", addition_b[0])
        image_data = self.artifact.get_reference_info(TestProjects.project_push_index_name, self.index_name, manifests_sha256_cli_ret[0], **TestProjects.USER_CLIENT)

        #10. Unable to Delete artifact in manifest list;
        self.artifact.delete_artifact(TestProjects.project_push_index_name, self.index_name, manifests_sha256_cli_ret[0], expect_status_code = 412, **TestProjects.USER_CLIENT)

        #11. Delete index successfully.
        self.artifact.delete_artifact(TestProjects.project_push_index_name, self.index_name, self.index_tag, **TestProjects.USER_CLIENT)

if __name__ == '__main__':
    unittest.main()

