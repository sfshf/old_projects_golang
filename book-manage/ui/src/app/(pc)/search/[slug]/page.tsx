'use client';
import React, { useEffect } from 'react';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Snackbar, { SnackbarOrigin } from '@mui/material/Snackbar';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import shared from '@/app/shared';
import { styled } from '@mui/material/styles';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell, { tableCellClasses } from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import TablePagination from '@mui/material/TablePagination';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import Preview from '../../preview/preview';
import Typography from '@mui/material/Typography';
import Link from '@mui/material/Link';

interface SearchStringItem {
  bookID: number;
  definitionID: number;
  string: string;
  level: string;
  type: string;
  partOfSpeech: string;
  index: number;
  definition: string;
}

interface HeadCell {
  disablePadding: boolean;
  id: string;
  label: string;
  minWidth?: number;
}

function EnhancedTableHead() {
  const headCells: readonly HeadCell[] = [
    {
      id: 'string',
      disablePadding: false,
      label: 'String',
      minWidth: 50,
    },
    {
      id: 'part_of_speech',
      disablePadding: false,
      label: 'POS',
      minWidth: 50,
    },
    {
      id: 'definition',
      disablePadding: false,
      label: 'Definition',
      minWidth: 300,
    },
    {
      id: 'level',
      disablePadding: false,
      label: 'Level',
      minWidth: 50,
    },
    {
      id: 'type',
      disablePadding: false,
      label: 'Type',
      minWidth: 50,
    },
    {
      id: 'bookID',
      disablePadding: false,
      label: 'BookID',
      minWidth: 50,
    },
    {
      id: 'index',
      disablePadding: false,
      label: 'Index',
      minWidth: 50,
    },

    {
      id: 'operation',
      disablePadding: false,
      label: 'Operation',
      minWidth: 50,
    },
  ];

  return (
    <TableHead>
      <TableRow>
        {headCells.map((headCell) => (
          <TableCell
            key={headCell.id}
            align="left"
            padding={headCell.disablePadding ? 'none' : 'normal'}
            style={{ minWidth: headCell.minWidth }}
            size="small"
          >
            {headCell.label}
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}

export default function Page({ params }: { params: { slug: string } }) {
  const [bookID, setBookID] = React.useState('');
  const [detailBookID, setDetailBookID] = React.useState(0);
  const [type, setType] = React.useState('');
  const [page, setPage] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(30);
  const [dataTotal, setDataTotal] = React.useState(0);
  const [errorToast, setErrorToast] = React.useState('');
  const [rows, setRows] = React.useState<SearchStringItem[] | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [detailDefinitionIndex, setDetailDefinitionIndex] = React.useState(0);

  const handleChangePage = (
    event: React.MouseEvent<HTMLButtonElement> | null,
    newPage: number
  ) => {
    setPage(newPage);
    searchStringPagination(newPage.toString(10), pageSize.toString(10));
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    setPageSize(parseInt(event.target.value, 10));
    setPage(0);
    searchStringPagination('0', event.target.value);
  };

  const StyledTableCell = styled(TableCell)(({ theme }) => ({
    [`&.${tableCellClasses.head}`]: {
      backgroundColor: theme.palette.common.black,
      color: theme.palette.common.white,
    },
    [`&.${tableCellClasses.body}`]: {
      fontSize: 14,
    },
  }));

  const StyledTableRow = styled(TableRow)(({ theme }) => ({
    '&:nth-of-type(odd)': {
      backgroundColor: theme.palette.action.hover,
    },
    // hide last border
    '&:last-child td, &:last-child th': {
      border: 0,
    },
  }));

  const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
    props,
    ref
  ) {
    return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
  });

  const searchStringPagination = async (page: string, pageSize: string) => {
    const password = shared.getPassword();
    if (password.length < 1) {
      setErrorToast('Password is empty');
      return;
    }

    setLoading(true);

    try {
      let param = new URLSearchParams({
        password,
        bookID,
        page,
        pageSize,
        searchText: params.slug.trim(),
        type,
      });
      const res = await fetch(
        `${shared.baseAPIURL}/book/search?` + param.toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setDataTotal(data.data.total);
      setRows(data.data.items);
    } catch (error) {
      const message = (error as Error).message;
      setErrorToast(message);
    }
  };

  const [openDetail, setOpenDetail] = React.useState(false);

  const handleDetailDialogClose = () => {
    setOpenDetail(false);
  };

  useEffect(() => {
    searchStringPagination(page.toString(10), pageSize.toString(10));
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
        <TextField
          label="BookID"
          defaultValue={bookID}
          value={bookID}
          onChange={(e) => {
            setBookID(e.target.value);
          }}
          variant="standard"
          sx={{ width: '33%' }}
        />
        <FormControl
          variant="standard"
          sx={{ width: '33%', m: 1, minWidth: 120 }}
        >
          <InputLabel>StringType</InputLabel>
          <Select
            value={type}
            onChange={(e) => {
              setType(e.target.value);
            }}
            label="StringType"
          >
            <MenuItem value="">
              <em>None</em>
            </MenuItem>
            <MenuItem value={'word'}>Word</MenuItem>
            <MenuItem value={'phrase'}>Phrase</MenuItem>
            <MenuItem value={'form'}>Form</MenuItem>
          </Select>
        </FormControl>
        <Button
          variant="contained"
          onClick={() => {
            searchStringPagination(page.toString(10), pageSize.toString(10));
          }}
        >
          Search
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
        <TableContainer component={Paper}>
          <Table
            sx={{
              minWidth: 700,
            }}
            aria-label="customized table"
          >
            <EnhancedTableHead />
            {rows && (
              <TableBody>
                {rows.map((row) => (
                  <StyledTableRow key={row.definitionID}>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '200px',
                      }}
                    >
                      {row.string}
                    </StyledTableCell>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '150px',
                      }}
                    >
                      {row.partOfSpeech}
                    </StyledTableCell>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '600px',
                      }}
                    >
                      {row.definition}
                    </StyledTableCell>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '150px',
                      }}
                    >
                      {row.level}
                    </StyledTableCell>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '150px',
                      }}
                    >
                      {row.type}
                    </StyledTableCell>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '150px',
                      }}
                    >
                      {row.bookID}
                    </StyledTableCell>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '150px',
                      }}
                    >
                      {row.index}
                    </StyledTableCell>
                    <StyledTableCell
                      align="left"
                      sx={{
                        width: '100px',
                      }}
                    >
                      {/* <Button
                        onClick={() => {
                          setDetailBookID(row.bookID);
                          setOpenDetail(true);
                          setDetailDefinitionIndex(row.index);
                        }}
                      >
                        Detail
                      </Button> */}
                      <Link
                        href={`/preview?definitionID=${row.definitionID}`}
                        target="_blank"
                      >
                        Jump
                      </Link>
                    </StyledTableCell>
                  </StyledTableRow>
                ))}
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
          count={dataTotal}
          page={page}
          onPageChange={handleChangePage}
          rowsPerPage={pageSize}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Stack>

      <Dialog
        fullWidth
        maxWidth="xl"
        open={openDetail}
        onClose={handleDetailDialogClose}
      >
        <DialogContent>
          <Stack spacing={5} alignItems="center">
            <Typography variant="h5">Definition Detail</Typography>
            <Preview
              read_only
              read_book_id={detailBookID}
              read_definition_index={detailDefinitionIndex}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            onClick={() => {
              setOpenDetail(false);
            }}
          >
            Close
          </Button>
        </DialogActions>
      </Dialog>

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
