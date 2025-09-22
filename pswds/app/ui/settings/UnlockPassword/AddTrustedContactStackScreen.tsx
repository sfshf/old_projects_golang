/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RootStackParamList } from '../../../navigation/routes';
import { useFocusEffect } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import { makeStyles, Input, Button, Icon, Text } from '@rneui/themed';
import { View } from 'react-native';
import { randomPassword } from '../../../common/cipher';
import Clipboard from '@react-native-clipboard/clipboard';
import { SafeAreaView, ScrollView, KeyboardAvoidingView } from 'react-native';
import { SlarkInfoContext } from '../../../contexts/slark';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { currentUnlockPasswordSetting } from '../../../services/unlockPassword';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';
import { encryptByUnlockPassword } from '../../../services/cipher';

interface TrustedContact {
  contactEmail: string;
  backupPassword: string;
  showPassword: boolean;
}

type UpdateDataAction =
  | {
      type: 'updateContactEmail';
      value: TrustedContact['contactEmail'];
    }
  | { type: 'updateBackupPassword'; value: TrustedContact['backupPassword'] }
  | { type: 'setShowPassword'; value: TrustedContact['showPassword'] };

type AddTrustedContactStackScreenProp =
  NativeStackScreenProps<RootStackParamList>;

function AddTrustedContactStackScreen({
  navigation,
}: AddTrustedContactStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const initTrustedContact: TrustedContact = {
    contactEmail: '',
    backupPassword: '',
    showPassword: true,
  };
  const trustedContactReducer = (
    entity: TrustedContact,
    action: UpdateDataAction,
  ) => {
    switch (action.type) {
      case 'updateContactEmail':
        return { ...entity, contactEmail: action.value };
      case 'updateBackupPassword':
        return { ...entity, backupPassword: action.value };
      case 'setShowPassword':
        return { ...entity, showPassword: action.value };
      default:
        return entity;
    }
  };
  const [trustedContact, dispatchTrustedContact] = React.useReducer(
    trustedContactReducer,
    initTrustedContact,
  );
  const { password } = React.useContext(UnlockPasswordContext);
  const validateParameters = React.useCallback((): boolean => {
    if (!slarkInfo) {
      return false;
    }
    if (!trustedContact.contactEmail) {
      setError(
        t('settings.unlockPassword.addTrustedContact.toast.emptyContactEmail'),
        t('app.toast.error'),
      );
      return false;
    }
    if (!trustedContact.backupPassword) {
      setError(
        t(
          'settings.unlockPassword.addTrustedContact.toast.emptyBackupPassword',
        ),
        t('app.toast.error'),
      );
      return false;
    }
    if (!password) {
      setError(
        t('app.toast.emptyUnlockPassword'),
        t('app.toast.internalError'),
      );
      return false;
    }
    return true;
  }, [slarkInfo, password, trustedContact]);

  const createContact = async () => {
    try {
      const valid = validateParameters();
      if (!valid) {
        return;
      }
      const backupCiphertext = encryptByUnlockPassword(
        trustedContact.backupPassword,
        password,
      );
      // 数据同步到后端
      setLoading(true);
      const respData = await post('/pswds/createTrustedContact/v1', {
        contactEmail: trustedContact.contactEmail,
        backupCiphertext,
      });
      setLoading(false);
      if (respData.code !== 0) {
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      setSuccess(t('app.toast.success'));
      navigation.goBack();
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  };

  useFocusEffect(
    React.useCallback(() => {
      if (!slarkInfo) {
        navigation.goBack();
        return;
      }
      let curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      if (!curSetting) {
        navigation.navigate('UnlockPasswordStack');
        return;
      }
      // generate a backup password
      const backupPassword = randomPassword({
        length: 20,
        useNumbers: true,
        useSymbols: true,
      });
      dispatchTrustedContact({
        type: 'updateBackupPassword',
        value: backupPassword,
      });
    }, [slarkInfo, navigation]),
  );

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  const onPressRefreshPassword = () => {
    const backupPassword = randomPassword({
      length: 20,
      useNumbers: true,
      useSymbols: true,
    });
    dispatchTrustedContact({
      type: 'updateBackupPassword',
      value: backupPassword,
    });
  };

  const onPressCopy = () => {
    Clipboard.setString(trustedContact.backupPassword);
  };

  const onPressShowPassword = () => {
    dispatchTrustedContact({
      type: 'setShowPassword',
      value: !trustedContact.showPassword,
    });
  };

  const onChangeContactEmail = (newText: string) => {
    newText = newText.trim();
    dispatchTrustedContact({
      type: 'updateContactEmail',
      value: newText,
    });
  };

  const onChangeBackupPassword = (newText: string) => {
    newText = newText.trim();
    dispatchTrustedContact({
      type: 'updateBackupPassword',
      value: newText,
    });
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView>
        <ScrollView>
          <View style={styles.body}>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                style={styles.input}
                labelStyle={styles.inputLabel}
                inputStyle={styles.inputStyle}
                label={
                  <View style={styles.row}>
                    <Text
                      h4
                      style={styles.passwordLabel}
                      h4Style={styles.inputLabel}>
                      {t(
                        'settings.unlockPassword.addTrustedContact.contactEmail',
                      )}
                    </Text>
                  </View>
                }
                placeholder={t(
                  'settings.unlockPassword.addTrustedContact.contactEmailPlaceholder',
                )}
                onChangeText={onChangeContactEmail}
                value={trustedContact.contactEmail}
              />
            </View>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                style={styles.input}
                labelStyle={styles.inputLabel}
                inputStyle={styles.inputStyle}
                label={
                  <>
                    <View style={styles.row}>
                      <Text
                        h4
                        style={styles.passwordLabel}
                        h4Style={styles.inputLabel}>
                        {t(
                          'settings.unlockPassword.addTrustedContact.backupPassword',
                        )}
                      </Text>
                    </View>
                    <View
                      style={[
                        styles.row,
                        { paddingLeft: '5%', paddingRight: '50%' },
                      ]}>
                      <Button
                        title={t(
                          'settings.unlockPassword.addTrustedContact.refreshBtn',
                        )}
                        size="sm"
                        radius={8}
                        containerStyle={styles.passwordLabelBtn}
                        titleStyle={styles.btnTitle}
                        onPress={onPressRefreshPassword}
                      />
                      <Button
                        title={t(
                          'settings.unlockPassword.addTrustedContact.copyBtn',
                        )}
                        size="sm"
                        radius={8}
                        containerStyle={styles.passwordLabelBtn}
                        titleStyle={styles.btnTitle}
                        onPress={onPressCopy}
                      />
                    </View>
                  </>
                }
                placeholder={t(
                  'settings.unlockPassword.addTrustedContact.backupPasswordPlaceholder',
                )}
                secureTextEntry={!trustedContact.showPassword}
                onChangeText={onChangeBackupPassword}
                value={trustedContact.backupPassword}
                rightIcon={
                  <Icon
                    type="font-awesome-5"
                    name={trustedContact.showPassword ? 'eye' : 'eye-slash'}
                    size={18}
                    onPress={onPressShowPassword}
                  />
                }
              />
            </View>
            <View style={styles.row}>
              <Button
                title={t('settings.unlockPassword.addTrustedContact.addBtn')}
                containerStyle={styles.commitBtn}
                titleStyle={styles.btnTitle}
                size="lg"
                radius={8}
                onPress={createContact}
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
  input: {
    width: '100%',
    color: theme.colors.black,
    paddingHorizontal: 12,
  },
  passwordLabel: { width: '100%' },
  passwordLabelBtn: { width: '50%', marginHorizontal: 8 },
  btnTitle: { fontSize: 20 },
  commitBtn: { width: '100%' },
}));

export default AddTrustedContactStackScreen;
