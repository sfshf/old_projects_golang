'use client';
import React from 'react';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import { post } from '@/app/shared';
import {
  removeLocalUserLoginState,
  setLocalUserLoginState,
  getLocalUserLoginState,
  LoginState,
} from '@/app/site/shared';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import Logout from '@mui/icons-material/Logout';
import PersonIcon from '@mui/icons-material/Person';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import TextField from '@mui/material/TextField';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import InputAdornment from '@mui/material/InputAdornment';
import FormControl from '@mui/material/FormControl';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import { keccak_256 } from '@noble/hashes/sha3';
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import CloseIcon from '@mui/icons-material/Close';
import DialogTitle from '@mui/material/DialogTitle';
import ListAltIcon from '@mui/icons-material/ListAlt';
import { useRouter, usePathname, useParams } from 'next/navigation';
import CustomAvatar from '@/app/site/CustomAvatar';
import CustomSnackbar from '@/components/CustomSnackbar';
import CustomBackdrop from '@/components/CustomBackdrop';
import SearchIcon from '@mui/icons-material/Search';
import Popover from '@mui/material/Popover';
import TuneIcon from '@mui/icons-material/Tune';
import { styled } from '@mui/material/styles';

const LoginDialog = ({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) => {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState('');
  const [errorToast, setErrorToast] = React.useState('');

  const [email, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [showPassword, setShowPassword] = React.useState(false);
  const handleClickShowPassword = () => setShowPassword((show) => !show);
  const handleMouseDownPassword = (
    event: React.MouseEvent<HTMLButtonElement>
  ) => {
    event.preventDefault();
  };
  const handleDialogClose = () => {
    setOpen(false);
    setEmail('');
    setPassword('');
  };
  const loginByEmail = (email: string, password: string) => {
    let passwordHash = Buffer.from(keccak_256(password)).toString('hex');
    post(
      false,
      '/slark/user/loginByEmail/v1',
      setLoading,
      { email: email, passwordHash: passwordHash },
      (respHeaders: any) => {},
      (respData: any) => {
        if (respData.data) {
          setLocalUserLoginState(respData.data);
          let loginState = getLocalUserLoginState();
          if (!loginState) {
            setErrorToast(
              'get local user login state fail, after loginByEmail'
            );
          }
        }
        setOpen(false);
        location.reload();
      },
      undefined,
      setErrorToast
    );
  };

  return (
    <Dialog fullWidth maxWidth="sm" open={open} onClose={handleDialogClose}>
      <DialogTitle>Log In Dialog</DialogTitle>
      <IconButton
        onClick={handleDialogClose}
        sx={(theme) => ({
          position: 'absolute',
          right: 8,
          top: 8,
          color: theme.palette.grey[500],
        })}
      >
        <CloseIcon />
      </IconButton>
      <DialogContent>
        <Stack direction="row" textAlign="center" justifyContent="center">
          <TextField
            id="outlined-basic"
            label="Email"
            variant="outlined"
            sx={{ m: 1, width: '500px' }}
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
            }}
          />
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="center">
          <FormControl sx={{ m: 1, width: '500px' }} variant="outlined">
            <InputLabel htmlFor="outlined-adornment-password">
              Password
            </InputLabel>
            <OutlinedInput
              id="outlined-adornment-password"
              type={showPassword ? 'text' : 'password'}
              endAdornment={
                <InputAdornment position="end">
                  <IconButton
                    aria-label="toggle password visibility"
                    onClick={handleClickShowPassword}
                    onMouseDown={handleMouseDownPassword}
                    edge="end"
                  >
                    {showPassword ? <VisibilityOff /> : <Visibility />}
                  </IconButton>
                </InputAdornment>
              }
              label="Password"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
              }}
            />
          </FormControl>
        </Stack>
        <Stack direction="row" textAlign="center" justifyContent="center">
          <Button
            variant="contained"
            sx={{ m: 1, width: '500px' }}
            onClick={() => {
              loginByEmail(email, password);
            }}
          >
            Sign in
          </Button>
        </Stack>
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

const CustomTypography = styled(Typography)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {
    fontSize: '26px',
  },
  [theme.breakpoints.up('sm')]: {},
}));

const CustomSearchStack = styled(Stack)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {
    margin: '10px',
    width: '300px',
  },
  [theme.breakpoints.up('sm')]: { margin: '20px', width: '600px' },
}));

export default function CustomHead({
  headText,
  loginInfo,
  siteID,
  setLoading,
  setErrorToast,
}: {
  headText: string;
  loginInfo: null | LoginState;
  siteID: null | number;
  setLoading: (loadin: boolean) => void;
  setErrorToast: (errorToast: string) => void;
}) {
  const router = useRouter();
  const pathname = usePathname();

  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = () => {
    setAnchorEl(null);
  };
  const handleClickLogout = () => {
    setAnchorEl(null);
    // LogOutBySession
    post(
      false,
      '/slark/user/logout/v1',
      setLoading,
      null,
      (respHeaders: any) => {},
      (respData: any) => {
        location.reload();
      },
      undefined,
      setErrorToast
    );
    removeLocalUserLoginState();
  };
  const [openLogin, setOpenLogin] = React.useState(false);
  const handleClickMyPostHistory = () => {
    setAnchorEl(null);
    let matches = pathname.match('/site/\\w+');
    if (matches) {
      router.push(matches[0] + '/user/post-history');
    }
  };
  const handleClickMyCommentHistory = () => {
    setAnchorEl(null);
    let matches = pathname.match('/site/\\w+');
    if (matches) {
      router.push(matches[0] + '/user/comment-history');
    }
  };
  const handleClickMyThumbupHistory = () => {
    setAnchorEl(null);
    let matches = pathname.match('/site/\\w+');
    if (matches) {
      router.push(matches[0] + '/user/thumbup-history');
    }
  };

  const [searchAnchorEl, setSearchAnchorEl] =
    React.useState<null | HTMLElement>(null);
  const openSearch = Boolean(searchAnchorEl);
  const handleSearchTooltipOpen = (event: React.MouseEvent<HTMLElement>) => {
    setSearchAnchorEl(event.currentTarget);
  };
  const handleSearchTooltipClose = () => {
    setSearchAnchorEl(null);
  };
  const [searchText, setSearchText] = React.useState('');
  const handleSearchTextChange = (e: any) => {
    setSearchText(e.target.value);
  };
  const handleSearch = (e: any) => {
    let matches = pathname.match('/site/\\w+');
    if (matches) {
      router.push(matches[0] + '/search?q=' + searchText);
    }
  };

  return (
    <>
      <AppBar>
        <Toolbar>
          <CustomTypography variant="h3" sx={{ flexGrow: 1 }}>
            {headText}
          </CustomTypography>
          {!loginInfo && (
            <Stack direction="row" spacing={1}>
              <Button
                variant="contained"
                size="small"
                onClick={() => {
                  setOpenLogin(true);
                }}
                startIcon={<PersonIcon />}
                sx={{
                  height: '40px',
                }}
              >
                Log In
              </Button>
            </Stack>
          )}
          {siteID && (
            <Stack
              direction="row"
              spacing={1}
              sx={{
                margin: '10px',
              }}
            >
              <Tooltip title="Search Post/Comment">
                <IconButton
                  size="large"
                  onClick={handleSearchTooltipOpen}
                  sx={{
                    color: 'white',
                    width: '40px',
                    height: '40px',
                    borderRadius: '0',
                  }}
                >
                  <SearchIcon />
                </IconButton>
              </Tooltip>
              <Popover
                anchorEl={searchAnchorEl}
                id="search-popover"
                open={openSearch}
                onClose={handleSearchTooltipClose}
                slotProps={{
                  paper: {
                    elevation: 0,
                    sx: {
                      overflow: 'visible',
                      filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
                      mt: 1.5,
                      '& .MuiAvatar-root': {
                        width: 32,
                        height: 32,
                        ml: -0.5,
                        mr: 1,
                      },
                      '&::before': {
                        content: '""',
                        display: 'block',
                        position: 'absolute',
                        top: 0,
                        right: 14,
                        width: 10,
                        height: 10,
                        bgcolor: 'background.paper',
                        transform: 'translateY(-50%) rotate(45deg)',
                        zIndex: 0,
                      },
                    },
                  },
                }}
                transformOrigin={{
                  horizontal: 'right',
                  vertical: 'top',
                }}
                anchorOrigin={{
                  horizontal: 'right',
                  vertical: 'bottom',
                }}
              >
                <CustomSearchStack direction="row" spacing={1}>
                  <FormControl fullWidth variant="outlined">
                    <InputLabel>Search</InputLabel>
                    <OutlinedInput
                      onChange={handleSearchTextChange}
                      endAdornment={
                        <InputAdornment position="end">
                          <IconButton
                            edge="end"
                            sx={{ borderRadius: '0' }}
                            onClick={handleSearch}
                          >
                            <SearchIcon color="primary" />
                          </IconButton>
                        </InputAdornment>
                      }
                      label="Search"
                    />
                  </FormControl>
                </CustomSearchStack>
              </Popover>
            </Stack>
          )}
          {loginInfo && (
            <Stack direction="row" spacing={1}>
              <Tooltip title="Account settings">
                <IconButton onClick={handleClick} size="large">
                  <CustomAvatar
                    text={loginInfo.nickname}
                    width="40px"
                    height="40px"
                  />
                </IconButton>
              </Tooltip>
              <Menu
                anchorEl={anchorEl}
                id="account-menu"
                open={open}
                onClose={handleClose}
                onClick={handleClose}
                slotProps={{
                  paper: {
                    elevation: 0,
                    sx: {
                      overflow: 'visible',
                      filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
                      mt: 1.5,
                      '& .MuiAvatar-root': {
                        width: 32,
                        height: 32,
                        ml: -0.5,
                        mr: 1,
                      },
                      '&::before': {
                        content: '""',
                        display: 'block',
                        position: 'absolute',
                        top: 0,
                        right: 14,
                        width: 10,
                        height: 10,
                        bgcolor: 'background.paper',
                        transform: 'translateY(-50%) rotate(45deg)',
                        zIndex: 0,
                      },
                    },
                  },
                }}
                transformOrigin={{
                  horizontal: 'right',
                  vertical: 'top',
                }}
                anchorOrigin={{
                  horizontal: 'right',
                  vertical: 'bottom',
                }}
              >
                <MenuItem onClick={handleClickLogout}>
                  <ListItemIcon>
                    <Logout fontSize="small" />
                  </ListItemIcon>
                  Logout
                </MenuItem>
                {siteID && (
                  <MenuItem onClick={handleClickMyPostHistory}>
                    <ListItemIcon>
                      <ListAltIcon fontSize="small" />
                    </ListItemIcon>
                    My Post History
                  </MenuItem>
                )}
                {siteID && (
                  <MenuItem onClick={handleClickMyCommentHistory}>
                    <ListItemIcon>
                      <ListAltIcon fontSize="small" />
                    </ListItemIcon>
                    My Comment History
                  </MenuItem>
                )}
                {siteID && (
                  <MenuItem onClick={handleClickMyThumbupHistory}>
                    <ListItemIcon>
                      <ListAltIcon fontSize="small" />
                    </ListItemIcon>
                    My Thumbup History
                  </MenuItem>
                )}
              </Menu>
            </Stack>
          )}
        </Toolbar>
      </AppBar>
      <Toolbar /> {/* important for position fixed AppBar */}
      {/* login dialog */}
      <LoginDialog open={openLogin} setOpen={setOpenLogin} />
    </>
  );
}
