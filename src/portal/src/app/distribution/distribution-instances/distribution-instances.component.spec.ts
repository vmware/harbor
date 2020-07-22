import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ClarityModule } from '@clr/angular';
import { SharedModule } from '../../shared/shared.module';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { DistributionInstancesComponent } from './distribution-instances.component';
import { PreheatService } from "../../../../ng-swagger-gen/services/preheat.service";
import { Instance } from '../../../../ng-swagger-gen/models/instance';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { Metadata } from '../../../../ng-swagger-gen/models/metadata';
import { DistributionSetupModalComponent } from '../distribution-setup-modal/distribution-setup-modal.component';

describe('DistributionInstanceComponent', () => {
  let component: DistributionInstancesComponent;
  let fixture: ComponentFixture<DistributionInstancesComponent>;

  const instance1: Instance = {
    name: 'Test1',
    default: true,
    enabled: true,
    description: 'Test1',
    endpoint: 'http://test.com',
    id: 1,
    setup_timestamp: new Date().getTime(),
    auth_mode: 'NONE',
    vendor: 'kraken',
    status: 'Healthy'
  };

  const instance2: Instance = {
    name: 'Test2',
    default: false,
    enabled: false,
    description: 'Test2',
    endpoint: 'http://test2.com',
    id: 2,
    setup_timestamp: new Date().getTime() + 3600000,
    auth_mode: 'BASIC',
    auth_info: {
      password: '123',
      username: 'abc'
    },
    vendor: 'kraken',
    status: 'Healthy'
  };

  const instance3: Instance = {
    name: 'Test3',
    default: false,
    enabled: true,
    description: 'Test3',
    endpoint: 'http://test3.com',
    id: 3,
    setup_timestamp: new Date().getTime() + 7200000,
    auth_mode: 'OAUTH',
    auth_info: {
      token: 'xxxxxxxxxxxxxxxxxxxx'
    },
    vendor: 'kraken',
    status: 'Unhealthy'
  };

  const mockedProviders: Metadata[] = [{
    'icon': 'https://raw.githubusercontent.com/alibaba/Dragonfly/master/docs/images/logo.png',
    'id': 'dragonfly',
    'maintainers': ['Jin Zhang/taiyun.zj@alibaba-inc.com'],
    'name': 'Dragonfly',
    'source': 'https://github.com/alibaba/Dragonfly',
    'version': '0.10.1'
  }, {
    'icon': 'https://github.com/uber/kraken/blob/master/assets/kraken-logo-color.svg',
    'id': 'kraken',
    'maintainers': ['mmpei/peimingming@corp.netease.com'],
    'name': 'Kraken',
    'source': 'https://github.com/uber/kraken',
    'version': '0.1.3'
  }];


  const fakedPreheatService = {
    ListInstancesResponse() {
      const res: HttpResponse<Array<Instance>> = new HttpResponse<Array<Instance>>({
        headers: new HttpHeaders({'x-total-count': '3'}),
        body: [instance1, instance2, instance3]
      });
      return of(res).pipe(delay(10));
    },
    ListProviders() {
      return of(mockedProviders).pipe(delay(10));
    }
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ClarityModule,
        TranslateModule,
        SharedModule,
        HttpClientTestingModule
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      providers: [
        { provide: PreheatService, useValue: fakedPreheatService }
      ],
      declarations: [
        DistributionInstancesComponent,
        DistributionSetupModalComponent
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DistributionInstancesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should render list and get providers', async () => {
    fixture.autoDetectChanges(true);
    await fixture.whenStable();
    expect(component.providers.length).toEqual(2);
    const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
    expect(rows.length).toEqual(3);
  });

  it('should open modal', async () => {
    fixture.autoDetectChanges(true);
    await fixture.whenStable();
    const addButton: HTMLButtonElement = fixture.nativeElement.querySelector("#new-instance");
    addButton.click();
    await fixture.whenStable();
    const modal: HTMLElement = fixture.nativeElement.querySelector("clr-modal");
    expect(modal).toBeTruthy();
  });
});
