/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, Keyboard, Alert, Text } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { makeStyles, Button, Input, useTheme } from '@rneui/themed';
import { SafeAreaView, ScrollView, KeyboardAvoidingView } from 'react-native';
import { HeaderBackButton } from '@react-navigation/elements';
import {
  ResponseCode_NotSetSecurityQuestions,
  ResponseCode_ResourceLimit,
} from '../../../common/http';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';
import RequestFamilyRecoverOverlay, {
  BackupMember,
} from '../../../components/RequestFamilyRecoverOverlay';
import { currentSlarkInfo } from '../../../services/slark';
import { useFocusEffect } from '@react-navigation/native';
import moment from 'moment';

type RecoverUnlockPasswordStackScreenProp =
  NativeStackScreenProps<RootStackParamList>;

function RecoverUnlockPasswordStackScreen({
  navigation,
}: RecoverUnlockPasswordStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { setLoading } = React.useContext(BackdropContext);
  const { setError } = React.useContext(SnackbarContext);
  const { password, setVisible } = React.useContext(UnlockPasswordContext);
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const [email, setEmail] = React.useState<string>('');
  const [checked, setChecked] = React.useState(slarkInfo !== null);

  const handleFailedResponse = (respData: any) => {
    if (respData.code === ResponseCode_NotSetSecurityQuestions) {
      Alert.alert(
        t('app.alert.mustSetSecurityQuestionsTitle'),
        t('app.alert.mustSetSecurityQuestionsMessage'),
        [
          {
            text: t('app.alert.okBtn'),
            style: 'destructive',
          },
        ],
      );
    } else if (respData.code === ResponseCode_ResourceLimit) {
      Alert.alert(
        t('app.alert.recoverRequestLimitTitle'),
        t('app.alert.recoverRequestLimitMessage'),
        [
          {
            text: t('app.alert.okBtn'),
            style: 'destructive',
          },
        ],
      );
    } else {
      setError(respData.message, t('app.toast.requestError'));
    }
  };
  const [backupMembers, setBackupMembers] = React.useState<BackupMember[]>([]);
  const [recoverVisible, setRecoverVisible] = React.useState(false);
  const doRecover = React.useCallback(async () => {
    Keyboard.dismiss();
    try {
      setLoading(true);
      const respData = await post('/pswds/recoverUnlockPassword/v1', {
        email,
      });
      setLoading(false);
      if (respData.code !== 0) {
        handleFailedResponse(respData);
        return;
      }
      setCanRecover(false);
      Alert.alert(
        t('app.alert.emailHasSentTitle'),
        t('app.alert.emailHasSentMessage') + email,
        [
          {
            text: t('app.alert.okBtn'),
            onPress: () => {},
          },
        ],
      );
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, password, email]);
  const doFamilyRecover = React.useCallback(async () => {
    Keyboard.dismiss();
    setRecoverVisible(true);
  }, []);

  const onChangeEmail = (newText: string) => {
    newText = newText.trim();
    setEmail(newText);
  };

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerLeft: () => (
        <HeaderBackButton
          style={styles.headerBack}
          onPress={() => {
            setEmail('');
            navigation.goBack();
            // 已登录，从锁屏跳转过来
            if (!password && currentSlarkInfo() !== null) {
              setVisible(true);
            }
          }}
        />
      ),
    });
  }, [navigation, password]);
  const [canRecover, setCanRecover] = React.useState(false);
  const [lastRecoveredAt, setLastRecoveredAt] = React.useState(0);
  const [nextRecoverTS, setNextRecoverTS] = React.useState(0);
  const [canFamilyRecover, setCanFamilyRecover] = React.useState(false);
  const [lastFamilyRecoveredAt, setLastFamilyRecoveredAt] = React.useState(0);
  const [nextFamilyRecoverTS, setNextFamilyRecoverTS] = React.useState(0);
  const checkUnlockPasswordBackups = React.useCallback(
    async (email: string) => {
      setLoading(true);
      const respData = await post('/pswds/checkUnlockPasswordBackups/v1', {
        email,
      });
      if (respData.code !== 0) {
        setLoading(false);
        handleFailedResponse(respData);
        return;
      }
      setLoading(false);
      // 如果没有任何备份
      if (respData.data.nullBackup) {
        setCanRecover(false);
        setCanFamilyRecover(false);
        Alert.alert(
          t('app.alert.nullBackupTitle'),
          t('app.alert.nullBackupMessage'),
          [
            {
              text: t('app.alert.okBtn'),
              style: 'cancel',
            },
          ],
        );
        return;
      }
      setCanRecover(respData.data.canRecover);
      setLastRecoveredAt(respData.data.lastRecoveredAt);
      setNextRecoverTS(respData.data.nextRecoverTS);
      setCanFamilyRecover(respData.data.canFamilyRecover);
      setBackupMembers(respData.data.backupMembers);
      setLastFamilyRecoveredAt(respData.data.lastFamilyRecoveredAt);
      setNextFamilyRecoverTS(respData.data.nextFamilyRecoverTS);
      setChecked(true);
    },
    [slarkInfo],
  );
  useFocusEffect(
    React.useCallback(() => {
      if (slarkInfo) {
        // 已登录，从锁屏跳转过来
        setEmail(slarkInfo.email);
        checkUnlockPasswordBackups(slarkInfo.email);
      } else {
        setEmail('');
      }
    }, []),
  );
  const doCheck = React.useCallback(async () => {
    Keyboard.dismiss();
    if (!email) {
      return;
    }
    await checkUnlockPasswordBackups(email);
  }, [slarkInfo, password, email]);

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView>
        <ScrollView>
          <View style={styles.body}>
            <View style={styles.row}>
              <Input
                editable={slarkInfo === null}
                autoCapitalize={'none'}
                labelStyle={styles.inputLabel}
                style={styles.input}
                placeholder={t(
                  'settings.unlockPassword.recoverUnlockPassword.emailPlaceholder',
                )}
                onChangeText={onChangeEmail}
                value={email}
              />
            </View>
            {checked && (
              <View style={[styles.row, { justifyContent: 'space-between' }]}>
                <View style={styles.half}>
                  {canRecover && (
                    <Button
                      title={t(
                        'settings.unlockPassword.recoverUnlockPassword.doRecoverBtn',
                      )}
                      containerStyle={styles.btn}
                      titleStyle={styles.btnTitle}
                      size="lg"
                      radius={8}
                      onPress={doRecover}
                    />
                  )}
                  {slarkInfo !== null && !canRecover && <></>}
                </View>
                <View style={styles.half}>
                  {canFamilyRecover && (
                    <Button
                      title={t(
                        'settings.unlockPassword.recoverUnlockPassword.doFamilyRecoverBtn',
                      )}
                      containerStyle={styles.btn}
                      titleStyle={styles.btnTitle}
                      size="lg"
                      radius={8}
                      onPress={doFamilyRecover}
                    />
                  )}
                  {slarkInfo !== null && !canFamilyRecover && <></>}
                </View>
              </View>
            )}
            {!checked && (
              <View style={[styles.row, { justifyContent: 'space-between' }]}>
                <Button
                  title={t(
                    'settings.unlockPassword.recoverUnlockPassword.doCheckBtn',
                  )}
                  disabled={!email}
                  containerStyle={styles.btn}
                  titleStyle={styles.btnTitle}
                  size="lg"
                  radius={8}
                  onPress={doCheck}
                />
              </View>
            )}
            {lastRecoveredAt > 0 && nextRecoverTS > 0 && (
              <View style={styles.row}>
                <Text style={styles.warningTip}>
                  {t(
                    'settings.unlockPassword.recoverUnlockPassword.recoverTip',
                    {
                      lastRecoveredAt: moment(lastRecoveredAt * 1000)
                        .local()
                        .format('YYYY-MM-DD HH:mm:ss'),
                      nextRecoverTS: moment(nextRecoverTS * 1000)
                        .local()
                        .format('YYYY-MM-DD HH:mm:ss'),
                    },
                  )}
                </Text>
              </View>
            )}
            {lastFamilyRecoveredAt > 0 && nextFamilyRecoverTS > 0 && (
              <View style={styles.row}>
                <Text style={styles.warningTip}>
                  {t(
                    'settings.unlockPassword.recoverUnlockPassword.familyRecoverTip',
                    {
                      lastFamilyRecoveredAt: moment(
                        lastFamilyRecoveredAt * 1000,
                      )
                        .local()
                        .format('YYYY-MM-DD HH:mm:ss'),
                      nextFamilyRecoverTS: moment(nextFamilyRecoverTS * 1000)
                        .local()
                        .format('YYYY-MM-DD HH:mm:ss'),
                    },
                  )}
                </Text>
              </View>
            )}
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
      <RequestFamilyRecoverOverlay
        visible={recoverVisible}
        setVisible={setRecoverVisible}
        email={email}
        backupMembers={backupMembers}
        setCanRecover={setCanRecover}
        setCanFamilyRecover={setCanFamilyRecover}
      />
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  headerBack: {
    marginLeft: -15,
    padding: 0,
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
  input: { height: 60 },
  btn: { width: '100%' },
  btnTitle: { fontSize: 12 },
  inputLabel: {
    fontSize: 20,
    fontWeight: 400,
    marginBottom: 20,
    color: theme.colors.black,
  },
  half: { width: '50%', padding: 1 },
  warningTip: {
    color: theme.colors.warning,
    textAlign: 'center',
  },
}));

export default RecoverUnlockPasswordStackScreen;
