export class GcJobData {
    id: number;
    job_name: string;
    job_kind: string;
    schedule: Schedule;
    job_status: string;
    job_parameters: string;
    job_uuid: string;
    creation_time: string;
    update_time: string;
    delete: boolean;
}

export class Schedule {
    type: string;
    cron: string;
}
export class GcJobViewModel {
    id: number;
    type: string;
    status: string;
    parameters: string;
    createTime: Date;
    updateTime: Date;
    details: string;
}


