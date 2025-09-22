/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { Keyboard, View } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Text, Input, Button, makeStyles, Icon } from '@rneui/themed';
import { useFocusEffect, usePreventRemove } from '@react-navigation/native';
import { z } from 'zod';
import { keccak_256 } from '@noble/hashes/sha3';
import { SafeAreaView, ScrollView, KeyboardAvoidingView } from 'react-native';
import moment from 'moment';
import {
  buildPasswordTableIndexesByUserID,
  changeUnlockPasswordAsync_Password,
  clearPasswordTableIndexes,
  clearUnsignedInPasswordsAsync,
  deletePasswordsByUserIDAsync,
  getPasswordsByUserIDAsync,
  insertPasswordsAsync,
  XoredPassword,
  xorPasswords,
  xorXoredPasswords,
} from '../../../common/sqlite/dao/password';
import {
  buildRecordTableIndexesByUserID,
  changeUnlockPasswordAsync_Record,
  clearRecordTableIndexes,
  clearUnsignedInRecordsAsync,
  deleteRecordsByUserIDAsync,
  getRecordsByUserIDAsync,
  insertRecordsAsync,
  XoredRecord,
  xorRecords,
  xorXoredRecords,
} from '../../../common/sqlite/dao/record';
import { dropTables } from '../../../common/sqlite';
import { storage } from '../../../common/mmkv';
import {
  currentUnlockPasswordSetting,
  UnlockPasswordSetting,
  updateAutoLockSetting,
  updateSecurityQuestionSetting,
  updateUnlockPasswordSetting,
} from '../../../services/unlockPassword';
import { syncBackup, uploadData } from '../../../services/backup';
import { SlarkInfo, updateSlarkInfo } from '../../../services/slark';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';
import {
  encryptedByUserPublicKey,
  userPrivateKey,
} from '../../../services/cipher';
import { getFamilyKey } from '../../../services/family';
import { xor_str } from '../../../common/sqlite/dao/utils';
import { upsertBackup } from '../../../common/mmkv/backup';
import { Password, Record } from '../../../common/sqlite/schema';

interface tooltips {
  tip1: string;
  openTip1: boolean;
  tip2: string;
  openTip2: boolean;
}

type TooltipsAction =
  | { type: 'setTip1'; value: tooltips['tip1'] }
  | { type: 'setOpenTip1'; value: tooltips['openTip1'] }
  | {
      type: 'setTip2';
      value: tooltips['tip2'];
    }
  | { type: 'setOpenTip2'; value: tooltips['openTip2'] };

const tooltipsReducer = (state: tooltips, action: TooltipsAction) => {
  switch (action.type) {
    case 'setTip1':
      return { ...state, tip1: action.value };
    case 'setOpenTip1':
      return { ...state, openTip1: action.value };
    case 'setTip2':
      return { ...state, tip2: action.value };
    case 'setOpenTip2':
      return { ...state, openTip2: action.value };
    default:
      return { ...state };
  }
};

type EditUnlockPasswordStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'EditUnlockPasswordStack'
>;

function EditUnlockPasswordStackScreen({
  navigation,
  route,
}: EditUnlockPasswordStackScreenProp): React.JSX.Element {
  const { setSlarkInfo } = React.useContext(SlarkInfoContext);
  const { password, setPassword } = React.useContext(UnlockPasswordContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const [currentSetting, setCurrentSetting] =
    React.useState<null | UnlockPasswordSetting>(null);
  // 长度限制为 6 - 32位之间
  const passwordSchema = z.string().regex(/^.{6,32}$/, {
    message: t(
      'settings.unlockPassword.editUnlockPassword.toast.passwordTypeError',
    ),
  });
  const [password1, setPassword1] = React.useState<string>('');
  const [password2, setPassword2] = React.useState<string>('');
  const [showPassword1, setShowPassword1] = React.useState<boolean>(true);
  const [showPassword2, setShowPassword2] = React.useState<boolean>(true);

  const validateParameters = (): boolean => {
    const validation1 = passwordSchema.safeParse(password1);
    if (!validation1.success) {
      setError(validation1.error.issues[0].message, t('app.toast.error'));
      return false;
    }
    const validation2 = passwordSchema.safeParse(password2);
    if (!validation2.success) {
      setError(validation2.error.issues[0].message, t('app.toast.error'));
      return false;
    }
    if (password1 !== password2) {
      setError(
        t(
          'settings.unlockPassword.editUnlockPassword.toast.inconsistentPasswords',
        ),
        t('app.toast.error'),
      );
      return false;
    }
    return true;
  };

  const handleUnsignedIn = React.useCallback(
    async (password1Hash: string) => {
      // 1. unlock password
      const oldSetting = currentUnlockPasswordSetting(-1);
      if (!oldSetting) {
        // no sqlite data, definitely.
        updateUnlockPasswordSetting(-1, {
          passwordHash: password1Hash,
        });
        return;
      }
      // maybe has sqlite data.
      await changeUnlockPasswordAsync_Password(
        oldSetting.passwordHash,
        password1Hash,
      );
      await changeUnlockPasswordAsync_Record(
        oldSetting.passwordHash,
        password1Hash,
      );
      updateUnlockPasswordSetting(-1, {
        passwordHash: password1Hash,
        supportFingerprint: oldSetting.supportFingerprint,
      });
      // build indexes
      clearPasswordTableIndexes();
      buildPasswordTableIndexesByUserID(password1Hash);
      clearRecordTableIndexes();
      buildRecordTableIndexesByUserID(password1Hash);
      setTimeout(() => {
        navigation.navigate('HomeStack', { screen: 'Home' });
      }, 0);
      setPassword(password1);
      setCurrentSetting({
        passwordHash: password1Hash,
      });
      setSuccess(t('app.toast.success'));
    },
    [password1, password, password2, route.params.loginInfo],
  );

  const getUpdatedPasswords = async (
    oldPasswordHash: string,
    passwordHash: string,
    updatedAt: number,
  ): Promise<XoredPassword[]> => {
    try {
      // NOTE：更新用户数据，所有的updatedAt（important！）
      const passwords: Password[] = xorXoredPasswords(
        oldPasswordHash,
        await getPasswordsByUserIDAsync(
          xor_str(oldPasswordHash, route.params.loginInfo!.userID.toString()),
        ),
      );
      for (let i = 0; i < passwords.length; i++) {
        passwords[i].updatedAt = updatedAt;
      }
      return xorPasswords(passwordHash, passwords);
    } catch (error) {
      throw error;
    }
  };

  const getUpdatedRecords = async (
    oldPasswordHash: string,
    passwordHash: string,
    updatedAt: number,
  ): Promise<XoredRecord[]> => {
    try {
      // NOTE：更新用户数据，所有的updatedAt（important！）
      const records: Record[] = xorXoredRecords(
        oldPasswordHash,
        await getRecordsByUserIDAsync(
          xor_str(oldPasswordHash, route.params.loginInfo!.userID.toString()),
        ),
      );
      for (let i = 0; i < records.length; i++) {
        records[i].updatedAt = updatedAt;
      }
      return xorRecords(passwordHash, records);
    } catch (error) {
      throw error;
    }
  };

  const updateAllData = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password1Hash: string,
      oldSetting: UnlockPasswordSetting,
    ): Promise<boolean> => {
      try {
        const updatedAt = moment().unix();
        const passwords: XoredPassword[] = await getUpdatedPasswords(
          oldSetting.passwordHash,
          password1Hash,
          updatedAt,
        );
        const records: XoredRecord[] = await getUpdatedRecords(
          oldSetting.passwordHash,
          password1Hash,
          updatedAt,
        );
        // (3) 上传更新后的解锁密码、encryptedFamilyKey和用户数据
        const familyKey = await getFamilyKey(password);
        const respData = await uploadData(
          slarkInfo,
          password1,
          password1Hash,
          familyKey.toString()
            ? encryptedByUserPublicKey(
                familyKey,
                userPrivateKey(password1).publicKey.toBytes(),
              ).toString('hex')
            : '',
          passwords,
          records,
        );
        if (respData.code !== 0) {
          setLoading(false);
          setError(respData.message, t('app.toast.syncBackupDataError'));
          return false;
        }
        // （4）删除本地的密保问题和可信联络人
        updateSecurityQuestionSetting(slarkInfo.userID, null);
        // （5）更新解锁密码设置
        updateUnlockPasswordSetting(slarkInfo.userID, {
          passwordHash: password1Hash,
          supportFingerprint: oldSetting.supportFingerprint,
        });
        // （6）更新到本地数据库
        await deletePasswordsByUserIDAsync(
          xor_str(oldSetting.passwordHash, slarkInfo.userID.toString()),
        );
        await insertPasswordsAsync(passwords);
        await deleteRecordsByUserIDAsync(
          xor_str(oldSetting.passwordHash, slarkInfo.userID.toString()),
        );
        await insertRecordsAsync(records);
        return true;
      } catch (error) {
        throw error;
      }
    },
    [password, password1, password2, route.params.loginInfo],
  );

  const uploadUnsignedData = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password1Hash: string,
      unsignedPasswordHash?: string,
      xoredUnsignedUserID?: string,
    ) => {
      try {
        const passwords: Password[] = [];
        const records: Record[] = [];
        if (unsignedPasswordHash && xoredUnsignedUserID) {
          passwords.push(
            ...xorXoredPasswords(
              unsignedPasswordHash,
              await getPasswordsByUserIDAsync(xoredUnsignedUserID),
            ),
          );
          for (let i = 0; i < passwords.length; i++) {
            passwords[i].userID = slarkInfo.userID;
          }
          records.push(
            ...xorXoredRecords(
              unsignedPasswordHash,
              await getRecordsByUserIDAsync(xoredUnsignedUserID),
            ),
          );
          for (let i = 0; i < records.length; i++) {
            records[i].userID = slarkInfo.userID;
          }
        }
        const uploadRespData = await uploadData(
          slarkInfo,
          password1,
          password1Hash,
          '',
          passwords.length > 0 ? xorPasswords(password1Hash, passwords) : [],
          records.length > 0 ? xorRecords(password1Hash, records) : [],
        );
        if (uploadRespData.code !== 0) {
          // 清理存储
          setSlarkInfo(null);
          setPassword('');
          storage.clearAll();
          await dropTables();
          await post('/slark/user/logout/v1');
          setLoading(false);
          setError(uploadRespData.message, t('app.toast.syncBackupDataError'));
          return false;
        }
      } catch (error) {
        throw error;
      }
    },
    [route.params.loginInfo, password, password1, password2],
  );
  const uploadNewData = React.useCallback(
    async (slarkInfo: SlarkInfo, password1Hash: string) => {
      try {
        const xoredUserID = xor_str(password1Hash, slarkInfo.userID.toString());
        // （1）本地创建backup记录
        const nowTS = moment().unix();
        upsertBackup({
          createdAt: nowTS,
          updatedAt: nowTS,
          userID: slarkInfo.userID,
          userPublicKey: userPrivateKey(password1).publicKey.toHex(),
          encryptedFamilyKey: null,
        });
        const unsignedSetting = currentUnlockPasswordSetting(-1);
        if (!unsignedSetting) {
          await uploadUnsignedData(
            slarkInfo,
            password1Hash,
            undefined,
            undefined,
          );
          updateUnlockPasswordSetting(slarkInfo.userID, {
            passwordHash: password1Hash,
          });
          return true;
        } else {
          const xoredUnsignedUserID = xor_str(
            unsignedSetting.passwordHash,
            '-1',
          );
          // （2）远程创建backup记录；将本地可能存在的未登录时的数据上传；
          await uploadUnsignedData(
            slarkInfo,
            password1Hash,
            unsignedSetting.passwordHash,
            xoredUnsignedUserID,
          );
          // （3）将本地未登录的数据的状态进行更新
          await clearUnsignedInPasswordsAsync(
            xoredUnsignedUserID,
            password1Hash,
            xoredUserID,
          );
          await clearUnsignedInRecordsAsync(
            xoredUnsignedUserID,
            password1Hash,
            xoredUserID,
          );
          updateUnlockPasswordSetting(slarkInfo.userID, {
            passwordHash: password1Hash,
            supportFingerprint: unsignedSetting.supportFingerprint,
          });
          // 清除未登录时的应用设置
          updateUnlockPasswordSetting(-1, null);
          updateAutoLockSetting(-1, null);
          return true;
        }
      } catch (error) {
        throw error;
      }
    },
    [password1, password, password2],
  );

  const handleSignedIn = React.useCallback(
    async (slarkInfo: SlarkInfo, password1Hash: string) => {
      setLoading(true);
      try {
        // 1. 本地有锁屏密码设置，是登录后的操作；
        let oldSetting = currentUnlockPasswordSetting(slarkInfo.userID);
        if (oldSetting) {
          // (1) 先同步一下后台数据
          let respData = await syncBackup(slarkInfo, password);
          if (respData.code !== 0) {
            setLoading(false);
            setError(respData.message, t('app.toast.requestError'));
            return;
          }
          // (2) 更新所有数据，并上传；
          if (!(await updateAllData(slarkInfo, password1Hash, oldSetting))) {
            setLoading(false);
            return;
          }
        } else {
          // 2. 本地没有锁屏密码设置，刚从注册页跳转过来；
          if (!(await uploadNewData(slarkInfo, password1Hash))) {
            setLoading(false);
            return;
          }
        }
        // build indexes
        clearPasswordTableIndexes();
        buildPasswordTableIndexesByUserID(password1Hash);
        clearRecordTableIndexes();
        buildRecordTableIndexesByUserID(password1Hash);
        // update slark info context
        setSlarkInfo(slarkInfo);
        updateSlarkInfo(slarkInfo);
        setTimeout(() => {
          navigation.navigate('HomeStack', { screen: 'Home' });
        }, 0);
        setPassword(password1);
        setCurrentSetting({
          passwordHash: password1Hash,
        });
        setLoading(false);
        setSuccess(t('app.toast.success'));
      } catch (error) {
        setLoading(false);
        setError(error as string, t('app.toast.internalError'));
      }
    },
    [password, password1, password2],
  );

  const createUnlockPassword = React.useCallback(async () => {
    // 1. validate parameters
    const valid = validateParameters();
    if (!valid) {
      return;
    }
    // 2. create unlock password
    const password1Hash = Buffer.from(keccak_256(password1)).toString('hex');
    // 2-1. 没有登录
    if (!route.params.loginInfo) {
      await handleUnsignedIn(password1Hash);
      return;
    }
    // 2-2. 有登录
    await handleSignedIn(route.params.loginInfo, password1Hash);
  }, [route.params.loginInfo, password, password1, password2]);

  const initTooltips: tooltips = {
    tip1: 'settings.unlockPassword.editUnlockPassword.alert.message1',
    openTip1: false,
    tip2: 'app.alert.mustSetUnlockPasswordMessage',
    openTip2: false,
  };

  const [tooltips, dispathTooltips] = React.useReducer(
    tooltipsReducer,
    initTooltips,
  );

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackButtonDisplayMode: 'minimal',
      title: t('settings.unlockPassword.editUnlockPassword.label'),
    });
  }, [navigation]);

  useFocusEffect(
    React.useCallback(() => {
      let curSetting = currentUnlockPasswordSetting(
        route.params.loginInfo ? route.params.loginInfo.userID : -1,
      );
      setCurrentSetting(curSetting);
      if (curSetting && !tooltips.openTip1) {
        dispathTooltips({ type: 'setOpenTip1', value: true });
      }
    }, [route.params.loginInfo, tooltips]),
  );

  usePreventRemove(
    route.params.loginInfo != null && currentSetting == null,
    () => {
      if (!tooltips.openTip2) {
        dispathTooltips({ type: 'setOpenTip2', value: true });
      }
    },
  );

  const onPressShowPassword1 = () => {
    setShowPassword1(!showPassword1);
  };

  const onPressShowPassword2 = () => {
    setShowPassword2(!showPassword2);
  };

  const onPressCreateUnlockPassword = async () => {
    Keyboard.dismiss();
    await createUnlockPassword();
  };

  const onChangePassword1 = (newText: string) => {
    newText = newText.trim();
    setPassword1(newText);
  };

  const onChangePassword2 = (newText: string) => {
    newText = newText.trim();
    setPassword2(newText);
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView>
        <ScrollView>
          <View style={styles.body}>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                autoFocus
                label={t(
                  'settings.unlockPassword.editUnlockPassword.password1',
                )}
                maxLength={32}
                labelStyle={styles.inputLabel}
                style={styles.inputStyle}
                placeholder={t(
                  'settings.unlockPassword.editUnlockPassword.password1Placeholder',
                )}
                secureTextEntry={!showPassword1}
                onChangeText={onChangePassword1}
                value={password1}
                rightIcon={
                  <Icon
                    type="font-awesome-5"
                    name={showPassword1 ? 'eye' : 'eye-slash'}
                    size={18}
                    onPress={onPressShowPassword1}
                  />
                }
              />
            </View>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                label={t(
                  'settings.unlockPassword.editUnlockPassword.password2',
                )}
                maxLength={32}
                labelStyle={styles.inputLabel}
                style={styles.inputStyle}
                placeholder={t(
                  'settings.unlockPassword.editUnlockPassword.password2Placeholder',
                )}
                secureTextEntry={!showPassword2}
                onChangeText={onChangePassword2}
                value={password2}
                rightIcon={
                  <Icon
                    type="font-awesome-5"
                    name={showPassword2 ? 'eye' : 'eye-slash'}
                    size={18}
                    onPress={onPressShowPassword2}
                  />
                }
              />
            </View>
            {tooltips.openTip1 && (
              <View style={styles.row}>
                <Icon
                  type="feather"
                  name="alert-triangle"
                  color="#f4d243"
                  containerStyle={styles.warnCtn}
                />
                <Text style={styles.warnText}>{t(tooltips.tip1)}</Text>
              </View>
            )}
            {tooltips.openTip2 && (
              <View style={styles.row}>
                <Icon
                  type="feather"
                  name="alert-triangle"
                  color="#f4d243"
                  containerStyle={styles.warnCtn}
                />
                <Text style={styles.warnText}>{t(tooltips.tip2)}</Text>
              </View>
            )}
            <View style={styles.row}>
              <Button
                title={t(
                  'settings.unlockPassword.editUnlockPassword.createBtn',
                )}
                containerStyle={styles.btn}
                titleStyle={styles.btnTitle}
                size="lg"
                radius={8}
                onPress={onPressCreateUnlockPassword}
              />
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
  body: {
    flex: 1,
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 16,
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  inputLabel: {
    fontSize: 20,
    color: theme.colors.black,
  },
  inputStyle: {
    height: 60,
  },
  btn: {
    width: '100%',
  },
  btnTitle: { fontSize: 20 },
  warnCtn: { marginHorizontal: 8 },
  warnText: {
    fontWeight: 'bold',
    marginHorizontal: 8,
    flex: 1,
    flexWrap: 'wrap',
  },
}));

export default EditUnlockPasswordStackScreen;
