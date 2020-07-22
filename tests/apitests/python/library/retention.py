# -*- coding: utf-8 -*-

import base
import swagger_client

class Retention(base.Base):
    def get_metadatas(self, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        metadatas, status_code, _ = client.retentions_metadatas_get_with_http_info()
        base._assert_status_code(expect_status_code, status_code)
        return metadatas

    def create_retention_policy(self, project_id, selector_repository="**", selector_tag="**", expect_status_code = 201, **kwargs):
        policy=swagger_client.RetentionPolicy(
            algorithm='or',
            rules=[
                swagger_client.RetentionRule(
                    disabled=False,
                    action="retain",
                    template="always",
                    params= {

                    },
                    scope_selectors={
                        "repository": [
                            {
                                "kind": "doublestar",
                                "decoration": "repoMatches",
                                "pattern": selector_repository
                            }
                        ]
                    },
                    tag_selectors=[
                        {
                            "kind": "doublestar",
                            "decoration": "matches",
                            "pattern": selector_tag
                        }
                    ]
                )
            ],
            trigger= {
                "kind": "Schedule",
                "settings": {
                    "cron": ""
                },
                "references": {
                }
            },
            scope= {
                "level": "project",
                "ref": project_id
            },
        )
        client = self._get_client(**kwargs)
        _, status_code, header = client.retentions_post_with_http_info(policy)
        base._assert_status_code(expect_status_code, status_code)
        return base._get_id_from_header(header)

    def get_retention_policy(self, retention_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        policy, status_code, _ = client.retentions_id_get_with_http_info(retention_id)
        base._assert_status_code(expect_status_code, status_code)
        return policy

    def update_retention_policy(self, retention_id, selector_repository="**", selector_tag="**", expect_status_code = 200, **kwargs):
        policy=swagger_client.RetentionPolicy(
            id=retention_id,
            algorithm='or',
            rules=[
                swagger_client.RetentionRule(
                    disabled=False,
                    action="retain",
                    template="always",
                    params= {

                    },
                    scope_selectors={
                        "repository": [
                            {
                                "kind": "doublestar",
                                "decoration": "repoMatches",
                                "pattern": selector_repository
                            }
                        ]
                    },
                    tag_selectors=[
                        {
                            "kind": "doublestar",
                            "decoration": "matches",
                            "pattern": selector_tag
                        }
                    ]
                )
            ],
            trigger= {
                "kind": "Schedule",
                "settings": {
                    "cron": ""
                },
                "references": {
                }
            },
            scope= {
                "level": "project",
                "ref": project_id
            },
        )
        client = self._get_client(**kwargs)
        _, status_code, _ = client.retentions_id_put_with_http_info(retention_id, policy)
        base._assert_status_code(expect_status_code, status_code)

    def update_retention_add_rule(self, retention_id, selector_repository="**", selector_tag="**", with_untag="True", expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        policy, status_code, _ = client.retentions_id_get_with_http_info(retention_id)
        base._assert_status_code(200, status_code)
        policy.rules.append(swagger_client.RetentionRule(
                                                disabled=False,
                                                action="retain",
                                                template="always",
                                                params= {

                                                },
                                                scope_selectors={
                                                    "repository": [
                                                        {
                                                            "kind": "doublestar",
                                                            "decoration": "repoMatches",
                                                            "pattern": selector_repository
                                                        }
                                                    ]
                                                },
                                                tag_selectors=[
                                                    {
                                                        "kind": "doublestar",
                                                        "decoration": "matches",
                                                        "extras":'["untagged":'+with_untag+']',
                                                        "pattern": selector_tag
                                                    }
                                                ]
                                            ))
        _, status_code, _ = client.retentions_id_put_with_http_info(retention_id, policy)
        base._assert_status_code(expect_status_code, status_code)

    def trigger_retention_policy(self, retention_id, dry_run=False, expect_status_code = 201, **kwargs):
        client = self._get_client(**kwargs)

        _, status_code, _ = client.retentions_id_executions_post_with_http_info(retention_id, {"dry_run":dry_run})
        base._assert_status_code(expect_status_code, status_code)

    def stop_retention_execution(self, retention_id, exec_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)

        r, status_code, _ = client.retentions_id_executions_eid_patch(retention_id, exec_id, {"action":"stop"})
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_executions(self, retention_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)

        r, status_code, _ = client.retentions_id_executions_get_with_http_info(retention_id)
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_exec_tasks(self, retention_id, exec_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)

        r, status_code, _ = client.retentions_id_executions_eid_tasks_get_with_http_info(retention_id, exec_id)
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_exec_task_log(self, retention_id, exec_id, task_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)

        r, status_code, _ = client.retentions_id_executions_eid_tasks_tid_get_with_http_info(retention_id, exec_id, task_id)
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_metadatas(self, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)

        r, status_code, _ = client.retentions_metadatas_get_with_http_info()
        base._assert_status_code(expect_status_code, status_code)
        return r
