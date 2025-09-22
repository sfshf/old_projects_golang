import { xchacha20poly1305 } from '@noble/ciphers/chacha';
import { Cipher } from '@noble/ciphers/utils';
import { keccak_256 } from '@noble/hashes/sha3';
import moment from 'moment';
import { removeLocalUserLoginState } from '@/app/site/shared';

let password: string | null;

export const baseAPIURL = process.env.NEXT_PUBLIC_SERVER_BASE_URL;

export const getApiKey = () => {
  // try to load from local storage
  if ((!password || password === '') && typeof localStorage !== 'undefined') {
    password = localStorage.getItem('password');
  }
  return password;
};

export const setApiKey = (newPassword: string) => {
  password = newPassword;
  if (typeof localStorage !== 'undefined' && password !== '') {
    localStorage.setItem('password', password);
  }
};

export const CookieKey_Session = 'LSessionID';

export const post = async (
  useGo: boolean,
  path: string,
  setLoading: (load: boolean) => void,
  reqData?: any,
  handleRespHeaders?: (headers: any) => void,
  successAction?: (respData: any) => void,
  setToast?: (message: string) => void,
  setErrorToast?: (message: string) => void
) => {
  setLoading(true);
  let postData = '';
  if (reqData) {
    postData = JSON.stringify(reqData);
  }
  let encryptKey;
  let nonce: Uint8Array;
  let aead: Cipher;
  if (useGo) {
    let ts = moment();
    encryptKey = keccak_256(ts.format('x') + '9C9B913EB1B6254F4737CE947');
    nonce = new Uint8Array(24);
    aead = xchacha20poly1305(encryptKey.slice(0, 32), nonce);
    let encryptedData = Buffer.from(
      aead.encrypt(new Uint8Array(Buffer.from(postData, 'utf-8')))
    ).toString('base64');
    postData = JSON.stringify({
      path: path,
      data: encryptedData,
      timestamp: ts.valueOf(),
    });
    path = '/go';
  }
  try {
    let reqOpts: any = {};
    reqOpts.method = 'POST';
    let headers: any = {};
    if (postData) {
      headers = {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
      };
      reqOpts['body'] = postData;
    }
    reqOpts['headers'] = headers;
    reqOpts['mode'] = 'cors';
    reqOpts['credentials'] = 'include';
    const resp = await fetch(baseAPIURL + path, reqOpts);
    setLoading(false);
    if (!resp.ok) {
      setErrorToast && setErrorToast('response no ok: ' + resp.statusText);
      return;
    }
    // handle headers
    if (handleRespHeaders) {
      handleRespHeaders(resp.headers);
    }
    const respData = await resp.json();
    if (respData.code !== 0) {
      if (
        respData.code === 107040100 || // invoker 401x
        respData.code === 100040100 || // slark 401x
        respData.code === 100040101 || // slark 401x
        respData.code === 100040102 || // slark 401x
        respData.code === 100040103 // slark 401x
      ) {
        // 没有登录态，清理本地登录状态
        removeLocalUserLoginState();
      }
      setErrorToast && setErrorToast(respData.debugMessage);
      return;
    }
    if (useGo) {
      let decryptedData = aead!.decrypt(
        new Uint8Array(Buffer.from(respData.data.encryptedData, 'base64'))
      );
      if (decryptedData) {
        respData.data = JSON.parse(Buffer.from(decryptedData).toString());
      }
    }
    if (successAction) {
      successAction(respData);
    }
    setToast && setToast('success');
    return respData;
  } catch (error) {
    const message = (error as Error).message;
    setErrorToast && setErrorToast(message);
  }
};

export const uploadFile = async (
  setLoading: (load: boolean) => void,
  reqData?: any,
  keyPrefix?: string,
  successAction?: (respData: any) => void,
  setToast?: (message: string) => void,
  setErrorToast?: (message: string) => void
) => {
  if (!reqData) {
    setErrorToast && setErrorToast('null file');
    return;
  }
  setLoading(true);
  try {
    let reqOpts: any = {};
    reqOpts.method = 'POST';
    reqOpts['body'] = reqData;
    reqOpts['mode'] = 'cors';
    reqOpts['credentials'] = 'include';
    const resp = await fetch(
      baseAPIURL + '/upload?keyPrefix=' + keyPrefix,
      reqOpts
    );
    setLoading(false);
    if (!resp.ok) {
      setErrorToast && setErrorToast('response no ok: ' + resp.statusText);
      return;
    }
    const respData = await resp.json();
    if (respData.code !== 0) {
      if (
        respData.code === 107040100 || // invoker 401x
        respData.code === 100040100 || // slark 401x
        respData.code === 100040101 || // slark 401x
        respData.code === 100040102 || // slark 401x
        respData.code === 100040103 // slark 401x
      ) {
        // 没有登录态，清理本地登录状态
        removeLocalUserLoginState();
      }
      setErrorToast && setErrorToast(respData.debugMessage);
      return;
    }
    if (successAction) {
      successAction(respData);
    }
    setToast && setToast('success');
    return respData;
  } catch (error) {
    const message = (error as Error).message;
    setErrorToast && setErrorToast(message);
  }
};
