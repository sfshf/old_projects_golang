'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { MutableRefObject, useEffect } from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '@/app/shared';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Stack from '@mui/material/Stack';
import Pagination from '@mui/material/Pagination';
import Link from '@mui/material/Link';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

const baseURL = process.env.NEXT_PUBLIC_SERVER_BASE_URL;

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'rpcServer', label: 'Rpc Server', minWidth: 100 },
  { id: 'status', label: 'Status', minWidth: 100 },
  { id: 'blockNumber', label: 'BlockNumber', minWidth: 100 },
  { id: 'delay', label: 'Delay', minWidth: 100 },
];

interface State {
  status: number;
  blockNumber: number;
  delay: number;
}

const StateRow = ({
  refresh,
  rpcServer,
  setErrorToast,
}: {
  refresh: boolean;
  rpcServer: string;
  setErrorToast: (err: string) => void;
}) => {
  const [loading, setLoading] = React.useState(false);
  const [state, setState] = React.useState<any>(null);
  const rpcServerDetection = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    const controller = new AbortController();
    setTimeout(() => {
      controller.abort('request timeout:' + rpcServer);
    }, 30000);
    setLoading(true);
    let postData: any = JSON.stringify({
      apiKey: password,
      serverUrl: rpcServer,
    });
    let reqOpts: any = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
      },
      body: postData,
      signal: controller.signal,
    };
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/doom/console/rpcServerDetection/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setState({
        status: resData.data.state.status,
        blockNumber: resData.data.state.blockNumber,
        delay: resData.data.state.delay,
      });
    } catch (error: any) {
      setLoading(false);
      setErrorToast(error);
    }
  };

  useEffect(() => {
    if (refresh) {
      rpcServerDetection();
      setState(null);
    }
  }, [refresh]);

  return (
    <TableRow key={rpcServer}>
      {columns.map((column) => {
        if (column.id === 'rpcServer') {
          return (
            <TableCell
              key={column.id}
              align={column.align}
              style={{ minWidth: column.minWidth }}
            >
              {rpcServer}
            </TableCell>
          );
        }
        if (loading) {
          return (
            <TableCell
              key={column.id}
              align={column.align}
              style={{ minWidth: column.minWidth }}
            >
              <CircularProgress size={10} color="inherit" />
            </TableCell>
          );
        } else {
          return (
            <TableCell
              key={column.id}
              align={column.align}
              style={{ minWidth: column.minWidth }}
            >
              {state ? state[column.id] : ''}
            </TableCell>
          );
        }
      })}
    </TableRow>
  );
};

export default function Page() {
  const [fresh, setFresh] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const stateList = [
    'https://rpc.ankr.com/eth',
    'https://eth-mainnet.public.blastapi.io',
    'https://eth.llamarpc.com',
    'https://ethereum.publicnode.com',
    'wss://mainnet.infura.io/ws/v3/ae1cb3ad3a4542e294f99a5f92be46c9',
    'https://rpc.ankr.com/eth/934224580d3a7fb3cc4f1d986571cd58b14eb1d7203a696842294d9acf721479',
    'https://eth-mainnet.g.alchemy.com/v2/YZdA4xns2oHf_FxL9gaXzzZktkj6wCAf',
  ];
  const [rpcServer, setRpcServer] = React.useState('');
  const [status, setStatus] = React.useState(0);
  const [blockNumber, setBlockNumber] = React.useState(0);
  const [delay, setDelay] = React.useState(0);

  const rpcServerDetection = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    const controller = new AbortController();
    setTimeout(() => {
      controller.abort('request timeout:' + rpcServer);
    }, 30000);
    setLoading(true);
    let postData: any = JSON.stringify({
      apiKey: password,
      serverUrl: rpcServer,
    });
    let reqOpts: any = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
      },
      body: postData,
      signal: controller.signal,
    };
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/doom/console/rpcServerDetection/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setStatus(resData.data.state.status);
      setBlockNumber(resData.data.state.blockNumber);
      setDelay(resData.data.state.delay);
    } catch (error: any) {
      setLoading(false);
      setErrorToast(error);
    }
  };

  useEffect(() => {
    setFresh(false);
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
      <Grid container spacing={4}>
        <Grid item xs={1}>
          <Button
            variant="contained"
            size="large"
            sx={{
              height: '50px',
            }}
            onClick={() => {
              setFresh(true);
              setTimeout(() => {
                setFresh(false);
              }, 5000);
            }}
          >
            Search
          </Button>
        </Grid>
        <Grid item xs={3}>
          <TextField
            id="outlined-basic"
            label="Outlined"
            variant="outlined"
            fullWidth
            value={rpcServer}
            onChange={(e) => {
              setRpcServer(e.target.value);
            }}
          />
        </Grid>
        <Grid item xs={1}>
          <Button
            variant="contained"
            size="large"
            sx={{
              height: '50px',
            }}
            onClick={() => {
              rpcServerDetection();
            }}
          >
            Detect
          </Button>
        </Grid>
        <Grid item xs={2}>
          <Typography variant="h4" gutterBottom>
            Status: {status}
          </Typography>
        </Grid>
        <Grid item xs={3}>
          <Typography variant="h4" gutterBottom>
            BlockNumber: {blockNumber}
          </Typography>
        </Grid>
        <Grid item xs={2}>
          <Typography variant="h4" gutterBottom>
            Delay: {delay}
          </Typography>
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
        <TableContainer>
          <Table stickyHeader aria-label="sticky table">
            <TableHead>
              <TableRow>
                {columns.map((column) => (
                  <TableCell
                    key={column.id}
                    align={column.align}
                    style={{ minWidth: column.minWidth }}
                  >
                    {column.label}
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {stateList.map((row: any) => {
                return (
                  <StateRow
                    key={row.rpcServer}
                    refresh={fresh}
                    rpcServer={row}
                    setErrorToast={setErrorToast}
                  />
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      </Stack>
      {/* <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop> */}
      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
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
    </Stack>
  );
}
