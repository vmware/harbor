import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ReplicationService } from "../../../services/replication.service";
import { TranslateService } from '@ngx-translate/core';
import { finalize } from "rxjs/operators";
import { Subscription, timer } from "rxjs";
import { ErrorHandler } from "../../../utils/error-handler/error-handler";
import { ReplicationJob, ReplicationTasks, Comparator, ReplicationJobItem, State } from "../../../services/interface";
import { CustomComparator, DEFAULT_PAGE_SIZE } from "../../../utils/utils";
import { RequestQueryParams } from "../../../services/RequestQueryParams";
import { REFRESH_TIME_DIFFERENCE } from '../../../entities/shared.const';

const executionStatus = 'InProgress';
const STATUS_MAP = {
  "Succeed": "Succeeded"
};
@Component({
  selector: 'replication-tasks',
  templateUrl: './replication-tasks.component.html',
  styleUrls: ['./replication-tasks.component.scss']
})
export class ReplicationTasksComponent implements OnInit, OnDestroy {
  isOpenFilterTag: boolean;
  inProgress: boolean = false;
  currentPage: number = 1;
  pageSize: number = DEFAULT_PAGE_SIZE;
  totalCount: number;
  loading = true;
  searchTask: string;
  defaultFilter = "resource_type";
  tasks: ReplicationTasks;
  taskItem: ReplicationTasks[] = [];
  tasksCopy: ReplicationTasks[] = [];
  stopOnGoing: boolean;
  executions: ReplicationJobItem[];
  timerDelay: Subscription;
  executionId: string;
  startTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
  ReplicationJob
  >("start_time", "date");
  endTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
    ReplicationJob
  >("end_time", "date");

  constructor(
    private translate: TranslateService,
    private router: Router,
    private replicationService: ReplicationService,
    private errorHandler: ErrorHandler,
    private route: ActivatedRoute,
  ) { }

  ngOnInit(): void {
    this.searchTask = '';
    this.executionId = this.route.snapshot.params['id'];
    const resolverData = this.route.snapshot.data;
    if (resolverData) {
      const replicationJob = <ReplicationJob>(resolverData["replicationTasksRoutingResolver"]);
      this.executions = replicationJob.data;
      this.clrLoadPage();
    }
  }
  getExecutionDetail(): void {
    this.inProgress = true;
    if (this.executionId) {
      this.replicationService.getExecutionById(this.executionId)
        .pipe(finalize(() => (this.inProgress = false)))
        .subscribe(res => {
          this.executions = res.data;
          this.clrLoadPage();
        },
        error => {
          this.errorHandler.error(error);
        });
    }
  }

  clrLoadPage(): void {
    if (!this.timerDelay) {
      this.timerDelay = timer(REFRESH_TIME_DIFFERENCE, REFRESH_TIME_DIFFERENCE).subscribe(() => {
        let count: number = 0;
        if (this.executions['status'] === executionStatus) {
          count++;
        }
        if (count > 0) {
          this.getExecutionDetail();
          let state: State = {
            page: {}
          };
          this.clrLoadTasks(state);
        } else {
          this.timerDelay.unsubscribe();
          this.timerDelay = null;
        }
      });
    }
  }

  public get trigger(): string {
    return this.executions && this.executions['trigger']
      ? this.executions['trigger']
      : "";
  }

  public get startTime(): Date {
    return this.executions && this.executions['start_time']
      ? this.executions['start_time']
      : null;
  }

  public get successNum(): string {
    return this.executions && this.executions['succeed'];
  }

  public get failedNum(): string {
    return this.executions && this.executions['failed'];
  }

  public get progressNum(): string {
    return this.executions && this.executions['in_progress'];
  }

  public get stoppedNum(): string {
    return this.executions && this.executions['stopped'];
  }

  stopJob() {
    this.stopOnGoing = true;
    this.replicationService.stopJobs(this.executionId)
    .subscribe(response => {
      this.stopOnGoing = false;
       this.getExecutionDetail();
       this.translate.get("REPLICATION.STOP_SUCCESS", { param: this.executionId }).subscribe((res: string) => {
          this.errorHandler.info(res);
       });
    },
    error => {
      this.errorHandler.error(error);
    });
  }

  viewLog(taskId: number | string): string {
    return this.replicationService.getJobBaseUrl() + "/executions/" + this.executionId + "/tasks/" + taskId + "/log";
  }

  ngOnDestroy() {
    if (this.timerDelay) {
      this.timerDelay.unsubscribe();
    }
  }

  clrLoadTasks(state: State): void {
      if (!state || !state.page || !this.executionId) {
        return;
      }

      let params: RequestQueryParams = new RequestQueryParams();
      params = params.set('page_size', this.pageSize + '').set('page', this.currentPage + '');
      if (this.searchTask && this.searchTask !== "") {
        params = params.set(this.defaultFilter, this.searchTask);
      }

      this.loading = true;
      this.replicationService.getReplicationTasks(this.executionId, params)
      .pipe(finalize(() => {
        this.loading = false;
      }))
      .subscribe(res => {
        if (res.headers) {
          let xHeader: string = res.headers.get("X-Total-Count");
          if (xHeader) {
            this.totalCount = parseInt(xHeader, 0);
          }
        }
        this.tasks = res.body; // Keep the data
      },
      error => {
        this.errorHandler.error(error);
      });
  }
  onBack(): void {
    this.router.navigate(["harbor", "replications"]);
  }

  selectFilter($event: any): void {
    this.defaultFilter = $event['target'].value;
    this.doSearch(this.searchTask);
  }

  // refresh icon
  refreshTasks(): void {
    this.currentPage = 1;
    let state: State = {
      page: {}
    };
    this.clrLoadTasks(state);
  }

  public doSearch(value: string): void {
    this.currentPage = 1;
    this.searchTask = value.trim();
    let state: State = {
      page: {}
    };
    this.clrLoadTasks(state);
  }

  openFilter(isOpen: boolean): void {
    this.isOpenFilterTag = isOpen;
  }

  getStatusStr(status: string): string {
    if (STATUS_MAP && STATUS_MAP[status]) {
      return STATUS_MAP[status];
    }
    return status;
  }

}
