import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from '../../utils/shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ReplicationComponent } from './replication.component';
import { ListReplicationRuleComponent } from '../list-replication-rule/list-replication-rule.component';
import { CreateEditRuleComponent } from '../create-edit-rule/create-edit-rule.component';
import { CronScheduleComponent } from '../cron-schedule/cron-schedule.component';
import { DatePickerComponent } from '../datetime-picker/datetime-picker.component';
import { FilterComponent } from '../filter/filter.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import {ReplicationRule, ReplicationJob, Endpoint} from '../../services/interface';
import { CronTooltipComponent } from "../cron-schedule/cron-tooltip/cron-tooltip.component";

import { ErrorHandler } from '../../utils/error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../../entities/service.config';
import { ReplicationService, ReplicationDefaultService } from '../../services/replication.service';
import { EndpointService, EndpointDefaultService } from '../../services/endpoint.service';
import { JobLogService, JobLogDefaultService, ReplicationJobItem } from '../../services';
import {ProjectDefaultService, ProjectService} from "../../services/project.service";
import {OperationService} from "../operation/operation.service";
import {FilterLabelComponent} from "../create-edit-rule/filter-label.component";
import {LabelPieceComponent} from "../label-piece/label-piece.component";
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { CURRENT_BASE_HREF } from "../../utils/utils";


describe('Replication Component (inline template)', () => {

  let mockRules: ReplicationRule[] = [
      {
          "id": 1,
          "name": "sync_01",
          "description": "",
          "filters": null,
          "trigger": {"type": "Manual", "trigger_settings": null},
          "error_job_count": 2,
          "deletion": false,
          "src_registry": {id: 3},
          "src_namespaces": ["name1"],
          "enabled": true,
          "override": true
      },
      {
          "id": 2,
          "name": "sync_02",
          "description": "",
          "filters": null,
          "trigger": {"type": "Manual", "trigger_settings": null},
          "error_job_count": 2,
          "deletion": false,
          "dest_registry": {id: 5},
          "src_namespaces": ["name1"],
          "enabled": true,
          "override": true
      }
  ];

  let mockJobs: ReplicationJobItem[] = [
    {
      id: 1,
      status: "stopped",
      policy_id: 1,
      trigger: "Manual",
      total: 0,
      failed: 0,
      succeed: 0,
      in_progress: 0,
      stopped: 0
    },
    {
      id: 2,
      status: "stopped",
      policy_id: 1,
      trigger: "Manual",
      total: 1,
      failed: 0,
      succeed: 1,
      in_progress: 0,
      stopped: 0
    },
    {
      id: 3,
      status: "stopped",
      policy_id: 2,
      trigger: "Manual",
      total: 1,
      failed: 1,
      succeed: 0,
      in_progress: 0,
      stopped: 0
    }
  ];

    let mockEndpoints: Endpoint[] = [
        {
            "id": 1,
            "credential": {
              "access_key": "admin",
              "access_secret": "",
              "type": "basic"
            },
            "description": "test",
            "insecure": false,
            "name": "target_01",
            "type": "Harbor",
            "url": "https://10.117.4.151"
        },
        {
            "id": 2,
            "credential": {
              "access_key": "admin",
              "access_secret": "",
              "type": "basic"
            },
            "description": "test",
            "insecure": false,
            "name": "target_02",
            "type": "Harbor",
            "url": "https://10.117.5.142"
        },
    ];

  let mockJob: ReplicationJob = {
    metadata: {xTotalCount: 3},
    data: mockJobs
  };

  let fixture: ComponentFixture<ReplicationComponent>;
  let fixtureCreate: ComponentFixture<CreateEditRuleComponent>;
  let comp: ReplicationComponent;
  let compCreate: CreateEditRuleComponent;

  let replicationService: ReplicationService;
  let endpointService: EndpointService;

  let spyRules: jasmine.Spy;
  let spyJobs: jasmine.Spy;
  let spyEndpoints: jasmine.Spy;

  let deGrids: DebugElement[];
  let deRules: DebugElement;
  let deJobs: DebugElement;

  let elRule: HTMLElement;
  let elJob: HTMLElement;

  let config: IServiceConfig = {
    replicationRuleEndpoint: CURRENT_BASE_HREF + '/policies/replication/testing'
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      schemas: [ CUSTOM_ELEMENTS_SCHEMA ],
      imports: [
        SharedModule,
        NoopAnimationsModule,
        RouterTestingModule
      ],
      declarations: [
        ReplicationComponent,
        ListReplicationRuleComponent,
        CreateEditRuleComponent,
        CronTooltipComponent,
        CronScheduleComponent,
        ConfirmationDialogComponent,
        DatePickerComponent,
        FilterComponent,
        InlineAlertComponent,
        FilterLabelComponent,
        LabelPieceComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ReplicationService, useClass: ReplicationDefaultService },
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: JobLogService, useClass: JobLogDefaultService },
        { provide: OperationService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ReplicationComponent);
    fixtureCreate = TestBed.createComponent(CreateEditRuleComponent);
    comp = fixture.componentInstance;
    compCreate = fixtureCreate.componentInstance;
    comp.projectId = 1;
    comp.search.ruleId = 1;

    replicationService = fixture.debugElement.injector.get(ReplicationService);

    endpointService = fixtureCreate.debugElement.injector.get(EndpointService);

    spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(of(mockRules));
    spyJobs = spyOn(replicationService, 'getExecutions').and.returnValues(of(mockJob));


    spyEndpoints = spyOn(endpointService, 'getEndpoints').and.returnValues(of(mockEndpoints));
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      deGrids = fixture.debugElement.queryAll(del => del.classes['datagrid']);
      fixture.detectChanges();
      expect(deGrids).toBeTruthy();
      expect(deGrids.length).toEqual(2);
    });
  });


  it('Should load replication rules', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      deRules = deGrids[0].query(By.css('datagrid-cell'));
      expect(deRules).toBeTruthy();
      fixture.detectChanges();
      elRule = deRules.nativeElement;
      expect(elRule).toBeTruthy();
      expect(elRule.textContent).toEqual('sync_01');
    });
  }));

  it('Should load replication jobs', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      deJobs = deGrids[1].query(By.css('datagrid-cell'));
      expect(deJobs).toBeTruthy();
      fixture.detectChanges();
      elJob = deJobs.nativeElement;
      fixture.detectChanges();
      expect(elJob).toBeTruthy();
      expect(elJob.textContent).toEqual('library/nginx');
    });
  }));

  it('Should filter replication rules by keywords', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doSearchRules('sync_01');
      fixture.detectChanges();
      let el: HTMLElement = deRules.nativeElement;
      fixture.detectChanges();
      expect(el.textContent.trim()).toEqual('sync_01');
    });
  }));

  it('Should filter replication jobs by keywords', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doSearchJobs('nginx');
      fixture.detectChanges();
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/nginx');
    });
  }));

  it('Should filter replication jobs by status', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/mysql');
    });
  }));

  it('Should filter replication jobs by date range', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/nginx');
    });
  }));

  it('function "getDuration" should work', () => {
    // ms level
    const item: ReplicationJobItem = {
      start_time: 1589340503637,
      end_time: 1589340503638,
      id: 3,
      status: "stopped",
      policy_id: 2,
      trigger: "Manual",
      total: 1,
      failed: 1,
      succeed: 0,
      in_progress: 0,
      stopped: 0
    };
    expect(comp.getDuration(item)).toEqual('1ms');
    // sec level
    item.start_time = 1589340503637;
    item.end_time = 1589340504638;
    expect(comp.getDuration(item)).toEqual('1s');
    // min level
    item.start_time = 1589340503637;
    item.end_time = 1589340564638;
    expect(comp.getDuration(item)).toEqual('1m1s');
    // hour level
    item.start_time = 1589340503637;
    item.end_time = 1589344164638;
    expect(comp.getDuration(item)).toEqual('61m1s');
    // day level
    item.start_time = "5/8/20,11:20 AM";
    item.end_time = "5/9/20,11:24 AM";
    expect(comp.getDuration(item)).toEqual('1444m');
  });
});
