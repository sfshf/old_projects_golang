/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import {
  makeStyles,
  Icon,
  Button,
  useTheme,
  ListItem,
  Avatar,
  Text,
} from '@rneui/themed';
import { iconBgColors } from '../../common/sqlite/schema';
import { SectionList, TouchableOpacity, View } from 'react-native';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { CompositeScreenProps } from '@react-navigation/native';
import { RootStackParamList, HomeTabsParamList } from '../../navigation/routes';
import { useTranslation } from 'react-i18next';
import { useFocusEffect } from '@react-navigation/native';
import { SafeAreaView } from 'react-native';
import { avatarIcon } from '../records/RecordsTabScreen';
import {
  getLatestUpdatedPasswordsAsync,
  getLatestUsedPasswordsAsync,
  getMostUsedPasswordsAsync,
  xorXoredPasswords,
} from '../../common/sqlite/dao/password';
import {
  getLatestUpdatedRecordsAsync,
  getLatestUsedRecordsAsync,
  getMostUsedRecordsAsync,
  xorXoredRecords,
} from '../../common/sqlite/dao/record';
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';
import { UnlockPasswordContext } from '../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../contexts/slark';
import { SnackbarContext } from '../../contexts/snackbar';

interface SectionType {
  title: string;
  open: boolean;
  data: any[];
}

type UpdateDataAction =
  | { type: 'updateLatestUsedData'; value: any[] }
  | { type: 'setLatestUsedOpen'; value: boolean }
  | {
      type: 'updateMostUsedData';
      value: any[];
    }
  | { type: 'setMostUsedOpen'; value: boolean }
  | { type: 'updateLatestUpdatedData'; value: any[] }
  | { type: 'setLatestUpdatedOpen'; value: boolean };

type HomeTabScreenProp = CompositeScreenProps<
  BottomTabScreenProps<HomeTabsParamList>,
  NativeStackScreenProps<RootStackParamList>
>;

function HomeTabScreen({ navigation }: HomeTabScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const styles = useStyles();
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { setError } = React.useContext(SnackbarContext);

  const initSections: SectionType[] = [
    {
      title: t('home.searchPassword.latestUsedBtn'),
      open: false,
      data: [],
    },
    {
      title: t('home.searchPassword.mostUsedBtn'),
      open: false,
      data: [],
    },
    {
      title: t('home.searchPassword.latestUpdatedBtn'),
      open: false,
      data: [],
    },
  ];

  const handleCopySectionsData = (
    copySections: SectionType[],
    title: string,
    value: any[],
  ): SectionType[] => {
    for (let i = 0; i < copySections.length; i++) {
      if (copySections[i].title == title) {
        copySections[i].data = value;
        break;
      }
    }
    return copySections;
  };

  const handleCopySectionsOpen = (
    copySections: SectionType[],
    title: string,
    value: boolean,
  ): SectionType[] => {
    for (let i = 0; i < copySections.length; i++) {
      if (copySections[i].title == title) {
        copySections[i].open = value;
        break;
      }
    }
    return copySections;
  };

  const sectionsReducer = (
    sections: SectionType[],
    action: UpdateDataAction,
  ) => {
    const copySections = [...sections];
    switch (action.type) {
      case 'updateLatestUsedData':
        return handleCopySectionsData(
          copySections,
          t('home.searchPassword.latestUsedBtn'),
          action.value,
        );
      case 'setLatestUsedOpen':
        return handleCopySectionsOpen(
          copySections,
          t('home.searchPassword.latestUsedBtn'),
          action.value,
        );
      case 'updateMostUsedData':
        return handleCopySectionsData(
          copySections,
          t('home.searchPassword.mostUsedBtn'),
          action.value,
        );
      case 'setMostUsedOpen':
        return handleCopySectionsOpen(
          copySections,
          t('home.searchPassword.mostUsedBtn'),
          action.value,
        );
      case 'updateLatestUpdatedData':
        return handleCopySectionsData(
          copySections,
          t('home.searchPassword.latestUpdatedBtn'),
          action.value,
        );
      case 'setLatestUpdatedOpen':
        return handleCopySectionsOpen(
          copySections,
          t('home.searchPassword.latestUpdatedBtn'),
          action.value,
        );
      default:
        return copySections;
    }
  };

  const [sections, dispatchSections] = React.useReducer(
    sectionsReducer,
    initSections,
  );

  const searchMostUsed = React.useCallback(async () => {
    try {
      let userID = slarkInfo ? slarkInfo.userID : -1;
      let curSetting = currentUnlockPasswordSetting(userID);
      if (!curSetting) {
        return;
      }
      const items: any[] = xorXoredPasswords(
        curSetting!.passwordHash,
        await getMostUsedPasswordsAsync(6),
      );
      if (items.length < 6) {
        const records = xorXoredRecords(
          curSetting!.passwordHash,
          await getMostUsedRecordsAsync(6 - items.length),
        );
        items.push(...records);
      }
      dispatchSections({ type: 'updateMostUsedData', value: items });
      dispatchSections({ type: 'setMostUsedOpen', value: true });
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    } finally {
    }
  }, [setError, t, slarkInfo]);
  const searchLatestUsed = React.useCallback(async () => {
    try {
      let userID = slarkInfo ? slarkInfo.userID : -1;
      let curSetting = currentUnlockPasswordSetting(userID);
      if (!curSetting) {
        return;
      }
      const items: any[] = xorXoredPasswords(
        curSetting!.passwordHash,
        await getLatestUsedPasswordsAsync(6),
      );
      if (items.length < 6) {
        const records = xorXoredRecords(
          curSetting!.passwordHash,
          await getLatestUsedRecordsAsync(6 - items.length),
        );
        items.push(...records);
      }
      dispatchSections({ type: 'updateLatestUsedData', value: items });
      dispatchSections({ type: 'setLatestUsedOpen', value: true });
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    } finally {
    }
  }, [setError, t, slarkInfo]);
  const searchLatestUpdated = React.useCallback(async () => {
    try {
      let userID = slarkInfo ? slarkInfo.userID : -1;
      let curSetting = currentUnlockPasswordSetting(userID);
      if (!curSetting) {
        return;
      }
      const items: any[] = xorXoredPasswords(
        curSetting!.passwordHash,
        await getLatestUpdatedPasswordsAsync(6),
      );
      if (items.length < 6) {
        const records = xorXoredRecords(
          curSetting!.passwordHash,
          await getLatestUpdatedRecordsAsync(6 - items.length),
        );
        items.push(...records);
      }
      dispatchSections({ type: 'updateLatestUpdatedData', value: items });
      dispatchSections({ type: 'setLatestUpdatedOpen', value: true });
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    } finally {
    }
  }, [setError, t, slarkInfo]);
  const onPressSearch = () => {
    navigation.navigate('SearchPasswordStack');
  };
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerRight: () => (
        <View>
          <TouchableOpacity style={styles.headerBtn} onPress={onPressSearch}>
            <Icon
              type="font-awesome-5"
              name="search"
              color={theme.colors.primary}
            />
          </TouchableOpacity>
        </View>
      ),
    });
  }, [navigation]);
  // NOTE：因为锁屏解锁后会刷新本地数据库，且调用动作为轻量操作，所以不用React.useCallback来阻止刷新；
  useFocusEffect(
    React.useCallback(() => {
      if (password) {
        searchMostUsed();
        searchLatestUsed();
        searchLatestUpdated();
      }
    }, [password]),
  );

  const onPressSectionHeader = (section: SectionType) => {
    return () => {
      switch (section.title) {
        case t('home.searchPassword.mostUsedBtn'):
          if (section.open) {
            dispatchSections({
              type: 'updateMostUsedData',
              value: [],
            });
            dispatchSections({
              type: 'setMostUsedOpen',
              value: false,
            });
          } else {
            searchMostUsed();
          }
          break;
        case t('home.searchPassword.latestUsedBtn'):
          if (section.open) {
            dispatchSections({
              type: 'updateLatestUsedData',
              value: [],
            });
            dispatchSections({
              type: 'setLatestUsedOpen',
              value: false,
            });
          } else {
            searchLatestUsed();
          }
          break;
        case t('home.searchPassword.latestUpdatedBtn'):
          if (section.open) {
            dispatchSections({
              type: 'updateLatestUpdatedData',
              value: [],
            });
            dispatchSections({
              type: 'setLatestUpdatedOpen',
              value: false,
            });
          } else {
            searchLatestUpdated();
          }
          break;
      }
    };
  };
  const onPressSectionItem = (item: any) => {
    return () => {
      if (item.recordType) {
        navigation.navigate('RecordDetailStack', { dataID: item.dataID });
      } else {
        navigation.navigate('PasswordDetailStack', { dataID: item.dataID });
      }
    };
  };

  return (
    <SafeAreaView style={styles.container}>
      <SectionList
        sections={sections}
        renderSectionHeader={({ section }: { section: SectionType }) => {
          return (
            <Button
              size="lg"
              buttonStyle={styles.menuButton}
              containerStyle={styles.menuButtonContainerStyle}
              titleStyle={styles.menuButtonTitleStyle}
              onPress={onPressSectionHeader(section)}>
              {section.title}
              {section.open && (
                <Icon
                  size={30}
                  name="keyboard-arrow-down"
                  color={theme.colors.primary}
                />
              )}
              {!section.open && (
                <Icon
                  size={20}
                  name="arrow-forward-ios"
                  color={theme.colors.primary}
                />
              )}
            </Button>
          );
        }}
        renderItem={({ item }) => {
          let icon: any = null;
          if (item.recordType) {
            icon = avatarIcon(item.recordType);
          }
          return (
            <ListItem
              key={item.dataID}
              bottomDivider
              onPress={onPressSectionItem(item)}>
              {!item.recordType && (
                <>
                  {item.website && (
                    <Avatar
                      size={40}
                      source={
                        item.website
                          ? { uri: item.website + '/favicon.ico' }
                          : {}
                      }
                      imageProps={{
                        style: { borderRadius: 8 },
                        PlaceholderContent: (
                          <Avatar
                            size={40}
                            title={item.title.slice(0, 2)}
                            containerStyle={[
                              styles.itemAvatar,
                              {
                                backgroundColor:
                                  iconBgColors[
                                    item.iconBgColor ? item.iconBgColor : 0
                                  ],
                              },
                            ]}
                          />
                        ),
                      }}
                    />
                  )}
                  {!item.website && (
                    <Avatar
                      size={40}
                      title={item.title.slice(0, 2)}
                      containerStyle={[
                        styles.itemAvatar,
                        {
                          backgroundColor:
                            iconBgColors[
                              item.iconBgColor ? item.iconBgColor : 0
                            ],
                        },
                      ]}
                    />
                  )}
                </>
              )}
              {item.recordType && (
                <Icon
                  size={40}
                  type={icon.type}
                  name={icon.name}
                  containerStyle={[
                    styles.itemAvatar,
                    {
                      backgroundColor:
                        iconBgColors[item.iconBgColor ? item.iconBgColor : 0],
                    },
                  ]}
                />
              )}
              <ListItem.Content>
                <ListItem.Title>
                  <View style={styles.itemTitle}>
                    <Text style={styles.itemTitleTitle}>{item.title}</Text>
                  </View>
                </ListItem.Title>
                {!item.recordType && (
                  <ListItem.Subtitle>{item.username}</ListItem.Subtitle>
                )}
              </ListItem.Content>
              {item.sharedAt &&
                item.sharedAt > 0 &&
                item.userID != slarkInfo?.userID && (
                  <Text style={styles.itemSharing}>
                    {t('familyShare.status.sharing')}
                  </Text>
                )}
              <ListItem.Chevron size={40} />
            </ListItem>
          );
        }}
        keyExtractor={(item: any) => item.dataID}
      />
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  menuButton: {
    backgroundColor: theme.colors.white,
  },
  headerBtn: {
    marginHorizontal: 8,
    padding: 0,
  },
  menuButtonContainerStyle: {
    width: '100%',
  },
  menuButtonTitleStyle: {
    flex: 1,
    fontSize: 20,
    color: theme.colors.primary,
    textAlign: 'left',
  },
  itemAvatar: {
    borderRadius: 8,
  },
  itemTitle: {
    flexDirection: 'row',
    alignContent: 'space-between',
  },
  itemTitleTitle: { fontSize: 20 },
  itemSharing: {
    fontSize: 12,
    backgroundColor: theme.colors.green0,
    padding: 4,
    borderRadius: 4,
  },
  itemShared: {
    fontSize: 12,
    backgroundColor: theme.colors.primary,
    marginHorizontal: 4,
    padding: 4,
    borderRadius: 4,
  },
}));

export default HomeTabScreen;
