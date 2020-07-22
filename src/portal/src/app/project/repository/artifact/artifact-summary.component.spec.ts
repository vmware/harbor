import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactSummaryComponent } from "./artifact-summary.component";
import { of } from "rxjs";
import { ClarityModule } from "@clr/angular";
import { NO_ERRORS_SCHEMA } from "@angular/core";
import { Artifact } from "../../../../../ng-swagger-gen/models/artifact";
import { ProjectService } from "../../../../lib/services";
import { ArtifactService } from "../../../../../ng-swagger-gen/services/artifact.service";
import { ErrorHandler } from "../../../../lib/utils/error-handler";
import { TranslateFakeLoader, TranslateLoader, TranslateModule } from "@ngx-translate/core";
import { ActivatedRoute, Router } from "@angular/router";
import { AppConfigService } from "../../../services/app-config.service";
import { Project } from "../../project";
import { AllPipesModule } from "../../../all-pipes/all-pipes.module";

describe('ArtifactSummaryComponent', () => {

  const mockedArtifact: Artifact = {
    id: 123,
    type: 'IMAGE'
  };

  const fakedProjectService = {
    getProject() {
      return of({name: 'test'});
    }
  };

  const fakedArtifactService = {
    getArtifact() {
       return of(mockedArtifact);
    }
  };
  let component: ArtifactSummaryComponent;
  let fixture: ComponentFixture<ArtifactSummaryComponent>;
  const mockActivatedRoute = {
    RouterparamMap: of({ get: (key) => 'value' }),
    snapshot: {
      params: {
        id: 1,
        repo: "test",
        digest: "ABC",
        subscribe: () => {
          return of(null);
        }
      },
      data: {
        artifactResolver: [mockedArtifact, new Project()]
      }
    },
    data: of({
      projectResolver: {
        ismember: true,
        role_name: 'maintainer',
      }
    })
  };
  const fakedAppConfigService = {
    getConfig: () => {
      return {with_admiral: false};
    }
  };
  const mockRouter = {
    navigate: () => { }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ClarityModule,
        AllPipesModule,
        TranslateModule.forRoot({
          loader: {
            provide: TranslateLoader,
            useClass: TranslateFakeLoader,
          }
        })
      ],
      declarations: [
        ArtifactSummaryComponent
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ],
      providers: [
        { provide: AppConfigService, useValue: fakedAppConfigService },
        { provide: Router, useValue: mockRouter },
        { provide: ActivatedRoute, useValue: mockActivatedRoute },
        { provide: ProjectService, useValue: fakedProjectService },
        { provide: ArtifactService, useValue: fakedArtifactService },
        ErrorHandler
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactSummaryComponent);
    component = fixture.componentInstance;
    component.repositoryName = 'demo';
    component.artifactDigest = 'sha: acf4234f';
    fixture.detectChanges();
  });

  it('should create and get artifactDetails', async () => {
    expect(component).toBeTruthy();
    await fixture.whenStable();
    expect(component.artifact.type).toEqual('IMAGE');
  });
});
