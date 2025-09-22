import { storage } from '../mmkv';

export const DefaultCookieSessionKey = 'LSessionID';
export type Response = {
  code: number;
  message: string;
  debugMessage: string;
  data?: any;
  lSessionID?: string;
};
export const CODE_SUCCESS = 0;
// response code
export const ResponseCode_DataPullAhead = 108040013; // 后台数据有更新；即本地数据滞后了导致的请求失败
export const ResponseCode_NotFound = 108040400; // 后台服务未找到相关资源；
export const ResponseCode_NoBackup = 108040012; // 后台没有当前用户的数据备份；
export const ResponseCode_DataFallBehind = 108040014; // 后台数据滞后；
export const ResponseCode_NotSetSecurityQuestions = 108040015; // 未设置密保
export const ResponseCode_ResourceLimit = 108040050; // 资源限制

export interface UseGoCache {
  useGo: boolean;
}

export const debugUseGoKey = 'cache_debug.useGo';

export const currentDebugUseGo = (): null | UseGoCache => {
  const curSetting = storage.getString(debugUseGoKey);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};

export const updateDebugUseGo = (obj: null | UseGoCache) => {
  if (!obj) {
    storage.delete(debugUseGoKey);
    return;
  }
  storage.set(debugUseGoKey, JSON.stringify(obj));
};
