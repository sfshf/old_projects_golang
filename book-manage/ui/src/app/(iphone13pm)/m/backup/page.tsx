'use client';
import * as React from 'react';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '@/app/shared';
import Backdrop from '@mui/material/Backdrop';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import { request } from 'http';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch, { SwitchProps } from '@mui/material/Switch';
import { styled } from '@mui/material/styles';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import CircularProgress, {
  CircularProgressProps,
} from '@mui/material/CircularProgress';
import Box from '@mui/material/Box';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import LinearProgress from '@mui/material/LinearProgress';
import CloseIcon from '@mui/icons-material/Close';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import IconButton from '@mui/material/IconButton';

interface BookItem {
  id: number;
  name: string;
  updatedAt: string;
  updatedAtText: string;
  description: string;
}

interface BackupLogItem {
  id: number;
  book_id: number;
  filepath: string;
  created_at: string;
}

interface MakeBackupRequest {
  password: string;
  book: string;
}

interface RegainBackupRequest {
  password: string;
  bookID: number;
  backupID: number;
}

interface SetCronJobRequest {
  password: string;
  startCron: string;
  scheduleSpec: string;
}

const CircularProgressWithLabel = (
  props: CircularProgressProps & { value: number }
) => {
  return (
    <Box sx={{ position: 'relative', display: 'inline-flex' }}>
      <CircularProgress variant="determinate" {...props} />
      <Box
        sx={{
          top: 0,
          left: 0,
          bottom: 0,
          right: 0,
          position: 'absolute',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Typography
          variant="caption"
          component="div"
          color="text.secondary"
        >{`${Math.round(props.value)}%`}</Typography>
      </Box>
    </Box>
  );
};

const CircularWithValueLabel = () => {
  const [progress, setProgress] = React.useState(10);

  React.useEffect(() => {
    const timer = setInterval(() => {
      setProgress((prevProgress) =>
        prevProgress >= 100 ? 0 : prevProgress + 10
      );
    }, 800);
    return () => {
      clearInterval(timer);
    };
  }, []);

  return <CircularProgressWithLabel value={progress} />;
};

export default function Page() {
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [bookRows, setBookRows] = React.useState<BookItem[] | null>(null);
  const [backupRows, setBackupRows] = React.useState<BackupLogItem[] | null>(
    null
  );

  const [loading, setLoading] = React.useState(false);

  const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
    props,
    ref
  ) {
    return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
  });

  const listBooks = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    try {
      const res = await fetch(
        `${shared.baseAPIURL}/book/allbooks?` +
          new URLSearchParams({ password }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      const books = data.data.books;
      if (books && books.length > 0) {
        books.forEach((book: BookItem) => {
          book.updatedAtText = new Date(book.updatedAt).toLocaleString();
        });
      }
      setBookRows(books);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  const listBackups = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    try {
      let param = new URLSearchParams({
        password,
      });
      const res = await fetch(
        `${shared.baseAPIURL}/backup/all?` + param.toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setBackupRows(data.data);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  interface Column {
    id: string;
    label: string;
    minWidth?: number;
    align?: 'right' | 'left';
    format?: (value: number) => string;
  }

  const bookColumns: readonly Column[] = [
    { id: 'id', label: 'Book\u00a0ID', minWidth: 50, align: 'left' },
    { id: 'name', label: 'Book\u00a0Name', minWidth: 100, align: 'right' },
    {
      id: 'updatedAtText',
      label: 'Update\u00a0At',
      minWidth: 150,
      align: 'right',
    },
  ];

  const backupColums: readonly Column[] = [
    { id: 'id', label: 'Backup\u00a0ID', minWidth: 100, align: 'right' },
    { id: 'filepath', label: 'Filepath', minWidth: 100, align: 'right' },
    {
      id: 'created_at',
      label: 'Created\u00a0At',
      minWidth: 100,
      align: 'right',
    },
  ];

  const makeBackup = async (book: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: MakeBackupRequest = {
      password: password,
      book: book,
    };

    try {
      const postData = JSON.stringify(reqData);
      const req = request(
        {
          path: shared.baseAPIURL + '/backup/make',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData),
          },
        },
        (res) => {
          res.setEncoding('utf8');
          res.on('data', (chunk) => {
            const respData = JSON.parse(chunk);
            if (respData.code !== 0) {
              setErrorToast(respData.message);
              return;
            }
            listBackups();
            setToast('success');
          });
        }
      );
      req.on('error', (e) => {
        throw e;
      });
      req.write(postData);
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const onClickRegainButton = async (bookID: number, backupID: number) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: RegainBackupRequest = {
      password: password,
      bookID: bookID,
      backupID: backupID,
    };

    try {
      const postData = JSON.stringify(reqData);
      const req = request(
        {
          path: shared.baseAPIURL + '/backup/regain',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData),
          },
        },
        (res) => {
          res.setEncoding('utf8');
          res.on('data', (chunk) => {
            const respData = JSON.parse(chunk);
            if (respData.code !== 0) {
              setErrorToast(respData.message);
              return;
            }
            setShowLogs(true);
            setLogs([]);
            setWordCount(0);
            setProgress(0);
            timerRef.current = setInterval(() => {
              checkRegaining();
            }, 200);
            setToast('success');
          });
        }
      );
      req.on('error', (e) => {
        throw e;
      });
      req.write(postData);
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const [showLogs, setShowLogs] = React.useState(false);
  const [logs, setLogs] = React.useState<string[]>([]);
  const [wordCount, setWordCount] = React.useState(0);
  const [progress, setProgress] = React.useState(0);
  const timerRef = React.useRef<NodeJS.Timer | null>(null);

  const handleClose = () => {
    setShowLogs(false);
  };

  const checkRegaining = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/backup/logs?` +
          new URLSearchParams({ password }).toString()
      );
      const data = await res.json();
      if (data.code !== 0) {
        setLogs((prev) => {
          return [data.message, ...prev];
        });
        return;
      }
      setLogs((prev) => {
        const newLogs = data.data.logs as string[];
        return [...newLogs.reverse(), ...prev];
      });
      setWordCount(parseInt(data.data.wordCount));
      setProgress(parseInt(data.data.progress));

      if (data.data.progress === 100) {
        clearInterval(timerRef.current as NodeJS.Timer);
        timerRef.current = null;
        setToast('Success: Book saved');
      } else if (data.data.error !== '') {
        clearInterval(timerRef.current as NodeJS.Timer);
        timerRef.current = null;
        setErrorToast(' Error :' + data.data.error);
      }
    } catch (error) {
      const message = (error as Error).message;
      setLogs((prev) => {
        return [message, ...prev];
      });
    }
  };

  const [openCronSetting, setOpenCronSetting] = React.useState(false);

  const handleCronSettingDialogClose = () => {
    setOpenCronSetting(false);
  };

  const IOSSwitch = styled((props: SwitchProps) => (
    <Switch
      focusVisibleClassName=".Mui-focusVisible"
      disableRipple
      {...props}
    />
  ))(({ theme }) => ({
    width: 42,
    height: 26,
    padding: 0,
    '& .MuiSwitch-switchBase': {
      padding: 0,
      margin: 2,
      transitionDuration: '300ms',
      '&.Mui-checked': {
        transform: 'translateX(16px)',
        color: '#fff',
        '& + .MuiSwitch-track': {
          backgroundColor:
            theme.palette.mode === 'dark' ? '#2ECA45' : '#65C466',
          opacity: 1,
          border: 0,
        },
        '&.Mui-disabled + .MuiSwitch-track': {
          opacity: 0.5,
        },
      },
      '&.Mui-focusVisible .MuiSwitch-thumb': {
        color: '#33cf4d',
        border: '6px solid #fff',
      },
      '&.Mui-disabled .MuiSwitch-thumb': {
        color:
          theme.palette.mode === 'light'
            ? theme.palette.grey[100]
            : theme.palette.grey[600],
      },
      '&.Mui-disabled + .MuiSwitch-track': {
        opacity: theme.palette.mode === 'light' ? 0.7 : 0.3,
      },
    },
    '& .MuiSwitch-thumb': {
      boxSizing: 'border-box',
      width: 22,
      height: 22,
    },
    '& .MuiSwitch-track': {
      borderRadius: 26 / 2,
      backgroundColor: theme.palette.mode === 'light' ? '#E9E9EA' : '#39393D',
      opacity: 1,
      transition: theme.transitions.create(['background-color'], {
        duration: 500,
      }),
    },
  }));

  const scheduleSpecs = ['@weekly', '@daily', '@hourly', '@every 60s'].sort();

  const [cronStart, setCronStart] = React.useState(false);
  const [startedOrStopedAt, setStartedOrStopedAt] = React.useState('');
  const [nextTime, setNextTime] = React.useState('');
  const [scheduleSpec, setScheduleSpec] = React.useState('');
  const [lastExecError, setLastExecError] = React.useState('');

  const onClickCronSetting = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    try {
      const res = await fetch(
        `${shared.baseAPIURL}/backup/cron_status?` +
          new URLSearchParams({ password }).toString()
      );

      setLoading(false);

      const respData = await res.json();
      if (respData.code !== 0) {
        setErrorToast(respData.message);
        return;
      }

      setCronStart(respData.data.started);
      setStartedOrStopedAt(
        new Date(respData.data.startedOrStopedAt).toLocaleString()
      );
      setNextTime(new Date(respData.data.nextTime).toLocaleString());
      setScheduleSpec(respData.data.scheduleSpec);
      setLastExecError(respData.data.lastExecError);
      setOpenCronSetting(true);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  const updateCronSetting = async (startCron: string, scheduleSpec: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: SetCronJobRequest = {
      password: password,
      startCron: startCron,
      scheduleSpec: scheduleSpec,
    };

    try {
      const postData = JSON.stringify(reqData);
      const req = request(
        {
          path: shared.baseAPIURL + '/backup/update_cron_setting',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData),
          },
        },
        (res) => {
          res.setEncoding('utf8');
          res.on('data', (chunk) => {
            const respData = JSON.parse(chunk);
            if (respData.code !== 0) {
              setErrorToast(respData.message);
              return;
            }

            setCronStart(respData.data.started);
            setStartedOrStopedAt(
              new Date(respData.data.startedOrStopedAt).toLocaleString()
            );
            setNextTime(new Date(respData.data.nextTime).toLocaleString());
            setScheduleSpec(respData.data.scheduleSpec);
            setLastExecError(respData.data.lastExecError);
            setToast('success');
          });
        }
      );
      req.on('error', (e) => {
        throw e;
      });
      req.write(postData);
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const onChangeStatusOfCron = (e: React.BaseSyntheticEvent) => {
    updateCronSetting(e.target.checked ? 'start' : 'stop', '');
  };

  const onChangeScheduleSpec = (e: SelectChangeEvent) => {
    setScheduleSpec(scheduleSpec);
    updateCronSetting('', e.target.value);
  };

  return (
    <Stack
      spacing={3}
      direction="column"
      alignItems="center"
      sx={{
        padding: '1rem',
      }}
    >
      <Stack
        spacing={3}
        direction="row"
        alignItems="center"
        sx={{
          padding: '1rem',
        }}
      >
        <Button
          variant="contained"
          size="small"
          onClick={() => {
            listBooks();
            listBackups();
          }}
        >
          Refresh
        </Button>
        <Button
          variant="contained"
          size="small"
          color="warning"
          onClick={() => {
            makeBackup('all');
          }}
        >
          Backup All
        </Button>
        <Button
          variant="contained"
          size="small"
          color="info"
          onClick={onClickCronSetting}
        >
          Cron Setting
        </Button>
      </Stack>
      <TableContainer component={Paper}>
        <Table aria-label="collapsible table">
          <TableHead>
            <TableRow>
              {bookColumns.map((column) => (
                <TableCell
                  key={column.id}
                  align={column.align}
                  style={{ minWidth: column.minWidth }}
                >
                  {column.label}
                </TableCell>
              ))}
              {backupColums.map((column) => (
                <TableCell
                  key={column.id}
                  align={column.align}
                  style={{ minWidth: column.minWidth }}
                >
                  {column.label}
                </TableCell>
              ))}
              <TableCell
                key="operattion"
                align="right"
                style={{ minWidth: '100px' }}
              >
                Operations
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {bookRows &&
              bookRows.map((item) => {
                let backups = new Array<BackupLogItem>();
                if (backupRows) {
                  for (let i = 0; i < backupRows.length; i++) {
                    if (backupRows[i].book_id === item.id) {
                      backups.push(backupRows[i]);
                    }
                  }
                }
                if (backups.length > 0) {
                  return (
                    <>
                      {backups.map((one, index) => {
                        if (index > 0) {
                          return (
                            <TableRow key={one.id}>
                              <TableCell align="right">{one.id}</TableCell>
                              <TableCell align="right">
                                {one.filepath}
                              </TableCell>
                              <TableCell align="right">
                                {new Date(one.created_at).toLocaleString()}
                              </TableCell>
                              <TableCell align="right">
                                <Button
                                  variant="contained"
                                  size="small"
                                  color="success"
                                  onClick={() => {
                                    onClickRegainButton(item.id, one.id);
                                  }}
                                >
                                  Regain
                                </Button>
                              </TableCell>
                            </TableRow>
                          );
                        } else {
                          return (
                            <TableRow key={one.id}>
                              <TableCell align="left" rowSpan={backups.length}>
                                <Typography>{item.id}</Typography>
                                <br />
                                <Button
                                  variant="contained"
                                  size="small"
                                  color="warning"
                                  onClick={() => {
                                    makeBackup(item.id.toString(10));
                                  }}
                                >
                                  Backup
                                </Button>
                              </TableCell>
                              <TableCell align="right" rowSpan={backups.length}>
                                {item.name}
                              </TableCell>
                              <TableCell align="right" rowSpan={backups.length}>
                                {item.updatedAtText}
                              </TableCell>
                              <TableCell align="right">{one.id}</TableCell>
                              <TableCell align="right">
                                {one.filepath}
                              </TableCell>
                              <TableCell align="right">
                                {new Date(one.created_at).toLocaleString()}
                              </TableCell>
                              <TableCell align="right">
                                <Button
                                  variant="contained"
                                  size="small"
                                  color="success"
                                  onClick={() => {
                                    onClickRegainButton(item.id, one.id);
                                  }}
                                >
                                  Regain
                                </Button>
                              </TableCell>
                            </TableRow>
                          );
                        }
                      })}
                    </>
                  );
                } else {
                  return (
                    <TableRow key={item.id}>
                      <TableCell align="left">
                        <Typography>{item.id}</Typography>
                        <br />
                        <Button
                          variant="contained"
                          size="small"
                          color="warning"
                          onClick={() => {
                            makeBackup(item.id.toString(10));
                          }}
                        >
                          Backup
                        </Button>
                      </TableCell>
                      <TableCell align="right">{item.name}</TableCell>
                      <TableCell align="right">{item.updatedAtText}</TableCell>
                    </TableRow>
                  );
                }
              })}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog
        fullWidth
        maxWidth="sm"
        open={openCronSetting}
        onClose={handleCronSettingDialogClose}
      >
        <DialogContent>
          <Stack
            spacing={3}
            m={1}
            direction="row"
            alignItems="center"
            sx={{
              padding: '1rem',
            }}
          >
            <Typography variant="subtitle2" gutterBottom sx={{ width: '60%' }}>
              Status of Cron :
            </Typography>
            <FormControlLabel
              control={
                <IOSSwitch
                  sx={{ m: 1 }}
                  defaultChecked={cronStart}
                  onChange={onChangeStatusOfCron}
                />
              }
              label={cronStart ? 'Start' : 'Stop'}
            />
          </Stack>
          <Stack
            spacing={3}
            m={1}
            direction="row"
            alignItems="center"
            sx={{
              padding: '1rem',
            }}
          >
            <Typography variant="subtitle2" gutterBottom sx={{ width: '60%' }}>
              {cronStart ? ' Started at' : ' Stoped at'} :
            </Typography>
            <Typography variant="body1">{startedOrStopedAt}</Typography>
          </Stack>
          <Stack
            spacing={3}
            m={1}
            direction="row"
            alignItems="center"
            sx={{
              padding: '1rem',
            }}
          >
            <Typography variant="subtitle2" gutterBottom sx={{ width: '60%' }}>
              Next executed at :
            </Typography>
            <Typography variant="body1">
              {cronStart ? nextTime : startedOrStopedAt}
            </Typography>
          </Stack>
          <Stack
            spacing={3}
            m={1}
            direction="row"
            alignItems="center"
            sx={{
              padding: '1rem',
            }}
          >
            <Typography variant="subtitle2" gutterBottom sx={{ width: '60%' }}>
              Schedule Spec :
            </Typography>
            <Select
              onChange={onChangeScheduleSpec}
              sx={{
                width: '150px',
              }}
              defaultValue={scheduleSpec}
              disabled={!cronStart}
            >
              {scheduleSpecs.map((item) => (
                <MenuItem key={item} value={item}>
                  {item}
                </MenuItem>
              ))}
            </Select>
          </Stack>
          <Stack
            spacing={3}
            m={1}
            direction="row"
            alignItems="center"
            sx={{
              padding: '1rem',
            }}
          >
            <Typography variant="subtitle2" gutterBottom sx={{ width: '60%' }}>
              Last execution error :
            </Typography>
            <Typography variant="body1">{lastExecError}</Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            size="small"
            onClick={() => {
              setOpenCronSetting(false);
            }}
          >
            OK
          </Button>
        </DialogActions>
      </Dialog>

      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
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
      <Dialog fullScreen open={showLogs} onClose={handleClose}>
        <AppBar sx={{ position: 'relative' }}>
          <Toolbar>
            <IconButton
              edge="start"
              color="inherit"
              onClick={handleClose}
              aria-label="close"
            >
              <CloseIcon />
            </IconButton>
            <Typography sx={{ ml: 2, flex: 1 }} variant="h6" component="div">
              Logs
            </Typography>
          </Toolbar>
        </AppBar>
        <LinearProgress
          sx={{
            height: '5px',
          }}
          variant="determinate"
          value={progress}
        />
        <Stack spacing={3} direction="column" alignItems="center">
          <Typography variant="h6" sx={{ paddingTop: '16px' }} component="div">
            Word Count: {wordCount} ; Progress: {progress}
          </Typography>
          <List
            sx={{
              width: '90%',
              bgcolor: 'background.paper',
              overflowY: 'scroll',
              maxHeight: '600px',
              height: '600px',
              paddingRight: '10px',
            }}
          >
            {logs.map((log, index) => {
              return (
                <ListItem key={index}>
                  <ListItemText primary={log} />
                </ListItem>
              );
            })}
          </List>
        </Stack>
      </Dialog>
    </Stack>
  );
}
