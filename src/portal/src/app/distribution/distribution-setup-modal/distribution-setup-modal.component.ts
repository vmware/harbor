import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { Component, EventEmitter, Input, OnInit, Output, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { TranslateService } from '@ngx-translate/core';
import { errorHandler } from '../../../lib/utils/shared/shared.utils';
import { PreheatService } from "../../../../ng-swagger-gen/services/preheat.service";
import { Instance } from "../../../../ng-swagger-gen/models/instance";
import { AuthMode } from "../distribution-interface";
import { clone } from '../../../lib/utils/utils';
import { InlineAlertComponent } from "../../shared/inline-alert/inline-alert.component";
import { ClrLoadingState } from "@clr/angular";
import { Metadata } from '../../../../ng-swagger-gen/models/metadata';
import { operateChanges, OperateInfo, OperationState } from '../../../lib/components/operation/operate';
import { OperationService } from '../../../lib/components/operation/operation.service';

@Component({
  selector: 'dist-setup-modal',
  templateUrl: './distribution-setup-modal.component.html',
  styleUrls: ['./distribution-setup-modal.component.scss']
})
export class DistributionSetupModalComponent implements OnInit {
  @Input()
  providers: Metadata[] = [];
  model: Instance;
  originModelForEdit: Instance;
  opened: boolean = false;
  editingMode: boolean = false;
  authData: {[key: string]: any} = {};
  @ViewChild('instanceForm', { static: true }) instanceForm: NgForm;
  @ViewChild(InlineAlertComponent, { static: false }) inlineAlert: InlineAlertComponent;
  saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;

  @Output()
  refresh: EventEmitter<any> = new EventEmitter<any>();

  constructor(
    private distributionService: PreheatService,
    private msgHandler: MessageHandlerService,
    private translate: TranslateService,
    private operationService: OperationService
  ) {}

  ngOnInit() {
    this.reset();
  }

  public get isValid(): boolean {
    return this.instanceForm && this.instanceForm.valid;
  }

  get title(): string {
    return this.editingMode
      ? 'DISTRIBUTION.EDIT_INSTANCE'
      : 'DISTRIBUTION.SETUP_NEW_INSTANCE';
  }

  authModeChange() {
    if (this.editingMode && this.model.auth_mode === this.originModelForEdit.auth_mode) {
      this.authData = clone(this.originModelForEdit.auth_info);
    } else {
      switch (this.model.auth_mode) {
        case AuthMode.BASIC:
          this.authData = {
            password: '',
            username: ''
          };
          break;
        case AuthMode.OAUTH:
          this.authData = {
            token: ''
          };
          break;
        default:
          this.authData = null;
          break;
      }
    }
  }

  _open() {
    this.inlineAlert.close();
    this.opened = true;
  }

  _close() {
    this.opened = false;
    this.reset();
  }

  reset() {
    this.model = {
      name: '',
      endpoint: '',
      enabled: true,
      vendor: '',
      auth_mode: AuthMode.NONE,
      auth_info: this.authData
    };
    this.instanceForm.reset();
  }

  cancel() {
    this._close();
  }

  submit() {
    if (this.editingMode) {
      const operMessageForEdit = new OperateInfo();
      operMessageForEdit.name = 'DISTRIBUTION.UPDATE_INSTANCE';
      operMessageForEdit.data.id = this.model.id;
      operMessageForEdit.state = OperationState.progressing;
      operMessageForEdit.data.name = this.model.name;
      this.operationService.publishInfo(operMessageForEdit);
      this.saveBtnState = ClrLoadingState.LOADING;
      const instance: Instance = clone(this.originModelForEdit);
      instance.endpoint = this.model.endpoint;
      instance.enabled = this.model.enabled;
      instance.description = this.model.description;
      instance.auth_mode = this.model.auth_mode;
      instance.auth_info = this.model.auth_info;
      this.distributionService.UpdateInstance({preheatInstanceName: this.model.name, instance: instance
        }).subscribe(
        response => {
          this.translate.get('DISTRIBUTION.UPDATE_SUCCESS').subscribe(msg => {
            operateChanges(operMessageForEdit, OperationState.success);
            this.msgHandler.info(msg);
          });
          this.saveBtnState = ClrLoadingState.SUCCESS;
          this._close();
          this.refresh.emit();
        },
        err => {
          const message = errorHandler(err);
          this.translate.get('DISTRIBUTION.UPDATE_FAILED').subscribe(msg => {
            this.translate.get(message).subscribe(errMsg => {
              operateChanges(operMessageForEdit, OperationState.failure, msg);
              this.inlineAlert.showInlineError(msg + ': ' + errMsg);
              this.saveBtnState = ClrLoadingState.ERROR;
            });
          });
          this.msgHandler.handleErrorPopupUnauthorized(err);
        }
      );
    } else {
      const operMessage = new OperateInfo();
      operMessage.name = 'DISTRIBUTION.CREATE_INSTANCE';
      operMessage.state = OperationState.progressing;
      operMessage.data.name = this.model.name;
      this.operationService.publishInfo(operMessage);
      this.saveBtnState = ClrLoadingState.LOADING;
      if (this.model.auth_mode !== AuthMode.NONE) {
        this.model.auth_info = this.authData;
      } else {
        delete this.model.auth_info;
      }
      this.distributionService.CreateInstance({instance: this.model}).subscribe(
        response => {
          this.translate.get('DISTRIBUTION.CREATE_SUCCESS').subscribe(msg => {
            operateChanges(operMessage, OperationState.success);
            this.msgHandler.info(msg);
          });
          this.saveBtnState = ClrLoadingState.SUCCESS;
          this._close();
          this.refresh.emit();
        },
        err => {
          const message = errorHandler(err);
          this.translate.get('DISTRIBUTION.CREATE_FAILED').subscribe(msg => {
            this.translate.get(message).subscribe(errMsg => {
              operateChanges(operMessage, OperationState.failure, msg);
              this.inlineAlert.showInlineError(msg + ': ' + errMsg);
              this.saveBtnState = ClrLoadingState.ERROR;
            });
          });
          this.msgHandler.handleErrorPopupUnauthorized(err);
        }
      );
    }
  }

  openSetupModal(editingMode: boolean, data?: Instance): void {
    this.editingMode = editingMode;
    this._open();
    if (editingMode) {
      this.model = clone(data);
      this.originModelForEdit = clone(data);
      this.authData = this.model.auth_info || {};
    }
  }

  hasChangesForEdit(): boolean {
    if ( this.editingMode) {
      if ( this.model.description !== this.originModelForEdit.description) {
        return true;
      }
      if ( this.model.endpoint !== this.originModelForEdit.endpoint) {
        return true;
      }
      if (this.model.auth_mode !== this.originModelForEdit.auth_mode) {
        return true;
      } else {
        if (this.model.auth_mode === AuthMode.BASIC) {
          if (this.originModelForEdit.auth_info['username'] !== this.authData['username']) {
            return true;
          }
          if (this.originModelForEdit.auth_info['password'] !== this.authData['password']) {
            return true;
          }
        }
        if (this.model.auth_mode === AuthMode.OAUTH) {
          if (this.originModelForEdit.auth_info['token'] !== this.authData['token']) {
            return true;
          }
        }
        return false;
      }
    }
    return true;
  }
}
