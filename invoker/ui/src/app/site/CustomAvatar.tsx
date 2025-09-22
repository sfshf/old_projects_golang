'use client';
import React from 'react';
import Avatar from '@mui/material/Avatar';
import { blue } from '@mui/material/colors';

export default function CustomAvatar({
  text,
  width,
  height,
}: {
  text: string;
  width: string;
  height: string;
}) {
  return (
    <>
      <Avatar
        sx={{
          width: width,
          height: height,
          bgcolor: blue[500],
        }}
      >
        {text ? text.charAt(0) : 'A'}
      </Avatar>
    </>
  );
}
