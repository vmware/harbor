import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CreateProjectComponent } from './create-project.component';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { EndpointDefaultService, EndpointService, ProjectService } from '../../../lib/services';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { ErrorHandler } from '../../../lib/utils/error-handler';
import { IServiceConfig, SERVICE_CONFIG } from '../../../lib/entities/service.config';
import { CURRENT_BASE_HREF } from '../../../lib/utils/utils';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('CreateProjectComponent', () => {
    let component: CreateProjectComponent;
    let fixture: ComponentFixture<CreateProjectComponent>;
    const mockProjectService = {
        checkProjectExists: function(name: string) {
            if (name === 'test') {
                return  of({status: 200}).pipe(delay(10));
            } else {
                return  of({status: 404}).pipe(delay(10));
            }
        },
        createProject: function () {
            return of(true);
        }
    };
    const mockMessageHandlerService = {
        showSuccess: function() {
        }
    };
    const config: IServiceConfig = {
        systemInfoEndpoint: CURRENT_BASE_HREF + "/endpoints/testing"
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                HttpClientTestingModule,
                BrowserAnimationsModule,
                FormsModule,
                ClarityModule,
                TranslateModule.forRoot()
            ],
            declarations: [CreateProjectComponent, InlineAlertComponent],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                TranslateService,
                { provide: SERVICE_CONFIG, useValue: config },
                {provide: ProjectService, useValue: mockProjectService},
                {provide: MessageHandlerService, useValue: mockMessageHandlerService},
                { provide: EndpointService, useClass: EndpointDefaultService },
                ErrorHandler
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateProjectComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open and close', async () => {
        let modelBody: HTMLDivElement;
        modelBody = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeFalsy();
        component.createProjectOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        modelBody = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeTruthy();
        const cancelButton: HTMLButtonElement = fixture.nativeElement.querySelector("#new-project-cancel");
        cancelButton.click();
        fixture.detectChanges();
        await fixture.whenStable();
        modelBody = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeFalsy();
    });

    it('should check project name', async () => {
        fixture.autoDetectChanges(true);
        component.createProjectOpened = true;
        await fixture.whenStable();
        const nameInput: HTMLInputElement = fixture.nativeElement.querySelector("#create_project_name");
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        let el: HTMLSpanElement;
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeTruthy();
        nameInput.value = "test";
        nameInput.dispatchEvent(new Event("input"));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeTruthy();
        nameInput.value = "test1";
        nameInput.dispatchEvent(new Event("input"));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeFalsy();
        const okButton: HTMLButtonElement = fixture.nativeElement.querySelector("#new-project-ok");
        okButton.click();
        await fixture.whenStable();
        const modelBody: HTMLDivElement = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeFalsy();
    });

    it('should enable proxy cache', async () => {
        component.enableProxyCache = true;
        component.isSystemAdmin = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const endpoint: HTMLDivElement = fixture.nativeElement.querySelector("#endpoint");
        expect(endpoint).toBeFalsy();
    });
});
