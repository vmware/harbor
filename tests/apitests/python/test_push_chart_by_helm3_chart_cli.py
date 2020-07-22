from __future__ import absolute_import


import unittest

import library.repository
import library.helm
from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.repo= Repository()
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_push_chart_password = "Aa123456"
        self.chart_file = "https://storage.googleapis.com/harbor-builds/helm-chart-test-files/harbor-0.2.0.tgz"
        self.archive = "harbor/"
        self.verion = "0.2.0"
        self.repo_name = "harbor_api_test"

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository chart(CA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_push_chart_name, self.repo_name, **TestProjects.USER_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_push_chart_id, **TestProjects.USER_CLIENT)

        #3. Delete user(UA).
        self.user.delete_user(TestProjects.user_id, **ADMIN_CLIENT)

    def testPushChartByHelmChartCLI(self):
        """
        Test case:
            Push Chart File By Helm Chart CLI
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push an chart(CA) to Harbor by helm3 registry/chart CLI successfully;
            4. List artifacts successfully;
            5. Get chart(CA) by reference successfully;
            6. Get addtion successfully;
            7. Delete chart by reference successfully.
        Tear down:
            1. Delete repository chart(CA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        #1. Create a new user(UA);
        TestProjects.user_id, user_name = self.user.create_user(user_password = self.user_push_chart_password, **ADMIN_CLIENT)
        TestProjects.USER_CLIENT=dict(endpoint = self.url, username = user_name, password = self.user_push_chart_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_push_chart_id, TestProjects.project_push_chart_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CLIENT)

        #3. Push an chart(CA) to Harbor by helm3 registry/chart CLI successfully;
        chart_cli_ret = library.helm.helm_chart_push_to_harbor(self.chart_file, self.archive,  harbor_server, TestProjects.project_push_chart_name, self.repo_name, self.verion, user_name, self.user_push_chart_password)
        print "chart_cli_ret:", chart_cli_ret

        #4. List artifacts successfully;
        artifacts = self.artifact.list_artifacts(TestProjects.project_push_chart_name, self.repo_name, **TestProjects.USER_CLIENT)
        self.assertEqual(artifacts[0].type, 'CHART')
        self.assertEqual(artifacts[0].tags[0].name, self.verion)

        #5.1 Get chart(CA) by reference successfully;
        artifact = self.artifact.get_reference_info(TestProjects.project_push_chart_name, self.repo_name, self.verion, **TestProjects.USER_CLIENT)
        self.assertEqual(artifact[0].type, 'CHART')
        self.assertEqual(artifact[0].tags[0].name, self.verion)

        #5.2 Chart bundle can be pulled by ctr successfully;
        #oci_ref = harbor_server+"/"+TestProjects.project_push_chart_name+"/"+self.repo_name+":"+self.verion
        #library.containerd.ctr_images_pull(user_name, self.user_push_chart_password, oci_ref)
        #library.containerd.ctr_images_list(oci_ref = oci_ref)

        #6. Get addtion successfully;
        addition_r = self.artifact.get_addition(TestProjects.project_push_chart_name, self.repo_name, self.verion, "readme.md", **TestProjects.USER_CLIENT)
        self.assertIn("Helm Chart for Harbor", addition_r[0])
        addition_d = self.artifact.get_addition(TestProjects.project_push_chart_name, self.repo_name, self.verion, "dependencies", **TestProjects.USER_CLIENT)
        self.assertIn("https://kubernetes-charts.storage.googleapis.com", addition_d[0])
        addition_v = self.artifact.get_addition(TestProjects.project_push_chart_name, self.repo_name, self.verion, "values.yaml", **TestProjects.USER_CLIENT)
        self.assertIn("adminserver", addition_v[0])

        #7. Delete chart by reference successfully.
        self.artifact.delete_artifact(TestProjects.project_push_chart_name, self.repo_name, self.verion, **TestProjects.USER_CLIENT)


if __name__ == '__main__':
    unittest.main()

