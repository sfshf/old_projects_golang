import moment from 'moment';
import { storage } from '../common/mmkv';
import {
  buildPasswordTableIndexesByUserID,
  cancelSharingPasswordsByUserIDAsync,
  clearPasswordTableIndexes,
  deletePasswordByDataIDAsync,
  deletePasswordsByUserIDAsync,
  deletePasswordsFromFamilyAsync,
  getPasswordByDataIDAsync,
  getPasswordsAsync,
  getPasswordsByUserIDAsync,
  getSharedPasswordsAsync,
  getUploadablePasswordsByUserIDAsync,
  insertPasswordAsync,
  sharingPasswordByDataIDAsync,
  upsertPasswordsAsync,
  XoredPassword,
  xorXoredPasswords,
} from '../common/sqlite/dao/password';
import {
  buildRecordTableIndexesByUserID,
  cancelSharingRecordsByUserIDAsync,
  clearRecordTableIndexes,
  deleteRecordByDataIDAsync,
  deleteRecordsByUserIDAsync,
  deleteRecordsFromFamilyAsync,
  getRecordByDataIDAsync,
  getRecordsAsync,
  getRecordsByUserIDAsync,
  getSharedRecordsAsync,
  getUploadableRecordsByUserIDAsync,
  insertRecordAsync,
  sharingRecordByDataIDAsync,
  upsertRecordsAsync,
  XoredRecord,
  xorXoredRecord,
  xorXoredRecords,
} from '../common/sqlite/dao/record';
import { Password, Record } from '../common/sqlite/schema';
import { SlarkInfo } from './slark';
import { CODE_SUCCESS, Response } from '../common/http';
import { keccak_256 } from '@noble/hashes/sha3';
import { updateOtherFamilyMembers } from './family';
import {
  currentSecurityQuestionSetting,
  currentUnlockPasswordSetting,
  updateSecurityQuestionSetting,
  updateUnlockPasswordSetting,
} from './unlockPassword';
import { insertSharedDataMembersAsync } from '../common/sqlite/dao/sharedDataMember';
import {
  decryptByUnlockPassword,
  decryptByXchacha20poly1305,
  decryptByUserPrivateKey,
  encryptByUnlockPassword,
  userPrivateKey,
} from './cipher';
import { post } from '../common/http/post';
import { xor_hex, xor_str } from '../common/sqlite/dao/utils';
import { Backup, currentBackup, upsertBackup } from '../common/mmkv/backup';

export interface BackupState {
  updatedAt: number;
}
export const backupStateKeyPrefix = 'cache_settings.backupState_';
export const currentBackupState = (userID: number): null | BackupState => {
  const curState = storage.getString(backupStateKeyPrefix + userID);
  if (curState) {
    return JSON.parse(curState);
  }
  return null;
};
export const updateBackupState = (userID: number, obj: null | BackupState) => {
  if (!obj) {
    storage.delete(backupStateKeyPrefix + userID);
    return;
  }
  storage.set(backupStateKeyPrefix + userID, JSON.stringify(obj));
};

export const checkUpdates = async (
  secretHex: string,
  slarkInfo: SlarkInfo,
): Promise<any> => {
  try {
    // （1）handle parameters
    const curBackupState = currentBackupState(slarkInfo.userID);
    const updatedAt = curBackupState ? curBackupState.updatedAt : 0;
    let sharedDataChecksum = 0;
    const sharedPasswords: Password[] = xorXoredPasswords(
      secretHex,
      await getSharedPasswordsAsync(),
    );
    for (let i = 0; i < sharedPasswords.length; i++) {
      let sharedAt = sharedPasswords[i].sharedAt
        ? sharedPasswords[i].sharedAt
        : 0;
      sharedDataChecksum += sharedAt ? sharedAt : 0;
    }
    const sharedRecords: Record[] = xorXoredRecords(
      secretHex,
      await getSharedRecordsAsync(),
    );
    for (let i = 0; i < sharedRecords.length; i++) {
      let sharedAt = sharedRecords[i].sharedAt ? sharedRecords[i].sharedAt : 0;
      sharedDataChecksum += sharedAt ? sharedAt : 0;
    }
    // （2）post request
    return await post('/pswds/checkUpdates/v1', {
      updatedAt,
      sharedDataChecksum,
    });
  } catch (error) {
    throw error;
  } finally {
  }
};
export const syncTimestamp = {
  lastSyncedAt: 0,
};
export const syncBackup = async (
  slarkInfo: SlarkInfo,
  password: string,
): Promise<Response> => {
  try {
    // （1）访问控制
    const nowTS = moment().unix();
    if (nowTS - syncTimestamp.lastSyncedAt < 60) {
      return { code: CODE_SUCCESS, message: 'Ok', debugMessage: '' };
    } else {
      syncTimestamp.lastSyncedAt = nowTS;
    }
    const passwordHash = Buffer.from(keccak_256(password)).toString('hex');
    // （2）checkUpdates
    const respData = await checkUpdates(passwordHash, slarkInfo);
    // （3）handle results
    if (respData.code === 0 && respData.data) {
      // 1. check backup state
      switch (respData.data.state) {
        case 'nothing':
          break;
        case 'upload':
          // 1. 获取用户本地密码记录差值
          let passwords: XoredPassword[] = [];
          if (respData.data.updatedAt > 0) {
            // upload diff password records
            passwords = await getUploadablePasswordsByUserIDAsync(
              xor_str(passwordHash, slarkInfo.userID.toString()),
              respData.data.updatedAt,
            );
          } else {
            // upload all password records
            passwords = await getPasswordsByUserIDAsync(
              xor_str(passwordHash, slarkInfo.userID.toString()),
            );
          }
          // 2. 获取用户本地非密码记录差值
          let records: XoredRecord[] = [];
          if (respData.data.updatedAt > 0) {
            // upload diff non password records
            records = await getUploadableRecordsByUserIDAsync(
              xor_str(passwordHash, slarkInfo.userID.toString()),
              respData.data.updatedAt,
            );
          } else {
            // upload all non password records
            records = await getRecordsByUserIDAsync(
              xor_str(passwordHash, slarkInfo.userID.toString()),
            );
          }
          await uploadData(
            slarkInfo,
            password,
            passwordHash,
            '',
            passwords,
            records,
          );
          break;
        case 'download':
          // handle later
          break;
      }
      // 2. download if need
      if (
        (respData.data.hasFamily && respData.data.sharedDataUpdated) ||
        respData.data.state === 'download'
      ) {
        await downloadData(slarkInfo, password);
      }
      // 3. otherFamilyMembers
      if (respData.data.otherFamilyMembers) {
        updateOtherFamilyMembers(slarkInfo.userID, {
          list: respData.data.otherFamilyMembers,
        });
      }
    }
    // （4）return reponse data；调用方处理 respData.code !== 0 的情况；
    return respData;
  } catch (error) {
    throw error;
  } finally {
  }
};

export const uploadData = async (
  slarkInfo: SlarkInfo,
  password: string,
  passwordHash: string,
  encryptedFamilyKey: string,
  passwords: XoredPassword[],
  records: XoredRecord[],
): Promise<Response> => {
  // 1. handle parameters
  if (!password) {
    throw 'password is empty';
  }
  // （1）user public key
  const backup: Backup | null = currentBackup();
  if (!backup || !backup.userPublicKey) {
    throw 'no user backup record';
  }
  // （2）user encrypted family key
  encryptedFamilyKey = encryptedFamilyKey
    ? encryptedFamilyKey
    : backup.encryptedFamilyKey
    ? backup.encryptedFamilyKey
    : '';

  try {
    // （3） handle passwords
    let pwdList: null | any[] = null;
    if (passwords.length > 0) {
      pwdList = [];
      for (let i = 0; i < passwords.length; i++) {
        pwdList.push({
          dataID: xor_hex(passwordHash, passwords[i].dataID),
          updatedAt: parseInt(xor_hex(passwordHash, passwords[i].updatedAt)),
          content: encryptByUnlockPassword(
            // 整体加密
            password,
            JSON.stringify(passwords[i]),
          ),
        });
      }
    }
    // （4） handle records
    let nprList: null | any[] = null;
    if (records.length > 0) {
      nprList = [];
      for (let i = 0; i < records.length; i++) {
        nprList.push({
          dataID: xor_hex(passwordHash, records[i].dataID),
          updatedAt: parseInt(xor_hex(passwordHash, records[i].updatedAt)),
          type: xor_hex(passwordHash, records[i].recordType),
          content: encryptByUnlockPassword(
            // 整体加密
            password,
            JSON.stringify(records[i]),
          ),
        });
      }
    }
    // 2. post request
    const updatedAt = moment().unix();
    const userPK = userPrivateKey(password).publicKey;
    const respData = await post('/pswds/uploadData/v1', {
      updatedAt,
      passwordHash,
      userPublicKey: userPK.toHex(),
      encryptedFamilyKey,
      pwdList,
      nprList,
    });
    // 3. 更新用户本地数据
    // （1）backup更新时间戳
    updateBackupState(slarkInfo.userID, {
      updatedAt: updatedAt,
    });
    // （2）backup数据库记录
    upsertBackup({
      ...backup,
      userPublicKey: userPK.toHex(),
      encryptedFamilyKey,
    });
    return respData;
  } catch (error) {
    throw error;
  } finally {
  }
};

export const downloadTimestamp = {
  lastDownloadAt: 0,
};
export const downloadData = async (
  slarkInfo: SlarkInfo,
  password: string,
): Promise<Response> => {
  try {
    // （1）访问控制
    const nowTS = moment().unix();
    if (nowTS - downloadTimestamp.lastDownloadAt < 1) {
      return { code: CODE_SUCCESS, message: 'Ok', debugMessage: '' };
    } else {
      downloadTimestamp.lastDownloadAt = nowTS;
    }
    const passwordHash = Buffer.from(keccak_256(password)).toString('hex');
    const xoredUserID = xor_str(passwordHash, slarkInfo.userID.toString());
    // （2）post request
    const curBackupState = currentBackupState(slarkInfo.userID);
    const updatedAt = curBackupState ? curBackupState.updatedAt : 0;
    const respData = await post('/pswds/downloadData/v1', {
      updatedAt,
    });
    // （3）handle results
    if (respData.code === 0 && respData.data) {
      // 1. 更新用户本地备份的更新时间戳
      updateBackupState(slarkInfo.userID, {
        updatedAt: respData.data.updatedAt,
      });
      // 2. 更新 unlockPasswordHash
      // TODO: 设计的逻辑漏洞：同一账号在不同平台上，一个修改了解锁密码，另一个在解锁后才download，缓存的密码与数据所需密码不一致
      const curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      if (curSetting) {
        updateUnlockPasswordSetting(slarkInfo.userID, {
          ...curSetting,
          passwordHash: respData.data.passwordHash,
        });
      } else {
        updateUnlockPasswordSetting(slarkInfo.userID, {
          passwordHash: respData.data.passwordHash,
        });
      }
      // 3. 更新用户本地的数据
      // 3-1. 解锁密码的密保问题
      if (respData.data.securityQuestions) {
        let curSecurityQuestionSetting = currentSecurityQuestionSetting(
          slarkInfo.userID,
        );
        if (curSecurityQuestionSetting) {
          updateSecurityQuestionSetting(slarkInfo.userID, {
            ...curSecurityQuestionSetting,
            questions: respData.data.securityQuestions,
          });
        } else {
          updateSecurityQuestionSetting(slarkInfo.userID, {
            questions: respData.data.securityQuestions,
          });
        }
      }
      // 3-2. 数据库数据
      // （1）password数据
      if (respData.data.pwdList) {
        let strList: string[] = JSON.parse(respData.data.pwdList);
        if (strList && strList.length > 0) {
          let list: XoredPassword[] = [];
          for (let i = 0; i < strList.length; i++) {
            const item = JSON.parse(
              decryptByUnlockPassword(password, strList[i]),
            ); // 整体解密
            if (item) {
              list.push(item);
            }
          }
          if (list.length > 0) {
            let passwords: XoredPassword[] = [];
            list.map(item => {
              passwords.push({
                ...item,
                sharedAt: null, // clear inconsistent sharing state
                sharingMembers: null,
                sharedToAll: null,
              });
            });
            await upsertPasswordsAsync(passwords);
          }
        }
      }
      // （2）record数据
      if (respData.data.nprList) {
        let strList: string[] = JSON.parse(respData.data.nprList);
        if (strList && strList.length > 0) {
          let list: XoredRecord[] = [];
          for (let i = 0; i < strList.length; i++) {
            const item = JSON.parse(
              decryptByUnlockPassword(password, strList[i]),
            ); // 整体解密
            if (item) {
              list.push(item);
            }
          }
          if (list.length > 0) {
            let records: XoredRecord[] = [];
            list.map(item => {
              records.push({
                ...item,
                sharedAt: null, // clear inconsistent sharing state
                sharingMembers: null,
                sharedToAll: null,
              });
            });
            await upsertRecordsAsync(records);
          }
        }
      }
      // NOTE: 本地溢出的数据会被认为，已被其他端进行了删除操作，所以本端会进行该删除操作以达到数据一致；
      if (respData.data.pwdIDList && respData.data.pwdIDList.length > 0) {
        const allRows: XoredPassword[] = await getPasswordsByUserIDAsync(
          xor_str(passwordHash, slarkInfo.userID.toString()),
        );
        for (const row of allRows) {
          if (
            !respData.data.pwdIDList.includes(xor_hex(passwordHash, row.dataID))
          ) {
            await deletePasswordByDataIDAsync(row.dataID);
          }
        }
      } else if (
        !respData.data.pwdIDList ||
        respData.data.pwdIDList.length === 0
      ) {
        await deletePasswordsByUserIDAsync(xoredUserID);
      }
      if (respData.data.nprIDList && respData.data.nprIDList.length > 0) {
        const allRows: XoredRecord[] = await getRecordsByUserIDAsync(
          xor_str(passwordHash, slarkInfo.userID.toString()),
        );
        for (const row of allRows) {
          if (
            !respData.data.nprIDList.includes(xor_hex(passwordHash, row.dataID))
          ) {
            await deleteRecordByDataIDAsync(row.dataID);
          }
        }
      } else if (
        !respData.data.nprIDList ||
        respData.data.nprIDList.length === 0
      ) {
        await deleteRecordsByUserIDAsync(xoredUserID);
      }
      // （3）家庭分享数据
      if (respData.data.encryptedFamilyKey) {
        // 1. encryptedFamilyKey
        const backup: Backup | null = currentBackup();
        if (backup) {
          upsertBackup({
            ...backup,
            encryptedFamilyKey: respData.data.encryptedFamilyKey,
          });
        }
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
      }
      // （4）家庭分享数据的分享成员
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
