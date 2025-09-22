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
import { makeStyles, Text } from '@rneui/themed';
import { View } from 'react-native';
import moment from 'moment';
import { SafeAreaView, ScrollView } from 'react-native';
import { currentDebugHttpLogs, HttpLog } from './DebugHttpLogStackScreen';

type DebugHttpLogDetailStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'DebugHttpLogDetailStack'
>;

function DebugHttpLogDetailStackScreen({
  navigation,
  route,
}: DebugHttpLogDetailStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const [entity, setEntity] = React.useState<null | HttpLog>(null);
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
    const curDebugHttpLogs = currentDebugHttpLogs();
    if (!curDebugHttpLogs) {
      return;
    }
    if (curDebugHttpLogs.logs.length <= route.params.index) {
      return;
    }
    setEntity(curDebugHttpLogs.logs[route.params.index]);
  }, [route.params.index]);

  return (
    <SafeAreaView style={styles.container}>
      {entity && (
        <ScrollView>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.useGo')}
            </Text>
            <Text style={styles.inputContainer}>{entity.useGo + ''}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.start')}
            </Text>
            <Text style={styles.inputContainer}>
              {moment(entity.start).local().format('YYYY-MM-DD HH:mm:ss.SSS')}
            </Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.end')}
            </Text>
            <Text style={styles.inputContainer}>
              {moment(entity.end).local().format('YYYY-MM-DD HH:mm:ss.SSS')}
            </Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.duration')}
            </Text>
            <Text style={styles.inputContainer}>{entity.duration + ''}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.host')}
            </Text>
            <Text style={styles.inputContainer}>{entity.host}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.path')}
            </Text>
            <Text style={styles.inputContainer}>{entity.path}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.reqHeaders')}
            </Text>
            <Text style={styles.inputContainer}>
              {entity.reqHeaders
                ? JSON.stringify(entity.reqHeaders, null, 2)
                : ''}
            </Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.reqBody')}
            </Text>
            <Text style={styles.inputContainer}>
              {entity.reqBody ? JSON.stringify(entity.reqBody, null, 2) : ''}
            </Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.respCode')}
            </Text>
            <Text style={styles.inputContainer}>{entity.respCode + ''}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.respHeaders')}
            </Text>
            <Text style={styles.inputContainer}>
              {entity.respHeaders
                ? JSON.stringify(entity.respHeaders.map, null, 2)
                : ''}
            </Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.inputLabel}>
              {t('debug.httpLogDetail.fields.respBody')}
            </Text>
            <Text style={styles.inputContainer}>
              {entity.respBody ? JSON.stringify(entity.respBody, null, 2) : ''}
            </Text>
          </View>
        </ScrollView>
      )}
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  row: {
    flex: 1,
    margin: 10,
  },
  inputLabel: {
    fontSize: 18,
    fontWeight: 'normal',
    color: theme.colors.black,
    marginBottom: 16,
  },
  inputContainer: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderRadius: 8,
    borderColor: theme.colors.black,
    color: theme.colors.black,
    minHeight: 50,
    fontSize: 18,
    padding: 8,
  },
}));

export default DebugHttpLogDetailStackScreen;
