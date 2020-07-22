
import { throwError as observableThrowError, Observable, of } from "rxjs";
import {Injectable, Inject} from "@angular/core";
import { HttpClient, HttpParams, HttpResponse } from "@angular/common/http";
import { catchError } from "rxjs/operators";

import { SERVICE_CONFIG, IServiceConfig } from "../entities/service.config";
import { Project } from "../components/project-policy-config/project";
import { ProjectPolicy } from "../components/project-policy-config/project-policy-config.component";
import {
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS,
  buildHttpRequestOptionsWithObserveResponse, CURRENT_BASE_HREF
} from "../utils/utils";

/**
 * Define the service methods to handle the Project related things.
 *
 **
 * @abstract
 * class ProjectService
 */
export abstract class ProjectService {
  /**
   * Get Informations about a specific Project.
   *
   * @abstract
   *  ** deprecated param {string|number} [projectId]
   * returns {(Observable<Project> )}
   *
   * @memberOf ProjectService
   */
  abstract getProject(
    projectId: number | string
  ): Observable<Project> ;

  /**
   * Update the specified project.
   *
   * @abstract
   *  ** deprecated param {(number | string)} projectId
   *  ** deprecated param {ProjectPolicy} projectPolicy
   * returns {(Observable<any>)}
   *
   * @memberOf EndpointService
   */
  abstract updateProjectPolicy(
    projectId: number | string,
    projectPolicy: ProjectPolicy,
    reuseSysCVEVAllowlist: string,
    projectAllowlist: object
  ): Observable<any>;

  /**
   * Get all projects
   *
   * @abstract
   *  ** deprecated param {string} name
   *  ** deprecated param {number} isPublic
   *  ** deprecated param {number} page
   *  ** deprecated param {number} pageSize
   * returns {(Observable<any>)}
   *
   * @memberOf EndpointService
   */
  abstract listProjects(
    name: string,
    isPublic?: number,
    page?: number,
    pageSize?: number
  ): Observable<HttpResponse<Project[]>>;
  abstract createProject(name: string, metadata: any, storageLimit: number, registryId: number): Observable<any>;
  abstract deleteProject(projectId: number): Observable<any>;
  abstract checkProjectExists(projectName: string): Observable<any>;
  abstract checkProjectMember(projectId: number): Observable<any>;
  abstract getProjectSummary(projectId: number): Observable<any>;
}

/**
 * Implement default service for project.
 *
 **
 * class ProjectDefaultService
 * extends {ProjectService}
 */
@Injectable()
export class ProjectDefaultService extends ProjectService {
  constructor(
    private http: HttpClient,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();
  }

  public getProject(
    projectId: number | string
  ): Observable<Project> {
    if (!projectId) {
      return observableThrowError("Bad argument");
    }
    let baseUrl: string = this.config.projectBaseEndpoint
      ? this.config.projectBaseEndpoint
      : CURRENT_BASE_HREF + "/projects";
    return this.http
      .get<Project>(`${baseUrl}/${projectId}`, HTTP_GET_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public updateProjectPolicy(
    projectId: number | string,
    projectPolicy: ProjectPolicy,
    reuseSysCVEVAllowlist: string,
    projectAllowlist: object
  ): any {
    let baseUrl: string = this.config.projectBaseEndpoint
      ? this.config.projectBaseEndpoint
      : CURRENT_BASE_HREF + "/projects";
    return this.http
      .put<any>(
        `${baseUrl}/${projectId}`,
        {
          metadata: {
              public: projectPolicy.Public ? "true" : "false",
              enable_content_trust: projectPolicy.ContentTrust ? "true" : "false",
              prevent_vul: projectPolicy.PreventVulImg ? "true" : "false",
              severity: projectPolicy.PreventVulImgSeverity,
              auto_scan: projectPolicy.ScanImgOnPush ? "true" : "false",
              reuse_sys_cve_allowlist: reuseSysCVEVAllowlist
          },
            cve_allowlist: projectAllowlist
        },
        HTTP_JSON_OPTIONS
      )
      .pipe(catchError(error => observableThrowError(error)));
  }
  public listProjects(name: string, isPublic?: number, page?: number, pageSize?: number): Observable<HttpResponse<Project[]>> {
    let params = new HttpParams();
    if (page && pageSize) {
      params = params.set('page', page + '').set('page_size', pageSize + '');
    }
    if (name && name.trim() !== "") {
      params = params.set('name', name);
    }
    if (isPublic !== undefined) {
      params = params.set('public', '' + isPublic);
    }
    return this.http
               .get<HttpResponse<Project[]>>(`${ CURRENT_BASE_HREF }/projects`, buildHttpRequestOptionsWithObserveResponse(params)).pipe(
               catchError(error => observableThrowError(error)), );
  }

  public createProject(name: string, metadata: any, storageLimit: number, registryId: number): Observable<any> {
    return this.http
               .post(`${ CURRENT_BASE_HREF }/projects`,
                 JSON.stringify({
                   'project_name': name, registry_id: +registryId, 'metadata': {
                     public: metadata.public ? 'true' : 'false'
                   },
                storage_limit: storageLimit
              })
                , HTTP_JSON_OPTIONS).pipe(
               catchError(error => observableThrowError(error)), );
  }

  public deleteProject(projectId: number): Observable<any> {
    return this.http
               .delete(`${ CURRENT_BASE_HREF }/projects/${projectId}`)
               .pipe(catchError(error => observableThrowError(error)));
  }

  public checkProjectExists(projectName: string): Observable<any> {
    return this.http
        .head(`${ CURRENT_BASE_HREF }/projects/?project_name=${projectName}`).pipe(
            catchError(error => {
              if (error && error.status === 404) {
                return of(error);
              }
              return observableThrowError(error);
            }));
  }

  public checkProjectMember(projectId: number): Observable<any> {
    return this.http
               .get(`${ CURRENT_BASE_HREF }/projects/${projectId}/members`, HTTP_GET_OPTIONS).pipe(
               catchError(error => observableThrowError(error)), );
  }
  public getProjectSummary(projectId: number): Observable<any> {
    return this.http
               .get(`${ CURRENT_BASE_HREF }/projects/${projectId}/summary`, HTTP_GET_OPTIONS).pipe(
               catchError(error => observableThrowError(error)), );
  }
}
