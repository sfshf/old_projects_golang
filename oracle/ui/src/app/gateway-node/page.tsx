'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { useEffect } from 'react';
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
import { request, RequestOptions } from 'http';
import Stack from '@mui/material/Stack';
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
  { id: 'id', label: 'ID', minWidth: 100 },
  { id: 'name', label: 'Name', minWidth: 100 },
  { id: 'domain', label: 'Domain', minWidth: 100 },
  { id: 'createdAt', label: 'CreatedAt', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [list, setList] = React.useState<any[] | null>(null);

  const listGatewayNodes = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/console/listGatewayNodes/v1', reqData, (respData) => {
      setList(respData.data.list);
    });
  };

  const post = async (
    path: string,
    reqData?: any,
    successAction?: (respData: any) => void
  ) => {
    setLoading(true);

    let postData;
    if (reqData) {
      postData = JSON.stringify(reqData);
    }
    try {
      let reqOpts: RequestOptions = {
        path: path,
        method: 'POST',
      };
      if (postData) {
        reqOpts.headers = {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
        };
      } else {
        reqOpts.headers = {
          'Content-Type': 'application/json',
        };
      }
      const req = request(reqOpts, (res) => {
        res.setEncoding('utf8');
        res.on('data', (chunk) => {
          const respData = JSON.parse(chunk);
          if (respData.code !== 0) {
            setErrorToast(respData.debugMessage);
            return;
          }
          if (successAction) {
            successAction(respData);
          }
          setToast('success');
        });
      });
      req.on('error', (e) => {
        throw e;
      });
      if (postData) {
        req.write(postData);
      }
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  useEffect(() => {
    listGatewayNodes();
  }, []);

  return (
    <main>
      <Stack direction="row" alignItems="center">
        <Button
          onClick={() => {
            listGatewayNodes();
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Refresh
        </Button>
      </Stack>
      {list && (
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
              {list.map((row) => {
                return (
                  <TableRow hover role="checkbox" tabIndex={-1} key={row.id}>
                    {columns.map((column) => {
                      const value = row[column.id];
                      if (column.id == 'createdAt') {
                        return (
                          <TableCell key={column.id} align={column.align}>
                            {moment(value)
                              .local()
                              .format('YYYY-MM-DD HH:mm:ss')}
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
          </Table>
        </TableContainer>
      )}
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
      <Snackbar
        open={toast !== ''}
        autoHideDuration={2000}
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
