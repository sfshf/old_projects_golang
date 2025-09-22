import { storage } from '../common/mmkv';
import { decryptByXchacha20poly1305, decryptByUserPrivateKey } from './cipher';
import { SlarkInfo } from './slark';
import i18n from '../common/locales';
import { Buffer } from 'buffer';
import { post } from '../common/http/post';
import {
  buildPasswordTableIndexesByUserID,
  cancelSharingPasswordsByUserIDAsync,
  clearPasswordTableIndexes,
  deletePasswordsFromFamilyAsync,
  getPasswordByDataIDAsync,
  insertPasswordAsync,
  sharingPasswordByDataIDAsync,
  XoredPassword,
} from '../common/sqlite/dao/password';
import {
  buildRecordTableIndexesByUserID,
  cancelSharingRecordsByUserIDAsync,
  clearRecordTableIndexes,
  deleteRecordsFromFamilyAsync,
  getRecordByDataIDAsync,
  insertRecordAsync,
  sharingRecordByDataIDAsync,
  XoredRecord,
} from '../common/sqlite/dao/record';
import { insertSharedDataMembersAsync } from '../common/sqlite/dao/sharedDataMember';
import { Backup, currentBackup, upsertBackup } from '../common/mmkv/backup';
import { keccak_256 } from '@noble/hashes/sha3';
import { xor_str } from '../common/sqlite/dao/utils';

// 1. other family members
export interface OtherFamilyMember {
  userID: number;
  email: string;
}

export interface OtherFamilyMembersCache {
  list: OtherFamilyMember[];
}

export const otherFamilyMembersKey = 'cache_settings.otherFamilyMembers_';

export const currentOtherFamilyMembers = (
  userID: number,
): null | OtherFamilyMembersCache => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  const curSetting = storage.getString(otherFamilyMembersKey + suffix);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};

export const updateOtherFamilyMembers = (
  userID: number,
  obj: null | OtherFamilyMembersCache,
) => {
  let suffix = 'notSignedIn';
  if (userID > 0) {
    suffix = userID.toString();
  }
  if (!obj) {
    storage.delete(otherFamilyMembersKey + suffix);
    return;
  }
  storage.set(otherFamilyMembersKey + suffix, JSON.stringify(obj));
};

export const deleteOtherFamilyMembers = (userID: number, other: number) => {
  let otherMembers = currentOtherFamilyMembers(userID);
  if (otherMembers) {
    let newOthers: OtherFamilyMember[] = [];
    for (let i = 0; i < otherMembers.list.length; i++) {
      if (otherMembers.list[i].userID !== other) {
        newOthers.push(otherMembers.list[i]);
      }
    }
    updateOtherFamilyMembers(userID, {
      list: newOthers,
    });
  }
};

export const getFamilyMemberEmail = (
  slarkInfo: SlarkInfo,
  userID: number,
): string => {
  const otherMembers = currentOtherFamilyMembers(slarkInfo.userID);
  if (!otherMembers) {
    return '';
  }
  for (let i = 0; i < otherMembers.list.length; i++) {
    if (otherMembers.list[i].userID === userID) {
      return otherMembers.list[i].email;
    }
  }
  return '';
};

// 2. family key
export const getFamilyKey = async (unlockPassword: string): Promise<Buffer> => {
  try {
    const backup: Backup | null = currentBackup();
    if (!backup) {
      throw i18n.t('app.toast.backupRecordError');
    }
    if (!backup.encryptedFamilyKey) {
      return Buffer.from('');
    }
    return decryptByUserPrivateKey(
      Buffer.from(backup.encryptedFamilyKey, 'hex'),
      unlockPassword,
    );
  } catch (error) {
    throw error;
  }
};

export const downloadSharedData = async (
  slarkInfo: SlarkInfo,
  password: string,
): Promise<any> => {
  try {
    const respData = await post('/pswds/downloadSharedData/v1');
    if (!respData || respData.code !== 0) {
      return respData;
    }
    // 更新本地目标用户的数据
    if (respData.data) {
      // 1. encryptedFamilyKey
      const backup: Backup | null = currentBackup();
      if (backup) {
        upsertBackup({
          ...backup,
          encryptedFamilyKey: respData.data.encryptedFamilyKey,
        });
      }
      const passwordHash = Buffer.from(keccak_256(password)).toString('hex');
      const xoredUserID = xor_str(passwordHash, slarkInfo.userID.toString());
      // 2. shared passwords or records
      const familyKey = decryptByUserPrivateKey(
        Buffer.from(respData.data.encryptedFamilyKey, 'hex'),
        password,
      );
      // delete local shared data
      await deletePasswordsFromFamilyAsync(passwordHash, xoredUserID);
      await cancelSharingPasswordsByUserIDAsync(xoredUserID);
      await deleteRecordsFromFamilyAsync(passwordHash, xoredUserID);
      await cancelSharingRecordsByUserIDAsync(xoredUserID);
      if (respData.data.sharingList.length > 0) {
        // update sharing states
        for (let i = 0; i < respData.data.sharingList.length; i++) {
          const content = JSON.parse(
            decryptByXchacha20poly1305(
              familyKey,
              respData.data.sharingList[i].content,
            ),
          );
          if (respData.data.sharingList[i].type === 'password') {
            const item: XoredPassword | null = await getPasswordByDataIDAsync(
              content.dataID,
            );
            if (item) {
              // update
              await sharingPasswordByDataIDAsync(
                item.dataID,
                xor_str(
                  passwordHash,
                  respData.data.sharingList[i].updatedAt.toString(),
                ),
                respData.data.sharingList[i].sharedToAll
                  ? xor_str(passwordHash, '1')
                  : null,
              );
            } else {
              // insert
              await insertPasswordAsync({
                ...content,
                sharedAt: xor_str(
                  passwordHash,
                  respData.data.sharingList[i].updatedAt.toString(),
                ),
                sharedToAll: respData.data.sharingList[i].sharedToAll
                  ? xor_str(passwordHash, '1')
                  : null,
              });
            }
          } else {
            const item: XoredRecord | null = await getRecordByDataIDAsync(
              content.dataID,
            );
            if (item) {
              // update
              await sharingRecordByDataIDAsync(
                item.dataID,
                xor_str(
                  passwordHash,
                  respData.data.sharingList[i].updatedAt.toString(),
                ),
                respData.data.sharingList[i].sharedToAll
                  ? xor_str(passwordHash, '1')
                  : null,
              );
            } else {
              // insert
              await insertRecordAsync({
                ...content,
                sharedAt: xor_str(
                  passwordHash,
                  respData.data.sharingList[i].updatedAt.toString(),
                ),
                sharedToAll: respData.data.sharingList[i].sharedToAll
                  ? xor_str(passwordHash, '1')
                  : null,
              });
            }
          }
        }
      }
      // family share
      if (respData.data.sharedData) {
        await insertSharedDataMembersAsync(
          passwordHash,
          respData.data.sharedData,
        );
      }
      // （5）build indexes
      clearPasswordTableIndexes();
      buildPasswordTableIndexesByUserID(passwordHash);
      clearRecordTableIndexes();
      buildRecordTableIndexesByUserID(passwordHash);
    }
    return respData;
  } catch (error) {
    throw error;
  } finally {
  }
};
