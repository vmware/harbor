import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ArtifactListPageComponent } from './artifact-list-page.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from "../../../shared/session.service";
import { AppConfigService } from "../../../services/app-config.service";
import { ArtifactService } from "../../../../../ng-swagger-gen/services/artifact.service";
import { ArtifactDefaultService } from "../artifact/artifact.service";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../lib/entities/service.config";
import { CURRENT_BASE_HREF } from "../../../../lib/utils/utils";

describe('ArtifactListPageComponent', () => {
    let component: ArtifactListPageComponent;
    let fixture: ComponentFixture<ArtifactListPageComponent>;
    const mockSessionService = {
        getCurrentUser: () => { }
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                project_creation_restriction: "",
                with_chartmuseum: "",
                with_notary: "",
                with_clair: "",
                with_admiral: "",
                registry_url: "",
            };
        }
    };
    const mockRouter = {
        navigate: () => { }
    };
    const mockArtifactService = {
        triggerUploadArtifact: {
            next: () => {}
        }
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: (key) => 'value' }),
        snapshot: {
            params: {
                id: 1,
            },
            parent: {
                params: { id: 1 },

            },
            data: {
                projectResolver: {
                    has_project_admin_role: true,
                    current_user_role_id: 3,
                }
            }
        },
        data: of({
            projectResolver: {
                ismember: true,
                role_name: 'maintainer',
            }
        }),
        params: {
            subscribe: () => {
                return of(null);
            }
        }
    };
    const config: IServiceConfig = {
        repositoryBaseEndpoint: CURRENT_BASE_HREF + "/repositories/testing"
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [ArtifactListPageComponent],
            providers: [
                TranslateService,
                ArtifactDefaultService,
                { provide: SERVICE_CONFIG, useValue: config },
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: Router, useValue: mockRouter },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: ArtifactService, useValue: mockArtifactService },
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactListPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
