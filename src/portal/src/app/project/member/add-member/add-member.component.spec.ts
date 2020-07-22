import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AddMemberComponent } from './add-member.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from "@angular/common/http/testing";
import { MemberService } from '../member.service';
import { UserService } from '../../../user/user.service';
import { of } from 'rxjs';
import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { ActivatedRoute } from '@angular/router';
import { ErrorHandler } from '../../../../lib/utils/error-handler';

describe('AddMemberComponent', () => {
    let component: AddMemberComponent;
    let fixture: ComponentFixture<AddMemberComponent>;
    const mockMemberService = {
        getUsersNameList: () => {
            return of([]);
        }
    };
    const mockUserService = {
        getUsersNameList: () => {
            return of([
                [], []
            ]);
        }
    };

    const mockMessageHandlerService = {
        showSuccess: () => { },
        handleError: () => { },
        isAppLevel: () => { },
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
            declarations: [AddMemberComponent],
            providers: [
                TranslateService,
                { provide: MemberService, useValue: mockMemberService },
                { provide: UserService, useValue: mockUserService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },
                ErrorHandler,
                {
                    provide: ActivatedRoute, useValue: {
                        RouterparamMap: of({ get: (key) => 'value' }),
                        snapshot: {
                            parent: {
                                params: { id: 1 }
                            },
                            data: 1
                        }
                    }
                }

            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AddMemberComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
