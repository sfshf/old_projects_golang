'use client';
import * as React from 'react';
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
import Box from '@mui/material/Box';
import TableSortLabel from '@mui/material/TableSortLabel';
import { visuallyHidden } from '@mui/utils';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import Backdrop from '@mui/material/Backdrop';
import CircularProgress from '@mui/material/CircularProgress';

interface OperateLogDTO {
  id: number;
  operator: string;
  operateTime: string;
  operateStatus: number;
  operateType: number;
  bookID: number;
  definitionID: number;
  otherOperateParams: string;
  error: string;
}

interface OrderProperty {
  id: number;
  operator: string;
  created_at: string;
  operate_status: number;
  operate_type: number;
  book_id: number;
  definition_id: number;
  other_operate_params: string;
  error: string;
}

type Order = 'asc' | 'desc';

interface EnhancedTableProps {
  numSelected: number;
  onRequestSort: (
    event: React.MouseEvent<unknown>,
    property: keyof OrderProperty
  ) => void;
  onSelectAllClick: (event: React.ChangeEvent<HTMLInputElement>) => void;
  order: Order;
  orderBy: string;
  rowCount: number;
}

interface HeadCell {
  disablePadding: boolean;
  id: keyof OrderProperty;
  label: string;
}

const headCells: readonly HeadCell[] = [
  {
    id: 'id',
    disablePadding: false,
    label: 'ID',
  },
  {
    id: 'book_id',
    disablePadding: false,
    label: 'BookID',
  },
  {
    id: 'operator',
    disablePadding: false,
    label: 'Operator',
  },
  {
    id: 'operate_type',
    disablePadding: false,
    label: 'OperateType',
  },
  {
    id: 'operate_status',
    disablePadding: false,
    label: 'OperateStatus',
  },
  {
    id: 'error',
    disablePadding: false,
    label: 'Error',
  },
  {
    id: 'created_at',
    disablePadding: false,
    label: 'OperateTime',
  },
];

function EnhancedTableHead(props: EnhancedTableProps) {
  const {
    onSelectAllClick,
    order,
    orderBy,
    numSelected,
    rowCount,
    onRequestSort,
  } = props;
  const createSortHandler =
    (property: keyof OrderProperty) => (event: React.MouseEvent<unknown>) => {
      onRequestSort(event, property);
    };

  return (
    <TableHead>
      <TableRow>
        {headCells.map((headCell) => (
          <TableCell
            key={headCell.id}
            align="right"
            padding={headCell.disablePadding ? 'none' : 'normal'}
            sortDirection={orderBy === headCell.id ? order : false}
          >
            <TableSortLabel
              active={orderBy === headCell.id}
              direction={orderBy === headCell.id ? order : 'asc'}
              onClick={createSortHandler(headCell.id)}
            >
              {headCell.label}
              {orderBy === headCell.id ? (
                <Box component="span" sx={visuallyHidden}>
                  {order === 'desc' ? 'sorted descending' : 'sorted ascending'}
                </Box>
              ) : null}
            </TableSortLabel>
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}

export default function Page() {
  const [operators, setOperators] = React.useState<string[] | null>(null);
  const [operator, setOperator] = React.useState('');
  const [bookID, setBookID] = React.useState('');
  const [operateStatus, setOperateStatus] = React.useState('');
  const [definitionID, setDefinitionID] = React.useState('');
  const [page, setPage] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(30);
  const [dataTotal, setDataTotal] = React.useState(0);
  const [errorToast, setErrorToast] = React.useState('');
  const [rows, setRows] = React.useState<OperateLogDTO[]>([]);

  const [order, setOrder] = React.useState<Order>('desc');
  const [orderBy, setOrderBy] =
    React.useState<keyof OrderProperty>('created_at');
  const [selected, setSelected] = React.useState<readonly string[]>([]);

  const [loading, setLoading] = React.useState(false);

  const handleOperatorSelectChange = (event: SelectChangeEvent) => {
    setOperator(event.target.value);
  };
  const handleOperateStatusSelectChange = (event: SelectChangeEvent) => {
    setOperateStatus(event.target.value);
  };

  const handleRequestSort = (
    event: React.MouseEvent<unknown>,
    property: keyof OrderProperty
  ) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
    getPagination(
      page.toString(10),
      pageSize.toString(10),
      isAsc ? 'desc' : 'asc',
      property
    );
  };

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelected = rows.map((n) => n.id.toString());
      setSelected(newSelected);
      return;
    }
    setSelected([]);
  };

  const handleChangePage = (
    event: React.MouseEvent<HTMLButtonElement> | null,
    newPage: number
  ) => {
    setPage(newPage);
    getPagination(newPage.toString(10), pageSize.toString(10), order, orderBy);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    setPageSize(parseInt(event.target.value, 10));
    setPage(0);
    getPagination('0', event.target.value, order, orderBy);
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

  const getPagination = async (
    page: string,
    pageSize: string,
    order: string,
    orderBy: string
  ) => {
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
        bookID,
        operateStatus,
        definitionID,
        page,
        pageSize,
        order,
        orderBy,
      });
      const res = await fetch(
        `${shared.baseAPIURL}/operate_log/pagination?` + param.toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setErrorToast(data.message);
        return;
      }
      setDataTotal(data.data.total);
      setRows(data.data.operateLogs);
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
    getStaffList(false);
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
          value={bookID}
          onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
            setBookID(event.target.value);
          }}
          variant="standard"
          sx={{ width: '25%' }}
        />
        <FormControl
          variant="standard"
          sx={{ width: '25%', m: 1, minWidth: 120 }}
        >
          <InputLabel>Operator</InputLabel>
          <Select
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
                    <em>{item}</em>
                  </MenuItem>
                );
              })}
          </Select>
        </FormControl>
        <FormControl
          variant="standard"
          sx={{ width: '25%', m: 1, minWidth: 120 }}
        >
          <InputLabel>OperateStatus</InputLabel>
          <Select
            value={operateStatus}
            onChange={handleOperateStatusSelectChange}
            label="OperateStatus"
          >
            <MenuItem value="">
              <em>None</em>
            </MenuItem>
            <MenuItem value={1}>Failure</MenuItem>
            <MenuItem value={2}>Success</MenuItem>
          </Select>
        </FormControl>
        <Button
          variant="contained"
          onClick={() => {
            getPagination(page.toString(), pageSize.toString(), order, orderBy);
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
          <Table sx={{ minWidth: 700 }} aria-label="customized table">
            <EnhancedTableHead
              numSelected={selected.length}
              order={order}
              orderBy={orderBy}
              onSelectAllClick={handleSelectAllClick}
              onRequestSort={handleRequestSort}
              rowCount={rows.length}
            />
            <TableBody>
              {rows.map((row) => (
                <StyledTableRow key={row.id}>
                  <StyledTableCell align="right">{row.id}</StyledTableCell>
                  <StyledTableCell align="right">{row.bookID}</StyledTableCell>
                  <StyledTableCell align="right">
                    {row.operator}
                  </StyledTableCell>
                  <StyledTableCell align="right">
                    {row.operateType}
                  </StyledTableCell>
                  <StyledTableCell align="right">
                    {row.operateStatus === 2
                      ? 'success'
                      : row.operateStatus === 1
                      ? 'failure'
                      : 'unknow'}
                  </StyledTableCell>
                  <StyledTableCell align="right">{row.error}</StyledTableCell>
                  <StyledTableCell align="right">
                    {row.operateTime}
                  </StyledTableCell>
                </StyledTableRow>
              ))}
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
