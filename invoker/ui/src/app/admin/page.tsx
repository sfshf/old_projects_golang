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
import Button from '@mui/material/Button';
import { getApiKey, post } from '@/app/shared';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import IconButton from '@mui/material/IconButton';
import AddIcon from '@mui/icons-material/Add';
import CloseIcon from '@mui/icons-material/Close';
import CheckIcon from '@mui/icons-material/Check';
import Divider, { dividerClasses } from '@mui/material/Divider';
import Box from '@mui/material/Box';
import { Site } from '@/app/model';

interface AdminInfo {
  userID: number;
  valid: boolean;
}

const AdminListDialog = ({
  open,
  setOpen,
  siteID,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
  siteID: number;
}) => {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const handleClose = () => {
    setOpen(false);
  };

  const [userID, setUserID] = React.useState<string>('');
  const [adminInfos, setAdminInfos] = React.useState<AdminInfo[] | null>(null);
  const [addAdmin, setAddAdmin] = React.useState(false);

  const getSiteAdmins = (siteID: number) => {
    post(
      false,
      '/invoker/admin/getSiteAdmins/v1',
      setLoading,
      {
        apiKey: getApiKey(),
        id: siteID,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setAdminInfos(respData.data.list);
        } else {
          setAdminInfos(null);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const addSiteAdmin = () => {
    if (!userID) {
      return;
    }
    post(
      false,
      '/invoker/admin/addSiteAdmin/v1',
      setLoading,
      {
        apiKey: getApiKey(),
        id: siteID,
        userID: parseInt(userID),
      },
      (respHeaders: any) => {},
      (respData: any) => {
        setAddAdmin(false);
        setUserID('');
        getSiteAdmins(siteID);
      },
      setToast,
      setErrorToast
    );
  };

  const deleteSiteAdmin = (userID: number) => {
    post(
      false,
      '/invoker/admin/deleteSiteAdmin/v1',
      setLoading,
      {
        apiKey: getApiKey(),
        id: siteID,
        userID: userID,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        getSiteAdmins(siteID);
      },
      setToast,
      setErrorToast
    );
  };

  React.useEffect(() => {
    if (open) {
      getSiteAdmins(siteID);
    }
  }, [siteID]);

  return (
    <Dialog maxWidth="lg" fullWidth open={open} onClose={handleClose}>
      <IconButton
        aria-label="close"
        onClick={() => {
          setOpen(false);
        }}
        sx={{
          position: 'absolute',
          right: 8,
          top: 8,
          color: (theme) => theme.palette.grey[500],
        }}
      >
        <CloseIcon />
      </IconButton>
      <DialogTitle>Administrator UserIDs:</DialogTitle>
      <DialogContent>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            borderColor: 'divider',
            borderRadius: 1,
            bgcolor: 'background.paper',
            color: 'text.secondary',
            '& svg': {
              m: 1,
            },
            [`& .${dividerClasses.root}`]: {
              mx: 0.5,
            },
          }}
        >
          <Stack
            sx={{
              width: '60%',
            }}
            spacing={2}
          >
            <IconButton
              onClick={() => {
                setAddAdmin(true);
              }}
            >
              <AddIcon color="primary" />
            </IconButton>
            {addAdmin && (
              <Stack
                direction="row"
                alignItems="center"
                sx={{
                  width: '100%',
                }}
              >
                <TextField
                  id="outlined-basic"
                  label="UserID"
                  variant="outlined"
                  fullWidth
                  value={userID}
                  onChange={(e) => {
                    if (e.target.value) {
                      setUserID(e.target.value);
                    } else {
                      setUserID('');
                    }
                  }}
                />
                <IconButton
                  onClick={() => {
                    addSiteAdmin();
                  }}
                >
                  <CheckIcon color="success" />
                </IconButton>
                <IconButton
                  onClick={() => {
                    setAddAdmin(false);
                    setUserID('');
                  }}
                >
                  <CloseIcon color="error" />
                </IconButton>
              </Stack>
            )}
          </Stack>
          <Divider orientation="vertical" variant="middle" flexItem />
          <Stack spacing={2} sx={{ width: '100%' }}>
            <List>
              {adminInfos &&
                adminInfos.map((row) => {
                  return (
                    <ListItem
                      key={row.userID}
                      disablePadding
                      sx={{ height: '50px' }}
                    >
                      <ListItemButton sx={{ height: '50px' }}>
                        <ListItemText
                          primary={row.userID}
                          sx={
                            row.valid
                              ? {
                                  marginLeft: '10%',
                                  color: 'green',
                                }
                              : {
                                  marginLeft: '10%',
                                  color: 'red',
                                }
                          }
                        />
                        <IconButton
                          sx={{
                            marginRight: '10%',
                          }}
                          onClick={() => {
                            deleteSiteAdmin(row.userID);
                          }}
                        >
                          <CloseIcon color="error" />
                        </IconButton>
                      </ListItemButton>
                    </ListItem>
                  );
                })}
            </List>
          </Stack>
        </Box>
      </DialogContent>
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

const SiteDialog = ({
  open,
  setOpen,
  siteID,
  isDelete,
  isCreate,
  name,
  setName,
  getSites,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
  siteID: number;
  isDelete: boolean;
  isCreate: boolean;
  name: string;
  setName: (name: string) => void;
  getSites: () => void;
}) => {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const handleClose = () => {
    setOpen(false);
  };
  const addSite = () => {
    post(
      false,
      '/invoker/admin/addSite/v1',
      setLoading,
      {
        apiKey: getApiKey(),
        name: name,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        getSites();
        setOpen(false);
      },
      setToast,
      setErrorToast
    );
  };

  const editSite = () => {
    post(
      false,
      '/invoker/admin/editSite/v1',
      setLoading,
      {
        apiKey: getApiKey(),
        id: siteID,
        name: name,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        getSites();
        setOpen(false);
      },
      setToast,
      setErrorToast
    );
  };

  const deleteSite = () => {
    post(
      false,
      '/invoker/admin/deleteSite/v1',
      setLoading,
      {
        apiKey: getApiKey(),
        id: siteID,
      },
      (respHeaders: any) => {},
      (respData: any) => {
        getSites();
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
              deleteSite();
            } else {
              if (isCreate) {
                addSite();
              } else {
                editSite();
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

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [sites, setSites] = React.useState<Site[] | null>(null);

  const getSites = () => {
    post(
      false,
      '/invoker/admin/getSites/v1',
      setLoading,
      { apiKey: getApiKey() },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setSites(respData.data.list);
        }
      },
      undefined,
      setErrorToast
    );
  };

  const [openSiteDialog, setOpenSiteDialog] = React.useState(false);
  const [siteID, setSiteID] = React.useState(0);
  const [isDelete, setIsDelete] = React.useState(false);
  const [isCreate, setIsCreate] = React.useState(false);
  const [name, setName] = React.useState('');

  const [openAdminsDialog, setOpenAdminsDialog] = React.useState(false);

  React.useEffect(() => {
    getSites();
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
          <Button onClick={getSites} variant="contained" sx={{ margin: '5px' }}>
            Refresh
          </Button>
          <Button
            onClick={() => {
              setOpenSiteDialog(true);
              setIsCreate(true);
              setIsDelete(false);
              setSiteID(0);
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
                      sx={{ minWidth: column.minWidth }}
                    >
                      {column.label}
                    </TableCell>
                  ))}
                </TableRow>
              </TableHead>
              {sites && (
                <TableBody>
                  {sites.map((row: any) => {
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
                                  color="success"
                                  variant="text"
                                  onClick={() => {
                                    setSiteID(row.id);
                                    setOpenAdminsDialog(true);
                                  }}
                                >
                                  Admins
                                </Button>
                                <Button
                                  size="small"
                                  color="primary"
                                  variant="text"
                                  onClick={() => {
                                    setOpenSiteDialog(true);
                                    setIsCreate(false);
                                    setIsDelete(false);
                                    setSiteID(row.id);
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
                                    setOpenSiteDialog(true);
                                    setIsDelete(true);
                                    setSiteID(row.id);
                                    setName(row.name);
                                  }}
                                >
                                  Delete
                                </Button>
                              </TableCell>
                            );
                          } else if (column.id === 'name') {
                            return (
                              <TableCell key={column.id} align={column.align}>
                                <Link href={'/site/' + value} target="_self">
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

      {/* admin list dialog */}
      <AdminListDialog
        open={openAdminsDialog}
        setOpen={setOpenAdminsDialog}
        siteID={siteID}
      />
      {/* site dialog */}
      <SiteDialog
        open={openSiteDialog}
        setOpen={setOpenSiteDialog}
        siteID={siteID}
        isDelete={isDelete}
        isCreate={isCreate}
        name={name}
        setName={setName}
        getSites={getSites}
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
