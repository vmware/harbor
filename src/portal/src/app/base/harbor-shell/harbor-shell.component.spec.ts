import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { SessionService } from '../../shared/session.service';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { HarborShellComponent } from './harbor-shell.component';
import { ClarityModule } from "@clr/angular";
import { of } from 'rxjs';
import { ConfigScannerService } from "../../config/scanner/config-scanner.service";
import { modalEvents } from '../modal-events.const';
import { AccountSettingsModalComponent } from '../../account/account-settings/account-settings-modal.component';
import { PasswordSettingComponent } from '../../account/password-setting/password-setting.component';
import { AboutDialogComponent } from '../../shared/about-dialog/about-dialog.component';
import { FormsModule } from '@angular/forms';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { AccountSettingsModalService } from '../../account/account-settings/account-settings-modal-service.service';
import { PasswordSettingService } from '../../account/password-setting/password-setting.service';
import { SkinableConfig } from '../../services/skinable-config.service';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';
import { AppConfigService } from "../../services/app-config.service";
import { ErrorHandler } from '../../../lib/utils/error-handler';

describe('HarborShellComponent', () => {
    let component: HarborShellComponent;
    let fixture: ComponentFixture<HarborShellComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        }
    };
    let fakeSearchTriggerService = {
        searchTriggerChan$: of('null')
        ,
        searchCloseChan$: of(null)
    };
    let mockMessageHandlerService = null;
    let mockAccountSettingsModalService = null;
    let mockPasswordSettingService = null;
    let mockSkinableConfig = {
        getProject: function () {
            return {
                introduction: {}
            };
        }
    };
    let fakeAppConfigService = {
        isLdapMode: function () {
            return true;
        },
        isHttpAuthMode: function () {
            return false;
        },
        isOidcMode: function () {
            return false;
        },
        getConfig: function () {
            return {
                with_clair: true
            };
        }
    };
    let fakeConfigScannerService = {
        getScanners() {
            return of(true);
        }
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                TranslateModule.forRoot(),
                ClarityModule,
                BrowserAnimationsModule,
                FormsModule
            ],
            declarations: [HarborShellComponent, AccountSettingsModalComponent
                , PasswordSettingComponent, AboutDialogComponent, InlineAlertComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: fakeSessionService },
                { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: ConfigScannerService, useValue: fakeConfigScannerService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },
                { provide: AccountSettingsModalService, useValue: mockAccountSettingsModalService },
                { provide: PasswordSettingService, useValue: mockPasswordSettingService },
                { provide: SkinableConfig, useValue: mockSkinableConfig },
                ErrorHandler
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(HarborShellComponent);
        component = fixture.componentInstance;
        component.showScannerInfo = true;
        component.accountSettingsModal = TestBed.createComponent(AccountSettingsModalComponent).componentInstance;
        component.accountSettingsModal.inlineAlert = TestBed.createComponent(InlineAlertComponent).componentInstance;
        component.pwdSetting = TestBed.createComponent(PasswordSettingComponent).componentInstance;
        component.aboutDialog = TestBed.createComponent(AboutDialogComponent).componentInstance;
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should open users profile', async () => {
        component.openModal({modalName: modalEvents.USER_PROFILE, modalFlag: false });
        await fixture.whenStable();
        const accountSettingsUsernameInput = fixture.nativeElement.querySelector("#account_settings_username");
        expect(accountSettingsUsernameInput).toBeTruthy();
    });
    it('should open users changPwd', async () => {
        component.openModal({modalName: modalEvents.CHANGE_PWD, modalFlag: false });
        await fixture.whenStable();
        const oldPasswordInput = fixture.nativeElement.querySelector("#oldPassword");
        expect(oldPasswordInput).toBeTruthy();
    });
    it('should open users about-dialog', async () => {
        component.openModal({modalName: modalEvents.ABOUT, modalFlag: false });
        await fixture.whenStable();
        const aboutVersionEl = fixture.nativeElement.querySelector(".about-version");
        expect(aboutVersionEl).toBeTruthy();
    });
});
