/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { makeStyles, Input, Button, Card } from '@rneui/themed';
import { View, Alert, ScrollView, KeyboardAvoidingView } from 'react-native';
import { z } from 'zod';
import { SafeAreaView } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';
import { encryptedByUserPublicKey } from '../../../services/cipher';
import { getFamilyKey } from '../../../services/family';

type InviteFamilyMemberStackScreenProp = NativeStackScreenProps<
  RootStackParamList,
  'InviteFamilyMemberStack'
>;

function InviteFamilyMemberStackScreen({
  navigation,
  route,
}: InviteFamilyMemberStackScreenProp): React.JSX.Element {
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const { t } = useTranslation();
  const styles = useStyles();
  const { setSuccess, setError, setWarning } =
    React.useContext(SnackbarContext);
  const { setLoading } = React.useContext(BackdropContext);
  const [email, setEmail] = React.useState('');

  const onChangeEmail = (newText: string) => {
    setEmail(newText);
  };

  const doInvite = React.useCallback(
    async (userPublicKey: string) => {
      try {
        if (!userPublicKey) {
          throw t(
            'settings.family.inviteFamilyMember.toast.emptyUserPublicKey',
          );
        }
        if (!slarkInfo) {
          throw t('app.toast.notSignedIn');
        }
        if (!password) {
          throw t('app.toast.emptyUnlockPassword');
        }
        const encryptedKey = encryptedByUserPublicKey(
          await getFamilyKey(password),
          new Uint8Array(Buffer.from(userPublicKey, 'hex')),
        );
        // post
        const respData = await post('/pswds/inviteFamilyMember/v1', {
          email,
          encryptedFamilyKey: encryptedKey.toString('hex'),
        });
        setLoading(false);
        if (respData.code !== 0) {
          setError(respData.message, t('app.toast.error'));
          return;
        }
        navigation.goBack();
        setSuccess(t('app.toast.success'));
      } catch (error) {
        setError(error as string, t('app.toast.internalError'));
        setLoading(false);
      }
    },
    [slarkInfo, password, email],
  );

  const emailSchema = z
    .string()
    .regex(
      /^(([^<>()[\]\.,;:\s@\"]+(\.[^<>()[\]\.,;:\s@\"]+)*)|(\".+\"))@(([^<>()[\]\.,;:\s@\"]+\.)+[^<>()[\]\.,;:\s@\"]{2,})$/,
      {
        message: t('settings.family.inviteFamilyMember.toast.malformedEmail'),
      },
    );

  const alertInvitationState = (message: string) => {
    Alert.alert('', message, [
      {
        text: t('settings.family.inviteFamilyMember.okBtn'),
        style: 'cancel',
      },
    ]);
  };

  const checkInvitationState = React.useCallback(
    async (data: any) => {
      switch (data.state) {
        case 'no_user':
          setLoading(false);
          alertInvitationState(
            t('settings.family.inviteFamilyMember.toast.invitationStateNoUser'),
          );
          return;
        case 'has_family':
          setLoading(false);
          alertInvitationState(
            t(
              'settings.family.inviteFamilyMember.toast.invitationStateHasFamily',
            ),
          );
          return;
        case 'has_invited':
          setLoading(false);
          alertInvitationState(
            t(
              'settings.family.inviteFamilyMember.toast.invitationStateHasInvited',
            ),
          );
          return;
        case 'invitable':
          await doInvite(data.userPublicKey);
          return;
      }
    },
    [email],
  );

  const inviteUser = React.useCallback(async () => {
    try {
      if (!email) {
        throw t('settings.family.inviteFamilyMember.toast.emptyEmail');
      }
      const validation = emailSchema.safeParse(email);
      if (!validation.success) {
        setError(validation.error.issues[0].message, t('app.toast.error'));
        return;
      }
      if (!slarkInfo) {
        throw t('app.toast.notSignedIn');
      }
      setLoading(true);
      // post
      const respData = await post('/pswds/checkUserAvailable/v1', {
        email,
      });
      if (respData.code !== 0) {
        setLoading(false);
        setError(respData.message, t('app.toast.error'));
        return;
      }
      // check invitation state
      await checkInvitationState(respData.data);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  }, [slarkInfo, password, email]);

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
        <ScrollView>
          <Card containerStyle={styles.card}>
            <View style={styles.row}>
              <Input
                autoCapitalize={'none'}
                multiline
                labelStyle={styles.inputLabel}
                inputStyle={styles.inputStyle}
                placeholder={t(
                  'settings.family.inviteFamilyMember.emailPlaceholder',
                )}
                onChangeText={onChangeEmail}
                value={email}
              />
            </View>
          </Card>
          <View style={styles.row}>
            <Button
              title={t('settings.family.inviteFamilyMember.inviteBtn')}
              containerStyle={styles.commitBtn}
              titleStyle={styles.btnTitle}
              size="lg"
              radius={8}
              onPress={inviteUser}
            />
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
  },
  card: { borderRadius: 8 },
  row: {
    flex: 1,
    flexDirection: 'row',
    marginVertical: 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  inputLabel: {
    fontSize: 18,
    fontWeight: 'normal',
    color: theme.colors.black,
    marginBottom: 16,
  },
  inputStyle: {
    fontSize: 18,
    fontWeight: 'normal',
  },
  passwordLabelItem: { width: '50%' },
  randomPasswordTitle: { fontSize: 16 },
  btnTitle: { fontSize: 20, fontWeight: 'normal' },
  commitBtn: { marginVertical: 10, width: '95%' },
}));

export default InviteFamilyMemberStackScreen;
