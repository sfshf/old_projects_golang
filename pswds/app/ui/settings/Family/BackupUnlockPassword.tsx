/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, ScrollView } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Button, makeStyles, ListItem, useTheme, Avatar } from '@rneui/themed';
import { SafeAreaView } from 'react-native';
import moment from 'moment';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';
import { FamilyMember } from './FamilyStackScreen';
import SetupFamilyBackupOverlay from '../../../components/SetupFamilyBackupOverlay';

export interface FamilyBackup {
  id: number;
  userID: number;
  email: string;
  createdAt: number;
}

type BackupUnlockPasswordStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'BackupUnlockPasswordStack'
>;

function BackupUnlockPasswordStackScreen({
  navigation,
  route,
}: BackupUnlockPasswordStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { setSuccess, setError, setWarning } =
    React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [familyMembers, setFamilyMembers] = React.useState<FamilyMember[]>(
    route.params.familyMembers,
  );
  const [selfBackups, setSelfBackups] = React.useState<FamilyBackup[]>([]);
  const [familyBackups, setFamilyBackups] = React.useState<FamilyBackup[]>([]);
  const [backupVisible, setBackupVisible] = React.useState(false);

  const onLoad = React.useCallback(async () => {
    try {
      // （1）检查用户登录状态
      if (!slarkInfo) {
        navigation.goBack();
        return;
      }
      setLoading(true);
      // （2）从远程获取家庭备份
      let respData = await post('/pswds/getFamilyBackups/v1');
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      if (respData.data.self) {
        setSelfBackups(respData.data.self);
      } else {
        setSelfBackups([]);
      }
      if (respData.data.family) {
        setFamilyBackups(respData.data.family);
      } else {
        setFamilyBackups([]);
      }
      setLoading(false);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, navigation]);

  React.useEffect(() => {
    onLoad();
  }, [onLoad]);
  // sharing backup
  const backuped = React.useCallback(
    (userID: number): null | FamilyBackup => {
      for (let i = 0; i < selfBackups.length; i++) {
        if (selfBackups[i].userID === userID) {
          return selfBackups[i];
        }
      }
      return null;
    },
    [selfBackups],
  );
  // shared backup
  const backupedBy = React.useCallback(
    (userID: number): null | FamilyBackup => {
      for (let i = 0; i < familyBackups.length; i++) {
        if (familyBackups[i].userID === userID) {
          return familyBackups[i];
        }
      }
      return null;
    },
    [familyBackups],
  );

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  const onPressSetupBackup = () => {
    setBackupVisible(true);
  };

  const onPressRecover = (item: FamilyMember) => {
    navigation.navigate('FamilyBackupRecoverStack', { member: item });
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView>
        {familyMembers &&
          familyMembers.map(item => {
            if (item.userID != slarkInfo?.userID) {
              return (
                <ListItem key={item.id} bottomDivider>
                  <Avatar
                    rounded
                    icon={{
                      name: 'person-outline',
                      type: 'material',
                      size: 26,
                    }}
                    containerStyle={styles.avatarContainer}
                  />
                  <ListItem.Content>
                    <ListItem.Title style={styles.itemTitle}>
                      {item.email}
                    </ListItem.Title>
                    {/* 被分享备份 */}
                    {backupedBy(item.userID) && (
                      <ListItem.Subtitle style={styles.itemSubtitle}>
                        {t(
                          'settings.family.backupUnlockPassword.sharedBackupAt',
                        ) +
                          moment(backupedBy(item.userID)!.createdAt * 1000)
                            .local()
                            .format('YYYY-MM-DD HH:mm:ss')}
                      </ListItem.Subtitle>
                    )}
                    {/* 分享备份 */}
                    {backuped(item.userID) && (
                      <ListItem.Subtitle style={styles.itemSubtitle}>
                        {t(
                          'settings.family.backupUnlockPassword.sharingBackupAt',
                        ) +
                          moment(backuped(item.userID)!.createdAt * 1000)
                            .local()
                            .format('YYYY-MM-DD HH:mm:ss')}
                      </ListItem.Subtitle>
                    )}
                  </ListItem.Content>
                  <Button
                    title={t(
                      'settings.family.backupUnlockPassword.recoverPasswordBtn',
                    )}
                    titleStyle={styles.recoverBtn}
                    radius={8}
                    onPress={() => {
                      onPressRecover(item);
                    }}
                  />
                </ListItem>
              );
            }
          })}
      </ScrollView>
      <View style={[styles.row, styles.bottomRow]}>
        <View style={styles.full}>
          <Button
            title={t('settings.family.backupUnlockPassword.backupBtn')}
            containerStyle={styles.bottomBtn}
            titleStyle={styles.bottomBtnTitle}
            radius={8}
            onPress={onPressSetupBackup}
          />
        </View>
      </View>
      <SetupFamilyBackupOverlay
        visible={backupVisible}
        setVisible={setBackupVisible}
        familyMembers={familyMembers}
        backupings={selfBackups}
        callback={onLoad}
      />
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  row: {
    flexDirection: 'row',
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  label: {
    fontSize: 18,
    fontWeight: 'bold',
    marginVertical: 10,
  },
  labelText: {
    fontSize: 16,
    fontWeight: 500,
  },
  wraped: { flexWrap: 'wrap' },
  createBtn: { width: '100%', marginVertical: 8 },
  createBtnTitle: { fontSize: 20 },
  headIcon: {},
  avatarContainer: {
    backgroundColor: theme.colors.surface,
    height: 30,
    width: 30,
  },
  itemTitle: { fontSize: 14, fontWeight: 500 },
  itemSubtitle: { fontSize: 10 },
  recoverBtn: {
    fontSize: 9,
    fontWeight: 300,
  },
  card: { borderRadius: 8 },
  invitationBtnRow: { justifyContent: 'flex-end', marginVertical: 8 },
  invitationBtn: { height: 40, marginHorizontal: 8 },
  invitationBtnTitle: { fontSize: 12 },
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

export default BackupUnlockPasswordStackScreen;
