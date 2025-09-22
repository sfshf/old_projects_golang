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
import * as clipboard from 'clipboard-polyfill';
import Typography from '@mui/material/Typography';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

interface Account {
  id: number;
  userID: number;
  userEmail: string;
  emailAccount: string;
  password: string;
}

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [list, setList] = React.useState<null | Account[]>(null);

  const getPrivacyEmailAccounts = () => {
    post(
      false,
      '',
      '/tester/getPrivacyEmailAccounts/v1',
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
    getPrivacyEmailAccounts();
  }, []);
  const copyToClipboard = (text: string) => {
    clipboard.writeText(text);
    setToast('copied');
  };

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
                <TableCell>Slark Account</TableCell>
                <TableCell>Account</TableCell>
                <TableCell>Password</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {list &&
                list.map((row) => (
                  <TableRow
                    key={'useTelegram'}
                    sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                  >
                    <TableCell>
                      <Stack direction="row">
                        <Typography sx={{ width: '50%' }}>
                          {row.userEmail}
                        </Typography>
                        <Button
                          sx={{ marginLeft: 20 }}
                          size="small"
                          variant="contained"
                          onClick={() => {
                            copyToClipboard(row.userEmail);
                          }}
                        >
                          Copy
                        </Button>
                      </Stack>
                    </TableCell>
                    <TableCell>
                      <Stack direction="row">
                        <Typography sx={{ width: '50%' }}>
                          {row.emailAccount}
                        </Typography>
                        <Button
                          sx={{ marginLeft: 20 }}
                          size="small"
                          variant="contained"
                          onClick={() => {
                            copyToClipboard(row.emailAccount);
                          }}
                        >
                          Copy
                        </Button>
                      </Stack>
                    </TableCell>
                    <TableCell>
                      <Stack direction="row">
                        <Typography sx={{ width: '50%' }}>
                          {row.password}
                        </Typography>
                        <Button
                          sx={{ marginLeft: 20 }}
                          size="small"
                          variant="contained"
                          onClick={() => {
                            copyToClipboard(row.password);
                          }}
                        >
                          Copy
                        </Button>
                      </Stack>
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
