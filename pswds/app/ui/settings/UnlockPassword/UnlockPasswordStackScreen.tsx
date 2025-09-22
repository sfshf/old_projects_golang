/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, TouchableOpacity } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Text, makeStyles, Icon, useTheme } from '@rneui/themed';
import { useFocusEffect } from '@react-navigation/native';
import { SafeAreaView } from 'react-native';
import { SlarkInfoContext } from '../../../contexts/slark';
import { currentUnlockPasswordSetting } from '../../../services/unlockPassword';

type UnlockPasswordStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function UnlockPasswordStackScreen({
  navigation,
}: UnlockPasswordStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();

  useFocusEffect(
    React.useCallback(() => {
      let userID = -1;
      if (slarkInfo) {
        userID = slarkInfo.userID;
      }
      let curUnlockPasswordSetting = currentUnlockPasswordSetting(userID);
      if (!curUnlockPasswordSetting) {
        navigation.goBack();
        return;
      }
    }, [slarkInfo, navigation]),
  );

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  const onPressVerificationMode = () => {
    navigation.navigate('VerificationModeStack');
  };

  const onPressEditUnlockPassword = () => {
    navigation.navigate('EditUnlockPasswordStack', {
      loginInfo: slarkInfo ? slarkInfo : undefined,
    });
  };

  const onPressAutoLock = () => {
    navigation.navigate('AutoLockStack');
  };

  const onPressSecurityQuestion = () => {
    navigation.navigate('SecurityQuestionStack');
  };

  // 可信联络人 停用
  // const onPressTrustedContact = () => {
  //   navigation.navigate('TrustedContactStack');
  // };

  return (
    <SafeAreaView style={styles.container}>
      <View>
        <TouchableOpacity
          style={[styles.row, styles.wraped]}
          onPress={onPressVerificationMode}>
          <View style={[styles.label]}>
            <Text style={styles.labelText}>
              {t('settings.unlockPassword.verificationMode.label')}
            </Text>
          </View>
          <Icon size={20} name="arrow-forward-ios" color={theme.colors.black} />
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.row, styles.wraped]}
          onPress={onPressEditUnlockPassword}>
          <View style={[styles.label]}>
            <Text style={styles.labelText}>
              {t('settings.unlockPassword.editUnlockPassword.label')}
            </Text>
          </View>
          <Icon size={20} name="arrow-forward-ios" color={theme.colors.black} />
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.row, styles.wraped]}
          onPress={onPressAutoLock}>
          <View style={[styles.label]}>
            <Text style={styles.labelText}>
              {t('settings.unlockPassword.autoLock.label')}
            </Text>
          </View>
          <Icon size={20} name="arrow-forward-ios" color={theme.colors.black} />
        </TouchableOpacity>
        {slarkInfo && (
          <>
            <TouchableOpacity
              style={[styles.row, styles.wraped]}
              onPress={onPressSecurityQuestion}>
              <View style={[styles.label]}>
                <Text style={styles.labelText}>
                  {t('settings.unlockPassword.securityQuestion.label')}
                </Text>
              </View>
              <Icon
                size={20}
                name="arrow-forward-ios"
                color={theme.colors.black}
              />
            </TouchableOpacity>
            {/* 可信联络人 停用 */}
            {/* <TouchableOpacity
              style={[styles.row, styles.wraped]}
              onPress={onPressTrustedContact}>
              <View style={[styles.label]}>
                <Text style={styles.labelText}>
                  {t('settings.unlockPassword.trustedContact.label')}
                </Text>
              </View>
              <Icon
                size={20}
                name="arrow-forward-ios"
                color={theme.colors.black}
              />
            </TouchableOpacity> */}
          </>
        )}
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

export default UnlockPasswordStackScreen;
