import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ResetPasswordComponent } from './reset-password.component';
import { FormsModule } from '@angular/forms';
import { PasswordSettingService } from '../password-setting.service';
import { RouterTestingModule } from '@angular/router/testing';
import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';
import { ErrorHandler } from '../../../../lib/utils/error-handler';

describe('ResetPasswordComponent', () => {
    let component: ResetPasswordComponent;
    let fixture: ComponentFixture<ResetPasswordComponent>;
    let fakePasswordSettingService = null;
    let fakeMessageHandlerService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                BrowserAnimationsModule,
                ClarityModule
            ],
            declarations: [ResetPasswordComponent, InlineAlertComponent],
            providers: [
                TranslateService,
                { provide: PasswordSettingService, useValue: fakePasswordSettingService },
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                ErrorHandler
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ResetPasswordComponent);
        component = fixture.componentInstance;
        component.inlineAlert = TestBed.createComponent(InlineAlertComponent).componentInstance;
        component.open();
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
