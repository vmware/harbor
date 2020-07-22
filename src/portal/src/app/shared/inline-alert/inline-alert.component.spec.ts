import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { InlineAlertComponent } from './inline-alert.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MessageHandlerService } from '../message-handler/message-handler.service';
import { ErrorHandler } from '../../../lib/utils/error-handler';

describe('InlineAlertComponent', () => {
    let component: InlineAlertComponent;
    let fixture: ComponentFixture<InlineAlertComponent>;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [InlineAlertComponent],
            providers: [
                TranslateService,
                ErrorHandler
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(InlineAlertComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
