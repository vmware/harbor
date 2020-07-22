import { Injectable } from '@angular/core';
import { GcJobData, GcJobViewModel } from './gcLog';

@Injectable()
export class GcViewModelFactory {
    public createJobViewModel(jobs: GcJobData[]): GcJobViewModel[] {
        let gcViewModels: GcJobViewModel[] = [];
        for (let job of jobs) {

            let createTime = new Date(job.creation_time);
            let updateTime = new Date(job.update_time);
            gcViewModels.push({
                id: job.id,
                type: job.schedule ? job.schedule.type : null,
                status: job.job_status,
                parameters: job.job_parameters,
                createTime: createTime,
                updateTime: updateTime,
                details: null
            });
        }
        return gcViewModels;
    }
}
