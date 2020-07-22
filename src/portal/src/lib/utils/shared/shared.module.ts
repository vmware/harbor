import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClientModule, HttpClientXsrfModule, HttpClient, HttpXsrfTokenExtractor } from '@angular/common/http';
import { ClarityModule } from '@clr/angular';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { TranslateModule, TranslateLoader, MissingTranslationHandler } from '@ngx-translate/core';
import { CookieService, CookieModule } from 'ngx-cookie';
import { MarkdownModule } from 'ngx-markdown';
import { IServiceConfig, SERVICE_CONFIG } from "../../entities/service.config";
import { TranslateHttpLoader } from "@ngx-translate/http-loader";
import { MyMissingTranslationHandler } from "../../i18n/missing-trans.handler";
import { TranslatorJsonLoader } from "../../i18n/local-json.loader";
import { ClipboardModule } from "../../components/third-party/ngx-clipboard";
import { environment } from '../../../environments/environment';

export function GeneralTranslatorLoader(http: HttpClient, config: IServiceConfig) {
    if (config && config.langMessageLoader === 'http') {
        const prefix: string = config.langMessagePathForHttpLoader ? config.langMessagePathForHttpLoader : "i18n/lang/";
        let suffix: string = config.langMessageFileSuffixForHttpLoader ? config.langMessageFileSuffixForHttpLoader : "-lang.json";
        if (environment && environment.buildTimestamp) {
            suffix += `?buildTimeStamp=${environment.buildTimestamp}`;
        }
        return new TranslateHttpLoader(http, prefix, suffix);
    } else {
        return new TranslatorJsonLoader(config);
    }
}
/**
 *
 * Module for sharing common modules
 *
 **
 * class SharedModule
 */
@NgModule({
    imports: [
        CommonModule,
        HttpClientModule,
        HttpClientXsrfModule.withOptions({
            cookieName: '__csrf',
            headerName: 'X-Harbor-CSRF-Token'
        }),
        FormsModule,
        ReactiveFormsModule,
        ClipboardModule,
        ClarityModule,
        CookieModule.forRoot(),
        MarkdownModule.forRoot(),
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useFactory: (GeneralTranslatorLoader),
                deps: [HttpClient, SERVICE_CONFIG]
            },
            missingTranslationHandler: {
                provide: MissingTranslationHandler,
                useClass: MyMissingTranslationHandler
            }
        }),
    ],
    exports: [
        CommonModule,
        HttpClientModule,
        FormsModule,
        ReactiveFormsModule,
        ClipboardModule,
        ClarityModule,
        CookieModule,
        MarkdownModule,
        TranslateModule,
    ],
    providers: [
        CookieService,
    ]
})
export class SharedModule { }
