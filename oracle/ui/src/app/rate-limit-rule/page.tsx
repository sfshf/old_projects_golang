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
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import FormControl from '@mui/material/FormControl';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';

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
  { id: 'type', label: 'Type', minWidth: 100 },
  { id: 'target', label: 'Target', minWidth: 100 },
  { id: 'capacity', label: 'Capacity', minWidth: 100 },
  { id: 'enabled', label: 'Enabled', minWidth: 100 },
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
  };
  const [isDelete, setIsDelete] = React.useState(false);
  const [isCreate, setIsCreate] = React.useState(false);
  const [id, setId] = React.useState(0);
  const [type, setType] = React.useState(0);
  const typeMsg = (type: number): string => {
    switch (type) {
      case 1:
        return 'service';
      case 2:
        return 'url_path';
    }
    return '';
  };
  const typeList = [
    {
      id: 'service',
      val: 1,
    },
    {
      id: 'url_path',
      val: 2,
    },
  ];
  const handleTypeChange = (event: SelectChangeEvent) => {
    setType(parseInt(event.target.value));
  };
  const [target, setTarget] = React.useState('');
  const [enabled, setEnabled] = React.useState(false);
  const enabledMsg = (enabled: boolean): string => {
    switch (enabled) {
      case true:
        return 'True';
      case false:
        return 'False';
    }
    return '';
  };
  const handleEnabledChange = (event: SelectChangeEvent) => {
    setEnabled(event.target.value == 'true');
  };
  const enabledList = [
    {
      id: 'True',
      val: 'true',
    },
    {
      id: 'False',
      val: 'false',
    },
  ];
  const [capacity, setCapacity] = React.useState(0);

  const listRateLimitRules = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/console/listRateLimitRules/v1', reqData, (respData) => {
      setList(respData.data.list);
    });
  };

  const addRateLimitRule = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      type: type,
      target: target,
      capacity: capacity,
      enabled: enabled,
    };
    post('/console/addRateLimitRule/v1', reqData, (respData) => {});
  };

  const deleteRateLimitRule = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      id: id,
    };
    post('/console/deleteRateLimitRule/v1', reqData, (respData) => {});
  };

  const updateRateLimitRule = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
      id: id,
      type: type,
      target: target,
      capacity: capacity,
      enabled: enabled,
    };
    post('/console/updateRateLimitRule/v1', reqData, (respData) => {});
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
    listRateLimitRules();
  }, []);

  return (
    <main>
      <Stack direction="row" alignItems="center">
        <Button
          onClick={() => {
            listRateLimitRules();
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
            setType(0);
            setTarget('');
            setCapacity(0);
            setEnabled(false);
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Create
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
                      } else if (column.id == 'type') {
                        return (
                          <TableCell key={column.id} align={column.align}>
                            {typeMsg(value)}
                          </TableCell>
                        );
                      } else if (column.id == 'enabled') {
                        return (
                          <TableCell
                            key={column.id}
                            align={column.align}
                            sx={
                              value
                                ? {
                                    color: 'green',
                                  }
                                : {
                                    color: 'red',
                                  }
                            }
                          >
                            {enabledMsg(value)}
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
                        color="primary"
                        variant="text"
                        onClick={() => {
                          setOpenDialog(true);
                          setIsCreate(false);
                          setIsDelete(false);
                          setId(row.id);
                          setType(row.type);
                          setTarget(row.target);
                          setCapacity(row.capacity);
                          setEnabled(row.enabled);
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
                          setId(row.id);
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
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Type
                </Typography>
                <FormControl sx={{ width: '100%' }}>
                  <InputLabel id="demo-simple-select-label">Type</InputLabel>
                  <Select
                    labelId="demo-simple-select-label"
                    id="demo-simple-select"
                    value={type.toString()}
                    label="Type"
                    onChange={handleTypeChange}
                  >
                    {typeList &&
                      typeList.map((item) => (
                        <MenuItem value={item.val} key={item.id}>
                          {item.id}
                        </MenuItem>
                      ))}
                  </Select>
                </FormControl>
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Target
                </Typography>
                <TextField
                  id="Target"
                  name="Target"
                  fullWidth
                  margin="normal"
                  value={target}
                  onChange={(e) => {
                    setTarget(e.target.value);
                  }}
                />
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Capacity
                </Typography>
                <TextField
                  id="Capacity"
                  name="Capacity"
                  type="number"
                  fullWidth
                  margin="normal"
                  value={capacity}
                  onChange={(e) => {
                    setCapacity(parseInt(e.target.value));
                  }}
                />
              </Stack>
              <Stack direction="row" alignItems="center">
                <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                  Enabled
                </Typography>
                <FormControl sx={{ width: '100%' }}>
                  <InputLabel id="demo-simple-select-label">Enabled</InputLabel>
                  <Select
                    labelId="demo-simple-select-label"
                    id="demo-simple-select"
                    value={enabled.toString()}
                    label="Enabled"
                    onChange={handleEnabledChange}
                  >
                    {enabledList &&
                      enabledList.map((item) => (
                        <MenuItem value={item.val} key={item.id}>
                          {item.id}
                        </MenuItem>
                      ))}
                  </Select>
                </FormControl>
              </Stack>
            </DialogContent>
          </>
        )}
        {isDelete && (
          <DialogContent>
            <Typography variant="h6" gutterBottom>
              Are you sure to delete rule {id} ?
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
                deleteRateLimitRule();
              } else {
                if (isCreate) {
                  addRateLimitRule();
                } else {
                  updateRateLimitRule();
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
