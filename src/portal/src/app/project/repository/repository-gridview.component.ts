import {
  Component,
  Input,
  Output,
  OnInit,
  ViewChild,
  EventEmitter,
  OnChanges,
  SimpleChanges,
  Inject, OnDestroy
} from "@angular/core";
import { forkJoin, Subscription } from "rxjs";
import { debounceTime, distinctUntilChanged, switchMap } from "rxjs/operators";
import { TranslateService } from "@ngx-translate/core";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { ClrDatagridStateInterface } from "@clr/angular";
import {
  RepositoryService as NewRepositoryService
} from "../../../../ng-swagger-gen/services/repository.service";
import {
  SystemInfo,
  SystemInfoService,
  UserPermissionService, USERSTATICPERMISSION
} from "../../../lib/services";
import { FilterComponent } from "../../../lib/components/filter/filter.component";
import {
  calculatePage,
  clone,
  DEFAULT_PAGE_SIZE,
  dbEncodeURIComponent,
  doFiltering,
  doSorting
} from "../../../lib/utils/utils";
import { IServiceConfig, SERVICE_CONFIG } from "../../../lib/entities/service.config";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { ConfirmationButtons, ConfirmationState, ConfirmationTargets } from "../../../lib/entities/shared.const";
import { operateChanges, OperateInfo, OperationState } from "../../../lib/components/operation/operate";
import {
  ConfirmationAcknowledgement,
  ConfirmationDialogComponent,
  ConfirmationMessage
} from "../../../lib/components/confirmation-dialog";
import { OperationService } from "../../../lib/components/operation/operation.service";
import { errorHandler } from "../../../lib/utils/shared/shared.utils";
import { Project } from "../project";
import { ActivatedRoute, Router } from "@angular/router";
import { SessionService } from "../../shared/session.service";
import { GridViewComponent } from "./gridview/grid-view.component";
import { Repository as NewRepository } from "../../../../ng-swagger-gen/models/repository";
import { StrictHttpResponse as __StrictHttpResponse } from '../../../../ng-swagger-gen/strict-http-response';


@Component({
  selector: "hbr-repository-gridview",
  templateUrl: "./repository-gridview.component.html",
  styleUrls: ["./repository-gridview.component.scss"],
})
export class RepositoryGridviewComponent implements OnChanges, OnInit, OnDestroy {
  signedCon: { [key: string]: any | string[] } = {};
  downloadLink: string;
  @Input() urlPrefix: string;
  projectId: number;
  hasProjectAdminRole: boolean;
  hasSignedIn: boolean;
  projectName: string;
  mode = 'standalone';
  @Output() repoProvisionEvent = new EventEmitter<NewRepository>();
  @Output() addInfoEvent = new EventEmitter<NewRepository>();

  lastFilteredRepoName: string;
  repositories: NewRepository[] = [];
  repositoriesCopy: NewRepository[] = [];
  systemInfo: SystemInfo;
  selectedRow: NewRepository[] = [];
  loading = true;

  isCardView: boolean;
  cardHover = false;
  listHover = false;

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: ClrDatagridStateInterface;

  @ViewChild("confirmationDialog", {static: false})
  confirmationDialog: ConfirmationDialogComponent;

  @ViewChild("gridView", {static: false}) gridView: GridViewComponent;
  hasCreateRepositoryPermission: boolean;
  hasDeleteRepositoryPermission: boolean;
  @ViewChild(FilterComponent, {static: true})
  filterComponent: FilterComponent;
  searchSub: Subscription;

  constructor(@Inject(SERVICE_CONFIG) private configInfo: IServiceConfig,
              private errorHandlerService: ErrorHandler,
              private translateService: TranslateService,
              private newRepoService: NewRepositoryService,
              private systemInfoService: SystemInfoService,
              private operationService: OperationService,
              private userPermissionService: UserPermissionService,
              private route: ActivatedRoute,
              private session: SessionService,
              private router: Router,
  ) {
    if (this.configInfo && this.configInfo.systemInfoEndpoint) {
      this.downloadLink = this.configInfo.systemInfoEndpoint + "/getcert";
    }
  }

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : "";
  }
  public get withAdmiral(): boolean {
    return this.mode === "admiral";
  }

  get canDownloadCert(): boolean {
    return this.systemInfo && this.systemInfo.has_ca_root;
  }

  goIntoRepo(repoEvt: NewRepository): void {
    let linkUrl = ['harbor', 'projects', repoEvt.project_id, 'repositories', repoEvt.name.substr(this.projectName.length + 1)];
    this.router.navigate(linkUrl);
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes["projectId"] && changes["projectId"].currentValue) {
      this.refresh();
    }
  }

  ngOnInit(): void {
    this.projectId = this.route.snapshot.parent.params['id'];
    let resolverData = this.route.snapshot.parent.data;
    if (resolverData) {
      let pro: Project = <Project>resolverData['projectResolver'];
      this.hasProjectAdminRole = pro.has_project_admin_role;
      this.projectName = pro.name;
    }
    this.hasSignedIn = this.session.getCurrentUser() !== null;
    // Get system info for tag views
   this.getSystemInfo();
    this.isCardView = this.mode === "admiral";
    this.lastFilteredRepoName = "";
    this.getHelmChartVersionPermission(this.projectId);
    if (!this.searchSub) {
      this.searchSub = this.filterComponent.filterTerms.pipe(
        debounceTime(500),
        distinctUntilChanged(),
        switchMap(repoName => {
          this.lastFilteredRepoName = repoName;
          this.currentPage = 1;
          // Pagination
          let params: NewRepositoryService.ListRepositoriesParams = {
            projectName: this.projectName,
            page: this.currentPage,
            pageSize: this.pageSize,
        };
        if (this.lastFilteredRepoName) {
          params.q = encodeURIComponent(`name=~${this.lastFilteredRepoName}`);
        }
          this.loading = true;
          return this.newRepoService.listRepositoriesResponse(params);
        })
      ).subscribe((repo: __StrictHttpResponse<Array<NewRepository>>) => {
        this.totalCount = +repo.headers.get('x-total-count');
        this.repositories = repo.body;
        this.loading = false;
      }, error => {
        this.loading = false;
        this.errorHandlerService.error(error);
      });
    }
  }
  getSystemInfo() {
      this.systemInfoService.getSystemInfo()
        .subscribe(systemInfo => (this.systemInfo = systemInfo)
          , error => this.errorHandlerService.error(error));
  }
  ngOnDestroy() {
    if (this.searchSub) {
      this.searchSub.unsubscribe();
      this.searchSub = null;
    }
  }

  confirmDeletion(message: ConfirmationAcknowledgement) {
    this.loading = true;
    // forkJoin(...repArr).subscribe(() => {
    if (message &&
      message.source === ConfirmationTargets.REPOSITORY &&
      message.state === ConfirmationState.CONFIRMED) {
      let repoLists = message.data;
      if (repoLists && repoLists.length) {
        let observableLists: any[] = [];
        repoLists.forEach(repo => {
          observableLists.push(this.delOperate(repo));
        });
        forkJoin(observableLists).subscribe((item) => {
          this.selectedRow = [];
          this.refresh();
          let st: ClrDatagridStateInterface = this.getStateAfterDeletion();
          if (!st) {
            this.refresh();
          } else {
            this.clrLoad(st);
          }
        }, error => {
          this.errorHandlerService.error(error);
          this.loading = false;
          this.refresh();
        });
      }
    }
  }

  delOperate(repo: NewRepository): Observable<any> {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DELETE_REPO';
    operMessage.data.id = repo.id;
    operMessage.state = OperationState.progressing;
    repo.name = repo.name.substr(this.projectName.length + 1);
    operMessage.data.name = repo.name;

    this.operationService.publishInfo(operMessage);
    return this.newRepoService
      .deleteRepository({
        repositoryName: dbEncodeURIComponent(repo.name),
        projectName: this.projectName
      })
      .pipe(map(
        response => {
          this.translateService.get('BATCH.DELETED_SUCCESS').subscribe(res => {
            operateChanges(operMessage, OperationState.success);
          });
        }), catchError(error => {
        const message = errorHandler(error);
        this.translateService.get(message).subscribe(res =>
          operateChanges(operMessage, OperationState.failure, res)
        );
        return observableThrowError(error);
      }));
  }

  doSearchRepoNames(repoName: string) {
    this.lastFilteredRepoName = repoName;
    this.currentPage = 1;
    let st: ClrDatagridStateInterface = this.currentState;
    if (!st || !st.page) {
      st = {page: {}};
    }
    st.page.size = this.pageSize;
    st.page.from = 0;
    st.page.to = this.pageSize - 1;
    this.clrLoad(st);
  }

  deleteRepos(repoLists: NewRepository[]) {
    if (repoLists && repoLists.length) {
      let repoNames: string[] = [];
      repoLists.forEach(repo => {
        repoNames.push(repo.name);
      });
      this.confirmationDialogSet(
        'REPOSITORY.DELETION_TITLE_REPO',
        '',
        repoNames.join(','),
        repoLists,
        'REPOSITORY.DELETION_SUMMARY_REPO',
        ConfirmationButtons.DELETE_CANCEL);
    }
  }

  confirmationDialogSet(summaryTitle: string, signature: string,
                        repoName: string, repoLists: NewRepository[],
                        summaryKey: string, button: ConfirmationButtons): void {
    this.translateService.get(summaryKey,
      {
        repoName: repoName,
        signedImages: signature,
      }).subscribe((res: string) => {
      summaryKey = res;
      let message = new ConfirmationMessage(
        summaryTitle,
        summaryKey,
        repoName,
        repoLists,
        ConfirmationTargets.REPOSITORY,
        button);
      this.confirmationDialog.open(message);


    });
  }

  itemAddInfoEvent(evt: any, repo: NewRepository): void {
    evt.stopPropagation();
    let repoCopy = clone(repo);
    repoCopy.name = this.registryUrl + ":443/" + repoCopy.name;
    this.addInfoEvent.emit(repoCopy);
  }

  deleteItemEvent(evt: any, item: NewRepository): void {
    evt.stopPropagation();
    this.deleteRepos([item]);
  }

  refresh() {
    this.doSearchRepoNames("");
  }

  loadNextPage() {
    this.currentPage = this.currentPage + 1;
    // Pagination
    let params: NewRepositoryService.ListRepositoriesParams = {
      projectName: this.projectName,
      page: this.currentPage,
      pageSize: this.pageSize
    };
    if (this.lastFilteredRepoName) {
      params.q = encodeURIComponent(`name=~${this.lastFilteredRepoName}`);
    }

    this.loading = true;
    this.newRepoService.listRepositoriesResponse(
      params
    )
      .subscribe((repo: __StrictHttpResponse<Array<NewRepository>>) => {
        this.totalCount = +repo.headers.get('x-total-count');
        this.repositoriesCopy = repo.body;
        this.repositories = this.repositories.concat(this.repositoriesCopy);
        this.loading = false;
      }, error => {
        this.loading = false;
        this.errorHandlerService.error(error);
      });
  }

  clrLoad(state: ClrDatagridStateInterface): void {
    if (!state || !state.page) {
      return;
    }
    this.selectedRow = [];
    // Keep it for future filtering and sorting
    this.currentState = state;

    let pageNumber: number = calculatePage(state);
    if (pageNumber <= 0) {
      pageNumber = 1;
    }

    // Pagination
    let params: NewRepositoryService.ListRepositoriesParams = {
      projectName: this.projectName,
      page: pageNumber,
      pageSize: this.pageSize
    };
    if (this.lastFilteredRepoName) {
      params.q = encodeURIComponent(`name=~${this.lastFilteredRepoName}`);
    }
    if (state.filters && state.filters.length) {
      state.filters.forEach(item => {
        params[item.property] = item.value;
      });
    }
    if (state.sort && state.sort.by) {
      // params = params.set(`sort`, `${(state.sort.reverse ? `-` : ``)}${state.sort.by as string}`);
    }
    this.loading = true;

    this.newRepoService.listRepositoriesResponse(
      params
    )
      .subscribe((repo: __StrictHttpResponse<Array<NewRepository>>) => {

        this.totalCount = +repo.headers.get('x-total-count');
        this.repositories = repo.body;
        // Do customising filtering and sorting
        this.repositories = doFiltering<NewRepository>(this.repositories, state);
        this.repositories = doSorting<NewRepository>(this.repositories, state);
        this.signedCon = {};
        this.loading = false;
      }, error => {
        this.loading = false;
        this.errorHandlerService.error(error);
      });
  }

  getStateAfterDeletion(): ClrDatagridStateInterface {
    let total: number = this.totalCount - 1;
    if (total <= 0) {
      return null;
    }

    let totalPages: number = Math.ceil(total / this.pageSize);
    let targetPageNumber: number = this.currentPage;

    if (this.currentPage > totalPages) {
      targetPageNumber = totalPages; // Should == currentPage -1
    }

    let st: ClrDatagridStateInterface = this.currentState;
    if (!st) {
      st = {page: {}};
    }
    st.page.size = this.pageSize;
    st.page.from = (targetPageNumber - 1) * this.pageSize;
    st.page.to = targetPageNumber * this.pageSize - 1;

    return st;
  }
  getImgLink(repo: NewRepository): string {
    return "/container-image-icons?container-image=" + repo.name;
  }

  showCard(cardView: boolean) {
    if (this.isCardView === cardView) {
      return;
    }
    this.isCardView = cardView;
    this.refresh();
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

  getHelmChartVersionPermission(projectId: number): void {

    let hasCreateRepositoryPermission = this.userPermissionService.getPermission(this.projectId,
      USERSTATICPERMISSION.REPOSITORY.KEY, USERSTATICPERMISSION.REPOSITORY.VALUE.CREATE);
    let hasDeleteRepositoryPermission = this.userPermissionService.getPermission(this.projectId,
      USERSTATICPERMISSION.REPOSITORY.KEY, USERSTATICPERMISSION.REPOSITORY.VALUE.DELETE);
    forkJoin(hasCreateRepositoryPermission, hasDeleteRepositoryPermission).subscribe(permissions => {
      this.hasCreateRepositoryPermission = permissions[0] as boolean;
      this.hasDeleteRepositoryPermission = permissions[1] as boolean;
    }, error => this.errorHandlerService.error(error));
  }
}
