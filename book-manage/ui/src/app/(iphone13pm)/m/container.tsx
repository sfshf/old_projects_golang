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
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Checkbox from '@mui/material/Checkbox';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Stack from '@mui/material/Stack';

const LINKS = [
  { text: 'Home', href: '/m' },
  { text: 'Update', href: '/m/update' },
  { text: 'Add', href: '/m/add' },
  { text: 'New Definition', href: '/m/new-definition' },
  { text: 'Definition List', href: '/m/list-definition' },
  { text: 'Preview', href: '/m/preview' },
  { text: 'Search', href: '/m/search' },
  { text: 'OperateLog', href: '/m/operate-log' },
  { text: 'WorkingEthics', href: '/m/working-ethics' },
  { text: 'Backup', href: '/m/backup' },
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
    if (window.innerWidth > 600) {
      if (pathName != '/m') {
        router.push(pathName.replace('/m', '') + '?' + params, {
          scroll: false,
        });
      }
    }
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

  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Stack
          spacing={2}
          width="100%"
          alignItems="center"
          sx={{
            paddingTop: '5px',
          }}
          direction="row"
        >
          <Typography
            width="35%"
            variant="subtitle2"
            gutterBottom
            className="p-5"
          >
            Password:
          </Typography>
          <TextField
            id="password-input"
            label="Password:"
            variant="filled"
            value={password}
            onChange={onPasswordChange}
          />
        </Stack>
        <Stack spacing={2} width="100%" direction="row" alignItems="center">
          <Typography
            width="35%"
            variant="subtitle2"
            gutterBottom
            className="p-5"
          >
            Authority Check:
          </Typography>
          <FormGroup>
            <FormControlLabel
              control={
                <Checkbox
                  defaultChecked={false}
                  checked={isAdmin}
                  onChange={onChangeIsAdmin}
                />
              }
              label="Is Admin"
            />
          </FormGroup>
        </Stack>
        <Stack spacing={2} width="100%" direction="row" alignItems="center">
          <Tabs
            value={value}
            onChange={handleChange}
            centered
            variant="scrollable"
            scrollButtons="auto"
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
        </Stack>
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
