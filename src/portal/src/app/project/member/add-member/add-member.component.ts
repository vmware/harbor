// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import {finalize, debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {
  Component,
  Input,
  EventEmitter,
  Output,
  ViewChild,
  AfterViewChecked,
  OnInit,
  OnDestroy, ChangeDetectionStrategy, ChangeDetectorRef
} from '@angular/core';
import { NgForm } from '@angular/forms';
import {ActivatedRoute} from "@angular/router";
import { Subject, forkJoin } from "rxjs";



import { TranslateService } from '@ngx-translate/core';

import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';
import { UserService } from '../../../user/user.service';
import {User} from "../../../user/user";
import {Project} from "../../project";
import { Member } from '../member';
import { MemberService } from '../member.service';
import { ErrorHandler } from '../../../../lib/utils/error-handler';


@Component({
  selector: 'add-member',
  templateUrl: 'add-member.component.html',
  styleUrls: ['add-member.component.scss'],
  providers: [UserService],
  changeDetection: ChangeDetectionStrategy.Default
})
export class AddMemberComponent implements AfterViewChecked, OnInit, OnDestroy {

  @Input() memberList: Member[] = [];
  member: Member = new Member();

  addMemberOpened: boolean;

  memberForm: NgForm;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @ViewChild('memberForm', {static: true})
  currentForm: NgForm;

  hasChanged: boolean;

  @ViewChild(InlineAlertComponent, {static: false})
  inlineAlert: InlineAlertComponent;

  @Input() projectId: number;
  @Output() added = new EventEmitter<boolean>();

  isMemberNameValid: boolean = true;
  memberTooltip: string = 'MEMBER.USERNAME_IS_REQUIRED';
  nameChecker: Subject<string> = new Subject<string>();
  checkOnGoing: boolean = false;
  selectUserName: string[] = [];
  userLists: User[];

  constructor(private memberService: MemberService,
    private userService: UserService,
    private errorHandle: ErrorHandler,
    private messageHandlerService: MessageHandlerService,
    private translateService: TranslateService,
    private route: ActivatedRoute,
    private ref: ChangeDetectorRef) { }

  ngOnInit(): void {
    let resolverData = this.route.snapshot.parent.data;
    let hasProjectAdminRole: boolean;
    if (resolverData) {
      hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
    }
    if (hasProjectAdminRole) {
      this.nameChecker.pipe(
        debounceTime(500),
        distinctUntilChanged(), )
        .subscribe((name: string) => {
          let cont = this.currentForm.controls['member_name'];
          if (cont) {
            this.isMemberNameValid = cont.valid;
            if (cont.valid) {
              this.checkOnGoing = true;
              this.ref.detectChanges();
              forkJoin(this.userService.getUsersNameList(cont.value, 20), this.memberService
              .listMembers(this.projectId, cont.value)).subscribe((res: Array<any>) => {
                this.userLists = res[0];
                if (res[1].filter(m => { return m.entity_name === cont.value; }).length > 0) {
                  this.isMemberNameValid = false;
                  this.memberTooltip = 'MEMBER.USERNAME_ALREADY_EXISTS';
                }
                this.checkOnGoing = false;
                if (this.userLists && this.userLists.length) {
                  this.selectUserName = [];
                  this.userLists.forEach(data => {
                    if (data.username.startsWith(cont.value) && !this.memberList.find(mem => mem.entity_name === data.username)) {
                      if (this.selectUserName.length < 10) {
                        this.selectUserName.push(data.username);
                      }
                    }
                  });
                }
                  let changeTimer = setInterval(() => this.ref.detectChanges(), 200);
                  setTimeout(() => {
                    clearInterval(changeTimer);
                  }, 2000);
              }, error => {
                this.checkOnGoing = false;
                this.ref.detectChanges();
              });
            } else {
              this.memberTooltip = 'MEMBER.USERNAME_IS_REQUIRED';
            }
          }
        });
    }
  }

  ngOnDestroy(): void {
    this.nameChecker.unsubscribe();
  }

  onSubmit(): void {
    if (!this.member.entity_name || this.member.entity_name.length === 0) { return; }
    this.memberService
      .addUserMember(this.projectId, {username: this.member.entity_name}, +this.member.role_id).pipe(
      finalize(() => {
        this.addMemberOpened = false;
        this.member.role_id = null;

        let changeTimer = setInterval(() => this.ref.detectChanges(), 200);
        setTimeout(() => {
          clearInterval(changeTimer);
        }, 2000);
      }
    ))
      .subscribe(
      () => {
        this.messageHandlerService.showSuccess('MEMBER.ADDED_SUCCESS');
        this.added.emit(true);
      },
      error => {
        this.errorHandle.error(error);
      });
  }

  selectedName(username: string) {
    this.member.entity_name = username;
    this.selectUserName = [];
  }

  onCancel() {
      this.addMemberOpened = false;
      this.memberForm.reset();
  }

  leaveInput() {
    this.selectUserName = [];
  }
  ngAfterViewChecked(): void {
    if (this.memberForm !== this.currentForm) {
      this.memberForm = this.currentForm;
    }
    if (this.memberForm) {
      this.memberForm.valueChanges.subscribe(data => {
        let memberName = data['member_name'];
        if (memberName && memberName !== '') {
          this.hasChanged = true;
        } else {
          this.hasChanged = false;
        }
      });
    }
  }

  openAddMemberModal(): void {
    this.currentForm.reset();
    this.member = new Member();
    this.addMemberOpened = true;
    this.hasChanged = false;
    this.member.role_id = 1;
    this.member.entity_name = '';
    this.isMemberNameValid = true;
    this.memberTooltip = 'MEMBER.USERNAME_IS_REQUIRED';
    this.selectUserName = [];
  }

  handleValidation(): void {
    let cont = this.currentForm.controls['member_name'];
    if (cont) {
      this.nameChecker.next(cont.value);
    }
  }

  public get isValid(): boolean {
    return this.currentForm &&
    this.currentForm.valid &&
    this.isMemberNameValid &&
    !this.checkOnGoing;
  }
}
