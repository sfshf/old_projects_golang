'use client';
import React from 'react';
import Link from 'next/link';
import { StyledEngineProvider } from '@mui/material/styles';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import useMediaQuery from '@mui/material/useMediaQuery';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { useParams, usePathname } from 'next/navigation';
import CustomHead from '@/app/site/CustomHead';
import { LoginState, getLocalUserLoginState } from '@/app/site/shared';
import { Site } from '@/app/model';
import { SiteIDContext } from '@/app/site/[site]/admin/context';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import { post } from '@/app/shared';

function Container({ children }: { children: React.ReactNode }) {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const params = useParams();
  const LINKS = [
    {
      text: 'Category',
      href: '/site/' + params.site + '/admin/category',
    },
  ];

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

  const [siteID, setSiteID] = React.useState(0);
  const [siteInfo, setSiteInfo] = React.useState<null | Site>(null);
  const [loginInfo, setLoginInfo] = React.useState<null | LoginState>(null);

  const getSite = (site: string) => {
    if (!site) {
      return;
    }
    post(
      false,
      '/invoker/site/getSite/v1',
      setLoading,
      { name: site },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setSiteInfo(respData.data);
          setSiteID(respData.data.id);
        }
      },
      undefined,
      setErrorToast
    );
  };

  React.useEffect(() => {
    setLoginInfo(getLocalUserLoginState());
    getSite(params.site as string);
    // If REF has been changed, do the stuff
    if (savedPathNameRef.current !== pathName) {
      setValue(pathName);
      // Update REF
      savedPathNameRef.current = pathName;
    }
  }, []);

  const handleChange = (event: React.SyntheticEvent, newValue: string) => {
    setValue(newValue);
  };

  return (
    <>
      <CustomHead
        headText={siteInfo ? 'Site ' + siteInfo.name : 'Site'}
        loginInfo={loginInfo}
        siteID={siteID}
        setLoading={setLoading}
        setErrorToast={setErrorToast}
      />
      <SiteIDContext.Provider value={siteID}>
        <StyledEngineProvider injectFirst>
          <ThemeProvider theme={theme}>
            <CssBaseline />
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
      </SiteIDContext.Provider>
      <CustomSnackbar
        toast={toast}
        setToast={setToast}
        errorToast={errorToast}
        setErrorToast={setErrorToast}
      />
      <CustomBackdrop loading={loading} />
    </>
  );
}
// export default withRouter(Container);
export default Container;
