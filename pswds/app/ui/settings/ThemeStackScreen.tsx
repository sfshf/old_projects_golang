/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, StyleSheet, useColorScheme } from 'react-native';
import { RootStackParamList } from '../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { CheckBox, useThemeMode } from '@rneui/themed';
import type { ThemeMode } from '@rneui/themed';
import { SafeAreaView } from 'react-native';
import { updateThemeSetting, currentThemeSetting } from '../../services/theme';
import { StatusBarStyleContext } from '../../contexts/statusbar';

type ThemeStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function ThemeStackScreen({
  navigation,
}: ThemeStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const { setMode } = useThemeMode();
  const { setStatusBarStyle } = React.useContext(StatusBarStyleContext);
  const [modes] = React.useState<string[]>(['default', 'light', 'dark']); // map to theme modes
  const curTheme = currentThemeSetting();
  const [checkList, setCheckList] = React.useState<boolean[]>([
    modes[0] === curTheme,
    modes[1] === curTheme,
    modes[2] === curTheme,
  ]);

  const systemColorScheme = useColorScheme();
  const onPressCheckItem = (index: number, item: string) => {
    return () => {
      updateThemeSetting(item);
      let tmpCheckList = [false, false, false];
      tmpCheckList[index] = true;
      setCheckList(tmpCheckList);
      let themeMode = item;
      if (item === 'default') {
        themeMode = systemColorScheme ? systemColorScheme : 'light';
      }
      setMode(themeMode as ThemeMode);
      setStatusBarStyle(themeMode === 'dark' ? 'light' : 'dark');
    };
  };

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  return (
    <SafeAreaView style={styles.container}>
      <View>
        {modes.map((item, index) => {
          return (
            <CheckBox
              key={index}
              checked={checkList[index]}
              title={t('settings.theme.' + item)}
              onPress={onPressCheckItem(index, item)}
            />
          );
        })}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
});

export default ThemeStackScreen;
