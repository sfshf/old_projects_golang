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
  CheckBox,
} from '@rneui/themed';
import { Pressable, View } from 'react-native';
import moment from 'moment';
import { Password, Record } from '../common/sqlite/schema';
import {
  getPasswordSharedDataMembersSync,
  sharingPasswordByDataIDAsync,
  xorPassword,
} from '../common/sqlite/dao/password';
import {
  getRecordSharedDataMembersSync,
  sharingRecordByDataIDAsync,
  xorRecord,
} from '../common/sqlite/dao/record';
import {
  updatePasswordSharedDataMembersAsync,
  updateRecordSharedDataMembersAsync,
} from '../common/sqlite/dao/sharedDataMember';
import {
  currentOtherFamilyMembers,
  getFamilyKey,
  OtherFamilyMember,
} from '../services/family';
import { UnlockPasswordContext } from '../contexts/unlockPassword';
import { SlarkInfoContext } from '../contexts/slark';
import { BackdropContext } from '../contexts/backdrop';
import { SnackbarContext } from '../contexts/snackbar';
import { post } from '../common/http/post';
import { encryptByXchacha20poly1305 } from '../services/cipher';
import { Response } from '../common/http';
import { currentUnlockPasswordSetting } from '../services/unlockPassword';
import { xor_str } from '../common/sqlite/dao/utils';

type MemberItemProps = {
  item: OtherFamilyMember;
  shareToAll: boolean;
  sharedTos: number[];
  setSharedTos: (val: number[]) => void;
};

const MemberItem = ({
  item,
  shareToAll,
  sharedTos,
  setSharedTos,
}: MemberItemProps) => {
  const styles = useMemberItemStyles();
  const [check, setCheck] = React.useState(false);

  React.useEffect(() => {
    for (let i = 0; i < sharedTos.length; i++) {
      if (sharedTos[i] === item.userID) {
        setCheck(true);
        return;
      }
    }
    setCheck(false);
  }, [item, sharedTos]);

  const onPressCheck = () => {
    if (!check) {
      setSharedTos([...sharedTos, item.userID]);
    } else {
      let newSlice = [];
      for (let i = 0; i < sharedTos.length; i++) {
        if (sharedTos[i] !== item.userID) {
          newSlice.push(sharedTos[i]);
        }
      }
      setSharedTos(newSlice);
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
        disabled={shareToAll}
      />
    </ListItem>
  );
};

const useMemberItemStyles = makeStyles(theme => ({
  avatarContainer: { backgroundColor: theme.colors.surface },
  itemSubtitle: { fontSize: 12 },
}));

export type FamilyShareOverlayProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  entity: Password | Record;
  setEntity: (newEntity: any) => void;
};

function FamilyShareOverlay({
  visible,
  setVisible,
  entity,
  setEntity,
}: FamilyShareOverlayProps): React.JSX.Element {
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
  const [otherFamilyMembers, setOtherFamilyMembers] = React.useState<
    OtherFamilyMember[]
  >([]);
  const [sharingMembers, setSharingMembers] = React.useState<number[]>([]); //正在分享中的成员
  const [checkedMembers, setCheckedMembers] = React.useState<number[]>([]); // 勾选中的成员
  const [sharedToAll, setSharedToAll] = React.useState(false);
  const [shareToAll, setShareToAll] = React.useState(false);
  const [type, setType] = React.useState<string>('');

  const onShow = () => {
    try {
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
      const curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      // （1）获取家庭信息
      const members = currentOtherFamilyMembers(slarkInfo.userID);
      if (members) {
        setOtherFamilyMembers(members.list);
      } else {
        setOtherFamilyMembers([]);
      }
      // （2）获取数据的分享信息
      let sharedDataMembers = null;
      if (type === 'password') {
        sharedDataMembers = getPasswordSharedDataMembersSync(
          curSetting!.passwordHash,
          xor_str(curSetting!.passwordHash, entity.dataID),
        );
      } else {
        sharedDataMembers = getRecordSharedDataMembersSync(
          curSetting!.passwordHash,
          xor_str(curSetting!.passwordHash, entity.dataID),
        );
      }
      setSharingMembers(sharedDataMembers ? sharedDataMembers : []);
      setCheckedMembers(sharedDataMembers ? sharedDataMembers : []);
      if (sharedDataMembers) {
        if (sharedDataMembers.length === 0) {
          if (members) {
            for (let i = 0; i < members.list.length; i++) {
              sharedDataMembers.push(members.list[i].userID);
            }
            setCheckedMembers(sharedDataMembers);
          }
        }
      }
    } catch (error) {
      setError(error as string, t('app.toast.error'));
      return;
    }
  };

  const onPressClose = () => {
    setVisible(false);
  };

  React.useEffect(() => {
    setType(
      (entity as Record).recordType
        ? (entity as Record).recordType
        : 'password',
    );
    setSharedToAll(entity.sharedToAll === 1);
    setShareToAll(entity.sharedToAll === 1);
  }, [entity]);

  const onPressShareToAll = () => {
    if (!slarkInfo) {
      return;
    }
    if (!shareToAll) {
      let newSlice = [];
      for (let i = 0; i < otherFamilyMembers.length; i++) {
        if (otherFamilyMembers[i].userID != slarkInfo.userID) {
          newSlice.push(otherFamilyMembers[i].userID);
        }
      }
      setCheckedMembers(newSlice);
    }
    setShareToAll(!shareToAll);
  };
  const updateSharedData = React.useCallback(
    async (entity: Password | Record) => {
      if (!slarkInfo) {
        return;
      }
      const curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      let newSlice: number[] = [];
      otherFamilyMembers.map(item => {
        if (item.userID != slarkInfo.userID) {
          newSlice.push(item.userID);
        }
      });
      if (type === 'password') {
        await sharingPasswordByDataIDAsync(
          xor_str(curSetting!.passwordHash, entity.dataID),
          entity.sharedAt !== null
            ? xor_str(curSetting!.passwordHash, entity.sharedAt.toString())
            : null,
          entity.sharedToAll !== null
            ? xor_str(curSetting!.passwordHash, entity.sharedToAll.toString())
            : null,
        );
        await updatePasswordSharedDataMembersAsync(
          xor_str(curSetting!.passwordHash, entity.dataID),
          xor_str(
            curSetting!.passwordHash,
            JSON.stringify(shareToAll ? newSlice : checkedMembers),
          ),
        );
      } else {
        await sharingRecordByDataIDAsync(
          xor_str(curSetting!.passwordHash, entity.dataID),
          entity.sharedAt !== null
            ? xor_str(curSetting!.passwordHash, entity.sharedAt.toString())
            : null,
          entity.sharedToAll !== null
            ? xor_str(curSetting!.passwordHash, entity.sharedToAll.toString())
            : null,
        );
        await updateRecordSharedDataMembersAsync(
          xor_str(curSetting!.passwordHash, entity.dataID),
          xor_str(
            curSetting!.passwordHash,
            JSON.stringify(shareToAll ? newSlice : checkedMembers),
          ),
        );
      }
    },
    [slarkInfo, otherFamilyMembers, checkedMembers, shareToAll, type],
  );
  const handleSharingData = React.useCallback(
    async (newEntity: Password | Record): Promise<Response> => {
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
      const curSetting = currentUnlockPasswordSetting(slarkInfo.userID);
      let respData;
      if (entity.sharedAt && entity.sharedAt > 0) {
        respData = await post('/pswds/manageSharingData/v1', {
          dataID: entity.dataID,
          sharingMembers: shareToAll ? [] : checkedMembers,
          stop: checkedMembers.length === 0,
        });
      } else {
        respData = await post('/pswds/shareDataToFamily/v1', {
          sharingMembers: shareToAll ? [] : checkedMembers,
          dataID: entity.dataID,
          type,
          content: encryptByXchacha20poly1305(
            await getFamilyKey(password),
            JSON.stringify(
              type === 'password'
                ? xorPassword(curSetting!.passwordHash, newEntity as Password)
                : xorRecord(curSetting!.passwordHash, newEntity as Record),
            ),
          ),
        });
      }
      await updateSharedData(newEntity);
      return respData;
    },
    [slarkInfo, shareToAll, checkedMembers, password, entity],
  );
  const onPressCommit = async () => {
    try {
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
      setVisible(false);
      const nowTS = moment().unix();
      let newEntity: Password | Record = {
        ...entity,
        updatedAt: nowTS,
        sharedAt: nowTS,
        sharedToAll: shareToAll ? 1 : null,
      };
      if (checkedMembers.length == 0) {
        newEntity.sharingMembers = null;
        newEntity.sharedAt = null;
        newEntity.sharedToAll = null;
      }
      setLoading(true);
      const respData = await handleSharingData(newEntity);
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.requestError'));
        return;
      }
      setEntity(newEntity);
      setLoading(false);
      setSuccess(t('app.toast.success'));
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
      {slarkInfo && otherFamilyMembers && (
        <>
          {otherFamilyMembers
            .filter(item => item.userID != slarkInfo.userID)
            .map(item => (
              <MemberItem
                key={item.userID}
                shareToAll={shareToAll}
                sharedTos={checkedMembers}
                setSharedTos={setCheckedMembers}
                item={item}
              />
            ))}
          <View style={styles.row}>
            <CheckBox
              checked={shareToAll}
              onPress={onPressShareToAll}
              title={t('familyShare.overlay.shareToAll')}
            />
          </View>
          <View style={styles.row}>
            <Button
              type="solid"
              disabled={
                shareToAll !== sharedToAll
                  ? false
                  : entity.sharedAt != null && entity.sharedAt > 0
                  ? checkedMembers.sort().toString() ===
                    sharingMembers.sort().toString()
                  : checkedMembers.length === 0
              }
              radius={8}
              color={theme.colors.primary}
              containerStyle={styles.btnContainer}
              titleStyle={styles.btnTitle}
              title={
                entity.sharedAt == null || entity.sharedAt === 0
                  ? t('familyShare.overlay.shareBtn')
                  : checkedMembers.length === 0
                  ? t('familyShare.overlay.stopShareBtn')
                  : t('familyShare.overlay.changeShareBtn')
              }
              onPress={onPressCommit}
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

export default FamilyShareOverlay;
