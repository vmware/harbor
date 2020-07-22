import {
  Component,
  OnInit,
  ViewChild,
  OnDestroy,
  ChangeDetectorRef
} from "@angular/core";
import { AddRobotComponent } from "./add-robot/add-robot.component";
import { ActivatedRoute, Router } from "@angular/router";
import { Robot } from "./robot";
import { Project } from "./../project";
import { finalize, catchError, map } from "rxjs/operators";
import { TranslateService } from "@ngx-translate/core";
import { Subscription, forkJoin, Observable, throwError as observableThrowError } from "rxjs";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { RobotService } from "./robot-account.service";
import { ConfirmationMessage } from "../../shared/confirmation-dialog/confirmation-message";
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import {
  ConfirmationTargets,
  ConfirmationState,
  ConfirmationButtons
} from "../../shared/shared.const";
import { OperationService } from "../../../lib/components/operation/operation.service";
import { UserPermissionService, USERSTATICPERMISSION } from "../../../lib/services";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { operateChanges, OperateInfo, OperationState } from "../../../lib/components/operation/operate";
import { errorHandler as errorHandlerFn } from "../../../lib/utils/shared/shared.utils";

@Component({
  selector: "app-robot-account",
  templateUrl: "./robot-account.component.html",
  styleUrls: ["./robot-account.component.scss"]
})
export class RobotAccountComponent implements OnInit, OnDestroy {
  @ViewChild(AddRobotComponent, {static: false})
  addRobotComponent: AddRobotComponent;
  selectedRow: Robot[] = [];
  robotsCopy: Robot[] = [];
  loading = false;
  searchRobot: string;
  projectName: string;
  timerHandler: any;
  batchChangeInfos: {};
  isDisabled: boolean;
  isDisabledTip: string = "ROBOT_ACCOUNT.DISABLE_ACCOUNT";
  robots: Robot[];
  projectId: number;
  subscription: Subscription;
  hasRobotCreatePermission: boolean;
  hasRobotUpdatePermission: boolean;
  hasRobotDeletePermission: boolean;
  constructor(
    private route: ActivatedRoute,
    private robotService: RobotService,
    private OperateDialogService: ConfirmationDialogService,
    private operationService: OperationService,
    private translate: TranslateService,
    private userPermissionService: UserPermissionService,
    private errorHandler: ErrorHandler,
    private ref: ChangeDetectorRef,
    private messageHandlerService: MessageHandlerService
  ) {
    this.subscription = OperateDialogService.confirmationConfirm$.subscribe(
      message => {
        if (
          message &&
          message.state === ConfirmationState.CONFIRMED &&
          message.source === ConfirmationTargets.ROBOT_ACCOUNT
        ) {
          this.delRobots(message.data);
        }
      }
    );
    this.forceRefreshView(2000);
  }

  ngOnInit(): void {
    this.projectId = +this.route.snapshot.parent.params["id"];
    let resolverData = this.route.snapshot.parent.data;
    if (resolverData) {
      let project = <Project>resolverData["projectResolver"];
      this.projectName = project.name;
    }
    this.searchRobot = "";
    this.retrieve();
    this.getPermissionsList(this.projectId);
  }
  getPermissionsList(projectId: number): void {
    let permissionsList = [];
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.ROBOT.KEY, USERSTATICPERMISSION.ROBOT.VALUE.CREATE));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.ROBOT.KEY, USERSTATICPERMISSION.ROBOT.VALUE.UPDATE));
    permissionsList.push(this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.ROBOT.KEY, USERSTATICPERMISSION.ROBOT.VALUE.DELETE));

    forkJoin(...permissionsList).subscribe(Rules => {
      this.hasRobotCreatePermission = Rules[0] as boolean;
      this.hasRobotUpdatePermission = Rules[1] as boolean;
      this.hasRobotDeletePermission = Rules[2] as boolean;

    }, error => this.errorHandler.error(error));
  }
  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
    if (this.timerHandler) {
      clearInterval(this.timerHandler);
      this.timerHandler = null;
    }
  }

  openAddRobotModal(): void {
    this.addRobotComponent.openAddRobotModal();
  }

  openDeleteRobotsDialog(robots: Robot[]) {
    let robotNames = robots.map(robot => robot.name).join(",");
    let deletionMessage = new ConfirmationMessage(
      "ROBOT_ACCOUNT.DELETION_TITLE",
      "ROBOT_ACCOUNT.DELETION_SUMMARY",
      robotNames,
      robots,
      ConfirmationTargets.ROBOT_ACCOUNT,
      ConfirmationButtons.DELETE_CANCEL
    );
    this.OperateDialogService.openComfirmDialog(deletionMessage);
  }

  delRobots(robots: Robot[]): void {
    if (robots && robots.length < 1) {
      return;
    }
    let robotsDelete$ = robots.map(robot => this.delOperate(robot));
    forkJoin(robotsDelete$)
      .pipe(
        finalize(() => {
          this.retrieve();
          this.selectedRow = [];
        })
      )
      .subscribe(() => { }
      , error => {
        this.errorHandler.error(error);
      });
  }

  delOperate(robot: Robot) {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = "OPERATION.DELETE_ROBOT";
    operMessage.data.id = robot.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = robot.name;
    this.operationService.publishInfo(operMessage);

    return this.robotService
      .deleteRobotAccount(this.projectId, robot.id)
      .pipe(
        map(
          () => operateChanges(operMessage, OperationState.success),
          catchError(error => {
            const errorMsg = errorHandlerFn(error);
            this.translate.get(errorMsg).subscribe(res =>
              operateChanges(operMessage, OperationState.failure, res)
            );
            return observableThrowError(error);
          }
        )
      ));
  }

  createAccount(created: boolean): void {
    if (created) {
      this.retrieve();
    }
  }

  forceRefreshView(duration: number): void {
    // Reset timer
    if (this.timerHandler) {
      clearInterval(this.timerHandler);
    }
    this.timerHandler = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => {
      if (this.timerHandler) {
        clearInterval(this.timerHandler);
        this.timerHandler = null;
      }
    }, duration);
  }

  doSearch(value: string): void {
    this.searchRobot = value;
    this.retrieve();
  }

  retrieve(): void {
    this.loading = true;
    this.selectedRow = [];
    this.robotService
      .listRobotAccount(this.projectId)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe(
        response => {
          this.robots = response.filter(x =>
            x.name.split('$')[1].includes(this.searchRobot)
          );
          this.robotsCopy = response.map(x => Object.assign({}, x));
          this.forceRefreshView(2000);
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  changeAccountStatus(robots: Robot): void {
    let id: number | string = robots[0].id;
    this.isDisabled = robots[0].disabled ? false : true;
    this.robotService
      .toggleDisabledAccount(this.projectId, id, this.isDisabled)
      .subscribe(response => {
        this.retrieve();
      });
  }
}
