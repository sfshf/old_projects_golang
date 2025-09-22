import React from 'react';
import {
  KeyboardAvoidingView,
  Platform,
  Pressable,
  SafeAreaView,
  Keyboard,
  Dimensions,
} from 'react-native';
import {useHeaderHeight} from '@react-navigation/elements';
import Animated from 'react-native-reanimated';
import {useSharedValue, withTiming, runOnJS} from 'react-native-reanimated';
import {makeStyles} from '@rneui/themed';

type Props = {
  children: React.ReactNode;
  goBack: () => void;
};

const BackgroundColor = 'rgba(0, 0, 0, 0.5)';
const TransparentColor = 'rgba(0, 0, 0, 0)';

const ModalScreen = ({children, goBack}: Props) => {
  const headerHeight = useHeaderHeight();
  const styles = useStyles();

  // may be rotated
  const windowHeight =
    Dimensions.get('window').height > Dimensions.get('window').width
      ? Dimensions.get('window').height
      : Dimensions.get('window').width;

  const translateY = useSharedValue(windowHeight);
  const backgroundColor = useSharedValue(TransparentColor);

  const showModal = () => {
    backgroundColor.value = withTiming(
      BackgroundColor,
      {duration: 200},
      finished => {
        if (finished === true) {
          translateY.value = withTiming(0);
        }
      },
    );
  };

  const dismissModal = () => {
    translateY.value = withTiming(windowHeight, {duration: 200}, finished => {
      if (finished === true) {
        backgroundColor.value = withTiming(
          TransparentColor,
          {duration: 200},
          _finished => {
            if (_finished === true) {
              // navigation.goBack();
              runOnJS(goBack)();
            }
          },
        );
      }
    });
  };

  React.useLayoutEffect(() => {
    showModal();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onClose = () => {
    if (Keyboard.isVisible()) {
      Keyboard.dismiss();
    } else {
      dismissModal();
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      keyboardVerticalOffset={headerHeight}>
      <Animated.View
        style={[styles.closeContainer, {backgroundColor: backgroundColor}]}>
        <Pressable style={styles.close} onPress={onClose} />
      </Animated.View>
      <Animated.View
        style={[
          styles.contentContainer,
          {transform: [{translateY: translateY}]},
        ]}>
        <SafeAreaView style={styles.childrenContainer}>{children}</SafeAreaView>
      </Animated.View>
    </KeyboardAvoidingView>
  );
};

const useStyles = makeStyles(theme => ({
  container: {
    flex: 1,
    flexDirection: 'column',
    justifyContent: 'flex-end',
    width: '100%',
    height: '100%',
  },
  closeContainer: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
  },
  close: {
    flex: 1,
  },
  contentContainer: {
    backgroundColor: theme.colors.surface,
    width: '100%',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    overflow: 'hidden',
  },
  childrenContainer: {height: '50%', width: '100%'},
}));

export default ModalScreen;
