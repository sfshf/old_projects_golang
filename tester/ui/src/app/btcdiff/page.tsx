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

interface BtcDiff {
  priceCB: number;
  priceBN: number;
  priceDiff: number;
  diffPercent: number;
  errorMessage: string;
  updatedAt: number;
}

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [btcDiff, setBtcDiff] = React.useState<null | BtcDiff>(null);
  const getBtcDiff = () => {
    post(
      false,
      '',
      '/tester/getBtcDiff/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword() },
      (respData: any) => {
        if (respData.data) {
          setBtcDiff(respData.data);
        }
      },
      setToast,
      setErrorToast
    );
  };

  useEffect(() => {
    getBtcDiff();
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
          Coinbase价:
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="subtitle1" gutterBottom>
          {btcDiff ? '$' + btcDiff.priceCB : 'unknown'}
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="h6" gutterBottom>
          Binance价:
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="subtitle1" gutterBottom>
          {btcDiff ? '$' + btcDiff.priceBN : 'unknown'}
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="h6" gutterBottom>
          溢价 :
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="subtitle1" gutterBottom>
          {btcDiff ? '$' + btcDiff.priceDiff : 'unknown'}
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="h6" gutterBottom>
          溢价百分比:
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="subtitle1" gutterBottom>
          {btcDiff ? btcDiff.diffPercent + '%' : 'unknown'}
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="h6" gutterBottom>
          更新时间:
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="subtitle1" gutterBottom>
          {btcDiff
            ? moment(btcDiff.updatedAt).local().format('YYYY-MM-DD HH:mm:ss')
            : 'unknown'}
        </Typography>
      </Stack>
      {btcDiff && btcDiff.errorMessage && (
        <>
          <Stack spacing={2} alignItems="center" direction="row">
            <Typography variant="h6" gutterBottom>
              报错信息:
            </Typography>
          </Stack>
          <Stack spacing={2} alignItems="center" direction="row">
            <Typography variant="subtitle1" gutterBottom>
              {btcDiff.errorMessage}
            </Typography>
          </Stack>
        </>
      )}

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
