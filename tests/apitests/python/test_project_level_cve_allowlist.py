from __future__ import absolute_import

import unittest
import swagger_client
import time

from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User


class TestProjectCVEAllowlist(unittest.TestCase):
    """
    Test case:
        Project Level CVE Allowlist
    Setup:
        1.Admin creates project(PA)
        2.Create user(RA)
        3.Add user(RA) as a guest of project(PA)
    Test Steps:
        1. User(RA) reads the project(PA), verify the "reuse_sys_cve_allowlist" is empty in the metadata, and the CVE allowlist is empty
        2. User(RA) updates the project CVE allowlist, verify it fails with Forbidden error.
        3. Admin user updates User(RA) as project admin.
        4. User(RA) updates the project CVE allowlist with expiration date and one item in the items list.
        5. User(RA) reads the project(PA), verify the CVE allowlist is updated as step 4
        6. User(RA) updates the project CVE allowlist removes expiration date and clean the items.
        7. User(RA) reads the project(PA), verify the CVE allowlist is updated as step 6
        8. User(RA) updates the project metadata to set "reuse_sys_cve_allowlist" to true.
        9. User(RA) reads the project(PA) verify the project metadata is updated.
    Tear Down:
        1. Remove User(RA) from project(PA) as member
        2. Delete project(PA)
        3. Delete User(RA)
    """

    def setUp(self):
        self.user = User()
        self.project = Project()
        user_ra_password = "Aa123456"
        print("Setup: Creating user for test")
        user_ra_id, user_ra_name = self.user.create_user(user_password=user_ra_password, **ADMIN_CLIENT)
        print("Created user: %s, id: %s" % (user_ra_name, user_ra_id))
        self.USER_RA_CLIENT = dict(endpoint=ADMIN_CLIENT["endpoint"],
                                   username=user_ra_name,
                                   password=user_ra_password)
        self.user_ra_id = int(user_ra_id)
        p_id, _ = self.project.create_project(metadata = {"public": "false"}, **ADMIN_CLIENT)
        self.project_pa_id = int(p_id)
        m_id = self.project.add_project_members(self.project_pa_id, self.user_ra_id, member_role_id=3, **ADMIN_CLIENT)
        self.member_id = int(m_id)

    def tearDown(self):
        print("Tearing down...")
        self.project.delete_project_member(self.project_pa_id, self.member_id, **ADMIN_CLIENT)
        self.project.delete_project(self.project_pa_id,**ADMIN_CLIENT)
        self.user.delete_user(self.user_ra_id, **ADMIN_CLIENT)

    def testProjectLevelCVEAllowlist(self):
        # User(RA) reads the project(PA), verify the "reuse_sys_cve_allowlist" is empty in the metadata,
        # and the CVE allowlist is empty
        p = self.project.get_project(self.project_pa_id, **self.USER_RA_CLIENT)
        self.assertIsNone(p.metadata.reuse_sys_cve_allowlist)
        self.assertEqual(0, len(p.cve_allowlist.items))

        # User(RA) updates the project CVE allowlist, verify it fails with Forbidden error.
        item_list = [swagger_client.CVEAllowlistItem(cve_id="CVE-2019-12310")]
        exp = int(time.time()) + 1000
        wl = swagger_client.CVEAllowlist(expires_at=exp, items=item_list)
        self.project.update_project(self.project_pa_id, cve_allowlist=wl, expect_status_code=403, **self.USER_RA_CLIENT)

        # Admin user updates User(RA) as project admin.
        self.project.update_project_member_role(self.project_pa_id,self.member_id, 1, **ADMIN_CLIENT)

        # User(RA) updates the project CVE allowlist with expiration date and one item in the items list.
        self.project.update_project(self.project_pa_id, cve_allowlist=wl, **self.USER_RA_CLIENT)
        p = self.project.get_project(self.project_pa_id, **self.USER_RA_CLIENT)
        self.assertEqual("CVE-2019-12310", p.cve_allowlist.items[0].cve_id)
        self.assertEqual(exp, p.cve_allowlist.expires_at)

        # User(RA) updates the project CVE allowlist with empty items list
        wl2 = swagger_client.CVEAllowlist(items=[])
        self.project.update_project(self.project_pa_id, cve_allowlist=wl2, **self.USER_RA_CLIENT)
        p = self.project.get_project(self.project_pa_id, **self.USER_RA_CLIENT)
        self.assertEqual(0, len(p.cve_allowlist.items))
        self.assertIsNone(p.cve_allowlist.expires_at)

        # User(RA) updates the project metadata to set "reuse_sys_cve_allowlist" to true.
        meta = swagger_client.ProjectMetadata(reuse_sys_cve_allowlist="true")
        self.project.update_project(self.project_pa_id, metadata=meta, **self.USER_RA_CLIENT)
        p = self.project.get_project(self.project_pa_id, **self.USER_RA_CLIENT)
        self.assertEqual("true", p.metadata.reuse_sys_cve_allowlist)


if __name__ == '__main__':
    unittest.main()
