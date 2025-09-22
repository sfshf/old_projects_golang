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
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import keccak256 from 'keccak256';
import { XChaCha20Poly1305 } from '@stablelib/xchacha20poly1305';
import moment from 'moment';

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
  { id: 'app', label: 'App', minWidth: 100 },
  { id: 'config', label: 'Config', minWidth: 100 },
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

  const [app, setApp] = React.useState('');
  const [config, setConfig] = React.useState('');
  const [isDelete, setIsDelete] = React.useState(false);
  const [isCreate, setIsCreate] = React.useState(false);

  const listAppConfigs = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/riki/console/listConfigs/v1', reqData, (respData) => {
      setList(respData.data.configList);
    });
  };

  const createAppConfig = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
      config: config,
    };
    post('/riki/console/createConfig/v1', reqData);
  };

  const updateAppConfig = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
      config: config,
    };
    post('/riki/console/updateConfig/v1', reqData);
  };

  const deleteAppConfig = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
    };
    post('/riki/console/deleteConfig/v1', reqData);
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

  const configExample = `{
  "appName": "test"
}`;

  useEffect(() => {
    listAppConfigs();
  }, []);

  return (
    <main>
      <div className="ml-10">
        <Button
          onClick={listAppConfigs}
          variant="contained"
          sx={{ margin: '5px' }}
        >
          Refresh
        </Button>
        <Button
          onClick={() => {
            setOpenDialog(true);
            setIsCreate(true);
            setIsDelete(false);
            setApp('');
            setConfig('');
          }}
          variant="contained"
          sx={{ margin: '5px' }}
        >
          Create
        </Button>
      </div>
      {list !== null ? (
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
                <TableCell
                  key="operations"
                  align="center"
                  style={{ minWidth: 100 }}
                >
                  Operations
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {list.map((row) => {
                return (
                  <TableRow hover role="checkbox" tabIndex={-1} key={row.id}>
                    {columns.map((column) => {
                      const value = row[column.id];
                      return (
                        <TableCell key={column.id} align={column.align}>
                          {column.format !== undefined &&
                          typeof value === 'number'
                            ? column.format(value)
                            : value}
                        </TableCell>
                      );
                    })}
                    <TableCell
                      key="download"
                      align="center"
                      style={{ minWidth: 100 }}
                      scope="row"
                    >
                      <Button
                        size="small"
                        color="primary"
                        variant="text"
                        onClick={() => {
                          setOpenDialog(true);
                          setIsCreate(false);
                          setIsDelete(false);
                          setApp(row.app);
                          setConfig(row.config);
                        }}
                      >
                        Update
                      </Button>
                      <Button
                        size="small"
                        color="warning"
                        variant="text"
                        onClick={() => {
                          setOpenDialog(true);
                          setIsDelete(true);
                          setApp(row.app);
                        }}
                      >
                        Delete
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      ) : null}
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
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  App
                </Typography>
                <TextField
                  id="app"
                  name="app"
                  fullWidth
                  margin="normal"
                  value={app}
                  onChange={(e) => {
                    setApp(e.target.value);
                  }}
                />
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Config
                </Typography>
                <TextField
                  id="config"
                  name="config"
                  fullWidth
                  margin="normal"
                  value={config}
                  onChange={(e) => {
                    setConfig(e.target.value);
                  }}
                  multiline
                  rows="9"
                />
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Config Example
                </Typography>
                <TextField
                  id="config"
                  name="config"
                  fullWidth
                  margin="normal"
                  value={configExample}
                  multiline
                  rows="9"
                  disabled
                />
              </Stack>
            </DialogContent>
          </>
        )}
        {isDelete && (
          <DialogContent>
            <Typography variant="h6" gutterBottom>
              Are you sure to delete {app} ?
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
                deleteAppConfig();
              } else {
                if (isCreate) {
                  createAppConfig();
                } else {
                  updateAppConfig();
                }
              }
            }}
          >
            Subscribe
          </Button>
        </DialogActions>
      </Dialog>
    </main>
  );
}
