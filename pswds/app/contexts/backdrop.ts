import React from 'react';

interface BackdropContextProp {
  setLoading: React.Dispatch<React.SetStateAction<boolean>>;
  setPrompt: React.Dispatch<React.SetStateAction<null | string>>;
}

export const BackdropContext = React.createContext<BackdropContextProp>({
  setLoading: () => {},
  setPrompt: () => {},
});
