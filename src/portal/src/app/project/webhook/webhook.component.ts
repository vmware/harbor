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
import { finalize } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { Component, OnInit, ViewChild } from '@angular/core';
import { AddWebhookComponent } from './add-webhook/add-webhook.component';
import { AddWebhookFormComponent } from './add-webhook-form/add-webhook-form.component';
import { ActivatedRoute } from '@angular/router';
import { LastTrigger, Webhook } from './webhook';
import { WebhookService } from './webhook.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { Project } from '../project';
import { ConfirmationButtons, ConfirmationState, ConfirmationTargets } from '../../shared/shared.const';

import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';
import { ConfirmationDialogComponent } from '../../shared/confirmation-dialog/confirmation-dialog.component';
import { clone } from '../../../lib/utils/utils';
import { forkJoin, Observable } from 'rxjs';
import { UserPermissionService, USERSTATICPERMISSION } from '../../../lib/services';
import { ClrLoadingState } from '@clr/angular';

@Component({
  templateUrl: './webhook.component.html',
  styleUrls: ['./webhook.component.scss']
})
export class WebhookComponent implements OnInit {
  @ViewChild(AddWebhookComponent, { static: false } )
  addWebhookComponent: AddWebhookComponent;
  @ViewChild(AddWebhookFormComponent, { static: false })
  addWebhookFormComponent: AddWebhookFormComponent;
  @ViewChild("confirmationDialogComponent", { static: false })
  confirmationDialogComponent: ConfirmationDialogComponent;
  lastTriggers: LastTrigger[] = [];
  lastTriggerCount: number = 0;
  projectId: number;
  projectName: string;
  selectedRow: Webhook[] = [];
  webhookList: Webhook[] = [];
  metadata: any;
  loadingMetadata: boolean = false;
  loadingWebhookList: boolean = false;
  loadingTriggers: boolean = false;
  hasCreatPermission: boolean = false;
  hasUpdatePermission: boolean = false;
  addBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  constructor(
    private route: ActivatedRoute,
    private translate: TranslateService,
    private webhookService: WebhookService,
    private messageHandlerService: MessageHandlerService,
    private userPermissionService: UserPermissionService,) { }

  ngOnInit() {
    this.projectId = +this.route.snapshot.parent.params['id'];
    let resolverData = this.route.snapshot.parent.data;
    if (resolverData) {
      let project = <Project>(resolverData["projectResolver"]);
      this.projectName = project.name;
    }
    this.getData();
    this.getPermissions();
  }
  getPermissions() {
    const permissionsList: Observable<boolean>[] = [];
    permissionsList.push(this.userPermissionService.getPermission(this.projectId,
      USERSTATICPERMISSION.WEBHOOK.KEY, USERSTATICPERMISSION.WEBHOOK.VALUE.CREATE));
    permissionsList.push(this.userPermissionService.getPermission(this.projectId,
      USERSTATICPERMISSION.WEBHOOK.KEY, USERSTATICPERMISSION.WEBHOOK.VALUE.UPDATE));
    this.addBtnState = ClrLoadingState.LOADING;
    forkJoin(...permissionsList).subscribe(Rules => {
      [this.hasCreatPermission, this.hasUpdatePermission] = Rules;
      this.addBtnState = ClrLoadingState.SUCCESS;
    }, error => {
      this.messageHandlerService.error(error);
      this.addBtnState = ClrLoadingState.ERROR;
    });
  }

  getData() {
    this.getMetadata();
    this.getLastTriggers();
    this.getWebhooks();
    this.selectedRow = [];
  }
  getMetadata() {
    this.loadingMetadata = true;
    this.webhookService.getWebhookMetadata(this.projectId)
      .pipe(finalize(() => (this.loadingMetadata = false)))
      .subscribe(
        response => {
          this.metadata = response;
          if (this.metadata && this.metadata.event_type) {
            // sort by text
            this.metadata.event_type.sort((a: string, b: string) => {
              if (this.eventTypeToText(a) === this.eventTypeToText(b)) {
                return 0;
              }
              return this.eventTypeToText(a) > this.eventTypeToText(b) ? 1 : -1;
            });
          }
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  getLastTriggers() {
    this.loadingTriggers = true;
    this.webhookService
      .listLastTrigger(this.projectId)
      .pipe(finalize(() => (this.loadingTriggers = false)))
      .subscribe(
        response => {
          this.lastTriggers = response;
          this.lastTriggerCount = response.length;
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  getWebhooks() {
    this.loadingWebhookList = true;
    this.webhookService
      .listWebhook(this.projectId)
      .pipe(finalize(() => (this.loadingWebhookList = false)))
      .subscribe(
        response => {
          this.webhookList = response;
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  switchWebhookStatus() {
    let content = '';
    this.translate.get(
      !this.selectedRow[0].enabled
        ? 'WEBHOOK.ENABLED_WEBHOOK_SUMMARY'
        : 'WEBHOOK.DISABLED_WEBHOOK_SUMMARY'
    , {name: this.selectedRow[0].name}).subscribe((res) => {
      content = res;
      let message = new ConfirmationMessage(
        !this.selectedRow[0].enabled ? 'WEBHOOK.ENABLED_WEBHOOK_TITLE' : 'WEBHOOK.DISABLED_WEBHOOK_TITLE',
        content,
        '',
        {},
        ConfirmationTargets.WEBHOOK,
        !this.selectedRow[0].enabled ? ConfirmationButtons.ENABLE_CANCEL : ConfirmationButtons.DISABLE_CANCEL
      );
      this.confirmationDialogComponent.open(message);
    });
  }

  confirmSwitch(message) {
    if (message &&
      message.source === ConfirmationTargets.WEBHOOK &&
      message.state === ConfirmationState.CONFIRMED) {
      if (JSON.stringify(message.data) === '{}') {
        this.webhookService
          .editWebhook(this.projectId, this.selectedRow[0].id,
            Object.assign({}, this.selectedRow[0], { enabled: !this.selectedRow[0].enabled }))
          .subscribe(
            response => {
              this.getData();
            },
            error => {
              this.messageHandlerService.handleError(error);
            }
          );
      } else {
        const observableLists: Observable<any>[] = [];
        message.data.forEach(item => {
          observableLists.push(this.webhookService.deleteWebhook(this.projectId, item.id));
        });
        forkJoin(...observableLists).subscribe(
          response => {
            this.getData();
          },
          error => {
            this.messageHandlerService.handleError(error);
          }
        );
      }
    }
  }

  editWebhook() {
    if (this.metadata) {
      this.addWebhookComponent.isOpen = true;
      this.addWebhookComponent.isEdit = true;
      this.addWebhookComponent.addWebhookFormComponent.isModify = true;
      this.addWebhookComponent.addWebhookFormComponent.webhook = clone(this.selectedRow[0]);
      this.addWebhookComponent.addWebhookFormComponent.originValue = clone(this.selectedRow[0]);
      this.addWebhookComponent.addWebhookFormComponent.webhook.event_types = clone(this.selectedRow[0].event_types);
    }
  }

  openAddWebhookModal(): void {
    this.addWebhookComponent.openAddWebhookModal();
  }
  newWebhook() {
    if (this.metadata) {
      this.addWebhookComponent.isOpen = true;
      this.addWebhookComponent.isEdit = false;
      this.addWebhookComponent.addWebhookFormComponent.isModify = false;
      this.addWebhookComponent.addWebhookFormComponent.currentForm.reset({notifyType: this.metadata.notify_type[0]});
      this.addWebhookComponent.addWebhookFormComponent.webhook = new Webhook();
      this.addWebhookComponent.addWebhookFormComponent.webhook.event_types = clone(this.metadata.event_type);
    }
  }
  success() {
   this.getData();
  }

  deleteWebhook() {
    const names: string[] = [];
    this.selectedRow.forEach(item => {
      names.push(item.name);
    });
    let content = '';
    this.translate.get(
         'WEBHOOK.DELETE_WEBHOOK_SUMMARY'
      , {names:  names.join(',')}).subscribe((res) => content = res);
    const msg: ConfirmationMessage = new ConfirmationMessage(
      "SCANNER.CONFIRM_DELETION",
      content,
      names.join(','),
      this.selectedRow,
      ConfirmationTargets.WEBHOOK,
      ConfirmationButtons.DELETE_CANCEL
    );
    this.confirmationDialogComponent.open(msg);
  }
  eventTypeToText(eventType: string): string {
    return this.webhookService.eventTypeToText(eventType);
  }
}
