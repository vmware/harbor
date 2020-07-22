import { ComponentFixture, ComponentFixtureAutoDetect, TestBed } from '@angular/core/testing';
import { HarborLibraryModule } from "../../../harbor-library.module";
import { IServiceConfig, SERVICE_CONFIG } from "../../../entities/service.config";
import { SystemSettingsComponent } from "./system-settings.component";
import { ConfigurationService, SystemInfoService } from "../../../services";
import { ErrorHandler } from "../../../utils/error-handler";
import { of } from "rxjs";
import { StringValueItem } from "../config";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { CURRENT_BASE_HREF } from "../../../utils/utils";
describe('SystemSettingsComponent', () => {
  let component: SystemSettingsComponent;
  let fixture: ComponentFixture<SystemSettingsComponent>;
  const config: IServiceConfig = {
    baseEndpoint: CURRENT_BASE_HREF + "/testing"
  };
  const mockedAllowlist = {
    id: 1,
    project_id: 1,
    expires_at: null,
    items: [
      {cve_id: 'CVE-2019-1234'}
    ]
  };
  const fakedSystemInfoService = {
    getSystemAllowlist() {
       return of(mockedAllowlist);
    },
    getSystemInfo() {
       return of({});
    },
    updateSystemAllowlist() {
      return of(true);
    }
  };
  const fakedErrorHandler = {
    info() {
      return null;
    }
  };
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
          HarborLibraryModule,
          BrowserAnimationsModule
      ],
       providers: [
           ConfigurationService,
           { provide: ErrorHandler, useValue: fakedErrorHandler },
           { provide: SystemInfoService, useValue: fakedSystemInfoService },
           { provide: SERVICE_CONFIG, useValue: config },
             // open auto detect
           { provide: ComponentFixtureAutoDetect, useValue: true }
       ]
    });
  });
  beforeEach(() => {
    fixture = TestBed.createComponent(SystemSettingsComponent);
    component = fixture.componentInstance;
    component.config.auth_mode = new StringValueItem("db_auth",  false );
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('cancel button should works', () => {
    component.systemAllowlist.items.push({cve_id: 'CVE-2019-456'});
    const readOnly: HTMLElement = fixture.nativeElement.querySelector('#repoReadOnly');
    readOnly.click();
    fixture.detectChanges();
    const cancel: HTMLButtonElement = fixture.nativeElement.querySelector('#config_system_cancel');
    cancel.click();
    fixture.detectChanges();
    expect(component.confirmationDlg.opened).toBeTruthy();
  });
  it('save button should works', () => {
    component.systemAllowlist.items[0].cve_id = 'CVE-2019-789';
    const readOnly: HTMLElement = fixture.nativeElement.querySelector('#repoReadOnly');
    readOnly.click();
    fixture.detectChanges();
    const save: HTMLButtonElement = fixture.nativeElement.querySelector('#config_system_save');
    save.click();
    fixture.detectChanges();
    expect(component.systemAllowlistOrigin.items[0].cve_id).toEqual('CVE-2019-789');
  });
});
