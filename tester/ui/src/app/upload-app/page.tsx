'use client';
import * as React from 'react';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import { MuiFileInput } from 'mui-file-input';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import { post } from '@/app/util';
import shared from '@/app/shared';
import TextField from '@mui/material/TextField';
import Link from '@mui/material/Link';

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'id', label: 'ID', minWidth: 100 },
  { id: 'appName', label: 'AppName', minWidth: 100 },
  { id: 'appVersions', label: 'AppVersions', minWidth: 200 },
];

interface AppVersion {
  version: number;
  download: string;
}

interface UploadedApp {
  id: string;
  appName: string;
  appVersions?: AppVersion[];
}

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

  const [file, setFile] = React.useState<File | null>(null);
  const handleChange = (newValue: File | null) => {
    setFile(newValue);
  };
  const buttonEnable = file !== null;

  const [appName, setAppName] = React.useState<string>('');
  const [appList, setAppList] = React.useState<null | UploadedApp[]>(null);

  const getUploadedApps = () => {
    post(
      false,
      '',
      '/tester/getUploadedApps/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword() },
      (respData: any) => {
        if (respData.data) {
          setAppList(respData.data.list);
        }
      },
      setToast,
      setErrorToast
    );
  };

  const uploadApp = () => {
    if (!file) {
      setErrorToast('file is null');
      return;
    }
    post(
      false,
      '',
      '/upload?fileName=' + file.name,
      setLoading,
      false,
      file,
      (respData: any) => {
        setFile(null);
        let fileHashName = respData.data.hashName;
        // upload record
        post(
          false,
          '',
          '/tester/uploadApp/v1',
          setLoading,
          true,
          { apiKey: shared.getPassword(), appName, appHashName: fileHashName },
          (respData: any) => {
            getUploadedApps();
          },
          setToast,
          setErrorToast
        );
      },
      setToast,
      setErrorToast
    );
  };

  React.useEffect(() => {
    getUploadedApps();
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
        <Stack spacing={1} direction="row">
          <TextField
            value={appName}
            onChange={(e) => {
              setAppName(e.target.value);
            }}
            id="outlined-basic"
            label="AppName"
            variant="outlined"
            sx={{ width: '400px' }}
          />
          <MuiFileInput
            label="Upload File"
            sx={{ width: '400px' }}
            value={file}
            onChange={handleChange}
          />
          <Button
            disabled={!buttonEnable}
            variant="contained"
            onClick={uploadApp}
          >
            Upload
          </Button>
        </Stack>
      </Stack>
      <Stack spacing={2} direction="row">
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
            {appList && (
              <TableBody>
                {appList.map((row: any) => {
                  return (
                    <TableRow tabIndex={-1} key={row.name}>
                      {columns.map((column) => {
                        const value = row[column.id];
                        if (column.id === 'appVersions') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              {value.map((item: any) => {
                                return (
                                  <Link
                                    sx={{ padding: '10px' }}
                                    href={item.download}
                                    key={item.version}
                                    target="_blank"
                                  >
                                    {item.version}
                                  </Link>
                                );
                              })}
                            </TableCell>
                          );
                        } else {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              {column.format !== undefined &&
                              typeof value === 'number'
                                ? column.format(value)
                                : value}
                            </TableCell>
                          );
                        }
                      })}
                    </TableRow>
                  );
                })}
              </TableBody>
            )}
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
