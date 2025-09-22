'use client';
import '@/app/globals.css';
import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import React from 'react';
import Link from 'next/link';
import { StyledEngineProvider } from '@mui/material/styles';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import useMediaQuery from '@mui/material/useMediaQuery';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import shared from '@/app/shared';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';

const LINKS = [
  { text: 'Home', href: '/' },
  { text: 'API', href: '/api' },
  { text: 'Mysql', href: '/mysql' },
  { text: 'Mongo', href: '/mongo' },
  { text: 'BTC Diff', href: '/btcdiff' },
  { text: 'Mempool CronStat', href: '/mempool-cronstat' },
  { text: 'Upload', href: '/upload' },
  { text: 'Slark Registration Captcha', href: '/slark-registration-captcha' },
  { text: 'Slark Login Captcha', href: '/slark-login-captcha' },
  { text: 'Upload App', href: '/upload-app' },
  { text: 'Notification Config', href: '/notification' },
  { text: 'Notification Log', href: '/notification-log' },
  { text: 'Privacy Email Account', href: '/privacy-email-account' },
];

function Container({ children }: { children: React.ReactNode }) {
  // console.log(router);
  const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)');
  const theme = React.useMemo(
    () =>
      createTheme({
        palette: {
          mode: prefersDarkMode ? 'dark' : 'light',
        },
      }),
    [prefersDarkMode]
  );
  const pathName = usePathname();

  const [value, setValue] = React.useState(pathName);

  // Save pathname on component mount into a REF
  const savedPathNameRef = React.useRef(pathName);
  const router = useRouter();
  const params = useSearchParams();
  React.useEffect(() => {
    // If REF has been changed, do the stuff
    if (savedPathNameRef.current !== pathName) {
      setValue(pathName);
      // Update REF
      savedPathNameRef.current = pathName;
    }
  }, [pathName, setValue]);

  const handleChange = (event: React.SyntheticEvent, newValue: string) => {
    setValue(newValue);
  };

  const [password, setPassword] = React.useState(shared.getPassword());

  const onPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const p = e.target.value;
    setPassword(p);
    shared.setPassword(p);
    setIsAdmin(false);
  };

  const [loading, setLoading] = React.useState(false);
  const [errorToast, setErrorToast] = React.useState('');

  const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
    props,
    ref
  ) {
    return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
  });

  const [isAdmin, setIsAdmin] = React.useState(false);

  const onChangeIsAdmin = async (event: React.BaseSyntheticEvent) => {
    if (event.target.checked) {
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
          `${shared.baseAPIURL}/system/check_admin?` + param.toString()
        );
        setLoading(false);
        const data = await res.json();
        if (data.code !== 0) {
          setErrorToast('只有管理员可以执⾏该操作');
          return;
        }
        setIsAdmin(data.data === true ? true : false);
      } catch (error) {
        const message = (error as Error).message;
        setErrorToast(message);
      }
    } else {
      setIsAdmin(event.target.checked);
    }
  };

  const matches = useMediaQuery('(min-width:600px)');

  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <div className="flex flex-row items-center">
          <Typography variant="h6" gutterBottom className="p-5">
            Password:
          </Typography>
          <TextField
            id="password-input"
            label="Password:"
            variant="filled"
            value={password}
            onChange={onPasswordChange}
          />
        </div>

        <Tabs
          value={value}
          onChange={handleChange}
          centered
          variant={'scrollable'}
        >
          {LINKS.map(({ text, href }) => {
            if (
              text === 'OperateLog' ||
              text === 'WorkingEthics' ||
              text === 'Backup'
            ) {
              if (isAdmin) {
                return (
                  <Tab
                    label={text}
                    key={href}
                    value={href}
                    href={href}
                    LinkComponent={Link}
                  />
                );
              }
            } else {
              return (
                <Tab
                  label={text}
                  key={href}
                  value={href}
                  href={href}
                  LinkComponent={Link}
                />
              );
            }
          })}
        </Tabs>

        {children}
      </ThemeProvider>
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
    </StyledEngineProvider>
  );
}
// export default withRouter(Container);
export default Container;
