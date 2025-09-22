'use client';
import React from 'react';
import Stack from '@mui/material/Stack';
import { post } from '@/app/shared';
import moment from 'moment';
import Typography from '@mui/material/Typography';
import {
  useRouter,
  useParams,
  useSearchParams,
  usePathname,
} from 'next/navigation';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import Divider from '@mui/material/Divider';
import { styled } from '@mui/material/styles';
import {
  setLocalUserLoginState,
  removeLocalUserLoginState,
  getLocalUserLoginState,
  LoginState,
} from '@/app/site/shared';
import CustomHead from '@/app/site/CustomHead';
import Container from '@mui/material/Container';
import SearchIcon from '@mui/icons-material/Search';
import CustomMenu from '@/app/site/[site]/CustomMenu';
import { Site, Category } from '@/app/model';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CustomAvatar from '@/app/site/CustomAvatar';
import Pagination from '@mui/material/Pagination';
import Link from '@mui/material/Link';
import { yellow } from '@mui/material/colors';
import Box from '@mui/material/Box';

interface MatchedInfo {
  postID: number;
  title: string;
  postPostedAt: number;
  postPostedBy: number;
  postPostedByString: string;
  postContent: string;
  commentID: number;
  commentContent: string;
  commentPostedAt: number;
  commentPostedBy: number;
  commentPostedByString: string;
  postTitleMatched: boolean;
  postContentMatched: boolean;
  commentContentMatched: boolean;
}

const CustomContentPage = styled(Container)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {},
  [theme.breakpoints.up('sm')]: {
    float: 'right',
    width: '80%',
  },
}));

export default function Page() {
  const queries = useSearchParams();
  const params = useParams();
  const router = useRouter();
  const pathname = usePathname();

  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [loginInfo, setLoginInfo] = React.useState<null | LoginState>(null);
  const [siteID, setSiteID] = React.useState<number>(0);
  const [siteInfo, setSiteInfo] = React.useState<null | Site>(null);
  const [curMenuText, setCurMenuText] = React.useState('');
  const [categories, setCategories] = React.useState<null | Category[]>(null);
  const [isSiteAdmin, setIsSiteAdmin] = React.useState(false);
  const [matchedInfos, setMatchedInfos] = React.useState<null | MatchedInfo[]>(
    null
  );
  const [matchedInfoTotal, setMatchedInfoTotal] = React.useState(0);
  const [pageNumber, setPageNumber] = React.useState(1); // ui from 1, but api from 0;
  const [pageSize, setPageSize] = React.useState(10);
  const aggregatedSearchPage = (site: string, searchText: string) => {
    if (!site) {
      return;
    }
    if (searchText.length > 200) {
      setErrorToast('search text too long');
      return;
    }
    post(
      false,
      '/invoker/site/aggregatedSearchPage/v1',
      setLoading,
      { site, searchText },
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
          setMatchedInfos(respData.data.matchedInfos);
          setMatchedInfoTotal(respData.data.total);
          // clear q
          let q = queries.get('q');
          if (q) {
            setSearchText(q);
            let matches = pathname.match('/site/\\w+/search');
            if (matches) {
              router.push(matches[0]);
            }
          }
        }
      },
      undefined,
      setErrorToast
    );
  };
  const searchPostComment = (
    site: string,
    pageNumber: number,
    pageSize: number
  ) => {
    if (!site || !searchText) {
      return;
    }
    if (searchText.length > 200) {
      setErrorToast('search text too long');
      return;
    }
    post(
      false,
      '/invoker/site/searchPostComment/v1',
      setLoading,
      { site, searchText, pageNumber, pageSize },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setMatchedInfos(respData.data.list);
          setMatchedInfoTotal(respData.data.total);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const [searchText, setSearchText] = React.useState('');
  const handleSearchTextChange = (e: any) => {
    setSearchText(e.target.value);
  };
  const handleOnClickSearch = () => {
    searchPostComment(params.site as string, pageNumber - 1, pageSize);
  };

  const prefixOfMatched = (src: string, q: string): string => {
    let index = src.indexOf(q);
    return src.slice(0, index);
  };
  const suffixOfMatched = (src: string, q: string): string => {
    let index = src.indexOf(q);
    return src.slice(index + q.length);
  };

  React.useEffect(() => {
    setLoginInfo(getLocalUserLoginState());
    let q = queries.get('q') as string;
    aggregatedSearchPage(params.site as string, q ? q : '');
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
        curText={curMenuText}
        categories={categories}
        isSiteAdmin={isSiteAdmin}
      />
      <CustomContentPage maxWidth="xl">
        <Box sx={{ width: '100%' }}>
          <Stack
            direction="row"
            sx={{
              height: '1vh',
              bgcolor: '#eceff1',
              alignContent: 'center',
              justifyContent: 'center',
            }}
          >
            <Typography variant="h4" gutterBottom>
              Search Page
            </Typography>
          </Stack>
          <Stack
            direction="row"
            spacing={2}
            sx={{
              height: '19vh',
              bgcolor: '#eceff1',
              alignContent: 'center',
              justifyContent: 'center',
            }}
          >
            <Stack
              sx={{
                width: '70%',
                alignContent: 'center',
                justifyContent: 'center',
                '& TextField': { margin: '0', height: '50px' },
              }}
            >
              <TextField
                fullWidth
                value={searchText}
                onChange={handleSearchTextChange}
                placeholder="Search"
              />
            </Stack>
            <Stack
              sx={{
                alignContent: 'center',
                justifyContent: 'center',
                '& button': { margin: '0', height: '50px', width: '150px' },
              }}
            >
              <Button
                startIcon={<SearchIcon />}
                variant="contained"
                onClick={handleOnClickSearch}
              >
                Search
              </Button>
            </Stack>
          </Stack>
          <Stack
            spacing={1}
            direction="row"
            sx={{
              padding: '1rem',
            }}
          >
            <Container maxWidth="xl">
              {matchedInfos &&
                matchedInfos.map((row) => (
                  <>
                    <Card>
                      <CardHeader
                        avatar={
                          <>
                            <CustomAvatar
                              text={
                                row.commentContentMatched
                                  ? row.commentPostedByString
                                  : row.postPostedByString
                              }
                              width="30px"
                              height="30px"
                            />
                          </>
                        }
                        action={<></>}
                        title={
                          row.commentContentMatched
                            ? row.commentPostedByString
                            : row.postPostedByString
                        }
                        subheader={
                          <>
                            <Stack>
                              {moment(
                                row.commentContentMatched
                                  ? row.commentPostedAt
                                  : row.postPostedAt
                              )
                                .local()
                                .format('YYYY-MM-DD HH:mm:ss')}
                            </Stack>
                            <Stack>
                              <Link
                                underline="hover"
                                onClick={() => {
                                  let matches = pathname.match('/site/\\w+');
                                  if (matches) {
                                    if (row.commentContentMatched) {
                                      router.push(
                                        matches[0] +
                                          '/post?id=' +
                                          row.postID +
                                          '&anchor=comment&anchor_id=' +
                                          row.commentID
                                      );
                                    } else {
                                      router.push(
                                        matches[0] +
                                          '/post?id=' +
                                          row.postID +
                                          '&anchor=post'
                                      );
                                    }
                                  }
                                }}
                                sx={{
                                  ':hover': {
                                    cursor: 'pointer',
                                  },
                                }}
                              >
                                <Typography
                                  sx={{ color: 'black' }}
                                  variant="h6"
                                >
                                  {row.postTitleMatched ? (
                                    <>
                                      <span>
                                        {prefixOfMatched(row.title, searchText)}
                                      </span>
                                      <span
                                        style={{
                                          color: yellow[900],
                                        }}
                                      >
                                        {searchText}
                                      </span>
                                      <span>
                                        {suffixOfMatched(row.title, searchText)}
                                      </span>
                                    </>
                                  ) : (
                                    row.title
                                  )}
                                </Typography>
                              </Link>
                            </Stack>
                          </>
                        }
                      />
                      <CardContent>
                        <Typography
                          variant="body2"
                          sx={{ color: 'text.secondary' }}
                        >
                          {row.commentContentMatched ? (
                            <>
                              <span>
                                {prefixOfMatched(
                                  row.commentContent,
                                  searchText
                                )}
                              </span>
                              <span
                                style={{
                                  color: yellow[900],
                                }}
                              >
                                {searchText}
                              </span>
                              <span>
                                {suffixOfMatched(
                                  row.commentContent,
                                  searchText
                                )}
                              </span>
                            </>
                          ) : (
                            <>
                              <span>
                                {prefixOfMatched(row.postContent, searchText)}
                              </span>
                              <span
                                style={{
                                  color: yellow[900],
                                }}
                              >
                                {searchText}
                              </span>
                              <span>
                                {suffixOfMatched(row.postContent, searchText)}
                              </span>
                            </>
                          )}
                        </Typography>
                      </CardContent>
                    </Card>
                    <Divider />
                  </>
                ))}
              {matchedInfoTotal > pageSize && (
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
                      searchPostComment(
                        params.site as string,
                        page - 1,
                        pageSize
                      );
                      setPageNumber(page);
                    }}
                    count={
                      matchedInfoTotal % pageSize > 0
                        ? (matchedInfoTotal - (matchedInfoTotal % pageSize)) /
                            pageSize +
                          1
                        : matchedInfoTotal / pageSize
                    }
                    variant="outlined"
                    shape="rounded"
                  />
                </Stack>
              )}
            </Container>
          </Stack>
        </Box>

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
