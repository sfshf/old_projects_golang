import 'fast-text-encoding';
import 'react-native-url-polyfill/auto';
import 'react-native-get-random-values';
// eslint-disable-next-line no-unused-vars
import Joi from 'joi';
import {install} from 'react-native-quick-crypto';

install();

global.Buffer = require('buffer').Buffer;
