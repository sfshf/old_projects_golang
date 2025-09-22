/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  makeStyles,
  Text,
  Input,
  Button,
  Card,
  Overlay,
  ButtonGroup,
  Icon,
  useTheme,
} from '@rneui/themed';
import { Pressable, View, TouchableOpacity } from 'react-native';
import Clipboard from '@react-native-clipboard/clipboard';
import {
  CameraView,
  useCameraPermissions,
  BarcodeScanningResult,
} from 'expo-camera';
import QRScannerView from './QRScanner';
import * as OTPAuth from 'otpauth';
import * as Progress from 'react-native-progress';
import RandomPasswordOverlay from './RandomPasswordOverlay';
import DatePicker from './DateTimePicker';

export interface OtherField {
  type: 'text' | 'url' | 'password' | 'one-time password' | 'date' | 'pin';
  key: string;
  value: string;
}

export type FieldCardProps = {
  readonly: boolean;
  isOTP: boolean;
  index: number;
  fields: OtherField[];
  setFields: (fields: OtherField[]) => void;
};

export const FieldCard = ({
  readonly,
  isOTP,
  index,
  fields,
  setFields,
}: FieldCardProps) => {
  const { t } = useTranslation();
  const styles = useFieldCardStyles();
  const { theme } = useTheme();
  const [field, setField] = React.useState(fields[index]);
  const [visible, setVisible] = React.useState(false);
  const [openDatePicker, setOpenDatePicker] = React.useState(false);

  React.useEffect(() => {
    setField(fields[index]);
  }, [index, fields]);

  const [copyPassword, setCopyPassword] = React.useState<boolean>(false);
  const onCopyPassword = () => {
    if (field.type === 'one-time password') {
      if (!totp) {
        return;
      }
      let val = totp.generate();
      if (!val) {
        return;
      }
      Clipboard.setString(val);
    } else {
      Clipboard.setString(field.value);
    }
    setCopyPassword(true);
    setField({
      ...field,
      value: t('passwords.passwordDetail.copyPrompt'),
    });
    setTimeout(() => {
      setCopyPassword(false);
      setField({
        ...field,
        value: field.value,
      });
    }, 1000);
  };

  const [qrscan, setQrscan] = React.useState(false);
  const [permission, requestPermission] = useCameraPermissions();
  const cameraViewRef = React.useRef<null | CameraView>(null);
  const onBarcodeScanned = async (result: BarcodeScanningResult) => {
    cameraViewRef.current && cameraViewRef.current.pausePreview();
    let data = result.data.trim();
    setField({ ...field, value: data });
    let tmp = fields;
    tmp[index] = { ...field, value: data };
    setFields(tmp);
    // close camera
    setQrscan(false);
    cameraViewRef.current && cameraViewRef.current.resumePreview();
  };

  const [totp, setTotp] = React.useState<null | OTPAuth.TOTP>(null);
  const [progress, setProgress] = React.useState(0);

  React.useEffect(() => {
    if (!readonly) {
      return;
    }
    if (copyPassword) {
      return;
    }
    if (isOTP) {
      if (!field.value) {
        return;
      }
      let aTotp: null | OTPAuth.TOTP = null;
      if (
        /^otpauth:\/\/([ht]otp)\/(?:[a-zA-Z0-9%]+:)?([^?]+)\?secret=([0-9A-Za-z]+)(?:.*(?:<?counter=)([0-9]+))?/.test(
          field.value,
        )
      ) {
        aTotp = OTPAuth.URI.parse(field.value) as OTPAuth.TOTP;
      } else {
        aTotp = new OTPAuth.TOTP({
          issuer: 'ACME',
          label: field.key,
          algorithm: 'SHA1',
          digits: 6,
          period: 30,
          secret: OTPAuth.Secret.fromUTF8(field.value),
        });
      }
      setTotp(aTotp);
      setInterval(() => {
        let pgs =
          1 -
          (aTotp.period - (Math.floor(Date.now() / 1000) % aTotp.period)) / 30;
        setProgress(pgs);
      }, 1000);
    }
  }, [readonly, isOTP, copyPassword, field]);

  const onPressMinus = (idx: number) => {
    return () => {
      let tmp = [...fields];
      tmp.splice(idx, 1);
      setFields(tmp);
    };
  };

  const onChangeTextFieldKey = (newText: string) => {
    setField({ ...field, key: newText });
    let tmp = fields;
    tmp[index] = { ...field, key: newText };
    setFields(tmp);
  };

  const onChangeTextFieldValue = (newText: string) => {
    setField({ ...field, value: newText });
    let tmp = fields;
    tmp[index] = { ...field, value: newText };
    setFields(tmp);
  };

  const onPressOpenRandomPassword = () => {
    setVisible(true);
  };

  const onPressOpenDatePicker = () => {
    setOpenDatePicker(true);
  };

  const onPressOpenQR = () => {
    setQrscan(true);
    if (permission) {
      if (!permission.granted) {
        requestPermission();
      }
    }
  };

  return (
    <>
      {!qrscan && (
        <Card containerStyle={styles.card} key={index}>
          <View style={styles.row}>
            {!readonly && (
              <View style={styles.left}>
                <TouchableOpacity onPress={onPressMinus(index)}>
                  <Icon
                    size={24}
                    type="feather"
                    name="minus-circle"
                    color={theme.colors.error}
                  />
                </TouchableOpacity>
              </View>
            )}
            <View style={styles.right}>
              {!readonly && (
                <View style={styles.row}>
                  <Input
                    autoCapitalize={'none'}
                    multiline
                    readOnly={readonly || isOTP}
                    style={styles.input}
                    labelStyle={styles.inputLabel}
                    inputStyle={styles.inputStyle}
                    placeholder={t(
                      'passwords.newPassword.otherFields.fieldPlaceholder',
                    )}
                    value={field.key}
                    onChangeText={onChangeTextFieldKey}
                  />
                </View>
              )}
              <TouchableOpacity
                disabled={
                  (field.type !== 'password' &&
                    field.type !== 'one-time password') ||
                  !readonly
                }
                onPress={onCopyPassword}>
                <View
                  style={styles.row}
                  pointerEvents={
                    field.type !== 'date' && !readonly ? 'auto' : 'none'
                  }>
                  <Input
                    inputContainerStyle={
                      field.value
                        ? {}
                        : field.type !== 'date'
                        ? {}
                        : { display: 'none' }
                    }
                    editable={field.type !== 'date'}
                    keyboardType={field.type !== 'pin' ? 'default' : 'numeric'}
                    autoCapitalize={'none'}
                    multiline
                    readOnly={readonly}
                    label={readonly ? field.key : ''}
                    style={styles.input}
                    labelStyle={styles.inputLabel}
                    inputStyle={styles.inputStyle}
                    placeholder={t(
                      'passwords.newPassword.otherFields.valuePlaceholder',
                    )}
                    value={
                      readonly
                        ? field.type === 'one-time password'
                          ? copyPassword
                            ? field.value
                            : totp?.generate()
                          : field.value
                        : field.value
                    }
                    rightIcon={
                      (field.type === 'password' ||
                        field.type === 'one-time password') &&
                      readonly &&
                      !copyPassword && (
                        <Icon type="font-awesome-5" name="copy" size={20} />
                      )
                    }
                    onChangeText={onChangeTextFieldValue}
                  />
                </View>
              </TouchableOpacity>
              {field.type === 'password' && !readonly && (
                <View style={styles.row}>
                  <Button
                    containerStyle={styles.normalBtn}
                    titleStyle={styles.normalBtnTitle}
                    title={t('passwords.newPassword.randomPassword')}
                    radius={8}
                    onPress={onPressOpenRandomPassword}
                  />
                </View>
              )}
              {field.type === 'date' && !readonly && (
                <View style={styles.row}>
                  <Button
                    containerStyle={styles.normalBtn}
                    titleStyle={styles.normalBtnTitle}
                    title={t('records.newRecord.datePicker.label')}
                    radius={8}
                    onPress={onPressOpenDatePicker}
                  />
                </View>
              )}
            </View>
            {isOTP && (
              <View style={styles.suffix}>
                {!readonly && (
                  <TouchableOpacity onPress={onPressOpenQR}>
                    <Icon
                      size={24}
                      type="material-community"
                      name="qrcode-scan"
                      color={theme.colors.primary}
                    />
                  </TouchableOpacity>
                )}
                {readonly && totp && (
                  <Progress.Circle
                    color="transparent"
                    unfilledColor={theme.colors.primary}
                    showsText
                    textStyle={styles.progressText}
                    formatText={prog => {
                      return Math.floor((1 - prog) * 30);
                    }}
                    progress={progress}
                    style={styles.progress}
                  />
                )}
              </View>
            )}
          </View>
          <RandomPasswordOverlay
            visible={visible}
            setVisible={setVisible}
            setValue={(password: string) => {
              setField({ ...field, value: password });
              let tmp = fields;
              tmp[index] = { ...field, value: password };
              setFields(tmp);
            }}
          />
          <DatePicker
            visible={openDatePicker}
            setVisible={setOpenDatePicker}
            value={field.value}
            setValue={(date: string) => {
              setField({ ...field, value: date });
              let tmp = fields;
              tmp[index] = { ...field, value: date };
              setFields(tmp);
            }}
          />
        </Card>
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
};

const useFieldCardStyles = makeStyles(theme => ({
  left: {
    flex: 2,
    justifyContent: 'flex-start',
  },
  right: {
    flex: 8,
  },
  suffix: {
    flex: 2,
    alignItems: 'center',
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  inputLabel: {
    fontSize: 18,
    fontWeight: 'normal',
    color: theme.colors.black,
  },
  inputStyle: {
    fontSize: 18,
    fontWeight: 'normal',
    borderBottomWidth: 0,
  },
  input: {
    width: '100%',
    color: theme.colors.black,
    paddingHorizontal: 12,
  },
  card: { borderRadius: 8 },
  normalBtn: { width: '90%' },
  normalBtnTitle: { fontSize: 16 },
  progress: {
    margin: 10,
  },
  progressText: { color: theme.colors.primary },
}));

export type FieldTypeOverlayProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  otherFields: null | OtherField[];
  setOtherFields: (fields: null | OtherField[]) => void;
};

function FieldTypeOverlay({
  visible,
  setVisible,
  otherFields,
  setOtherFields,
}: FieldTypeOverlayProps): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const toggleOverlay = () => {
    setVisible(!visible);
  };
  const [fieldTypes, setFieldTypes] = React.useState<string[]>([]);
  const onShow = () => {
    // check OTP item
    let has = false;
    if (otherFields) {
      for (let i = 0; i < otherFields.length; i++) {
        if (otherFields[i].type === 'one-time password') {
          has = true;
          break;
        }
      }
    }
    let types = [
      t('passwords.newPassword.otherFields.fieldType1'),
      t('passwords.newPassword.otherFields.fieldType2'),
      t('passwords.newPassword.otherFields.fieldType3'),
      t('passwords.newPassword.otherFields.fieldType5'),
      t('passwords.newPassword.otherFields.fieldType6'),
    ];
    if (has) {
      setFieldTypes(types);
    } else {
      types.push(t('passwords.newPassword.otherFields.fieldType4'));
      setFieldTypes(types);
    }
  };
  const onPressClose = () => {
    setVisible(false);
  };
  const onPressTypeButton = (index: number) => {
    let tmp: OtherField[] = [];
    if (otherFields) {
      tmp = otherFields;
    }
    switch (fieldTypes[index]) {
      case t('passwords.newPassword.otherFields.fieldType1'):
        tmp.push({
          type: 'text',
          key: 'text',
          value: '',
        });
        break;
      case t('passwords.newPassword.otherFields.fieldType2'):
        tmp.push({
          type: 'url',
          key: 'url',
          value: '',
        });
        break;
      case t('passwords.newPassword.otherFields.fieldType3'):
        tmp.push({
          type: 'password',
          key: 'password',
          value: '',
        });
        break;
      case t('passwords.newPassword.otherFields.fieldType4'):
        tmp.push({
          type: 'one-time password',
          key: 'one-time password',
          value: '',
        });
        break;
      case t('passwords.newPassword.otherFields.fieldType5'):
        tmp.push({
          type: 'date',
          key: 'date',
          value: '',
        });
        break;
      case t('passwords.newPassword.otherFields.fieldType6'):
        tmp.push({
          type: 'pin',
          key: 'pin',
          value: '',
        });
        break;
      default:
        break;
    }
    setOtherFields(tmp);
    setVisible(false);
  };
  return (
    <Overlay
      fullScreen
      overlayStyle={styles.container}
      isVisible={visible}
      onShow={onShow}
      onBackdropPress={toggleOverlay}>
      <View style={styles.topline}>
        <Pressable style={styles.pressable} onPress={onPressClose} />
      </View>
      <View style={styles.row}>
        <Text h4 style={styles.headText}>
          {t('passwords.newPassword.otherFields.label')}
        </Text>
        <ButtonGroup
          buttons={fieldTypes}
          onPress={onPressTypeButton}
          containerStyle={styles.groupContainer}
          buttonContainerStyle={styles.groupBtnContainer}
          textStyle={styles.groupBtnText}
          vertical
        />
      </View>
    </Overlay>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    marginTop: '200%',
  },
  topline: {
    height: 10,
    alignItems: 'center',
  },
  pressable: {
    height: 4,
    width: 40,
    marginVertical: 1,
    borderRadius: 2,
    backgroundColor: theme.colors.surface,
  },
  row: {
    marginVertical: 24,
    marginHorizontal: 8,
  },
  headText: { textAlign: 'center' },
  groupContainer: { marginVertical: 30 },
  groupBtnContainer: { borderRadius: 8, height: 50 },
  groupBtnText: { fontSize: 18 },
}));

export default FieldTypeOverlay;
