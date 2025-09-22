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
import { makeStyles, Text, Button, Input, Icon, Card } from '@rneui/themed';
import { useTheme } from '@rneui/themed';
import {
  Alert,
  Keyboard,
  Platform,
  TouchableOpacity,
  View,
} from 'react-native';
import { iconBgColors, Record } from '../../common/sqlite/schema';
import moment from 'moment';
import { z } from 'zod';
import { SafeAreaView, ScrollView, KeyboardAvoidingView } from 'react-native';
import FieldTypeOverlay, {
  FieldCard,
  OtherField,
} from '../../components/FieldTypeOverlay';
import {
  bankAccountFields,
  creditCardFields,
  DateInputLabel,
  driverLicenseFields,
  identityFields,
  initRecord,
  inputContainerStyle,
  inputEditable,
  passportFields,
} from './NewRecordStackScreen';
import { avatarIcon } from './RecordsTabScreen';
import FamilyShareOverlay from '../../components/FamilyShareOverlay';
import {
  updateRecordAsync,
  deleteRecordByDataIDAsync,
  getRecordByDataIDAsync,
  addRecordUseByDataIDAsync,
  XoredRecord,
  xorXoredRecord,
  xorRecord,
} from '../../common/sqlite/dao/record';
import {
  ResponseCode_DataPullAhead,
  ResponseCode_NotFound,
} from '../../common/http';
import { useFocusEffect } from '@react-navigation/native';
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
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';
import { aes256GCM_secp256k1Encrypt } from '../../common/cipher';
import QRScannerView from '../../components/QRScanner';
import { keccak_256 } from '@noble/hashes/sha3';
import { xor_str } from '../../common/sqlite/dao/utils';

type RecordDetailStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'RecordDetailStack'
>;

function RecordDetailStackScreen({
  navigation,
  route,
}: RecordDetailStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { setLoading } = React.useContext(BackdropContext);
  const { setSuccess, setError, setWarning } =
    React.useContext(SnackbarContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const [entity, setEntity] = React.useState<Record>(initRecord);
  const [transLabel, setTransLabel] = React.useState('');
  const [fields, setFields] = React.useState<string[]>([]);
  const [record, setRecord] = React.useState<Record>(initRecord);
  const [otherFields, setOtherFields] = React.useState<null | OtherField[]>(
    null,
  );
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
      const result: XoredRecord | null = await getRecordByDataIDAsync(
        xoredDataID,
      );
      if (!result) {
        setError(t('app.toast.notFoundError'), t('app.toast.error'));
        return;
      }
      const xoredResult = xorXoredRecord(curSetting!.passwordHash, result);
      setEntity(xoredResult);
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
      if (!route.params) {
        navigation.goBack();
        return;
      }
      // record type
      switch (xoredResult.recordType) {
        case 'identity':
          setTransLabel('identity');
          setFields(identityFields);
          break;
        case 'credit card':
          setTransLabel('creditCard');
          setFields(creditCardFields);
          break;
        case 'bank account':
          setTransLabel('bankAccount');
          setFields(bankAccountFields);
          break;
        case 'driver license':
          setTransLabel('driverLicense');
          setFields(driverLicenseFields);
          break;
        case 'passport':
          setTransLabel('passport');
          setFields(passportFields);
          break;
      }
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  }, [route.params.dataID, slarkInfo]);
  const genReqData = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password: string,
      record: Record,
    ): Promise<any> => {
      const passwordHash = Buffer.from(keccak_256(password)).toString('hex');
      const xoredRecord = xorRecord(passwordHash, record);
      const content = JSON.stringify(xoredRecord);
      let sharedData = '';
      if (
        record.sharedAt &&
        record.sharedAt > 0 &&
        record.userID == slarkInfo.userID
      ) {
        sharedData = encryptByXchacha20poly1305(
          await getFamilyKey(password),
          content,
        );
      }
      return {
        updatedAt: record.updatedAt,
        dataID: record.dataID,
        content: encryptByUnlockPassword(
          // 整体加密
          password,
          content,
        ),
        sharedData,
      };
    },
    [],
  );

  const updateRecord = async (
    slarkInfo: SlarkInfo,
    password: string,
    record: Record,
  ): Promise<boolean> => {
    try {
      setLoading(true);
      const respData = await post(
        '/pswds/updateNonPasswordRecord/v1',
        await genReqData(slarkInfo, password, record),
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
          return false;
        } else {
          setLoading(false);
          setError(respData.message, t('app.toast.error'));
          return false;
        }
      }
      updateBackupState(slarkInfo.userID, {
        updatedAt: record.updatedAt,
      });
      setLoading(false);
      return true;
    } catch (error) {
      throw error;
    }
  };

  const [edit, setEdit] = React.useState(false);

  const navHeaderLeft = React.useCallback(() => {
    const icon = avatarIcon(entity.recordType);
    return (
      <Icon
        size={30}
        type={icon.type}
        name={icon.name}
        containerStyle={[
          styles.itemAvatar,
          {
            backgroundColor:
              iconBgColors[entity.iconBgColor ? entity.iconBgColor : 0],
          },
        ]}
      />
    );
  }, [entity, styles]);

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

  useFocusEffect(
    React.useCallback(() => {
      onLoad();
    }, [onLoad]),
  );

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

  const toptSchema = z
    .string()
    .regex(
      /^otpauth:\/\/([ht]otp)\/(?:[a-zA-Z0-9%]+:)?([^?]+)\?secret=([0-9A-Za-z]+)(?:.*(?:<?counter=)([0-9]+))?/,
      {
        message: t('records.newRecord.toast.invalidTOTP'),
      },
    );
  const pinSchema = z.string().regex(/^\d{4,12}$/, {
    message: t('records.newRecord.toast.pinTypeError'),
  });

  const genNewEntity = React.useCallback(
    (record: Record): null | Record => {
      let entity = { ...record };
      entity.updatedAt = moment().unix();
      if (entity.pin) {
        const validation = pinSchema.safeParse(entity.pin);
        if (!validation.success) {
          setError(validation.error.issues[0].message, t('app.toast.error'));
          return null;
        }
      }
      if (otherFields) {
        for (let i = 0; i < otherFields.length; i++) {
          if (otherFields[i].type === 'one-time password') {
            // validate ont-time password format, refer to https://github.com/google/google-authenticator/wiki/Key-Uri-Format
            const validation = toptSchema.safeParse(otherFields[i].value);
            if (!validation.success) {
              setError(
                validation.error.issues[0].message,
                t('app.toast.error'),
              );
              return null;
            }
          } else if (otherFields[i].type === 'pin') {
            const validation = pinSchema.safeParse(otherFields[i].value);
            if (!validation.success) {
              setError(
                validation.error.issues[0].message,
                t('app.toast.error'),
              );
              return null;
            }
          }
        }
        entity.others = JSON.stringify(otherFields);
      }
      return entity;
    },
    [otherFields],
  );

  const doEdit = React.useCallback(async () => {
    try {
      const curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      // 1. validate parameters
      const newEntity = genNewEntity(record);
      if (!newEntity) {
        return;
      }
      if (slarkInfo) {
        const ok = await updateRecord(slarkInfo, password, newEntity);
        if (!ok) {
          return;
        }
      }
      // 3. handle results
      // 更新到本地数据库
      await updateRecordAsync(xorRecord(curSetting!.passwordHash, newEntity));
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
    // auth check
    if (entity.userID > 0) {
      if (!slarkInfo || entity.userID != slarkInfo.userID) {
        setEdit(false);
        return;
      }
    }
    Alert.alert(
      '',
      t('records.recordDetail.editPrompt') + record.title + ' ?',
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
        const respData = await post('/pswds/deleteNonPasswordRecord/v1', {
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
      await deleteRecordByDataIDAsync(entity.dataID);
      setLoading(false);
      setSuccess(t('app.toast.success'));
      navigation.popTo('HomeStack', { screen: 'Records' });
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, entity]);

  const onChangeText = (field: string) => (newText: string) => {
    if (field === 'pin') {
      newText = newText.replace(/\D/g, '');
    }
    let tmp = { ...record } as any;
    tmp[field] = newText;
    setRecord(tmp);
  };

  const inputLabel = (field: string): string | React.JSX.Element => {
    switch (field) {
      case 'title':
        return '';
      case 'birthDate':
      case 'expiryDate':
      case 'validFrom':
      case 'issuedOn':
        if (edit) {
          return (
            <DateInputLabel
              label={t('records.newRecord.inputs.' + transLabel + '.' + field)}
              btnLabel={t('records.newRecord.datePicker.label')}
              field={field}
              record={record!}
              setRecord={setRecord}
            />
          );
        }
      default:
        return t('records.newRecord.inputs.' + transLabel + '.' + field);
    }
  };
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
      let curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      if (!curSetting || curSetting.passwordHash === '') {
        return;
      }
      cameraViewRef.current && cameraViewRef.current.pausePreview();
      try {
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
        setLoading(false);
        if (respData.code !== 0) {
          cameraViewRef.current && cameraViewRef.current.resumePreview();
          setError(respData.message, t('app.toast.requestError'));
          return;
        }
        setSuccess(t('app.toast.success'));
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
        setLoading(true);
        const ok = await updateRecord(slarkInfo, password, newEntity);
        if (!ok) {
          return;
        }
        await addRecordUseByDataIDAsync(
          curSetting!.passwordHash,
          xor_str(curSetting!.passwordHash, newEntity.dataID),
        );
        setEntity(newEntity);
        setLoading(false);
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
      t('records.recordDetail.deletePrompt') + record.title + ' ?',
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
              {record &&
                fields.map((field: string) => {
                  if (!edit) {
                    if ((record as any)[field]) {
                      return (
                        <Card key={field} containerStyle={styles.card}>
                          <View style={styles.row}>
                            <Input
                              readOnly={true}
                              inputContainerStyle={
                                (record as any)[field]
                                  ? {}
                                  : inputContainerStyle(field)
                              }
                              multiline
                              labelStyle={styles.inputLabel}
                              inputStyle={styles.inputStyle}
                              label={inputLabel(field)}
                              placeholder={t(
                                'records.newRecord.inputs.' +
                                  transLabel +
                                  '.' +
                                  field +
                                  'Placeholder',
                              )}
                              value={(record as any)[field]}
                            />
                          </View>
                        </Card>
                      );
                    } else {
                      return <></>;
                    }
                  } else {
                    return (
                      <Card key={field} containerStyle={styles.card}>
                        <View style={styles.row}>
                          <Input
                            inputContainerStyle={
                              (record as any)[field]
                                ? {}
                                : inputContainerStyle(field)
                            }
                            editable={inputEditable(field)}
                            keyboardType={
                              field !== 'pin'
                                ? 'default'
                                : Platform.OS === 'android'
                                ? 'numeric'
                                : 'number-pad'
                            }
                            autoCapitalize={'none'}
                            multiline
                            labelStyle={styles.inputLabel}
                            inputStyle={styles.inputStyle}
                            label={inputLabel(field)}
                            placeholder={t(
                              'records.newRecord.inputs.' +
                                transLabel +
                                '.' +
                                field +
                                'Placeholder',
                            )}
                            onChangeText={onChangeText(field)}
                            value={(record as any)[field]}
                          />
                        </View>
                      </Card>
                    );
                  }
                })}
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
                  } else {
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
                      title={t('records.newRecord.addFieldBtn')}
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
                      {t('records.recordDetail.lastEditedLabel') +
                        moment(entity.updatedAt * 1000)
                          .local()
                          .format('YYYY-MM-DD HH:mm:ss')}
                    </Text>
                  </View>
                  {slarkInfo && entity.sharedAt && (
                    <View style={[styles.row, styles.marginRow]}>
                      <Text>
                        {entity.userID === slarkInfo.userID
                          ? t('records.recordDetail.sharingSince') +
                            moment(entity.sharedAt * 1000)
                              .local()
                              .format('YYYY-MM-DD HH:mm:ss')
                          : t('records.recordDetail.sharedBy') +
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

export default RecordDetailStackScreen;
