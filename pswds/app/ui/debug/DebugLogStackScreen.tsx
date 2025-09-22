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
import { SafeAreaView } from 'react-native';
import moment from 'moment';
import { getLogs, Log } from '../../common/log';
import { useTheme } from '@rneui/themed';

type DebugLogStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function DebugLogStackScreen({
  navigation,
}: DebugLogStackScreenProp): React.JSX.Element {
  const styles = useStyles();
  const { theme } = useTheme();
  const [list, setList] = React.useState<null | Log[]>(null);
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
    const debugLogs = getLogs();
    if (debugLogs) {
      setList(debugLogs);
    }
  }, []);

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView>
        {list &&
          list.map((item, index) => {
            return (
              <ListItem key={index} bottomDivider>
                <ListItem.Content>
                  <ListItem.Title>
                    <View style={styles.itemTitle}>
                      <Text style={styles.itemTitleTitle}>
                        {moment(item.timestamp)
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss.SSS')}
                      </Text>
                    </View>
                  </ListItem.Title>
                  <ListItem.Subtitle>
                    <View style={styles.itemTitle}>
                      <Text
                        style={[
                          styles.itemTitleTitle,
                          item.level === 'info'
                            ? {}
                            : { color: theme.colors.error },
                        ]}>
                        {item.message}
                      </Text>
                    </View>
                  </ListItem.Subtitle>
                </ListItem.Content>
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
  itemTitleTitle: { fontSize: 10 },
  itemSubTitle: { fontSize: 8 },
}));

export default DebugLogStackScreen;
