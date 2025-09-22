import { db } from '..';
import moment from 'moment';
import { insertErrorLog } from '../../log';
import {
  getPasswordByDataIDAsync,
  updatePasswordSharingMembersByDataIDAsync,
  XoredPassword,
} from './password';
import {
  getRecordByDataIDAsync,
  updateRecordSharingMembersByDataIDAsync,
  XoredRecord,
} from './record';
import { xor_str } from './utils';

export const insertSharedDataMembersAsync = async (
  secretHex: string,
  rawList: any,
) => {
  try {
    for (let i = 0; i < rawList.length; i++) {
      const xoredDataID = xor_str(secretHex, rawList[i].dataID);
      const password: null | XoredPassword = await getPasswordByDataIDAsync(
        xoredDataID,
      );
      if (!password) {
        const record: null | XoredRecord = await getRecordByDataIDAsync(
          xoredDataID,
        );
        if (!record) {
          continue;
        } else {
          await updateRecordSharingMembersByDataIDAsync(
            xoredDataID,
            xor_str(secretHex, JSON.stringify(rawList[i].sharingMembers)),
          );
        }
      } else {
        await updatePasswordSharingMembersByDataIDAsync(
          xoredDataID,
          xor_str(secretHex, JSON.stringify(rawList[i].sharingMembers)),
        );
      }
    }
  } catch (error) {
    throw error;
  }
};

// updates
export const updatePasswordSharedDataMembersAsync = async (
  xoredDataID: string,
  xoredSharingMembers: string | null,
) => {
  try {
    const entity: null | XoredPassword = await getPasswordByDataIDAsync(
      xoredDataID,
    );
    if (!entity) {
      return;
    } else {
      await updatePasswordSharingMembersByDataIDAsync(
        xoredDataID,
        xoredSharingMembers,
      );
    }
  } catch (error) {
    throw error;
  }
};

export const updateRecordSharedDataMembersAsync = async (
  xoredDataID: string,
  xoredSharingMembers: string | null,
) => {
  try {
    const entity: null | XoredRecord = await getRecordByDataIDAsync(
      xoredDataID,
    );
    if (!entity) {
      return;
    } else {
      await updateRecordSharingMembersByDataIDAsync(
        xoredDataID,
        xoredSharingMembers,
      );
    }
  } catch (error) {
    throw error;
  }
};

// deletes

export const clearSharedDataMembersAsync = async () => {
  const nowTS = moment();
  try {
    await db.runAsync(
      `UPDATE password 
            SET sharingMembers=$sharingMembers 
            WHERE 1=1;`,
      {
        $sharingMembers: null,
      },
    );
    await db.runAsync(
      `UPDATE record 
            SET sharingMembers=$sharingMembers 
            WHERE 1=1;`,
      {
        $sharingMembers: null,
      },
    );
  } catch (error) {
    insertErrorLog({
      level: 'error',
      timestamp: nowTS.valueOf(),
      message: error as string,
    });
    throw error;
  } finally {
  }
};

export const deleteSharedDataMembersByDataIDAsync = async (dataID: string) => {
  const nowTS = moment();
  try {
    const password: null | XoredPassword = await getPasswordByDataIDAsync(
      dataID,
    );
    if (!password) {
      const record: null | XoredRecord = await getRecordByDataIDAsync(dataID);
      if (!record) {
        return;
      } else {
        await updateRecordSharingMembersByDataIDAsync(dataID, '');
      }
    } else {
      await updatePasswordSharingMembersByDataIDAsync(dataID, '');
    }
  } catch (error) {
    insertErrorLog({
      level: 'error',
      timestamp: nowTS.valueOf(),
      message: error as string,
    });
    throw error;
  }
};

export const deleteSharedDataMembersByUserIDAsync = async (
  xoredUserID: string,
) => {
  const nowTS = moment();
  try {
    await db.runAsync(
      `UPDATE password 
            SET sharingMembers=$sharingMembers 
            WHERE userID=$userID;`,
      {
        $sharingMembers: null,
        $userID: xoredUserID,
      },
    );
    await db.runAsync(
      `UPDATE record 
            SET sharingMembers=$sharingMembers 
             WHERE userID=$userID;`,
      {
        $sharingMembers: null,
        $userID: xoredUserID,
      },
    );
  } catch (error) {
    insertErrorLog({
      level: 'error',
      timestamp: nowTS.valueOf(),
      message: error as string,
    });
    throw error;
  } finally {
  }
};

// gets
