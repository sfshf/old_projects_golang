'use client';
import React, { useEffect } from 'react';
import Stack from '@mui/material/Stack';
import { post } from '@/app/shared';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import Link from '@mui/material/Link';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import CustomHead from '@/app/site/CustomHead';
import { getLocalUserLoginState, LoginState } from '@/app/site/shared';

interface Site {
  id: number;
  name: string;
}

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [errorToast, setErrorToast] = React.useState('');

  const [sites, setSites] = React.useState<Site[] | null>(null);

  const getSites = () => {
    post(
      false,
      '/invoker/site/getSites/v1',
      setLoading,
      null,
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data && respData.data.list.length > 0) {
          setSites(respData.data.list);
        }
      },
      undefined,
      setErrorToast
    );
  };
  const [loginInfo, setLoginInfo] = React.useState<null | LoginState>(null);

  useEffect(() => {
    // fetch login state
    let loginState = getLocalUserLoginState();
    if (loginState) {
      setLoginInfo(loginState);
    } else {
      setLoginInfo(null);
    }
    getSites();
  }, []);

  return (
    <>
      <CustomHead
        headText="Invoker"
        loginInfo={loginInfo}
        siteID={null}
        setLoading={setLoading}
        setErrorToast={setErrorToast}
      />
      <Stack
        spacing={3}
        direction="row"
        sx={{
          padding: '1rem',
          marginLeft: '5%',
          marginRight: '5%',
        }}
      >
        <Typography variant="h5">Site List:</Typography>
      </Stack>
      <Stack
        spacing={3}
        direction="row"
        sx={{
          padding: '1rem',
          marginLeft: '5%',
          marginRight: '5%',
        }}
      >
        <Box sx={{ width: '100%' }}>
          <List>
            {sites &&
              sites.map((row, idx) => {
                return (
                  <ListItem key={row.id} disablePadding sx={{ height: '80px' }}>
                    <ListItemButton>
                      <Link href={'/site/' + row.name} target="_self">
                        <ListItemText primary={idx + 1 + '. ' + row.name} />
                      </Link>
                    </ListItemButton>
                  </ListItem>
                );
              })}
          </List>
        </Box>
      </Stack>
      <CustomSnackbar errorToast={errorToast} setErrorToast={setErrorToast} />
      <CustomBackdrop loading={loading} />
    </>
  );
}
