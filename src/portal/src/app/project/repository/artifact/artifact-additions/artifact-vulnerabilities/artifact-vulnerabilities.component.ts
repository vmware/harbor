import { Component, Input, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { AdditionsService } from "../additions.service";
import { ClrDatagridComparatorInterface, ClrLoadingState } from "@clr/angular";
import { finalize } from "rxjs/operators";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import {
  ScannerVo,
  ScanningResultService,
  UserPermissionService,
  USERSTATICPERMISSION,
  VulnerabilityItem
} from "../../../../../../lib/services";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";
import {
  DEFAULT_SUPPORTED_MIME_TYPE,
  SEVERITY_LEVEL_MAP,
  VULNERABILITY_SEVERITY
} from "../../../../../../lib/utils/utils";
import { ChannelService } from "../../../../../../lib/services/channel.service";
import { ResultBarChartComponent } from "../../../vulnerability-scanning/result-bar-chart.component";
import { Subscription } from "rxjs";
import { Artifact } from "../../../../../../../ng-swagger-gen/models/artifact";

@Component({
  selector: 'hbr-artifact-vulnerabilities',
  templateUrl: './artifact-vulnerabilities.component.html',
  styleUrls: ['./artifact-vulnerabilities.component.scss']
})
export class ArtifactVulnerabilitiesComponent implements OnInit, OnDestroy {
  @Input()
  vulnerabilitiesLink: AdditionLink;
  @Input()
  projectName: string;
  @Input()
  projectId: number;
  @Input()
  repoName: string;
  @Input()
  digest: string;
  @Input() artifact: Artifact;
  scan_overview: any;
  scanner: ScannerVo;

  scanningResults: VulnerabilityItem[] = [];
  loading: boolean = false;
  hasEnabledScanner: boolean = false;
  scanBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  severitySort: ClrDatagridComparatorInterface<VulnerabilityItem>;
  hasScanningPermission: boolean = false;
  onSendingScanCommand: boolean = false;
  hasShowLoading: boolean = false;
  @ViewChild(ResultBarChartComponent, {static: false})
  resultBarChartComponent: ResultBarChartComponent;
  sub: Subscription;
  hasViewInitWithDelay: boolean = false;
  constructor(
    private errorHandler: ErrorHandler,
    private additionsService: AdditionsService,
    private userPermissionService: UserPermissionService,
    private scanningService: ScanningResultService,
    private channel: ChannelService,
  ) {
    const that = this;
    this.severitySort = {
      compare(a: VulnerabilityItem, b: VulnerabilityItem): number {
        return that.getLevel(a) - that.getLevel(b);
      }
    };
  }

  ngOnInit() {
    this.getVulnerabilities();
    this.getScanningPermission();
    this.getProjectScanner();
    if (!this.sub) {
      this.sub = this.channel.ArtifactDetail$.subscribe(tag => {
        this.getVulnerabilities();
      });
    }
    setTimeout(() => {
      this.hasViewInitWithDelay = true;
    }, 0);
  }
  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
      this.sub = null;
    }
  }
  getVulnerabilities() {
    if (this.vulnerabilitiesLink
      && !this.vulnerabilitiesLink.absolute
      && this.vulnerabilitiesLink.href) {
      if (!this.hasShowLoading) {
        this.loading = true;
        this.hasShowLoading = true;
      }
      this.additionsService.getDetailByLink(this.vulnerabilitiesLink.href)
        .pipe(finalize(() => {
          this.loading = false;
          this.hasShowLoading = false;
        }))
        .subscribe(
          res  => {
            this.scan_overview = res;
            if (this.scan_overview && this.scan_overview[DEFAULT_SUPPORTED_MIME_TYPE]) {
              this.scanningResults = this.scan_overview[DEFAULT_SUPPORTED_MIME_TYPE].vulnerabilities || [];
              // sort
              if (this.scanningResults) {
                this.scanningResults.sort(((a, b) => this.getLevel(b) - this.getLevel(a)));
              }
              this.scanner = this.scan_overview[DEFAULT_SUPPORTED_MIME_TYPE].scanner;
            }
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }
  getScanningPermission(): void {
    const permissions = [
      { resource: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE.CREATE },
    ];
    this.userPermissionService.hasProjectPermissions(this.projectId, permissions).subscribe((results: Array<boolean>) => {
      this.hasScanningPermission = results[0];
      // only has label permission
    }, error => this.errorHandler.error(error));
  }
  getProjectScanner(): void {
    this.hasEnabledScanner = false;
    this.scanBtnState = ClrLoadingState.LOADING;
    this.scanningService.getProjectScanner(this.projectId)
      .subscribe(response => {
        if (response && "{}" !== JSON.stringify(response) && !response.disabled
          && response.health === "healthy") {
          this.scanBtnState = ClrLoadingState.SUCCESS;
          this.hasEnabledScanner = true;
        } else {
          this.scanBtnState = ClrLoadingState.ERROR;
        }
      }, error => {
        this.scanBtnState = ClrLoadingState.ERROR;
      });
  }
  getLevel(v: VulnerabilityItem): number {
    if (v && v.severity && SEVERITY_LEVEL_MAP[v.severity]) {
      return SEVERITY_LEVEL_MAP[v.severity];
    }
    return 0;
  }
  refresh(): void {
    this.getVulnerabilities();
  }

  severityText(severity: string): string {
    switch (severity) {
      case VULNERABILITY_SEVERITY.CRITICAL:
        return 'VULNERABILITY.SEVERITY.CRITICAL';
      case VULNERABILITY_SEVERITY.HIGH:
        return 'VULNERABILITY.SEVERITY.HIGH';
      case VULNERABILITY_SEVERITY.MEDIUM:
        return 'VULNERABILITY.SEVERITY.MEDIUM';
      case VULNERABILITY_SEVERITY.LOW:
        return 'VULNERABILITY.SEVERITY.LOW';
      case VULNERABILITY_SEVERITY.NEGLIGIBLE:
        return 'VULNERABILITY.SEVERITY.NEGLIGIBLE';
      case VULNERABILITY_SEVERITY.UNKNOWN:
        return 'VULNERABILITY.SEVERITY.UNKNOWN';
      default:
        return 'UNKNOWN';
    }
  }
  scanNow() {
    this.onSendingScanCommand = true;
    this.channel.publishScanEvent(this.repoName + "/" + this.digest);
  }
  submitFinish(e: boolean) {
    this.onSendingScanCommand = e;
  }
  shouldShowBar(): boolean {
    return this.hasViewInitWithDelay && this.resultBarChartComponent
      && (this.resultBarChartComponent.queued || this.resultBarChartComponent.scanning || this.resultBarChartComponent.error);
  }
  hasScanned(): boolean {
    return this.hasViewInitWithDelay && this.resultBarChartComponent
      && !(this.resultBarChartComponent.completed
        || this.resultBarChartComponent.error
        || this.resultBarChartComponent.queued
        || this.resultBarChartComponent.scanning);
  }
  handleScanOverview(scanOverview: any): any {
    if (scanOverview) {
      return scanOverview[DEFAULT_SUPPORTED_MIME_TYPE];
    }
    return null;
  }
}
