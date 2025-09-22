/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, ScrollView, TouchableOpacity, Alert } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import {
  Button,
  makeStyles,
  ListItem,
  useTheme,
  Icon,
  Text,
  Avatar,
  Card,
} from '@rneui/themed';
import { SafeAreaView } from 'react-native';
import moment from 'moment';
import CreateFamilyOverlay from '../../../components/CreateFamilyOverlay';
import {
  cancelSharingPasswordsByUserIDAsync,
  deletePasswordsByUserIDAsync,
  deletePasswordsFromFamilyAsync,
} from '../../../common/sqlite/dao/password';
import {
  cancelSharingRecordsByUserIDAsync,
  deleteRecordsByUserIDAsync,
  deleteRecordsFromFamilyAsync,
} from '../../../common/sqlite/dao/record';
import {
  clearSharedDataMembersAsync,
  deleteSharedDataMembersByUserIDAsync,
} from '../../../common/sqlite/dao/sharedDataMember';
import { useActionSheet } from '@expo/react-native-action-sheet';
import {
  deleteOtherFamilyMembers,
  downloadSharedData,
  OtherFamilyMember,
  updateOtherFamilyMembers,
} from '../../../services/family';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';
import { decryptByUserPrivateKey } from '../../../services/cipher';
import { SlarkInfo } from '../../../services/slark';
import {
  Backup,
  currentBackup,
  upsertBackup,
} from '../../../common/mmkv/backup';
import { currentUnlockPasswordSetting } from '../../../services/unlockPassword';
import { xor_str } from '../../../common/sqlite/dao/utils';

export interface FamilyMember {
  id: number;
  userID: number;
  email: string;
  familyID: string;
  joinedAt: number;
  isAdmin: boolean;
  isMe: boolean;
  userPublicKey: string;
}

export interface FamilyInfo {
  hasFamily: boolean;
  description: string;
  familyMembers: FamilyMember[];
  sharedNumbers: number;
}

interface FamilyInvitation {
  hasInvitation: boolean;
  id: number;
  invitedBy: string;
  invitedAt: number;
  encryptedFamilyKey: string;
}

type FamilyStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function FamilyStackScreen({
  navigation,
}: FamilyStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { setLoading } = React.useContext(BackdropContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const [familyInfo, setFamilyInfo] = React.useState<null | FamilyInfo>(null);
  const [familyInvitation, setFamilyInvitation] =
    React.useState<null | FamilyInvitation>(null);

  const navHeaderRight = React.useCallback(
    (familyInfo: null | FamilyInfo) => () => {
      if (!slarkInfo) {
        return <></>;
      }
      if (!familyInfo) {
        return <></>;
      }
      if (!familyInfo.familyMembers || familyInfo.familyMembers.length === 0) {
        return <></>;
      }
      let valid = false;
      for (let i = 0; i < familyInfo.familyMembers.length; i++) {
        if (familyInfo.familyMembers[i].userID === slarkInfo.userID) {
          valid = true;
          if (!familyInfo.familyMembers[i].isAdmin) {
            return <></>;
          }
        }
      }
      if (!valid) {
        return <></>;
      }
      const onPressInvite = () => {
        navigation.navigate('FamilyMessageStack');
      };
      return (
        <TouchableOpacity style={styles.headIcon} onPress={onPressInvite}>
          <Icon type="antdesign" name="message1" color={theme.colors.primary} />
        </TouchableOpacity>
      );
    },
    [slarkInfo, styles, familyInfo],
  );

  const [isAdmin, setIsAdmin] = React.useState(false);
  const handleResponseData = React.useCallback(
    (slarkInfo: SlarkInfo, data: any) => {
      if (data.hasFamily) {
        if (data.familyMembers && data.familyMembers.length > 0) {
          // NOTE: update local cache
          let otherMembers: OtherFamilyMember[] = [];
          for (let i = 0; i < data.familyMembers.length; i++) {
            if (data.familyMembers[i].userID === slarkInfo.userID) {
              data.familyMembers[i].isMe = true;
              setIsAdmin(data.familyMembers[i].isAdmin);
            } else {
              data.familyMembers[i].isMe = false;
              otherMembers.push({
                userID: data.familyMembers[i].userID,
                email: data.familyMembers[i].email,
              });
            }
          }
          updateOtherFamilyMembers(slarkInfo.userID, { list: otherMembers });
        }
      }
    },
    [],
  );

  const checkFamilyInvitation = async () => {
    try {
      const respData = await post('/pswds/checkFamilyInvitation/v1');
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      if (respData.data.hasInvitation) {
        setFamilyInvitation(respData.data);
      } else {
        setFamilyInvitation(null);
      }
    } catch (error) {
      throw error;
    }
  };

  const onLoad = React.useCallback(async () => {
    try {
      // （1）检查用户登录状态
      if (!slarkInfo) {
        navigation.goBack();
        return;
      }
      setLoading(true);
      // （2）从远程获取家庭信息
      let respData = await post('/pswds/getFamilyInfo/v1');
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      handleResponseData(slarkInfo, respData.data);
      setFamilyInfo(respData.data);
      // 如果当前用户没有家庭，再检查其有没有家庭邀请
      if (!respData.data.hasFamily) {
        await checkFamilyInvitation();
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

  React.useEffect(() => {
    // 导航栏：管理员功能
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
      title: t('settings.family.label'),
      headerRight: navHeaderRight(familyInfo),
    });
  }, [familyInfo, isAdmin]);

  const [createFamilyVisible, setCreateFamilyVisible] = React.useState(false);

  const onPressCreateFamily = () => {
    setCreateFamilyVisible(true);
  };

  const validateEncryptedFamilyKey = (
    encryptedFK: string,
    password: string,
  ): boolean => {
    try {
      if (familyInvitation) {
        const familyKey = decryptByUserPrivateKey(
          Buffer.from(encryptedFK, 'hex'),
          password,
        ).toString();
        if (familyKey) {
          return true;
        } else {
          return false;
        }
      } else {
        return false;
      }
    } catch (error) {
      return false;
    }
  };

  const checkPreludes = React.useCallback(
    async (action: boolean): Promise<Backup> => {
      try {
        if (!slarkInfo) {
          throw t('app.toast.notSignedIn');
        }
        if (!familyInvitation || familyInvitation.id <= 0) {
          throw t('app.toast.parameterError');
        }
        const backup: Backup | null = currentBackup();
        if (!backup || !backup.userPublicKey) {
          throw 'no user backup record';
        }
        // 检查解锁密码是否是在处理家庭请求之前被修改了
        if (
          action &&
          !validateEncryptedFamilyKey(
            familyInvitation.encryptedFamilyKey,
            password,
          )
        ) {
          throw 'unlock password has changed before process family invitation. Please reject the invitation';
        }
        return backup;
      } catch (error) {
        throw error;
      }
    },
    [slarkInfo, password, familyInvitation],
  );

  const handleProcessResult = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password: string,
      action: boolean,
      backup: Backup,
    ) => {
      if (action) {
        // 情况一：接受邀请
        // a. 将encryptedFamilyKey存入数据库
        upsertBackup({
          ...backup,
          encryptedFamilyKey: familyInvitation!.encryptedFamilyKey,
        });
        // b. 下载家庭共享数据
        const respData = await downloadSharedData(slarkInfo, password);
        if (respData.code !== 0) {
          setLoading(false);
          setError(respData.message, t('app.toast.requestError'));
          return;
        }
        setLoading(false);
        // c. 清理掉invitation状态信息
        setFamilyInvitation(null);
        // d. 拉取家庭信息
        setTimeout(() => {
          onLoad();
        }, 0);
        return;
      } else {
        // 情况二：拒绝邀请
        // 清理掉当前invitation状态信息
        setLoading(false);
        setFamilyInvitation(null);
        // 检查有没有其他邀请
        await checkFamilyInvitation();
      }
    },
    [familyInvitation],
  );

  const processInvitation = React.useCallback(
    (action: boolean) => async () => {
      try {
        // 1. handle parameters
        const backup = await checkPreludes(action);
        setLoading(true);
        // 2. post request
        let respData = await post('/pswds/processFamilyInvitation/v1', {
          id: familyInvitation!.id,
          accept: action,
        });
        if (respData.code !== 0) {
          setLoading(false);
          setError(respData.message, t('app.toast.requestError'));
          return;
        }
        // 3. handle results
        await handleProcessResult(slarkInfo!, password, action, backup);
      } catch (error) {
        setLoading(false);
        setError(error as string, t('app.toast.internalError'));
      }
    },
    [slarkInfo, password, familyInvitation],
  );

  const { showActionSheetWithOptions } = useActionSheet();

  const handleAdminAuthority = async (
    slarkInfo: SlarkInfo,
    familyInfo: FamilyInfo,
    item: FamilyMember,
  ) => {
    try {
      setLoading(true);
      let respData = await post('/pswds/handleAdminAuthority/v1', {
        userID: item.userID,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      if (item.isAdmin) {
        item.isAdmin = false;
        if (slarkInfo.userID === item.userID) {
          setIsAdmin(false);
        }
      } else {
        item.isAdmin = true;
      }
      for (let i = 0; i < familyInfo.familyMembers!.length; i++) {
        if (familyInfo.familyMembers![i].userID === item.userID) {
          familyInfo.familyMembers![i] = item;
          break;
        }
      }
      setFamilyInfo(familyInfo);
      setLoading(false);
      setSuccess(t('app.toast.success'));
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const removeFamilyMember = async (
    slarkInfo: SlarkInfo,
    familyInfo: FamilyInfo,
    item: FamilyMember,
  ) => {
    try {
      setLoading(true);
      let respData = await post('/pswds/removeFamilyMember/v1', {
        userID: item.userID,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      let newFamilyMembers = [];
      for (let i = 0; i < familyInfo.familyMembers!.length; i++) {
        if (familyInfo.familyMembers![i].userID !== item.userID) {
          newFamilyMembers.push(familyInfo.familyMembers![i]);
        }
      }
      familyInfo.familyMembers = newFamilyMembers;
      // refresh data
      deleteOtherFamilyMembers(slarkInfo.userID, item.userID);
      const curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      const xoredUserID = xor_str(
        curSetting!.passwordHash,
        item.userID.toString(),
      );
      await deleteSharedDataMembersByUserIDAsync(xoredUserID);
      await deletePasswordsByUserIDAsync(xoredUserID);
      await deleteRecordsByUserIDAsync(xoredUserID);
      setFamilyInfo(familyInfo);
      setLoading(false);
      setSuccess(t('app.toast.success'));
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const actionSheetCallback =
    (options: any[], item: FamilyMember) =>
    async (selectedIndex: number | undefined) => {
      if (!slarkInfo) {
        return;
      }
      if (!familyInfo) {
        return;
      }
      if (!familyInfo.familyMembers || familyInfo.familyMembers.length == 0) {
        return;
      }
      switch (options[selectedIndex as number]) {
        case t('familyOperation.overlay.operations.1'):
        case t('familyOperation.overlay.operations.2'):
          await handleAdminAuthority(slarkInfo, familyInfo, item);
          break;
        case t('familyOperation.overlay.operations.3'):
          await removeFamilyMember(slarkInfo, familyInfo, item);
          break;
        default:
          break;
      }
    };

  const onPressThreeDot = (item: FamilyMember) => {
    let options = [];
    let cancelButtonIndex = -1;
    if (item.isAdmin) {
      options.push(t('familyOperation.overlay.operations.2'));
      cancelButtonIndex = 1;
    } else {
      options.push(t('familyOperation.overlay.operations.1'));
      options.push(t('familyOperation.overlay.operations.3'));
      cancelButtonIndex = 2;
    }
    options.push(t('familyOperation.overlay.operations.4'));
    showActionSheetWithOptions(
      {
        options,
        cancelButtonIndex,
      },
      actionSheetCallback(options, item),
    );
  };

  const doLeaveFamily = async () => {
    if (!slarkInfo) {
      return;
    }
    // 1. post
    setLoading(true);
    let respData = await post('/pswds/leaveFamily/v1');
    if (respData.code !== 0) {
      setLoading(false);
      setError(respData.message, t('app.toast.requestError'));
      return;
    }
    setLoading(false);
    setSuccess(t('app.toast.success'));
    // 2. reload
    onLoad();
    // 3. refresh all data
    const curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
    const xoredUserID = xor_str(
      curSetting!.passwordHash,
      slarkInfo.userID.toString(),
    );
    await clearSharedDataMembersAsync();
    await deletePasswordsFromFamilyAsync(curSetting!.passwordHash, xoredUserID);
    await cancelSharingPasswordsByUserIDAsync(xoredUserID);
    await deleteRecordsFromFamilyAsync(curSetting!.passwordHash, xoredUserID);
    await cancelSharingRecordsByUserIDAsync(xoredUserID);
    upsertBackup({
      ...currentBackup()!,
      encryptedFamilyKey: null,
    });
  };

  const onPressLeaveFamily = () => {
    Alert.alert('', t('settings.family.leaveFamilyPrompt'), [
      {
        text: t('app.alert.cancelBtn'),
        style: 'cancel',
      },
      {
        text: t('app.alert.confirmBtn'),
        onPress: doLeaveFamily,
      },
    ]);
  };

  const doRemoveFamily = async () => {
    if (!slarkInfo) {
      return;
    }
    // 1. post
    setLoading(true);
    let respData = await post('/pswds/removeFamily/v1');
    setLoading(false);
    if (respData.code !== 0) {
      setError(respData.message, t('app.toast.requestError'));
      return;
    }
    setSuccess(t('app.toast.success'));
    // 2. reload
    onLoad();
    // 3. refresh all data
    const curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
    const xoredUserID = xor_str(
      curSetting!.passwordHash,
      slarkInfo.userID.toString(),
    );
    await clearSharedDataMembersAsync();
    await deletePasswordsFromFamilyAsync(curSetting!.passwordHash, xoredUserID);
    await cancelSharingPasswordsByUserIDAsync(xoredUserID);
    await deleteRecordsFromFamilyAsync(curSetting!.passwordHash, xoredUserID);
    await cancelSharingRecordsByUserIDAsync(xoredUserID);
    upsertBackup({
      ...currentBackup()!,
      encryptedFamilyKey: null,
    });
  };

  const onPressRemoveFamily = () => {
    Alert.alert('', t('settings.family.removeFamilyPrompt'), [
      {
        text: t('app.alert.cancelBtn'),
        style: 'cancel',
      },
      {
        text: t('app.alert.confirmBtn'),
        onPress: doRemoveFamily,
      },
    ]);
  };

  const onPressInviteMember = () => {
    navigation.navigate('InviteFamilyMemberStack');
  };

  const onPressBackupUnlockPassword = () => {
    if (!familyInfo) {
      return;
    }
    navigation.navigate('BackupUnlockPasswordStack', {
      familyMembers: familyInfo.familyMembers,
    });
  };

  return (
    <SafeAreaView style={styles.container}>
      {familyInfo && (
        <>
          <ScrollView>
            {!familyInfo.hasFamily && (
              <>
                <View style={styles.row}>
                  <Button
                    title={t('settings.family.createFamilyBtn')}
                    containerStyle={styles.createBtn}
                    titleStyle={styles.createBtnTitle}
                    size="lg"
                    radius={8}
                    onPress={onPressCreateFamily}
                  />
                </View>
                {familyInvitation && (
                  <Card containerStyle={styles.card}>
                    <Card.Title>
                      {t('settings.family.familyInvitation.title')}
                    </Card.Title>
                    <Card.Divider />
                    <View style={styles.row}>
                      <Text>
                        {t('settings.family.familyInvitation.invitedBy') +
                          familyInvitation.invitedBy}
                      </Text>
                    </View>
                    <View style={styles.row}>
                      <Text>
                        {t('settings.family.familyInvitation.invitedAt') +
                          moment(familyInvitation.invitedAt * 1000)
                            .local()
                            .format('YYYY-MM-DD HH:mm:ss')}
                      </Text>
                    </View>
                    <View style={[styles.row, styles.invitationBtnRow]}>
                      <Button
                        title={t('settings.family.familyInvitation.rejectBtn')}
                        containerStyle={styles.invitationBtn}
                        titleStyle={styles.invitationBtnTitle}
                        radius={8}
                        color={theme.colors.error}
                        onPress={processInvitation(false)}
                      />
                      <Button
                        title={t('settings.family.familyInvitation.acceptBtn')}
                        containerStyle={styles.invitationBtn}
                        titleStyle={styles.invitationBtnTitle}
                        radius={8}
                        color={theme.colors.primary}
                        onPress={processInvitation(true)}
                      />
                    </View>
                  </Card>
                )}
              </>
            )}
            {familyInfo.hasFamily && (
              <>
                <View style={styles.row}>
                  <Text style={styles.label}>
                    {t('settings.family.familyName') + familyInfo.description}
                  </Text>
                </View>
              </>
            )}
            {familyInfo.familyMembers &&
              familyInfo.familyMembers.map(item => (
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
                    <ListItem.Title style={item.isMe ? styles.isMe : {}}>
                      {item.email}
                    </ListItem.Title>
                    <ListItem.Subtitle
                      style={[
                        styles.itemSubtitle,
                        item.isMe ? styles.isMe : {},
                      ]}>
                      {t('settings.family.joinedAt') +
                        moment(item.joinedAt * 1000)
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss')}
                    </ListItem.Subtitle>
                  </ListItem.Content>
                  {item.isAdmin && (
                    <Text style={styles.itemAdmin}>
                      {t('settings.family.isAdmin')}
                    </Text>
                  )}
                  {isAdmin && (
                    <TouchableOpacity
                      style={{
                        opacity: item.userID != slarkInfo?.userID ? 1 : 0,
                      }}
                      onPress={() => {
                        onPressThreeDot(item);
                      }}>
                      <Icon type="entypo" name="dots-three-horizontal" />
                    </TouchableOpacity>
                  )}
                </ListItem>
              ))}
          </ScrollView>
          {/* 有家庭、非管理员 */}
          {familyInfo.hasFamily && !isAdmin && (
            <View style={[styles.row, styles.bottomRow]}>
              <View style={styles.half}>
                <Button
                  title={t('settings.family.leaveFamilyBtn')}
                  color={theme.colors.error}
                  containerStyle={styles.bottomBtn}
                  titleStyle={styles.bottomBtnTitle}
                  radius={8}
                  onPress={onPressLeaveFamily}
                />
              </View>
              <View style={styles.half}>
                <Button
                  title={t('settings.family.backupUnlockPasswordBtn')}
                  containerStyle={styles.bottomBtn}
                  titleStyle={styles.bottomBtnTitle}
                  radius={8}
                  onPress={onPressBackupUnlockPassword}
                />
              </View>
            </View>
          )}
          {/* 有家庭、管理员、家庭成员只剩一个 */}
          {familyInfo.hasFamily &&
            isAdmin &&
            familyInfo.familyMembers?.length === 1 && (
              <View style={[styles.row, styles.bottomRow]}>
                <View style={styles.thirty}>
                  <Button
                    title={t('settings.family.removeFamilyBtn')}
                    color={theme.colors.error}
                    containerStyle={styles.bottomBtn}
                    titleStyle={styles.bottomBtnTitle}
                    radius={8}
                    onPress={onPressRemoveFamily}
                  />
                </View>
                <View style={styles.thirty}>
                  <Button
                    title={t('settings.family.inviteMemberBtn')}
                    containerStyle={styles.bottomBtn}
                    titleStyle={styles.bottomBtnTitle}
                    radius={8}
                    onPress={onPressInviteMember}
                  />
                </View>
                <View style={styles.forty}>
                  <Button
                    title={t('settings.family.backupUnlockPasswordBtn')}
                    containerStyle={styles.bottomBtn}
                    titleStyle={styles.bottomBtnTitle}
                    radius={8}
                    onPress={onPressBackupUnlockPassword}
                  />
                </View>
              </View>
            )}
          {/* 有家庭、管理员、家庭成员超过一个 */}
          {familyInfo.hasFamily &&
            isAdmin &&
            familyInfo.familyMembers?.length !== 1 && (
              <View style={[styles.row, styles.bottomRow]}>
                <View style={styles.half}>
                  <Button
                    title={t('settings.family.inviteMemberBtn')}
                    containerStyle={styles.bottomBtn}
                    titleStyle={styles.bottomBtnTitle}
                    radius={8}
                    onPress={onPressInviteMember}
                  />
                </View>
                <View style={styles.half}>
                  <Button
                    title={t('settings.family.backupUnlockPasswordBtn')}
                    containerStyle={styles.bottomBtn}
                    titleStyle={styles.bottomBtnTitle}
                    radius={8}
                    onPress={onPressBackupUnlockPassword}
                  />
                </View>
              </View>
            )}
        </>
      )}
      <CreateFamilyOverlay
        visible={createFamilyVisible}
        setVisible={setCreateFamilyVisible}
        successCallback={onLoad}
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
  avatarContainer: { backgroundColor: theme.colors.surface },
  itemSubtitle: { fontSize: 12 },
  card: { borderRadius: 8 },
  invitationBtnRow: { justifyContent: 'flex-end', marginVertical: 8 },
  invitationBtn: { height: 40, marginHorizontal: 8 },
  invitationBtnTitle: { fontSize: 12 },
  bottomBtn: {
    marginVertical: 8,
    width: '100%',
  },
  bottomBtnTitle: {
    fontSize: 10,
    fontWeight: 'normal',
  },
  itemAdmin: {
    fontSize: 12,
    backgroundColor: theme.colors.primary,
    padding: 4,
    borderRadius: 4,
  },
  bottomRow: { bottom: 10, position: 'absolute' },
  isMe: { color: theme.colors.green0 },
  half: { width: '50%', padding: 1 },
  thirty: { width: '30%', padding: 1 },
  forty: { width: '40%', padding: 1 },
}));

export default FamilyStackScreen;
