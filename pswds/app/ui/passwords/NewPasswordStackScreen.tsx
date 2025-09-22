/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { RootStackParamList } from '../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { makeStyles, Text, Input, Button, Card } from '@rneui/themed';
import { Keyboard, View, ScrollView, KeyboardAvoidingView } from 'react-native';
import moment from 'moment';
import { v4 as uuidv4 } from 'uuid';
import InputToolbar from '../../components/InputToolbar';
import { useFocusEffect } from '@react-navigation/native';
import { z } from 'zod';
import { SafeAreaView } from 'react-native';
import { Password } from '../../common/sqlite/schema';
import FieldTypeOverlay, {
  FieldCard,
  OtherField,
} from '../../components/FieldTypeOverlay';
import {
  getFrequentlyUsedUsernamesByUserIDAsync,
  insertPasswordAsync,
  insertPasswordTableIndexes,
  xorPassword,
  xorXoredPassword,
} from '../../common/sqlite/dao/password';
import { ResponseCode_DataPullAhead } from '../../common/http';
import { downloadData, updateBackupState } from '../../services/backup';
import { UnlockPasswordContext } from '../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../contexts/slark';
import { BackdropContext } from '../../contexts/backdrop';
import { SnackbarContext } from '../../contexts/snackbar';
import { post } from '../../common/http/post';
import { encryptByUnlockPassword } from '../../services/cipher';
import { SlarkInfo } from '../../services/slark';
import { xor_str } from '../../common/sqlite/dao/utils';
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';

export type NewPasswordAction =
  | { type: 'setTitle'; value: Password['title'] }
  | { type: 'setWebsite'; value: Password['website'] }
  | {
      type: 'setUsername';
      value: Password['username'];
    }
  | { type: 'setPassword'; value: Password['password'] }
  | { type: 'setNotes'; value: Password['notes'] }
  | { type: 'setOthers'; value: Password['others'] };

export const initPassword: Password = {
  dataID: '',
  createdAt: 0,
  updatedAt: 0,
  userID: 0,
  title: '',
  username: '',
  password: '',
  website: '',
  notes: '',
  others: null,
  usedAt: null,
  usedCount: null,
  iconBgColor: 0,
  sharedAt: null,
  sharingMembers: null,
  sharedToAll: null,
};

export const passwordReducer = (state: Password, action: NewPasswordAction) => {
  switch (action.type) {
    case 'setTitle':
      return { ...state, title: action.value };
    case 'setWebsite':
      return { ...state, website: action.value };
    case 'setUsername':
      return { ...state, username: action.value };
    case 'setPassword':
      return { ...state, password: action.value };
    case 'setNotes':
      return { ...state, notes: action.value };
    default:
      return { ...state };
  }
};

type NewPasswordStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'NewPasswordStack'
>;

function NewPasswordStackScreen({
  navigation,
  route,
}: NewPasswordStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { setSuccess, setError, setWarning } =
    React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [record, dispatchRecord] = React.useReducer(
    passwordReducer,
    initPassword,
  );
  const [keyboardDidShow, setKeyboardDisShow] = React.useState(false);
  const [editUsername, setEditUsername] = React.useState(false);
  const [frequentlyUsedUsernames, setFrequentlyUsedUsernames] = React.useState<
    string[]
  >([]);

  const findFrequentlyUsedUsernames = React.useCallback(async () => {
    try {
      let userID = slarkInfo ? slarkInfo.userID : -1;
      const curSetting = currentUnlockPasswordSetting(userID);
      const items: string[] = await getFrequentlyUsedUsernamesByUserIDAsync(
        xor_str(curSetting!.passwordHash, userID.toString()),
      );
      const usernames: string[] = [];
      items.map(item => {
        usernames.push(xor_str(curSetting!.passwordHash, item));
      });
      setFrequentlyUsedUsernames(usernames);
    } catch (error) {}
  }, [slarkInfo]);

  useFocusEffect(
    React.useCallback(() => {
      findFrequentlyUsedUsernames();
    }, [findFrequentlyUsedUsernames]),
  );

  Keyboard.addListener('keyboardDidHide', () => {
    setKeyboardDisShow(false);
  });

  Keyboard.addListener('keyboardDidShow', () => {
    setKeyboardDisShow(true);
  });

  const websiteSchema = z
    .string()
    .regex(/^(https:\/\/)?(www\.)?[a-zA-Z0-9-]+\.[a-zA-Z]+(\/[^\s]*)?$/, {
      message: t('passwords.newPassword.toast.invalidWebsite'),
    });

  const toptSchema = z
    .string()
    .regex(
      /^otpauth:\/\/([ht]otp)\/(?:[a-zA-Z0-9%]+:)?([^?]+)\?secret=([0-9A-Za-z]+)(?:.*(?:<?counter=)([0-9]+))?/,
      {
        message: t('passwords.newPassword.toast.invalidTOTP'),
      },
    );
  const [otherFields, setOtherFields] = React.useState<null | OtherField[]>(
    null,
  );
  const genNewEntity = React.useCallback((): Password | null => {
    let userID = slarkInfo ? slarkInfo.userID : -1;
    const nowTS = moment().unix();
    const entity: Password = {
      ...record,
      dataID: uuidv4(),
      createdAt: nowTS,
      updatedAt: nowTS,
      userID: userID,
      others: null,
      usedAt: 0,
      usedCount: 0,
      iconBgColor: Math.round(Math.random() * 4),
    };
    // validate others fields
    if (otherFields) {
      for (let i = 0; i < otherFields.length; i++) {
        if (otherFields[i].type === 'one-time password') {
          // validate ont-time password format, refer to https://github.com/google/google-authenticator/wiki/Key-Uri-Format
          const validation = toptSchema.safeParse(otherFields[i].value);
          if (!validation.success) {
            setError(validation.error.issues[0].message, t('app.toast.error'));
            return null;
          }
        }
      }
      entity.others = JSON.stringify(otherFields);
    }
    return entity;
  }, [slarkInfo, otherFields, record]);

  const createNewPassword = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password: string,
      passwordHash: string,
      entity: Password,
    ) => {
      const updatedAt = entity.updatedAt;
      const xoredEntity = xorPassword(passwordHash, entity);
      setLoading(true);
      const respData = await post('/pswds/createPasswordRecord/v1', {
        updatedAt: updatedAt,
        dataID: entity.dataID,
        content: encryptByUnlockPassword(
          // 整体加密
          password,
          JSON.stringify(xoredEntity),
        ),
      });
      if (respData.code !== 0) {
        if (respData.code === ResponseCode_DataPullAhead) {
          // need to sync data
          const downloadRespData = await downloadData(slarkInfo, password);
          if (downloadRespData.code === 0) {
            // 数据同步成功，警告用户重试操作
            setLoading(false);
            setWarning(
              t('app.toast.afterSyncBackupData'),
              t('app.toast.error'),
            );
          }
          return;
        } else {
          setLoading(false);
          setError(respData.message, t('app.toast.error'));
          return;
        }
      }
      updateBackupState(slarkInfo.userID, {
        updatedAt: updatedAt,
      });
    },
    [],
  );

  const validateParameters = React.useCallback((): boolean => {
    if (!record.title) {
      setError(t('passwords.newPassword.emptyTitle'), t('app.toast.error'));
      return false;
    }
    if (record.website) {
      const validation = websiteSchema.safeParse(record.website);
      if (!validation.success) {
        setError(validation.error.issues[0].message, t('app.toast.error'));
        return false;
      }
      if (!record.website.startsWith('https://')) {
        record.website = 'https://' + record.website;
      }
      if (record.website.endsWith('/')) {
        record.website = record.website.slice(0, record.website.length - 1);
      }
    }
    if (!record.username) {
      setError(t('passwords.newPassword.emptyUsername'), t('app.toast.error'));
      return false;
    }
    if (!record.password) {
      setError(t('passwords.newPassword.emptyPassword'), t('app.toast.error'));
      return false;
    }
    return true;
  }, [record]);

  const newPassword = React.useCallback(async () => {
    // 1. validate parameters
    const valid = validateParameters();
    if (!valid) {
      return;
    }
    try {
      if (!password) {
        throw t('app.toast.emptyUnlockPassword');
      }
      const curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      const newEntity = genNewEntity();
      if (!newEntity) {
        return;
      }
      // 数据同步到后端
      if (slarkInfo) {
        await createNewPassword(
          slarkInfo,
          password,
          curSetting!.passwordHash,
          newEntity,
        );
      }
      const xoredEntity = xorPassword(curSetting!.passwordHash, newEntity);
      await insertPasswordAsync(xoredEntity);
      // insert password indexes
      insertPasswordTableIndexes(curSetting!.passwordHash, xoredEntity);
      setLoading(false);
      setSuccess(t('app.toast.success'));
      navigation.goBack();
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, password, record]);

  React.useEffect(() => {
    if (route.params?.password) {
      dispatchRecord({
        type: 'setPassword',
        value: route.params.password,
      });
    }
  }, [route.params?.password]);

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  const [visible, setVisible] = React.useState(false);

  const onPressOpenRandomPassword = () => {
    navigation.navigate('RandomPasswordStack');
  };

  const onPressAddField = () => {
    setVisible(true);
  };

  const onChangeTitle = (newText: string) => {
    dispatchRecord({
      type: 'setTitle',
      value: newText,
    });
  };

  const onChangeWebsite = (newText: string) => {
    dispatchRecord({
      type: 'setWebsite',
      value: newText,
    });
  };

  const onFocusUsername = () => {
    setEditUsername(true);
  };

  const onBlurUsername = () => {
    setEditUsername(false);
  };

  const onChangeUsername = (newText: string) => {
    dispatchRecord({
      type: 'setUsername',
      value: newText,
    });
  };

  const onChangePassword = (newText: string) => {
    dispatchRecord({
      type: 'setPassword',
      value: newText,
    });
  };

  const onChangeNotes = (newText: string) => {
    dispatchRecord({
      type: 'setNotes',
      value: newText,
    });
  };

  const usernameAccessoryViewID = 'usernameAccessoryViewID';

  return (
    <>
      <SafeAreaView style={styles.container}>
        <KeyboardAvoidingView>
          <ScrollView>
            <Card containerStyle={styles.card}>
              <View style={styles.row}>
                <Input
                  autoCapitalize={'none'}
                  multiline
                  labelStyle={styles.inputLabel}
                  inputStyle={styles.inputStyle}
                  placeholder={t('passwords.newPassword.titlePlaceholder')}
                  onChangeText={onChangeTitle}
                  value={record.title}
                />
              </View>
            </Card>
            <Card containerStyle={styles.card}>
              <View style={styles.row}>
                <Input
                  autoCapitalize={'none'}
                  multiline
                  labelStyle={styles.inputLabel}
                  inputStyle={styles.inputStyle}
                  label={t('passwords.newPassword.website')}
                  placeholder={t('passwords.newPassword.websitePlaceholder')}
                  onChangeText={onChangeWebsite}
                  value={record.website!}
                />
              </View>
            </Card>
            <Card containerStyle={styles.card}>
              <View style={styles.row}>
                <Input
                  inputAccessoryViewID={usernameAccessoryViewID}
                  autoCapitalize={'none'}
                  multiline
                  labelStyle={styles.inputLabel}
                  inputStyle={styles.inputStyle}
                  label={t('passwords.newPassword.username')}
                  placeholder={t('passwords.newPassword.usernamePlaceholder')}
                  onFocus={onFocusUsername}
                  onBlur={onBlurUsername}
                  onChangeText={onChangeUsername}
                  value={record.username}
                />
              </View>
            </Card>
            <Card containerStyle={styles.card}>
              <View style={styles.row}>
                <Input
                  autoCapitalize={'none'}
                  multiline
                  labelStyle={styles.inputLabel}
                  inputStyle={styles.inputStyle}
                  label={
                    <View style={styles.row}>
                      <Text
                        h4
                        style={styles.passwordLabelItem}
                        h4Style={styles.inputLabel}>
                        {t('passwords.newPassword.password')}
                      </Text>
                      <Button
                        title={t('passwords.newPassword.randomPassword')}
                        radius={8}
                        containerStyle={styles.passwordLabelItem}
                        titleStyle={styles.randomPasswordTitle}
                        onPress={onPressOpenRandomPassword}
                      />
                    </View>
                  }
                  placeholder={t('passwords.newPassword.passwordPlaceholder')}
                  onChangeText={onChangePassword}
                  value={record.password}
                />
              </View>
            </Card>
            {otherFields &&
              otherFields.length > 0 &&
              otherFields.map((item, idx) => {
                if (item.type === 'one-time password') {
                  return (
                    <FieldCard
                      key={item.key + idx}
                      readonly={false}
                      isOTP={true}
                      index={idx}
                      fields={otherFields}
                      setFields={setOtherFields}
                    />
                  );
                }
              })}
            <Card containerStyle={styles.card}>
              <View style={styles.row}>
                <Input
                  autoCapitalize={'none'}
                  multiline
                  labelStyle={styles.inputLabel}
                  inputStyle={styles.inputStyle}
                  label={t('passwords.newPassword.notes')}
                  placeholder={t('passwords.newPassword.notesPlaceholder')}
                  numberOfLines={5}
                  onChangeText={onChangeNotes}
                  value={record.notes!}
                />
              </View>
            </Card>
            {otherFields &&
              otherFields.length > 0 &&
              otherFields.map((item, idx) => {
                if (item.type !== 'one-time password') {
                  return (
                    <FieldCard
                      key={item.key + idx}
                      readonly={false}
                      isOTP={false}
                      index={idx}
                      fields={otherFields}
                      setFields={setOtherFields}
                    />
                  );
                }
              })}
            <Card containerStyle={styles.card}>
              <Button
                type="clear"
                title={t('passwords.newPassword.addFieldBtn')}
                containerStyle={styles.normalBtn}
                titleStyle={styles.btnTitle}
                size="lg"
                radius={8}
                onPress={onPressAddField}
              />
            </Card>
            <View style={styles.row}>
              <Button
                title={t('passwords.newPassword.newBtn')}
                containerStyle={styles.commitBtn}
                titleStyle={styles.btnTitle}
                size="lg"
                radius={8}
                onPress={newPassword}
              />
            </View>
            <FieldTypeOverlay
              visible={visible}
              setVisible={setVisible}
              otherFields={otherFields}
              setOtherFields={setOtherFields}
            />
          </ScrollView>
        </KeyboardAvoidingView>
      </SafeAreaView>
      {editUsername &&
        keyboardDidShow &&
        frequentlyUsedUsernames.length > 0 && (
          <InputToolbar
            inputAccessoryViewID={usernameAccessoryViewID}
            list={frequentlyUsedUsernames}
            setValue={val => {
              dispatchRecord({
                type: 'setUsername',
                value: val,
              });
            }}
          />
        )}
    </>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  card: { borderRadius: 8 },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  inputLabel: {
    fontSize: 18,
    fontWeight: 'normal',
    color: theme.colors.black,
    marginBottom: 16,
  },
  inputStyle: {
    fontSize: 18,
    fontWeight: 'normal',
  },
  passwordLabelItem: { width: '50%' },
  randomPasswordTitle: { fontSize: 16 },
  btnTitle: { fontSize: 20, fontWeight: 'normal' },
  normalBtn: { width: '100%' },
  commitBtn: { marginVertical: 10, width: '95%' },
}));

export default NewPasswordStackScreen;
