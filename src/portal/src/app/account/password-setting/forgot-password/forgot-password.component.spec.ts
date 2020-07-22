import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormsModule } from '@angular/forms';
import { ForgotPasswordComponent } from './forgot-password.component';
import { ClarityModule } from "@clr/angular";
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { PasswordSettingService } from '../password-setting.service';
import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { of } from 'rxjs';
import { ErrorHandler } from '../../../../lib/utils/error-handler';

describe('ForgotPasswordComponent', () => {
    let component: ForgotPasswordComponent;
    let fixture: ComponentFixture<ForgotPasswordComponent>;
    let fakePasswordSettingService = {
        sendResetPasswordMail: () => of(null)
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ForgotPasswordComponent, InlineAlertComponent],
            imports: [
                FormsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                BrowserAnimationsModule
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                TranslateService,
                ErrorHandler,
                { provide: PasswordSettingService, useValue: fakePasswordSettingService }
            ]
        }).compileComponents();
    }));
    let el;
    beforeEach(() => {
        fixture = TestBed.createComponent(ForgotPasswordComponent);
        component = fixture.componentInstance;
        component.inlineAlert = TestBed.createComponent(InlineAlertComponent).componentInstance;
        component.open();
        el = fixture.debugElement;
        fixture.autoDetectChanges();

    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should validate email', async () => {
        await fixture.whenStable();
        let resetPwdInput: HTMLInputElement = fixture.nativeElement.querySelector('#reset_pwd_email');
        expect(resetPwdInput).toBeTruthy();
        resetPwdInput.value = '1234567';
        resetPwdInput.dispatchEvent(new Event('input'));
        resetPwdInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const resetPwdError = fixture.nativeElement.querySelector('#reset_pwd_email-error');
        expect(resetPwdError.innerText).toEqual(' TOOLTIP.EMAIL ');
        // success
        resetPwdInput.value = '1234567@qq.com';
        resetPwdInput.dispatchEvent(new Event('input'));
        resetPwdInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const resetPwdError1 = fixture.nativeElement.querySelector('#reset_pwd_email-error');
        expect(resetPwdError1).toBeNull();

    });
    it('should send email to back end', async () => {
        await fixture.whenStable();
        let resetPwdInput: HTMLInputElement = fixture.nativeElement.querySelector('#reset_pwd_email');
        resetPwdInput.value = '1234567@qq.com';
        resetPwdInput.dispatchEvent(new Event('input'));
        resetPwdInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        expect(el.nativeElement.querySelector('#submit-btn').disabled).toBeFalsy();
        const submitBtn = fixture.nativeElement.querySelector('#submit-btn');
        submitBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const alertText: HTMLSpanElement = fixture.nativeElement.querySelector('.alert-text');
        expect(alertText.innerText).toEqual(' RESET_PWD.SUCCESS ');
    });

});
