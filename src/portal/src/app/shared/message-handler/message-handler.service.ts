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
import { Injectable } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { AlertType } from '../../shared/shared.const';
import { MessageService } from '../../global-message/message.service';
import { SessionService } from '../../shared/session.service';
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { errorHandler } from "../../../lib/utils/shared/shared.utils";
import { httpStatusCode } from "../../../lib/entities/shared.const";


@Injectable()
export class MessageHandlerService implements ErrorHandler {

    constructor(
        private msgService: MessageService,
        private translate: TranslateService,
        private session: SessionService) { }

    // Handle the error and map it to the suitable message
    // base on the status code of error.

    public handleError(error: any | string): void {
        if (!error) {
            return;
        }
        let msg = errorHandler(error);

        if (!(error.statusCode || error.status)) {
            this.msgService.announceMessage(500, msg, AlertType.DANGER);
        } else {
            let code = error.statusCode || error.status;
            if (code === httpStatusCode.Unauthorized) {
                this.msgService.announceAppLevelMessage(code, msg, AlertType.DANGER);
                // Session is invalid now, clare session cache
                this.session.clear();
            } else {
                this.msgService.announceMessage(code, msg, AlertType.DANGER);
            }
        }
    }
    public handleErrorPopupUnauthorized(error: any | string): void {

        if (!(error.statusCode || error.status)) {
            return;
        }
        let msg = errorHandler(error);
        let code = error.statusCode || error.status;
        if (code === httpStatusCode.Unauthorized) {
            this.msgService.announceAppLevelMessage(code, msg, AlertType.DANGER);
            // Session is invalid now, clare session cache
            this.session.clear();
        }
    }

    public handleReadOnly(): void {
        this.msgService.announceAppLevelMessage(503, 'REPO_READ_ONLY', AlertType.WARNING);
    }

    public showError(message: string, params: any): void {
        if (!params) {
            params = {};
        }
        this.translate.get(message, params).subscribe((res: string) => {
            this.msgService.announceMessage(500, res, AlertType.DANGER);
        });
    }

    public showSuccess(message: string): void {
        if (message && message.trim() !== "") {
            this.msgService.announceMessage(200, message, AlertType.SUCCESS);
        }
    }

    public showInfo(message: string): void {
        if (message && message.trim() !== "") {
            this.msgService.announceMessage(200, message, AlertType.INFO);
        }
    }

    public showWarning(message: string): void {
        if (message && message.trim() !== "") {
            this.msgService.announceMessage(400, message, AlertType.WARNING);
        }
    }

    public clear(): void {
        this.msgService.clear();
    }

    public isAppLevel(error: any): boolean {
        return error && error.statusCode === httpStatusCode.Unauthorized;
    }

    public error(error: any): void {
        this.handleError(error);
    }

    public warning(warning: any): void {
        this.showWarning(warning);
    }

    public info(info: any): void {
        this.showSuccess(info);
    }

    public log(log: any): void {
        this.showInfo(log);
    }
}
