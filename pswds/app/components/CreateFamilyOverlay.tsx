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
  Input,
  Button,
  Card,
  Overlay,
  useTheme,
} from '@rneui/themed';
import {
  View,
  ScrollView,
  KeyboardAvoidingView,
  Pressable,
} from 'react-native';
import { v4 as uuidv4 } from 'uuid';
import { UnlockPasswordContext } from '../contexts/unlockPassword';
import { SlarkInfoContext } from '../contexts/slark';
import { BackdropContext } from '../contexts/backdrop';
import { SnackbarContext } from '../contexts/snackbar';
import { post } from '../common/http/post';
import { encryptedByUserPublicKey, familyKey } from '../services/cipher';
import { Backup, currentBackup, upsertBackup } from '../common/mmkv/backup';

export type CreateFamilyOverlayProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  successCallback: () => void;
};

function CreateFamilyOverlay({
  visible,
  setVisible,
  successCallback,
}: CreateFamilyOverlayProps): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [description, setDescription] = React.useState(uuidv4().slice(0, 18));

  const onChangeDescription = (newText: string) => {
    setDescription(newText);
  };

  const toggleOverlay = () => {
    setVisible(!visible);
  };

  const onPressClose = () => {
    setVisible(false);
  };

  const onShow = React.useCallback(async () => {
    try {
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
    } catch (error) {
    } finally {
    }
  }, [slarkInfo]);

  const createFamily = async () => {
    try {
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
      if (!description) {
        throw t('settings.family.createFamily.toast.emptyDescription');
      }
      const backup: Backup | null = currentBackup();
      if (!backup || !backup.userPublicKey) {
        throw t('app.toast.internalError');
      }
      setVisible(false);
      // 生成familyKey
      const encryptedKey = encryptedByUserPublicKey(
        familyKey(password),
        new Uint8Array(Buffer.from(backup.userPublicKey, 'hex')),
      );
      // post
      setLoading(true);
      const respData = await post('/pswds/createFamily/v1', {
        description,
        encryptedFamilyKey: encryptedKey.toString('hex'),
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.error'));
        return;
      }
      // 将encryptedFamilyKey存入本地数据库
      upsertBackup({
        ...backup,
        encryptedFamilyKey: encryptedKey.toString('hex'),
      });
      setLoading(false);
      setSuccess(t('app.toast.success'));
      successCallback();
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const onPressCancel = () => {
    setVisible(false);
  };

  return (
    <Overlay
      fullScreen
      overlayStyle={styles.container}
      isVisible={visible}
      onShow={onShow}
      onBackdropPress={toggleOverlay}>
      <KeyboardAvoidingView>
        <View style={styles.topline}>
          <Pressable style={styles.pressable} onPress={onPressClose} />
        </View>
        <ScrollView>
          <Card containerStyle={styles.card}>
            <Input
              autoCapitalize={'none'}
              multiline
              maxLength={20}
              labelStyle={styles.inputLabel}
              inputStyle={styles.inputStyle}
              placeholder={t(
                'settings.family.createFamily.descriptionPlaceholder',
              )}
              onChangeText={onChangeDescription}
              value={description}
            />
          </Card>
          <View style={styles.row}>
            <Button
              title={t('app.alert.commitBtn')}
              containerStyle={styles.normalBtn}
              titleStyle={styles.btnTitle}
              size="lg"
              radius={8}
              onPress={createFamily}
            />
          </View>
          <View style={styles.row}>
            <Button
              title={t('app.alert.cancelBtn')}
              color={theme.colors.error}
              containerStyle={styles.normalBtn}
              titleStyle={styles.btnTitle}
              size="lg"
              radius={8}
              onPress={onPressCancel}
            />
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
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
  card: { borderRadius: 8 },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginTop: 8,
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
  btnTitle: { fontSize: 14, fontWeight: 'normal' },
  normalBtn: { width: '91%', marginVertical: 1 },
}));

export default CreateFamilyOverlay;
