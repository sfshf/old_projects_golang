import * as SQLite from 'expo-sqlite';
import { passwordSchema, recordSchema } from './schema';
import { insertErrorLog } from '../log';
import moment from 'moment';
import { databaseVersion, migrationSQLs } from './migration';
import { currentDBVersion, upsertDBVersion } from '../mmkv/dbVersion';

export const DatabaseName_PSWDS = 'db_pswds';
export const db = SQLite.openDatabaseSync(DatabaseName_PSWDS);
export const migrateTables = () => {
  const nowTS = moment();
  try {
    // 1. 检查 db version 记录；
    const curDBVersion = currentDBVersion();
    if (!curDBVersion) {
      // （1）没有，则插入一条版本记录
      upsertDBVersion({
        createdAt: nowTS.unix(),
        updatedAt: nowTS.unix(),
        version: databaseVersion,
      });
      // （2）创建sqlite表；
      createTables();
      // （3）执行migration语句；
      for (let i = 0; i < databaseVersion; i++) {
        db.execSync(migrationSQLs[i]);
      }
    } else {
      // 检查数据版本，如果数据表有改动，则执行相应改动语句；
      if (curDBVersion.version < databaseVersion) {
        // （1）更新db_version记录；
        upsertDBVersion({
          ...curDBVersion,
          updatedAt: nowTS.unix(),
          version: databaseVersion,
        });
        // （2）执行migration语句；
        for (let i = curDBVersion.version; i < databaseVersion; i++) {
          db.execSync(migrationSQLs[i]);
        }
      }
    }
    // // test migrations
    // const testResult = db.runSync(
    //   `INSERT INTO test_migration (
    // message) VALUES (
    // $message);`,
    //   { $message: 'test migration' },
    // );
    // if (testResult.changes === 0) {
    //   console.log('=============>test migration fail');
    // } else if (testResult.changes === 1) {
    //   console.log('=============>test migration success');
    // }
  } catch (error) {
    console.log('[sqlite] migrate tables error:', error as string);
    insertErrorLog({
      level: 'error',
      timestamp: nowTS.valueOf(),
      message: error as string,
    });
  }
};
export const createTables = () => {
  try {
    db.execSync(`PRAGMA journal_mode = WAL;` + passwordSchema + recordSchema);
  } catch (error) {
    console.log('[sqlite] create tables error:', error as string);
    throw error;
  }
};
export const dropTables = async () => {
  try {
    await db.execAsync(`DROP TABLE IF EXISTS password;`);
    await db.execAsync(`DROP TABLE IF EXISTS record;`);
    migrateTables();
  } catch (error) {
    throw error;
  }
};
