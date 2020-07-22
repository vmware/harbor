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
import { Component, Output, EventEmitter } from "@angular/core";
import { TranslateService } from "@ngx-translate/core";

import { errorHandler } from "../../utils/shared/shared.utils";
import { Observable,  Subscription } from "rxjs";
import { ErrorHandler } from "../../../lib/utils/error-handler";

@Component({
  selector: "hbr-inline-alert",
  templateUrl: "./inline-alert.component.html",
  styleUrls: ["./inline-alert.component.scss"]
})
export class InlineAlertComponent {
  inlineAlertType: string = "danger";
  inlineAlertClosable: boolean = false;
  alertClose: boolean = true;
  displayedText: string = "";
  showCancelAction: boolean = false;
  useAppLevelStyle: boolean = false;
  timer: Subscription | null = null;
  count: number = 0;
  blinking: boolean = false;

  @Output() confirmEvt = new EventEmitter<boolean>();

  constructor(private translate: TranslateService,
    private errHandler: ErrorHandler,
    ) {}

  public get errorMessage(): string {
    return this.displayedText;
  }

  // Show error message inline
  public showInlineError(error: any): void {
    this.displayedText = errorHandler(error);
    if (this.displayedText) {
      this.translate
        .get(this.displayedText)
        .subscribe((res: string) => (this.displayedText = res));
    }
    this.errHandler.handleErrorPopupUnauthorized(error);

    this.inlineAlertType = "danger";
    this.showCancelAction = false;
    this.inlineAlertClosable = true;
    this.alertClose = false;
    this.useAppLevelStyle = false;
  }

  // Show confirmation info with action button
  public showInlineConfirmation(warning: any): void {
    this.displayedText = "";
    if (warning && warning.message) {
      this.translate
        .get(warning.message)
        .subscribe((res: string) => (this.displayedText = res));
    }
    this.inlineAlertType = "warning";
    this.showCancelAction = true;
    this.inlineAlertClosable = false;
    this.alertClose = false;
    this.useAppLevelStyle = false;
  }

  // Show inline sccess info
  public showInlineSuccess(info: any): void {
    this.displayedText = "";
    if (info && info.message) {
      this.translate
        .get(info.message)
        .subscribe((res: string) => (this.displayedText = res));
    }
    this.inlineAlertType = "success";
    this.showCancelAction = false;
    this.inlineAlertClosable = true;
    this.alertClose = false;
    this.useAppLevelStyle = false;
  }

  // Close alert
  public close(): void {
    this.alertClose = true;
  }

  public blink() {}

  confirmCancel(): void {
    this.confirmEvt.emit(true);
  }
}
