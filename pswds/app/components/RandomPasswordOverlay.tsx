/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import {useTranslation} from 'react-i18next';
import {
  makeStyles,
  Text,
  Button,
  Overlay,
  useTheme,
  Switch,
  Slider,
} from '@rneui/themed';
import {Pressable, View, TextInput} from 'react-native';
import {randomPassword} from '../common/cipher';

export interface RandomPasswordState {
  password: string;
  charLength: number;
  includeNumber: boolean;
  includeSymbol: boolean;
}

export type RandomPasswordAction =
  | {type: 'setPassword'; value: RandomPasswordState['password']}
  | {type: 'setCharLength'; value: RandomPasswordState['charLength']}
  | {
      type: 'setIncludeNumber';
      value: RandomPasswordState['includeNumber'];
    }
  | {type: 'setIncludeSymbol'; value: RandomPasswordState['includeSymbol']};

export const initRandomPasswordState: RandomPasswordState = {
  charLength: 16,
  includeNumber: true,
  includeSymbol: true,
  password: randomPassword({
    length: 16,
    useNumbers: true,
    useSymbols: true,
  }),
};

export const randomPasswordStateReducer = (
  state: RandomPasswordState,
  action: RandomPasswordAction,
) => {
  switch (action.type) {
    case 'setPassword':
      if (state.password !== action.value) {
        return {...state, password: action.value};
      }
      break;
    case 'setCharLength':
      if (state.charLength !== action.value) {
        return {...state, charLength: action.value};
      }
      break;
    case 'setIncludeNumber':
      if (state.includeNumber !== action.value) {
        return {...state, includeNumber: action.value};
      }
      break;
    case 'setIncludeSymbol':
      if (state.includeSymbol !== action.value) {
        return {...state, includeSymbol: action.value};
      }
      break;
  }
  return state;
};

export type RandomPasswordOverlayProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  setValue: (value: string) => void;
};

function RandomPasswordOverlay({
  visible,
  setVisible,
  setValue,
}: RandomPasswordOverlayProps): React.JSX.Element {
  const {t} = useTranslation();
  const styles = useStyles();
  const {theme} = useTheme();
  const toggleOverlay = () => {
    setVisible(!visible);
  };
  const [randomPasswordState, dispatchRandomPasswordState] = React.useReducer(
    randomPasswordStateReducer,
    initRandomPasswordState,
  );
  const onPressTopClose = () => {
    setVisible(false);
  };
  const onPressUse = () => {
    setValue(randomPasswordState.password);
    setVisible(false);
  };

  return (
    <Overlay
      fullScreen
      overlayStyle={styles.container}
      isVisible={visible}
      onBackdropPress={toggleOverlay}>
      <View style={styles.topline}>
        <Pressable style={styles.pressable} onPress={onPressTopClose} />
      </View>
      <TextInput
        style={styles.pswdText}
        multiline
        numberOfLines={2}
        readOnly
        scrollEnabled
        value={randomPasswordState.password}
      />
      <View style={styles.row}>
        <Text style={styles.fieldLabel}>
          {randomPasswordState.charLength +
            ' ' +
            t('passwords.newPassword.randomPasswordDialog.charLength')}
        </Text>
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
      <View style={styles.row}>
        <Text style={styles.fieldLabel}>
          {t('passwords.newPassword.randomPasswordDialog.includeNumber')}
        </Text>
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
      <View style={styles.row}>
        <Text style={styles.fieldLabel}>
          {t('passwords.newPassword.randomPasswordDialog.includeSymbol')}
        </Text>
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
      <View style={styles.row}>
        <Button
          type="solid"
          radius={8}
          color={theme.colors.primary}
          containerStyle={styles.btnContainer}
          titleStyle={styles.useBtnTitle}
          title={t('passwords.newPassword.randomPasswordDialog.useBtn')}
          onPress={onPressUse}
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
    flexDirection: 'row',
    marginVertical: 24,
    marginHorizontal: 8,
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  btnContainer: {width: '100%'},
  useBtnTitle: {fontSize: 20},
  pswdText: {
    marginVertical: 8,
    marginHorizontal: 8,
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
  charLengthSlider: {width: '50%'},
}));

export default RandomPasswordOverlay;
