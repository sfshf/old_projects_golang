'use client';
import * as React from 'react';
import Stack from '@mui/material/Stack';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import Button from '@mui/material/Button';
import { MuiFileInput } from 'mui-file-input';
import TextField from '@mui/material/TextField';
import { useSearchParams } from 'next/navigation';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar, { SnackbarOrigin } from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import shared from '@/app/shared';
import { TransitionProps } from '@mui/material/transitions';
import Dialog from '@mui/material/Dialog';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import CloseIcon from '@mui/icons-material/Close';
import LinearProgress from '@mui/material/LinearProgress';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Slide from '@mui/material/Slide';
import List from '@mui/material/List';

const Transition = React.forwardRef(function Transition(
  props: TransitionProps & {
    children: React.ReactElement;
  },
  ref: React.Ref<unknown>
) {
  return <Slide direction="up" ref={ref} {...props} />;
});

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

export default function BasicButtons() {
  const searchParams = useSearchParams();
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [file, setFile] = React.useState<File | null>(null);
  const [updateInfo, setUpdateInfo] = React.useState(false);
  const [strictMode, setStrictMode] = React.useState(false);
  const [bookID, setBookID] = React.useState(searchParams.get('bookID') || '');
  const [name, setName] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [showLogs, setShowLogs] = React.useState(false);
  const [logs, setLogs] = React.useState<string[]>([]);
  const [wordCount, setWordCount] = React.useState(0);
  const [progress, setProgress] = React.useState(0);

  const timerRef = React.useRef<NodeJS.Timer | null>(null);

  const handleChange = (newValue: File | null) => {
    setFile(newValue);
  };

  const buttonEnable =
    bookID.length >= 9 &&
    (updateInfo ? name.length > 1 && description.length > 1 : file !== null);

  async function upload() {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    const baseURL = shared.baseAPIURL;
    const formData = new FormData();
    formData.append('bookID', bookID);
    formData.append('password', password);
    if (updateInfo) {
      formData.append('name', name);
      formData.append('updateInfo', 'true');
      formData.append('description', description);
    } else {
      formData.append('file', file as File);
      formData.append('strictMode', strictMode ? 'true' : 'false');
      formData.append('updateInfo', 'false');
    }
    try {
      const res = await fetch(`${baseURL}/book/update`, {
        method: 'POST',
        body: formData,
      });
      const data = await res.json();
      setLoading(false);
      setFile(null);
      if (data.code !== 0) {
        // console.log(data.message);
        setErrorToast(data.message);
        return;
      }
      if (!updateInfo) {
        // show logs
        setShowLogs(true);
        setLogs([]);
        setWordCount(0);
        setProgress(0);
        timerRef.current = setInterval(() => {
          checkUploading();
        }, 200);
        // setTimeout(() => {
        //   if (timerRef.current !== null) {
        //     clearInterval(timerRef.current as NodeJS.Timer);
        //     timerRef.current = null;
        //   }
        // }, 100000);
      } else {
        setToast('Success: Book Info Updated');
      }
    } catch (error) {
      // console.log(error);
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  }

  async function checkUploading() {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    // setLoading(true);
    const baseURL = shared.baseAPIURL;
    try {
      const res = await fetch(
        `${baseURL}/book/logs?` + new URLSearchParams({ password }).toString()
      );
      const data = await res.json();
      // setLoading(false);
      if (data.code !== 0) {
        // console.log(data.message);

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

      // setToast("Success: Book added, Book ID: " + data.data.bookID);
    } catch (error) {
      // console.log(error);
      const message = (error as Error).message;
      setLogs((prev) => {
        return [message, ...prev];
      });
    }
  }

  const handleClose = () => {
    setShowLogs(false);
  };

  return (
    <Stack
      spacing={5}
      direction="column"
      alignItems="center"
      sx={{
        padding: '1rem',
      }}
    >
      <TextField
        required
        label="BookID"
        value={bookID}
        onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
          setBookID(event.target.value);
        }}
        variant="standard"
      />
      <FormControlLabel
        control={<Switch checked={updateInfo} />}
        label="Only Update Info"
        onChange={(e, checked) => setUpdateInfo(checked)}
      />
      {!updateInfo && (
        <FormControlLabel
          control={<Switch checked={strictMode} />}
          label="Strict Mode: "
          onChange={(e, checked) => setStrictMode(checked)}
        />
      )}

      {updateInfo ? (
        <FormGroup>
          <TextField
            required
            label="name"
            value={name}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
              setName(event.target.value);
            }}
            variant="standard"
          />
          <TextField
            required
            label="description"
            value={description}
            onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
              setDescription(event.target.value);
            }}
            variant="standard"
          />
        </FormGroup>
      ) : (
        <MuiFileInput
          label="Upload CSV File"
          value={file}
          onChange={handleChange}
        />
      )}

      {!updateInfo && (
        <Typography variant="h6" component="div">
          The system is updated; here, only support updating the book by adding
          new entries. If you want to edit or delete any entry, please use the
          Preview page.
        </Typography>
      )}
      <Button disabled={!buttonEnable} variant="contained" onClick={upload}>
        Update
      </Button>
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
      <Dialog
        fullScreen
        open={showLogs}
        onClose={handleClose}
        TransitionComponent={Transition}
      >
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
