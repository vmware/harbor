import { Injectable, Inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Observable, throwError as observableThrowError } from "rxjs";

import { IServiceConfig, SERVICE_CONFIG } from "../entities/service.config";
import {
  buildHttpRequestOptions,
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS, CURRENT_BASE_HREF
} from "../utils/utils";
import { RequestQueryParams } from "./RequestQueryParams";
import { Endpoint, ReplicationRule, PingEndpoint } from "./interface";
import { catchError, map } from "rxjs/operators";

const ADAPTERS_MAP = {
  "ali-acr": "Alibaba ACR",
  "aws-ecr": "Aws ECR",
  "azure-acr": "Azure ACR",
  "docker-hub": "Docker Hub",
  "docker-registry": "Docker Registry",
  "gitlab": "Gitlab",
  "google-gcr": "Google GCR",
  "harbor": "Harbor",
  "helm-hub": "Helm Hub",
  "huawei-SWR": "Huawei SWR",
  "jfrog-artifactory": "jFrog Artifactory",
  "quay-io": "Quay.io"
};



/**
 * Define the service methods to handle the endpoint related things.
 *
 **
 * @abstract
 * class EndpointService
 */
export abstract class EndpointService {
  /**
   * Get all the endpoints.
   * Set the argument 'endpointName' to return only the endpoints match the name pattern.
   *
   * @abstract
   *  ** deprecated param {string} [endpointName]
   *  ** deprecated param {RequestQueryParams} [queryParams]
   * returns {(Observable<Endpoint[]> | Endpoint[])}
   *
   * @memberOf EndpointService
   */
  abstract getEndpoints(
    endpointName?: string,
    queryParams?: RequestQueryParams
  ): Observable<Endpoint[]>;

  /**
   * Get the specified endpoint.
   *
   * @abstract
   *  ** deprecated param {(number | string)} endpointId
   * returns {(Observable<Endpoint> | Endpoint)}
   *
   * @memberOf EndpointService
   */
  abstract getEndpoint(
    endpointId: number | string
  ): Observable<Endpoint>;

  /**
   * Create new endpoint.
   *
   * @abstract
   *  ** deprecated param {Endpoint} endpoint
   * returns {(Observable<any>)}
   *
   * @memberOf EndpointService
   */
  abstract getAdapters(): Observable<any>;

  /**
   * Create new endpoint.
   *
   * @abstract
   *  ** deprecated param {Adapter} adapter
   * returns {(Observable<any> | any)}
   *
   * @memberOf EndpointService
   */
  abstract createEndpoint(
    endpoint: Endpoint
  ): Observable<any>;

  /**
   * Update the specified endpoint.
   *
   * @abstract
   *  ** deprecated param {(number | string)} endpointId
   *  ** deprecated param {Endpoint} endpoint
   * returns {(Observable<any>)}
   *
   * @memberOf EndpointService
   */
  abstract updateEndpoint(
    endpointId: number | string,
    endpoint: Endpoint
  ): Observable<any>;

  /**
   * Delete the specified endpoint.
   *
   * @abstract
   *  ** deprecated param {(number | string)} endpointId
   * returns {(Observable<any>)}
   *
   * @memberOf EndpointService
   */
  abstract deleteEndpoint(
    endpointId: number | string
  ): Observable<any>;

  /**
   * Ping the specified endpoint.
   *
   * @abstract
   *  ** deprecated param {Endpoint} endpoint
   * returns {(Observable<any>)}
   *
   * @memberOf EndpointService
   */
  abstract pingEndpoint(
    endpoint: PingEndpoint
  ): Observable<any>;

  /**
   * Check endpoint whether in used with specific replication rule.
   *
   * @abstract
   *  ** deprecated param {{number | string}} endpointId
   * returns {{Observable<any>}}
   */
  abstract getEndpointWithReplicationRules(
    endpointId: number | string
  ): Observable<any>;

  abstract getAdapterText(adapter: string): string;
}

/**
 * Implement default service for endpoint.
 *
 **
 * class EndpointDefaultService
 * extends {EndpointService}
 */
@Injectable()
export class EndpointDefaultService extends EndpointService {
  _endpointUrl: string;

  constructor(
    @Inject(SERVICE_CONFIG) config: IServiceConfig,
    private http: HttpClient
  ) {
    super();
    this._endpointUrl = config.targetBaseEndpoint
      ? config.targetBaseEndpoint
      : CURRENT_BASE_HREF + "/registries";
  }

  public getEndpoints(
    endpointName?: string,
    queryParams?: RequestQueryParams
  ): Observable<Endpoint[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    if (endpointName) {
      queryParams = queryParams.set("name", endpointName);
    }
    let requestUrl: string = `${this._endpointUrl}`;
    return this.http
      .get(requestUrl, buildHttpRequestOptions(queryParams))
      .pipe(map(response => response as Endpoint[])
      , catchError(error => observableThrowError(error)));
  }

  public getEndpoint(
    endpointId: number | string
  ): Observable<Endpoint> {
    if (!endpointId || endpointId <= 0) {
      return observableThrowError("Bad request argument.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}`;
    return this.http
      .get(requestUrl, HTTP_GET_OPTIONS)
      .pipe(map(response => response as Endpoint)
      , catchError(error => observableThrowError(error)));
  }

  public getAdapters(): Observable<any> {
    return this.http
    .get(`${ CURRENT_BASE_HREF }/replication/adapters`)
    .pipe(catchError(error => observableThrowError(error)));
}

  public createEndpoint(
    endpoint: Endpoint
  ): Observable<any> {
    if (!endpoint) {
      return  observableThrowError("Invalid endpoint.");
    }
    let requestUrl: string = `${this._endpointUrl}`;
    return this.http
      .post<any>(requestUrl, JSON.stringify(endpoint), HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public updateEndpoint(
    endpointId: number | string,
    endpoint: Endpoint
  ): Observable<any> {
    if (!endpointId || endpointId <= 0) {
      return  observableThrowError("Bad request argument.");
    }
    if (!endpoint) {
      return  observableThrowError("Invalid endpoint.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}`;
    return this.http
      .put<any>(requestUrl, JSON.stringify(endpoint), HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public deleteEndpoint(
    endpointId: number | string
  ): Observable<any> {
    if (!endpointId || endpointId <= 0) {
      return  observableThrowError("Bad request argument.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}`;
    return this.http
      .delete<any>(requestUrl)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public pingEndpoint(
    endpoint: Endpoint
  ): Observable<any> {
    if (!endpoint) {
      return  observableThrowError("Invalid endpoint.");
    }
    let requestUrl: string = `${this._endpointUrl}/ping`;
    return this.http
      .post<any>(requestUrl, endpoint, HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public getEndpointWithReplicationRules(
    endpointId: number | string
  ): Observable<any> {
    if (!endpointId || endpointId <= 0) {
      return  observableThrowError("Bad request argument.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}/policies`;
    return this.http
      .get(requestUrl, HTTP_GET_OPTIONS)
      .pipe(map(response => response as ReplicationRule[])
      , catchError(error => observableThrowError(error)));
  }

  getAdapterText(adapter: string): string {
    if (ADAPTERS_MAP && ADAPTERS_MAP[adapter]) {
      return ADAPTERS_MAP[adapter];
    }
    return adapter;
  }
}
