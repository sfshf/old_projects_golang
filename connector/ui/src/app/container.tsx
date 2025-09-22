'use client';
import './globals.css';
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
import { usePathname, useRouter } from 'next/navigation';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import shared from './shared';

const LINKS = [
  { text: 'ApiKey', href: '/' },
  { text: 'Config', href: '/config' },
  { text: 'Password', href: '/password' },
  { text: 'PrivateKey', href: '/privatekey' },
  { text: 'Keystore Log', href: '/keystore/log' },
  { text: 'Keystore Monitor', href: '/keystore/monitor' },
  { text: 'Connector Manage Log', href: '/log' },
];

function Container({ children }: { children: React.ReactNode }) {
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
  const handleChange = (event: React.SyntheticEvent, newValue: string) => {
    setValue(newValue);
  };
  const [password, setPassword] = React.useState(shared.getPassword());
  const onPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const p = e.target.value;
    setPassword(p);
    shared.setPassword(p);
  };
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
    </StyledEngineProvider>
  );
}

export default Container;
