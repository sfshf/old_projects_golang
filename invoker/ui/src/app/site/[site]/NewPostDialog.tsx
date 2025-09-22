'use client';
import React from 'react';
import { post, uploadFile } from '@/app/shared';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import { styled } from '@mui/material/styles';
import CloseIcon from '@mui/icons-material/Close';
import Badge from '@mui/material/Badge';
import pica from 'pica';
import { Category } from '@/app/model';

export default function NewPostDialog({
  site_id,
  category_id,
  open,
  handleClose,
  successAction,
  categories,
}: {
  site_id: number;
  category_id: number;
  open: boolean;
  handleClose: (event?: object, reason?: string) => void;
  successAction?: () => void;
  categories?: null | Category[];
}) {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [categoryID, setCategoryID] = React.useState(category_id);
  const handleCategoryChange = (event: SelectChangeEvent) => {
    setCategoryID(parseInt(event.target.value));
  };
  const [title, setTitle] = React.useState('');
  const [content, setContent] = React.useState('');
  const [imgFile, setImgFile] = React.useState<null | File>(null);
  const [imgUrl, setImgUrl] = React.useState<null | string>(null);

  const addPost = () => {
    if (site_id === 0 || categoryID === 0) {
      setErrorToast('invalid site or category');
      return;
    }
    // upload image first, if has
    if (imgFile) {
      uploadFile(
        setLoading,
        imgFile,
        'images',
        (respData: any) => {
          let image =
            'https://d2y6ia7j6nkf8t.cloudfront.net/images/' +
            respData.data.hashName;
          // post addPost
          post(
            false,
            '/invoker/site/addPost/v1',
            setLoading,
            {
              siteID: site_id,
              categoryID,
              title,
              content,
              image,
            },
            (respHeaders: any) => {},
            (respData: any) => {
              setCategoryID(0);
              setTitle('');
              setContent('');
              setImgFile(null);
              setImgUrl('');
              // call successAction, if has
              if (successAction) {
                successAction();
              }
              handleClose();
            },
            setToast,
            setErrorToast
          );
        },
        setToast,
        setErrorToast
      );
    } else {
      // post addPost
      post(
        false,
        '/invoker/site/addPost/v1',
        setLoading,
        {
          siteID: site_id,
          categoryID,
          title,
          content,
        },
        (respHeaders: any) => {},
        (respData: any) => {
          setCategoryID(0);
          setTitle('');
          setContent('');
          setImgFile(null);
          setImgUrl('');
          // call successAction, if has
          if (successAction) {
            successAction();
          }
          handleClose();
        },
        setToast,
        setErrorToast
      );
    }
  };

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

  React.useEffect(() => {
    setCategoryID(category_id);
  }, [site_id, category_id, categories]);

  return (
    <>
      <Dialog maxWidth="lg" fullWidth open={open} onClose={handleClose}>
        <DialogTitle>New Post</DialogTitle>
        <DialogContent>
          {category_id === 0 && categories && (
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                Category
              </Typography>
              <Select
                value={categoryID.toString()}
                onChange={handleCategoryChange}
                fullWidth
              >
                {categories &&
                  categories.map((item) => {
                    return (
                      <MenuItem key={item.id} value={item.id}>
                        {item.name}
                      </MenuItem>
                    );
                  })}
              </Select>
            </Stack>
          )}
          <Stack direction="row" alignItems="center">
            <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
              Title
            </Typography>
            <TextField
              id="title"
              name="title"
              fullWidth
              margin="normal"
              value={title}
              onChange={(e) => {
                setTitle(e.target.value);
              }}
            />
          </Stack>
          <Stack direction="row" alignItems="center">
            <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
              Content
            </Typography>
            <TextField
              id="content"
              name="content"
              fullWidth
              multiline
              minRows={10}
              margin="normal"
              value={content}
              onChange={(e) => {
                setContent(e.target.value);
              }}
            />
          </Stack>
          <Stack direction="row" alignItems="center" marginTop="10px">
            <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
              Image
            </Typography>
            {!imgUrl && (
              <Button
                variant="contained"
                component="label"
                sx={{
                  width: '100px',
                  height: '50px',
                }}
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
            {imgUrl && (
              <Badge
                badgeContent={
                  <IconButton
                    onClick={() => {
                      setImgFile(null);
                      setImgUrl(null);
                    }}
                  >
                    <CloseIcon color="error" />
                  </IconButton>
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
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              handleClose();
              setCategoryID(0);
              setTitle('');
              setContent('');
              setImgFile(null);
              setImgUrl('');
            }}
            variant="contained"
            color="error"
            sx={{ margin: '5px', width: '100px' }}
          >
            Cancel
          </Button>
          <Button
            onClick={() => {
              addPost();
            }}
            variant="contained"
            sx={{ margin: '5px', width: '100px' }}
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
    </>
  );
}
