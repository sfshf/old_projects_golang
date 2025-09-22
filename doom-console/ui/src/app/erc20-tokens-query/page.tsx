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
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';

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

  const [contractAddress, setContractAddress] = React.useState('');
  const [address, setAddress] = React.useState('');
  const [type, setType] = React.useState('');
  const [name, setName] = React.useState('');
  const [symbol, setSymbol] = React.useState('');
  const [decimals, setDecimals] = React.useState(0);
  const [priced, setPriced] = React.useState(false);
  const [checked, setChecked] = React.useState(false);

  const erc20TokensQuery = async () => {
    if (contractAddress === '') {
      setErrorToast('Contract Address is empty');
      return;
    }
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    let postData: any = JSON.stringify({
      apiKey: password,
      contractAddress,
    });
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
        `${shared.baseAPIURL}/doom/console/erc20TokensQuery/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setAddress(resData.data.address);
      setType(resData.data.type);
      setName(resData.data.name);
      setSymbol(resData.data.symbol);
      setDecimals(resData.data.decimals);
      setPriced(resData.data.priced);
      setChecked(resData.data.checked);
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
              erc20TokensQuery();
            }}
          >
            Search
          </Button>
        </Grid>
        <Grid item xs={4}>
          <FormControl fullWidth>
            <TextField
              label="Contract Address"
              value={contractAddress}
              onChange={(e) => {
                setContractAddress(e.target.value);
              }}
              id="outlined-basic"
              variant="outlined"
            />
          </FormControl>
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
          Address：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {address}
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
          Type：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {type}
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
          Name：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {name}
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
          Symbol：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {symbol}
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
          Decimals：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {decimals}
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
          Priced：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {priced ? 'True' : 'False'}
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
          Checked：
        </Typography>
        <Typography variant="overline" display="block" gutterBottom>
          {checked ? 'True' : 'False'}
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
