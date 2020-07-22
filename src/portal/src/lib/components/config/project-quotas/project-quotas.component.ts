import { Component, Input, Output, EventEmitter, ViewChild, SimpleChanges, OnChanges } from '@angular/core';
import { Configuration } from '../config';
import {
  Quota, State, Comparator, ClrDatagridComparatorInterface, QuotaHardLimitInterface, QuotaHard
} from '../../../services/interface';
import {
  clone, isEmpty, getChanges, getSuitableUnit, calculatePage, CustomComparator
  , getByte, GetIntegerAndUnit
} from '../../../utils/utils';
import { ErrorHandler } from '../../../utils/error-handler';
import { QuotaUnits, QuotaUnlimited, QUOTA_DANGER_COEFFICIENT, QUOTA_WARNING_COEFFICIENT } from '../../../entities/shared.const';
import { EditProjectQuotasComponent } from './edit-project-quotas/edit-project-quotas.component';
import {
  ConfigurationService
} from '../../../services';
import { TranslateService } from '@ngx-translate/core';
import { forkJoin } from 'rxjs';
import { QuotaService } from "../../../services/quota.service";
import { Router } from '@angular/router';
import { finalize } from 'rxjs/operators';
const quotaSort = {
  storage: "used.storage",
  sortType: 'string'
};
const QuotaType = 'project';

@Component({
  selector: 'project-quotas',
  templateUrl: './project-quotas.component.html',
  styleUrls: ['./project-quotas.component.scss']
})
export class ProjectQuotasComponent implements OnChanges {

  config: Configuration = new Configuration();
  @ViewChild('editProjectQuotas', {static: false})
  editQuotaDialog: EditProjectQuotasComponent;
  loading = true;
  quotaHardLimitValue: QuotaHardLimitInterface;
  currentState: State;

  @Output() configChange: EventEmitter<Configuration> = new EventEmitter<Configuration>();
  @Output() refreshAllconfig: EventEmitter<Configuration> = new EventEmitter<Configuration>();
  quotaList: Quota[] = [];
  originalConfig: Configuration;
  currentPage = 1;
  totalCount = 0;
  pageSize = 15;
  quotaDangerCoefficient: number = QUOTA_DANGER_COEFFICIENT;
  quotaWarningCoefficient: number = QUOTA_WARNING_COEFFICIENT;
  @Input()
  get allConfig(): Configuration {
    return this.config;
  }
  set allConfig(cfg: Configuration) {
    this.config = cfg;
    this.configChange.emit(this.config);
  }
  storageComparator: Comparator<Quota> = new CustomComparator<Quota>(quotaSort.storage, quotaSort.sortType);
  selectedRow: Quota[] = [];

  constructor(
    private configService: ConfigurationService,
    private quotaService: QuotaService,
    private translate: TranslateService,
    private router: Router,
    private errorHandler: ErrorHandler) { }

  editQuota() {
    if (this.selectedRow && this.selectedRow.length === 1) {
      const defaultTexts = [this.translate.get('QUOTA.EDIT_PROJECT_QUOTAS')
        , this.translate.get('QUOTA.SET_QUOTAS', { params: this.selectedRow[0].ref.name })
        , this.translate.get('QUOTA.STORAGE_QUOTA')];
      forkJoin(...defaultTexts).subscribe(res => {
        const defaultTextsObj = {
          editQuota: res[0],
          setQuota: res[1],
          storageQuota: res[2],
          quotaHardLimitValue: this.selectedRow[0],
          isSystemDefaultQuota: false
        };
        this.editQuotaDialog.openEditQuotaModal(defaultTextsObj);
      });
    }
  }

  editDefaultQuota(quotaHardLimitValue: QuotaHardLimitInterface) {
    const defaultTexts = [this.translate.get('QUOTA.EDIT_DEFAULT_PROJECT_QUOTAS'), this.translate.get('QUOTA.SET_DEFAULT_QUOTAS')
      , this.translate.get('QUOTA.STORAGE_DEFAULT_QUOTA')];
    forkJoin(...defaultTexts).subscribe(res => {
      const defaultTextsObj = {
        editQuota: res[0],
        setQuota: res[1],
        storageQuota: res[2],
        quotaHardLimitValue: quotaHardLimitValue,
        isSystemDefaultQuota: true
      };
      this.editQuotaDialog.openEditQuotaModal(defaultTextsObj);

    });
  }
  public getChanges() {
    let allChanges = getChanges(this.originalConfig, this.config);
    if (allChanges) {
      return this.getQuotaChanges(allChanges);
    }
    return null;
  }

  getQuotaChanges(allChanges) {
    let changes = {};
    for (let prop in allChanges) {
      if (prop === 'storage_per_project') {
        changes[prop] = allChanges[prop];
      }
    }
    return changes;
  }

  public saveConfig(configQuota): void {
    this.allConfig.storage_per_project.value = +configQuota.storage === QuotaUnlimited ?
      configQuota.storage : getByte(configQuota.storage, configQuota.storageUnit);
    let changes = this.getChanges();
    if (!isEmpty(changes)) {
      this.loading = true;
      this.configService.saveConfigurations(changes)
        .pipe(finalize(() => {
          this.loading = false;
          this.editQuotaDialog.openEditQuota = false;
        }))
        .subscribe(response => {
          this.refreshAllconfig.emit();
          this.errorHandler.info('CONFIG.SAVE_SUCCESS');
        }
          , error => {
            this.errorHandler.error(error);
          });
    } else {
      // Inprop situation, should not come here
      this.translate.get('CONFIG.NO_CHANGE').subscribe(res => {
        this.editQuotaDialog.inlineAlert.showInlineError(res);
      });
    }
  }

  confirmEdit(event) {
    if (event.isSystemDefaultQuota) {
      this.saveConfig(event.formValue);
    } else {
      this.saveCurrentQuota(event);
    }
  }
  saveCurrentQuota(event) {
    let storage = +event.formValue.storage === QuotaUnlimited ?
      +event.formValue.storage : getByte(+event.formValue.storage, event.formValue.storageUnit);
    let rep: QuotaHard = { hard: { storage } };
    this.loading = true;
    this.quotaService.updateQuota(event.id, rep).subscribe(res => {
      this.editQuotaDialog.openEditQuota = false;
      this.getQuotaList(this.currentState);
      this.errorHandler.info('QUOTA.SAVE_SUCCESS');
    }, error => {
      this.editQuotaDialog.inlineAlert.showInlineError(error);
      this.loading = false;
    });
  }

  getquotaHardLimitValue() {
    const storageNumberAndUnit = this.allConfig.storage_per_project ? this.allConfig.storage_per_project.value : QuotaUnlimited;
    const storageLimit = storageNumberAndUnit;
    const storageUnit = this.getIntegerAndUnit(storageNumberAndUnit, 0).partCharacterHard;
    this.quotaHardLimitValue = { storageLimit, storageUnit };
  }
  getQuotaList(state: State) {
    if (!state || !state.page) {
      return;
    }
    // Keep state for future filtering and sorting
    this.currentState = state;

    let pageNumber: number = calculatePage(state);
    if (pageNumber <= 0) { pageNumber = 1; }
    let sortBy: any = '';
    if (state.sort) {
      sortBy = state.sort.by as string | ClrDatagridComparatorInterface<any>;
      sortBy = sortBy.fieldName ? sortBy.fieldName : sortBy;
      sortBy = state.sort.reverse ? `-${sortBy}` : sortBy;
    }
    this.loading = true;

    this.quotaService.getQuotaList(QuotaType, pageNumber, this.pageSize, sortBy).pipe(finalize(() => {
      this.loading = false;
      this.selectedRow = [];
    })).subscribe(res => {
      if (res.headers) {
        let xHeader: string = res.headers.get("X-Total-Count");
        if (xHeader) {
          this.totalCount = parseInt(xHeader, 0);
        }
      }
      this.quotaList = res.body.filter((quota) => {
        return quota.ref !== null;
      }) as Quota[];
    }, error => {
      this.errorHandler.error(error);
    });
  }
  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["allConfig"]) {
      this.originalConfig = clone(this.config);
      this.getquotaHardLimitValue();
    }
  }
  getSuitableUnit(value) {
    const QuotaUnitsCopy = clone(QuotaUnits);
    return getSuitableUnit(value, QuotaUnitsCopy);
  }
  getIntegerAndUnit(valueHard, valueUsed) {
    return GetIntegerAndUnit(valueHard
      , clone(QuotaUnits), valueUsed, clone(QuotaUnits));
  }

  goToLink(proId) {
    let linkUrl = ["harbor", "projects", proId, "summary"];
    this.router.navigate(linkUrl);
  }
  refresh() {
    const state: State = {
      page: {
        from: 0,
        to: 14,
        size: 15
      },
    };
    this.getQuotaList(state);
    this.selectedRow = [];
  }
}
