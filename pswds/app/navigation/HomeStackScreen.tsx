/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { RootStackParamList, HomeTabsParamList } from './routes';
import HomeTabScreen from '../ui/home/HomeTabScreen';
import PasswordsTabScreen from '../ui/passwords/PasswordsTabScreen';
import SettingsTabScreen from '../ui/settings/SettingsTabScreen';
import { useTranslation } from 'react-i18next';
import { Icon } from '@rneui/themed';
import RecordsTabScreen from '../ui/records/RecordsTabScreen';
import { useNavigation } from '@react-navigation/native';

const Tabs = createBottomTabNavigator<HomeTabsParamList>();

type HomeStackScreenProp = NativeStackScreenProps<RootStackParamList>;

const HomeTabBarIcon = () => (
  <Icon type="font-awesome-5" name="home" size={20} color="grey" />
);
const PasswordsTabBarIcon = () => (
  <Icon type="font-awesome-5" name="th-list" size={20} color="grey" />
);
const RecordsTabBarIcon = () => (
  <Icon type="font-awesome-5" name="list-alt" size={20} color="grey" />
);
const SettingsTabBarIcon = () => (
  <Icon type="font-awesome-5" name="wrench" size={20} color="grey" />
);

function HomeStackScreen({}: HomeStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const navigation = useNavigation();
  navigation.getState();
  return (
    <Tabs.Navigator>
      <Tabs.Screen
        name="Home"
        component={HomeTabScreen}
        options={{
          tabBarIcon: HomeTabBarIcon,
          title: t('home.tabTitle'),
        }}
      />
      <Tabs.Screen
        name="Passwords"
        component={PasswordsTabScreen}
        options={{
          tabBarIcon: PasswordsTabBarIcon,
          title: t('passwords.tabTitle'),
        }}
      />
      <Tabs.Screen
        name="Records"
        component={RecordsTabScreen}
        options={{
          tabBarIcon: RecordsTabBarIcon,
          title: t('records.tabTitle'),
        }}
      />
      <Tabs.Screen
        name="Settings"
        component={SettingsTabScreen}
        options={{
          tabBarIcon: SettingsTabBarIcon,
          title: t('settings.tabTitle'),
        }}
      />
    </Tabs.Navigator>
  );
}

export default HomeStackScreen;
