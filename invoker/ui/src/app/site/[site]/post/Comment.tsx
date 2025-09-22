'use client';
import React from 'react';
import { post as httpPost } from '@/app/shared';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid2';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import moment from 'moment';
import CustomAvatar from '@/app/site/CustomAvatar';
import IconButton from '@mui/material/IconButton';
import CommentIcon from '@mui/icons-material/Comment';
import CardActions from '@mui/material/CardActions';
import Collapse from '@mui/material/Collapse';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import EditIcon from '@mui/icons-material/Edit';
import DoneIcon from '@mui/icons-material/Done';
import CloseIcon from '@mui/icons-material/Close';
import { getLocalUserLoginState } from '@/app/site/shared';
import DeleteIcon from '@mui/icons-material/Delete';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import Pagination from '@mui/material/Pagination';
import FavoriteBorderIcon from '@mui/icons-material/FavoriteBorder';
import FavoriteIcon from '@mui/icons-material/Favorite';
import { Category, Comment } from '@/app/model';
import { styled } from '@mui/material/styles';
import { green } from '@mui/material/colors';

const CustomNewCommentDialog = ({
  open,
  setOpen,
  postID,
  rootCommentID,
  atWho,
  atWhoString,
  successAction,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
  postID: number;
  rootCommentID: number;
  atWho: number;
  atWhoString: string;
  successAction?: () => void;
}) => {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [content, setContent] = React.useState('');

  const handleClose = () => {
    setOpen(false);
  };

  const addComment = async () => {
    await httpPost(
      false,
      '/invoker/site/addComment/v1',
      setLoading,
      {
        postID: postID,
        rootCommentID: rootCommentID,
        atWho,
        content,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        setContent('');
        if (successAction) {
          successAction();
        }
        setOpen(false);
      },
      setToast,
      setErrorToast
    );
  };

  return (
    <Dialog maxWidth="lg" fullWidth open={open} onClose={handleClose}>
      <DialogTitle>New A Comment</DialogTitle>
      <DialogContent>
        {atWho > 0 && (
          <Stack direction="row" alignItems="center">
            <Typography
              variant="h4"
              sx={{ color: '#0000FF', marginRight: '10px' }}
            >
              Reply to:
            </Typography>
            <Typography variant="h6">{atWhoString}</Typography>
          </Stack>
        )}
        <Stack direction="row" alignItems="center">
          <TextField
            id="comment"
            name="comment"
            fullWidth
            margin="normal"
            value={content}
            multiline
            minRows={10}
            onChange={(e) => {
              setContent(e.target.value);
            }}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button
          onClick={() => {
            handleClose();
          }}
        >
          Close
        </Button>
        <Button
          onClick={() => {
            addComment();
          }}
        >
          Post
        </Button>
      </DialogActions>
      <CustomSnackbar
        toast={toast}
        setToast={setToast}
        errorToast={errorToast}
        setErrorToast={setErrorToast}
      />
      <CustomBackdrop loading={loading} />
    </Dialog>
  );
};

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

const CommentCard = ({
  anchor,
  anchorID,
  anchorInfo,
  siteID,
  categoryInfo,
  isSiteAdmin,
  cardSX,
  outerComments,
  setOuterComments,
  comment,
  isFirstLevel,
  setLoading,
  setToast,
  setErrorToast,
}: {
  anchor: null | string;
  anchorID: null | number;
  anchorInfo: null | AnchorInfo;
  siteID: number;
  categoryInfo: null | Category;
  isSiteAdmin: boolean;
  cardSX: any;
  outerComments: Comment[];
  setOuterComments: (comments: Comment[]) => void;
  comment: Comment;
  isFirstLevel: boolean;
  setLoading: (loading: boolean) => void;
  setToast: (toast: string) => void;
  setErrorToast: (errorToast: string) => void;
}) => {
  const [pageNumber, setPageNumber] = React.useState(1); // ui from 1, but api from 0;
  const [pageSize, setPageSize] = React.useState(10);
  const [comments, setComments] = React.useState<Comment[] | null>(null);
  const [total, setTotal] = React.useState(0);
  const [isPoster, setIsPoster] = React.useState(false);
  const [commentsLength, setCommentsLength] = React.useState(0);

  const getComments = async (
    commentID: number,
    pageNumber: number,
    pageSize: number
  ) => {
    await httpPost(
      false,
      '/invoker/site/getComments/v1',
      setLoading,
      {
        commentID,
        pageSize,
        pageNumber,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          if (respData.data.list) {
            setComments(respData.data.list);
            setCommentsLength(respData.data.list.length);
          }
          setTotal(respData.data.total);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const [expanded, setExpanded] = React.useState(false);

  const handleExpandClick = () => {
    let toExpand = !expanded;
    setExpanded(toExpand);
    if (toExpand) {
      if (!comments || comments.length != commentsLength) {
        getComments(comment.id, pageNumber - 1, pageSize);
      }
    }
  };

  const [isUpdateComment, setIsUpdateComment] = React.useState(false);
  const [content, setContent] = React.useState('');
  const editComment = async () => {
    await httpPost(
      false,
      '/invoker/site/editComment/v1',
      setLoading,
      {
        id: comment.id,
        content,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        comment.content = content;
        setContent('');
        setIsUpdateComment(false);
      },
      setToast,
      setErrorToast
    );
  };

  const [openDeleteDialog, setOpenDeleteDialog] = React.useState(false);
  const handleCloseDeleteDialog = () => {
    setOpenDeleteDialog(false);
  };

  const deleteComment = async () => {
    await httpPost(
      false,
      '/invoker/site/deleteComment/v1',
      setLoading,
      {
        id: comment.id,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        let tmpComments: Comment[] = [];
        for (let i = 0; i < outerComments.length; i++) {
          if (outerComments[i].id == comment.id) {
            continue;
          }
          tmpComments.push(outerComments[i]);
        }
        setOuterComments(tmpComments);
      },
      setToast,
      setErrorToast
    );
  };

  const [thumbup, setThumbup] = React.useState(comment.thumbup);
  const thumbupComment = (
    siteID: number,
    categoryID: number,
    postID: number,
    commentID: number
  ) => {
    let loginState = getLocalUserLoginState();
    if (!loginState) {
      return;
    }
    httpPost(
      false,
      '/invoker/site/thumbupComment/v1',
      setLoading,
      {
        siteID,
        categoryID,
        postID,
        commentID,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (thumbup) {
          comment.thumbups--;
          comment.thumbup = false;
          setThumbup(false);
        } else {
          comment.thumbups++;
          comment.thumbup = true;
          setThumbup(true);
        }
      },
      setToast,
      setErrorToast
    );
  };

  const [openDialog, setOpenDialog] = React.useState(false);
  const [atWho, setAtWho] = React.useState(0);
  const [atWhoString, setAtWhoString] = React.useState('');
  const cardRef = React.useRef<null | HTMLDivElement>(null);

  React.useEffect(() => {
    setExpanded(false);
    setComments(null);
    let userState = getLocalUserLoginState();
    setIsPoster(userState && userState.userID == comment.postedBy);
    setThumbup(comment.thumbup);
    if (anchorID && anchorInfo) {
      if (
        anchorInfo.rootCommentID > 0 &&
        anchorInfo.rootCommentID == comment.id &&
        isFirstLevel
      ) {
        setExpanded(true);
        setPageNumber(anchorInfo.secondLevelPageNumber);
        getComments(comment.id, anchorInfo.secondLevelPageNumber, pageSize);
      }
      if (anchorID == comment.id && cardRef.current) {
        cardRef.current.scrollIntoView({
          behavior: 'smooth',
          block: 'nearest',
          inline: 'center',
        });
      }
    }
  }, [comment]);

  return (
    <>
      <FadeOutBgcolorCard
        sx={
          anchor && anchor == 'comment' && anchorID == comment.id
            ? { ...cardSX, animation: 'fadeOut 3s ease-out' }
            : { ...cardSX }
        }
        ref={cardRef}
      >
        <CardHeader
          avatar={
            <CustomAvatar
              text={comment.postedByString}
              width="20px"
              height="20px"
            />
          }
          action={
            <>
              {!isUpdateComment && (
                <>
                  {comment.thumbups}
                  <IconButton
                    size="small"
                    aria-label="thumbup the comment"
                    onClick={() => {
                      thumbupComment(
                        siteID,
                        categoryInfo ? categoryInfo.id : 0,
                        comment.postID,
                        comment.id
                      );
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
              {isPoster && !isUpdateComment && (
                <IconButton
                  size="small"
                  aria-label="edit the comment"
                  onClick={() => {
                    setContent(comment.content);
                    setIsUpdateComment(true);
                  }}
                >
                  <EditIcon />
                </IconButton>
              )}
              {(isPoster || isSiteAdmin) && !isUpdateComment && (
                <IconButton
                  size="small"
                  aria-label="delete the comment"
                  onClick={() => {
                    setOpenDeleteDialog(true);
                  }}
                >
                  <DeleteIcon />
                </IconButton>
              )}
              {!isUpdateComment && (
                <>
                  {getLocalUserLoginState() && (
                    <IconButton
                      size="small"
                      aria-label="add a comment"
                      onClick={() => {
                        // if the comment is after second level comment, it should have atWho
                        if (comment.rootCommentID) {
                          setAtWho(comment.postedBy);
                          setAtWhoString(comment.postedByString);
                        }
                        setOpenDialog(true);
                      }}
                    >
                      <CommentIcon />
                    </IconButton>
                  )}
                </>
              )}
              {isUpdateComment && (
                <>
                  <IconButton
                    size="small"
                    aria-label="commit the edit"
                    onClick={() => {
                      editComment();
                    }}
                  >
                    <DoneIcon />
                  </IconButton>
                  <IconButton
                    size="small"
                    aria-label="cancel the edit"
                    onClick={() => {
                      setContent('');
                      setIsUpdateComment(false);
                    }}
                  >
                    <CloseIcon />
                  </IconButton>
                </>
              )}
            </>
          }
          title={<Stack>{comment.postedByString}</Stack>}
          subheader={
            <>
              <Stack>
                {moment(comment.postedAt).local().format('YYYY-MM-DD HH:mm:ss')}
              </Stack>
              {comment.atWho > 0 && (
                <Stack sx={{ color: '#FFA500' }}>
                  {'@ ' + comment.atWhoString}
                </Stack>
              )}
            </>
          }
        />
        <CardContent>
          {!isUpdateComment && (
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {comment.content}
            </Typography>
          )}
          {isUpdateComment && (
            <TextField
              id="content"
              name="content"
              margin="normal"
              fullWidth
              value={content}
              multiline
              minRows={10}
              onChange={(e) => {
                setContent(e.target.value);
              }}
            />
          )}
        </CardContent>
        {comment.replies > 0 && (
          <>
            <Button
              onClick={handleExpandClick}
              endIcon={expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
            >
              {comment.replies == 1 ? '1 Reply' : comment.replies + ' Replies'}
            </Button>
            <CardActions disableSpacing>
              <Collapse
                in={expanded}
                timeout="auto"
                unmountOnExit
                sx={{ width: '100%' }}
              >
                {comments &&
                  comments.map((item, idx) => {
                    return (
                      <CommentCard
                        anchor={anchor}
                        anchorID={anchorID}
                        anchorInfo={anchorInfo}
                        siteID={siteID}
                        categoryInfo={categoryInfo}
                        isSiteAdmin={isSiteAdmin}
                        cardSX={{ width: '98%', float: 'right' }}
                        key={item.id}
                        outerComments={comments}
                        setOuterComments={setComments}
                        comment={item}
                        isFirstLevel={false}
                        setLoading={setLoading}
                        setToast={setToast}
                        setErrorToast={setErrorToast}
                      />
                    );
                  })}
                {comments && total > comments.length && (
                  <Stack
                    direction="row"
                    alignItems="center"
                    justifyContent="center"
                    sx={{
                      paddingTop: '5px',
                      width: '100%',
                    }}
                  >
                    <Pagination
                      size="small"
                      page={pageNumber}
                      defaultPage={pageNumber}
                      onChange={(e, page) => {
                        getComments(comment.id, page - 1, pageSize);
                        setPageNumber(page);
                      }}
                      count={
                        total % pageSize > 0
                          ? (total - (total % pageSize)) / pageSize + 1
                          : total / pageSize
                      }
                      variant="outlined"
                      shape="rounded"
                    />
                  </Stack>
                )}
              </Collapse>
            </CardActions>
          </>
        )}
      </FadeOutBgcolorCard>

      {/* new a comment dialog */}
      <CustomNewCommentDialog
        open={openDialog}
        setOpen={setOpenDialog}
        postID={comment.postID}
        rootCommentID={
          comment.rootCommentID ? comment.rootCommentID : comment.id
        }
        atWho={atWho}
        atWhoString={atWhoString}
        successAction={() => {
          setAtWho(0);
          if (comment.rootCommentID == 0) {
            comment.replies++;
          } else {
            if (comments) {
              for (let i = 0; i < comments?.length; i++) {
                if (comment.rootCommentID == comments[i].id) {
                  comments[i].replies = comments[i].replies + 1;
                  setComments(comments);
                  break;
                }
              }
            }
          }
          getComments(comment.id, pageNumber - 1, pageSize);
        }}
      />

      {/* delete comment dialog */}
      <Dialog
        maxWidth="lg"
        fullWidth
        open={openDeleteDialog}
        onClose={handleCloseDeleteDialog}
      >
        <DialogContent>
          <Stack direction="row" alignItems="center">
            <Typography variant="h4">
              Are you sure to delete the comment?
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              handleCloseDeleteDialog();
            }}
          >
            Close
          </Button>
          <Button
            onClick={() => {
              deleteComment();
              setOpenDeleteDialog(false);
            }}
          >
            Confirm
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default function CommentComponent({
  anchor,
  anchorID,
  anchorInfo,
  siteID,
  categoryInfo,
  isSiteAdmin,
  postID,
  originComments,
  originTotal,
}: {
  anchor: null | string;
  anchorID: null | number;
  anchorInfo: null | AnchorInfo;
  siteID: number;
  categoryInfo: null | Category;
  isSiteAdmin: boolean;
  postID: number;
  originComments: Comment[] | null;
  originTotal: number;
}) {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [pageNumber, setPageNumber] = React.useState(1); // ui from 1, but api from 0;
  const [pageSize, setPageSize] = React.useState(20);
  const [comments, setComments] = React.useState<Comment[] | null>(
    originComments
  );
  const [total, setTotal] = React.useState(originTotal);

  const getComments = async (pageNumber: number, pageSize: number) => {
    await httpPost(
      false,
      '/invoker/site/getComments/v1',
      setLoading,
      {
        postID,
        pageSize,
        pageNumber,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setComments(respData.data.list);
          setTotal(respData.data.total);
        }
      },
      undefined,
      setErrorToast
    );
  };
  const [openDialog, setOpenDialog] = React.useState(false);

  React.useEffect(() => {
    setComments(originComments);
    setTotal(originTotal);
    if (anchorID && anchorInfo) {
      setPageNumber(anchorInfo.firstLevelPageNumber + 1);
    }
  }, []);

  return (
    <>
      <Grid container spacing={1} sx={{ width: '100%' }}>
        <Grid
          size={12}
          sx={{
            marginTop: '10px',
            display: 'flex',
            flexDirection: 'row-reverse',
          }}
        >
          {getLocalUserLoginState() && (
            <Button
              onClick={() => {
                setOpenDialog(true);
              }}
              variant="contained"
            >
              New A Comment
            </Button>
          )}
        </Grid>
        {/* comment list */}
        {comments &&
          comments.map((item, idx) => {
            return (
              <Grid key={idx} size={12}>
                <CommentCard
                  anchor={anchor}
                  anchorID={anchorID}
                  anchorInfo={anchorInfo}
                  siteID={siteID}
                  categoryInfo={categoryInfo}
                  isSiteAdmin={isSiteAdmin}
                  cardSX={{ width: '100%' }}
                  outerComments={comments}
                  setOuterComments={setComments}
                  comment={item}
                  isFirstLevel={true}
                  setLoading={setLoading}
                  setToast={setToast}
                  setErrorToast={setErrorToast}
                />
              </Grid>
            );
          })}
        {total > pageSize && (
          <Grid size={12}>
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
                  getComments(page - 1, pageSize);
                  setPageNumber(page);
                }}
                count={
                  total % pageSize > 0
                    ? (total - (total % pageSize)) / pageSize + 1
                    : total / pageSize
                }
                variant="outlined"
                shape="rounded"
              />
            </Stack>
          </Grid>
        )}
      </Grid>

      {/* new a comment dialog */}
      <CustomNewCommentDialog
        open={openDialog}
        setOpen={setOpenDialog}
        postID={postID}
        rootCommentID={0}
        atWho={0}
        atWhoString={''}
        successAction={() => {
          getComments(pageNumber - 1, pageSize);
        }}
      />

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
