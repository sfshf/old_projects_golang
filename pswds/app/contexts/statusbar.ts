import React from 'react';

interface StatusBarStyleContextProp {
  statusBarStyle: 'light' | 'dark';
  setStatusBarStyle: React.Dispatch<React.SetStateAction<'light' | 'dark'>>;
}

export const StatusBarStyleContext =
  React.createContext<StatusBarStyleContextProp>({
    statusBarStyle: 'light',
    setStatusBarStyle: () => {},
  });
