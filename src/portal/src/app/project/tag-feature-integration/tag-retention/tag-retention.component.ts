// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AddRuleComponent } from "./add-rule/add-rule.component";
import {ClrDatagridStateInterface, ClrDatagridStringFilterInterface} from "@clr/angular";
import { TagRetentionService } from "./tag-retention.service";
import { Retention, Rule } from "./retention";

import { Project } from "../../project";

import { finalize } from "rxjs/operators";
import { CronScheduleComponent } from "../../../../lib/components/cron-schedule";
import { ErrorHandler } from "../../../../lib/utils/error-handler";
import { OriginCron } from "../../../../lib/services";
import {clone, DEFAULT_PAGE_SIZE} from "../../../../lib/utils/utils";

const MIN = 60000;
const SEC = 1000;
const MIN_STR = "min";
const SEC_STR = "sec";
const SCHEDULE_TYPE = {
    NONE: "None",
    DAILY: "Daily",
    WEEKLY: "Weekly",
    HOURLY: "Hourly",
    CUSTOM: "Custom"
};
const DECORATION = {
    MATCHES: "matches",
    EXCLUDES: "excludes",
};
const RUNNING: string = "Running";
const PENDING: string = "pending";
const TIMEOUT: number = 5000;
@Component({
    selector: 'tag-retention',
    templateUrl: './tag-retention.component.html',
    styleUrls: ['./tag-retention.component.scss']
})
export class TagRetentionComponent implements OnInit {
    serialFilter: ClrDatagridStringFilterInterface<any> = {
        accepts(item: any, search: string): boolean {
            return item.id.toString().indexOf(search) !== -1;
        }
    };
    statusFilter: ClrDatagridStringFilterInterface<any> = {
        accepts(item: any, search: string): boolean {
            return item.status.toLowerCase().indexOf(search.toLowerCase()) !== -1;
        }
    };
    dryRunFilter: ClrDatagridStringFilterInterface<any> = {
        accepts(item: any, search: string): boolean {
            let str = item.dry_run ? 'YES' : 'NO';
            return str.indexOf(search) !== -1;
        }
    };
    projectId: number;
    isRetentionRunOpened: boolean = false;
    isAbortedOpened: boolean = false;
    isConfirmOpened: boolean = false;
    cron: string;
    selectedItem: any = null;
    ruleIndex: number = -1;
    index: number = -1;
    retentionId: number;
    retention: Retention = new Retention();
    editIndex: number;
    executionList = [];
    executionId: number;
    historyList = [];
    loadingExecutions: boolean = true;
    loadingHistories: boolean = true;
    label: string = 'TAG_RETENTION.TRIGGER';
    loadingRule: boolean = false;
    currentPage: number = 1;
    pageSize: number = DEFAULT_PAGE_SIZE;
    totalCount: number = 0;
    currentLogPage: number = 1;
    totalLogCount: number = 0;
    logPageSize: number = 5;
    isDetailOpened: boolean = false;
    @ViewChild('cronScheduleComponent')
    cronScheduleComponent: CronScheduleComponent;
    @ViewChild('addRule') addRuleComponent: AddRuleComponent;
    constructor(
        private route: ActivatedRoute,
        private tagRetentionService: TagRetentionService,
        private errorHandler: ErrorHandler,
    ) {
    }
    originCron(): OriginCron {
        let originCron: OriginCron = {
            type: SCHEDULE_TYPE.NONE,
            cron: ""
        };
        originCron.cron = this.retention.trigger.settings.cron;
        if (originCron.cron === "") {
            originCron.type = SCHEDULE_TYPE.NONE;
        } else if (originCron.cron === "0 0 * * * *") {
            originCron.type = SCHEDULE_TYPE.HOURLY;
        } else if (originCron.cron === "0 0 0 * * *") {
            originCron.type = SCHEDULE_TYPE.DAILY;
        } else if (originCron.cron === "0 0 0 * * 0") {
            originCron.type = SCHEDULE_TYPE.WEEKLY;
        } else {
            originCron.type = SCHEDULE_TYPE.CUSTOM;
        }
        return originCron;
    }

    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        this.retention.scope = {
            level: "project",
            ref: this.projectId
        };
        this.refreshAfterCreatRetention();
        this.getMetadata();
    }
    openConfirm(cron: string) {
      if (cron) {
        this.isConfirmOpened = true;
        this.cron = cron;
      } else {
        this.updateCron(cron);
      }
    }
    closeConfirm() {
      this.isConfirmOpened = false;
      this.updateCron(this.cron);
    }
    updateCron(cron: string) {
        let retention: Retention = clone(this.retention);
        retention.trigger.settings.cron = cron;
        if (!this.retentionId) {
            this.tagRetentionService.createRetention(retention).subscribe(
                response => {
                    this.cronScheduleComponent.isEditMode = false;
                    this.refreshAfterCreatRetention();
                }, error => {
                    this.errorHandler.error(error);
                });
        } else {
            this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
                response => {
                    this.cronScheduleComponent.isEditMode = false;
                    this.getRetention();
                }, error => {
                    this.errorHandler.error(error);
                });
        }
    }
    getMetadata() {
        this.tagRetentionService.getRetentionMetadata().subscribe(
            response => {
                this.addRuleComponent.metadata = response;
            }, error => {
                this.errorHandler.error(error);
            });
    }

    getRetention() {
        if (this.retentionId) {
            this.tagRetentionService.getRetention(this.retentionId).subscribe(
                response => {
                    if (response && response.rules && response.rules.length > 0) {
                        response.rules.forEach(item => {
                            if (!item.params) {
                                item.params = {};
                            }
                            this.setRuleUntagged(item);

                        });
                    }
                    this.retention = response;
                    this.loadingRule = false;
                }, error => {
                    this.errorHandler.error(error);
                    this.loadingRule = false;
                });
        }
    }

    editRuleByIndex(index) {
        this.editIndex = index;
        this.addRuleComponent.rule = clone(this.retention.rules[index]);
        this.addRuleComponent.editRuleOrigin = clone(this.retention.rules[index]);
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = false;
        this.ruleIndex = -1;
    }
    toggleDisable(index, isActionDisable) {
        let retention: Retention = clone(this.retention);
        retention.rules[index].disabled = isActionDisable;
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
            response => {
                this.getRetention();
            }, error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            });
    }
    deleteRule(index) {
        let retention: Retention = clone(this.retention);
        retention.rules.splice(index, 1);
        // if rules is empty, clear schedule.
        if (retention.rules && retention.rules.length === 0) {
          retention.trigger.settings.cron = "";
        }
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
            response => {
                this.getRetention();
            }, error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            });
    }
    setRuleUntagged(rule) {
        if (!rule.tag_selectors[0].extras) {
            if (rule.tag_selectors[0].decoration === DECORATION.MATCHES) {
                rule.tag_selectors[0].extras = JSON.stringify({untagged: true});
            }
            if (rule.tag_selectors[0].decoration === DECORATION.EXCLUDES) {
                rule.tag_selectors[0].extras = JSON.stringify({untagged: false});

            }
        } else {
            let extras = JSON.parse(rule.tag_selectors[0].extras);
            if (extras.untagged === undefined) {
                if (rule.tag_selectors[0].decoration === DECORATION.MATCHES) {
                    extras.untagged = true;
                }
                if (rule.tag_selectors[0].decoration === DECORATION.EXCLUDES) {
                    extras.untagged = false;
                }
                rule.tag_selectors[0].extras = JSON.stringify(extras);
            }
        }
    }
    openAddRule() {
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = true;
        this.addRuleComponent.rule = new Rule();
    }

    runRetention() {
        this.isRetentionRunOpened = false;
        this.tagRetentionService.runNowTrigger(this.retentionId).subscribe(
            response => {
                this.refreshList();
            }, error => {
                this.errorHandler.error(error);
            });
    }

    whatIfRun() {
        this.tagRetentionService.whatIfRunTrigger(this.retentionId).subscribe(
            response => {
                this.refreshList();
            }, error => {
                this.errorHandler.error(error);
            });
    }

    refreshList(state?: ClrDatagridStateInterface) {
        this.index = -1 ;
        this.selectedItem = null;
        this.loadingExecutions = true;
        if (this.retentionId) {
            if (state && state.page) {
                this.pageSize = state.page.size;
            }
            this.tagRetentionService.getRunNowList(this.retentionId, this.currentPage, this.pageSize)
              .pipe(finalize(() => this.loadingExecutions = false))
              .subscribe(
                  (response: any) => {
                    // Get total count
                    if (response.headers) {
                        let xHeader: string = response.headers.get("x-total-count");
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.executionList = response.body as Array<any>;
                    TagRetentionComponent.calculateDuration(this.executionList);
                }, error => {
                    this.errorHandler.error(error);
                });
        } else {
          setTimeout(() => {
            this.loadingExecutions = false;
          });
        }
    }

    static calculateDuration(arr: Array<any>) {
        if (arr && arr.length > 0) {
            for (let i = 0; i < arr.length; i++) {
                if (arr[i].end_time && arr[i].start_time) {
                    let duration = new Date(arr[i].end_time).getTime() - new Date(arr[i].start_time).getTime();
                    let min = Math.floor(duration / MIN);
                    let sec = Math.floor((duration % MIN) / SEC);
                    arr[i]['duration'] = "";
                    if ((min || sec) && duration > 0) {
                        if (min) {
                            arr[i]['duration'] += '' + min + MIN_STR;
                        }
                        if (sec) {
                            arr[i]['duration'] += '' + sec + SEC_STR;
                        }
                    } else {
                        arr[i]['duration'] = "0";
                    }
                } else {
                    arr[i]['duration'] = "N/A";
                }
            }
        }
    }

    abortRun() {
        this.isAbortedOpened = true;
        this.tagRetentionService.AbortRun(this.retentionId, this.selectedItem.id).subscribe(
            res => {
                this.refreshList();
            }, error => {
                this.errorHandler.error(error);
            });
    }

    abortRetention() {
        this.isAbortedOpened = false;
    }

    openEditor(index) {
        if (this.ruleIndex !== index) {
            this.ruleIndex = index;
        } else {
            this.ruleIndex = -1;
        }
    }

    loadLog() {
        if (this.isDetailOpened) {
            setTimeout(() => {// when this.isDetailOpened is true, need to wait ngCheck finished
                this.loadingHistories = true;
                this.tagRetentionService.getExecutionHistory(this.retentionId, this.executionId, this.currentLogPage, this.logPageSize)
                    .pipe(finalize(() => this.loadingHistories = false))
                    .subscribe(
                        (response: any) => {
                            // Get total count
                            if (response.headers) {
                                let xHeader: string = response.headers.get("x-total-count");
                                if (xHeader) {
                                    this.totalLogCount = parseInt(xHeader, 0);
                                }
                            }
                            this.historyList = response.body as Array<any>;
                            TagRetentionComponent.calculateDuration(this.historyList);
                            if (this.historyList && this.historyList.length
                                && this.historyList.some(item => {
                                    return item.status === RUNNING || item.status === PENDING;
                                })) {
                                setTimeout(() => {
                                    this.loadLog();
                                }, TIMEOUT);
                            }
                        }, error => {
                            this.errorHandler.error(error);
                        });
            }, 0);
        }
    }
    openDetail(index, executionId) {
        if (this.index !== index) {
            this.index = index;
            this.historyList = [];
            this.executionId = executionId;
            this.isDetailOpened = true;
        } else {
            this.index = -1;
            this.isDetailOpened = false;
        }
    }

    refreshAfterCreatRetention() {
        this.tagRetentionService.getProjectInfo(this.projectId).subscribe(
            response => {
                this.retentionId = response.metadata.retention_id;
                this.refreshList();
                this.getRetention();
            }, error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            });
    }

    clickAdd(rule) {
        this.loadingRule = true;
        this.addRuleComponent.onGoing = true;
        if (this.addRuleComponent.isAdd) {
            let retention: Retention = clone(this.retention);
            retention.rules.push(rule);
            if (!this.retentionId) {
                this.tagRetentionService.createRetention(retention).subscribe(
                    response => {
                        this.refreshAfterCreatRetention();
                        this.addRuleComponent.close();
                        this.addRuleComponent.onGoing = false;
                    }, error => {
                        if (error && error.error && error.error.message) {
                            error = this.tagRetentionService.getI18nKey(error.error.message);
                        }
                        this.addRuleComponent.inlineAlert.showInlineError(error);
                        this.loadingRule = false;
                        this.addRuleComponent.onGoing = false;
                    });
            } else {
                this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
                    response => {
                        this.getRetention();
                        this.addRuleComponent.close();
                        this.addRuleComponent.onGoing = false;
                    }, error => {
                        this.loadingRule = false;
                        this.addRuleComponent.onGoing = false;
                      if (error && error.error && error.error.message) {
                          error = this.tagRetentionService.getI18nKey(error.error.message);
                      }
                      this.addRuleComponent.inlineAlert.showInlineError(error);
                    });
            }
        } else {
            let retention: Retention = clone(this.retention);
            retention.rules[this.editIndex] = rule;
            this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
                response => {
                    this.getRetention();
                    this.addRuleComponent.close();
                    this.addRuleComponent.onGoing = false;
                }, error => {
                    if (error && error.error && error.error.message) {
                        error = this.tagRetentionService.getI18nKey(error.error.message);
                    }
                    this.addRuleComponent.inlineAlert.showInlineError(error);
                    this.loadingRule = false;
                    this.addRuleComponent.onGoing = false;
                });
        }
    }

    seeLog(executionId, taskId) {
        this.tagRetentionService.seeLog(this.retentionId, executionId, taskId);
    }

    formatPattern(pattern: string): string {
        return pattern.replace(/[{}]/g, "");
    }

    getI18nKey(str: string) {
        return this.tagRetentionService.getI18nKey(str);
    }
    clrLoad(state: ClrDatagridStateInterface) {
        this.refreshList(state);
    }
    /**
     *
     * @param extras Json string
     */
    showUntagged(extras) {
        return JSON.parse(extras).untagged;
    }
}
