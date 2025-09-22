/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import {
  AppState,
  View,
  Alert,
  KeyboardAvoidingView,
  ScrollView,
  Platform,
  useColorScheme,
} from 'react-native';
import {
  Overlay,
  Button,
  makeStyles,
  Icon,
  Input,
  useThemeMode,
  ThemeMode,
  Dialog,
  useTheme,
  Text,
} from '@rneui/themed';
import { useTranslation } from 'react-i18next';
import { storage } from '../common/mmkv';
import { useNavigation } from '@react-navigation/native';
import { RootStackParamList } from '../navigation/routes';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { keccak_256 } from '@noble/hashes/sha3';
import * as LocalAuthentication from '../services/localAuthentication';
import moment from 'moment';
import { dropTables } from '../common/sqlite';
import {
  currentSlarkInfo,
  SlarkInfo,
  updateSlarkInfo,
} from '../services/slark';
import {
  currentAutoLockSetting,
  currentUnlockPasswordSetting,
  updateUnlockPasswordSetting,
} from '../services/unlockPassword';
import { syncBackup } from '../services/backup';
import { UnlockPasswordContext } from '../contexts/unlockPassword';
import { SlarkInfoContext } from '../contexts/slark';
import { BackdropContext } from '../contexts/backdrop';
import { SnackbarContext } from '../contexts/snackbar';
import { post } from '../common/http/post';
import { currentThemeSetting } from '../services/theme';
import { StatusBarStyleContext } from '../contexts/statusbar';

interface tooltips {
  tip: string;
  errTip: string;
}

type TooltipsAction =
  | {
      type: 'tips';
      value: tooltips;
    }
  | {
      type: 'tip';
      value: tooltips['tip'];
    }
  | { type: 'errTip'; value: tooltips['errTip'] };

const tooltipsReducer = (state: tooltips, action: TooltipsAction) => {
  switch (action.type) {
    case 'tips':
      return { ...action.value };
    case 'tip':
      return { ...state, tip: action.value };
    case 'errTip':
      return { ...state, errTip: action.value };
    default:
      return { ...state };
  }
};

interface LockScreenProp {
  children?: React.ReactNode;
}

function LockScreen({ children }: LockScreenProp): React.JSX.Element {
  const styles = useStyles();
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { slarkInfo, setSlarkInfo } = React.useContext(SlarkInfoContext);
  const { setError } = React.useContext(SnackbarContext);
  const navigation =
    useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const [visible, setVisible] = React.useState<boolean>(false);
  const { setLoading } = React.useContext(BackdropContext);
  const [localLoading, setLocalLoading] = React.useState(false);

  // tips' states
  const initTooltips: tooltips = {
    tip: 'lockScreen.tips.1',
    errTip: '',
  };
  const [tooltips, dispathTooltips] = React.useReducer(
    tooltipsReducer,
    initTooltips,
  );

  const logoutBySession = async () => {
    try {
      // 1. 清除内存中的账号信息、清除内存中的密码
      setSlarkInfo(null);
      updateSlarkInfo(null);
      updateUnlockPasswordSetting(slarkInfo ? slarkInfo.userID : -1, null);
      setPassword('');
      // 2. mmkv中存的数据sqlite中存的数据
      storage.clearAll();
      await dropTables();
      setVisible(false);
      dispathTooltips({
        type: 'tips',
        value: { tip: 'lockScreen.tips.1', errTip: '' },
      });
      if (slarkInfo) {
        setLocalLoading(true);
        await post('/slark/user/logout/v1');
        setLocalLoading(false);
      }
      navigation.navigate('SignInStack');
    } catch (error) {
      Alert.alert(t('app.toast.internalError'), error as string, [
        {
          text: t('app.alert.okBtn'),
          style: 'destructive',
        },
      ]);
    } finally {
    }
  };

  const onPressRecoverUnlockPassword = () => {
    setVisible(false);
    dispathTooltips({
      type: 'tips',
      value: { tip: 'lockScreen.tips.1', errTip: '' },
    });
    navigation.navigate('RecoverUnlockPasswordStack');
  };

  const [password, setPassword] = React.useState(''); // 密码缓存

  const syncData = React.useCallback(async () => {
    if (!slarkInfo) {
      return;
    }
    try {
      setLoading(true);
      await syncBackup(slarkInfo, password);
      setLoading(false);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
      return;
    }
  }, [slarkInfo]);

  const [needToLogin, setNeedToLogin] = React.useState(true);
  const loginBySession = React.useCallback(
    async (slarkInfo: SlarkInfo, password: string): Promise<boolean> => {
      try {
        // a. 使用session进行登录
        let respData = await post('/slark/user/loginBySession/v1');
        if (respData.code !== 0) {
          // 1. LoginBySession失败
          // 1-1. 如果是401*:
          if (respData.code === 100040102 || respData.code === 100040103) {
            // 执行LoginBySecondaryPassword，进行登录
            respData = await post('/slark/loginBySecondaryPassword/v1', {
              email: slarkInfo.email,
              passwordHash: Buffer.from(keccak_256(password)).toString('hex'),
            });
            if (respData.code !== 0) {
              setLocalLoading(false);
              dispathTooltips({
                type: 'tips',
                value: { tip: '', errTip: 'lockScreen.errTips.4' },
              });
              return false;
            }
          } else {
            setLocalLoading(false);
            // 1-2. 如果不是401*:
            dispathTooltips({
              type: 'tips',
              value: { tip: '', errTip: 'lockScreen.errTips.3' },
            });
            return false;
          }
        } else {
          if (respData.lSessionID) {
            updateSlarkInfo({
              ...respData.data,
              lSessionID: respData.lSessionID,
            });
          } else {
            if (Platform.OS !== 'ios') {
              dispathTooltips({
                type: 'tips',
                value: { tip: '', errTip: 'lockScreen.errTips.3' },
              });
            }
          }
        }
        return true;
      } catch (error) {
        throw error;
      }
    },
    [],
  );
  const handleUnlock = React.useCallback(async () => {
    try {
      // 1. 已登录
      if (slarkInfo) {
        // （1）用户刚重启了APP
        if (needToLogin) {
          setLocalLoading(true);
          const ok = await loginBySession(slarkInfo, password);
          if (!ok) {
            return;
          }
          // 2. LoginBySession成功
          // 2-1. 同步数据
          const respData = await syncBackup(slarkInfo, password);
          if (respData.code !== 0) {
            setLocalLoading(false);
            dispathTooltips({
              type: 'tips',
              value: { tip: '', errTip: respData.message },
            });
            return;
          }
          setLocalLoading(false);
          // 2-2. 解开锁屏
          setVisible(false);
          dispathTooltips({
            type: 'tips',
            value: { tip: 'lockScreen.tips.1', errTip: '' },
          });
          navigation.navigate('HomeStack', { screen: 'Home' });
          return;
        } else {
          // 不需要LoginBySession的话，在进入应用后再进行数据同步
          setTimeout(async () => {
            syncData();
          }, 0);
        }
      }
      // 2. 未登录
      setVisible(false);
      dispathTooltips({
        type: 'tips',
        value: { tip: 'lockScreen.tips.1', errTip: '' },
      });
      navigation.navigate('HomeStack', { screen: 'Home' });
      return;
    } catch (error) {
      throw error;
    }
  }, [slarkInfo, password, needToLogin]);

  const verifyUnlockPassword = React.useCallback(async () => {
    try {
      // 1. check unlock password input
      if (!password) {
        dispathTooltips({
          type: 'tips',
          value: { tip: '', errTip: 'lockScreen.errTips.1' },
        });
        return;
      }
      // 2. check unlock password
      let curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      if (!curSetting) {
        return;
      }
      if (
        Buffer.from(keccak_256(password)).toString('hex') ===
        curSetting.passwordHash
      ) {
        // unlock password is right
        await handleUnlock();
      } else {
        // unlock password is wrong
        dispathTooltips({
          type: 'tips',
          value: { tip: '', errTip: 'lockScreen.errTips.2' },
        });
      }
    } catch (error) {
      dispathTooltips({
        type: 'tips',
        value: { tip: '', errTip: error as string },
      });
    }
  }, [slarkInfo, password]);

  // auto lock
  const appState = React.useRef(AppState.currentState);
  // support fingerprint
  const [supportFingerprint, setSupportFingerprint] = React.useState(false);

  const authenticateFingerprint = React.useCallback(async () => {
    try {
      let userID = slarkInfo ? slarkInfo.userID : -1;
      let curSetting = currentUnlockPasswordSetting(userID);
      if (!curSetting || !curSetting.supportFingerprint) {
        return;
      }
      if (Platform.OS === 'ios') {
        const result = await LocalAuthentication.requestFaceIDPermission();
        if (!result.granted) {
          dispathTooltips({
            type: 'tips',
            value: { tip: '', errTip: 'lockScreen.errTips.2' },
          });
          return;
        }
      }
      const passwd = await LocalAuthentication.getGenericPassword(
        userID.toString(),
      );
      if (passwd) {
        // 只有指纹验证成功才能拿到解锁密码
        setPassword(passwd);
        // 解锁密码验证成功后：
        await handleUnlock();
      }
    } catch (error) {
      dispathTooltips({
        type: 'tips',
        value: { tip: '', errTip: error as string },
      });
    }
  }, [slarkInfo]);

  const notSignedIn = React.useCallback(async () => {
    // 1. 检查有没有未登录的解锁密码
    let curSetting = currentUnlockPasswordSetting(-1);
    if (!curSetting) {
      // 1-1、没有未登录的解锁密码，引导至登录页
      navigation.navigate('SignInStack');
      return;
    } else {
      // 1-2、有未登录的解锁密码
      setSupportFingerprint(
        curSetting.supportFingerprint ? curSetting.supportFingerprint : false,
      );
      setVisible(true);
      setPassword(''); // 锁定时，要清理掉内存中的解锁密码和密钥
      if (curSetting.supportFingerprint) {
        authenticateFingerprint();
      }
    }
  }, [navigation, authenticateFingerprint]);

  const hasSignedIn = React.useCallback(
    async (loginInfo: SlarkInfo) => {
      // 1. 检查登录用户的unlock password设置
      let curUnlockPasswordSetting = currentUnlockPasswordSetting(
        loginInfo.userID,
      );
      // 首次登录页/注册页之后，如果没有设置解锁密码，则不会有解锁密码配置；引导至创建解锁密码页
      if (!curUnlockPasswordSetting) {
        navigation.navigate('EditUnlockPasswordStack', { loginInfo });
        return;
      }
      // 2. 更新锁屏页面的状态
      setSupportFingerprint(
        curUnlockPasswordSetting.supportFingerprint
          ? curUnlockPasswordSetting.supportFingerprint
          : false,
      ); // 是否启用了指纹验证
      setVisible(true);
      setPassword(''); // 锁定时，要清理掉内存中的解锁密码和密钥
      if (curUnlockPasswordSetting.supportFingerprint) {
        authenticateFingerprint();
      }
    },
    [slarkInfo, authenticateFingerprint],
  );

  const passwordRef = React.useRef<null | any>(null);
  // auto lock effect
  const lastTS = React.useRef<number>(0);
  // auto update app theme
  const { setMode } = useThemeMode();
  const { setStatusBarStyle } = React.useContext(StatusBarStyleContext);
  const systemColorScheme = useColorScheme();
  React.useEffect(() => {
    let userID = slarkInfo ? slarkInfo.userID : -1;
    const subscription = AppState.addEventListener('change', nextAppState => {
      if (appState.current === 'active' && nextAppState === 'background') {
        appState.current = nextAppState;
        lastTS.current = moment().valueOf();
        return;
      }
      if (appState.current === 'background' && nextAppState === 'active') {
        appState.current = nextAppState;
        if (!password) {
          return;
        }
        // 1. check lock screen
        const nowTS = moment().valueOf();
        let timeLag = 30 * 1000;
        const curSetting = currentAutoLockSetting(userID);
        if (curSetting) {
          timeLag = curSetting.timeLag > 0 ? curSetting.timeLag * 1000 - 1 : 0;
        }
        if (nowTS - lastTS.current > timeLag) {
          lastTS.current = nowTS;
          // APP从后台进入前台，需要锁屏
          setVisible(true);
          setPassword(''); // 锁定时，要清理掉内存中的解锁密码和密钥
          setNeedToLogin(false); // APP退到后台，未重启APP，则不需要LoginBySession
          // 是否启用了指纹验证
          let curSetting = currentUnlockPasswordSetting(userID);
          if (curSetting && curSetting.supportFingerprint) {
            setSupportFingerprint(
              curSetting.supportFingerprint
                ? curSetting.supportFingerprint
                : false,
            );
            authenticateFingerprint();
          }
        } else {
          // 不需要锁屏，但要检查同步数据
          syncData();
        }
        // 2. check system color theme
        if (systemColorScheme) {
          let curSetting = currentThemeSetting();
          if (curSetting === 'default') {
            curSetting = systemColorScheme;
          }
          setMode(curSetting as ThemeMode);
          setStatusBarStyle(curSetting === 'dark' ? 'light' : 'dark');
        }
      }
    });
    return () => {
      subscription.remove();
    };
  }, [slarkInfo, systemColorScheme, password]);

  const onChangePassword = (newText: string) => {
    setPassword(newText.trim());
  };

  React.useEffect(() => {
    // 1. app inspect
    // (1) 登录成功后，password已存值
    const loginInfo = currentSlarkInfo();
    let curSetting = currentUnlockPasswordSetting(
      loginInfo ? loginInfo.userID : -1,
    );
    if (
      curSetting &&
      Buffer.from(keccak_256(password)).toString('hex') ===
        curSetting.passwordHash
    ) {
      setVisible(false);
      return;
    }
    //（2）重启后，password没存值
    if (!loginInfo) {
      // (2-1) 没有用户登录
      notSignedIn();
    } else {
      // (2-2) 有用户登录
      hasSignedIn(loginInfo);
    }
  }, [slarkInfo]); // 不依赖 password state，所以只会拿到password的初始快照/context设置快照

  const onShow = () => {
    // auto focus on password input
    if (passwordRef.current && passwordRef.current.focus) {
      setTimeout(() => passwordRef.current.focus(), 0);
    }
  };

  return (
    <>
      <Overlay
        fullScreen
        overlayStyle={styles.container}
        isVisible={visible}
        onShow={onShow}>
        <KeyboardAvoidingView>
          <ScrollView>
            <Text style={styles.title}>{t('lockScreen.title')}</Text>
            <Icon size={50} type="feather" name="lock" />
            {!supportFingerprint && (
              <View style={styles.content}>
                {tooltips.tip && (
                  <Text style={styles.normalTip}>{t(tooltips.tip)}</Text>
                )}
                {tooltips.errTip && (
                  <Text style={styles.errorTip}>{t(tooltips.errTip)}</Text>
                )}
                {!localLoading && (
                  <>
                    <Input
                      autoCapitalize={'none'}
                      ref={passwordRef}
                      containerStyle={styles.inputContainer}
                      inputContainerStyle={styles.inputInnerContainer}
                      value={password}
                      onChangeText={onChangePassword}
                    />
                    <Button
                      title={t(
                        'settings.unlockPassword.editUnlockPassword.verifyBtn',
                      )}
                      containerStyle={styles.verifyButtonContainer}
                      titleStyle={styles.verifyButtonText}
                      onPress={verifyUnlockPassword}
                    />
                  </>
                )}
                {localLoading && (
                  <Dialog.Loading loadingProps={{ size: 'small' }} />
                )}
              </View>
            )}
            {supportFingerprint && (
              <View style={styles.content}>
                {tooltips.tip && (
                  <Text style={styles.normalTip}>{t(tooltips.tip)}</Text>
                )}
                {tooltips.errTip && (
                  <Text style={styles.errorTip}>{t(tooltips.errTip)}</Text>
                )}
                {!localLoading && (
                  <Icon
                    containerStyle={styles.fingerprintBtn}
                    size={50}
                    type={
                      Platform.OS === 'ios' ? 'material-community' : '"entypo"'
                    }
                    name={
                      Platform.OS === 'ios' ? 'face-recognition' : 'fingerprint'
                    }
                    onPress={authenticateFingerprint}
                  />
                )}
                {localLoading && (
                  <Dialog.Loading loadingProps={{ size: 'small' }} />
                )}
              </View>
            )}
            {!localLoading && (
              <View style={styles.bottomRow}>
                <Text style={styles.logout} onPress={logoutBySession}>
                  {t('lockScreen.logout')}
                </Text>
                <Text
                  style={styles.recoverTipText}
                  onPress={onPressRecoverUnlockPassword}>
                  {t('lockScreen.recoverUnlockPassword')}
                </Text>
              </View>
            )}
          </ScrollView>
        </KeyboardAvoidingView>
      </Overlay>
      <UnlockPasswordContext.Provider
        value={{
          password,
          setPassword,
          setVisible,
          supportFingerprint,
          setSupportFingerprint,
        }}>
        {children}
      </UnlockPasswordContext.Provider>
    </>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    backgroundColor: theme.colors.background,
  },
  title: {
    paddingTop: 50,
    fontSize: 25,
    fontWeight: '700',
    textAlign: 'center',
    paddingBottom: 40,
  },
  icon: {
    width: 217 / 2.4,
    height: 158 / 2.4,
    marginLeft: 'auto',
    marginRight: 'auto',
  },
  content: {
    height: 200,
    marginHorizontal: 20,
  },
  normalTip: {
    paddingTop: 30,
    textAlign: 'center',
  },
  errorTip: {
    paddingTop: 30,
    color: theme.colors.error,
    textAlign: 'center',
  },
  inputContainer: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    marginTop: 20,
    borderRadius: 8,
    borderColor: theme.colors.black,
    color: theme.colors.black,
  },
  inputInnerContainer: {
    borderBottomWidth: 0,
  },
  verifyButtonContainer: {
    marginTop: 20,
    borderRadius: 8,
  },
  verifyButtonText: {
    textAlign: 'center',
    fontSize: 20,
    fontWeight: '700',
  },
  fingerprintBtn: {
    height: 150,
    alignItems: 'center',
    justifyContent: 'center',
  },
  bottomRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginVertical: 20,
    marginHorizontal: 20,
  },
  logout: {
    color: theme.colors.secondary,
  },
  recoverTipText: {
    color: theme.colors.warning,
  },
}));

export default LockScreen;
