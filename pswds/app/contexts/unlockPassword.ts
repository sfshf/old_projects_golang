import React from 'react';

// unlock password context
interface UnlockPasswordContextProp {
  password: string; // 解锁密码
  setPassword: React.Dispatch<React.SetStateAction<string>>;
  setVisible: React.Dispatch<React.SetStateAction<boolean>>;
  supportFingerprint: boolean;
  setSupportFingerprint: React.Dispatch<React.SetStateAction<boolean>>;
}

export const UnlockPasswordContext =
  React.createContext<UnlockPasswordContextProp>({
    password: '',
    setPassword: () => {},
    setVisible: () => {},
    supportFingerprint: false,
    setSupportFingerprint: () => {},
  });
