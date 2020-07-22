import { ComponentFixture, TestBed, async } from "@angular/core/testing";
import { By } from "@angular/platform-browser";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { DebugElement } from "@angular/core";

import { SharedModule } from "../../utils/shared/shared.module";
import { EndpointComponent } from "./endpoint.component";
import { FilterComponent } from "../filter/filter.component";
import { ConfirmationDialogComponent } from "../confirmation-dialog/confirmation-dialog.component";
import { CreateEditEndpointComponent } from "../create-edit-endpoint/create-edit-endpoint.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ErrorHandler } from "../../utils/error-handler/error-handler";
import { Endpoint } from "../../services/interface";
import {
  EndpointService,
  EndpointDefaultService
} from "../../services/endpoint.service";
import { IServiceConfig, SERVICE_CONFIG } from "../../entities/service.config";
import { OperationService } from "../operation/operation.service";

import { click, CURRENT_BASE_HREF } from "../../utils/utils";
import { of } from "rxjs";
import { HttpClientTestingModule } from "@angular/common/http/testing";
import { HttpClient } from "@angular/common/http";
import { AppConfigService } from '../../../app/services/app-config.service';

describe("EndpointComponent (inline template)", () => {
  let adapterInfoMockData = {
    "ali-acr": {
      "endpoint_pattern": {
        "endpoint_type": "EndpointPatternTypeList",
        "endpoints": [
          {
            "key": "cn-hangzhou",
            "value": "https://registry.cn-hangzhou.aliyuncs.com"
          },
          {
            "key": "cn-shanghai",
            "value": "https://registry.cn-shanghai.aliyuncs.com"
          },
          {
            "key": "cn-qingdao",
            "value": "https://registry.cn-qingdao.aliyuncs.com"
          },
          {
            "key": "cn-beijing",
            "value": "https://registry.cn-beijing.aliyuncs.com"
          },
          {
            "key": "cn-zhangjiakou",
            "value": "https://registry.cn-zhangjiakou.aliyuncs.com"
          },
          {
            "key": "cn-huhehaote",
            "value": "https://registry.cn-huhehaote.aliyuncs.com"
          },
          {
            "key": "cn-shenzhen",
            "value": "https://registry.cn-shenzhen.aliyuncs.com"
          },
          {
            "key": "cn-chengdu",
            "value": "https://registry.cn-chengdu.aliyuncs.com"
          },
          {
            "key": "cn-hongkong",
            "value": "https://registry.cn-hongkong.aliyuncs.com"
          },
          {
            "key": "ap-southeast-1",
            "value": "https://registry.ap-southeast-1.aliyuncs.com"
          },
          {
            "key": "ap-southeast-2",
            "value": "https://registry.ap-southeast-2.aliyuncs.com"
          },
          {
            "key": "ap-southeast-3",
            "value": "https://registry.ap-southeast-3.aliyuncs.com"
          },
          {
            "key": "ap-southeast-5",
            "value": "https://registry.ap-southeast-5.aliyuncs.com"
          },
          {
            "key": "ap-northeast-1",
            "value": "https://registry.ap-northeast-1.aliyuncs.com"
          },
          {
            "key": "ap-south-1",
            "value": "https://registry.ap-south-1.aliyuncs.com"
          },
          {
            "key": "eu-central-1",
            "value": "https://registry.eu-central-1.aliyuncs.com"
          },
          {
            "key": "eu-west-1",
            "value": "https://registry.eu-west-1.aliyuncs.com"
          },
          {
            "key": "us-west-1",
            "value": "https://registry.us-west-1.aliyuncs.com"
          },
          {
            "key": "us-east-1",
            "value": "https://registry.us-east-1.aliyuncs.com"
          },
          {
            "key": "me-east-1",
            "value": "https://registry.me-east-1.aliyuncs.com"
          }
        ]
      },
      "credential_pattern": null
    },
    "aws-ecr": {
      "endpoint_pattern": {
        "endpoint_type": "EndpointPatternTypeList",
        "endpoints": [
          {
            "key": "ap-northeast-1",
            "value": "https://api.ecr.ap-northeast-1.amazonaws.com"
          },
          {
            "key": "us-east-1",
            "value": "https://api.ecr.us-east-1.amazonaws.com"
          },
          {
            "key": "us-east-2",
            "value": "https://api.ecr.us-east-2.amazonaws.com"
          },
          {
            "key": "us-west-1",
            "value": "https://api.ecr.us-west-1.amazonaws.com"
          },
          {
            "key": "us-west-2",
            "value": "https://api.ecr.us-west-2.amazonaws.com"
          },
          {
            "key": "ap-east-1",
            "value": "https://api.ecr.ap-east-1.amazonaws.com"
          },
          {
            "key": "ap-south-1",
            "value": "https://api.ecr.ap-south-1.amazonaws.com"
          },
          {
            "key": "ap-northeast-2",
            "value": "https://api.ecr.ap-northeast-2.amazonaws.com"
          },
          {
            "key": "ap-southeast-1",
            "value": "https://api.ecr.ap-southeast-1.amazonaws.com"
          },
          {
            "key": "ap-southeast-2",
            "value": "https://api.ecr.ap-southeast-2.amazonaws.com"
          },
          {
            "key": "ca-central-1",
            "value": "https://api.ecr.ca-central-1.amazonaws.com"
          },
          {
            "key": "eu-central-1",
            "value": "https://api.ecr.eu-central-1.amazonaws.com"
          },
          {
            "key": "eu-west-1",
            "value": "https://api.ecr.eu-west-1.amazonaws.com"
          },
          {
            "key": "eu-west-2",
            "value": "https://api.ecr.eu-west-2.amazonaws.com"
          },
          {
            "key": "eu-west-3",
            "value": "https://api.ecr.eu-west-3.amazonaws.com"
          },
          {
            "key": "eu-north-1",
            "value": "https://api.ecr.eu-north-1.amazonaws.com"
          },
          {
            "key": "sa-east-1",
            "value": "https://api.ecr.sa-east-1.amazonaws.com"
          }
        ]
      },
      "credential_pattern": null
    },
    "docker-hub": {
      "endpoint_pattern": {
        "endpoint_type": "EndpointPatternTypeFix",
        "endpoints": [
          {
            "key": "hub.docker.com",
            "value": "https://hub.docker.com"
          }
        ]
      },
      "credential_pattern": null
    },
    "google-gcr": {
      "endpoint_pattern": {
        "endpoint_type": "EndpointPatternTypeList",
        "endpoints": [
          {
            "key": "gcr.io",
            "value": "https://gcr.io"
          },
          {
            "key": "us.gcr.io",
            "value": "https://us.gcr.io"
          },
          {
            "key": "eu.gcr.io",
            "value": "https://eu.gcr.io"
          },
          {
            "key": "asia.gcr.io",
            "value": "https://asia.gcr.io"
          }
        ]
      },
      "credential_pattern": {
        "access_key_type": "AccessKeyTypeFix",
        "access_key_data": "_json_key",
        "access_secret_type": "AccessSecretTypeFile",
        "access_secret_data": "No Change"
      }
    }
  };
  let fakedHttp = {
    get() {
      return of(adapterInfoMockData);
    }
  };
  let mockData: Endpoint[] = [
    {
      id: 1,
      credential: {
        access_key: "admin",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_01",
      type: "Harbor",
      url: "https://10.117.4.151"
    },
    {
      id: 2,
      credential: {
        access_key: "AAA",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_02",
      type: "Harbor",
      url: "https://10.117.5.142"
    },
    {
      id: 3,
      credential: {
        access_key: "admin",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_03",
      type: "Harbor",
      url: "https://101.1.11.111"
    },
    {
      id: 4,
      credential: {
        access_key: "admin",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_04",
      type: "Harbor",
      url: "https://4.4.4.4"
    }
  ];
  let mockOne: Endpoint[] = [
    {
      id: 1,
      credential: {
        access_key: "admin",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_01",
      type: "Harbor",
      url: "https://10.117.4.151"
    }
  ];

  let mockAdapters = ['harbor', 'docker hub'];

  let comp: EndpointComponent;
  let fixture: ComponentFixture<EndpointComponent>;
  let config: IServiceConfig = {
    systemInfoEndpoint: CURRENT_BASE_HREF + "/endpoints/testing"
  };
  const mockAppConfigService = {
    getConfig: () => {
      return {
        project_creation_restriction: "",
        with_chartmuseum: ""
      };
    }
  };

  let endpointService: EndpointService;
  let spy: jasmine.Spy;
  let spyAdapter: jasmine.Spy;
  let spyOnRules: jasmine.Spy;
  let spyOne: jasmine.Spy;
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, NoopAnimationsModule, HttpClientTestingModule],
      declarations: [
        FilterComponent,
        ConfirmationDialogComponent,
        CreateEditEndpointComponent,
        InlineAlertComponent,
        EndpointComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: OperationService },
        { provide: HttpClient, useValue: fakedHttp },
        { provide: AppConfigService, useValue: mockAppConfigService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(EndpointComponent);
    comp = fixture.componentInstance;
    endpointService = fixture.debugElement.injector.get(EndpointService);
    spy = spyOn(endpointService, "getEndpoints").and.returnValues(
      of(mockData)
    );

    spyAdapter = spyOn(endpointService, "getAdapters").and.returnValue(
      of(mockAdapters)
    );

    spyOnRules = spyOn(
      endpointService,
      "getEndpointWithReplicationRules"
    ).and.returnValue(of([]));
    spyOne = spyOn(endpointService, "getEndpoint").and.returnValue(
      of(mockOne[0])
    );
    fixture.detectChanges();
  });

  it("should retrieve endpoint data", () => {
    fixture.detectChanges();
    expect(spy.calls.any()).toBeTruthy();
  });

  it("should endpoint be initialized", () => {
    fixture.detectChanges();
    expect(config.systemInfoEndpoint).toEqual(CURRENT_BASE_HREF + "/endpoints/testing");
  });

  it("should open create endpoint modal", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.editTargets(mockOne);
      fixture.detectChanges();
      expect(comp.target.name).toEqual("target_01");
    });
  }));

  it("should filter endpoints by keyword", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doSearchTargets("target_02");
      fixture.detectChanges();
      expect(comp.targets.length).toEqual(1);
    });
  }));

  it("should render data", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(
        By.css("datagrid-cell")
      );
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el.textContent).toEqual("target_01");
    });
  }));

  it("should open creation endpoint", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      let de: DebugElement = fixture.debugElement.query(By.css("btn-link"));
      expect(de).toBeTruthy();
      fixture.detectChanges();
      click(de);
      fixture.detectChanges();
      let deInput: DebugElement = fixture.debugElement.query(By.css("input"));
      expect(deInput).toBeTruthy();
    });
  }));

  it("should open to edit existing endpoint", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      let de: DebugElement = fixture.debugElement.query(
        del => del.classes["action-item"]
      );
      expect(de).toBeTruthy();
      fixture.detectChanges();
      click(de);
      fixture.detectChanges();
      let deInput: DebugElement = fixture.debugElement.query(By.css("input"));
      expect(deInput).toBeTruthy();
      let elInput: HTMLElement = deInput.nativeElement;
      expect(elInput).toBeTruthy();
      expect(elInput.textContent).toEqual("target_01");
    });
  }));
});
