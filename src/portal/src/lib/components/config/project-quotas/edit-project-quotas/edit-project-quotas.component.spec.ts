import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { EditProjectQuotasComponent } from './edit-project-quotas.component';
import { SERVICE_CONFIG, IServiceConfig } from '../../../../entities/service.config';
import { EditQuotaQuotaInterface } from '../../../../services';
import { HarborLibraryModule } from '../../../../harbor-library.module';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { CURRENT_BASE_HREF } from "../../../../utils/utils";
import { ErrorHandler } from '../../../../utils/error-handler';

describe('EditProjectQuotasComponent', () => {
  let component: EditProjectQuotasComponent;
  let fixture: ComponentFixture<EditProjectQuotasComponent>;
  let config: IServiceConfig = {
    quotaUrl: CURRENT_BASE_HREF + "/quotas/testing"
  };
  const mockedEditQuota: EditQuotaQuotaInterface = {
    editQuota: "Edit Default Project Quotas",
    setQuota: "Set the default project quotas when creating new projects",
    storageQuota: "Default storage consumption",
    quotaHardLimitValue: {storageLimit: -1, storageUnit: "Byte"},
    isSystemDefaultQuota: true
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        HarborLibraryModule,
        BrowserAnimationsModule
      ],
      providers: [
        { provide: SERVICE_CONFIG, useValue: config },
        ErrorHandler
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(EditProjectQuotasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });
  it('should create', () => {
    expect(component).toBeTruthy();
  });
  // ToDo update it with storage edit?
  // it('should open', async () => {
  //   component.openEditQuota = true;
  //   fixture.detectChanges();
  //   await fixture.whenStable();
  //   component.openEditQuotaModal(mockedEditQuota);
  //   fixture.detectChanges();
  //   await fixture.whenStable();
  //   let countInput: HTMLInputElement = fixture.nativeElement.querySelector('#count');
  //   countInput.value = "100";
  //   countInput.dispatchEvent(new Event("input"));
  //   fixture.detectChanges();
  //   await fixture.whenStable();
  //   expect(component.isValid).toBeTruthy();
  // });
});
