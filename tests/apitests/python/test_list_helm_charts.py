from __future__ import absolute_import

import unittest

from testutils import ADMIN_CLIENT, CHART_API_CLIENT
from testutils import TEARDOWN
from library.user import User
from library.project import Project
from library.chart import Chart

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.chart= Chart()
        self.project= Project()
        self.user= User()

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete chart file;
        self.chart.delete_chart_with_version(TestProjects.project_chart_name, TestProjects.CHART_NAME, TestProjects.VERSION, **CHART_API_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_chart_id, **TestProjects.USER_CHART_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_chart_id, **ADMIN_CLIENT)

    def testListHelmCharts(self):
        """
        Test case:
            List Helm Charts
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Upload a chart file to project(PA);
            4. Chart file should be exist in project(PA).
        Tear down:
            1. Delete chart file;
            2. Delete project(PA);
            3. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        chart_api_url = CHART_API_CLIENT['endpoint']
        user_chart_password = 'Aa123456'
        TestProjects.CHART_NAME = 'mariadb'
        TestProjects.VERSION = '4.3.1'

        #1. Create a new user(UA);
        TestProjects.user_chart_id, user_chart_name = self.user.create_user(user_password = user_chart_password, **ADMIN_CLIENT)

        TestProjects.USER_CHART_CLIENT=dict(endpoint = url, username = user_chart_name, password = user_chart_password)

        TestProjects.API_CHART_CLIENT=dict(endpoint = chart_api_url, username = user_chart_name, password = user_chart_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_chart_id, TestProjects.project_chart_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CHART_CLIENT)

        #3. Upload a chart file to project(PA);
        self.chart.upload_chart(TestProjects.project_chart_name, r'./tests/apitests/python/mariadb-{}.tgz'.format(TestProjects.VERSION), **TestProjects.API_CHART_CLIENT)

        #4. Chart file should be exist in project(PA).
        self.chart.chart_should_exist(TestProjects.project_chart_name, TestProjects.CHART_NAME, **TestProjects.API_CHART_CLIENT)

if __name__ == '__main__':
    unittest.main()