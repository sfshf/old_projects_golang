'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { useEffect } from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from './shared';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { request } from 'http';
import { useSearchParams } from 'next/navigation';
import IconButton from '@mui/material/IconButton';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import InputAdornment from '@mui/material/InputAdornment';
import FormControl from '@mui/material/FormControl';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import { keccak_256 } from '@noble/hashes/sha3';
import cookie from 'react-cookies';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

const baseURL = process.env.NEXT_PUBLIC_SERVER_BASE_URL;

interface ValidateSession {
  sso?: string;
  sig?: string;
}

interface SessionUser {
  avatar: string;
  username: string;
  email: string;
}

interface SignInRequest {
  email?: string;
  passwordHash?: string;
  sso: string;
  sig: string;
}

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [showPassword, setShowPassword] = React.useState(false);

  const handleClickShowPassword = () => setShowPassword((show) => !show);

  const handleMouseDownPassword = (
    event: React.MouseEvent<HTMLButtonElement>
  ) => {
    event.preventDefault();
  };

  const [email, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [sessionID, setSessionID] = React.useState('');
  const [sessionUser, setSessionUser] = React.useState<SessionUser | null>(
    null
  );

  const DefaultCookieSessionKey = 'LSessionID';

  const searchParams = useSearchParams();

  const detectSession = async () => {
    let sessionID = cookie.load(DefaultCookieSessionKey);
    let sso = searchParams.get('sso') as string;
    let sig = searchParams.get('sig') as string;

    if (!sessionID) {
      return;
    }

    setLoading(true);

    const reqData: ValidateSession = {
      sso: sso,
      sig: sig,
    };

    try {
      const postData = JSON.stringify(reqData);
      const req = request(
        {
          path: shared.baseAPIURL + '/session/validate/v1',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData),
            Cookie: DefaultCookieSessionKey + '=' + sessionID,
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
            setSessionID(sessionID);
            if (sso && sig) {
              setSessionUser(respData.data);
            } else {
              // redirect to the other page due to it's not sso
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

  const signIn = async () => {
    let sso = searchParams.get('sso');
    let sig = searchParams.get('sig');
    sso = sso ? sso : '';
    sig = sig ? sig : '';
    let sessionID = cookie.load(DefaultCookieSessionKey);

    setLoading(true);

    const reqData: SignInRequest = {
      sso: sso,
      sig: sig,
      email: email,
      passwordHash: Buffer.from(keccak_256(password)).toString('hex'),
    };

    try {
      const postData = JSON.stringify(reqData);
      let headers;
      if (sessionID) {
        headers = {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
          Cookie: DefaultCookieSessionKey + '=' + sessionID,
        };
      } else {
        headers = {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
        };
      }
      const req = request(
        {
          path: shared.baseAPIURL + '/signIn/v1',
          method: 'POST',
          headers: headers,
        },
        (res) => {
          res.setEncoding('utf8');
          res.on('data', (chunk) => {
            const respData = JSON.parse(chunk);
            if (respData.code !== 0) {
              setErrorToast(respData.message);
              return;
            }
            setToast('success');
            if (sso && sig) {
              window.location.replace(respData.data);
            } else {
              // redirect to the other page
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

  useEffect(() => {
    detectSession();
  });

  return (
    <main>
      {!sessionID && (
        <Stack marginTop="150px" spacing={3}>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <Typography variant="h1" gutterBottom>
              WELCOME
            </Typography>
          </Stack>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <TextField
              id="outlined-basic"
              label="Email"
              variant="outlined"
              sx={{ width: '500px' }}
              value={email}
              onChange={(e) => {
                setEmail(e.target.value);
              }}
            />
          </Stack>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <FormControl sx={{ m: 1, width: '500px' }} variant="outlined">
              <InputLabel htmlFor="outlined-adornment-password">
                Password
              </InputLabel>
              <OutlinedInput
                id="outlined-adornment-password"
                type={showPassword ? 'text' : 'password'}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={handleClickShowPassword}
                      onMouseDown={handleMouseDownPassword}
                      edge="end"
                    >
                      {showPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                }
                label="Password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                }}
              />
            </FormControl>
          </Stack>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <Button
              variant="contained"
              sx={{ width: '500px' }}
              onClick={() => {
                signIn();
              }}
            >
              Sign in
            </Button>
          </Stack>
        </Stack>
      )}
      {sessionID && sessionUser && (
        <Stack marginTop="150px" spacing={3}>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <Typography variant="h5" gutterBottom>
              Your slark account {sessionUser.username} has been detected.
            </Typography>
          </Stack>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <Typography variant="h6" gutterBottom>
              Do you want to continue signing in Discourse with{' '}
              {sessionUser.email}?
            </Typography>
          </Stack>
          <Stack direction="row" textAlign="center" justifyContent="center">
            <Button
              variant="contained"
              sx={{ width: '500px' }}
              onClick={() => {
                signIn();
              }}
            >
              Sign in
            </Button>
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
