import React from 'react';
import { Modal, View } from 'react-native';
import { makeStyles, Text, Icon } from '@rneui/themed';
import { CameraView, CameraType, BarcodeScanningResult } from 'expo-camera';
import { useTranslation } from 'react-i18next';

const QRScannerView = ({
  cameraViewRef,
  onBarcodeScanned,
  setQrscan,
}: {
  cameraViewRef: React.MutableRefObject<CameraView | null>;
  onBarcodeScanned: (result: BarcodeScanningResult) => void;
  setQrscan: (scan: boolean) => void;
}) => {
  const styles = useStyles();
  const { t } = useTranslation();

  const [facing, setFacing] = React.useState<CameraType>('back');
  const toggleCameraFacing = () => {
    setFacing(current => (current === 'back' ? 'front' : 'back'));
  };
  const onPressClose = () => {
    setQrscan(false);
  };
  return (
    <Modal transparent={true}>
      <CameraView
        ref={cameraViewRef}
        barcodeScannerSettings={{
          barcodeTypes: ['qr'],
        }}
        onBarcodeScanned={onBarcodeScanned}
        style={styles.camera}
        facing={facing}>
        <View style={styles.header}>
          <Text style={styles.text}>
            {t('passwords.passwordDetail.qrcodeScanner.label')}
          </Text>
        </View>
        <View style={styles.body}>
          <View style={styles.frame} />
        </View>
        <View style={styles.footer}>
          <Icon
            size={60}
            type="material"
            name="flip-camera-ios"
            onPress={toggleCameraFacing}
          />
          <Icon size={60} type="material" name="close" onPress={onPressClose} />
        </View>
      </CameraView>
    </Modal>
  );
};

const useStyles = makeStyles(theme => ({
  camera: {
    flex: 1,
  },
  header: {
    flex: 1,
    flexDirection: 'row',
    backgroundColor: 'transparent',
    marginVertical: 10,
    marginHorizontal: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  body: {
    flex: 4,
    flexDirection: 'row',
    backgroundColor: 'transparent',
    justifyContent: 'center',
    alignItems: 'center',
  },
  frame: {
    marginHorizontal: 50,
    width: 350,
    height: 350,
    borderWidth: 3,
    borderColor: 'green',
  },
  footer: {
    flex: 1,
    flexDirection: 'row',
    backgroundColor: 'transparent',
    margin: 64,
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  text: {
    fontSize: 24,
    fontWeight: 'bold',
    color: theme.colors.white,
  },
}));

export default QRScannerView;
