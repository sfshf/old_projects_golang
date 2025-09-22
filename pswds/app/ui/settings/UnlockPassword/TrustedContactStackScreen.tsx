/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, TouchableOpacity, SectionList } from 'react-native';
import type { BottomTabNavigationProp } from '@react-navigation/bottom-tabs';
import type { CompositeNavigationProp } from '@react-navigation/native';
import type {
  NativeStackScreenProps,
  NativeStackNavigationProp,
} from '@react-navigation/native-stack';
import {
  RootStackParamList,
  HomeTabsParamList,
} from '../../../navigation/routes';
import { makeStyles, Button, useTheme, Dialog, Text } from '@rneui/themed';
import { useTranslation } from 'react-i18next';
import { useFocusEffect } from '@react-navigation/native';
import { SafeAreaView } from 'react-native';
import { SlarkInfoContext } from '../../../contexts/slark';
import { currentUnlockPasswordSetting } from '../../../services/unlockPassword';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';

const DeleteDialog = ({
  toDeleted,
  setToDeleted,
}: {
  toDeleted: null | TrustedContact;
  setToDeleted: (visible: null | TrustedContact) => void;
}) => {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = useDpdStyles();
  const [visible, setVisible] = React.useState(false);

  React.useEffect(() => {
    setVisible(toDeleted != null && !toDeleted.deleted);
  }, [toDeleted]);

  const doDelete = React.useCallback(async () => {
    if (!slarkInfo) {
      return;
    }
    if (!toDeleted) {
      return;
    }
    setVisible(false);
    try {
      // 数据同步到后端
      setLoading(true);
      const respData = await post('/pswds/deleteTrustedContact/v1', {
        id: toDeleted.id,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.error'));
        return;
      }
      setLoading(false);
      setSuccess(t('app.toast.success'));
      setToDeleted({ ...toDeleted, deleted: true });
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    } finally {
    }
  }, [slarkInfo, toDeleted]);

  const onPressCancel = () => {
    setToDeleted(null);
  };

  return (
    <Dialog overlayStyle={styles.overlayStyle} isVisible={visible}>
      <View style={styles.container}>
        <View style={styles.row}>
          <Text style={styles.title}>
            {t('passwords.passwordDetail.deletePrompt') +
              toDeleted?.contactEmail}
          </Text>
        </View>
        <View style={styles.row}>
          <Button
            type="solid"
            radius={8}
            color={theme.colors.error}
            containerStyle={styles.btnContainer}
            titleStyle={styles.btnTitle}
            title={t('app.alert.cancelBtn')}
            onPress={onPressCancel}
          />
          <Button
            type="solid"
            radius={8}
            color={theme.colors.primary}
            containerStyle={styles.btnContainer}
            titleStyle={styles.btnTitle}
            title={t('app.alert.confirmBtn')}
            onPress={doDelete}
          />
        </View>
      </View>
    </Dialog>
  );
};

const useDpdStyles = makeStyles(() => ({
  overlayStyle: {
    width: '80%',
    height: '30%',
    borderRadius: 8,
  },
  title: { fontSize: 24 },
  container: {
    flex: 1,
    width: '100%',
    alignItems: 'center',
    justifyContent: 'center',
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    margin: 10,
  },
  btnContainer: { width: 100, height: 80, margin: 10 },
  btnTitle: { fontSize: 20 },
}));

type ItemProps = {
  entity: TrustedContact;
  setToDeleted: (entity: null | TrustedContact) => void;
  navigation?:
    | CompositeNavigationProp<
        BottomTabNavigationProp<HomeTabsParamList>,
        NativeStackNavigationProp<RootStackParamList>
      >
    | CompositeNavigationProp<
        NativeStackNavigationProp<RootStackParamList>,
        BottomTabNavigationProp<HomeTabsParamList>
      >;
};

const Item = ({ entity, setToDeleted }: ItemProps) => {
  const { t } = useTranslation();
  const styles = useItemStyles();
  const { theme } = useTheme();
  const onPressDelete = () => {
    setToDeleted(entity);
    return;
  };
  return (
    <TouchableOpacity disabled style={styles.container} onPress={() => {}}>
      <View style={styles.row}>
        <View style={styles.emailColumn}>
          <Text style={styles.content}>{entity.contactEmail}</Text>
        </View>
        <View style={styles.deleteColumn}>
          <Button
            title={t('settings.unlockPassword.trustedContact.deleteBtn')}
            size="lg"
            color={theme.colors.error}
            containerStyle={styles.deleteBtn}
            titleStyle={styles.btnTitle}
            onPress={onPressDelete}
          />
        </View>
      </View>
    </TouchableOpacity>
  );
};

const useItemStyles = makeStyles(theme => ({
  container: {
    flex: 1,
    backgroundColor: theme.colors.grey5,
    borderRadius: 8,
    padding: 8,
    marginVertical: 4,
    height: 80,
  },
  row: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
  },
  emailColumn: {
    flex: 3,
    justifyContent: 'center',
    alignContent: 'center',
  },
  deleteColumn: {
    flex: 1,
    justifyContent: 'center',
    alignContent: 'center',
  },
  content: {
    fontSize: 16,
  },
  deleteBtn: {
    width: 100,
  },
  btnTitle: { fontSize: 20 },
}));

interface SectionType {
  title: string;
  data: TrustedContact[];
}

type UpdateDataAction = {
  type: 'updateList';
  value: TrustedContact[];
};

interface TrustedContact {
  id: number;
  contactEmail: string;
  deleted: boolean;
}

type TrustedContactStackScreenProp = NativeStackScreenProps<RootStackParamList>;

function TrustedContactStackScreen({
  navigation,
}: TrustedContactStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const initSections: SectionType[] = [
    {
      title: 'contactList',
      data: [],
    },
    {
      title: 'addBtn',
      data: [],
    },
  ];
  const sectionsReducer = (
    sections: SectionType[],
    action: UpdateDataAction,
  ) => {
    const copySections = [...sections];
    switch (action.type) {
      case 'updateList':
        for (let i = 0; i < copySections.length; i++) {
          if (copySections[i].title == 'contactList') {
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
  const [count, setCount] = React.useState(0);
  const { setLoading } = React.useContext(BackdropContext);
  const { setError } = React.useContext(SnackbarContext);
  const onPressAdd = () => {
    navigation.navigate('AddTrustedContactStack');
  };
  const [toDeleted, setToDeleted] = React.useState<null | TrustedContact>(null);
  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
    if (!slarkInfo) {
      navigation.goBack();
      return;
    }
    let curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
    if (!curSetting) {
      navigation.navigate('UnlockPasswordStack');
      return;
    }
  }, [slarkInfo, navigation]);
  const onFocusEffect = React.useCallback(async () => {
    if (!slarkInfo) {
      return;
    }
    try {
      setLoading(true);
      const respData = await post('/pswds/getTrustedContacts/v1');
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      if (respData.data && respData.data.list.length > 0) {
        dispatchSections({
          type: 'updateList',
          value: respData.data.list,
        });
        setCount(respData.data.list.length);
      }
      setLoading(false);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, setError, setLoading, t]);
  useFocusEffect(
    React.useCallback(() => {
      onFocusEffect();
    }, [onFocusEffect]),
  );
  // 删除后刷新;
  const onDelete = React.useCallback(async () => {
    if (!slarkInfo) {
      return;
    }
    if (toDeleted && toDeleted.deleted) {
      try {
        setLoading(true);
        const respData = await post('/pswds/getTrustedContacts/v1');
        if (respData.code !== 0) {
          setLoading(false);
          setError(respData.message, t('app.toast.requestError'));
          return;
        }
        if (respData.data && respData.data.list.length > 0) {
          dispatchSections({
            type: 'updateList',
            value: respData.data.list,
          });
          setCount(respData.data.list.length);
        } else {
          dispatchSections({
            type: 'updateList',
            value: [],
          });
          setCount(0);
        }
        setLoading(false);
      } catch (error) {
        setLoading(false);
        setError(error as string, t('app.toast.internalError'));
      }
    }
  }, [slarkInfo, setError, setLoading, t, toDeleted]);
  React.useEffect(() => {
    onDelete();
  }, [onDelete]);
  return (
    <SafeAreaView style={styles.container}>
      <View>
        <DeleteDialog toDeleted={toDeleted} setToDeleted={setToDeleted} />
        <SectionList
          sections={sections}
          renderSectionHeader={({ section: { title } }) => {
            if (title === 'addBtn' && count < 3) {
              return (
                <Button
                  title={t('settings.unlockPassword.trustedContact.addBtn')}
                  size="lg"
                  containerStyle={styles.btnContainer}
                  titleStyle={styles.btnTitle}
                  onPress={onPressAdd}
                />
              );
            } else {
              return <View />;
            }
          }}
          renderItem={({ item }) => (
            <Item entity={item} setToDeleted={setToDeleted} />
          )}
          keyExtractor={item => item.contactEmail}
        />
      </View>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(() => ({
  container: {
    flex: 1,
    marginTop: 8,
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  btnContainer: { width: '100%' },
  btnTitle: { fontSize: 20 },
}));

export default TrustedContactStackScreen;
