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
import { Card, makeStyles, Switch, Text } from '@rneui/themed';
import { useTheme } from '@rneui/themed';
import { useFocusEffect } from '@react-navigation/native';
import { SafeAreaView, View } from 'react-native';
import { encode } from '@doomjs/animated-qrcode';
import QRCode from 'react-native-qrcode-svg';
import { randomDigits } from '../../common/cipher';
import { keccak_256 } from '@noble/hashes/sha3';
import { SlarkInfoContext } from '../../contexts/slark';
import { encryptByXchacha20poly1305 } from '../../services/cipher';
import {
  currentQrPinSetting,
  QrSetting,
  updateQrPinSetting,
} from '../../services/qrCode';
import { useWindowDimensions } from 'react-native';
import { Password } from '../../common/sqlite/schema';

interface TransObject {
  pin: boolean;
  data: string;
}

type DataQRStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'DataQRStack'
>;

function DataQRStackScreen({
  navigation,
  route,
}: DataQRStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { width } = useWindowDimensions();
  const [setting, setSetting] = React.useState<QrSetting>({
    pin: false,
    onlyPassword: true,
  });
  const [pinCode, setPinCode] = React.useState('');

  useFocusEffect(
    React.useCallback(() => {
      const qrSetting = currentQrPinSetting(slarkInfo ? slarkInfo.userID : -1);
      if (qrSetting) {
        setSetting(qrSetting);
        if (qrSetting.pin) {
          setPinCode(randomDigits(4));
        }
      }
    }, [slarkInfo, route.params.entity]),
  );

  const [title, setTitle] = React.useState('/');
  const [value, setValue] = React.useState('');
  const timerRef = React.useRef<null | NodeJS.Timeout>(null);

  React.useEffect(() => {
    let entityJson = '';
    if (
      setting.onlyPassword &&
      'username' in route.params.entity &&
      'password' in route.params.entity
    ) {
      entityJson = JSON.stringify({
        title: (route.params.entity as Password).title,
        username: (route.params.entity as Password).username,
        password: (route.params.entity as Password).password,
      });
    } else {
      entityJson = JSON.stringify(route.params.entity);
    }
    let obj: null | TransObject = null;
    if (setting.pin) {
      let code = pinCode;
      if (!code) {
        code = randomDigits(4);
        setPinCode(code);
      }
      obj = {
        pin: true,
        data: encryptByXchacha20poly1305(keccak_256(code), entityJson),
      };
    } else {
      obj = { pin: false, data: entityJson };
    }
    if (!obj) {
      return;
    }
    const fragments = encode(JSON.stringify(obj), 320, 40);
    if (timerRef.current) clearInterval(timerRef.current);
    let a = 0;
    let b = fragments.length;
    timerRef.current = setInterval(() => {
      if (a == b) {
        a = 0;
      }
      setValue(fragments[a]);
      setTitle(a + 1 + '/' + b);
      a++;
    }, 1000);
  }, [route.params.entity, setting]);

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.row}>
        <Text style={styles.fieldLabel}>{t('dataQR.usePin')}</Text>
        <Switch
          trackColor={{
            false: theme.colors.grey3,
            true: theme.colors.primary,
          }}
          value={setting.pin}
          onValueChange={value => {
            setSetting({ ...setting, pin: value as boolean });
            updateQrPinSetting(slarkInfo ? slarkInfo.userID : -1, {
              ...setting,
              pin: value as boolean,
            });
            if ((value as boolean) === true) {
              setPinCode(randomDigits(4));
            } else {
              setPinCode('');
            }
          }}
        />
      </View>
      {'username' in route.params.entity &&
        'password' in route.params.entity && (
          <View style={styles.row}>
            <Text style={styles.fieldLabel}>{t('dataQR.onlyPassword')}</Text>
            <Switch
              trackColor={{
                false: theme.colors.grey3,
                true: theme.colors.primary,
              }}
              value={setting.onlyPassword}
              onValueChange={value => {
                setSetting({ ...setting, onlyPassword: value as boolean });
                updateQrPinSetting(slarkInfo ? slarkInfo.userID : -1, {
                  ...setting,
                  onlyPassword: value as boolean,
                });
              }}
            />
          </View>
        )}

      {pinCode && (
        <>
          <View style={styles.row}>
            <Text style={styles.fieldLabel}>{t('dataQR.pinCode')}</Text>
            <Text style={styles.fieldLabel}>{pinCode}</Text>
          </View>
        </>
      )}
      {value && (
        <Card containerStyle={styles.card}>
          <QRCode size={width - 40} value={value} />
          <View style={styles.tip}>
            <Text>{title}</Text>
          </View>
        </Card>
      )}
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  tip: {
    alignItems: 'center',
    justifyContent: 'center',
  },
  card: { borderRadius: 8, alignSelf: 'center' },
  row: {
    flexDirection: 'row',
    marginVertical: 8,
    marginHorizontal: 8,
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  fieldLabel: {
    alignContent: 'center',
  },
}));

export default DataQRStackScreen;
