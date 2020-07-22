import { ComponentFixture, TestBed, async, fakeAsync, tick } from '@angular/core/testing';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { SharedModule } from "../../utils/shared/shared.module";
import { ErrorHandler } from "../../utils/error-handler/error-handler";
import { IServiceConfig, SERVICE_CONFIG } from '../../entities/service.config';
import {FilterLabelComponent} from "./filter-label.component";
import {LabelService} from "../../services/label.service";
import { RouterTestingModule } from '@angular/router/testing';
import { CURRENT_BASE_HREF } from '../../utils/utils';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { Label } from '../../services';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';


describe("FilterLabelComponent", () => {
  let fixture: ComponentFixture<FilterLabelComponent>;
  let comp: FilterLabelComponent;
  const fakedLabel1: Label = {
    id: 1,
    name: "dd",
    description: "fff",
    color: "#CD3517",
    scope: "g",
    project_id: 0,
    creation_time: "2020-04-20T08:08:39.540765Z",
    update_time: "2020-04-20T08:08:39.540765Z",
    deleted: false
  };
  const fakedLabel2: Label = {
    id: 2,
    name: "ff",
    description: "fff",
    color: "#CD3518",
    scope: "p",
    project_id: 1,
    creation_time: "2020-04-20T08:08:39.540765Z",
    update_time: "2020-04-20T08:08:39.540765Z",
    deleted: false
  };

  const config: IServiceConfig = {
    replicationBaseEndpoint: CURRENT_BASE_HREF + "/replication/testing",
    targetBaseEndpoint: CURRENT_BASE_HREF + "/registries/testing"
  };
  const fakedLabelService = {
    getGLabels() {
      return of([fakedLabel1]).pipe(delay(0));
    },
    getPLabels() {
      return of([fakedLabel2]).pipe(delay(0));
    }
  };
  const fakedErrorHandler = {
    error() {
      return null;
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, NoopAnimationsModule, RouterTestingModule],
      declarations: [
        FilterLabelComponent,
      ],
      providers: [
        { provide: ErrorHandler, useValue: fakedErrorHandler },
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: LabelService, useValue: fakedLabelService}
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(FilterLabelComponent);
    comp = fixture.componentInstance;
    fixture.detectChanges();
    comp.openFilterLabelPanel = true;
  });

  it("Should create", () => {
    expect(comp).toBeTruthy();
  });
  it("Should render and filter", fakeAsync( async () => {
    comp.labelLists = [{
      iconsShow: true,
      label: fakedLabel1,
      show: true,
    },
      {
        iconsShow: true,
        label: fakedLabel2,
        show: true,
      }];
    fixture.detectChanges();
    let buttons: HTMLCollection = fixture.nativeElement.getElementsByClassName('labelBtn');
    expect(buttons.length).toEqual(2);
    const input: HTMLInputElement = fixture.nativeElement.querySelector('.filterInput');
    input.value = 'dd';
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('keyup'));
    tick(600);
    fixture.detectChanges();
    await fixture.whenStable();
    expect(comp.labelLists[1].show).toBeFalsy();
  }));
});
