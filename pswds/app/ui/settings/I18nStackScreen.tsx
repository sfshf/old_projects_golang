/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, StyleSheet } from 'react-native';
import { RootStackParamList } from '../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { changeLanguage, currentLanguage } from '../../common/locales';
import { CheckBox } from '@rneui/themed';
import { SafeAreaView } from 'react-native';

type I18nStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function I18nStackScreen({
  navigation,
}: I18nStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const curLang = currentLanguage();
  const [langs] = React.useState<string[]>(['default', 'zh', 'en']); // map to locale translations
  const [checkList, setCheckList] = React.useState<boolean[]>([
    langs[0] === curLang,
    langs[1] === curLang,
    langs[2] === curLang,
  ]);

  const onPressCheckItem = (index: number, item: string) => {
    return () => {
      const tmpCheckList = [false, false, false];
      tmpCheckList[index] = true;
      changeLanguage(item);
      setCheckList(tmpCheckList);
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
        {langs.map((item, index) => {
          return (
            <CheckBox
              key={index}
              checked={checkList[index]}
              title={t('settings.language.' + item)}
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

export default I18nStackScreen;
