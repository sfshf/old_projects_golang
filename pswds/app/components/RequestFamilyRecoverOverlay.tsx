/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  makeStyles,
  Overlay,
  ListItem,
  Avatar,
  useTheme,
  Icon,
} from '@rneui/themed';
import { Alert, Pressable, TouchableOpacity, View } from 'react-native';
import { OtherFamilyMember } from '../services/family';
import { SlarkInfoContext } from '../contexts/slark';
import { BackdropContext } from '../contexts/backdrop';
import { SnackbarContext } from '../contexts/snackbar';
import { post } from '../common/http/post';
import { ResponseCode_ResourceLimit } from '../common/http';

type MemberItemProps = {
  email: string;
  item: OtherFamilyMember;
  setVisible: (visible: boolean) => void;
  setCanRecover: (can: boolean) => void;
  setCanFamilyRecover: (can: boolean) => void;
};

const MemberItem = ({
  email,
  item,
  setVisible,
  setCanRecover,
  setCanFamilyRecover,
}: MemberItemProps) => {
  const styles = useMemberItemStyles();
  const { t } = useTranslation();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);

  const onPressSend = React.useCallback(async () => {
    try {
      setVisible(false);
      setLoading(true);
      const respData = await post('/pswds/requestFamilyRecover/v1', {
        email,
        userID: item.userID,
      });
      if (respData.code !== 0) {
        setLoading(false);
        if (respData.code === ResponseCode_ResourceLimit) {
          Alert.alert(
            t('app.alert.recoverRequestLimitTitle'),
            t('app.alert.recoverRequestLimitMessage'),
            [
              {
                text: t('app.alert.okBtn'),
                style: 'destructive',
              },
            ],
          );
        } else {
          setError(respData.message, t('app.toast.requestError'));
        }
        return;
      }
      setLoading(false);
      setCanFamilyRecover(false);
      Alert.alert(
        t('app.alert.emailHasSentTitle'),
        t('app.alert.emailHasSentMessage') + email,
        [
          {
            text: t('app.alert.okBtn'),
            onPress: () => {},
          },
        ],
      );
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [email, item]);

  return (
    <ListItem key={item.userID} bottomDivider>
      <Avatar
        rounded
        icon={{
          name: 'person-outline',
          type: 'material',
          size: 26,
        }}
        containerStyle={styles.avatarContainer}
      />
      <ListItem.Content>
        <ListItem.Title>{item.email}</ListItem.Title>
      </ListItem.Content>
      <TouchableOpacity onPress={onPressSend}>
        <Icon type="material" name="send" />
      </TouchableOpacity>
    </ListItem>
  );
};

const useMemberItemStyles = makeStyles(theme => ({
  avatarContainer: { backgroundColor: theme.colors.surface },
  itemSubtitle: { fontSize: 12 },
}));

export interface BackupMember {
  userID: number;
  email: string;
}

export type RequestFamilyRecoverOverlayProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  email: string;
  backupMembers: BackupMember[];
  setCanRecover: (can: boolean) => void;
  setCanFamilyRecover: (can: boolean) => void;
};

function RequestFamilyRecoverOverlay({
  visible,
  setVisible,
  email,
  backupMembers,
  setCanRecover,
  setCanFamilyRecover,
}: RequestFamilyRecoverOverlayProps): React.JSX.Element {
  const styles = useStyles();

  const toggleOverlay = () => {
    setVisible(!visible);
  };

  const onPressClose = () => {
    setVisible(false);
  };

  return (
    <Overlay
      fullScreen
      overlayStyle={styles.container}
      isVisible={visible}
      onBackdropPress={toggleOverlay}>
      <View style={styles.topline}>
        <Pressable style={styles.pressable} onPress={onPressClose} />
      </View>
      {backupMembers && (
        <>
          {backupMembers.map(item => (
            <MemberItem
              key={item.userID}
              email={email}
              item={item}
              setVisible={setVisible}
              setCanRecover={setCanRecover}
              setCanFamilyRecover={setCanFamilyRecover}
            />
          ))}
        </>
      )}
    </Overlay>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    marginTop: '200%',
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
    marginHorizontal: 8,
    marginVertical: 8,
  },
  headText: { textAlign: 'center' },
  groupContainer: { marginVertical: 30 },
  groupBtnContainer: { borderRadius: 8, height: 50 },
  groupBtnText: { fontSize: 18 },
  btnContainer: { width: '100%' },
  btnTitle: { fontSize: 20 },
}));

export default RequestFamilyRecoverOverlay;
