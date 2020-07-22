import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AppConfigService } from "../../services/app-config.service";
import { QUOTA_DANGER_COEFFICIENT, QUOTA_WARNING_COEFFICIENT, QuotaUnits } from "../../../lib/entities/shared.const";
import {
  Endpoint,
  EndpointService,
  ProjectService,
  UserPermissionService,
  USERSTATICPERMISSION
} from '../../../lib/services';
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { clone, GetIntegerAndUnit, getSuitableUnit as getSuitableUnitFn } from "../../../lib/utils/utils";
import { SessionService } from '../../shared/session.service';
import { Project } from '../project';

@Component({
  selector: 'summary',
  templateUrl: './summary.component.html',
  styleUrls: ['./summary.component.scss']
})
export class SummaryComponent implements OnInit {
  showProjectMemberInfo: boolean;
  showQuotaInfo: boolean;

  projectId: number;
  summaryInformation: any;
  quotaDangerCoefficient: number = QUOTA_DANGER_COEFFICIENT;
  quotaWarningCoefficient: number = QUOTA_WARNING_COEFFICIENT;
  endpoint: Endpoint;
  constructor(
    private projectService: ProjectService,
    private userPermissionService: UserPermissionService,
    private errorHandler: ErrorHandler,
    private appConfigService: AppConfigService,
    private route: ActivatedRoute,
  ) { }

  ngOnInit() {
    this.projectId = this.route.snapshot.parent.params['id'];
    const permissions = [
      {resource: USERSTATICPERMISSION.MEMBER.KEY, action: USERSTATICPERMISSION.MEMBER.VALUE.LIST},
      {resource: USERSTATICPERMISSION.QUOTA.KEY, action: USERSTATICPERMISSION.QUOTA.VALUE.READ},
    ];

    this.userPermissionService.hasProjectPermissions(this.projectId, permissions).subscribe((results: Array<boolean>) => {
      this.showProjectMemberInfo = results[0];
      this.showQuotaInfo = results[1];
    });

    this.projectService.getProjectSummary(this.projectId).subscribe(res => {
      this.summaryInformation = res;
    }, error => {
      this.errorHandler.error(error);
    });
  }
  getSuitableUnit(value) {
    const QuotaUnitsCopy = clone(QuotaUnits);
    return getSuitableUnitFn(value, QuotaUnitsCopy);
  }

  getIntegerAndUnit(hardValue, usedValue) {
    return GetIntegerAndUnit(hardValue, clone(QuotaUnits), usedValue, clone(QuotaUnits));
  }

  public get withHelmChart(): boolean {
    return this.appConfigService.getConfig().with_chartmuseum;
  }
}
