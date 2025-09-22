'use client';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { request } from 'http';
import shared from '../../../shared';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';
import keccak256 from 'keccak256';
import { XChaCha20Poly1305 } from '@stablelib/xchacha20poly1305';
import moment from 'moment';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

export default function Page({
  params,
}: {
  params: { referral_code: string };
}) {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [valid, setValid] = React.useState(false);

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

  const checkReferralCode = async () => {
    const reqData = {
      referralCode: params.referral_code,
    };
    post('/alchemist/checkReferralCode/v1', reqData, (respData) => {
      setValid(respData.data.valid);
    });
  };

  useEffect(() => {
    checkReferralCode();
  }, []);

  return (
    <main>
      <Stack marginTop="150px" spacing={3}>
        {valid && (
          <>
            <Stack direction="row" textAlign="center" justifyContent="center">
              <Typography variant="h3" gutterBottom>
                Referral Code:
              </Typography>
            </Stack>
            <Stack direction="row" textAlign="center" justifyContent="center">
              <Typography variant="h3" gutterBottom>
                {params.referral_code}
              </Typography>
            </Stack>
          </>
        )}
        {!valid && (
          <Stack direction="row" textAlign="center" justifyContent="center">
            <Typography variant="h3" gutterBottom>
              Not valid Referral code
            </Typography>
          </Stack>
        )}
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
