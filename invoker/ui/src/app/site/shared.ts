import { post } from '@/app/shared';

export interface UserInfo {
  userID: number;
  nickname: string;
  email: string;
  phone: string;
}

export interface LoginState {
  userID: number;
  nickname: string;
  email: string;
  phone: string;
  expiry: number;
}

export const LocalStorageKey_UserLoginState = 'UserLoginState';

export let loginState: null | LoginState = null;

export const removeLocalUserLoginState = () => {
  if (typeof localStorage === 'undefined') {
    throw 'localStorage is undefined';
  }
  localStorage.removeItem(LocalStorageKey_UserLoginState);
  loginState = null;
};

export const setLocalUserLoginState = (origin: UserInfo) => {
  if (typeof localStorage === 'undefined') {
    throw 'localStorage is undefined';
  }
  const now = new Date();
  const ttl: number = 1000 * 60 * 60 * 24 * 30; // 30 days
  const item: LoginState = {
    userID: origin.userID,
    nickname: origin.nickname,
    email: origin.email,
    phone: origin.phone,
    expiry: now.getTime() + ttl,
  };
  loginState = item;
  localStorage.setItem(LocalStorageKey_UserLoginState, JSON.stringify(item));
};

export const getLocalUserLoginState = () => {
  if (loginState) {
    return loginState;
  }
  if (typeof localStorage === 'undefined') {
    return null;
  }
  const userStateStr = localStorage.getItem(LocalStorageKey_UserLoginState);
  if (!userStateStr) {
    return null;
  }
  const item = JSON.parse(userStateStr);
  const now = new Date();
  if (now.getTime() > item.expiry) {
    localStorage.removeItem(LocalStorageKey_UserLoginState);
    return null;
  } else {
    return item;
  }
};

export const loginBySession = ({
  setLoading,
  setErrorToast,
}: {
  setLoading: (loading: boolean) => void;
  setErrorToast: (errorToast: string) => void;
}) => {
  post(
    false,
    '/slark/user/loginBySession/v1',
    setLoading,
    null,
    (respHeaders: any) => {},
    (respData: any) => {
      if (respData.code != 0) {
        // loginBySession failed
        return;
      }
      if (respData.data) {
        setLocalUserLoginState(respData.data);
        let loginState = getLocalUserLoginState();
        if (!loginState) {
          setErrorToast(
            'get local user login state fail, after loginBySession'
          );
        }
      }
    },
    undefined,
    setErrorToast
  );
};
