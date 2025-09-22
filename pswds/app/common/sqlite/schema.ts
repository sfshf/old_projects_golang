// password schema
export interface Password {
  dataID: string;
  createdAt: number;
  updatedAt: number;
  userID: number;
  title: string;
  website: string | null;
  username: string;
  password: string;
  notes: string | null;
  others: string | null;
  usedAt: number | null;
  usedCount: number | null;
  iconBgColor: number | null;
  sharedAt: number | null;
  sharedToAll: number | null;
  sharingMembers: string | null;
}
export const passwordTable = 'password';
export const passwordSchema = `
CREATE TABLE IF NOT EXISTS password (
  dataID TEXT NOT NULL UNIQUE, 
  createdAt TEXT NOT NULL,
  updatedAt TEXT NOT NULL,
  userID TEXT NOT NULL,
  title TEXT NOT NULL, 
  website TEXT NULL,
  username TEXT NOT NULL,
  password TEXT NOT NULL,
  notes TEXT NULL,
  others TEXT NULL,
  usedAt TEXT NULL,
  usedCount TEXT NULL,
  iconBgColor TEXT NULL,
  sharedAt TEXT NULL,
  sharedToAll TEXT NULL, 
  sharingMembers TEXT NULL);`;
export const iconBgColors = [
  '#2089dc',
  '#ad1457',
  '#52c41a',
  '#faad14',
  '#ff190c',
];

// record (non password) schema
export interface Record {
  dataID: string;
  createdAt: number;
  updatedAt: number;
  userID: number;
  recordType: string;
  title: string;
  iconBgColor: number | null;
  usedAt: number | null;
  usedCount: number | null;
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
  sharedAt: number | null;
  sharedToAll: number | null;
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
export const recordTable = 'record';
export const recordSchema = `
CREATE TABLE IF NOT EXISTS record (
  dataID TEXT NOT NULL UNIQUE, 
  createdAt TEXT NOT NULL,
  updatedAt TEXT NOT NULL,
  userID TEXT NOT NULL,
  recordType TEXT NOT NULL, 
  title TEXT NOT NULL, 
  iconBgColor TEXT NULL, 
  usedAt TEXT NULL, 
  usedCount TEXT NULL, 
  -- mixed fields
  phone TEXT NULL, 
  type TEXT NULL, 
  number TEXT NULL, 
  address TEXT NULL, 
  fullName TEXT NULL, 
  birthDate TEXT NULL, 
  gender TEXT NULL, 
  pin TEXT NULL, 
  expiryDate TEXT NULL, 
  others TEXT NULL, 
  sharedAt TEXT NULL, 
  sharedToAll TEXT NULL,
  sharingMembers TEXT NULL, 
  -- identity fields
  firstName TEXT NULL, 
  lastName TEXT NULL, 
  job TEXT NULL, 
  socialSecurityNumber TEXT NULL, 
  idNumber TEXT NULL, 
  -- credit card fields
  cardholderName TEXT NULL, 
  verificationNumber TEXT NULL, 
  validFrom TEXT NULL, 
  issuingBank TEXT NULL, 
  -- bank account fields
  bankName TEXT NULL, 
  nameOnAccount TEXT NULL, 
  routingNumber TEXT NULL, 
  branch TEXT NULL, 
  accountNumber TEXT NULL, 
  swift TEXT NULL, 
  -- driver license fields
  height TEXT NULL,
  licenseClass TEXT NULL,
  state TEXT NULL,
  country TEXT NULL,
  -- passport fields
  issuingCountry TEXT NULL,
  nationality TEXT NULL,
  issuingAuthority TEXT NULL,
  birthPlace TEXT NULL,
  issuedOn TEXT NULL);`;
export declare type RecordType =
  | 'identity'
  | 'credit card'
  | 'bank account'
  | 'driver license'
  | 'passport';
