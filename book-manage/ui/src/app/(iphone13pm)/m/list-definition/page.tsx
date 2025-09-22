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
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import Link from '@mui/material/Link';
import TablePagination from '@mui/material/TablePagination';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Grid from '@mui/material/Grid';

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
  { id: 'string', label: 'String', minWidth: 100 },
  { id: 'partOfSpeech', label: 'POS', minWidth: 100 },
  { id: 'definition', label: 'Definition', minWidth: 400 },
  { id: 'type', label: 'Type', minWidth: 100 },
  { id: 'index', label: 'Index', minWidth: 100 },
];

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [cefrLevel, setCefrLevel] = React.useState('');
  const [cefrLevels, setCefrLevels] = React.useState<any[]>([]);
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

  const [list, setList] = React.useState<any[] | null>(null);
  const [page, setPage] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(50);
  const [total, setTotal] = React.useState(0);
  const handleChangePage = (
    event: React.MouseEvent<HTMLButtonElement> | null,
    newPage: number
  ) => {
    setPage(newPage);
    listDefinition(newPage.toString(), pageSize.toString());
  };
  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    setPageSize(parseInt(event.target.value, 10));
    setPage(0);
    listDefinition('0', event.target.value);
  };

  const listDefinition = async (page: string, pageSize: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }
    setLoading(true);
    try {
      let param = new URLSearchParams({
        password,
        cefrLevel,
        page,
        pageSize,
      });
      const res = await fetch(
        `${shared.baseAPIURL}/book/list_definition?` + param.toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setTotal(data.data.total);
      setList(data.data.list);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  useEffect(() => {
    getCefrLevels('');
  }, []);

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
        <Grid item xs={3}>
          <Button
            variant="contained"
            size="large"
            sx={{
              height: '50px',
            }}
            onClick={() => {
              listDefinition(page.toString(), pageSize.toString());
            }}
          >
            Search
          </Button>
        </Grid>
        <Grid item xs={8}>
          <FormControl fullWidth>
            <InputLabel>Cefr Level</InputLabel>
            <Select
              onChange={(e) => {
                setCefrLevel(e.target.value as string);
              }}
              label="Cefr Level"
              sx={{
                height: '50px',
              }}
            >
              {cefrLevels.map((item) => (
                <MenuItem key={item.bookID} value={item.level}>
                  {item.level}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
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
                <TableCell
                  key="operations"
                  align="center"
                  style={{ minWidth: 100 }}
                >
                  Operations
                </TableCell>
              </TableRow>
            </TableHead>
            {list != null && (
              <TableBody>
                {list.map((row) => {
                  return (
                    <TableRow hover role="checkbox" tabIndex={-1} key={row.id}>
                      {columns.map((column) => {
                        const value = row[column.id];
                        return (
                          <TableCell key={column.id} align={column.align}>
                            {column.format !== undefined &&
                            typeof value === 'number'
                              ? column.format(value)
                              : value}
                          </TableCell>
                        );
                      })}
                      <TableCell
                        key="download"
                        align="center"
                        style={{ minWidth: 100 }}
                        scope="row"
                      >
                        <Link
                          href={`/preview?definitionID=${row.definitionID}`}
                          target="_blank"
                        >
                          Jump
                        </Link>
                      </TableCell>
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
        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 30, 50, 100]}
          component="div"
          count={total}
          page={page}
          onPageChange={handleChangePage}
          rowsPerPage={pageSize}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Stack>

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
