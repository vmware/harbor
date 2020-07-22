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
  Output,
  EventEmitter,
  ViewChild,
  AfterViewChecked,
  ChangeDetectorRef,
  OnDestroy,
  OnInit
} from "@angular/core";
import { NgForm } from "@angular/forms";
import { Subscription, throwError as observableThrowError } from "rxjs";
import { TranslateService } from "@ngx-translate/core";

import { EndpointService } from "../../services/endpoint.service";
import { ErrorHandler } from "../../utils/error-handler";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { Endpoint, PingEndpoint } from "../../services/interface";
import { clone, compareValue, CURRENT_BASE_HREF, isEmptyObject } from "../../utils/utils";
import { HttpClient } from "@angular/common/http";
import { catchError } from "rxjs/operators";
import { AppConfigService } from '../../../app/services/app-config.service';

const FAKE_PASSWORD = "rjGcfuRu";
const FAKE_JSON_KEY = "No Change";
const METADATA_URL = CURRENT_BASE_HREF + "/replication/adapterinfos";
const HELM_HUB = "helm-hub";
@Component({
  selector: "hbr-create-edit-endpoint",
  templateUrl: "./create-edit-endpoint.component.html",
  styleUrls: ["./create-edit-endpoint.component.scss"]
})
export class CreateEditEndpointComponent
  implements AfterViewChecked, OnDestroy, OnInit {
  modalTitle: string;
  urlDisabled: boolean = false;
  editDisabled: boolean = false;
  controlEnabled: boolean = false;
  createEditDestinationOpened: boolean;
  staticBackdrop: boolean = true;
  closable: boolean = false;
  editable: boolean;
  adapterList: string[];
  endpointList: any[] = [];
  target: Endpoint = this.initEndpoint();
  selectedType: string;
  initVal: Endpoint;
  targetForm: NgForm;
  @ViewChild("targetForm", {static: false}) currentForm: NgForm;
  targetEndpoint;
  testOngoing: boolean;
  onGoing: boolean;
  endpointId: number | string;

  @ViewChild(InlineAlertComponent, {static: false}) inlineAlert: InlineAlertComponent;

  @Output() reload = new EventEmitter<boolean>();

  timerHandler: any;
  valueChangesSub: Subscription;
  formValues: { [key: string]: string } | any;
  adapterInfo: object;
  showEndpointList: boolean = false;
  endpointOnHover: boolean = false;
  constructor(
    private endpointService: EndpointService,
    private errorHandler: ErrorHandler,
    private translateService: TranslateService,
    private ref: ChangeDetectorRef,
    private http: HttpClient,
    private appConfigService: AppConfigService,
  ) {}

  ngOnInit(): void {
    this.getAdapters();
    this.getAdapterInfo();
  }
  getAdapters() {
    this.endpointService.getAdapters().subscribe(
      adapters => {
        this.adapterList = adapters || [];
        if (!this.appConfigService.getConfig().with_chartmuseum) { // disable helm-hub
          for (let i = 0; i < this.adapterList.length; i++) {
            if (this.adapterList[i] === HELM_HUB) {
              this.adapterList.splice(i, 1);
            }
          }
        }
      },
      error => {
        this.errorHandler.error(error);
      }
    );
  }
  getAdapterInfo() {
    this.http.get(METADATA_URL)
        .pipe(catchError(error => observableThrowError(error)))
        .subscribe(
            response => {
                this.adapterInfo = response;
            }, error => {
                this.errorHandler.error(error);
            });
  }
  isNormalCredential(): boolean {
    return !(this.adapterInfo && this.target && this.target.type
        && this.adapterInfo[this.target.type]
        && this.adapterInfo[this.target.type].credential_pattern);
  }
  selectedEndpoint(endpoint: string) {
    this.targetForm.controls.endpointUrl.reset(endpoint);
    this.showEndpointList = false;
    this.endpointOnHover = false;
  }
  blur() {
    if (!this.endpointOnHover) {
      this.showEndpointList = false;
    }
  }
  public get isValid(): boolean {
    return (
      !this.testOngoing &&
      !this.onGoing &&
      this.targetForm &&
      this.targetForm.valid &&
      this.editable &&
      !compareValue(this.target, this.initVal)
    );
  }

  public get inProgress(): boolean {
    return this.onGoing || this.testOngoing;
  }

  setInsecureValue($event: any) {
    this.target.insecure = !$event;
  }

  ngOnDestroy(): void {
    if (this.valueChangesSub) {
      this.valueChangesSub.unsubscribe();
    }
  }

  initEndpoint(): Endpoint {
    return {
      credential: {
        access_key: "",
        access_secret: "",
        type: "basic"
      },
      description: "",
      insecure: false,
      name: "",
      type: "harbor",
      url: ""
    };
  }

  initPingEndpoint(): PingEndpoint {
    return {
      access_key: "",
      access_secret: "",
      description: "",
      insecure: false,
      name: "",
      type: "harbor",
      url: ""
    };
  }

  open(): void {
    this.createEditDestinationOpened = true;
  }

  close(): void {
    this.createEditDestinationOpened = false;
  }

  reset(): void {
    // Reset status variables
    this.testOngoing = false;
    this.onGoing = false;

    // Reset data
    this.target = this.initEndpoint();
    this.initVal = this.initEndpoint();
    this.formValues = null;
    this.endpointId = "";
    this.inlineAlert.close();
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

  openCreateEditTarget(editable: boolean, targetId?: number | string) {
    this.editable = editable;
    // reset
    this.reset();
    if (targetId) {
      this.endpointId = targetId;
      this.translateService
        .get("DESTINATION.TITLE_EDIT")
        .subscribe(res => (this.modalTitle = res));
      this.endpointService.getEndpoint(targetId).subscribe(
        target => {
          this.target = target;
          this.urlDisabled = this.target.type === 'docker-hub';
          // Keep data cache
          this.initVal = clone(target);
          this.initVal.credential.access_secret = this.target.type === 'google-gcr' ? FAKE_JSON_KEY : FAKE_PASSWORD;
          this.target.credential.access_secret = this.target.type === 'google-gcr' ? FAKE_JSON_KEY : FAKE_PASSWORD;

          // Open the modal now
          this.open();
          this.editDisabled = true;
          this.forceRefreshView(2000);
        },
        error => this.errorHandler.error(error)
      );
    } else {
      this.urlDisabled = false;
      this.endpointId = "";
      this.translateService
        .get("DESTINATION.TITLE_ADD")
        .subscribe(res => (this.modalTitle = res));
      // Directly open the modal
      this.open();
      this.editDisabled = false;
    }
  }

  adapterChange($event): void {
    this.targetForm.controls.endpointUrl.reset("");
    let selectValue = this.targetForm.controls.adapter.value;
    this.urlDisabled = false;
    if (this.isNormalCredential()) {
      this.targetForm.controls.access_key.setValue("");
    } else {
      this.targetForm.controls.access_key.setValue(this.adapterInfo[this.target.type].credential_pattern.access_key_data);
    }
    if (this.adapterInfo && this.adapterInfo[selectValue]
        && this.adapterInfo[selectValue].endpoint_pattern
        && this.adapterInfo[selectValue].endpoint_pattern.endpoints) {
      this.endpointList = this.adapterInfo[selectValue].endpoint_pattern.endpoints;
      if (this.endpointList.length === 1) {
        this.target.url = this.endpointList[0].value;
      }
    } else {
      this.endpointList = [];
    }
  }

  testConnection() {
    let payload: PingEndpoint = this.initPingEndpoint();
    if (!this.endpointId) {
      payload.name = this.target.name;
      payload.description = this.target.description;
      payload.type = this.target.type;
      payload.url = this.target.url;
      payload.access_key = this.target.credential.access_key;
      payload.access_secret = this.target.credential.access_secret;
      payload.insecure = this.target.insecure;
    } else {
      let changes: { [key: string]: any } = this.getChanges();
      for (let prop of Object.keys(payload)) {
        delete payload[prop];
      }
      payload.id = this.target.id;
      if (!isEmptyObject(changes)) {
        let changekeys: { [key: string]: any } = Object.keys(this.getChanges());
        changekeys.forEach((key: string) => {
          payload[key] = changes[key];
        });
      }
    }

    this.testOngoing = true;
    this.endpointService.pingEndpoint(payload).subscribe(
      response => {
        this.inlineAlert.showInlineSuccess({
          message: "DESTINATION.TEST_CONNECTION_SUCCESS"
        });
        this.forceRefreshView(2000);
        this.testOngoing = false;
      },
      error => {
        this.inlineAlert.showInlineError("DESTINATION.TEST_CONNECTION_FAILURE");
        this.forceRefreshView(2000);
        this.testOngoing = false;
      }
    );
  }

  onSubmit() {
    if (this.endpointId) {
      this.updateEndpoint();
    } else {
      this.addEndpoint();
    }
  }

  addEndpoint() {
    if (this.onGoing) {
      return; // Avoid duplicated submitting
    }
    this.onGoing = true;
    this.endpointService.createEndpoint(this.target).subscribe(
      response => {
        this.translateService
          .get("DESTINATION.CREATED_SUCCESS")
          .subscribe(res => this.errorHandler.info(res));
        this.reload.emit(true);
        this.onGoing = false;
        this.close();
        this.forceRefreshView(2000);
      },
      error => {
        this.onGoing = false;
        this.inlineAlert.showInlineError(error);
        this.forceRefreshView(2000);
      }
    );
  }

  updateEndpoint() {
    if (this.onGoing) {
      return; // Avoid duplicated submitting
    }

    let payload: Endpoint = this.initEndpoint();
    for (let prop of Object.keys(payload)) {
      delete payload[prop];
    }
    let changes: { [key: string]: any } = this.getChanges();
    if (isEmptyObject(changes)) {
      return;
    }

    let changekeys: { [key: string]: any } = Object.keys(changes);

    changekeys.forEach((key: string) => {
      payload[key] = changes[key];
    });

    if (!this.target.id) {
      return;
    }

    this.onGoing = true;
    this.endpointService.updateEndpoint(this.target.id, payload).subscribe(
      response => {
        this.translateService
          .get("DESTINATION.UPDATED_SUCCESS")
          .subscribe(res => this.errorHandler.info(res));
        this.reload.emit(true);
        this.close();
        this.onGoing = false;
        this.forceRefreshView(2000);
      },
      error => {
        this.inlineAlert.showInlineError(error);
        this.onGoing = false;
        this.forceRefreshView(2000);
      }
    );
  }

  onCancel() {
    let changes: { [key: string]: any } = this.getChanges();
    if (!isEmptyObject(changes)) {
      this.inlineAlert.showInlineConfirmation({
        message: "ALERT.FORM_CHANGE_CONFIRMATION"
      });
    } else {
      this.close();
      if (this.targetForm) {
        this.targetForm.reset();
      }
    }
  }

  confirmCancel(confirmed: boolean) {
    this.inlineAlert.close();
    this.close();
  }

  ngAfterViewChecked(): void {
    if (this.targetForm !== this.currentForm) {
      this.targetForm = this.currentForm;
      if (this.targetForm) {
        this.valueChangesSub = this.targetForm.valueChanges.subscribe(
          (data: { [key: string]: string } | any) => {
            if (data) {
              // To avoid invalid change publish events
              let keyNumber: number = 0;
              for (let key in data) {
                // Empty string "" is accepted
                if (data[key] !== null) {
                  keyNumber++;
                }
              }
              if (keyNumber !== 5) {
                return;
              }

              if (!compareValue(this.formValues, data)) {
                this.formValues = data;
              }
            }
          }
        );
      }
    }
  }
  getChanges(): { [key: string]: any | any[] } {
    let changes: { [key: string]: any | any[] } = {};
    if (!this.target || !this.initVal) {
      return changes;
    }
    for (let prop of Object.keys(this.target)) {
      let field: any = this.initVal[prop];
      if (typeof field !== "object") {
        if (!compareValue(field, this.target[prop])) {
          changes[prop] = this.target[prop];
          // Number
          if (typeof field === "number") {
            changes[prop] = +changes[prop];
          }

          // Trim string value
          if (typeof field === "string") {
            changes[prop] = ("" + changes[prop]).trim();
          }
        }
      } else {
        for (let pro of Object.keys(field)) {
          if (!compareValue(field[pro], this.target[prop][pro])) {
            changes[pro] = this.target[prop][pro];
            // Number
            if (typeof field[pro] === "number") {
              changes[pro] = +changes[pro];
            }

            // Trim string value
            if (typeof field[pro] === "string") {
              changes[pro] = ("" + changes[pro]).trim();
            }
          }
        }
      }
    }
    return changes;
  }

  getAdapterText(adapter: string): string {
    return this.endpointService.getAdapterText(adapter);
  }
}
