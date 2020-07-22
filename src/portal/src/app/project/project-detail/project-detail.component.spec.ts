import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ProjectDetailComponent } from './project-detail.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { of } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { AppConfigService } from "../../services/app-config.service";
import { ProjectService, UserPermissionService } from "../../../lib/services";
import { ErrorHandler } from "../../../lib/utils/error-handler";

describe('ProjectDetailComponent', () => {
    let component: ProjectDetailComponent;
    let fixture: ComponentFixture<ProjectDetailComponent>;
    const mockSessionService = {
        getCurrentUser: () => {
            return of({
                user_id: 1
            });
        }
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                with_admiral: true,
                with_chartmuseum: true,
            };
        },

    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        }
    };
    const mockProjectService = null;
    const mockErrorHandler = {
        error() { }
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: (key) => 'value' }),
        snapshot: {
            params: { id: 1 },
            data: 1,
            children: [
                {
                    routeConfig:
                        { path: "" }
                }
            ]
        },
        data: of({
            projectResolver: {
                ismember: true,
                role_name: 'maintainer',
            }
        })
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
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [ProjectDetailComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: UserPermissionService, useValue: mockUserPermissionService },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: ProjectService, useValue: mockProjectService },
                {
                    provide: ActivatedRoute, useValue: mockActivatedRoute
                }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectDetailComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
