import { Component, OnInit, Input, ViewChild, OnDestroy } from '@angular/core';
import { Observable, of, forkJoin, Subject, Subscription } from 'rxjs';
import { map, catchError, finalize, debounceTime, distinctUntilChanged, switchMap } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { NgForm } from '@angular/forms';
import { AVAILABLE_TIME } from "../../artifact-list-page/artifact-list/artifact-list-tab/artifact-list-tab.component";
import {
  ConfirmationAcknowledgement,
  ConfirmationDialogComponent,
  ConfirmationMessage
} from "../../../../../lib/components/confirmation-dialog";
import { OperationService } from "../../../../../lib/components/operation/operation.service";
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { ConfirmationButtons, ConfirmationState, ConfirmationTargets } from "../../../../../lib/entities/shared.const";
import { operateChanges, OperateInfo, OperationState } from "../../../../../lib/components/operation/operate";
import { errorHandler } from "../../../../../lib/utils/shared/shared.utils";
import { ArtifactFront as Artifact } from "../artifact";
import { ArtifactService } from '../../../../../../ng-swagger-gen/services/artifact.service';
import { Tag } from '../../../../../../ng-swagger-gen/models/tag';
import {
  UserPermissionService, USERSTATICPERMISSION, SystemInfoService, SystemInfo
} from "../../../../../lib/services";
import { ClrDatagridStateInterface } from '@clr/angular';
import {
  DEFAULT_PAGE_SIZE,
  calculatePage,
  dbEncodeURIComponent,
  doFiltering,
  doSorting
} from '../../../../../lib/utils/utils';
import { AppConfigService } from "../../../../services/app-config.service";

class InitTag {
  name = "";
}
const DeleteTagWithNotoryCommand1 = 'notary -s https://';
const DeleteTagWithNotoryCommand2 = ':4443 -d ~/.docker/trust remove -p ';
@Component({
  selector: 'artifact-tag',
  templateUrl: './artifact-tag.component.html',
  styleUrls: ['./artifact-tag.component.scss']
})
export class ArtifactTagComponent implements OnInit, OnDestroy {
  @Input() artifactDetails: Artifact;
  @Input() projectName: string;
  @Input() projectId: number;
  @Input() repositoryName: string;
  newTagName = new InitTag();
  newTagForm: NgForm;
  @ViewChild("newTagForm", { static: true }) currentForm: NgForm;
  selectedRow: Tag[] = [];
  isTagNameExist = false;
  newTagformShow = false;
  loading = true;
  openTag = false;
  availableTime = AVAILABLE_TIME;
  @ViewChild("confirmationDialog", { static: false })
  confirmationDialog: ConfirmationDialogComponent;
  hasDeleteTagPermission: boolean;
  hasCreateTagPermission: boolean;

  totalCount: number = 0;
  allTags: Tag[] = [];
  currentTags: Tag[] = [];
  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  tagNameChecker: Subject<string> = new Subject<string>();
  tagNameCheckSub: Subscription;
  tagNameCheckOnGoing = false;
  systemInfo: SystemInfo;
  constructor(
    private operationService: OperationService,
    private artifactService: ArtifactService,
    private translateService: TranslateService,
    private userPermissionService: UserPermissionService,
    private systemInfoService: SystemInfoService,
    private appConfigService: AppConfigService,
    private errorHandlerService: ErrorHandler

  ) { }
  ngOnInit() {
    this.getImagePermissionRule(this.projectId);
    this.invalidCreateTag();
    this.systemInfoService.getSystemInfo()
      .subscribe(systemInfo => this.systemInfo = systemInfo, error => this.errorHandlerService.error(error));
  }
  checkTagName(name) {
      let listArtifactParams: ArtifactService.ListArtifactsParams = {
        projectName: this.projectName,
        repositoryName: dbEncodeURIComponent(this.repositoryName),
        withLabel: true,
        withScanOverview: true,
        withTag: true,
        q: encodeURIComponent(`tags=${name}`)
      };
      return this.artifactService.listArtifacts(listArtifactParams)
      .pipe(finalize(() => this.tagNameCheckOnGoing = false));
  }
  invalidCreateTag() {
    if (!this.tagNameCheckSub) {
      this.tagNameCheckSub = this.tagNameChecker
        .pipe(debounceTime(200))
        .pipe(distinctUntilChanged())
        .pipe(switchMap(name => {
          this.tagNameCheckOnGoing = true;
          this.isTagNameExist = false;
          return this.checkTagName(name);
        }))
        .subscribe(response => {
        // tag existing
        if (response && response.length) {
            this.isTagNameExist = true;
        }
        }, error => {
          this.errorHandlerService.error(error);
        });
    }
  }
  getCurrentArtifactTags(state: ClrDatagridStateInterface) {
    if (!state || !state.page) {
      return ;
    }
    let pageNumber: number = calculatePage(state);
      if (pageNumber <= 0) { pageNumber = 1; }
    let params: ArtifactService.ListTagsParams = {
      projectName: this.projectName,
      repositoryName: dbEncodeURIComponent(this.repositoryName),
      reference: this.artifactDetails.digest,
      page: pageNumber,
      withSignature: true,
      withImmutableStatus: true,
      pageSize: this.pageSize
    };
    this.artifactService.listTagsResponse(params).pipe(finalize(() => {
      this.loading = false;
    })).subscribe(res => {
      if (res.headers) {
        let xHeader: string = res.headers.get("x-total-count");
        if (xHeader) {
          this.totalCount = Number.parseInt(xHeader);
        }
      }
      this.currentTags = res.body;
      // Do customising filtering and sorting
      this.currentTags = doFiltering<Tag>(this.currentTags, state);
      this.currentTags = doSorting<Tag>(this.currentTags, state);
    }, error => {
      this.errorHandlerService.error(error);
    });
  }
  getImagePermissionRule(projectId: number): void {
    const permissions = [
      { resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE },
    { resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.CREATE },

    ];
    this.userPermissionService.hasProjectPermissions(this.projectId, permissions).subscribe((results: Array<boolean>) => {
      this.hasDeleteTagPermission = results[0];
      this.hasCreateTagPermission = results[1];
    }, error => this.errorHandlerService.error(error));
  }

  addTag() {
    this.newTagformShow = true;

  }
  cancelAddTag() {
    this.newTagformShow = false;
    this.newTagName = new InitTag();
  }
  saveAddTag() {
    // const tag: NewTag = {name: this.newTagName};
    const createTagParams: ArtifactService.CreateTagParams = {
      projectName: this.projectName,
      repositoryName: dbEncodeURIComponent(this.repositoryName),
      reference: this.artifactDetails.digest,
      tag:  this.newTagName
    };
    this.loading = true;
    this.artifactService.createTag(createTagParams).subscribe(res => {
      this.newTagformShow = false;
      this.newTagName = new InitTag();
      this.currentPage = 1;
      let st: ClrDatagridStateInterface = { page: {from: 0, to: this.pageSize - 1, size: this.pageSize} };
      this.getCurrentArtifactTags(st);
    }, error => {
      this.loading = false;
      this.errorHandlerService.error(error);
    });
  }
  removeTag() {
    if (this.selectedRow && this.selectedRow.length) {
      let tagNames: string[] = [];
      this.selectedRow.forEach(artifact => {
        tagNames.push(artifact.name);
      });
      let titleKey: string, summaryKey: string, content: string, buttons: ConfirmationButtons;
      titleKey = "REPOSITORY.DELETION_TITLE_TAG";
      summaryKey = "REPOSITORY.DELETION_SUMMARY_TAG";
      buttons = ConfirmationButtons.DELETE_CANCEL;
      content = tagNames.join(" , ");

      let message = new ConfirmationMessage(
        titleKey,
        summaryKey,
        content,
        this.selectedRow,
        ConfirmationTargets.TAG,
        buttons);
      this.confirmationDialog.open(message);
    }
  }
  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (message &&
      message.source === ConfirmationTargets.TAG
      && message.state === ConfirmationState.CONFIRMED) {
      let tagList: Tag[] = message.data;
      if (tagList && tagList.length) {
        let observableLists: any[] = [];
        tagList.forEach(tag => {
          observableLists.push(this.delOperate(tag));
        });
        this.loading = true;
        forkJoin(...observableLists).subscribe((deleteResult) => {
          // if delete one success  refresh list
          let deleteSuccessList = [];
          let deleteErrorList = [];
          deleteResult.forEach(result => {
            if (!result) {
              // delete success
              deleteSuccessList.push(result);
            } else {
              deleteErrorList.push(result);
            }
          });
          this.selectedRow = [];
          if (deleteSuccessList.length === deleteResult.length) {
            // all is success
            this.currentPage = 1;
            let st: ClrDatagridStateInterface = { page: {from: 0, to: this.pageSize - 1, size: this.pageSize} };
            this.getCurrentArtifactTags(st);
          } else if (deleteErrorList.length === deleteResult.length) {
            // all is error
            this.loading = false;
            this.errorHandlerService.error(deleteResult[deleteResult.length - 1]);
          } else {
            // some artifact delete success but it has error delete things
            this.errorHandlerService.error(deleteErrorList[deleteErrorList.length - 1]);
            // if delete one success  refresh list
            this.currentPage = 1;
            let st: ClrDatagridStateInterface = { page: {from: 0, to: this.pageSize - 1, size: this.pageSize} };
            this.getCurrentArtifactTags(st);
          }
        });
      }
    }
  }

  delOperate(tag: Tag): Observable<any> | null {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DELETE_TAG';
    operMessage.state = OperationState.progressing;
    operMessage.data.name = tag.name;
    this.operationService.publishInfo(operMessage);

    if (tag.signed) {
      forkJoin(this.translateService.get("BATCH.DELETED_FAILURE"),
        this.translateService.get("REPOSITORY.DELETION_SUMMARY_TAG_DENIED")).subscribe(res => {
          let wrongInfo: string = res[1] + DeleteTagWithNotoryCommand1 + this.registryUrl +
            DeleteTagWithNotoryCommand2 +
            this.registryUrl + "/" + this.repositoryName +
            " " + name;
          operateChanges(operMessage, OperationState.failure, wrongInfo);
        });
        return of(null);
    } else {
     const deleteTagParams: ArtifactService.DeleteTagParams = {
      projectName: this.projectName,
      repositoryName: dbEncodeURIComponent(this.repositoryName),
      reference: this.artifactDetails.digest,
      tagName: tag.name
    };
    return this.artifactService.deleteTag(deleteTagParams)
      .pipe(map(
        response => {
          this.translateService.get("BATCH.DELETED_SUCCESS")
            .subscribe(res => {
              operateChanges(operMessage, OperationState.success);
            });
        }), catchError(error => {
          const message = errorHandler(error);
          this.translateService.get(message).subscribe(res =>
            operateChanges(operMessage, OperationState.failure, res)
          );
          return of(error);
        }));
      }
  }

  existValid(name) {
    if (name) {
      this.tagNameChecker.next(name);
    } else {
      this.isTagNameExist = false;
    }
  }
  toggleTagListOpenOrClose() {
    this.openTag = !this.openTag;
    this.newTagformShow = false;
  }
  hasImmutableOnTag(): boolean {
    return this.selectedRow.some((artifact) => artifact.immutable);
  }
  refresh() {
    this.loading = true;
    this.currentPage = 1;
    let st: ClrDatagridStateInterface = { page: {from: 0, to: this.pageSize - 1, size: this.pageSize} };
    this.getCurrentArtifactTags(st);
  }
  ngOnDestroy(): void {
    this.tagNameCheckSub.unsubscribe();
  }
  get withNotary(): boolean {
    return this.appConfigService.getConfig().with_notary;
  }
  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : '';
  }
}
