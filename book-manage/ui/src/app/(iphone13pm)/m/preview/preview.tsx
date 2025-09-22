'use client';
import Backdrop from '@mui/material/Backdrop';
import React from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '@/app/shared';
import Stack from '@mui/material/Stack';
import { useSearchParams } from 'next/navigation';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import { request } from 'http';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Checkbox from '@mui/material/Checkbox';
import {
  FieldRow,
  Data,
  UpdatePreviewRequest,
  DeletePreviewRequest,
} from './row';
// VolumeUp

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

export default function Preview({
  read_only,
  read_book_id,
  read_definition_index,
}: {
  read_only: boolean;
  read_book_id?: number;
  read_definition_index?: number;
}) {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');
  const baseURL = shared.baseAPIURL;

  const searchParams = useSearchParams();
  const [bookID, setBookID] = React.useState(searchParams.get('bookID') || '');
  const [index, setIndex] = React.useState(-1);
  const [jumpTo, setJumpTo] = React.useState<number | undefined>(undefined);
  const [count, setCount] = React.useState(0);
  const [data, setData] = React.useState<Data | null>(null);
  const [audioPlaying, setAudioPlaying] = React.useState(false);
  const [withComment, setWithComment] = React.useState(false);

  async function fetchData(bookID: string, index: number) {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(
        `${baseURL}/book/preview?` +
          new URLSearchParams({
            password,
            bookID,
            index: String(index),
            withComment: String(withComment),
          }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        // console.log(data.message)
        setErrorToast(data.message);
        setData(null);
        return;
      }
      // console.log(data)
      setData({
        book_id: parseInt(bookID, 10),
        string_id: data.data.item.definition.string_id,
        string: data.data.item.string,
        type: data.data.item.type,
        sort_value: data.data.item.sortValue,
        definition_id: data.data.item.definition.id,
        definition: data.data.item.definition.definition,
        definition_translations: data.data.item.definitionTranslations,
        part_of_speech: data.data.item.definition.part_of_speech,
        specific_type: data.data.item.definition.specific_type,
        pronunciation_ipa: data.data.item.definition.pronunciation_ipa,
        pronunciation_ipa_weak:
          data.data.item.definition.pronunciation_ipa_weak,
        pronunciation_ipa_other:
          data.data.item.definition.pronunciation_ipa_other,
        pronunciation_text: data.data.item.definition.pronunciation_text,
        cefr_level: data.data.item.definition.cefr_level,
        example1_id: data.data.item.examples[0]
          ? data.data.item.examples[0].id
          : 0,
        example1: data.data.item.examples[0]
          ? data.data.item.examples[0].content
          : '',
        example2_id: data.data.item.examples[1]
          ? data.data.item.examples[1].id
          : 0,
        example2: data.data.item.examples[1]
          ? data.data.item.examples[1].content
          : '',
        example3_id: data.data.item.examples[2]
          ? data.data.item.examples[2].id
          : 0,
        example3: data.data.item.examples[2]
          ? data.data.item.examples[2].content
          : '',
        positions1: data.data.item.examples[0]
          ? data.data.item.examples[0].word_positions
          : '',
        positions2: data.data.item.examples[1]
          ? data.data.item.examples[1].word_positions
          : '',
        positions3: data.data.item.examples[2]
          ? data.data.item.examples[2].word_positions
          : '',
        example_translations: data.data.item.exampleTranslations,
        relatedForms: data.data.item.relatedForms,
        definition_comment_id: data.data.item.definitionComment.id,
        definition_comment: data.data.item.definitionComment.content,
        index: data.data.index,
      });
      setCount(data.data.total);
      setIndex(data.data.index);
    } catch (error) {
      // console.log(error)
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
      setData(null);
    }
  }

  const [openDelete, setOpenDelete] = React.useState(false);

  const handleDeleteDialogClose = () => {
    setOpenDelete(false);
  };

  const onClickConfirmDelete = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: DeletePreviewRequest = {
      password: password,
      bookID: data!.book_id,
      definitionID: data!.definition_id,
      field: 'preview_item',
    };

    try {
      const postData = JSON.stringify(reqData);
      const req = request(
        {
          path: shared.baseAPIURL + '/book/delete_preview',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData),
          },
        },
        (res) => {
          res.setEncoding('utf8');
          res.on('data', (chunk) => {
            const respData = JSON.parse(chunk);
            if (respData.code !== 0) {
              setErrorToast(respData.message);
              return;
            }
            fetchData(bookID, index);
            setToast('success');
            setOpenDelete(false);
          });
        }
      );
      req.on('error', (e) => {
        throw e;
      });
      req.write(postData);
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const ITEM_HEIGHT = 48;
  const ITEM_PADDING_TOP = 8;
  const MenuProps = {
    PaperProps: {
      style: {
        maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
        width: 250,
      },
    },
  };
  const partOfSpeechPredefines = (partOfSpeech: string) => {
    let res: string[] = ['also'];
    partOfSpeech.split(',').map((val) => {
      switch (val) {
        case 'verb':
          res.push(
            'present simple',
            'present participle',
            'past simple',
            'past participle'
          );
          break;
        case 'noun':
          res.push('plural');
          break;
        case 'adjective':
          res.push('comparative', 'superlative');
          break;
        case 'adverb':
          res.push('comparative', 'superlative');
          break;
      }
    });
    return res.sort();
  };

  const [openUpdateForm, setOpenUpdateForm] = React.useState(false);
  const [updateFormData, setUpdateFormData] =
    React.useState<UpdatePreviewRequest | null>(null);
  const [needToRefresh, setNeedToRefresh] = React.useState(false);

  const handleUpdateFormDialogClose = () => {
    setOpenUpdateForm(false);
  };

  const onClickConfirmForm = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: UpdatePreviewRequest = {
      password: password,
      bookID: data!.book_id,
      definitionID: data!.definition_id,
      field: 'form',

      formStringID: updateFormData?.formStringID,
      form: updateFormData?.form,
      formString: updateFormData?.formString,
      pronunciation: updateFormData?.pronunciation,
    };

    try {
      const postData = JSON.stringify(reqData);
      const req = request(
        {
          path: shared.baseAPIURL + '/book/update_preview',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData),
          },
        },
        (res) => {
          res.setEncoding('utf8');
          res.on('data', (chunk) => {
            const respData = JSON.parse(chunk);
            if (respData.code !== 0) {
              setErrorToast(respData.message);
              return;
            }
            setToast('success');
            fetchData(bookID, index);
            setOpenUpdateForm(false);
          });
        }
      );
      req.on('error', (e) => {
        throw e;
      });
      req.write(postData);
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const fetchDefinition = async (definitionID: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(
        shared.baseAPIURL +
          '/book/definition_info?' +
          new URLSearchParams({
            password,
            definitionID,
          }).toString()
      );
      const respData = await res.json();
      if (respData.code !== 0) {
        setErrorToast(respData.message);
        return;
      }
      if (index == -1) {
        setBookID(respData.data.bookID);
        setIndex(respData.data.index);
        fetchData(respData.data.bookID, respData.data.index);
      }
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  React.useEffect(() => {
    let definitionIDQuery = searchParams.get('definitionID');
    if (definitionIDQuery) {
      fetchDefinition(definitionIDQuery);
    }

    if (needToRefresh) {
      fetchData(bookID, index);
      setNeedToRefresh(false);
    }
    if (read_only) {
      setBookID(read_book_id!.toString(10));
      fetchData(read_book_id!.toString(10), read_definition_index!);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    needToRefresh,
    read_only,
    read_book_id,
    read_definition_index,
    bookID,
    index,
  ]);

  return (
    <Container>
      <Stack
        spacing={2}
        width="100%"
        sx={{
          alignItems: 'center',
          padding: '5px',
        }}
      >
        {data && (
          <Stack spacing={1} width="100%">
            <FieldRow
              read_only={read_only}
              field="string"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              audioPlaying={audioPlaying}
              setAudioPlaying={setAudioPlaying}
            />
            <FieldRow
              read_only={read_only}
              field="type"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="definition"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              formIndex={0}
              setOpenUpdateForm={setOpenUpdateForm}
              setUpdateFormData={setUpdateFormData}
              setNeedToRefresh={setNeedToRefresh}
            />
            <FieldRow
              read_only={read_only}
              field="part_of_speech"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              formIndex={0}
              setOpenUpdateForm={setOpenUpdateForm}
              setUpdateFormData={setUpdateFormData}
            />
            <FieldRow
              read_only={read_only}
              field="specific_type"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="pronunciation_ipa"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              audioPlaying={audioPlaying}
              setAudioPlaying={setAudioPlaying}
            />
            <FieldRow
              read_only={read_only}
              field="pronunciation_ipa_weak"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              audioPlaying={audioPlaying}
              setAudioPlaying={setAudioPlaying}
            />
            <FieldRow
              read_only={read_only}
              field="pronunciation_ipa_other"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              audioPlaying={audioPlaying}
              setAudioPlaying={setAudioPlaying}
            />
            <FieldRow
              read_only={read_only}
              field="pronunciation_text"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              audioPlaying={audioPlaying}
              setAudioPlaying={setAudioPlaying}
            />
            <FieldRow
              read_only={read_only}
              field="cefr_level"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              refreshData={fetchData}
            />
            {data.relatedForms &&
              data.relatedForms.length > 0 &&
              data.relatedForms.map((item, index) => {
                return (
                  <FieldRow
                    read_only={read_only}
                    key={index}
                    field="form"
                    data={data}
                    setData={setData}
                    setLoading={setLoading}
                    setToast={setToast}
                    setErrorToast={setErrorToast}
                    audioPlaying={audioPlaying}
                    setAudioPlaying={setAudioPlaying}
                    formIndex={index}
                    setOpenUpdateForm={setOpenUpdateForm}
                    setUpdateFormData={setUpdateFormData}
                    setNeedToRefresh={setNeedToRefresh}
                  />
                );
              })}
            {data.type !== 'phrase' && !read_only && (
              <Button
                variant="contained"
                size="small"
                sx={{
                  width: '150px',
                  height: '25px',
                }}
                onClick={() => {
                  setOpenUpdateForm(true);
                  setUpdateFormData({
                    password: '',
                    definitionID: 0,
                    bookID: 0,
                    field: 'form',
                    formStringID: 0,
                    form: '',
                    formString: '',
                    pronunciation: '',
                  });
                }}
              >
                Add a form
              </Button>
            )}
            <FieldRow
              read_only={read_only}
              field="example_1"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              setNeedToRefresh={setNeedToRefresh}
            />
            <FieldRow
              read_only={read_only}
              field="position_1"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="example_2"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              setNeedToRefresh={setNeedToRefresh}
            />
            <FieldRow
              read_only={read_only}
              field="position_2"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="example_3"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              setNeedToRefresh={setNeedToRefresh}
            />
            <FieldRow
              read_only={read_only}
              field="position_3"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="sort_value"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="definition_comment"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="quick_search"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
            <FieldRow
              read_only={read_only}
              field="definition_operatelog"
              data={data}
              setData={setData}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
            />
          </Stack>
        )}
        {data && !read_only && (
          <>
            <Stack direction="row" spacing={3} alignItems="center">
              <Button
                disabled={index === 0}
                variant="contained"
                size="small"
                sx={{ width: '50px', height: '25px' }}
                onClick={() => fetchData(bookID, index - 1)}
              >
                Pre
              </Button>
              <Typography variant="h5" gutterBottom fontSize={13}>
                {index}/{count - 1}
              </Typography>
              <Button
                disabled={count === 0 || index === count - 1}
                variant="contained"
                size="small"
                sx={{ width: '50px', height: '25px' }}
                onClick={() => fetchData(bookID, index + 1)}
              >
                Next
              </Button>
            </Stack>
            <Stack direction="row" spacing={3} alignItems="center">
              <TextField
                label="Jump To"
                variant="outlined"
                value={jumpTo}
                placeholder="index start from 0"
                onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                  setJumpTo(
                    event.target.value === ''
                      ? undefined
                      : Number(event.target.value)
                  );
                }}
              />
              <Button
                disabled={jumpTo === undefined}
                variant="contained"
                size="small"
                sx={{ width: '50px', height: '25px' }}
                onClick={() => fetchData(bookID, jumpTo!)}
              >
                Jump
              </Button>
              <Button
                variant="contained"
                size="small"
                sx={{ width: '50px', height: '25px' }}
                color="warning"
                onClick={() => {
                  setOpenDelete(true);
                }}
              >
                Delete
              </Button>
            </Stack>
          </>
        )}
        {!read_only && (
          <>
            <Stack direction="row" spacing={3} alignItems="center">
              <TextField
                required
                label="BookID"
                value={bookID}
                onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                  setBookID(event.target.value);
                }}
                variant="standard"
              />
              <Button
                disabled={bookID.length < 9}
                variant="contained"
                size="small"
                sx={{ width: '50px', height: '25px' }}
                onClick={() => fetchData(bookID, 0)}
              >
                Refresh
              </Button>
            </Stack>
            <Stack direction="row" spacing={3} alignItems="center">
              <FormGroup>
                <FormControlLabel
                  control={
                    <Checkbox
                      size="medium"
                      checked={withComment}
                      onChange={(e) => {
                        setWithComment(e.target.checked);
                      }}
                    />
                  }
                  label="With Comment"
                />
              </FormGroup>
            </Stack>
          </>
        )}
      </Stack>

      {/* update form dialog */}
      {data && (
        <Dialog
          fullWidth
          maxWidth="md"
          open={openUpdateForm}
          onClose={handleUpdateFormDialogClose}
        >
          <DialogContent>
            <Typography variant="h6" gutterBottom>
              Properties of form:
            </Typography>
            <Stack spacing={1} direction="row" alignItems="center">
              <Typography
                variant="body1"
                gutterBottom
                fontSize={13}
                sx={{
                  width: '50%',
                }}
              >
                Form:
              </Typography>
              <Select
                value={updateFormData?.form}
                onChange={(e) => {
                  setUpdateFormData({
                    password: '',
                    definitionID: 0,
                    bookID: 0,
                    field: 'form',
                    formStringID: updateFormData?.formStringID,
                    form: e.target.value,
                    formString: updateFormData?.formString,
                    pronunciation: updateFormData?.pronunciation,
                  });
                }}
                MenuProps={MenuProps}
                fullWidth
              >
                {partOfSpeechPredefines(data.part_of_speech).map((name) => (
                  <MenuItem key={name} value={name}>
                    {name}
                  </MenuItem>
                ))}
              </Select>
            </Stack>
            <Stack spacing={1} direction="row" alignItems="center">
              <Typography
                variant="body1"
                gutterBottom
                fontSize={13}
                sx={{
                  width: '50%',
                }}
              >
                String:
              </Typography>
              <TextField
                fullWidth
                defaultValue={updateFormData?.formString}
                onChange={(e) => {
                  let tmpData = updateFormData;
                  tmpData!.formString = e.target.value;
                  setUpdateFormData(tmpData);
                }}
              />
            </Stack>
            <Stack spacing={1} direction="row" alignItems="center">
              <Typography
                variant="body1"
                gutterBottom
                fontSize={13}
                sx={{
                  width: '50%',
                }}
              >
                PronunciationIPA:
              </Typography>
              <TextField
                fullWidth
                defaultValue={updateFormData?.pronunciation}
                onChange={(e) => {
                  let tmpData = updateFormData;
                  tmpData!.pronunciation = e.target.value;
                  setUpdateFormData(tmpData);
                }}
              />
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button
              variant="contained"
              size="small"
              sx={{ width: '50px', height: '25px' }}
              color="error"
              onClick={() => {
                setOpenUpdateForm(false);
                setUpdateFormData({
                  password: '',
                  definitionID: 0,
                  bookID: 0,
                  field: 'form',
                  formStringID: 0,
                  form: '',
                  formString: '',
                  pronunciation: '',
                });
              }}
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              size="small"
              sx={{ width: '50px', height: '25px' }}
              color="success"
              onClick={onClickConfirmForm}
            >
              Confirm
            </Button>
          </DialogActions>
        </Dialog>
      )}

      {/* delete preview dialog */}
      <Dialog
        fullWidth
        maxWidth="sm"
        open={openDelete}
        onClose={handleDeleteDialogClose}
      >
        <DialogContent>
          <Typography variant="h6" gutterBottom fontSize={13}>
            Confirm to delete the preview ?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            size="small"
            sx={{ width: '50px', height: '25px' }}
            color="error"
            onClick={() => {
              setOpenDelete(false);
            }}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            size="small"
            sx={{ width: '50px', height: '25px' }}
            color="success"
            onClick={onClickConfirmDelete}
          >
            Confirm
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={errorToast !== ''}
        autoHideDuration={5000}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
        onClose={() => setErrorToast('')}
      >
        <Alert
          onClose={() => setErrorToast('')}
          severity="error"
          sx={{ width: '100%' }}
        >
          {errorToast}
        </Alert>
      </Snackbar>
      <Snackbar
        open={toast !== ''}
        autoHideDuration={3000}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
        onClose={() => setToast('')}
      >
        <Alert
          onClose={() => setToast('')}
          severity="success"
          sx={{ width: '100%' }}
        >
          {toast}
        </Alert>
      </Snackbar>
      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
    </Container>
  );
}
