import { storage } from '../common/mmkv';

export interface SlarkInfo {
  userID: number;
  nickname: string;
  email: string;
  phone: string;
  lSessionID: string;
}
export const slarkInfoSettingKey = 'cache_settings.signin.slarkInfo';
export const currentSlarkInfo = (): null | SlarkInfo => {
  const curSlarkInfo = storage.getString(slarkInfoSettingKey);
  if (curSlarkInfo) {
    return JSON.parse(curSlarkInfo);
  }
  return null;
};
export const updateSlarkInfo = (obj: null | SlarkInfo) => {
  if (!obj) {
    storage.delete(slarkInfoSettingKey);
    return;
  }
  storage.set(slarkInfoSettingKey, JSON.stringify(obj));
};
