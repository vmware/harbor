# -*- coding: utf-8 -*-

import time
import re
import base
import swagger_client
from swagger_client.rest import ApiException

class System(base.Base):
    def get_gc_history(self, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_get_with_http_info()
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def get_gc_status_by_id(self, job_id, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_id_get_with_http_info(job_id)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def get_gc_log_by_id(self, job_id, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_id_log_get_with_http_info(job_id)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def get_gc_schedule(self, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        try:
            data, status_code, _ = client.system_gc_schedule_get_with_http_info()
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def set_gc_schedule(self, schedule_type = 'None', cron = None, expect_status_code = 200, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)
        gc_schedule = swagger_client.AdminJobSchedule()
        gc_schedule.type = schedule_type
        if cron is not None:
            gc_schedule.cron = cron
        try:
            data, status_code, _ = client.system_gc_schedule_put_with_http_info(gc_schedule)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Get configuration response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Get configuration result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return data

    def create_gc_schedule(self, schedule_type, cron = None, expect_status_code = 201, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)

        gc_parameters = {'delete_untagged':True}

        gc_schedule = swagger_client.AdminJobScheduleObj()
        gc_schedule.type = schedule_type
        if cron is not None:
            gc_schedule.cron = cron

        gc_job = swagger_client.AdminJobSchedule()
        gc_job.schedule = gc_schedule
        gc_job.parameters = gc_parameters

        try:
            _, status_code, header = client.system_gc_schedule_post_with_http_info(gc_job)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Create GC schedule response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Create GC schedule result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return base._get_id_from_header(header)

    def create_scan_all_schedule(self, schedule_type, cron = None, expect_status_code = 201, expect_response_body = None, **kwargs):
        client = self._get_client(**kwargs)
        scanschedule = swagger_client.AdminJobScheduleObj()
        scanschedule.type = schedule_type
        if cron is not None:
            scanschedule.cron = cron

        scan_all_schedule = swagger_client.AdminJobSchedule(scanschedule)
        try:
            _, status_code, header = client.system_scan_all_schedule_post_with_http_info(scan_all_schedule)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Create Scan All schedule response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Create Scan All schedule result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)
        return base._get_id_from_header(header)

    def scan_now(self, **kwargs):
        scan_all_id = self.create_scan_all_schedule('Manual', **kwargs)
        return scan_all_id

    def gc_now(self, **kwargs):
        gc_id = self.create_gc_schedule('Manual', **kwargs)
        return gc_id

    def validate_gc_job_status(self, gc_id, expected_gc_status, **kwargs):
        get_gc_status_finish = False
        timeout_count = 20
        while timeout_count > 0:
            time.sleep(5)
            status = self.get_gc_status_by_id(gc_id, **kwargs)
            print("GC job No: {}, status: {}".format(timeout_count, status.job_status))
            if status.job_status == expected_gc_status:
                get_gc_status_finish = True
                break
            timeout_count = timeout_count - 1

        if not (get_gc_status_finish):
            raise Exception("GC status is not as expected '{}' actual GC status is '{}'".format(expected_gc_status, status.job_status))

    def validate_deletion_success(self, gc_id, **kwargs):
        log_content = self.get_gc_log_by_id(gc_id, **kwargs)
        key_message = "manifests eligible for deletion"
        key_message_pos = log_content.find(key_message)
        full_message = log_content[key_message_pos-30 : key_message_pos + len(key_message)]
        deleted_files_count_list = re.findall(r'\s+(\d+)\s+blobs\s+and\s+\d+\s+manifests\s+eligible\s+for\s+deletion', full_message)

        if len(deleted_files_count_list) != 1:
            raise Exception(r"Fail to get blobs eligible for deletion in log file, failure is {}.".format(len(deleted_files_count_list)))
        deleted_files_count = int(deleted_files_count_list[0])
        if deleted_files_count == 0:
            raise Exception(r"Get blobs eligible for deletion count is {}, while we expect more than 1.".format(deleted_files_count))

    def set_cve_allowlist(self, expires_at=None, expected_status_code=200, *cve_ids, **kwargs):
        client = self._get_client(**kwargs)
        cve_list = [swagger_client.CVEAllowlistItem(cve_id=c) for c in cve_ids]
        allowlist = swagger_client.CVEAllowlist(expires_at=expires_at, items=cve_list)
        try:
            r = client.system_cve_allowlist_put_with_http_info(allowlist=allowlist, _preload_content=False)
        except Exception as e:
            base._assert_status_code(expected_status_code, e.status)
        else:
            base._assert_status_code(expected_status_code, r[1])

    def get_cve_allowlist(self, **kwargs):
        client = self._get_client(**kwargs)
        return client.system_cve_allowlist_get()

    def get_project_quota(self, reference, reference_id, **kwargs):
        params={}
        params['reference'] = reference
        params['reference_id'] = reference_id

        client = self._get_client(**kwargs)
        data, status_code, _ = client.quotas_get_with_http_info(**params)
        base._assert_status_code(200, status_code)
        return data
