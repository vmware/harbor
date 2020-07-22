import { async, ComponentFixture, fakeAsync, getTestBed, TestBed, tick } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { GlobalSearchComponent } from './global-search.component';
import { SearchTriggerService } from './search-trigger.service';
import { FormsModule } from '@angular/forms';
import { AppConfigService } from '../../services/app-config.service';
import { SkinableConfig } from "../../services/skinable-config.service";
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';

describe('GlobalSearchComponent', () => {
    let component: GlobalSearchComponent;
    let fixture: ComponentFixture<GlobalSearchComponent>;
    let fakeSearchTriggerService = {
        searchClearChan$: {
            subscribe: function () {
            }
        },
        triggerSearch() {
            return undefined;
        }
    };
    let fakeAppConfigService = {
        isIntegrationMode: function () {
            return true;
        }
    };
    let fakeSkinableConfig = {
        getProject: function () {
            return {
                introduction: {}
            };
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule
            ],
            declarations: [GlobalSearchComponent],
            providers: [
                TranslateService,
                { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: SkinableConfig, useValue: fakeSkinableConfig }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(GlobalSearchComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should trigger search', fakeAsync(async () => {
        const  service: SearchTriggerService = TestBed.get(SearchTriggerService);
        const spy: jasmine.Spy = spyOn(service,  'triggerSearch').and.callThrough();
        const input: HTMLInputElement = fixture.nativeElement.querySelector('#search_input');
        expect(input).toBeTruthy();
        input.value = 'test';
        input.dispatchEvent(new Event('keyup'));
        tick(500);
        fixture.detectChanges();
        await fixture.whenStable();
        expect(spy.calls.count()).toEqual(1);
    }));
});
