'use client';
import * as React from 'react';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import { MuiFileInput } from 'mui-file-input';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import Typography from '@mui/material/Typography';
import { TransitionProps } from '@mui/material/transitions';
import Slide from '@mui/material/Slide';
import { post } from '@/app/util';

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

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [file, setFile] = React.useState<File | null>(null);
  const [fileHashName, setFileHashName] = React.useState('');

  const handleChange = (newValue: File | null) => {
    setFile(newValue);
  };

  const buttonEnable = file !== null;

  const upload = () => {
    post(
      false,
      '',
      '/upload',
      setLoading,
      false,
      file,
      (respData: any) => {
        setFile(null);
        if (respData.data) {
          setFileHashName(respData.data.hashName);
        }
      },
      setToast,
      setErrorToast
    );
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
      <MuiFileInput label="Upload File" value={file} onChange={handleChange} />
      <Button disabled={!buttonEnable} variant="contained" onClick={upload}>
        Upload
      </Button>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="h6" gutterBottom>
          File Hash Name:
        </Typography>
      </Stack>
      <Stack spacing={2} alignItems="center" direction="row">
        <Typography variant="subtitle1" gutterBottom>
          {fileHashName}
        </Typography>
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
    </Stack>
  );
}
