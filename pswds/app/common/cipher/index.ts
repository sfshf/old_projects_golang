import { Cipher } from '@noble/ciphers/utils';
import * as secp from '@noble/secp256k1';
import { sha256 } from '@noble/hashes/sha256';
import { gcm } from '@noble/ciphers/aes';

export const randomPassword = ({
  length,
  useNumbers,
  useSymbols,
}: {
  length: number;
  useNumbers: boolean;
  useSymbols: boolean;
}): string => {
  let charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
  let newPassword = '';
  let mustNumber = '';
  if (useNumbers) {
    const numbers = '0123456789';
    charset += numbers;
    mustNumber = numbers.charAt(Math.floor(Math.random() * numbers.length));
  }
  let mustSymbol = '';
  if (useSymbols) {
    const symbols = '_!@#$%^&*()';
    charset += symbols;
    mustSymbol = symbols.charAt(Math.floor(Math.random() * symbols.length));
  }
  let mustNumberIndex = -1;
  if (mustNumber) {
    mustNumberIndex = Math.floor(Math.random() * length);
  }
  let mustSymbolIndex = -1;
  if (mustSymbol) {
    mustSymbolIndex = Math.floor(Math.random() * length);
  }
  for (let i = 0; i < length; i++) {
    if (i === mustNumberIndex) {
      newPassword += mustNumber;
    } else if (i === mustSymbolIndex) {
      newPassword += mustSymbol;
    } else {
      newPassword += charset.charAt(Math.floor(Math.random() * charset.length));
    }
  }
  return newPassword;
};

export const randomDigits = (length: number): string => {
  let charset = '0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += charset.charAt(Math.floor(Math.random() * charset.length));
  }
  return result;
};

// AES-256-GCM
export const aesAEAD = (
  key: Uint8Array,
  nonce: Uint8Array,
  AAD?: Uint8Array,
): Cipher => {
  return gcm(key, nonce, AAD);
};

export const aes256GCM_secp256k1Encrypt = (
  pubKey: Uint8Array,
  plaintext: Uint8Array,
): Uint8Array => {
  const ephemeralPrivKey = secp.utils.randomPrivateKey();
  const ephemeralPubKey = secp.getPublicKey(ephemeralPrivKey, false);
  const sharedSecret = secp.getSharedSecret(ephemeralPrivKey, pubKey, false);
  const cipherKey = sha256.create().update(sharedSecret).digest();
  const nonce = new Uint8Array(16);
  crypto.getRandomValues(nonce);
  const aead = aesAEAD(cipherKey, nonce);
  const ciphertext = aead.encrypt(plaintext);
  const tag = ciphertext.slice(ciphertext.length - nonce.length);
  const cipherText = ciphertext.slice(0, ciphertext.length - tag.length);
  const result = new Uint8Array(
    ephemeralPubKey.length + 16 + ciphertext.length,
  );
  result.set(ephemeralPubKey, 0);
  result.set(nonce, ephemeralPubKey.length);
  result.set(tag, ephemeralPubKey.length + nonce.length);
  result.set(cipherText, ephemeralPubKey.length + nonce.length + tag.length);
  return result;
};

export const aes256GCM_secp256k1Decrypt = (
  privKey: Uint8Array,
  ciphertext: Uint8Array,
): Uint8Array => {
  // ciphertext cannot be less than length of public key (65) + nonce (16) + tag (16)
  if (ciphertext.length <= 1 + 32 + 32 + 16 + 16) {
    throw 'invalid length of ciphertext';
  }
  const ephemeralPubKey = ciphertext.slice(0, 65);
  const nonce = ciphertext.slice(
    ephemeralPubKey.length,
    ephemeralPubKey.length + 16,
  );
  const tag = ciphertext.slice(
    ephemeralPubKey.length + 16,
    ephemeralPubKey.length + 16 + 16,
  );
  const remain = ciphertext.slice(
    ephemeralPubKey.length + nonce.length + tag.length,
  );
  const cipherText = new Uint8Array(remain.length + tag.length);
  cipherText.set(remain, 0);
  cipherText.set(tag, remain.length);
  const sharedSecret = secp.getSharedSecret(privKey, ephemeralPubKey, false);
  const cipherKey = sha256.create().update(sharedSecret).digest();
  const aead = aesAEAD(cipherKey, nonce);
  const plaintext = aead.decrypt(cipherText);
  return plaintext;
};
