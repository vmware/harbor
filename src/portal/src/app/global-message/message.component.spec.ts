import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ElementRef } from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ClarityModule } from '@clr/angular';
import { Message } from './message';
import { MessageService } from './message.service';
import { MessageComponent } from './message.component';
import { AlertType } from '../shared/shared.const';

describe('MessageComponent', () => {
    let component: MessageComponent;
    let fixture: ComponentFixture<MessageComponent>;
    let fakeElementRef = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ClarityModule,
                RouterTestingModule,
                TranslateModule.forRoot()
            ],
            declarations: [MessageComponent],
            providers: [
                MessageService,
                TranslateService,
                {provide: ElementRef, useValue: fakeElementRef}
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(MessageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open mask layer when unauthorized', async () => {
        component.globalMessageOpened = true;
        component.globalMessage = Message.newMessage(401, "unauthorized", AlertType.DANGER);
        fixture.detectChanges();
        await fixture.whenStable();
        const ele: HTMLDivElement = fixture.nativeElement.querySelector(".mask-layer");
        expect(ele).toBeTruthy();
    });

    it("should not open mask layer when it's not unauthorized", async () => {
        component.globalMessageOpened = true;
        component.globalMessage = Message.newMessage(403, "forbidden", AlertType.WARNING);
        fixture.detectChanges();
        await fixture.whenStable();
        const ele: HTMLDivElement = fixture.nativeElement.querySelector(".mask-layer");
        expect(ele).toBeFalsy();
    });
});
