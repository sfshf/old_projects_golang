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
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select, { SelectChangeEvent } from '@mui/material/Select';
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
  { id: 'keyID', label: 'KeyID', minWidth: 100 },
  { id: 'app', label: 'App', minWidth: 100 },
  { id: 'passwordHash', label: 'PasswordHash', minWidth: 100 },
  { id: 'createdAt', label: 'CreatedAt', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [list, setList] = React.useState<any[] | null>(null);
  const [openDialog, setOpenDialog] = React.useState(false);
  const handleCloseDialog = () => {
    setOpenDialog(false);
    setIsShow(false);
    setIsCreate(false);
    setIsDelete(false);
    setDialogApp('');
    setDialogKeyID('');
    setPassword('');
  };
  const [row, setRow] = React.useState<any>(null);

  const [app, setApp] = React.useState('');
  const handleSelectChange = (event: SelectChangeEvent) => {
    setApp(event.target.value as string);
  };
  const [dialogApp, setDialogApp] = React.useState('');
  const handleDialogSelectChange = (event: SelectChangeEvent) => {
    setDialogApp(event.target.value as string);
  };
  const [dialogKeyID, setDialogKeyID] = React.useState('');

  const [appList, setAppList] = React.useState<string[] | null>(null);
  const [isDelete, setIsDelete] = React.useState(false);
  const [isCreate, setIsCreate] = React.useState(false);
  const [isShow, setIsShow] = React.useState(false);
  const [password, setPassword] = React.useState('');

  const getAllApps = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/riki/console/getAllApps/v1', reqData, (respData) => {
      setAppList(respData.data.list);
    });
  };

  const listPassword = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
    };
    post('/riki/console/listPassword/v1', reqData, (respData) => {
      setList(respData.data.list);
    });
  };

  const addPassword = async (app: string, keyID: string) => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
      keyID: keyID,
    };
    post('/riki/console/addPassword/v1', reqData);
  };

  const fetchPassword = async (
    app: string,
    keyID: string,
    passwordHash: string
  ) => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
      keyID: keyID,
    };
    post('/riki/console/fetchPassword/v1', reqData, (respData) => {
      setPassword(respData.data.password);
    });
  };

  const removePassword = async (app: string, keyID: string) => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
      keyID: keyID,
    };
    post('/riki/console/removePassword/v1', reqData);
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

  useEffect(() => {
    getAllApps();
    listPassword();
  }, []);

  return (
    <main>
      <Stack direction="row" alignItems="center">
        <Button
          onClick={() => {
            listPassword();
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
            setDialogApp('');
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Create
        </Button>
        <FormControl sx={{ width: '200px', margin: '5px' }}>
          <InputLabel id="demo-simple-select-label">App</InputLabel>
          <Select
            labelId="demo-simple-select-label"
            id="demo-simple-select"
            value={app}
            label="App"
            onChange={handleSelectChange}
          >
            <MenuItem value="">NULL</MenuItem>
            {appList &&
              appList.map((item) => (
                <MenuItem key={item} value={item}>
                  {item}
                </MenuItem>
              ))}
          </Select>
        </FormControl>
      </Stack>
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
                    <TableCell
                      key="id"
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
                          setIsShow(true);
                          setIsCreate(false);
                          setIsDelete(false);
                          fetchPassword(row.app, row.keyID, row.passwordHash);
                        }}
                      >
                        Show
                      </Button>
                      <Button
                        size="small"
                        color="warning"
                        variant="text"
                        onClick={() => {
                          setOpenDialog(true);
                          setIsDelete(true);
                          setRow(row);
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
        {isShow && (
          <>
            <DialogTitle>Password</DialogTitle>
            <DialogContent>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  {password}
                </Typography>
              </Stack>
            </DialogContent>
          </>
        )}
        {isCreate && (
          <>
            <DialogTitle>Properties</DialogTitle>
            <DialogContent>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  App
                </Typography>
                <FormControl fullWidth>
                  <InputLabel id="demo-simple-select-label">App</InputLabel>
                  <Select
                    labelId="demo-simple-select-label"
                    id="demo-simple-select"
                    value={dialogApp}
                    label="App"
                    onChange={handleDialogSelectChange}
                  >
                    {appList &&
                      appList.map((item) => (
                        <MenuItem key={item} value={item}>
                          {item}
                        </MenuItem>
                      ))}
                  </Select>
                </FormControl>
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  KeyID
                </Typography>
                <TextField
                  id="keyID"
                  name="keyID"
                  fullWidth
                  margin="normal"
                  value={dialogKeyID}
                  onChange={(e) => {
                    setDialogKeyID(e.target.value);
                  }}
                />
              </Stack>
            </DialogContent>
          </>
        )}
        {isDelete && (
          <DialogContent>
            <Typography variant="h6" gutterBottom>
              Are you sure to remove {row && row.keyID} ?
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
          {!isShow && (
            <Button
              onClick={() => {
                if (isDelete) {
                  removePassword(row.app, row.keyID);
                } else {
                  if (isCreate) {
                    addPassword(dialogApp, dialogKeyID);
                  } else {
                  }
                }
              }}
            >
              Subscribe
            </Button>
          )}
        </DialogActions>
      </Dialog>
    </main>
  );
}
