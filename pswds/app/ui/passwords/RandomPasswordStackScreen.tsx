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
import { makeStyles, Text, Button, Switch, Slider } from '@rneui/themed';
import { View, TextInput, TouchableOpacity } from 'react-native';
import { useTheme } from '@rneui/themed';
import Clipboard from '@react-native-clipboard/clipboard';
import ModalScreen from '../../components/ModalScreen';
import { randomPassword } from '../../common/cipher';

interface RandomPasswordState {
  password: string;
  charLength: number;
  includeNumber: boolean;
  includeSymbol: boolean;
}

type RandomPasswordAction =
  | { type: 'setPassword'; value: RandomPasswordState['password'] }
  | { type: 'setCharLength'; value: RandomPasswordState['charLength'] }
  | {
      type: 'setIncludeNumber';
      value: RandomPasswordState['includeNumber'];
    }
  | { type: 'setIncludeSymbol'; value: RandomPasswordState['includeSymbol'] };

const initRandomPasswordState: RandomPasswordState = {
  charLength: 16,
  includeNumber: true,
  includeSymbol: true,
  password: randomPassword({
    length: 16,
    useNumbers: true,
    useSymbols: true,
  }),
};

const randomPasswordStateReducer = (
  state: RandomPasswordState,
  action: RandomPasswordAction,
) => {
  switch (action.type) {
    case 'setPassword':
      if (state.password !== action.value) {
        return { ...state, password: action.value };
      }
      break;
    case 'setCharLength':
      if (state.charLength !== action.value) {
        return { ...state, charLength: action.value };
      }
      break;
    case 'setIncludeNumber':
      if (state.includeNumber !== action.value) {
        return { ...state, includeNumber: action.value };
      }
      break;
    case 'setIncludeSymbol':
      if (state.includeSymbol !== action.value) {
        return { ...state, includeSymbol: action.value };
      }
      break;
  }
  return state;
};

type RandomPasswordStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function RandomPasswordStackScreen({
  navigation,
}: RandomPasswordStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();

  const [randomPasswordState, dispatchRandomPasswordState] = React.useReducer(
    randomPasswordStateReducer,
    initRandomPasswordState,
  );

  const [copyPassword, setCopyPassword] = React.useState<boolean>(false);

  const onPressCopy = () => {
    if (!randomPasswordState.password) {
      return;
    }
    Clipboard.setString(randomPasswordState.password);
    setCopyPassword(true);
    dispatchRandomPasswordState({
      type: 'setPassword',
      value: t('passwords.newPassword.randomPasswordDialog.copyPrompt'),
    });
    setTimeout(() => {
      setCopyPassword(false);
      dispatchRandomPasswordState({
        type: 'setPassword',
        value: randomPasswordState.password,
      });
    }, 1000);
  };

  const onPressUse = () => {
    navigation.popTo('NewPasswordStack', {
      password: randomPasswordState.password,
    });
  };

  return (
    <ModalScreen
      goBack={() => {
        navigation.goBack();
      }}>
      <View style={styles.container}>
        <View style={styles.topBtn} />
        <View style={styles.row}>
          <TouchableOpacity
            style={[
              styles.pswdTextRow,
              copyPassword ? { backgroundColor: theme.colors.grey4 } : {},
            ]}
            onPress={onPressCopy}>
            <TextInput
              style={styles.pswdText}
              multiline
              numberOfLines={2}
              readOnly
              scrollEnabled
              value={randomPasswordState.password}
            />
          </TouchableOpacity>
        </View>
        <View style={styles.row}>
          <View style={styles.column}>
            <Text style={styles.fieldLabel}>
              {randomPasswordState.charLength +
                ' ' +
                t('passwords.newPassword.randomPasswordDialog.charLength')}
            </Text>
          </View>
          <View style={styles.column}>
            <Slider
              style={styles.charLengthSlider}
              thumbTintColor={theme.colors.primary}
              value={randomPasswordState.charLength}
              onValueChange={(value: number) => {
                if (value !== randomPasswordState.charLength) {
                  let password = randomPassword({
                    length: value,
                    useNumbers: randomPasswordState.includeNumber,
                    useSymbols: randomPasswordState.includeSymbol,
                  });
                  dispatchRandomPasswordState({
                    type: 'setPassword',
                    value: password,
                  });
                  dispatchRandomPasswordState({
                    type: 'setCharLength',
                    value: value,
                  });
                }
              }}
              maximumValue={32}
              minimumValue={8}
              step={1}
              allowTouchTrack
            />
          </View>
        </View>
        <View style={styles.row}>
          <View style={styles.column}>
            <Text style={styles.fieldLabel}>
              {t('passwords.newPassword.randomPasswordDialog.includeNumber')}
            </Text>
          </View>
          <View style={styles.column}>
            <Switch
              trackColor={{
                false: theme.colors.grey3,
                true: theme.colors.primary,
              }}
              value={randomPasswordState.includeNumber}
              onValueChange={value => {
                let password = randomPassword({
                  length: randomPasswordState.charLength,
                  useNumbers: value as boolean,
                  useSymbols: randomPasswordState.includeSymbol,
                });
                dispatchRandomPasswordState({
                  type: 'setPassword',
                  value: password,
                });
                dispatchRandomPasswordState({
                  type: 'setIncludeNumber',
                  value: value as boolean,
                });
              }}
            />
          </View>
        </View>
        <View style={styles.row}>
          <View style={styles.column}>
            <Text style={styles.fieldLabel}>
              {t('passwords.newPassword.randomPasswordDialog.includeSymbol')}
            </Text>
          </View>
          <View style={styles.column}>
            <Switch
              trackColor={{
                false: theme.colors.black,
                true: theme.colors.primary,
              }}
              value={randomPasswordState.includeSymbol}
              onValueChange={value => {
                let password = randomPassword({
                  length: randomPasswordState.charLength,
                  useNumbers: randomPasswordState.includeNumber,
                  useSymbols: value as boolean,
                });
                dispatchRandomPasswordState({
                  type: 'setPassword',
                  value: password,
                });

                dispatchRandomPasswordState({
                  type: 'setIncludeSymbol',
                  value: value as boolean,
                });
              }}
            />
          </View>
        </View>
        <View style={styles.row}>
          <View style={styles.column}>
            <Button
              type="solid"
              radius={8}
              color={theme.colors.primary}
              containerStyle={styles.btnContainer}
              titleStyle={styles.btnTitle}
              title={t('passwords.newPassword.randomPasswordDialog.useBtn')}
              onPress={onPressUse}
            />
          </View>
        </View>
      </View>
    </ModalScreen>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    height: '180%',
    width: '100%',
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingBottom: 16,
  },
  topBtn: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: theme.colors.background,
    marginBottom: 16,
    marginTop: 8,
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginVertical: 8,
    marginHorizontal: 8,
  },
  column: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  btnContainer: { width: '100%' },
  btnTitle: { fontSize: 20 },
  pswdTextRow: {
    width: '100%',
  },
  pswdText: {
    borderWidth: 2,
    borderRadius: 8,
    borderColor: theme.colors.black,
    fontSize: 20,
    height: 60,
  },
  fieldLabel: {
    fontSize: 20,
    fontWeight: 'bold',
    color: theme.colors.black,
    alignContent: 'center',
  },
  charLengthSlider: { width: '50%' },
}));

export default RandomPasswordStackScreen;
