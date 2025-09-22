import { xchacha20poly1305 } from '@noble/ciphers/chacha';
import { keccak_256 } from '@noble/hashes/sha3';
import {
  PrivateKey,
  encrypt as eciesEncrypt,
  decrypt as eciesDecrypt,
} from 'eciesjs';
import moment from 'moment';
import { Buffer } from 'buffer';

export const decryptByUserPrivateKey = (
  ciphertext: Uint8Array,
  unlockPassword: string,
): Buffer => {
  return eciesDecrypt(userPrivateKey(unlockPassword).secret, ciphertext);
};

export const encryptedByUserPublicKey = (
  plaintext: Uint8Array,
  userPublicKey: Uint8Array,
): Buffer => {
  return eciesEncrypt(userPublicKey, plaintext);
};

export const familyKey = (unlockPassword: string): Uint8Array => {
  return keccak_256(
    unlockPassword +
      moment().valueOf() +
      'C6093FD9CC143F9F058938868B2DF2DAF9A91D28',
  );
};

export const userPrivateKey = (unlockPassword: string): PrivateKey => {
  return PrivateKey.fromHex(
    Buffer.from(
      keccak_256(unlockPassword + '4838B106FCE9647BDF1E7877BF73CE8B0BAD5F97'),
    ).toString('hex'),
  );
};

export const encryptByUnlockPassword = (
  unlockPassword: string,
  plaintext: string,
): string => {
  // 加解密的密钥 keccak_256（ password +  salt），salt为 9C9B913EB1B6254F4737CE947EFD16F16E916F
  let encryptKey = keccak_256(
    unlockPassword + '9C9B913EB1B6254F4737CE947EFD16F16E916F',
  );
  return encryptByXchacha20poly1305(encryptKey, plaintext);
};

export const encryptByXchacha20poly1305 = (
  key: Uint8Array,
  plaintext: string,
): string => {
  let nonce = new Uint8Array(24);
  let aead = xchacha20poly1305(key.slice(0, 32), nonce);
  let encryptedb64 = Buffer.from(
    aead.encrypt(new Uint8Array(Buffer.from(plaintext, 'utf-8'))),
  ).toString('base64');
  return encryptedb64;
};

export const decryptByUnlockPassword = (
  unlockPassword: string,
  encryptedb64: string,
): string => {
  let encryptKey = keccak_256(
    unlockPassword + '9C9B913EB1B6254F4737CE947EFD16F16E916F',
  );
  return decryptByXchacha20poly1305(encryptKey, encryptedb64);
};

export const decryptByXchacha20poly1305 = (
  key: Uint8Array,
  encryptedb64: string,
): string => {
  let nonce = new Uint8Array(24);
  let aead = xchacha20poly1305(key.slice(0, 32), nonce);
  let plainbytes = aead.decrypt(
    new Uint8Array(Buffer.from(encryptedb64, 'base64')),
  );
  return Buffer.from(plainbytes).toString('utf-8');
};
