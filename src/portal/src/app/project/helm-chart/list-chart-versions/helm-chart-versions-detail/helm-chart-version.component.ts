import {
  Component,
  Input,
  OnInit,
  ViewChild,
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Output,
  EventEmitter
} from "@angular/core";
import { Observable, forkJoin, throwError as observableThrowError } from "rxjs";
import { finalize, map, catchError } from "rxjs/operators";
import { TranslateService } from "@ngx-translate/core";
import { HelmChartVersion, HelmChartMaintainer } from "../../helm-chart.interface.service";
import { HelmChartService } from "../../helm-chart.service";
import { ConfirmationAcknowledgement } from "../../../../shared/confirmation-dialog/confirmation-state-message";
import { ConfirmationDialogComponent } from "../../../../shared/confirmation-dialog/confirmation-dialog.component";
import { ConfirmationMessage } from "../../../../shared/confirmation-dialog/confirmation-message";
import {
  ConfirmationButtons,
  ConfirmationTargets,
  ConfirmationState,
  DefaultHelmIcon,
  ResourceType,
} from "../../../../shared/shared.const";
import {
  Label,
  LabelService,
  State,
  SystemInfo,
  SystemInfoService,
  UserPermissionService, USERSTATICPERMISSION
} from "../../../../../lib/services";
import { DEFAULT_PAGE_SIZE, downloadFile } from "../../../../../lib/utils/utils";
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { OperationService } from "../../../../../lib/components/operation/operation.service";
import { operateChanges, OperateInfo, OperationState } from "../../../../../lib/components/operation/operate";
import { errorHandler as errorHandlerFn } from "../../../../../lib/utils/shared/shared.utils";

@Component({
  selector: "hbr-helm-chart-version",
  templateUrl: "./helm-chart-version.component.html",
  styleUrls: ["./helm-chart-version.component.scss"],
})
export class ChartVersionComponent implements OnInit {
  signedCon: { [key: string]: any | string[] } = {};
  @Input() projectId: number;
  @Input() projectName: string;
  @Input() chartName: string;
  @Input() roleName: string;
  @Input() hasSignedIn: boolean;
  @Input() chartDefaultIcon: string = DefaultHelmIcon;
  @Output() versionClickEvt = new EventEmitter<string>();
  @Output() backEvt = new EventEmitter<any>();


  lastFilteredVersionName: string;
  chartVersions: HelmChartVersion[] = [];
  systemInfo: SystemInfo;
  selectedRows: HelmChartVersion[] = [];
  labels: Label[] = [];
  loading = true;
  resourceType = ResourceType.CHART_VERSION;

  isCardView: boolean;
  cardHover = false;
  listHover = false;

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: State;

  chartFile: File;
  provFile: File;

  addLabelHeaders = 'HELM_CHART.ADD_LABEL_TO_CHART_VERSION';

  @ViewChild("confirmationDialog", {static: false})
  confirmationDialog: ConfirmationDialogComponent;
  hasAddRemoveHelmChartVersionPermission: boolean;
  hasDownloadHelmChartVersionPermission: boolean;
  hasDeleteHelmChartVersionPermission: boolean;
  constructor(
    private errorHandler: ErrorHandler,
    private systemInfoService: SystemInfoService,
    private helmChartService: HelmChartService,
    private resrouceLabelService: LabelService,
    public userPermissionService: UserPermissionService,
    private operationService: OperationService,
    private translateService: TranslateService,
  ) { }

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : "";
  }

  ngOnInit(): void {
    // Get system info for tag views
    this.systemInfoService.getSystemInfo()
      .subscribe(systemInfo => (this.systemInfo = systemInfo)
      , error => this.errorHandler.error(error));
    this.refresh();
    this.getLabels();
    this.lastFilteredVersionName = "";
    this.getHelmChartVersionPermission(this.projectId);
  }

  updateFilterValue(value: string) {
    this.lastFilteredVersionName = value;
    this.refresh();
  }

  getLabels() {
    forkJoin(this.resrouceLabelService.getLabels("g"), this.resrouceLabelService.getProjectLabels(this.projectId))
      .subscribe(
        (labels) => {
          this.labels = [].concat(...labels);
        });
  }

  refresh() {
    this.loading = true;
    this.helmChartService
      .getChartVersions(this.projectName, this.chartName)
      .pipe(finalize(() => {
        this.selectedRows = [];
        this.loading = false;
      }))
      .subscribe(
        versions => {
          this.chartVersions = versions.filter(x => x.version.includes(this.lastFilteredVersionName));
          this.totalCount = versions.length;
        },
        err => {
          this.errorHandler.error(err);
        }
      );
  }

  getMaintainerString(maintainers: HelmChartMaintainer[]) {
    if (!maintainers || maintainers.length < 1) {
      return "";
    }

    let maintainer_string = maintainers[0].name;
    if (maintainers.length > 1) {
      maintainer_string = `${maintainer_string} (${maintainers.length - 1} others)`;
    }
    return maintainer_string;
  }

  onVersionClick(version: HelmChartVersion) {
    this.versionClickEvt.emit(version.version);
  }

  deleteVersion(version: HelmChartVersion): Observable<any> {
    // init operation info
    let operateMsg = new OperateInfo();
    operateMsg.name = "OPERATION.DELETE_CHART_VERSION";
    operateMsg.data.id = version.digest;
    operateMsg.state = OperationState.progressing;
    operateMsg.data.name = `${version.name}:${version.version}`;
    this.operationService.publishInfo(operateMsg);

    return this.helmChartService
      .deleteChartVersion(this.projectName, this.chartName, version.version)
      .pipe(map(
        () => operateChanges(operateMsg, OperationState.success)),
        catchError( error => {
          const message = errorHandlerFn(error);
          this.translateService.get(message).subscribe(res =>
            operateChanges(operateMsg, OperationState.failure, res)
          );
          return observableThrowError(error);
        }
      ));
  }

  deleteVersions(versions: HelmChartVersion[]) {
    if (versions && versions.length < 1) { return; }
    let successCount: number;
    let totalCount = this.chartVersions.length;
    let versionObs = versions.map(v => this.deleteVersion(v));
    forkJoin(versionObs).pipe(finalize(() => {
      if (totalCount !== successCount) {
        this.refresh();
      }
    })).subscribe(res => {
      successCount = res.filter(r => r.state === OperationState.success).length;
      if (totalCount === successCount) {
        this.backEvt.emit();
      }
    }, error => {
      this.errorHandler.error(error);
    });
  }

  versionDownload(evt?: Event, item?: HelmChartVersion) {
    if (evt) {
      evt.stopPropagation();
    }
    let selectedVersion: HelmChartVersion;

    if (item) {
      selectedVersion = item;
    } else {
      // return if selected version less then 1
      if (this.selectedRows.length < 1) {
        return;
      }
      selectedVersion = this.selectedRows[0];
    }
    if (!selectedVersion) {
      return;
    }

    let filename = selectedVersion.urls[0];
    this.helmChartService.downloadChart(this.projectName, filename).subscribe(
      res => {
        downloadFile(res);
      },
      error => {
        this.errorHandler.error(error);
      }
    );
  }

  showCard(cardView: boolean) {
    if (this.isCardView === cardView) {
      return;
    }
    this.isCardView = cardView;
  }

  mouseEnter(itemName: string) {
    if (itemName === "card") {
      this.cardHover = true;
    } else {
      this.listHover = true;
    }
  }

  mouseLeave(itemName: string) {
    if (itemName === "card") {
      this.cardHover = false;
    } else {
      this.listHover = false;
    }
  }

  isHovering(itemName: string) {
    if (itemName === "card") {
      return this.cardHover;
    } else {
      return this.listHover;
    }
  }

  onChartFileChangeEvent(event) {
    if (event.target.files && event.target.files.length > 0) {
      this.chartFile = event.target.files[0];
    }
  }
  onProvFileChangeEvent(event) {
    if (event.target.files && event.target.files.length > 0) {
      this.provFile = event.target.files[0];
    }
  }

  deleteVersionCard(env: Event, version: HelmChartVersion) {
    env.stopPropagation();
    this.openVersionDeleteModal([version]);
  }

  openVersionDeleteModal(versions?: HelmChartVersion[]) {
    if (!versions) {
      versions = this.selectedRows;
    }
    let versionNames = versions.map(v => v.version).join(",");
    let message = new ConfirmationMessage(
      "HELM_CHART.DELETE_CHART_VERSION_TITLE",
      "HELM_CHART.DELETE_CHART_VERSION",
      versionNames,
      versions,
      ConfirmationTargets.HELM_CHART_VERSION,
      ConfirmationButtons.DELETE_CANCEL
    );
    this.confirmationDialog.open(message);
  }

  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (
      message &&
      message.source === ConfirmationTargets.HELM_CHART_VERSION &&
      message.state === ConfirmationState.CONFIRMED
    ) {
      let versions = message.data;
      this.deleteVersions(versions);
    }
  }

  getImgLink(v: HelmChartVersion) {
    if (v.icon) {
      return v.icon;
    } else {
      return DefaultHelmIcon;
    }
  }

  getDefaultIcon(v: HelmChartVersion) {
    v.icon = this.chartDefaultIcon;
  }

  getStatusString(chartVersion: HelmChartVersion) {
    if (chartVersion.deprecated) {
      return "HELM_CHART.DEPRECATED";
    } else {
      return "HELM_CHART.ACTIVE";
    }
  }

  onLabelChange(version: HelmChartVersion) {
    this.resrouceLabelService.getChartVersionLabels(this.projectName, this.chartName, version.version)
      .subscribe(labels => {
        let versionIdx = this.chartVersions.findIndex(v => v.name === version.name);
        this.chartVersions[versionIdx].labels = labels;
      });
  }

  getHelmChartVersionPermission(projectId: number): void {

    let hasAddRemoveHelmChartVersionPermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.HELM_CHART_VERSION_LABEL.KEY, USERSTATICPERMISSION.HELM_CHART_VERSION_LABEL.VALUE.CREATE);
    let hasDownloadHelmChartVersionPermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.HELM_CHART_VERSION.KEY, USERSTATICPERMISSION.HELM_CHART_VERSION.VALUE.READ);
    let hasDeleteHelmChartVersionPermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.HELM_CHART_VERSION.KEY, USERSTATICPERMISSION.HELM_CHART_VERSION.VALUE.DELETE);
    forkJoin(hasAddRemoveHelmChartVersionPermission, hasDownloadHelmChartVersionPermission, hasDeleteHelmChartVersionPermission)
    .subscribe(permissions => {
      this.hasAddRemoveHelmChartVersionPermission = permissions[0] as boolean;
      this.hasDownloadHelmChartVersionPermission = permissions[1] as boolean;
      this.hasDeleteHelmChartVersionPermission = permissions[2] as boolean;
    }, error => this.errorHandler.error(error));
  }
}
