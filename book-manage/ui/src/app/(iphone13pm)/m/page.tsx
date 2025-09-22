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
import Link from 'next/link';
import Stack from '@mui/material/Stack';
import { usePathname, useRouter } from 'next/navigation';

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
  { id: 'id', label: 'BookID', minWidth: 100 },
  { id: 'name', label: 'Book\u00a0Name', minWidth: 100 },
  { id: 'updatedAtText', label: 'Update\u00a0At', minWidth: 100 },
  { id: 'updatedAt', label: 'timestamp', minWidth: 100 },
  { id: 'description', label: 'description', minWidth: 100 },
];

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [list, setList] = React.useState<any[] | null>(null);

  async function fetchData() {
    const password = shared.getPassword();
    if (password.length < 1) {
      setToast('Password is empty');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(
        `${baseURL}/book/allbooks?` +
          new URLSearchParams({ password }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setToast(data.message);
        return;
      }
      const books = data.data.books;
      books.forEach((book: any) => {
        book.updatedAtText = new Date(book.updatedAt).toLocaleString();
      });
      setList(books);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setToast(message);
    }
  }

  async function downloadBook(bookID: number, name: string) {
    const password = shared.getPassword();
    if (password.length < 1) {
      setToast('Password is empty');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(
        `${baseURL}/book/csv?` +
          new URLSearchParams({ password, book: String(bookID) }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        // console.log(data.message);
        setToast(data.message);
        return;
      }
      const path = data.data.path;
      const url = shared.baseAPIURL + path;
      var link = document.createElement('a');
      link.download = name + '.csv';
      link.target = '_blank';
      // Construct the URI
      link.href = url;
      document.body.appendChild(link);
      link.click();
      // Cleanup the DOM
      document.body.removeChild(link);
    } catch (error) {
      // console.log(error);
      const message = (error as Error).message;
      setLoading(false);
      setToast(message);
    }
  }

  async function downloadBundle(bookID: number, name: string) {
    const password = shared.getPassword();
    if (password.length < 1) {
      setToast('Password is empty');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(
        `${baseURL}/book/bundle?` +
          new URLSearchParams({ password, book: String(bookID) }).toString()
      );
      setLoading(false);
      const data = await res.json();
      if (data.code !== 0) {
        setToast(data.message);
        return;
      }
      const path = data.data.path;
      const url = shared.baseAPIURL + path;
      var link = document.createElement('a');
      link.download = name + '.bundle';
      link.target = '_blank';
      // Construct the URI
      link.href = url;
      document.body.appendChild(link);
      link.click();
      // Cleanup the DOM
      document.body.removeChild(link);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setToast(message);
    }
  }

  const router = useRouter();
  const pathName = usePathname();
  React.useEffect(() => {
    console.log(pathName);
    if (window.innerWidth > 600) {
      router.push('/', { scroll: false });
    }
  }, []);

  return (
    <main>
      <Stack
        spacing={2}
        width="25%"
        sx={{
          alignItems: 'center',
          padding: '20px',
        }}
      >
        <Button onClick={fetchData} variant="contained" size="small">
          Refresh
        </Button>
      </Stack>
      <Stack
        spacing={2}
        width="100%"
        sx={{
          alignItems: 'center',
          padding: '20px',
        }}
      >
        {list !== null ? (
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
                    key="download"
                    align="center"
                    style={{ minWidth: 100 }}
                  >
                    Download
                    {/* <Button variant="contained">Download</Button> */}
                  </TableCell>
                  <TableCell
                    key="update"
                    align="center"
                    style={{ minWidth: 100 }}
                  >
                    Update
                  </TableCell>
                  <TableCell
                    key="preview"
                    align="center"
                    style={{ minWidth: 100 }}
                  >
                    Preview
                  </TableCell>
                  <TableCell
                    key="bundle"
                    align="center"
                    style={{ minWidth: 100 }}
                  >
                    Bundle
                  </TableCell>
                </TableRow>
              </TableHead>
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
                      >
                        <Button
                          variant="contained"
                          onClick={() => downloadBook(row.id, row.name)}
                        >
                          Download
                        </Button>
                      </TableCell>

                      <TableCell
                        key="update"
                        align="center"
                        style={{ minWidth: 100 }}
                      >
                        <Button
                          variant="contained"
                          href={`/m/update?bookID=${row.id}`}
                          LinkComponent={Link}
                        >
                          Update
                        </Button>
                      </TableCell>
                      <TableCell
                        key="preview"
                        align="center"
                        style={{ minWidth: 100 }}
                      >
                        <Button
                          href={`/m/preview?bookID=${row.id}`}
                          LinkComponent={Link}
                          variant="contained"
                        >
                          Preview
                        </Button>
                      </TableCell>
                      <TableCell
                        key="bundle"
                        align="center"
                        style={{ minWidth: 100 }}
                      >
                        <Button
                          variant="contained"
                          color="error"
                          onClick={() => downloadBundle(row.id, row.name)}
                        >
                          Bundle
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>
        ) : null}
      </Stack>

      <Snackbar
        open={toast !== ''}
        autoHideDuration={4500}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
        onClose={() => setToast('')}
      >
        <Alert
          onClose={() => setToast('')}
          severity="error"
          sx={{ width: '100%' }}
        >
          {toast}
        </Alert>
      </Snackbar>
      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
        // onClick={handleClose}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
    </main>
  );
}
