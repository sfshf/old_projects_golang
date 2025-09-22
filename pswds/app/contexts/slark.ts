import React from 'react';
import { SlarkInfo } from '../services/slark';

type SlarkInfoContextProp = {
  slarkInfo: null | SlarkInfo;
  setSlarkInfo: React.Dispatch<React.SetStateAction<null | SlarkInfo>>;
};

export const SlarkInfoContext = React.createContext<SlarkInfoContextProp>({
  slarkInfo: null,
  setSlarkInfo: () => {},
});
