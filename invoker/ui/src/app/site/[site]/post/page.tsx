'use client';
import React from 'react';
import Stack from '@mui/material/Stack';
import { post, uploadFile } from '@/app/shared';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import moment from 'moment';
import {
  useRouter,
  useParams,
  useSearchParams,
  usePathname,
} from 'next/navigation';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import CommentComponent from '@/app/site/[site]/post/Comment';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import Divider from '@mui/material/Divider';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CustomAvatar from '@/app/site/CustomAvatar';
import IconButton from '@mui/material/IconButton';
import EditIcon from '@mui/icons-material/Edit';
import DoneIcon from '@mui/icons-material/Done';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import Grid from '@mui/material/Grid2';
import { styled } from '@mui/material/styles';
import {
  removeLocalUserLoginState,
  setLocalUserLoginState,
  getLocalUserLoginState,
  LoginState,
} from '@/app/site/shared';
import CustomHead from '@/app/site/CustomHead';
import CustomMenu from '@/app/site/[site]/CustomMenu';
import Container from '@mui/material/Container';
import Badge from '@mui/material/Badge';
import pica from 'pica';
import FavoriteBorderIcon from '@mui/icons-material/FavoriteBorder';
import FavoriteIcon from '@mui/icons-material/Favorite';
import { Site, Category, Post, Comment } from '@/app/model';
import { green } from '@mui/material/colors';

interface AnchorInfo {
  rootCommentID: number;
  firstLevelPageNumber: number;
  secondLevelPageNumber: number;
  secondLevelTotal: number;
}

const FadeOutBgcolorCard = styled(Card)(({ theme }) => ({
  '@keyframes fadeOut': {
    '0%': {
      backgroundColor: green[100],
    },
    '100%': {},
  },
}));

const VisuallyHiddenInput = styled('input')({
  clip: 'rect(0 0 0 0)',
  clipPath: 'inset(50%)',
  height: 1,
  overflow: 'hidden',
  position: 'absolute',
  bottom: 0,
  left: 0,
  whiteSpace: 'nowrap',
  width: 1,
});

const CustomContentPage = styled(Container)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {},
  [theme.breakpoints.up('sm')]: {
    float: 'right',
    width: '80%',
  },
}));

const CustomPostTitle = styled(Typography)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: { fontSize: '24px' },
  [theme.breakpoints.up('sm')]: { fontSize: '36px' },
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
  const [categoryInfo, setCategoryInfo] = React.useState<null | Category>(null);
  const [categories, setCategories] = React.useState<null | Category[]>(null);
  const [anchor, setAnchor] = React.useState<null | string>(null); // null/post/comment
  const [anchorID, setAnchorID] = React.useState<null | number>(null);
  const [anchorInfo, setAnchorInfo] = React.useState<null | AnchorInfo>(null);

  const [curPost, setCurPost] = React.useState<Post | null>(null); // post data backup
  let blankPost: Post = {
    id: 0,
    siteID: 0,
    category: '',
    categoryID: 0,
    title: '',
    postedAt: 0,
    postedBy: 0,
    postedByString: '',
    content: '',
    image: '',
    state: 0,
    replies: 0,
    views: 0,
    activity: 0,
    thumbups: 0,
    thumbup: false,
  };
  const [newPost, setNewPost] = React.useState<Post>(blankPost);
  const [thumbup, setThumbup] = React.useState(false);
  const [imgFile, setImgFile] = React.useState<null | File>(null);
  const [imgUrl, setImgUrl] = React.useState<null | string>(null);
  const [originComments, setOriginComments] = React.useState<Comment[] | null>(
    null
  );
  const [originTotal, setOriginTotal] = React.useState(0);
  const [isWritingTitle, setIsWritingTitle] = React.useState(false);
  const [isWritingContent, setIsWritingContent] = React.useState(false);
  const [isPoster, setIsPoster] = React.useState(false);

  const getPostDetail = (postID: number) => {
    post(
      false,
      '/invoker/site/getPostDetail/v1',
      setLoading,
      { id: postID },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setCurPost(respData.data);
          setNewPost(respData.data);
          setThumbup(respData.data.thumbup);
          let userState = getLocalUserLoginState();
          setIsPoster(userState && userState.userID == respData.data.postedBy);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const editPost = (newPost: Post) => {
    // upload image first, if has changed
    if (imgFile) {
      uploadFile(
        setLoading,
        imgFile,
        'images',
        (respData: any) => {
          let image =
            'https://d2y6ia7j6nkf8t.cloudfront.net/images/' +
            respData.data.hashName;
          // post editPost
          post(
            false,
            '/invoker/site/editPost/v1',
            setLoading,
            {
              id: newPost.id,
              title: newPost.title,
              content: newPost.content,
              image: image,
            },
            (respHeaders: any) => {},
            (respData: any) => {
              // refresh the post
              let postID = queries.get('id');
              if (postID) {
                getPostDetail(parseInt(postID));
              }
              setImgFile(null);
              setImgUrl(image);
            },
            setToast,
            setErrorToast
          );
        },
        setToast,
        setErrorToast
      );
    } else {
      post(
        false,
        '/invoker/site/editPost/v1',
        setLoading,
        {
          id: newPost.id,
          title: newPost.title,
          content: newPost.content,
          image: imgUrl,
        },
        (respHeaders: any) => {},
        (respData: any) => {
          let postID = queries.get('id');
          if (postID) {
            getPostDetail(parseInt(postID));
          }
        },
        setToast,
        setErrorToast
      );
    }
  };

  const [openDialog, setOpenDialog] = React.useState(false);
  const handleCloseDialog = () => {
    setOpenDialog(false);
  };

  const deletePost = (id: number) => {
    post(
      false,
      '/invoker/site/deletePost/v1',
      setLoading,
      {
        id: id,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (categoryInfo) {
          router.push(
            '/site/' + params.site + '/category?name=' + categoryInfo.name
          );
        } else {
          router.push('/site/' + params.site);
        }
      },
      setToast,
      setErrorToast
    );
  };

  const [isSiteAdmin, setIsSiteAdmin] = React.useState(false);
  const aggregatedPostPage = async (
    site: string,
    postID: number,
    anchorCommentID: number
  ) => {
    await post(
      false,
      '/invoker/site/aggregatedPostPage/v1',
      setLoading,
      { site, postID, anchorCommentID },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.code != 0) {
          setErrorToast('fail to get site info of ' + site);
          return;
        }
        if (respData.data) {
          // login state
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
          // isSiteAdmin
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
          // isPoster
          setIsPoster(
            respData.data.loginInfo &&
              respData.data.loginInfo.userID == respData.data.post.postedBy
          );
          // page state
          setCategoryInfo(respData.data.categoryInfo);
          setCategories(respData.data.categories);
          setCurPost(respData.data.post);
          setNewPost(respData.data.post);
          setThumbup(respData.data.post.thumbup);
          setOriginComments(respData.data.post.comments);
          setOriginTotal(respData.data.post.commentTotal);
          if (respData.data.post.image) {
            setImgUrl(respData.data.post.image);
          }
          // comment anchor info
          if (respData.data.anchorInfo) {
            setAnchorInfo(respData.data.anchorInfo);
          }
          // clear anchor
          let anchor = queries.get('anchor');
          if (anchor) {
            setAnchor(anchor);
            let matches = pathname.match('/site/\\w+/post');
            if (matches) {
              router.push(matches[0] + '?id=' + postID);
            }
          }
        }
      },
      undefined,
      setErrorToast
    );
  };

  const thumbupPost = (siteID: number, categoryID: number, postID: number) => {
    let loginState = getLocalUserLoginState();
    if (!loginState) {
      return;
    }
    post(
      false,
      '/invoker/site/thumbupPost/v1',
      setLoading,
      {
        siteID,
        categoryID,
        postID,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (thumbup) {
          let post = { ...newPost };
          post.thumbups--;
          post.thumbup = false;
          setNewPost(post);
          setThumbup(false);
          if (curPost) {
            let post2 = { ...curPost };
            post2.thumbups--;
            post2.thumbup = false;
            setCurPost(post2);
          }
        } else {
          let post = { ...newPost };
          post.thumbups++;
          post.thumbup = true;
          setNewPost(post);
          setThumbup(true);
          if (curPost) {
            let post2 = { ...curPost };
            post2.thumbups++;
            post2.thumbup = true;
            setCurPost(post2);
          }
        }
      },
      setToast,
      setErrorToast
    );
  };

  React.useEffect(() => {
    let postID = queries.get('id');
    let anchorID = queries.get('anchor_id');
    if (anchorID) {
      setAnchorID(parseInt(anchorID));
    }
    if (params.site && postID) {
      // aggregated api, with login state
      aggregatedPostPage(
        params.site as string,
        parseInt(postID),
        anchorID ? parseInt(anchorID) : 0
      );
    }
  }, [params.site as string, queries.get('id') as string]);

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
        curText={categoryInfo ? categoryInfo.name : ''}
        categories={categories}
        isSiteAdmin={isSiteAdmin}
      />
      <CustomContentPage maxWidth="xl">
        <Box sx={{ width: '100%' }}>
          <Stack spacing={1} direction="row" alignItems="center">
            <Grid
              container
              spacing={1}
              sx={{ placeItems: 'center', width: '100%' }}
            >
              <Grid size={11}>
                {!isWritingTitle && (
                  <CustomPostTitle gutterBottom>
                    {newPost.title}
                  </CustomPostTitle>
                )}
                {isWritingTitle && (
                  <TextField
                    id="title"
                    name="title"
                    margin="normal"
                    fullWidth
                    value={newPost.title}
                    onChange={(e) => {
                      let post = { ...newPost };
                      post.title = e.target.value;
                      setNewPost(post);
                    }}
                  />
                )}
              </Grid>
              <Grid size={1}>
                <Stack
                  direction="row"
                  spacing={1}
                  sx={{
                    display: 'flex',
                    flexDirection: 'row-reverse',
                  }}
                >
                  {/* 管理员只能删除别人的 回复或者帖子， 不能修改别人的内容 */}
                  {isPoster && !isWritingTitle && (
                    <IconButton
                      size="small"
                      aria-label="edit the post title"
                      onClick={() => {
                        setIsWritingTitle(true);
                      }}
                    >
                      <EditIcon />
                    </IconButton>
                  )}
                  {isWritingTitle && (
                    <>
                      <IconButton
                        size="small"
                        aria-label="cancel the edit"
                        onClick={() => {
                          setIsWritingTitle(false);
                        }}
                      >
                        <CloseIcon />
                      </IconButton>
                      <IconButton
                        size="small"
                        aria-label="commit the edit"
                        onClick={() => {
                          editPost(newPost);
                          setIsWritingTitle(false);
                        }}
                      >
                        <DoneIcon />
                      </IconButton>
                    </>
                  )}
                </Stack>
              </Grid>
            </Grid>
          </Stack>
          <Divider />
          <Stack
            spacing={1}
            direction="row"
            sx={{ minHeight: '300px', width: '100%' }}
          >
            <FadeOutBgcolorCard
              sx={
                anchor && anchor == 'post'
                  ? { width: '100%', animation: 'fadeOut 3s ease-out' }
                  : { width: '100%' }
              }
            >
              <CardHeader
                avatar={
                  <CustomAvatar
                    text={newPost.postedByString}
                    width="30px"
                    height="30px"
                  />
                }
                action={
                  <>
                    {!isWritingContent && (
                      <>
                        {newPost.thumbups}
                        <IconButton
                          size="small"
                          aria-label="thumbup the post"
                          onClick={() => {
                            if (categoryInfo && curPost) {
                              thumbupPost(siteID, categoryInfo.id, curPost.id);
                            }
                          }}
                        >
                          {thumbup ? (
                            <FavoriteIcon color="error" />
                          ) : (
                            <FavoriteBorderIcon />
                          )}
                        </IconButton>
                      </>
                    )}
                    {/* 管理员只能删除别人的 回复或者帖子， 不能修改别人的内容 */}
                    {isPoster && !isWritingContent && (
                      <IconButton
                        size="small"
                        aria-label="edit the post content"
                        onClick={() => {
                          setIsWritingContent(true);
                        }}
                      >
                        <EditIcon />
                      </IconButton>
                    )}
                    {(isPoster || isSiteAdmin) && !isWritingContent && (
                      <IconButton
                        size="small"
                        aria-label="delete the post"
                        onClick={() => {
                          setOpenDialog(true);
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    )}
                    {isWritingContent && (
                      <>
                        <IconButton
                          size="small"
                          aria-label="commit the edit"
                          onClick={() => {
                            editPost(newPost);
                            setIsWritingContent(false);
                          }}
                        >
                          <DoneIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          aria-label="cancel the edit"
                          onClick={() => {
                            setIsWritingContent(false);
                            setImgFile(null);
                            if (curPost) {
                              setImgUrl(curPost.image);
                            }
                          }}
                        >
                          <CloseIcon />
                        </IconButton>
                      </>
                    )}
                  </>
                }
                title={newPost.postedByString}
                subheader={
                  <>
                    <Stack>
                      {moment(newPost.postedAt)
                        .local()
                        .format('YYYY-MM-DD HH:mm:ss')}
                    </Stack>
                  </>
                }
              />
              <CardContent>
                {!isWritingContent && (
                  <>
                    <Typography
                      variant="body2"
                      sx={{
                        color: 'text.secondary',
                      }}
                    >
                      {newPost.content}
                    </Typography>
                  </>
                )}
                {isWritingContent && (
                  <>
                    <TextField
                      id="content"
                      name="content"
                      margin="normal"
                      fullWidth
                      value={newPost.content}
                      multiline
                      minRows={10}
                      onChange={(e) => {
                        let post = { ...newPost };
                        post.content = e.target.value;
                        setNewPost(post);
                      }}
                    />
                    {!imgUrl && (
                      <Button
                        component="label"
                        sx={{
                          width: '300px',
                          height: '200px',
                          border: '1px dashed grey',
                        }}
                        onClick={() => {}}
                      >
                        <VisuallyHiddenInput
                          type="file"
                          onChange={(e) => {
                            let files = e.target.files;
                            if (files) {
                              let file = files[0];
                              if (!/image/i.test(file.type)) {
                                setErrorToast(
                                  'File ' + file.name + ' is not an image.'
                                );
                                return;
                              }
                              // resize the image, if it is gt 1M
                              let imgUrl = URL.createObjectURL(file);
                              if (file.size <= 1000000) {
                                setImgUrl(imgUrl);
                                setImgFile(file);
                                return;
                              }
                              let image = new Image();
                              image.src = imgUrl;
                              image.onload = () => {
                                let scale = 1000000 / file.size;
                                let width = image.width * scale;
                                let height = image.height * scale;
                                let picaInst = pica();
                                let to = document.createElement('canvas');
                                to.width = width;
                                to.height = height;
                                picaInst
                                  .resize(image, to)
                                  .then((result) =>
                                    picaInst.toBlob(result, file.type, 0.7)
                                  )
                                  .then((blob) => {
                                    let newImgUrl = URL.createObjectURL(blob);
                                    setImgUrl(newImgUrl);
                                    let newFile = new File([blob], file.name, {
                                      type: file.type,
                                      lastModified: file.lastModified,
                                    });
                                    setImgFile(newFile);
                                  });
                              };
                            }
                          }}
                        />
                        +
                      </Button>
                    )}
                  </>
                )}
                {imgUrl && (
                  <Badge
                    sx={{
                      marginTop: '20px',
                    }}
                    badgeContent={
                      isWritingContent && (
                        <IconButton
                          size="small"
                          onClick={() => {
                            setImgFile(null);
                            setImgUrl(null);
                          }}
                        >
                          <CloseIcon color="error" />
                        </IconButton>
                      )
                    }
                  >
                    <img
                      src={imgUrl ? imgUrl : ''}
                      alt={imgUrl}
                      loading="lazy"
                      width="600px"
                      height="300px"
                    />
                  </Badge>
                )}
              </CardContent>
            </FadeOutBgcolorCard>
          </Stack>
          {originComments && (
            <Stack spacing={1} direction="row" alignItems="center">
              <CommentComponent
                anchor={anchor}
                anchorID={anchorID}
                anchorInfo={anchorInfo}
                siteID={siteID}
                categoryInfo={categoryInfo}
                isSiteAdmin={isSiteAdmin}
                postID={newPost.id}
                originComments={originComments}
                originTotal={originTotal}
              ></CommentComponent>
            </Stack>
          )}
        </Box>

        {/* delete post dialog */}
        <Dialog
          maxWidth="lg"
          fullWidth
          open={openDialog}
          onClose={handleCloseDialog}
        >
          <DialogContent>
            <Typography variant="h6" gutterBottom>
              Are you sure to delete the post ?
            </Typography>
          </DialogContent>
          <DialogActions>
            <Button
              onClick={() => {
                handleCloseDialog();
              }}
            >
              Close
            </Button>
            <Button
              onClick={() => {
                let postID = queries.get('id');
                if (postID) {
                  deletePost(parseInt(postID));
                }
              }}
            >
              Confirm
            </Button>
          </DialogActions>
        </Dialog>

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
