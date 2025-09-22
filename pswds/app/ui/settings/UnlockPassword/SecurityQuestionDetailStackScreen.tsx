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
import { Text, makeStyles, Input } from '@rneui/themed';
import { useFocusEffect } from '@react-navigation/native';
import { SafeAreaView } from 'react-native';
import { SlarkInfoContext } from '../../../contexts/slark';
import { currentSecurityQuestionSetting } from '../../../services/unlockPassword';
import { SnackbarContext } from '../../../contexts/snackbar';

interface Questions {
  question1: string;
  question2: string;
  question3: string;
}

type SecurityQuestionDetailStackScreenProp =
  NativeStackScreenProps<RootStackParamList>;

function SecurityQuestionDetailStackScreen({
  navigation,
}: SecurityQuestionDetailStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const { setError } = React.useContext(SnackbarContext);
  const [questions, setQuestions] = React.useState<null | Questions>(null);

  useFocusEffect(
    React.useCallback(() => {
      if (!slarkInfo) {
        setError(t('app.toast.notSignedIn'), t('app.toast.error'));
        navigation.goBack();
        return;
      }
      let curSecurityQuestionSetting = currentSecurityQuestionSetting(
        slarkInfo.userID,
      );
      if (curSecurityQuestionSetting) {
        if (curSecurityQuestionSetting.questions) {
          setQuestions(JSON.parse(curSecurityQuestionSetting.questions));
        }
      }
    }, [slarkInfo, navigation, setError, t]),
  );

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.column}>
        <Text style={styles.inputLabel}>
          {t('settings.unlockPassword.securityQuestion.question1Title')}
        </Text>
        <Input
          autoCapitalize={'none'}
          readOnly
          multiline
          scrollEnabled
          containerStyle={[styles.inputContainer, styles.inputContainer2]}
          inputContainerStyle={styles.inputInnerContainer}
          value={questions?.question1}
        />
      </View>
      <View style={styles.column}>
        <Text style={styles.inputLabel}>
          {t('settings.unlockPassword.securityQuestion.question2Title')}
        </Text>
        <Input
          autoCapitalize={'none'}
          readOnly
          multiline
          scrollEnabled
          containerStyle={[styles.inputContainer, styles.inputContainer2]}
          inputContainerStyle={styles.inputInnerContainer}
          value={questions?.question2}
        />
      </View>
      <View style={styles.column}>
        <Text style={styles.inputLabel}>
          {t('settings.unlockPassword.securityQuestion.question3Title')}
        </Text>
        <Input
          autoCapitalize={'none'}
          readOnly
          multiline
          scrollEnabled
          containerStyle={[styles.inputContainer, styles.inputContainer2]}
          inputContainerStyle={styles.inputInnerContainer}
          value={questions?.question3}
        />
      </View>
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
}));

export default SecurityQuestionDetailStackScreen;
