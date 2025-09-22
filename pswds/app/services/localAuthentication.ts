import * as Keychain from 'react-native-keychain';
import i18n from '../common/locales';
import { check, request, PERMISSIONS, RESULTS } from 'react-native-permissions';

export const setFingerprintPassword = async (
  username: string,
  password: string,
) => {
  try {
    return await Keychain.setGenericPassword(username, password, {
      accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      authenticationPrompt: {
        title: i18n.t('lockScreen.authenticationPromptTitle'),
        cancel: i18n.t('app.alert.cancelBtn'),
      },
      accessControl: Keychain.ACCESS_CONTROL.BIOMETRY_CURRENT_SET,
      securityLevel: Keychain.SECURITY_LEVEL.SECURE_HARDWARE,
      storage: Keychain.STORAGE_TYPE.RSA,
      service: 'nextsurfer.pswds.test' + username,
    });
  } catch (error) {
    throw error;
  }
};

export const hasFingerprintPassword = async (username: string) => {
  try {
    return await Keychain.hasGenericPassword({
      service: 'nextsurfer.pswds.test' + username,
    });
  } catch (error) {
    throw error;
  }
};

export const deleteFingerprintPassword = async (username: string) => {
  try {
    return await Keychain.resetGenericPassword({
      service: 'nextsurfer.pswds.test' + username,
    });
  } catch (error) {
    throw error;
  }
};

export const getGenericPassword = async (username: string) => {
  try {
    const result = await Keychain.getGenericPassword({
      accessControl: Keychain.ACCESS_CONTROL.BIOMETRY_CURRENT_SET,
      authenticationPrompt: {
        title: i18n.t('lockScreen.authenticationPromptTitle'),
        cancel: i18n.t('app.alert.cancelBtn'),
      },
      rules: Keychain.SECURITY_RULES.NONE,
      service: 'nextsurfer.pswds.test' + username,
    });
    if (result === false) {
      return '';
    } else {
      return result.password;
    }
  } catch (error) {
    throw error;
  }
};

export const requestFaceIDPermission = async () => {
  const permission = await check(PERMISSIONS.IOS.FACE_ID);
  switch (permission) {
    case RESULTS.GRANTED:
      return { granted: true, message: 'ok' };
    case RESULTS.BLOCKED:
      return { granted: false, message: 'blocked' };
    case RESULTS.DENIED:
      const requestPermission = await request(PERMISSIONS.IOS.FACE_ID);
      if (requestPermission === RESULTS.GRANTED) {
        return { granted: false, message: 'denied' };
      }
      return { granted: false, message: 'denied' };
    case RESULTS.LIMITED:
      return { granted: false, message: 'limited' };
    case RESULTS.UNAVAILABLE:
      return { granted: false, message: 'unavailable' };
    default:
      return { granted: false, message: 'default' };
  }
};
