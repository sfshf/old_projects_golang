/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { TouchableOpacity, View } from 'react-native';
import { RootStackParamList } from '../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Icon, makeStyles, Switch, Text, useTheme } from '@rneui/themed';
import { SafeAreaView } from 'react-native';
import { storage } from '../../common/mmkv';
import { SlarkInfoContext } from '../../contexts/slark';
import { currentDebugUseGo, updateDebugUseGo } from '../../common/http';

export interface LogHttpCache {
  logHttp: boolean;
}
export const debugLogHttpKey = 'cache_debug.logHttp';
export const currentDebugLogHttp = (): null | LogHttpCache => {
  const curSetting = storage.getString(debugLogHttpKey);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  let cfg = null;
  if (__DEV__) {
    cfg = { logHttp: true };
  } else {
    cfg = { logHttp: false };
  }
  updateDebugLogHttp(cfg);
  return cfg;
};
export const updateDebugLogHttp = (obj: null | LogHttpCache) => {
  if (!obj) {
    storage.delete(debugLogHttpKey);
    return;
  }
  storage.set(debugLogHttpKey, JSON.stringify(obj));
};

type DebugStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function DebugStackScreen({
  navigation,
}: DebugStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const [useGo, setUseGo] = React.useState<boolean>(true);
  const [logHttp, setLogHttp] = React.useState<boolean>(true);

  React.useEffect(() => {
    // use go
    let curUseGoSetting = currentDebugUseGo();
    if (curUseGoSetting) {
      setUseGo(curUseGoSetting.useGo);
    } else {
      updateDebugUseGo({ useGo: true });
      setUseGo(true);
    }
    // log http
    let curLogHttpSetting = currentDebugLogHttp();
    if (curLogHttpSetting) {
      setLogHttp(curLogHttpSetting.logHttp);
    } else {
      let logHttp = true;
      if (!__DEV__) {
        logHttp = false;
      }
      updateDebugLogHttp({ logHttp });
      setLogHttp(logHttp);
    }
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  const onValueChangeUseGo = (value: boolean) => {
    setUseGo(value);
    let userID = -1;
    if (slarkInfo) {
      userID = slarkInfo.userID;
    }
    updateDebugUseGo({ useGo: value });
  };

  const onValueChangeLogHttp = (value: boolean) => {
    setLogHttp(value);
    let userID = -1;
    if (slarkInfo) {
      userID = slarkInfo.userID;
    }
    updateDebugLogHttp({ logHttp: value });
  };

  const onPressHttpLogs = () => {
    navigation.navigate('DebugHttpLogStack');
  };

  const onPressLogs = () => {
    navigation.navigate('DebugLogStack');
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.row}>
        <Text style={styles.labelText}>{t('debug.useGo')}</Text>
        <Switch value={useGo} onValueChange={onValueChangeUseGo} />
      </View>
      <View style={styles.row}>
        <Text style={styles.labelText}>{t('debug.logHttp')}</Text>
        <Switch value={logHttp} onValueChange={onValueChangeLogHttp} />
      </View>
      <TouchableOpacity style={styles.row} onPress={onPressHttpLogs}>
        <View style={styles.label}>
          <Text style={styles.labelText}>{t('debug.httpLogs.label')}</Text>
        </View>
        <Icon size={20} name="arrow-forward-ios" color={theme.colors.black} />
      </TouchableOpacity>
      <TouchableOpacity style={styles.row} onPress={onPressLogs}>
        <View style={styles.label}>
          <Text style={styles.labelText}>{t('debug.logs.label')}</Text>
        </View>
        <Icon size={20} name="arrow-forward-ios" color={theme.colors.black} />
      </TouchableOpacity>
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
    marginVertical: 4,
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  labelText: {
    fontSize: 16,
  },
  wraped: { flexWrap: 'wrap' },
  label: {
    flex: 2,
  },
}));

export default DebugStackScreen;
