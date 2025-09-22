import { storage } from '../mmkv';
// database version schema
export interface DBVersion {
  createdAt: number;
  updatedAt: number;
  version: number;
}

export const dbVersionKey = 'cache_db.version';

export const currentDBVersion = (): null | DBVersion => {
  const curSetting = storage.getString(dbVersionKey);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};

export const upsertDBVersion = (obj: null | DBVersion) => {
  if (!obj) {
    storage.delete(dbVersionKey);
    return;
  }
  storage.set(dbVersionKey, JSON.stringify(obj));
};
