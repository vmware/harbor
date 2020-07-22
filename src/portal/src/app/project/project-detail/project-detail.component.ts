// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, OnInit, HostListener, AfterViewInit, OnDestroy, ChangeDetectorRef } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../project';
import { SessionService } from '../../shared/session.service';
import { AppConfigService } from "../../services/app-config.service";
import { forkJoin, Subject, Subscription } from "rxjs";
import { UserPermissionService, USERSTATICPERMISSION } from "../../../lib/services";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { debounceTime } from 'rxjs/operators';
import { DOWN, SHOW_ELLIPSIS_WIDTH, UP } from './project-detail.const';


@Component({
  selector: 'project-detail',
  templateUrl: 'project-detail.component.html',
  styleUrls: ['project-detail.component.scss']
})
export class ProjectDetailComponent implements OnInit, AfterViewInit, OnDestroy {
  hasSignedIn: boolean;
  currentProject: Project;

  isMember: boolean;
  roleName: string;
  projectId: number;
  hasProjectReadPermission: boolean;
  hasHelmChartsListPermission: boolean;
  hasRepositoryListPermission: boolean;
  hasMemberListPermission: boolean;
  hasLabelListPermission: boolean;
  hasLabelCreatePermission: boolean;
  hasLogListPermission: boolean;
  hasConfigurationListPermission: boolean;
  hasRobotListPermission: boolean;
  hasTagRetentionPermission: boolean;
  hasWebhookListPermission: boolean;
  hasScannerReadPermission: boolean;
  tabLinkNavList = [
    {
      linkName: "summary",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.SUMMARY",
      permissions: () => this.hasProjectReadPermission
    },
    {
      linkName: "repositories",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.REPOSITORIES",
      permissions: () => this.hasRepositoryListPermission
    },
    {
      linkName: "helm-charts",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.HELMCHART",
      permissions: () => this.withHelmChart && this.hasHelmChartsListPermission
    },
    {
      linkName: "members",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.USERS",
      permissions: () => this.hasMemberListPermission
    },
    {
      linkName: "labels",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.LABELS",
      permissions: () => (this.hasLabelListPermission && this.hasLabelCreatePermission) && !this.withAdmiral
    },
    {
      linkName: "scanner",
      tabLinkInOverflow: false,
      showTabName: "SCANNER.SCANNER",
      permissions: () => this.hasScannerReadPermission
    },
    {
      linkName: "tag-strategy",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.POLICY",
      permissions: () => this.hasTagRetentionPermission
    },
    {
      linkName: "robot-account",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.ROBOT_ACCOUNTS",
      permissions: () => this.hasRobotListPermission
    },
    {
      linkName: "webhook",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.WEBHOOKS",
      permissions: () => this.hasWebhookListPermission
    },
    {
      linkName: "logs",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.LOGS",
      permissions: () => this.hasLogListPermission
    },
    {
      linkName: "configs",
      tabLinkInOverflow: false,
      showTabName: "PROJECT_DETAIL.CONFIG",
      permissions: () => this.isSessionValid && this.hasConfigurationListPermission
    }
  ];
  previousWindowWidth: number;
  private _subject = new Subject<string>();
  private _subscription: Subscription;
  isProxyCacheProject: boolean = false;
  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private sessionService: SessionService,
    private appConfigService: AppConfigService,
    private userPermissionService: UserPermissionService,
    private errorHandler: ErrorHandler,
    private cdf: ChangeDetectorRef) {
    this.hasSignedIn = this.sessionService.getCurrentUser() !== null;
    this.route.data.subscribe(data => {
      this.currentProject = <Project>data['projectResolver'];
      if (this.currentProject.registry_id) {
        this.isProxyCacheProject = true;
      }
      this.isMember = this.currentProject.is_member;
      this.roleName = this.currentProject.role_name;
    });
  }
  ngOnInit() {
    this.projectId = this.route.snapshot.params['id'];
    this.getPermissionsList(this.projectId);
    if (!this._subscription) {
      this._subscription = this._subject.pipe(
          debounceTime(100)
      ).subscribe(
          type => {
            if (type === DOWN) {
              this.resetTabsForDownSize();
            } else {
              this.resetTabsForUpSize();
            }
          }
      );
    }
  }

  ngAfterViewInit() {
    this.previousWindowWidth = window.innerWidth;
    setTimeout(() => {
      this.resetTabsForDownSize();
    }, 0);
  }
  ngOnDestroy() {
    if (this._subscription) {
      this._subscription.unsubscribe();
      this._subscription = null;
    }
  }
  getPermissionsList(projectId: number): void {
    let permissionsList = [];
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.PROJECT.KEY, USERSTATICPERMISSION.PROJECT.VALUE.READ));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.LOG.KEY, USERSTATICPERMISSION.LOG.VALUE.LIST));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.CONFIGURATION.KEY, USERSTATICPERMISSION.CONFIGURATION.VALUE.READ));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.MEMBER.KEY, USERSTATICPERMISSION.MEMBER.VALUE.LIST));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.LABEL.KEY, USERSTATICPERMISSION.LABEL.VALUE.LIST));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.REPOSITORY.KEY, USERSTATICPERMISSION.REPOSITORY.VALUE.LIST));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.HELM_CHART.KEY, USERSTATICPERMISSION.HELM_CHART.VALUE.LIST));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.ROBOT.KEY, USERSTATICPERMISSION.ROBOT.VALUE.LIST));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.LABEL.KEY, USERSTATICPERMISSION.LABEL.VALUE.CREATE));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.TAG_RETENTION.KEY, USERSTATICPERMISSION.TAG_RETENTION.VALUE.READ));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.WEBHOOK.KEY, USERSTATICPERMISSION.WEBHOOK.VALUE.LIST));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.SCANNER.KEY, USERSTATICPERMISSION.SCANNER.VALUE.READ));

    forkJoin(...permissionsList).subscribe(Rules => {
      [this.hasProjectReadPermission, this.hasLogListPermission, this.hasConfigurationListPermission, this.hasMemberListPermission
        , this.hasLabelListPermission, this.hasRepositoryListPermission, this.hasHelmChartsListPermission, this.hasRobotListPermission
        , this.hasLabelCreatePermission, this.hasTagRetentionPermission, this.hasWebhookListPermission,
      this.hasScannerReadPermission] = Rules;
    }, error => this.errorHandler.error(error));
  }

  public get isSessionValid(): boolean {
    return this.sessionService.getCurrentUser() != null;
  }

  public get withAdmiral(): boolean {
    return this.appConfigService.getConfig().with_admiral;
  }

  public get withHelmChart(): boolean {
    return this.appConfigService.getConfig().with_chartmuseum;
  }

  backToProject(): void {
    if (window.sessionStorage) {
      window.sessionStorage.setItem('fromDetails', 'true');
    }
    this.router.navigate(['/harbor', 'projects']);
  }
  isDefaultTab(tab, index) {
    return this.route.snapshot.children[0].routeConfig.path !== tab.linkName && index === 0;
  }

  isTabLinkInOverFlow() {
    return this.tabLinkNavList.some(tab => {
      return tab.tabLinkInOverflow && this.route.snapshot.children[0].routeConfig.path === tab.linkName;
    });
  }

  @HostListener('window:resize', ['$event'])
  onResize(event) {
    if (this.previousWindowWidth) {
       // down size
      if (this.previousWindowWidth > event.target.innerWidth) {
          this._subject.next(DOWN);
      } else { // up size
          this._subject.next(UP);
      }
    }
    this.previousWindowWidth = event.target.innerWidth;
  }

  resetTabsForDownSize(): void {
    this.tabLinkNavList.filter(item => !item.tabLinkInOverflow).forEach((item, index) => {
      const tabEle: HTMLElement = document.getElementById('project-' + item.linkName);
      // strengthen code
      if (tabEle && tabEle.getBoundingClientRect) {
        const right: number = window.innerWidth - document.getElementById('project-' + item.linkName).getBoundingClientRect().right;
        if (right < SHOW_ELLIPSIS_WIDTH) {
          this.tabLinkNavList[index].tabLinkInOverflow = true;
        }
      }
    });
  }
  resetTabsForUpSize() {
    // 1.Set tabLinkInOverflow to false for all tabs(show all tabs)
    for ( let i = 0; i < this.tabLinkNavList.length; i++) {
      this.tabLinkNavList[i].tabLinkInOverflow = false;
    }
    // 2.Manually  detect changes to rerender dom
    this.cdf.detectChanges();
    // 3. Hide overflowed tabs
    this.resetTabsForDownSize();
  }
}
