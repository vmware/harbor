
import {filter} from 'rxjs/operators';
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
import { Component, Input, Output, OnDestroy, EventEmitter, ChangeDetectionStrategy, ChangeDetectorRef, OnInit } from '@angular/core';
import { Router, NavigationEnd } from '@angular/router';
import { State } from '../../../lib/services/interface';
import { Repository } from '../../../../ng-swagger-gen/models/repository';

import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import {Subscription} from "rxjs";

@Component({
  selector: 'list-repository-ro',
  templateUrl: 'list-repository-ro.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListRepositoryROComponent implements OnInit, OnDestroy {

  @Input() projectId: number;
  @Input() repositories: Repository[];

  @Output() paginate = new EventEmitter<State>();

  routerSubscription: Subscription;

  constructor(
    private router: Router,
    private searchTrigger: SearchTriggerService,
    private ref: ChangeDetectorRef) {
    this.router.routeReuseStrategy.shouldReuseRoute = function() {
        return false;
    };
    this.routerSubscription = this.router.events.pipe(filter(event => event instanceof NavigationEnd))
        .subscribe((event) => {
         // trick the Router into believing it's last link wasn't previously loaded
         this.router.navigated = false;
         // if you need to scroll back to top, here is the right place
         window.scrollTo(0, 0);
    });
  }

  ngOnInit(): void {
    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 1000);
  }

  ngOnDestroy(): void {
    this.routerSubscription.unsubscribe();
  }

  refresh(state: State) {
    if (this.repositories) {
      this.paginate.emit(state);
    }
  }

  public gotoLink(projectId: number, repoName: string): void {
    this.searchTrigger.closeSearch(true);
    let projectName = repoName.split('/')[0];
    let repositorieName = projectName ? repoName.substr(projectName.length + 1) : repoName;
    let linkUrl = ['harbor', 'projects', projectId, 'repositories', repositorieName ];
    this.router.navigate(linkUrl);
  }

}
