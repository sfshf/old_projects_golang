/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { ScrollView, TouchableWithoutFeedback, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { HomeTabsParamList, RootStackParamList } from '../../navigation/routes';
import { Icon, Card, Text } from '@rneui/themed';
import { useTranslation } from 'react-i18next';
import { useTheme, makeStyles, Button } from '@rneui/themed';
import { storage } from '../../common/mmkv';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { CompositeScreenProps } from '@react-navigation/native';
import { getBuildNumber, getVersion } from 'react-native-device-info';
import { SafeAreaView } from 'react-native';
import { dropTables } from '../../common/sqlite';
import { UnlockPasswordContext } from '../../contexts/unlockPassword';
import { updateSlarkInfo } from '../../services/slark';
import { SlarkInfoContext } from '../../contexts/slark';
import { BackdropContext } from '../../contexts/backdrop';
import { SnackbarContext } from '../../contexts/snackbar';
import { post } from '../../common/http/post';
import { updateUnlockPasswordSetting } from '../../services/unlockPassword';

type SettingsTabScreenProp = CompositeScreenProps<
  BottomTabScreenProps<HomeTabsParamList>,
  NativeStackScreenProps<RootStackParamList>
>;

function SettingsTabScreen({
  navigation,
}: SettingsTabScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = useStyles();
  const { setLoading } = React.useContext(BackdropContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { slarkInfo, setSlarkInfo } = React.useContext(SlarkInfoContext);
  const { setPassword, setVisible } = React.useContext(UnlockPasswordContext);

  const onPressLogin = () => {
    navigation.navigate('SignInStack');
  };

  const onPressLogout = async () => {
    try {
      setLoading(true);
      const respData = await post('/slark/user/logout/v1');
      setLoading(false);
      if (respData.code !== 0) {
        setError(respData.message, t('app.toast.error'));
        return;
      }
      // 1. 清除内存中的账号信息、清除内存中的密码
      updateUnlockPasswordSetting(slarkInfo ? slarkInfo.userID : -1, null);
      setPassword('');
      updateSlarkInfo(null);
      setSlarkInfo(null);
      // 2. mmkv中存的数据sqlite中存的数据
      storage.clearAll();
      await dropTables();
      setVisible(false);
      setSuccess(t('app.toast.success'));
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const onPressFamilyShare = () => {
    navigation.navigate('FamilyStack');
  };

  const onPressPrivacyEmail = () => {
    navigation.navigate('PrivacyEmailStack', {});
  };

  const onPressI18n = () => {
    navigation.navigate('I18nStack');
  };

  const onPressTheme = () => {
    navigation.navigate('ThemeStack');
  };

  const onPressUnlockPassword = () => {
    navigation.navigate('UnlockPasswordStack');
  };

  const onPressForum = () => {
    navigation.navigate('ForumStack');
  };

  const countRef = React.useRef(0);
  const countTimoutRef = React.useRef<null | NodeJS.Timeout>(null);

  const onPressVersion = () => {
    if (countRef.current === 4) {
      countRef.current = 0;
      navigation.navigate('DebugStack');
    } else {
      countRef.current = countRef.current + 1;
    }
    if (countTimoutRef.current == null) {
      countTimoutRef.current = setTimeout(() => {
        countRef.current = 0;
        countTimoutRef.current = null;
      }, 3000);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView>
        <Card containerStyle={styles.card}>
          <Card.Title>
            <Text h4 style={styles.cardTitleText}>
              {t('settings.cards.title1')}
            </Text>
          </Card.Title>
          <Card.Title>
            {slarkInfo && (
              <Text style={[styles.cardTitleText, styles.accountText]}>
                {slarkInfo.email}
              </Text>
            )}
          </Card.Title>
          <Card.Divider />
          {!slarkInfo && (
            <Button
              icon={
                <Icon
                  type="antdesign"
                  name="login"
                  style={styles.menuButtonIcon}
                />
              }
              size="lg"
              buttonStyle={styles.menuButton}
              containerStyle={styles.menuButtonContainerStyle}
              titleStyle={styles.menuButtonTitleStyle}
              onPress={onPressLogin}>
              {t('settings.signin.label')}
              <Icon
                size={20}
                name="arrow-forward-ios"
                color={theme.colors.black}
              />
            </Button>
          )}
          {slarkInfo && (
            <>
              <Button
                icon={
                  <Icon
                    type="antdesign"
                    name="logout"
                    style={styles.menuButtonIcon}
                  />
                }
                size="lg"
                buttonStyle={styles.menuButton}
                containerStyle={styles.menuButtonContainerStyle}
                titleStyle={styles.menuButtonTitleStyle}
                onPress={onPressLogout}>
                {t('settings.signout.label')}
                <Icon
                  size={20}
                  name="arrow-forward-ios"
                  color={theme.colors.black}
                />
              </Button>
              <Button
                icon={
                  <Icon
                    type="fontisto"
                    name="email"
                    style={styles.menuButtonIcon}
                  />
                }
                size="lg"
                buttonStyle={styles.menuButton}
                containerStyle={styles.menuButtonContainerStyle}
                titleStyle={styles.menuButtonTitleStyle}
                onPress={onPressPrivacyEmail}>
                {t('settings.email.label')}
                <Icon
                  size={20}
                  name="arrow-forward-ios"
                  color={theme.colors.black}
                />
              </Button>
              <Button
                icon={
                  <Icon
                    type="entypo"
                    name="share"
                    style={styles.menuButtonIcon}
                  />
                }
                size="lg"
                buttonStyle={styles.menuButton}
                containerStyle={styles.menuButtonContainerStyle}
                titleStyle={styles.menuButtonTitleStyle}
                onPress={onPressFamilyShare}>
                {t('settings.family.label')}
                <Icon
                  size={20}
                  name="arrow-forward-ios"
                  color={theme.colors.black}
                />
              </Button>
            </>
          )}
        </Card>
        <Card containerStyle={styles.card}>
          <Card.Title>
            <Text h4 style={styles.cardTitleText}>
              {t('settings.cards.title2')}
            </Text>
          </Card.Title>
          <Card.Divider />
          <Button
            size="lg"
            icon={
              <Icon
                type="font-awesome"
                name="language"
                style={styles.menuButtonIcon}
              />
            }
            buttonStyle={styles.menuButton}
            containerStyle={styles.menuButtonContainerStyle}
            titleStyle={styles.menuButtonTitleStyle}
            onPress={onPressI18n}>
            {t('settings.language.label')}
            <Icon
              size={20}
              name="arrow-forward-ios"
              color={theme.colors.black}
            />
          </Button>
          <Button
            size="lg"
            icon={
              <Icon
                type="material-community"
                name="theme-light-dark"
                style={styles.menuButtonIcon}
              />
            }
            buttonStyle={styles.menuButton}
            containerStyle={styles.menuButtonContainerStyle}
            titleStyle={styles.menuButtonTitleStyle}
            onPress={onPressTheme}>
            {t('settings.theme.label')}
            <Icon
              size={20}
              name="arrow-forward-ios"
              color={theme.colors.black}
            />
          </Button>
          <Button
            size="lg"
            icon={
              <Icon type="feather" name="lock" style={styles.menuButtonIcon} />
            }
            buttonStyle={styles.menuButton}
            containerStyle={styles.menuButtonContainerStyle}
            titleStyle={styles.menuButtonTitleStyle}
            onPress={onPressUnlockPassword}>
            {t('settings.unlockPassword.label')}
            <Icon
              size={20}
              name="arrow-forward-ios"
              color={theme.colors.black}
            />
          </Button>
        </Card>
        <Card containerStyle={styles.card}>
          <Card.Title>
            <Text h4 style={styles.cardTitleText}>
              {t('settings.cards.title3')}
            </Text>
          </Card.Title>
          <Card.Divider />
          <Button
            size="lg"
            icon={
              <Icon
                type="material-community"
                name="forum"
                style={styles.menuButtonIcon}
              />
            }
            buttonStyle={styles.menuButton}
            containerStyle={styles.menuButtonContainerStyle}
            titleStyle={styles.menuButtonTitleStyle}
            onPress={onPressForum}>
            {t('settings.forum.label')}
            <Icon
              size={20}
              name="arrow-forward-ios"
              color={theme.colors.black}
            />
          </Button>
          <TouchableWithoutFeedback onPress={onPressVersion}>
            <View style={styles.versionBtn}>
              <Icon
                type="octicon"
                name="versions"
                style={styles.menuButtonIcon}
              />
              <Text style={styles.versionTitle}>{t('settings.version')}</Text>
              <Text style={styles.versionText}>
                {getVersion() + '(' + getBuildNumber() + ')'}
              </Text>
            </View>
          </TouchableWithoutFeedback>
        </Card>
      </ScrollView>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  menuButton: {
    backgroundColor: theme.colors.white,
  },
  menuButtonIcon: { marginRight: 10 },
  menuButtonContainerStyle: {
    width: '100%',
  },
  menuButtonTitleStyle: {
    flex: 1,
    fontSize: 20,
    color: theme.colors.black,
    textAlign: 'left',
  },
  cardTitleText: { textAlign: 'left' },
  accountText: { fontSize: 14 },
  versionBtn: {
    flexDirection: 'row',
    marginHorizontal: 16,
    alignItems: 'center',
  },
  versionTitle: { fontSize: 20, color: theme.colors.black },
  versionText: { fontWeight: 'bold', fontSize: 16, marginLeft: '50%' },
  card: { borderRadius: 8 },
}));

export default SettingsTabScreen;
