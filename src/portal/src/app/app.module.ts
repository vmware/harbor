// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { BrowserModule } from '@angular/platform-browser';
import { NgModule, APP_INITIALIZER, LOCALE_ID, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { AppComponent } from './app.component';
import { InterceptHttpService } from './services/intercept-http.service';

import { BaseModule } from './base/base.module';
import { HarborRoutingModule } from './harbor-routing.module';
import { SharedModule } from './shared/shared.module';
import { AccountModule } from './account/account.module';
import { SignInModule } from './sign-in/sign-in.module';
import { ConfigurationModule } from './config/config.module';
import { DeveloperCenterModule } from './dev-center/dev-center.module';
import { registerLocaleData } from '@angular/common';

import { TranslateService } from "@ngx-translate/core";
import { AppConfigService } from './services/app-config.service';
import { SkinableConfig } from "./services/skinable-config.service";
import { ProjectConfigComponent } from './project/project-config/project-config.component';

import zh from '@angular/common/locales/zh-Hans';
import es from '@angular/common/locales/es';
import localeFr from '@angular/common/locales/fr';
import localePt from '@angular/common/locales/pt-PT';
import localeTr from '@angular/common/locales/tr';
import { DevCenterComponent } from './dev-center/dev-center.component';
import { VulnerabilityPageComponent } from './vulnerability-page/vulnerability-page.component';
import { GcPageComponent } from './gc-page/gc-page.component';
import { OidcOnboardModule } from './oidc-onboard/oidc-onboard.module';
import { LicenseModule } from './license/license.module';
import { InterrogationServicesComponent } from "./interrogation-services/interrogation-services.component";
import { LabelsComponent } from './labels/labels.component';
import { ProjectQuotasComponent } from './project-quotas/project-quotas.component';
import { HarborLibraryModule } from "../lib/harbor-library.module";
import { HTTP_INTERCEPTORS } from '@angular/common/http';
import { AllPipesModule } from './all-pipes/all-pipes.module';
import { DistributionModule } from './distribution/distribution.module';
registerLocaleData(zh, 'zh-cn');
registerLocaleData(es, 'es-es');
registerLocaleData(localeFr, 'fr-fr');
registerLocaleData(localePt, 'pt-br');
registerLocaleData(localeTr, 'tr-tr');

export function initConfig(configService: AppConfigService, skinableService: SkinableConfig) {
    return () => {
        skinableService.getCustomFile().subscribe();
        configService.load().subscribe();
    };
}

export function getCurrentLanguage(translateService: TranslateService) {
    return translateService.currentLang;
}

@NgModule({
    declarations: [
        AppComponent,
        ProjectConfigComponent,
        VulnerabilityPageComponent,
        GcPageComponent,
        InterrogationServicesComponent,
        LabelsComponent,
        ProjectQuotasComponent
    ],
    imports: [
        BrowserModule,
        SharedModule,
        BaseModule,
        AccountModule,
        SignInModule,
        HarborRoutingModule,
        ConfigurationModule,
        DeveloperCenterModule,
        OidcOnboardModule,
        LicenseModule,
        HarborLibraryModule,
        AllPipesModule,
        DistributionModule,
    ],
    exports: [
    ],
    providers: [
        AppConfigService,
        SkinableConfig,
        {
            provide: APP_INITIALIZER,
            useFactory: initConfig,
            deps: [AppConfigService, SkinableConfig],
            multi: true
        },
        { provide: LOCALE_ID, useValue: "en-US" },
        { provide: HTTP_INTERCEPTORS, useClass: InterceptHttpService, multi: true }

    ],
    schemas: [
        CUSTOM_ELEMENTS_SCHEMA
    ],
    bootstrap: [AppComponent]
})
export class AppModule { }
