/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { ScrollView, TouchableOpacity, View } from 'react-native';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { CompositeScreenProps } from '@react-navigation/native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RootStackParamList, HomeTabsParamList } from '../../navigation/routes';
import { useFocusEffect } from '@react-navigation/native';
import {
  makeStyles,
  Button,
  ListItem,
  Avatar,
  Text,
  Icon,
  useTheme,
} from '@rneui/themed';
import { useTranslation } from 'react-i18next';
import { iconBgColors, Password } from '../../common/sqlite/schema';
import { SafeAreaView } from 'react-native';
import {
  getPasswordsAsync,
  xorXoredPasswords,
} from '../../common/sqlite/dao/password';
import { SlarkInfoContext } from '../../contexts/slark';
import { SnackbarContext } from '../../contexts/snackbar';
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';

type PasswordsTabScreenProp = CompositeScreenProps<
  BottomTabScreenProps<HomeTabsParamList>,
  NativeStackScreenProps<RootStackParamList>
>;

function PasswordsTabScreen({
  navigation,
}: PasswordsTabScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setError } = React.useContext(SnackbarContext);

  const [list, setList] = React.useState<null | Password[]>(null);

  const inspect = async () => {
    try {
      const curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      const items: Password[] = xorXoredPasswords(
        curSetting!.passwordHash,
        await getPasswordsAsync(),
      );
      if (items.length > 0) {
        setList(items);
      } else {
        setList(null);
      }
    } catch (error) {
      setError(error as string, t('app.toast.internalError'));
    }
  };
  const onPressSearch = () => {
    navigation.navigate('SearchPasswordStack');
  };
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerRight: () => (
        <View style={styles.headerRow}>
          <TouchableOpacity style={styles.headerBtn} onPress={onPressSearch}>
            <Icon
              type="font-awesome-5"
              name="search"
              color={theme.colors.primary}
            />
          </TouchableOpacity>
          <TouchableOpacity style={styles.headerBtn} onPress={onPressAdd}>
            <Icon
              type="font-awesome-5"
              name="plus"
              color={theme.colors.primary}
            />
          </TouchableOpacity>
        </View>
      ),
    });
  }, [navigation]);

  useFocusEffect(
    React.useCallback(() => {
      inspect();
    }, []),
  );

  const onPressListItem = (item: Password) => {
    return () => {
      navigation.navigate('PasswordDetailStack', { dataID: item.dataID });
    };
  };

  const onPressAdd = () => {
    navigation.navigate('NewPasswordStack', {});
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView>
        {list &&
          list.map(item => {
            return (
              <ListItem
                key={item.dataID}
                bottomDivider
                onPress={onPressListItem(item)}>
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
                  <ListItem.Title>
                    <View style={styles.itemTitle}>
                      <Text style={styles.itemTitleTitle}>{item.title}</Text>
                    </View>
                  </ListItem.Title>
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
            );
          })}
        {!list && (
          <View style={styles.row}>
            <Button
              title={t('passwords.firstOneBtn')}
              containerStyle={styles.firstOneBtn}
              titleStyle={styles.firstOneBtnTitle}
              size="lg"
              radius={8}
              onPress={onPressAdd}
            />
          </View>
        )}
      </ScrollView>
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
  headerRow: { flexDirection: 'row' },
  headerBtn: {
    marginHorizontal: 8,
    padding: 0,
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  itemAvatar: {
    borderRadius: 8,
  },
  firstOneBtn: { width: '100%' },
  firstOneBtnTitle: { fontSize: 20 },
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
    padding: 4,
    borderRadius: 4,
  },
}));

export default PasswordsTabScreen;
