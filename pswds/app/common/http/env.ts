/* eslint-disable @typescript-eslint/no-unused-vars */
export type EVN = {
  API_URL: string;
  API_ENCRYPTED: boolean;
};

const prod: EVN = {
  API_URL: 'https://api.n1xt.net',
  API_ENCRYPTED: true,
};

const dev: EVN = {
  API_URL: 'https://api.test.n1xt.net',
  API_ENCRYPTED: true,
};

// change env here
export const env = dev;
