/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { useTranslation } from 'react-i18next';
import { makeStyles, Button, useTheme, Tooltip } from '@rneui/themed';
import { View, useWindowDimensions } from 'react-native';
import { Password, Record } from '../common/sqlite/schema';
import { SlarkInfoContext } from '../contexts/slark';
import { currentBackup } from '../common/mmkv/backup';

export type RecordOperationTooltipProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  entity: Password | Record;
  onPressEdit: () => void;
  onPressShare: () => void;
  onPressTransfer: () => void;
  onPressTransfer2: () => void;
  onPressDelete: () => void;
};

function RecordOperationTooltip({
  open,
  setOpen,
  entity,
  onPressEdit,
  onPressShare,
  onPressTransfer,
  onPressTransfer2,
  onPressDelete,
}: RecordOperationTooltipProps): React.JSX.Element {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { width } = useWindowDimensions();
  const styles = useStyles();
  const { slarkInfo } = React.useContext(SlarkInfoContext);
  const [hasFamily, setHasFamily] = React.useState(false);
  React.useEffect(() => {
    if (slarkInfo) {
      const backup = currentBackup();
      if (backup && backup.encryptedFamilyKey) {
        setHasFamily(true);
      } else {
        setHasFamily(false);
      }
    } else {
      setHasFamily(false);
    }
  }, [entity]);
  return (
    <Tooltip
      backgroundColor={theme.colors.grey4}
      pointerStyle={{ marginLeft: 35, left: (width / 3) * 1 }}
      containerStyle={{ marginLeft: width / 7 }}
      visible={open}
      height={180}
      width={140}
      onClose={() => {
        setOpen(false);
      }}
      popover={
        <View>
          <Button
            buttonStyle={styles.buttonStyle}
            titleStyle={styles.titleStyle}
            title={t('app.button.edit')}
            icon={{ name: 'edit', type: 'antdesign', size: 16 }}
            iconContainerStyle={styles.iconContainerStyle}
            onPress={onPressEdit}
          />
          {hasFamily && (
            <Button
              buttonStyle={styles.buttonStyle}
              titleStyle={styles.titleStyle}
              title={t('app.button.share')}
              icon={{ name: 'share', type: 'entypo', size: 16 }}
              iconContainerStyle={styles.iconContainerStyle}
              onPress={onPressShare}
            />
          )}
          <Button
            buttonStyle={styles.buttonStyle}
            titleStyle={styles.titleStyle}
            title={t('app.button.transfer')}
            icon={{
              name: 'monitor-share',
              type: 'material-community',
              size: 16,
            }}
            iconContainerStyle={styles.iconContainerStyle}
            onPress={onPressTransfer}
          />
          <Button
            buttonStyle={styles.buttonStyle}
            titleStyle={styles.titleStyle}
            title={t('app.button.transfer2')}
            icon={{
              name: 'qr-code-scanner',
              type: 'material',
              size: 16,
            }}
            iconContainerStyle={styles.iconContainerStyle}
            onPress={onPressTransfer2}
          />
          <Button
            buttonStyle={styles.buttonStyle}
            titleStyle={[styles.titleStyle, styles.deleteTitleStyle]}
            title={t('app.button.delete')}
            icon={{
              name: 'delete',
              type: 'antdesign',
              size: 16,
            }}
            iconContainerStyle={styles.iconContainerStyle}
            onPress={onPressDelete}
          />
        </View>
      }
    />
  );
}

const useStyles = makeStyles(theme => ({
  buttonStyle: {
    margin: 4,
    padding: 4,
    backgroundColor: theme.colors.grey4,
  },
  titleStyle: {
    width: 80,
    color: theme.colors.primary,
    fontSize: 10,
    textAlign: 'left',
  },
  deleteTitleStyle: { color: theme.colors.error },
  iconContainerStyle: { marginRight: 10 },
}));

export default RecordOperationTooltip;
