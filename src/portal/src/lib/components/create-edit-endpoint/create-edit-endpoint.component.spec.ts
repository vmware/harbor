import {
  ComponentFixture,
  TestBed,
  async
} from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from "../../utils/shared/shared.module";

import { FilterComponent } from "../filter/filter.component";

import { CreateEditEndpointComponent } from "./create-edit-endpoint.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ErrorHandler } from "../../utils/error-handler/error-handler";
import { Endpoint } from "../../services/interface";
import {
  EndpointService,
  EndpointDefaultService
} from "../../services/endpoint.service";
import { IServiceConfig, SERVICE_CONFIG } from "../../entities/service.config";
import { of } from "rxjs";
import { HttpClient } from "@angular/common/http";
import { HttpClientTestingModule } from "@angular/common/http/testing";
import { CURRENT_BASE_HREF } from "../../utils/utils";
import { AppConfigService } from '../../../app/services/app-config.service';

describe("CreateEditEndpointComponent (inline template)", () => {
  let mockData: Endpoint = {
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
  };
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
  let mockAdapters = ['harbor', 'docker hub'];

  let comp: CreateEditEndpointComponent;
  let fixture: ComponentFixture<CreateEditEndpointComponent>;

  let config: IServiceConfig = {
    systemInfoEndpoint: CURRENT_BASE_HREF + "/endpoints/testing"
  };

  let endpointService: EndpointService;
  let http: HttpClient;

  let spy: jasmine.Spy;
  let spyAdapter: jasmine.Spy;
  const mockAppConfigService = {
    getConfig: () => {
      return {
        project_creation_restriction: "",
        with_chartmuseum: ""
      };
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, NoopAnimationsModule],
      declarations: [
        FilterComponent,
        CreateEditEndpointComponent,
        InlineAlertComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: HttpClient, useValue: fakedHttp },
        { provide: AppConfigService, useValue: mockAppConfigService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateEditEndpointComponent);
    comp = fixture.componentInstance;
    endpointService = fixture.debugElement.injector.get(EndpointService);
    http = fixture.debugElement.injector.get(HttpClient);
    spyAdapter = spyOn(endpointService, "getAdapters").and.returnValue(
      of(mockAdapters)
    );
    spy = spyOn(endpointService, "getEndpoint").and.returnValue(
      of(mockData)
    );
    fixture.detectChanges();

    comp.openCreateEditTarget(true, 1);
    fixture.detectChanges();
  });

  it("should be created", () => {
    fixture.detectChanges();
    expect(comp).toBeTruthy();
  });

  it("should get endpoint be called", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      expect(spy.calls.any()).toBeTruthy();
    });
  }));
  it("should get adapterInfo", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      expect(comp.adapterInfo).toBeTruthy();
    });
  }));

  it("should get endpoint and open modal", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      expect(comp.target.name).toEqual("target_01");
    });
  }));

  it("should endpoint be initialized", () => {
    fixture.detectChanges();
    expect(config.systemInfoEndpoint).toEqual(CURRENT_BASE_HREF + "/endpoints/testing");
  });
});
