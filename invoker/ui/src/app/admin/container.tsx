'use client';
import React from 'react';
import Link from 'next/link';
import { StyledEngineProvider } from '@mui/material/styles';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import useMediaQuery from '@mui/material/useMediaQuery';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { usePathname } from 'next/navigation';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { getApiKey, setApiKey } from '@/app/shared';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';

const LINKS = [
  { text: 'Site', href: '/admin' }, // default page
  { text: 'Mysql', href: '/admin/mysql' },
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
  React.useEffect(() => {
    // If REF has been changed, do the stuff
    if (savedPathNameRef.current !== pathName) {
      setValue(pathName);
      // console.log(pathName);
      // Update REF
      savedPathNameRef.current = pathName;
    }
  }, []);

  const handleChange = (event: React.SyntheticEvent, newValue: string) => {
    setValue(newValue);
  };

  const [password, setPassword] = React.useState(getApiKey());

  const onPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const p = e.target.value;
    setPassword(p);
    setApiKey(p);
  };

  const [loading, setLoading] = React.useState(false);
  const [errorToast, setErrorToast] = React.useState('');

  const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
    props,
    ref
  ) {
    return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
  });

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

        <Tabs value={value} onChange={handleChange} centered>
          {LINKS.map(({ text, href }) => {
            return (
              <Tab
                label={text}
                key={href}
                value={href}
                href={href}
                LinkComponent={Link}
              />
            );
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
