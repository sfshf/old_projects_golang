'use client';
import * as React from 'react';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';
import { post } from '@/app/util';
import shared from '@/app/shared';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Switch from '@mui/material/Switch';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import AddIcon from '@mui/icons-material/Add';
import Dialog from '@mui/material/Dialog';
import TextField from '@mui/material/TextField';
import CloseIcon from '@mui/icons-material/Close';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

interface Config {
  useTelegram: boolean;
  useWxpusher: boolean;
  wxpusherUIDs: string[];
  useEmail: boolean;
  emails: string[];
}

type ConfigAction = { type: 'setConfig'; value: Config };

const initConfig: Config = {
  useTelegram: false,
  useWxpusher: false,
  wxpusherUIDs: [],
  useEmail: false,
  emails: [],
};

const configReducer = (state: Config, action: ConfigAction) => {
  switch (action.type) {
    case 'setConfig':
      return action.value;
    default:
      return { ...state };
  }
};

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const [config, dispatchConfig] = React.useReducer(configReducer, initConfig);
  const getMessageNotificationConfig = () => {
    post(
      false,
      '',
      '/tester/getMessageNotificationConfig/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword() },
      (respData: any) => {
        if (respData.data) {
          dispatchConfig({ type: 'setConfig', value: respData.data });
        }
      },
      setToast,
      setErrorToast
    );
  };
  const [addMode, setAddMode] = React.useState<'uid' | 'email' | ''>('');
  const updateMessageNotificationConfig = (
    useTelegram: boolean,
    useWxpusher: boolean,
    wxpusherUIDs: string[],
    useEmail: boolean,
    emails: string[]
  ) => {
    post(
      false,
      '',
      '/tester/updateMessageNotificationConfig/v1',
      setLoading,
      true,
      {
        apiKey: shared.getPassword(),
        useTelegram,
        useWxpusher,
        wxpusherUIDs,
        useEmail,
        emails,
      },
      (respData) => {
        dispatchConfig({
          type: 'setConfig',
          value: {
            useTelegram,
            useWxpusher,
            wxpusherUIDs,
            useEmail,
            emails,
          },
        });
      },
      setToast,
      setErrorToast
    );
  };
  const sendMessageNotification = () => {
    post(
      false,
      '',
      '/tester/sendMessageNotification/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword() },
      undefined,
      setToast,
      setErrorToast
    );
  };
  const [input, setInput] = React.useState('');

  React.useEffect(() => {
    getMessageNotificationConfig();
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
        <TableContainer sx={{ width: '40%' }} component={Paper}>
          <Table aria-label="simple table">
            <TableBody>
              <TableRow
                key={'useTelegram'}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  <Typography variant="h6" gutterBottom>
                    Use Telegram:
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Switch
                    checked={config.useTelegram}
                    onChange={(e) => {
                      updateMessageNotificationConfig(
                        e.target.checked,
                        config.useWxpusher,
                        config.wxpusherUIDs,
                        config.useEmail,
                        config.emails
                      );
                    }}
                  />
                </TableCell>
              </TableRow>
              <TableRow
                key={'useWxpusher'}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  <Typography variant="h6" gutterBottom>
                    Use Wxpusher:
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Switch
                    checked={config.useWxpusher}
                    onChange={(e) => {
                      updateMessageNotificationConfig(
                        config.useTelegram,
                        e.target.checked,
                        config.wxpusherUIDs,
                        config.useEmail,
                        config.emails
                      );
                    }}
                  />
                </TableCell>
              </TableRow>
              <TableRow
                key={'wxpusherUIDs'}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  <Typography variant="h6" gutterBottom>
                    Wxpusher UIDs:
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <List>
                    {config.wxpusherUIDs.map((uid) => (
                      <ListItem
                        key={uid}
                        sx={{ borderBottom: '1px solid #979494' }}
                      >
                        <ListItemText primary={uid} />
                        <ListItemButton
                          sx={{ justifyContent: 'center' }}
                          onClick={() => {
                            let uids = config.wxpusherUIDs.filter(
                              (item) => item !== uid
                            );
                            updateMessageNotificationConfig(
                              config.useTelegram,
                              config.useWxpusher,
                              uids,
                              config.useEmail,
                              config.emails
                            );
                          }}
                        >
                          <ListItemIcon sx={{ justifyContent: 'center' }}>
                            <CloseIcon />
                          </ListItemIcon>
                        </ListItemButton>
                      </ListItem>
                    ))}
                    {config.wxpusherUIDs.length < 3 && (
                      <ListItem>
                        <ListItemButton
                          sx={{ justifyContent: 'center' }}
                          onClick={() => {
                            setAddMode('uid');
                          }}
                        >
                          <ListItemIcon sx={{ justifyContent: 'center' }}>
                            <AddIcon />
                          </ListItemIcon>
                        </ListItemButton>
                      </ListItem>
                    )}
                  </List>
                </TableCell>
              </TableRow>
              <TableRow
                key={'useEmail'}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  <Typography variant="h6" gutterBottom>
                    Use Email:
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Switch
                    checked={config.useEmail}
                    onChange={(e) => {
                      updateMessageNotificationConfig(
                        config.useTelegram,
                        config.useWxpusher,
                        config.wxpusherUIDs,
                        e.target.checked,
                        config.emails
                      );
                    }}
                  />
                </TableCell>
              </TableRow>
              <TableRow
                key={'emails'}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  <Typography variant="h6" gutterBottom>
                    Emails:
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <List>
                    {config.emails.map((email) => (
                      <ListItem
                        key={email}
                        sx={{ borderBottom: '1px solid #979494' }}
                      >
                        <ListItemText primary={email} />
                        <ListItemButton
                          sx={{ justifyContent: 'center' }}
                          onClick={() => {
                            let emails = config.emails.filter(
                              (item) => item !== email
                            );
                            updateMessageNotificationConfig(
                              config.useTelegram,
                              config.useWxpusher,
                              config.wxpusherUIDs,
                              config.useEmail,
                              emails
                            );
                          }}
                        >
                          <ListItemIcon sx={{ justifyContent: 'center' }}>
                            <CloseIcon />
                          </ListItemIcon>
                        </ListItemButton>
                      </ListItem>
                    ))}
                    {config.emails.length < 2 && (
                      <ListItem>
                        <ListItemButton
                          sx={{ justifyContent: 'center' }}
                          onClick={() => {
                            setAddMode('email');
                          }}
                        >
                          <ListItemIcon sx={{ justifyContent: 'center' }}>
                            <AddIcon />
                          </ListItemIcon>
                        </ListItemButton>
                      </ListItem>
                    )}
                  </List>
                </TableCell>
              </TableRow>
              <TableRow
                key={'useTelegram'}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
              >
                <TableCell component="th" scope="row">
                  <Typography variant="h6" gutterBottom>
                    Send Test Message:
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Button
                    variant="contained"
                    onClick={() => {
                      sendMessageNotification();
                    }}
                  >
                    Test
                  </Button>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
        <Dialog
          onClose={() => {
            setAddMode('');
          }}
          open={addMode !== ''}
        >
          <Stack
            spacing={1}
            sx={{
              padding: '1rem',
              width: '500px',
              height: '150px',
            }}
            direction="column"
            alignItems="center"
            justifyContent="center"
          >
            <TextField
              onChange={(e) => {
                setInput(e.target.value);
              }}
              id="outlined-basic"
              label={addMode === 'uid' ? 'Wxpusher UID' : 'Email'}
              variant="outlined"
              sx={{ width: '100%' }}
            />
            <Stack spacing={1} direction="row">
              <Button
                variant="contained"
                onClick={() => {
                  setAddMode('');
                  if (addMode === 'uid') {
                    let uids = [...config.wxpusherUIDs];
                    uids.push(input);
                    updateMessageNotificationConfig(
                      config.useTelegram,
                      config.useWxpusher,
                      uids,
                      config.useEmail,
                      config.emails
                    );
                  } else if (addMode === 'email') {
                    let emails = [...config.emails];
                    emails.push(input);
                    updateMessageNotificationConfig(
                      config.useTelegram,
                      config.useWxpusher,
                      config.wxpusherUIDs,
                      config.useEmail,
                      emails
                    );
                  }
                }}
              >
                Commit
              </Button>
              <Button
                color="error"
                variant="contained"
                onClick={() => {
                  setAddMode('');
                }}
              >
                Cancel
              </Button>
            </Stack>
          </Stack>
        </Dialog>
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
