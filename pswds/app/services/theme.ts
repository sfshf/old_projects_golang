import { storage } from '../common/mmkv';
import { createTheme } from '@rneui/themed';
import type { ThemeMode } from '@rneui/themed';
import { useColorScheme } from 'react-native';

export const themeSettingKey = 'cache_settings.theme';

export const currentThemeSetting = (): string => {
  const curTheme = storage.getString(themeSettingKey);
  if (!curTheme) {
    return 'default';
  }
  return curTheme;
};

export const updateThemeSetting = (theme: string) => {
  storage.set(themeSettingKey, theme);
};

export const lightColors = {
  green0: '#9FF127',
  green1: '#30F127',
  green2: '#27F197',
  pink0: '#EABDD8',
  pink1: 'E98CB9',
  pink2: '#E41694',
  purple0: '#F84EF3',
  purple1: '#7F19D8',
  purple2: '#8622AB',
  surface: '#DFE0E2',
};

export const darkColors = {
  green0: '#9FF127',
  green1: '#30F127',
  green2: '#27F197',
  pink0: '#EABDD8',
  pink1: '#E98CB9',
  pink2: '#E41694',
  purple0: '#F84EF3',
  purple1: '#7F19D8',
  purple2: '#8622AB',
  surface: '#36383B',
};
