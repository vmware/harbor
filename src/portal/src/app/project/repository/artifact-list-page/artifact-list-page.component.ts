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
import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ArtifactListComponent } from "./artifact-list/artifact-list.component";
import { ArtifactDefaultService } from "../artifact/artifact.service";
import { AppConfigService } from "../../../services/app-config.service";
import { SessionService } from "../../../shared/session.service";
import { ArtifactClickEvent } from "../../../../lib/services";
import { Project } from "../../project";

@Component({
  selector: 'artifact-list-page',
  templateUrl: 'artifact-list-page.component.html',
  styleUrls: ['./artifact-list-page.component.scss']
})
export class ArtifactListPageComponent implements OnInit {

  projectId: number;
  projectName: string;
  projectMemberRoleId: number;
  repoName: string;
  referArtifactNameArray: string[] = [];
  hasProjectAdminRole: boolean = false;
  isGuest: boolean;
  registryUrl: string;

  @ViewChild(ArtifactListComponent, {static: false})
  repositoryComponent: ArtifactListComponent;
  depth: string;
  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private artifactService: ArtifactDefaultService,
    private appConfigService: AppConfigService,
    private session: SessionService) {
    this.route.params.subscribe(params => {
      this.depth = this.route.snapshot.params['depth'];
      if (this.depth) {
        const arr: string[] = this.depth.split('-');
        this.referArtifactNameArray = arr.slice(0, arr.length - 1);
      }
    });
  }

  ngOnInit() {
    this.projectId = this.route.snapshot.params['id'];
    if (!this.projectId) {
      this.projectId = this.route.snapshot.parent.params['id'];
    }
    let resolverData = this.route.snapshot.data;
    if (resolverData) {
      this.projectName = (<Project>resolverData['projectResolver']).name;
      this.hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
      this.isGuest = (<Project>resolverData['projectResolver']).current_user_role_id === 3;
      this.projectMemberRoleId = (<Project>resolverData['projectResolver']).current_user_role_id;
    }
    this.repoName = this.route.snapshot.params['repo'];
    this.registryUrl = this.appConfigService.getConfig().registry_url;
  }

  get withNotary(): boolean {
    return this.appConfigService.getConfig().with_notary;
  }
  get withAdmiral(): boolean {
    return this.appConfigService.getConfig().with_admiral;
  }

  get hasSignedIn(): boolean {
    return this.session.getCurrentUser() !== null;
  }

  hasChanges(): boolean {
    return this.repositoryComponent.hasChanges();
  }
  watchGoBackEvt(projectId: string| number): void {
    this.router.navigate(["harbor", "projects", projectId, "repositories"]);
  }
  goProBack(): void {
    this.router.navigate(["harbor", "projects"]);
  }
  backInitRepo() {
    this.router.navigate(["harbor", "projects", this.projectId, "repositories", this.repoName]);
  }
  jumpDigest(index: number) {
    const arr: string[] = this.referArtifactNameArray.slice(0, index + 1 );
    if ( arr && arr.length) {
      this.router.navigate(["harbor", "projects", this.projectId, "repositories", this.repoName, "depth", arr.join('-')]);
    } else {
      this.router.navigate(["harbor", "projects", this.projectId, "repositories", this.repoName]);
    }
  }
}
