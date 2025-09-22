'use client';
import React from 'react';
import { post } from '@/app/shared';
import { useRouter, usePathname, useParams } from 'next/navigation';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import moment from 'moment';
import Stack from '@mui/material/Stack';
import Pagination from '@mui/material/Pagination';
import CustomHead from '@/app/site/CustomHead';
import {
  setLocalUserLoginState,
  LoginState,
  removeLocalUserLoginState,
} from '@/app/site/shared';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import Container from '@mui/material/Container';
import Divider from '@mui/material/Divider';
import CustomAvatar from '@/app/site/CustomAvatar';
import { Site, Thumbup } from '@/app/model';
import Link from '@mui/material/Link';

export default function Page() {
  const router = useRouter();
  const pathname = usePathname();
  const params = useParams();

  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [pageNumber, setPageNumber] = React.useState(1); // ui from 1, but api from 0;
  const [pageSize, setPageSize] = React.useState(10);

  const [loginInfo, setLoginInfo] = React.useState<null | LoginState>(null);
  const [siteID, setSiteID] = React.useState<number>(0);
  const [siteInfo, setSiteInfo] = React.useState<null | Site>(null);
  const [isSiteAdmin, setIsSiteAdmin] = React.useState(false);
  const [thumbups, setThumbups] = React.useState<Thumbup[] | null>(null);
  const [thumbupTotal, setThumbupTotal] = React.useState(0);
  const thumbupHistory = (
    site: string,
    pageNumber: number,
    pageSize: number
  ) => {
    post(
      false,
      '/invoker/user/thumbupHistory/v1',
      setLoading,
      {
        site,
        pageNumber,
        pageSize,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.code != 0) {
          setErrorToast('fail to get site info of ' + site);
          return;
        }
        if (respData.data) {
          if (respData.data.loginInfo) {
            setLoginInfo(respData.data.loginInfo);
            setLocalUserLoginState(respData.data.loginInfo);
          } else {
            setLoginInfo(null);
            removeLocalUserLoginState();
          }
          if (respData.data.siteInfo) {
            setSiteInfo(respData.data.siteInfo);
            setSiteID(respData.data.siteInfo.id);
          } else {
            setSiteInfo(null);
            setSiteID(0);
          }
          if (
            respData.data.loginInfo &&
            respData.data.siteInfo &&
            respData.data.siteInfo.admins.includes(
              respData.data.loginInfo.userID
            )
          ) {
            setIsSiteAdmin(true);
          } else {
            setIsSiteAdmin(false);
          }
          setThumbups(respData.data.list);
          setThumbupTotal(respData.data.total);
        }
      },
      undefined,
      setErrorToast
    );
  };

  React.useEffect(() => {
    if (params.site) {
      thumbupHistory(params.site as string, pageNumber - 1, pageSize);
    }
  }, []);

  return (
    <>
      <CustomHead
        headText={siteInfo ? 'Site ' + siteInfo.name : 'Site'}
        loginInfo={loginInfo}
        siteID={siteID}
        setLoading={setLoading}
        setErrorToast={setErrorToast}
      />
      <Stack direction="row" alignItems="center" sx={{ width: '100%' }}>
        <Typography variant="h4" gutterBottom>
          My Thumbup History
        </Typography>
      </Stack>
      <Container maxWidth="xl">
        {thumbups &&
          thumbups.map((row) => (
            <>
              <Card>
                <CardHeader
                  avatar={
                    <>
                      <CustomAvatar
                        text={row.postedByString}
                        width="30px"
                        height="30px"
                      />
                    </>
                  }
                  action={<></>}
                  title={row.postedByString}
                  subheader={
                    <>
                      <Stack direction="row">
                        <Stack>
                          {moment(row.postedAt)
                            .local()
                            .format('YYYY-MM-DD HH:mm:ss')}
                        </Stack>
                      </Stack>
                      <Stack direction="row">
                        <Link
                          underline="hover"
                          onClick={() => {
                            let matches = pathname.match('/site/\\w+');
                            if (matches) {
                              router.push(
                                matches[0] + '/post?id=' + row.postID
                              );
                            }
                          }}
                          sx={{
                            ':hover': {
                              cursor: 'pointer',
                            },
                          }}
                        >
                          <Typography sx={{ color: 'black' }} variant="h6">
                            {row.title}
                          </Typography>
                        </Link>
                      </Stack>
                    </>
                  }
                />
                <CardContent>{row.content}</CardContent>
              </Card>
              <Divider />
            </>
          ))}
        {thumbupTotal > pageSize && (
          <Stack
            direction="row"
            alignItems="center"
            justifyContent="center"
            sx={{ marginTop: '10px', width: '100%' }}
          >
            <Pagination
              page={pageNumber}
              defaultPage={pageNumber}
              onChange={(e, page) => {
                thumbupHistory(params.site as string, page - 1, pageSize);
                setPageNumber(page);
              }}
              count={
                thumbupTotal % pageSize > 0
                  ? (thumbupTotal - (thumbupTotal % pageSize)) / pageSize + 1
                  : thumbupTotal / pageSize
              }
              variant="outlined"
              shape="rounded"
            />
          </Stack>
        )}
      </Container>

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
