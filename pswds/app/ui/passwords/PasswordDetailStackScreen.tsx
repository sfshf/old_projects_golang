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
import {
  makeStyles,
  Text,
  Button,
  Input,
  Icon,
  Card,
  Avatar,
} from '@rneui/themed';
import { useTheme } from '@rneui/themed';
import Clipboard from '@react-native-clipboard/clipboard';
import { initPassword } from './NewPasswordStackScreen';
import { Alert, Keyboard, TouchableOpacity, View } from 'react-native';
import { iconBgColors, Password } from '../../common/sqlite/schema';
import moment from 'moment';
import InputToolbar from '../../components/InputToolbar';
import { useFocusEffect } from '@react-navigation/native';
import { z } from 'zod';
import { SafeAreaView, ScrollView, KeyboardAvoidingView } from 'react-native';
import FieldTypeOverlay, {
  FieldCard,
  OtherField,
} from '../../components/FieldTypeOverlay';
import FamilyShareOverlay from '../../components/FamilyShareOverlay';
import {
  getFrequentlyUsedUsernamesByUserIDAsync,
  updatePasswordAsync,
  addPasswordUseByDataIDAsync,
  deletePasswordByDataIDAsync,
  getPasswordByDataIDAsync,
  xorXoredPassword,
  XoredPassword,
  xorPassword,
} from '../../common/sqlite/dao/password';
import {
  ResponseCode_DataPullAhead,
  ResponseCode_NotFound,
} from '../../common/http';
import { SlarkInfo } from '../../services/slark';
import { downloadData, updateBackupState } from '../../services/backup';
import { getFamilyKey, getFamilyMemberEmail } from '../../services/family';
import { UnlockPasswordContext } from '../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../contexts/slark';
import { BackdropContext } from '../../contexts/backdrop';
import { SnackbarContext } from '../../contexts/snackbar';
import { post } from '../../common/http/post';
import {
  encryptByUnlockPassword,
  encryptByXchacha20poly1305,
} from '../../services/cipher';
import RecordOperationTooltip from '../../components/RecordOperationTooltip';
import {
  BarcodeScanningResult,
  CameraView,
  useCameraPermissions,
} from 'expo-camera';
import QRScannerView from '../../components/QRScanner';
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';
import { aes256GCM_secp256k1Encrypt } from '../../common/cipher';
import { xor_str } from '../../common/sqlite/dao/utils';
import { keccak_256 } from '@noble/hashes/sha3';

type PasswordDetailStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'PasswordDetailStack'
>;

function PasswordDetailStackScreen({
  navigation,
  route,
}: PasswordDetailStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { setLoading } = React.useContext(BackdropContext);
  const { setSuccess, setError, setWarning } =
    React.useContext(SnackbarContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const [entity, setEntity] = React.useState<Password>(initPassword);
  const [record, setRecord] = React.useState<Password>(initPassword);
  const [otherFields, setOtherFields] = React.useState<null | OtherField[]>(
    null,
  );
  const [frequentlyUsedUsernames, setFrequentlyUsedUsernames] = React.useState<
    string[]
  >([]);
  const findFrequentlyUsedUsernames = React.useCallback(async () => {
    try {
      let userID = slarkInfo ? slarkInfo.userID : -1;
      const curSetting = currentUnlockPasswordSetting(userID);
      const items: any[] = await getFrequentlyUsedUsernamesByUserIDAsync(
        xor_str(curSetting!.passwordHash, userID.toString()),
      );
      const usernames: string[] = [];
      items.map(item => {
        usernames.push(xor_str(curSetting!.passwordHash, item));
      });
      setFrequentlyUsedUsernames(usernames);
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo]);
  const onLoad = React.useCallback(async () => {
    try {
      if (!route.params.dataID) {
        navigation.goBack();
        return;
      }
      const curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      const xoredDataID = xor_str(
        curSetting!.passwordHash,
        route.params.dataID,
      );
      // 1. 加载数据
      const result: XoredPassword | null = await getPasswordByDataIDAsync(
        xoredDataID,
      );
      if (!result) {
        setError(t('app.toast.notFoundError'), t('app.toast.error'));
        return;
      }
      const xoredResult = xorXoredPassword(curSetting!.passwordHash, result);
      setEntity(xoredResult);
      // 2. 加载常用usernames
      await findFrequentlyUsedUsernames();
      // 3. entity to record
      setRecord({
        ...xoredResult,
      });
      // others
      if (xoredResult.others) {
        setOtherFields(JSON.parse(xoredResult.others));
      } else {
        setOtherFields(null);
      }
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  }, [route.params.dataID, slarkInfo]);
  const [keyboardDidShow, setKeyboardDisShow] = React.useState(false);
  const [editUsername, setEditUsername] = React.useState(false);

  const genReqData = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password: string,
      updatedAt: number,
      newEntity: Password,
    ): Promise<any> => {
      try {
        const passwordHash = Buffer.from(keccak_256(password)).toString('hex');
        const xoredEntity = xorPassword(passwordHash, newEntity);
        const content = JSON.stringify(xoredEntity);
        // （1）解密出familyKey，加密数据
        let sharedData = '';
        if (
          newEntity.sharedAt &&
          newEntity.sharedAt > 0 &&
          newEntity.userID == slarkInfo.userID
        ) {
          sharedData = encryptByXchacha20poly1305(
            await getFamilyKey(password),
            content,
          );
        }
        return {
          updatedAt: updatedAt,
          dataID: newEntity.dataID,
          content: encryptByUnlockPassword(
            // 整体加密
            password,
            content,
          ),
          sharedData,
        };
      } catch (error) {
        throw error;
      }
    },
    [slarkInfo, record],
  );

  const updatePasswordRecord = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password: string,
      newEntity: Password,
    ): Promise<any> => {
      try {
        if (!password) {
          throw t('app.toast.emptyUnlockPassword');
        }
        const updatedAt = newEntity.updatedAt;
        // （2）post
        const respData = await post(
          '/pswds/updatePasswordRecord/v1',
          await genReqData(slarkInfo, password, updatedAt, newEntity),
        );
        if (!respData || respData.code !== 0) {
          return respData;
        }
        updateBackupState(slarkInfo.userID, {
          updatedAt: updatedAt,
        });
        return respData;
      } catch (error) {
        throw error;
      }
    },
    [slarkInfo, password, record],
  );

  useFocusEffect(
    React.useCallback(() => {
      onLoad();
    }, [onLoad]),
  );

  Keyboard.addListener('keyboardDidHide', () => {
    setKeyboardDisShow(false);
  });

  Keyboard.addListener('keyboardDidShow', () => {
    setKeyboardDisShow(true);
  });

  const [copyUsername, setCopyUsername] = React.useState<boolean>(false);

  const onCopyUsername = () => {
    if (!record.password) {
      return;
    }
    Clipboard.setString(record.username);
    setCopyUsername(true);
    setRecord({
      ...record,
      username: t('passwords.passwordDetail.copyPrompt'),
    });
    setTimeout(() => {
      setCopyUsername(false);
      setRecord({
        ...record,
        username: record.username,
      });
    }, 1000);
  };

  const [copyPassword, setCopyPassword] = React.useState<boolean>(false);

  const onCopyPassword = React.useCallback(async () => {
    try {
      if (!record.password) {
        return;
      }
      const curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      Clipboard.setString(record.password);
      setCopyPassword(true);
      setRecord({
        ...record,
        password: t('passwords.passwordDetail.copyPrompt'),
      });
      setTimeout(() => {
        setCopyPassword(false);
        setRecord({
          ...record,
          password: record.password,
        });
      }, 1000);
      // copy password, is an 'use' operation
      let newEntity = { ...entity };
      const nowTS = moment().unix();
      newEntity.updatedAt = nowTS;
      newEntity.usedAt = nowTS;
      newEntity.usedCount = Number(entity.usedCount) + 1;
      // 数据同步到后端
      if (slarkInfo) {
        await updatePasswordRecord(slarkInfo, password, newEntity);
      }
      await addPasswordUseByDataIDAsync(
        curSetting!.passwordHash,
        newEntity.dataID,
      );
      setEntity(newEntity);
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  }, [password, record, slarkInfo, entity, otherFields]);

  const [edit, setEdit] = React.useState(false);

  const navHeaderLeft = React.useCallback(() => {
    if (!edit) {
      return (
        <>
          {entity.website && (
            <Avatar
              source={
                entity.website ? { uri: entity.website + '/favicon.ico' } : {}
              }
              imageProps={{
                style: { borderRadius: 8 },
                PlaceholderContent: (
                  <Avatar
                    title={entity.title.substring(0, 2)}
                    containerStyle={[
                      styles.itemAvatar,
                      {
                        backgroundColor:
                          iconBgColors[
                            entity.iconBgColor ? entity.iconBgColor : 0
                          ],
                      },
                    ]}
                  />
                ),
              }}
            />
          )}
          {!entity.website && (
            <Avatar
              title={entity.title.substring(0, 2)}
              containerStyle={[
                styles.itemAvatar,
                {
                  backgroundColor:
                    iconBgColors[entity.iconBgColor ? entity.iconBgColor : 0],
                },
              ]}
            />
          )}
        </>
      );
    }
    return <></>;
  }, [edit, entity, styles]);

  const [familyShareVisible, setFamilyShareVisible] = React.useState(false);
  const [openToolTip, setOpenToolTip] = React.useState(false);
  const navHeaderRight = React.useCallback(() => {
    return (
      <>
        {!edit && (
          <TouchableOpacity
            style={styles.headIcon}
            onPressOut={() => {
              setOpenToolTip(!openToolTip);
            }}>
            <Icon type="entypo" name="dots-three-horizontal" />
          </TouchableOpacity>
        )}
      </>
    );
  }, [edit, styles]);
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
      headerLeft: navHeaderLeft,
      headerRight:
        entity.userID === -1 || entity.userID === slarkInfo?.userID
          ? navHeaderRight
          : undefined,
    });
  }, [navigation, navHeaderLeft, navHeaderRight]);

  const [visible, setVisible] = React.useState(false);

  const onPressOpenAddField = () => {
    setVisible(true);
  };

  const websiteSchema = z
    .string()
    .regex(/^(https:\/\/)?(www\.)?[a-zA-Z0-9-]+\.[a-zA-Z]+(\/[^\s]*)?$/, {
      message: t('passwords.passwordDetail.toast.invalidWebsite'),
    });

  const toptSchema = z
    .string()
    .regex(
      /^otpauth:\/\/([ht]otp)\/(?:[a-zA-Z0-9%]+:)?([^?]+)\?secret=([0-9A-Za-z]+)(?:.*(?:<?counter=)([0-9]+))?/,
      {
        message: t('passwords.newPassword.toast.invalidTOTP'),
      },
    );

  const validateParameters = React.useCallback((): boolean => {
    if (record.website) {
      const validation = websiteSchema.safeParse(record.website);
      if (!validation.success) {
        setError(validation.error.issues[0].message, t('app.toast.error'));
        return false;
      }
      if (!record.website.startsWith('https://')) {
        record.website = 'https://' + record.website;
      }
      if (record.website.endsWith('/')) {
        record.website = record.website.slice(0, record.website.length - 1);
      }
    }
    return true;
  }, [record]);

  const genNewEntity = React.useCallback((): null | Password => {
    let newEntity = { ...entity };
    newEntity.updatedAt = moment().unix();
    newEntity.title = record.title;
    newEntity.website = record.website;
    newEntity.username = record.username;
    newEntity.password = record.password;
    newEntity.notes = record.notes;
    if (otherFields) {
      for (let i = 0; i < otherFields.length; i++) {
        if (otherFields[i].type === 'one-time password') {
          // validate ont-time password format, refer to https://github.com/google/google-authenticator/wiki/Key-Uri-Format
          const validation = toptSchema.safeParse(otherFields[i].value);
          if (!validation.success) {
            setError(validation.error.issues[0].message, t('app.toast.error'));
            return null;
          }
        }
      }
      newEntity.others = JSON.stringify(otherFields);
    }
    return newEntity;
  }, [entity, record, otherFields]);

  const doEdit = React.useCallback(async () => {
    try {
      const passwordHash = Buffer.from(keccak_256(password)).toString('hex');
      // 1. handle parameters
      const valid = validateParameters();
      if (!valid) {
        return;
      }
      const newEntity = genNewEntity();
      if (!newEntity) {
        return;
      }
      if (slarkInfo) {
        // 2. post request, if signed in
        setLoading(true);
        const respData = await updatePasswordRecord(
          slarkInfo,
          password,
          newEntity,
        );
        if (respData.code !== 0) {
          if (
            respData.code === ResponseCode_DataPullAhead ||
            respData.code === ResponseCode_NotFound
          ) {
            // need to sync data, if local data fall behind
            const downloadRespData = await downloadData(slarkInfo, password);
            if (downloadRespData.code === 0) {
              setLoading(false);
              // 数据同步成功，警告用户重试操作
              setWarning(
                t('app.toast.afterSyncBackupData'),
                t('app.toast.error'),
              );
            }
            return;
          } else {
            setLoading(false);
            setError(respData.message, t('app.toast.error'));
            return;
          }
        }
      }
      // 3. handle results
      // 更新到本地数据库
      await updatePasswordAsync(xorPassword(passwordHash, newEntity));
      setEntity(newEntity);
      setLoading(false);
      setEdit(false);
      setSuccess(t('app.toast.success'));
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
      return;
    }
  }, [slarkInfo, entity, record, otherFields]);

  const onPressCommit = () => {
    if (entity.userID > 0) {
      if (!slarkInfo || entity.userID != slarkInfo.userID) {
        setEdit(false);
        return;
      }
    }
    Alert.alert(
      '',
      t('passwords.passwordDetail.editPrompt') + record.title + ' ?',
      [
        {
          text: t('app.alert.cancelBtn'),
          style: 'cancel',
        },
        {
          text: t('app.alert.confirmBtn'),
          onPress: doEdit,
        },
      ],
    );
  };

  const onPressCancel = () => {
    Keyboard.dismiss();
    onLoad();
    setEdit(false);
  };

  const doDelete = React.useCallback(async () => {
    try {
      if (slarkInfo) {
        const updatedAt = moment().unix();
        setLoading(true);
        const respData = await post('/pswds/deletePasswordRecord/v1', {
          updatedAt,
          dataID: entity.dataID,
        });
        if (respData.code !== 0) {
          if (
            respData.code === ResponseCode_DataPullAhead ||
            respData.code === ResponseCode_NotFound
          ) {
            const downloadRespData = await downloadData(slarkInfo, password);
            if (downloadRespData.code === 0) {
              setLoading(false);
              setWarning(
                t('app.toast.afterSyncBackupData'),
                t('app.toast.error'),
              );
            }
            return;
          } else {
            setLoading(false);
            setError(respData.message, t('app.toast.error'));
            return;
          }
        }
        updateBackupState(slarkInfo.userID, {
          updatedAt: updatedAt,
        });
      }
      await deletePasswordByDataIDAsync(entity.dataID);
      setLoading(false);
      setSuccess(t('app.toast.success'));
      navigation.popTo('HomeStack', { screen: 'Passwords' });
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, entity]);

  const onChangeTitle = (newText: string) => {
    setRecord({ ...record, title: newText });
  };

  const onChangeWebsite = (newText: string) => {
    setRecord({
      ...record,
      website: newText,
    });
  };

  const onFocusUsername = () => {
    setEditUsername(true);
  };

  const onBlurUsername = () => {
    setEditUsername(false);
  };

  const onChangeUsername = (newText: string) => {
    setRecord({
      ...record,
      username: newText,
    });
  };

  const onChangePassword = (newText: string) => {
    setRecord({
      ...record,
      password: newText,
    });
  };

  const onChangeNotes = (newText: string) => {
    setRecord({ ...record, notes: newText });
  };

  const usernameAccessoryViewID = 'usernameAccessoryViewID';

  const onPressEdit = () => {
    setOpenToolTip(false);
    setEdit(true);
  };
  const onPressShare = async () => {
    setOpenToolTip(false);
    setFamilyShareVisible(true);
  };
  const onPressTransfer = () => {
    setOpenToolTip(false);
    navigation.navigate('DataQRStack', { entity: entity });
  };
  const [qrscan, setQrscan] = React.useState(false);
  const [permission, requestPermission] = useCameraPermissions();
  const cameraViewRef = React.useRef<null | CameraView>(null);
  const onPressTransfer2 = () => {
    setOpenToolTip(false);
    setQrscan(true);
    if (permission) {
      if (!permission.granted) {
        requestPermission();
      }
    }
  };

  const onBarcodeScanned = React.useCallback(
    async (result: BarcodeScanningResult) => {
      if (!slarkInfo) {
        return;
      }
      let lockSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      if (!lockSetting || lockSetting.passwordHash === '') {
        return;
      }
      cameraViewRef.current && cameraViewRef.current.pausePreview();
      try {
        const passwordHash = Buffer.from(keccak_256(password)).toString('hex');
        // 1. parse data from qrcode
        const data = JSON.parse(result.data);
        // 2. encrypt password
        const pubkey = new Uint8Array(Buffer.from(data.publicKey, 'hex'));
        const cipherbytes = aes256GCM_secp256k1Encrypt(
          pubkey,
          new Uint8Array(Buffer.from(JSON.stringify(entity), 'utf-8')),
        );
        // 3. upload
        setLoading(true);
        const respData = await post('/pswds/uploadAirdropData/v1', {
          uuid: data.uuid,
          cipherText: Buffer.from(cipherbytes).toString('hex'),
        });
        if (respData.code !== 0) {
          setLoading(false);
          cameraViewRef.current && cameraViewRef.current.resumePreview();
          setError(respData.message, t('app.toast.requestError'));
          return;
        }
        // 4. close camera
        setQrscan(false);
        cameraViewRef.current && cameraViewRef.current.resumePreview();
        // 5. transfer password, is an 'use' operation
        let newEntity = { ...entity };
        const nowTS = moment().unix();
        newEntity.updatedAt = nowTS;
        newEntity.usedAt = nowTS;
        newEntity.usedCount = Number(entity.usedCount) + 1;
        // 数据同步到后端
        const respData2 = await updatePasswordRecord(
          slarkInfo,
          password,
          newEntity,
        );
        if (respData2.code !== 0) {
          if (respData2.code === ResponseCode_DataPullAhead) {
            // need to sync data
            const downloadRespData = await downloadData(slarkInfo, password);
            if (downloadRespData.code === 0) {
              // 数据同步成功，警告用户重试操作
              setLoading(false);
              setWarning(
                t('app.toast.afterSyncBackupData'),
                t('app.toast.error'),
              );
              return;
            }
            setLoading(false);
            return;
          } else {
            setLoading(false);
            setError(respData2.message, t('app.toast.error'));
            return;
          }
        }
        await addPasswordUseByDataIDAsync(
          passwordHash,
          xor_str(passwordHash, newEntity.dataID),
        );
        setEntity(newEntity);
        setLoading(false);
        setSuccess(t('app.toast.success'));
      } catch (error) {
        setLoading(false);
        setError(error as string, t('app.toast.internalError'));
      } finally {
      }
    },
    [slarkInfo, password, entity, cameraViewRef.current],
  );

  const onPressDelete = () => {
    setOpenToolTip(false);
    Alert.alert(
      '',
      t('passwords.passwordDetail.deletePrompt') + record.title + ' ?',
      [
        {
          text: t('app.alert.cancelBtn'),
          style: 'cancel',
        },
        {
          text: t('app.alert.confirmBtn'),
          onPress: doDelete,
        },
      ],
    );
  };

  return (
    <>
      {!qrscan && (
        <SafeAreaView style={styles.container}>
          {(entity.userID === -1 || entity.userID === slarkInfo?.userID) && (
            <RecordOperationTooltip
              open={openToolTip}
              setOpen={setOpenToolTip}
              entity={entity}
              onPressEdit={onPressEdit}
              onPressShare={onPressShare}
              onPressTransfer={onPressTransfer}
              onPressTransfer2={onPressTransfer2}
              onPressDelete={onPressDelete}
            />
          )}
          <KeyboardAvoidingView>
            <ScrollView>
              <Card containerStyle={styles.card}>
                <View style={styles.row}>
                  <Input
                    autoCapitalize={'none'}
                    multiline
                    readOnly={!edit}
                    labelStyle={styles.inputLabel}
                    inputStyle={styles.inputStyle}
                    placeholder={t('passwords.passwordDetail.titlePlaceholder')}
                    value={record.title}
                    onChangeText={onChangeTitle}
                  />
                </View>
              </Card>
              <Card containerStyle={styles.card}>
                <View style={styles.row}>
                  <Input
                    autoCapitalize={'none'}
                    multiline
                    readOnly={!edit}
                    labelStyle={styles.inputLabel}
                    inputStyle={styles.inputStyle}
                    label={t('passwords.passwordDetail.website')}
                    placeholder={t(
                      'passwords.passwordDetail.websitePlaceholder',
                    )}
                    value={record.website!}
                    onChangeText={onChangeWebsite}
                  />
                </View>
              </Card>
              <Card containerStyle={styles.card}>
                <TouchableOpacity disabled={edit} onPress={onCopyUsername}>
                  <View
                    style={styles.row}
                    pointerEvents={edit ? 'auto' : 'none'}>
                    <Input
                      autoCapitalize={'none'}
                      multiline
                      readOnly={!edit}
                      labelStyle={styles.inputLabel}
                      inputStyle={styles.inputStyle}
                      label={t('passwords.passwordDetail.username')}
                      placeholder={t(
                        'passwords.passwordDetail.usernamePlaceholder',
                      )}
                      value={record.username}
                      rightIcon={
                        !edit &&
                        !copyUsername && (
                          <Icon type="font-awesome-5" name="copy" size={20} />
                        )
                      }
                      onFocus={onFocusUsername}
                      onBlur={onBlurUsername}
                      onChangeText={onChangeUsername}
                    />
                  </View>
                </TouchableOpacity>
              </Card>
              <Card containerStyle={styles.card}>
                <TouchableOpacity disabled={edit} onPress={onCopyPassword}>
                  <View
                    style={styles.row}
                    pointerEvents={edit ? 'auto' : 'none'}>
                    <Input
                      autoCapitalize={'none'}
                      multiline
                      readOnly={!edit}
                      labelStyle={styles.inputLabel}
                      inputStyle={styles.inputStyle}
                      label={t('passwords.passwordDetail.password')}
                      placeholder={t(
                        'passwords.passwordDetail.passwordPlaceholder',
                      )}
                      value={record.password}
                      rightIcon={
                        !edit &&
                        !copyPassword && (
                          <Icon type="font-awesome-5" name="copy" size={20} />
                        )
                      }
                      onChangeText={onChangePassword}
                    />
                  </View>
                </TouchableOpacity>
              </Card>
              {otherFields &&
                otherFields.length > 0 &&
                otherFields.map((item, idx) => {
                  if (item.type === 'one-time password') {
                    return (
                      <FieldCard
                        key={item.key + idx}
                        readonly={!edit}
                        isOTP={true}
                        index={idx}
                        fields={otherFields}
                        setFields={setOtherFields}
                      />
                    );
                  }
                })}
              {(record.notes || edit) && (
                <Card containerStyle={styles.card}>
                  <View style={styles.row}>
                    <Input
                      autoCapitalize={'none'}
                      multiline
                      readOnly={!edit}
                      labelStyle={styles.inputLabel}
                      inputStyle={styles.inputStyle}
                      label={t('passwords.passwordDetail.notes')}
                      placeholder={t(
                        'passwords.passwordDetail.notesPlaceholder',
                      )}
                      numberOfLines={5}
                      value={record.notes!}
                      onChangeText={onChangeNotes}
                    />
                  </View>
                </Card>
              )}
              {otherFields &&
                otherFields.length > 0 &&
                otherFields.map((item, idx) => {
                  if (item.type !== 'one-time password') {
                    return (
                      <FieldCard
                        key={item.key + idx}
                        readonly={!edit}
                        isOTP={false}
                        index={idx}
                        fields={otherFields}
                        setFields={setOtherFields}
                      />
                    );
                  }
                })}
              {edit && (
                <>
                  <Card containerStyle={styles.card}>
                    <Button
                      type="clear"
                      title={t('passwords.newPassword.addFieldBtn')}
                      containerStyle={styles.normalBtn}
                      titleStyle={styles.btnTitle}
                      size="lg"
                      radius={8}
                      onPress={onPressOpenAddField}
                    />
                  </Card>
                  <View style={styles.row}>
                    <Button
                      title={t('app.alert.commitBtn')}
                      containerStyle={styles.commitBtn}
                      titleStyle={styles.btnTitle}
                      color={theme.colors.primary}
                      size="lg"
                      radius={8}
                      onPress={onPressCommit}
                    />
                  </View>
                  <View style={styles.row}>
                    <Button
                      title={t('app.alert.cancelBtn')}
                      containerStyle={styles.cancelBtn}
                      titleStyle={styles.btnTitle}
                      color={theme.colors.error}
                      size="lg"
                      radius={8}
                      onPress={onPressCancel}
                    />
                  </View>
                </>
              )}
              {!edit && (
                <>
                  <View style={[styles.row, styles.marginRow]}>
                    <Text>
                      {t('passwords.passwordDetail.lastEditedLabel') +
                        moment(entity.updatedAt * 1000)
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss')}
                    </Text>
                  </View>
                  {slarkInfo && entity.sharedAt && (
                    <View style={[styles.row, styles.marginRow]}>
                      <Text>
                        {entity.userID === slarkInfo.userID
                          ? t('passwords.passwordDetail.sharingSince') +
                            moment(entity.sharedAt * 1000)
                              .local()
                              .format('YYYY-MM-DD HH:mm:ss')
                          : t('passwords.passwordDetail.sharedBy') +
                            getFamilyMemberEmail(slarkInfo, entity.userID)}
                      </Text>
                    </View>
                  )}
                </>
              )}
              <FieldTypeOverlay
                visible={visible}
                setVisible={setVisible}
                otherFields={otherFields}
                setOtherFields={setOtherFields}
              />
              <FamilyShareOverlay
                visible={familyShareVisible}
                setVisible={setFamilyShareVisible}
                entity={entity}
                setEntity={setEntity}
              />
            </ScrollView>
          </KeyboardAvoidingView>
          {editUsername &&
            keyboardDidShow &&
            frequentlyUsedUsernames.length > 0 && (
              <InputToolbar
                inputAccessoryViewID={usernameAccessoryViewID}
                list={frequentlyUsedUsernames}
                setValue={val => {
                  setRecord({
                    ...record,
                    username: val,
                  });
                }}
              />
            )}
        </SafeAreaView>
      )}
      {qrscan && (
        <QRScannerView
          cameraViewRef={cameraViewRef}
          onBarcodeScanned={onBarcodeScanned}
          setQrscan={setQrscan}
        />
      )}
    </>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  card: { borderRadius: 8 },
  headIcon: {
    padding: 1,
    marginHorizontal: 1,
  },
  itemAvatar: {
    borderRadius: 8,
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  marginRow: {
    marginTop: 20,
  },
  inputLabel: {
    fontSize: 18,
    fontWeight: 'normal',
    color: theme.colors.black,
    marginBottom: 16,
  },
  inputStyle: {
    fontSize: 18,
    fontWeight: 'normal',
  },
  notesHeight: {
    height: 80,
  },
  deleteBtn: { marginVertical: 8, width: '95%' },
  commitBtn: { marginTop: 8, width: '95%' },
  cancelBtn: { width: '95%' },
  btnTitle: { fontSize: 20, fontWeight: 'normal' },
  normalBtn: { width: '100%' },
  itemSharing: {
    fontSize: 8,
    backgroundColor: theme.colors.green0,
    marginHorizontal: 4,
    padding: 4,
    borderRadius: 4,
  },
  itemShared: {
    fontSize: 8,
    backgroundColor: theme.colors.primary,
    marginHorizontal: 4,
    padding: 4,
    borderRadius: 4,
  },
}));

export default PasswordDetailStackScreen;
