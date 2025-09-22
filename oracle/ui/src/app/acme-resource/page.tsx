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
  { id: 'domain', label: 'Domain', minWidth: 100 },
  { id: 'expectedRefreshedAt', label: 'ExpectedRefreshedAt', minWidth: 100 },
  { id: 'createdAt', label: 'CreatedAt', minWidth: 100 },
  { id: 'updatedAt', label: 'UpdatedAt', minWidth: 100 },
  { id: 'expiresAt', label: 'ExpiresAt', minWidth: 100 },
  { id: 'operations', label: 'Operations', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [list, setList] = React.useState<any[] | null>(null);

  const listAcmeResources = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/console/listAcmeResources/v1', reqData, (respData) => {
      setList(respData.data.list);
    });
  };

  const renewAcmeResource = async (domain: string) => {
    const reqData = {
      apiKey: shared.getPassword(),
      domain,
    };
    post('/console/renewAcmeResource/v1', reqData, (respData) => {});
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
      fetch(path, reqOpts)
        .then((resp) => {
          setLoading(false);
          return resp.json();
        })
        .then((data) => {
          if (data.code !== 0) {
            setErrorToast(data.debugMessage);
            return;
          }
          if (successAction) {
            successAction(data);
          }
          setToast('success');
        })
        .then((err) => {
          throw err;
        });
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  useEffect(() => {
    listAcmeResources();
  }, []);

  return (
    <main>
      <Stack direction="row" alignItems="center">
        <Button
          onClick={() => {
            listAcmeResources();
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Refresh
        </Button>
      </Stack>
      {list && (
        <TableContainer>
          <Table>
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
                      if (
                        column.id == 'createdAt' ||
                        column.id == 'expectedRefreshedAt'
                      ) {
                        return (
                          <TableCell key={column.id} align={column.align}>
                            {value > 0
                              ? moment(value)
                                  .local()
                                  .format('YYYY-MM-DD HH:mm:ss')
                              : 'NULL'}
                          </TableCell>
                        );
                      } else if (column.id == 'updatedAt') {
                        return (
                          <>
                            <TableCell key={column.id} align={column.align}>
                              {value > 0
                                ? moment(value)
                                    .local()
                                    .format('YYYY-MM-DD HH:mm:ss')
                                : 'NULL'}
                            </TableCell>
                            <TableCell key={column.id} align={column.align}>
                              {value > 0
                                ? moment(value)
                                    .add(90, 'days')
                                    .local()
                                    .format('YYYY-MM-DD HH:mm:ss')
                                : 'NULL'}
                            </TableCell>
                          </>
                        );
                      } else if (column.id == 'operations') {
                        return (
                          <TableCell key={column.id} align={column.align}>
                            {row['createdAt'] == 0 ? (
                              <Button
                                variant="text"
                                size="small"
                                onClick={() => {
                                  renewAcmeResource(row.domain);
                                }}
                              >
                                Renew
                              </Button>
                            ) : (
                              <></>
                            )}
                          </TableCell>
                        );
                      } else if (column.id != 'expiresAt') {
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
