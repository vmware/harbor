import { Component, Input, OnInit, OnChanges, Output, EventEmitter, ChangeDetectorRef, SimpleChanges } from "@angular/core";
import { LabelService } from "../../services/label.service";
import { Label } from "../../services/interface";
import { ErrorHandler } from "../../utils/error-handler/error-handler";
import { Subject, forkJoin, Observable, throwError as observableThrowError } from "rxjs";
import { debounceTime, distinctUntilChanged } from "rxjs/operators";
import { map, catchError } from "rxjs/operators";

export interface LabelState {
    iconsShow: boolean;
    label: Label;
    show: boolean;
}

@Component({
    selector: "hbr-filter-label",
    templateUrl: "./filter-label.component.html",
    styleUrls: ["./filter-label.component.scss"]
})
export class FilterLabelComponent implements OnInit, OnChanges {

    openFilterLabelPanel: boolean;
    labelLists: LabelState[] = [];
    filterLabelName = '';
    labelNameFilter: Subject<string> = new Subject<string>();
    @Input() isOpen: boolean;
    @Input() projectId: number;
    @Input() selectedLabelInfo: Label[];
    @Output() selectedLabels = new EventEmitter<LabelState[]>();
    @Output() closePanelEvent = new EventEmitter();

    constructor(private labelService: LabelService,
        private errorHandler: ErrorHandler) { }

    ngOnInit(): void {
        forkJoin(this.getGLabels(), this.getPLabels()).subscribe(() => {
            this.selectedLabelInfo.forEach(info => {
                if (this.labelLists.length) {
                    let lab = this.labelLists.find(data => data.label.id === info.id);
                    if (lab) { this.selectOper(lab); }
                }
            });
        }, error => {
            this.errorHandler.error(error);
        });

        this.labelNameFilter
            .pipe(debounceTime(500))
            .pipe(distinctUntilChanged())
            .subscribe((name: string) => {
                if (this.filterLabelName.length) {
                    this.labelLists.forEach(data => {
                        data.show = data.label.name.indexOf(this.filterLabelName) !== -1;
                    });
                }
            });
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes['isOpen']) { this.openFilterLabelPanel = changes['isOpen'].currentValue; }
    }

    getGLabels() {
        return this.labelService.getGLabels().pipe(map((res: Label[]) => {
            if (res.length) {
                res.forEach(data => {
                    this.labelLists.push({ 'iconsShow': false, 'label': data, 'show': true });
                });
            }
        })
            , catchError(error => observableThrowError(error)));
    }

    getPLabels() {
        if (this.projectId && this.projectId > 0) {
            return this.labelService.getPLabels(this.projectId).pipe(map((res1: Label[]) => {
                if (res1.length) {
                    res1.forEach(data => {
                        this.labelLists.push({ 'iconsShow': false, 'label': data, 'show': true });
                    });
                }
            })
                , catchError(error => observableThrowError(error)));
        }
    }

    handleInputFilter(): void {
        if (this.filterLabelName.length) {
            this.labelNameFilter.next(this.filterLabelName);
        } else {
            this.labelLists.every(data => data.show = true);
        }
    }

    selectLabel(labelInfo: LabelState): void {
        if (labelInfo) {
            let isClick = true;
            if (!labelInfo.iconsShow) {
                this.selectOper(labelInfo, isClick);
            } else {
                this.unSelectOper(labelInfo, isClick);
            }
        }
    }

    selectOper(labelInfo: LabelState, isClick?: boolean): void {
        // set the selected label in front
        this.labelLists.splice(this.labelLists.indexOf(labelInfo), 1);
        this.labelLists.some((data, i) => {
            if (!data.iconsShow) {
                this.labelLists.splice(i, 0, labelInfo);
                return true;
            }
        });
        // when is the last one
        if (this.labelLists.every(data => data.iconsShow === true)) {
            this.labelLists.push(labelInfo);
        }

        labelInfo.iconsShow = true;
        if (isClick) {
            this.selectedLabels.emit(this.labelLists);
        }

    }

    unSelectOper(labelInfo: LabelState, isClick?: boolean): void {
        this.sortOperation(this.labelLists, labelInfo);

        labelInfo.iconsShow = false;
        if (isClick) {
            this.selectedLabels.emit(this.labelLists);
        }
    }

    // insert the unselected label to groups with the same icons
    sortOperation(labelList: LabelState[], labelInfo: LabelState): void {
        labelList.some((data, i) => {
            if (!data.iconsShow) {
                if (data.label.scope === labelInfo.label.scope) {
                    labelList.splice(i, 0, labelInfo);
                    labelList.splice(labelList.indexOf(labelInfo, 0), 1);
                    return true;
                }
                if (data.label.scope !== labelInfo.label.scope && i === labelList.length - 1) {
                    labelList.push(labelInfo);
                    labelList.splice(labelList.indexOf(labelInfo), 1);
                    return true;
                }
            }
            if (data.iconsShow && i === labelList.length - 1) {
                labelList.push(labelInfo);
                labelList.splice(labelList.indexOf(labelInfo), 1);
                return true;
            }
        });
    }

    closeFilter(): void {
        this.closePanelEvent.emit();
    }
}
