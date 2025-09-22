/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { ScrollView } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RootStackParamList } from '../../../navigation/routes';
import { useFocusEffect } from '@react-navigation/native';
import {
  makeStyles,
  ButtonGroup,
  Icon,
  ListItem,
  useTheme,
  Button,
} from '@rneui/themed';
import { useTranslation } from 'react-i18next';
import moment from 'moment';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';

interface Email {
  id: number;
  mailbox: string;
  uid: number;
  sentBy: string;
  sentAt: string;
  subject: string;
}

type PrivacyEmailStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'PrivacyEmailStack'
>;

function PrivacyEmailStackScreen({
  navigation,
  route,
}: PrivacyEmailStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { setError } = React.useContext(SnackbarContext);
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [total, setTotal] = React.useState<number>(0);
  const [pageNum, setPageNum] = React.useState<number>(0);
  const [pageSize] = React.useState<number>(20);
  const [list, setList] = React.useState<null | Email[]>(null);
  const [paginationDisabled, setPaginationDiabled] = React.useState([0, 1, 2]);
  const [arrowColors, setArrowColors] = React.useState([
    theme.colors.grey4,
    '',
  ]);
  const [hasAccount, setHasAccount] = React.useState(true);
  const [accounts, setAccounts] = React.useState<null | string[]>(null);

  const onLoad = React.useCallback(async () => {
    if (!slarkInfo) {
      navigation.goBack();
      return;
    }
    setLoading(true);
    const respData = await post('/pswds/getPrivacyEmails/v1', {
      pageNum,
      pageSize,
    });
    setLoading(false);
    if (respData.code !== 0) {
      setError(respData.message, t('app.toast.requestError'));
      return;
    }
    setHasAccount(respData.data.hasAccount);
    setAccounts(respData.data.accounts);
    if (respData.data.hasAccount) {
      if (respData.data.total > 0) {
        let disabled = [0, 2];
        let colors = ['', ''];
        let totalPage =
          respData.data.total % pageSize == 0
            ? Math.floor(respData.data.total / pageSize)
            : Math.floor(respData.data.total / pageSize) + 1;
        if (pageNum == 0) {
          disabled.push(1);
          colors[0] = theme.colors.grey4;
        } else if (pageNum + 1 == totalPage) {
          disabled.push(3);
          colors[1] = theme.colors.grey4;
        }
        setTotal(respData.data.total);
        setPaginationDiabled(disabled);
        setArrowColors(colors);
        if (respData.data && respData.data.list.length > 0) {
          setList(respData.data.list);
        } else {
          setList(null);
        }
      }
    }
  }, [
    slarkInfo,
    setError,
    setLoading,
    t,
    pageNum,
    pageSize,
    theme.colors.grey4,
  ]);

  React.useEffect(() => {
    onLoad();
  }, [onLoad]);

  React.useEffect(() => {
    // 导航栏
    let showedAccount = '';
    if (accounts && accounts.length > 0) {
      showedAccount = accounts[0];
    }
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
      title: showedAccount,
    });
  }, [navigation, accounts]);

  useFocusEffect(
    React.useCallback(() => {
      if (route.params.deleted) {
        if (list) {
          let tmp = [];
          let deleted: null | Email = null;
          for (let i = 0; i < list.length; i++) {
            if (list[i].id !== route.params.deleted) {
              tmp.push(list[i]);
            } else {
              deleted = list[i];
            }
          }
          if (deleted) {
            for (let i = 0; i < tmp.length; i++) {
              if (
                tmp[i].mailbox === deleted.mailbox &&
                tmp[i].uid > deleted.uid
              ) {
                tmp[i].uid -= 1;
              }
            }
          }
          setList(tmp);
        }
      }
    }, [route.params.deleted, list]),
  );

  const handlePaginationChange = async (value: number) => {
    if (!slarkInfo) {
      return;
    }
    let totalPage =
      total % pageSize == 0
        ? Math.floor(total / pageSize)
        : Math.floor(total / pageSize) + 1;
    let newPageNum = 0;
    switch (value) {
      case 0:
        return;
      case 1:
        // check pagination
        if (pageNum + 1 <= 1) {
          return;
        }
        // get list
        newPageNum = pageNum - 1;
        break;
      case 2:
        return;
      case 3:
        // check pagination
        if (pageNum + 1 >= totalPage) {
          return;
        }
        newPageNum = pageNum + 1;
        break;
    }
    setPageNum(newPageNum);
  };

  const applyEmailAccount = async () => {
    if (!slarkInfo) {
      return;
    }
    setLoading(true);
    const respData = await post('/pswds/addPrivacyEmailAccount/v1');
    setLoading(false);
    if (respData.code !== 0) {
      setError(respData.message, t('app.toast.requestError'));
      return;
    }
    await onLoad();
  };

  const onPressListItem = (item: Email) => () => {
    navigation.navigate('PrivacyEmailDetailStack', { id: item.id });
  };

  return (
    <ScrollView style={styles.container}>
      {hasAccount && (
        <>
          {list &&
            list.map(item => (
              <ListItem
                key={item.id}
                bottomDivider
                onPress={onPressListItem(item)}>
                <Icon type="fontisto" name="email" />
                <ListItem.Content>
                  <ListItem.Title>
                    {Buffer.from(item.subject, 'base64').toString('utf-8')}
                  </ListItem.Title>
                  <ListItem.Subtitle>
                    {t('settings.email.itemFrom') + item.sentBy}
                  </ListItem.Subtitle>
                  <ListItem.Subtitle>
                    {t('settings.email.itemDate') +
                      moment(item.sentAt).local().format('YYYY-MM-DD HH:mm:ss')}
                  </ListItem.Subtitle>
                </ListItem.Content>
                <ListItem.Chevron size={40} />
              </ListItem>
            ))}
          {total > pageSize && (
            <ButtonGroup
              disabled={paginationDisabled}
              buttons={[
                t('settings.email.paginationTotal') + total,
                <Icon name="arrow-back-ios" size={20} color={arrowColors[0]} />,
                t('settings.email.paginationPage') +
                  (pageNum + 1) +
                  '/' +
                  (total % pageSize == 0
                    ? Math.floor(total / pageSize)
                    : Math.floor(total / pageSize) + 1),
                <Icon
                  name="arrow-forward-ios"
                  size={20}
                  color={arrowColors[1]}
                />,
              ]}
              onPress={handlePaginationChange}
            />
          )}
        </>
      )}
      {!hasAccount && (
        <>
          <Button
            title={t('settings.email.applyEmailAccountBtn')}
            containerStyle={styles.applyEmailAccountBtn}
            titleStyle={styles.applyEmailAccountTitle}
            size="lg"
            radius={8}
            onPress={applyEmailAccount}
          />
        </>
      )}
    </ScrollView>
  );
}

const useStyles = makeStyles(() => ({
  container: {
    flex: 1,
    marginTop: 8,
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  applyEmailAccountBtn: { width: '100%' },
  applyEmailAccountTitle: { fontSize: 20 },
}));

export default PrivacyEmailStackScreen;
