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
  Button,
} from '@rneui/themed';
import { Pressable, View } from 'react-native';
import { OtherFamilyMember } from '../services/family';
import { UnlockPasswordContext } from '../contexts/unlockPassword';
import { SlarkInfoContext } from '../contexts/slark';
import { BackdropContext } from '../contexts/backdrop';
import { SnackbarContext } from '../contexts/snackbar';
import { post } from '../common/http/post';
import { encryptedByUserPublicKey } from '../services/cipher';
import { FamilyMember } from '../ui/settings/Family/FamilyStackScreen';
import { FamilyBackup } from '../ui/settings/Family/BackupUnlockPassword';

type MemberItemProps = {
  item: OtherFamilyMember;
  backupedTos: number[];
  setBackupedTos: (val: number[]) => void;
};

const MemberItem = ({ item, backupedTos, setBackupedTos }: MemberItemProps) => {
  const styles = useMemberItemStyles();
  const [check, setCheck] = React.useState(false);

  React.useEffect(() => {
    for (let i = 0; i < backupedTos.length; i++) {
      if (backupedTos[i] === item.userID) {
        setCheck(true);
        return;
      }
    }
    setCheck(false);
  }, [item, backupedTos]);

  const onPressCheck = () => {
    if (!check) {
      setBackupedTos([...backupedTos, item.userID]);
    } else {
      let newSlice = [];
      for (let i = 0; i < backupedTos.length; i++) {
        if (backupedTos[i] !== item.userID) {
          newSlice.push(backupedTos[i]);
        }
      }
      setBackupedTos(newSlice);
    }
    setCheck(!check);
  };

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
      <ListItem.CheckBox
        // Use ThemeProvider to change the defaults of the checkbox
        iconType="material-community"
        checkedIcon="checkbox-marked"
        uncheckedIcon="checkbox-blank-outline"
        checked={check}
        onPress={onPressCheck}
      />
    </ListItem>
  );
};

const useMemberItemStyles = makeStyles(theme => ({
  avatarContainer: { backgroundColor: theme.colors.surface },
  itemSubtitle: { fontSize: 12 },
}));

export type SetupFamilyBackupOverlayProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  familyMembers: FamilyMember[];
  backupings: FamilyBackup[];
  callback: () => void;
};

function SetupFamilyBackupOverlay({
  visible,
  setVisible,
  familyMembers,
  backupings,
  callback,
}: SetupFamilyBackupOverlayProps): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { password } = React.useContext(UnlockPasswordContext);

  const toggleOverlay = () => {
    setVisible(!visible);
  };

  const { setLoading } = React.useContext(BackdropContext);
  const [backupedMembers, setBackupedMembers] = React.useState<number[]>([]); // 已备份的成员
  const [checkedMembers, setCheckedMembers] = React.useState<number[]>([]); // 勾选中的成员

  const onShow = React.useCallback(() => {
    try {
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
      const backuped: number[] = [];
      backupings.map(item => backuped.push(item.userID));
      setBackupedMembers(backuped);
      setCheckedMembers(backuped);
    } catch (error) {
      setError(error as string, t('app.toast.error'));
      return;
    }
  }, [slarkInfo, familyMembers, backupings]);

  const onPressClose = () => {
    setVisible(false);
  };

  const onPressBackup = async () => {
    try {
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
      setVisible(false);
      setLoading(true);
      const familyBackups: any[] = [];
      checkedMembers.map(userID => {
        for (let i = 0; i < familyMembers.length; i++) {
          if (familyMembers[i].userID === userID) {
            familyBackups.push({
              userID: userID,
              email: familyMembers[i].email,
              ciphertext: encryptedByUserPublicKey(
                Buffer.from(password),
                new Uint8Array(
                  Buffer.from(familyMembers[i].userPublicKey, 'hex'),
                ),
              ).toString('hex'),
            });
            break;
          }
        }
      });
      const respData = await post('/pswds/setFamilyBackup/v1', {
        set: checkedMembers.length !== 0,
        familyBackups,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      setLoading(false);
      setSuccess(t('app.toast.success'));
      callback();
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
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
      {slarkInfo && familyMembers && (
        <>
          {familyMembers
            .filter(item => item.userID != slarkInfo.userID)
            .map(item => (
              <MemberItem
                key={item.userID}
                backupedTos={checkedMembers}
                setBackupedTos={setCheckedMembers}
                item={item}
              />
            ))}
          <View style={styles.row}>
            <Button
              type="solid"
              disabled={
                checkedMembers.sort().toString() ===
                backupedMembers.sort().toString()
              }
              radius={8}
              color={theme.colors.primary}
              containerStyle={styles.btnContainer}
              titleStyle={styles.btnTitle}
              title={t('settings.family.setupFamilyBackupOverlay.backupBtn')}
              onPress={onPressBackup}
            />
          </View>
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

export default SetupFamilyBackupOverlay;
