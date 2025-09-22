/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { View, TouchableOpacity } from 'react-native';
import { RootStackParamList } from '../../../navigation/routes';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { useTranslation } from 'react-i18next';
import { Text, makeStyles, useTheme, Icon } from '@rneui/themed';
import { useFocusEffect } from '@react-navigation/native';
import { SafeAreaView } from 'react-native';
import { SlarkInfoContext } from '../../../contexts/slark';
import { currentSecurityQuestionSetting } from '../../../services/unlockPassword';

type SecurityQuestionStackScreenProp =
  NativeStackScreenProps<RootStackParamList>;

function SecurityQuestionStackScreen({
  navigation,
}: SecurityQuestionStackScreenProp): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const [has, setHas] = React.useState(false);

  useFocusEffect(
    React.useCallback(() => {
      if (!slarkInfo) {
        navigation.goBack();
        return;
      }
      let curSetting = currentSecurityQuestionSetting(slarkInfo.userID);
      setHas(curSetting !== null);
    }, [slarkInfo, navigation, setHas]),
  );

  React.useEffect(() => {
    // 导航栏
    navigation.setOptions({
      headerBackVisible: true,
      headerBackButtonDisplayMode: 'minimal',
    });
  }, [navigation]);

  const onPressSecurityQuestionDetail = () => {
    navigation.navigate('SecurityQuestionDetailStack');
  };

  const onPressEditSecurityQuestion = () => {
    navigation.navigate('EditSecurityQuestionStack');
  };

  return (
    <SafeAreaView style={styles.container}>
      <View>
        {has && (
          <>
            <View style={[styles.row, styles.wraped]}>
              <View style={[styles.label]}>
                <Text style={styles.labelText}>
                  {t(
                    'settings.unlockPassword.securityQuestion.hasSecurityQuestion',
                  )}
                </Text>
              </View>
            </View>
            <TouchableOpacity
              style={[styles.row, styles.wraped]}
              onPress={onPressSecurityQuestionDetail}>
              <View style={[styles.label]}>
                <Text style={styles.labelText}>
                  {t(
                    'settings.unlockPassword.securityQuestion.securityQuestionDetail.label',
                  )}
                </Text>
              </View>
              <Icon
                size={20}
                name="arrow-forward-ios"
                color={theme.colors.black}
              />
            </TouchableOpacity>
          </>
        )}
        <TouchableOpacity
          style={[styles.row, styles.wraped]}
          onPress={onPressEditSecurityQuestion}>
          <View style={[styles.label]}>
            <Text style={styles.labelText}>
              {has
                ? t(
                    'settings.unlockPassword.securityQuestion.editSecurityQuestion.label2',
                  )
                : t(
                    'settings.unlockPassword.securityQuestion.editSecurityQuestion.label1',
                  )}
            </Text>
          </View>
          <Icon size={20} name="arrow-forward-ios" color={theme.colors.black} />
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
}

const useStyles = makeStyles(() => ({
  container: {
    flex: 1,
  },
  row: {
    width: '100%',
    flexDirection: 'row',
    marginHorizontal: 8,
    paddingHorizontal: 8,
    padding: 10,
  },
  wraped: { flexWrap: 'wrap' },
  label: {
    flex: 2,
  },
  labelText: {
    fontSize: 16,
    fontWeight: 500,
  },
}));

export default SecurityQuestionStackScreen;
