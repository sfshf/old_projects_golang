import { storage } from '../mmkv';
// backup schema
export interface Backup {
  createdAt: number;
  updatedAt: number;
  userID: number;
  userPublicKey: string;
  encryptedFamilyKey: string | null;
}

export const backupKey = 'cache_db.backup';

export const currentBackup = (): null | Backup => {
  const curSetting = storage.getString(backupKey);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};

export const upsertBackup = (obj: null | Backup) => {
  if (!obj) {
    storage.delete(backupKey);
    return;
  }
  storage.set(backupKey, JSON.stringify(obj));
};
