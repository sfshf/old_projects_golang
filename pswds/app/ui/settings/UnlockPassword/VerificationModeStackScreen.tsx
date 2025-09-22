/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { Platform, View } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Text, makeStyles, useTheme, Switch } from '@rneui/themed';
import { useFocusEffect } from '@react-navigation/native';
import * as LocalAuthentication from '../../../services/localAuthentication';
import { SafeAreaView } from 'react-native';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import {
  currentUnlockPasswordSetting,
  UnlockPasswordSetting,
  updateUnlockPasswordSetting,
} from '../../../services/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';

type VerificationModeStackScreenProp =
  NativeStackScreenProps<RootStackParamList>;

function VerificationModeStackScreen({
  navigation,
}: VerificationModeStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { supportFingerprint, setSupportFingerprint } = React.useContext(
    UnlockPasswordContext,
  );
  const { setLoading } = React.useContext(BackdropContext);
  const { setError } = React.useContext(SnackbarContext);
  const { password, setPassword } = React.useContext(UnlockPasswordContext);

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  useFocusEffect(
    React.useCallback(() => {
      let userID = slarkInfo ? slarkInfo.userID : -1;
      let curUnlockPasswordSetting = currentUnlockPasswordSetting(userID);
      if (!curUnlockPasswordSetting) {
        navigation.navigate('EditUnlockPasswordStack', {
          loginInfo: slarkInfo ? slarkInfo : undefined,
        });
        return;
      }
    }, [slarkInfo, navigation]),
  );

  const handleNotSupport = React.useCallback(
    async (
      userID: number,
      support: boolean,
      curSetting: UnlockPasswordSetting,
    ) => {
      try {
        const result = await LocalAuthentication.hasFingerprintPassword(
          userID.toString(),
        );
        if (!result) {
          setSupportFingerprint(!support);
          return;
        }
        // 情形二：禁用指纹验证，将解锁密码缓存，并删除解锁密码存储
        const passwd = await LocalAuthentication.getGenericPassword(
          userID.toString(),
        );
        if (passwd) {
          setPassword(passwd);
          await LocalAuthentication.deleteFingerprintPassword(
            userID.toString(),
          );
          updateUnlockPasswordSetting(userID, {
            ...curSetting,
            supportFingerprint: false,
          });
        } else {
          setSupportFingerprint(!support);
        }
      } catch (error) {
        throw error;
      }
    },
    [slarkInfo],
  );

  const handleSupport = React.useCallback(
    async (
      userID: number,
      support: boolean,
      oldSetting: UnlockPasswordSetting,
    ) => {
      try {
        if (Platform.OS === 'ios') {
          const result = await LocalAuthentication.requestFaceIDPermission();
          if (!result.granted) {
            setSupportFingerprint(!support);
            return;
          }
        }
        const result = await LocalAuthentication.setFingerprintPassword(
          userID.toString(),
          password,
        );
        if (!result) {
          setSupportFingerprint(!support);
          return;
        }
        const success = await LocalAuthentication.getGenericPassword(
          userID.toString(),
        );
        if (success) {
          updateUnlockPasswordSetting(userID, {
            ...oldSetting,
            supportFingerprint: true,
          });
          // 删除解锁密码缓存
          setPassword('');
        }
        const curSetting = currentUnlockPasswordSetting(userID);
        if (!curSetting || !curSetting.supportFingerprint) {
          setSupportFingerprint(!support);
          LocalAuthentication.deleteFingerprintPassword(userID.toString());
        }
      } catch (error) {
        throw error;
      }
    },
    [],
  );

  const authenticateFingerprint = React.useCallback(
    async (support: boolean) => {
      try {
        let userID = slarkInfo ? slarkInfo.userID : -1;
        let curSetting = currentUnlockPasswordSetting(userID);
        if (!curSetting) {
          return;
        }
        setSupportFingerprint(support);
        if (!support) {
          await handleNotSupport(userID, support, curSetting);
          return;
        } else {
          // 情形一：启用指纹验证，将解锁密码存储
          await handleSupport(userID, support, curSetting);
        }
      } catch (error) {
        setError(error as string, t('app.toast.internalError'));
      }
    },
    [slarkInfo],
  );

  return (
    <SafeAreaView style={styles.container}>
      <View style={[styles.row, styles.wraped]}>
        <View style={[styles.label]}>
          <Text style={styles.labelText}>
            {t('settings.unlockPassword.verificationMode.supportFingerprint')}
          </Text>
        </View>
        <Switch
          trackColor={{
            false: theme.colors.grey3,
            true: theme.colors.primary,
          }}
          value={supportFingerprint}
          onValueChange={value => {
            authenticateFingerprint(value);
          }}
        />
      </View>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(() => ({
  container: {
    flex: 1,
  },
  row: {
    flexDirection: 'row',
    marginHorizontal: 8,
    paddingHorizontal: 8,
    padding: 10,
  },
  label: {
    flex: 2,
  },
  labelText: {
    fontSize: 16,
    fontWeight: 500,
  },
  wraped: { flexWrap: 'wrap' },
}));

export default VerificationModeStackScreen;
