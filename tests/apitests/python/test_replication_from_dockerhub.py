from __future__ import absolute_import
import unittest

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.replication import Replication
from library.registry import Registry
from library.artifact import Artifact
from library.repository import Repository
import swagger_client

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.replication = Replication()
        self.registry = Registry()
        self.artifact = Artifact()
        self.repo = Repository()
        self.image = "alpine"
        self.tag = "latest"

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete rule(RA);
        self.replication.delete_replication_rule(TestProjects.rule_id, **ADMIN_CLIENT)

        #2. Delete registry(TA);
        self.registry.delete_registry(TestProjects.registry_id, **ADMIN_CLIENT)

        #1. Delete repository(RA);
        self.repo.delete_repoitory(TestProjects.project_name, self.image, **TestProjects.USER_add_rule_CLIENT)

        #3. Delete project(PA);
        self.project.delete_project(TestProjects.project_add_rule_id, **TestProjects.USER_add_rule_CLIENT)

        #4. Delete user(UA);
        self.user.delete_user(TestProjects.user_add_rule_id, **ADMIN_CLIENT)

    def testReplicationFromDockerhub(self):
        """
        Test case:
            Replication From Dockerhub
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Create a new registry;
            4. Create a new rule for this registry;
            5. Check rule should be exist;
            6. Trigger the rule;
            7. Wait for completion of this replication job;
            8. Check image is replicated into target project successfully.
        Tear down:
            1. Delete rule(RA);
            2. Delete registry(TA);
            3. Delete project(PA);
            4. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        user_add_rule_password = "Aa123456"

        #1. Create user(UA)
        TestProjects.user_add_rule_id, user_add_rule_name = self.user.create_user(user_password = user_add_rule_password, **ADMIN_CLIENT)

        TestProjects.USER_add_rule_CLIENT=dict(endpoint = url, username = user_add_rule_name, password = user_add_rule_password)

        #2.1. Create private project(PA) by user(UA)
        TestProjects.project_add_rule_id, TestProjects.project_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_add_rule_CLIENT)

        #2.2. Get private project of uesr-001, uesr-001 can see only one private project which is project-001
        self.project.projects_should_exist(dict(public=False), expected_count = 1,
            expected_project_id = TestProjects.project_add_rule_id, **TestProjects.USER_add_rule_CLIENT)

        #3. Create a new registry;
        TestProjects.registry_id, _ = self.registry.create_registry("https://hub.docker.com", registry_type="docker-hub", access_key = "", access_secret = "", insecure=False, **ADMIN_CLIENT)
        print "TestProjects.registry_id:", TestProjects.registry_id

        #4. Create a pull-based rule for this registry;
        TestProjects.rule_id, rule_name = self.replication.create_replication_policy(src_registry=swagger_client.Registry(id=int(TestProjects.registry_id)),
                                            dest_namespace=TestProjects.project_name,
                                            filters=[swagger_client.ReplicationFilter(type="name",value="library/"+self.image),swagger_client.ReplicationFilter(type="tag",value=self.tag)],
                                            **ADMIN_CLIENT)

        #5. Check rule should be exist;
        self.replication.check_replication_rule_should_exist(TestProjects.rule_id, rule_name, **ADMIN_CLIENT)

        #6. Trigger the rule;
        self.replication.trigger_replication_executions(TestProjects.rule_id, **ADMIN_CLIENT)

        #7. Wait for completion of this replication job;
        self.replication.wait_until_jobs_finish(TestProjects.rule_id, **ADMIN_CLIENT)

        #8. Check image is replicated into target project successfully.
        artifact = self.artifact.get_reference_info(TestProjects.project_name, self.image, self.tag, **ADMIN_CLIENT)
        print artifact


if __name__ == '__main__':
    unittest.main()
