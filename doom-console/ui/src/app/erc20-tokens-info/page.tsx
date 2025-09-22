'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { useEffect } from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '@/app/shared';
import Stack from '@mui/material/Stack';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [headerNumber, setHeaderNumber] = React.useState(0);
  const [toBlockNumber, setToBlockNumber] = React.useState(0);
  const [numberDiff, setNumberDiff] = React.useState(0);
  const [days, setDays] = React.useState(0);
  const [totalTokens, setTotalTokens] = React.useState(0);

  const erc20TokensInfo = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    let postData: any = JSON.stringify({ apiKey: password });
    let reqOpts: any = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
      },
      body: postData,
    };
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/doom/console/erc20TokensInfo/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setHeaderNumber(resData.data.headerNumber);
      setToBlockNumber(resData.data.toBlockNumber);
      setNumberDiff(resData.data.numberDiff);
      setDays(resData.data.days);
      setTotalTokens(resData.data.totalTokens);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  return (
    <Stack
      spacing={3}
      direction="column"
      alignItems="center"
      sx={{
        padding: '1rem',
      }}
    >
      <Grid container spacing={4}>
        <Grid item xs={1}>
          <Button
            variant="contained"
            size="large"
            sx={{
              height: '50px',
            }}
            onClick={() => {
              erc20TokensInfo();
            }}
          >
            Search
          </Button>
        </Grid>
      </Grid>
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
          width: '100%',
        }}
        direction="row"
      >
        <Typography variant="h6" gutterBottom>
          当前eth链上最新的blocknumber：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {headerNumber}
        </Typography>
      </Stack>
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
          width: '100%',
        }}
        direction="row"
      >
        <Typography variant="h6" gutterBottom>
          数据库里记录的blocknumber：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {toBlockNumber}
        </Typography>
      </Stack>
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
          width: '100%',
        }}
        direction="row"
      >
        <Typography variant="h6" gutterBottom>
          差额：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {numberDiff}
        </Typography>
      </Stack>
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
          width: '100%',
        }}
        direction="row"
      >
        <Typography variant="h6" gutterBottom>
          估算天数：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {days}
        </Typography>
      </Stack>
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
          width: '100%',
        }}
        direction="row"
      >
        <Typography variant="h6" gutterBottom>
          目前数据库里的erc20tokens的总数：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {totalTokens}
        </Typography>
      </Stack>

      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
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
    </Stack>
  );
}
