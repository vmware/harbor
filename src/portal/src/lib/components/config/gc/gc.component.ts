import {
  Component,
  Input,
  Output,
  EventEmitter,
  ViewChild,
  OnInit
} from "@angular/core";
import { TranslateService } from "@ngx-translate/core";
import { GcJobViewModel } from "./gcLog";
import { GcViewModelFactory } from "./gc.viewmodel.factory";
import { GcRepoService } from "./gc.service";
import {
  SCHEDULE_TYPE_NONE,
  ONE_MINITUE,
  THREE_SECONDS, GCSchedule
} from "./gc.const";
import { ErrorHandler } from "../../../utils/error-handler";
import { CronScheduleComponent } from "../../cron-schedule/cron-schedule.component";
import { OriginCron } from '../../../services/interface';
import { finalize } from "rxjs/operators";
@Component({
  selector: "gc-config",
  templateUrl: "./gc.component.html",
  styleUrls: ["./gc.component.scss"]
})
export class GcComponent implements OnInit {
  jobs: Array<GcJobViewModel> = [];
  schedule: GCSchedule = {};
  originCron: OriginCron;
  disableGC: boolean = false;
  getLabelCurrent = 'GC.CURRENT_SCHEDULE';
  @Output() loadingGcStatus = new EventEmitter<boolean>();
  @ViewChild(CronScheduleComponent, {static: false})
  CronScheduleComponent: CronScheduleComponent;
  shouldDeleteUntagged: boolean;
  dryRunOnGoing: boolean = false;
  constructor(
    private gcRepoService: GcRepoService,
    private gcViewModelFactory: GcViewModelFactory,
    private errorHandler: ErrorHandler,
    private translate: TranslateService
  ) {
    translate.setDefaultLang("en-us");
  }

  ngOnInit() {
    this.getCurrentSchedule();
    this.getJobs();
  }

  getCurrentSchedule() {
    this.loadingGcStatus.emit(true);
    this.gcRepoService.getSchedule()
    .pipe(finalize(() => {
      this.loadingGcStatus.emit(false);
    }))
    .subscribe(schedule => {
      this.initSchedule(schedule);
    }, error => {
      this.errorHandler.error(error);
    });
  }

  public initSchedule(schedule: GCSchedule) {
    if (schedule && schedule.schedule !== null) {
      this.schedule = schedule;
      this.originCron = this.schedule.schedule;
    } else {
      this.originCron = {
        type: SCHEDULE_TYPE_NONE,
        cron: ''
      };
    }
    if (schedule && schedule.job_parameters) {
      this.shouldDeleteUntagged = JSON.parse(schedule.job_parameters).delete_untagged;
    } else {
      this.shouldDeleteUntagged = false;
    }
  }

  getJobs() {
    this.gcRepoService.getJobs().subscribe(jobs => {
      this.jobs = this.gcViewModelFactory.createJobViewModel(jobs);
    });
  }

  gcNow(): void {
    this.disableGC = true;
    setTimeout(() => {
      this.enableGc();
    }, ONE_MINITUE);

    this.gcRepoService.manualGc(this.shouldDeleteUntagged, false).subscribe(
      response => {
        this.translate.get("GC.MSG_SUCCESS").subscribe((res: string) => {
          this.errorHandler.info(res);
        });
      },
      error => {
        this.errorHandler.error(error);
      }
    );
  }

  dryRun() {
    this.dryRunOnGoing = true;
    this.gcRepoService.manualGc(this.shouldDeleteUntagged, true)
      .pipe(finalize(() => this.dryRunOnGoing = false))
      .subscribe(
      response => {
        this.translate.get("GC.DRY_RUN_SUCCESS").subscribe((res: string) => {
          this.errorHandler.info(res);
        });
      },
      error => {
        this.errorHandler.error(error);
      }
    );
  }

  private enableGc() {
    this.disableGC = false;
  }

  private resetSchedule(cron) {
    this.schedule = {
      schedule: {
        type: this.CronScheduleComponent.scheduleType,
        cron: cron
      }
    };
    if (!cron) {
      this.shouldDeleteUntagged = false;
    }
    this.getJobs();
  }

  scheduleGc(cron: string) {
    let schedule = this.schedule;
    if (schedule && schedule.schedule && schedule.schedule.type !== SCHEDULE_TYPE_NONE) {
      this.gcRepoService.putScheduleGc(this.shouldDeleteUntagged, this.CronScheduleComponent.scheduleType, cron).subscribe(
        response => {
          this.translate
            .get("GC.MSG_SCHEDULE_RESET")
            .subscribe((res) => {
              this.errorHandler.info(res);
              this.CronScheduleComponent.resetSchedule();
            });
          this.resetSchedule(cron);
        },
        error => {
          this.errorHandler.error(error);
        }
      );
    } else {
      this.gcRepoService.postScheduleGc(this.shouldDeleteUntagged, this.CronScheduleComponent.scheduleType, cron).subscribe(
        response => {
          this.translate.get("GC.MSG_SCHEDULE_SET").subscribe((res) => {
            this.errorHandler.info(res);
            this.CronScheduleComponent.resetSchedule();
          });
          this.resetSchedule(cron);
        },
        error => {
          this.errorHandler.error(error);
        }
      );
    }
  }
}
