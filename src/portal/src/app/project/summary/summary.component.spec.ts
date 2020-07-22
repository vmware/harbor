import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { AppConfigService } from "../../services/app-config.service";
import { SummaryComponent } from './summary.component';
import { EndpointDefaultService, EndpointService, ProjectService, UserPermissionService } from '../../../lib/services';
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { IServiceConfig, SERVICE_CONFIG } from '../../../lib/entities/service.config';
import { CURRENT_BASE_HREF } from '../../../lib/utils/utils';
import { SessionService } from '../../shared/session.service';


describe('SummaryComponent', () => {
  let component: SummaryComponent;
  let fixture: ComponentFixture<SummaryComponent>;
  let fakeAppConfigService = {
    getConfig() {
      return {
        with_chartmuseum: false
      };
    }
  };
  let fakeProjectService = {
    getProjectSummary: function () {
      return of();
    }
  };
  let fakeErrorHandler = null;
  let fakeUserPermissionService = {
    hasProjectPermissions: function() {
      return of([true, true]);
    }
  };
  const config: IServiceConfig = {
    systemInfoEndpoint: CURRENT_BASE_HREF + "/endpoints/testing"
  };

  const fakedSessionService = {
    getCurrentUser() {
      return {
        has_admin_role: true
      };
    }
  };

  const fakedEndpointService = {
    getEndpoint() {
      return of({
        name: "test",
        url: "https://test.com"
      });
    }
  };

  const mockedSummaryInformation = {
    repo_count: 0,
    chart_count: 0,
    project_admin_count: 1,
    maintainer_count: 0,
    developer_count: 0,
    registry: {
      name: "test",
      url: "https://test.com"
    }
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [SummaryComponent],
      imports: [
        ClarityModule,
        TranslateModule.forRoot()
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      providers: [
        TranslateService,
        { provide: AppConfigService, useValue: fakeAppConfigService },
        { provide: ProjectService, useValue: fakeProjectService },
        { provide: ErrorHandler, useValue: fakeErrorHandler },
        { provide: UserPermissionService, useValue: fakeUserPermissionService },
        { provide: EndpointService, useValue: fakedEndpointService },
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: SessionService, useValue:  fakedSessionService},
        {
          provide: ActivatedRoute, useValue: {
            paramMap: of({ get: (key) => 'value' }),
            snapshot: {
              parent: {
                params: { id: 1 },
                data: {
                  projectResolver: {registry_id: 3}
                }
              },
            }
          }
        },
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(SummaryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should show proxy cache endpoint', async () => {
    component.summaryInformation = mockedSummaryInformation;
    fixture.detectChanges();
    await fixture.whenStable();
    const endpoint: HTMLElement = fixture.nativeElement.querySelector("#endpoint");
    expect(endpoint).toBeTruthy();
    expect(endpoint.innerText).toEqual("test-https://test.com");
  });
});
