/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { DefaultTheme, NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { RootStackParamList } from './routes';
import HomeStackScreen from './HomeStackScreen';
import I18nStackScreen from '../ui/settings/I18nStackScreen';
import ThemeStackScreen from '../ui/settings/ThemeStackScreen';
import ForumStackScreen from '../ui/settings/ForumStackScreen';
import SignInStackScreen from '../ui/settings/Identity/SignInStackScreen';
import SignUpStackScreen from '../ui/settings/Identity/SignUpStackScreen';
import { useTranslation } from 'react-i18next';
import { useTheme } from '@rneui/themed';
import NewPasswordStackScreen from '../ui/passwords/NewPasswordStackScreen';
import PasswordDetailStackScreen from '../ui/passwords/PasswordDetailStackScreen';
import UnlockPasswordStackScreen from '../ui/settings/UnlockPassword/UnlockPasswordStackScreen';
import EditUnlockPasswordStackScreen from '../ui/settings/UnlockPassword/EditUnlockPasswordStackScreen';
import LockScreen from '../components/LockScreen';
import SearchPasswordStackScreen from '../ui/passwords/SearchPasswordStackScreen';
import AutoLockScreenStackScreen from '../ui/settings/UnlockPassword/AutoLockStackScreen';
import SecurityQuestionStackScreen from '../ui/settings/UnlockPassword/SecurityQuestionStackScreen';
import SecurityQuestionDetailStackScreen from '../ui/settings/UnlockPassword/SecurityQuestionDetailStackScreen';
import EditSecurityQuestionStackScreen from '../ui/settings/UnlockPassword/EditSecurityQuestionStackScreen';
import RecoverUnlockPasswordStackScreen from '../ui/settings/UnlockPassword/RecoverUnlockPasswordStackScreen';
import VerificationModeStackScreen from '../ui/settings/UnlockPassword/VerificationModeStackScreen';
import RandomPasswordStackScreen from '../ui/passwords/RandomPasswordStackScreen';
import AddTrustedContactStackScreen from '../ui/settings/UnlockPassword/AddTrustedContactStackScreen';
import PrivacyEmailStackScreen from '../ui/settings/PrivacyEmail/PrivacyEmailStackScreen';
import PrivacyEmailDetailStackScreen from '../ui/settings/PrivacyEmail/PrivacyEmailDetailStackScreen';
import NewRecordStackScreen from '../ui/records/NewRecordStackScreen';
import RecordDetailStackScreen from '../ui/records/RecordDetailStackScreen';
import FamilyStackScreen from '../ui/settings/Family/FamilyStackScreen';
import InviteFamilyMemberStackScreen from '../ui/settings/Family/InviteFamilyMemberStackScreen';
import DebugStackScreen from '../ui/debug/DebugStackScreen';
import DebugHttpLogStackScreen from '../ui/debug/DebugHttpLogStackScreen';
import DebugHttpLogDetailStackScreen from '../ui/debug/DebugHttpLogDetailStackScreen';
import DebugLogStackScreen from '../ui/debug/DebugLogStackScreen';
import FamilyMessageStackScreen from '../ui/settings/Family/FamilyMessageStackScreen';
import DataQRStackScreen from '../ui/passwords/DataQRStackScreen';
import { currentSlarkInfo, SlarkInfo } from '../services/slark';
import { SlarkInfoContext } from '../contexts/slark';
import BackupUnlockPasswordStackScreen from '../ui/settings/Family/BackupUnlockPassword';
import FamilyBackupRecoverStackScreen from '../ui/settings/Family/FamilyBackupRecoverStackScreen';
import { currentUnlockPasswordSetting } from '../services/unlockPassword';
import { buildPasswordTableIndexesByUserID } from '../common/sqlite/dao/password';
import { buildRecordTableIndexesByUserID } from '../common/sqlite/dao/record';

const Stack = createNativeStackNavigator<RootStackParamList>();

function RootStack(): React.JSX.Element {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const [slarkInfo, setSlarkInfo] = React.useState<null | SlarkInfo>(null);
  React.useEffect(() => {
    // 1. slark info
    let cachedInfo = currentSlarkInfo();
    if (cachedInfo) {
      setSlarkInfo({ ...cachedInfo });
    }
    // 2. sqlite indexes
    const userID = cachedInfo ? cachedInfo.userID : -1;
    const curSetting = currentUnlockPasswordSetting(userID);
    if (curSetting) {
      buildPasswordTableIndexesByUserID(curSetting!.passwordHash);
      buildRecordTableIndexesByUserID(curSetting!.passwordHash);
    }
  }, []);
  return (
    <SlarkInfoContext.Provider value={{ slarkInfo, setSlarkInfo }}>
      <NavigationContainer
        theme={{
          dark: theme.mode === 'dark',
          colors: {
            primary: theme.colors.primary,
            background: theme.colors.background,
            card: theme.colors.white,
            text: theme.colors.black,
            border: theme.colors.divider,
            notification: theme.colors.warning,
          },
          fonts: DefaultTheme.fonts,
        }}>
        <LockScreen>
          <Stack.Navigator>
            <Stack.Group
              screenOptions={{
                headerShown: false,
              }}>
              <Stack.Screen name="HomeStack" component={HomeStackScreen} />
            </Stack.Group>
            <Stack.Group>
              {/* passwords screens */}
              <Stack.Screen
                name="DataQRStack"
                component={DataQRStackScreen}
                options={{
                  title: t('dataQR.label'),
                }}
              />
              <Stack.Screen
                name="NewPasswordStack"
                component={NewPasswordStackScreen}
                options={{
                  title: t('passwords.newPassword.label'),
                }}
              />
              <Stack.Screen
                name="PasswordDetailStack"
                component={PasswordDetailStackScreen}
                options={{
                  title: '',
                }}
              />
              <Stack.Screen
                name="SearchPasswordStack"
                component={SearchPasswordStackScreen}
                options={{
                  title: t('passwords.searchPassword.label'),
                }}
              />
              {/* records screens */}
              <Stack.Screen
                name="NewRecordStack"
                component={NewRecordStackScreen}
                options={{
                  title: t('records.newRecord.label'),
                }}
              />
              <Stack.Screen
                name="RecordDetailStack"
                component={RecordDetailStackScreen}
                options={{
                  title: '',
                }}
              />
              {/* settings screens */}
              <Stack.Screen
                name="SignInStack"
                component={SignInStackScreen}
                options={{
                  title: t('settings.signin.label'),
                }}
              />
              <Stack.Screen
                name="SignUpStack"
                component={SignUpStackScreen}
                options={{
                  title: t('settings.signup.label'),
                }}
              />
              <Stack.Screen
                name="PrivacyEmailStack"
                component={PrivacyEmailStackScreen}
                options={{
                  title: t('settings.email.label'),
                }}
              />
              <Stack.Screen
                name="PrivacyEmailDetailStack"
                component={PrivacyEmailDetailStackScreen}
                options={{
                  title: t('settings.emailDetail.label'),
                }}
              />
              <Stack.Screen
                name="FamilyStack"
                component={FamilyStackScreen}
                options={{
                  title: t('settings.family.label'),
                }}
              />
              <Stack.Screen
                name="InviteFamilyMemberStack"
                component={InviteFamilyMemberStackScreen}
                options={{
                  title: t('settings.family.inviteFamilyMember.label'),
                }}
              />
              <Stack.Screen
                name="BackupUnlockPasswordStack"
                component={BackupUnlockPasswordStackScreen}
                options={{
                  title: t('settings.family.backupUnlockPassword.label'),
                }}
              />
              <Stack.Screen
                name="FamilyBackupRecoverStack"
                component={FamilyBackupRecoverStackScreen}
              />
              <Stack.Screen
                name="FamilyMessageStack"
                component={FamilyMessageStackScreen}
                options={{
                  title: t('settings.family.familyMessage.label'),
                }}
              />
              <Stack.Screen
                name="I18nStack"
                component={I18nStackScreen}
                options={{
                  title: t('settings.language.label'),
                }}
              />
              <Stack.Screen
                name="ThemeStack"
                component={ThemeStackScreen}
                options={{
                  title: t('settings.theme.label'),
                }}
              />
              <Stack.Screen
                name="ForumStack"
                component={ForumStackScreen}
                options={{
                  title: t('settings.forum.label'),
                }}
              />
              <Stack.Screen
                name="UnlockPasswordStack"
                component={UnlockPasswordStackScreen}
                options={{
                  title: t('settings.unlockPassword.label'),
                }}
              />
              <Stack.Screen
                name="VerificationModeStack"
                component={VerificationModeStackScreen}
                options={{
                  title: t('settings.unlockPassword.verificationMode.label'),
                }}
              />
              <Stack.Screen
                name="EditUnlockPasswordStack"
                component={EditUnlockPasswordStackScreen}
                options={{}}
              />
              <Stack.Screen
                name="AutoLockStack"
                component={AutoLockScreenStackScreen}
                options={{
                  title: t('settings.unlockPassword.autoLock.label'),
                }}
              />
              <Stack.Screen
                name="SecurityQuestionStack"
                component={SecurityQuestionStackScreen}
                options={{
                  title: t('settings.unlockPassword.securityQuestion.label'),
                }}
              />
              <Stack.Screen
                name="SecurityQuestionDetailStack"
                component={SecurityQuestionDetailStackScreen}
                options={{
                  title: t(
                    'settings.unlockPassword.securityQuestion.securityQuestionDetail.label',
                  ),
                }}
              />
              <Stack.Screen
                name="EditSecurityQuestionStack"
                component={EditSecurityQuestionStackScreen}
                options={{}}
              />
              <Stack.Screen
                name="RecoverUnlockPasswordStack"
                component={RecoverUnlockPasswordStackScreen}
                options={{
                  title: t(
                    'settings.unlockPassword.recoverUnlockPassword.label',
                  ),
                }}
              />
              {/* 可信联络人 停用 */}
              {/* <Stack.Screen
                name="TrustedContactStack"
                component={TrustedContactStackScreen}
                options={{
                  title: t('settings.unlockPassword.trustedContact.label'),
                }}
              /> */}
              <Stack.Screen
                name="AddTrustedContactStack"
                component={AddTrustedContactStackScreen}
                options={{
                  title: t('settings.unlockPassword.addTrustedContact.label'),
                }}
              />
            </Stack.Group>
            {/* fullScreenModal */}
            <Stack.Group
              screenOptions={{
                presentation: 'transparentModal',
                headerShown: false,
                animation: 'none',
              }}>
              <Stack.Screen
                name="RandomPasswordStack"
                component={RandomPasswordStackScreen}
              />
            </Stack.Group>
            {/* debug screens */}
            <Stack.Group>
              <Stack.Screen
                name="DebugStack"
                component={DebugStackScreen}
                options={{
                  title: t('debug.label'),
                }}
              />
              <Stack.Screen
                name="DebugHttpLogStack"
                component={DebugHttpLogStackScreen}
                options={{
                  title: t('debug.httpLogs.label'),
                }}
              />
              <Stack.Screen
                name="DebugHttpLogDetailStack"
                component={DebugHttpLogDetailStackScreen}
                options={{
                  title: t('debug.httpLogDetail.label'),
                }}
              />
              <Stack.Screen
                name="DebugLogStack"
                component={DebugLogStackScreen}
                options={{
                  title: t('debug.logs.label'),
                }}
              />
            </Stack.Group>
          </Stack.Navigator>
        </LockScreen>
      </NavigationContainer>
    </SlarkInfoContext.Provider>
  );
}

export default RootStack;
