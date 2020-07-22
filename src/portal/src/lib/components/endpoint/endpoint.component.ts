// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import {
    Component,
    OnInit,
    OnDestroy,
    ViewChild,
    ChangeDetectionStrategy,
    ChangeDetectorRef
} from "@angular/core";
import { Subscription, Observable, forkJoin, throwError as observableThrowError } from "rxjs";
import { TranslateService } from "@ngx-translate/core";
import { Comparator } from "../../services/interface";

import { Endpoint } from "../../services/interface";
import { EndpointService } from "../../services/endpoint.service";

import { ErrorHandler } from "../../utils/error-handler";
import { map, catchError, finalize } from "rxjs/operators";
import { ConfirmationMessage } from "../confirmation-dialog/confirmation-message";
import { ConfirmationAcknowledgement } from "../confirmation-dialog/confirmation-state-message";
import { ConfirmationDialogComponent } from "../confirmation-dialog/confirmation-dialog.component";

import {
    ConfirmationTargets,
    ConfirmationState,
    ConfirmationButtons
} from "../../entities/shared.const";

import { CreateEditEndpointComponent } from "../create-edit-endpoint/create-edit-endpoint.component";
import { CustomComparator } from "../../utils/utils";
import { errorHandler as errorHandFn } from "../../utils/shared/shared.utils";

import { operateChanges, OperateInfo, OperationState } from "../operation/operate";
import { OperationService } from "../operation/operation.service";


@Component({
    selector: "hbr-endpoint",
    templateUrl: "./endpoint.component.html",
    styleUrls: ["./endpoint.component.scss"],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class EndpointComponent implements OnInit, OnDestroy {
    @ViewChild(CreateEditEndpointComponent, {static: false})
    createEditEndpointComponent: CreateEditEndpointComponent;

    @ViewChild("confirmationDialog", {static: false})
    confirmationDialogComponent: ConfirmationDialogComponent;

    targets: Endpoint[];
    target: Endpoint;

    targetName: string;
    subscription: Subscription;

    loading: boolean = false;

    creationTimeComparator: Comparator<Endpoint> = new CustomComparator<Endpoint>(
        "creation_time",
        "date"
    );

    timerHandler: any;
    selectedRow: Endpoint[] = [];

    get initEndpoint(): Endpoint {
        return {
            credential: {
                access_key: "",
                access_secret: "",
                type: ""
            },
            description: "",
            insecure: false,
            name: "",
            type: "",
            url: "",
        };
    }

    constructor(private endpointService: EndpointService,
        private errorHandler: ErrorHandler,
        private translateService: TranslateService,
        private operationService: OperationService,
        private ref: ChangeDetectorRef) {
    }

    ngOnInit(): void {
        this.targetName = "";
        this.retrieve();
    }

    ngOnDestroy(): void {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    }

    selectedChange(): void {
        this.forceRefreshView(5000);
    }

    retrieve(): void {
        this.loading = true;
        this.selectedRow = [];
        this.endpointService.getEndpoints(this.targetName).pipe(finalize(() => {
            this.loading = false;
            this.forceRefreshView(1000);
        }))
            .subscribe(targets => {
                this.targets = targets || [];
            }, error => {
                this.errorHandler.error(error);
            });
    }

    doSearchTargets(targetName: string) {
        this.targetName = targetName;
        this.retrieve();
    }

    refreshTargets() {
        this.retrieve();
    }

    reload($event: any) {
        this.targetName = "";
        this.retrieve();
    }

    openModal() {
        this.createEditEndpointComponent.openCreateEditTarget(true);
        this.target = this.initEndpoint;
    }

    editTargets(targets: Endpoint[]) {
        if (targets && targets.length === 1) {
            let target = targets[0];
            let editable = true;
            if (!target.id) {
                return;
            }
            let id: number | string = target.id;
            this.createEditEndpointComponent.openCreateEditTarget(editable, id);
        }
    }

    deleteTargets(targets: Endpoint[]) {
        if (targets && targets.length) {
            let targetNames: string[] = [];
            targets.forEach(target => {
                targetNames.push(target.name);
            });
            let deletionMessage = new ConfirmationMessage(
                'REPLICATION.DELETION_TITLE_TARGET',
                'REPLICATION.DELETION_SUMMARY_TARGET',
                targetNames.join(', ') || '',
                targets,
                ConfirmationTargets.TARGET,
                ConfirmationButtons.DELETE_CANCEL);
            this.confirmationDialogComponent.open(deletionMessage);
        }
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (message &&
            message.source === ConfirmationTargets.TARGET &&
            message.state === ConfirmationState.CONFIRMED) {
            let targetLists: Endpoint[] = message.data;
            if (targetLists && targetLists.length) {
                let observableLists: any[] = [];
                targetLists.forEach(target => {
                    observableLists.push(this.delOperate(target));
                });
                forkJoin(...observableLists)
                .pipe(finalize(() => {
                    this.selectedRow = [];
                    this.reload(true);
                    this.forceRefreshView(2000);
                }))
                .subscribe((item) => {
                }, error => {
                    this.errorHandler.error(error);
                });
            }
        }
    }
    delOperate(target: Endpoint): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_REGISTRY';
        operMessage.data.id = target.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = target.name;
        this.operationService.publishInfo(operMessage);

        return this.endpointService
            .deleteEndpoint(target.id)
            .pipe(map(
                response => {
                    this.translateService.get('BATCH.DELETED_SUCCESS')
                        .subscribe(res => {
                            operateChanges(operMessage, OperationState.success);
                        });
                })
                , catchError(error => {
                    const message = errorHandFn(error);
                    this.translateService.get(message).subscribe(res =>
                        operateChanges(operMessage, OperationState.failure, res)
                    );
                    return observableThrowError(error);
                }
                ));
    }

    // Forcely refresh the view
    forceRefreshView(duration: number): void {
        // Reset timer
        if (this.timerHandler) {
            clearInterval(this.timerHandler);
        }
        this.timerHandler = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => {
            if (this.timerHandler) {
                clearInterval(this.timerHandler);
                this.timerHandler = null;
            }
        }, duration);
    }

    getAdapterText(adapter: string): string {
        return this.endpointService.getAdapterText(adapter);
    }
}
