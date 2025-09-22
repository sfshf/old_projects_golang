import React from 'react';

interface SnackbarContextProp {
  setSuccess: (success: string, guide?: string, timeout?: number) => void;
  setError: (error: string, guide?: string, timeout?: number) => void;
  setWarning: (warning: string, guide?: string, timeout?: number) => void;
}

export const SnackbarContext = React.createContext<SnackbarContextProp>({
  setSuccess: () => {},
  setError: () => {},
  setWarning: () => {},
});
