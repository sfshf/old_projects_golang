/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import {
  makeStyles,
  Icon,
  Button,
  useTheme,
  ListItem,
  Text,
} from '@rneui/themed';
import Clipboard from '@react-native-clipboard/clipboard';
import {
  Alert,
  FlatList,
  Pressable,
  SectionList,
  TouchableOpacity,
  View,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { SafeAreaView } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import { SlarkInfoContext } from '../../../contexts/slark';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SnackbarContext } from '../../../contexts/snackbar';
import { BackdropContext } from '../../../contexts/backdrop';
import { post } from '../../../common/http/post';
import { decryptByUserPrivateKey } from '../../../services/cipher';
import moment from 'moment';

interface Recover {
  self: boolean;
  id: number;
  createdAt: number;
  userID: number;
  email: string;
  checkedAt: number;
  ciphertext: string;
  plaintext: string;
}

type FamilyBackupRecoverStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'FamilyBackupRecoverStack'
>;

function FamilyBackupRecoverStackScreen({
  navigation,
  route,
}: FamilyBackupRecoverStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const styles = useStyles();
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [list, setList] = React.useState<Recover[]>([]);

  const onLoad = React.useCallback(async () => {
    try {
      // （1）检查用户登录状态
      if (!slarkInfo) {
        navigation.goBack();
        return;
      }
      setLoading(true);
      // （2）从远程获取家庭信息
      let respData = await post('/pswds/getFamilyBackupRecovers/v1', {
        userID: route.params.member.userID,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      // （3）处理 recovers
      const newList: Recover[] = [];
      // 某人（主动）找我帮忙
      if (respData.data.family) {
        respData.data.family.map((item: any) => {
          if (item.userID === route.params.member.userID) {
            newList.push({
              self: false,
              id: item.id,
              createdAt: item.createdAt,
              userID: item.userID,
              email: item.email,
              checkedAt: item.checkedAt,
              ciphertext: item.ciphertext,
              plaintext:
                item.ciphertext &&
                Buffer.from(
                  decryptByUserPrivateKey(
                    Buffer.from(item.ciphertext, 'hex'),
                    password,
                  ),
                ).toString('utf-8'),
            });
          }
        });
      }
      // 我（被动）帮助某人找回
      if (respData.data.self) {
        respData.data.self.map((item: any) => {
          if (item.userID === route.params.member.userID) {
            newList.push({
              self: true,
              id: item.id,
              createdAt: item.createdAt,
              userID: item.userID,
              email: item.email,
              checkedAt: item.checkedAt,
              ciphertext: item.ciphertext,
              plaintext:
                item.ciphertext &&
                Buffer.from(
                  decryptByUserPrivateKey(
                    Buffer.from(item.ciphertext, 'hex'),
                    password,
                  ),
                ).toString('utf-8'),
            });
          }
        });
      }
      // sort the list
      setList(
        newList.sort((a: Recover, b: Recover) => {
          return b.createdAt - a.createdAt;
        }),
      );
      setLoading(false);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, navigation, password, route.params.member]);

  React.useEffect(() => {
    onLoad();
  }, [onLoad]);

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
      title: route.params.member.email,
    });
  }, [navigation]);

  const onPressHelpRecover = async () => {
    try {
      setLoading(true);
      let respData = await post('/pswds/helpFamilyRecover/v1', {
        userID: route.params.member.userID,
        email: route.params.member.email,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      setLoading(false);
      Alert.alert(
        t('app.alert.emailHasSentTitle'),
        t('app.alert.emailHasSentMessage') + route.params.member.email,
        [
          {
            text: t('app.alert.okBtn'),
          },
        ],
      );
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const onPressCopy = (txt: string) => {
    Clipboard.setString(txt);
  };

  return (
    <SafeAreaView style={styles.container}>
      <FlatList
        data={list}
        renderItem={({ item }) => (
          <ListItem key={item.id} bottomDivider>
            <ListItem.Content>
              <View style={styles.textRow}>
                <Text h4 h4Style={styles.itemTitle}>
                  {item.self
                    ? t(
                        'settings.family.familyBackupRecover.labels.myRecoverRequest',
                      )
                    : t(
                        'settings.family.familyBackupRecover.labels.theirRecoverRequest',
                      )}
                </Text>
              </View>
              <View style={styles.textRow}>
                <Text
                  h4
                  h4Style={[styles.itemSubtitle, styles.itemSubtitleLabel]}>
                  {t('settings.family.familyBackupRecover.labels.sentAt')}
                </Text>
                <Text
                  h4
                  h4Style={[styles.itemSubtitle, styles.itemSubtitleContent]}>
                  {moment(item.createdAt * 1000)
                    .local()
                    .format('YYYY-MM-DD HH:mm:ss')}
                </Text>
              </View>
              {item.ciphertext === '' && (
                <>
                  {/* 可能性一：主动 -> 未被确认 */}
                  {!item.self && item.checkedAt === 0 && (
                    <View style={styles.textRow}>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleLabel,
                        ]}>
                        {t(
                          'settings.family.familyBackupRecover.labels.needToConfirm',
                        )}
                      </Text>
                    </View>
                  )}
                  {/* 可能性二：被动 -> 没超过3天，且未被拒绝 */}
                  {item.self && item.checkedAt === 0 && (
                    <View style={styles.textRow}>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleLabel,
                        ]}>
                        {t(
                          'settings.family.familyBackupRecover.labels.waitUntil',
                        )}
                      </Text>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleContent,
                        ]}>
                        {moment(item.createdAt * 1000)
                          .add(3, 'd')
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss')}
                      </Text>
                    </View>
                  )}
                  {/* 可能性三：被拒绝 */}
                  {item.self && item.checkedAt > 0 && (
                    <View style={styles.textRow}>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleLabel,
                        ]}>
                        {t(
                          'settings.family.familyBackupRecover.labels.emailRejectedAt',
                        )}
                      </Text>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleContent,
                        ]}>
                        {moment(item.checkedAt * 1000)
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss')}
                      </Text>
                    </View>
                  )}
                </>
              )}
              {/* 有密码的可能性：1. 对方主动请求；2. 我帮助他已过拒绝期，且在7天内 */}
              {item.plaintext && (
                <>
                  {!item.self && (
                    <View style={styles.textRow}>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleLabel,
                        ]}>
                        {t(
                          'settings.family.familyBackupRecover.labels.willExpiredAt',
                        )}
                      </Text>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleContent,
                        ]}>
                        {moment(item.createdAt * 1000)
                          .add(1, 'd')
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss')}
                      </Text>
                    </View>
                  )}
                  {item.self && (
                    <View style={styles.textRow}>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleLabel,
                        ]}>
                        {t(
                          'settings.family.familyBackupRecover.labels.willExpiredAt',
                        )}
                      </Text>
                      <Text
                        h4
                        h4Style={[
                          styles.itemSubtitle,
                          styles.itemSubtitleContent,
                        ]}>
                        {moment(item.createdAt * 1000)
                          .add(7, 'd')
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss')}
                      </Text>
                    </View>
                  )}
                  <View style={styles.textRow}>
                    <Text
                      h4
                      h4Style={[styles.itemSubtitle, styles.itemSubtitleLabel]}>
                      {t(
                        'settings.family.familyBackupRecover.labels.plaintext',
                      )}
                    </Text>
                    <Text h4 h4Style={styles.itemSubtitle}>
                      {item.plaintext}
                    </Text>
                    <TouchableOpacity
                      style={styles.copy}
                      onPress={() => {
                        onPressCopy(item.plaintext);
                      }}>
                      <Text h4 h4Style={styles.itemSubtitle}>
                        {t('app.button.copy')}
                      </Text>
                    </TouchableOpacity>
                  </View>
                </>
              )}
            </ListItem.Content>
          </ListItem>
        )}
        keyExtractor={(item: Recover) => item.id.toString()}
      />
      <View style={[styles.row, styles.bottomRow]}>
        <View style={styles.full}>
          <Button
            title={t('settings.family.familyBackupRecover.helpRecoverBtn')}
            containerStyle={styles.bottomBtn}
            titleStyle={styles.bottomBtnTitle}
            radius={8}
            onPress={onPressHelpRecover}
          />
        </View>
      </View>
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
  headerBtn: {
    marginHorizontal: 8,
    padding: 0,
  },
  menuButtonContainerStyle: {
    width: '100%',
  },
  menuButtonTitleStyle: {
    flex: 1,
    fontSize: 20,
    color: theme.colors.primary,
    textAlign: 'left',
  },
  itemAvatar: {
    borderRadius: 8,
  },
  itemTitle: { fontSize: 14, fontWeight: 500 },
  itemSharing: {
    fontSize: 12,
    backgroundColor: theme.colors.green0,
    padding: 4,
    borderRadius: 4,
  },
  itemShared: {
    fontSize: 12,
    backgroundColor: theme.colors.primary,
    marginHorizontal: 4,
    padding: 4,
    borderRadius: 4,
  },
  row: {
    flexDirection: 'row',
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  textRow: {
    flexDirection: 'row',
    marginVertical: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  label: {
    fontSize: 18,
    fontWeight: 'bold',
    marginVertical: 10,
  },
  itemSubtitle: { fontSize: 10, textAlign: 'center' },
  itemSubtitleLabel: { marginHorizontal: 5, marginVertical: 1 },
  itemSubtitleContent: { marginHorizontal: 5, marginVertical: 1 },
  copy: {
    marginHorizontal: 20,
    padding: 3,
    borderRadius: 4,
    backgroundColor: theme.colors.primary,
  },
  bottomBtn: {
    marginVertical: 8,
    width: '100%',
  },
  bottomBtnTitle: {
    fontSize: 12,
    fontWeight: 'normal',
  },
  itemAdmin: {
    fontSize: 12,
    backgroundColor: theme.colors.primary,
    padding: 4,
    borderRadius: 4,
  },
  bottomRow: { bottom: 50, position: 'absolute' },
  full: { width: '100%', padding: 1 },
}));

export default FamilyBackupRecoverStackScreen;
