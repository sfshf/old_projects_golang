'use client';
import React from 'react';
import Button from '@mui/material/Button';
import shared from '@/app/shared';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import VolumeUp from '@mui/icons-material/VolumeUp';
import Downloading from '@mui/icons-material/Downloading';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import { Theme, useTheme } from '@mui/material/styles';
import { request } from 'http';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import CloseIcon from '@mui/icons-material/Close';
import CheckIcon from '@mui/icons-material/Check';
import * as clipboard from 'clipboard-polyfill';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';

export interface Form {
  String: string;
  StringID: number;
  Form: string;
  Definition: string;
  DefinitionID: number;
  Pronunciation: string;
}

export interface Translation {
  id: number;
  item_type: string;
  item_id: number;
  content: string;
  language_code: string;
}

export interface Data {
  book_id: number;
  string_id: number;
  string: string;
  type: string;
  sort_value: number;
  definition_id: number;
  definition: string;
  definition_translations?: Translation[];
  part_of_speech: string;
  specific_type: string;
  pronunciation_ipa: string;
  pronunciation_ipa_weak: string;
  pronunciation_ipa_other: string;
  pronunciation_text: string;
  cefr_level: string;
  example1_id: number;
  example1: string;
  positions1: string;
  example2_id: number;
  example2: string;
  positions2: string;
  example3_id: number;
  example3: string;
  positions3: string;
  example_translations?: Translation[];
  relatedForms?: Form[];
  definition_comment_id: number;
  definition_comment: string;
  index: number;
}

export interface UpdatePreviewRequest {
  password: string;
  definitionID: number;
  bookID: number;
  field?: string;
  stringID?: number;
  string?: string;
  definition?: string;
  partOfSpeech?: string;
  specificType?: string;
  pronunciationIpa?: string;
  pronunciationIpaWeak?: string;
  pronunciationIpaOther?: string;
  pronunciationText?: string;
  exampleID?: number;
  example?: string;
  position?: string;
  definitionCommentID?: number;
  definitionComment?: string;
  formStringID?: number;
  form?: string;
  formString?: string;
  pronunciation?: string;
  sortValue?: number;
  translationID?: number;
  translationContent?: string;
  languageCode?: string;
}

export interface DeletePreviewRequest {
  password: string;
  definitionID: number;
  bookID: number;
  field?: string;
  exampleID?: number;
  definitionCommentID?: number;
  formStringID?: number;
  translationID?: number;
}

const TranslationBoard = ({
  read_only,
  book_id,
  definition_id,
  item_type,
  item_id,
  translations,
  translIndex,
  setTranslations,
  setLoading,
  setToast,
  setErrorToast,
  setNeedToRefresh,
}: {
  read_only: boolean;
  book_id: number;
  definition_id: number;
  item_type: string;
  item_id: number;
  translations: Translation[];
  translIndex: number;
  setTranslations: (transls: Translation[]) => void;
  setLoading: (load: boolean) => void;
  setToast: (message: string) => void;
  setErrorToast: (message: string) => void;
  setNeedToRefresh?: (need: boolean) => void;
}) => {
  React.useEffect(() => {
    setTranslID(
      translations && translIndex > -1 ? translations[translIndex].id : 0
    );
    setLanguageCode(
      translations && translIndex > -1
        ? translations[translIndex].language_code
        : ''
    );
    setTranslContent(
      translations && translIndex > -1 ? translations[translIndex].content : ''
    );
  }, [translations, translIndex]);
  const [translID, setTranslID] = React.useState(
    translations && translIndex > -1 ? translations[translIndex].id : 0
  );
  const [languageCode, setLanguageCode] = React.useState(
    translations && translIndex > -1
      ? translations[translIndex].language_code
      : ''
  );
  const [translContent, setTranslContent] = React.useState(
    translations && translIndex > -1 ? translations[translIndex].content : ''
  );

  const [writingTransl, setWritingTransl] = React.useState(false);

  const languageCodePredefines = ['zh'].sort();

  const onClickConfirmUpdate = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: UpdatePreviewRequest = {
      password: password,
      bookID: book_id,
      definitionID: definition_id,
      translationID: translID,
      translationContent: translContent,
      languageCode: languageCode,
    };
    switch (item_type) {
      case 'definition':
        reqData.field = 'definition_translation';
        break;
      case 'example':
        reqData.exampleID = item_id;
        reqData.field = 'example_translation';
        break;
    }

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

            let tmpTransls = [...translations];
            if (translIndex === -1) {
              let one: Translation = {
                id: respData.data,
                item_type: item_type,
                item_id: item_id,
                content: translContent,
                language_code: languageCode,
              };
              tmpTransls.push(one);
              setTranslID(respData.data);
            } else {
              tmpTransls[translIndex].content = translContent;
              tmpTransls[translIndex].language_code = languageCode;
            }
            setTranslations(tmpTransls);

            setNeedToRefresh!(true);

            setWritingTransl(false);

            setToast('success');
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

  const onClickConfirmDelete = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: DeletePreviewRequest = {
      password: password,
      bookID: book_id,
      definitionID: definition_id,
      translationID: translID,
    };

    switch (item_type) {
      case 'definition':
        reqData.field = 'definition_translation';
        break;
      case 'example':
        reqData.exampleID = item_id;
        reqData.field = 'example_translation';
        break;
    }

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

            let tmpTransls = [...translations];
            tmpTransls.splice(translIndex, 1);
            setTranslations(tmpTransls);

            setNeedToRefresh!(true);

            setToast('success');
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

  return (
    <Stack direction="row" alignItems="center">
      {!read_only && translID === 0 && !writingTransl && (
        <Stack
          sx={{
            width: '20%',
          }}
        >
          <IconButton
            onClick={() => {
              setWritingTransl(true);
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
        {(translID > 0 || writingTransl) && (
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
                Language Code:
              </Typography>
              <Select
                onChange={(e) => {
                  setLanguageCode(e.target.value as string);
                }}
                sx={{
                  width: '800px',
                }}
                defaultValue={
                  translIndex > -1
                    ? translations[translIndex].language_code
                    : ''
                }
                value={languageCode}
                disabled={!writingTransl}
              >
                {languageCodePredefines.map((name) => (
                  <MenuItem key={name} value={name}>
                    {name}
                  </MenuItem>
                ))}
              </Select>
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
                Content:
              </Typography>
              <TextField
                sx={{
                  width: '800px',
                }}
                multiline
                value={translContent}
                onChange={(e) => {
                  setTranslContent(e.target.value);
                }}
                disabled={!writingTransl}
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
        {!read_only && translID > 0 && !writingTransl && (
          <>
            <IconButton
              onClick={() => {
                setWritingTransl(true);
              }}
            >
              <EditIcon color="secondary" />
            </IconButton>
            <IconButton onClick={onClickConfirmDelete}>
              <DeleteIcon color="error" />
            </IconButton>
          </>
        )}
        {writingTransl && (
          <>
            <IconButton onClick={onClickConfirmUpdate}>
              <CheckIcon color="success" />
            </IconButton>
            <IconButton
              onClick={() => {
                setWritingTransl(false);
                setLanguageCode(
                  translIndex > -1
                    ? translations[translIndex].language_code
                    : ''
                );
                setTranslContent(
                  translIndex > -1 ? translations[translIndex].content : ''
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

export const FieldRow = ({
  read_only,
  field,
  data,
  setData,
  setLoading,
  setToast,
  setErrorToast,
  audioPlaying,
  setAudioPlaying,
  formIndex,
  setOpenUpdateForm,
  setUpdateFormData,
  setNeedToRefresh,
  refreshData,
}: {
  read_only: boolean;
  field: string;
  data: Data;
  setData: (message: Data) => void;
  setLoading: (load: boolean) => void;
  setToast: (message: string) => void;
  setErrorToast: (message: string) => void;
  audioPlaying?: boolean;
  setAudioPlaying?: (playing: boolean) => void;
  formIndex?: number;
  setOpenUpdateForm?: (open: boolean) => void;
  setUpdateFormData?: (req: UpdatePreviewRequest | null) => void;
  setNeedToRefresh?: (need: boolean) => void;
  refreshData?: (bookID: string, index: number) => void;
}) => {
  // dynamic data
  const labels = (field: string): string => {
    switch (field) {
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
      case 'pronunciation_text':
        return 'Pronunciation text';
      case 'form':
        if (data.relatedForms![formIndex!].Pronunciation === '') {
          return `Form [${data.relatedForms![formIndex!].Form}] `;
        } else {
          return `${data.relatedForms![formIndex!].String}   / ${
            data.relatedForms![formIndex!].Pronunciation
          } /`;
        }
      case 'example_1':
        return 'Example 1';
      case 'position_1':
        return 'Positions 1';
      case 'example_2':
        return 'Example 2';
      case 'position_2':
        return 'Positions 2';
      case 'example_3':
        return 'Example 3';
      case 'position_3':
        return 'Positions 3';
      case 'cefr_level':
        return 'Cefr level';
      case 'sort_value':
        return 'SortValue';
      case 'definition_comment':
        return 'Comment';
      case 'quick_search':
        return 'Quick Search';
      case 'definition_operatelog':
        return 'Operate Log';
      default:
        return '';
    }
  };
  const texts = (field: string, data: Data): string => {
    switch (field) {
      case 'string':
        return data.string;
      case 'type':
        return data.type;
      case 'definition':
        return data.definition;
      case 'part_of_speech':
        return data.part_of_speech;
      case 'specific_type':
        return data.specific_type;
      case 'pronunciation_ipa':
        return data.pronunciation_ipa;
      case 'pronunciation_ipa_weak':
        return data.pronunciation_ipa_weak;
      case 'pronunciation_ipa_other':
        return data.pronunciation_ipa_other;
      case 'pronunciation_text':
        return data.pronunciation_text;
      case 'form':
        return data.relatedForms![formIndex!].String;
      case 'example_1':
        return data.example1;
      case 'position_1':
        return data.positions1;
      case 'example_2':
        return data.example2;
      case 'position_2':
        return data.positions2;
      case 'example_3':
        return data.example3;
      case 'position_3':
        return data.positions3;
      case 'cefr_level':
        return data.cefr_level;
      case 'sort_value':
        return data.sort_value ? data.sort_value.toString(10) : '0';
      case 'definition_comment':
        return data.definition_comment;
      default:
        return '';
    }
  };
  const disableEdits = (field: string): boolean => {
    switch (field) {
      case 'string':
        return false;
      case 'type':
        return true;
      case 'definition':
        return false;
      case 'part_of_speech':
        return data.type === 'phrase';
      case 'specific_type':
        return false;
      case 'pronunciation_ipa':
        return data.part_of_speech === 'phrase';
      case 'pronunciation_ipa_weak':
        return data.part_of_speech === 'phrase';
      case 'pronunciation_ipa_other':
        return data.part_of_speech === 'phrase';
      case 'pronunciation_text':
        return false;
      case 'form':
        return false;
      case 'example_1':
        return false;
      case 'position_1':
        return true;
      case 'example_2':
        return false;
      case 'position_2':
        return true;
      case 'example_3':
        return false;
      case 'position_3':
        return true;
      case 'cefr_level':
        return true;
      case 'sort_value':
        return false;
      case 'definition_comment':
        return false;
      default:
        return true;
    }
  };
  const disableDeletes = (field: string): boolean => {
    switch (field) {
      case 'string':
        return true;
      case 'type':
        return true;
      case 'definition':
        return true;
      case 'part_of_speech':
        return true;
      case 'specific_type':
        return false;
      case 'pronunciation_ipa':
        return true;
      case 'pronunciation_ipa_weak':
        return false;
      case 'pronunciation_ipa_other':
        return false;
      case 'pronunciation_text':
        return false;
      case 'form':
        return false;
      case 'example_1':
        return false;
      case 'position_1':
        return true;
      case 'example_2':
        return false;
      case 'position_2':
        return true;
      case 'example_3':
        return false;
      case 'position_3':
        return true;
      case 'cefr_level':
        return true;
      case 'sort_value':
        return true;
      case 'definition_comment':
        return false;
      default:
        return true;
    }
  };
  // clipboard
  const copyToClipboard = (text: string) => {
    clipboard.writeText(text);
  };

  // common states
  const [txt, setTxt] = React.useState(texts(field, data));
  const [prevTxt, setPrevTxt] = React.useState(texts(field, data));

  const [isPartOfSpeech, setIsPartOfSpeech] = React.useState(false);

  const [partOfSpeech, setPartOfSpeech] = React.useState<string[]>([]);

  const ipas = (field: string): string | undefined => {
    switch (field) {
      case 'pronunciation_ipa':
        return data.pronunciation_ipa;
      case 'pronunciation_ipa_weak':
        return data.pronunciation_ipa_weak;
      case 'pronunciation_ipa_other':
        return data.pronunciation_ipa_other;
      case 'form':
        return data.relatedForms![formIndex!].Pronunciation !== ''
          ? data.relatedForms![formIndex!].Pronunciation
          : '';
      default:
        return undefined;
    }
  };
  const needToCheckWordPronunciation = (data: Data): boolean => {
    // we need to check if the word's ipa is right.
    // So, we allow user to speak the word to see
    // the difference between the word's ipa and the word.
    return (
      data !== null &&
      !data.string.includes(' ') &&
      data.pronunciation_text === ''
    );
  };
  const p_texts = (field: string, data: Data): string | undefined => {
    switch (field) {
      case 'string':
        return needToCheckWordPronunciation(data) ? data.string : '';
      case 'pronunciation_text':
        return data.pronunciation_text;
      default:
        return undefined;
    }
  };

  // speech states
  const [playing, setPlaying] = React.useState(false);
  const [ipaEqTxt, setipaEqTxt] = React.useState(
    ipas(field)! === texts(field, data)
  );

  // fetch speech data
  const fetchData = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    // setLoading(true)
    try {
      let param = new URLSearchParams({
        password,
        ipa: ipaEqTxt ? prevTxt : ipas(field)!,
      }).toString();
      if (p_texts(field, data) !== undefined) {
        param = new URLSearchParams({
          password,
          text: p_texts(field, data)!,
        }).toString();
      }
      setPlaying(true);
      setAudioPlaying!(true);

      const res = await fetch(`${shared.baseAPIURL}/book/tts?` + param);
      // setLoading(false)
      const respData = await res.json();
      if (respData.code !== 0) {
        setErrorToast(respData.message);
        setPlaying(false);
        setAudioPlaying!(false);
        return;
      }
      const audio = new Audio();
      audio.src = respData.data.path;
      audio.play();
      audio.onended = () => {
        setPlaying(false);
        setAudioPlaying!(false);
      };
    } catch (error) {
      const message = (error as Error).message;
      // setLoading(false)
      setErrorToast(message);
      setPlaying(false);
      setAudioPlaying!(false);
    }
  };

  const onClickWaitingOrSpeak = () => {
    fetchData();
  };

  // partOfSpeech selections
  const theme = useTheme();
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
  const partOfSpeechPredefines = [
    'noun',
    'pronoun',
    'verb',
    'adjective',
    'adverb',
    'preposition',
    'conjunction',
    'interjection',
    'article',
    'determiner',
    'predeterminer',
  ].sort();
  function getStyles(name: string, names: string[], theme: Theme) {
    return {
      fontWeight:
        names.indexOf(name) === -1
          ? theme.typography.fontWeightRegular
          : theme.typography.fontWeightMedium,
    };
  }
  const handlePartOfSpeechChange = (
    event: SelectChangeEvent<typeof partOfSpeech>
  ) => {
    const {
      target: { value },
    } = event;
    setPartOfSpeech(
      // On autofill we get a stringified value.
      typeof value === 'string' ? value.split(',').sort() : value
    );
    setTxt(typeof value === 'string' ? value : value.sort().join(','));
  };

  // position states
  const [pos, setPos] = React.useState('');
  const [prevPos, setPrevPos] = React.useState('');

  const wordPositions = (field: string, data: Data): string => {
    switch (field) {
      case 'example_1':
        return data.positions1;
      case 'example_2':
        return data.positions2;
      case 'example_3':
        return data.positions3;
      default:
        return '';
    }
  };
  const positionStringToNumbers = (positions: string): number[] => {
    let positionNumbers = positions.split(',').map((item) => Number(item));
    let positionIndexs: number[] = [];
    for (let i = 0; i < positionNumbers.length; i += 2) {
      positionIndexs.push(positionNumbers[i]);
      positionIndexs.push(positionNumbers[i + 1]);
    }
    return positionIndexs;
  };
  const [positionIndexs, setPositionIndexs] = React.useState<number[]>(
    positionStringToNumbers(wordPositions(field, data))
  );
  const [prevPositionIndexs, setPrevPositionIndexs] = React.useState<number[]>(
    positionStringToNumbers(wordPositions(field, data))
  );

  const isExample = (field: string): boolean => {
    return (
      field === 'example_1' || field === 'example_2' || field === 'example_3'
    );
  };

  const exampleToPositionField = (field: string): string => {
    switch (field) {
      case 'example_1':
        return 'position_1';
      case 'example_2':
        return 'position_2';
      case 'example_3':
        return 'position_3';
      default:
        return '';
    }
  };

  const onExampleChange = async (e: React.BaseSyntheticEvent) => {
    setTxt(e.target.value);

    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    try {
      let param = new URLSearchParams({
        password,
        example: e.target.value,
        definitionID: data.definition_id.toString(10),
      }).toString();

      const res = await fetch(
        shared.baseAPIURL + '/book/example_position?' + param
      );
      const respData = await res.json();
      if (respData.code !== 0) {
        setErrorToast(respData.message);
        return;
      }

      setPos(respData.data);
      setPositionIndexs(positionStringToNumbers(respData.data));
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const onPositionChange = (e: React.BaseSyntheticEvent) => {
    setPos(e.target.value);

    setPositionIndexs(positionStringToNumbers(e.target.value));
  };

  const onClickConfirmUpdate = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    const reqData: UpdatePreviewRequest = {
      password: password,
      bookID: data.book_id,
      definitionID: data.definition_id,
      field: field,
    };

    switch (field) {
      case 'string':
        reqData.stringID = data.string_id;
        reqData.string = txt;
        break;
      case 'type':
        break;
      case 'definition':
        reqData.definition = txt;
        break;
      case 'part_of_speech':
        reqData.partOfSpeech = txt;
        break;
      case 'specific_type':
        reqData.specificType = txt;
        break;
      case 'pronunciation_ipa':
        reqData.pronunciationIpa = txt;
        break;
      case 'pronunciation_ipa_weak':
        reqData.pronunciationIpaWeak = txt;
        break;
      case 'pronunciation_ipa_other':
        reqData.pronunciationIpaOther = txt;
        break;
      case 'pronunciation_text':
        reqData.pronunciationText = txt;
        break;
      case 'example_1':
        reqData.exampleID = data.example1_id;
        reqData.example = txt;
        reqData.position = pos;
        break;
      case 'example_2':
        reqData.exampleID = data.example2_id;
        reqData.example = txt;
        reqData.position = pos;
        break;
      case 'example_3':
        reqData.exampleID = data.example3_id;
        reqData.example = txt;
        reqData.position = pos;
        break;
      case 'sort_value':
        reqData.sortValue = parseInt(txt, 10);
      case 'definition_comment':
        reqData.definitionCommentID = data.definition_comment_id;
        reqData.definitionComment = txt;
        break;
    }
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

            let tmpData = data;
            switch (field) {
              case 'string':
                tmpData.string_id = respData.data;
                setPrevTxt(txt.trim());
                setData(tmpData);
                break;
              case 'definition':
                setNeedToRefresh!(true);
                break;
              case 'part_of_speech':
                tmpData.part_of_speech = txt;
                setPrevTxt(txt.trim());
                setData(tmpData);
                break;
              case 'example_1':
                setNeedToRefresh!(true);
                break;
              case 'example_2':
                setNeedToRefresh!(true);
                break;
              case 'example_3':
                setNeedToRefresh!(true);
                break;
              case 'definition_comment':
                tmpData.definition_comment_id = respData.data;
                tmpData.definition_comment = txt;
                setPrevTxt(txt.trim());
                setData(tmpData);
                break;
            }

            setToast('success');
            setOpenUpdate(false);
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

  const [openUpdate, setOpenUpdate] = React.useState(false);

  const handleUpdateDialogClose = () => {
    setOpenUpdate(false);
  };

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
      bookID: data.book_id,
      definitionID: data.definition_id,
      field: field,
    };

    switch (field) {
      case 'form':
        reqData.formStringID = data.relatedForms![formIndex!].StringID;
        break;
      case 'example_1':
        reqData.exampleID = data.example1_id;
        break;
      case 'example_2':
        reqData.exampleID = data.example2_id;
        break;
      case 'example_3':
        reqData.exampleID = data.example3_id;
        break;
      case 'definition_comment':
        reqData.definitionCommentID = data.definition_comment_id;
        break;
    }

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

            let tmpData = data;
            switch (field) {
              case 'definition':
                setNeedToRefresh!(true);
                break;
              case 'form':
                setNeedToRefresh!(true);
                break;
              case 'example_1':
                setNeedToRefresh!(true);
                break;
              case 'example_2':
                setNeedToRefresh!(true);
                break;
              case 'example_3':
                setNeedToRefresh!(true);
                break;
              case 'definition_comment':
                tmpData.definition_comment_id = 0;
                setPrevTxt('');
                setTxt('');
                setData(tmpData);
                break;
            }

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

  const [openChange, setOpenChange] = React.useState(false);
  const [cefrLevel, setCefrLevel] = React.useState(data.cefr_level);
  const [sortValue, setSortValue] = React.useState(data.sort_value);
  const [cefrLevels, setCefrLevels] = React.useState<any[]>([]);

  const handleChangeDialogClose = () => {
    setOpenUpdate(false);
    setSortValue(data.sort_value);
  };

  const getCefrLevels = async (cur_level: string) => {
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
            cur_level,
          }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setCefrLevels(data.data.list);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const getNextSortValue = async (cur_level: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/book/next_sort_value?` +
          new URLSearchParams({
            password,
            cur_level,
          }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setSortValue(data.data.nextSortValue);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const onClickConfirmChange = async () => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    let postData: any = JSON.stringify({
      password,
      bookID: data.book_id,
      definitionID: data.definition_id,
      cefrLevel,
      sortValue,
    });
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
        `${shared.baseAPIURL}/book/update_cefr_level`,
        reqOpts
      );
      setLoading(false);
      const resData = await res.json();
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      // refresh data
      refreshData!(data.book_id.toString(), data.index);
      setOpenChange(false);
      setSortValue(data.sort_value);
      setToast('success');
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const [openTranslView, setOpenTranslView] = React.useState(false);
  const [itemType, setItemType] = React.useState('');
  const [itemID, setItemID] = React.useState(0);
  const [translations, setTranslations] = React.useState<Translation[] | null>(
    null
  );

  const onClickTranslView = () => {
    setOpenTranslView(true);
    parseTranslations();
  };

  const hasTranslationView = (): boolean => {
    if (field === 'definition') {
      return !read_only || data.definition_translations!.length > 0;
    }

    let transls: Translation[] = new Array<Translation>();
    let example_id: number = 0;
    switch (field) {
      case 'example_1':
        example_id = data.example1_id;
        break;
      case 'example_2':
        example_id = data.example2_id;
        break;
      case 'example_3':
        example_id = data.example3_id;
        break;
      default:
        return false;
    }

    if (data.example_translations && data.example_translations.length > 0) {
      for (let i = 0; i < data.example_translations.length; i++) {
        if (data.example_translations[i].item_id === example_id) {
          transls.push(data.example_translations[i]);
        }
      }
      return !read_only || transls.length > 0;
    } else {
      return !read_only && example_id > 0;
    }
  };

  const parseTranslations = () => {
    if (field === 'definition') {
      setItemType('definition');
      setItemID(data.definition_id);
      setTranslations(data.definition_translations!);
      return;
    }

    if (data.example_translations && data.example_translations.length > 0) {
      let transls: Translation[] = new Array<Translation>();
      let example_id: number = 0;
      switch (field) {
        case 'example_1':
          example_id = data.example1_id;
          break;
        case 'example_2':
          example_id = data.example2_id;
          break;
        case 'example_3':
          example_id = data.example3_id;
          break;
      }
      setItemType('example');
      setItemID(example_id);
      for (let i = 0; i < data.example_translations.length; i++) {
        if (data.example_translations[i].item_id === example_id) {
          transls.push(data.example_translations[i]);
        }
      }
      setTranslations(transls);
    }
  };

  const handleTranslViewDialogClose = () => {
    setOpenTranslView(false);
  };

  const getPreviewLatestLogs = async (definitionID: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(
        `${shared.baseAPIURL}/operate_log/preview_latest_logs?` +
          new URLSearchParams({
            password,
            definitionID,
          }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      if (data.data.operateLogs.length > 0) {
        let msg = '';
        for (let i = 0; i < data.data.operateLogs.length; i++) {
          msg +=
            data.data.operateLogs[i].operator +
            '\t' +
            data.data.operateLogs[i].operateType +
            '\t' +
            data.data.operateLogs[i].operateTime +
            '\n';
        }
        setPrevTxt(msg);
      } else {
        setPrevTxt('');
      }
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  // react hook
  React.useEffect(() => {
    setTxt(texts(field, data));
    setPrevTxt(texts(field, data));
    switch (field) {
      case 'part_of_speech':
        setIsPartOfSpeech(true);
        break;
      case 'example_1':
        setPrevPos(data.positions1);
        setPos(data.positions1);
        setPrevPositionIndexs(positionStringToNumbers(data.positions1));
        setPositionIndexs(positionStringToNumbers(data.positions1));
        break;
      case 'example_2':
        setPrevPos(data.positions2);
        setPos(data.positions2);
        setPrevPositionIndexs(positionStringToNumbers(data.positions2));
        setPositionIndexs(positionStringToNumbers(data.positions2));
        break;
      case 'example_3':
        setPrevPos(data.positions3);
        setPos(data.positions3);
        setPrevPositionIndexs(positionStringToNumbers(data.positions3));
        setPositionIndexs(positionStringToNumbers(data.positions3));
        break;
      case 'definition_operatelog':
        getPreviewLatestLogs(data.definition_id.toString());
        break;
    }
  }, [data, field]);

  return (
    <Stack direction="row" alignItems="center">
      <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
        {labels(field)}:
      </Typography>
      <Stack
        direction="row"
        alignItems="center"
        sx={{
          width: '75%',
        }}
      >
        <Paper
          variant="outlined"
          sx={
            isExample(field)
              ? {
                  paddingLeft: '10px',
                  paddingRight: '10px',
                  display: 'flex',
                  width: '90%',
                  flexWrap: 'wrap',
                }
              : {
                  paddingLeft: '10px',
                  paddingRight: '10px',
                  width: '90%',
                }
          }
        >
          {!isExample(field) &&
            field != 'definition_operatelog' &&
            field != 'quick_search' && (
              <Typography
                variant="body1"
                sx={{ whiteSpace: 'pre-wrap' }}
                gutterBottom
              >
                {prevTxt}
              </Typography>
            )}
          {field == 'definition_operatelog' && (
            <TextField
              id="outlined-basic"
              variant="standard"
              minRows={3}
              maxRows={5}
              multiline
              value={prevTxt}
              fullWidth
              disabled
            />
          )}
          {isExample(field) && (
            <Typography
              variant="body1"
              sx={{ whiteSpace: 'pre-wrap' }}
              gutterBottom
            >
              {prevTxt.split('').map((item, index) => {
                for (let j = 0; j < prevPositionIndexs.length; j += 2) {
                  if (
                    index >= prevPositionIndexs[j] &&
                    index < prevPositionIndexs[j] + prevPositionIndexs[j + 1]
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
          )}
          {field == 'quick_search' && (
            <Box
              sx={{
                typography: 'body1',
                overflow: 'scroll',
              }}
            >
              <Stack>
                <Stack direction="row" alignItems="left" spacing={1}>
                  <Typography>Definition:</Typography>
                  {data.definition.split(/\s+/).map((word, idx) => (
                    <Link
                      href={`/search/${word.replace(
                        /[\u2000-\u206F\u2E00-\u2E7F\\'!"#$%&()*+,\-.\/:;<=>?@\[\]^_`{|}~]/g,
                        ''
                      )}`}
                      target="_blank"
                      key={idx}
                    >
                      {word}
                    </Link>
                  ))}
                </Stack>
                <Stack direction="row" alignItems="left" spacing={1}>
                  <Typography>Example1:</Typography>
                  {data.example1.split(/\s+/).map((word, idx) => (
                    <Link
                      href={`/search/${word.replace(
                        /[\u2000-\u206F\u2E00-\u2E7F\\'!"#$%&()*+,\-.\/:;<=>?@\[\]^_`{|}~]/g,
                        ''
                      )}`}
                      target="_blank"
                      key={idx}
                    >
                      {word}
                    </Link>
                  ))}
                </Stack>
                <Stack direction="row" alignItems="left" spacing={1}>
                  <Typography>Example2:</Typography>
                  {data.example2.split(/\s+/).map((word, idx) => (
                    <Link
                      href={`/search/${word.replace(
                        /[\u2000-\u206F\u2E00-\u2E7F\\'!"#$%&()*+,\-.\/:;<=>?@\[\]^_`{|}~]/g,
                        ''
                      )}`}
                      target="_blank"
                      key={idx}
                    >
                      {word}
                    </Link>
                  ))}
                </Stack>
                <Stack direction="row" alignItems="left" spacing={1}>
                  <Typography>Example3:</Typography>
                  {data.example3.split(/\s+/).map((word, idx) => (
                    <Link
                      href={`/search/${word.replace(
                        /[\u2000-\u206F\u2E00-\u2E7F\\'!"#$%&()*+,\-.\/:;<=>?@\[\]^_`{|}~]/g,
                        ''
                      )}`}
                      target="_blank"
                      key={idx}
                    >
                      {word}
                    </Link>
                  ))}
                </Stack>
              </Stack>
            </Box>
          )}
        </Paper>
        {/* translation view button */}
        {hasTranslationView() && (
          <Button
            size="small"
            sx={{
              width: '100px',
              left: '5px',
            }}
            onClick={onClickTranslView}
          >
            TRANSL VIEW
          </Button>
        )}
        {/* waiting or speak button */}
        {((p_texts(field, data) !== undefined && p_texts(field, data) !== '') ||
          (ipaEqTxt && prevTxt)) &&
          !read_only && (
            <Button
              disabled={audioPlaying}
              onClick={onClickWaitingOrSpeak}
              startIcon={playing ? <Downloading /> : <VolumeUp />}
              size="small"
              sx={{
                width: '80px',
                left: '5px',
              }}
            >
              {playing ? 'Waiting' : 'Speak'}
            </Button>
          )}
        {/* copy button */}
        {field == 'string' && (
          <Button
            variant="contained"
            size="small"
            sx={{
              width: '50px',
              left: '30px',
            }}
            onClick={() => {
              copyToClipboard(txt);
            }}
          >
            Copy
          </Button>
        )}

        {/* edit button */}
        {!read_only && (
          <Button
            variant="contained"
            size="small"
            sx={
              disableEdits(field)
                ? {
                    width: '50px',
                    left: '35px',
                    visibility: 'hidden',
                  }
                : {
                    width: '50px',
                    left: '35px',
                  }
            }
            onClick={() => {
              if (field !== 'form') {
                if (isPartOfSpeech) {
                  setPartOfSpeech(
                    typeof txt === 'string' ? txt.split(',') : txt
                  );
                }
                setOpenUpdate(true);
              } else {
                setOpenUpdateForm!(true);
                setUpdateFormData!({
                  password: '',
                  definitionID: 0,
                  bookID: 0,
                  field: 'form',
                  formStringID: data.relatedForms![formIndex!].StringID,
                  form: data.relatedForms![formIndex!].Form,
                  formString: data.relatedForms![formIndex!].String,
                  pronunciation: data.relatedForms![formIndex!].Pronunciation,
                });
              }
            }}
            disabled={disableEdits(field)}
          >
            Edit
          </Button>
        )}

        {/* delete button */}
        {!read_only && (
          <Button
            variant="contained"
            size="small"
            color="warning"
            sx={
              disableDeletes(field) || !prevTxt
                ? {
                    width: '50px',
                    left: '40px',
                    visibility: 'hidden',
                  }
                : {
                    width: '50px',
                    left: '40px',
                  }
            }
            onClick={() => {
              setOpenDelete(true);
            }}
            disabled={disableDeletes(field)}
          >
            Delete
          </Button>
        )}
        {/* change button */}
        {!read_only && (
          <Button
            variant="contained"
            size="small"
            sx={
              field != 'cefr_level' || !prevTxt
                ? {
                    width: '50px',
                    left: '40px',
                    display: 'none',
                  }
                : {
                    width: '50px',
                    left: '40px',
                  }
            }
            onClick={() => {
              // fetch cefr levels
              getCefrLevels(data.cefr_level);
              setOpenChange(true);
            }}
          >
            Change
          </Button>
        )}
      </Stack>

      {/* change dialog */}
      <Dialog
        fullWidth
        maxWidth="md"
        open={openChange}
        onClose={handleChangeDialogClose}
      >
        <DialogContent>
          <Typography variant="h6" gutterBottom>
            Before:
          </Typography>
          <Stack direction="row" alignItems="center">
            <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
              {labels(field)}:
            </Typography>
            <Typography
              variant="body1"
              sx={{
                whiteSpace: 'pre-wrap',
                width: '800px',
              }}
              gutterBottom
            >
              {prevTxt}
            </Typography>
          </Stack>
          <Stack direction="row" alignItems="center">
            <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
              {labels('sort_value')}:
            </Typography>
            <Typography
              variant="body1"
              sx={{
                whiteSpace: 'pre-wrap',
                width: '800px',
              }}
              gutterBottom
            >
              {data.sort_value}
            </Typography>
          </Stack>
          <Typography variant="h6" gutterBottom>
            After:
          </Typography>
          <Stack direction="row" alignItems="center">
            <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
              {labels(field)}:
            </Typography>
            <Select
              onChange={(e) => {
                setCefrLevel(e.target.value as string);
                getNextSortValue(e.target.value as string);
              }}
              sx={{
                width: '800px',
              }}
            >
              {cefrLevels.map((item) => (
                <MenuItem key={item.bookID} value={item.level}>
                  {item.level}
                </MenuItem>
              ))}
            </Select>
          </Stack>
          <Stack direction="row" alignItems="center">
            <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
              {labels('sort_value')}:
            </Typography>
            <TextField
              sx={{
                width: '800px',
              }}
              value={sortValue}
              onChange={(e) => {
                setSortValue(parseInt(e.target.value as string));
              }}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="error"
            onClick={() => {
              setTxt(prevTxt);
              setPos(prevPos);
              setOpenChange(false);
              setSortValue(data.sort_value);
            }}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            color="success"
            onClick={onClickConfirmChange}
          >
            Confirm
          </Button>
        </DialogActions>
      </Dialog>

      {/* update dialog */}
      <Dialog
        fullWidth
        maxWidth="md"
        open={openUpdate}
        onClose={handleUpdateDialogClose}
      >
        <DialogContent>
          <Typography variant="h6" gutterBottom>
            Before:
          </Typography>
          {!isExample(field) && (
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                {labels(field)}:
              </Typography>
              <Typography
                variant="body1"
                sx={{
                  whiteSpace: 'pre-wrap',
                  width: '800px',
                }}
                gutterBottom
              >
                {prevTxt}
              </Typography>
            </Stack>
          )}
          {isExample(field) && (
            <Stack
              direction="row"
              alignItems="center"
              style={{
                flexWrap: 'wrap',
              }}
            >
              <Typography variant="body1" gutterBottom sx={{ width: '15%' }}>
                {labels(field)}:
              </Typography>
              <Typography
                variant="body1"
                sx={{ whiteSpace: 'pre-wrap' }}
                gutterBottom
              >
                {prevTxt.split('').map((item, index) => {
                  for (let j = 0; j < prevPositionIndexs.length; j += 2) {
                    if (
                      index >= prevPositionIndexs[j] &&
                      index < prevPositionIndexs[j] + prevPositionIndexs[j + 1]
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
          )}
          {isExample(field) && (
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '15%' }}>
                {labels(exampleToPositionField(field))}:
              </Typography>
              <Typography
                variant="body1"
                sx={{ whiteSpace: 'pre-wrap' }}
                gutterBottom
              >
                {prevPos}
              </Typography>
            </Stack>
          )}
          <Typography variant="h6" gutterBottom>
            After:
          </Typography>
          {!isExample(field) && (
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                {labels(field)}:
              </Typography>
              {!isPartOfSpeech && (
                <TextField
                  sx={{
                    width: '800px',
                    height:
                      field === 'definition' || field === 'definition_comment'
                        ? '160px'
                        : '40px',
                  }}
                  multiline={
                    field === 'definition' || field === 'definition_comment'
                  }
                  rows={
                    field === 'definition' || field === 'definition_comment'
                      ? 3
                      : 1
                  }
                  defaultValue={txt}
                  value={txt}
                  onChange={(e) => {
                    if (field === 'sort_value') {
                      let num = parseInt(e.target.value, 10).toString(10);
                      setTxt(num !== 'NaN' ? num : '');
                    } else {
                      setTxt(e.target.value);
                    }
                  }}
                />
              )}
              {isPartOfSpeech && (
                <Select
                  multiple
                  value={partOfSpeech}
                  onChange={handlePartOfSpeechChange}
                  MenuProps={MenuProps}
                  sx={{
                    width: '800px',
                  }}
                >
                  {partOfSpeechPredefines.map((name) => (
                    <MenuItem
                      key={name}
                      value={name}
                      style={getStyles(name, partOfSpeech, theme)}
                    >
                      {name}
                    </MenuItem>
                  ))}
                </Select>
              )}
            </Stack>
          )}
          {isExample(field) && (
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '17%' }}>
                {labels(field)}:
              </Typography>
              <TextField
                sx={{
                  width: '800px',
                  height: '160px',
                }}
                multiline
                rows={3}
                defaultValue={txt}
                value={txt}
                onChange={onExampleChange}
              />
            </Stack>
          )}
          {isExample(field) && (
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '17%' }}>
                {labels(exampleToPositionField(field))}:
              </Typography>
              <TextField
                sx={{
                  width: '800px',
                }}
                defaultValue={pos}
                value={pos}
                onChange={onPositionChange}
              />
            </Stack>
          )}
          {isExample(field) && (
            <Stack
              direction="row"
              alignItems="center"
              style={{
                flexWrap: 'wrap',
              }}
            >
              <Typography variant="body1" gutterBottom sx={{ width: '15%' }}>
                Result:
              </Typography>
              <Typography
                variant="body1"
                sx={{ whiteSpace: 'pre-wrap' }}
                gutterBottom
              >
                {txt.split('').map((item, index) => {
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
          )}
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="error"
            onClick={() => {
              setTxt(prevTxt);
              setPos(prevPos);
              setOpenUpdate(false);
            }}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            color="success"
            onClick={onClickConfirmUpdate}
          >
            Confirm
          </Button>
        </DialogActions>
      </Dialog>

      {/* delete dialog */}
      <Dialog
        fullWidth
        maxWidth="sm"
        open={openDelete}
        onClose={handleDeleteDialogClose}
      >
        <DialogContent>
          <Typography variant="h6" gutterBottom>
            Confirm to delete ?
          </Typography>
          <Stack direction="row" alignItems="center">
            <Typography variant="body1" gutterBottom sx={{ width: '50%' }}>
              {labels(field)}:
            </Typography>
            <Typography
              variant="body1"
              sx={{
                whiteSpace: 'pre-wrap',
                width: '800px',
              }}
              gutterBottom
            >
              {prevTxt}
            </Typography>
          </Stack>
          {isExample(field) && (
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '50%' }}>
                {labels(exampleToPositionField(field))}:
              </Typography>
              <Typography
                variant="body1"
                sx={{
                  whiteSpace: 'pre-wrap',
                  width: '800px',
                }}
                gutterBottom
              >
                {prevPos}
              </Typography>
            </Stack>
          )}
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="error"
            onClick={() => {
              setOpenDelete(false);
            }}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            color="success"
            onClick={onClickConfirmDelete}
          >
            Confirm
          </Button>
        </DialogActions>
      </Dialog>

      {/* translation dialog */}
      <Dialog
        fullWidth
        maxWidth="md"
        open={openTranslView}
        onClose={handleTranslViewDialogClose}
      >
        <IconButton
          aria-label="close"
          onClick={() => {
            setOpenTranslView(false);
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
              Translations:
            </Typography>
          </Stack>

          {translations &&
            translations.length > 0 &&
            translations.map((transl, index) => {
              if (index + 1 < translations.length) {
                return (
                  <>
                    <TranslationBoard
                      read_only={read_only}
                      book_id={data.book_id}
                      definition_id={data.definition_id}
                      item_type={itemType}
                      item_id={itemID}
                      translations={translations}
                      translIndex={index}
                      setTranslations={setTranslations}
                      setLoading={setLoading}
                      setToast={setToast}
                      setErrorToast={setErrorToast}
                      setNeedToRefresh={setNeedToRefresh}
                    />
                    <Divider variant="middle" />
                  </>
                );
              } else {
                return (
                  <>
                    <TranslationBoard
                      read_only={read_only}
                      book_id={data.book_id}
                      definition_id={data.definition_id}
                      item_type={itemType}
                      item_id={itemID}
                      translations={translations}
                      translIndex={index}
                      setTranslations={setTranslations}
                      setLoading={setLoading}
                      setToast={setToast}
                      setErrorToast={setErrorToast}
                      setNeedToRefresh={setNeedToRefresh}
                    />
                    <Divider variant="middle" />
                    <TranslationBoard
                      read_only={read_only}
                      book_id={data.book_id}
                      definition_id={data.definition_id}
                      item_type={itemType}
                      item_id={itemID}
                      translations={translations}
                      translIndex={-1}
                      setTranslations={setTranslations}
                      setLoading={setLoading}
                      setToast={setToast}
                      setErrorToast={setErrorToast}
                      setNeedToRefresh={setNeedToRefresh}
                    />
                  </>
                );
              }
            })}

          {(!translations || translations.length === 0) && (
            <TranslationBoard
              read_only={read_only}
              book_id={data.book_id}
              definition_id={data.definition_id}
              item_type={itemType}
              item_id={itemID}
              translations={[]}
              translIndex={-1}
              setTranslations={setTranslations}
              setLoading={setLoading}
              setToast={setToast}
              setErrorToast={setErrorToast}
              setNeedToRefresh={setNeedToRefresh}
            />
          )}
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            onClick={() => {
              setOpenTranslView(false);
            }}
          >
            OK
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
};
