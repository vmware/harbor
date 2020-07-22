export class AuthMode {
  static NONE = 'NONE';
  static BASIC = 'BASIC';
  static OAUTH = 'OAUTH';
  static CUSTOM = 'CUSTOM';
}

export enum PreheatingStatusEnum {
  // front status
  NOT_PREHEATED = 'NOT_PREHEATED',
  // back-end status
  PENDING = 'PENDING',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
  FAIL = 'FAIL',
}
