/**
 * @format
 */

import { AppRegistry } from 'react-native';
import 'react-native-get-random-values';
import App from './app/App';
import { name as appName } from './app.json';
import { launchI18n } from './app/common/locales';
import './app/common/sqlite';
import { migrateTables } from './app/common/sqlite';

launchI18n();
migrateTables();

AppRegistry.registerComponent(appName, () => App);
