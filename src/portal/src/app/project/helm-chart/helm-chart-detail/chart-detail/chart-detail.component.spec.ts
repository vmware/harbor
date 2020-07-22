import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ChartDetailComponent } from './chart-detail.component';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { HelmChartService } from "../../helm-chart.service";
import { of } from 'rxjs';
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { SystemInfoService } from "../../../../../lib/services";
describe('ChartDetailComponent', () => {
    let component: ChartDetailComponent;
    let fixture: ComponentFixture<ChartDetailComponent>;
    const mockErrorHandler = {
        error: function () { }
    };
    const mockSystemInfoService = {
        getSystemInfo: function () {
            return of(
                {
                    "with_notary": false,
                    "with_admiral": false,
                    "admiral_endpoint": "",
                    "auth_mode": "oidc_auth",
                    "registry_url": "nightly-oidc.harbor.io",
                    "external_url": "https://nightly-oidc.harbor.io",
                    "project_creation_restriction": "everyone",
                    "self_registration": false,
                    "has_ca_root": false,
                    "harbor_version": "dev",
                    "registry_storage_provider_name": "filesystem",
                    "read_only": false,
                    "with_chartmuseum": true,
                    "notification_enable": true
                }
            );
        }
    };
    const mockHelmChartService = {
        getChartDetail: function () {
            return of(
                {
                    "metadata": {
                      "name": "harbor",
                      "home": "https://github.com/vmware/harbor",
                      "sources": [
                        "https://github.com/vmware/harbor/tree/master/contrib/helm/harbor"
                      ],
                      "version": "0.2.0",
                      "description": "Ane",
                      "keywords": [
                        "vmware",
                        "docker",
                        "registry",
                        "harbor"
                      ],
                      "maintainers": [
                        {
                          "name": "Jessde Hu",
                          "email": "huh@qq.com"
                        },
                        {
                          "name": "paulczar",
                          "email": "username@qq.com"
                        }
                      ],
                      "engine": "",
                      "icon": "ht",
                      "appVersion": "1.5.0",
                      "urls": [
                        ""
                      ],
                      "created": "201940492141Z",
                      "digest": ""
                    },
                    "dependencies": [
                      {
                        "name": "redis",
                        "version": "3.2.5",
                        "repository": ""
                      }
                    ],
                    "values": {
                      "adminserver.image.pullPolicy": "IfNotPresent"
                    },
                    "files": {
                      "README.md": "",
                      "values.yaml": ""
                    },
                    "security": {
                      "signature": {
                        "signed": false,
                        "prov_file": ""
                      }
                    },
                    "labels": []
                  }
            );
        },
        downloadChart: function () { }
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                ClarityModule,
                FormsModule
            ],
            declarations: [ChartDetailComponent],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                TranslateService,
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: SystemInfoService, useValue: mockSystemInfoService },
                { provide: HelmChartService, useValue: mockHelmChartService },
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailComponent);
        component = fixture.componentInstance;
        component.chartName = 'chart';
        component.chartVersion = 'chart-version';
        component.project = {
            "project_id": 1,
            "owner_id": 1,
            "name": "library",
            "creation_time": new Date(),
            "creation_time_str": "123",
            "update_time": new Date(),
            "deleted": 1,
            "owner_name": "",
            "togglable": true,
            "current_user_role_id": 1,
            "has_project_admin_role": true,
            "is_member": true,
            "role_name": 'maintainer',
            "repo_count": 0,
            "chart_count": 1,
            "registry_id" : 0,
            "metadata": {
                "public": "true",
                "enable_content_trust": "string",
                "prevent_vul": "string",
                "severity": 'string',
                "auto_scan": true,
                "retention_id": 1
            }
        };
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
