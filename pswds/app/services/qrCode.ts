import { storage } from '../common/mmkv';

export interface QrSetting {
  pin: boolean;
  onlyPassword: boolean;
}

export const qrPinSettingKey = 'cache_settings.qrPin_';

export const currentQrPinSetting = (userID: number): null | QrSetting => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  const curSetting = storage.getString(qrPinSettingKey + suffix);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};

export const updateQrPinSetting = (userID: number, obj: null | QrSetting) => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  if (!obj) {
    storage.delete(qrPinSettingKey + suffix);
    return;
  }
  storage.set(qrPinSettingKey + suffix, JSON.stringify(obj));
};
