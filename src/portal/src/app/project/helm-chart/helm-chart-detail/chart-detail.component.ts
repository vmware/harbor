import { ActivatedRoute, Router } from '@angular/router';
import { Component, OnInit } from "@angular/core";
import { Project } from '../../project';
import { SessionService } from './../../../shared/session.service';
import { SessionUser } from './../../../shared/session-user';

@Component({
  selector: "project-chart-detail",
  templateUrl: "./chart-detail.component.html",
  styleUrls: ["./chart-detail.component.scss"]
})
export class HelmChartDetailComponent implements OnInit {

  projectId: number | string;
  project: Project;
  projectName: string;
  chartName: string;
  chartVersion: string;
  currentUser: SessionUser;
  hasProjectAdminRole: boolean;
  roleName: string;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private session: SessionService
  ) {}

  ngOnInit() {
    // Get projectId from route params snapshot.
    this.projectId = +this.route.snapshot.params['id'];
    this.chartName = this.route.snapshot.params['chart'];
    this.chartVersion = this.route.snapshot.params['version'];
    // Get current user from registered resolver.
    this.currentUser = this.session.getCurrentUser();
    let resolverData = this.route.snapshot.data;
    if (resolverData) {
      this.project = <Project>(resolverData["projectResolver"]);
      this.roleName = this.project.role_name;
      this.projectName = this.project.name;
      this.hasProjectAdminRole = this.project.has_project_admin_role;
    }
  }

  gotoProjectList() {
    this.router.navigateByUrl("/harbor/projects");
  }

  gotoChartList() {
    this.router.navigateByUrl(`/harbor/projects/${this.projectId}/helm-charts`);
  }

  gotoChartVersion() {
    this.router.navigateByUrl(`/harbor/projects/${this.projectId}/helm-charts/${this.chartName}/versions`);
  }
}
