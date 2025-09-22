'use client';
import React from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import { post } from '@/app/shared';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import { Category } from '@/app/model';
import { SiteIDContext } from '@/app/site/[site]/admin/context';

const CategoryDialog = ({
  open,
  setOpen,
  siteID,
  categoryID,
  isDelete,
  isCreate,
  name,
  setName,
  getCategories,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
  siteID: number;
  categoryID: number;
  isDelete: boolean;
  isCreate: boolean;
  name: string;
  setName: (name: string) => void;
  getCategories: () => void;
}) => {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const handleClose = () => {
    setOpen(false);
  };

  const addCategory = () => {
    post(
      false,
      '/invoker/site/addCategory/v1',
      setLoading,
      {
        siteID: siteID,
        name: name,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        getCategories();
        setOpen(false);
      },
      setToast,
      setErrorToast
    );
  };

  const editCategory = () => {
    post(
      false,
      '/invoker/site/editCategory/v1',
      setLoading,
      {
        id: categoryID,
        name: name,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        getCategories();
        setOpen(false);
      },
      setToast,
      setErrorToast
    );
  };

  const deleteCategory = () => {
    post(
      false,
      '/invoker/site/deleteCategory/v1',
      setLoading,
      {
        id: categoryID,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        getCategories();
        setOpen(false);
      },
      setToast,
      setErrorToast
    );
  };

  return (
    <Dialog maxWidth="lg" fullWidth open={open} onClose={handleClose}>
      {!isDelete && (
        <>
          <DialogTitle>Properties</DialogTitle>
          <DialogContent>
            <Stack direction="row" alignItems="center">
              <Typography variant="body1" gutterBottom sx={{ width: '25%' }}>
                Name
              </Typography>
              <TextField
                id="name"
                name="name"
                fullWidth
                margin="normal"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                }}
              />
            </Stack>
          </DialogContent>
        </>
      )}
      {isDelete && (
        <DialogContent>
          <Typography variant="h6" gutterBottom>
            Are you sure to delete {name} ?
          </Typography>
        </DialogContent>
      )}
      <DialogActions>
        <Button
          onClick={() => {
            handleClose();
          }}
        >
          Close
        </Button>
        <Button
          onClick={() => {
            if (isDelete) {
              deleteCategory();
            } else {
              if (isCreate) {
                addCategory();
              } else {
                editCategory();
              }
            }
          }}
        >
          Confirm
        </Button>
      </DialogActions>
      <CustomSnackbar
        toast={toast}
        setToast={setToast}
        errorToast={errorToast}
        setErrorToast={setErrorToast}
      />
      <CustomBackdrop loading={loading} />
    </Dialog>
  );
};

interface Column {
  id: string;
  label: string;
  minWidth?: number;
  align?: 'right';
  format?: (value: number) => string;
}

const columns: readonly Column[] = [
  { id: 'id', label: 'ID', minWidth: 100 },
  { id: 'name', label: 'Name', minWidth: 200 },
  { id: 'operations', label: 'Operations', minWidth: 200 },
];

export default function Page() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const siteID = React.useContext(SiteIDContext);
  const [categories, setCategories] = React.useState<Category[] | null>(null);
  const getCategories = () => {
    console.log('siteID:', siteID);
    if (!siteID) {
      return;
    }
    post(
      false,
      '/invoker/site/getCategories/v1',
      setLoading,
      { siteID: siteID },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setCategories(respData.data.list);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const [openCategoryDialog, setOpenCategoryDialog] = React.useState(false);
  const [isDelete, setIsDelete] = React.useState(false);
  const [isCreate, setIsCreate] = React.useState(false);
  const [name, setName] = React.useState('');
  const [categoryID, setCategoryID] = React.useState(0);

  React.useEffect(() => {
    if (siteID) {
      getCategories();
    }
  }, [siteID]);

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
            onClick={getCategories}
            variant="contained"
            sx={{ margin: '5px' }}
          >
            Refresh
          </Button>
          <Button
            onClick={() => {
              setOpenCategoryDialog(true);
              setIsCreate(true);
              setIsDelete(false);
              setCategoryID(0);
              setName('');
            }}
            variant="contained"
            sx={{ margin: '5px' }}
          >
            Create
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
              {categories && (
                <TableBody>
                  {categories.map((row: any) => {
                    return (
                      <TableRow
                        hover
                        role="checkbox"
                        tabIndex={-1}
                        key={row.id}
                      >
                        {columns.map((column) => {
                          const value = row[column.id];
                          if (column.id === 'operations') {
                            return (
                              <TableCell key={column.id} align={column.align}>
                                <Button
                                  size="small"
                                  color="primary"
                                  variant="text"
                                  onClick={() => {
                                    setOpenCategoryDialog(true);
                                    setIsCreate(false);
                                    setIsDelete(false);
                                    setCategoryID(row.id);
                                    setName(row.name);
                                  }}
                                >
                                  Update
                                </Button>
                                <Button
                                  size="small"
                                  color="warning"
                                  variant="text"
                                  onClick={() => {
                                    setOpenCategoryDialog(true);
                                    setIsDelete(true);
                                    setCategoryID(row.id);
                                    setName(row.name);
                                  }}
                                >
                                  Delete
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
      </Stack>

      {/* category dialog */}
      <CategoryDialog
        open={openCategoryDialog}
        setOpen={setOpenCategoryDialog}
        siteID={siteID}
        categoryID={categoryID}
        isCreate={isCreate}
        isDelete={isDelete}
        getCategories={getCategories}
        name={name}
        setName={setName}
      />

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
