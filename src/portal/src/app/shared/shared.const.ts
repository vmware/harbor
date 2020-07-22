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
export const supportedLangs = ['en-us', 'zh-cn', 'zh-tw', 'es-es', 'fr-fr', 'pt-br', 'tr-tr'];
export const enLang = "en-us";
export const languageNames = {
  "en-us": "English",
  "zh-cn": "中文简体",
  "zh-tw": "中文繁體",
  "es-es": "Español",
  "fr-fr": "Français",
  "pt-br": "Português do Brasil",
  "tr-tr": "Türkçe"
};
export const enum AlertType {
  DANGER, WARNING, INFO, SUCCESS
}

export const enum ConfirmationTargets {
  EMPTY,
  PROJECT,
  PROJECT_MEMBER,
  ROBOT_ACCOUNT,
  USER,
  POLICY,
  TOGGLE_CONFIRM,
  TARGET,
  REPOSITORY,
  TAG,
  CONFIG,
  CONFIG_ROUTE,
  CONFIG_TAB,
  HELM_CHART,
  HELM_CHART_VERSION,
  WEBHOOK,
  SCANNER,
  INSTANCE
}

export const enum ActionType {
  ADD_NEW, EDIT
}

export const AdmiralQueryParamKey = "admiral_redirect_url";
export const HarborQueryParamKey = "harbor_redirect_url";
export const CookieKeyOfAdmiral = "admiral.endpoint.latest";
export const enum ConfirmationState {
  NA, CONFIRMED, CANCEL
}
export const enum ConfirmationButtons {
  CONFIRM_CANCEL, YES_NO, DELETE_CANCEL, CLOSE, ENABLE_CANCEL, DISABLE_CANCEL, SWITCH_CANCEL
}

export const ProjectTypes = { 0: 'PROJECT.ALL_PROJECTS', 1: 'PROJECT.PRIVATE_PROJECTS', 2: 'PROJECT.PUBLIC_PROJECTS' };

export const RoleInfo = {
  1: "MEMBER.PROJECT_ADMIN",
  2: "MEMBER.DEVELOPER",
  3: "MEMBER.GUEST",
  4: "MEMBER.PROJECT_MAINTAINER",
  5: "MEMBER.LIMITED_GUEST",
};

export const RoleMapping = {
  "projectAdmin": "MEMBER.PROJECT_ADMIN",
  "maintainer": "MEMBER.PROJECT_MAINTAINER",
  "developer": "MEMBER.DEVELOPER",
  "guest": "MEMBER.GUEST",
  "limitedGuest": "MEMBER.LIMITED_GUEST",
};

export const ProjectRoles = [
  { id: 1, value: "MEMBER.PROJECT_ADMIN" },
  { id: 2, value: "MEMBER.DEVELOPER" },
  { id: 3, value: "MEMBER.GUEST" },
  { id: 4, value: "MEMBER.PROJECT_MAINTAINER" },
  { id: 5, value: "MEMBER.LIMITED_GUEST" },
];

export enum Roles {
  PROJECT_ADMIN = 1,
  PROJECT_MAINTAINER = 4,
  DEVELOPER = 2,
  GUEST = 3,
  LIMITED_GUEST = 5,
  OTHER = 0,
}
export const DefaultHelmIcon = '/images/helm-gray.svg';
export enum ResourceType {
  REPOSITORY = 1,
  CHART_VERSION = 2,
  REPOSITORY_TAG = 3,
}
