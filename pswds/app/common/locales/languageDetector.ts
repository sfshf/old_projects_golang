import {type LanguageDetectorModule} from 'i18next';
import {getLocales} from 'react-native-localize';

const LanguageDetector: LanguageDetectorModule = {
  type: 'languageDetector',
  init: () => {},
  detect: () => getLocales()[0].languageCode,
  cacheUserLanguage: () => {},
};

export default LanguageDetector;
