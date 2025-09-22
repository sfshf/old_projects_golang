/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import {KeyboardAvoidingView, Platform, InputAccessoryView} from 'react-native';
import {makeStyles, ButtonGroup} from '@rneui/themed';

interface InputToolbarProp {
  inputAccessoryViewID: string;
  list: string[];
  setValue: (val: string) => void;
}

function InputToolbar({
  inputAccessoryViewID,
  list,
  setValue,
}: InputToolbarProp): React.JSX.Element {
  const styles = useStyles();
  const onPressButton = (idx: number) => {
    setValue(list[idx]);
  };
  return (
    <>
      {Platform.OS === 'android' && (
        <KeyboardAvoidingView behavior="padding">
          <ButtonGroup
            buttons={list}
            onPress={onPressButton}
            containerStyle={styles.btnContainerStyle}
          />
        </KeyboardAvoidingView>
      )}
      {Platform.OS === 'ios' && (
        <InputAccessoryView nativeID={inputAccessoryViewID}>
          <ButtonGroup
            buttons={list}
            onPress={onPressButton}
            containerStyle={styles.btnContainerStyle}
          />
        </InputAccessoryView>
      )}
    </>
  );
}

const useStyles = makeStyles(() => ({
  btnContainerStyle: {
    margin: 3,
  },
}));

export default InputToolbar;
