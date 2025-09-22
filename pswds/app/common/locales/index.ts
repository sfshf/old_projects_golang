import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import en from './en/translation.json';
import zh from './zh/translation.json';
import LanguageDetector from './languageDetector';
import {storage} from '../mmkv';
import {getLocales} from 'react-native-localize';

const resources = {
  en: {
    translation: en,
  },
  zh: {
    translation: zh,
  },
};

export const languageSettingKey = 'cache_settings.language';

export const currentLanguage = (): string => {
  let curLang = storage.getString(languageSettingKey);
  if (!curLang) {
    return 'default';
  }
  return curLang;
};

export const changeLanguage = (lang: string) => {
  if (!lang) {
    lang = 'default';
  }
  storage.set(languageSettingKey, lang);
  if (lang === 'default') {
    lang = getLocales()[0].languageCode;
    const langs = ['en', 'zh'];
    // system default
    let support = false;
    for (let i = 0; i < langs.length; i++) {
      if (langs[i] === lang) {
        support = true;
        break;
      }
    }
    if (!support) {
      lang = 'en';
    }
  }
  i18n.changeLanguage(lang);
};

export const launchI18n = () => {
  i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
      debug: true,
      resources,
      ns: ['translation'],
      defaultNS: 'translation',
      fallbackLng: 'en',
      interpolation: {
        escapeValue: false,
      },
    });
  // load local language setting
  changeLanguage(currentLanguage());
};

export default i18n;
