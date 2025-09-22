'use client';
import React from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Stack from '@mui/material/Stack';
import { getApiKey, post } from '@/app/shared';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'database', label: 'Database', minWidth: 100 },
  { id: 'size', label: 'Size (MB)', minWidth: 100 },
];

interface SystemInfo {
  total: number;
  free: number;
}

interface DatabaseInfo {
  database: string;
  size: number;
}

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [systemInfo, setSystemInfo] = React.useState<null | SystemInfo>(null);
  const [databaseInfos, setDatabaseInfos] = React.useState<
    null | DatabaseInfo[]
  >(null);

  const getStorageInfo = () => {
    post(
      false,
      '/invoker/admin/getStorageInfo/v1',
      setLoading,
      { apiKey: getApiKey() },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setSystemInfo(respData.data.systemInfo);
          setDatabaseInfos(respData.data.databaseInfos);
        }
      },
      undefined,
      setErrorToast
    );
  };

  React.useEffect(() => {
    getStorageInfo();
  }, []);

  return (
    <>
      <Stack
        spacing={3}
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
            onClick={getStorageInfo}
            variant="contained"
            sx={{ margin: '5px' }}
          >
            Refresh
          </Button>
          <Typography variant="h5" gutterBottom>
            Total Disk (GB): {systemInfo && systemInfo.total}
          </Typography>
          <Typography variant="h5" gutterBottom>
            Free Disk (GB): {systemInfo && systemInfo.free}
          </Typography>
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
                      sx={{ minWidth: column.minWidth }}
                    >
                      {column.label}
                    </TableCell>
                  ))}
                </TableRow>
              </TableHead>
              {databaseInfos && (
                <TableBody>
                  {databaseInfos.map((row: any) => {
                    return (
                      <TableRow
                        hover
                        role="checkbox"
                        tabIndex={-1}
                        key={row.name}
                      >
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
                      </TableRow>
                    );
                  })}
                </TableBody>
              )}
            </Table>
          </TableContainer>
        </Stack>
      </Stack>

      <CustomSnackbar
        toast={toast}
        setToast={setToast}
        errorToast={errorToast}
        setErrorToast={setErrorToast}
      />
      <CustomBackdrop loading={loading} />
    </>
  );
}
