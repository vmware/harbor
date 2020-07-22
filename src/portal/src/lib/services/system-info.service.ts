import {ElementRef, Inject, Injectable} from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import {SystemCVEAllowlist, SystemInfo} from './interface';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';
import { CURRENT_BASE_HREF, HTTP_GET_OPTIONS, HTTP_JSON_OPTIONS } from "../utils/utils";

/**
 * Get System information about current backend server.
 * @abstract
 * class
 */
export abstract class SystemInfoService {
  /**
   *  Get global system information.
   *  @abstract
   *  returns
   */
  abstract getSystemInfo(): Observable<SystemInfo>;
  /**
   *  get system CEVAllowlist
   */
  abstract getSystemAllowlist(): Observable<SystemCVEAllowlist>;
  /**
   *  update systemCVEAllowlist
   * @param systemCVEAllowlist
   */
  abstract updateSystemAllowlist(systemCVEAllowlist: SystemCVEAllowlist): Observable<any>;
  /**
   *  set null to the date type input
   * @param ref
   */
  abstract resetDateInput(ref: ElementRef);
}

@Injectable()
export class SystemInfoDefaultService extends SystemInfoService {
  constructor(
    @Inject(SERVICE_CONFIG) private config: IServiceConfig,
    private http: HttpClient) {
    super();
  }
  getSystemInfo(): Observable<SystemInfo> {
    let url = this.config.systemInfoEndpoint ? this.config.systemInfoEndpoint : CURRENT_BASE_HREF + '/systeminfo';
    return this.http.get(url, HTTP_GET_OPTIONS)
      .pipe(map(systemInfo => systemInfo as SystemInfo)
      , catchError(error => observableThrowError(error)));
  }
  public getSystemAllowlist(): Observable<SystemCVEAllowlist> {
    return this.http.get(CURRENT_BASE_HREF + "/system/CVEAllowlist", HTTP_GET_OPTIONS)
        .pipe(map(systemCVEAllowlist => systemCVEAllowlist as SystemCVEAllowlist)
            , catchError(error => observableThrowError(error)));
  }
  public updateSystemAllowlist(systemCVEAllowlist: SystemCVEAllowlist): Observable<any> {
    return this.http.put(CURRENT_BASE_HREF + "/system/CVEAllowlist", JSON.stringify(systemCVEAllowlist), HTTP_JSON_OPTIONS)
        .pipe(map(response => response)
            , catchError(error => observableThrowError(error)));
  }
  public resetDateInput(ref: ElementRef) {
    if (ref) {
      ref.nativeElement.value = null ;
    }
  }
}

