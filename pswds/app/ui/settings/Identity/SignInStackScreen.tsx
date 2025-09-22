/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { makeStyles, Text, Input, Button } from '@rneui/themed';
import { View } from 'react-native';
import { SafeAreaView, KeyboardAvoidingView, ScrollView } from 'react-native';
import { usePreventRemove } from '@react-navigation/native';
import { keccak_256 } from '@noble/hashes/sha3';
import moment from 'moment';
import {
  buildPasswordTableIndexesByUserID,
  cancelSharingPasswordsByUserIDAsync,
  clearPasswordTableIndexes,
  clearUnsignedInPasswordsAsync,
  deletePasswordByDataIDAsync,
  deletePasswordsByUserIDAsync,
  deletePasswordsFromFamilyAsync,
  getPasswordByDataIDAsync,
  getPasswordsByUserIDAsync,
  insertPasswordAsync,
  sharingPasswordByDataIDAsync,
  upsertPasswordsAsync,
  XoredPassword,
  xorPasswords,
  xorXoredPasswords,
} from '../../../common/sqlite/dao/password';
import {
  buildRecordTableIndexesByUserID,
  cancelSharingRecordsByUserIDAsync,
  clearRecordTableIndexes,
  clearUnsignedInRecordsAsync,
  deleteRecordByDataIDAsync,
  deleteRecordsByUserIDAsync,
  deleteRecordsFromFamilyAsync,
  getRecordByDataIDAsync,
  getRecordsByUserIDAsync,
  insertRecordAsync,
  sharingRecordByDataIDAsync,
  upsertRecordsAsync,
  XoredRecord,
  xorRecords,
  xorXoredRecords,
} from '../../../common/sqlite/dao/record';
import { insertSharedDataMembersAsync } from '../../../common/sqlite/dao/sharedDataMember';
import { dropTables } from '../../../common/sqlite';
import { storage } from '../../../common/mmkv';
import {
  currentSecurityQuestionSetting,
  currentUnlockPasswordSetting,
  UnlockPasswordSetting,
  updateAutoLockSetting,
  updateSecurityQuestionSetting,
  updateUnlockPasswordSetting,
} from '../../../services/unlockPassword';
import {
  checkUpdates,
  updateBackupState,
  uploadData,
} from '../../../services/backup';
import { SlarkInfo, updateSlarkInfo } from '../../../services/slark';
import { updateOtherFamilyMembers } from '../../../services/family';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';
import {
  decryptByUnlockPassword,
  decryptByXchacha20poly1305,
  decryptByUserPrivateKey,
  userPrivateKey,
} from '../../../services/cipher';
import { upsertBackup } from '../../../common/mmkv/backup';
import { xor_hex, xor_str } from '../../../common/sqlite/dao/utils';
import { Password, Record } from '../../../common/sqlite/schema';

type SignInStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function SignInStackScreen({
  navigation,
}: SignInStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { setLoading } = React.useContext(BackdropContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const [email, setEmail] = React.useState<string>('');
  const [localPassword, setLocalPassword] = React.useState<string>('');
  const { slarkInfo, setSlarkInfo } = React.useContext(SlarkInfoContext);
  const { setPassword } = React.useContext(UnlockPasswordContext);

  const simpleLogout = async (message: string) => {
    try {
      // （1）清理内存
      setSlarkInfo(null);
      setPassword('');
      // （2）清理存储
      storage.clearAll();
      await dropTables();
      // （3）发送 http
      await post('/slark/user/logout/v1');
    } catch (error) {
    } finally {
      setLoading(false);
      setError(message, t('app.toast.error'));
    }
  };

  const getOtherFamilyMembers = async (
    passwordHash: string,
    loginInfo: SlarkInfo,
  ) => {
    try {
      let respData = await checkUpdates(passwordHash, loginInfo);
      if (respData.code !== 0) {
        await simpleLogout(respData.message);
        return;
      }
      if (respData.data.otherFamilyMembers) {
        updateOtherFamilyMembers(loginInfo.userID, {
          list: respData.data.otherFamilyMembers,
        });
      }
    } catch (error) {
      throw error;
    }
  };

  const handleSecurityQuestions = (loginInfo: SlarkInfo, data: any) => {
    if (data.securityQuestions) {
      const curSetting = currentSecurityQuestionSetting(loginInfo.userID);
      if (curSetting) {
        updateSecurityQuestionSetting(loginInfo.userID, {
          ...curSetting,
          questions: data.securityQuestions,
        });
      } else {
        updateSecurityQuestionSetting(loginInfo.userID, {
          questions: data.securityQuestions,
        });
      }
    }
  };
  const handleBackupAsync = async (loginInfo: SlarkInfo, data: any) => {
    try {
      const nowTS = moment().unix();
      upsertBackup({
        createdAt: nowTS,
        updatedAt: nowTS,
        userID: loginInfo.userID,
        userPublicKey: userPrivateKey(localPassword).publicKey.toHex(),
        encryptedFamilyKey: data.encryptedFamilyKey,
      });
    } catch (error) {
      throw error;
    }
  };
  const handlePasswordsAsync = async (
    passwordHash: string,
    loginInfo: SlarkInfo,
    data: any,
  ) => {
    try {
      if (data.pwdList) {
        let list: XoredPassword[] = [];
        (JSON.parse(data.pwdList) as string[]).map(item => {
          const obj = JSON.parse(decryptByUnlockPassword(localPassword, item)); // 整体解密
          if (obj) {
            list.push(obj);
          }
        });
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
      // NOTE: 本地溢出的数据会被认为，已被其他端进行了删除操作，所以本端会进行该删除操作以达到数据一致；
      if (data.pwdIDList && data.pwdIDList.length > 0) {
        const allRows: XoredPassword[] = await getPasswordsByUserIDAsync(
          xor_str(passwordHash, loginInfo.userID.toString()),
        );
        allRows.map(async row => {
          if (!data.pwdIDList.includes(xor_hex(passwordHash, row.dataID))) {
            await deletePasswordByDataIDAsync(row.dataID);
          }
        });
      } else if (!data.pwdIDList || data.pwdIDList.length === 0) {
        await deletePasswordsByUserIDAsync(
          xor_str(passwordHash, loginInfo.userID.toString()),
        );
      }
    } catch (error) {
      throw error;
    }
  };

  const handleRecordsAsync = async (
    passwordHash: string,
    loginInfo: SlarkInfo,
    data: any,
  ) => {
    try {
      if (data.nprList) {
        let list: XoredRecord[] = [];
        (JSON.parse(data.nprList) as string[]).map(item => {
          const obj = JSON.parse(decryptByUnlockPassword(localPassword, item)); // 整体解密
          if (obj) {
            list.push(obj);
          }
        });
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
      // NOTE: 本地溢出的数据会被认为，已被其他端进行了删除操作，所以本端会进行该删除操作以达到数据一致；
      if (data.nprIDList && data.nprIDList.length > 0) {
        const allRows: XoredRecord[] = await getRecordsByUserIDAsync(
          xor_str(passwordHash, loginInfo.userID.toString()),
        );
        allRows.map(async row => {
          if (!data.nprIDList.includes(xor_hex(passwordHash, row.dataID))) {
            await deleteRecordByDataIDAsync(row.dataID);
          }
        });
      } else if (!data.nprIDList || data.nprIDList.length === 0) {
        await deleteRecordsByUserIDAsync(
          xor_str(passwordHash, loginInfo.userID.toString()),
        );
      }
    } catch (error) {
      throw error;
    }
  };
  const handleSharingPassword = async (
    passwordHash: string,
    content: any,
    item: any,
  ) => {
    try {
      const entity: XoredPassword | null = await getPasswordByDataIDAsync(
        content.dataID,
      );
      if (entity) {
        // update
        await sharingPasswordByDataIDAsync(
          entity.dataID,
          xor_str(passwordHash, item.updatedAt.toString()),
          item.sharedToAll ? xor_str(passwordHash, '1') : null,
        );
      } else {
        // insert
        await insertPasswordAsync({
          ...content,
          sharedAt: xor_str(passwordHash, item.updatedAt.toString()),
          sharedToAll: item.sharedToAll ? xor_str(passwordHash, '1') : null,
        });
      }
    } catch (error) {
      throw error;
    }
  };
  const handleSharingRecord = async (
    passwordHash: string,
    content: XoredRecord,
    item: any,
  ) => {
    try {
      const record: XoredRecord | null = await getRecordByDataIDAsync(
        content.dataID,
      );
      if (record) {
        // update
        await sharingRecordByDataIDAsync(
          record.dataID,
          xor_str(passwordHash, item.updatedAt.toString()),
          item.sharedToAll ? xor_str(passwordHash, '1') : null,
        );
      } else {
        // insert
        await insertRecordAsync({
          ...content,
          sharedAt: xor_str(passwordHash, item.updatedAt.toString()),
          sharedToAll: item.sharedToAll ? xor_str(passwordHash, '1') : null,
        });
      }
    } catch (error) {
      throw error;
    }
  };
  const handleFamilySharingData = async (
    loginInfo: SlarkInfo,
    password: string,
    passwordHash: string,
    data: any,
  ) => {
    try {
      if (data.encryptedFamilyKey) {
        const familyKey = decryptByUserPrivateKey(
          Buffer.from(data.encryptedFamilyKey, 'hex'),
          password,
        );
        const xoredUserID = xor_str(passwordHash, loginInfo.userID.toString());
        // delete local shared data
        await deletePasswordsFromFamilyAsync(passwordHash, xoredUserID);
        await cancelSharingPasswordsByUserIDAsync(xoredUserID);
        await deleteRecordsFromFamilyAsync(passwordHash, xoredUserID);
        await cancelSharingRecordsByUserIDAsync(xoredUserID);
        if (data.sharingList.length > 0) {
          // update sharing states
          data.sharingList.map(async (item: any) => {
            const content = JSON.parse(
              decryptByXchacha20poly1305(familyKey, item.content),
            );
            if (item.type === 'password') {
              await handleSharingPassword(passwordHash, content, item);
            } else {
              await handleSharingRecord(passwordHash, content, item);
            }
          });
        }
      }
    } catch (error) {
      throw error;
    }
  };
  const handleUnsignedIn = async (
    loginInfo: SlarkInfo,
    password: string,
    passwordHash: string,
    data: any,
  ): Promise<boolean> => {
    try {
      const unsignedSetting = currentUnlockPasswordSetting(-1);
      if (unsignedSetting) {
        // a. 获取未登录时的数据
        const xoredUnsignedUserID = xor_str(unsignedSetting.passwordHash, '-1');
        let passwords: Password[] = xorXoredPasswords(
          unsignedSetting.passwordHash,
          await getPasswordsByUserIDAsync(xoredUnsignedUserID),
        );
        let records: Record[] = xorXoredRecords(
          unsignedSetting.passwordHash,
          await getRecordsByUserIDAsync(xoredUnsignedUserID),
        );
        // b. 将未登录的数据更新，再同步到后端
        for (let i = 0; i < passwords.length; i++) {
          passwords[i].userID = loginInfo.userID;
        }
        for (let i = 0; i < records.length; i++) {
          records[i].userID = loginInfo.userID;
        }
        if (passwords.length > 0 || records.length > 0) {
          let uploadRespData = await uploadData(
            loginInfo,
            password,
            data.passwordHash,
            data.encryptedFamilyKey,
            xorPasswords(passwordHash, passwords),
            xorRecords(passwordHash, records),
          );
          if (uploadRespData.code !== 0) {
            await simpleLogout(uploadRespData.message);
            return false;
          }
        }
        // c. 将本地未登录的数据的状态进行更新
        const xoredUserID = xor_str(passwordHash, loginInfo.userID.toString());
        await clearUnsignedInPasswordsAsync(
          xoredUnsignedUserID,
          passwordHash,
          xoredUserID,
        );
        await clearUnsignedInRecordsAsync(
          xoredUnsignedUserID,
          passwordHash,
          xoredUserID,
        );
        updateUnlockPasswordSetting(loginInfo.userID, {
          passwordHash: data.passwordHash,
          supportFingerprint: unsignedSetting.supportFingerprint,
        });
        // 清除未登录时的解锁密码设置
        updateUnlockPasswordSetting(-1, null);
        updateAutoLockSetting(-1, null);
      } else {
        // 如果本地没有未登录的解锁密码设置
        updateUnlockPasswordSetting(loginInfo.userID, {
          passwordHash: data.passwordHash,
        });
      }
      return true;
    } catch (error) {
      throw error;
    }
  };
  const handleSignInResult = async (
    passwordHash: string,
    loginInfo: SlarkInfo,
  ) => {
    try {
      // （1）用户进行登录，就一定会有解锁密码；所以登录成功后直接拉取后台数据；
      // （1-1）家庭其他成员列表
      await getOtherFamilyMembers(passwordHash, loginInfo);
      // （1-2）解锁密码的密保问题、user public key、encrypted family key、password列表、record列表、家庭共享数据列表、家庭共享数据的分享成员列表
      let respData = await post('/pswds/downloadData/v1', {
        updatedAt: 0,
      });
      if (respData.code !== 0) {
        await simpleLogout(respData.message);
        return;
      }
      // a. 解锁密码的密保问题
      handleSecurityQuestions(loginInfo, respData.data);
      // b. 本地 backup 记录
      await handleBackupAsync(loginInfo, respData.data);
      // c. password列表
      await handlePasswordsAsync(passwordHash, loginInfo, respData.data);
      // d. record列表
      await handleRecordsAsync(passwordHash, loginInfo, respData.data);
      // e. 家庭共享数据列表
      await handleFamilySharingData(
        loginInfo,
        localPassword,
        passwordHash,
        respData.data,
      );
      // f. 家庭共享数据的分享成员列表
      if (respData.data.sharedData) {
        await insertSharedDataMembersAsync(
          passwordHash,
          respData.data.sharedData,
        );
      }
      // g. 更新用户本地备份的更新时间戳
      updateBackupState(loginInfo.userID, {
        updatedAt: respData.data.updatedAt,
      });
      // （1-3）如果本地有未登录的解锁密码设置
      const ok = await handleUnsignedIn(
        loginInfo,
        localPassword,
        passwordHash,
        respData.data,
      );
      if (!ok) {
        return;
      }
      // （1-4）build indexes
      clearPasswordTableIndexes();
      buildPasswordTableIndexesByUserID(passwordHash);
      clearRecordTableIndexes();
      buildRecordTableIndexesByUserID(passwordHash);
      // （1-5）缓存用户登录信息、解锁密码
      setSlarkInfo(loginInfo);
      updateSlarkInfo(loginInfo);
      setPassword(localPassword);
      setLoading(false);
      // （1-6）显示登录成功
      setSuccess(t('app.toast.success'));
      // （1-7）正常跳转
      navigation.navigate('HomeStack', { screen: 'Passwords' });
    } catch (error) {
      throw error;
    }
  };

  const loginBySecondaryPassword = async () => {
    // 1. handle parameters
    if (!email) {
      setError(t('settings.signin.toast.emptyEmail'), t('app.toast.error'));
      return;
    }
    if (!localPassword) {
      setError(
        t('settings.signin.toast.emptyUnlockPassword'),
        t('app.toast.error'),
      );
      return;
    }
    const passwordHash = Buffer.from(keccak_256(localPassword)).toString('hex');
    try {
      // 2. post request
      setLoading(true);
      let respData = await post('/slark/loginBySecondaryPassword/v1', {
        email,
        passwordHash,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.error'));
        return;
      }
      // 3. handle results
      const loginInfo: SlarkInfo = {
        ...respData.data,
        lSessionID: respData.lSessionID,
      };
      await handleSignInResult(passwordHash, loginInfo);
    } catch (error) {
      await simpleLogout(error as string);
    }
  };

  const onPressSignIn = async () => {
    await loginBySecondaryPassword();
  };

  const onPressNoAccount = () => {
    navigation.navigate('EditUnlockPasswordStack', {});
  };

  const onPressSignUp = () => {
    navigation.navigate('SignUpStack');
  };

  const onChangeEmail = (newText: string) => {
    newText = newText.trim();
    setEmail(newText);
  };

  const onChangePassword = (newText: string) => {
    newText = newText.trim();
    setLocalPassword(newText);
  };

  const [currentSetting, setCurrentSetting] =
    React.useState<null | UnlockPasswordSetting>(null);

  React.useEffect(() => {
    let curSetting = currentUnlockPasswordSetting(
      slarkInfo ? slarkInfo.userID : -1,
    );
    if (slarkInfo) {
      // 有登录信息，但是没有设置解锁密码
      if (!curSetting) {
        setTimeout(() => {
          navigation.navigate('UnlockPasswordStack');
        }, 0);
        return;
      }
    }
    setCurrentSetting(curSetting);
    // 导航栏
    navigation.setOptions({
      headerBackVisible: curSetting !== null,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  usePreventRemove(slarkInfo == null && currentSetting == null, () => {});

  const onPressRecoverUnlockPassword = () => {
    navigation.navigate('RecoverUnlockPasswordStack');
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView>
        <ScrollView>
          <View style={styles.body}>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                label={t('settings.signin.email')}
                labelStyle={styles.inputLabel}
                inputStyle={styles.inputStyle}
                placeholder={t('settings.signin.emailPlaceholder')}
                onChangeText={onChangeEmail}
                value={email}
              />
            </View>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                label={t('settings.signin.unlockPassword')}
                labelStyle={styles.inputLabel}
                inputStyle={styles.inputStyle}
                placeholder={t('settings.signin.unlockPasswordPlaceholder')}
                onChangeText={onChangePassword}
                value={localPassword}
              />
            </View>
            <View style={styles.row}>
              <Button
                title={t('settings.signin.label')}
                containerStyle={styles.signInBtn}
                titleStyle={styles.signInTitle}
                size="lg"
                radius={8}
                onPress={onPressSignIn}
              />
            </View>
            <View style={[styles.row, styles.tipContainer]}>
              <Text />
              <Text
                style={styles.recoverTipText}
                onPress={onPressRecoverUnlockPassword}>
                {t('lockScreen.recoverUnlockPassword')}
              </Text>
            </View>
            {currentSetting == null && (
              <View style={[styles.row, styles.tipContainer]}>
                <Text />
                <Text
                  style={styles.noAccountTipText}
                  onPress={onPressNoAccount}>
                  {t('settings.signin.noAccount')}
                </Text>
              </View>
            )}
            <View style={[styles.row, styles.tipContainer]}>
              <Text />
              <Text style={styles.signUpTipText} onPress={onPressSignUp}>
                {t('settings.signin.goToSignUp')}
              </Text>
            </View>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
    marginTop: 8,
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  head: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 16,
    fontSize: 36,
    fontWeight: 'bold',
  },
  body: {
    flex: 1,
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 8,
    marginHorizontal: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  inputLabel: {
    fontSize: 20,
    color: theme.colors.black,
  },
  inputStyle: {
    height: 60,
  },
  tipContainer: { justifyContent: 'space-between' },
  noAccountTipText: {
    color: theme.colors.primary,
    fontSize: 16,
  },
  signUpTipText: {
    color: theme.colors.primary,
    fontSize: 16,
  },
  signInBtn: {
    width: '100%',
  },
  signInTitle: { fontSize: 20 },
  recoverTipText: {
    color: theme.colors.warning,
  },
}));

export default SignInStackScreen;
