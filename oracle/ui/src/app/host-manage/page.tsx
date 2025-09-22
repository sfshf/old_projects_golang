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
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
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
  { id: 'rawURL', label: 'RawURL', minWidth: 200 },
  { id: 'createdAt', label: 'CreatedAt', minWidth: 100 },
  { id: 'updatedAt', label: 'UpdatedAt', minWidth: 100 },
  { id: 'operations', label: 'Operations', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [list, setList] = React.useState<any[] | null>(null);
  const [openDialog, setOpenDialog] = React.useState(false);
  const handleCloseDialog = () => {
    setOpenDialog(false);
  };
  const [isDelete, setIsDelete] = React.useState(false);
  const [isCreate, setIsCreate] = React.useState(false);
  const [id, setId] = React.useState(0);
  const [domain, setDomain] = React.useState('');
  const [rawURL, setRawURL] = React.useState('');

  const [oracleGatewayHostname, setOracleGatewayHostname] =
    React.useState('api.test.n1xt.net');

  const listHostnames = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/console/listHostnames/v1', reqData, (respData) => {
      setList(respData.data.list);
      for (let i = 0; i < respData.data.list.length; i++) {
        if (respData.data.list[i].domain === 'api.n1xt.net') {
          setOracleGatewayHostname('api.n1xt.net');
          break;
        }
      }
    });
  };

  const createHostname = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      domain,
      rawURL,
    };
    post('/console/createHostname/v1', reqData, (respData) => {});
  };

  const updateHostname = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      id,
      rawURL,
    };
    post('/console/updateHostname/v1', reqData, (respData) => {});
  };

  const deleteHostname = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      id,
    };
    post('/console/deleteHostname/v1', reqData, (respData) => {});
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
          if (resp.ok) {
            return resp.json();
          } else {
            setErrorToast('request error:' + resp.status);
            throw 'request error:' + resp.status;
          }
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
    listHostnames();
  }, []);

  return (
    <main>
      <Stack direction="row" alignItems="center">
        <Button
          onClick={() => {
            listHostnames();
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Refresh
        </Button>
        <Button
          onClick={() => {
            setOpenDialog(true);
            setIsCreate(true);
            setIsDelete(false);
            setId(0);
            setDomain('');
            setRawURL('');
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Create
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
                        column.id == 'updatedAt'
                      ) {
                        return (
                          <TableCell key={column.id} align={column.align}>
                            {moment(value)
                              .local()
                              .format('YYYY-MM-DD HH:mm:ss')}
                          </TableCell>
                        );
                      } else if (column.id == 'operations') {
                        return (
                          <TableCell key={column.id} align={column.align}>
                            {row.domain != oracleGatewayHostname && (
                              <>
                                <Button
                                  variant="text"
                                  size="small"
                                  onClick={() => {
                                    setOpenDialog(true);
                                    setIsCreate(false);
                                    setIsDelete(false);
                                    setId(row.id);
                                    setDomain(row.domain);
                                    setRawURL(row.rawURL);
                                  }}
                                >
                                  Update
                                </Button>
                                <Button
                                  variant="text"
                                  size="small"
                                  onClick={() => {
                                    setOpenDialog(true);
                                    setIsDelete(true);
                                    setId(row.id);
                                  }}
                                >
                                  Delete
                                </Button>
                              </>
                            )}
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
      <Dialog
        maxWidth="lg"
        fullWidth
        open={openDialog}
        onClose={handleCloseDialog}
      >
        {!isDelete && (
          <>
            <DialogTitle>Properties</DialogTitle>
            <DialogContent>
              {isCreate && (
                <Stack direction="row" alignItems="center">
                  <Typography
                    variant="body1"
                    gutterBottom
                    sx={{ width: '25%' }}
                  >
                    Domain
                  </Typography>
                  <TextField
                    fullWidth
                    margin="normal"
                    value={domain}
                    onChange={(e) => {
                      setDomain(e.target.value);
                    }}
                  />
                </Stack>
              )}
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  RawURL
                </Typography>
                <TextField
                  fullWidth
                  margin="normal"
                  value={rawURL}
                  onChange={(e) => {
                    setRawURL(e.target.value);
                  }}
                />
              </Stack>
            </DialogContent>
          </>
        )}
        {isDelete && (
          <DialogContent>
            <Typography variant="h6" gutterBottom>
              Are you sure to delete {domain} ?
            </Typography>
          </DialogContent>
        )}
        <DialogActions>
          <Button
            onClick={() => {
              handleCloseDialog();
            }}
          >
            Close
          </Button>
          <Button
            onClick={() => {
              if (isDelete) {
                deleteHostname();
              } else {
                if (isCreate) {
                  createHostname();
                } else {
                  updateHostname();
                }
              }
            }}
          >
            Subscribe
          </Button>
        </DialogActions>
        <Backdrop
          sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
          open={loading}
        >
          <CircularProgress color="inherit" />
        </Backdrop>
      </Dialog>
    </main>
  );
}
