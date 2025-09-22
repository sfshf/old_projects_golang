import { xchacha20poly1305 } from '@noble/ciphers/chacha';
import { Cipher } from '@noble/ciphers/utils';
import { keccak_256 } from '@noble/hashes/sha3';
import moment from 'moment';
import i18n from '../locales';
import { env } from '../http/env';
import { currentDebugLogHttp } from '../../ui/debug/DebugStackScreen';
import {
  currentDebugHttpLogs,
  HttpLog,
  updateDebugHttpLogs,
} from '../../ui/debug/DebugHttpLogStackScreen';
import { CODE_SUCCESS, currentDebugUseGo, Response } from '../http';
import { insertInfoLog } from '../log';

export const post = async (
  path: string,
  data?: any,
  abortController?: AbortController,
): Promise<Response> => {
  // 1. pre handles
  const originPath = path;
  let reqData = '';
  if (data) {
    reqData = JSON.stringify(data);
  }
  let encryptKey;
  let nonce: Uint8Array;
  let aead: Cipher | undefined;
  let log: HttpLog | undefined;
  let useGo = env.API_ENCRYPTED;
  let curDebugUseGoSetting = currentDebugUseGo();
  if (curDebugUseGoSetting) {
    useGo = curDebugUseGoSetting.useGo;
  } else {
    useGo = true; // default is true
  }
  let curDebugLogHttpSetting = currentDebugLogHttp();
  if (curDebugLogHttpSetting) {
    if (curDebugLogHttpSetting.logHttp) {
      log = {
        useGo,
        start: 0,
        end: 0,
        duration: 0,
        host: '',
        path: '',
        cookie: '',
        reqBody: null,
        reqHeaders: null,
        respCode: 0,
        respHeaders: null,
        respBody: null,
      };
    } else {
      log = undefined;
    }
  }
  const startTS = moment();
  if (useGo) {
    encryptKey = keccak_256(startTS.format('x') + '9C9B913EB1B6254F4737CE947');
    nonce = new Uint8Array(24);
    aead = xchacha20poly1305(encryptKey.slice(0, 32), nonce);
    let encryptedData = Buffer.from(
      aead.encrypt(new Uint8Array(Buffer.from(reqData, 'utf-8'))),
    ).toString('base64');
    reqData = JSON.stringify({
      path: path,
      data: encryptedData,
      timestamp: startTS.valueOf(),
    });
    path = '/go';
  }
  try {
    let reqOpts: any = {};
    reqOpts.method = 'POST';
    let headers: any = {};
    headers['Content-Type'] = 'application/json'; // explicit content type: application/json
    if (reqData) {
      headers['Content-Length'] = Buffer.byteLength(reqData);
      reqOpts.body = reqData;
    }
    reqOpts.headers = headers;
    // cors config
    reqOpts.mode = 'cors';
    reqOpts.credentials = 'include';
    // abort controller
    // error: [AbortError: Aborted]
    if (abortController) {
      reqOpts.signal = abortController.signal;
    } else {
      const controller = new AbortController();
      setTimeout(() => controller.abort(), 30000);
      reqOpts.signal = controller.signal;
    }
    // 2. post
    const resp = await fetch(env.API_URL + path, reqOpts);
    // 3. handle response
    const endTS = moment();
    // 3-1. handle response status
    if (!resp.ok) {
      if (log) {
        // (1) assign values
        log.useGo = useGo;
        log.start = startTS.valueOf();
        log.end = endTS.valueOf();
        log.duration = log.end - log.start;
        log.host = env.API_URL;
        log.path = originPath;
        log.reqBody = data;
        log.reqHeaders = headers;
        log.respCode = resp.status;
        log.respHeaders = resp.headers;
        log.respBody = resp.body;
        // (2) insert log
        const curDebugHttpLogs = currentDebugHttpLogs();
        if (curDebugHttpLogs) {
          curDebugHttpLogs.logs.push(log);
          if (curDebugHttpLogs.logs.length == 21) {
            curDebugHttpLogs.logs.shift();
          }
          updateDebugHttpLogs(curDebugHttpLogs);
        } else {
          updateDebugHttpLogs({ logs: [log] });
        }
      }
      const message = `request (${path}) error: http code: ${resp.status}`;
      // debug info log
      insertInfoLog({
        timestamp: endTS.valueOf(),
        level: 'info',
        message: message,
      });
      return {
        code: resp.status,
        message: i18n.t('app.toast.requestError'),
        debugMessage: message,
        data: null,
      };
    }
    // 3-2. handle response body
    const respData = await resp.json();
    if (respData.code !== CODE_SUCCESS) {
      if (log) {
        // (1) assign values
        log.useGo = useGo;
        log.start = startTS.valueOf();
        log.end = endTS.valueOf();
        log.duration = log.end - log.start;
        log.host = env.API_URL;
        log.path = originPath;
        log.reqBody = data;
        log.reqHeaders = headers;
        log.respCode = resp.status;
        log.respHeaders = resp.headers;
        log.respBody = respData;
        // (2) insert log
        const curDebugHttpLogs = currentDebugHttpLogs();
        if (curDebugHttpLogs) {
          curDebugHttpLogs.logs.push(log);
          if (curDebugHttpLogs.logs.length == 21) {
            curDebugHttpLogs.logs.shift();
          }
          updateDebugHttpLogs(curDebugHttpLogs);
        } else {
          updateDebugHttpLogs({ logs: [log] });
        }
      }
      // debug info log
      insertInfoLog({
        timestamp: endTS.valueOf(),
        level: 'info',
        message: respData.debugMessage,
      });
      return {
        code: respData.code,
        message: respData.message,
        debugMessage: respData.debugMessage,
        data: null,
      };
    }
    if (aead) {
      if (respData.data && respData.data.encryptedData) {
        let decryptedData = aead.decrypt(
          new Uint8Array(Buffer.from(respData.data.encryptedData, 'base64')),
        );
        if (decryptedData && decryptedData.length > 0) {
          respData.data = JSON.parse(Buffer.from(decryptedData).toString());
        } else {
          respData.data = null;
        }
      }
    }
    // 3-3. handle response haeders
    if (log) {
      // (1) assign values
      log.useGo = useGo;
      log.start = startTS.valueOf();
      log.end = endTS.valueOf();
      log.duration = log.end - log.start;
      log.host = env.API_URL;
      log.path = originPath;
      log.reqBody = data;
      log.reqHeaders = headers;
      log.respCode = resp.status;
      log.respHeaders = resp.headers;
      log.respBody = respData;
      // (2) insert log
      const curDebugHttpLogs = currentDebugHttpLogs();
      if (curDebugHttpLogs) {
        curDebugHttpLogs.logs.push(log);
        if (curDebugHttpLogs.logs.length == 21) {
          curDebugHttpLogs.logs.shift();
        }
        updateDebugHttpLogs(curDebugHttpLogs);
      } else {
        updateDebugHttpLogs({ logs: [log] });
      }
    }
    return {
      code: respData.code,
      message: respData.message,
      debugMessage: respData.debugMessage,
      data: respData.data,
    };
  } catch (error) {
    let message = '';
    if (typeof error === 'string') {
      message = error.toUpperCase();
    } else if (error instanceof Error) {
      message = error.message;
    }
    return {
      code: 500,
      message: i18n.t('app.toast.requestError'),
      debugMessage: message,
      data: null,
    };
  }
};
