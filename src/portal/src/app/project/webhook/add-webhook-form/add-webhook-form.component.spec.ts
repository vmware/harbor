import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AddWebhookFormComponent } from './add-webhook-form.component';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { WebhookService } from "../webhook.service";
import { MessageHandlerService } from "../../../shared/message-handler/message-handler.service";
import { of } from 'rxjs';
import { Webhook } from "../webhook";
import { InlineAlertComponent } from "../../../shared/inline-alert/inline-alert.component";
import { ErrorHandler } from '../../../../lib/utils/error-handler';

describe('AddWebhookFormComponent', () => {
    let component: AddWebhookFormComponent;
    let fixture: ComponentFixture<AddWebhookFormComponent>;
    const mockWebhookService = {
        getCurrentUser: () => {
            return of(null);
        },
        createWebhook() {
            return of(null);
        },
        editWebhook() {
            return of(null);
        },
        testEndpoint() {
            return of(null);
        },
        eventTypeToText(eventType: string) {
            return eventType;
        }
    };
    const mockMessageHandlerService = {
        handleError: () => { }
    };
    const mockedWehook: Webhook = {
        id: 1,
        project_id: 1,
        name: 'test',
        description: 'just a test webhook',
        targets: [{
            address: 'https://test.com',
            type: 'http',
            attachment:  null,
           auth_header: null,
           skip_cert_verify: true,
        }],
        event_types: [
          'projectQuota'
        ],
        creator: null,
        creation_time: null,
        update_time: null,
        enabled: true,
    };
    const mockedMetadata = {
        "event_type": [
            "projectQuota",
            "pullImage",
            "scanningFailed",
            "uploadChart",
            "deleteChart",
            "downloadChart",
            "scanningCompleted",
            "pushImage",
            "deleteImage"
        ],
        "notify_type": [
            "http",
            "slack"
        ]
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [AddWebhookFormComponent,
                InlineAlertComponent,
            ],
            providers: [
                TranslateService,
                { provide: WebhookService, useValue: mockWebhookService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },
                ErrorHandler


            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AddWebhookFormComponent);
        component = fixture.componentInstance;
        component.metadata = mockedMetadata;
        fixture.detectChanges();
    });

    it('should create', async () => {
        expect(component).toBeTruthy();
        const cancelButtonForAdd: HTMLButtonElement = fixture.nativeElement.querySelector("#add-webhook-cancel");
        expect(cancelButtonForAdd).toBeTruthy();
        component.isModify = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const cancelButtonForEdit: HTMLButtonElement = fixture.nativeElement.querySelector("#edit-webhook-cancel");
        expect(cancelButtonForEdit).toBeTruthy();
    });
    it("should occur a 'name is required' error",  async () => {
        await fixture.whenStable();
        fixture.autoDetectChanges(true);
        const nameInput: HTMLInputElement = fixture.nativeElement.querySelector("#name");
        nameInput.value = "test";
        nameInput.dispatchEvent(new Event('input'));
        nameInput.value = null;
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        const errorEle: HTMLElement = fixture.nativeElement.querySelector("clr-control-error");
        expect(errorEle.innerText).toEqual('WEBHOOK.NAME_REQUIRED');
    });
    it("add button should work", async () => {
        const spy: jasmine.Spy = spyOn(component, 'add').and.returnValue(undefined);
        component.webhook = mockedWehook;
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.isValid).toBeTruthy();
        const addButton: HTMLButtonElement = fixture.nativeElement.querySelector("#new-webhook-continue");
        addButton.dispatchEvent(new Event('click'));
        fixture.detectChanges();
        await fixture.whenStable();
        expect(spy.calls.count()).toEqual(1);
    });
});
