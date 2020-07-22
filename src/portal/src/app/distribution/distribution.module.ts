import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DistributionInstancesComponent } from './distribution-instances/distribution-instances.component';
import { DistributionSetupModalComponent } from './distribution-setup-modal/distribution-setup-modal.component';
import { SharedModule } from '../shared/shared.module';

@NgModule({
  imports: [CommonModule, SharedModule],
  declarations: [
    DistributionSetupModalComponent,
    DistributionInstancesComponent
  ],
})
export class DistributionModule {}
