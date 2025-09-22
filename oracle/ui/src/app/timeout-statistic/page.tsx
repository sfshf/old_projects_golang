'use client';
import Backdrop from '@mui/material/Backdrop';
import React, { useEffect } from 'react';
import CircularProgress from '@mui/material/CircularProgress';
import Button from '@mui/material/Button';
import Snackbar from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '@/app/shared';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Stack from '@mui/material/Stack';
import moment from 'moment-timezone';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import Pagination from '@mui/material/Pagination';
import dayjs, { Dayjs } from 'dayjs';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'path', label: 'Path', minWidth: 100 },
  { id: 'application', label: 'Application', minWidth: 100 },
  { id: 'service', label: 'Service', minWidth: 100 },
  { id: 'date', label: 'Date', minWidth: 100 },
  { id: 'count', label: 'Count', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [total, setTotal] = React.useState(0);
  const [list, setList] = React.useState<any[] | null>(null);

  const [pageSize, setPageSize] = React.useState(30);
  const [pageNumber, setPageNumber] = React.useState(0);
  const [applicationID, setApplicationID] = React.useState('0');
  const handleApplicationChange = (event: SelectChangeEvent) => {
    let applicationID = event.target.value as string;
    setApplicationID(applicationID);
    setServiceID('0');
    listServices(applicationID);
    setPageNumber(0);
  };
  const [serviceID, setServiceID] = React.useState('0');
  const handleServiceChange = (event: SelectChangeEvent) => {
    let serviceID = event.target.value as string;
    setServiceID(serviceID);
    listServicePaths(serviceID);
    setPageNumber(0);
  };
  const [path, setPath] = React.useState('');
  const handlePathChange = (event: SelectChangeEvent) => {
    setPath(event.target.value as string);
    setPageNumber(0);
  };
  const [date, setDate] = React.useState<Dayjs | null>(dayjs());
  const handleDateChange = (value: any, context: any) => {
    setDate(dayjs(value.format('YYYY-MM-DD')));
    setPageNumber(0);
    listTimeoutStatistics(
      pageSize,
      0,
      applicationID,
      serviceID,
      path,
      value.format('YYYY-MM-DD')
    );
  };

  const listTimeoutStatistics = async (
    pageSize: number,
    pageNumber: number,
    applicationID: string,
    serviceID: string,
    path: string,
    date: string
  ) => {
    const reqData = {
      apiKey: shared.getPassword(),
      pageSize: pageSize,
      pageNumber: pageNumber,
      applicationID: parseInt(applicationID),
      serviceID: parseInt(serviceID),
      path: path,
      date: date,
    };
    post('/console/listTimeoutStatistics/v1', reqData, (respData) => {
      setTotal(respData.data.total);
      setList(respData.data.list);
    });
  };

  const [applicationList, setApplicationList] = React.useState<any[] | null>(
    null
  );
  const listApplications = async () => {
    const reqData = {
      apiKey: shared.getPassword(),
    };
    post('/console/listApplications/v1', reqData, (respData) => {
      setApplicationList(respData.data.list);
    });
  };

  const [serviceList, setServiceList] = React.useState<any[] | null>(null);
  const listServices = async (applicationID: string) => {
    if (applicationID == '' || applicationID == '0') {
      setServiceList([]);
      setPathList([]);
      return;
    }
    const reqData = {
      apiKey: shared.getPassword(),
      applicationID: parseInt(applicationID),
    };
    post('/console/listServices/v1', reqData, (respData) => {
      setServiceList(respData.data.list);
    });
  };

  const [pathList, setPathList] = React.useState<any[] | null>(null);
  const listServicePaths = async (serviceID: string) => {
    if (serviceID == '' || serviceID == '0') {
      setPathList([]);
      return;
    }
    const reqData = {
      apiKey: shared.getPassword(),
      serviceID: parseInt(serviceID),
    };
    post('/console/listServicePaths/v1', reqData, (respData) => {
      setPathList(respData.data.list);
    });
  };

  const post = async (
    path: string,
    reqData?: any,
    successAction?: (respData: any) => void
  ) => {
    setLoading(true);
    let postData = '';
    if (reqData) {
      postData = JSON.stringify(reqData);
    }
    try {
      let reqOpts: any = {};
      reqOpts.method = 'POST';
      let headers: any = {};
      if (postData) {
        headers = {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
        };
        reqOpts['body'] = postData;
      }
      reqOpts['headers'] = headers;
      fetch(path, reqOpts)
        .then((resp) => {
          return resp.json();
        })
        .then((data) => {
          if (data.code !== 0) {
            setErrorToast(data.debugMessage);
            return;
          }
          if (successAction) {
            successAction(data);
          }
        })
        .then((err) => {
          throw err;
        });
      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setErrorToast(message);
    }
  };

  const handleChangePage = (
    event: React.ChangeEvent<unknown>,
    newPage: number
  ) => {
    setPageNumber(newPage - 1);
    listTimeoutStatistics(
      pageSize,
      newPage - 1,
      applicationID,
      serviceID,
      path,
      date ? date.format('YYYY-MM-DD') : ''
    );
  };

  useEffect(() => {
    listApplications();
    listTimeoutStatistics(
      pageSize,
      pageNumber,
      applicationID,
      serviceID,
      path,
      date ? date.format('YYYY-MM-DD') : ''
    );
  }, []);

  return (
    <Stack
      spacing={5}
      direction="column"
      alignItems="center"
      sx={{
        padding: '1rem',
      }}
    >
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
          width: '100%',
        }}
        direction="row"
      >
        <Button
          onClick={() => {
            listTimeoutStatistics(
              pageSize,
              pageNumber,
              applicationID,
              serviceID,
              path,
              date ? date.format('YYYY-MM-DD') : ''
            );
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Refresh
        </Button>
        <Button
          onClick={() => {
            setApplicationID('0');
            setServiceList([]);
            setServiceID('0');
            setPathList([]);
            setPath('');
            setDate(null);
            setPageNumber(0);
            listTimeoutStatistics(pageSize, 0, '0', '0', '', '');
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Reset
        </Button>
        <FormControl
          sx={{
            width: '15%',
          }}
        >
          <InputLabel id="demo-simple-select-label">Aplication</InputLabel>
          <Select
            labelId="demo-simple-select-label"
            id="demo-simple-select"
            value={applicationID}
            label="Aplication"
            onChange={handleApplicationChange}
          >
            {applicationList &&
              applicationList.map((item) => (
                <MenuItem value={item.id} key={item.id}>
                  {item.name}
                </MenuItem>
              ))}
          </Select>
        </FormControl>
        <FormControl
          sx={{
            width: '15%',
          }}
        >
          <InputLabel id="demo-simple-select-label">Service</InputLabel>
          <Select
            labelId="demo-simple-select-label"
            id="demo-simple-select"
            value={serviceID}
            label="Service"
            onChange={handleServiceChange}
          >
            {serviceList &&
              serviceList.map((item) => (
                <MenuItem value={item.id} key={item.id}>
                  {item.name}
                </MenuItem>
              ))}
          </Select>
        </FormControl>
        <FormControl
          sx={{
            width: '15%',
          }}
        >
          <InputLabel id="demo-simple-select-label">Path</InputLabel>
          <Select
            labelId="demo-simple-select-label"
            id="demo-simple-select"
            value={path}
            label="Path"
            onChange={handlePathChange}
          >
            {pathList &&
              pathList.map((item) => (
                <MenuItem value={item} key={item}>
                  {item}
                </MenuItem>
              ))}
          </Select>
        </FormControl>
        <LocalizationProvider dateAdapter={AdapterDayjs}>
          <DatePicker
            value={date}
            onChange={handleDateChange}
            format="YYYY-MM-DD"
          />
        </LocalizationProvider>
        <Button
          onClick={() => {
            let today = dayjs();
            setPageNumber(0);
            setDate(today);
            listTimeoutStatistics(
              pageSize,
              0,
              applicationID,
              serviceID,
              path,
              today.format('YYYY-MM-DD')
            );
          }}
          variant="contained"
          sx={{ height: '50px', width: '100px', margin: '5px' }}
        >
          Today
        </Button>
      </Stack>
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
          width: '100%',
        }}
        direction="row"
      >
        <TableContainer>
          <Table stickyHeader aria-label="sticky table">
            <TableHead>
              <TableRow>
                {columns.map((column) => (
                  <TableCell
                    key={column.id}
                    align={column.align}
                    style={{ minWidth: column.minWidth }}
                  >
                    {column.label}
                  </TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {list &&
                list.map((row) => {
                  return (
                    <TableRow hover role="checkbox" tabIndex={-1} key={row.id}>
                      {columns.map((column) => {
                        const value = row[column.id];
                        if (column.id == 'createdAt') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              {moment(value)
                                .tz('Asia/Shanghai')
                                .format('YYYY-MM-DD HH:mm:ss')}
                            </TableCell>
                          );
                        } else if (column.id == 'date') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              {moment(value)
                                .tz('Asia/Shanghai')
                                .format('YYYY-MM-DD')}
                            </TableCell>
                          );
                        } else {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              {column.format !== undefined &&
                              typeof value === 'number'
                                ? column.format(value)
                                : value}
                            </TableCell>
                          );
                        }
                      })}
                    </TableRow>
                  );
                })}
            </TableBody>
          </Table>
        </TableContainer>
      </Stack>
      <Stack
        spacing={2}
        alignItems="center"
        sx={{
          padding: '1rem',
        }}
        direction="row"
      >
        <Pagination
          color="primary"
          count={
            total % pageSize === 0
              ? total / pageSize
              : Math.floor(total / pageSize) + 1
          }
          page={pageNumber + 1}
          onChange={handleChangePage}
        />
      </Stack>
      <Snackbar
        open={errorToast !== ''}
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
        autoHideDuration={2000}
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
    </Stack>
  );
}
