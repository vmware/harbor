import {
  Component,
  OnInit,
  Input,
  ViewChild,
  Output,
  EventEmitter,
} from "@angular/core";
import { Webhook } from "../webhook";
import { NgForm } from "@angular/forms";
import { ClrLoadingState } from "@clr/angular";
import { finalize } from "rxjs/operators";
import { WebhookService } from "../webhook.service";
import { InlineAlertComponent } from "../../../shared/inline-alert/inline-alert.component";
import { compareValue } from '../../../../lib/utils/utils';

@Component({
  selector: 'add-webhook-form',
  templateUrl: './add-webhook-form.component.html',
  styleUrls: ['./add-webhook-form.component.scss']
})
export class AddWebhookFormComponent implements OnInit {
  closable: boolean = true;
  checking: boolean = false;
  checkBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  webhookForm: NgForm;
  submitting: boolean = false;
  @Input() projectId: number;
  webhook: Webhook = new Webhook();
  originValue: Webhook;
  isModify: boolean;
  @Input() isOpen: boolean;
  @Output() close = new EventEmitter<boolean>();
  @ViewChild("webhookForm", { static: true }) currentForm: NgForm;
  @ViewChild(InlineAlertComponent, { static: false }) inlineAlert: InlineAlertComponent;
  @Input()
  metadata: any;
  @Output() notify = new EventEmitter<Webhook>();
  constructor(
    private webhookService: WebhookService,
  ) { }

  ngOnInit() {
  }
  onCancel() {
    this.close.emit(false);
    this.currentForm.reset();
    this.inlineAlert.close();
  }

  add() {
    this.submitting = true;
    this.webhookService.createWebhook(this.projectId, this.webhook)
      .pipe(finalize(() => (this.submitting = false)))
      .subscribe(
        response => {
          this.notify.emit();
          this.inlineAlert.close();
        },
        error => {
            this.inlineAlert.showInlineError(error);
        }
      );
  }

  save() {
    this.submitting = true;
    this.webhookService.editWebhook(this.projectId, this.webhook.id, this.webhook)
      .pipe(finalize(() => (this.submitting = false)))
      .subscribe(
        response => {
          this.inlineAlert.close();
          this.notify.emit();
        },
        error => {
          this.inlineAlert.showInlineError(error);
        }
      );
  }

  setCertValue($event: any): void {
    this.webhook.targets[0].skip_cert_verify = !$event;
  }

  public get isValid(): boolean {
    return (
      this.currentForm &&
      this.currentForm.valid &&
      !this.submitting &&
      !this.checking &&
        this.hasEventType()
    );
  }
  hasChange(): boolean {
    return !compareValue(this.originValue, this.webhook);
  }

  setEventType(eventType) {
    if (this.webhook.event_types.indexOf(eventType) === -1) {
      this.webhook.event_types.push(eventType);
    } else {
      this.webhook.event_types.splice(this.webhook.event_types.findIndex(item => item === eventType), 1);
    }
  }
  getEventType(eventType): boolean {
    return eventType && this.webhook.event_types.indexOf(eventType) !== -1;
  }
  hasEventType(): boolean {
    return this.metadata
      && this.metadata.event_type
      && this.metadata.event_type.length > 0
      && this.webhook.event_types
      && this.webhook.event_types.length > 0;
  }
  eventTypeToText(eventType: string): string {
    return this.webhookService.eventTypeToText(eventType);
  }
}
