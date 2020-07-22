import {
    Component, Input,
    OnInit
} from "@angular/core";
import { ConfigScannerService } from "../config-scanner.service";
import { finalize } from "rxjs/operators";
import { ScannerMetadata } from "../scanner-metadata";
import { DatePipe } from "@angular/common";
import { TranslateService } from "@ngx-translate/core";
import { ErrorHandler } from "../../../../lib/utils/error-handler";
import {DATABASE_NEXT_UPDATE_PROPERTY, DATABASE_UPDATED_PROPERTY} from "../../../../lib/utils/utils";

@Component({
    selector: 'scanner-metadata',
    templateUrl: 'scanner-metadata.html',
    styleUrls: ['./scanner-metadata.scss']
})
export class ScannerMetadataComponent implements  OnInit {
    @Input() uid: string;
    loading: boolean = false;
    scannerMetadata: ScannerMetadata;
    constructor(private configScannerService: ConfigScannerService,
                private errorHandler: ErrorHandler) {
    }
    ngOnInit(): void {
        this.loading = true;
        this.configScannerService.getScannerMetadata(this.uid)
            .pipe(finalize(() => this.loading = false))
            .subscribe(response => {
                this.scannerMetadata = response;
            }, error => {
                this.errorHandler.error(error);
            });
    }
    parseDate(item: any): string {
        if (this.hasValue(item) && this.hasDateValue(item)) {
            return new DatePipe('en-us').transform(item.value, 'short');
        }
        if (this.hasValue(item)) {
            return item.value;
        }
        return '';
    }
    hasValue(item: any): boolean {
        return item && item.value;
    }
    hasDateValue(item: any): boolean {
        switch (item.key) {
            case DATABASE_UPDATED_PROPERTY:
            case DATABASE_NEXT_UPDATE_PROPERTY:
                return true;
            default:
                return false;
        }
    }
    toString(arr: string[]) {
        if (arr && arr.length > 0) {
            return "[" + arr.join(" , ") + "]";
        }
        return arr;
    }
}
