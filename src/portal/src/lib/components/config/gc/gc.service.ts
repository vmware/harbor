import { Injectable } from '@angular/core';
import { Observable, Subscription, Subject, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { GcApiRepository } from './gc.api.repository';
import { ErrorHandler } from '../../../utils/error-handler';
import { GcJobData } from './gcLog';


@Injectable()
export class GcRepoService {

    constructor(
        private gcApiRepository: GcApiRepository,
        private errorHandler: ErrorHandler) {
    }

    public manualGc(shouldDeleteUntagged: boolean, isDryRun: boolean): Observable<any> {
        const param = {
            "schedule": {
                "type": "Manual"
            },
            parameters: {
                delete_untagged: shouldDeleteUntagged,
                dry_run: isDryRun
            }
        };
        return this.gcApiRepository.postSchedule(param);
    }

    public getJobs(): Observable <GcJobData []> {
        return this.gcApiRepository.getJobs();
    }

    public getLog(id): Observable <any> {
        return this.gcApiRepository.getLog(id);
    }

    public getSchedule(): Observable <any> {
        return this.gcApiRepository.getSchedule();
    }

    public postScheduleGc(shouldDeleteUntagged: boolean, type, cron): Observable <any> {
        let param = {
            "schedule": {
                "type": type,
                "cron": cron,
            },
            parameters: {
                delete_untagged: shouldDeleteUntagged
            }
        };

        return this.gcApiRepository.postSchedule(param);
    }

    public putScheduleGc(shouldDeleteUntagged, type, cron): Observable <any> {
        let param = {
            "schedule": {
                "type": type,
                "cron": cron,
            },
            parameters: {
                delete_untagged: shouldDeleteUntagged
            }
        };

        return this.gcApiRepository.putSchedule(param);
    }

    public getLogLink(id): string  {
        return this.gcApiRepository.getLogLink(id);
    }
}
