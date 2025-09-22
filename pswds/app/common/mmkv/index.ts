import {MMKV, Mode} from 'react-native-mmkv';

export const storage = new MMKV({
  id: 'pswds-storage',
  encryptionKey: 'fJVfDpPciWfym6KK6dblaEmw',
  mode: Mode.MULTI_PROCESS,
  readOnly: false,
});
