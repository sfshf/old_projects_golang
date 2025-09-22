/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { ScrollView, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RootStackParamList } from '../../navigation/routes';
import { makeStyles, ListItem, Text } from '@rneui/themed';
import { useTranslation } from 'react-i18next';
import { SafeAreaView } from 'react-native';
import { storage } from '../../common/mmkv';
import moment from 'moment';
import { SlarkInfoContext } from '../../contexts/slark';

export type HttpLog = {
  useGo: boolean; //是否是加密请求
  start: number; // 开始时间
  end: number; // 结束时间
  duration: number; // 耗时
  host: string; //请求host
  path: string; // 请求path
  cookie: string; // cookie
  reqBody: any; // 请求体
  reqHeaders: any; // 请求头
  respCode: number; // 响应码
  respHeaders: any; // 响应头
  respBody: any; // 响应体
};
export interface HttpLogsCache {
  logs: HttpLog[];
}
export const debugHttpLogsKey = 'cache_debug.httpLogs';
export const currentDebugHttpLogs = (): null | HttpLogsCache => {
  const curSetting = storage.getString(debugHttpLogsKey);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};
export const updateDebugHttpLogs = (obj: null | HttpLogsCache) => {
  if (!obj) {
    storage.delete(debugHttpLogsKey);
    return;
  }
  storage.set(debugHttpLogsKey, JSON.stringify(obj));
};

type DebugHttpLogStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function DebugHttpLogStackScreen({
  navigation,
}: DebugHttpLogStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const [list, setList] = React.useState<null | HttpLog[]>(null);
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
    const curDebugHttpLogs = currentDebugHttpLogs();
    if (curDebugHttpLogs) {
      setList(curDebugHttpLogs.logs.reverse());
    }
  }, []);

  const onPressListItem = (index: number) => {
    return () => {
      navigation.navigate('DebugHttpLogDetailStack', { index });
    };
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView>
        {list &&
          list.map((item, index) => {
            return (
              <ListItem
                key={index}
                bottomDivider
                onPress={onPressListItem(list.length - index - 1)}>
                <ListItem.Content>
                  <ListItem.Title>
                    <View style={styles.itemTitle}>
                      <Text style={styles.itemTitleTitle}>
                        {moment(item.start)
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss.SSS')}
                      </Text>
                    </View>
                  </ListItem.Title>
                  <ListItem.Subtitle>{item.path}</ListItem.Subtitle>
                </ListItem.Content>
                <ListItem.Chevron size={40} />
              </ListItem>
            );
          })}
      </ScrollView>
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
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  itemTitle: {
    flexDirection: 'row',
    alignContent: 'space-between',
  },
  itemTitleTitle: { fontSize: 20 },
}));

export default DebugHttpLogStackScreen;
