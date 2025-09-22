'use client';
import * as React from 'react';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';
import { post } from '@/app/util';
import shared from '@/app/shared';
import Paper from '@mui/material/Paper';
import moment from 'moment';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

interface Log {
  id: string;
  mode: number;
  message: string;
  createdAt: number;
}

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [list, setList] = React.useState<null | Log[]>(null);
  const modetoString = (mode: number) => {
    let result = '';
    switch (mode) {
      case 1:
        return 'Telegram';
      case 2:
        return 'Wxpusher';
      case 4:
        return 'Email';
      case 3:
        return 'Telegram|Wxpusher';
      case 5:
        return 'Telegram|Email';
      case 6:
        return 'Wxpusher|Email';
      case 7:
        return 'Telegram|Wxpusher|Email';
    }
  };
  const getMessageNotificationLogs = () => {
    post(
      false,
      '',
      '/tester/getMessageNotificationLogs/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword() },
      (respData: any) => {
        if (respData.data) {
          setList(respData.data.list);
        }
      },
      setToast,
      setErrorToast
    );
  };

  React.useEffect(() => {
    getMessageNotificationLogs();
  }, []);

  return (
    <>
      <Stack
        spacing={1}
        sx={{
          padding: '1rem',
        }}
        direction="column"
        alignItems="center"
      >
        <TableContainer sx={{}} component={Paper}>
          <Table aria-label="simple table">
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell align="right">Mode</TableCell>
                <TableCell align="right">Message</TableCell>
                <TableCell align="right">CreatedAt</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {list &&
                list.map((row) => (
                  <TableRow
                    key={'useTelegram'}
                    sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                  >
                    <TableCell component="th" scope="row">
                      {row.id}
                    </TableCell>
                    <TableCell align="right">
                      {modetoString(row.mode)}
                    </TableCell>
                    <TableCell align="right">{row.message}</TableCell>
                    <TableCell align="right">
                      {moment(row.createdAt * 1000)
                        .local()
                        .format('YYYY-MM-DD HH:mm:ss')}
                    </TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </TableContainer>
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
        autoHideDuration={3000}
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
    </>
  );
}
