/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { StyleSheet } from 'react-native';
import { Chip } from '@rneui/themed';
import { Icon } from '@rneui/themed';
import { SnackbarContext } from '../contexts/snackbar';

interface SnackbarProp {
  children?: React.ReactNode;
}

function Snackbar({ children }: SnackbarProp): React.JSX.Element {
  const [msg, setMsg] = React.useState<string>('');
  const timerRef = React.useRef<null | NodeJS.Timeout>(null);
  const [color, setColor] = React.useState<'success' | 'error' | 'warning'>(
    'success',
  );
  const setSuccess = React.useCallback(
    (success: string, guide?: string, timeout?: number) => {
      if (!timeout) {
        timeout = 3;
      }
      timerRef.current = setTimeout(() => {
        setMsg('');
        timerRef.current && clearTimeout(timerRef.current);
        timerRef.current = null;
      }, 1000 * timeout);
      let prompt = success;
      if (guide) {
        prompt = guide + ': ' + prompt;
      }
      setMsg(prompt);
      setColor('success');
    },
    [],
  );
  const setError = React.useCallback(
    (error: string, guide?: string, timeout?: number) => {
      if (!timeout) {
        timeout = 10;
      }
      timerRef.current = setTimeout(() => {
        setMsg('');
        timerRef.current && clearTimeout(timerRef.current);
        timerRef.current = null;
      }, 1000 * timeout);
      let prompt = error;
      if (guide) {
        prompt = guide + ': ' + prompt;
      }
      setMsg(prompt);
      setColor('error');
    },
    [],
  );
  const setWarning = React.useCallback(
    (warning: string, guide?: string, timeout?: number) => {
      if (!timeout) {
        timeout = 5;
      }
      timerRef.current = setTimeout(() => {
        setMsg('');
        timerRef.current && clearTimeout(timerRef.current);
        timerRef.current = null;
      }, 1000 * timeout);
      let prompt = warning;
      if (guide) {
        prompt = guide + ': ' + prompt;
      }
      setMsg(prompt);
      setColor('warning');
    },
    [],
  );
  const onPressClose = () => {
    setMsg('');
    timerRef.current && clearTimeout(timerRef.current);
    timerRef.current = null;
  };

  return (
    <>
      <SnackbarContext.Provider value={{ setSuccess, setError, setWarning }}>
        {children}
      </SnackbarContext.Provider>
      {msg !== '' && (
        <Chip
          title={msg}
          color={color}
          icon={
            <Icon
              type="antdesign"
              onPress={onPressClose}
              name="close"
              color="white"
              size={20}
            />
          }
          iconRight
          size="md"
          radius={8}
          containerStyle={styles.chipContainer}
        />
      )}
    </>
  );
}

const styles = StyleSheet.create({
  chipContainer: {
    position: 'absolute',
    width: '100%',
    bottom: 0,
  },
});

export default Snackbar;
