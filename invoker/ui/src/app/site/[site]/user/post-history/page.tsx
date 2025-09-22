'use client';
import React from 'react';
import { post } from '@/app/shared';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
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
import { blue } from '@mui/material/colors';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import { Site, Post } from '@/app/model';

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
  const [posts, setPosts] = React.useState<Post[] | null>(null);
  const [postTotal, setPostTotal] = React.useState(0);
  const postHistory = (site: string, pageNumber: number, pageSize: number) => {
    post(
      false,
      '/invoker/user/postHistory/v1',
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
          setPosts(respData.data.list);
          setPostTotal(respData.data.total);
        }
      },
      undefined,
      setErrorToast
    );
  };

  React.useEffect(() => {
    if (params.site) {
      postHistory(params.site as string, pageNumber - 1, pageSize);
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
          My Post History
        </Typography>
      </Stack>
      <Stack direction="row" alignItems="center" sx={{ width: '100%' }}>
        <TableContainer>
          <Table>
            <TableHead
              sx={{
                borderStyle: 'none none solid none',
                borderWidth: '3px',
                borderColor: 'rgba(177, 175, 175, 0.5)',
              }}
            >
              <TableRow>
                <TableCell sx={{ color: '#929698' }}>Post</TableCell>
                <TableCell
                  sx={{ color: '#929698', width: '100px' }}
                ></TableCell>
                <TableCell
                  sx={{
                    color: '#929698',
                    width: '50px',
                  }}
                >
                  Replies
                </TableCell>
                <TableCell sx={{ color: '#929698', width: '50px' }}>
                  Views
                </TableCell>
                <TableCell sx={{ color: '#929698', width: '100px' }}>
                  Activity
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {posts &&
                posts.map((row) => (
                  <TableRow key={row.id} hover component={Stack}>
                    <TableCell component="th" scope="row">
                      <Link
                        underline="hover"
                        color="black"
                        sx={{
                          '&.MuiLink-underlineHover:hover': {
                            color: blue[800],
                            cursor: 'pointer',
                          },
                        }}
                        onClick={() => {
                          let matches = pathname.match('/site/\\w+');
                          if (matches) {
                            router.push(matches[0] + '/post?id=' + row.id);
                          }
                        }}
                      >
                        {row.title}
                      </Link>
                    </TableCell>
                    <TableCell align="center"></TableCell>
                    <TableCell align="center">{row.replies}</TableCell>
                    <TableCell align="center">{row.views}</TableCell>
                    <TableCell align="center">
                      {moment
                        .duration(moment().diff(moment(row.activity)))
                        .humanize()}
                    </TableCell>
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Stack>
      {postTotal > pageSize && (
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
              postHistory(params.site as string, page - 1, pageSize);
              setPageNumber(page);
            }}
            count={
              postTotal % pageSize > 0
                ? (postTotal - (postTotal % pageSize)) / pageSize + 1
                : postTotal / pageSize
            }
            variant="outlined"
            shape="rounded"
          />
        </Stack>
      )}

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
