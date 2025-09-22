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
import { makeStyles, Text, Input, Button, Card } from '@rneui/themed';
import { View, ScrollView, KeyboardAvoidingView, Platform } from 'react-native';
import moment from 'moment';
import { v4 as uuidv4 } from 'uuid';
import { z } from 'zod';
import { SafeAreaView } from 'react-native';
import { Record } from '../../common/sqlite/schema';
import FieldTypeOverlay, {
  FieldCard,
  OtherField,
} from '../../components/FieldTypeOverlay';
import DateTimePicker from '../../components/DateTimePicker';
import {
  insertRecordAsync,
  insertRecordTableIndexes,
  xorRecord,
  xorXoredRecord,
} from '../../common/sqlite/dao/record';
import { ResponseCode_DataPullAhead } from '../../common/http';
import { downloadData, updateBackupState } from '../../services/backup';
import { UnlockPasswordContext } from '../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../contexts/slark';
import { BackdropContext } from '../../contexts/backdrop';
import { SnackbarContext } from '../../contexts/snackbar';
import { post } from '../../common/http/post';
import { encryptByUnlockPassword } from '../../services/cipher';
import { SlarkInfo } from '../../services/slark';
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';

// record fields
export const identityFields = [
  'title',
  'firstName',
  'lastName',
  'gender',
  'birthDate',
  'job',
  'address',
  'phone',
  'socialSecurityNumber',
  'idNumber',
];
export const creditCardFields = [
  'title',
  'cardholderName',
  'type',
  'number',
  'verificationNumber',
  'pin',
  'expiryDate',
  'validFrom',
  'issuingBank',
];
export const bankAccountFields = [
  'title',
  'bankName',
  'nameOnAccount',
  'type',
  'routingNumber',
  'branch',
  'accountNumber',
  'swift',
  'pin',
  'phone',
];
export const driverLicenseFields = [
  'title',
  'fullName',
  'address',
  'birthDate',
  'gender',
  'height',
  'number',
  'licenseClass',
  'state',
  'country',
  'expiryDate',
];
export const passportFields = [
  'title',
  'issuingCountry',
  'type',
  'number',
  'fullName',
  'gender',
  'nationality',
  'issuingAuthority',
  'birthDate',
  'birthPlace',
  'issuedOn',
  'expiryDate',
];

type DateInputLabelProps = {
  label: string;
  btnLabel: string;
  field: string;
  record: Record;
  setRecord: (record: Record) => void;
};

export const DateInputLabel = ({
  label,
  btnLabel,
  field,
  record,
  setRecord,
}: DateInputLabelProps) => {
  const styles = useDateInputLabelStyles();
  const [openDatetimePicker, setOpenDatetimePicker] = React.useState(false);
  const onPressDatetimePicker = () => {
    setOpenDatetimePicker(true);
  };
  return (
    <View style={styles.row}>
      <Text h4 style={styles.passwordLabelItem} h4Style={styles.inputLabel}>
        {label}
      </Text>
      <Button
        title={btnLabel}
        radius={8}
        containerStyle={styles.passwordLabelItem}
        titleStyle={styles.randomPasswordTitle}
        onPress={onPressDatetimePicker}
      />
      <DateTimePicker
        visible={openDatetimePicker}
        setVisible={setOpenDatetimePicker}
        value={(record as any)[field]}
        setValue={(date: string) => {
          let tmp = { ...record } as any;
          tmp[field] = date;
          setRecord(tmp);
        }}
      />
    </View>
  );
};

const useDateInputLabelStyles = makeStyles(theme => ({
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  inputLabel: {
    fontSize: 18,
    fontWeight: 'normal',
    color: theme.colors.black,
    marginBottom: 16,
  },
  passwordLabelItem: { width: '50%' },
  randomPasswordTitle: { fontSize: 16 },
}));

export const initRecord: Record = {
  dataID: '',
  createdAt: 0,
  updatedAt: 0,
  userID: 0,
  recordType: '',
  title: '',
  iconBgColor: 0,
  usedAt: 0,
  usedCount: 0,
  sharedAt: null,
  sharingMembers: null,
  sharedToAll: null,
  // mixed fields
  phone: '',
  type: '',
  number: '',
  address: '',
  fullName: '',
  birthDate: '',
  gender: '',
  pin: '',
  expiryDate: '',
  others: '',
  // identity fields
  firstName: '',
  lastName: '',
  job: '',
  socialSecurityNumber: '',
  idNumber: '',
  // credit card fields
  cardholderName: '',
  verificationNumber: '',
  validFrom: '',
  issuingBank: '',
  // bank account fields
  bankName: '',
  nameOnAccount: '',
  routingNumber: '',
  branch: '',
  accountNumber: '',
  swift: '',
  // driver license fields
  height: '',
  licenseClass: '',
  state: '',
  country: '',
  // passport fields
  issuingCountry: '',
  nationality: '',
  issuingAuthority: '',
  birthPlace: '',
  issuedOn: '',
};

export const inputContainerStyle = (field: string): any => {
  switch (field) {
    case 'birthDate':
    case 'expiryDate':
    case 'validFrom':
    case 'issuedOn':
      return { display: 'none' };
    default:
      return {};
  }
};

export const inputEditable = (field: string): boolean => {
  switch (field) {
    case 'birthDate':
    case 'expiryDate':
    case 'validFrom':
    case 'issuedOn':
      return false;
    default:
      return true;
  }
};

type NewRecordStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'NewRecordStack'
>;

function NewRecordStackScreen({
  navigation,
  route,
}: NewRecordStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { setSuccess, setError, setWarning } =
    React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [transLabel, setTransLabel] = React.useState('');
  const [fields, setFields] = React.useState<string[]>([]);
  const [record, setRecord] = React.useState<null | Record>(null);

  React.useEffect(() => {
    if (!route.params) {
      navigation.goBack();
      return;
    }
    if (!route.params.recordType) {
      navigation.goBack();
      return;
    }
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
    switch (route.params.recordType) {
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
    setRecord(initRecord);
  }, [navigation]);

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

  const validateParameters = React.useCallback((): boolean => {
    try {
      if (!record) {
        return false;
      }
      if (!record.title) {
        setError(t('records.newRecord.emptyTitle'), t('app.toast.error'));
        return false;
      }
      if (record.pin) {
        const validation = pinSchema.safeParse(record.pin);
        if (!validation.success) {
          setError(validation.error.issues[0].message, t('app.toast.error'));
          return false;
        }
      }
      if (!password) {
        throw t('app.toast.emptyUnlockPassword');
      }
      return true;
    } catch (error) {
      throw error;
    }
  }, [record, password]);

  const [otherFields, setOtherFields] = React.useState<null | OtherField[]>(
    null,
  );

  const genNewEntity = React.useCallback((): null | Record => {
    const nowTS = moment().unix();
    const entity: Record = {
      ...record!,
      dataID: uuidv4(),
      createdAt: nowTS,
      updatedAt: nowTS,
      userID: slarkInfo ? slarkInfo.userID : -1,
      recordType: route.params.recordType,
      iconBgColor: Math.round(Math.random() * 4),
    };
    // validate others fields
    if (otherFields) {
      for (let i = 0; i < otherFields.length; i++) {
        if (otherFields[i].type === 'one-time password') {
          // validate ont-time password format, refer to https://github.com/google/google-authenticator/wiki/Key-Uri-Format
          const validation = toptSchema.safeParse(otherFields[i].value);
          if (!validation.success) {
            setError(validation.error.issues[0].message, t('app.toast.error'));
            return null;
          }
        } else if (otherFields[i].type === 'pin') {
          const validation = pinSchema.safeParse(otherFields[i].value);
          if (!validation.success) {
            setError(validation.error.issues[0].message, t('app.toast.error'));
            return null;
          }
        }
      }
      entity.others = JSON.stringify(otherFields);
    }
    return entity;
  }, [slarkInfo, otherFields, record]);

  const createNewRecord = React.useCallback(
    async (
      slarkInfo: SlarkInfo,
      password: string,
      passwordHash: string,
      newEntity: Record,
    ): Promise<boolean> => {
      try {
        const updatedAt = newEntity.updatedAt;
        const xoredRecord = xorRecord(passwordHash, newEntity);
        setLoading(true);
        const respData = await post('/pswds/createNonPasswordRecord/v1', {
          updatedAt: updatedAt,
          dataID: newEntity.dataID,
          type: newEntity.recordType,
          content: encryptByUnlockPassword(
            // 整体加密
            password,
            JSON.stringify(xoredRecord),
          ),
        });
        if (respData.code !== 0) {
          if (respData.code === ResponseCode_DataPullAhead) {
            // need to sync data
            const downloadRespData = await downloadData(slarkInfo, password);
            if (downloadRespData.code === 0) {
              // 数据同步成功，警告用户重试操作
              setLoading(false);
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
          updatedAt: updatedAt,
        });
        setLoading(false);
        return true;
      } catch (error) {
        setLoading(false);
        throw error;
      }
    },
    [slarkInfo, password, otherFields],
  );

  const newRecord = async () => {
    try {
      // 1. validate parameters
      const valid = validateParameters();
      if (!valid) {
        return;
      }
      const curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      const newEntity = genNewEntity();
      if (!newEntity) {
        return;
      }
      // 数据同步到后端
      if (slarkInfo) {
        const ok = await createNewRecord(
          slarkInfo,
          password,
          curSetting!.passwordHash,
          newEntity,
        );
        if (!ok) {
          return;
        }
      }
      const xoredEntity = xorRecord(curSetting!.passwordHash, newEntity);
      await insertRecordAsync(xoredEntity);
      // insert record indexes
      insertRecordTableIndexes(curSetting!.passwordHash, xoredEntity);
      setSuccess(t('app.toast.success'));
      navigation.goBack();
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const [visible, setVisible] = React.useState(false);

  const onPressAddField = () => {
    setVisible(true);
  };

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
        return (
          <DateInputLabel
            label={t('records.newRecord.inputs.' + transLabel + '.' + field)}
            btnLabel={t('records.newRecord.datePicker.label')}
            field={field}
            record={record!}
            setRecord={setRecord}
          />
        );
      default:
        return t('records.newRecord.inputs.' + transLabel + '.' + field);
    }
  };

  return (
    <>
      <SafeAreaView style={styles.container}>
        <KeyboardAvoidingView>
          <ScrollView>
            {record &&
              fields.map((field: string) => (
                <Card key={field} containerStyle={styles.card}>
                  <View style={styles.row}>
                    <Input
                      inputContainerStyle={
                        (record as any)[field] ? {} : inputContainerStyle(field)
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
              ))}
            {otherFields &&
              otherFields.length > 0 &&
              otherFields.map((item, idx) => {
                if (item.type === 'one-time password') {
                  return (
                    <FieldCard
                      key={item.key + idx}
                      readonly={false}
                      isOTP={true}
                      index={idx}
                      fields={otherFields}
                      setFields={setOtherFields}
                    />
                  );
                }
              })}
            {otherFields &&
              otherFields.length > 0 &&
              otherFields.map((item, idx) => {
                if (item.type !== 'one-time password') {
                  return (
                    <FieldCard
                      key={item.key + idx}
                      readonly={false}
                      isOTP={false}
                      index={idx}
                      fields={otherFields}
                      setFields={setOtherFields}
                    />
                  );
                }
              })}
            <Card containerStyle={styles.card}>
              <Button
                type="clear"
                title={t('records.newRecord.addFieldBtn')}
                containerStyle={styles.normalBtn}
                titleStyle={styles.btnTitle}
                size="lg"
                radius={8}
                onPress={onPressAddField}
              />
            </Card>
            <View style={styles.row}>
              <Button
                title={t('records.newRecord.newBtn')}
                containerStyle={styles.commitBtn}
                titleStyle={styles.btnTitle}
                size="lg"
                radius={8}
                onPress={newRecord}
              />
            </View>
            <FieldTypeOverlay
              visible={visible}
              setVisible={setVisible}
              otherFields={otherFields}
              setOtherFields={setOtherFields}
            />
          </ScrollView>
        </KeyboardAvoidingView>
      </SafeAreaView>
    </>
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
    marginVertical: 2,
    alignItems: 'center',
    justifyContent: 'center',
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
  passwordLabelItem: { width: '50%' },
  randomPasswordTitle: { fontSize: 16 },
  btnTitle: { fontSize: 20, fontWeight: 'normal' },
  normalBtn: { width: '100%' },
  commitBtn: { marginVertical: 10, width: '95%' },
}));

export default NewRecordStackScreen;
