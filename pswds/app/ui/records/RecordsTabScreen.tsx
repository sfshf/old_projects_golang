/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { ScrollView, View, Pressable, TouchableOpacity } from 'react-native';
import type {
  BottomTabNavigationProp,
  BottomTabScreenProps,
} from '@react-navigation/bottom-tabs';
import type {
  CompositeNavigationProp,
  CompositeScreenProps,
} from '@react-navigation/native';
import type {
  NativeStackNavigationProp,
  NativeStackScreenProps,
} from '@react-navigation/native-stack';
import { RootStackParamList, HomeTabsParamList } from '../../navigation/routes';
import { useFocusEffect } from '@react-navigation/native';
import {
  makeStyles,
  Button,
  ListItem,
  Overlay,
  Text,
  ButtonGroup,
  Icon,
  useTheme,
} from '@rneui/themed';
import { useTranslation } from 'react-i18next';
import { iconBgColors, Record, RecordType } from '../../common/sqlite/schema';
import { SafeAreaView } from 'react-native';
import {
  getRecordsAsync,
  xorXoredRecords,
} from '../../common/sqlite/dao/record';
import { SlarkInfoContext } from '../../contexts/slark';
import { SnackbarContext } from '../../contexts/snackbar';
import { currentUnlockPasswordSetting } from '../../services/unlockPassword';

export interface RecordIconValue {
  name: string;
  type?: string;
}

export const avatarIcon = (recordType: string): RecordIconValue => {
  let type = 'antdesign';
  let name = 'question';
  switch (recordType) {
    case 'identity':
      type = 'material';
      name = 'perm-identity';
      break;
    case 'credit card':
      type = 'antdesign';
      name = 'creditcard';
      break;
    case 'bank account':
      type = 'antdesign';
      name = 'bank';
      break;
    case 'driver license':
      type = 'font-awesome';
      name = 'drivers-license';
      break;
    case 'passport':
      type = 'fontisto';
      name = 'passport';
      break;
  }
  return { name, type };
};

export type RecordTypeOverlayProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  navigation: CompositeNavigationProp<
    BottomTabNavigationProp<
      HomeTabsParamList,
      keyof HomeTabsParamList,
      undefined
    >,
    NativeStackNavigationProp<RootStackParamList>
  >;
};

export const RecordTypeOverlay = ({
  visible,
  setVisible,
  navigation,
}: RecordTypeOverlayProps) => {
  const { t } = useTranslation();
  const styles = useRecordTypeOverlayStyles();
  const toggleOverlay = () => {
    setVisible(!visible);
  };
  const [recordTypes, setRecordTypes] = React.useState<string[]>([]);

  const onShow = () => {
    setRecordTypes([
      t('records.recordType1'),
      t('records.recordType2'),
      t('records.recordType3'),
      t('records.recordType4'),
      t('records.recordType5'),
    ]);
  };

  const onPressClose = () => {
    setVisible(false);
  };

  const onPressTypeButton = (index: number) => {
    let selected = '';
    switch (recordTypes[index]) {
      case t('records.recordType1'):
        selected = 'identity';
        break;
      case t('records.recordType2'):
        selected = 'credit card';
        break;
      case t('records.recordType3'):
        selected = 'bank account';
        break;
      case t('records.recordType4'):
        selected = 'driver license';
        break;
      case t('records.recordType5'):
        selected = 'passport';
        break;
      default:
        break;
    }
    if (!selected) {
      return;
    }
    navigation.navigate('NewRecordStack', {
      recordType: selected as RecordType,
    });
    setVisible(false);
  };

  return (
    <Overlay
      fullScreen
      overlayStyle={styles.container}
      isVisible={visible}
      onShow={onShow}
      onBackdropPress={toggleOverlay}>
      <View style={styles.topline}>
        <Pressable style={styles.pressable} onPress={onPressClose} />
      </View>
      <View style={styles.row}>
        <Text h4 style={styles.headText}>
          {t('records.recordTypeOverlay.label')}
        </Text>
        <ButtonGroup
          buttons={recordTypes}
          onPress={onPressTypeButton}
          containerStyle={styles.groupContainer}
          buttonContainerStyle={styles.groupBtnContainer}
          textStyle={styles.groupBtnText}
          vertical
        />
      </View>
    </Overlay>
  );
};

const useRecordTypeOverlayStyles = makeStyles(theme => ({
  container: {
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    marginTop: '200%',
    backgroundColor: theme.colors.surface,
  },
  topline: {
    height: 10,
    alignItems: 'center',
  },
  pressable: {
    height: 4,
    width: 40,
    marginVertical: 1,
    borderRadius: 2,
    backgroundColor: theme.colors.surface,
  },
  row: {
    marginVertical: 24,
    marginHorizontal: 8,
  },
  headText: { textAlign: 'center' },
  groupContainer: {
    marginVertical: 30,
    backgroundColor: theme.colors.surface,
  },
  groupBtnContainer: { borderRadius: 8, height: 50 },
  groupBtnText: { fontSize: 18 },
}));

type RecordsTabScreenProp = CompositeScreenProps<
  BottomTabScreenProps<HomeTabsParamList>,
  NativeStackScreenProps<RootStackParamList>
>;

function RecordsTabScreen({
  navigation,
}: RecordsTabScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setError } = React.useContext(SnackbarContext);
  const [visible, setVisible] = React.useState(false);
  const [list, setList] = React.useState<null | Record[]>(null);
  const inspect = async () => {
    try {
      const curSetting = currentUnlockPasswordSetting(
        slarkInfo ? slarkInfo.userID : -1,
      );
      const items: Record[] = xorXoredRecords(
        curSetting!.passwordHash,
        await getRecordsAsync(),
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
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerRight: () => (
        <View>
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

  const onPressAdd = () => {
    setVisible(true);
  };
  const onPressListItem = (item: Record) => {
    return () => {
      navigation.navigate('RecordDetailStack', { dataID: item.dataID });
    };
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView>
        {list &&
          list.map(item => {
            const icon = avatarIcon(item.recordType);
            return (
              <ListItem
                key={item.dataID}
                bottomDivider
                onPress={onPressListItem(item)}>
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
                <ListItem.Content>
                  <ListItem.Title>
                    <View style={styles.itemTitle}>
                      <Text style={styles.itemTitleTitle}>{item.title}</Text>
                    </View>
                  </ListItem.Title>
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
              title={t('records.firstOneBtn')}
              containerStyle={styles.firstOneBtn}
              titleStyle={styles.firstOneBtnTitle}
              size="lg"
              radius={8}
              onPress={onPressAdd}
            />
          </View>
        )}
        <RecordTypeOverlay
          visible={visible}
          setVisible={setVisible}
          navigation={navigation}
        />
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

export default RecordsTabScreen;
