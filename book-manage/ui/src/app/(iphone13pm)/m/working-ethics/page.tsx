'use client';
import * as React from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import Snackbar, { SnackbarOrigin } from '@mui/material/Snackbar';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';
import shared from '@/app/shared';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import { DemoContainer } from '@mui/x-date-pickers/internals/demo';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import Timeline from '@mui/lab/Timeline';
import TimelineItem from '@mui/lab/TimelineItem';
import TimelineSeparator from '@mui/lab/TimelineSeparator';
import TimelineConnector from '@mui/lab/TimelineConnector';
import TimelineContent from '@mui/lab/TimelineContent';
import TimelineDot from '@mui/lab/TimelineDot';
import TimelineOppositeContent, {
  timelineOppositeContentClasses,
} from '@mui/lab/TimelineOppositeContent';
import dayjs from 'dayjs';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

interface GetWorkingEthicsResponseData {
  list: WorkingEthicData[];
  totalDuration: number;
}

interface WorkingEthicData {
  duration: number;
  operateLogs: OperateLog[];
}

interface OperateLog {
  id: number;
  operator: string;
  operateTime: string;
  operateStatus: number;
  operateType: string;
  bookID: number;
  definitionID: number;
  otherOperateParams: string;
  error: string;
}

export default function Page() {
  const [operators, setOperators] = React.useState<string[] | null>(null);
  const [operator, setOperator] = React.useState('');
  const [operateDate, setOperateDate] = React.useState(dayjs().startOf('day'));

  const handleOperatorSelectChange = (event: SelectChangeEvent) => {
    setOperator(event.target.value);
  };

  const [errorToast, setErrorToast] = React.useState('');
  const [loading, setLoading] = React.useState(false);

  const [workingEthics, setWorkingEthics] =
    React.useState<GetWorkingEthicsResponseData | null>(null);

  const getWorkingEthics = async (operator: string, date: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    try {
      let param = new URLSearchParams({
        password,
        operator,
        date,
      });
      const res = await fetch(
        `${shared.baseAPIURL}/operate_log/workingEthics?` + param.toString()
      );

      setLoading(false);

      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setWorkingEthics(data.data);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  const getStaffList = async (nonAdmin: boolean) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    try {
      let param = new URLSearchParams({
        password,
        nonAdmin: nonAdmin.toString(),
      });
      const res = await fetch(
        `${shared.baseAPIURL}/system/staffs?` + param.toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setOperators(data.data);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  React.useEffect(() => {
    getStaffList(true);
  }, []);

  return (
    <Stack
      spacing={2}
      direction="column"
      alignItems="center"
      sx={{
        padding: '1rem',
      }}
    >
      <FormControl
        variant="standard"
        sx={{
          width: '80%',
          m: 1,
          minWidth: 120,
        }}
      >
        <InputLabel>Operator</InputLabel>
        <Select
          defaultValue={operator}
          value={operator}
          onChange={handleOperatorSelectChange}
          label="Operator"
        >
          <MenuItem value="">
            <em>None</em>
          </MenuItem>
          {operators &&
            operators.map((item, index) => {
              return (
                <MenuItem key={index} value={item}>
                  {item}
                </MenuItem>
              );
            })}
        </Select>
      </FormControl>
      <FormControl
        variant="standard"
        sx={{
          width: '80%',
          m: 1,
          minWidth: 120,
        }}
      >
        <LocalizationProvider dateAdapter={AdapterDayjs}>
          <DemoContainer components={['DatePicker']}>
            <DatePicker
              label="Date"
              defaultValue={operateDate}
              onChange={(val) => {
                setOperateDate(val!);
              }}
              sx={{
                width: '100%',
              }}
            />
          </DemoContainer>
        </LocalizationProvider>
      </FormControl>
      <Button
        variant="contained"
        onClick={() => {
          getWorkingEthics(operator, operateDate.format('YYYY-MM-DD'));
        }}
      >
        Search
      </Button>

      {workingEthics && (
        <Typography
          sx={{
            width: '300px',
          }}
        >
          {'Total working duration: '}
          {workingEthics.totalDuration / 60 >= 1
            ? Math.floor(workingEthics.totalDuration / 60) + 'm'
            : ''}
          {(workingEthics.totalDuration % 60) + 's'}
        </Typography>
      )}

      <Timeline
        sx={{
          [`& .${timelineOppositeContentClasses.root}`]: {
            flex: 0.2,
          },
        }}
      >
        {workingEthics?.list &&
          workingEthics.list.map((item, index) => {
            return (
              <TimelineItem key={index}>
                <TimelineOppositeContent color="textSecondary">
                  <Typography
                    sx={{
                      width: '180px',
                    }}
                  >
                    {item.operateLogs.length > 1
                      ? item.operateLogs[item.operateLogs.length - 1]
                          .operateTime +
                        ' To ' +
                        item.operateLogs[0].operateTime
                      : item.operateLogs[0].operateTime}
                  </Typography>
                </TimelineOppositeContent>
                <TimelineSeparator>
                  <TimelineDot
                    color={item.operateLogs.length > 2 ? 'success' : 'warning'}
                  />
                  {index < workingEthics.list.length - 1 && (
                    <TimelineConnector />
                  )}
                </TimelineSeparator>
                <TimelineContent>
                  <Typography
                    sx={{
                      width: '180px',
                    }}
                  >
                    {'working duration: '}
                    {item.duration / 60 >= 1
                      ? Math.floor(item.duration / 60) + 'm'
                      : ''}
                    {(item.duration % 60) + 's'}
                  </Typography>
                </TimelineContent>
              </TimelineItem>
            );
          })}
      </Timeline>

      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
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
    </Stack>
  );
}
