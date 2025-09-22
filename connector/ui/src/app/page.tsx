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
import moment from 'moment';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select, { SelectChangeEvent } from '@mui/material/Select';
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
  { id: 'id', label: 'ID', minWidth: 100 },
  { id: 'app', label: 'App', minWidth: 100 },
  { id: 'name', label: 'Name', minWidth: 100 },
  { id: 'keyID', label: 'KeyID', minWidth: 100 },
  { id: 'permission', label: 'Permission', minWidth: 100 },
  { id: 'createdAt', label: 'CreatedAt', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [list, setList] = React.useState<any[] | null>(null);
  const [row, setRow] = React.useState<any>(null);
  const [openDialog, setOpenDialog] = React.useState(false);
  const handleCloseDialog = () => {
    setOpenDialog(false);
    setApp('');
    setKeyID('');
    setName('');
    setPermission('read');
    setIsCreate(false);
    setIsDelete(false);
  };

  const [appList, setAppList] = React.useState<string[]>([]);
  const [keyIDList, setKeyIDList] = React.useState<string[]>([]);

  const [isDelete, setIsDelete] = React.useState(false);
  const [isCreate, setIsCreate] = React.useState(false);

  const [app, setApp] = React.useState('pswd_');
  const [keyID, setKeyID] = React.useState('');
  const [name, setName] = React.useState('');
  const [permission, setPermission] = React.useState('read');
  const handleAppChange = (event: SelectChangeEvent) => {
    setApp(event.target.value as string);
  };
  const handleKeyIDChange = (event: SelectChangeEvent) => {
    setKeyID(event.target.value as string);
  };
  const handlePermissionChange = (event: SelectChangeEvent) => {
    setPermission(event.target.value as string);
  };

  const listApiKey = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/riki/console/listApiKey/v1', reqData, (respData) => {
      setList(respData.data.list);
    });
  };

  const addApiKey = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      app: app,
      keyID: keyID,
      name: name,
      permission: permission,
    };
    post('/riki/console/addApiKey/v1', reqData);
  };

  const removeApiKey = async (id: number) => {
    const reqData = {
      apiKey: shared.getPassword(),
      id: id,
    };
    post('/riki/console/removeApiKey/v1', reqData);
  };

  const getAllApps = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/riki/console/getAllApps/v1', reqData, (respData) => {
      setAppList(respData.data.list);
    });
  };

  const getAllKeyIDs = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/riki/console/getAllKeyIDs/v1', reqData, (respData) => {
      setKeyIDList(respData.data.list);
    });
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
    listApiKey();
  }, []);

  return (
    <main>
      <div className="ml-10">
        <Button onClick={listApiKey} variant="contained" sx={{ margin: '5px' }}>
          Refresh
        </Button>
        <Button
          onClick={() => {
            setOpenDialog(true);
            setIsCreate(true);
            setIsDelete(false);
            setApp('pswd_');
            setKeyID('');
            setName('');
            setPermission('read');
            getAllApps();
            getAllKeyIDs();
          }}
          variant="contained"
          sx={{ margin: '5px' }}
        >
          Create
        </Button>
      </div>
      {list != null ? (
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
                      key="download"
                      align="center"
                      style={{ minWidth: 100 }}
                      scope="row"
                    >
                      <Button
                        size="small"
                        color="warning"
                        variant="text"
                        onClick={() => {
                          setOpenDialog(true);
                          setIsCreate(false);
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
        {!isDelete && (
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
                    value={app}
                    label="App"
                    onChange={handleAppChange}
                  >
                    {appList.map((row) => {
                      return (
                        <MenuItem key={row} value={row}>
                          {row}
                        </MenuItem>
                      );
                    })}
                  </Select>
                </FormControl>
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  KeyID
                </Typography>
                <FormControl fullWidth>
                  <InputLabel id="demo-simple-select-label">KeyID</InputLabel>
                  <Select
                    labelId="demo-simple-select-label"
                    id="demo-simple-select"
                    value={keyID}
                    label="KeyID"
                    onChange={handleKeyIDChange}
                  >
                    {keyIDList.map((row) => {
                      return (
                        <MenuItem key={row} value={row}>
                          {row}
                        </MenuItem>
                      );
                    })}
                  </Select>
                </FormControl>
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Name
                </Typography>
                <TextField
                  id="name"
                  name="name"
                  fullWidth
                  margin="normal"
                  value={name}
                  onChange={(e) => {
                    setName(e.target.value);
                  }}
                />
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Permission
                </Typography>
                <FormControl fullWidth>
                  <InputLabel id="demo-simple-select-label">
                    Permission
                  </InputLabel>
                  <Select
                    labelId="demo-simple-select-label"
                    id="demo-simple-select"
                    value={permission}
                    label="Permission"
                    onChange={handlePermissionChange}
                  >
                    <MenuItem value="write">Write</MenuItem>
                    <MenuItem value="read">Read</MenuItem>
                  </Select>
                </FormControl>
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
          <Button
            onClick={() => {
              if (isDelete) {
                removeApiKey(row.id);
              } else {
                if (isCreate) {
                  addApiKey();
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
