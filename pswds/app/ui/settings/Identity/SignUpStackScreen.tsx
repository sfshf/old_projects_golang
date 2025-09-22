/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { makeStyles, Input, Button } from '@rneui/themed';
import { View, KeyboardAvoidingView, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';

export const CaptchaButton = ({
  type,
  email,
}: {
  type: 'register' | 'login';
  email: string;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const styles = useCbStyles();
  const { setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const tickerRef = React.useRef<null | NodeJS.Timeout>(null);
  const [title, setTitle] = React.useState<string>(
    t('settings.signup.fetchCaptcha'),
  );

  const sendRegistrationEmailCaptcha = async () => {
    if (!email) {
      setError(t('settings.signup.toast.emptyEmail'), t('app.toast.error'));
      return;
    }
    try {
      let reqUrl = '';
      if (type === 'register') {
        reqUrl = '/slark/user/sendRegistrationEmailCaptcha/v1';
      } else if (type === 'login') {
        reqUrl = '/slark/sendLoginEmailCode/v1';
      }
      setLoading(true);
      const respData = await post(reqUrl, { email });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      let limit = 60;
      tickerRef.current = setInterval(() => {
        if (limit === 0) {
          tickerRef.current && clearInterval(tickerRef.current);
          tickerRef.current = null;
          setTitle(t('settings.signup.fetchCaptcha'));
          return;
        } else {
          setTitle(limit + 's');
          limit -= 1;
        }
      }, 1000);
      setLoading(false);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  };

  return (
    <View style={styles.container}>
      <Button
        title={title}
        titleStyle={styles.titleStyle}
        type="clear"
        radius={8}
        disabled={tickerRef.current != null}
        onPress={sendRegistrationEmailCaptcha}
      />
    </View>
  );
};

const useCbStyles = makeStyles(() => ({
  container: { width: 100, height: 60 },
  titleStyle: { fontSize: 16 },
}));

type SignUpStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function SignUpStackScreen({
  navigation,
}: SignUpStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const [email, setEmail] = React.useState<string>('');
  const [captcha, setCaptcha] = React.useState<string>('');
  const { setLoading } = React.useContext(BackdropContext);
  const { setError } = React.useContext(SnackbarContext);
  const registerByEmail = async () => {
    // 1. handle parameters
    if (!email) {
      setError(t('settings.signup.toast.emptyEmail'), t('app.toast.error'));
      return;
    }
    if (!captcha) {
      setError(t('settings.signup.toast.emptyCaptcha'), t('app.toast.error'));
      return;
    }
    setLoading(true);
    try {
      // 2. post request
      let respData = await post('/slark/user/registerByEmailCaptcha/v1', {
        email,
        captcha,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.error'));
        return;
      }
      // 3. handle results
      // 注册成功后，因为没有解锁密码，所以直接跳转到设置解锁密码界面
      setLoading(false);
      navigation.navigate('EditUnlockPasswordStack', {
        loginInfo: {
          ...respData.data,
          lSessionID: respData.lSessionID,
        },
      });
      return;
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    } finally {
      setLoading(false);
    }
  };
  const onChangeEmail = (newText: string) => {
    newText = newText.trim();
    setEmail(newText);
  };
  const onChangeCaptcha = (newText: string) => {
    newText = newText.trim();
    setCaptcha(newText);
  };
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView>
        <ScrollView>
          <View style={styles.body}>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                label={t('settings.signup.email')}
                labelStyle={styles.inputLabel}
                inputStyle={styles.inputStyle}
                placeholder={t('settings.signup.emailPlaceholder')}
                onChangeText={onChangeEmail}
                value={email}
                rightIcon={<CaptchaButton type="register" email={email} />}
              />
            </View>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                label={t('settings.signup.captcha')}
                labelStyle={styles.inputLabel}
                inputStyle={styles.inputStyle}
                placeholder={t('settings.signup.captchaPlaceholder')}
                onChangeText={onChangeCaptcha}
                value={captcha}
                keyboardType="numeric"
              />
            </View>
            <View style={styles.row}>
              <Button
                title={t('settings.signup.label')}
                containerStyle={styles.signUpBtn}
                size="lg"
                radius={8}
                onPress={registerByEmail}
              />
            </View>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
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
  inputLabel: {
    fontSize: 20,
    color: theme.colors.black,
  },
  inputStyle: {
    height: 60,
  },
  signUpBtn: {
    width: '100%',
  },
}));

export default SignUpStackScreen;
