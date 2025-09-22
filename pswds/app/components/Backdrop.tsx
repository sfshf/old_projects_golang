/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { Dialog, Text } from '@rneui/themed';
import { BackdropContext } from '../contexts/backdrop';

interface BackdropProp {
  children?: React.ReactNode;
}

function Backdrop({ children }: BackdropProp): React.JSX.Element {
  const [loading, setLoading] = React.useState<boolean>(false);
  const [prompt, setPrompt] = React.useState<null | string>(null);
  return (
    <>
      <BackdropContext.Provider value={{ setPrompt, setLoading }}>
        {children}
      </BackdropContext.Provider>
      <Dialog isVisible={loading}>
        {prompt && <Text h4>{prompt}</Text>}
        <Dialog.Loading loadingProps={{ size: 'small' }} />
      </Dialog>
    </>
  );
}

export default Backdrop;
