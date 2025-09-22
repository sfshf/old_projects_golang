'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { useEffect } from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import { post } from '@/app/util';
import shared from '@/app/shared';
import moment from 'moment';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'database', label: 'Database', minWidth: 100 },
  { id: 'size', label: 'Size (MB)', minWidth: 100 },
];

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [lastTS, setLastTS] = React.useState<number>(0); // it's a error msg if lastTS==0
  const [msg, setMsg] = React.useState<string>('');
  const getTxsMempoolInfo = () => {
    post(
      false,
      '',
      '/tester/getTxsMempoolInfo/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword() },
      (respData: any) => {
        if (respData.data) {
          setLastTS(respData.data.lastTS);
          setMsg(respData.data.msg);
        }
      },
      setToast,
      setErrorToast
    );
  };

  useEffect(() => {
    getTxsMempoolInfo();
  }, []);

  return (
    <Stack
      spacing={3}
      direction="column"
      alignItems="center"
      sx={{
        padding: '1rem',
      }}
    >
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="h6" gutterBottom>
          Status:
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography
          variant="subtitle1"
          gutterBottom
          sx={lastTS > 0 ? { color: 'green' } : { color: 'red' }}
        >
          {lastTS > 0 ? 'OK' : 'ERROR'}
        </Typography>
      </Stack>
      {lastTS > 0 && (
        <>
          <Stack spacing={2} alignItems="center" direction="row">
            <Typography variant="h6" gutterBottom>
              Last Time:
            </Typography>
          </Stack>
          <Stack spacing={2} alignItems="center" direction="row">
            <Typography variant="subtitle1" gutterBottom>
              {moment(lastTS).local().format('YYYY-MM-DD HH:mm:ss')}
            </Typography>
          </Stack>
        </>
      )}
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="h6" gutterBottom>
          Message:
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="subtitle1" gutterBottom>
          {msg}
        </Typography>
      </Stack>

      <Snackbar
        open={errorToast !== ''}
        autoHideDuration={5000}
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
      <Snackbar
        open={toast !== ''}
        autoHideDuration={4500}
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
      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
    </Stack>
  );
}
