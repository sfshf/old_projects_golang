import { storage } from '../common/mmkv';

// 1. unlock password setting
export interface UnlockPasswordSetting {
  passwordHash: string; // 验证解锁密码是否正确，哈希值 keccak_256（ password ）
  supportFingerprint?: boolean; // 指纹验证为本地配置，不会同步到后台
}
export const unlockPasswordSettingKey = 'cache_settings.unlockPassword_';
export const currentUnlockPasswordSetting = (
  userID: number,
): null | UnlockPasswordSetting => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  const curSetting = storage.getString(unlockPasswordSettingKey + suffix);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};
export const updateUnlockPasswordSetting = (
  userID: number,
  obj: null | UnlockPasswordSetting,
) => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  if (!obj) {
    storage.delete(unlockPasswordSettingKey + suffix);
    return;
  }
  storage.set(unlockPasswordSettingKey + suffix, JSON.stringify(obj));
};

// 2. auto lock screen setting
export interface AutoLockSetting {
  timeLag: number;
}
export const autoLockSettingKey = 'cache_settings.unlockPassword.autoLock_';
export const currentAutoLockSetting = (
  userID: number,
): null | AutoLockSetting => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  const curSetting = storage.getString(autoLockSettingKey + suffix);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};
export const updateAutoLockSetting = (
  userID: number,
  obj: null | AutoLockSetting,
) => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  if (!obj) {
    storage.delete(autoLockSettingKey + suffix);
    return;
  }
  storage.set(autoLockSettingKey + suffix, JSON.stringify(obj));
};

// 3. security question setting
export interface SecurityQuestionSetting {
  questions: string;
}
export const securityQuestionSettingKey =
  'cache_settings.unlockPassword.securityQuestion_';
export const currentSecurityQuestionSetting = (
  userID: number,
): null | SecurityQuestionSetting => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  const curSetting = storage.getString(securityQuestionSettingKey + suffix);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};
export const updateSecurityQuestionSetting = (
  userID: number,
  obj: null | SecurityQuestionSetting,
) => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  if (!obj) {
    storage.delete(securityQuestionSettingKey + suffix);
    return;
  }
  storage.set(securityQuestionSettingKey + suffix, JSON.stringify(obj));
};
