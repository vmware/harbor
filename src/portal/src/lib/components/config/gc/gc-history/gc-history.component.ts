import { Component, OnInit, OnDestroy } from '@angular/core';
import { GcRepoService } from "../gc.service";
import { GcJobViewModel } from "../gcLog";
import { GcViewModelFactory } from "../gc.viewmodel.factory";
import { ErrorHandler } from "../../../../utils/error-handler";
import { Subscription, timer } from "rxjs";
import { REFRESH_TIME_DIFFERENCE } from '../../../../entities/shared.const';
const JOB_STATUS = {
  PENDING: "pending",
  RUNNING: "running"
};
const YES: string = 'Yes';
const NO: string = 'No';
@Component({
  selector: 'gc-history',
  templateUrl: './gc-history.component.html',
  styleUrls: ['./gc-history.component.scss']
})
export class GcHistoryComponent implements OnInit, OnDestroy {
  jobs: Array<GcJobViewModel> = [];
  loading: boolean;
  timerDelay: Subscription;
  constructor(
    private gcRepoService: GcRepoService,
    private gcViewModelFactory: GcViewModelFactory,
    private errorHandler: ErrorHandler
    ) {}

  ngOnInit() {
    this.getJobs();
  }

  getJobs() {
    this.loading = true;
    this.gcRepoService.getJobs().subscribe(jobs => {
      this.jobs = this.gcViewModelFactory.createJobViewModel(jobs);
      this.loading = false;
      // to avoid some jobs not finished.
      if (!this.timerDelay) {
        this.timerDelay = timer(REFRESH_TIME_DIFFERENCE, REFRESH_TIME_DIFFERENCE).subscribe(() => {
          let count: number = 0;
          this.jobs.forEach(job => {
            if (
              job['status'] === JOB_STATUS.PENDING ||
              job['status'] === JOB_STATUS.RUNNING
            ) {
              count++;
            }
          });
          if (count > 0) {
            this.getJobs();
          } else {
            this.timerDelay.unsubscribe();
            this.timerDelay = null;
          }
        });
      }
    }, error => {
        this.errorHandler.error(error);
        this.loading = false;
    });
  }

  isDryRun(param: string): string {
    if (param) {
      const paramObj: any = JSON.parse(param);
      if (paramObj && paramObj.dry_run) {
        return YES;
      }
    }
    return NO;
  }

  ngOnDestroy() {
    if (this.timerDelay) {
      this.timerDelay.unsubscribe();
    }
  }

  getLogLink(id): string {
    return this.gcRepoService.getLogLink(id);
  }

}
