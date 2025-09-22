'use client';
import React from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Stack from '@mui/material/Stack';
import Link from '@mui/material/Link';

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'name', label: 'Name', minWidth: 100 },
  { id: 'url', label: 'URL', minWidth: 400 },
];

export default function Home() {
  const [testDeployList, setTestDeployList] = React.useState([
    {
      name: 'Tester',
      url: 'http://43.198.255.33:8080/',
    },
    {
      name: 'Oracle Console',
      url: 'http://3.214.8.157:8866/',
    },
    {
      name: 'Riki Console',
      url: 'https://riki.test.n1xt.net/',
    },
    {
      name: 'Alchemist Console',
      url: 'https://rp.test.n1xt.net/',
    },
    {
      name: 'Doom Console',
      url: 'http://43.198.255.33:80/',
    },
    {
      name: 'Doom Helper',
      url: 'https://doom.test.n1xt.net/',
    },
    {
      name: 'Book Manager',
      url: 'http://44.219.239.87/',
    },
    {
      name: 'Slark SSO',
      url: 'https://sso.test.n1xt.net/',
    },
    {
      name: 'Status',
      url: 'https://status.test.n1xt.net/',
    },
    {
      name: 'Forum',
      url: 'https://forum.test.n1xt.net/',
    },
  ]);

  return (
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
            {testDeployList && (
              <TableBody>
                {testDeployList.map((row: any) => {
                  return (
                    <TableRow
                      hover
                      role="checkbox"
                      tabIndex={-1}
                      key={row.name}
                    >
                      {columns.map((column) => {
                        const value = row[column.id];
                        if (column.id === 'url') {
                          return (
                            <TableCell key={column.id} align={column.align}>
                              <Link href={value} target="_blank">
                                {value}
                              </Link>
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
    </Stack>
  );
}
