'use client';
import React from 'react';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

export default function CustomSnackbar({
  errorToast,
  setErrorToast,
  toast,
  setToast,
}: {
  errorToast: string;
  setErrorToast: (msg: string) => void;
  toast?: string;
  setToast?: (msg: string) => void;
}) {
  return (
    <>
      <Snackbar
        open={errorToast !== ''}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
        onClose={() => setErrorToast('')}
      >
        <Alert
          onClose={() => setErrorToast('')}
          severity="error"
          sx={{ width: '100%' }}
        >
          {errorToast}
        </Alert>
      </Snackbar>
      {toast && setToast && (
        <Snackbar
          open={toast !== ''}
          autoHideDuration={1000}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
          onClose={() => setToast('')}
        >
          <Alert
            onClose={() => setToast('')}
            severity="success"
            sx={{ width: '100%' }}
          >
            {toast}
          </Alert>
        </Snackbar>
      )}
    </>
  );
}
