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
import moment from 'moment';

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

  const [v2Info, setV2Info] = React.useState<any>(null);
  const [v3Info, setV3Info] = React.useState<any>(null);

  const uniswapInfo = async () => {
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
        `${shared.baseAPIURL}/doom/console/uniswapInfo/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setV2Info(resData.data.v2Info);
      setV3Info(resData.data.v3Info);
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
              uniswapInfo();
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
        <Typography variant="h4" gutterBottom>
          Uniswap V2 Info：
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
          当前实时的lp总数：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {v2Info ? v2Info.totalRealTime : 'NULL'}
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
          现在数据库存的lp总数：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {v2Info ? v2Info.totalInDB : 'NULL'}
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
          {v2Info ? v2Info.diffValue : 'NULL'}
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
        <Typography variant="h4" gutterBottom>
          Uniswap V3 Info：
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
          现在数据库存的pool总数：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {v3Info ? v3Info.totalInDB : 'NULL'}
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
          上一次脚本执行时间：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {v3Info
            ? moment(v3Info.timestamp).local().format('YYYY-MM-DD HH:mm:ss')
            : 'NULL'}
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
