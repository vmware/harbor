import { Component, OnInit, Input, OnChanges, SimpleChanges } from '@angular/core';
import { DatePipe } from "@angular/common";
import { TranslateService } from "@ngx-translate/core";
import { Artifact } from "../../../../../../ng-swagger-gen/models/artifact";
import { isObject } from "../../../../../lib/utils/utils";

enum Types {
  CREATED = 'created',
  TYPE = 'type',
  MEDIA_TYPE = 'media_type',
  MANIFEST_MEDIA_TYPE = 'manifest_media_type',
  DIGEST = 'digest',
  SIZE = 'size',
  PUSH_TIME = 'push_time',
  PULL_TIME = 'pull_time',
}

@Component({
  selector: 'artifact-common-properties',
  templateUrl: './artifact-common-properties.component.html',
  styleUrls: ['./artifact-common-properties.component.scss']
})
export class ArtifactCommonPropertiesComponent implements OnInit, OnChanges {
  @Input() artifactDetails: Artifact;
  commonProperties: { [key: string]: any } = {};

  constructor() {
  }

  ngOnInit() {
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes && changes["artifactDetails"]) {
      if (this.artifactDetails) {
        Object.assign(this.commonProperties, this.artifactDetails.extra_attrs, this.artifactDetails.annotations);
        for (let name in this.commonProperties) {
          if (this.commonProperties.hasOwnProperty(name)) {
            if (typeof (this.commonProperties[name]) === 'object') {
              if (this.commonProperties[name] === null) {
                this.commonProperties[name] = '';
              } else {
                this.commonProperties[name] = JSON.stringify(this.commonProperties[name]);
              }
            }
            if (name === Types.CREATED) {
              this.commonProperties[name] = new DatePipe('en-us')
                .transform(this.commonProperties[name], 'short');
            }
          }
        }
      }
    }
  }

  hasCommonProperties(): boolean {
    return JSON.stringify(this.commonProperties) !== '{}';
  }
}
