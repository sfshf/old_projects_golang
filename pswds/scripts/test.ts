import { keccak_256 } from '@noble/hashes/sha3';

// 异或
export const xor_hex = (secretHex: string, hex: string): string => {
  return Buffer.from(
    xor(Buffer.from(secretHex, 'hex'), Buffer.from(hex, 'hex')),
  ).toString();
};
export const xor_str = (secretHex: string, origin: string): string => {
  return Buffer.from(
    xor(Buffer.from(secretHex, 'hex'), Buffer.from(origin)),
  ).toString('hex');
};
export const xor = (secret: Uint8Array, origin: Uint8Array): Uint8Array => {
  let index = 0;
  const handled = new Uint8Array(origin.length);
  for (let i = 0; i < origin.length; i++) {
    let right = next(secret, index);
    handled[i] = origin[i] ^ right.val;
    index = right.next;
  }
  return handled;
};

const next = (
  secret: Uint8Array,
  index: number,
): { val: number; next: number } => {
  if (index === secret.length) {
    index = 0;
  }
  return { val: secret[index], next: index + 1 };
};

console.log(
  '===>result1:',
  Buffer.from(
    xor(
      Buffer.from('111111'),
      xor(Buffer.from('111111'), Buffer.from('asdfasd')),
    ),
  ).toString(),
);

console.log(
  '===>result2:',
  Buffer.from(
    xor_hex(
      Buffer.from(keccak_256('111111')).toString('hex'),
      xor_str(Buffer.from(keccak_256('111111')).toString('hex'), 'qwerqwer'),
    ),
  ).toString(),
);
