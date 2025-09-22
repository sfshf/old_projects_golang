'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { useEffect } from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '@/app/shared';
import Stack from '@mui/material/Stack';
import TextField, { TextFieldVariants } from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import FormControl from '@mui/material/FormControl';
import IconButton from '@mui/material/IconButton';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import CloseIcon from '@mui/icons-material/Close';
import CheckIcon from '@mui/icons-material/Check';
import Divider from '@mui/material/Divider';
import InputLabel from '@mui/material/InputLabel';
import FormHelperText from '@mui/material/FormHelperText';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

const FormBoard = ({
  forms,
  formIndex,
  setForms,
  partOfSpeech,
}: {
  forms: any[];
  formIndex: number;
  setForms: (forms: any[]) => void;
  partOfSpeech: any;
}) => {
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

  const [writingForm, setWritingForm] = React.useState(false);
  const onClickConfirmUpdate = () => {
    if (form === '') {
      setFormError(true);
      return;
    }
    if (formString === '') {
      setFormStringError(true);
      return;
    }

    let tmpForms = [...forms];
    if (formIndex === -1) {
      tmpForms.push({
        form: form,
        formString: formString,
        pronunciation: pronunciation,
      });
      setFormIdx(tmpForms.length - 1);
    } else {
      tmpForms[formIndex].form = form;
      tmpForms[formIndex].formString = formString;
      tmpForms[formIndex].pronunciation = pronunciation;
    }
    setForms(tmpForms);
    setWritingForm(false);
  };
  const onClickConfirmDelete = () => {
    let tmpForms = [...forms];
    tmpForms.splice(formIdx, 1);
    setForms(tmpForms);
  };

  const [formIdx, setFormIdx] = React.useState(formIndex);
  const [form, setForm] = React.useState(
    forms && formIndex > -1 ? forms[formIndex].form : ''
  );
  const [formString, setFormString] = React.useState(
    forms && formIndex > -1 ? forms[formIndex].formString : ''
  );
  const [pronunciation, setPronunciation] = React.useState(
    forms && formIndex > -1 ? forms[formIndex].pronunciation : ''
  );
  const [formError, setFormError] = React.useState(false);
  const [formStringError, setFormStringError] = React.useState(false);

  React.useEffect(() => {
    setForm(forms && formIndex > -1 ? forms[formIndex].form : '');
    setFormString(forms && formIndex > -1 ? forms[formIndex].formString : '');
    setPronunciation(
      forms && formIndex > -1 ? forms[formIndex].pronunciation : ''
    );
  }, [forms, formIndex]);

  return (
    <Stack direction="row" alignItems="center">
      {formIdx < 0 && !writingForm && (
        <Stack
          sx={{
            width: '20%',
          }}
        >
          <IconButton
            onClick={() => {
              setWritingForm(true);
            }}
          >
            <AddIcon color="success" />
          </IconButton>
        </Stack>
      )}

      <Stack
        sx={{
          width: '90%',
        }}
      >
        {(formIdx >= 0 || writingForm) && (
          <>
            <Stack
              direction="row"
              alignItems="center"
              spacing={1}
              sx={{
                m: 1,
              }}
            >
              <Typography
                variant="body1"
                gutterBottom
                sx={{
                  width: '25%',
                }}
              >
                Form:
              </Typography>
            </Stack>
            <Stack
              direction="row"
              alignItems="center"
              spacing={1}
              sx={{
                m: 1,
              }}
            >
              <FormControl sx={{ m: 1, width: '800px' }} error={formError}>
                <InputLabel id="demo-simple-select-error-label">
                  Form
                </InputLabel>
                <Select
                  labelId="demo-simple-select-error-label"
                  id="demo-simple-select-error"
                  defaultValue={formIdx > -1 ? forms[formIdx].form : ''}
                  value={form}
                  label="Form"
                  onChange={(e) => {
                    setForm(e.target.value as string);
                    setFormError(false);
                  }}
                  disabled={!writingForm}
                >
                  {partOfSpeechPredefines(partOfSpeech).map((name) => (
                    <MenuItem key={name} value={name}>
                      {name}
                    </MenuItem>
                  ))}
                </Select>
                {formError && <FormHelperText>Required Field</FormHelperText>}
              </FormControl>
            </Stack>
            <Stack
              direction="row"
              alignItems="center"
              spacing={1}
              sx={{
                m: 1,
              }}
            >
              <Typography
                variant="body1"
                gutterBottom
                sx={{
                  width: '25%',
                }}
              >
                String:
              </Typography>
            </Stack>
            <Stack
              direction="row"
              alignItems="center"
              spacing={1}
              sx={{
                m: 1,
              }}
            >
              <FormControl
                sx={{ m: 1, width: '800px' }}
                error={formStringError}
              >
                <TextField
                  multiline
                  value={formString}
                  onChange={(e) => {
                    setFormString(e.target.value);
                    setFormStringError(false);
                  }}
                  disabled={!writingForm}
                  error={formStringError}
                />
                {formStringError && (
                  <FormHelperText>Required Field</FormHelperText>
                )}
              </FormControl>
            </Stack>
            <Stack
              direction="row"
              alignItems="center"
              spacing={1}
              sx={{
                m: 1,
              }}
            >
              <Typography
                variant="body1"
                gutterBottom
                sx={{
                  width: '25%',
                }}
              >
                Pronunciation:
              </Typography>
            </Stack>
            <Stack
              direction="row"
              alignItems="center"
              spacing={1}
              sx={{
                m: 1,
              }}
            >
              <TextField
                sx={{
                  width: '800px',
                }}
                multiline
                value={pronunciation}
                onChange={(e) => {
                  setPronunciation(e.target.value);
                }}
                disabled={!writingForm}
              />
            </Stack>
          </>
        )}
      </Stack>
      <Stack
        sx={{
          width: '20%',
        }}
      >
        {formIdx >= 0 && !writingForm && (
          <>
            <IconButton
              onClick={() => {
                setWritingForm(true);
              }}
            >
              <EditIcon color="secondary" />
            </IconButton>
            <IconButton onClick={onClickConfirmDelete}>
              <DeleteIcon color="error" />
            </IconButton>
          </>
        )}
        {writingForm && (
          <>
            <IconButton onClick={onClickConfirmUpdate}>
              <CheckIcon color="success" />
            </IconButton>
            <IconButton
              onClick={() => {
                setWritingForm(false);
                setForm(formIdx > -1 ? forms[formIdx].form : '');
                setFormString(formIdx > -1 ? forms[formIdx].formString : '');
                setPronunciation(
                  formIdx > -1 ? forms[formIdx].pronunciation : ''
                );
              }}
            >
              <CloseIcon color="error" />
            </IconButton>
          </>
        )}
      </Stack>
    </Stack>
  );
};

const FieldRow = ({
  field,
  data,
  setData,
  multiline,
  minRows,
  maxRows,
  variant,
  setHasExample,
  type,
  setType,
  partOfSpeech,
  setPartOfSpeech,
}: {
  field: string;
  data: any;
  setData: (data: any) => void;
  multiline?: boolean;
  minRows?: number;
  maxRows?: number;
  variant?: TextFieldVariants;
  setHasExample?: (has: boolean) => void;
  type?: any;
  setType?: (type: string) => void;
  partOfSpeech?: any;
  setPartOfSpeech?: (partOfSpeech: string) => void;
}) => {
  const labels = (field: string): string => {
    switch (field) {
      case 'cefr_level':
        return 'Cefr Level';
      case 'string':
        return 'String';
      case 'type':
        return 'Type';
      case 'definition':
        return 'Definition';
      case 'part_of_speech':
        return 'Part of speech';
      case 'specific_type':
        return 'Specific type';
      case 'pronunciation_ipa':
        return 'Pronunciation ipa';
      case 'pronunciation_ipa_weak':
        return 'Pronunciation ipa weak';
      case 'pronunciation_ipa_other':
        return 'Pronunciation ipa other';
      case 'forms':
        return 'Forms';
      case 'pronunciation_text':
        return 'Pronunciation text';
      case 'example_1':
        return 'Example 1';
      case 'example_2':
        return 'Example 2';
      case 'example_3':
        return 'Example 3';
      case 'sort_value':
        return 'SortValue';
      case 'definition_comment':
        return 'Comment';
      default:
        return '';
    }
  };
  const typeList = ['word', 'phrase'];
  const posList = [
    'determiner',
    'adverb',
    'preposition',
    'noun',
    'conjunction',
    'verb',
    'pronoun',
    'adjective',
    'phrase',
  ];

  const dataField = (field: string): string => {
    let key: string = field;
    switch (field) {
      case 'string':
        return 'string';
      case 'book_id':
        return 'bookID';
      case 'cefr_level':
        return 'cefrLevel';
      case 'type':
        return 'type';
      case 'definition':
        return 'definition';
      case 'part_of_speech':
        return 'partOfSpeech';
      case 'specific_type':
        return 'specificType';
      case 'pronunciation_ipa':
        return 'pronunciationIpa';
      case 'pronunciation_ipa_weak':
        return 'pronunciationIpaWeak';
      case 'pronunciation_ipa_other':
        return 'pronunciationIpaOther';
      case 'forms':
        return 'forms';
      case 'pronunciation_text':
        return 'pronunciationText';
      case 'example_1':
        return 'example1';
      case 'position_1':
        return 'position1';
      case 'example_2':
        return 'example2';
      case 'position_2':
        return 'position2';
      case 'example_3':
        return 'example3';
      case 'position_3':
        return 'position3';
      case 'sort_value':
        return 'sortValue';
      case 'definition_comment':
        return 'definitionComment';
      default:
        return '';
    }
  };
  const [txt, setTxt] = React.useState(data[dataField(field)]);
  const setField = (field: string, val: string) => {
    setTxt(val);
    if (setHasExample) {
      if (val) {
        setHasExample(true);
      } else {
        setHasExample(false);
      }
    }
    if (setType) {
      setType(val);
    }
    if (setPartOfSpeech) {
      setPartOfSpeech(val);
    }
    if (field === 'sort_value') {
      // update data state
      data[dataField(field)] = parseInt(val);
      setData(data);
    } else {
      data[dataField(field)] = val;
      setData(data);
    }
  };

  const isExample = (field: string) => {
    return (
      field === 'example_1' || field == 'example_2' || field == 'example_3'
    );
  };
  const positionLabel = (field: string) => {
    switch (field) {
      case 'example_1':
        return 'Positions 1';
      case 'example_2':
        return 'Positions 2';
      case 'example_3':
        return 'Positions 3';
      default:
        return 'Invalid Example Filed';
    }
  };
  const positionField = (field: string) => {
    switch (field) {
      case 'example_1':
        return 'positions1';
      case 'example_2':
        return 'positions2';
      case 'example_3':
        return 'positions3';
      default:
        return 'Invalid Example Filed';
    }
  };
  const [position, setPosition] = React.useState(data[positionField(field)]);
  const positionStringToNumbers = (positions: string): number[] | null => {
    if (!positions) {
      return null;
    }
    let positionNumbers = positions.split(',').map((item) => Number(item));
    let positionIndexs: number[] = [];
    for (let i = 0; i < positionNumbers.length; i += 2) {
      positionIndexs.push(positionNumbers[i]);
      positionIndexs.push(positionNumbers[i + 1]);
    }
    return positionIndexs;
  };
  const [positionIndexs, setPositionIndexs] = React.useState<number[] | null>(
    positionStringToNumbers(data[positionField(field)])
  );
  const exampleResultField = (field: string) => {
    switch (field) {
      case 'example_1':
        return 'Example1 Result';
      case 'example_2':
        return 'Example2 Result';
      case 'example_3':
        return 'Example3 Result';
      default:
        return 'Invalid Example Filed';
    }
  };

  const [openFormView, setOpenFormView] = React.useState(false);
  const handleFormViewDialogClose = () => {
    setOpenFormView(false);
  };
  const [forms, setForms] = React.useState<any[]>([]);

  return (
    <>
      <Stack direction="row" alignItems="center">
        {/* label */}
        <Typography variant="subtitle2" gutterBottom sx={{ width: '40%' }}>
          {labels(field)}:
        </Typography>
        {/* input */}
        {field === 'type' && (
          <FormControl fullWidth>
            <Select
              labelId="demo-simple-select-label"
              id="demo-simple-select"
              value={txt}
              onChange={(e) => {
                setField(field, e.target.value);
              }}
            >
              {typeList &&
                typeList.map((item) => {
                  return (
                    <MenuItem key={item} value={item}>
                      {item}
                    </MenuItem>
                  );
                })}
            </Select>
          </FormControl>
        )}
        {field === 'part_of_speech' && (
          <FormControl fullWidth>
            <Select
              labelId="demo-simple-select-label"
              id="demo-simple-select"
              value={type === 'phrase' ? 'phrase' : txt}
              onChange={(e) => {
                setField(field, e.target.value);
              }}
            >
              {posList.map((item) => {
                if (type === 'word' && item !== 'phrase') {
                  return (
                    <MenuItem key={item} value={item}>
                      {item}
                    </MenuItem>
                  );
                } else if (type === 'phrase' && item === 'phrase') {
                  return (
                    <MenuItem key={item} value={item}>
                      {item}
                    </MenuItem>
                  );
                }
              })}
            </Select>
          </FormControl>
        )}
        {field === 'forms' && (
          <Button
            size="large"
            onClick={() => {
              setOpenFormView(true);
              data[dataField(field)] = forms;
              setData(data);
            }}
          >
            Forms View
          </Button>
        )}
        {field !== 'type' &&
          field !== 'part_of_speech' &&
          field !== 'forms' && (
            <TextField
              variant={variant ? variant : 'standard'}
              minRows={minRows}
              maxRows={maxRows}
              multiline={multiline}
              value={txt}
              fullWidth
              onChange={(e) => {
                setField(field, e.target.value);
              }}
            />
          )}
      </Stack>
      {/* example positions and view */}
      {isExample(field) && (
        <>
          <Stack direction="row" alignItems="center">
            <Typography variant="subtitle2" gutterBottom sx={{ width: '40%' }}>
              {positionLabel(field)}:
            </Typography>
            <TextField
              variant={variant ? variant : 'standard'}
              minRows={minRows}
              maxRows={maxRows}
              multiline={multiline}
              value={position}
              fullWidth
              onChange={(e) => {
                setPosition(e.target.value);
                setPositionIndexs(positionStringToNumbers(e.target.value));
                // update data state
                data[positionField(field)] = e.target.value;
                setData(data);
              }}
            />
          </Stack>
          <Stack direction="row" alignItems="center">
            <Typography variant="subtitle2" gutterBottom sx={{ width: '40%' }}>
              {exampleResultField(field)}:
            </Typography>
            <Typography
              variant="body1"
              sx={{ whiteSpace: 'pre-wrap' }}
              gutterBottom
            >
              {positionIndexs &&
                txt.split('').map((item: any, index: any) => {
                  for (let j = 0; j < positionIndexs.length; j += 2) {
                    if (
                      index >= positionIndexs[j] &&
                      index < positionIndexs[j] + positionIndexs[j + 1]
                    ) {
                      return (
                        <span
                          key={index}
                          style={{
                            color: 'red',
                          }}
                        >
                          {item}
                        </span>
                      );
                    }
                  }
                  return <span key={index}>{item}</span>;
                })}
            </Typography>
          </Stack>
        </>
      )}
      {/* forms dialog */}
      <Dialog
        fullWidth
        maxWidth="md"
        open={openFormView}
        onClose={handleFormViewDialogClose}
      >
        <IconButton
          aria-label="close"
          onClick={() => {
            setOpenFormView(false);
          }}
          sx={{
            position: 'absolute',
            right: 8,
            top: 8,
            color: (theme) => theme.palette.grey[500],
          }}
        >
          <CloseIcon />
        </IconButton>
        <DialogContent>
          <Stack
            direction="row"
            alignItems="center"
            spacing={2}
            sx={{
              m: 2,
            }}
          >
            <Typography variant="h6" gutterBottom>
              Forms:
            </Typography>
          </Stack>
          {forms.length > 0 &&
            forms.map((form, index) => {
              if (index + 1 < forms.length) {
                return (
                  <>
                    <FormBoard
                      forms={forms}
                      formIndex={index}
                      setForms={setForms}
                      partOfSpeech={partOfSpeech}
                    />
                    <Divider variant="middle" />
                  </>
                );
              } else {
                return (
                  <>
                    <FormBoard
                      forms={forms}
                      formIndex={index}
                      setForms={setForms}
                      partOfSpeech={partOfSpeech}
                    />
                    <Divider variant="middle" />
                    <FormBoard
                      forms={forms}
                      formIndex={-1}
                      setForms={setForms}
                      partOfSpeech={partOfSpeech}
                    />
                  </>
                );
              }
            })}

          {forms.length === 0 && (
            <FormBoard
              forms={forms}
              formIndex={-1}
              setForms={setForms}
              partOfSpeech={partOfSpeech}
            />
          )}
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            onClick={() => {
              // update data state
              data[dataField(field)] = forms;
              setData(data);
              setOpenFormView(false);
            }}
          >
            OK
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [bookID, setBookID] = React.useState('');
  const [cefrLevel, setCefrLevel] = React.useState('');
  const [data, setData] = React.useState<any>({
    bookID: 0,
    cefrLevel: '',
    string: '',
    type: '',
    sortValue: 0,
    definition: '',
    partOfSpeech: '',
    specificType: '',
    pronunciationIpa: '',
    pronunciationIpaWeak: '',
    pronunciationIpaOther: '',
    forms: [],
    pronunciationText: '',
    example1: '',
    positions1: '',
    example2: '',
    positions2: '',
    example3: '',
    positions3: '',
  });

  const [openCommit, setOpenCommit] = React.useState(false);

  const handleCommitDialogClose = () => {
    setOpenCommit(false);
  };

  const onClickConfirmCommit = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    let postData: any = JSON.stringify({ ...data, password });
    let reqOpts: any = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
      },
      body: postData,
    };
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/book/new_definition`,
        reqOpts
      );
      setLoading(false);
      const resData = await res.json();
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setOpenCommit(false);
      setToast('success');
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const [cefrLevels, setCefrLevels] = React.useState<any[]>([]);
  const fetchCefrLevel = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/book/cefr_levels?` +
          new URLSearchParams({
            password,
          }).toString()
      );
      setLoading(false);
      const respData = await res.json();
      if (respData.code !== 0) {
        setErrorToast(respData.message);
        return;
      }
      setCefrLevels(respData.data.list);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const [hasExample1, setHasExample1] = React.useState(false);
  const [hasExample2, setHasExample2] = React.useState(false);

  const [type, setType] = React.useState('');
  const [partOfSpeech, setPartOfSpeech] = React.useState('');

  useEffect(() => {
    fetchCefrLevel();
  }, []);

  return (
    <Container
      maxWidth="md"
      sx={{ justifyContent: 'center', marginTop: '50px', marginBottom: '60px' }}
    >
      <Stack spacing={4} width="100%">
        <Stack direction="row" alignItems="center">
          <Typography variant="subtitle2" gutterBottom sx={{ width: '40%' }}>
            Cefr Level:
          </Typography>
          <FormControl fullWidth>
            <Select
              labelId="demo-simple-select-label"
              id="demo-simple-select"
              value={cefrLevel}
              onChange={(e) => {
                setCefrLevel(e.target.value);
                let i = 0;
                for (; i < cefrLevels.length; i++) {
                  if (cefrLevels[i].level === e.target.value) {
                    setBookID(cefrLevels[i].bookID);
                    break;
                  }
                }
                // update data state
                data.bookID = cefrLevels[i].bookID;
                data.cefrLevel = e.target.value;
                setData(data);
              }}
            >
              {cefrLevels &&
                cefrLevels.map((item) => {
                  return (
                    <MenuItem key={item.bookID} value={item.level}>
                      {item.level}
                    </MenuItem>
                  );
                })}
            </Select>
          </FormControl>
        </Stack>
        {bookID && (
          <>
            <FieldRow field="string" data={data} setData={setData} />
            <FieldRow
              field="type"
              data={data}
              setData={setData}
              setType={setType}
            />
            <FieldRow
              multiline
              maxRows={10}
              minRows={3}
              variant="outlined"
              field="definition"
              data={data}
              setData={setData}
            />
            <FieldRow
              field="part_of_speech"
              data={data}
              setData={setData}
              type={type}
              setPartOfSpeech={setPartOfSpeech}
            />
            <FieldRow field="specific_type" data={data} setData={setData} />
            {type === 'word' && (
              <>
                <FieldRow
                  field="pronunciation_ipa"
                  data={data}
                  setData={setData}
                />
                <FieldRow
                  field="pronunciation_ipa_weak"
                  data={data}
                  setData={setData}
                />
                <FieldRow
                  field="pronunciation_ipa_other"
                  data={data}
                  setData={setData}
                />
                <FieldRow
                  field="forms"
                  data={data}
                  setData={setData}
                  partOfSpeech={partOfSpeech}
                />
              </>
            )}
            {type === 'phrase' && (
              <FieldRow
                field="pronunciation_text"
                data={data}
                setData={setData}
              />
            )}
            <FieldRow
              field="example_1"
              data={data}
              setData={setData}
              setHasExample={setHasExample1}
            />
            {hasExample1 && (
              <FieldRow
                field="example_2"
                data={data}
                setData={setData}
                setHasExample={setHasExample2}
              />
            )}
            {hasExample2 && (
              <FieldRow field="example_3" data={data} setData={setData} />
            )}
            <FieldRow field="sort_value" data={data} setData={setData} />
            <Button
              variant="contained"
              onClick={() => {
                setOpenCommit(true);
              }}
            >
              Commit
            </Button>
          </>
        )}
      </Stack>

      <Dialog
        fullWidth
        maxWidth="sm"
        open={openCommit}
        onClose={handleCommitDialogClose}
      >
        <DialogContent>
          <Typography variant="h6" gutterBottom>
            Confirm to commit the new definition ?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="error"
            onClick={() => {
              setOpenCommit(false);
            }}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            color="success"
            onClick={onClickConfirmCommit}
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
