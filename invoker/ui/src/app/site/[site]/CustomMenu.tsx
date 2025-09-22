'use client';
import React from 'react';
import { useRouter, usePathname } from 'next/navigation';
import ListItemText from '@mui/material/ListItemText';
import Divider from '@mui/material/Divider';
import MenuList from '@mui/material/MenuList';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import Stack from '@mui/material/Stack';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import SquareIcon from '@mui/icons-material/Square';
import FormatListBulletedIcon from '@mui/icons-material/FormatListBulleted';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';
import MenuIcon from '@mui/icons-material/Menu';
import { styled } from '@mui/material/styles';
import IconButton from '@mui/material/IconButton';
import Link from '@mui/material/Link';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import { Category } from '@/app/model';

const CustomMenuIcon = styled(IconButton)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {},
  [theme.breakpoints.up('sm')]: {
    display: 'none',
  },
}));

const CustomMenuItem = ({
  curText,
  text,
  icon,
  onClick,
}: {
  curText: string;
  text: string;
  icon: React.ReactElement;
  onClick?: (e: any) => void;
}) => {
  return (
    <MenuItem
      component={Link}
      sx={
        curText === text
          ? {
              backgroundColor: '#27a1f1',
            }
          : {}
      }
      onClick={(e) => {
        onClick && onClick(e);
      }}
    >
      <ListItemIcon>{icon}</ListItemIcon>
      <ListItemText>{text}</ListItemText>
    </MenuItem>
  );
};

const CustomMenuList = styled(MenuList)(({ theme }) => ({
  [theme.breakpoints.down('sm')]: {
    display: 'none',
  },
  [theme.breakpoints.up('sm')]: {
    float: 'left',
    width: '20%',
    position: 'fixed',
  },
}));

export default function CustomMenu({
  curText,
  categories,
  isSiteAdmin,
}: {
  curText: string;
  categories: null | Category[];
  isSiteAdmin: boolean;
}) {
  const [expand, setExpand] = React.useState(true);

  const [open, setOpen] = React.useState(false);
  const toggleDrawer = (newOpen: boolean) => () => {
    setOpen(newOpen);
  };

  const pathname = usePathname();
  const router = useRouter();

  return (
    <>
      {/* menu container */}
      <Stack direction="row">
        <CustomMenuIcon
          color="primary"
          onClick={() => {
            setOpen(true);
          }}
          size="large"
        >
          <MenuIcon />
        </CustomMenuIcon>
        <SwipeableDrawer
          anchor="top"
          onClose={toggleDrawer(false)}
          onOpen={toggleDrawer(true)}
          disableSwipeToOpen={false}
          ModalProps={{
            keepMounted: true,
          }}
          open={open}
        >
          <MenuList>
            <CustomMenuItem
              curText={curText}
              text="Categories"
              icon={
                expand ? (
                  <ExpandMoreIcon fontSize="small" />
                ) : (
                  <ChevronRightIcon fontSize="small" />
                )
              }
              onClick={() => {
                setExpand(!expand);
              }}
            />
            {expand &&
              categories &&
              categories.map((item, idx) => {
                return (
                  <CustomMenuItem
                    curText={curText}
                    key={item.id}
                    text={item.name}
                    icon={<SquareIcon fontSize="small" />}
                    onClick={() => {
                      let matches = pathname.match('/site/\\w+');
                      if (matches) {
                        router.push(matches[0] + '/category?name=' + item.name);
                      }
                      setOpen(false);
                    }}
                  />
                );
              })}
            <Divider />
            <CustomMenuItem
              curText={curText}
              text="All Categories"
              icon={<FormatListBulletedIcon fontSize="small" />}
              onClick={() => {
                let matches = pathname.match('/site/\\w+');
                if (matches) {
                  router.push(matches[0]);
                }
                setOpen(false);
              }}
            />
            {isSiteAdmin && (
              <>
                <Divider />
                <CustomMenuItem
                  curText={curText}
                  text="Admin"
                  icon={<AdminPanelSettingsIcon fontSize="small" />}
                  onClick={() => {
                    let matches = pathname.match('/site/\\w+');
                    if (matches && matches.length > 0) {
                      router.push(matches[0] + '/admin');
                    }
                  }}
                />
              </>
            )}
          </MenuList>
        </SwipeableDrawer>
      </Stack>
      <CustomMenuList>
        <CustomMenuItem
          curText={curText}
          text="Categories"
          icon={
            expand ? (
              <ExpandMoreIcon fontSize="small" />
            ) : (
              <ChevronRightIcon fontSize="small" />
            )
          }
          onClick={() => {
            setExpand(!expand);
          }}
        />
        {expand &&
          categories &&
          categories.map((item, idx) => {
            return (
              <CustomMenuItem
                curText={curText}
                key={item.id}
                text={item.name}
                icon={<SquareIcon fontSize="small" />}
                onClick={() => {
                  let matches = pathname.match('/site/\\w+');
                  if (matches) {
                    router.push(matches[0] + '/category?name=' + item.name);
                  }
                }}
              />
            );
          })}
        <Divider />
        <CustomMenuItem
          curText={curText}
          text="All Categories"
          icon={<FormatListBulletedIcon fontSize="small" />}
          onClick={() => {
            let matches = pathname.match('/site/\\w+');
            if (matches) {
              router.push(matches[0]);
            }
          }}
        />
        {isSiteAdmin && (
          <>
            <Divider />
            <CustomMenuItem
              curText={curText}
              text="Admin"
              icon={<AdminPanelSettingsIcon fontSize="small" />}
              onClick={() => {
                let matches = pathname.match('/site/\\w+');
                if (matches && matches.length > 0) {
                  router.push(matches[0] + '/admin');
                }
              }}
            />
          </>
        )}
      </CustomMenuList>
    </>
  );
}
