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
import Pagination from '@mui/material/Pagination';
import Link from '@mui/material/Link';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

const baseURL = process.env.NEXT_PUBLIC_SERVER_BASE_URL;

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'id', label: 'ID', minWidth: 100 },
  { id: 'symbol', label: 'Symbol', minWidth: 50 },
  { id: 'name', label: 'Name', minWidth: 100 },
  { id: 'type', label: 'Is ERC20', minWidth: 50 },
  { id: 'address', label: 'Address', minWidth: 400 },
  { id: 'checked', label: 'Checked', minWidth: 50 },
  { id: 'operations', label: 'Operations', minWidth: 100 },
];

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [total, setTotal] = React.useState(0);
  const [list, setList] = React.useState<any[] | null>(null);
  const [pageSize, setPageSize] = React.useState(40);
  const [pageNumber, setPageNumber] = React.useState(0);
  const [checked, setChecked] = React.useState(false);
  const handleChangePage = (
    event: React.ChangeEvent<unknown>,
    newPage: number
  ) => {
    setPageNumber(newPage - 1);
    listErrorTokens(pageSize, newPage - 1, checked);
  };

  const listErrorTokens = async (
    pageSize: number,
    pageNumber: number,
    checked: boolean
  ) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    let postData: any = JSON.stringify({
      apiKey: password,
      pageSize,
      pageNumber,
      checked,
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
        `${shared.baseAPIURL}/doom/console/listErrorTokens/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setTotal(resData.data.total);
      setList(resData.data.list);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  const checkErrorToken = async (id: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    let postData: any = JSON.stringify({
      apiKey: password,
      id,
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
        `${shared.baseAPIURL}/doom/console/checkErrorToken/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setToast('success');
      // refresh list
      listErrorTokens(pageSize, pageNumber, checked);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  const detectErrorToken = async (id: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    let postData: any = JSON.stringify({
      apiKey: password,
      id,
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
        `${shared.baseAPIURL}/doom/console/detectErrorToken/v1`,
        reqOpts
      );
      const resData = await res.json();
      setLoading(false);
      if (resData.code !== 0) {
        setErrorToast(resData.message);
        return;
      }
      setToast(resData.data.result);
      // refresh list
      listErrorTokens(pageSize, pageNumber, checked);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  return (
    <Stack
      spacing={3}
      direction="column"
      alignItems="center"
      sx={{
        padding: '1rem',
      }}
    >
      <Grid container spacing={4}>
        <Grid item xs={1}>
          <Button
            variant="contained"
            size="large"
            sx={{
              height: '50px',
            }}
            onClick={() => {
              listErrorTokens(pageSize, pageNumber, checked);
            }}
          >
            Search
          </Button>
        </Grid>
        <Grid item xs={1}>
          <Stack
            spacing={1}
            direction="row"
            alignItems="center"
            sx={{
              padding: '1rem',
            }}
          >
            <FormControlLabel
              control={
                <Switch
                  checked={checked}
                  onChange={(e) => {
                    setChecked(e.target.checked);
                    listErrorTokens(pageSize, pageNumber, e.target.checked);
                  }}
                />
              }
              label="Checked"
            />
          </Stack>
        </Grid>
        <Grid item xs={1}>
          <Stack
            spacing={1}
            direction="row"
            alignItems="center"
            sx={{
              padding: '1rem',
            }}
          >
            <Typography variant="h4" gutterBottom>
              Total:
            </Typography>
            <Typography variant="h4" gutterBottom>
              {total}
            </Typography>
          </Stack>
        </Grid>
      </Grid>
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
            {list != null && (
              <TableBody>
                {list.map((row) => {
                  return (
                    <TableRow hover role="checkbox" tabIndex={-1} key={row.id}>
                      {columns.map((column) => {
                        const value = row[column.id];
                        if (column.id === 'address') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              <Link
                                href={`https://etherscan.io/address/${value}`}
                                target="_blank"
                              >
                                {value}
                              </Link>
                            </TableCell>
                          );
                        } else if (column.id === 'type') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              {value == 'ERC20' ? 'True' : 'False'}
                            </TableCell>
                          );
                        } else if (column.id === 'checked') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              {value ? 'True' : 'False'}
                            </TableCell>
                          );
                        } else if (column.id == 'operations') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              <Button
                                variant="text"
                                size="small"
                                onClick={() => {
                                  detectErrorToken(row.id);
                                }}
                              >
                                Detect
                              </Button>
                              <Button
                                variant="text"
                                size="small"
                                onClick={() => {
                                  checkErrorToken(row.id);
                                }}
                              >
                                {checked ? 'Uncheck' : 'Check'}
                              </Button>
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
            )}
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
