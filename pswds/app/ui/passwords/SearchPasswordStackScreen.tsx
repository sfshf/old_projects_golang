/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { makeStyles, Input, ListItem, Avatar, Text } from '@rneui/themed';
import { iconBgColors, Password } from '../../common/sqlite/schema';
import { View, SectionList } from 'react-native';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { CompositeScreenProps } from '@react-navigation/native';
import { RootStackParamList, HomeTabsParamList } from '../../navigation/routes';
import { useTranslation } from 'react-i18next';
import { SafeAreaView, KeyboardAvoidingView } from 'react-native';
import {
  searchPasswordsByTitleAsync,
  XoredPassword,
  xorXoredPasswords,
} from '../../common/sqlite/dao/password';
import { SlarkInfoContext } from '../../contexts/slark';
import { SnackbarContext } from '../../contexts/snackbar';
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';

interface SectionType {
  title: string;
  data: Password[];
}

type UpdateDataAction = { type: 'updateSearchData'; value: Password[] };

type SearchPasswordStackScreenProp = CompositeScreenProps<
  NativeStackScreenProps<RootStackParamList>,
  BottomTabScreenProps<HomeTabsParamList>
>;

function SearchPasswordStackScreen({
  navigation,
}: SearchPasswordStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const styles = useStyles();
  const { t } = useTranslation();
  const { setError } = React.useContext(SnackbarContext);
  const [searchText, setSearchText] = React.useState('');

  const initSections: SectionType[] = [
    {
      title: 'searchInput',
      data: [],
    },
  ];

  const sectionsReducer = (
    sections: SectionType[],
    action: UpdateDataAction,
  ) => {
    const copySections = [...sections];
    switch (action.type) {
      case 'updateSearchData':
        for (let i = 0; i < copySections.length; i++) {
          if (copySections[i].title == 'searchInput') {
            copySections[i].data = action.value;
            break;
          }
        }
        return copySections;
      default:
        return copySections;
    }
  };

  const [sections, dispatchSections] = React.useReducer(
    sectionsReducer,
    initSections,
  );

  const searchList = async (query: string) => {
    try {
      let curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      const items: Password[] = xorXoredPasswords(
        curSetting!.passwordHash,
        await searchPasswordsByTitleAsync(query, 10),
      );
      dispatchSections({ type: 'updateSearchData', value: items });
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const onPressListItem = (item: Password) => {
    return () => {
      navigation.navigate('PasswordDetailStack', { dataID: item.dataID });
    };
  };

  const onChangeSearchText = (newText: string) => {
    newText = newText.trim();
    setSearchText(newText);
    if (newText) {
      searchList(newText);
    } else {
      dispatchSections({ type: 'updateSearchData', value: [] });
    }
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
      <KeyboardAvoidingView>
        <SectionList
          sections={sections}
          renderSectionHeader={({ section: { title } }) => {
            if (title === 'searchInput') {
              return (
                <Input
                  autoCapitalize={'none'}
                  containerStyle={styles.inputContainer}
                  inputContainerStyle={styles.innerInputContainer}
                  placeholder={t(
                    'passwords.searchPassword.searchInputPlaceholder',
                  )}
                  value={searchText}
                  onChangeText={onChangeSearchText}
                />
              );
            } else {
              return <View />;
            }
          }}
          renderItem={({ item }) => (
            <ListItem bottomDivider onPress={onPressListItem(item)}>
              {item.website && (
                <Avatar
                  size={40}
                  source={
                    item.website ? { uri: item.website + '/favicon.ico' } : {}
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
                        iconBgColors[item.iconBgColor ? item.iconBgColor : 0],
                    },
                  ]}
                />
              )}
              <ListItem.Content>
                <View style={styles.itemTitle}>
                  <Text style={styles.itemTitleTitle}>{item.title}</Text>
                </View>
                <ListItem.Subtitle>{item.username}</ListItem.Subtitle>
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
          )}
          keyExtractor={(item: Password) => item.title}
        />
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
    marginTop: 8,
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  body: {
    flex: 12,
    marginTop: 20,
  },
  inputContainer: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    height: 60,
    borderWidth: 1,
    borderRadius: 8,
    borderColor: theme.colors.black,
    color: theme.colors.black,
    marginVertical: 8,
  },
  innerInputContainer: {
    borderBottomWidth: 0,
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

export default SearchPasswordStackScreen;
