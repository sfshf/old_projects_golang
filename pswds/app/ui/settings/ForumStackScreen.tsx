/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { RootStackParamList } from '../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { WebView } from 'react-native-webview';
import { DefaultCookieSessionKey } from '../../common/http/index';
import { makeStyles } from '@rneui/themed';
import { SafeAreaView } from 'react-native';
import { SlarkInfoContext } from '../../contexts/slark';

type ForumStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function ForumStackScreen({
  navigation,
}: ForumStackScreenProp): React.JSX.Element {
  const styles = useStyles();
  const uri = 'https://invoker.test.n1xt.net/site/Pswds';
  const { slarkInfo } = React.useContext(SlarkInfoContext);

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  return (
    <SafeAreaView style={styles.container}>
      <WebView
        source={{
          uri: uri,
          headers: {
            Cookie:
              DefaultCookieSessionKey + '=' + slarkInfo
                ? slarkInfo?.lSessionID
                : '',
          },
        }}
        sharedCookiesEnabled={true}
      />
    </SafeAreaView>
  );
}

const useStyles = makeStyles(() => ({
  container: {
    flex: 1,
  },
}));

export default ForumStackScreen;
