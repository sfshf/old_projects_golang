/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { RootStackParamList } from '../../../navigation/routes';
import type {
  NativeStackScreenProps,
  NativeStackNavigationProp,
} from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { makeStyles, Text, Button, Dialog, Icon } from '@rneui/themed';
import { useTheme } from '@rneui/themed';
import { View } from 'react-native';
import moment from 'moment';
import { SafeAreaView, ScrollView } from 'react-native';
import { WebView } from 'react-native-webview';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';

interface Attachment {
  filename: string;
  size: number;
  content: string;
}

interface Content {
  contentType: string;
  content: string;
}

interface Email {
  id: number;
  mailbox: string;
  uid: number;
  sentBy: string;
  sentAt: string;
  subject: string;
  contents?: Content[];
  attachments?: Attachment[];
}

const ConfirmDeleteDialog = ({
  entity,
  visible,
  setVisible,
  navigation,
}: {
  entity: Email;
  visible: boolean;
  setVisible: (visible: boolean) => void;
  navigation: NativeStackNavigationProp<RootStackParamList>;
}) => {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const { t } = useTranslation();
  const { theme } = useTheme();
  const styles = useDpdStyles();
  const [subject, setSubject] = React.useState('');

  React.useEffect(() => {
    setSubject(Buffer.from(entity.subject, 'base64').toString('utf-8'));
  }, [entity]);

  const doConfirm = React.useCallback(async () => {
    try {
      setVisible(false);
      if (slarkInfo) {
        // 数据同步到后端
        setLoading(true);
        const respData = await post('/pswds/deletePrivacyEmail/v1', {
          id: entity.id,
        });
        if (respData.code !== 0) {
          setLoading(false);
          setError(respData.message, t('app.toast.error'));
          return;
        }
      }
      setLoading(false);
      setSuccess(t('app.toast.success'));
      setTimeout(() => {
        navigation.popTo('PrivacyEmailStack', { deleted: entity.id });
      }, 0);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    } finally {
    }
  }, [slarkInfo, entity]);

  const onPressCancel = () => {
    setVisible(false);
  };

  return (
    <Dialog overlayStyle={styles.overlayStyle} isVisible={visible}>
      <View style={styles.container}>
        <View style={styles.row}>
          <Text style={styles.title}>
            {t('settings.emailDetail.deletePrompt') + subject + ' ?'}
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
            onPress={doConfirm}
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
    margin: 8,
  },
  btnContainer: { width: 100, height: 80, margin: 8 },
  btnTitle: { fontSize: 18, fontWeight: 'normal' },
}));

type PrivacyEmailDetailStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'PrivacyEmailDetailStack'
>;

function PrivacyEmailDetailStackScreen({
  navigation,
  route,
}: PrivacyEmailDetailStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setLoading } = React.useContext(BackdropContext);
  const { setError } = React.useContext(SnackbarContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const [entity, setEntity] = React.useState<null | Email>(null);
  const [openConfirmDialog, setOpenConfirmDialog] =
    React.useState<boolean>(false);

  const [isWebView, setIsWebView] = React.useState(false);
  const checkWebView = (content: string): boolean => {
    return (
      content.trim().startsWith('<html>') && content.trim().endsWith('</html>')
    );
  };
  const onLoad = React.useCallback(async () => {
    if (!slarkInfo) {
      return;
    }
    try {
      setLoading(true);
      const respData = await post('/pswds/getPrivacyEmail/v1', {
        id: route.params.id,
      });
      setLoading(false);
      if (respData.code !== 0) {
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      if (respData.data.contents.length > 0) {
        for (let i = 0; i < respData.data.contents.length; i++) {
          respData.data.contents[i].content = Buffer.from(
            respData.data.contents[i].content,
            'base64',
          ).toString('utf-8');
          if (checkWebView(respData.data.contents[i].content)) {
            setIsWebView(true);
          }
        }
      }
      setEntity(respData.data);
    } catch (error) {}
  }, [slarkInfo, route.params.id, setLoading, setError, t]);
  React.useEffect(() => {
    onLoad();
  }, [onLoad]);

  const navHeaderRight = React.useCallback(() => {
    const onPressOpen = () => {
      setOpenConfirmDialog(true);
    };
    if (slarkInfo) {
      return (
        <Icon
          style={styles.headIcon}
          type="antdesign"
          name="delete"
          onPress={onPressOpen}
        />
      );
    }
    return <></>;
  }, [slarkInfo, styles]);
  React.useEffect(() => {
    // 2. 导航栏
    navigation.setOptions({
      headerBackButtonDisplayMode: 'minimal',
      headerRight: navHeaderRight,
    });
  }, [navigation, navHeaderRight]);

  return (
    <>
      {entity && (
        <>
          <SafeAreaView style={styles.margins}>
            <Text h4>
              {Buffer.from(entity?.subject, 'base64').toString('utf-8')}
            </Text>
            <Text style={styles.sentByLabel}>
              {t('settings.emailDetail.sentByLabel') + entity?.sentBy}
            </Text>
            <Text style={styles.sentAtLabel}>
              {t('settings.emailDetail.sentAtLabel') +
                moment(entity?.sentAt).local().format('YYYY-MM-DD HH:mm:ss')}
            </Text>
          </SafeAreaView>
          {!isWebView && (
            <ScrollView style={styles.margins}>
              <Text />
              {entity.contents &&
                entity.contents.length > 0 &&
                entity.contents.map(item => {
                  if (item.contentType === 'text') {
                    return <Text>{item.content}</Text>;
                  }
                  //  else if (item.contentType === 'image') {
                  //   return (
                  //     <Image
                  //       source={{uri: item.content}}
                  //       containerStyle={styles.attachmentImage}
                  //     />
                  //   );
                  // }
                  else {
                    return <></>;
                  }
                })}
              {entity.attachments && entity.attachments.length > 0 && (
                <Text style={styles.attachmentLabel}>
                  {t('settings.emailDetail.attachmentLabel')}
                </Text>
              )}
              {entity.attachments &&
                entity.attachments.map(item => {
                  return (
                    <View key={item.filename}>
                      <Text>
                        {t('settings.emailDetail.attachmentFilenameLabel') +
                          item.filename}
                      </Text>
                      <Text>
                        {t('settings.emailDetail.attachmentFilesizeLabel') +
                          item.size +
                          'B'}
                      </Text>
                      {/* {item.content && (
                        <Image
                          source={{uri: item.content}}
                          containerStyle={styles.attachmentImage}
                        />
                      )} */}
                    </View>
                  );
                })}
            </ScrollView>
          )}
          {isWebView &&
            entity &&
            entity.contents &&
            entity.contents.length > 0 &&
            entity.contents.map(item => {
              if (checkWebView(item.content)) {
                return (
                  <WebView
                    source={{
                      html: item.content,
                    }}
                  />
                );
              } else {
                return <></>;
              }
            })}
          <ConfirmDeleteDialog
            entity={entity}
            visible={openConfirmDialog}
            setVisible={setOpenConfirmDialog}
            navigation={navigation}
          />
        </>
      )}
    </>
  );
}

const useStyles = makeStyles(() => ({
  margins: {
    marginVertical: 16,
    marginHorizontal: 16,
  },
  headIcon: {
    padding: 10,
    marginHorizontal: 8,
  },
  sentByLabel: { marginTop: 10, fontWeight: 'bold' },
  sentAtLabel: { fontWeight: 'bold' },
  attachmentLabel: { marginTop: 30, fontWeight: 'bold' },
  attachmentImage: { height: 300, width: '100%' },
}));

export default PrivacyEmailDetailStackScreen;
