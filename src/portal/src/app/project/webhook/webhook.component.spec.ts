import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { WebhookComponent } from './webhook.component';
import { ActivatedRoute } from '@angular/router';
import { WebhookService } from './webhook.service';
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { delay } from "rxjs/operators";
import { Webhook } from "./webhook";
import { AddWebhookFormComponent } from "./add-webhook-form/add-webhook-form.component";
import { InlineAlertComponent } from "../../shared/inline-alert/inline-alert.component";
import { AddWebhookComponent } from "./add-webhook/add-webhook.component";
import { ConfirmationDialogComponent } from "../../../lib/components/confirmation-dialog";
import { UserPermissionService } from '../../../lib/services';
import { ErrorHandler } from '../../../lib/utils/error-handler';

describe('WebhookComponent', () => {
    let component: WebhookComponent;
    let fixture: ComponentFixture<WebhookComponent>;
    const mockMessageHandlerService = {
        handleError: () => { }
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
    const mockWebhookService = {
        listLastTrigger: () => {
            return of([]).pipe(delay(0));
        },
        listWebhook: () => {
            return of([mockedWehook
            ]).pipe(delay(0));
        },
        getWebhookMetadata() {
            return of(mockedMetadata).pipe(delay(0));
        },
        editWebhook() {
            return of(true);
        },
        eventTypeToText(eventType: string) {
            return eventType;
        }
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: (key) => 'value' }),
        snapshot: {
            parent: {
                params: { id: 1 },
                data: {
                    projectResolver: {
                        ismember: true,
                        name: 'library',
                    }
                }
            }
        }
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true).pipe(delay(0));
        }
    };

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
            declarations: [WebhookComponent,
                AddWebhookComponent,
                AddWebhookFormComponent,
                InlineAlertComponent,
                ConfirmationDialogComponent
            ],
            providers: [
                TranslateService,
                ErrorHandler,
                { provide: WebhookService, useValue: mockWebhookService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: UserPermissionService, useValue: mockUserPermissionService },
            ]
        })
            .compileComponents();
    }));

    beforeEach(async () => {
        fixture = TestBed.createComponent(WebhookComponent);
        component = fixture.componentInstance;
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
    });

    it('should create', async () => {
        expect(component).toBeTruthy();
    });
    it('should get webhook list', async () => {
        const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
        expect(rows.length).toEqual(1);
    });
    it('should open modal', async () => {
        component.newWebhook();
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement = fixture.nativeElement.querySelector(".modal-body");
        expect(body).toBeTruthy();
        const title: HTMLElement = fixture.nativeElement.querySelector(".modal-title");
        expect(title.innerText).toEqual('WEBHOOK.ADD_WEBHOOK');
    });
    it('should open edit modal', async () => {
        component.webhookList[0].name = 'test';
        component.selectedRow[0] = component.webhookList[0];
        component.editWebhook();
        fixture.detectChanges();
        await fixture.whenStable();
        const body: HTMLElement = fixture.nativeElement.querySelector(".modal-body");
        expect(body).toBeTruthy();
        const title: HTMLElement = fixture.nativeElement.querySelector(".modal-title");
        expect(title.innerText).toEqual('WEBHOOK.EDIT_WEBHOOK');
        const nameInput: HTMLInputElement = fixture.nativeElement.querySelector("#name");
        expect(nameInput.value).toEqual('test');
    });
    it('should disable webhook', async () => {
        await fixture.whenStable();
        component.selectedRow[0] = component.webhookList[0];
        component.webhookList[0].enabled = true;
        component.switchWebhookStatus();
        fixture.detectChanges();
        await fixture.whenStable();
        const button: HTMLButtonElement = fixture.nativeElement.querySelector("#dialog-action-disable");
        button.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const body: HTMLElement = fixture.nativeElement.querySelector(".modal-body");
        expect(body).toBeFalsy();
    });
    it('should enable webhook', async () => {
        await fixture.whenStable();
        component.webhookList[0].enabled = false;
        component.selectedRow[0] = component.webhookList[0];
        component.switchWebhookStatus();
        fixture.detectChanges();
        await fixture.whenStable();
        const buttonEnable: HTMLButtonElement = fixture.nativeElement.querySelector("#dialog-action-enable");
        buttonEnable.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const bodyEnable: HTMLElement = fixture.nativeElement.querySelector(".modal-body");
        expect(bodyEnable).toBeFalsy();
    });
});



