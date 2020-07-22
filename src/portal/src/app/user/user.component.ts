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
import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import { Subscription, Observable, forkJoin } from "rxjs";
import { TranslateService } from '@ngx-translate/core';
import { ConfirmationState, ConfirmationTargets, ConfirmationButtons } from '../shared/shared.const';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { SessionService } from '../shared/session.service';
import { AppConfigService } from '../services/app-config.service';
import { NewUserModalComponent } from './new-user-modal.component';
import { UserService } from './user.service';
import { User } from './user';
import { ChangePasswordComponent } from "./change-password/change-password.component";
import { map, catchError } from 'rxjs/operators';
import { throwError as observableThrowError } from "rxjs";
import { OperationService } from "../../lib/components/operation/operation.service";
import { operateChanges, OperateInfo, OperationState } from "../../lib/components/operation/operate";
import { errorHandler } from "../../lib/utils/shared/shared.utils";

/**
 * NOTES:
 *   Pagination for this component is a temporary workaround solution. It will be replaced in future release.
 *
 **
 * class UserComponent
 * @implements {OnInit}
 * @implements {OnDestroy}
 */

@Component({
    selector: 'harbor-user',
    templateUrl: 'user.component.html',
    styleUrls: ['user.component.scss'],
    providers: [UserService],
})

export class UserComponent implements OnInit, OnDestroy {
    users: User[] = [];
    selectedRow: User[] = [];
    ISADMNISTRATOR: string = "USER.ENABLE_ADMIN_ACTION";

    currentTerm: string;
    totalCount: number = 0;
    currentPage: number = 1;
    pageSize: number = 15;
    timerHandler: any;

    private onGoing: boolean = true;
    private adminMenuText: string = "";
    private adminColumn: string = "";
    private deletionSubscription: Subscription;

    @ViewChild(NewUserModalComponent, {static: true})
    newUserDialog: NewUserModalComponent;
    @ViewChild(ChangePasswordComponent, {static: true})
    changePwdDialog: ChangePasswordComponent;

    constructor(
        private userService: UserService,
        private translate: TranslateService,
        private deletionDialogService: ConfirmationDialogService,
        private msgHandler: MessageHandlerService,
        private session: SessionService,
        private appConfigService: AppConfigService,
        private operationService: OperationService) {
        this.deletionSubscription = deletionDialogService.confirmationConfirm$.subscribe(confirmed => {
            if (confirmed &&
                confirmed.source === ConfirmationTargets.USER &&
                confirmed.state === ConfirmationState.CONFIRMED) {
                this.delUser(confirmed.data);
            }
        });
    }

    isMySelf(uid: number): boolean {
        let currentUser = this.session.getCurrentUser();
        if (currentUser) {
            if (currentUser.user_id === uid) {
                return true;
            }
        }

        return false;
    }

    get onlySelf(): boolean {
        return this.selectedRow.length === 1 && this.isMySelf(this.selectedRow[0].user_id);
    }

    public get canCreateUser(): boolean {
        let appConfig = this.appConfigService.getConfig();
        if (appConfig) {
            return !(appConfig.auth_mode === 'ldap_auth' || appConfig.auth_mode === 'uaa_auth' || appConfig.auth_mode === 'oidc_auth');
        } else {
            return true;
        }
    }

    public get ifSameRole(): boolean {
        let usersRole: number[] = [];
        this.selectedRow.forEach(user => {
            if (user.user_id === 0 || this.isMySelf(user.user_id)) {
                return false;
            }
            if (user.sysadmin_flag) {
                usersRole.push(1);
            } else {
                usersRole.push(0);
            }
        });
        if (usersRole.length && usersRole.every(num => num === 0)) {
            this.ISADMNISTRATOR = 'USER.ENABLE_ADMIN_ACTION';
            return true;
        }
        if (usersRole.length && usersRole.every(num => num === 1)) {
            this.ISADMNISTRATOR = 'USER.DISABLE_ADMIN_ACTION';
            return true;
        }
        return false;
    }

    isSystemAdmin(u: User): string {
        if (!u) {
            return "{{MISS}}";
        }
        let key: string = u.sysadmin_flag ? "USER.IS_ADMIN" : "USER.IS_NOT_ADMIN";
        this.translate.get(key).subscribe((res: string) => this.adminColumn = res);
        return this.adminColumn;
    }

    adminActions(u: User): string {
        if (!u) {
            return "{{MISS}}";
        }
        let key: string = u.sysadmin_flag ? "USER.DISABLE_ADMIN_ACTION" : "USER.ENABLE_ADMIN_ACTION";
        this.translate.get(key).subscribe((res: string) => this.adminMenuText = res);
        return this.adminMenuText;
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    ngOnInit(): void {
    }

    ngOnDestroy(): void {
        if (this.deletionSubscription) {
            this.deletionSubscription.unsubscribe();
        }

        if (this.timerHandler) {
            clearInterval(this.timerHandler);
            this.timerHandler = null;
        }
    }

    openChangePwdModal(): void {
        if (this.selectedRow.length === 1) {
            this.changePwdDialog.open(this.selectedRow[0].user_id);
        }
    }

    // Filter items by keywords
    doFilter(terms: string): void {
        this.selectedRow = [];
        this.currentTerm = terms.trim();
        this.currentPage = 1;
        this.onGoing = true;
        this.getUserListByPaging();
    }

    // Disable the admin role for the specified user
    changeAdminRole(): void {
        let observableLists: any[] = [];
        if (this.selectedRow.length) {
            if (this.ISADMNISTRATOR === 'USER.ENABLE_ADMIN_ACTION') {
                for (let i = 0; i < this.selectedRow.length; i++) {
                    // Double confirm user is existing
                    if (this.selectedRow[i].user_id === 0 || this.isMySelf(this.selectedRow[i].user_id)) {
                        continue;
                    }
                    let updatedUser: User = new User();
                    updatedUser.user_id = this.selectedRow[i].user_id;

                    updatedUser.sysadmin_flag = true; // Set as admin
                    observableLists.push(this.userService.updateUserRole(updatedUser));
                }
            }
            if (this.ISADMNISTRATOR === 'USER.DISABLE_ADMIN_ACTION') {
                for (let i = 0; i < this.selectedRow.length; i++) {
                    // Double confirm user is existing
                    if (this.selectedRow[i].user_id === 0 || this.isMySelf(this.selectedRow[i].user_id)) {
                        continue;
                    }
                    let updatedUser: User = new User();
                    updatedUser.user_id = this.selectedRow[i].user_id;

                    updatedUser.sysadmin_flag = false; // Set as none admin
                    observableLists.push(this.userService.updateUserRole(updatedUser));
                }
            }

            forkJoin(...observableLists).subscribe(() => {
                this.selectedRow = [];
                this.refresh();
            }, error => {
                this.selectedRow = [];
                this.msgHandler.handleError(error);
            });
        }
    }

    // Delete the specified user
    deleteUsers(users: User[]): void {
        let userArr: string[] = [];
        if (this.onlySelf) {
            return;
        }

        if (users && users.length) {
            users.forEach(user => {
                userArr.push(user.username);
            });
        }
        // Confirm deletion
        let msg: ConfirmationMessage = new ConfirmationMessage(
            "USER.DELETION_TITLE",
            "USER.DELETION_SUMMARY",
            userArr.join(','),
            users,
            ConfirmationTargets.USER,
            ConfirmationButtons.DELETE_CANCEL
        );
        this.deletionDialogService.openComfirmDialog(msg);
    }

    delUser(users: User[]): void {
        let observableLists: any[] = [];
        if (users && users.length) {
            users.forEach(user => {
                observableLists.push(this.delOperate(user));
            });

            forkJoin(...observableLists).subscribe((item) => {
                this.selectedRow = [];
                this.currentTerm = '';
                this.refresh();
            }, error => {
                this.msgHandler.handleError(error);
            });
        }
    }

    delOperate(user: User): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_USER';
        operMessage.data.id = user.user_id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = user.username;
        this.operationService.publishInfo(operMessage);

        if (this.isMySelf(user.user_id)) {
            return this.translate.get('BATCH.DELETED_FAILURE').pipe(map(res => {
                operateChanges(operMessage, OperationState.failure, res);
            }));
        }

        return this.userService.deleteUser(user.user_id).pipe(map(() => {
            this.translate.get('BATCH.DELETED_SUCCESS').subscribe(res => {
                operateChanges(operMessage, OperationState.success);
            });
        }), catchError(error => {
            const message = errorHandler(error);
            this.translate.get(message).subscribe(res =>
                operateChanges(operMessage, OperationState.failure, res)
            );
            return observableThrowError(error);
        }));
    }

    // Refresh the user list
    refreshUser(): void {
        this.selectedRow = [];
        // Start to get
        this.currentTerm = '';
        this.onGoing = true;
        this.getUserListByPaging();
    }

    // Add new user
    addNewUser(): void {
        if (!this.canCreateUser) {
            return; // No response to this hacking action
        }
        this.newUserDialog.open();
    }

    // Add user to the user list
    addUserToList(user: User): void {
        // Currently we can only add it by reloading all
        this.refresh();
    }

    // Data loading
    load(state: any): void {
        this.selectedRow = [];
        this.onGoing = true;
        this.getUserListByPaging();
    }

    refresh(): void {
        this.currentPage = 1; // Refresh pagination
        this.refreshUser();
    }

    getUserListByPaging() {
        this.userService.getUserListByPaging(this.currentPage, this.pageSize, this.currentTerm)
            .subscribe(response => {
                // Get total count
                if (response.headers) {
                    let xHeader: string = response.headers.get("X-Total-Count");
                    if (xHeader) {
                        this.totalCount = parseInt(xHeader, 0);
                    }
                }
                this.users = response.body as User[];
                this.onGoing = false;
            }, error => {
                this.msgHandler.handleError(error);
                this.onGoing = false;
            });
    }
}
