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
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Grid from '@mui/material/Grid2';
import moment from 'moment';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Pagination from '@mui/material/Pagination';
import IconButton from '@mui/material/IconButton';
import CreateIcon from '@mui/icons-material/Create';
import { styled } from '@mui/material/styles';
import CustomHead from '@/app/site/CustomHead';
import {
  setLocalUserLoginState,
  LoginState,
  removeLocalUserLoginState,
} from '@/app/site/shared';
import Container from '@mui/material/Container';
import CustomMenu from '@/app/site/[site]/CustomMenu';
import { blue } from '@mui/material/colors';
import Link from '@mui/material/Link';
import { Site, Category, Post } from '@/app/model';
import NewPostDialog from '@/app/site/[site]/NewPostDialog';

const CustomAddPostButton = styled(Button)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: { display: 'none' },
  [theme.breakpoints.up('sm')]: {},
}));

const CustomAddPostIconButton = styled(IconButton)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {},
  [theme.breakpoints.up('sm')]: {
    display: 'none',
  },
}));

function a11yProps(index: number) {
  return {
    id: `simple-tab-${index}`,
    'aria-controls': `simple-tabpanel-${index}`,
  };
}

const CustomContentPage = styled(Container)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {},
  [theme.breakpoints.up('sm')]: {
    float: 'right',
    width: '80%',
  },
}));

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [loginInfo, setLoginInfo] = React.useState<null | LoginState>(null);
  const [siteID, setSiteID] = React.useState<number>(0);
  const [siteInfo, setSiteInfo] = React.useState<null | Site>(null);
  const [categories, setCategories] = React.useState<null | Category[]>(null);

  const getCategories = (siteID: number) => {
    post(
      false,
      '/invoker/site/getCategories/v1',
      setLoading,
      { siteID },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data && respData.data.list.length > 0) {
          setCategories(respData.data.list);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const [posts, setPosts] = React.useState<Post[] | null>(null);
  const [openNewPostDialog, setOpenNewPostDialog] = React.useState(false);
  const handleCloseNewPostDialog = () => {
    setOpenNewPostDialog(false);
  };

  const [sortedByActivity, setSortedByActivity] = React.useState(true);
  const [sortedByViews, setSortedByViews] = React.useState(false);
  const [sortedByReplies, setSortedByReplies] = React.useState(false);
  const [pageNumber, setPageNumber] = React.useState(1); // ui from 1, but api from 0;
  const [pageSize, setPageSize] = React.useState(10);
  const [postTotal, setPostTotal] = React.useState(0);
  const [value, setValue] = React.useState(0);

  const handleChange = (event: React.SyntheticEvent, newValue: number) => {
    let tmpSortedByActivity = true;
    let tmpSortedByViews = false;
    let tmpSortedByReplies = false;
    switch (newValue) {
      case 0:
        tmpSortedByActivity = true;
        tmpSortedByViews = false;
        tmpSortedByReplies = false;
        break;
      case 1:
        tmpSortedByActivity = false;
        tmpSortedByViews = true;
        tmpSortedByReplies = false;
        break;
      case 2:
        tmpSortedByActivity = false;
        tmpSortedByViews = false;
        tmpSortedByReplies = true;
        break;
    }
    setSortedByActivity(tmpSortedByActivity);
    setSortedByViews(tmpSortedByViews);
    setSortedByReplies(tmpSortedByReplies);
    setValue(newValue);
    getPosts(
      siteID,
      tmpSortedByActivity,
      tmpSortedByViews,
      tmpSortedByReplies,
      pageNumber - 1,
      pageSize
    );
  };

  const router = useRouter();
  const pathname = usePathname();
  const params = useParams();
  const [isSiteAdmin, setIsSiteAdmin] = React.useState(false);
  const aggregatedSitePage = async (site: string) => {
    await post(
      false,
      '/invoker/site/aggregatedSitePage/v1',
      setLoading,
      { site },
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
          setCategories(respData.data.categories);
          setPosts(respData.data.posts);
          setPostTotal(respData.data.postTotal);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const getPosts = (
    siteID: number,
    sortedByActivity: boolean,
    sortedByViews: boolean,
    sortedByReplies: boolean,
    pageNumber: number,
    pageSize: number
  ) => {
    post(
      false,
      '/invoker/site/getPosts/v1',
      setLoading,
      {
        siteID,
        sortedByActivity,
        sortedByViews,
        sortedByReplies,
        pageNumber,
        pageSize,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          if (respData.data.list.length > 0) {
            setPosts(respData.data.list);
          }
          setPostTotal(respData.data.total);
        }
      },
      undefined,
      setErrorToast
    );
  };

  React.useEffect(() => {
    if (params.site) {
      // aggregated api, with login state
      aggregatedSitePage(params.site as string);
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
      <CustomMenu
        curText={'All Categories'}
        categories={categories}
        isSiteAdmin={isSiteAdmin}
      />
      <CustomContentPage maxWidth="xl">
        <Grid container spacing={2} sx={{ marginBottom: '3px' }}>
          <Grid size={10}>
            <Tabs
              value={value}
              onChange={handleChange}
              aria-label="basic tabs example"
            >
              <Tab label="Latest" {...a11yProps(0)} />
              <Tab label="Most View" {...a11yProps(1)} />
              <Tab label="Most Replies" {...a11yProps(2)} />
            </Tabs>
          </Grid>
          <Grid size={2}>
            {loginInfo && (
              <CustomAddPostButton
                onClick={() => {
                  getCategories(siteID);
                  setOpenNewPostDialog(true);
                }}
                variant="contained"
                startIcon={<CreateIcon />}
              >
                New A Post
              </CustomAddPostButton>
            )}
            {loginInfo && (
              <CustomAddPostIconButton
                onClick={() => {
                  setOpenNewPostDialog(true);
                }}
                size="large"
                sx={{ backgroundColor: '#1976d2' }}
              >
                <CreateIcon sx={{ color: 'white' }} />
              </CustomAddPostIconButton>
            )}
          </Grid>
        </Grid>
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
                getPosts(
                  siteID,
                  sortedByActivity,
                  sortedByViews,
                  sortedByReplies,
                  page - 1,
                  pageSize
                );
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

        {/* new a post dialog */}
        <NewPostDialog
          site_id={siteID}
          category_id={0}
          open={openNewPostDialog}
          handleClose={handleCloseNewPostDialog}
          successAction={() => {
            getPosts(
              siteID,
              sortedByActivity,
              sortedByViews,
              sortedByReplies,
              pageNumber - 1,
              pageSize
            );
          }}
          categories={categories}
        />

        <CustomSnackbar
          toast={toast}
          setToast={setToast}
          errorToast={errorToast}
          setErrorToast={setErrorToast}
        />
        <CustomBackdrop loading={loading} />
      </CustomContentPage>
    </>
  );
}
