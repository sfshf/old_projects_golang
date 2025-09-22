/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Text, makeStyles } from '@rneui/themed';
import { useFocusEffect } from '@react-navigation/native';
import PickerView from '../../../components/PickerView';
import { SafeAreaView } from 'react-native';
import { SlarkInfoContext } from '../../../contexts/slark';
import {
  currentAutoLockSetting,
  currentUnlockPasswordSetting,
  updateAutoLockSetting,
} from '../../../services/unlockPassword';

export enum AutoLockTime {
  Immediately,
  After30Seconds,
  After1Minute,
  After10Minutes,
  After30Minutes,
  After1Hour,
  Never,
}

export const AutoLockTimeValueMap = {
  [AutoLockTime.Immediately]: 0,
  [AutoLockTime.After30Seconds]: 30,
  [AutoLockTime.After1Minute]: 60,
  [AutoLockTime.After10Minutes]: 600,
  [AutoLockTime.After30Minutes]: 1800,
  [AutoLockTime.After1Hour]: 3600,
  [AutoLockTime.Never]: 999999999,
};

export const DEFAULT_AUTO_LOCK_TIME = AutoLockTime.After30Seconds;

type AutoLockStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function AutoLockStackScreen({
  navigation,
}: AutoLockStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { slarkInfo } = React.useContext(SlarkInfoContext);

  const autoLockTimeTexts = [
    t('settings.unlockPassword.autoLock.lockTimes.immediately'),
    t('settings.unlockPassword.autoLock.lockTimes.after30Seconds'),
    t('settings.unlockPassword.autoLock.lockTimes.after1Minute'),
    t('settings.unlockPassword.autoLock.lockTimes.after10Minutes'),
    t('settings.unlockPassword.autoLock.lockTimes.after30Minutes'),
    t('settings.unlockPassword.autoLock.lockTimes.after1Hour'),
    t('settings.unlockPassword.autoLock.lockTimes.never'),
  ];

  const selectedIndex = (): number => {
    let userID = -1;
    if (slarkInfo) {
      userID = slarkInfo.userID;
    }
    const curAutoLockSetting = currentAutoLockSetting(userID);
    if (!curAutoLockSetting) {
      return DEFAULT_AUTO_LOCK_TIME;
    }
    let res = DEFAULT_AUTO_LOCK_TIME;
    Object.entries(AutoLockTimeValueMap).forEach(([key, value]) => {
      if (value === curAutoLockSetting.timeLag) {
        res = parseInt(key, 10) as AutoLockTime;
      }
    });
    return res;
  };

  const onSelect = (index: AutoLockTime) => {
    let userID = -1;
    if (slarkInfo) {
      userID = slarkInfo.userID;
    }
    updateAutoLockSetting(userID, {
      timeLag: AutoLockTimeValueMap[index],
    });
  };

  useFocusEffect(
    React.useCallback(() => {
      let userID = -1;
      if (slarkInfo) {
        userID = slarkInfo.userID;
      }
      let curSetting = currentUnlockPasswordSetting(userID);
      if (!curSetting) {
        navigation.navigate('UnlockPasswordStack');
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

  return (
    <SafeAreaView style={styles.container}>
      <View>
        <Text style={styles.normalText}>
          {t('settings.unlockPassword.autoLock.caption')}
        </Text>
        <PickerView
          options={autoLockTimeTexts}
          onSelect={onSelect}
          selectedIndex={selectedIndex()}
        />
      </View>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
    marginHorizontal: 8,
  },
  normalText: {
    marginTop: 20,
    fontSize: 16,
    textAlign: 'center',
    width: '100%',
    marginBottom: 40,
    color: theme.colors.black,
  },
}));

export default AutoLockStackScreen;
