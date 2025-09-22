import { NavigatorScreenParams } from '@react-navigation/native';
import { Password, Record, RecordType } from '../common/sqlite/schema';
import { SlarkInfo } from '../services/slark';
import { FamilyMember } from '../ui/settings/Family/FamilyStackScreen';

export type RootStackParamList = {
  // 1. home routes
  HomeStack: NavigatorScreenParams<HomeTabsParamList>;
  // 2. passwords routes
  RandomPasswordStack: undefined;
  NewPasswordStack: { password?: string };
  PasswordDetailStack: { dataID: string };
  SearchPasswordStack: undefined;
  DataQRStack: { entity: Password | Record };
  // 3. records routes
  NewRecordStack: { recordType: RecordType };
  RecordDetailStack: { dataID: string };
  // 4. settings routes
  SignInStack: undefined;
  SignUpStack: undefined;
  PrivacyEmailStack: { deleted?: number };
  PrivacyEmailDetailStack: { id: number };
  FamilyStack: undefined;
  InviteFamilyMemberStack: undefined;
  FamilyMessageStack: undefined;
  BackupUnlockPasswordStack: { familyMembers: FamilyMember[] };
  FamilyBackupRecoverStack: { member: FamilyMember };
  I18nStack: undefined;
  ThemeStack: undefined;
  ForumStack: undefined;
  /// unlock password stack screens
  UnlockPasswordStack: undefined;
  VerificationModeStack: undefined;
  EditUnlockPasswordStack: { loginInfo?: SlarkInfo };
  AutoLockStack: undefined;
  SecurityQuestionStack: undefined;
  SecurityQuestionDetailStack: undefined;
  EditSecurityQuestionStack: undefined;
  RecoverUnlockPasswordStack: undefined;
  // TrustedContactStack: undefined;  // 可信联络人 停用
  AddTrustedContactStack: undefined;
  /// debug stack screens
  DebugStack: undefined;
  DebugHttpLogStack: undefined;
  DebugHttpLogDetailStack: { index: number };
  DebugLogStack: undefined;
};

export type HomeTabsParamList = {
  Home: undefined;
  Passwords: undefined;
  Records: undefined;
  Settings: undefined;
};
