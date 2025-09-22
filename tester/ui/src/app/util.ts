import { xchacha20poly1305 } from '@noble/ciphers/chacha';
import { Cipher } from '@noble/ciphers/utils';
import { keccak_256 } from '@noble/hashes/sha3';
import moment from 'moment';
import * as secp from '@noble/secp256k1';
import { sha256 } from '@noble/hashes/sha256';
import { gcm } from '@noble/ciphers/aes';
import { randomBytes } from '@noble/ciphers/webcrypto';

export const DefaultCookieSessionKey = 'LSessionID';

export const post = async (
  useGo: boolean,
  sessionID: string,
  path: string,
  setLoading: (load: boolean) => void,
  useJson?: boolean,
  reqData?: any,
  successAction?: (respData: any) => void,
  setToast?: (message: string) => void,
  setErrorToast?: (message: string) => void
) => {
  setLoading(true);
  let postData = '';
  if (useJson && reqData) {
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
    if (useJson) {
      let encryptedData = Buffer.from(
        aead.encrypt(new Uint8Array(Buffer.from(postData, 'utf-8')))
      ).toString('base64');
      postData = JSON.stringify({
        path: path,
        data: encryptedData,
        timestamp: ts.valueOf(),
      });
      path = '/go';
    } else {
      let encryptedData = Buffer.from(
        aead.encrypt(new Uint8Array(Buffer.from(reqData)))
      ).toString('base64');
      postData = JSON.stringify({
        path: path,
        data: encryptedData,
        timestamp: ts.valueOf(),
      });
      path = '/go';
    }
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
    } else {
      reqOpts['body'] = reqData;
    }
    if (sessionID != '') {
      headers['X-SessionID'] =
        DefaultCookieSessionKey +
        '=' +
        sessionID +
        '; Path=/; Max-Age=600; HttpOnly';
    }
    reqOpts['headers'] = headers;
    const resp = await fetch('https://api.test.n1xt.net' + path, reqOpts);
    setLoading(false);
    const respData = await resp.json();
    if (useGo) {
      let decryptedData = aead!.decrypt(
        new Uint8Array(Buffer.from(respData.data.encryptedData, 'base64'))
      );
      if (decryptedData && decryptedData.length > 0) {
        respData.data = JSON.parse(Buffer.from(decryptedData).toString());
      } else {
        respData.data = null;
      }
    }
    if (respData.code !== 0) {
      setErrorToast && setErrorToast(respData.debugMessage);
      return respData;
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

// AES-256-GCM
export const aesAEAD = (
  key: Uint8Array,
  nonce: Uint8Array,
  AAD?: Uint8Array
): Cipher => {
  return gcm(key, nonce, AAD);
};

export const aes256GCM_secp256k1Encrypt = (
  pubKey: Uint8Array,
  plaintext: Uint8Array
): Uint8Array => {
  const ephemeralPrivKey = secp.utils.randomPrivateKey();
  const ephemeralPubKey = secp.getPublicKey(ephemeralPrivKey);
  const sharedSecret = secp.getSharedSecret(ephemeralPrivKey, pubKey);
  const cipherKey = sha256.create().update(sharedSecret).digest();
  const nonce = randomBytes(24);
  const aead = aesAEAD(cipherKey, nonce, ephemeralPubKey);
  const ciphertext = aead.encrypt(plaintext);
  const prefix = new Uint8Array([
    ephemeralPubKey.length,
    ephemeralPubKey.length >> 8,
    ephemeralPubKey.length >> 16,
    ephemeralPubKey.length >> 24,
  ]);
  const result = new Uint8Array(
    4 + ephemeralPubKey.length + 24 + ciphertext.length
  );
  result.set(prefix, 0);
  result.set(ephemeralPubKey, prefix.length);
  result.set(nonce, prefix.length + ephemeralPubKey.length);
  result.set(ciphertext, prefix.length + ephemeralPubKey.length + nonce.length);
  return result;
};

export const aes256GCM_secp256k1Decrypt = (
  privKey: Uint8Array,
  ciphertext: Uint8Array
): Uint8Array => {
  const prefix = ciphertext.slice(0, 4);
  const ephemeralPubKeyLength =
    prefix[0] | (prefix[1] << 8) | (prefix[2] << 16) | (prefix[3] << 24);
  const ephemeralPubKey = ciphertext.slice(4, 4 + ephemeralPubKeyLength);
  const nonce = ciphertext.slice(
    4 + ephemeralPubKeyLength,
    4 + ephemeralPubKeyLength + 24
  );
  const encryptedMessage = ciphertext.slice(4 + ephemeralPubKeyLength + 24);
  const sharedSecret = secp.getSharedSecret(privKey, ephemeralPubKey);
  const cipherKey = sha256.create().update(sharedSecret).digest();
  const aead = aesAEAD(cipherKey, nonce, ephemeralPubKey);
  const plaintext = aead.decrypt(encryptedMessage);
  console.log('decrypt result', Buffer.from(plaintext).toString('utf-8'));
  return plaintext;
};
