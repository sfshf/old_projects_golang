'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { useEffect } from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '../../shared';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import moment from 'moment';
import keccak256 from 'keccak256';
import { XChaCha20Poly1305 } from '@stablelib/xchacha20poly1305';

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
  { id: 'date', label: 'Date', minWidth: 100 },
  { id: 'count', label: 'Count', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [app, setApp] = React.useState('');
  const [duration, setDuration] = React.useState(1);
  const [dateList, setDateList] = React.useState<string[]>([]);
  const [countList, setCountList] = React.useState<number[]>([]);
  const [currentCount, setCurrentCount] = React.useState(0);
  const [appList, setAppList] = React.useState<string[]>([]);

  const getAllApps = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    const reqData = {
      password: password,
    };
    post('/alchemist/console/getAllApps/v1', reqData, (respData) => {
      setAppList(respData.data.list);
    });
  };

  const listSubsciptionCounts = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    if (app == '') {
      setErrorToast('app is empty');
      return;
    }

    if (dateList.length == 0) {
      setErrorToast('invalid date duration');
      return;
    }

    const reqData = {
      password: password,
      app: app,
      startDate: dateList[dateList.length - 1],
      endDate: dateList[0],
    };
    post(
      '/alchemist/console/listSubscriptionCounts/v1',
      reqData,
      (respData) => {
        setCountList(respData.data.list);
      }
    );
  };

  const getCurrentSubscriptionCount = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    if (app == '') {
      setErrorToast('app is empty');
      return;
    }

    const reqData = {
      password: password,
      app: app,
    };
    post(
      '/alchemist/console/getCurrentSubscriptionCount/v1',
      reqData,
      (respData) => {
        setCurrentCount(respData.data.count);
      }
    );
  };

  const post = async (
    path: string,
    reqData?: any,
    successAction?: (respData: any) => void
  ) => {
    setLoading(true);
    let postData = '';
    if (reqData) {
      postData = JSON.stringify(reqData);
    }
    let encryptKey;
    let nonce: Uint8Array;
    let aead: XChaCha20Poly1305;
    // use /go
    let ts = moment();
    encryptKey = keccak256(ts.format('x') + '9C9B913EB1B6254F4737CE947');
    nonce = new Uint8Array(24);
    aead = new XChaCha20Poly1305(new Uint8Array(encryptKey).slice(0, 32));
    let encryptedData = Buffer.from(
      aead.seal(nonce, new Uint8Array(Buffer.from(postData, 'utf-8')))
    ).toString('base64');
    postData = JSON.stringify({
      path: path,
      data: encryptedData,
      timestamp: ts.valueOf(),
    });
    path = '/go';
    try {
      let reqOpts: any = {};
      reqOpts.method = 'POST';
      let headers: any = {};
      if (postData) {
        headers = {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
        };
        reqOpts['body'] = postData;
      }
      reqOpts['headers'] = headers;
      fetch(shared.kongAddress + path, reqOpts)
        .then((resp) => {
          return resp.json();
        })
        .then((data) => {
          if (data.code !== 0) {
            setErrorToast(data.debugMessage);
            return;
          }
          // use /go
          let decryptedData = aead.open(
            nonce,
            Buffer.from(data.data.encryptedData, 'base64')
          );
          if (decryptedData) {
            data.data = JSON.parse(Buffer.from(decryptedData).toString());
          }
          if (successAction) {
            successAction(data);
          }
          setToast('success');
        })
        .then((err) => {
          throw err;
        });

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const latestWeekDates = (): string[] => {
    var currentDate = new Date();
    currentDate.setHours(0);
    currentDate.setMinutes(0);
    currentDate.setSeconds(0);
    currentDate.setMilliseconds(0);
    var dates: string[] = [];

    for (var i = 1; i < 8; i++) {
      var newDate = new Date(currentDate.getTime() - i * 24 * 60 * 60 * 1000);
      var formattedDate = moment(newDate).format('YYYY-MM-DD');
      dates.push(formattedDate);
    }
    return dates;
  };

  const latestMonthDates = (): string[] => {
    var currentDate = new Date();
    currentDate.setHours(0);
    currentDate.setMinutes(0);
    currentDate.setSeconds(0);
    currentDate.setMilliseconds(0);
    var dates: string[] = [];

    for (var i = 1; i < 31; i++) {
      var newDate = new Date(currentDate.getTime() - i * 24 * 60 * 60 * 1000);
      var formattedDate = moment(newDate).format('YYYY-MM-DD');
      dates.push(formattedDate);
    }
    return dates;
  };

  const lastMonthDates = (): string[] => {
    var date = new Date();
    date.setMonth(date.getMonth(), 0);
    var dates: string[] = [];
    var daysInMonth = date.getDate();

    for (var i = 1; i < daysInMonth + 1; i++) {
      var day = new Date(date.getFullYear(), date.getMonth(), i);
      var formattedDay = moment(day).format('YYYY-MM-DD');
      dates.push(formattedDay);
    }
    return dates.reverse();
  };

  useEffect(() => {
    getAllApps();
    switch (duration) {
      case 1:
        setDateList(latestWeekDates());
        break;
      case 2:
        setDateList(latestMonthDates());
        break;
      case 3:
        setDateList(lastMonthDates());
        break;
    }
  }, []);

  return (
    <main>
      <div className="ml-10">
        <Stack direction="row" alignItems="center">
          <Button
            onClick={() => {
              listSubsciptionCounts();
            }}
            variant="contained"
            sx={{ margin: '20px' }}
          >
            Refresh
          </Button>
          <FormControl sx={{ margin: '20px', width: '200px' }}>
            <InputLabel id="demo-simple-select-label">App</InputLabel>
            <Select
              value={app}
              label="App"
              onChange={(e) => {
                setApp(e.target.value as string);
              }}
            >
              <MenuItem value={''}>NULL</MenuItem>
              {appList &&
                appList.map((e) => (
                  <MenuItem key={e} value={e}>
                    {e}
                  </MenuItem>
                ))}
            </Select>
          </FormControl>
          <FormControl sx={{ margin: '20px', width: '200px' }}>
            <InputLabel id="demo-simple-select-label">Duration</InputLabel>
            <Select
              value={duration}
              label="Duration"
              onChange={(e) => {
                let val = e.target.value as number;
                switch (val) {
                  case 1:
                    setDateList(latestWeekDates());
                    break;
                  case 2:
                    setDateList(latestMonthDates());
                    break;
                  case 3:
                    setDateList(lastMonthDates());
                    break;
                }
                setDuration(val);
                setCountList([]);
              }}
            >
              <MenuItem value={1}>Lastest Week</MenuItem>
              <MenuItem value={2}>Lastest Month</MenuItem>
              <MenuItem value={3}>Last Month</MenuItem>
            </Select>
          </FormControl>
          <Typography
            variant="h6"
            gutterBottom
            sx={{ margin: '20px', width: '200px' }}
          >
            Current Count: {currentCount}
          </Typography>
        </Stack>
      </div>

      {dateList && countList && (
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
              {dateList.map((row, idx) => {
                return (
                  <TableRow key={row}>
                    <TableCell>{row}</TableCell>
                    <TableCell>{countList[idx]}</TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
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
    </main>
  );
}
