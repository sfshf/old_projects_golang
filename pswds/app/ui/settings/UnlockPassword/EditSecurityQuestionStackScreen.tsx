/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Text, makeStyles, useTheme, Button, Input } from '@rneui/themed';
import { keccak_256 } from '@noble/hashes/sha3';
import { xchacha20poly1305 } from '@noble/ciphers/chacha';
import { useFocusEffect } from '@react-navigation/native';
import { encrypt } from 'eciesjs';
import { SafeAreaView, ScrollView, KeyboardAvoidingView } from 'react-native';
import moment from 'moment';
import { syncBackup, updateBackupState } from '../../../services/backup';
import {
  currentSecurityQuestionSetting,
  updateSecurityQuestionSetting,
} from '../../../services/unlockPassword';
import { UnlockPasswordContext } from '../../../contexts/unlockPassword';
import { SlarkInfoContext } from '../../../contexts/slark';
import { BackdropContext } from '../../../contexts/backdrop';
import { SnackbarContext } from '../../../contexts/snackbar';
import { post } from '../../../common/http/post';

interface Questions {
  question1: string;
  question2: string;
  question3: string;
}

interface Answers {
  answer1: string;
  answer2: string;
  answer3: string;
}

export const PswdsApiKey_KeyID = 'pswd_pswds';
export const PswdsApiKey = 'fJVfDpPciWfym6KK6dblaEmw';
export const PswdsPrivateKey_KeyID = 'pswds-001';

type EditSecurityQuestionStackScreenProp =
  NativeStackScreenProps<RootStackParamList>;

function EditSecurityQuestionStackScreen({
  navigation,
}: EditSecurityQuestionStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { setSuccess, setError } = React.useContext(SnackbarContext);
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { password } = React.useContext(UnlockPasswordContext);
  const [publickKey, setPublicKey] = React.useState(''); // hex string
  const [questions, setQuestions] = React.useState<null | Questions>(null);
  const [answers, setAnswers] = React.useState<null | Answers>(null);
  const { setLoading } = React.useContext(BackdropContext);
  const validateParameters = (): boolean => {
    try {
      if (!slarkInfo) {
        setError(t('app.toast.notSignedIn'), t('app.toast.error'));
        return false;
      }
      // validate questions and answers
      if (
        !questions ||
        !questions.question1 ||
        !questions.question2 ||
        !questions.question3
      ) {
        setError(
          t(
            'settings.unlockPassword.securityQuestion.editSecurityQuestion.toast.emptyQuestion',
          ),
          t('app.toast.error'),
        );
        return false;
      }
      if (
        !answers ||
        !answers.answer1 ||
        !answers.answer2 ||
        !answers.answer3
      ) {
        setError(
          t(
            'settings.unlockPassword.securityQuestion.editSecurityQuestion.toast.emptyAnswer',
          ),
          t('app.toast.error'),
        );
        return false;
      }
      return true;
    } catch (error) {
      throw error;
    }
  };

  const genReqData = React.useCallback(
    (updatedAt: number): any => {
      const answerHash = Buffer.from(
        keccak_256(
          answers!.answer1.toLowerCase() +
            answers!.answer2.toLowerCase() +
            answers!.answer3.toLowerCase(),
        ),
      ).toString('hex');
      const nonce = new Uint8Array(24);
      crypto.getRandomValues(nonce);
      const hashOfHash = Buffer.from(keccak_256(answerHash)).toString('hex');
      let plaintext = JSON.stringify({
        question1: questions!.question1,
        question2: questions!.question2,
        question3: questions!.question3,
        encryptedPassword: Buffer.from(
          xchacha20poly1305(
            new Uint8Array(Buffer.from(answerHash, 'utf-8')).slice(0, 32),
            nonce,
          ).encrypt(new Uint8Array(Buffer.from(password, 'utf-8'))),
        ).toString('base64'),
        answerHash,
        hashOfHash,
        nonce: Buffer.from(nonce).toString('base64'),
      });
      // upload the cipherText
      return {
        updatedAt,
        securityQuestions: JSON.stringify(questions),
        securityQuestionsCiphertext: Buffer.from(
          encrypt(
            publickKey,
            Buffer.from(
              Buffer.from(plaintext, 'utf-8').toString('base64'),
              'utf-8',
            ),
          ),
        ).toString('base64'),
      };
    },
    [answers, questions, slarkInfo],
  );

  const uploadSecurityQuestions = React.useCallback(async () => {
    try {
      const updatedAt = moment().unix();
      const respData = await post(
        '/pswds/uploadSecurityQuestions/v1',
        await genReqData(updatedAt),
      );
      if (!respData) {
        setLoading(false);
        setError(
          t(
            'settings.securityQuestion.editSecurityQuestion.toast.uploadSecurityQuestionAnswersError',
          ),
          t('app.toast.internalError'),
        );
        return;
      } else {
        if (respData.code !== 0) {
          setLoading(false);
          setError(respData.message, t('app.toast.internalError'));
          return;
        }
      }
      // 更新用户本地备份的更新时间戳
      updateBackupState(slarkInfo!.userID, {
        updatedAt: updatedAt,
      });
      // 更新本地密保问题
      updateSecurityQuestionSetting(slarkInfo!.userID, {
        questions: JSON.stringify(questions),
      });
      setLoading(false);
      setSuccess(respData.message, t('app.toast.success'));
      navigation.goBack();
    } catch (error) {
      throw error;
    }
  }, [slarkInfo, password, answers, questions]);

  const doCommit = async () => {
    try {
      const valid = validateParameters();
      if (!valid) {
        return;
      }
      // 1. sync user data
      setLoading(true);
      const syncRespData = await syncBackup(slarkInfo!, password);
      if (syncRespData.code !== 0) {
        setLoading(false);
        setError(syncRespData.message, t('app.toast.requestError'));
        return;
      }
      // 2. upload security questions
      await uploadSecurityQuestions();
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  };

  const checkPublicKey = async () => {
    try {
      setLoading(true);
      const respData1 = await post('/riki/checkKeyExisting/v1', {
        apiKey: PswdsApiKey,
        keyID: PswdsPrivateKey_KeyID,
      });
      if (respData1.code !== 0 || !respData1.data.existing) {
        setLoading(false);
        setError(respData1.message, t('app.toast.internalError'));
        return;
      }
      const respData2 = await post('/riki/getPublicKey/v1', {
        apiKey: PswdsApiKey,
        keyID: PswdsPrivateKey_KeyID,
      });

      if (respData2.code !== 0) {
        setLoading(false);
        setError(respData2.message, t('app.toast.internalError'));
        return;
      }
      setPublicKey(respData2.data.publicKey);
      setLoading(false);
    } catch (error) {
      setLoading(false);
      setError(error as string, t('app.toast.internalError'));
    }
  };

  useFocusEffect(
    React.useCallback(() => {
      if (!slarkInfo) {
        setError(t('app.toast.notSignedIn'), t('app.toast.error'));
        navigation.goBack();
        return;
      }
      navigation.setOptions({
        headerBackButtonDisplayMode: 'minimal',
        title:
          currentSecurityQuestionSetting(slarkInfo.userID) !== null
            ? t(
                'settings.unlockPassword.securityQuestion.editSecurityQuestion.label2',
              )
            : t(
                'settings.unlockPassword.securityQuestion.editSecurityQuestion.label1',
              ),
      });
      let curSecurityQuestionSetting = currentSecurityQuestionSetting(
        slarkInfo.userID,
      );
      if (curSecurityQuestionSetting) {
        if (curSecurityQuestionSetting.questions) {
          setQuestions(JSON.parse(curSecurityQuestionSetting.questions));
        }
      }
      checkPublicKey();
    }, [slarkInfo, navigation, t, setError]),
  );

  const onChangeQuestion1 = (newText: string) => {
    newText = newText.trim();
    if (questions) {
      setQuestions({
        ...questions,
        question1: newText,
      });
    } else {
      setQuestions({
        question1: newText,
        question2: '',
        question3: '',
      });
    }
  };

  const onChangeAnswer1 = (newText: string) => {
    newText = newText.trim();
    if (answers) {
      setAnswers({
        ...answers,
        answer1: newText,
      });
    } else {
      setAnswers({
        answer1: newText,
        answer2: '',
        answer3: '',
      });
    }
  };

  const onChangeQuestion2 = (newText: string) => {
    newText = newText.trim();
    if (questions) {
      setQuestions({
        ...questions,
        question2: newText,
      });
    } else {
      setQuestions({
        question1: '',
        question2: newText,
        question3: '',
      });
    }
  };

  const onChangeAnswer2 = (newText: string) => {
    newText = newText.trim();
    if (answers) {
      setAnswers({
        ...answers,
        answer2: newText,
      });
    } else {
      setAnswers({
        answer1: '',
        answer2: newText,
        answer3: '',
      });
    }
  };

  const onChangeQuestion3 = (newText: string) => {
    newText = newText.trim();
    if (questions) {
      setQuestions({
        ...questions,
        question3: newText,
      });
    } else {
      setQuestions({
        question1: '',
        question2: '',
        question3: newText,
      });
    }
  };

  const onChangeAnswer3 = (newText: string) => {
    newText = newText.trim();
    if (answers) {
      setAnswers({
        ...answers,
        answer3: newText,
      });
    } else {
      setAnswers({
        answer1: '',
        answer2: '',
        answer3: newText,
      });
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView>
        <ScrollView>
          <View style={styles.column}>
            <Text style={styles.inputLabel}>
              {t('settings.unlockPassword.securityQuestion.question1Title')}
            </Text>
            <Input
              autoCapitalize={'none'}
              multiline
              scrollEnabled
              containerStyle={[styles.inputContainer, styles.inputContainer2]}
              inputContainerStyle={styles.inputInnerContainer}
              value={questions?.question1}
              onChangeText={onChangeQuestion1}
            />
          </View>
          <View style={styles.column}>
            <Text style={styles.inputLabel}>
              {t('settings.unlockPassword.securityQuestion.answer1Title')}
            </Text>
            <Input
              autoCapitalize={'none'}
              multiline
              scrollEnabled
              containerStyle={[styles.inputContainer, styles.inputContainer2]}
              inputContainerStyle={styles.inputInnerContainer}
              value={answers?.answer1}
              onChangeText={onChangeAnswer1}
            />
          </View>
          <View style={styles.column}>
            <Text style={styles.inputLabel}>
              {t('settings.unlockPassword.securityQuestion.question2Title')}
            </Text>
            <Input
              autoCapitalize={'none'}
              multiline
              scrollEnabled
              containerStyle={[styles.inputContainer, styles.inputContainer2]}
              inputContainerStyle={styles.inputInnerContainer}
              value={questions?.question2}
              onChangeText={onChangeQuestion2}
            />
          </View>
          <View style={styles.column}>
            <Text style={styles.inputLabel}>
              {t('settings.unlockPassword.securityQuestion.answer2Title')}
            </Text>
            <Input
              autoCapitalize={'none'}
              multiline
              scrollEnabled
              containerStyle={[styles.inputContainer, styles.inputContainer2]}
              inputContainerStyle={styles.inputInnerContainer}
              value={answers?.answer2}
              onChangeText={onChangeAnswer2}
            />
          </View>
          <View style={styles.column}>
            <Text style={styles.inputLabel}>
              {t('settings.unlockPassword.securityQuestion.question3Title')}
            </Text>
            <Input
              autoCapitalize={'none'}
              multiline
              scrollEnabled
              containerStyle={[styles.inputContainer, styles.inputContainer2]}
              inputContainerStyle={styles.inputInnerContainer}
              value={questions?.question3}
              onChangeText={onChangeQuestion3}
            />
          </View>
          <View style={styles.column}>
            <Text style={styles.inputLabel}>
              {t('settings.unlockPassword.securityQuestion.answer3Title')}
            </Text>
            <Input
              autoCapitalize={'none'}
              multiline
              scrollEnabled
              containerStyle={[styles.inputContainer, styles.inputContainer2]}
              inputContainerStyle={styles.inputInnerContainer}
              value={answers?.answer3}
              onChangeText={onChangeAnswer3}
            />
          </View>
          <View style={styles.column}>
            <Button
              title={t('app.alert.commitBtn')}
              containerStyle={styles.commitBtn}
              titleStyle={styles.commitTitle}
              color={theme.colors.primary}
              size="lg"
              radius={8}
              onPress={doCommit}
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
    marginHorizontal: 8,
    paddingHorizontal: 8,
  },
  column: {
    flex: 1,
    margin: 10,
  },
  inputLabel: {
    fontSize: 20,
    fontWeight: 400,
    marginBottom: 20,
    color: theme.colors.black,
  },
  inputContainer: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderRadius: 8,
    borderColor: theme.colors.black,
    color: theme.colors.black,
  },
  inputContainer2: {
    height: 60,
  },
  inputInnerContainer: {
    borderBottomWidth: 0,
  },
  commitBtn: { marginTop: 20, height: 60 },
  commitTitle: { fontSize: 20 },
}));

export default EditSecurityQuestionStackScreen;
