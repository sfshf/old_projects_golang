/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import RootStack from './navigation/RootStack';
import { createTheme, ThemeMode, ThemeProvider } from '@rneui/themed';
import Backdrop from './components/Backdrop';
import Snackbar from './components/Snackbar';
import { Buffer } from 'buffer';
global.Buffer = Buffer;
import { StatusBar } from 'expo-status-bar';
import { ActionSheetProvider } from '@expo/react-native-action-sheet';
import { currentThemeSetting, darkColors, lightColors } from './services/theme';
import { useColorScheme } from 'react-native';
import { StatusBarStyleContext } from './contexts/statusbar';

function App(): React.JSX.Element {
  const systemColorScheme = useColorScheme();
  const createdTheme = React.useMemo(() => {
    return createTheme({
      lightColors,
      darkColors,
      mode:
        currentThemeSetting() === 'default'
          ? systemColorScheme === 'dark'
            ? 'dark'
            : 'light'
          : (currentThemeSetting() as ThemeMode),
    });
  }, [systemColorScheme]);
  const [statusBarStyle, setStatusBarStyle] = React.useState<'light' | 'dark'>(
    'light',
  );
  React.useEffect(() => {
    // 1. theme
    let setting = currentThemeSetting();
    if (setting === 'default') {
      setting = systemColorScheme ? systemColorScheme : 'light';
    }
    // 2. status bar
    setStatusBarStyle(setting === 'dark' ? 'light' : 'dark');
  }, []);

  return (
    <ThemeProvider theme={createdTheme}>
      <ActionSheetProvider>
        <>
          <StatusBar style={statusBarStyle} />
          <StatusBarStyleContext.Provider
            value={{ statusBarStyle, setStatusBarStyle }}>
            <Snackbar>
              <Backdrop>
                <RootStack />
              </Backdrop>
            </Snackbar>
          </StatusBarStyleContext.Provider>
        </>
      </ActionSheetProvider>
    </ThemeProvider>
  );
}

export default App;
