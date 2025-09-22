// 从0开始；
// 如果为0，则不用执行migrationSQLs里的语句；
// 如果为1，则要执行migrationSQLs里索引为0的语句；
// 如果为2，则要执行migrationSQLs里索引为1的语句；
export const databaseVersion = 0;
export const migrationSQLs = [
  `CREATE TABLE IF NOT EXISTS test_migration (
  id INTEGER PRIMARY KEY NOT NULL, 
  message TEXT NOT NULL);`, // 测试：先以databaseVersion为0进行运行，然后将databaseVersion改为1，然后将./index.ts里相应的测试代码注解去掉来进行测试应用重启；
  ``,
  ``,
  ``,
];
