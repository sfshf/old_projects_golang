/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { makeStyles, Card, Text } from '@rneui/themed';
import { View, ScrollView, KeyboardAvoidingView } from 'react-native';
import { SafeAreaView } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import moment from 'moment';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';

interface Message {
  id: number;
  createdAt: number;
  creator: string;
  target: string;
  operation: number;
}

type FamilyMessageStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'FamilyMessageStack'
>;

function FamilyMessageStackScreen({
  navigation,
  route,
}: FamilyMessageStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [messages, setMessages] = React.useState<Message[]>([]);

  const onLoad = React.useCallback(async () => {
    try {
      setLoading(true);
      let respData = await post('/pswds/getFamilyMessages/v1');
      setLoading(false);
      if (respData.code !== 0) {
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      if (respData.data) {
        setMessages(respData.data.list);
      }
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  }, []);

  React.useEffect(() => {
    onLoad();
  }, [onLoad]);

  React.useEffect(() => {
    // 检查用户登录状态
    if (!slarkInfo) {
      navigation.goBack();
      return;
    }
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  const operations = (operation: number): string => {
    switch (operation) {
      case 1:
        return t('settings.family.familyMessage.operations.1');
      case 2:
        return t('settings.family.familyMessage.operations.2');
      case 3:
        return t('settings.family.familyMessage.operations.3');
      case 4:
        return t('settings.family.familyMessage.operations.4');
      case 5:
        return t('settings.family.familyMessage.operations.5');
      case 6:
        return t('settings.family.familyMessage.operations.6');
      case 7:
        return t('settings.family.familyMessage.operations.7');
      case 8:
        return t('settings.family.familyMessage.operations.8');
      default:
        return '';
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView>
        <ScrollView>
          {messages.map(item => (
            <Card key={item.id} containerStyle={styles.card}>
              <View style={styles.row}>
                <View style={styles.label}>
                  <Text>{t('settings.family.familyMessage.createdAt')}</Text>
                </View>
                <View style={styles.value}>
                  <Text>
                    {moment(item.createdAt * 1000)
                      .local()
                      .format('YYYY-MM-DD HH:mm:ss')}
                  </Text>
                </View>
              </View>
              <View style={styles.row}>
                <View style={styles.label}>
                  <Text>{t('settings.family.familyMessage.creator')}</Text>
                </View>
                <View style={styles.value}>
                  <Text>{item.creator}</Text>
                </View>
              </View>
              <View style={styles.row}>
                <View style={styles.label}>
                  <Text>{t('settings.family.familyMessage.target')}</Text>
                </View>
                <View style={styles.value}>
                  <Text>{item.target}</Text>
                </View>
              </View>
              <View style={styles.row}>
                <View style={styles.label}>
                  <Text>{t('settings.family.familyMessage.operation')}</Text>
                </View>
                <View style={styles.value}>
                  <Text>{operations(item.operation)}</Text>
                </View>
              </View>
            </Card>
          ))}
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
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
    marginVertical: 1,
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  label: { width: '30%' },
  value: { width: '70%' },
}));

export default FamilyMessageStackScreen;
