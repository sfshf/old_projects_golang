'use client';
import React, { useEffect } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import moment from 'moment';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import { post } from '@/app/util';
import shared from '@/app/shared';

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
  const handleRequestBody = (obj: any): any => {
    if (obj == null) {
      return '';
    }
    for (let key in obj) {
      if (obj.hasOwnProperty(key)) {
        if (typeof obj[key] === 'string' && obj[key].includes('TIMESTAMP=')) {
          let valStr = obj[key].substring(obj[key].indexOf('=') + 1);
          let ts = moment(); // n
          if (valStr.includes('+')) {
            let deltaStr = valStr.substring(valStr.indexOf('+') + 1);
            if (deltaStr.includes('s')) {
              obj[key] = ts
                .add(
                  parseInt(deltaStr.substring(0, deltaStr.indexOf('s'))),
                  's'
                )
                .valueOf();
            } else if (deltaStr.includes('h')) {
              obj[key] = ts
                .add(
                  parseInt(deltaStr.substring(0, deltaStr.indexOf('h'))),
                  'h'
                )
                .valueOf();
            } else if (deltaStr.includes('d')) {
              obj[key] = ts
                .add(
                  parseInt(deltaStr.substring(0, deltaStr.indexOf('d'))),
                  'd'
                )
                .valueOf();
            }
          } else if (obj[key].includes('-')) {
            let deltaStr = valStr.substring(valStr.indexOf('-') + 1);
            if (deltaStr.includes('s')) {
              obj[key] = ts
                .subtract(
                  parseInt(deltaStr.substring(0, deltaStr.indexOf('s'))),
                  's'
                )
                .valueOf();
            } else if (deltaStr.includes('h')) {
              obj[key] = ts
                .subtract(
                  parseInt(deltaStr.substring(0, deltaStr.indexOf('h'))),
                  'h'
                )
                .valueOf();
            } else if (deltaStr.includes('d')) {
              obj[key] = ts
                .subtract(
                  parseInt(deltaStr.substring(0, deltaStr.indexOf('d'))),
                  'd'
                )
                .valueOf();
            }
          } else {
            obj[key] = ts.valueOf();
          }
        }
      }
    }
    return JSON.stringify(obj, null, 4);
  };
  const [app, setApp] = React.useState('');
  const [appList, setAppList] = React.useState<string[]>([]);
  const getApps = () => {
    post(
      false,
      '',
      '/tester/getApps/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword() },
      (respData: any) => {
        if (respData.data) {
          setAppList(respData.data.list);
        }
      },
      setToast,
      setErrorToast
    );
  };
  const [apiList, setApiList] = React.useState<any[]>([]);
  const getAPITestcases = (app: string) => {
    post(
      false,
      '',
      '/tester/getAPITestcases/v1',
      setLoading,
      true,
      { apiKey: shared.getPassword(), app },
      (respData: any) => {
        if (respData.data) {
          setApiList(respData.data.list);
        } else {
          setApiList([]);
          setReqBody('');
          setRespBody('');
        }
      },
      setToast,
      setErrorToast
    );
  };
  const [api, setApi] = React.useState('');
  const [path, setPath] = React.useState('');
  const [sessionIDList, setSessionIDList] = React.useState<any>([
    { key: 'Account 1', value: 'CBDB7E012057495DAA44A9DDE989BFE5' },
    { key: 'Account 2', value: 'DCAC40C131C547F7A503855593EBB2D3' },
    { key: 'Account 3', value: 'FF2F4033094947A7B6A6AA1BFC613A8C' },
    { key: 'None', value: 'None' },
  ]);
  const [sessionID, setSessionID] = React.useState(
    'CBDB7E012057495DAA44A9DDE989BFE5'
  );
  const [reqBody, setReqBody] = React.useState('');
  const [respBody, setRespBody] = React.useState('');
  const [useGo, setUseGo] = React.useState(true);

  const doPost = () => {
    post(
      useGo,
      sessionID,
      path,
      setLoading,
      true,
      reqBody ? JSON.parse(handleRequestBody(eval('(' + reqBody + ')'))) : null,
      (respData) => {
        setRespBody(JSON.stringify(respData, null, 4));
      },
      setToast,
      setErrorToast
    );
  };

  useEffect(() => {
    getApps();
  }, []);

  return (
    <main>
      <Stack marginX="200px" marginTop="50px" spacing={2}>
        <Stack direction="row" textAlign="center" justifyContent="left">
          <Typography align="left" variant="h6" gutterBottom width="40%">
            Step 1: Select App
          </Typography>
          <FormControl fullWidth>
            <Select
              value={app}
              onChange={(e) => {
                setApp(e.target.value);
                getAPITestcases(e.target.value);
              }}
            >
              {appList &&
                appList.map((item: any) => (
                  <MenuItem value={item} key={item}>
                    {item}
                  </MenuItem>
                ))}
            </Select>
          </FormControl>
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="left">
          <Typography align="left" variant="h6" gutterBottom width="40%">
            Step 2: Select Api
          </Typography>
          <FormControl fullWidth>
            <Select
              labelId="demo-simple-select-label"
              id="demo-simple-select"
              value={api}
              onChange={(e) => {
                let apiName = e.target.value;
                setApi(apiName);
                for (let i = 0; i < apiList.length; i++) {
                  if (apiList[i].name == apiName) {
                    setPath(apiList[i].path);
                    if (apiList[i].body) {
                      let bodyString = apiList[i].body
                        .replace(/\\/g, '')
                        .replace(/\s+/g, '');
                      setReqBody(
                        JSON.stringify(eval('(' + bodyString + ')'), null, 4)
                      );
                    } else {
                      setReqBody('');
                    }
                    break;
                  }
                }
                setRespBody('');
              }}
            >
              {apiList &&
                apiList.map((item: any) => (
                  <MenuItem value={item.name} key={item.path}>
                    {item.name + '::' + item.path}
                  </MenuItem>
                ))}
            </Select>
          </FormControl>
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="left">
          <Typography align="left" variant="h6" gutterBottom width="40%">
            Step 3: Select Account
          </Typography>
          <FormControl fullWidth>
            <Select
              labelId="demo-simple-select-label"
              id="demo-simple-select"
              value={sessionID}
              onChange={(e) => {
                let val = e.target.value;
                if (val === 'None' || val === '') {
                  setSessionID('');
                } else {
                  setSessionID(val);
                }
              }}
            >
              {sessionIDList &&
                sessionIDList.map((item: any) => (
                  <MenuItem value={item.value} key={item.key}>
                    {item.key}
                  </MenuItem>
                ))}
            </Select>
          </FormControl>
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="left">
          <Typography align="left" variant="h6" gutterBottom width="40%">
            Step 4: Request Body
          </Typography>
          <FormControlLabel
            control={
              <Checkbox
                checked={useGo}
                onChange={(e) => {
                  setUseGo(e.target.checked);
                }}
              />
            }
            label="Send encrypted request"
          />
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="left">
          <TextField
            id="outlined-basic"
            variant="outlined"
            fullWidth
            value={reqBody}
            onChange={(e) => {
              setReqBody(e.target.value);
            }}
            multiline
            rows={8}
          />
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="left">
          <Typography align="left" variant="h6" gutterBottom width="40%">
            Step 5: Post Request
          </Typography>
          <Button
            variant="contained"
            size="medium"
            fullWidth
            onClick={() => {
              doPost();
            }}
          >
            Post
          </Button>
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="left">
          <Typography align="left" variant="h6" gutterBottom width="40%">
            Step 6: Results
          </Typography>
          {/* <TextField
            id="outlined-basic"
            variant="outlined"
            value={tsNum}
            onChange={(e) => {
              setTsNum(parseInt(e.target.value));
            }}
            sx={{
              width: '20%',
            }}
          />
          <Button
            variant="text"
            size="medium"
            onClick={() => {
              doTransfer();
            }}
          >
            {'==>'}
          </Button>
          <TextField
            id="outlined-basic"
            variant="outlined"
            value={tsStr}
            disabled
            sx={{
              width: '30%',
            }}
          /> */}
        </Stack>
        <TextField
          style={{ flex: 1 }}
          id="outlined-basic"
          variant="outlined"
          fullWidth
          multiline
          minRows={15}
          value={respBody}
          disabled
        />
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
