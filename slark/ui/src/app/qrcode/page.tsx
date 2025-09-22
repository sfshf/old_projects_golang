'use client';
import QRCode from 'qrcode';
import React, { useCallback, useEffect } from 'react';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import shared from '../shared';
import { request } from 'http';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

interface SessionUser {
  avatar: string;
  username: string;
  email: string;
  sessionID: string;
}

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [sessionUser, setSessionUser] = React.useState<SessionUser | null>(
    null
  );
  const timerRef = React.useRef<NodeJS.Timer | null>(null);

  const checkLogin = async (token: string) => {
    if (sessionUser) {
      clearInterval(timerRef.current as NodeJS.Timer);
      timerRef.current = null;
      return;
    }
    // setLoading(true);
    try {
      const postData = JSON.stringify({ token: token });
      const req = request(
        {
          path: shared.baseAPIURL + '/qrcode/login/check/v1',
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
              clearInterval(timerRef.current as NodeJS.Timer);
              timerRef.current = null;
              setErrorToast(respData.message);
              return;
            } else {
              if (respData.message === 'success') {
                clearInterval(timerRef.current as NodeJS.Timer);
                timerRef.current = null;
                setSessionUser(respData.data);
              }
              return;
            }
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
      setToast(message);
    }
  };

  const getToken = async () => {
    setLoading(true);

    try {
      const req = request(
        {
          path: shared.baseAPIURL + '/qrcode/token/v1',
          method: 'POST',
        },
        (res) => {
          res.setEncoding('utf8');
          res.on('data', (chunk) => {
            const respData = JSON.parse(chunk);
            if (respData.code !== 0) {
              setErrorToast(respData.message);
              return;
            }

            // generate qr code
            let canvas = document.getElementById('qrcode');
            let loginUrl = 'n1xt://slark/qrLogin?token=' + respData.data;
            QRCode.toCanvas(canvas, loginUrl, { width: 300 }, (err) => {
              if (err) console.error(err);
            });

            timerRef.current = setInterval(() => {
              checkLogin(respData.data);
            }, 5000);
          });
        }
      );
      req.on('error', (e) => {
        throw e;
      });
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setToast(message);
    }
  };

  useEffect(() => {
    getToken();
  }, []);

  return (
    <main>
      <Stack marginTop="150px" spacing={3}>
        <Stack direction="row" textAlign="center" justifyContent="center">
          <Typography variant="h3" gutterBottom>
            Please scan the QR code, and log in !
          </Typography>
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="center">
          <canvas id="qrcode"></canvas>
        </Stack>
      </Stack>
      {sessionUser && (
        <Stack marginTop="150px" spacing={3}>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <Typography variant="h5" gutterBottom>
              Hello {sessionUser.username}, you have logged in.
            </Typography>
          </Stack>
        </Stack>
      )}

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
        autoHideDuration={4500}
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
