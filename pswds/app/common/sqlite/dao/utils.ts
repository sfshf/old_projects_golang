// 异或
export const xor_hex = (secretHex: string, hex: string): string => {
  return Buffer.from(
    xor(Buffer.from(secretHex, 'hex'), Buffer.from(hex, 'hex')),
  ).toString();
};
// NOTE：不使用 hex，sqlite数据库会出现字符转义异常，导致数据出错！！！
export const xor_str = (secretHex: string, origin: string): string => {
  if (origin === '') {
    return '';
  }
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
