from __future__ import absolute_import

import unittest

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library.user import User
from library.system import System
from library.project import Project
from library.repository import Repository
from library.repository import push_image_to_project
from testutils import harbor_server
from library.base import _assert_status_code
from library.repository import push_special_image_to_project
from library.artifact import Artifact

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.system = System()
        self.project = Project()
        self.user = User()
        self.repo = Repository()
        self.artifact = Artifact()
        self.repo_name = "test_repo"
        self.repo_name_untag = "test_untag"
        self.tag = "v1.0"

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == True, "Test data won't be erased.")
    def test_ClearData(self):
        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_gc_id, **TestProjects.USER_GC_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_gc_id, **ADMIN_CLIENT)

    def testGarbageCollection(self):
        """
        Test case:
            Garbage Collection
        Test step and expected result:
            1. Create a new user(UA);
            2. Create project(PA) and project(PB) by user(UA);
            3. Push a image in project(PA) and then delete repository by admin;
            4. Get repository by user(UA), it should get nothing;
            5. Tigger garbage collection operation;
            6. Check garbage collection job was finished;
            7. Get garbage collection log, check there is a number of files was deleted;
            8. Push a image in project(PB) by admin and delete the only tag;
            9. Tigger garbage collection operation;
            10. Check garbage collection job was finished;
            11. Repository with untag image should be still there;
            12. But no any artifact in repository anymore.
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        admin_name = ADMIN_CLIENT["username"]
        admin_password = ADMIN_CLIENT["password"]
        user_gc_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_gc_id, user_gc_name = self.user.create_user(user_password = user_gc_password, **ADMIN_CLIENT)

        TestProjects.USER_GC_CLIENT=dict(endpoint = url, username = user_gc_name, password = user_gc_password)

        #2. Create project(PA) and project(PB) by user(UA);
        TestProjects.project_gc_id, TestProjects.project_gc_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_GC_CLIENT)
        TestProjects.project_gc_untag_id, TestProjects.project_gc_untag_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_GC_CLIENT)

        #3. Push a image in project(PA) and then delete repository by admin;
        push_special_image_to_project(TestProjects.project_gc_name, harbor_server, admin_name, admin_password, self.repo_name, ["latest", "v1.2.3"])
        self.repo.delete_repoitory(TestProjects.project_gc_name, self.repo_name, **TestProjects.USER_GC_CLIENT)

        #4. Get repository by user(UA), it should get nothing;
        repo_data = self.repo.list_repositories(TestProjects.project_gc_name, **TestProjects.USER_GC_CLIENT)
        _assert_status_code(len(repo_data), 0)

        #5. Tigger garbage collection operation;
        gc_id = self.system.gc_now(**ADMIN_CLIENT)

        #6. Check garbage collection job was finished;
        self.system.validate_gc_job_status(gc_id, "finished", **ADMIN_CLIENT)

        #7. Get garbage collection log, check there is a number of files was deleted;
        self.system.validate_deletion_success(gc_id, **ADMIN_CLIENT)

        #8. Push a image in project(PB) by admin and delete the only tag;
        push_special_image_to_project(TestProjects.project_gc_untag_name, harbor_server, admin_name, admin_password, self.repo_name_untag, [self.tag])
        self.artifact.delete_tag(TestProjects.project_gc_untag_name, self.repo_name_untag, self.tag, self.tag, **ADMIN_CLIENT)

        #9. Tigger garbage collection operation;
        gc_id = self.system.gc_now(**ADMIN_CLIENT)

        #10. Check garbage collection job was finished;
        self.system.validate_gc_job_status(gc_id, "finished", **ADMIN_CLIENT)

        #11. Repository with untag image should be still there;
        repo_data_untag = self.repo.list_repositories(TestProjects.project_gc_untag_name, **TestProjects.USER_GC_CLIENT)
        print "repo_data_untag:", repo_data_untag
        _assert_status_code(len(repo_data_untag), 1)
        self.assertEqual(TestProjects.project_gc_untag_name + "/" + self.repo_name_untag , repo_data_untag[0].name)

        #12. But no any artifact in repository anymore.
        artifacts = self.artifact.list_artifacts(TestProjects.project_gc_untag_name, self.repo_name_untag, **TestProjects.USER_GC_CLIENT)
        self.assertEqual(artifacts,[])



if __name__ == '__main__':
    unittest.main()