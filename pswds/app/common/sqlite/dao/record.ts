import { SQLiteRunResult } from 'expo-sqlite';
import { db } from '..';
import { Record } from '../schema';
import moment from 'moment';
import { insertErrorLog } from '../../log';
import { xor_hex, xor_str } from './utils';

export interface XoredRecord {
  dataID: string;
  createdAt: string;
  updatedAt: string;
  userID: string;
  recordType: string;
  title: string;
  iconBgColor: string | null;
  usedAt: string | null;
  usedCount: string | null;
  // mixed fields
  phone: string | null;
  type: string | null;
  number: string | null;
  address: string | null;
  fullName: string | null;
  birthDate: string | null;
  gender: string | null;
  pin: string | null;
  expiryDate: string | null;
  others: string | null;
  sharedAt: string | null;
  sharedToAll: string | null;
  sharingMembers: string | null;
  // identity fields
  firstName: string | null;
  lastName: string | null;
  job: string | null;
  socialSecurityNumber: string | null;
  idNumber: string | null;
  // credit card fields
  cardholderName: string | null;
  verificationNumber: string | null;
  validFrom: string | null;
  issuingBank: string | null;
  // bank account fields
  bankName: string | null;
  nameOnAccount: string | null;
  routingNumber: string | null;
  branch: string | null;
  accountNumber: string | null;
  swift: string | null;
  // driver license fields
  height: string | null;
  licenseClass: string | null;
  state: string | null;
  country: string | null;
  // passport fields
  issuingCountry: string | null;
  nationality: string | null;
  issuingAuthority: string | null;
  birthPlace: string | null;
  issuedOn: string | null;
}
export const xorRecord = (secretHex: string, entity: Record): XoredRecord => {
  return {
    dataID: xor_str(secretHex, entity.dataID),
    createdAt: xor_str(secretHex, entity.createdAt.toString()),
    updatedAt: xor_str(secretHex, entity.updatedAt.toString()),
    userID: xor_str(secretHex, entity.userID.toString()),
    recordType: xor_str(secretHex, entity.recordType),
    title: xor_str(secretHex, entity.title),
    iconBgColor: entity.iconBgColor
      ? xor_str(secretHex, entity.iconBgColor.toString())
      : null,
    usedAt: entity.usedAt ? xor_str(secretHex, entity.usedAt.toString()) : null,
    usedCount: entity.usedCount
      ? xor_str(secretHex, entity.usedCount.toString())
      : null,
    // mixed fields
    phone: entity.phone ? xor_str(secretHex, entity.phone) : null,
    type: entity.type ? xor_str(secretHex, entity.type) : null,
    number: entity.number ? xor_str(secretHex, entity.number) : null,
    address: entity.address ? xor_str(secretHex, entity.address) : null,
    fullName: entity.fullName ? xor_str(secretHex, entity.fullName) : null,
    birthDate: entity.birthDate ? xor_str(secretHex, entity.birthDate) : null,
    gender: entity.gender ? xor_str(secretHex, entity.gender) : null,
    pin: entity.pin ? xor_str(secretHex, entity.pin) : null,
    expiryDate: entity.expiryDate
      ? xor_str(secretHex, entity.expiryDate)
      : null,
    others: entity.others ? xor_str(secretHex, entity.others) : null,
    sharingMembers: entity.sharingMembers
      ? xor_str(secretHex, entity.sharingMembers)
      : null,
    sharedAt: entity.sharedAt
      ? xor_str(secretHex, entity.sharedAt.toString())
      : null,
    sharedToAll: entity.sharedToAll
      ? xor_str(secretHex, entity.sharedToAll.toString())
      : null,
    // identity fields
    firstName: entity.firstName ? xor_str(secretHex, entity.firstName) : null,
    lastName: entity.lastName ? xor_str(secretHex, entity.lastName) : null,
    job: entity.job ? xor_str(secretHex, entity.job) : null,
    socialSecurityNumber: entity.socialSecurityNumber
      ? xor_str(secretHex, entity.socialSecurityNumber)
      : null,
    idNumber: entity.idNumber ? xor_str(secretHex, entity.idNumber) : null,
    // credit card fields
    cardholderName: entity.cardholderName
      ? xor_str(secretHex, entity.cardholderName)
      : null,
    verificationNumber: entity.verificationNumber
      ? xor_str(secretHex, entity.verificationNumber)
      : null,
    validFrom: entity.validFrom ? xor_str(secretHex, entity.validFrom) : null,
    issuingBank: entity.issuingBank
      ? xor_str(secretHex, entity.issuingBank)
      : null,
    // bank account fields
    bankName: entity.bankName ? xor_str(secretHex, entity.bankName) : null,
    nameOnAccount: entity.nameOnAccount
      ? xor_str(secretHex, entity.nameOnAccount)
      : null,
    routingNumber: entity.routingNumber
      ? xor_str(secretHex, entity.routingNumber)
      : null,
    branch: entity.branch ? xor_str(secretHex, entity.branch) : null,
    accountNumber: entity.accountNumber
      ? xor_str(secretHex, entity.accountNumber)
      : null,
    swift: entity.swift ? xor_str(secretHex, entity.swift) : null,
    // driver license fields
    height: entity.height ? xor_str(secretHex, entity.height) : null,
    licenseClass: entity.licenseClass
      ? xor_str(secretHex, entity.licenseClass)
      : null,
    state: entity.state ? xor_str(secretHex, entity.state) : null,
    country: entity.country ? xor_str(secretHex, entity.country) : null,
    // passport fields
    issuingCountry: entity.issuingCountry
      ? xor_str(secretHex, entity.issuingCountry)
      : null,
    nationality: entity.nationality
      ? xor_str(secretHex, entity.nationality)
      : null,
    issuingAuthority: entity.issuingAuthority
      ? xor_str(secretHex, entity.issuingAuthority)
      : null,
    birthPlace: entity.birthPlace
      ? xor_str(secretHex, entity.birthPlace)
      : null,
    issuedOn: entity.issuedOn ? xor_str(secretHex, entity.issuedOn) : null,
  };
};
export const xorRecords = (
  secretHex: string,
  entities: Record[],
): XoredRecord[] => {
  let result: XoredRecord[] = [];
  entities.map(item => {
    result.push(xorRecord(secretHex, item));
  });
  return result;
};
export const xorXoredRecord = (
  secretHex: string,
  entity: XoredRecord,
): Record => {
  return {
    dataID: xor_hex(secretHex, entity.dataID),
    createdAt: parseInt(xor_hex(secretHex, entity.createdAt)),
    updatedAt: parseInt(xor_hex(secretHex, entity.updatedAt)),
    userID: parseInt(xor_hex(secretHex, entity.userID)),
    recordType: xor_hex(secretHex, entity.recordType),
    title: xor_hex(secretHex, entity.title),
    iconBgColor: entity.iconBgColor
      ? parseInt(xor_hex(secretHex, entity.iconBgColor))
      : null,
    usedAt: entity.usedAt ? parseInt(xor_hex(secretHex, entity.usedAt)) : null,
    usedCount: entity.usedCount
      ? parseInt(xor_hex(secretHex, entity.usedCount))
      : null,
    // mixed fields
    phone: entity.phone ? xor_hex(secretHex, entity.phone) : null,
    type: entity.type ? xor_hex(secretHex, entity.type) : null,
    number: entity.number ? xor_hex(secretHex, entity.number) : null,
    address: entity.address ? xor_hex(secretHex, entity.address) : null,
    fullName: entity.fullName ? xor_hex(secretHex, entity.fullName) : null,
    birthDate: entity.birthDate ? xor_hex(secretHex, entity.birthDate) : null,
    gender: entity.gender ? xor_hex(secretHex, entity.gender) : null,
    pin: entity.pin ? xor_hex(secretHex, entity.pin) : null,
    expiryDate: entity.expiryDate
      ? xor_hex(secretHex, entity.expiryDate)
      : null,
    others: entity.others ? xor_hex(secretHex, entity.others) : null,
    sharingMembers: entity.sharingMembers
      ? xor_hex(secretHex, entity.sharingMembers)
      : null,
    sharedAt: entity.sharedAt
      ? parseInt(xor_hex(secretHex, entity.sharedAt))
      : null,
    sharedToAll: entity.sharedToAll
      ? parseInt(xor_hex(secretHex, entity.sharedToAll))
      : null,
    // identity fields
    firstName: entity.firstName ? xor_hex(secretHex, entity.firstName) : null,
    lastName: entity.lastName ? xor_hex(secretHex, entity.lastName) : null,
    job: entity.job ? xor_hex(secretHex, entity.job) : null,
    socialSecurityNumber: entity.socialSecurityNumber
      ? xor_hex(secretHex, entity.socialSecurityNumber)
      : null,
    idNumber: entity.idNumber ? xor_hex(secretHex, entity.idNumber) : null,
    // credit card fields
    cardholderName: entity.cardholderName
      ? xor_hex(secretHex, entity.cardholderName)
      : null,
    verificationNumber: entity.verificationNumber
      ? xor_hex(secretHex, entity.verificationNumber)
      : null,
    validFrom: entity.validFrom ? xor_hex(secretHex, entity.validFrom) : null,
    issuingBank: entity.issuingBank
      ? xor_hex(secretHex, entity.issuingBank)
      : null,
    // bank account fields
    bankName: entity.bankName ? xor_hex(secretHex, entity.bankName) : null,
    nameOnAccount: entity.nameOnAccount
      ? xor_hex(secretHex, entity.nameOnAccount)
      : null,
    routingNumber: entity.routingNumber
      ? xor_hex(secretHex, entity.routingNumber)
      : null,
    branch: entity.branch ? xor_hex(secretHex, entity.branch) : null,
    accountNumber: entity.accountNumber
      ? xor_hex(secretHex, entity.accountNumber)
      : null,
    swift: entity.swift ? xor_hex(secretHex, entity.swift) : null,
    // driver license fields
    height: entity.height ? xor_hex(secretHex, entity.height) : null,
    licenseClass: entity.licenseClass
      ? xor_hex(secretHex, entity.licenseClass)
      : null,
    state: entity.state ? xor_hex(secretHex, entity.state) : null,
    country: entity.country ? xor_hex(secretHex, entity.country) : null,
    // passport fields
    issuingCountry: entity.issuingCountry
      ? xor_hex(secretHex, entity.issuingCountry)
      : null,
    nationality: entity.nationality
      ? xor_hex(secretHex, entity.nationality)
      : null,
    issuingAuthority: entity.issuingAuthority
      ? xor_hex(secretHex, entity.issuingAuthority)
      : null,
    birthPlace: entity.birthPlace
      ? xor_hex(secretHex, entity.birthPlace)
      : null,
    issuedOn: entity.issuedOn ? xor_hex(secretHex, entity.issuedOn) : null,
  };
};
export const xorXoredRecords = (
  secretHex: string,
  entities: XoredRecord[],
): Record[] => {
  let result: Record[] = [];
  entities.map(item => {
    result.push(xorXoredRecord(secretHex, item));
  });
  return result;
};

// inserts

export const insertRecordAsyncStatement = async () => {
  return await db.prepareAsync(
    `INSERT INTO record (
          dataID, 
          createdAt, 
          updatedAt, 
          userID, 
          recordType, 
          title, 
          iconBgColor, 
          usedAt, 
          usedCount, 
          -- mixed fields
          phone, 
          type, 
          number, 
          address, 
          fullName, 
          birthDate, 
          gender, 
          pin, 
          expiryDate, 
          others, 
          sharingMembers, 
          sharedAt, 
          sharedToAll, 
          -- identity fields
          firstName, 
          lastName, 
          job, 
          socialSecurityNumber, 
          idNumber, 
          -- credit card fields
          cardholderName, 
          verificationNumber, 
          validFrom, 
          issuingBank, 
          -- bank account fields
          bankName, 
          nameOnAccount, 
          routingNumber, 
          branch, 
          accountNumber, 
          swift, 
          -- driver license fields
          height,
          licenseClass,
          state,
          country,
          -- passport fields
          issuingCountry,
          nationality,
          issuingAuthority,
          birthPlace,
          issuedOn) VALUES (
          $dataID, 
          $createdAt, 
          $updatedAt, 
          $userID, 
          $recordType, 
          $title, 
          $iconBgColor, 
          $usedAt, 
          $usedCount, 
          -- mixed fields
          $phone, 
          $type, 
          $number, 
          $address, 
          $fullName, 
          $birthDate, 
          $gender, 
          $pin, 
          $expiryDate, 
          $others, 
          $sharingMembers, 
          $sharedAt, 
          $sharedToAll, 
          -- identity fields
          $firstName, 
          $lastName, 
          $job, 
          $socialSecurityNumber, 
          $idNumber, 
          -- credit card fields
          $cardholderName, 
          $verificationNumber, 
          $validFrom, 
          $issuingBank, 
          -- bank account fields
          $bankName, 
          $nameOnAccount, 
          $routingNumber, 
          $branch, 
          $accountNumber, 
          $swift, 
          -- driver license fields
          $height,
          $licenseClass,
          $state,
          $country,
          -- passport fields
          $issuingCountry,
          $nationality,
          $issuingAuthority,
          $birthPlace,
          $issuedOn);`,
  );
};

export const recordToInsertable = (entity: XoredRecord) => {
  return {
    $dataID: entity.dataID,
    $createdAt: entity.createdAt,
    $updatedAt: entity.updatedAt,
    $userID: entity.userID,
    $recordType: entity.recordType,
    $title: entity.title,
    $iconBgColor: entity.iconBgColor,
    $usedAt: entity.usedAt,
    $usedCount: entity.usedCount,
    // mixed fields
    $phone: entity.phone,
    $type: entity.type,
    $number: entity.number,
    $address: entity.address,
    $fullName: entity.fullName,
    $birthDate: entity.birthDate,
    $gender: entity.gender,
    $pin: entity.pin,
    $expiryDate: entity.expiryDate,
    $others: entity.others,
    $sharingMembers: entity.sharingMembers,
    $sharedAt: entity.sharedAt,
    $sharedToAll: entity.sharedToAll,
    // identity fields
    $firstName: entity.firstName,
    $lastName: entity.lastName,
    $job: entity.job,
    $socialSecurityNumber: entity.socialSecurityNumber,
    $idNumber: entity.idNumber,
    // credit card fields
    $cardholderName: entity.cardholderName,
    $verificationNumber: entity.verificationNumber,
    $validFrom: entity.validFrom,
    $issuingBank: entity.issuingBank,
    // bank account fields
    $bankName: entity.bankName,
    $nameOnAccount: entity.nameOnAccount,
    $routingNumber: entity.routingNumber,
    $branch: entity.branch,
    $accountNumber: entity.accountNumber,
    $swift: entity.swift,
    // driver license fields
    $height: entity.height,
    $licenseClass: entity.licenseClass,
    $state: entity.state,
    $country: entity.country,
    // passport fields
    $issuingCountry: entity.issuingCountry,
    $nationality: entity.nationality,
    $issuingAuthority: entity.issuingAuthority,
    $birthPlace: entity.birthPlace,
    $issuedOn: entity.issuedOn,
  };
};

export const insertRecordAsync = async (entity: XoredRecord) => {
  const statement = await insertRecordAsyncStatement();
  const nowTS = moment();
  try {
    const item = recordToInsertable(entity);
    await statement.executeAsync<XoredRecord>(item);
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

export const insertRecordsAsync = async (entities: XoredRecord[]) => {
  const statement = await insertRecordAsyncStatement();
  const nowTS = moment();
  try {
    entities.map(async item => {
      await statement.executeAsync<XoredRecord>(recordToInsertable(item));
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

export const upsertRecordsAsync = async (entities: XoredRecord[]) => {
  try {
    entities.map(async item => {
      let record: XoredRecord | null = await getRecordByDataIDAsync(
        item.dataID,
      );
      if (record) {
        await updateRecordAsync(item);
      } else {
        await insertRecordAsync(item);
      }
    });
  } catch (error) {
    throw error;
  } finally {
  }
};

// updates

export const updateRecordAsyncStatement = async () => {
  return await db.prepareAsync(
    `UPDATE record 
      SET updatedAt=$updatedAt, 
          title=$title, 
          iconBgColor=$iconBgColor, 
          usedAt=$usedAt, 
          usedCount=$usedCount, 
          -- mixed fields
          phone=$phone, 
          type=$type, 
          number=$number, 
          address=$address, 
          fullName=$fullName, 
          birthDate=$birthDate, 
          gender=$gender, 
          pin=$pin, 
          expiryDate=$expiryDate, 
          others=$others, 
          sharingMembers=$sharingMembers, 
          sharedAt=$sharedAt, 
          sharedToAll=$sharedToAll, 
          -- identity fields
          firstName=$firstName, 
          lastName=$lastName, 
          job=$job, 
          socialSecurityNumber=$socialSecurityNumber, 
          idNumber=$idNumber,
          -- credit card fields
          cardholderName=$cardholderName, 
          verificationNumber=$verificationNumber, 
          validFrom=$validFrom, 
          issuingBank=$issuingBank, 
          -- bank account fields
          bankName=$bankName, 
          nameOnAccount=$nameOnAccount, 
          routingNumber=$routingNumber, 
          branch=$branch, 
          accountNumber=$accountNumber, 
          swift=$swift, 
          -- driver license fields
          height=$height,
          licenseClass=$licenseClass,
          state=$state,
          country=$country,
          -- passport fields
          issuingCountry=$issuingCountry,
          nationality=$nationality,
          issuingAuthority=$issuingAuthority,
          birthPlace=$birthPlace,
          issuedOn=$issuedOn
        WHERE dataID=$dataID;`,
  );
};

export const recordToUpdatable = (entity: XoredRecord) => {
  return {
    $dataID: entity.dataID,
    $updatedAt: entity.updatedAt,
    $title: entity.title,
    $iconBgColor: entity.iconBgColor,
    $usedAt: entity.usedAt,
    $usedCount: entity.usedCount,
    // mixed fields
    $phone: entity.phone,
    $type: entity.type,
    $number: entity.number,
    $address: entity.address,
    $fullName: entity.fullName,
    $birthDate: entity.birthDate,
    $gender: entity.gender,
    $pin: entity.pin,
    $expiryDate: entity.expiryDate,
    $others: entity.others,
    $sharingMembers: entity.sharingMembers,
    $sharedAt: entity.sharedAt,
    $sharedToAll: entity.sharedToAll,
    // identity fields
    $firstName: entity.firstName,
    $lastName: entity.lastName,
    $job: entity.job,
    $socialSecurityNumber: entity.socialSecurityNumber,
    $idNumber: entity.idNumber,
    // credit card fields
    $cardholderName: entity.cardholderName,
    $verificationNumber: entity.verificationNumber,
    $validFrom: entity.validFrom,
    $issuingBank: entity.issuingBank,
    // bank account fields
    $bankName: entity.bankName,
    $nameOnAccount: entity.nameOnAccount,
    $routingNumber: entity.routingNumber,
    $branch: entity.branch,
    $accountNumber: entity.accountNumber,
    $swift: entity.swift,
    // driver license fields
    $height: entity.height,
    $licenseClass: entity.licenseClass,
    $state: entity.state,
    $country: entity.country,
    // passport fields
    $issuingCountry: entity.issuingCountry,
    $nationality: entity.nationality,
    $issuingAuthority: entity.issuingAuthority,
    $birthPlace: entity.birthPlace,
    $issuedOn: entity.issuedOn,
  };
};

export const updateRecordAsync = async (entity: XoredRecord) => {
  const statement = await updateRecordAsyncStatement();
  const nowTS = moment();
  try {
    await statement.executeAsync<XoredRecord>(recordToUpdatable(entity));
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

export const clearUnsignedInRecordsAsync = async (
  xoredUnsignedUserID: string,
  secretHex: string,
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE record 
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

export const cancelSharingRecordsByUserIDAsync = async (
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE record 
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

export const sharingRecordByDataIDAsync = async (
  xoredDataID: string,
  xoredSharedAt: string | null,
  xoredSharedToAll: string | null,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE record 
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

export const addRecordUseByDataIDAsync = async (
  secretHex: string,
  xoredDataID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    const entity: XoredRecord | null = await db.getFirstAsync(
      'SELECT * FROM record WHERE dataID = ?;',
      xoredDataID,
    );
    if (!entity) {
      throw 'invalid record data id ' + xoredDataID;
    }
    return db.runAsync(
      `UPDATE record 
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

export const updateRecordSharingMembersByDataIDAsync = async (
  xoredDataID: string,
  xoredSharingMembers: string | null,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `UPDATE record 
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

export const changeUnlockPasswordAsync_Record = async (
  oldPasswordHash: string,
  newPasswordHash: string,
) => {
  const nowTS = moment();
  try {
    // 1. remove old indexes
    clearRecordTableIndexes();
    // 2. iterate
    for await (const row of db.getEachAsync<XoredRecord>(
      'SELECT * FROM record;',
    )) {
      // 2-1. delete
      await deleteRecordByDataIDAsync(row.dataID);
      // 2-2. insert new entity
      const plainEntity = xorXoredRecord(oldPasswordHash, row);
      const xoredEntity = xorRecord(newPasswordHash, plainEntity);
      await insertRecordAsync(xoredEntity);
      // 2-3. insert indexes
      insertRecordTableIndex_UpdatedAt(newPasswordHash, xoredEntity);
      insertRecordTableIndex_UsedAt(newPasswordHash, xoredEntity);
      insertRecordTableIndex_UsedCount(newPasswordHash, xoredEntity);
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

export const deleteRecordsFromFamilyAsync = async (
  secretHex: string,
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync(
      `DELETE FROM record WHERE userID != $userID AND userID != $unsigned;`,
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

export const deleteRecordsByUserIDAsync = async (
  xoredUserID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync('DELETE FROM record WHERE userID = $userID;', {
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

export const deleteRecordByDataIDAsync = async (
  xoredDataID: string,
): Promise<SQLiteRunResult> => {
  const nowTS = moment();
  try {
    return db.runAsync('DELETE FROM record WHERE dataID = $dataID;', {
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

export const getRecordByDataIDAsync = async (
  xoredDataID: string,
): Promise<XoredRecord | null> => {
  const nowTS = moment();
  try {
    return db.getFirstAsync(
      'SELECT * FROM record WHERE dataID = ?;',
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

export const getRecordsAsync = async (): Promise<XoredRecord[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredRecord[] = [];
    for (let i = 0; i < recordTableIndex_UpdatedAt.length; i++) {
      const result = await getRecordByDataIDAsync(
        recordTableIndex_UpdatedAt[i].xoredDataID,
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

export const getMostUsedRecordsAsync = async (
  limit: number,
): Promise<XoredRecord[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredRecord[] = [];
    for (let i = 0; i < recordTableIndex_UsedCount.length; i++) {
      if (i === limit) {
        break;
      }
      const result = await getRecordByDataIDAsync(
        recordTableIndex_UsedCount[i].xoredDataID,
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

export const getLatestUsedRecordsAsync = async (
  limit: number,
): Promise<XoredRecord[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredRecord[] = [];
    for (let i = 0; i < recordTableIndex_UsedAt.length; i++) {
      if (i === limit) {
        break;
      }
      const result = await getRecordByDataIDAsync(
        recordTableIndex_UsedAt[i].xoredDataID,
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

export const getLatestUpdatedRecordsAsync = async (
  limit: number,
): Promise<XoredRecord[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredRecord[] = [];
    for (let i = 0; i < recordTableIndex_UpdatedAt.length; i++) {
      if (i === limit) {
        break;
      }
      const result = await getRecordByDataIDAsync(
        recordTableIndex_UpdatedAt[i].xoredDataID,
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

export const getRecordsByUserIDAsync = async (
  xoredUserID: string,
): Promise<XoredRecord[]> => {
  const nowTS = moment();
  try {
    return db.getAllAsync(
      'SELECT * FROM record WHERE userID = ?;',
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

export const getSharedRecordsAsync = async (): Promise<XoredRecord[]> => {
  const nowTS = moment();
  try {
    return db.getAllAsync(`SELECT * FROM record WHERE sharedAt IS NOT NULL;`);
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

export const getUploadableRecordsByUserIDAsync = async (
  xoredUserID: string,
  updatedAt: number,
): Promise<XoredRecord[]> => {
  const nowTS = moment();
  try {
    // search from index
    const list: XoredRecord[] = [];
    for (let i = 0; i < recordTableIndex_UpdatedAt.length; i++) {
      if (recordTableIndex_UpdatedAt[i].xoredUserID !== xoredUserID) {
        continue;
      }
      if (recordTableIndex_UpdatedAt[i].updatedAt > updatedAt) {
        const result = await getRecordByDataIDAsync(
          recordTableIndex_UpdatedAt[i].xoredDataID,
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

export const getRecordSharedDataMembersSync = (
  secretHex: string,
  xoredDataID: string,
): null | number[] => {
  const nowTS = moment();
  try {
    const entity: XoredRecord | null = db.getFirstSync(
      'SELECT * FROM record WHERE dataID = ?;',
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

export const buildRecordTableIndexesByUserID = (secretHex: string) => {
  const nowTS = moment();
  try {
    console.log(
      '>>>>>>>>>>> 开始构建Record的索引：',
      nowTS.local().format('YYYY-MM-DD HH:mm:ss.SSS'),
    );
    for (const row of db.getEachSync<XoredRecord>(
      'SELECT dataID, title, updatedAt, usedCount, usedAt FROM record;',
    )) {
      insertRecordTableIndexes(secretHex, row);
    }
    const overTS = moment();
    console.log(
      '>>>>>>>>>>> 结束构建Record的索引：',
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
export const clearRecordTableIndexes = () => {
  recordTableIndex_UsedCount = [];
  recordTableIndex_UpdatedAt = [];
  recordTableIndex_UsedAt = [];
};

export const insertRecordTableIndexes = (
  secretHex: string,
  row: XoredRecord,
) => {
  insertRecordTableIndex_UpdatedAt(secretHex, row);
  insertRecordTableIndex_UsedAt(secretHex, row);
  insertRecordTableIndex_UsedCount(secretHex, row);
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
interface recordUsedCountIndex {
  usedCount: number | null;
  xoredDataID: string;
  xoredUserID: string;
}
let recordTableIndex_UsedCount: recordUsedCountIndex[] = [];
const insertRecordTableIndex_UsedCount = (
  secretHex: string,
  row: XoredRecord,
) => {
  if (indexExists(recordTableIndex_UsedCount, row.dataID)) {
    return;
  }
  if (row.usedCount === null) {
    recordTableIndex_UsedCount.push({
      xoredDataID: row.dataID,
      usedCount: row.usedCount,
      xoredUserID: row.userID,
    });
    return;
  }
  let usedCount =
    typeof row.usedCount === 'string'
      ? parseInt(xor_str(secretHex, row.usedCount))
      : row.usedCount;
  for (let i = 0; i < recordTableIndex_UsedCount.length; i++) {
    if (
      recordTableIndex_UsedCount[i].usedCount === null ||
      recordTableIndex_UsedCount[i].usedCount! < usedCount
    ) {
      let newIndex: recordUsedCountIndex[] = [];
      if (i === 0) {
        newIndex.push({
          xoredDataID: row.dataID,
          usedCount: usedCount,
          xoredUserID: row.userID,
        });
      } else {
        newIndex = [...recordTableIndex_UsedCount.slice(0, i)];
        newIndex.push({
          xoredDataID: row.dataID,
          usedCount: usedCount,
          xoredUserID: row.userID,
        });
      }
      newIndex.push(...recordTableIndex_UsedCount.slice(i));
      recordTableIndex_UsedCount = newIndex;
      return;
    }
  }
  recordTableIndex_UsedCount.push({
    xoredDataID: row.dataID,
    usedCount: usedCount,
    xoredUserID: row.userID,
  });
};
const spliceRecordTableIndex_UsedCount = (xoredDataID: string) => {
  for (let i = 0; i < recordTableIndex_UsedCount.length; i++) {
    if (recordTableIndex_UsedCount[i].xoredDataID == xoredDataID) {
      let newIndex: recordUsedCountIndex[] = [];
      if (i === 0) {
        newIndex.push(...recordTableIndex_UsedCount.slice(1));
      } else {
        newIndex.push(
          ...recordTableIndex_UsedCount.slice(0, i),
          ...recordTableIndex_UsedCount.slice(i + 1),
        );
      }
      recordTableIndex_UsedCount = newIndex;
      return;
    }
  }
};

// index usedAt：从大到小
interface recordUsedAtIndex {
  usedAt: number | null;
  xoredDataID: string;
  xoredUserID: string;
}
let recordTableIndex_UsedAt: recordUsedAtIndex[] = [];
const insertRecordTableIndex_UsedAt = (secretHex: string, row: XoredRecord) => {
  if (indexExists(recordTableIndex_UsedAt, row.dataID)) {
    return;
  }
  if (row.usedAt === null) {
    recordTableIndex_UsedAt.push({
      xoredDataID: row.dataID,
      usedAt: row.usedAt,
      xoredUserID: row.userID,
    });
    return;
  }
  let usedAt =
    typeof row.usedAt === 'string'
      ? parseInt(xor_str(secretHex, row.usedAt))
      : row.usedAt;
  for (let i = 0; i < recordTableIndex_UsedAt.length; i++) {
    if (
      recordTableIndex_UsedAt[i].usedAt === null ||
      recordTableIndex_UsedAt[i].usedAt! < usedAt
    ) {
      let newIndex: recordUsedAtIndex[] = [];
      if (i === 0) {
        newIndex.push({
          xoredDataID: row.dataID,
          usedAt: usedAt,
          xoredUserID: row.userID,
        });
      } else {
        newIndex = [...recordTableIndex_UsedAt.slice(0, i)];
        newIndex.push({
          xoredDataID: row.dataID,
          usedAt: usedAt,
          xoredUserID: row.userID,
        });
      }
      newIndex.push(...recordTableIndex_UsedAt.slice(i));
      recordTableIndex_UsedAt = newIndex;
      return;
    }
  }
  recordTableIndex_UsedAt.push({
    xoredDataID: row.dataID,
    usedAt: usedAt,
    xoredUserID: row.userID,
  });
};
const spliceRecordTableIndex_UsedAt = (xoredDataID: string) => {
  for (let i = 0; i < recordTableIndex_UsedAt.length; i++) {
    if (recordTableIndex_UsedAt[i].xoredDataID == xoredDataID) {
      let newIndex: recordUsedAtIndex[] = [];
      if (i === 0) {
        newIndex.push(...recordTableIndex_UsedAt.slice(1));
      } else {
        newIndex.push(
          ...recordTableIndex_UsedAt.slice(0, i),
          ...recordTableIndex_UsedAt.slice(i + 1),
        );
      }
      recordTableIndex_UsedAt = newIndex;
      return;
    }
  }
};

// index updatedAt：从大到小，支持 title文本搜索
interface recordUpdatedAtIndex {
  updatedAt: number;
  title: string;
  xoredDataID: string;
  xoredUserID: string;
}
let recordTableIndex_UpdatedAt: recordUpdatedAtIndex[] = [];
const insertRecordTableIndex_UpdatedAt = (
  secretHex: string,
  row: XoredRecord,
) => {
  if (indexExists(recordTableIndex_UpdatedAt, row.dataID)) {
    return;
  }
  const updatedAt = parseInt(xor_str(secretHex, row.updatedAt));
  if (row.usedAt === null) {
    recordTableIndex_UpdatedAt.push({
      xoredDataID: row.dataID,
      updatedAt: updatedAt,
      title: row.title,
      xoredUserID: row.userID,
    });
    return;
  }
  for (let i = 0; i < recordTableIndex_UpdatedAt.length; i++) {
    if (recordTableIndex_UpdatedAt[i].updatedAt < updatedAt) {
      let newIndex: recordUpdatedAtIndex[] = [];
      if (i === 0) {
        newIndex.push({
          xoredDataID: row.dataID,
          updatedAt: updatedAt,
          title: row.title,
          xoredUserID: row.userID,
        });
      } else {
        newIndex = [...recordTableIndex_UpdatedAt.slice(0, i)];
        newIndex.push({
          xoredDataID: row.dataID,
          updatedAt: updatedAt,
          title: row.title,
          xoredUserID: row.userID,
        });
      }
      newIndex.push(...recordTableIndex_UpdatedAt.slice(i));
      recordTableIndex_UpdatedAt = newIndex;
      return;
    }
  }
  recordTableIndex_UpdatedAt.push({
    xoredDataID: row.dataID,
    updatedAt: updatedAt,
    title: row.title,
    xoredUserID: row.userID,
  });
};
const spliceRecordTableIndex_UpdatedAt = (xoredDataID: string) => {
  for (let i = 0; i < recordTableIndex_UpdatedAt.length; i++) {
    if (recordTableIndex_UpdatedAt[i].xoredDataID == xoredDataID) {
      let newIndex: recordUpdatedAtIndex[] = [];
      if (i === 0) {
        newIndex.push(...recordTableIndex_UpdatedAt.slice(1));
      } else {
        newIndex.push(
          ...recordTableIndex_UpdatedAt.slice(0, i),
          ...recordTableIndex_UpdatedAt.slice(i + 1),
        );
      }
      recordTableIndex_UpdatedAt = newIndex;
      return;
    }
  }
};
