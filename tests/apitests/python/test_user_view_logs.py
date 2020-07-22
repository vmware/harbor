from __future__ import absolute_import

import unittest
import time

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from testutils import TestResult
from library.user import User
from library.projectV2 import ProjectV2
from library.project import Project
from library.repository import Repository
from library.repository import push_image_to_project
from testutils import harbor_server

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        test_result = TestResult()
        self.test_result= test_result

        project = Project()
        self.project= project

        user = User()
        self.user= user

        repo = Repository()
        self.repo= repo

        projectv2 = ProjectV2()
        self.projectv2= projectv2

    @classmethod
    def tearDown(self):
        self.test_result.get_final_result()
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete project(PA);
        self.project.delete_project(TestProjects.project_user_view_logs_id, **TestProjects.USER_USER_VIEW_LOGS_CLIENT)

        #2. Delete user(UA);
        self.user.delete_user(TestProjects.user_user_view_logs_id, **ADMIN_CLIENT)

    def testUserViewLogs(self):
        """
        Test case:
            User View Logs
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA), in project(PA), there should be 1 'create' log record;;
            3. Push a new image(IA) in project(PA) by admin, in project(PA), there should be 1 'push' log record;;
            4. Delete repository(RA) by user(UA), in project(PA), there should be 1 'delete' log record;;
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        admin_name = ADMIN_CLIENT["username"]
        admin_password = ADMIN_CLIENT["password"]
        user_content_trust_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_user_view_logs_id, user_user_view_logs_name = self.user.create_user(user_password = user_content_trust_password, **ADMIN_CLIENT)

        TestProjects.USER_USER_VIEW_LOGS_CLIENT=dict(endpoint = url, username = user_user_view_logs_name, password = user_content_trust_password)

        #2.1 Create a new project(PA) by user(UA);
        TestProjects.project_user_view_logs_id, project_user_view_logs_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_USER_VIEW_LOGS_CLIENT)
        time.sleep(2)

        #2.2 In project(PA), there should be 1 'create' log record;
        operation = "create"
        log_count = self.projectv2.filter_project_logs(project_user_view_logs_name, user_user_view_logs_name, project_user_view_logs_name, "project", operation, **TestProjects.USER_USER_VIEW_LOGS_CLIENT)
        if log_count != 1:
            self.test_result.add_test_result("1 - Failed to get log with user:{}, resource:{}, resource_type:{} and operation:{}, expect count 1, but actual is {}.".
                                             format(user_user_view_logs_name, project_user_view_logs_name, "project", operation, log_count))

        #3.1 Push a new image(IA) in project(PA) by admin;
        repo_name, tag = push_image_to_project(project_user_view_logs_name, harbor_server, admin_name, admin_password, "tomcat", "latest")
        time.sleep(2)

        #3.2 In project(PA), there should be 1 'push' log record;
        operation = "create"
        log_count = self.projectv2.filter_project_logs(project_user_view_logs_name,  admin_name, r'{}:{}'.format(repo_name, tag), "artifact", operation, **TestProjects.USER_USER_VIEW_LOGS_CLIENT)
        if log_count != 1:
            self.test_result.add_test_result("2 - Failed to get log with user:{}, resource:{}, resource_type:{} and operation:{}, expect count 1, but actual is {}.".
                                             format(user_user_view_logs_name, project_user_view_logs_name, "artifact", operation, log_count))
        #4.1 Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(project_user_view_logs_name, repo_name.split('/')[1], **TestProjects.USER_USER_VIEW_LOGS_CLIENT)
        time.sleep(6)

        #4.2 In project(PA), there should be 1 'delete' log record;
        operation = "delete"
        log_count = self.projectv2.filter_project_logs(project_user_view_logs_name, user_user_view_logs_name, repo_name, "repository", operation, **TestProjects.USER_USER_VIEW_LOGS_CLIENT)
        if log_count != 1:
            self.test_result.add_test_result("5 - Failed to get log with user:{}, resource:{}, resource_type:{} and operation:{}, expect count 1, but actual is {}.".
                                             format(user_user_view_logs_name, project_user_view_logs_name, "repository", operation, log_count))

if __name__ == '__main__':
    unittest.main()