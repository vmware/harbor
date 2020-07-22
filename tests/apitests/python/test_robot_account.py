from __future__ import absolute_import

import unittest

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library.user import User
from library.project import Project
from library.repository import Repository
from library.repository import pull_harbor_image
from library.repository import push_image_to_project
from testutils import harbor_server
from library.base import _assert_status_code

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.repo = Repository()

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_ra_name_a, TestProjects.repo_name_in_project_a.split('/')[1], **TestProjects.USER_RA_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_ra_name_b, TestProjects.repo_name_in_project_b.split('/')[1], **TestProjects.USER_RA_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_ra_name_c, TestProjects.repo_name_in_project_c.split('/')[1], **TestProjects.USER_RA_CLIENT)
        self.repo.delete_repoitory(TestProjects.project_ra_name_a, TestProjects.repo_name_pa.split('/')[1], **TestProjects.USER_RA_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_ra_id_a, **TestProjects.USER_RA_CLIENT)
        self.project.delete_project(TestProjects.project_ra_id_b, **TestProjects.USER_RA_CLIENT)
        self.project.delete_project(TestProjects.project_ra_id_c, **TestProjects.USER_RA_CLIENT)

        #3. Delete user(UA).
        self.user.delete_user(TestProjects.user_ra_id, **ADMIN_CLIENT)

    def testRobotAccount(self):
        """
        Test case:
            Robot Account
        Test step and expected result:
			1. Create user(UA);
			2. Create private project(PA), private project(PB) and public project(PC) by user(UA);
			3. Push image(ImagePA) to project(PA), image(ImagePB) to project(PB) and image(ImagePC) to project(PC) by user(UA);
			4. Create a new robot account(RA) with pull and push privilige in project(PA) by user(UA);
			5. Check robot account info, it should has both pull and push priviliges;
			6. Pull image(ImagePA) from project(PA) by robot account(RA), it must be successful;
			7. Push image(ImageRA) to project(PA) by robot account(RA), it must be successful;
			8. Push image(ImageRA) to project(PB) by robot account(RA), it must be not successful;
			9. Pull image(ImagePB) from project(PB) by robot account(RA), it must be not successful;
			10. Pull image from project(PC), it must be successful;
			11. Push image(ImageRA) to project(PC) by robot account(RA), it must be not successful;
			12. Update action property of robot account(RA);
			13. Pull image(ImagePA) from project(PA) by robot account(RA), it must be not successful;
			14. Push image(ImageRA) to project(PA) by robot account(RA), it must be not successful;
			15. Delete robot account(RA), it must be not successful.
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        admin_name = ADMIN_CLIENT["username"]
        admin_password = ADMIN_CLIENT["password"]
        user_ra_password = "Aa123456"
        image_project_a = "haproxy"
        image_project_b = "hello-world"
        image_project_c = "httpd"
        image_robot_account = "alpine"
        tag = "latest"

        print "#1. Create user(UA);"
        TestProjects.user_ra_id, user_ra_name = self.user.create_user(user_password = user_ra_password, **ADMIN_CLIENT)
        TestProjects.USER_RA_CLIENT=dict(endpoint = url, username = user_ra_name, password = user_ra_password)

        print "#2. Create private project(PA), private project(PB) and public project(PC) by user(UA);"
        TestProjects.project_ra_id_a, TestProjects.project_ra_name_a = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_RA_CLIENT)
        TestProjects.project_ra_id_b, TestProjects.project_ra_name_b = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_RA_CLIENT)
        TestProjects.project_ra_id_c, TestProjects.project_ra_name_c = self.project.create_project(metadata = {"public": "true"}, **TestProjects.USER_RA_CLIENT)

        print "#3. Push image(ImagePA) to project(PA), image(ImagePB) to project(PB) and image(ImagePC) to project(PC) by user(UA);"
        TestProjects.repo_name_in_project_a, tag_a = push_image_to_project(TestProjects.project_ra_name_a, harbor_server, user_ra_name, user_ra_password, image_project_a, tag)
        TestProjects.repo_name_in_project_b, tag_b = push_image_to_project(TestProjects.project_ra_name_b, harbor_server, user_ra_name, user_ra_password, image_project_b, tag)
        TestProjects.repo_name_in_project_c, tag_c = push_image_to_project(TestProjects.project_ra_name_c, harbor_server, user_ra_name, user_ra_password, image_project_c, tag)

        print "#4. Create a new robot account(RA) with pull and push privilige in project(PA) by user(UA);"
        robot_id, robot_account = self.project.add_project_robot_account(TestProjects.project_ra_id_a, TestProjects.project_ra_name_a,
                                                                         2441000531 ,**TestProjects.USER_RA_CLIENT)
        print robot_account.name
        print robot_account.token

        print "#5. Check robot account info, it should has both pull and push priviliges;"
        data = self.project.get_project_robot_account_by_id(TestProjects.project_ra_id_a, robot_id, **TestProjects.USER_RA_CLIENT)
        _assert_status_code(robot_account.name, data.name)

        print "#6. Pull image(ImagePA) from project(PA) by robot account(RA), it must be successful;"
        pull_harbor_image(harbor_server, robot_account.name, robot_account.token, TestProjects.repo_name_in_project_a, tag_a)

        print "#7. Push image(ImageRA) to project(PA) by robot account(RA), it must be successful;"
        TestProjects.repo_name_pa, _ = push_image_to_project(TestProjects.project_ra_name_a, harbor_server, robot_account.name, robot_account.token, image_robot_account, tag)

        print "#8. Push image(ImageRA) to project(PB) by robot account(RA), it must be not successful;"
        push_image_to_project(TestProjects.project_ra_name_b, harbor_server, robot_account.name, robot_account.token, image_robot_account, tag, expected_error_message = "unauthorized to access repository")

        print "#9. Pull image(ImagePB) from project(PB) by robot account(RA), it must be not successful;"
        pull_harbor_image(harbor_server, robot_account.name, robot_account.token, TestProjects.repo_name_in_project_b, tag_b, expected_error_message = "unauthorized to access repository")

        print "#10. Pull image from project(PC), it must be successful;"
        pull_harbor_image(harbor_server, robot_account.name, robot_account.token, TestProjects.repo_name_in_project_c, tag_c)

        print "#11. Push image(ImageRA) to project(PC) by robot account(RA), it must be not successful;"
        push_image_to_project(TestProjects.project_ra_name_c, harbor_server, robot_account.name, robot_account.token, image_robot_account, tag, expected_error_message = "unauthorized to access repository")

        print "#12. Update action property of robot account(RA);"
        self.project.disable_project_robot_account(TestProjects.project_ra_id_a, robot_id, True, **TestProjects.USER_RA_CLIENT)

        print "#13. Pull image(ImagePA) from project(PA) by robot account(RA), it must be not successful;"
        pull_harbor_image(harbor_server, robot_account.name, robot_account.token, TestProjects.repo_name_in_project_a, tag_a, expected_login_error_message = "unauthorized: authentication required")

        print "#14. Push image(ImageRA) to project(PA) by robot account(RA), it must be not successful;"
        push_image_to_project(TestProjects.project_ra_name_a, harbor_server, robot_account.name, robot_account.token, image_robot_account, tag, expected_login_error_message = "unauthorized: authentication required")

        print "#15. Delete robot account(RA), it must be not successful."
        self.project.delete_project_robot_account(TestProjects.project_ra_id_a, robot_id, **TestProjects.USER_RA_CLIENT)

if __name__ == '__main__':
    unittest.main()