import { SQLiteRunResult } from 'expo-sqlite';
import { db } from '..';
import { Password } from '../schema';
import moment from 'moment';
import { insertErrorLog } from '../../log';
import { xor_hex, xor_str } from './utils';
import { currentUnlockPasswordSetting } from '../../../services/unlockPassword';

export interface XoredPassword {
  dataID: string;
  createdAt: string;
  updatedAt: string;
  userID: string;
  title: string;
  website: string | null;
  username: string;
  password: string;
  notes: string | null;
  others: string | null;
  usedAt: string | null;
  usedCount: string | null;
  iconBgColor: string | null;
  sharedAt: string | null;
  sharedToAll: string | null;
  sharingMembers: string | null;
}
export const xorPassword = (
  secretHex: string,
  entity: Password,
): XoredPassword => {
  return {
    dataID: xor_str(secretHex, entity.dataID),
    createdAt: xor_str(secretHex, entity.createdAt.toString()),
    updatedAt: xor_str(secretHex, entity.updatedAt.toString()),
    userID: xor_str(secretHex, entity.userID.toString()),
    title: xor_str(secretHex, entity.title),
    website: entity.website ? xor_str(secretHex, entity.website) : null,
    username: xor_str(secretHex, entity.username),
    password: xor_str(secretHex, entity.password),
    notes: entity.notes ? xor_str(secretHex, entity.notes) : null,
    others: entity.others ? xor_str(secretHex, entity.others) : null,
    usedAt: entity.usedAt ? xor_str(secretHex, entity.usedAt.toString()) : null,
    usedCount: entity.usedCount
      ? xor_str(secretHex, entity.usedCount.toString())
      : null,
    iconBgColor: entity.iconBgColor
      ? xor_str(secretHex, entity.iconBgColor.toString())
      : null,
    sharingMembers: entity.sharingMembers
      ? xor_str(secretHex, entity.sharingMembers)
      : null,
    sharedAt: entity.sharedAt
      ? xor_str(secretHex, entity.sharedAt.toString())
      : null,
    sharedToAll: entity.sharedToAll
      ? xor_str(secretHex, entity.sharedToAll.toString())
      : null,
  };
};
export const xorPasswords = (
  secretHex: string,
  entities: Password[],
): XoredPassword[] => {
  let result: XoredPassword[] = [];
  entities.map(item => {
    result.push(xorPassword(secretHex, item));
  });
  return result;
};
export const xorXoredPassword = (
  secretHex: string,
  entity: XoredPassword,
): Password => {
  return {
    dataID: xor_hex(secretHex, entity.dataID),
    createdAt: parseInt(xor_hex(secretHex, entity.createdAt)),
    updatedAt: parseInt(xor_hex(secretHex, entity.updatedAt)),
    userID: parseInt(xor_hex(secretHex, entity.userID)),
    title: xor_hex(secretHex, entity.title),
    website: entity.website ? xor_hex(secretHex, entity.website) : null,
    username: xor_hex(secretHex, entity.username),
    password: xor_hex(secretHex, entity.password),
    notes: entity.notes ? xor_hex(secretHex, entity.notes) : null,
    others: entity.others ? xor_hex(secretHex, entity.others) : null,
    usedAt: entity.usedAt ? parseInt(xor_hex(secretHex, entity.usedAt)) : null,
    usedCount: entity.usedCount
      ? parseInt(xor_hex(secretHex, entity.usedCount))
      : null,
    iconBgColor: entity.iconBgColor
      ? parseInt(xor_hex(secretHex, entity.iconBgColor))
      : null,
    sharedAt: entity.sharedAt
      ? parseInt(xor_hex(secretHex, entity.sharedAt))
      : null,
    sharedToAll: entity.sharedToAll
      ? parseInt(xor_hex(secretHex, entity.sharedToAll))
      : null,
    sharingMembers: entity.sharingMembers
      ? xor_hex(secretHex, entity.sharingMembers)
      : null,
  };
};
export const xorXoredPasswords = (
  secretHex: string,
  entities: XoredPassword[],
): Password[] => {
  let result: Password[] = [];
  entities.map(item => {
    result.push(xorXoredPassword(secretHex, item));
  });
  return result;
};

// inserts

export const insertPasswordAsyncStatement = async () => {
  return await db.prepareAsync(
    `INSERT INTO password (
      dataID, 
      createdAt, 
      updatedAt, 
      userID, 
      title, 
      website, 
      username, 
      password, 
      notes, 
      others, 
      usedAt, 
      usedCount, 
      iconBgColor, 
      sharingMembers, 
      sharedAt, 
      sharedToAll) VALUES (
      $dataID, 
      $createdAt, 
      $updatedAt, 
      $userID, 
      $title, 
      $website, 
      $username, 
      $password, 
      $notes, 
      $others, 
      $usedAt, 
      $usedCount, 
      $iconBgColor, 
      $sharingMembers, 
      $sharedAt, 
      $sharedToAll);`,
  );
};

export const passwordToInsertable = (entity: XoredPassword) => {
  return {
    $dataID: entity.dataID,
    $createdAt: entity.createdAt,
    $updatedAt: entity.updatedAt,
    $userID: entity.userID,
    $title: entity.title,
    $website: entity.website,
    $username: entity.username,
    $password: entity.password,
    $notes: entity.notes,
    $others: entity.others,
    $usedAt: entity.usedAt,
    $usedCount: entity.usedCount,
    $iconBgColor: entity.iconBgColor,
    $sharingMembers: entity.sharingMembers,
    $sharedAt: entity.sharedAt,
    $sharedToAll: entity.sharedToAll,
  };
};

export const insertPasswordAsync = async (entity: XoredPassword) => {
  const statement = await insertPasswordAsyncStatement();
  const nowTS = moment();
  try {
    await statement.executeAsync<XoredPassword>(passwordToInsertable(entity));
  } catch (error) {
    insertErrorLog({
      level: 'error',
      timestamp: nowTS.valueOf(),
      message: error as string,
    });
    throw error;
  } finally {
    await statement.finalizeAsync();
  }
};

export const insertPasswordsAsync = async (entities: XoredPassword[]) => {
  const statement = await insertPasswordAsyncStatement();
  const nowTS = moment();
  try {
    entities.map(async item => {
      await statement.executeAsync<XoredPassword>(passwordToInsertable(item));
    });
  } catch (error) {
    insertErrorLog({
      level: 'error',
      timestamp: nowTS.valueOf(),
      message: error as string,
    });
    throw error;
  } finally {
    await statement.finalizeAsync();
  }
};

export const upsertPasswordsAsync = async (entities: XoredPassword[]) => {
  const nowTS = moment();
  try {
    entities.map(async item => {
      let record: XoredPassword | null = await getPasswordByDataIDAsync(
        item.dataID,
      );
      if (record) {
        await updatePasswordAsync(item);
      } else {
        await insertPasswordAsync(item);
      }
    });
  } catch (error) {
    throw error;
  } finally {
  }
};

// updates

export const updatePasswordAsyncStatement = async () => {
  return await db.prepareAsync(
    `UPDATE password 
      SET updatedAt=$updatedAt,
          title=$title,
          website=$website, 
          username=$username, 
          password=$password, 
          notes=$notes, 
          others=$others, 
          usedAt=$usedAt, 
          usedCount=$usedCount, 
          iconBgColor=$iconBgColor, 
          sharingMembers=$sharingMembers, 
          sharedAt=$sharedAt, 
          sharedToAll=$sharedToAll
        WHERE dataID=$dataID;`,
  );
};

export const passwordToUpdatable = (entity: XoredPassword) => {
  return {
    $dataID: entity.dataID,
    $updatedAt: entity.updatedAt,
    $title: entity.title,
    $website: entity.website,
    $username: entity.username,
    $password: entity.password,
    $notes: entity.notes,
    $others: entity.others,
    $usedAt: entity.usedAt,
    $usedCount: entity.usedCount,
    $iconBgColor: entity.iconBgColor,
    $sharingMembers: entity.sharingMembers,
    $sharedAt: entity.sharedAt,
    $sharedToAll: entity.sharedToAll,
  };
};

export const updatePasswordAsync = async (entity: XoredPassword) => {
  const statement = await updatePasswordAsyncStatement();
  const nowTS = moment();
  try {
    await statement.executeAsync<XoredPassword>(passwordToUpdatable(entity));
  } catch (error) {
    insertErrorLog({
      level: 'error',
      timestamp: nowTS.valueOf(),
      message: error as string,
    });
    throw error;
  } finally {
    await statement.finalizeAsync();
  }
};

export const clearUnsignedInPasswordsAsync = async (
  xoredUnsignedUserID: string,
  secretHex: string,
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE password 
        SET userID=$userID,
          updatedAt=$updatedAt 
        WHERE userID=$oldUserID;`,
      {
        $oldUserID: xoredUnsignedUserID,
        $userID: xoredUserID,
        $updatedAt: xor_str(secretHex, nowTS.unix().toString()),
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

export const cancelSharingPasswordsByUserIDAsync = async (
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE password 
        SET sharedAt=$sharedAt, 
            sharingMembers=$sharingMembers, 
            sharedToAll=$sharedToAll 
        WHERE userID=$userID;`,
      {
        $userID: xoredUserID,
        $sharedAt: null,
        $sharingMembers: null,
        $sharedToAll: null,
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

export const sharingPasswordByDataIDAsync = async (
  xoredDataID: string,
  xoredSharedAt: string | null,
  xoredSharedToAll: string | null,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE password 
        SET sharedAt=$sharedAt, 
            sharedToAll=$sharedToAll 
        WHERE dataID=$dataID;`,
      {
        $dataID: xoredDataID,
        $sharedAt: xoredSharedAt,
        $sharedToAll: xoredSharedToAll,
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

export const addPasswordUseByDataIDAsync = async (
  secretHex: string,
  xoredDataID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment().unix();
  try {
    const entity: XoredPassword | null = await db.getFirstAsync(
      'SELECT * FROM password WHERE dataID = ?;',
      xoredDataID,
    );
    if (!entity) {
      throw 'invalid password data id ' + xoredDataID;
    }
    return db.runAsync(
      `UPDATE password 
      SET updatedAt=$updatedAt, 
          usedAt=$usedAt, 
          usedCount=$usedCount 
      WHERE dataID=$dataID;`,
      {
        $dataID: xoredDataID,
        $updatedAt: xor_str(secretHex, nowTS.toString()),
        $usedAt: xor_str(secretHex, nowTS.toString()),
        $usedCount: xor_str(
          secretHex,
          (entity.usedCount
            ? parseInt(xor_str(secretHex, entity.usedCount)) + 1
            : 1
          ).toString(),
        ),
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

export const updatePasswordSharingMembersByDataIDAsync = async (
  xoredDataID: string,
  xoredSharingMembers: string | null,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE password 
        SET sharingMembers=$sharingMembers 
        WHERE dataID=$dataID;`,
      {
        $dataID: xoredDataID,
        $sharingMembers: xoredSharingMembers,
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

export const changeUnlockPasswordAsync_Password = async (
  oldPasswordHash: string,
  newPasswordHash: string,
) => {
  const nowTS = moment();
  try {
    // 1. remove old indexes
    clearPasswordTableIndexes();
    // 2. iterate
    for await (const row of db.getEachAsync<XoredPassword>(
      'SELECT * FROM password;',
    )) {
      // 2-1. delete
      await deletePasswordByDataIDAsync(row.dataID);
      // 2-2. insert new entity
      const plainEntity = xorXoredPassword(oldPasswordHash, row);
      const xoredEntity = xorPassword(newPasswordHash, plainEntity);
      await insertPasswordAsync(xoredEntity);
      // 2-3. insert indexes
      insertPasswordTableIndex_UpdatedAt(newPasswordHash, xoredEntity);
      insertPasswordTableIndex_UsedAt(newPasswordHash, xoredEntity);
      insertPasswordTableIndex_UsedCount(newPasswordHash, xoredEntity);
    }
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

// deletes

export const deletePasswordsFromFamilyAsync = async (
  secretHex: string,
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `DELETE FROM password WHERE userID != $userID AND userID != $unsigned;`,
      {
        $userID: xoredUserID,
        $unsigned: xor_str(secretHex, '-1'),
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

export const deletePasswordsByUserIDAsync = async (
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync('DELETE FROM password WHERE userID = $userID;', {
      $userID: xoredUserID,
    });
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

export const deletePasswordByDataIDAsync = async (
  xoredDataID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync('DELETE FROM password WHERE dataID = $dataID;', {
      $dataID: xoredDataID,
    });
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

export const getPasswordByDataIDAsync = async (
  xoredDataID: string,
): Promise<XoredPassword | null> => {
  const nowTS = moment();
  try {
    return db.getFirstAsync(
      'SELECT * FROM password WHERE dataID = ?;',
      xoredDataID,
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

export const getPasswordsAsync = async (): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredPassword[] = [];
    for (let i = 0; i < passwordTableIndex_UpdatedAt.length; i++) {
      const result = await getPasswordByDataIDAsync(
        passwordTableIndex_UpdatedAt[i].xoredDataID,
      );
      if (result) {
        list.push(result);
      }
    }
    return list;
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

export const searchPasswordsByTitleAsync = async (
  query: string,
  limit: number,
): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredPassword[] = [];
    for (let i = 0; i < passwordTableIndex_UpdatedAt.length; i++) {
      if (i === limit) {
        break;
      }
      if (passwordTableIndex_UpdatedAt[i].title.includes(query)) {
        const result = await getPasswordByDataIDAsync(
          passwordTableIndex_UpdatedAt[i].xoredDataID,
        );
        if (result) {
          list.push(result);
        }
      }
    }
    return list;
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

export const getMostUsedPasswordsAsync = async (
  limit: number,
): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredPassword[] = [];
    for (let i = 0; i < passwordTableIndex_UsedCount.length; i++) {
      if (i === limit) {
        break;
      }
      const result = await getPasswordByDataIDAsync(
        passwordTableIndex_UsedCount[i].xoredDataID,
      );
      if (result) {
        list.push(result);
      }
    }
    return list;
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

export const getLatestUsedPasswordsAsync = async (
  limit: number,
): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredPassword[] = [];
    for (let i = 0; i < passwordTableIndex_UsedAt.length; i++) {
      if (i === limit) {
        break;
      }
      const result = await getPasswordByDataIDAsync(
        passwordTableIndex_UsedAt[i].xoredDataID,
      );
      if (result) {
        list.push(result);
      }
    }
    return list;
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

export const getLatestUpdatedPasswordsAsync = async (
  limit: number,
): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredPassword[] = [];
    for (let i = 0; i < passwordTableIndex_UpdatedAt.length; i++) {
      if (i === limit) {
        break;
      }
      const result = await getPasswordByDataIDAsync(
        passwordTableIndex_UpdatedAt[i].xoredDataID,
      );
      if (result) {
        list.push(result);
      }
    }
    return list;
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

export const getFrequentlyUsedUsernamesByUserIDAsync = async (
  xoredUserID: string,
): Promise<string[]> => {
  const nowTS = moment();
  try {
    const result: any[] = db.getAllSync(
      `SELECT username FROM password WHERE userID = ? GROUP BY username ORDER BY COUNT(username) DESC LIMIT 3;`,
      xoredUserID,
    );
    if (result.length === 0) {
      return [];
    }
    const list: string[] = [];
    for (let i = 0; i < result.length; i++) {
      list.push(result[i].username);
    }
    return list;
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

export const getPasswordsByUserIDAsync = async (
  xoredUserID: string,
): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    return db.getAllAsync(
      'SELECT * FROM password WHERE userID = ?;',
      xoredUserID,
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

export const getSharedPasswordsAsync = async (): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    return db.getAllAsync(`SELECT * FROM password WHERE sharedAt IS NOT NULL;`);
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

export const getUploadablePasswordsByUserIDAsync = async (
  xoredUserID: string,
  updatedAt: number,
): Promise<XoredPassword[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredPassword[] = [];
    for (let i = 0; i < passwordTableIndex_UpdatedAt.length; i++) {
      if (passwordTableIndex_UpdatedAt[i].xoredUserID !== xoredUserID) {
        continue;
      }
      if (passwordTableIndex_UpdatedAt[i].updatedAt > updatedAt) {
        const result = await getPasswordByDataIDAsync(
          passwordTableIndex_UpdatedAt[i].xoredDataID,
        );
        if (result) {
          list.push(result);
        }
      }
    }
    return list;
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

export const getPasswordSharedDataMembersSync = (
  secretHex: string,
  xoredDataID: string,
): null | number[] => {
  const nowTS = moment();
  try {
    const entity: XoredPassword | null = db.getFirstSync(
      'SELECT * FROM password WHERE dataID = ?;',
      xoredDataID,
    );
    if (entity && entity.sharingMembers) {
      return JSON.parse(xor_hex(secretHex, entity.sharingMembers));
    }
    return null;
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

// 索引

export const buildPasswordTableIndexesByUserID = (secretHex: string) => {
  const nowTS = moment();
  try {
    console.log(
      '>>>>>>>>>>> 开始构建Password的索引：',
      nowTS.local().format('YYYY-MM-DD HH:mm:ss.SSS'),
    );
    for (const row of db.getEachSync<XoredPassword>(
      'SELECT dataID, title, updatedAt, usedCount, usedAt FROM password;',
    )) {
      insertPasswordTableIndexes(secretHex, row);
    }
    const overTS = moment();
    console.log(
      '>>>>>>>>>>> 结束构建Password的索引：',
      overTS.local().format('YYYY-MM-DD HH:mm:ss.SSS'),
      '； 用时（单位：s）：',
      moment.duration(overTS.diff(nowTS)).asSeconds(),
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

export const insertPasswordTableIndexes = (
  secretHex: string,
  row: XoredPassword,
) => {
  insertPasswordTableIndex_UpdatedAt(secretHex, row);
  insertPasswordTableIndex_UsedAt(secretHex, row);
  insertPasswordTableIndex_UsedCount(secretHex, row);
};
export const clearPasswordTableIndexes = () => {
  passwordTableIndex_UsedCount = [];
  passwordTableIndex_UpdatedAt = [];
  passwordTableIndex_UsedAt = [];
};

const indexExists = (indexes: any[], xoredDataID: string): boolean => {
  for (let i = 0; i < indexes.length; i++) {
    if (indexes[i].xoredDataID === xoredDataID) {
      return true;
    }
  }
  return false;
};

// index usedCount：从大到小排序
interface passwordUsedCountIndex {
  usedCount: number | null;
  xoredDataID: string;
  xoredUserID: string;
}
let passwordTableIndex_UsedCount: passwordUsedCountIndex[] = [];
const insertPasswordTableIndex_UsedCount = (
  secretHex: string,
  row: XoredPassword,
) => {
  if (indexExists(passwordTableIndex_UsedCount, row.dataID)) {
    return;
  }
  if (row.usedCount === null) {
    passwordTableIndex_UsedCount.push({
      xoredDataID: row.dataID,
      usedCount: row.usedCount,
      xoredUserID: row.userID,
    });
    return;
  }
  let usedCount = parseInt(xor_str(secretHex, row.usedCount));
  for (let i = 0; i < passwordTableIndex_UsedCount.length; i++) {
    if (
      passwordTableIndex_UsedCount[i].usedCount === null ||
      passwordTableIndex_UsedCount[i].usedCount! < usedCount
    ) {
      let newIndex: passwordUsedCountIndex[] = [];
      if (i === 0) {
        newIndex.push({
          xoredDataID: row.dataID,
          usedCount: usedCount,
          xoredUserID: row.userID,
        });
      } else {
        newIndex = [...passwordTableIndex_UsedCount.slice(0, i)];
        newIndex.push({
          xoredDataID: row.dataID,
          usedCount: usedCount,
          xoredUserID: row.userID,
        });
      }
      newIndex.push(...passwordTableIndex_UsedCount.slice(i));
      passwordTableIndex_UsedCount = newIndex;
      return;
    }
  }
  passwordTableIndex_UsedCount.push({
    xoredDataID: row.dataID,
    usedCount: usedCount,
    xoredUserID: row.userID,
  });
};
const splicePasswordTableIndex_UsedCount = (xoredDataID: string) => {
  for (let i = 0; i < passwordTableIndex_UsedCount.length; i++) {
    if (passwordTableIndex_UsedCount[i].xoredDataID == xoredDataID) {
      let newIndex: passwordUsedCountIndex[] = [];
      if (i === 0) {
        newIndex.push(...passwordTableIndex_UsedCount.slice(1));
      } else {
        newIndex.push(
          ...passwordTableIndex_UsedCount.slice(0, i),
          ...passwordTableIndex_UsedCount.slice(i + 1),
        );
      }
      passwordTableIndex_UsedCount = newIndex;
      return;
    }
  }
};

// index usedAt：从大到小
interface passwordUsedAtIndex {
  usedAt: number | null;
  xoredDataID: string;
  xoredUserID: string;
}
let passwordTableIndex_UsedAt: passwordUsedAtIndex[] = [];
const insertPasswordTableIndex_UsedAt = (
  secretHex: string,
  row: XoredPassword,
) => {
  if (indexExists(passwordTableIndex_UsedAt, row.dataID)) {
    return;
  }
  if (row.usedAt === null) {
    passwordTableIndex_UsedAt.push({
      xoredDataID: row.dataID,
      usedAt: row.usedAt,
      xoredUserID: row.userID,
    });
    return;
  }
  let usedAt = parseInt(xor_str(secretHex, row.usedAt));
  for (let i = 0; i < passwordTableIndex_UsedAt.length; i++) {
    if (
      passwordTableIndex_UsedAt[i].usedAt === null ||
      passwordTableIndex_UsedAt[i].usedAt! < usedAt
    ) {
      let newIndex: passwordUsedAtIndex[] = [];
      if (i === 0) {
        newIndex.push({
          xoredDataID: row.dataID,
          usedAt: usedAt,
          xoredUserID: row.userID,
        });
      } else {
        newIndex = [...passwordTableIndex_UsedAt.slice(0, i)];
        newIndex.push({
          xoredDataID: row.dataID,
          usedAt: usedAt,
          xoredUserID: row.userID,
        });
      }
      newIndex.push(...passwordTableIndex_UsedAt.slice(i));
      passwordTableIndex_UsedAt = newIndex;
      return;
    }
  }
  passwordTableIndex_UsedAt.push({
    xoredDataID: row.dataID,
    usedAt: usedAt,
    xoredUserID: row.userID,
  });
};
const splicePasswordTableIndex_UsedAt = (xoredDataID: string) => {
  for (let i = 0; i < passwordTableIndex_UsedAt.length; i++) {
    if (passwordTableIndex_UsedAt[i].xoredDataID == xoredDataID) {
      let newIndex: passwordUsedAtIndex[] = [];
      if (i === 0) {
        newIndex.push(...passwordTableIndex_UsedAt.slice(1));
      } else {
        newIndex.push(
          ...passwordTableIndex_UsedAt.slice(0, i),
          ...passwordTableIndex_UsedAt.slice(i + 1),
        );
      }
      passwordTableIndex_UsedAt = newIndex;
      return;
    }
  }
};

// index updatedAt：从大到小，支持 title文本搜索
interface passwordUpdatedAtIndex {
  updatedAt: number;
  title: string;
  xoredDataID: string;
  xoredUserID: string;
}
let passwordTableIndex_UpdatedAt: passwordUpdatedAtIndex[] = [];
const insertPasswordTableIndex_UpdatedAt = (
  secretHex: string,
  row: XoredPassword,
) => {
  if (indexExists(passwordTableIndex_UpdatedAt, row.dataID)) {
    return;
  }
  if (row.updatedAt === null) {
    passwordTableIndex_UpdatedAt.push({
      xoredDataID: row.dataID,
      updatedAt: row.updatedAt,
      title: row.title,
      xoredUserID: row.userID,
    });
    return;
  }
  let updatedAt = parseInt(xor_str(secretHex, row.updatedAt));
  for (let i = 0; i < passwordTableIndex_UpdatedAt.length; i++) {
    if (passwordTableIndex_UpdatedAt[i].updatedAt < updatedAt) {
      let newIndex: passwordUpdatedAtIndex[] = [];
      if (i === 0) {
        newIndex.push({
          xoredDataID: row.dataID,
          updatedAt: updatedAt,
          title: row.title,
          xoredUserID: row.userID,
        });
      } else {
        newIndex = [...passwordTableIndex_UpdatedAt.slice(0, i)];
        newIndex.push({
          xoredDataID: row.dataID,
          updatedAt: updatedAt,
          title: row.title,
          xoredUserID: row.userID,
        });
      }
      newIndex.push(...passwordTableIndex_UpdatedAt.slice(i));
      passwordTableIndex_UpdatedAt = newIndex;
      return;
    }
  }
  passwordTableIndex_UpdatedAt.push({
    xoredDataID: row.dataID,
    updatedAt: updatedAt,
    title: row.title,
    xoredUserID: row.userID,
  });
};
const splicePasswordTableIndex_UpdatedAt = (xoredDataID: string) => {
  for (let i = 0; i < passwordTableIndex_UpdatedAt.length; i++) {
    if (passwordTableIndex_UpdatedAt[i].xoredDataID == xoredDataID) {
      let newIndex: passwordUpdatedAtIndex[] = [];
      if (i === 0) {
        newIndex.push(...passwordTableIndex_UpdatedAt.slice(1));
      } else {
        newIndex.push(
          ...passwordTableIndex_UpdatedAt.slice(0, i),
          ...passwordTableIndex_UpdatedAt.slice(i + 1),
        );
      }
      passwordTableIndex_UpdatedAt = newIndex;
      return;
    }
  }
};
